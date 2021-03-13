package github

import (
	js "github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/go-utils"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/writer"
	"github.com/urfave/cli/v2"
	"path"
	"time"
)

func init() {
	command := cli.Command{
		Name: "action",
		Flags: []cli.Flag{
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
			return extractActionUsage(store, dest)
		},
	}
	RegisterGithubExtract(command)

}

type GithubWorkflowRun struct {
	Org              string
	Repo             string
	Id               int64
	RunNumber        int64
	CreatedAt        int64
	UpdatedAt        *int64
	Status           string
	Conclusion       string
	WorkflowId       int
	FirstJobStarted  int64
	LastJobCompleted int64
	JobDetails       bool
	Jobs             int
	JobSeconds       int64
}

func extractActionUsage(store kv.KV, dest string) error {
	writer, err := writer.NewWriter(path.Join(dest, "github-action-job"), "csv", new(GithubWorkflowRun))
	if err != nil {
		return err
	}
	defer writer.Close()
	p := util.CreateProgress()
	err = store.Iterate(RepoDir(""), func(orgKey string) error {
		return store.Iterate(orgKey, func(repoKey string) error {
			return extractRepoActionUsage(store, path.Base(orgKey), path.Base(repoKey), writer, p)
		})
	})
	p.End()
	return err
}

func extractRepoActionUsage(store kv.KV, org string, repo string, w writer.Writer, p *util.Progress) error {
	workflows, err := store.List(path.Join("github", "actions", org, repo, "workflowruns"))
	if err != nil {
		return err
	}
	if len(workflows) == 0 {

	}
	for _, workflowKey := range workflows {
		workflowRuns, err := store.List(path.Join("github", "actions", org, repo, "workflowruns", path.Base(workflowKey)))
		if err != nil {
			return err
		}
		for _, workflowRunKey := range workflowRuns {

			workflowRun, err := js.AsJson(store.Get(workflowRunKey))
			if err != nil {
				return err
			}
			if js.MS(workflowRun, "status") != "completed" {
				continue
			}

			record := GithubWorkflowRun{
				Org:        org,
				Repo:       repo,
				Id:         js.MN64(workflowRun, "id"),
				RunNumber:  js.MN64(workflowRun, "run_number"),
				CreatedAt:  js.ME(time.RFC3339, workflowRun, "created_at"),
				UpdatedAt:  js.MEP(time.RFC3339, workflowRun, "updated_at"),
				Status:     js.MS(workflowRun, "status"),
				Conclusion: js.MS(workflowRun, "conclusion"),
				WorkflowId: js.MN(workflowRun, "workflow_id"),
			}

			jobPath := path.Join("github", "actions", org, repo, "jobs", path.Base(workflowKey), path.Base(workflowRunKey))
			if store.Contains(jobPath) {
				jobs, err := js.AsJson(store.Get(jobPath))
				if err != nil {
					return err
				}
				record.JobDetails = true
				record.Jobs = len(js.L(js.M(jobs, "jobs")))

				for _, job := range js.L(js.M(jobs, "jobs")) {
					startedAt := js.MEP(time.RFC3339, job, "started_at")
					completedAt := js.MEP(time.RFC3339, job, "completed_at")
					if startedAt != nil {
						if record.FirstJobStarted == 0 || record.FirstJobStarted > *startedAt {
							record.FirstJobStarted = *startedAt
						}
					}
					if completedAt != nil {
						if record.LastJobCompleted == 0 || record.LastJobCompleted < *completedAt {
							record.LastJobCompleted = *completedAt
						}
					}
					if startedAt != nil && completedAt != nil {
						duration := *completedAt - *startedAt
						record.JobSeconds += duration / 1000
					}
				}
			}
			err = w.Write(record)

			if err != nil {
				return err
			}
			p.Increment()
		}
	}
	return nil
}
