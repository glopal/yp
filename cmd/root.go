package cmd

import (
	"log"

	"github.com/glopal/yamlplus/yamlp"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "yamlplus",
		Short: "Extended yaml rendering",
		Run: func(cmd *cobra.Command, args []string) {
			err := yamlp.Run(args, yamlp.OmitDotFiles())
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}
