package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var editCommand = &cobra.Command{
	Use:   "edit",
	Args:  cobra.MinimumNArgs(1),
	Short: "Edit a file",
	RunE:  makeRunE(config.runEditCommand),
}

type editCommandConfig struct {
	apply  bool
	diff   bool
	prompt bool
}

func init() {
	rootCommand.AddCommand(editCommand)

	persistentFlags := editCommand.PersistentFlags()
	persistentFlags.BoolVarP(&config.edit.apply, "apply", "a", false, "apply edit after editing")
	persistentFlags.BoolVarP(&config.edit.diff, "diff", "d", false, "print diff after editing")
	persistentFlags.BoolVarP(&config.edit.prompt, "prompt", "p", false, "prompt before applying (implies --diff)")
}

func (c *Config) runEditCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	if c.edit.prompt {
		c.edit.diff = true
	}
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(targetState, args)
	if err != nil {
		return err
	}
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	argv := []string{}
	for _, entry := range entries {
		argv = append(argv, filepath.Join(c.SourceDir, entry.SourceName()))
	}
	if !c.edit.diff && !c.edit.apply {
		return c.exec(append([]string{editor}, argv...))
	}
	if c.Verbose {
		fmt.Printf("%s %s\n", editor, strings.Join(argv, " "))
	}
	cmd := exec.Command(editor, argv...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	readOnlyFS := vfs.NewReadOnlyFS(fs)
	applyActuator := c.getDefaultActuator(fs)
	for i, entry := range entries {
		anyActuator := chezmoi.NewAnyActuator(chezmoi.NullActuator)
		var actuator chezmoi.Actuator = anyActuator
		if c.edit.diff {
			actuator = chezmoi.NewLoggingActuator(os.Stdout, actuator)
		}
		if err := entry.Apply(readOnlyFS, targetState.TargetDir, targetState.Umask, actuator); err != nil {
			return err
		}
		if c.edit.apply && anyActuator.Actuated() {
			if c.edit.prompt {
				choice, err := prompt(fmt.Sprintf("Apply %s", args[i]), "ynqa")
				if err != nil {
					return err
				}
				switch choice {
				case 'y':
				case 'n':
					continue
				case 'q':
					return nil
				case 'a':
					c.edit.prompt = false
				}
			}
			if err := entry.Apply(readOnlyFS, targetState.TargetDir, targetState.Umask, applyActuator); err != nil {
				return err
			}
		}
	}
	return nil
}
