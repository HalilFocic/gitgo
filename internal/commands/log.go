package commands

import (
    "fmt"
    "path/filepath"
    "github.com/HalilFocic/gitgo/internal/commit"
    "github.com/HalilFocic/gitgo/internal/refs"
)

type LogCommand struct {
    rootPath string
    maxCount int
}

func NewLogCommand(rootPath string, maxCount int) *LogCommand {
    if maxCount <= 0 {
        maxCount = -1
    }
    return &LogCommand{
        rootPath: rootPath,
        maxCount: maxCount,
    }
}

func (c *LogCommand) Execute() error {
    objectsPath := filepath.Join(c.rootPath, ".gitgo", "objects")
    
    headRef, err := refs.ReadHead(c.rootPath)
    if err != nil {
        return fmt.Errorf("failed to read HEAD: %v", err)
    }

    currentCommitHash := headRef.Target
    commitCount := 0

    for currentCommitHash != "" {
        if c.maxCount != -1 && commitCount >= c.maxCount {
            break
        }

        currentCommit, err := commit.Read(objectsPath, currentCommitHash)
        if err != nil {
            return fmt.Errorf("failed to read commit %s: %v", currentCommitHash, err)
        }

        fmt.Printf("commit %s\n", currentCommitHash)
        fmt.Printf("Author: %s\n", currentCommit.Author)
        fmt.Printf("Date: %v\n", currentCommit.AuthorDate.Format("Mon Jan 2 15:04:05 2006 -0700"))
        fmt.Printf("\n    %s\n\n", currentCommit.Message)

        currentCommitHash = currentCommit.ParentHash
        commitCount++
    }

    if commitCount == 0 {
        fmt.Println("No commits found")
    }

    return nil
}

