package github

import (
	gh "github.com/elek/go-utils/github"
	js "github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/urfave/cli/v2"
	"path"
	"strings"
	"time"
)

func init() {
	command := cli.Command{
		Name: "action-job",
		Flags: []cli.Flag{
		},
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c)
			if err != nil {
				return err
			}
			return fetchActionJobs(store)
		},
	}
	RegisterGithubUpdate(command)

}

func fetchActionJobs(store kv.KV) error {
	//now := time.Now()
	//deadLine := now.Add(-90 * 24 * time.Hour)
	location, err := time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}
	deadLine := time.Date(2021, 1, 30, 0, 0, 0, 0, location)
	return store.Iterate(path.Join("github", "actions"), func(orgKey string) error {
		return store.Iterate(orgKey, func(repoKey string) error {

			workflowRunPath := path.Join(repoKey, "workflowruns")
			if store.Contains(workflowRunPath) {
				return store.Iterate(workflowRunPath, func(workflowKey string) error {
					return store.Iterate(workflowKey, func(workflowRunKey string) error {
						jobKey := strings.Replace(workflowRunKey, "/workflowruns/", "/jobs/", 1)
						if !store.Contains(jobKey) {
							workflow, err := js.AsJson(store.Get(workflowRunKey))
							if err != nil {
								return err
							}
							if js.MS(workflow, "status") != "completed" {
								return nil
							}
							createdAt, err := time.Parse(time.RFC3339, js.MS(workflow, "created_at"))
							if err != nil {
								return err
							}
							if createdAt.After(deadLine) {
								result, err := gh.ReadGithubApiV3(js.MS(workflow, "jobs_url"))
								if err != nil {
									return err
								}
								err = store.Put(jobKey, result)
								if err != nil {
									return err
								}

								println(jobKey)
							}
						}
						return nil
					})
				})
			} else {
				if store.Contains(path.Join(repoKey, "workflows-downloaded")) {
					println("rm -rf " + repoKey)
				}
				return nil
			}
		})
	})

	return nil
}
