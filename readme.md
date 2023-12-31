## Copy Jira Key
Copies the url from a focused Chrome window if the URL matches `Jira URL` input in the workflow config, and returns the Jira key eg. JIRA-2727.

Afterwards it either copies it into your clipboard or starts tracking with Toggl.

## Installation
* [Download the latest release](https://github.com/atoftegaard-git/alfred-jira-toggl/releases)
* Open the downloaded file in Finder.
* Make sure [Toggl Track](https://toggl.com/track/time-tracking-mac/) is installed
* If running on macOS Catalina or later, you _**MUST**_ add Alfred to the list of security exceptions for running unsigned software. See [this guide](https://github.com/deanishe/awgo/wiki/Catalina) for instructions on how to do this.

## Prerequisites

When trying to start a toggl tracking through the workflow for the first time, you'll be asked for a Toggl API token, this can be found [here](https://track.toggl.com/profile).

To start tracking with this workflow, you'll need a workspace ID, this can be found with the following command, using the API token found above:
`curl -u <API_TOKEN>:api_token -H "Content-Type: application/json" -X GET https://api.track.toggl.com/api/v9/workspaces`

## Actions
* ⌘ + .   - Copy the jira issue of the active chrome window to your clipboard

* ⌘ ⇧ + . - Copy the jira issue of the active chrome window and start a toggl entry with it.
    * If an entry is already running, you will be prompted if you want to start a new one or if you want to override the description in the running tracker.
