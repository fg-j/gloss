package internal

import (
	"encoding/json"
	"fmt"
	"time"
)

type RepositoryContainer struct {
	Repository Repository
	Error      error
}

type Repository struct {
	Name  string `json:"full_name"`
	URL   string `json:"url"`
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
}

//go:generate faux --interface Clock --output fakes/clock.go
type Clock interface {
	Now() time.Time
}

func (r *Repository) GetRecentIssues(client Client, clock Clock) ([]Issue, error) {
	timeString := clock.Now().UTC().Add(-30 * 24 * time.Hour).Format(time.RFC3339)

	body, err := client.Get(fmt.Sprintf("/repos/%s/issues", r.Name),
		"state=all",
		"per_page=100",
		fmt.Sprintf("since=%s", timeString))
	if err != nil {
		return nil, fmt.Errorf("getting recent issues: %s", err)
	}

	issues := []Issue{}
	err = json.Unmarshal(body, &issues)
	if err != nil {
		return nil, fmt.Errorf("getting recent issues: could not unmarshal JSON '%s' : %s", string(body), err)
	}

	return issues, nil
}
