package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const giraEnvPrefix = "GIRA"
const jiraIssuePatternEnv = "JIRA_ISSUE_PATTERN"
const jiraTokenEnv = "JIRA_TOKEN"
const jiraUserEnv = "JIRA_USER"
const jiraUrlEnv = "JIRA_URL"

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
	config := readConfiguration()

	// TODO: replace with cobra?
	command := "delete"

	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "delete":
		deleteLocalBranches(config)
	case "branch":
		createLocalBranchFromJiraIssue(config)
	default:
		fmt.Printf("Unknown command '%s'\n", command)
	}

	os.Exit(0)
}

func readConfiguration() *Config {
	viper.SetDefault(withGiraPrefix(jiraIssuePatternEnv), "")
	viper.SetDefault(withGiraPrefix(jiraTokenEnv), "")
	viper.SetDefault(withGiraPrefix(jiraUserEnv), "")
	viper.SetDefault(withGiraPrefix(jiraUrlEnv), "")

	viper.SetEnvPrefix(giraEnvPrefix)
	viper.BindEnv(jiraIssuePatternEnv)
	viper.BindEnv(jiraTokenEnv)
	viper.BindEnv(jiraUserEnv)
	viper.BindEnv(jiraUrlEnv)

	viper.SetConfigType("toml")
	viper.SetConfigName(".gira")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/")

	viper.ReadInConfig()

	return &Config{
		issuePattern: viper.GetString(jiraIssuePatternEnv),
		jiraToken:    viper.GetString(jiraTokenEnv),
		jiraUser:     viper.GetString(jiraUserEnv),
		jiraURL:      viper.GetString(jiraUrlEnv),
	}
}

func withGiraPrefix(s string) string {
	return giraEnvPrefix + "_" + s
}

func (c Config) HasJira() bool {
	return c.jiraToken != "" && c.jiraUser != "" && c.jiraURL != ""
}

func createLocalBranchFromJiraIssue(config *Config) {
	issueKey := ""
	if len(os.Args) >= 3 {
		issueKey = os.Args[2]
	}

	if issueKey == "" {
		fmt.Println("The branch command requires a Jira issue key")
		os.Exit(1)
	}

	if !config.HasJira() {
		fmt.Println("Jira configuration required")
		os.Exit(1)
	}

	title, _, err := fetchJiraInfo(issueKey, config)
	if err != nil {
		fmt.Printf("Error fetching issue %s from Jira: %s\n", issueKey, err)
	}

	branchName := formatBranchName(title, issueKey, "-")

	repo, err := gitRepository()
	if err != nil {
		fmt.Printf("Git problem: %s\n", err)
		os.Exit(1)
	}

	headRef, err := repo.Head()
	if err != nil {
		fmt.Printf("Git problem: %s\n", err)
		os.Exit(1)
	}

	newBranchRefName := plumbing.NewBranchReferenceName(branchName)
	newRef := plumbing.NewHashReference(newBranchRefName, headRef.Hash())
	err = repo.Storer.SetReference(newRef)
	if err != nil {
		fmt.Printf("Git problem: %s\n", err)
	}

	checkout := exec.Command("git", "checkout", branchName)
	err = checkout.Run()
	if err != nil {
		fmt.Printf("Git problem: %s\n", err)
	}
}

func formatBranchName(title, prefix, delimiter string) string {
	nonChars := regexp.MustCompile("[^A-Za-z0-9]")
	repeatedSpaces := regexp.MustCompile("\\s{2,}")
	cleanedTitle := nonChars.ReplaceAllLiteralString(title, " ")
	normalizedTitle := repeatedSpaces.ReplaceAllLiteralString(cleanedTitle, " ")
	normalizedTitle = strings.TrimSuffix(normalizedTitle, " ")

	return prefix + delimiter + strings.ReplaceAll(strings.ToLower(normalizedTitle), " ", delimiter)
}

func deleteLocalBranches(config *Config) {
	repo, err := gitRepository()
	if err != nil {
		fmt.Printf("Git problem: %s\n", err)
		os.Exit(1)
	}
	branches, err := getBranches(repo)
	if err != nil {
		fmt.Printf("Git problem: %s\n", err)
		os.Exit(1)
	}

	for _, branch := range branches {
		_, status, err := fetchJiraInfo(branch.jiraIssueKey(config.issuePattern), config)
		if err != nil {
			fmt.Printf("Error requesting Jira informtion: %s\n", err)
		}

		branch.JiraStatus = status
	}

	cancelled := showUserSelection(branches)
	if cancelled {
		os.Exit(1)
	}

	err = deleteSelectedBranches(repo, branches)
	if err != nil {
		fmt.Printf("Error deleting branch(es): %s\n", err)
		os.Exit(1)
	}
}

func gitRepository() (*git.Repository, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func getBranches(repo *git.Repository) ([]*Branch, error) {
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

func fetchJiraInfo(issueKey string, config *Config) (title, status string, err error) {
	if !config.HasJira() || config.issuePattern == "" || issueKey == "" {
		return "", "", nil
	}

	basicAuth := jira.BasicAuthTransport{
		Username: config.jiraUser,
		Password: config.jiraToken,
	}

	client, err := jira.NewClient(basicAuth.Client(), config.jiraURL)
	if err != nil {
		return "", "", err
	}

	issue, _, err := client.Issue.Get(issueKey, &jira.GetQueryOptions{})
	if err != nil {
		return "", "", err
	}

	return issue.Fields.Summary, issue.Fields.Status.Name, nil
}

func (b Branch) jiraIssueKey(pattern string) string {
	re := regexp.MustCompile(pattern)

	found := re.Find([]byte(b.DisplayName()))

	return string(found)
}

func (b Branch) DisplayName() string {
	return strings.TrimPrefix(b.Name.String(), gitBranchRefPrefix)
}

func showUserSelection(branches []*Branch) bool {
	userCancelled := true

	app := tview.NewApplication()

	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	for r := 0; r < len(branches); r++ {
		branch := branches[r]

		selectionCell := tview.NewTableCell("").SetAlign(tview.AlignLeft)
		decorateSelectedCell(selectionCell, branch.SelectedForDelete)

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
			userCancelled = false
			app.Stop()
		}
	}).SetSelectedFunc(func(row int, column int) {
		branches[row].SelectedForDelete = !branches[row].SelectedForDelete

		decorateSelectedCell(table.GetCell(row, 0), branches[row].SelectedForDelete)
	})

	frame := tview.
		NewFrame(table).
		SetBorders(2, 2, 2, 2, 4, 4).
		AddText("gira: Select branches to delete", true, tview.AlignCenter, tcell.ColorWhite).
		AddText("<ctrl-c>: immediately quit", false, tview.AlignCenter, tcell.ColorWhite).
		AddText("<esc>: quit and delete", false, tview.AlignCenter, tcell.ColorWhite).
		AddText("<enter>: (de)select", false, tview.AlignCenter, tcell.ColorWhite)

	if err := app.SetRoot(frame, true).SetFocus(frame).Run(); err != nil {
		panic(err)
	}

	return userCancelled
}

func decorateSelectedCell(cell *tview.TableCell, selected bool) {
	if selected {
		cell.SetText("X")
		cell.SetTextColor(tcell.ColorRed)
	} else {
		cell.SetText(" ")
		cell.SetTextColor(tcell.ColorBlack)
	}

	cell.SetBackgroundColor(tcell.ColorBlack)
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

func deleteSelectedBranches(repo *git.Repository, branches []*Branch) error {
	for _, branch := range branches {
		if branch.SelectedForDelete {
			err := repo.Storer.RemoveReference(branch.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
