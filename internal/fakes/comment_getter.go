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
	GetFirstContactTimeCall struct {
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
func (f *CommentGetter) GetFirstContactTime(param1 internal.Client, param2 internal.Clock, param3 ...string) (float64, error) {
	f.GetFirstContactTimeCall.Lock()
	defer f.GetFirstContactTimeCall.Unlock()
	f.GetFirstContactTimeCall.CallCount++
	f.GetFirstContactTimeCall.Receives.Client = param1
	f.GetFirstContactTimeCall.Receives.Clock = param2
	f.GetFirstContactTimeCall.Receives.IgnoredUsers = param3
	if f.GetFirstContactTimeCall.Stub != nil {
		return f.GetFirstContactTimeCall.Stub(param1, param2, param3...)
	}
	return f.GetFirstContactTimeCall.Returns.Float64, f.GetFirstContactTimeCall.Returns.Error
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
func (f *CommentGetter) GetUserLogin() string {
	f.GetUserLoginCall.Lock()
	defer f.GetUserLoginCall.Unlock()
	f.GetUserLoginCall.CallCount++
	if f.GetUserLoginCall.Stub != nil {
		return f.GetUserLoginCall.Stub()
	}
	return f.GetUserLoginCall.Returns.String
}
