package main

import (
	"fmt"
	"os"
	"time"

	"gloss/internal"

	"github.com/aclements/go-moremath/stats"
)

func CalculateFirstContactTimeMetric(config Config) error {
	start := time.Now()
	if os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Println("Please set GITHUB_TOKEN")
		os.Exit(1)

		in := getOrgReposChan(config.Organizations, config.Server)

		fmt.Printf("Running with %d workers...\nUse --workers to set.\n\n", config.NumWorkers)

		var responseTimes []float64
		var workers []<-chan internal.TimeContainer
		for i := 0; i < config.NumWorkers; i++ {
			workers = append(workers, worker(i, config.Server, in))
		}

		for timeContainer := range merge(workers...) {
			if err := timeContainer.Error; err != nil {
				fmt.Printf("failed to calculate merge times: %s\n", err)
				os.Exit(1)
			}
			responseTimes = append(responseTimes, timeContainer.Time)
		}
		responseTimesSample := stats.Sample{Xs: responseTimes}
		fmt.Printf("\nMerge Time Stats\nFor %d pull requests\n    Average: %f hours\n    Median %f hours\n    95th Percentile: %f hours\n",
			len(responseTimesSample.Xs),
			(responseTimesSample.Mean() / 60),
			(responseTimesSample.Quantile(0.5) / 60),
			(responseTimesSample.Quantile(0.95) / 60))

		duration := time.Since(start)
		fmt.Printf("Execution took %f seconds.\n", duration.Seconds())
	}

	return nil
}

func worker(id int, client internal.Client, input <-chan internal.RepositoryContainer) chan internal.TimeContainer {
	output := make(chan internal.TimeContainer)

	go func() {
		for repo := range input {
			if repo.Error != nil {
				output <- internal.TimeContainer{Error : repo.Error}
				close(output)
			}
			repo.Repository.GetFirstContactTimes(client, clock, output)
		}
	}
return output
}

func merge(...<-chan internal.TimeContainer) chan internal.TimeContainer{
return nil
}

func getOrgReposChan(orgs []string, serverURI string) chan internal.RepositoryContainer {
	output := make(chan internal.RepositoryContainer)
	go func() {
		for _, org := range orgs {
			repos, err := internal.GetOrgRepos(org, serverURI)
			if err != nil {
				output <- internal.RepositoryContainer{Error: fmt.Errorf("failed to get repositories: %s", err)}
			}
			for _, repo := range repos {
				repoContainer := internal.RepositoryContainer{
					Repository: repo,
					Error:      nil,
				}
				output <- repoContainer
			}
		}
		close(output)
	}()
	return output
}
