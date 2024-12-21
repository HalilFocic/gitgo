package main

import (
	"flag"
	"fmt"
	"github.com/HalilFocic/gitgo/internal/commands"
	"github.com/HalilFocic/gitgo/internal/repository"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: gitgo <command> [<args>]")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		initCmd.Parse(os.Args[2:])
		_, err := repository.Init(cwd)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Initialized empty gitgo repository in %s\n", cwd)

	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		addCmd.Parse(os.Args[2:])
		if addCmd.NArg() < 1 {
			fmt.Println("error: path required for 'add'")
			os.Exit(1)
		}
		for _, path := range addCmd.Args() {
			cmd := commands.NewAddCommand(cwd, path)
			if err := cmd.Execute(); err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
		}
		fmt.Printf("Sucessfully added %d files to index.\n", len(addCmd.Args()))
	case "remove":
		rmCmd := flag.NewFlagSet("remove", flag.ExitOnError)
		rmCmd.Parse(os.Args[2:])
		if rmCmd.NArg() < 1 {
			fmt.Println("error: path required for 'remove'")
			os.Exit(1)
		}
		for _, path := range rmCmd.Args() {
			cmd := commands.NewRemoveCommand(cwd, path)
			if err := cmd.Execute(); err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
		}
		fmt.Printf("Sucessfully removed %d files to index.\n", len(rmCmd.Args()))

	case "commit":
		commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)
		message := commitCmd.String("m", "", "commit message")
		commitCmd.Parse(os.Args[2:])
		if *message == "" {
			fmt.Println("error: -m flag required")
			os.Exit(1)
		}
		cmd := commands.NewCommitCommand(cwd, *message, "User <user@example.com>")
		if err := cmd.Execute(); err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Commit added sucessfully\n")

	case "branch":
		branchCmd := flag.NewFlagSet("branch", flag.ExitOnError)
		create := branchCmd.Bool("c", false, "create new branch")
		delete := branchCmd.Bool("d", false, "delete branch")
		branchCmd.Parse(os.Args[2:])

		if *create && branchCmd.NArg() == 1 {
			cmd := commands.NewBranchCommand(cwd, branchCmd.Arg(0), "create")
			if err := cmd.Execute(); err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
		} else if *delete && branchCmd.NArg() == 1 {
			cmd := commands.NewBranchCommand(cwd, branchCmd.Arg(0), "delete")
			if err := cmd.Execute(); err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
		} else {
			cmd := commands.NewBranchCommand(cwd, "", "list")
			if err := cmd.Execute(); err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
		}

	case "checkout":
		checkoutCmd := flag.NewFlagSet("checkout", flag.ExitOnError)
		checkoutCmd.Parse(os.Args[2:])
		if checkoutCmd.NArg() != 1 {
			fmt.Println("error: branch name or commit hash required")
			os.Exit(1)
		}
		cmd := commands.NewCheckoutCommand(cwd, checkoutCmd.Arg(0))
		if err := cmd.Execute(); err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
	case "log":
		logCmd := flag.NewFlagSet("log", flag.ExitOnError)
		maxCount := logCmd.Int("n", -1, "limit number of commits")
		logCmd.Parse(os.Args[2:])

		cmd := commands.NewLogCommand(cwd, *maxCount)
		if err := cmd.Execute(); err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
