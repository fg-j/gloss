package fakes

import (
	"gloss/internal"
	"sync"
)

type CommentGetter struct {
	GetCreatedAtCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			String string
		}
		Stub func() string
	}
	GetFirstReplyCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Client       internal.Client
			IgnoredUsers []string
		}
		Returns struct {
			Comment internal.Comment
			Error   error
		}
		Stub func(internal.Client, ...string) (internal.Comment, error)
	}
	GetFirstResponseTimeCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Client       internal.Client
			Clock        internal.Clock
			IgnoredUsers []string
		}
		Returns struct {
			Float64 float64
			Error   error
		}
		Stub func(internal.Client, internal.Clock, ...string) (float64, error)
	}
	GetNumberCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			Int int
		}
		Stub func() int
	}
	GetUserLoginCall struct {
		sync.Mutex
		CallCount int
		Returns   struct {
			String string
		}
		Stub func() string
	}
}

func (f *CommentGetter) GetCreatedAt() string {
	f.GetCreatedAtCall.Lock()
	defer f.GetCreatedAtCall.Unlock()
	f.GetCreatedAtCall.CallCount++
	if f.GetCreatedAtCall.Stub != nil {
		return f.GetCreatedAtCall.Stub()
	}
	return f.GetCreatedAtCall.Returns.String
}
func (f *CommentGetter) GetFirstReply(param1 internal.Client, param2 ...string) (internal.Comment, error) {
	f.GetFirstReplyCall.Lock()
	defer f.GetFirstReplyCall.Unlock()
	f.GetFirstReplyCall.CallCount++
	f.GetFirstReplyCall.Receives.Client = param1
	f.GetFirstReplyCall.Receives.IgnoredUsers = param2
	if f.GetFirstReplyCall.Stub != nil {
		return f.GetFirstReplyCall.Stub(param1, param2...)
	}
	return f.GetFirstReplyCall.Returns.Comment, f.GetFirstReplyCall.Returns.Error
}
func (f *CommentGetter) GetFirstResponseTime(param1 internal.Client, param2 internal.Clock, param3 ...string) (float64, error) {
	f.GetFirstResponseTimeCall.Lock()
	defer f.GetFirstResponseTimeCall.Unlock()
	f.GetFirstResponseTimeCall.CallCount++
	f.GetFirstResponseTimeCall.Receives.Client = param1
	f.GetFirstResponseTimeCall.Receives.Clock = param2
	f.GetFirstResponseTimeCall.Receives.IgnoredUsers = param3
	if f.GetFirstResponseTimeCall.Stub != nil {
		return f.GetFirstResponseTimeCall.Stub(param1, param2, param3...)
	}
	return f.GetFirstResponseTimeCall.Returns.Float64, f.GetFirstResponseTimeCall.Returns.Error
}
func (f *CommentGetter) GetNumber() int {
	f.GetNumberCall.Lock()
	defer f.GetNumberCall.Unlock()
	f.GetNumberCall.CallCount++
	if f.GetNumberCall.Stub != nil {
		return f.GetNumberCall.Stub()
	}
	return f.GetNumberCall.Returns.Int
}
func (f *CommentGetter) GetUserLogin() string {
	f.GetUserLoginCall.Lock()
	defer f.GetUserLoginCall.Unlock()
	f.GetUserLoginCall.CallCount++
	if f.GetUserLoginCall.Stub != nil {
		return f.GetUserLoginCall.Stub()
	}
	return f.GetUserLoginCall.Returns.String
}
