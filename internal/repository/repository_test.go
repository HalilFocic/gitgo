package repository

import (
	"os"
	"path/filepath"
	//"path/filepath"
	"testing"
)

func TestInitRepository(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directgory: %v", err)
	}
	os.RemoveAll(filepath.Join(cwd, ".gitgo"))

	t.Run("1.1: Initialize new repository", func(t *testing.T) {
		_, err := Init(".")
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		// Check if .gitgo directory exists
		if _, err := os.Stat(".gitgo"); os.IsNotExist(err) {
			t.Error("Gitgo directory was not created")
		}

		// Check essential directories
		dirs := []string{
			".gitgo/objects",
			".gitgo/refs",
			".gitgo/refs/heads",
		}

		for _, dir := range dirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Required directory not created: %s", dir)
			}
		}

		// Repository should not initialize if already exists
		_, err = Init(".")
		if err == nil {
			t.Error("Should not initialize repository in existing .gitgo directory")
		}
	})
	t.Run("2.2: IsRepository validation", func(t *testing.T) {
		// Should return true for valid repository
		if !IsRepository(".") {
			t.Error("IsRepository() returned false for valid repository")
		}

		// Should return false for non-existent path
		if IsRepository("./non-existent-path") {
			t.Error("IsRepository() returned true for non-existent path")
		}

		// Should return false if missing critical directories
		// Remove objects directory
		os.RemoveAll(filepath.Join(".gitgo", "objects"))
		if IsRepository(".") {
			t.Error("IsRepository() returned true for repository with missing objects directory")
		}

		// Cleanup and create new repository for next test
		os.RemoveAll(".gitgo")
		Init(".")

		// Remove refs directory
		os.RemoveAll(filepath.Join(".gitgo", "refs"))
		if IsRepository(".") {
			t.Error("IsRepository() returned true for repository with missing refs directory")
		}
	})
}
func TestRepositoryPaths(t *testing.T) {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Clean up and create new repository
	os.RemoveAll(filepath.Join(cwd, ".gitgo"))
	repo, err := Init(".")
	if err != nil {
		t.Fatalf("Failed to initialize test repository: %v", err)
	}

	t.Run("2.1: Test ObjectPath", func(t *testing.T) {
		expected := filepath.Join(repo.GitgoDir, "objects")
		if repo.ObjectPath() != expected {
			t.Errorf("ObjectPath() = %v, want %v", repo.ObjectPath(), expected)
		}
	})

	t.Run("2.2: Test RefsPath", func(t *testing.T) {
		expected := filepath.Join(repo.GitgoDir, "refs")
		if repo.RefsPath() != expected {
			t.Errorf("RefsPath() = %v, want %v", repo.RefsPath(), expected)
		}
	})
}

func TestIsRepository(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	t.Run("3.1: Valid repository detection", func(t *testing.T) {
		// Clean up and create new repository
		os.RemoveAll(filepath.Join(cwd, ".gitgo"))
		_, err := Init(".")
		if err != nil {
			t.Fatalf("Failed to initialize test repository: %v", err)
		}

		if !IsRepository(".") {
			t.Error("IsRepository() = false, want true for valid repository")
		}
	})

	t.Run("3.2: Invalid repository detection", func(t *testing.T) {
		// Test non-existent directory
		if IsRepository("./non-existent") {
			t.Error("IsRepository() = true, want false for non-existent directory")
		}

		// Test incomplete repository structure
		os.RemoveAll(filepath.Join(cwd, ".gitgo", "objects"))
		if IsRepository(".") {
			t.Error("IsRepository() = true, want false for repository with missing objects directory")
		}
	})
}
