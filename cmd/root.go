package cmd

import (
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spenserblack/gh-browse-tree/internal/browser"
	"github.com/spenserblack/gh-browse-tree/internal/display"
	"github.com/spf13/cobra"
)

var repo string

var rootCmd = &cobra.Command{
	Use:   "gh-browse-tree",
	Short: "Browse the file tree remotely via the API",
	RunE: func(cmd *cobra.Command, args []string) error {
		stdout := cmd.OutOrStdout()
		stderr := cmd.ErrOrStderr()
		stdin := cmd.InOrStdin()

		var (
			r   repository.Repository
			err error
		)
		if repo == "" {
			r, err = repository.Current()
		} else {
			r, err = repository.Parse(repo)
		}
		if err != nil {
			return err
		}

		display, err := display.Default()
		quitIfErr(stderr, err)

		client, err := api.NewRESTClient(api.ClientOptions{
			Host: r.Host,
		})
		quitIfErr(stderr, err)
		quitIfErr(stderr, browser.Start(r, client, stdin, stdout, stderr, display))

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&repo, "repo", "R", "", "Select another repository using the [HOST/]OWNER/REPO format")
}
