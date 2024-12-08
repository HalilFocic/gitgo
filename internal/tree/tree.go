package tree

import (
	"fmt"
	"sort"
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

func (tree *Tree) Write(path string) (string, error) {
	//TODO: Implement writing tree content to disk
	return "", nil
}
func (tree *Tree) Read() (*Tree, error) {
    //TODO :Read tree from disk
	return &Tree{
		entries: []TreeEntry{},
	}, nil
}
