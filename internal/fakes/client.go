package fakes

import "sync"

type Client struct {
	GetCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path   string
			Params []string
		}
		Returns struct {
			ByteSlice []byte
			Error     error
		}
		Stub func(string, ...string) ([]byte, error)
	}
}

func (f *Client) Get(param1 string, param2 ...string) ([]byte, error) {
	f.GetCall.Lock()
	defer f.GetCall.Unlock()
	f.GetCall.CallCount++
	f.GetCall.Receives.Path = param1
	f.GetCall.Receives.Params = param2
	if f.GetCall.Stub != nil {
		return f.GetCall.Stub(param1, param2...)
	}
	return f.GetCall.Returns.ByteSlice, f.GetCall.Returns.Error
}
