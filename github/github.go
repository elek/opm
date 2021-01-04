package github

import (
	"github.com/elek/opm/runner"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var GithubUpdateCommands []*cli.Command
var GithubExtractCommands []*cli.Command

func RegisterGithubUpdate(command cli.Command) {
	if GithubUpdateCommands == nil {
		GithubUpdateCommands = make([]*cli.Command,0)
	}
	GithubUpdateCommands = append(GithubUpdateCommands, &command)
}


func RegisterGithubExtract(command cli.Command) {
	if GithubExtractCommands == nil {
		GithubExtractCommands = make([]*cli.Command,0)
	}
	GithubExtractCommands = append(GithubExtractCommands, &command)
}


func CreateGithubCommand() *cli.Command {
	GithubExtractCommands := append(GithubExtractCommands, &cli.Command{
		Name:"all",
		Description: "Execute all the other extract subcommands",
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c);
			if err != nil {
				return err
			}
			dest,err := runner.DestDir(c);
			if err != nil {
				return err
			}
			log.Info().Msg("Extracting pull requests")
			err = githubPrExtract(store,dest, c.String("org"), c.String("format"))
			if err != nil {
				return err
			}
			err = githubPrExtract(store,dest, c.String("org"), c.String("format"))
			if err != nil {
				return err
			}
			return nil
		},
	})
	return &cli.Command{
		Name:        "github",
		Description: "Retrieve and extract information fromGitHub",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "org",
				Value: "apache",
				Usage: "Github organization",
			},
		},
		Subcommands: []*cli.Command{
			&cli.Command{
				Name:        "update",
				Subcommands: GithubUpdateCommands,
			},
			&cli.Command{
				Name:        "extract",
				Subcommands: GithubExtractCommands,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Value: "csv",
						Usage: "Format of the extract ('csv' or 'parquet')",
					},
				},
			},
		}}
}
