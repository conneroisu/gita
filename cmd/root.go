/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"

	_ "embed"

	groq "github.com/conneroisu/go-groq"
	"github.com/spf13/cobra"
)

var (
	//go:embed prompt.md
	prompt         string
	promptTemplate = template.Must(template.New("prompt").Parse(prompt))
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gita",
	Short: "Groq Based Git Commit Message Generator.",
	Long: `
Gita is a groq based git commit message generator.
	
It utilizes the groq api to generate a commit message based on the current git repository state.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		diff, err := getDiff()
		if err != nil {
			return err
		}
		fmt.Println(diff)
		w := bytes.NewBuffer([]byte{})
		err = fillTemplate(w, diff)
		if err != nil {
			return err
		}
		client := groq.NewClient(os.Getenv("GROQ_KEY"), http.DefaultClient)
		resp, err := client.Chat(groq.ChatRequest{
			Model:  "llama-3.1-405b-reasoning",
			Stream: false,
			Messages: []groq.Message{
				{
					Role:    "system",
					Content: w.String(),
				},
			},
		})
		if err != nil {
			return err
		}
		cmd.Println(resp.Choices[0].Message.Content)
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gita.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
