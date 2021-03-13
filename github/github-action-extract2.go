package github

import (
	"github.com/elek/go-utils"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/writer"
	"github.com/urfave/cli/v2"
	"github.com/valyala/fastjson"
	"path"
	"time"
)

func init() {
	command := cli.Command{
		Name: "action2",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "repo",
				Usage: "Repository url",
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
			return extractActionUsage2(store, dest)
		},
	}
	RegisterGithubExtract(command)

}

func extractActionUsage2(store kv.KV, dest string) error {
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

func extractRepoActionUsage2(store kv.KV, org string, repo string, w writer.Writer, p *util.Progress) error {
	workflows, err := store.List(path.Join("github", "actions", org, repo, "workflowruns"))
	if err != nil {
		return err
	}
	if len(workflows) == 0 {

	}
	var parser fastjson.Parser

	for _, workflowKey := range workflows {
		workflowRuns, err := store.List(path.Join("github", "actions", org, repo, "workflowruns", path.Base(workflowKey)))
		if err != nil {
			return err
		}
		for _, workflowRunKey := range workflowRuns {

			rawJson, err := store.Get(workflowRunKey)
			if err != nil {
				return err
			}

			workflowRun, err := parser.ParseBytes(rawJson)
			if err != nil {
				return err
			}

			if string(workflowRun.GetStringBytes("status")) != "completed" {
				continue
			}
			workflowRun.GetStringBytes()
			record := GithubWorkflowRun{
				Org:        org,
				Repo:       repo,
				Id:         workflowRun.GetInt64("id"),
				RunNumber:  workflowRun.GetInt64("run_number"),
				CreatedAt:  *epoch(workflowRun, "created_at"),
				UpdatedAt:  epoch(workflowRun, "updated_at"),
				Status:     string(workflowRun.GetStringBytes("status")),
				Conclusion: string(workflowRun.GetStringBytes("conclusion")),
				WorkflowId: 0,
			}

			jobPath := path.Join("github", "actions", org, repo, "jobs", path.Base(workflowKey), path.Base(workflowRunKey))
			if store.Contains(jobPath) {
				jobJson, err := store.Get(jobPath)
				if err != nil {
					return err
				}

				jobs, err := parser.ParseBytes(jobJson)
				if err != nil {
					return err
				}
				record.JobDetails = true
				jobItems := jobs.GetArray("jobs")
				record.Jobs = len(jobItems)

				for _, job := range jobItems {
					startedAt := epoch(job, "started_at")
					completedAt := epoch(job, "completed_at")
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

func epoch(run *fastjson.Value, key string) *int64 {
	value := run.GetStringBytes(key)
	if value == nil {
		return nil
	}

	t, err := time.Parse(time.RFC3339, string(value))
	if err != nil {
		panic(err)
	}
	res := t.Unix() * 1000
	return &res
}
