package blob

import "errors"

type Blob struct {
	hash string

	content []byte
}

func New(content []byte) (*Blob, error) {
	return nil, errors.New("Not implemented")
}

func (b *Blob) Hash() string {
	return ""
}

func (b *Blob) Content() []byte

// store writes the compressed blob to the objects directory
func (b *Blob) Store(objectsDir string) error {
	return nil
}

//Read function reads a blob from the objects directory by its hash

func Read(objectsDir, hash string) (*Blob, error) {
	return nil, errors.New("Not implemented")
}

func (b *Blob) calculateHash() error { return nil }
