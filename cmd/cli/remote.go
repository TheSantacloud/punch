package cli

import (
	"fmt"

	"github.com/dormunis/punch/pkg/config"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [remote]",
	Short: "pushes the current day to the remote server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "manage remote servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		for k, v := range Config.Remotes {
			if verbose {
				fmt.Printf("%s (%s)\n", k, v.(config.Remote).Type())
			} else {
				fmt.Println(k)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
