package main

type Config struct {
	Organizations []string
	Server        string
	NumWorkers    int
}

func main() {
	config := Config{}
	config.Organizations = []string{"paketo-buildpacks"}
	config.Server = "https://api.github.com"
	config.NumWorkers = 1
	CalculateFirstContactTimeMetric(config)
}
