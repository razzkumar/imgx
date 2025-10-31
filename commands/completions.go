package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

const bashSetup = `# Bash completion setup for imgx
# Add this to your ~/.bashrc or ~/.bash_profile:

# Option 1: Source directly
PROG=imgx source <(imgx --generate-shell-completion)

# Option 2: Save to file (recommended)
imgx --generate-shell-completion > /etc/bash_completion.d/imgx
# or for user-only:
imgx --generate-shell-completion > ~/.local/share/bash-completion/completions/imgx
`

const zshSetup = `# Zsh completion setup for imgx
# Add this to your ~/.zshrc:

# Option 1: Source directly
autoload -U compinit; compinit
PROG=imgx source <(imgx --generate-shell-completion)

# Option 2: Save to file (recommended)
# First, ensure completion directory is in fpath
# Add to ~/.zshrc:
fpath=(~/.zsh/completion $fpath)
autoload -U compinit; compinit

# Then save completion file:
imgx --generate-shell-completion > ~/.zsh/completion/_imgx
`

const fishSetup = `# Fish completion setup for imgx
# Fish completion is automatic once the file is in the right place.

# Save completion file:
imgx --generate-shell-completion > ~/.config/fish/completions/imgx.fish

# Reload completions (optional):
fish_update_completions
`

func CompletionsCommand() *cli.Command {
	return &cli.Command{
		Name:      "completions",
		Usage:     "Show shell completion setup instructions",
		ArgsUsage: "[SHELL]",
		Description: `Display instructions for setting up shell completions for imgx.

imgx uses urfave/cli's built-in completion system. Use the --generate-shell-completion
flag to generate completions dynamically.

Supported shells: bash, zsh, fish

Examples:
  # Show instructions for all shells
  imgx completions

  # Show instructions for bash only
  imgx completions bash

  # Generate completion script
  imgx --generate-shell-completion`,
		Action: completionsAction,
	}
}

func completionsAction(ctx context.Context, cmd *cli.Command) error {
	shell := ""
	if cmd.NArg() > 0 {
		shell = cmd.Args().First()
	}

	if shell == "" {
		// Show all shells
		fmt.Println("Shell Completion Setup for imgx")
		fmt.Println("================================")
		fmt.Println()
		fmt.Println("imgx supports dynamic shell completions using the --generate-shell-completion flag.")

		return nil
	}

	switch shell {
	case "bash":
		fmt.Print(bashSetup)
	case "zsh":
		fmt.Print(zshSetup)
	case "fish":
		fmt.Print(fishSetup)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}

	return nil
}
