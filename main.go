package main

import (
	"flag"
	"os"
)

type Config struct {
	Organizations []string
	Server        string
	NumWorkers    int
}

type arrayFlags []string

func (f *arrayFlags) String() string {
	return "some string"
}

func (f *arrayFlags) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	config := Config{}
	config.NumWorkers = 1

	responseTimesCommand := flag.NewFlagSet("response-times", flag.ExitOnError)

	// response-times subcommand flag pointers
	var orgFlags arrayFlags
	responseTimesServer := responseTimesCommand.String("server", "https://api.github.com", "Server from which to collect response time data.")
	responseTimesCommand.Var(&orgFlags, "org", `org(s) from which to collect response time data. (Can be passed multiple times)`)

	if len(os.Args) < 2 {
		panic("Subcommand needed")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "response-times":
		responseTimesCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if responseTimesCommand.Parsed() {
		config.Server = *responseTimesServer
		config.Organizations = orgFlags

		CalculateResponseTimeMetric(config)
		os.Exit(0)
	}
	os.Exit(1)
}
