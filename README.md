# Preflight
Preflight is a checklist manager which integrates with Todoist. It will take a list of items from either Trello or a config file and push them to your Todoist inbox either on a regular schedule (e.g. every Tuesday and Thursday at 9AM) or when you invoke it from a command line.

## Installation
1. Install go and set GOPATH.
2. Download source with `go get github.com/jsutton9/preflight`.
3. Build the executable with `go build -o preflight github.com/jsutton9/preflight/main`.

## Commands
- `preflight config CONFIG_FILE` sets the configuration to that in the specified config file.
- `preflight update` updates your Todoist inbox according to the configured schedule. Set a cron job for this.
- `preflight TEMPLATE_NAME` adds to-do items from the specified template to your Todoist inbox.

## Configuration
Preflight is configured with a JSON file. This file contains one JSON object with the following fields:
- "api\_token": This is your Todoist API token. Find it in the web app at *gear icon* > *Todoist Settings* > *Account* > *API token*.
- "timezone": IANA time zone name, e.g. "America/Denver"
- "trello": (optional) An object with the following fields:
  - "key": (optional) [Trello developer API key](https://trello.com/app-key)
  - "token": (optional) manual Trello token (see above link)
  - "board\_name": (optional) Trello board title
  These global settings may be overriden by per-template settings.
- "templates": An object mapping the name of each template to a configuration object with any of these fields:
  - "tasks": (optional) A list of to-do item strings
  - "trello": (optional) An object with the following fields:
    - "key": optional if set globally
    - "token": optional if set globally
    - "board\_name": optional if set globally
    - "list\_name": Trello list title
  - "schedule": (optional) An object with the following fields:
    - "start": Time at which to add items to inbox, e.g. "17:00"
    - "end": (optional) Time at which to remove from inbox if still there
    - "interval": (optional, default 1) Minimum interval in days between posts, e.g. 3 for every three days
    - "days": (optional, default every day) List of days of the week, e.g. ["Monday", "Wed", "fri"]
