package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/deanishe/awgo/util"
	"github.com/ncruces/zenity"
)

type WorkflowConfig struct {
	APIToken    string
	WorkspaceID int    `env:"workspace_id"`
	JiraURL     string `env:"jira_url"`
}

type CurrentTogglTrack struct {
	ID          int
	Description string
}

const (
	repo            = "atoftegaard-git/alfred-jira-toggl"
	keychainAccount = "alfred-jira-toggl"
	updateJobName   = "checkForUpdates"
)

var (
	wf                           *aw.Workflow
	cfg                          *WorkflowConfig
	clearAuthFlag                bool
	authFlag                     bool
	stopTogglEntryFlag           bool
	startTogglEntryFlag          bool
	copyIssueKeyFlag             bool
	overrideDescriptionFlag      bool
	checkRunningFlag             bool
	overrideIssueKeyFlag         string
	addToEmptyDescriptionFlag    bool
	checkForUpdatesFlag          bool
	promptForUpdateAvailableFlag bool
	doUpdateFlag                 bool
)

func init() {
	wf = aw.New(
		update.GitHub(repo),
		aw.MagicPrefix("fuck"),
	)
}

func GetURL() string {
	script := `
    tell application "Google Chrome"
        get URL of active tab of first window
    end tell
    `
	output, err := util.RunAS(script)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Issue:", output)
	return output
}

func run() {
	if err := cli.Parse(wf.Args()); err != nil {
		wf.FatalError(err)
	}

	if checkForUpdatesFlag {
		wf.Configure(aw.TextErrors(true))
		log.Println("Checking for updates...")
		if err := wf.CheckForUpdate(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
		log.Println("Running update check in background...")
		cmd := exec.Command(os.Args[0], "--check-for-updates")
		if err := wf.RunInBackground(updateJobName, cmd); err != nil {
			log.Printf("Error starting update check: %s", err)
		}
	}

	if promptForUpdateAvailableFlag {
		if wf.UpdateAvailable() {
			fmt.Printf("prompt")
			return
		}
	}

	if doUpdateFlag {
		wf.Updater.Install()
		return
	}

	if clearAuthFlag {
		err := wf.Keychain.Delete(keychainAccount)

		if err != nil {
			fmt.Printf("Error deleting toggl api token from keychain")
			log.Fatal(err)
		}
		fmt.Printf("Toggl api token deleted from keychain")
		return
	}

	if authFlag {
		_, api_token, err := zenity.Password(zenity.Title("Enter Toggl API token"))
		if err != nil {
			log.Fatal(err)
		}

		err = wf.Keychain.Set(keychainAccount, api_token)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	api_token, err := wf.Keychain.Get(keychainAccount)
	if err != nil {
		cmd := exec.Command(os.Args[0], "-auth")

		_, err := util.RunCmd(cmd)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Toggl api token added to keychain")
		return
	}
	cfg = &WorkflowConfig{APIToken: api_token}
	err = wf.Config.To(cfg)
	if err != nil {
		log.Fatal(err)
	}

	av := aw.NewArgVars()
	if checkRunningFlag {
		res := GetCurrentTracking()
		if res != "not running" {
			av.Var("running", "true")
			var currentTrackBody CurrentTogglTrack
			json.Unmarshal([]byte(res), &currentTrackBody)
			if currentTrackBody.Description == "" {
				av.Var("prompt", "false")
			} else {
				av.Var("prompt", "true")
			}
		} else {
			av.Var("running", "false")
		}
		if err := av.Send(); err != nil {
			panic(err)
		}
		return
	}

	var issue string
	url := GetURL()
	if strings.HasPrefix(url, cfg.JiraURL+"/browse") {
		issue = regexp.MustCompile("[a-zA-Z]+-[0-9]+").FindString(url)
	}

	if addToEmptyDescriptionFlag {
		res := GetCurrentTracking()
		if res != "not running" {
			var currentTrackBody CurrentTogglTrack
			json.Unmarshal([]byte(res), &currentTrackBody)
			if currentTrackBody.Description == "" {
				av.Var("prompt", "false")
				log.Println("Description is empty, adding issue to currently running entry")
				av.Arg(AddDescription(overrideIssueKeyFlag, currentTrackBody.ID))
			} else {
				av.Var("prompt", "true")
			}
			if err := av.Send(); err != nil {
				panic(err)
			}
		}
		return
	}

	if overrideIssueKeyFlag != "" && !overrideDescriptionFlag {
		StartTracking(overrideIssueKeyFlag)
		return
	}

	if stopTogglEntryFlag {
		res := GetCurrentTracking()
		if res != "not running" {
			var currentTrackBody CurrentTogglTrack
			json.Unmarshal([]byte(res), &currentTrackBody)
			err := StopTogglEntry(currentTrackBody.ID)
			if err != nil {
				log.Fatal(err)
				fmt.Printf("Current toggl could not be stopped")
			} else {
				fmt.Printf("Current toggl entry stopped")
			}
		}
		return
	}

	if overrideDescriptionFlag {
		if overrideIssueKeyFlag != "" {
			res := GetCurrentTracking()
			var currentTrackBody CurrentTogglTrack
			json.Unmarshal([]byte(res), &currentTrackBody)
			if currentTrackBody.Description != "" {
				log.Println("Overriding description")
				av.Arg(AddDescription(overrideIssueKeyFlag, currentTrackBody.ID))
			}
		} else {
			res := GetCurrentTracking()
			var currentTrackBody CurrentTogglTrack
			json.Unmarshal([]byte(res), &currentTrackBody)
			if currentTrackBody.Description != "" {
				log.Println("Overriding description")
				av.Arg(AddDescription(issue, currentTrackBody.ID))
			}
		}
		if err := av.Send(); err != nil {
			panic(err)
		}
		return
	}

	if startTogglEntryFlag {
		res := GetCurrentTracking()
		if res == "not running" {
			StartTracking(issue)
		} else {
			var currentTrackBody CurrentTogglTrack
			json.Unmarshal([]byte(res), &currentTrackBody)
			if currentTrackBody.Description == "" {
				av.Var("prompt", "false")
				log.Println("Description is empty, adding issue to currently running entry")
				av.Arg(AddDescription(issue, currentTrackBody.ID))
			} else {
				av.Var("prompt", "true")
			}
			if err := av.Send(); err != nil {
				panic(err)
			}
		}
		return
	}

	if copyIssueKeyFlag {
		if issue != "" {
			fmt.Printf(issue)
		}
		return
	}
	return
}

func main() {
	wf.Run(run)
}
