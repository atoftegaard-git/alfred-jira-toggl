package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "os/exec"
    "regexp"
    "strings"

    aw "github.com/deanishe/awgo"
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
    keychainAccount = "alfred-jira-toggl"
)

var (
    wf                      *aw.Workflow
    cfg                     *WorkflowConfig
    clearAuthFlag           bool
    authFlag                bool
    stopTogglEntryFlag      bool
    startTogglEntryFlag     bool
    overrideDescriptionFlag bool
    checkRunningFlag        bool
    overrideIssueKeyFlag    string
)

func init() {
    wf = aw.New()
    flag.BoolVar(&clearAuthFlag, "clear-auth", false, "clear toggl api token from keychain")
    flag.BoolVar(&authFlag, "auth", false, "adds toggl api token to keychain")
    flag.BoolVar(&stopTogglEntryFlag, "stop-entry", false, "stops current toggl entry")
    flag.BoolVar(&startTogglEntryFlag, "start-entry", false, "starts a new empty toggl entry")
    flag.BoolVar(&overrideDescriptionFlag, "override-description", false, "overrides description in current toggl entry")
    flag.BoolVar(&checkRunningFlag, "check", false, "checks if toggl is currently running")
    flag.StringVar(&overrideIssueKeyFlag, "issue-key", "", "enables the user to pass a issue key to the workflow")
}

func CheckRunning() string {
    res := GetCurrentTracking()
    if res != "not running" {
        fmt.Printf("Toggl is running")
        return "running"
    } else {
        fmt.Printf("Toggl is not currently running")
        return "not-running"
    }
}

func OverrideIssueKey(issueKey string) {
    StartTracking(issueKey)
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
    wf.Args()
    flag.Parse()

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

    if checkRunningFlag {
        CheckRunning()
        return
    }

    if overrideIssueKeyFlag != "" {
        StartTracking(overrideIssueKeyFlag)
        return
    }

    if startTogglEntryFlag {
        StartTracking("")
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

    var issue string
    url := GetURL()
    if strings.HasPrefix(url, cfg.JiraURL+"/browse") {
        issue = regexp.MustCompile("[a-zA-Z]+-[0-9]+").FindString(url)
    } else if os.Getenv("action") == "toggl" {
        StartTracking("")
        return
    }
    if os.Getenv("action") == "copy" {
        if issue != "" {
            fmt.Printf(issue)
        }
        return
    }

    av := aw.NewArgVars()
    if overrideDescriptionFlag {
        res := GetCurrentTracking()
        var currentTrackBody CurrentTogglTrack
        json.Unmarshal([]byte(res), &currentTrackBody)
        if currentTrackBody.Description != "" {
            log.Println("Overriding description")
            av.Arg(AddDescription(issue, currentTrackBody.ID))
        }
        return
    }

    res := GetCurrentTracking()
    if res == "not running" {
        StartTracking(issue)
    } else {
        var currentTrackBody CurrentTogglTrack
        json.Unmarshal([]byte(res), &currentTrackBody)
        if currentTrackBody.Description == "" {
            av.Var("prompt","false")
            log.Println("Description is empty, adding issue to currently running entry")
            av.Arg(AddDescription(issue, currentTrackBody.ID))
        } else {
            av.Var("prompt","true")
        }
        if err := av.Send(); err != nil {
            panic(err)
        }
    }
    return
}

func main() {
    wf.Run(run)
}
