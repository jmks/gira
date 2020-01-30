package main

import (
	"testing"
)

func TestConfiguration(t *testing.T) {
	t.Run("HasJira", func(t *testing.T) {
		tests := []struct {
			token string
			user  string
			url   string
			valid bool
		}{
			{token: "", user: "", url: "", valid: false},
			{token: "abc", user: "", url: "", valid: false},
			{token: "", user: "abc", url: "", valid: false},
			{token: "", user: "", url: "abc", valid: false},
			{token: "abc", user: "anonymous", url: "atlassian", valid: true},
		}

		for _, test := range tests {
			result := Config{
				issuePattern: "",
				jiraToken:    test.token,
				jiraUser:     test.user,
				jiraURL:      test.url,
			}
			if result.HasJira() != test.valid {
				t.Errorf("HashJira() wanted: %t, got: %t", test.valid, result.HasJira())
			}
		}
	})
}

func TestBranch(t *testing.T) {
	t.Run("DisplayName hides the refs/heads/ prefix", func(t *testing.T) {
		branch := Branch{Name: "refs/heads/helpful-branch-name"}

		want := "helpful-branch-name"
		got := branch.DisplayName()

		if got != want {
			t.Errorf("Expected display name to be %s, but was %s", want, got)
		}
	})
}

func TestJiraTitleToBranchName(t *testing.T) {
	tests := []struct {
		title     string
		prefix    string
		delimiter string
		want      string
	}{
		{"Hello World", "P", "-", "P-hello-world"},
		{"Extra  Space", "P", "-", "P-extra-space"},
	}

	for _, test := range tests {
		got := formatBranchName(test.title, test.prefix, test.delimiter)

		if got != test.want {
			t.Errorf("Wanted %s, got %s", test.want, got)
		}
	}
}
