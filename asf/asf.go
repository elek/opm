package asf

import (
	"encoding/json"
	js "github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/writer"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
)

func CreateAsfCommand() *cli.Command {
	return &cli.Command{
		Name:        "asf",
		Description: "Retrieve and extract committer/PMC information from public ASF sources",
		Subcommands: []*cli.Command{
			{
				Name: "update",
				Action: func(c *cli.Context) error {
					store, err := runner.CreateRepo(c)
					if err != nil {
						return err
					}
					return asfUpdate(store)
				},
			},
			{
				Name: "extract",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Value: "csv",
						Usage: "Format of the extract ('csv' or 'parquet')",
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
					return asfExtract(store, dest)
				},
			},
		}}
}

type GithubToApache struct {
	GithubName string
	ApacheName string
}

func asfExtract(store kv.KV, destDir string) error {
	err := extractGithubToApache(store, destDir)
	if err != nil {
		return err
	}
	err = extractMembership(store, destDir)
	if err != nil {
		return err
	}
	return nil
}

type ApacheMembership struct {
	ApacheName string
	Project    string
	Role       string
}

func extractMembership(store kv.KV, destDir string) error {
	dest, err := writer.NewWriter(path.Join(destDir, "asf-membership"), "csv", new(GithubToApache))
	if err != nil {
		return err
	}
	defer dest.Close()

	keys, err := store.List(path.Join("asf", "project"))
	if err != nil {
		return err
	}
	for _, key := range keys {
		projectInfo, err := js.AsJson(store.Get(key))
		if err != nil {
			return err
		}

		projectName := path.Base(key)
		exported := make(map[string]bool)
		for _, name := range js.L(js.M(projectInfo, "owners")) {
			err = dest.Write(ApacheMembership{
				ApacheName: name.(string),
				Project:    projectName,
				Role:       "pmc",
			})
			exported[name.(string)] = true
			if err != nil {
				return err
			}

		}

		for _, name := range js.L(js.M(projectInfo, "members")) {
			if _, found := exported[name.(string)]; !found {
				err = dest.Write(ApacheMembership{
					ApacheName: name.(string),
					Project:    projectName,
					Role:       "committer",
				})
				if err != nil {
					return err
				}
			}

		}
	}
	return nil
}

func extractGithubToApache(store kv.KV, destDir string) error {
	dest, err := writer.NewWriter(path.Join(destDir, "asf-github-to-apache"), "csv", new(GithubToApache))
	if err != nil {
		return err
	}
	defer dest.Close()

	committers, err := store.List(path.Join("asf", "committers"))
	if err != nil {
		return err
	}
	for _, committer := range committers {
		committerInfo, err := js.AsJson(store.Get(committer))
		if err != nil {
			return err
		}
		for _, githubUser := range js.L(js.M(committerInfo, "githubUsername")) {
			err = dest.Write(GithubToApache{
				GithubName: githubUser.(string),
				ApacheName: js.MS(committerInfo, "id"),
			})
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func asfUpdate(store kv.KV) error {
	err := downloadGithubUsers(store)
	if err != nil {
		return err
	}

	err = downloadProjects(store)
	if err != nil {
		return err
	}

	err = downloadCommittees(store)
	if err != nil {
		return err
	}
	return nil
}

func downloadGithubUsers(store kv.KV) error {
	user := os.Getenv("ASF_USER")
	password := os.Getenv("ASF_PASSWORD")
	if password == "" || user == "" {
		return errors.New("ASF_USER and ASF_PASSWORD env variables are required")
	}
	req, err := http.NewRequest("GET", "https://whimsy.apache.org/roster/committer/index.json", nil)
	req.SetBasicAuth(user, password)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return errors.New("HTTP error code  " + strconv.Itoa(resp.StatusCode))
	}

	response, err := js.AsJsonList(ioutil.ReadAll(resp.Body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	for _, content := range js.L(response) {

		data, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			return err
		}
		store.Put(path.Join("asf", "committers", js.MS(content, "id")), data)
	}
	return nil

}

func downloadProjects(store kv.KV) error {

	req, err := http.NewRequest("GET", "https://people.apache.org/public/public_ldap_projects.json", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return errors.New("HTTP error code  " + strconv.Itoa(resp.StatusCode))
	}

	response, err := js.AsJson(ioutil.ReadAll(resp.Body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	for name, content := range js.M(response, "projects").(map[string]interface{}) {
		proj, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			return err
		}
		store.Put(path.Join("asf", "project", name), proj)
	}
	return nil

}

func downloadCommittees(store kv.KV) error {

	req, err := http.NewRequest("GET", "https://people.apache.org/public/committee-info.json", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return errors.New("HTTP error code  " + strconv.Itoa(resp.StatusCode))
	}

	response, err := js.AsJson(ioutil.ReadAll(resp.Body))
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	for name, content := range js.M(response, "committees").(map[string]interface{}) {
		proj, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			return err
		}
		store.Put(path.Join("asf", "committee", name), proj)
	}
	return nil

}
