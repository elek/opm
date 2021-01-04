package github

import (
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/writer"
	"github.com/urfave/cli/v2"
	"path"
	"strings"
	"time"
)

func init() {
	cmd := cli.Command{
		Name: "contribution",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "filter",
				Value: "",
				Usage: "filter to match any repo in the current store in org/repo format",
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
			return extractContribution(store, dest, c.String("org"), c.String("format"))
		},
	}
	RegisterGithubExtract(cmd)
}

type Contribution struct {
	Org           string
	Repo          string
	Type          string
	Identifier    string
	SubIdentifier string
	Date          int64
	Own           bool
	Author        string
}

func extractContribution(store kv.KV, dest string, filter string, format string) error {
	output, err := writer.NewWriter("github-contribution", format, new(Contribution))
	if err != nil {
		return err
	}


	orgList, err := store.List(RepoDir(""))
	if err != nil {
		return err
	}

	for _, orgKey := range orgList {
		org := strings.Split(path.Base(orgKey),".")[0]
		repoList, err := store.List(RepoDir(org))
		if err != nil {
			return err
		}

		for _, repo := range repoList {
			repoName := path.Base(repo)

			prList, err := store.List(PrDir(org, repoName))
			if err != nil {
				return err
			}
			for _, prName := range prList {
				pr, err := json.AsJson(store.Get(prName))
				if err != nil {
					return err
				}

				output.Write(Contribution{
					Org:        org,
					Repo:       repoName,
					Type:       "PR_CREATED",
					Date:       json.MT(time.RFC3339, pr, "createdAt"),
					Own:        true,
					Identifier: json.MNS(pr, "number"),
					Author:     json.MS(pr, "author", "login"),
				})

				for _, comment := range json.L(json.M(pr, "comments", "nodes")) {

					output.Write(Contribution{
						Org:        org,
						Repo:       repoName,
						Type:       "PR_COMMENTED",
						Date:       json.MT(time.RFC3339, comment, "createdAt"),
						Own:        json.MS(pr, "author", "login") == json.MS(comment, "author", "login"),
						Identifier: json.MNS(pr, "number"),
						Author:     json.MS(comment, "author", "login"),
					})
				}

				for _, review := range json.L(json.M(pr, "reviews", "nodes")) {

					output.Write(Contribution{
						Org:        org,
						Repo:       repoName,
						Type:       "PR_REVIEW_" + json.MS(review, "state"),
						Date:       json.MT(time.RFC3339, review, "updatedAt"),
						Own:        json.MS(pr, "author", "login") == json.MS(review, "author", "login"),
						Identifier: json.MNS(pr, "number"),
						Author:     json.MS(review, "author", "login"),
					})
				}

				if json.MB(pr, "merged") {

					output.Write(Contribution{
						Org:        org,
						Repo:       repoName,
						Type:       "PR_MERGED",
						Date:       json.MT(time.RFC3339, pr, "mergedAt"),
						Own:        json.MS(pr, "author", "login") == json.MS(pr, "mergedBy", "login"),
						Identifier: json.MNS(pr, "number"),
						Author:     json.MS(pr, "mergedBy", "login"),
					})
				}
			}

		}
	}
	return output.Close()

}
