package refs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// "fmt"
// "os"
// "path/filepath"

const (
	RefTypeCommit = iota
	RefTypeSymbolic
)

const (
	HeadFile = "HEAD"
	RefsDir  = "refs"
	HeadsDir = "refs/heads"
)

type Reference struct {
	Name     string
	Type     int
	Target   string
	rootPath string
}

func ReadRef(rootPath, name string) (Reference, error) {
	refPath := filepath.Join(rootPath, ".gitgo", name)

	content, err := os.ReadFile(refPath)
	if err != nil {
		return Reference{}, fmt.Errorf("failed to read reference %s: %v", name, err)
	}
	ref := Reference{
		Name:     name,
		rootPath: rootPath,
	}

	text := strings.TrimSpace(string(content))
	if strings.HasPrefix(text, "ref: ") {
		ref.Type = RefTypeSymbolic
		ref.Target = strings.TrimPrefix(text, "ref: ")
	} else {
		ref.Type = RefTypeCommit
		ref.Target = text
	}
	return ref, nil
}

func ReadHead(rootPath string) (Reference, error) {
	return ReadRef(rootPath, HeadFile)
}

func UpdateRef(rootPath, name, target string, isSymbolic bool) error {
	fullPath := filepath.Join(rootPath, ".gitgo", name)

	var content string
	if isSymbolic {
		content = "ref: " + target + "\n"
	} else {
		content = target + "\n"
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %v", name, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write reference %s: %v", name, err)
	}
	return nil
}
