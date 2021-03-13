package github

import (
	"encoding/json"
	gh "github.com/elek/go-utils/github"
	js "github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/urfave/cli/v2"
	"path"
)

func init() {
	command := cli.Command{
		Name: "action-workflow",
		Flags: []cli.Flag{
		},
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c)
			if err != nil {
				return err
			}
			return fetchActionUsage(store)
		},
	}
	RegisterGithubUpdate(command)

}

func fetchActionUsage(store kv.KV) error {
	orgs, err := store.List(RepoDir(""))
	if err != nil {
		return err
	}
	for _, org := range orgs {

		repos, err := store.List(org)
		if err != nil {
			return err
		}
		for _, repo := range repos {
			err = fetchRepoActionUsage(store, path.Base(org), path.Base(repo))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func fetchRepoActionUsage(store kv.KV, org string, repo string) error {
	markerPath := path.Join("github", "actions", org, repo, "workflows-downloaded")
	if store.Contains(markerPath) {
		//don't re-download project if already downloaded
		return nil
	}
	url := "https://api.github.com/repos/" + org + "/" + repo + "/actions/runs?per_page=100"
	written := 0
	err := gh.ReadAllGithubApiV3(url, func(data []byte, err error) error {
		result, err := js.AsJson(data, err)
		if err != nil {
			return err
		}
		for _, run := range js.L(js.M(result, "workflow_runs")) {

			runJson, err := json.MarshalIndent(run, "", "   ")
			if err != nil {
				return err
			}
			workflowId := js.MNS(run, "workflow_id")
			runNumber := js.MNS(run, "run_number")
			err = store.Put(path.Join("github", "actions", org, repo, "workflowruns", workflowId, runNumber), runJson)
			if err != nil {
				return err
			}
			written++
		}
		return nil
	})
	if err != nil {
		return err
	}
	if written > 0 {
		err = store.Put(markerPath, []byte("done"))
		if err != nil {
			return err
		}
	}
	return nil
}
