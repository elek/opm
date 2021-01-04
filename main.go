package main

import (
	"fmt"
	"github.com/elek/opm/github"
	_ "github.com/elek/opm/github"
	"github.com/elek/opm/jira"
	_ "github.com/elek/opm/jira"
	"github.com/elek/opm/ponymail"
	_ "github.com/elek/opm/ponymail"
	"github.com/elek/opm/runner"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	runner.App.Commands = append(runner.App.Commands, github.CreateGithubCommand())
	runner.App.Commands = append(runner.App.Commands, jira.CreateJiraCommand())
	runner.App.Commands = append(runner.App.Commands, ponymail.CreatePonymailCommand())
	err := runner.App.Run(os.Args)
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(-1)
	}
}
