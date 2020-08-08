package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "gopack",
	Short: "Gopack is a go package management tool",
	Long: "Gopack is a package management tool which focuses mainly on simplicity of package management in golang",
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
