package commands

import (
	"fmt"

	"github.com/HalilFocic/gitgo/internal/staging"
)

type RemoveCommand struct {
	rootPath string
	path     string
}

func NewRemoveCommand(rootPath, path string) *RemoveCommand {
	return &RemoveCommand{
		rootPath: rootPath,
		path:     path,
	}
}

func (c *RemoveCommand) Execute() error {
	index, err := staging.New(c.rootPath)
	if err != nil {
		return fmt.Errorf("failed to read staging area: %v", err)
	}
	if err := index.Remove(c.path); err != nil {
		return fmt.Errorf("failed to remove file %s from index: %v", c.path, err)
	}
	return nil
}
