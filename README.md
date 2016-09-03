# Preflight
Preflight is a checklist manager which integrates with Todoist. It will take a list of items from either Trello or a config file and push them to your Todoist inbox either on a regular schedule (e.g. every Tuesday and Thursday at 9AM) or when you invoke it from a command line or the API.

## Installation
1. Install go and set GOPATH.
2. Download source with `go get github.com/jsutton9/preflight`
3. Build executable
  - for cli: `go build -o preflight github.com/jsutton9/preflight/cli`
  - for api: `go build -o preflight-api github.com/jsutton9/preflight/api`
4. Install and start mongodb.

## Configuration
The API server requires a json configuration file containing an object with these fields:
- port: (default 443) port on which to serve API
- certFile: SSL certificate file
- keyFile: SSL key file
- errLog: (default stderr) file to which to log errors
- databaseServer: (default localhost) IP or domain of MongoDB server
- databaseUsersCollection: (default users) name of MongoDB collection to use for users
- trelloAppKey: (optional) Trello application key
- secretFile: (default: /etc/preflight/secret) file in which to store secret for authentication between nodes

## Integrations
- You will need a Todoist API token, which you can find in the web app at *gear icon* > *Todoist Settings* > *Account* > *API token*.
- For optional Trello integration, you will need a [Trello developer API key](https://trello.com/app-key) and manual Trello token

## API
Before running the API server, you will need to register the server in MongoDB if you haven't already. Use `./preflight register-node CONFIG\_FILE`. This will generate a node secret and write it to the database.
Start the API server with `./preflight-api CONFIG\_FILE`.

API requests must all use https.
Requests may be authenticated in three ways:
- with a client token, by adding parameter `token={token-secret}`. A token can have these permissions:
  - checklistRead
  - checklistWrite
  - checklistInvoke
  - generalRead
  - generalWrite
- with a node secret, by adding parameter `nodeSecret={node-secret}`
  - This may be used as a substitute for a client token if you also include parameter `user={user-id}`.
- with basic authentication, when creating a new client token

Currently supported API calls:
- POST /users
  - authentication: node secret
  - body: `{"email": $EMAIL, "password": $PASSWORD}`
- DELETE /users/{user-id}
  - authentication: node secret
- GET /checklists
  - authentication: checklistRead
  - response body: json list of checklists (see Checklists section)
- GET /checklists/{checklist-id}
  - authentication: checklistRead
  - response body: json checklist (see Checklists section)
- POST /checklists
  - authentication: checklistWrite
  - body: `{"name": $NAME, "checklist": $CHECKLIST}`
    - CHECKLIST as in Checklist section
  - response Location header: URL for new checklist
- POST /checklists/{checklist-id}/invoke
  - authentication: checklistInvoke
- PUT /checklists/{checklist-id}
  - authentication: checklistWrite
  - body: a checklist (see Checklist section)
- DELETE /checklists/{checklist-id}
  - authentication: checklistWrite
- GET /tokens
  - authentication: generalRead
  - response body: json list of checklist objects (see Tokens section), with secrets removed
- POST /tokens
  - authentication: basic auth
  - body: `{"permissions": $PERMISSIONS, "expiryHours": $HOURS_UNTIL_EXPIRATION, "description": $DESCRIPTION_STRING}`
    - PERMISSIONS as in Token object
  - response body: token object (see Tokens section)
- DELETE /tokens/{token-id}
  - authentication: generalWrite
- GET /settings
  - authentication: generalRead
  - response body: `{"timezone": $TIMEZONE, "trelloBoard": $TRELLO_BOARD}`
- PUT /settings/timezone
  - authentication: generalWrite
  - body: IANA timezone string, e.g. "America/Denver"
- PUT /settings/trelloBoard
  - authentication: generalWrite
  - body: name of Trello board

## Checklists
A checklist is represented by a json object with the following fields:
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

## Tokens
A user token is represented by a json object with the following fields:
  - id: randomly generated identifier
  - secret: randomly generated secret
  - permissions: object with these boolean fields:
    - checklistRead
    - checklistWrite
    - checklistInvoke
    - generalRead
    - generalWrite
  - expiry: ISO-8601 timestamp when the token expires
  - description: client-provided description string
