package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"regexp"
	"strings"
)

const jiraIssuePatternEnv = "GIRA_JIRA_ISSUE_PATTERN"
const jiraTokenEnv = "GIRA_JIRA_TOKEN"
const jiraUserEnv = "GIRA_JIRA_USER"
const jiraUrlEnv = "GIRA_JIRA_URL"

const gitBranchRefPrefix = "refs/heads/"

type Config struct {
	issuePattern string
	jiraToken    string
	jiraUser     string
	jiraURL      string
}

type Branch struct {
	Name              plumbing.ReferenceName
	JiraStatus        string
	SelectedForDelete bool
}

func main() {
	config := newConfig(
		os.Getenv(jiraIssuePatternEnv),
		os.Getenv(jiraTokenEnv),
		os.Getenv(jiraUserEnv),
		os.Getenv(jiraUrlEnv),
	)

	branches, err := findGitBranches()
	if err != nil {
		fmt.Printf("Git problem: %s", err)
		os.Exit(1)
	}

	err = addBranchStatusFromJira(branches, config)
	if err != nil {
		fmt.Printf("Error requesting Jira informtion: %s", err)
	}

	showUserSelection(branches)

	for _, branch := range branches {
		if branch.SelectedForDelete {
			fmt.Println(branch.DisplayName())
		}
	}
}

func newConfig(issuePattern, jiraToken, jiraUser, jiraURL string) *Config {
	return &Config{
		issuePattern: issuePattern,
		jiraToken:    jiraToken,
		jiraUser:     jiraUser,
		jiraURL:      jiraURL,
	}
}

func (c Config) HasJira() bool {
	return c.jiraToken != "" && c.jiraUser != "" && c.jiraURL != ""
}

func findGitBranches() ([]*Branch, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}

	iter, err := repo.Branches()
	if err != nil {
		return nil, err
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

	return branches, nil
}

func addBranchStatusFromJira(branches []*Branch, config *Config) error {
	if !config.HasJira() {
		return nil
	}

	basicAuth := jira.BasicAuthTransport{
		Username: config.jiraUser,
		Password: config.jiraToken,
	}

	client, err := jira.NewClient(basicAuth.Client(), config.jiraURL)
	if err != nil {
		return err
	}

	for _, branch := range branches {
		if config.issuePattern == "" {
			continue
		}

		key := branch.jiraIssueKey(config.issuePattern)
		if key == "" {
			continue
		}

		issue, _, err := client.Issue.Get(key, &jira.GetQueryOptions{})
		if err != nil {
			return err
		}

		branch.JiraStatus = issue.Fields.Status.Name
	}

	return nil
}

func (b Branch) jiraIssueKey(pattern string) string {
	re := regexp.MustCompile(pattern)

	found := re.Find([]byte(b.Name.String()))

	return string(found)
}

func (b Branch) DisplayName() string {
	return strings.TrimLeft(b.Name.String(), gitBranchRefPrefix)
}

func showUserSelection(branches []*Branch) {
	app := tview.NewApplication()

	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	for r := 0; r < len(branches); r++ {
		branch := branches[r]

		selectionCell := tview.NewTableCell(getSelectedText(branch.SelectedForDelete)).
			SetTextColor(getSelectedTextColor(branch.SelectedForDelete)).
			SetAlign(tview.AlignLeft)
		branchCell := tview.NewTableCell(branch.DisplayName()).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft).
			SetExpansion(1)

		table.SetCell(r, 0, selectionCell)
		table.SetCell(r, 1, newStatusCell(branch.JiraStatus))
		table.SetCell(r, 2, branchCell)
	}

	table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
	}).SetSelectedFunc(func(row int, column int) {
		branches[row].SelectedForDelete = !branches[row].SelectedForDelete

		table.GetCell(row, 0).
			SetText(getSelectedText(branches[row].SelectedForDelete)).
			SetTextColor(getSelectedTextColor(branches[row].SelectedForDelete))
	})

	frame := tview.NewFrame(table).
		SetBorders(2, 2, 2, 2, 4, 4).
		AddText("Select branches to delete with <enter>", true, tview.AlignCenter, tcell.ColorWhite)

	if err := app.SetRoot(frame, true).SetFocus(frame).Run(); err != nil {
		panic(err)
	}
}

func getSelectedText(selected bool) string {
	if selected {
		return "X"
	} else {
		return " "
	}
}

func getSelectedTextColor(selected bool) tcell.Color {
	if selected {
		return tcell.ColorRed
	}

	return tcell.ColorWhite
}

func newStatusCell(status string) *tview.TableCell {
	color := tcell.ColorLightBlue

	switch status {
	case "Done":
		color = tcell.ColorGreen
	case "Discarded":
		color = tcell.ColorGreen
	case "To Do":
		color = tcell.ColorGrey
	case "Ready for Dev":
		color = tcell.ColorGrey
	}

	return tview.NewTableCell(status).
		SetTextColor(color).
		SetAlign(tview.AlignLeft)
}