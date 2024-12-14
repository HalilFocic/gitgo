package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HalilFocic/gitgo/internal/repository"
	"github.com/HalilFocic/gitgo/internal/staging"
)

func TestCommitCommand(t *testing.T) {
	t.Run("1.1: Basic commit creation", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}

		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		_, err = repository.Init(testDir)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		files := map[string]string{
			"main.go":     "main content",
			"lib/util.go": "util content",
			"lib/test.go": "test content",
		}

		if err := os.Chdir(testDir); err != nil {
			t.Fatalf("Failed to change to test directory: %v", err)
		}
		defer os.Chdir(cwd)

		for path, content := range files {
			fullPath := filepath.Join(".", path)
			err := os.MkdirAll(filepath.Dir(fullPath), 0755)
			if err != nil {
				t.Fatalf("Failed to create directory for %s: %v", path, err)
			}
			err = os.WriteFile(fullPath, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", path, err)
			}
		}

		idx, err := staging.New(".")
		if err != nil {
			t.Fatalf("Failed to create staging area: %v", err)
		}

		for path := range files {
			err = idx.Add(path)
			if err != nil {
				t.Fatalf("Failed to stage file %s: %v", path, err)
			}
		}

		cmd := NewCommitCommand(
			".",
			"Initial commit",
			"Test User <test@example.com>",
		)

		err = cmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute commit: %v", err)
		}

		objectsDir := filepath.Join(testDir, ".gitgo", "objects")

		var objectCount int
		err = filepath.Walk(objectsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				objectCount++
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to count objects: %v", err)
		}

		if objectCount < 6 {
			t.Errorf("Expected at least 6 objects, got %d", objectCount)
		}
	})
	t.Run("1.2: Empty staging area", func(t *testing.T) {
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		_, err := repository.Init(testDir)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		cmd := NewCommitCommand(
			testDir,
			"Empty commit",
			"Test User <test@example.com>",
		)

		err = cmd.Execute()
		if err == nil {
			t.Error("Expected error for empty staging area")
		}
	})

	t.Run("1.3: Directory structure preservation", func(t *testing.T) {
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		_, err := repository.Init(testDir)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		files := map[string]string{
			"src/main.go":          "main content",
			"src/lib/util/util.go": "util content",
			"src/lib/test/test.go": "test content",
		}
		if err := os.Chdir(testDir); err != nil {
			t.Fatalf("Failed to change to test directory: %v", err)
		}
		defer os.Chdir(cwd)
		for path, content := range files {
			fullPath := filepath.Join(".", path)
			err := os.MkdirAll(filepath.Dir(fullPath), 0755)
			if err != nil {
				t.Fatalf("Failed to create directory for %s: %v", path, err)
			}
			err = os.WriteFile(fullPath, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", path, err)
			}
		}

		idx, err := staging.New(".")
		if err != nil {
			t.Fatalf("Failed to create staging area: %v", err)
		}

		for path := range files {
			err = idx.Add(path)
			if err != nil {
				t.Fatalf("Failed to stage file %s: %v", path, err)
			}
		}

		cmd := NewCommitCommand(
			testDir,
			"Nested directories",
			"Test User <test@example.com>",
		)

		err = cmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute commit: %v", err)
		}

		objectsDir := filepath.Join(testDir, ".gitgo", "objects")
		var objectCount int
		err = filepath.Walk(objectsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				objectCount++
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to count objects: %v", err)
		}

		if objectCount < 8 {
			t.Errorf("Expected at least 8 objects, got %d", objectCount)
		}
	})

}
