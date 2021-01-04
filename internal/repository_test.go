package internal_test

import (
	"fmt"
	"testing"
	"time"

	. "gloss/internal"
	"gloss/internal/fakes"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testRepository(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect
	var repo Repository
	var apiClient = &fakes.Client{}
	var clock = &fakes.Clock{}

	repo = Repository{
		Name: "example-org/example-repo",
	}
	context("GetRecentIssues", func() {
		it.Before(func() {
			clock.NowCall.Returns.Time = time.Date(2001, time.January, 1, 20, 20, 20, 0, time.UTC).Add(30 * 24 * time.Hour)
			apiClient.GetCall.Returns.ByteSlice = []byte(`[
{
	"created_at" : "2001-01-01T20:20:20Z",
	"comments" : 1,
	"comments_url" : "test-url.com"
}]`)
		})

		it("returns the issues from the repo", func() {
			issues, err := repo.GetRecentIssues(apiClient, clock)
			Expect(err).NotTo(HaveOccurred())
			Expect(apiClient.GetCall.Receives.Path).To(Equal("/repos/example-org/example-repo/issues"))
			Expect(apiClient.GetCall.Receives.Params).To(ContainElement("per_page=100"))
			Expect(apiClient.GetCall.Receives.Params).To(ContainElement("since=2001-01-01T20:20:20Z"))

			testIssue := Issue{
				CreatedAt:   "2001-01-01T20:20:20Z",
				NumComments: 1,
				CommentsURL: "test-url.com",
			}
			Expect(issues).To(ContainElement(testIssue))
		})

		context("failure cases", func() {
			context("when get request fails", func() {

				it.Before(func() {
					apiClient.GetCall.Returns.Error = fmt.Errorf("something went wrong with HTTP GET")
				})
				it("returns the error", func() {
					_, err := repo.GetRecentIssues(apiClient, clock)
					Expect(err).To(MatchError("getting recent issues: something went wrong with HTTP GET"))
				})
			})

			context("when JSON cannot be unmarshalled into object", func() {

				it.Before(func() {
					apiClient.GetCall.Returns.ByteSlice = []byte("{invalidJSON")
				})
				it("returns the error", func() {
					_, err := repo.GetRecentIssues(apiClient, clock)
					Expect(err).To(MatchError("getting recent issues: could not unmarshal JSON '{invalidJSON' : invalid character 'i' looking for beginning of object key string"))
				})
			})
		})
	})
}
