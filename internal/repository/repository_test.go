package repository

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/HalilFocic/gitgo/internal/config"
)

func TestInitRepository(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	os.RemoveAll(filepath.Join(cwd, config.GitDirName))

	t.Run("1.1: Initialize new repository", func(t *testing.T) {
		_, err := Init(".")
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		// Check if .gitgo directory exists
		if _, err := os.Stat(config.GitDirName); os.IsNotExist(err) {
			t.Errorf("%s directory was not created", config.GitDirName)
		}

		// Check essential directories
		dirs := []string{
			config.GitDirName,
			filepath.Join(config.GitDirName, "objects"),
			filepath.Join(config.GitDirName, "refs"),
			filepath.Join(config.GitDirName, "refs/heads"),
		}
		for _, dir := range dirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Required directory not created: %s", dir)
			}
		}

		// Repository should not initialize if already exists
		_, err = Init(".")
		if err == nil {
			t.Errorf("Should not initialize repository in existing %s directory", config.GitDirName)
		}
	})
	t.Run("1.2: IsRepository validation", func(t *testing.T) {
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
		os.RemoveAll(filepath.Join(config.GitDirName, "objects"))
		if IsRepository(".") {
			t.Error("IsRepository() returned true for repository with missing objects directory")
		}

		// Cleanup and create new repository for next test
		os.RemoveAll(config.GitDirName)
		Init(".")

		// Remove refs directory
		os.RemoveAll(filepath.Join(config.GitDirName, "refs"))
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
	os.RemoveAll(filepath.Join(cwd, config.GitDirName))
	repo, err := Init(".")
	defer os.RemoveAll(filepath.Join(cwd,config.GitDirName))

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
		os.RemoveAll(filepath.Join(cwd, config.GitDirName))
		_, err := Init(".")
		defer os.RemoveAll(filepath.Join(cwd, config.GitDirName))
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
		os.RemoveAll(filepath.Join(cwd, config.GitDirName, "objects"))
		if IsRepository(".") {
			t.Error("IsRepository() = true, want false for repository with missing objects directory")
		}
	})
}
