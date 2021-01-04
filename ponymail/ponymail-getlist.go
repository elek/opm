package ponymail

import (
	gojson "encoding/json"
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/urfave/cli/v2"
	"path"
	"strings"
	"time"
)

func init() {
	cmd := cli.Command{
		Name:        "getlists",
		Description: "Download the available ponymail mailing lists",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "filter",
				Value: "",
				Usage: "Filter to restrict results (eg. hadoop.apache.org/dev or hadoop.apache.org)",
			},
		},
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c)
			if err != nil {
				return err
			}
			return maillists(store, c.String("filter"))
		},
	}
	RegisterPonymail(cmd)
}

func maillists(store kv.KV, filter string) error {
	result, err := json.AsJson(httpCall("preferences.lua"))
	if err != nil {
		return err
	}
	for domain, lists := range json.M(result, "lists").(map[string]interface{}) {
		for list, metadata := range lists.(map[string]interface{}) {

			_, err := gojson.Marshal(metadata)
			if err != nil {
				return err
			}
			name := path.Join(domain, list)
			if filter == "" || strings.HasPrefix(name, filter) {

				url := "stats.lua?list=" + list + "&domain=" + domain
				result, err := httpCall(url)
				if err != nil {
					return err
				}

				err = store.Put(path.Join("ponymail", "lists", domain, list), result)
				if err != nil {
					return err
				}
				time.Sleep(3 * time.Second)

			}
		}
	}
	return nil
}
