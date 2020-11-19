package fakes

import (
	"net/http"
	"sync"
)

type HTTPClient struct {
	DoCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Req *http.Request
		}
		Returns struct {
			Response *http.Response
			Error    error
		}
		Stub func(*http.Request) (*http.Response, error)
	}
}

func (f *HTTPClient) Do(param1 *http.Request) (*http.Response, error) {
	f.DoCall.Lock()
	defer f.DoCall.Unlock()
	f.DoCall.CallCount++
	f.DoCall.Receives.Req = param1
	if f.DoCall.Stub != nil {
		return f.DoCall.Stub(param1)
	}
	return f.DoCall.Returns.Response, f.DoCall.Returns.Error
}
