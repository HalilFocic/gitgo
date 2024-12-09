package tree

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	RegularFileMode = 0100644 // Regular file
	ExecutableMode  = 0100755 // Executable file
	DirectoryMode   = 0040000 // Directory
)

type TreeEntry struct {
	Name string
	Hash string
	Mode int
}

// Implemented purely to use sort.Sort
type TreeEntries []TreeEntry

func (te TreeEntries) Len() int {
	return len(te)
}
func (te TreeEntries) Swap(i, j int)      { te[i], te[j] = te[j], te[i] }
func (te TreeEntries) Less(i, j int) bool { return te[i].Name < te[j].Name }

type Tree struct {
	entries []TreeEntry
}

func New() *Tree {
	return &Tree{
		entries: []TreeEntry{},
	}
}

func (tree *Tree) AddEntry(name, hash string, filemode int) error {
	if len(name) == 0 {
		return fmt.Errorf("entry name cannot be empty")
	}
	if len(hash) == 0 {
		return fmt.Errorf("entry hash cannot be empty")
	}
	if len(hash) != 40 {
		return fmt.Errorf("invalid hash length: expected 40 characters, got %d", len(hash))
	}
	if filemode != RegularFileMode && filemode != ExecutableMode && filemode != DirectoryMode {
		return fmt.Errorf("Invalid file mode.")
	}
	if strings.Contains(name, "/") {
		return fmt.Errorf("Entry name cannot contain '/', this should be handled by seperate tree")
	}
	for _, tEntry := range tree.entries {
		if tEntry.Name == name {
			return fmt.Errorf("Name %s already exists inside this tree", tEntry.Name)
		}
	}
	tree.entries = append(tree.entries, TreeEntry{
		Name: name,
		Hash: hash,
		Mode: filemode,
	})
	sort.Sort(TreeEntries(tree.entries))
	return nil
}

func (tree *Tree) Entries() []TreeEntry {
	return tree.entries
}

func (tree *Tree) Write(objectsPath string) (string, error) {
	var buffer bytes.Buffer

	contentSize := 0
	for _, entry := range tree.entries {
		contentSize += len(strconv.FormatInt(int64(entry.Mode), 8)) + 1 + len(entry.Name) + 1 + 20
	}

	fmt.Fprintf(&buffer, "tree %d\x00", contentSize)

	for _, entry := range tree.entries {
		fmt.Fprintf(&buffer, "%o ", entry.Mode)

		buffer.WriteString(entry.Name)
		buffer.WriteByte(0)

		hashBytes, err := hex.DecodeString(entry.Hash)
		if err != nil {
			return "", fmt.Errorf("invalid hash %s: %v", entry.Hash, err)
		}
		buffer.Write(hashBytes)
	}

	var compressed bytes.Buffer
	zWriter := zlib.NewWriter(&compressed)
	if _, err := zWriter.Write(buffer.Bytes()); err != nil {
		return "", fmt.Errorf("compression failed: %v", err)
	}
	zWriter.Close()

	hash := sha1.Sum(compressed.Bytes())
	hashString := hex.EncodeToString(hash[:])

	hashPath := filepath.Join(objectsPath, hashString[:2], hashString[2:])
	if err := os.MkdirAll(filepath.Dir(hashPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create object directory: %v", err)
	}
	if err := os.WriteFile(hashPath, compressed.Bytes(), 0755); err != nil {
		return "", fmt.Errorf("failed to write object file: %v", err)
	}

	return hashString, nil
}
func Read(objectsPath, hash string) (*Tree, error) {
	hashPath := filepath.Join(objectsPath, hash[:2], hash[2:])
	compressed, err := os.ReadFile(hashPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read object file: %v", err)
	}

	zReader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %v", err)
	}
	defer zReader.Close()

	var buffer bytes.Buffer

	if _, err := io.Copy(&buffer, zReader); err != nil {
		return nil, fmt.Errorf("failed to decompress data: %v", err)
	}

	data := buffer.Bytes()
	header := bytes.SplitN(data, []byte{0}, 2)
	if len(header) != 2 {
		return nil, fmt.Errorf("invalid object format")
	}

	headerParts := bytes.Fields(header[0])
	if len(headerParts) != 2 || string(headerParts[0]) != "tree" {
		return nil, fmt.Errorf("not a tree object")
	}
	tree := New()
	content := header[1]
	for len(content) > 0 {

		spaceIndex := bytes.IndexByte(content, ' ')
		if spaceIndex == -1 {
			return nil, fmt.Errorf("invalid entry format")
		}

		nullIdx := bytes.IndexByte(content[spaceIndex+1:], 0)

		nullIdx += spaceIndex + 1

		mode, err := strconv.ParseInt(string(content[:spaceIndex]), 8, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid mode: %v", err)
		}

		name := string(content[spaceIndex+1 : nullIdx])

		if len(content[nullIdx+1:]) < 20 {
			return nil, fmt.Errorf("truncated entry")
		}

		hash := hex.EncodeToString(content[nullIdx+1 : nullIdx+21])

		if err := tree.AddEntry(name, hash, int(mode)); err != nil {
			return nil, fmt.Errorf("failed to add entry: %v", err)
		}
		content = content[nullIdx+21:]
	}
	return tree, nil
}
