# Preflight
Preflight is a checklist manager which integrates with Todoist. It will take a list of items from either Trello or a config file and push them to your Todoist inbox either on a regular schedule (e.g. every Tuesday and Thursday at 9AM) or when you invoke it from a command line.

## Installation
1. Install go and set GOPATH.
2. Download source with `go get github.com/jsutton9/preflight`.
3. Build the executable with `go build -o preflight github.com/jsutton9/preflight/cli`.
4. Install and start mongodb.
5. Run preflight with no arguments to get a list of available commands.

## Integrations
- You will need a Todoist API token, which you can find in the web app at *gear icon* > *Todoist Settings* > *Account* > *API token*.
- For optional Trello integration, you will need a [Trello developer API key](https://trello.com/app-key) and manual Trello token

## Checklists
Checklists are configured using json files. Such a file contains one json object with the following fields:
- tasksSource: May be "preflight" or "trello"
- tasksTarget: Must be "todoist"
- isScheduled: true iff the checklist is to be added to your inbox on a regular schedule
- tasks: (optional) List of to-do item strings
- trello: (optional) An object with the following fields:
  - board: Title of Trello board
  - name: Title of Trello list
- schedule: (optional) An object with the following fields:
  - interval: (optional, default 1) Minimum interval in days between posts, e.g. 3 for every three days
  - days: (optional, default every day) List of days of the week, e.g. ["Monday", "Wed", "fri"]
  - start: Time at which to add items to inbox, e.g. "17:00"
  - end: (optional) Time at which to remove from inbox if still there
