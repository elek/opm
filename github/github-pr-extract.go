package github

import (
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/writer"
	"github.com/urfave/cli/v2"
	"path"
	"time"
)

func init() {
	cmd:= cli.Command{
		Name: "pr",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "org",
				Value: "apache",
				Usage: "Github organization",
			},
		},
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c);
			if err != nil {
				return err
			}
			dest,err := runner.DestDir(c);
			if err != nil {
				return err
			}
			return githubPrExtract(store,dest, c.String("org"), c.String("format"))
		},
	}
	RegisterGithubExtract(cmd)
}

type GithubPr struct {
	Repo           string  `parquet:"name=repo, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Number         int32   `parquet:"name=number, type=INT32"`
	CreatedAt      int64   `parquet:"name=createdAt, type=TIMESTAMP_MILLIS"`
	UpdatedAt      *int64  `parquet:"name=updatedAt, type=TIMESTAMP_MILLIS"`
	ClosedAt       *int64  `parquet:"name=closedAt, type=TIMESTAMP_MILLIS"`
	Merged         bool    `parquet:"name=merged, type=BOOLEAN"`
	Closed         bool    `parquet:"name=closed, type=BOOLEAN"`
	IsDraft        bool    `parquet:"name=isDraft, type=BOOLEAN"`
	Author         string  `parquet:"name=author, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Reviews        int32   `parquet:"name=reviews, type=INT32"`
	Comments       int32   `parquet:"name=comments, type=INT32"`
	MergedBy       string  `parquet:"name=mergedBy, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Title          string  `parquet:"name=title, type=UTF8, encoding=PLAIN"`
	FirstResponse  *int64  `parquet:"name=firtsResponse, type=UTF8, encoding=PLAIN"`
	FirstResponder *string `parquet:"name=firstResponder, type=UTF8, encoding=PLAIN_DICTIONARY"`
}

type GithubPrComment struct {
	Org        string `parquet:"name=org, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Repo       string `parquet:"name=repo, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Number     int32  `parquet:"name=number, type=INT32"`
	CreatedAt  int64  `parquet:"name=createdAt, type=TIMESTAMP_MILLIS"`
	Author     string `parquet:"name=author, type=UTF8, encoding=PLAIN_DICTIONARY"`
	AuthorRole string `parquet:"name=role, type=UTF8, encoding=PLAIN_DICTIONARY"`
}

func githubPrExtract(store kv.KV, dir string, org string, format string) error {

	repos, err := store.List(RepoDir(org))
	if err != nil {
		return err
	}

	dest, err := writer.NewWriter("github-pr",format, new(GithubPr))
	if err != nil {
		return err
	}

	commentsDest, err := writer.NewWriter("github-pr-comment",format, new(GithubPrComment))
	if err != nil {
		return err
	}

	for _, repo := range repos {
		println(repo)
		prs, err := store.List(PrDir(org, path.Base(repo)))
		if err != nil {
			continue
			//return err
		}

		for _, pr := range prs {
			js, err := json.AsJson(store.Get(pr))
			if err != nil {
				return err
			}
			createdAt, err := time.Parse(time.RFC3339, json.MS(js, "createdAt"))
			if err != nil {
				return err
			}
			prNumber := json.MN32(js, "number")
			pr := GithubPr{
				Repo:      path.Base(repo),
				Number:    prNumber,
				CreatedAt: createdAt.Unix() * 1000,
				Merged:    json.MB(js, "merged"),
				Closed:    json.MB(js, "closed"),
				IsDraft:   json.MB(js, "isDraft"),
				Author:    json.MS(js, "author", "login"),
				Reviews:   json.MN32(js, "reviews", "totalCount"),
				Comments:  json.MN32(js, "comments", "totalCount"),
				MergedBy:  json.MS(js, "mergedBy", "login"),
				Title:     json.MS(js, "title"),
			}

			if json.MS(js, "closedAt") != "" {
				closedAt, err := time.Parse(time.RFC3339, json.MS(js, "closedAt"))
				if err != nil {
					return err
				}
				val := closedAt.Unix() * 1000
				pr.ClosedAt = &val
			}

			if json.MS(js, "updatedAt") != "" {
				updatedAt, err := time.Parse(time.RFC3339, json.MS(js, "updatedAt"))
				if err != nil {
					return err
				}
				val := updatedAt.Unix() * 1000
				pr.UpdatedAt = &val
			}

			first := int64(0)
			firstResponder := ""

			comments := json.M(js, "comments", "nodes")
			for _, comment := range json.L(comments) {

				commentCreated := json.MT(time.RFC3339, comment, "createdAt")
				actor := json.MS(comment, "author", "login")
				if (first == 0 || first > commentCreated) && (actor != "hadoop-yetus") {
					first = commentCreated
					firstResponder = json.MS(comment, "author", "login")
				}

				err = commentsDest.Write(GithubPrComment{
					org,
					path.Base(repo),
					prNumber,
					commentCreated,
					actor,
					json.MS(comment, "authorAssociation"),
				})
				if err != nil {
					return err
				}
			}

			for _, review := range json.L(json.M(js, "reviews", "nodes")) {
				timestamp := json.MT(time.RFC3339, review, "updatedAt")
				actor := json.MS(review, "author", "login")
				if (first == 0 || first > timestamp) && (actor != "hadoop-yetus") {
					first = timestamp
					firstResponder = actor
				}
			}

			if first != 0 {
				pr.FirstResponse = &first
				pr.FirstResponder = &firstResponder
			}

			err = dest.Write(pr)
			if err != nil {
				return err
			}

		}

	}
	_ = dest.Close()
	_ = commentsDest.Close()
	return nil
}
