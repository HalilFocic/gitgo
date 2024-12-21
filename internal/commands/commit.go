package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/HalilFocic/gitgo/internal/commit"
	"github.com/HalilFocic/gitgo/internal/staging"
	"github.com/HalilFocic/gitgo/internal/tree"
)

type CommitCommand struct {
	rootPath string
	message  string
	author   string
}

func NewCommitCommand(rootPath, message, author string) *CommitCommand {
	return &CommitCommand{
		rootPath: rootPath,
		message:  message,
		author:   author,
	}
}

func (c *CommitCommand) Execute() error {
	index, err := staging.New(c.rootPath)
	if err != nil {
		return fmt.Errorf("failed to read staging area: %v", err)
	}

	entries := index.Entries()
	if len(entries) == 0 {
		return fmt.Errorf("nothing to commit, staging area is empty")
	}

	root := c.groupEntriesByDirectory(entries)

	objectsPath := filepath.Join(c.rootPath, ".gitgo", "objects")
	treeHash, err := c.createTreeFromNode(root, objectsPath)
	if err != nil {
		return fmt.Errorf("failed to create tree: %v", err)
	}

	headContent, err := os.ReadFile(filepath.Join(c.rootPath, ".gitgo", "HEAD"))
	if err != nil {
		return err
	}

	headRef := strings.TrimSpace(string(headContent))
	if !strings.HasPrefix(headRef, "ref: refs/heads/") {
		return fmt.Errorf("invalid HEAD format")
	}

	branchName := strings.TrimPrefix(headRef, "ref: refs/heads/")
	branchPath := filepath.Join(c.rootPath, ".gitgo", "refs", "heads", branchName)
	parentHash := ""
	if hash, err := os.ReadFile(branchPath); err == nil {
		parentHash = strings.TrimSpace(string(hash))
	}

	newCommit, err := commit.New(treeHash, parentHash, c.author, c.message)
	if err != nil {
		return fmt.Errorf("failed to create commit: %v", err)
	}

	commitHash, err := newCommit.Write(objectsPath)
	if err != nil {
		return fmt.Errorf("failed to write commit :%v", err)
	}

	if err := os.WriteFile(branchPath, []byte(commitHash), 0644); err != nil {
		return fmt.Errorf("failed to update branch reference: %v", err)
	}
	index.Clear()
	return nil
}

type pathNode struct {
	files    map[string]staging.Entry
	children map[string]*pathNode
}

func NewPathNode() *pathNode {
	return &pathNode{
		files:    make(map[string]staging.Entry),
		children: make(map[string]*pathNode),
	}
}

func (c *CommitCommand) groupEntriesByDirectory(entries []*staging.Entry) *pathNode {
	root := NewPathNode()

	for _, entry := range entries {
		parts := strings.Split(entry.Path, "/")
		current := root
		for i := 0; i < len(parts)-1; i++ {
			dirName := parts[i]
			if _, exists := current.children[dirName]; !exists {
				current.children[dirName] = NewPathNode()
			}
			current = current.children[dirName]
		}

		filename := parts[len(parts)-1]
		current.files[filename] = *entry
	}

	return root
}

func getFileMode(mode fs.FileMode) int {
	if mode&fs.ModeDir != 0 {
		return tree.DirectoryMode
	}
	if mode&0111 != 0 {
		return tree.ExecutableMode
	}
	return tree.RegularFileMode
}

func (c *CommitCommand) createTreeFromNode(node *pathNode, objectsPath string) (string, error) {
	t := tree.New()

	for dirName, childNode := range node.children {
		childHash, err := c.createTreeFromNode(childNode, objectsPath)
		if err != nil {
			return "", fmt.Errorf("failed to create tree for %s: %v", dirName, err)
		}

		err = t.AddEntry(dirName, childHash, tree.DirectoryMode)
	}
	for fileName, entry := range node.files {
		err := t.AddEntry(fileName, entry.Hash, getFileMode(entry.Mode))
		if err != nil {
			return "", fmt.Errorf("failed to add entry %s: %v", fileName, err)
		}
	}

	hash, err := t.Write(objectsPath)
	if err != nil {
		return "", fmt.Errorf("failed to write tree: %v", err)
	}
	return hash, nil
}
