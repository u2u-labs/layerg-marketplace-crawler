/*
Copyright © 2024 LayerG team
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/u2u-labs/layerg-crawler/cmd/libs"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "layerg-crawler",
		Short: "Start a multichain crawler",
		Long:  `Start a multichain crawler.`,
		Run: func(cmd *cobra.Command, args []string) {
			//go startApi(cmd, args)
			//go startWorker(cmd, args)
			startCrawler(cmd, args)
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.layerg-crawler.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find working directory.
		home, err := os.Getwd()
		cobra.CheckErr(err)

		// Search config in working directory with name ".layerg-crawler" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".layerg-crawler")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	libs.PusherClient = libs.NewPusherClient(
		viper.GetString("PUSHER_APP_ID"),
		viper.GetString("PUSHER_KEY"),
		viper.GetString("PUSHER_SECRET"),
		viper.GetString("PUSHER_CLUSTER"),
	)
	if libs.PusherClient == nil {
		log.Fatal("Failed to initialize Pusher client")
	}

	libs.QUEUE_URL = viper.GetString("QUEUE_URL")
	libs.AWS_REGION = viper.GetString("AWS_REGION")
	libs.InitSQSClient(libs.InitAWSConfig())
}
