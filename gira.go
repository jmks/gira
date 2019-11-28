package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"regexp"
	"strings"
)

const jiraTokenEnv = "GIRA_JIRA_TOKEN"
const jiraUserEnv = "GIRA_JIRA_USER"
const jiraUrlEnv = "GIRA_JIRA_URL"
const jiraIssuePatternEnv = "GIRA_JIRA_ISSURE_PATTERN"
const gitBranchRefPrefix = "refs/heads/"

type Branch struct {
	Name       plumbing.ReferenceName
	JiraStatus string
}

func main() {
	branches := findGitBranches()
	addBranchStatusFromJira(branches)

	printByStatus(branches)
}

func findGitBranches() []*Branch {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not open current directory!")
		fmt.Println(err)
		os.Exit(1)
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		fmt.Println("Are you working in a git directory?")
		os.Exit(1)
	}

	iter, err := repo.Branches()
	if err != nil {
		fmt.Println("Could not read branches:")
		fmt.Println(err)
		os.Exit(1)
	}

	branches := []*Branch{}
	iter.ForEach(func(b *plumbing.Reference) error {
		if strings.HasPrefix(b.Name().String(), gitBranchRefPrefix) {
			branches = append(branches, &Branch{
				Name: b.Name(),
			})
		}
		return nil
	})

	return branches
}

func addBranchStatusFromJira(branches []*Branch) {
	basicAuth := jira.BasicAuthTransport{
		Username: os.Getenv(jiraUserEnv),
		Password: os.Getenv(jiraTokenEnv),
	}

	client, err := jira.NewClient(basicAuth.Client(), os.Getenv(jiraUrlEnv))
	if err != nil {
		fmt.Println("Error creating client")
		fmt.Println(err)
		os.Exit(1)
	}

	for _, branch := range branches {
		key := branch.jiraIssueKey(os.Getenv(jiraIssuePatternEnv))
		if key == "" {
			continue
		}

		issue, _, err := client.Issue.Get(key, &jira.GetQueryOptions{})
		if err != nil {
			fmt.Printf("Error finding issue for %s\n", key)
			fmt.Println(err)
		}

		branch.JiraStatus = issue.Fields.Status.Name
	}
}

// TODO: how to make regexp configurable and safe?
func (b Branch) jiraIssueKey(pattern string) string {
	re := regexp.MustCompile(pattern)

	found := re.Find([]byte(b.Name.String()))

	return string(found)
}

func printByStatus(branches []*Branch) {
	byStatus := map[string][]string{}

	for _, branch := range branches {
		displayStatus := branch.JiraStatus
		if displayStatus == "" {
			displayStatus = "Not following " + jiraIssuePatternEnv
		}

		_, ok := byStatus[displayStatus]
		if ok {
			byStatus[displayStatus] = append(byStatus[displayStatus], branch.Name.String())
		} else {
			byStatus[displayStatus] = []string{branch.Name.String()}
		}
	}

	for status, bs := range byStatus {
		fmt.Println(status)
		fmt.Println(strings.Repeat("-", len(status)))
		for _, b := range bs {
			fmt.Println(b)
		}
		fmt.Println("")
	}
}
