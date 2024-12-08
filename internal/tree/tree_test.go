package tree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTreeOperations(t *testing.T) {
	t.Run("1.1: Create new tree and add entries", func(t *testing.T) {
		tree := New()

		err := tree.AddEntry("main.go", "abc123", RegularFileMode)
		if err != nil {
			t.Fatalf("Failed to add file entry: %v", err)
		}

		err = tree.AddEntry("lib", "def456", DirectoryMode)
		if err != nil {
			t.Fatalf("Failed to add directory entry: %v", err)
		}

		entries := tree.Entries()
		if len(entries) != 2 {
			t.Fatalf("Expected 2 entries, got %d", len(entries))
		}

		if entries[0].Name != "main.go" {
			t.Errorf("Expected name 'main.go', got %s", entries[0].Name)
		}
		if entries[0].Hash != "abc123" {
			t.Errorf("Expected hash 'abc123', got %s", entries[0].Hash)
		}
		if entries[0].Mode != RegularFileMode {
			t.Errorf("Expected mode %o, got %o", RegularFileMode, entries[0].Mode)
		}
	})

	t.Run("1.2: Invalid entry handling", func(t *testing.T) {
		tree := New()

		err := tree.AddEntry("", "abc123", RegularFileMode)
		if err == nil {
			t.Errorf("Expected error for empty name, got none")
		}

		err = tree.AddEntry("test.txt", "", RegularFileMode)
		if err == nil {
			t.Errorf("Expected error for empty hash, got none")
		}

		err = tree.AddEntry("test.txt", "abc123", 0777)
		if err == nil {
			t.Error("Expected error for invalid mode, got none")
		}
	})

	t.Run("1.3: Storage and retrieval", func(t *testing.T) {
		tree := New()
		tree.AddEntry("file1.txt", "abc123", RegularFileMode)
		tree.AddEntry("dir", "def456", DirectoryMode)

		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}

		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "objects"), 0755)
		defer os.RemoveAll(testDir)

		hash, err := tree.Write(filepath.Join(testDir, ".gitgo", "objects"))
		if err != nil {
			t.Fatalf("Failed to write tree: %v", err)
		}

		readTree, err := Read(filepath.Join(testDir, ".gitgo", "objects"), hash)
		if err != nil {
			t.Fatalf("Failed to read tree: %v", err)
		}

		originalEntries := tree.Entries()
		readEntries := readTree.Entries()

		if len(originalEntries) != len(readEntries) {
			t.Errorf("Entry count mismatch: got %d, want %d",
				len(readEntries), len(originalEntries))
		}

		for i, original := range originalEntries {
			read := readEntries[i]
			if original.Name != read.Name {
				t.Errorf("Name mismatch: got %s, want %s",
					read.Name, original.Name)
			}
			if original.Hash != read.Hash {
				t.Errorf("Hash mismatch: got %s, want %s",
					read.Hash, original.Hash)
			}
			if original.Mode != read.Mode {
				t.Errorf("Mode mismatch: got %o, want %o",
					read.Mode, original.Mode)
			}
		}
	})

}
