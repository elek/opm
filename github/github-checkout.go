package github

import (
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
	"path"
)

func init() {
	cmd := cli.Command{
		Name: "checkout",
		Action: func(c *cli.Context) error {
			store, err := runner.CreateRepo(c);
			if err != nil {
				return err
			}
			dest,err := runner.DestDir(c);
			if err != nil {
				return err
			}
			return githubClone(store, dest)
		},
	}
	RegisterGithubUpdate(cmd)
}

func githubClone(store kv.KV, dest string) error {

	orgs, err := store.List(RepoDir(""))
	if err != nil {
		return err
	}

	for _, orgPath := range orgs {
		org := path.Base(orgPath)
		repos, err := store.List(orgPath)
		if err != nil {
			return err
		}

		for _, repo := range repos {
			repoName := path.Base(repo)
			repoDir := path.Join(dest, "checkout", org, repoName)
			_ = os.MkdirAll(path.Dir(repoDir),0755)
			err = gitClone(repoDir, org, repoName)
			if err != nil {
				log.Warn().Msg(err.Error())
				continue
			}

		}

	}
	return nil
}


func gitClone(coDir string, org string, repoName string) error {
	if _, err := os.Stat(coDir); os.IsNotExist(err) {
		gitClone := exec.Cmd{
			Path: "/usr/bin/git",
			Args: []string{
				"git",
				"clone",
				"--filter=blob:none",
				"--sparse",
				"git://github.com/" + org + "/" + repoName + ".git",
			},
			Dir:    path.Dir(coDir),
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}
		err = gitClone.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
