package commit

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type Commit struct {
	TreeHash   string
	ParentHash string
	Author     string
	AuthorDate time.Time
	Message    string
}

func New(treeHash string, parentHash string, author string, message string) (*Commit, error) {
	if len(treeHash) != 40 {
		return nil, fmt.Errorf("expected treehash to have length 40, got %d", len(treeHash))
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(treeHash) {
		return nil, fmt.Errorf("tree hash must contain only hex characters")
	}
	if parentHash != "" && len(parentHash) != 40 {
		return nil, fmt.Errorf("if present, parent hash must be 40 characters, got %d", len(parentHash))
	}
	if parentHash != "" && !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(parentHash) {
		return nil, fmt.Errorf("parent hash must contain only hex characters")
	}
	if len(message) == 0 {
		return nil, fmt.Errorf("commit message cannot be empty")
	}
	authorRegex := regexp.MustCompile(`^([^<]+)\s+<([^>]+)>$`)
	if !authorRegex.MatchString(author) {
		return nil, fmt.Errorf("invalid author format, must be 'Name <email>'")
	}
	commit := Commit{
		TreeHash:   treeHash,
		ParentHash: parentHash,
		Author:     author,
		AuthorDate: time.Now(),
		Message:    message,
	}
	return &commit, nil

}

func (c *Commit) Write(objectsPath string) (string, error) {
	timestamp := c.AuthorDate.Unix()
	timezone := c.AuthorDate.Format("-0700")

	content := fmt.Sprintf("tree %s\n", c.TreeHash)
	if c.ParentHash != "" {
		content += fmt.Sprintf("parent %s\n", c.ParentHash)
	}
	content += fmt.Sprintf("author %s %d %s\n\n%s",
		c.Author,
		timestamp,
		timezone,
		c.Message)
	data := fmt.Sprintf("commit %d\x00%s", len(content), content)
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write([]byte(data)); err != nil {
		return "", fmt.Errorf("failed to compress data: %v", err)
	}
	zw.Close()

	hash := sha1.Sum(compressed.Bytes())
	hashStr := hex.EncodeToString(hash[:])
	hashPath := filepath.Join(objectsPath, hashStr[:2], hashStr[2:])
	if err := os.MkdirAll(filepath.Dir(hashPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create object directory: %v", err)
	}

	// Write file
	if err := os.WriteFile(hashPath, compressed.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write object file: %v", err)
	}

	return hashStr, nil
}

func Read(objectsPath, hash string) (*Commit, error) {
	hashPath := filepath.Join(objectsPath, hash[:2], hash[2:])
	compressed, err := os.ReadFile(hashPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read object file: %v", err)
	}

	zr, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %v", err)
	}
	defer zr.Close()

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, zr); err != nil {
		return nil, fmt.Errorf("failed to decompress data: %v", err)
	}

	data := buffer.Bytes()
	parts := bytes.SplitN(data, []byte{0}, 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format")
	}

	header := bytes.Fields(parts[0])
	if len(header) != 2 || string(header[0]) != "commit" {
		return nil, fmt.Errorf("not a commit object")
	}

	content := parts[1]
	lines := bytes.Split(content, []byte{'\n'})

	var treeHash, parentHash, author string
	var authorTime time.Time
	var message string

	messageStart := 0
	for i, line := range lines {
		if len(line) == 0 {
			messageStart = i + 1
			break
		}

		fields := bytes.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid line format")
		}

		switch string(fields[0]) {
		case "tree":
			treeHash = string(fields[1])
		case "parent":
			parentHash = string(fields[1])
		case "author":
			if len(fields) < 4 {
				return nil, fmt.Errorf("invalid author line")
			}
			authorEnd := len(fields) - 2
			author = string(bytes.Join(fields[1:authorEnd], []byte(" ")))

			timestamp, err := strconv.ParseInt(string(fields[authorEnd]), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid timestamp: %v", err)
			}

			timezone := string(fields[authorEnd+1])
			tzHours, err := strconv.Atoi(timezone[1:3])
			if err != nil {
				return nil, fmt.Errorf("invalid timezone hours: %v", err)
			}
			tzMinutes, err := strconv.Atoi(timezone[3:])
			if err != nil {
				return nil, fmt.Errorf("invalid timezone minutes: %v", err)
			}
			tzOffset := (tzHours*60 + tzMinutes) * 60
			if timezone[0] == '-' {
				tzOffset = -tzOffset
			}
			authorTime = time.Unix(timestamp, 0).In(time.FixedZone("", tzOffset))
		}
	}

	message = string(bytes.Join(lines[messageStart:], []byte{'\n'}))

	commit := &Commit{
		TreeHash:   treeHash,
		ParentHash: parentHash,
		Author:     author,
		AuthorDate: authorTime,
		Message:    message,
	}

	return commit, nil
}
