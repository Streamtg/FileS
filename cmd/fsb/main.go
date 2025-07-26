package main

import (
	"EverythingSuckz/fsb/config"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const versionString = "3.1.0"

var rootCmd = &cobra.Command{
	Use:               "fsb [command]",
	Short:             "Telegram File Stream Bot",
	Long:              "Telegram Bot to generate direct streamable links for Telegram media.",
	Example:           "fsb run --port 8080",
	Version:           versionString,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	// Establece flags de configuración desde el archivo de configuración
	config.SetFlagsFromConfig(runCmd)

	// Registra los subcomandos
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(sessionCmd)

	// Establece el template de versión
	rootCmd.SetVersionTemplate(fmt.Sprintf(`Telegram File Stream Bot version %s`, versionString))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
