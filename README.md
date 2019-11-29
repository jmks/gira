# Gira

A tool for managing *GI*t branches with ji*RA*.

## Description

I frequently have many feature branches. However, CI/Github actions may rebase the feature branch before it is merged to mainline.
Thus, I have many local branches that are not merged into mainline that I would like to remove, with the knowledge the issue is done.

Even if Jira is not configured, or the branch is not tracked in Jira, `gira` can still help you multi-select branches to be deleted.

![](/screenshot.png?raw=true "Selection screen")

## Environment Variables

Currently requires the following environment variables to be set up:

| Variable                | Value                                                                                                                 |
| -------------           | -------------                                                                                                         |
| GIRA_JIRA_ISSUE_PATTERN | A pattern for identifying a Jira Issue key from a branch name e.g. PM-1701-my-feature-branch has the pattern 'PM-\d+' |
| GIRA_JIRA_TOKEN         | Jira API TOKEN                                                                                                        |
| GIRA_JIRA_USER          | User of the Jira API token e.g. an email address                                                                      |
| GIRA_JIRA_URL           | Jira base URL e.g. https://issues.apache.org/jira/                                                                    |

## License

This project is released under the terms of the [MIT license](http://en.wikipedia.org/wiki/MIT_License).
