package jira

import "path"

func ProjectDir() string {
	return path.Join("jira")
}

func IssueDir(project string) string {
	return path.Join(ProjectDir(), project)
}
