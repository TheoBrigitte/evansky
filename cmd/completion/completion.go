package completion

import (
	"os"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

$ source <(evansky completion bash)

# To load completions for each session, execute once:
Linux:
  $ evansky completion bash > /etc/bash_completion.d/evansky
MacOS:
  $ evansky completion bash > /usr/local/etc/bash_completion.d/evansky

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ evansky completion zsh > "${fpath[1]}/_evansky"

# You will need to start a new shell for this setup to take effect.

Fish:

$ evansky completion fish | source

# To load completions for each session, execute once:
$ evansky completion fish > ~/.config/fish/completions/evansky.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}
