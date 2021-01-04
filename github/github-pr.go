package github

import (
	gojson "encoding/json"
	"errors"
	gh "github.com/elek/go-utils/github"
	"github.com/elek/go-utils/incremental"
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/markbates/pkger"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"
)

func init() {
	command:= cli.Command{
		Name: "pr",
		Flags: []cli.Flag{
		},
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c)
			if err != nil {
				return err
			}
			return fetch(store)
		},
	}
	RegisterGithubUpdate(command)

}



func fetch(store kv.KV) error {
	list, err := store.List(RepoDir(""))
	if err != nil {
		return err
	}
	for _, org := range list {
		orgName := strings.Split(path.Base(org),".")[0]
		err = orgFetch(store, orgName)
		if err != nil {
			return err
		}
	}
	return nil
}

func orgFetch(store kv.KV, org string) error {
	list, err := store.List(RepoDir(org))
	if err != nil {
		return err
	}
	for _, repo := range list {
		repoName := path.Base(repo)
		err = githubFetch(store, org, repoName)
		if err != nil {
			return err
		}
	}
	return nil
}

func githubFetch(store kv.KV, org string, repo string) error {
	query, err := pkger.Open("/graphql/prs.graphql")
	if err != nil {
		return err
	}
	defer query.Close()
	originalGraphql, err := ioutil.ReadAll(query)
	if err != nil {
		return err
	}

	inc := incremental.Incremental{
		Store: store,
		Key:   path.Join(PrData(org, repo, "pr-last")),
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
			for _, predge := range json.L(json.M(response, "data", "repository", "pullRequests", "edges")) {
				pr := json.M(predge, "node")
				number := json.MN(pr, "number")
				err = persist(store, org, repo, number, pr)
				if err != nil {
					return processed, err
				}

				updatedAt, err := time.Parse(time.RFC3339, json.MS(pr, "updatedAt"))
				if err != nil {
					return processed, err
				}

				if !updatedAt.After(lastUpdated) {
					log.Info().Msgf("All the older PRs are already downloaded for %s/%s", org, repo)
					break outer
				}

				if processed == lastUpdated {
					processed = updatedAt
				}

				page++

			}
			log.Info().Msgf("All the messages are persisted from this batch (%s/%s), saved prs: %d", org, repo, page)
			//hasNextPage = false
			hasNextPage = json.M(response, "data", "repository", "pullRequests", "pageInfo", "hasNextPage").(bool)
			if hasNextPage {
				cursor := json.MS(response, "data", "repository", "pullRequests", "pageInfo", "endCursor")
				graphql = strings.Replace(firstGraphql, "pullRequests(", "pullRequests(after:\""+cursor+"\",", 1)
				time.Sleep(1 * time.Second)
			}
		}
		return processed, nil
	}
	_, err = inc.Update(download)
	return err
}

func persist(kv kv.KV, org string, repo string, prnum int, pr interface{}) error {
	json, err := gojson.MarshalIndent(pr, "", "   ")
	if err != nil {
		return err
	}
	err = kv.Put(PrFile(org, repo, strconv.Itoa(prnum)), json)
	if err != nil {
		return err
	}
	return nil
}
