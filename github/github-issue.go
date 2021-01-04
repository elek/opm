package github

import (
	gojson "encoding/json"
	gh "github.com/elek/go-utils/github"
	"github.com/elek/go-utils/incremental"
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"
)

func init() {
	cmd := cli.Command{
		Name: "issue",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "org",
				Value: "apache",
				Usage: "Download github issues",
			},
		},
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c);
			if err != nil {
				return err
			}
			return issueFetchOrg(store, c.String("org"))
		},
	}
	RegisterGithubUpdate(cmd)
}

func issueFetchOrg(store kv.KV, org string) error {
	list, err := store.List(RepoDir(org))
	if err != nil {
		return err
	}
	for _, repo := range list {
		repoName := path.Base(repo)
		err = issueFetchRepo(store, org, repoName)
		if err != nil {
			return err
		}
	}
	return nil
}

func issueFetchRepo(store kv.KV, org string, repo string) error {
	query, err := pkger.Open("/graphql/issues.graphql")
	if err != nil {
		return errors.Wrap(err, "embedded issue.graphql couldn't be loaded")
	}
	defer query.Close()
	originalGraphql, err := ioutil.ReadAll(query)
	if err != nil {
		return err
	}

	inc := incremental.Incremental{
		Store: store,
		Key:   path.Join(IssueData(org, repo,  "issue-last")),
	}

	download := func(lastUpdated time.Time) (time.Time, error) {
		processed := lastUpdated
		firstGraphql := strings.Replace(string(originalGraphql), "hadoop-ozone", repo, -1)
		firstGraphql = strings.Replace(firstGraphql, "apache", org, -1)
		graphql := firstGraphql
		hasNextPage := true
		page := 0
	outer:
		for hasNextPage {
			response, err := json.AsJson(gh.ReadGithubApiV4Query([]byte(graphql)))
			if err != nil {
				return processed, err
			}
			if json.M(response, "errors") != nil {
				firstError := json.L(json.M(response, "errors"))[0]
				return processed, errors.New(json.MS(firstError, "message"))
			}
			if json.MN(response,"data","repository","issues","totalCount") == 0 {
				return processed,nil
			}
			for _, issueedge := range json.L(json.M(response, "data", "repository", "issues", "edges")) {
				issue := json.M(issueedge, "node")
				number := json.MN(issue, "number")
				err = persistIssue(store, org, repo, number, issue)
				if err != nil {
					return processed, err
				}

				updatedAt, err := time.Parse(time.RFC3339, json.MS(issue, "updatedAt"))
				if err != nil {
					return processed, err
				}

				if !updatedAt.After(lastUpdated) {
					log.Info().Msgf("All the older issues are already downloaded for %s/%s", org, repo)
					break outer
				}

				if processed == lastUpdated {
					processed = updatedAt
				}

				page++

			}
			log.Info().Msgf("All the messages are persisted from this batch (%s/%s), saved issues: %d", org, repo, page)
			//hasNextPage = false
			hasNextPage = json.M(response, "data", "repository", "issues", "pageInfo", "hasNextPage").(bool)
			if hasNextPage {
				cursor := json.MS(response, "data", "repository", "issues", "pageInfo", "endCursor")
				graphql = strings.Replace(firstGraphql, "issues(", "issues(after:\""+cursor+"\",", 1)
				time.Sleep(200 * time.Millisecond)
			}
		}
		return processed, nil
	}
	_,err = inc.Update(download)
	return err
}

func persistIssue(kv kv.KV, org string, repo string, prnum int, pr interface{}) error {
	json, err := gojson.MarshalIndent(pr, "", "   ")
	if err != nil {
		return err
	}
	err = kv.Put(IssueFile( org, repo, strconv.Itoa(prnum)), json)
	if err != nil {
		return err
	}
	return nil
}
