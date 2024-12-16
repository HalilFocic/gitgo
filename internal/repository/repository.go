package repository

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Repository struct {
	Path     string
	GitgoDir string
}

func Init(path string) (*Repository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	gitGoPath := filepath.Join(absPath, ".gitgo")

	file, err := os.Stat(gitGoPath)
	if err == nil && file != nil {
		return nil, errors.New("repository already exists in this repository")
	}

	objectsPath := filepath.Join(gitGoPath, "objects")
	refsPath := filepath.Join(gitGoPath, "refs")
	headsPath := filepath.Join(gitGoPath, "refs/heads")

	indexPath := filepath.Join(gitGoPath, "index")
	err = os.MkdirAll(gitGoPath, 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(objectsPath, 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(refsPath, 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(headsPath, 0755)
	if err != nil {
		return nil, err
	}
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return nil, err
	}

	indexFile.Close()
	headPath := filepath.Join(gitGoPath, "HEAD")
	err = os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0755)
	if err != nil {
		os.RemoveAll(gitGoPath)
		return nil, fmt.Errorf("failed to create HEAD file: %v", err)
	}
	return &Repository{
		Path:     absPath,
		GitgoDir: gitGoPath,
	}, nil
}

func IsRepository(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	gitGoPath := filepath.Join(absPath, ".gitgo")
	dirs := []string{
		".",
		"./objects",
		"./refs",
		"./refs/heads",
	}
	for _, d := range dirs {
		p := filepath.Join(gitGoPath, d)
		file, err := os.Stat(p)
		if file == nil || err != nil {
			return false
		}
	}
	return true
}

func (r *Repository) ObjectPath() string {
	return filepath.Join(r.GitgoDir, "objects")
}

func (r *Repository) RefsPath() string {
	return filepath.Join(r.GitgoDir, "refs")
}
