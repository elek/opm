package github

import (
	csv2 "encoding/csv"
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"os"
	"path"
)

func init() {
	//runner.App.Commands = append(runner.App.Commands, cli.Command{
	//	Name: "github-team-extract",
	//	Flags: []cli.Flag{
	//		cli.StringFlag{
	//			Name:  "org",
	//			Value: "apache",
	//			Usage: "Github organization",
	//		},
	//		cli.StringFlag{
	//			Name:     "team",
	//			Value:    "",
	//			Required: true,
	//			Usage:    "Name of the team to download",
	//		},
	//	},
	//	Action: func(c *cli.Context) error {
	//		store, err := kv.Create(c.Args().Get(0))
	//		if err != nil {
	//			return err
	//		}
	//		return extractTeam(store, c.String("org"), c.String("team"))
	//	},
	//})

}

func extractTeam(store kv.KV, org string, team string) error {
	reposFile, err := os.Create("github-" + org + "-" + team + ".csv")
	if err != nil {
		return err
	}

	defer reposFile.Close()
	csv := csv2.NewWriter(reposFile)
	csv.Write([]string{
		"login", "name", "company", "email", "location", "webSite"	,
	})
	logins, err := store.List(TeamDir(org, team))
	if err != nil {
		return err
	}
	for _, login := range logins {
		user, err := json.AsJson(store.Get(TeamUserFile(org, team, path.Base(login))))
		if err != nil {
			return err
		}

		csv.Write([]string{
			json.MS(user, "login"),
			json.MS(user, "name"),
			json.MS(user, "company"),
			json.MS(user, "email"),
			json.MS(user, "location"),
			json.MS(user, "webSite"),
		})
	}
	csv.Flush()
	return err
}
