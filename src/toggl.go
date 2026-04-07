package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var (
	httpClient = &http.Client{}
	togglBaseURL = "https://api.track.toggl.com/api/v9"
)

func executeRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(cfg.APIToken, "api_token")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 402 {
		resetsIn := resp.Header.Get("X-Toggl-Quota-Resets-In")
		msg := "Toggl API quota reached."
		if resetsIn != "" {
			msg = fmt.Sprintf("%s Please wait %s seconds before trying again.", msg, resetsIn)
		}
		wf.FatalError(fmt.Errorf("%s", msg))
	}

	return resp, nil
}

func GetCurrentTracking() (msg string) {
	url := fmt.Sprintf("%s/me/time_entries/current", togglBaseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
		return ""
	}

	resp, err := executeRequest(req)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			wf.FatalError(err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	if string(body) == "null" {
		log.Printf("Time tracking not running")
		return "not running"
	} else {
		log.Printf("Time tracking already running")
		log.Println("JSON returned from currently runnning tracking:", string(body))
		return string(body)
	}
}

func GetProjectNameFromID(projectID int) (msg string) {
	togglProjectURL := fmt.Sprintf("%s/workspaces/%d/projects/%d", togglBaseURL, cfg.WorkspaceID, projectID)

	log.Printf("Project ID to look up: %d ", projectID)

	req, err := http.NewRequest(http.MethodGet, togglProjectURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := executeRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			wf.FatalError(err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("JSON returned from getting project", string(body))

	var currentProjectBody CurrentTogglProject
	togglErr := json.Unmarshal([]byte(body), &currentProjectBody)
	if err != nil {
		wf.FatalError(togglErr)
	}
	return currentProjectBody.Name
}

func StartTracking(issue string) string {
	now := time.Now()
	var unix_start_time = -1 * now.Unix()
	start_date := now.Format(time.RFC3339)

	create_tracking_jsonbody := fmt.Sprintf(`{"created_with": "alfred", "description": "%s", "duration": %d, "start": "%s", "workspace_id": %d}`, issue, unix_start_time, start_date, cfg.WorkspaceID)
	log.Println("Payload to start tracking:", create_tracking_jsonbody)
	log.Println("Wworkspace id:", cfg.WorkspaceID)
	bodyBuffered := bytes.NewBuffer([]byte(create_tracking_jsonbody))
	togglUrl := fmt.Sprintf("%s/workspaces/%d/time_entries", togglBaseURL, cfg.WorkspaceID)

	req, err := http.NewRequest(http.MethodPost, togglUrl, bodyBuffered)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := executeRequest(req)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			wf.FatalError(err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var msg string
	if issue == "" {
		msg = "Tracking started"
	} else {
		msg = fmt.Sprintf("Tracking started for %s", issue)
	}
	log.Println("JSON returned from newly started tracking:", string(body))
	return msg
}

func AddDescription(description string, currentTrackID int) (msg string) {
	currentTogglTrackUrl := fmt.Sprintf("%s/workspaces/%d/time_entries/%d", togglBaseURL, cfg.WorkspaceID, currentTrackID)

	newDescription := fmt.Sprintf(`{"workspace_id":%d,"description":"%s"}`, cfg.WorkspaceID, description)
	log.Println("Payload to add description:", newDescription)
	bodyBuffered := bytes.NewBuffer([]byte(newDescription))

	req, err := http.NewRequest(http.MethodPut, currentTogglTrackUrl, bodyBuffered)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := executeRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			wf.FatalError(err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("JSON returned from editing description", string(body))
	return fmt.Sprintf("%s added to current toggl entry", description)
}

func StopTogglEntry(currentTrackID int) error {
	currentTogglTrackUrl := fmt.Sprintf("%s/workspaces/%d/time_entries/%d/stop", togglBaseURL, cfg.WorkspaceID, currentTrackID)
	req, err := http.NewRequest(http.MethodPatch, currentTogglTrackUrl, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	resp, err := executeRequest(req)
	if err != nil {
		log.Println(err)
		return err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			wf.FatalError(err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Current Toggl entry stopped:", string(body))
	return nil
}
