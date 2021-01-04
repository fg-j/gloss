package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/sclevine/spec"
)

const exampleOrgReposResponse string = `
[
{
"full_name" : "example-org/example-repo",
"owner" : {
		"login": "example-org"
	},
"url" : "http://api.example.com/repos/example-org/example-repo"
}
]
`

const otherExampleOrgReposResponse string = `
[
{
"full_name" : "other-example-org/example-repo",
"owner" : {
		"login": "other-example-org"
	},
"url" : "http://api.example.com/repos/other-example-org/example-repo"
}
]
`

const issuesResponse string = `
[
{
"created_at" : "2021-01-01T00:00:00Z",
"comments" : 1,
"comments_url" : "http://api.example.com/repos/example-org/example-repo/issues/1/comments",
"number" : 1,
"user" : {
	"login" : "example-user"
	}
}
]
`

const commentsResponse string = `
[
{
"user" : {
	"login" : "example-maintainer",
	"type" : "user"
	},
"created_at" : "2021-01-01T00:10:00Z"
}
]
`

const notFoundResponse string = `
{
  "message": "Not Found"
}
`

func testResponseTimes(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		mockGithubServer *httptest.Server
		gloss            string
		err              error
	)

	it.Before(func() {
		mockGithubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)
				return
			}

			switch req.URL.Path {
			case "/orgs/example-org/repos":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, exampleOrgReposResponse)
			case "/orgs/other-example-org/repos":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, otherExampleOrgReposResponse)
			case "/orgs/invalid-org/repos":
				fmt.Fprintln(w, notFoundResponse)
			case "/repos/example-org/example-repo/issues":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, issuesResponse)
			case "/repos/other-example-org/example-repo/issues":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, issuesResponse)
			case "/repos/example-org/example-repo/issues/1/comments":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, commentsResponse)
			case "/repos/other-example-org/example-repo/issues/1/comments":
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, commentsResponse)
			default:
				fmt.Fprintln(w, "unknown path")
				t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
			}
		}))

		gloss, err = gexec.Build("gloss")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		mockGithubServer.Close()
		gexec.CleanupBuildArtifacts()
	})

	context("given an auth token is provided", func() {
		it.Before(func() {
			os.Setenv("GITHUB_TOKEN", "sometoken")
		})

		context("and the provided org exists", func() {
			it("correctly prints average, median, 95th percentile first response times on issues", func() {
				command := exec.Command(gloss, "response-times", "--server", mockGithubServer.URL, "--org", "example-org")
				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

				out := string(buffer.Contents())

				Expect(out).To(ContainLines(
					`example-org/example-repo #1 by example-user received response from example-maintainer in 10.000000 minutes`,
				))
				Expect(out).To(ContainLines(
					`For 1 issues/PRs`,
					`    Average: 0.006944 days`,
					`    Median 0.006944 days`,
					`    95th Percentile: 0.006944 days`,
				))
			})
		})

		context("and the provided multiple orgs exist", func() {
			it("correctly prints average, median, 95th percentile first response times on issues", func() {
				command := exec.Command(gloss, "response-times", "--server", mockGithubServer.URL, "--org", "example-org", "--org", "other-example-org")
				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

				out := string(buffer.Contents())

				Expect(out).To(ContainLines(
					`example-org/example-repo #1 by example-user received response from example-maintainer in 10.000000 minutes`,
				))
				Expect(out).To(ContainLines(
					`other-example-org/example-repo #1 by example-user received response from example-maintainer in 10.000000 minutes`,
				))
				Expect(out).To(ContainLines(
					`For 2 issues/PRs`,
					`    Average: 0.006944 days`,
					`    Median 0.006944 days`,
					`    95th Percentile: 0.006944 days`,
				))
			})
		})

		context("and the provided org does not exist", func() {
			it("fails with an informative error", func() {
				command := exec.Command(gloss, "response-times", "--server", mockGithubServer.URL, "--org", "invalid-org")
				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })

				out := string(buffer.Contents())

				Expect(out).To(ContainLines(
					`failed to calculate merge times: failed to get repositories: getting org repos for invalid-org: could not unmarshal response: '`,
				))
			})
		})
	})

	context("given no auth token is provided", func() {
		it.Before(func() {
			os.Setenv("GITHUB_TOKEN", "")
		})

		it("exits saying that an auth token is required", func() {
			command := exec.Command(gloss, "response-times", "--server", mockGithubServer.URL, "--org", "invalid-org")
			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })

			out := string(buffer.Contents())

			Expect(out).To(ContainLines(
				`Please set GITHUB_TOKEN`,
			))
		})
	})
}
