package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gloss/internal"

	"github.com/aclements/go-moremath/stats"
	"github.com/paketo-buildpacks/packit/chronos"
)

func CalculateFirstContactTimeMetric(config Config) error {
	start := time.Now()
	if os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Println("Please set GITHUB_TOKEN")
		os.Exit(1)
	}

	fmt.Println("creating client")
	apiClient := internal.NewAPIClient(config.Server)
	fmt.Println("getting org repos")
	in := getOrgReposChan(config.Organizations, apiClient)

	fmt.Printf("Running with %d workers...\nUse --workers to set.\n\n", config.NumWorkers)

	var responseTimes []float64
	var workers []<-chan internal.TimeContainer
	for i := 0; i < config.NumWorkers; i++ {
		workers = append(workers, worker(i, apiClient, in))
	}

	for timeContainer := range merge(workers...) {
		if err := timeContainer.Error; err != nil {
			fmt.Printf("failed to calculate merge times: %s\n", err)
			os.Exit(1)
		}
		responseTimes = append(responseTimes, timeContainer.Time)
	}
	responseTimesSample := stats.Sample{Xs: responseTimes}
	fmt.Printf("\nResponse Time Stats\nFor %d issues/PRs\n    Average: %f days\n    Median %f days\n    95th Percentile: %f days\n",
		len(responseTimesSample.Xs),
		(responseTimesSample.Mean() / (60 * 24)),
		(responseTimesSample.Quantile(0.5) / (60 * 24)),
		(responseTimesSample.Quantile(0.95) / (60 * 24)))

	duration := time.Since(start)
	fmt.Printf("Execution took %f seconds.\n", duration.Seconds())

	return nil
}

func worker(id int, client internal.Client, input <-chan internal.RepositoryContainer) chan internal.TimeContainer {
	output := make(chan internal.TimeContainer)

	go func() {
		for repo := range input {
			fmt.Printf("Repository: %s\n\n", repo.Repository.Name)
			if repo.Error != nil {
				output <- internal.TimeContainer{Error: repo.Error}
				close(output)
			}
			repo.Repository.GetFirstContactTimes(client, chronos.DefaultClock, output)
			fmt.Println("")
		}
		close(output)
	}()
	return output
}

func merge(ws ...<-chan internal.TimeContainer) chan internal.TimeContainer {
	var wg sync.WaitGroup
	output := make(chan internal.TimeContainer)

	getTimes := func(c <-chan internal.TimeContainer) {
		for timeContainer := range c {
			output <- timeContainer
		}
		wg.Done()
	}
	wg.Add(len(ws))
	for _, w := range ws {
		go getTimes(w)
	}
	go func() {
		wg.Wait()
		close(output)
	}()
	return output
}

func getOrgReposChan(orgNames []string, client internal.Client) chan internal.RepositoryContainer {
	output := make(chan internal.RepositoryContainer)
	go func() {
		for _, orgName := range orgNames {
			org := internal.Organization{Name: orgName}
			repos, err := org.GetRepos(client)
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
