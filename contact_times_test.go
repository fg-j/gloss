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

func testContactTimes(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		mockGithubServer    *httptest.Server
		mockGithubServerURI string
		gloss               string
		err                 error
	)

	it.Before(func() {

		mockGithubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)
				return
			}

			switch req.URL.Path {
			default:
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

	context("given a valid auth token and org are provided", func() {
		it.Before(func() {
			os.Setenv("GITHUB_TOKEN", "some-token")
		})

		it("correctly prints average, median, 95th percentile first contact times on issues", func() {
			command := exec.Command(gloss, "first-contact-times", "--server", mockGithubServer.URL, "--org example-org")
			fmt.Println(mockGithubServerURI)
			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)

			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			out := string(buffer.Contents())

			Expect(out).To(ContainLines(
				`example-org/example-repo #1 received response from example-maintainer in 10.0000 minutes`,
			))
		})
	})
}
