package github

import (
	"path"
)


func RepoDir(org string) string {
	return path.Join("github", "repos", org)
}

func RepoFile(org string, repo string) string {
	return path.Join(RepoDir(org), repo)
}

func TeamDir(org string, team string) string {
	return path.Join("github", "teams", org, team)

}

func TeamUserFile(org string, team string, user string) string {
	return path.Join(TeamDir(org, team), user)
}

func PrDir(org string, repo string) string {
	return path.Join("github", "prs", org, repo, "pr")
}

func PrData(org string, repo string, name string) string {
	return path.Join("github", "prs", org, repo, name)
}

func PrFile(org string, repo string, number string) string {
	return path.Join(PrDir(org, repo), number+".json")
}

func IssueData(org string, repo string, name string) string {
	return path.Join("github", "issues", org, repo, name)
}

func IssueDir(org string, repo string) string {
	return path.Join("github", "issues", org, repo, "issue")
}

func IssueFile(org string, repo string, number string) string {
	return path.Join(IssueDir(org, repo), number+".json")
}
