package github

import (
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/writer"
	"github.com/urfave/cli/v2"
	"path"
)

func init() {
	cmd := cli.Command{
		Name: "repo",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "org",
				Value: "apache",
				Usage: "Github organization",
			},
		},
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c)
			if err != nil {
				return err
			}
			dest, err := runner.DestDir(c)
			if err != nil {
				return err
			}
			return githubRepoExtract(store, dest, c.String("org"),c.String("format"))
		},
	}
	RegisterGithubExtract(cmd)

}

type Repo struct {
	Name            string
	WatcherCount    int
	StargazersCount int
	Size            int
	OpenIssuesCount int
	ForksCount      int
}

func githubRepoExtract(store kv.KV, dir string, org string, format string) error {
	repos, err := store.List(RepoDir(org))
	if err != nil {
		return err
	}

	repoWriter, err := writer.NewWriter("github-repo", format, new(Repo))
	if err != nil {
		return err
	}

	for _, repo := range repos {
		js, err := json.AsJson(store.Get(RepoFile(org, path.Base(repo))))
		if err != nil {
			return err
		}
		repoWriter.Write(Repo{
			json.MS(js, "name"),
			json.MN(js, "watchers_count"),
			json.MN(js, "stargazers_count"),
			json.MN(js, "size"),
			json.MN(js, "open_issues_count"),
			json.MN(js, "forks_count"),
		})

	}
	err = repoWriter.Close()
	if err != nil {
		return err
	}
	return nil
}
