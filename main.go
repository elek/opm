package main

import (
	"fmt"
	"github.com/elek/opm/asf"
	"github.com/elek/opm/github"
	_ "github.com/elek/opm/github"
	"github.com/elek/opm/jira"
	_ "github.com/elek/opm/jira"
	"github.com/elek/opm/ponymail"
	_ "github.com/elek/opm/ponymail"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/youtube"
	"github.com/rs/zerolog"
	"os"
	"runtime/pprof"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	if os.Getenv("OPM_PROFILE") != "" {

		f, err := os.Create(os.Getenv("OPM_PROFILE"))
		if err != nil {
			panic(err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}
	runner.App.Name = "Open-source project data downloader"
	runner.App.Description = "Utility to get base data for metrics about open source project development"
	runner.App.Commands = append(runner.App.Commands, github.CreateGithubCommand())
	runner.App.Commands = append(runner.App.Commands, jira.CreateJiraCommand())
	runner.App.Commands = append(runner.App.Commands, ponymail.CreatePonymailCommand())
	runner.App.Commands = append(runner.App.Commands, asf.CreateAsfCommand())
	runner.App.Commands = append(runner.App.Commands, youtube.CreateYoutubeCommand())
	err := runner.App.Run(os.Args)
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(-1)
	}
}
