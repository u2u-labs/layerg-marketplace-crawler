package cmd

import (
	"github.com/spf13/cobra"
)

// workerCmd represents the api command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start a multichain crawler API server",
	Long:  `Start a multichain crawler API server.`,
	Run:   startWorker,
}

func init() {
	rootCmd.AddCommand(workerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// apiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// apiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
