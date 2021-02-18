package jira

import (
	"encoding/json"
	jirautil "github.com/elek/go-utils/jira"
	js "github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/urfave/cli/v2"
	"path"
	"strings"
)

func init() {
	cmd := cli.Command{
		Name: "project",
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c)
			if err != nil {
				return err
			}
			return JiraUpdateRepo(store, c.String("filter"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "filter",
			},
		},
	}
	RegisterJiraUpdate(cmd)

}

func JiraUpdateRepo(store kv.KV, filter string) error {
	jiraClient := jirautil.Jira{
		Url: "https://issues.apache.org/jira",
	}

	projects, err := js.AsJsonList(jiraClient.ListProject())
	if err != nil {
		return err
	}
	for _, project := range projects {
		key := js.MS(project, "key")
		if filter == "" || strings.HasPrefix(key, filter) {
			rawProject, err := json.MarshalIndent(project, "", "  ")
			if err != nil {
				return err
			}
			err = store.Put(path.Join("jira", "projects", key), rawProject)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
