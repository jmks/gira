# Gira

A tool for managing *GI*t branches with ji*RA*.

Pronounced "gear-ra" (g like "git" or "gif" :trollface:)

## Description

I frequently have many feature branches. However, CI/Github actions may rebase the feature branch before it is merged to mainline.
Thus, I have many local branches that are not merged into mainline that I would like to remove, with the knowledge the issue is done.

Even if Jira is not configured, or the branch is not tracked in Jira, `gira` can still help you multi-select branches to be deleted.

![](/screenshot.png?raw=true "Selection screen")

## Configuration

Configure with Environment Variables or a Configuration File. Note that the Environment values will override the Configuration File values.

### Environment Variables

| Variable                | Value                                                                                                                 |
| -------------           | -------------                                                                                                         |
| GIRA_JIRA_ISSUE_PATTERN | A pattern for identifying a Jira Issue key from a branch name e.g. PM-1701-my-feature-branch has the pattern 'PM-\d+' |
| GIRA_JIRA_TOKEN         | Jira API TOKEN                                                                                                        |
| GIRA_JIRA_USER          | User of the Jira API token e.g. an email address                                                                      |
| GIRA_JIRA_URL           | Jira base URL e.g. https://issues.apache.org/jira/                                                                    |

### Configuration File

Uses the configuration file `.gira.toml` (technically, the format can be any understood by [viper](https://github.com/spf13/viper#reading-config-files)).

`gira` will first look in the current directory for `./.gira.toml`, then at `$HOME/.config/.gira.toml`.

Look at the TOML [sample](/.gira.sample.toml).

Note that if you copy the example, to rename the file as described above, e.g. `cp .gira.sample.toml ./.gira.toml`

## TODO

* When creating a branch name, if last char is special character, the branch should not end with the delimiter `TASS-some-branch-`
* Add a log to a file
* Does api not work with TASS issues?
* Create a [worker pool](https://gobyexample.com/worker-pools) for Jira requests
* Show `git ls-remotes` info in delete UI
* Move branch delimiter to configuration
* Move branch naming pattern to configuration
* Update README with gif
* Add more tests :crying_cat_face:
* configure command to set configuration

## License

This project is released under the terms of the [MIT license](http://en.wikipedia.org/wiki/MIT_License).
