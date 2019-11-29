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
			result := newConfig("", test.token, test.user, test.url).HasJira()
			if result != test.valid {
				t.Errorf("HashJira() wanted: %t, got: %t", test.valid, result)
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
