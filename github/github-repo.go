package github

import (
	gojson "encoding/json"
	gh "github.com/elek/go-utils/github"
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func init() {
	repo := cli.Command{
		Name: "repo",
		Description: "Download metadata of Github repository",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "org",
				Value: "apache",
				Usage: "Github organization",
			},
			&cli.StringFlag{
				Name:  "repo",
				Value: "",
				Usage: "Name of github repository",
			},
		},
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c);
			if err != nil {
				return err
			}
			repo := c.String("repo")
			if repo == "" {
				return githubRepos(store, c.String("org"))
			} else {
				return githubRepo(store, c.String("org"), repo)
			}
		},
	}
	RegisterGithubUpdate(repo)
}

func githubRepos(store kv.KV, org string) error {

	processRepoList := func(data []byte, err error) error {
		repos, err := json.AsJsonList(data, err)
		if err != nil {
			return errors.Wrap(err, "Can't retrieve github repos for org "+org)
		}
		for _, repo := range repos {
			err := persistRepo(store,org,repo)
			if err != nil {
				return err
			}
		}
		return nil
	}
	url := "https://api.github.com/orgs/" + org + "/repos?per_page=100"
	return gh.ReadAllGithubApiV3(url, processRepoList)
}

func persistRepo(store kv.KV, org string, repo interface{}) error {
	data, err := gojson.MarshalIndent(repo, "", "   ")
	if err != nil {
		return errors.Wrap(err, "Can't marshall json repo fragment")
	}
	err = store.Put(RepoFile(org, json.MS(repo, "name")), data)
	if err != nil {
		return errors.Wrap(err, "Can't save repo file")
	}
	return nil
}

func githubRepo(store kv.KV, org string, repo string) error {
	url := "https://api.github.com/repos/"+org+"/"+repo
	r, err := json.AsJson(gh.ReadGithubApiV3(url))
	if err != nil {
		return err
	}
	err = persistRepo(store,org, r)
	if err != nil {
		return err
	}

	return nil
}
