package commands

import (
	"fmt"
	"github.com/HalilFocic/gitgo/internal/refs"
)

type BranchCommand struct {
	rootPath string
	name     string
	action   string
}

func NewBranchCommand(rootPath, name, action string) *BranchCommand {
	return &BranchCommand{
		rootPath: rootPath,
		name:     name,
		action:   action,
	}
}

func (c *BranchCommand) Execute() error {
	switch c.action {
	case "create":
		head, err := refs.ReadHead(c.rootPath)
		if err != nil {
			return fmt.Errorf("failed to read HEAD: %v", err)
		}
		if head.Type == refs.RefTypeSymbolic {
			ref, err := refs.ReadRef(c.rootPath, head.Target)
			if err != nil {
				return fmt.Errorf("failed to read current branch: %v", err)
			}
			head = ref
		}
		if err := refs.CreateBranch(c.rootPath, c.name, head.Target); err != nil {
			return fmt.Errorf("failed to create branch: %v", err)
		}

	case "delete":
		if err := refs.DeleteBranch(c.rootPath, c.name); err != nil {
			return fmt.Errorf("failed to delete branch: %v", err)
		}

	case "list":
		branches, err := refs.ListBranches(c.rootPath)
		if err != nil {
			return fmt.Errorf("failed to list branches: %v", err)
		}

		head, err := refs.ReadHead(c.rootPath)
		if err != nil {
			return fmt.Errorf("failed to read HEAD: %v", err)
		}

		currentBranch := ""
		if head.Type == refs.RefTypeSymbolic {
			currentBranch = head.Target[len("refs/heads/"):]
		}

		for _, branch := range branches {
			if branch == currentBranch {
				fmt.Printf("* %s\n", branch)
			} else {
				fmt.Printf("  %s\n", branch)
			}
		}

	default:
		return fmt.Errorf("unknown branch action: %s", c.action)
	}

	return nil
}
