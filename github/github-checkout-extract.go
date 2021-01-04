package github

import (
	"bufio"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/writer"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
	"path"
	"strings"
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
			return githubCheckoutExtract(store, dest)
		},
	}
	RegisterGithubExtract(cmd)
}

type GitCommit struct {
	Org string
	Repo string
	Hash string
	AuthorTime string
	AuthorEmail string
	AuthorName string
	CommitterTime string
	CommitterEmail string
	CommitterName string

}
func githubCheckoutExtract(store kv.KV, dest string) error {

	output, err := writer.NewWriter("github-commit","csv",new(GitCommit))
	if err != nil {
		return err
	}

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
			err := gitLogExport(repoDir, org, repoName, output)
			if err != nil {
				return err
			}

		}

	}
	return output.Close()

}

func gitLogExport(coDir string, org string, repoName string, output writer.Writer) error {
	sep := "_|*|_"
	fields := []string{"%H", "%at", "%ae", "%an", "%ct", "%ce", "%cn"}
	filter := strings.Join(fields, sep)
	gitLog := exec.Cmd{
		Path: "/usr/bin/git",
		Args: []string{
			"git",
			"-c",
			"pager.log=false",
			"log",
			"--all",
			"--pretty=format:" + filter,
		},
		Dir:    coDir,
		Stderr: os.Stderr,
	}

	stdout, err := gitLog.StdoutPipe()
	if err != nil {
		return err
	}

	if err := gitLog.Start(); err != nil {
		return err
	}

	reader := bufio.NewReader(stdout)
	for {
		r, _, _ := reader.ReadLine()
		if r == nil {
			break
		}
		parts := strings.Split(string(r), sep)
		err = output.Write(GitCommit{
			Org: org,
			Repo: repoName,
			Hash: parts[0],
			AuthorTime: parts[1],
			AuthorEmail: parts[2],
			AuthorName: parts[3],
			CommitterTime: parts[4],
			CommitterEmail: parts[5],
			CommitterName: parts[6],
		})
		if err != nil {
			return err
		}
	}
	err = gitLog.Wait()
	if err != nil {
		return err
	}

	return nil
}
