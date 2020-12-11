package internal_test

import (
	"fmt"
	"gloss/internal"
	"gloss/internal/fakes"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testOrganization(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect
	var client = &fakes.Client{}
	var org = internal.Organization{Name: "example-org"}

	context("when there are repos associated with the org", func() {
		it.Before(func() {
			client.GetCall.Returns.ByteSlice = []byte(`
[{
	"name": "example-repo-one",
  "full_name": "example-org/example-repo-one",
  "owner": {
    "login": "example-org",
    "type": "Organization"
	},
	"url": "https://api.github.com/repos/example-org/example-repo-one"
	},
	{
	"name": "example-repo-two",
  "full_name": "example-org/example-repo-two",
  "owner": {
    "login": "example-org",
    "type": "Organization"
	},
	"url": "https://api.github.com/repos/example-org/example-repo-two"
	}
]
`)
			// TODO: dot import internal packages in tests
		})
		it("returns an array of the repos", func() {
			repos, err := org.GetRepos(client)

			repoOne := internal.Repository{
				Name: "example-org/example-repo-one",
				URL:  "https://api.github.com/repos/example-org/example-repo-one",
			}
			repoOne.Owner.Login = "example-org"

			repoTwo := internal.Repository{
				Name: "example-org/example-repo-two",
				URL:  "https://api.github.com/repos/example-org/example-repo-two",
			}
			repoTwo.Owner.Login = "example-org"

			expectedRepos := []internal.Repository{repoOne, repoTwo}

			Expect(err).NotTo(HaveOccurred())
			Expect(repos).To(Equal(expectedRepos))

		})
	})
	context("failure cases", func() {
		context("when the org repos cannot be gotten from the server", func() {
			it.Before(func() {
				client.GetCall.Returns.Error = fmt.Errorf("some problem with the API")
			})
			it("returns the error", func() {
				_, err := org.GetRepos(client)
				Expect(err).To(MatchError("getting org repos for example-org: some problem with the API"))
			})
		})
		//TODO: make JSON unmarshal error messages consistent OR factor out JSON unmarshalling
		context("when the response JSON cannot be unmarshalled", func() {
			it.Before(func() {
				client.GetCall.Returns.ByteSlice = []byte("\nsome-garbage\n")
			})

			it("returns the error", func() {
				_, err := org.GetRepos(client)
				Expect(err).To(MatchError("getting org repos for example-org: could not unmarshal response: '\nsome-garbage\n': invalid character 's' looking for beginning of value"))
			})
		})
	})
}
