package blob

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBlobOperations(t *testing.T) {
    t.Run("1.2: Create and store blob", func(t *testing.T) {
        content := []byte("test content")
        
        // Create new blob
        b, err := New(content)
        if err != nil {
            t.Fatalf("Failed to create blob: %v", err)
        }

        // Verify hash format
        if len(b.Hash()) != 40 {
            t.Errorf("Invalid hash length: got %d, want 40", len(b.Hash()))
        }

        // Store the blob
        err = b.Store(".gitgo/objects")
        if err != nil {
            t.Fatalf("Failed to store blob: %v", err)
        }

        // Check if file exists in correct location
        hash := b.Hash()
        objectPath := filepath.Join(".gitgo/objects", hash[:2], hash[2:])
        if _, err := os.Stat(objectPath); os.IsNotExist(err) {
            t.Error("Blob file was not created in correct location")
        }
    })

    t.Run("1.3: Read blob content", func(t *testing.T) {
        content := []byte("test content")
        originalBlob, _ := New(content)
        originalBlob.Store(".gitgo/objects")

        // Read blob back
        readBlob, err := Read(".gitgo/objects", originalBlob.Hash())
        if err != nil {
            t.Fatalf("Failed to read blob: %v", err)
        }
        // Verify content matches
        if !bytes.Equal(readBlob.Content(), content) {
            t.Error("Retrieved content doesn't match original")
        }
    })
}
