package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/HalilFocic/gitgo/internal/blob"
	"github.com/HalilFocic/gitgo/internal/commit"
	"github.com/HalilFocic/gitgo/internal/refs"
	"github.com/HalilFocic/gitgo/internal/tree"
)

type CheckoutCommand struct {
	rootPath string
	target   string
}

func NewCheckoutCommand(rootPath, target string) *CheckoutCommand {
	return &CheckoutCommand{
		rootPath: rootPath,
		target:   target,
	}
}

func (c *CheckoutCommand) Execute() error {
	branchRef := filepath.Join("refs", "heads", c.target)
	ref, err := refs.ReadRef(c.rootPath, branchRef)

	var commitHash string
	if err == nil {
		commitHash = ref.Target
		if err := refs.WriteHead(c.rootPath, branchRef, true); err != nil {
			return fmt.Errorf("failed to update HEAD: %v", err)
		}
	} else {
		if len(c.target) != 40 {
			return fmt.Errorf("invalid reference: %s", c.target)
		}
		commitHash = c.target
		if err := refs.WriteHead(c.rootPath, commitHash, false); err != nil {
			return fmt.Errorf("failed to update HEAD: %v", err)
		}
	}

	com, err := commit.Read(filepath.Join(c.rootPath, ".gitgo", "objects"), commitHash)
	if err != nil {
		return fmt.Errorf("failed to read commit: %v", err)
	}

	rootTree, err := tree.Read(filepath.Join(c.rootPath, ".gitgo", "objects"), com.TreeHash)
	if err != nil {
		return fmt.Errorf("failed to read tree: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(c.rootPath, "*"))
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}
	for _, f := range files {
		if filepath.Base(f) != ".gitgo" {
			os.RemoveAll(f)
		}
	}

	if err := c.writeTree(rootTree, c.rootPath); err != nil {
		return fmt.Errorf("failed to write files: %v", err)
	}

	return nil
}

func (c *CheckoutCommand) writeTree(t *tree.Tree, path string) error {
	for _, entry := range t.Entries() {
		fullPath := filepath.Join(path, entry.Name)

		if entry.Mode == tree.DirectoryMode {
			os.MkdirAll(fullPath, 0755)
			subTree, err := tree.Read(filepath.Join(c.rootPath, ".gitgo", "objects"), entry.Hash)
			if err != nil {
				return fmt.Errorf("failed to read subtree: %v", err)
			}
			if err := c.writeTree(subTree, fullPath); err != nil {
				return err
			}
		} else {
			b, err := blob.Read(filepath.Join(c.rootPath, ".gitgo", "objects"), entry.Hash)
			if err != nil {
				return fmt.Errorf("failed to read blob: %v", err)
			}
			if err := os.WriteFile(fullPath, b.Content(), fs.FileMode(entry.Mode)); err != nil {
				return fmt.Errorf("failed to write file: %v", err)
			}
		}
	}
	return nil
}
