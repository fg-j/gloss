package internal

import (
	"encoding/json"
	"fmt"
)

type Organization struct {
	Name string
}

// TODO change to Client and not APIClient
// TODO add ignore set of repos
func (o *Organization) GetRepos(client Client) ([]Repository, error) {
	body, err := client.Get(fmt.Sprintf("orgs/%s/repos", o.Name), "per_page=100")
	if err != nil {
		return nil, fmt.Errorf("getting org repos for %s: %s", o.Name, err)
	}

	repos := []Repository{}
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return nil, fmt.Errorf("getting org repos for %s: could not unmarshal response: '%s': %s", o.Name, string(body), err)
	}
	return repos, nil
}
