package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/charmbracelet/huh"
	"github.com/conneroisu/groq-go"
	"github.com/spf13/viper"
)

// CommitMessages is a list of commit messages.
type CommitMessages string

func (c CommitMessages) HuhOptions() []huh.Option[string] {
	opts := huh.NewOptions[string]()
	for _, line := range strings.Split(string(c), "\n") {
		opts = append(opts, huh.Option[string]{
			Key:   line,
			Value: line,
		})
	}
	return opts
}

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(home, ".gita")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	err = viper.ReadInConfig()
	if err != nil {
		err = createConfig(configDir)
		if err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		err = viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("failed to read config after creating it: %w", err)
		}
	}

	diff, err := getDiff()
	if err != nil {
		return err
	}

	prompt, err := fillTemplate(diff)
	if err != nil {
		return err
	}

	client, err := groq.NewClient(viper.GetString("groq_key"))
	if err != nil {
		return err
	}

	resp, err := client.ChatCompletion(ctx, groq.ChatCompletionRequest{
		Model: groq.ModelLlama3370BVersatile,
		Messages: []groq.ChatCompletionMessage{
			{
				Role:    "system",
				Content: prompt,
			},
		},
	})
	if err != nil {
		return err
	}
	if len(resp.Choices) == 0 {
		return errors.New("no choices from chat completion")
	}
	commitMsgs := CommitMessages(resp.Choices[0].Message.Content)

	var selected []string
	err = huh.NewMultiSelect[string]().
		Title("Select a commit message.").
		Limit(10).
		Options(commitMsgs.HuhOptions()...).
		Value(&selected).
		Run()
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp("", "commitmsg")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(strings.Join(selected, "\n")))
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "commit", "-F", tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var (
	//go:embed prompt.tmpl
	promptStr      string
	promptTemplate = template.Must(template.New("prompt").Parse(promptStr))
)

func fillTemplate(diff string) (string, error) {
	var (
		w   bytes.Buffer
		err error
	)
	err = promptTemplate.Execute(&w, struct {
		Diff string
	}{
		Diff: diff,
	})
	return w.String(), err
}

func getDiff() (string, error) {
	var w, ew bytes.Buffer
	cmd := exec.Command(
		"git",
		"diff",
		"--cached",
		"--",
		":!go.sum",
	)
	cmd.Stdout = &w
	cmd.Stderr = &ew
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	if ew.Len() > 0 {
		return "", errors.New(ew.String())
	}
	return w.String(), nil
}

func createConfig(configDir string) error {
	var key string
	err := huh.NewInput().
		Title("What's your groq key?").
		Value(&key).
		Run()
	if err != nil {
		return err
	}

	viper.Set("groq_key", key)

	// Create gita directory if it doesn't exist
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.yaml")
	err = viper.WriteConfigAs(configFile)
	if err != nil {
		return err
	}

	return nil
}
