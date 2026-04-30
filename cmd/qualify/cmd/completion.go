// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for Ark.

To load completions:

Bash:
  $ source <(ark completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ ark completion bash > /etc/bash_completion.d/ark
  # macOS:
  $ ark completion bash > $(brew --prefix)/etc/bash_completion.d/ark

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ ark completion zsh > "${fpath[1]}/_ark"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ ark completion fish | source

  # To load completions for each session, execute once:
  $ ark completion fish > ~/.config/fish/completions/ark.fish

PowerShell:
  PS> ark completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> ark completion powershell > ark.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}
