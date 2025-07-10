package main

import (
	"autotaskViewer/web"
	"flag"
	"fmt"
	"time"
)

var version string

func main() {

	validateFlags()

	// default version string
	// overridden by github workflow release version
	if version == "" {
		version = "dev-" + time.Now().Format("15:04:05-Jan-02-2006")
	}

	w := web.NewWebApp(
		*logHttp,
		*pollRate,
		*port,
		*saveFilePath,
		*verboseApi,
		*apiStart,
		*apiEnd,
		version,
	)

	w.Start()

}

const defaultPollRate = 30
const defaultPort = 8880
const defaultApiStart = 6
const defaultApiEnd = 18

var logHttp = flag.Bool("loghttp", false, "Enable HTTP request logging")
var pollRate = flag.Int("pollrate", defaultPollRate, "API poll interval in seconds")
var port = flag.Int("port", defaultPort, "webserver listening port")
var saveFilePath = flag.String("filepath", "secrets.gob", "Relative filepath to save encrypted secrets")
var verboseApi = flag.Bool("verboseapi", false, "verbose API call info")
var apiStart = flag.Int("apistart", defaultApiStart, "hour (24hr format) to start API calls")
var apiEnd = flag.Int("apiend", defaultApiEnd, "hour (24hr format) to end API calls")

func validateFlags() {
	flag.Parse()

	if *port < 1 || *port > 65535 {
		fmt.Printf("Invalid port %d, using default port %d\n", *port, defaultPort)
		*port = defaultPort
	}

	if *pollRate < 1 || *pollRate > 7200 {
		fmt.Printf("Invalid pollrate %d\n    min allowed = 1, max allowed = 7200 (2 hours)\n    using default poll rate %d\n", *pollRate, defaultPollRate)
		*port = defaultPort
	}

	if *apiEnd <= *apiStart {
		fmt.Printf("Invalid active hours: end not after start\n    received start:%d, end:%d.\n    using defaults (%d-%d)\n", *apiStart, *apiEnd, defaultApiStart, defaultApiEnd)
		*apiStart = defaultApiStart
		*apiEnd = defaultApiEnd
	} else if *apiEnd > 23 || *apiEnd < 1 {
		fmt.Printf("Invalid active hours: end hour invalid\n    received start:%d, end:%d.\n    using defaults (%d-%d)\n", *apiStart, *apiEnd, defaultApiStart, defaultApiEnd)
		*apiStart = defaultApiStart
		*apiEnd = defaultApiEnd
	} else if *apiStart < 0 || *apiStart > 22 {
		fmt.Printf("Invalid active hours: start hour invalid\n    received start:%d, end:%d.\n    using defaults (%d-%d)\n", *apiStart, *apiEnd, defaultApiStart, defaultApiEnd)
		*apiStart = defaultApiStart
		*apiEnd = defaultApiEnd
	}
}
