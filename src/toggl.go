package main

import (
	"encoding/json"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func GetCurrentTracking() (msg string) {
	req, err := http.NewRequest(http.MethodGet, "https://api.track.toggl.com/api/v9/me/time_entries/current", nil)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(cfg.APIToken, "api_token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	defer resp.Body.Close()
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
	togglProjectURL := fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%d/projects/%d", cfg.WorkspaceID, projectID)

    log.Printf("Project ID to look up: %d ", projectID)

	req, err := http.NewRequest(http.MethodGet, togglProjectURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(cfg.APIToken, "api_token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("JSON returned from getting project", string(body))

    var currentProjectBody CurrentTogglProject
    json.Unmarshal([]byte(body), &currentProjectBody)
    return currentProjectBody.Name
}

func StartTracking(issue string) string {
	now := time.Now()
	var unix_start_time = -1 * now.Unix()
	start_date := now.Format(time.RFC3339)

	create_tracking_jsonbody := fmt.Sprintf(`{"created_with": "alfred", "description": "%s", "duration": %d, "start": "%s", "workspace_id": %d}`, issue, unix_start_time, start_date, cfg.WorkspaceID)
	log.Println("Payload to start tracking:", create_tracking_jsonbody)
	bodyBuffered := bytes.NewBuffer([]byte(create_tracking_jsonbody))
	togglUrl := fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%d/time_entries", cfg.WorkspaceID)

	req, err := http.NewRequest(http.MethodPost, togglUrl, bodyBuffered)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(cfg.APIToken, "api_token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
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
	currentTogglTrackUrl := fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%d/time_entries/%d", cfg.WorkspaceID, currentTrackID)

	newDescription := fmt.Sprintf(`{"workspace_id":%d,"description":"%s"}`, cfg.WorkspaceID, description)
	log.Println("Payload to add description:", newDescription)
	bodyBuffered := bytes.NewBuffer([]byte(newDescription))

	req, err := http.NewRequest(http.MethodPut, currentTogglTrackUrl, bodyBuffered)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(cfg.APIToken, "api_token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("JSON returned from editing description", string(body))
	return fmt.Sprintf("%s added to current toggl entry", description)
}

func StopTogglEntry(currentTrackID int) error {
	currentTogglTrackUrl := fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%d/time_entries/%d/stop", cfg.WorkspaceID, currentTrackID)
	req, err := http.NewRequest(http.MethodPatch, currentTogglTrackUrl, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(cfg.APIToken, "api_token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Current Toggl entry stopped:", string(body))
	return nil
}
