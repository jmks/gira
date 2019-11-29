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

const jiraTokenEnv = "GIRA_JIRA_TOKEN"
const jiraUserEnv = "GIRA_JIRA_USER"
const jiraUrlEnv = "GIRA_JIRA_URL"
const jiraIssuePatternEnv = "GIRA_JIRA_ISSUE_PATTERN"
const gitBranchRefPrefix = "refs/heads/"

type Branch struct {
	Name              plumbing.ReferenceName
	JiraStatus        string
	SelectedForDelete bool
}

func main() {
	branches := findGitBranches()
	addBranchStatusFromJira(branches)

	showUserSelection(branches)
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
		branchCell := tview.NewTableCell(branch.Name.String()).
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
