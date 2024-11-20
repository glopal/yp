package cmd

import (
	"log"

	"github.com/glopal/yp/yplib"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "yamlplus",
		Short: "Extended yaml rendering",
		Run: func(cmd *cobra.Command, args []string) {
			err := yplib.WithOptions(yplib.OmitDotFiles()).Load(args...).Out()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}
