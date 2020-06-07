package cmd

import (
	"envgen/generator"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Version:       "2.2.0",
	SilenceErrors: true,
	Use:           "envgen <configFilePath> [envFile1] ... [envFileN]",
	Short:         "envgen generates env files for sub packages",
	Long:          "envgen is CLI tool that generates .env files for subpackages in your project based on a configuration file",
	Args:          cobra.MinimumNArgs(1),
	RunE:          runGenerator,
}

func Execute() {
	envFiles := os.Args[2:]
	if len(envFiles) > 0 {
		err := godotenv.Load(envFiles...)
		if err != nil {
			rootCmd.PrintErr("Error loading .env file")
		}
	}

	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErr(err)
		os.Exit(1)
	}
}

func runGenerator(cmd *cobra.Command, args []string) error {
	gen, err := generator.New(args[0])
	if err != nil {
		return err
	}

	gen.GenerateFiles()
	fmt.Println("Finished env files generation!")

	return nil
}
