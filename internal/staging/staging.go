package staging

import (
	"bufio"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/HalilFocic/gitgo/internal/blob"
	"github.com/HalilFocic/gitgo/internal/repository"
)

type Entry struct {
	Path     string
	Hash     string
	Mode     os.FileMode
	Size     int64
	Modified time.Time
}

type Index struct {
	entries map[string]*Entry
	root    string
}

func New(root string) (*Index, error) {
	absPath, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, errors.New("path does not exist")
	}
	if !repository.IsRepository(absPath) {
		return nil, errors.New("not a gitgo repository")
	}

	entries := make(map[string]*Entry)
	return &Index{
		root:    root,
		entries: entries,
	}, nil
}

func (idx *Index) Add(path string) error {
	absPath := filepath.Join(idx.root, path)
	objectsPath := filepath.Join(idx.root, ".gitgo/objects")
	relPath, err := filepath.Rel(idx.root, path)
	if err != nil {
		return err
	}
	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("path %s is outside repository", path)
	}

	fileStat, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if fileStat.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlinks are not supported")
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}
	b, err := blob.New(content)
	if err != nil {
		return err
	}

	err = b.Store(objectsPath)
	if err != nil {
		return err
	}
	entry := Entry{
		Path:     relPath,
		Hash:     b.Hash(),
		Mode:     fileStat.Mode(),
		Size:     fileStat.Size(),
		Modified: fileStat.ModTime(),
	}
	idx.entries[relPath] = &entry
	return nil
}

func (idx *Index) Remove(path string) error {
	relPath, err := filepath.Rel(idx.root, path)
	if err != nil {
		return err
	}
	_, ok := idx.entries[relPath]
	if !ok {
		return fmt.Errorf("File is not in index entries")
	}
	delete(idx.entries, relPath)
	return nil
}

func (idx *Index) Entries() []*Entry {
	slice := make([]*Entry, 0, len(idx.entries))
	for _, entry := range idx.entries {
		slice = append(slice, entry)
	}
	return slice
}

func (idx *Index) IsStaged(path string) bool {
	relPath, err := filepath.Rel(idx.root, path)
	if err != nil {
		return false
	}
	_, ok := idx.entries[relPath]
	return ok
}

type IndexHeader struct {
	signature  [4]byte
	version    uint32
	numEntries uint32
}

type IndexEntry struct {
	Ctimesec  uint32 // Ctime is creation time and Mtim is modification time
	Ctimenano uint32
	Mtimesec  uint32
	Mtimenano uint32
	Dev       uint32
	Ino       uint32
	Mode      uint32
	Size      uint32
	Hash      [20]byte
	Flags     uint16
	Path      []byte
}

func (idx *Index) Write() error {
	indexPath := filepath.Join(idx.root, ".gitgo", "index")
	file, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("Failed to create index file: %v", err)
	}
	defer file.Close()

	hash := sha1.New()
	writer := bufio.NewWriter(io.MultiWriter(file, hash))

	header := IndexHeader{
		signature:  [4]byte{'D', 'I', 'R', 'C'},
		version:    2,
		numEntries: uint32(len(idx.entries)),
	}

	if err := binary.Write(writer, binary.BigEndian, header.signature); err != nil {
		return fmt.Errorf("failed to write signature: %v", err)
	}

	if err := binary.Write(writer, binary.BigEndian, header.version); err != nil {
		return fmt.Errorf("failed to write signature: %v", err)
	}

	if err := binary.Write(writer, binary.BigEndian, header.numEntries); err != nil {
		return fmt.Errorf("failed to write signature: %v", err)
	}

	paths := []string{}
	for path := range idx.entries {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		entry := idx.entries[path]

		indexEntry := IndexEntry{
			Ctimesec:  uint32(entry.Modified.Unix()),
			Ctimenano: uint32(entry.Modified.Nanosecond()),
			Mtimesec:  uint32(entry.Modified.Unix()),
			Mtimenano: uint32(entry.Modified.Nanosecond()),
			Dev:       0,
			Ino:       0,
			Path:      []byte(path),
			Mode:      uint32(entry.Mode),
			Size:      uint32(entry.Size),
			Flags:     uint16(len(entry.Path)),
		}

		hashBytes, err := hex.DecodeString(entry.Hash)
		if err != nil {
			return fmt.Errorf("failed to decode hash %v", err)
		}

		copy(indexEntry.Hash[:], hashBytes)

		if err := binary.Write(writer, binary.BigEndian, indexEntry.Ctimesec); err != nil {
			return fmt.Errorf("failed to write ctime sec: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, indexEntry.Ctimenano); err != nil {
			return fmt.Errorf("failed to write ctime nano: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, indexEntry.Mtimesec); err != nil {
			return fmt.Errorf("failed to write mtime sec: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, indexEntry.Mtimenano); err != nil {
			return fmt.Errorf("failed to write mtime nano: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, indexEntry.Dev); err != nil {
			return fmt.Errorf("failed to write dev: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, indexEntry.Ino); err != nil {
			return fmt.Errorf("failed to write ino: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, indexEntry.Mode); err != nil {
			return fmt.Errorf("failed to write mode: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, indexEntry.Size); err != nil {
			return fmt.Errorf("failed to write size: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, indexEntry.Hash); err != nil {
			return fmt.Errorf("failed to write hash: %v", err)
		}
		if err := binary.Write(writer, binary.BigEndian, indexEntry.Flags); err != nil {
			return fmt.Errorf("failed to write flags: %v", err)
		}
		if _, err := writer.Write(indexEntry.Path); err != nil {
			return fmt.Errorf("failed to write path: %v", err)
		}

		if err := writer.WriteByte(0); err != nil {
			return fmt.Errorf("failed to write path terminator: %v", err)
		}

		padding := 8 - ((62 + len(indexEntry.Path) + 1) % 8)
		if padding < 8 {
			zeros := make([]byte, padding)
			if _, err := writer.Write(zeros); err != nil {
				return fmt.Errorf("failed to write padding: %v", err)
			}
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %v", err)
	}
	if _, err := file.Write(hash.Sum(nil)); err != nil {
		return fmt.Errorf("failed to write checksum: %v", err)
	}

	return nil

}

func (idx *Index) Read() error {
	indexPath := filepath.Join(idx.root, ".gitgo", "index")

	file, err := os.Open(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			idx.Clear()
			return nil
		}
		return fmt.Errorf("failed to open index file %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	header := IndexHeader{}

	if err := binary.Read(reader, binary.BigEndian, &header.signature); err != nil {
		return fmt.Errorf("failed to read signature %v", err)
	}

	if err := binary.Read(reader, binary.BigEndian, &header.version); err != nil {
		return fmt.Errorf("failed to read version %v", err)
	}

	if err := binary.Read(reader, binary.BigEndian, &header.numEntries); err != nil {
		return fmt.Errorf("failed to read num entries %v", err)
	}
	idx.Clear()

	for i := uint32(0); i < header.numEntries; i++ {
		indexEntry := IndexEntry{}

		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Ctimesec); err != nil {
			return fmt.Errorf("failed to read ctimesec %v", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Ctimenano); err != nil {
			return fmt.Errorf("failed to read ctimenano %v", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Mtimesec); err != nil {
			return fmt.Errorf("failed to read mtimesec %v", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Mtimenano); err != nil {
			return fmt.Errorf("failed to read mtimenano %v", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Dev); err != nil {
			return fmt.Errorf("failed to read dev %v", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Ino); err != nil {
			return fmt.Errorf("failed to read ino %v", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Mode); err != nil {
			return fmt.Errorf("failed to read mode %v", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Size); err != nil {
			return fmt.Errorf("failed to read size %v", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Hash); err != nil {
			return fmt.Errorf("failed to read hash %v", err)
		}
		if err := binary.Read(reader, binary.BigEndian, &indexEntry.Flags); err != nil {
			return fmt.Errorf("failed to read flags %v", err)
		}

		path, err := reader.ReadBytes(0)
		if err != nil {
			return fmt.Errorf("failed tyo read path: %v", err)
		}
		path = path[:len(path)-1]

		padding := 8 - ((62 + len(path) + 1) % 8)
		if padding < 8 {
			if _, err := reader.Discard(padding); err != nil {
				return fmt.Errorf("failed to skip padding %v", err)
			}
		}

		entry := &Entry{
			Path:     string(path),
			Hash:     hex.EncodeToString(indexEntry.Hash[:]),
			Mode:     os.FileMode(indexEntry.Mode),
			Size:     int64(indexEntry.Size),
			Modified: time.Unix(int64(indexEntry.Mtimesec), int64(indexEntry.Mtimenano)),
		}
		idx.entries[entry.Path] = entry
	}
	return nil
}

func (idx *Index) Clear() {
	idx.entries = make(map[string]*Entry)
}
