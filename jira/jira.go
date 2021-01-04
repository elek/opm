package jira

import (
	"github.com/urfave/cli/v2"
)

var JiraUpdateCommands []*cli.Command
var JiraExtractCommands []*cli.Command

func RegisterJiraUpdate(command cli.Command) {
	if JiraUpdateCommands == nil {
		JiraUpdateCommands = make([]*cli.Command, 0)
	}
	JiraUpdateCommands = append(JiraUpdateCommands, &command)
}

func RegisterJiraExtract(command cli.Command) {
	if JiraExtractCommands == nil {
		JiraExtractCommands = make([]*cli.Command, 0)
	}
	JiraExtractCommands = append(JiraExtractCommands, &command)
}

func CreateJiraCommand() *cli.Command {

	return &cli.Command{
		Name:        "jira",
		Description: "Retrieve and extract information from Jira",
		Flags: []cli.Flag{

		},
		Subcommands: []*cli.Command{
			{
				Name:        "update",
				Subcommands: JiraUpdateCommands,
			},
			{
				Name:        "extract",
				Subcommands: JiraExtractCommands,
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
