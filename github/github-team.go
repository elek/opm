package github

import (
	gojson "encoding/json"
	"errors"
	gh "github.com/elek/go-utils/github"
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/markbates/pkger"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"strings"
	"time"
)

func init() {
	repo := cli.Command{
		Name:        "repo",
		Description: "Download team data from github",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "org",
				Value: "apache",
				Usage: "Github organization",
			},
			&cli.StringFlag{
				Name:     "team",
				Value:    "",
				Required: true,
				Usage:    "Name of github team to download",
			},
		},
		Action: func(c *cli.Context) error {
			store, err := kv.Create(c.Args().Get(0))
			if err != nil {
				return err
			}

			return fetchTeam(store, c.String("org"), c.String("team"))
		},
	}
	RegisterGithubUpdate(repo)
}

func fetchTeam(store kv.KV, org string, team string) error {
	query, err := pkger.Open("/graphql/team.graphql")
	if err != nil {
		return err
	}
	defer query.Close()
	originalGraphql, err := ioutil.ReadAll(query)
	if err != nil {
		return err
	}

	firstGraphql := strings.Replace(string(originalGraphql), "apache-committers", team, -1)
	firstGraphql = strings.Replace(firstGraphql, "apache", org, -1)
	graphql := firstGraphql
	hasNextPage := true
	page := 0
	for hasNextPage {
		response, err := json.AsJson(gh.ReadGithubApiV4Query([]byte(graphql)))
		if err != nil {
			return err
		}
		if json.M(response, "errors") != nil {
			firstError := json.L(json.M(response, "errors"))[0]
			return errors.New(json.MS(firstError, "message"))
		}
		for _, edge := range json.L(json.M(response, "data", "organization", "team", "members", "edges")) {
			user := json.M(edge, "node")
			login := json.MS(user, "login")

			persistenceKey := TeamUserFile(org, team, login)

			json, err := gojson.MarshalIndent(user, "", "   ")
			if err != nil {
				return err
			}

			err = store.Put(persistenceKey, json)
			if err != nil {
				return err
			}

			page++

		}
		log.Info().Msgf("All the users are persisted from this batch (%s/%s), saved users: %d", org, team, page)
		hasNextPage = json.MB(response, "data", "organization", "team", "members", "pageInfo", "hasNextPage")
		if hasNextPage {
			cursor := json.MS(response, "data", "organization", "team", "members", "pageInfo", "endCursor")
			graphql = strings.Replace(firstGraphql, "members {", "members(after:\""+cursor+"\"){", 1)
			time.Sleep(200 * time.Millisecond)
		}
	}

	return err
}
