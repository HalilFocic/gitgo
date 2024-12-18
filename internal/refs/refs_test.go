package refs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReferences(t *testing.T) {
	t.Run("1.1: Read HEAD reference", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}

		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo"), 0755)
		defer os.RemoveAll(testDir)

		headPath := filepath.Join(testDir, ".gitgo", "HEAD")
		err = os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create HEAD file: %v", err)
		}

		ref, err := ReadRef(testDir, "HEAD")
		if err != nil {
			t.Fatalf("Failed to read HEAD: %v", err)
		}

		if ref.Type != RefTypeSymbolic {
			t.Error("Expected HEAD to be symbolic reference")
		}
		if ref.Target != "refs/heads/main" {
			t.Errorf("Wrong target: got %s, want refs/heads/main", ref.Target)
		}
	})

	t.Run("1.2: Read branch reference", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}

		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "refs", "heads"), 0755)
		defer os.RemoveAll(testDir)

		commitHash := "1234567890123456789012345678901234567890"
		branchPath := filepath.Join(testDir, ".gitgo", "refs", "heads", "main")
		err = os.WriteFile(branchPath, []byte(commitHash), 0644)
		if err != nil {
			t.Fatalf("Failed to create branch file: %v", err)
		}

		ref, err := ReadRef(testDir, "refs/heads/main")
		if err != nil {
			t.Fatalf("Failed to read branch: %v", err)
		}

		if ref.Type != RefTypeCommit {
			t.Error("Expected branch to be commit reference")
		}
		if ref.Target != commitHash {
			t.Errorf("Wrong target: got %s, want %s", ref.Target, commitHash)
		}
	})

	t.Run("1.3: Invalid reference", func(t *testing.T) {
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		_, err := ReadRef(testDir, "nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent reference")
		}
	})
	t.Run("2.1: Create and delete branch", func(t *testing.T) {
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "refs", "heads"), 0755)
		defer os.RemoveAll(testDir)

		err := WriteHead(testDir, "refs/heads/main", true)
		if err != nil {
			t.Fatalf("Failed to write HEAD: %v", err)
		}

		commitHash := "1234567890123456789012345678901234567890"

		err = CreateBranch(testDir, "dev", commitHash)
		if err != nil {
			t.Fatalf("Failed to create branch: %v", err)
		}

		ref, err := ReadRef(testDir, "refs/heads/dev")
		if err != nil {
			t.Fatalf("Failed to read created branch: %v", err)
		}
		if ref.Target != commitHash {
			t.Errorf("Branch points to wrong commit: got %s, want %s", ref.Target, commitHash)
		}

		err = CreateBranch(testDir, "dev", commitHash)
		if err == nil {
			t.Error("Expected error when creating duplicate branch")
		}

		err = DeleteBranch(testDir, "dev")
		if err != nil {
			t.Fatalf("Failed to delete branch: %v", err)
		}

		_, err = ReadRef(testDir, "refs/heads/dev")
		if err == nil {
			t.Error("Branch still exists after deletion")
		}
	})

	t.Run("2.2: Cannot delete current branch", func(t *testing.T) {
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "refs", "heads"), 0755)
		defer os.RemoveAll(testDir)

		commitHash := "1234567890123456789012345678901234567890"

		err := CreateBranch(testDir, "main", commitHash)
		if err != nil {
			t.Fatalf("Failed to create main branch: %v", err)
		}
		err = WriteHead(testDir, "refs/heads/main", true)
		if err != nil {
			t.Fatalf("Failed to write HEAD: %v", err)
		}

		err = DeleteBranch(testDir, "main")
		if err == nil {
			t.Error("Should not be able to delete current branch")
		}
	})

	t.Run("2.3: Branch operations with detached HEAD", func(t *testing.T) {
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "refs", "heads"), 0755)
		defer os.RemoveAll(testDir)

		commitHash := "1234567890123456789012345678901234567890"

		err := WriteHead(testDir, commitHash, false)
		if err != nil {
			t.Fatalf("Failed to create detached HEAD: %v", err)
		}

		err = CreateBranch(testDir, "feature", commitHash)
		if err != nil {
			t.Fatalf("Failed to create branch in detached HEAD: %v", err)
		}

		err = DeleteBranch(testDir, "feature")
		if err != nil {
			t.Fatalf("Failed to delete branch in detached HEAD: %v", err)
		}
	})

	t.Run("2.4: Invalid branch names", func(t *testing.T) {
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "refs", "heads"), 0755)
		defer os.RemoveAll(testDir)

		commitHash := "1234567890123456789012345678901234567890"
		invalidNames := []string{
			"",
			"branch/with/slash",
			".",
			"..",
		}

		for _, name := range invalidNames {
			err := CreateBranch(testDir, name, commitHash)
			if err == nil {
				t.Errorf("Expected error for invalid branch name: %q", name)
			}
		}
	})
}
