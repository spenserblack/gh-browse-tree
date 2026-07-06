// Package browser provides the file browser using the GitHub API.
package browser

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/prompter"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// browser is the file browser state.
type browser struct {
	// repo is the Git repository information.
	repo repository.Repository
	// rest is the REST client.
	rest *api.RESTClient
	// display is the program used to display file contents, which are received from stdin.
	// The length must always be at least 1. It is a collection of the command and any arguments.
	display []string
	// path is the current path in the repository.
	path string
	// prompter is the utility used to prompt for choices.
	prompter *prompter.Prompter
	// stdout is where commands should write to.
	stdout *os.File
}

// Start starts the browser.
func Start(
	repo repository.Repository,
	rest *api.RESTClient,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	display []string,
) error {
	stdoutF, ok := stdout.(*os.File)
	if !ok {
		return errors.New("Couldn't cast stdout to a file")
	}
	stderrF, ok := stderr.(*os.File)
	if !ok {
		return errors.New("Couldn't cast stderr to a file")
	}
	stdinF, ok := stdin.(*os.File)
	if !ok {
		return errors.New("Couldn't cast stdin to a file")
	}
	prompter := prompter.New(stdinF, stdoutF, stderrF)
	b := browser{
		repo:     repo,
		rest:     rest,
		prompter: prompter,
		display:  display,
		path:     "",
		stdout:   stdoutF,
	}
	for {
		contents, err := b.GetContents()
		if err != nil {
			return err
		}
		entry, up, err := b.Prompt(contents)
		if err != nil {
			return err
		}
		if up {
			b.path = path.Dir(b.path)
		} else if entry.Type == "dir" {
			b.path = entry.Path
		} else if entry.Type == "file" {
			if err := b.DisplayFile(entry.Path); err != nil {
				return err
			}
		} else {
			panic("unreachable: unsupported type")
		}
	}
}

// GetContents gets the contents at the current path.
func (b browser) GetContents() ([]entry, error) {
	response := []entry{}
	err := b.rest.Get(b.Endpoint(), &response)
	return response, err
}

// Prompt prompts for a single entry selection.
func (b browser) Prompt(entries []entry) (entry entry, up bool, err error) {
	const parent string = ".."
	choices := make([]string, 0, len(entries)+1)
	choices = append(choices, parent)
	for _, entry := range entries {
		choices = append(choices, entry.Name)
	}
	index, err := b.prompter.Select(b.path, "", choices)
	if err != nil {
		return entry, false, err
	}
	if choices[index] == parent {
		up = true
		return
	}
	return entries[index-1], false, nil
}

// BaseEndpoint gets the base endpoint for the contents API.
func (b browser) BaseEndpoint() string {
	return fmt.Sprintf("repos/%s/%s/contents", b.repo.Name, b.repo.Owner)
}

// Endpoint gets the endpoint for the browser.
func (b browser) Endpoint() string {
	endpoint := b.BaseEndpoint()
	if b.path != "" && b.path != "." {
		endpoint = fmt.Sprintf("%s/%s", endpoint, b.path)
	}
	return endpoint
}

// DisplayFile displays a file's contents.
func (b browser) DisplayFile(path string) error {
	display := exec.Command(b.display[0], b.display[1:]...)
	display.Stdout = b.stdout
	stdin, err := display.StdinPipe()
	if err != nil {
		return err
	}
	endpoint := b.BaseEndpoint()
	endpoint = fmt.Sprintf("%s/%s", endpoint, path)
	response := struct {
		// Content is a base64 encoded string.
		Content string
	}{}
	err = b.rest.Get(endpoint, &response)
	if err != nil {
		return err
	}
	decoded, err := base64.StdEncoding.DecodeString(response.Content)
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		stdin.Write(decoded)
	}()
	return display.Run()
}
