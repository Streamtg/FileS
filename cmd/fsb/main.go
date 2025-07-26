Skip to content
Navigation Menu
EverythingSuckz
TG-FileStreamBot

Type / to search
Code
Issues
29
Pull requests
3
Discussions
Actions
Projects
1
Security
Insights
Files
Go to file
t
.github
.vscode
cmd/fsb
main.go
run.go
session.go
config
internal
pkg
.gitattributes
.gitignore
.goreleaser.yaml
Dockerfile
LICENSE
Procfile
README.md
app.json
docker-compose.yaml
fsb.sample.env
go.mod
go.sum
goreleaser.Dockerfile
TG-FileStreamBot/cmd/fsb
/main.go
EverythingSuckz
EverythingSuckz
chore(version): release 3.1.0
6261681
 · 
5 months ago

Code

Blame
37 lines (31 loc) · 851 Bytes
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
	Long:              "Telegram Bot to generate direct streamable links for telegram media.",
	Example:           "fsb run --port 8080",
	Version:           versionString,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	config.SetFlagsFromConfig(runCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(sessionCmd)
	rootCmd.SetVersionTemplate(fmt.Sprintf(`Telegram File Stream Bot version %s`, versionString))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
