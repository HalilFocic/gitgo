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

func WriteHead(rootPath, target string, isSymbol bool) error {
	return UpdateRef(rootPath, HeadFile, target, isSymbol)
}

func CreateBranch(rootPath, name, commitHash string) error {
	if strings.Contains(name, "/") {
		return fmt.Errorf("branch cannot contain slashes")
	}
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	branchRef := filepath.Join("refs", "heads", name)
	if _, err := ReadRef(rootPath, branchRef); err == nil {
		return fmt.Errorf("branch %s already exists", name)
	}
	return UpdateRef(rootPath, branchRef, commitHash, false)
}

func DeleteBranch(rootPath, name string) error {
	branchRef := filepath.Join("refs", "heads", name)
	_, err := ReadRef(rootPath, branchRef)
	if err != nil {
		return fmt.Errorf("branch %s does not exist", name)
	}
	head, err := ReadHead(rootPath)
	if err != nil {
		return fmt.Errorf("failed to read head: %v", err)
	}

	if head.Type == RefTypeSymbolic && head.Target == branchRef {
		return fmt.Errorf("cannot delete current branch %s", name)
	}
	branchPath := filepath.Join(rootPath, ".gitgo", branchRef)
	if err := os.Remove(branchPath); err != nil {
		return fmt.Errorf("failed to delete branch %s: %v", name, err)
	}
	return nil
}

func ListBranches(rootPath string) ([]string, error) {
	headsDir := filepath.Join(rootPath, ".gitgo", "refs", "heads")

	files, err := os.ReadDir(headsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read refs directory: %v", err)
	}

	var branches []string
	for _, file := range files {
		if !file.IsDir() {
			branches = append(branches, file.Name())
		}
	}

	return branches, nil
}
