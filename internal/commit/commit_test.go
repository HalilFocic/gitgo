package commit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommitCreation(t *testing.T) {
	t.Run("1.1: valid commit creation", func(t *testing.T) {
		treeHash := "1234567890123456789012345678901234567890"
		parentHash := "abcdef1234567890abcdef1234567890abcdef12"
		author := "John Doe <john@example.com>"
		message := "Initial commit"

		commit, err := New(treeHash, parentHash, author, message)
		if err != nil {
			t.Fatalf("Failed to create valid commit: %v", err)
		}

		if commit.TreeHash != treeHash {
			t.Errorf("TreeHash = %s; want %s", commit.TreeHash, treeHash)
		}
		if commit.ParentHash != parentHash {
			t.Errorf("ParentHash = %s; want %s", commit.ParentHash, parentHash)
		}
		if commit.Author != author {
			t.Errorf("Author = %s; want %s", commit.Author, author)
		}
		if commit.Message != message {
			t.Errorf("Message = %s; want %s", commit.Message, message)
		}
		if commit.AuthorDate.IsZero() {
			t.Error("AuthorDate should not be zero")
		}
	})
	t.Run("1.2: Test first commit", func(t *testing.T) {
		treeHash := "1234567890123456789012345678901234567890"
		author := "John Doe <john@example.com>"
		message := "First commit"

		commit, err := New(treeHash, "", author, message)
		if err != nil {
			t.Fatalf("Failed to create first commit: %v", err)
		}

		if commit.ParentHash != "" {
			t.Errorf("First commit should have empty ParentHash, got %s", commit.ParentHash)
		}
	})
	t.Run("1.3: Invalid tree hash", func(t *testing.T) {
		cases := []struct {
			hash string
			desc string
		}{
			{"123", "too short"},
			{"1234567890123456789012345678901234567890extra", "too long"},
			{"123456789012345678901234567890123456789g", "non-hex character"},
			{"", "empty"},
		}

		for _, tc := range cases {
			_, err := New(tc.hash, "", "John Doe <john@example.com>", "test")
			if err == nil {
				t.Errorf("Expected error for invalid tree hash (%s)", tc.desc)
			}
		}
	})
	t.Run("1.4: Invalid parent hash", func(t *testing.T) {
		validTree := "1234567890123456789012345678901234567890"
		cases := []struct {
			hash string
			desc string
		}{
			{"123", "too short"},
			{"1234567890123456789012345678901234567890extra", "too long"},
			{"123456789012345678901234567890123456789g", "non-hex character"},
		}

		for _, tc := range cases {
			_, err := New(validTree, tc.hash, "John Doe <john@example.com>", "test")
			if err == nil {
				t.Errorf("Expected error for invalid parent hash (%s)", tc.desc)
			}
		}
	})
	t.Run("1.5: Invalid author format", func(t *testing.T) {
		validTree := "1234567890123456789012345678901234567890"
		cases := []struct {
			author string
			desc   string
		}{
			{"John Doe", "missing email"},
			{"<john@example.com>", "missing name"},
			{"John Doe john@example.com", "missing brackets"},
			{"", "empty"},
		}

		for _, tc := range cases {
			_, err := New(validTree, "", tc.author, "test")
			if err == nil {
				t.Errorf("Expected error for invalid author format (%s)", tc.desc)
			}
		}
	})

	t.Run("1.6: Invalid message", func(t *testing.T) {
		validTree := "1234567890123456789012345678901234567890"
		validAuthor := "John Doe <john@example.com>"

		_, err := New(validTree, "", validAuthor, "")
		if err == nil {
			t.Error("Expected error for empty message")
		}
	})

	t.Run("1.7: Multi-line message", func(t *testing.T) {
		validTree := "1234567890123456789012345678901234567890"
		validAuthor := "John Doe <john@example.com>"
		message := "First line\nSecond line\nThird line"

		commit, err := New(validTree, "", validAuthor, message)
		if err != nil {
			t.Fatalf("Failed to create commit with multi-line message: %v", err)
		}

		if commit.Message != message {
			t.Errorf("Message not preserved exactly. Got %q, want %q", commit.Message, message)
		}
	})

}

func TestCommitStorage(t *testing.T) {
	t.Run("2.1: Write and read commit", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}

		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "objects"), 0755)
		defer os.RemoveAll(testDir)

		treeHash := "1234567890123456789012345678901234567890"
		parentHash := "abcdef1234567890abcdef1234567890abcdef12"
		author := "John Doe <john@example.com>"
		message := "Test commit message"

		commit, err := New(treeHash, parentHash, author, message)
		if err != nil {
			t.Fatalf("Failed to create commit: %v", err)
		}

		hash, err := commit.Write(filepath.Join(testDir, ".gitgo", "objects"))
		if err != nil {
			t.Fatalf("Failed to write commit: %v", err)
		}

		readCommit, err := Read(filepath.Join(testDir, ".gitgo", "objects"), hash)
		if err != nil {
			t.Fatalf("Failed to read commit: %v", err)
		}

		if readCommit.TreeHash != commit.TreeHash {
			t.Errorf("Tree hash mismatch: got %s, want %s", readCommit.TreeHash, commit.TreeHash)
		}
		if readCommit.ParentHash != commit.ParentHash {
			t.Errorf("Parent hash mismatch: got %s, want %s", readCommit.ParentHash, commit.ParentHash)
		}
		if readCommit.Author != commit.Author {
			t.Errorf("Author mismatch: got %s, want %s", readCommit.Author, commit.Author)
		}
		if readCommit.Message != commit.Message {
			t.Errorf("Message mismatch: got %s, want %s", readCommit.Message, commit.Message)
		}

		/*
						This additional logic was added since we didn't store nanoseconds
						on disk and regular date comparison was failing. With this logic we check if difference between
			            written and read commit is greater than 1 second.
		*/
		timeDiff := readCommit.AuthorDate.Sub(commit.AuthorDate)
		if timeDiff.Seconds() > 1 {
			t.Errorf("Dates differ by more than 1 second: got %v, want %v",
				readCommit.AuthorDate, commit.AuthorDate)
		}
	})

	t.Run("2.2: Multi-line commit message", func(t *testing.T) {
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "objects"), 0755)
		defer os.RemoveAll(testDir)

		message := "First line\nSecond line\nThird line"
		commit, _ := New(
			"1234567890123456789012345678901234567890",
			"",
			"John Doe <john@example.com>",
			message,
		)

		hash, err := commit.Write(filepath.Join(testDir, ".gitgo", "objects"))
		if err != nil {
			t.Fatalf("Failed to write commit: %v", err)
		}

		readCommit, err := Read(filepath.Join(testDir, ".gitgo", "objects"), hash)
		if err != nil {
			t.Fatalf("Failed to read commit: %v", err)
		}

		if readCommit.Message != message {
			t.Errorf("Multi-line message not preserved.\nGot:\n%s\nWant:\n%s",
				readCommit.Message, message)
		}
	})

	t.Run("2.3: First commit (no parent)", func(t *testing.T) {
		cwd, _ := os.Getwd()
		testDir := filepath.Join(cwd, "testdata")
		os.RemoveAll(testDir)
		os.MkdirAll(filepath.Join(testDir, ".gitgo", "objects"), 0755)
		defer os.RemoveAll(testDir)

		commit, _ := New(
			"1234567890123456789012345678901234567890",
			"",
			"John Doe <john@example.com>",
			"First commit",
		)

		hash, err := commit.Write(filepath.Join(testDir, ".gitgo", "objects"))
		if err != nil {
			t.Fatalf("Failed to write commit: %v", err)
		}

		readCommit, err := Read(filepath.Join(testDir, ".gitgo", "objects"), hash)
		if err != nil {
			t.Fatalf("Failed to read commit: %v", err)
		}

		if readCommit.ParentHash != "" {
			t.Errorf("Expected empty parent hash, got %s", readCommit.ParentHash)
		}
	})
}
