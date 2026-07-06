// Package cmd provides the CLI.
package cmd

import (
	"fmt"
	"io"
	"os"
)

// Execute executes the CLI.
func Execute() {
	quitIfErr(os.Stderr, rootCmd.Execute())
}

// quitIfErr takes an error value and quits if it is not nil.
func quitIfErr(w io.Writer, err error) {
	if err != nil {
		quitWithMessage(w, err)
	}
}

// quitWithMessage quits with a message printed to stderr, then exits.
func quitWithMessage(w io.Writer, msg any) {
	fmt.Fprintln(w, msg)
	os.Exit(1)
}
