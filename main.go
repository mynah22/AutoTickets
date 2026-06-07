package main

import (
	"AutoTickets/web"
	"flag"
	"fmt"
	"os"
	"strconv"
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

const envPrefix = "AUTOTICKETS_"

func validateFlags() {
	flag.Parse()

	setFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	if !setFlags["loghttp"] {
		*logHttp = getEnvBool("LOGHTTP", *logHttp)
	}
	if !setFlags["pollrate"] {
		*pollRate = getEnvInt("POLL_RATE", *pollRate)
	}
	if !setFlags["port"] {
		*port = getEnvInt("PORT", *port)
	}
	if !setFlags["filepath"] {
		*saveFilePath = getEnvString("FILEPATH", *saveFilePath)
	}
	if !setFlags["verboseapi"] {
		*verboseApi = getEnvBool("VERBOSEAPI", *verboseApi)
	}
	if !setFlags["apistart"] {
		*apiStart = getEnvInt("API_START", *apiStart)
	}
	if !setFlags["apiend"] {
		*apiEnd = getEnvInt("API_END", *apiEnd)
	}

	if *port < 1 || *port > 65535 {
		fmt.Printf("Invalid port %d, using default port %d\n", *port, defaultPort)
		*port = defaultPort
	}

	if *pollRate < 1 || *pollRate > 7200 {
		fmt.Printf("Invalid pollrate %d\n    min allowed = 1, max allowed = 7200 (2 hours)\n    using default poll rate %d\n", *pollRate, defaultPollRate)
		*pollRate = defaultPollRate
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

func getEnvString(name, defaultValue string) string {
	if value := os.Getenv(envPrefix + name); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(name string, defaultValue int) int {
	if value := os.Getenv(envPrefix + name); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("Invalid %s=%q, using %d\n", envPrefix+name, value, defaultValue)
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}

func getEnvBool(name string, defaultValue bool) bool {
	if value := os.Getenv(envPrefix + name); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			fmt.Printf("Invalid %s=%q, using %t\n", envPrefix+name, value, defaultValue)
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}
