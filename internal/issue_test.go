package internal_test

import (
	"fmt"
	"gloss/internal"
	"gloss/internal/fakes"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testIssue(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect
	var issue internal.Issue
	var client = &fakes.Client{}
	it.Before(func() {
	})
	context("GetFirstReply", func() {
		context("when there are no replies on an issue", func() {
			it.Before(func() {
				issue = internal.Issue{
					CreatedAt:   "some-time",
					NumComments: 0,
				}
			})

			it("returns an empty comment and no error", func() {
				reply, err := issue.GetFirstReply(client)

				Expect(err).NotTo(HaveOccurred())
				Expect(reply).To(Equal(internal.Comment{}))
			})
		})
		context("when there is one reply on an issue from someone other than the author", func() {
			it.Before(func() {
				issue = internal.Issue{
					NumComments: 1,
					CommentsURL: "www.example.com",
				}
				issue.User.Login = "originalPoster"

				client.GetCall.Returns.ByteSlice = []byte(`
[
  {
    "user": {
      "login": "replyGuy",
      "type": "User"
    },
		"created_at": "2001-01-01T00:00:00Z"
  }
]
`)
			})

			it("returns the reply", func() {
				reply, err := issue.GetFirstReply(client)

				Expect(err).NotTo(HaveOccurred())
				expected := internal.Comment{}
				expected.User.Login = "replyGuy"
				expected.User.Type = "User"
				expected.CreatedAt = "2001-01-01T00:00:00Z"
				Expect(reply).To(Equal(expected))
			})
		})

		context("when the reply on an issue is from the OP", func() {
			it.Before(func() {
				issue = internal.Issue{
					NumComments: 1,
					CommentsURL: "www.example.com",
				}
				issue.User.Login = "originalPoster"

				client.GetCall.Returns.ByteSlice = []byte(`
[
  {
    "user": {
      "login": "originalPoster",
      "type": "User"
    },
		"created_at": "2001-01-01T00:00:00Z"
  }
]
`)
			})

			it("does not return the reply", func() {
				reply, err := issue.GetFirstReply(client)

				Expect(err).NotTo(HaveOccurred())
				Expect(reply).To(Equal(internal.Comment{}))
			})
		})

		context("when the reply on an issue is from an ignored user account", func() {
			it.Before(func() {
				issue = internal.Issue{
					NumComments: 1,
					CommentsURL: "www.example.com",
				}
				issue.User.Login = "originalPoster"

				client.GetCall.Returns.ByteSlice = []byte(`
[
  {
    "user": {
      "login": "ignoredUser",
      "type": "User"
    },
		"created_at": "2001-01-01T00:00:00Z"
  }
]
`)
			})

			it("does not return the reply", func() {
				reply, err := issue.GetFirstReply(client, "ignoredUser")

				Expect(err).NotTo(HaveOccurred())
				Expect(reply).To(Equal(internal.Comment{}))
			})
		})
		context("failure cases", func() {
			context("when the comment URL cannot be parsed", func() {
				it.Before(func() {
					issue = internal.Issue{
						NumComments: 1,
						CommentsURL: "some-garbage\n",
					}
				})
				it("returns the error", func() {
					_, err := issue.GetFirstReply(client)
					Expect(err).To(MatchError(`parsing comments url: parse "some-garbage\n": net/url: invalid control character in URL`))
				})
			})

			context("when the comments get request fails", func() {
				it.Before(func() {
					issue = internal.Issue{
						NumComments: 1,
						CommentsURL: "www.example.com",
					}
					client.GetCall.Returns.Error = fmt.Errorf("some http GET issue")
				})
				it("returns the error", func() {
					_, err := issue.GetFirstReply(client)
					Expect(err).To(MatchError("getting issue comments: some http GET issue"))
				})
			})

			context("when the comments JSON cannot be unmarshalled", func() {
				it.Before(func() {
					issue = internal.Issue{
						NumComments: 1,
						CommentsURL: "www.example.com",
					}
					client.GetCall.Returns.ByteSlice = []byte("[[")
				})
				it("returns the error", func() {
					_, err := issue.GetFirstReply(client)
					Expect(err).To(MatchError("getting issue comments: could not unmarshal JSON '[[' : unexpected end of JSON input"))
				})
			})
		})
	})

	context("GetFirstContactTime", func() {
		var clock = &fakes.Clock{}
		context("when there is a reply on an issue", func() {
			it.Before(func() {
				issue = internal.Issue{
					NumComments: 1,
					CommentsURL: "www.example.com",
					CreatedAt:   "2001-01-01T00:00:00Z",
				}
				issue.User.Login = "originalPoster"

				client.GetCall.Returns.ByteSlice = []byte(`
[
  {
    "user": {
      "login": "replyGuy",
      "type": "User"
    },
		"created_at": "2001-01-01T00:10:00Z"
  }
]
`)
			})

			it("returns the difference between the issue creation time and reply time in minutes", func() {
				contactTime, err := issue.GetFirstContactTime(client, clock)

				Expect(err).NotTo(HaveOccurred())
				Expect(contactTime).To(Equal(float64(10)))
			})
		})

		context("when there are no replies on an issue", func() {
			it.Before(func() {
				issue = internal.Issue{
					CreatedAt:   "2001-01-01T00:00:00Z",
					NumComments: 0,
				}
				clock.NowCall.Returns.Time = time.Date(2001, time.January, 1, 1, 0, 0, 0, time.UTC)
			})

			it("returns the difference between the issue creation and the current time in minutes", func() {
				contactTime, err := issue.GetFirstContactTime(client, clock)

				Expect(err).NotTo(HaveOccurred())
				Expect(contactTime).To(Equal(float64(60)))
			})
		})

		context("when the reply on an issue is from the OP", func() {
			it.Before(func() {
				issue = internal.Issue{
					NumComments: 1,
					CommentsURL: "www.example.com",
					CreatedAt:   "2001-01-01T00:00:00Z",
				}
				issue.User.Login = "originalPoster"

				client.GetCall.Returns.ByteSlice = []byte(`
[
  {
    "user": {
      "login": "originalPoster",
      "type": "User"
    },
		"created_at": "2001-01-01T00:00:00Z"
  }
]
`)
				clock.NowCall.Returns.Time = time.Date(2001, time.January, 1, 1, 0, 0, 0, time.UTC)
			})

			it("it returns the difference between the issue creation time and current time in minutes", func() {
				contactTime, err := issue.GetFirstContactTime(client, clock)

				Expect(err).NotTo(HaveOccurred())
				Expect(contactTime).To(Equal(float64(60)))
			})
		})

		context("when the reply on an issue is from an ignored user account", func() {
			it.Before(func() {
				issue = internal.Issue{
					NumComments: 1,
					CommentsURL: "www.example.com",
					CreatedAt:   "2001-01-01T00:00:00Z",
				}
				issue.User.Login = "originalPoster"

				client.GetCall.Returns.ByteSlice = []byte(`
[
  {
    "user": {
      "login": "ignoredUser",
      "type": "User"
    },
		"created_at": "2001-01-01T00:00:00Z"
  }
]
`)
				clock.NowCall.Returns.Time = time.Date(2001, time.January, 1, 1, 0, 0, 0, time.UTC)
			})

			it("it returns the difference between the issue creation time and current time in minutes", func() {
				contactTime, err := issue.GetFirstContactTime(client, clock, "ignoredUser")

				Expect(err).NotTo(HaveOccurred())
				Expect(contactTime).To(Equal(float64(60)))
			})
		})

		context("failure cases", func() {
			context.Pend("when there is an error getting the first reply from an issue", func() {
				it.Before(func() {
				})

				it("sends the error in a container in the channel", func() {
				})
			})

			context.Pend("when there is an error parsing the first comment creation time", func() {
				it.Before(func() {
				})

				it("sends the error in a container in the channel", func() {
				})
			})

			context.Pend("when there is an error parsing the issue's creation time", func() {
				it.Before(func() {
				})

				it("sends the error in a container in the channel", func() {
				})
			})
		})
	})

	context("GetCreatedAt", func() {
		context("when the issue has a CreatedAt field", func() {
			it.Before(func() {
				issue = internal.Issue{
					CreatedAt: "some-time",
				}
			})

			it("returns the value of the issue's CreatedAt field", func() {
				createdAt := issue.GetCreatedAt()

				Expect(createdAt).To(Equal("some-time"))
			})
		})

		context("when the issue has an unset CreatedAt field", func() {
			it.Before(func() {
				issue = internal.Issue{}
			})

			it("returns the an empty string", func() {
				createdAt := issue.GetCreatedAt()

				Expect(createdAt).To(Equal(""))
			})
		})
	})

	context("GetUserLogin", func() {
		context("when the issue has a User.Login field", func() {
			it.Before(func() {
				issue = internal.Issue{}
				issue.User.Login = "some-user"
			})

			it("returns the value of the issue's User.Login field", func() {
				createdAt := issue.GetUserLogin()

				Expect(createdAt).To(Equal("some-user"))
			})
		})

		context("when the issue has an unset User.Login field", func() {
			it.Before(func() {
				issue = internal.Issue{}
			})

			it("returns an empty string", func() {
				createdAt := issue.GetUserLogin()

				Expect(createdAt).To(Equal(""))
			})
		})
	})
}
