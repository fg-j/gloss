package internal_test

import (
	"bytes"
	"fmt"
	. "gloss/internal"
	"gloss/internal/fakes"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testAPIClient(t *testing.T, context spec.G, it spec.S) {
	var apiClient APIClient
	var httpClient *fakes.HTTPClient
	var Expect = NewWithT(t).Expect

	httpClient = &fakes.HTTPClient{}

	apiClient = NewAPIClient("https://test-server.com", httpClient)

	context("Get", func() {
		//TODO: Add test that auth token is added
		context("when an endpoint is provided", func() {
			it.Before(func() {
				doBody := ioutil.NopCloser(bytes.NewReader([]byte("some body")))
				httpClient.DoCall.Returns.Response = &http.Response{StatusCode: 200, Body: doBody}
			})
			it("makes an HTTP request to the provided endpoint", func() {
				_, err := apiClient.Get("/my/test/endpoint")

				Expect(err).NotTo(HaveOccurred())
				Expect(httpClient.DoCall.Receives.Req.URL.Host).To(Equal("test-server.com"))
				Expect(httpClient.DoCall.Receives.Req.URL.Scheme).To(Equal("https"))
				Expect(httpClient.DoCall.Receives.Req.URL.Path).To(Equal("/my/test/endpoint"))
			})

			it("returns the httpClient's response", func() {
				body, _ := apiClient.Get("/my/test/endpoint")

				Expect(string(body)).To(Equal("some body"))
			})
		})

		context("when params are provided", func() {
			it.Before(func() {
				doBody := ioutil.NopCloser(bytes.NewReader([]byte("some body")))
				httpClient.DoCall.Returns.Response = &http.Response{StatusCode: 200, Body: doBody}
			})

			it("makes an HTTP request with those params", func() {
				_, err := apiClient.Get("/my/test/endpoint", "per_page=100", "state=open")

				Expect(err).NotTo(HaveOccurred())
				Expect(httpClient.DoCall.Receives.Req.URL.RawQuery).To(Equal("per_page=100&state=open"))
			})
		})

		context("failure cases", func() {
			context("when server URL cannot be parsed", func() {
				it.Before(func() {
					apiClient.ServerURL = "some-garbage\n"
				})

				it("returns the error", func() {
					_, err := apiClient.Get("")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(`could not parse server URL: parse "some-garbage\n": net/url: invalid control character in URL`))
				})
			})

			context("when client fails to make HTTP request", func() {
				it.Before(func() {
					httpClient.DoCall.Returns.Error = fmt.Errorf("something failed")
				})
				it("returns the error", func() {
					_, err := apiClient.Get("/my/endpoint")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("client couldn't make HTTP request: something failed"))
				})
			})
		})
	})
}
