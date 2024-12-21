package blob

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Blob struct {
	hash string

	content []byte
}

func New(content []byte) (*Blob, error) {
	header := fmt.Sprintf("blob %d%c", len(content), 0)
	combined := append([]byte(header), content...)
	sumResult := sha1.Sum(combined)
	hash := hex.EncodeToString(sumResult[:])
	b := Blob{
		hash:    hash,
		content: content,
	}
	return &b, nil
}

func (b *Blob) Hash() string {
	return b.hash
}

func (b *Blob) Content() []byte {
	return b.content
}

// store writes the compressed blob to the objects directory
func (b *Blob) Store(objectsDir string) error {
	directory := b.hash[:2]
	fileName := b.hash[2:]
	directoryPath := filepath.Join(objectsDir, directory)

	err := os.MkdirAll(directoryPath, 0755)
	if err != nil {
		return err
	}
	filePath := filepath.Join(directoryPath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := zlib.NewWriter(file)
	defer writer.Close()
	header := fmt.Sprintf("blob %d%c", len(b.content), 0)
	combined := append([]byte(header), b.content...)

	_, err = writer.Write(combined)
	return err
}

//Read function reads a blob from the objects directory by its hash

func Read(objectsDir, hash string) (*Blob, error) {
	directory := hash[:2]
	fileName := hash[2:]
	fullFilePath := filepath.Join(objectsDir, directory, fileName)
	file, err := os.OpenFile(fullFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader, err := zlib.NewReader(file)

	if err != nil {
		return nil, err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(content, []byte("blob ")) {
		return nil, fmt.Errorf("Invalid blob header: doesn't start with 'blob'")
	}
	nullIndex := bytes.IndexByte(content, 0)
	if nullIndex == -1 {
		return nil, fmt.Errorf("Invalid blob header: no full byte found")
	}
	header := string(content[:nullIndex])
	var length int
	_, err = fmt.Sscanf(header, "blob %d", &length)
	if err != nil {
		return nil, err
	}
	actualContent := content[nullIndex+1:]
	if len(actualContent) != length {
		return nil, fmt.Errorf("Content length mismatch: expected %d, got %d", length, len(actualContent))
	}

	header = fmt.Sprintf("blob %d%c", len(actualContent), 0)
	combined := append([]byte(header), actualContent...)
	sumResult := sha1.Sum(combined)
	hashResult := hex.EncodeToString(sumResult[:])
	if hashResult != hash {
		return nil, fmt.Errorf("Hash mismatch, expected %s, got %s", hash, hashResult)
	}
	return &Blob{
		hash:    hash,
		content: actualContent,
	}, nil
}

func (b *Blob) calculateHash() error { return nil }
