package commands

import (
	"fmt"
	"github.com/HalilFocic/gitgo/internal/staging"
)

type AddCommand struct {
	rootPath string
	path     string
}

func NewAddCommand(rootPath, path string) *AddCommand {
	return &AddCommand{
		rootPath: rootPath,
		path:     path,
	}
}

func (c *AddCommand) Execute() error {
	index, err := staging.New(c.rootPath)
	if err != nil {
		return fmt.Errorf("failed to read staging area: %v", err)
	}

	if err := index.Add(c.path); err != nil {
		return fmt.Errorf("failed to add %s: %v", c.path, err)
	}

	return nil
}
