# AutoTickets

[![Tests](https://github.com/mynah22/AutoTickets/actions/workflows/test.yml/badge.svg)](https://github.com/mynah22/AutoTickets/actions/workflows/test.yml)
[![Release](https://github.com/mynah22/AutoTickets/actions/workflows/release.yml/badge.svg)](https://github.com/mynah22/AutoTickets/actions/workflows/release.yml)

A webserver that leverages the Autotask API to serve a simple webpage that shows open tickets that have not been assigned.

- [AutoTickets](#autotickets)
  - [Features](#features)
    - [Backend](#backend)
    - [Frontend](#frontend)
    - [Websockets](#websockets)
  - [Build instructions](#build-instructions)
  - [Usage instructions](#usage-instructions)
    - [Runtime flags](#runtime-flags)
      - [Runtime flags examples](#runtime-flags-examples)
  - [Technical explanations](#technical-explanations)
    - [Determining state change of open tickets](#determining-state-change-of-open-tickets)
    - [Determining if new ticket has been received (client)](#determining-if-new-ticket-has-been-received-client)
  - [Project structure](#project-structure)
    - [Packages](#packages)
    - [Other files / folders](#other-files--folders)
  - [Notes for production use](#notes-for-production-use)

## Features

### Backend

- Written in Golang
- Uses the Echo framework to simplify webserver code
- Stores API secrets on disk as an encrypted file
  - all set up / unlocking of secrets is done through the web UI
- Only polls API during a specified active period
- Uses templates to dynamically render pages
- Server Parameters can be overridden by launching the executable with optional flags
- Use of mutexes on important data structures ensures thread-safety of values in memory

### Frontend

- Simple JS / HTML frontend
  - no JS imports used
  - very cross platform / cross browser compatible
- Only keeps server connection alive while tab is visible
- Instantly detects server failure
  - displays an error message with instructions on restarting server
  - redirects user when server comes back online
- Page blink when new ticket comes in

### Websockets

- This project uses Websockets for bidirectional communication between client and server
  - server can push data to client ASAP, no need to wait for request from client
  - greatly reduces background comms; instead of client constantly requesting new data, server only pushes when data changes

## Build instructions

1. Install go `1.24` or above
2. Copy this repository locally
3. `cd` into repo folder
4. `go mod tidy`
5. `go build`
6. compiled executable will be `autotaskViewer` (*nix / mac) or `autotaskViewer.exe` (windows)

## Usage instructions

1. launch executable file. You may need to `chmod +x ./autotaskViewer` on *nix/mac
2. browse to [http://localhost:8880](http://localhost:8880) (port will be different if launched with `-port` flag)
3. on first run you will have to provide API secrets to the server, as well as a password for encrypting secrets
4. on subsequent runs you will have to provide the password to decrypt secrets
5. after providing / unlocking secrets, page will display unassigned tickets
6. page automatically updates as soon as server detects a change in the list of open tickets
7. server will not poll API outside of active hours

### Runtime flags

This project parses flags in order to easily adjust server parameters. They are as follows:

- `loghttp`
  - Enables http request logging (default: false)
- `pollrate`
  - Sets interval (in seconds) at which API is queried (default: 30)
- `port`
  - Webserver's listening port (default: 8880)
- `filepath`
  - Relative path of encrypted secrets file (default: "secrets.gob")
- `verboseapi`
  - Prints verbose output of functions related to API calls (default: false)

#### Runtime flags examples

- `./autotaskViewer -loghttp -pollrate 60`
  - launches with http logging enabled and API polling set to occur every 60 seconds
- `./autotaskViewer -port 80 -filepath "newSecrets.gob" -verboseapi`
  - launches with server listening on port 80. Will load secrets from / save secrets to "newSecrets.gob". Prints verbose messages when API functions run.

## Technical explanations

While nothing in this project uses novel techniques, some of the strategies employed are worth explaining

### Determining state change of open tickets

The server will only send messages to websocket clients when it detects that the list of open tickets has changed.

There are various methodologies one could use to determine if this has occurred, but this project does the following:

1. When api is queried, ticket titles are sorted, then concatenated into a single string
2. This string is then hashed
3. Hash of sorted ticket titles is then stored in the `tickets.TicketsCollection` value
4. When api is queried next, the new hash is compared to the old hash. Websocket broadcast only occurs if they differ

### Determining if new ticket has been received (client)

Since the server will only communicate when titles of open tickets change, the client can always assume that data has changed and render the new content.

In order to alert user of new ticket (via background blinking on web page), the page needs to have a way to determine if the ticket is new. Since tickets can become renamed, or unassigned after having an assigned resource, some care is needed in making this determination

The client decides if data includes a new ticket by following these steps:

1. No action is taken on first websockets message from server. First load will never trigger an alert
2. The newest creation date of open tickets is stored by JS as the 'last newest date'
3. When the server sends new data, the newest date is obtained from that data.
4. The 'newest date' of the new data is compared to the 'last newest date'
   - if the new data has any tickets newer than the 'last newest date' then it must include a new ticket. Alert is triggered, and the newest date stored as 'last newest date'
   - if the newest ticket in the new data is not newer than the 'last newest date' then there is no new ticket. No alert is triggered, and 'last newest date' stays the same

## Project structure

This project is laid out in the following way:

### Packages

- `package main`
  - implements runtime flags and flag checking
  - uses public functions from the `web` package to set up and run the project
- `package web`
  - primary `WebApp` type and associated methods / structs are defined here
  - all non-`main` packages are used by the `WebApp` type
  - defines webserver routes, file I/O for encrypted secrets, and template rendering
  - split into multiple files to facilitate management:
    - `webApp.go` defines the `WebApp`, public methods, and api polling methods, in addition to misc helpers
    - `routes.go` defines all standard http route handler methods
    - `webSockets.go` defines `wsClient` type, and websocket handler / websocket broadcast methods
- `package tickets`
  - data structures & methods for Autotask tickets
- `package secrets`
  - data structures & methods for managing api secrets / file encryption & decryption
- `package api`
  - implements API call to Autotask

### Other files / folders

- `templates/`
  - contains `.html` files that the server dynamically renders at time of request
- `static/`
  - contains static files (images) that are not dynamically rendered
- `secrets.gob`
  - encrypted go binary file where API secrets are stored
  - is not actually a valid `.gob` format; the `.gob` byte slice is encrypted and combined with nonce/salt before saving to disc

## Notes for production use

This project was designed to be run and accessed on the same machine.

This server is fast and light weight, and should easily be able to handle a large amount of concurrent connections. However, there are a few things that should be done if you are planning on hosting this for multiple users:

1. Use a reverse proxy
   - Allows for easier use of standard ports like `80` and `443`
   - Host rules can allow for multiple web servers to be successfully accessed while running on the same host machine
   - Greatly simplifies deployment of HTTPS certificates (via ACME or manual distribution)
   - Allows for useful middlewares to run between client and server
   - No changes to server code needed, all desired security features can be implemented in the reverse proxy
   - Likely more performant than manually implementing changes in the server
   - Traefik is a great choice
2. Set up HTTPS
   - Again, a reverse proxy is recommended, but either way HTTPS should be used
   - Server sends potentially sensitive ticket data to clients, and clients send very sensitive password / api details when setting up / unlocking secrets
     - If running only on localhost, no packets leave the computer so all outside actors remain unable to sniff the traffic
     - These sensitive details will be transmitted IN THE CLEAR if traffic leaves the local machine and HTTPS is not used
3. Set up common-sense middleware
   - Rate limiting
   - Disable or restrict access to the `/secrets` path
   - IP limiting
   - Anything else you find reasonable