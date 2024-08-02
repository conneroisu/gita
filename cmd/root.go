package cmd

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"os"

	_ "embed"

	groq "github.com/conneroisu/groq-go"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		diff, err := getDiff()
		if err != nil {
			return err
		}
		if len(diff) < 9 {
			return fmt.Errorf("no diff found")
		}
		fmt.Println(diff)
		w := bytes.NewBuffer([]byte{})
		err = fillTemplate(w, diff)
		if err != nil {
			return err
		}
		client, err := groq.NewClient(os.Getenv("GROQ_KEY"))
		if err != nil {
			return err
		}
		resp, err := client.Chat(ctx, groq.ChatRequest{
			Model: "llama3-8b-8192",
			Messages: []groq.Message{
				{
					Role:    "system",
					Content: w.String(),
				},
			},
			TopP:      0.9,
			MaxTokens: 1024,
			Stop:      nil,
			Stream:    false,
			Format: struct {
				Type groq.Format "json:\"type\""
			}{
				Type: groq.FormatText,
			},
		})
		if err != nil {
			return err
		}
		log.Println(resp)
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
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
