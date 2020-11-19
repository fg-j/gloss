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
	var Eventually = NewWithT(t).Eventually
	var Consistently = NewWithT(t).Consistently
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

	context("GetFirstContactTimes", func() {
		var timeChan chan TimeContainer
		var issues []CommentGetter
		context("when given a set of issues", func() {
			it.Before(func() {
				issue := &fakes.CommentGetter{}
				issue.GetFirstReplyCall.Returns.Comment = Comment{CreatedAt: "2001-01-01T21:20:20Z"}
				issue.GetCreatedAtCall.Returns.String = "2001-01-01T20:20:20Z"

				issues = []CommentGetter{issue}
			})
			it("writes the first reply time for an issue to the output channel", func() {
				timeChan = make(chan TimeContainer)
				go repo.GetFirstContactTimes(apiClient, issues, clock, timeChan)

				Eventually(<-timeChan).Should(Equal(TimeContainer{Time: 60, Error: nil}))
			})
		})
		context("when an issue has been opened by a bot", func() {
			it.Before(func() {
				realIssue := &fakes.CommentGetter{}
				realIssue.GetFirstReplyCall.Returns.Comment = Comment{CreatedAt: "2001-01-01T21:20:20Z"}
				realIssue.GetCreatedAtCall.Returns.String = "2001-01-01T20:20:20Z"

				botIssue := &fakes.CommentGetter{}
				botIssue.GetFirstReplyCall.Returns.Comment = Comment{CreatedAt: "2001-01-01T20:21:20Z"}
				botIssue.GetCreatedAtCall.Returns.String = "2001-01-01T20:20:20Z"
				botIssue.GetUserLoginCall.Returns.String = "paketo-bot"

				issues = []CommentGetter{realIssue, botIssue}
			})

			it("does not include reply time for the bot issue", func() {
				timeChan = make(chan TimeContainer)
				go repo.GetFirstContactTimes(apiClient, issues, clock, timeChan)

				Eventually(<-timeChan).Should(Equal(TimeContainer{Time: 60, Error: nil}))
				Consistently(<-timeChan).ShouldNot(Equal(TimeContainer{Time: 1, Error: nil}))
			})
		})

		context("when an issue has no reply", func() {
			it.Before(func() {
				issue := &fakes.CommentGetter{}
				issue.GetFirstReplyCall.Returns.Comment = Comment{}
				issue.GetCreatedAtCall.Returns.String = "2001-01-01T20:20:20Z"
				clock.NowCall.Returns.Time = time.Date(2001, time.January, 1, 20, 20, 20, 0, time.UTC).Add(1 * time.Hour)

				issues = []CommentGetter{issue}
			})
			it("returns the time between run time and issue opening", func() {
				timeChan = make(chan TimeContainer)
				go repo.GetFirstContactTimes(apiClient, issues, clock, timeChan)

				Eventually(<-timeChan).Should(Equal(TimeContainer{Time: 60, Error: nil}))
			})
		})
		context("failure cases", func() {
			context("when there is an error getting the first reply from an issue", func() {
				it.Before(func() {
					issue := &fakes.CommentGetter{}
					issue.GetFirstReplyCall.Returns.Error = fmt.Errorf("some problem getting reply")
					issues = []CommentGetter{issue}
				})

				it("sends the error in a container in the channel", func() {
					timeChan = make(chan TimeContainer)
					go repo.GetFirstContactTimes(apiClient, issues, clock, timeChan)

					Eventually((<-timeChan).Error).Should(MatchError(fmt.Errorf("could not get first reply: some problem getting reply")))

				})
			})

			context("when there is an error parsing the first comment creation time", func() {
				it.Before(func() {
					issue := &fakes.CommentGetter{}
					issue.GetFirstReplyCall.Returns.Comment = Comment{CreatedAt: "some-garbage"}
					issues = []CommentGetter{issue}
				})

				it("sends the error in a container in the channel", func() {
					timeChan = make(chan TimeContainer)
					go repo.GetFirstContactTimes(apiClient, issues, clock, timeChan)

					Eventually((<-timeChan).Error).Should(MatchError(fmt.Errorf(`could not parse first reply time: parsing time "some-garbage" as "2006-01-02T15:04:05Z07:00": cannot parse "some-garbage" as "2006"`)))

				})
			})

			context("when there is an error parsing the issue's creation time", func() {
				it.Before(func() {
					issue := &fakes.CommentGetter{}
					issue.GetFirstReplyCall.Returns.Comment = Comment{}
					issue.GetCreatedAtCall.Returns.String = "some-garbage"
					issues = []CommentGetter{issue}
				})

				it("sends the error in a container in the channel", func() {
					timeChan = make(chan TimeContainer)
					go repo.GetFirstContactTimes(apiClient, issues, clock, timeChan)

					Eventually((<-timeChan).Error).Should(MatchError(fmt.Errorf(`could not parse issue creation time: parsing time "some-garbage" as "2006-01-02T15:04:05Z07:00": cannot parse "some-garbage" as "2006"`)))
				})
			})
		})
	})
}
