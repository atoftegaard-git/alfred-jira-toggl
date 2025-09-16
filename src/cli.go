package main

import "flag"

var (
	cli  = flag.NewFlagSet("alfred-jira-toggl", flag.ContinueOnError)
)

func init() {
	cli.BoolVar(&clearAuthFlag, "clear-auth", false, "clear toggl api token from keychain")
	cli.BoolVar(&authFlag, "auth", false, "adds toggl api token to keychain")
	cli.BoolVar(&stopTogglEntryFlag, "stop-entry", false, "stops current toggl entry")
	cli.BoolVar(&startTogglEntryFlag, "start-entry", false, "starts a new empty toggl entry")
	cli.BoolVar(&copyIssueKeyFlag, "copy-issue-key", false, "copies jira key from url")
	cli.BoolVar(&overrideDescriptionFlag, "override-description", false, "overrides description in current toggl entry")
	cli.BoolVar(&checkRunningFlag, "check", false, "checks if toggl is currently running")
	cli.StringVar(&overrideIssueKeyFlag, "issue-key", "", "enables the user to pass a issue key to the workflow")
	cli.BoolVar(&addToEmptyDescriptionFlag, "add-description", false, "add issue key to empty description")
	cli.BoolVar(&checkForUpdatesFlag, "check-for-updates", false, "check for updates")
	cli.BoolVar(&promptForUpdateAvailableFlag, "prompt-for-updates", false, "prompt for updates")
	cli.BoolVar(&doUpdateFlag, "update", false, "Update")
}
