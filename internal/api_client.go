package internal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

//go:generate faux --interface HTTPClient --output fakes/http_client.go
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

//go:generate faux --interface Client --output fakes/client.go
type Client interface {
	Get(path string, params ...string) ([]byte, error)
}

type APIClient struct {
	ServerURL string
	client    HTTPClient
}

func NewAPIClient(serverURL string, httpClient HTTPClient) APIClient {
	return APIClient{ServerURL: serverURL,
		client: httpClient}
}

func (c *APIClient) Get(path string, params ...string) ([]byte, error) {

	uri, err := url.Parse(c.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse server URL: %s", err)
	}

	uri.Path = path
	if len(params) > 0 {
		uri.RawQuery = params[0]
		for i := range params {
			if i != 0 {
				uri.RawQuery = fmt.Sprintf("%s&%s", uri.RawQuery, params[i])
			}
		}
	}

	request, _ := http.NewRequest("GET", uri.String(), nil)
	request.Header.Add("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("client couldn't make HTTP request: %s", err)
	}

	return ioutil.ReadAll(response.Body)
}
