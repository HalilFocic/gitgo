package tree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTreeOperations(t *testing.T) {
	t.Run("1.1: Create new tree and add entries", func(t *testing.T) {
		tree := New()
		fileHash := "1234567890123456789012345678901234567890"
		dirHash := "abcdefabcdefabcdefabcdefabcdefabcdefabcd"
		err := tree.AddEntry("main.go", fileHash, RegularFileMode)
		if err != nil {
			t.Fatalf("Failed to add file entry: %v", err)
		}

		err = tree.AddEntry("lib", dirHash, DirectoryMode)
		if err != nil {
			t.Fatalf("Failed to add directory entry: %v", err)
		}

		entries := tree.Entries()
		if len(entries) != 2 {
			t.Fatalf("Expected 2 entries, got %d", len(entries))
		}

		// 1 was used instead of zero because m is after l alphabetically
		// since we sorted the entries, this makes impact
		if entries[1].Name != "main.go" {
			t.Errorf("Expected name 'main.go', got %s", entries[0].Name)
		}
		if entries[1].Hash != fileHash {
			t.Errorf("Expected hash 'abc123', got %s", entries[0].Hash)
		}
		if entries[1].Mode != RegularFileMode {
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

func TestTreeEdgeCases(t *testing.T) {
	t.Run("2.1: Special characters in filename", func(t *testing.T) {
		tree := New()
		hash := "1234567890123456789012345678901234567890"
        // This test tries to add files with special characters
		specialNames := []string{
			"hello world.txt",
			"!@#$%.txt",
			"über.txt",
			"file名前.txt",
		}

		for _, name := range specialNames {
			err := tree.AddEntry(name, hash, RegularFileMode)
			if err != nil {
				t.Errorf("Failed to add file with special name '%s': %v", name, err)
			}
		}

		// Test write/read with special characters
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "objects"), 0755)
		defer os.RemoveAll(testDir)

		treeHash, err := tree.Write(filepath.Join(testDir, ".gitgo", "objects"))
		if err != nil {
			t.Fatalf("Failed to write tree: %v", err)
		}

		readTree, err := Read(filepath.Join(testDir, ".gitgo", "objects"), treeHash)
		if err != nil {
			t.Fatalf("Failed to read tree: %v", err)
		}

		for _, name := range specialNames {
			found := false
			for _, entry := range readTree.Entries() {
				if entry.Name == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Failed to find entry '%s' after read", name)
			}
		}
	})

	t.Run("2.2: Empty tree", func(t *testing.T) {
		tree := New()

		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "objects"), 0755)
		defer os.RemoveAll(testDir)

		hash, err := tree.Write(filepath.Join(testDir, ".gitgo", "objects"))
		if err != nil {
			t.Fatalf("Failed to write empty tree: %v", err)
		}

		readTree, err := Read(filepath.Join(testDir, ".gitgo", "objects"), hash)
		if err != nil {
			t.Fatalf("Failed to read empty tree: %v", err)
		}

		if len(readTree.Entries()) != 0 {
			t.Errorf("Expected empty tree, got %d entries", len(readTree.Entries()))
		}
	})

	t.Run("2.3: Invalid UTF-8", func(t *testing.T) {
		tree := New()
		hash := "1234567890123456789012345678901234567890"

		// Create invalid UTF-8 string
		invalidName := string([]byte{0xFF, 0xFE, 0xFD})

		err := tree.AddEntry(invalidName, hash, RegularFileMode)
		if err == nil {
			t.Error("Expected error for invalid UTF-8 filename")
		}
	})
}
