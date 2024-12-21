package staging

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/HalilFocic/gitgo/internal/repository"
)

func TestStaingArea(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory %v", err)
	}
	testDir := filepath.Join(cwd, "testdata")
	os.RemoveAll(testDir)

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to created test directory %v", err)
	}
	defer os.RemoveAll(testDir)

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory %v", err)
	}
	defer os.Chdir(cwd)

	_, err = repository.Init(".")
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}
	t.Run("1.1: Add file to staging", func(t *testing.T) {
		content := []byte("test content")
		if err := os.WriteFile("test.txt", content, 0644); err != nil {
			t.Fatalf("Failed to create test.txt file: %v", err)
		}

		index, err := New(".")
		if err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}
		defer index.Clear()
		err = index.Add("test.txt")
		if err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}
		entries := index.Entries()
		if len(entries) != 1 {
			t.Fatalf("Expected 1 entry, got %d", len(entries))
		}
		if entries[0].Path != "test.txt" {
			t.Fatalf("Wrong path, expected test.txt, got %s", entries[0].Path)
		}
	})

	t.Run("1.2: Multiple file staging", func(t *testing.T) {
		os.Remove(filepath.Join(".gitgo", "index"))
		index, err := New(".")
		if err != nil {
			t.Fatalf("failed to create index: %v", err)
		}
		defer index.Clear()
		files := []string{"a.txt", "b.txt", "dir/c.txt"}

		for _, f := range files {
			dir := filepath.Dir(f)
			if dir != "." {
				if err = os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory for %s: %v", f, err)
				}
			}

			if err := os.WriteFile(f, []byte("content"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", f, err)
			}
		}

		for _, f := range files {
			err := index.Add(f)
			if err != nil {
				t.Fatalf("Failed to add %s: %v", f, err)
			}
		}

		if len(index.Entries()) != len(files) {
			t.Fatalf("Expected %d entries, got %d", len(files), len(index.Entries()))
		}
	})

	t.Run("1.3 Update staged file", func(t *testing.T) {
		os.Remove(filepath.Join(".gitgo", "index"))
		index, err := New(".")
		if err != nil {
			t.Fatalf("Failed to create index %v", err)
		}
		defer index.Clear()

		if err = os.WriteFile("update.txt", []byte("initial"), 0644); err != nil {
			t.Fatalf("Failed to create update.txt file: %v", err)
		}
		err = index.Add("update.txt")
		if err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}
		initialHash := index.Entries()[0].Hash

		if err := os.WriteFile("update.txt", []byte("updated"), 0644); err != nil {
			t.Fatalf("Failed to update the update.txt file:%v", err)
		}
		err = index.Add("update.txt")
		if err != nil {
			t.Fatalf("Failed to add updated file:%v", err)
		}

		updatedHash := index.Entries()[0].Hash

		if initialHash == updatedHash {
			t.Fatalf("Hash should change when file is updated")
		}
	})

	t.Run("1.4 Remove file from staging", func(t *testing.T) {
		os.Remove(filepath.Join(".gitgo", "index"))
		index, err := New(".")

		if err != nil {
			t.Fatalf("Failed to create index %v", err)
		}
		defer index.Clear()

		if err := os.WriteFile("remove.txt", []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = index.Add("remove.txt")
		if err != nil {
			t.Fatalf("Failed to add remove.txt to index: %v", err)
		}
		if !index.IsStaged("remove.txt") {
			t.Fatalf("remove.txt should have been staged")
		}
		err = index.Remove("remove.txt")
		if err != nil {
			t.Fatalf("Failed to remove file from staing: %v", err)
		}
		if index.IsStaged("remove.txt") {
			t.Fatalf("File should not be staged after removal")
		}
	})
	t.Run("1.5: Write and read index", func(t *testing.T) {
		os.Remove(filepath.Join(".gitgo", "index"))
		index, err := New(".")
		if err != nil {
			t.Fatalf("Failed to created index: %v", err)
		}

		files := []string{"write.txt", "read.txt"}

		for _, f := range files {
			if err := os.WriteFile(f, []byte("content"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", f, err)
			}
			if err := index.Add(f); err != nil {
				t.Fatalf("Failed to add file %s: %v", f, err)
			} else {
				fmt.Printf("added to index : %v\n", f)
			}
		}
		if err := index.Write(); err != nil {
			t.Fatalf("Failed to write to index :%v", err)
		}
		originalEntries := index.Entries()

		newIndex, err := New(".")
		if err != nil {
			t.Fatalf("Failed to create second index:%v", err)
		}

		newEntries := newIndex.Entries()
		if len(originalEntries) != len(newEntries) {
			t.Fatalf("Entries count missmatch: expected %d, go %d", len(originalEntries), len(newEntries))
		}
	})

	t.Run("1.6: Clear index", func(t *testing.T) {
		os.Remove(filepath.Join(".gitgo", "index"))
		index, err := New(".")
		if err != nil {
			t.Fatalf("Failed to create index: %v", err)
		}

		// Add a file
		if err := os.WriteFile("clear.txt", []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = index.Add("clear.txt")
		if err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		if len(index.Entries()) == 0 {
			t.Error("Index should have entries before clear")
		}
		index.Clear()

		if len(index.Entries()) != 0 {
			t.Error("Index should have no entries after clear")
		}
	})
}
