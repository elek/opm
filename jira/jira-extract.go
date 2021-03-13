package jira

import (
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/util"
	"github.com/elek/opm/writer"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"path"
	"strconv"
)

func init() {
	cmd := cli.Command{
		Name: "issue",
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c)
			if err != nil {
				return err
			}
			dest, err := runner.DestDir(c)
			if err != nil {
				return err
			}
			return jiraExtract(store, dest, c.String("format"))
		},
	}
	RegisterJiraExtract(cmd)

}

func jiraExtract(store kv.KV, dir string, format string) error {

	store, err := kv.Create(dir)
	if err != nil {
		return err
	}

	output, err := writer.NewWriter("jira-issue", format, new(JiraIssue))
	if err != nil {
		return err
	}
	defer output.Close()

	commentOutput, err := writer.NewWriter("jira-comment", format, new(JiraComment))
	if err != nil {
		return err
	}
	defer commentOutput.Close()

	contributionOutput, err := writer.NewWriter("jira-contribution", format, new(JiraComment))
	if err != nil {
		return err
	}
	defer contributionOutput.Close()

	projects, err := store.List(path.Join("jira", "issues"))
	if err != nil {
		return err
	}
	for _, project := range projects {
		log.Info().Msgf("Processing project %s", project)
		err := exportProject(store, output, commentOutput, contributionOutput, path.Base(project))
		if err != nil {
			return err
		}

	}
	return output.Close()
}

type JiraIssue struct {
	Project         string        `parquet:"name=project, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Key             string        `parquet:"name=key, type=UTF8, encoding=PLAIN"`
	Id              int64         `parquet:"name=id, type=INT64"`
	Summary         string        `parquet:"name=repo, type=UTF8, encoding=PLAIN"`
	Creator         string        `parquet:"name=creator, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Reporter        *string       `parquet:"name=reporter, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Assignee        *string       `parquet:"name=assignee, type=UTF8, encoding=PLAIN_DICTIONARY"`
	AssigneeDisplay *string       `parquet:"name=assigneeDisplay, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Created         int64         `parquet:"name=created, type=TIMESTAMP_MILLIS"`
	Updated         int64         `parquet:"name=updated, type=TIMESTAMP_MILLIS"`
	Resolved        *int64        `parquet:"name=resolved, type=TIMESTAMP_MILLIS"`
	Resolution      *string       `parquet:"name=status, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Status          string        `parquet:"name=resolution, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Votes           int32         `parquet:"name=votes, type=INT32"`
	Watches         int32         `parquet:"name=watches, type=INT32"`
	Priority        string        `parquet:"name=priority, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Parent          *string       `parquet:"name=parent, type=UTF8, encoding=PLAIN"`
}

type JiraComment struct {
	Project       string `parquet:"name=project, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Key           string `parquet:"name=key, type=UTF8, encoding=PLAIN"`
	Id            string `parquet:"name=id, type=UTF8, encoding=PLAIN"`
	AuthorDisplay string `parquet:"name=authorDisplay, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Author        string `parquet:"name=author, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Created       int64  `parquet:"name=created, type=TIMESTAMP_MILLIS"`
}

type JiraContribution struct {
	Project       string
	Type          string
	Identifier    string
	SubIdentifier string
	Date          int64
	Author        string
	Owner         string
}

func exportProject(store kv.KV, issueWriter writer.Writer, commentWriter writer.Writer, contributionWriter writer.Writer, project string) error {
	repoKey := path.Join("jira", "issues", project)
	progress := util.CreateProgress()
	prs, err := store.List(repoKey)
	if err != nil {
		return err
	}
	timeFormat := "2006-01-02T15:04:05.999-0700"
	for _, pr := range prs {
		js, err := json.AsJson(store.Get(pr))
		if err != nil {
			return err
		}
		n, _ := strconv.Atoi(json.MS(js, "number"))
		issue := JiraIssue{
			Project:         project,
			Key:             json.MS(js, "key"),
			Id:              int64(n),
			Summary:         json.MS(js, "fields", "summary"),
			Creator:         json.MS(js, "fields", "creator", "key"),
			Reporter:        json.MSP(js, "fields", "reporter", "key"),
			Assignee:        json.MSP(js, "fields", "assignee", "key"),
			AssigneeDisplay: json.MSP(js, "fields", "assignee", "displayName"),
			Created:         json.ME(timeFormat, js, "fields", "created"),
			Updated:         json.ME(timeFormat, js, "fields", "updated"),
			Resolved:        json.MEP(timeFormat, js, "fields", "resolutiondate"),
			Resolution:      json.MSP(js, "fields", "resolution", "name"),
			Status:          json.MS(js, "fields", "status", "name"),
			Votes:           json.MN32(js, "fields", "votes", "votes"),
			Watches:         json.MN32(js, "fields", "watches", "watchCount"),
			Priority:        json.MS(js, "fields", "priority", "name"),
			Parent:          json.MSP(js, "fields", "parent", "key"),
		}

		err = issueWriter.Write(issue)
		if err != nil {
			return err
		}

		err = contributionWriter.Write(JiraContribution{
			Project:       project,
			Type:          "JIRA_CREATED",
			Identifier:    json.MS(js, "key"),
			SubIdentifier: "",
			Date:          json.ME(timeFormat, js, "fields", "created"),
			Author:        json.MS(js, "fields", "creator", "key"),
			Owner:         json.MS(js, "fields", "creator", "key"),
		})
		if err != nil {
			return err
		}

		for _, comment := range json.L(json.M(js, "fields", "comment", "comments")) {

			err = commentWriter.Write(JiraComment{
				project,
				json.MS(js, "key"),
				json.MS(comment, "id"),
				json.MS(comment, "author", "displayName"),
				json.MS(comment, "author", "key"),
				json.ME(timeFormat, comment, "created"),
			})
			if err != nil {
				return err
			}

			err = contributionWriter.Write(JiraContribution{
				Project:       project,
				Type:          "JIRA_COMMENT",
				Identifier:    json.MS(js, "key"),
				SubIdentifier: json.MS(comment, "id"),
				Date:          json.ME(timeFormat, comment, "created"),
				Author:        json.MS(comment, "author", "key"),
				Owner:         json.MS(js, "fields", "creator", "key"),
			})
			if err != nil {
				return err
			}
		}

		progress.Increment()
	}
	return nil
}
