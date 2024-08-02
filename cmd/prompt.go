package cmd

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

// fillTemplate fills the template with the given diff.
func fillTemplate(w io.Writer, diff string) error {
	return promptTemplate.Execute(w, struct {
		Diff string
	}{
		Diff: diff,
	})
}

// getDiff returns the diff from the current git repository.
// It returns an error if the diff cannot be retrieved.
// Equivalent to `git diff --cached`.
func getDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	w := bytes.NewBuffer([]byte{})
	ew := bytes.NewBuffer([]byte{})
	cmd.Stdout = w
	cmd.Stderr = ew
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	if ew.Len() > 0 {
		return "", errors.New(ew.String())
	}
	return w.String(), nil
}
