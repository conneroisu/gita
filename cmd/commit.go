package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/conneroisu/groq-go"
)

// commit commits the change.
func commit(resp *groq.ChatResponse) error {

	// split the response into lines
	lines := strings.Split(resp.Choices[0].Message.Content, "\n")
	ms := huh.NewMultiSelect[string]()
	opts := huh.NewOptions[string]()

	for _, line := range lines {
		opts = append(opts, huh.Option[string]{
			Key:   line,
			Value: line,
		})
	}
	var commitMessages *[]string
	ms.Options(opts...)
	ms.Title("Select a commit message.")
	ms.Limit(10)
	ms.Value(commitMessages)
	err := ms.Run()
	if err != nil {
		return err
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "commitmsg")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	// Write the selected messages to the temporary file
	_, err = tmpFile.Write(toBody(commitMessages))
	if err != nil {
		return err
	}
	tmpFile.Close()

	// Open the temporary file in the editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	editorCmd := exec.Command(editor, tmpFile.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	err = editorCmd.Run()
	if err != nil {
		return err
	}

	// Check if the file is not empty and commit
	fileInfo, err := os.Stat(tmpFile.Name())
	if err != nil {
		return err
	}

	if fileInfo.Size() > 0 {
		gitCommitCmd := exec.Command("git", "commit", "-F", tmpFile.Name())
		gitCommitCmd.Stdin = os.Stdin
		gitCommitCmd.Stdout = os.Stdout
		gitCommitCmd.Stderr = os.Stderr
		err = gitCommitCmd.Run()
		if err != nil {
			fmt.Println("Error committing changes:", err)
		}
	} else {
		fmt.Println("Commit message is empty, commit aborted.")
	}
	return nil
}

func toBody(commitMessages *[]string) []byte {
	var body string
	for _, commitMessage := range *commitMessages {
		body += commitMessage + "\n"
	}
	return []byte(body)
}
