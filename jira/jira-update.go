package jira

import (
	gojson "encoding/json"
	"fmt"
	"github.com/elek/go-utils/incremental"
	jirautil "github.com/elek/go-utils/jira"
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"math"
	"path"
	"strings"
	"time"
)

func init() {
	cmd := cli.Command{
		Name: "issue",
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c)
			if err != nil {
				return err
			}
			return jira(store, c.String("query"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "query",
			},
		},
	}
	RegisterJiraUpdate(cmd)

}

func jira(store kv.KV, query string) error {
	jiraClient := jirautil.Jira{
		Url: "https://issues.apache.org/jira",
	}

	inc := incremental.Incremental{
		Store: store,
		Key:   path.Join("jira", "last"),
	}

	retrieved := math.MaxInt32

	updater := func(lastUpdate time.Time) (time.Time, error) {
		lastTime := lastUpdate

		getter := func() ([]byte, error) {
			searchQuery := "updated >= %d order by updated asc"
			if query != "" {
				searchQuery = query + " AND " + searchQuery
			}
			return jiraClient.ReadSearch(fmt.Sprintf(searchQuery, lastUpdate.Unix()*1000))
		}
		hddsIssues, err := json.AsJson(getter())
		if err != nil {
			return lastTime, errors.Wrap(err, "Can't execute Jira query")
		}
		retrieved = 0
		issues := json.L(json.M(hddsIssues, "issues"))
		for _, issue := range issues {
			key := json.MS(issue, "key")
			retrieved++
			content, err := gojson.MarshalIndent(issue, "", "   ")
			if err != nil {
				return lastTime, errors.Wrap(err, "Can't marshall json for key "+key)
			}
			updated := json.MS(issue, "fields", "updated")
			lastTime, err = time.Parse("2006-01-02T15:04:05.999999999-0700", updated)
			if err != nil {
				return lastTime, errors.Wrap(err, "Time couldn't parse "+updated)
			}
			parts := strings.Split(key, "-")
			store.Put(path.Join("jira", parts[0], key), content)
		}
		return lastTime, nil
	}
	next := true
	for next {
		var err error
		next, err = inc.Update(updater)
		if err != nil {
			return errors.Wrap(err, "Error on getting new issues")
		}
		log.Info().Msgf("Retrieved %d entries, up to date until %s", retrieved, OrPanic(store.Get(path.Join("jira", "last"))))
		if retrieved > 0 {
			time.Sleep(2 * time.Second)
		}
	}
	return nil
}

func OrPanic(value interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return value
}
