// Package display gets a display program.
package display

import (
	"runtime"

	"github.com/cli/safeexec"
)

// Default is the path to the default application to display code.
func Default() ([]string, error) {
	name := "less"
	var args []string
	if runtime.GOOS == "windows" {
		name = "more"
		args = []string{"/C"}
	}

	path, err := safeexec.LookPath(name)
	if err != nil {
		return nil, err
	}
	cmd := []string{path}
	cmd = append(cmd, args...)
	return cmd, nil
}
