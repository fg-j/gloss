package internal

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type TimeContainer struct {
	Time  float64
	Error error
}

type Issue struct {
	CreatedAt   string `json:"created_at"`
	NumComments int    `json:"comments"`
	CommentsURL string `json:"comments_url"`
	User        struct {
		Login string `json:"login"`
	} `json:"user"`
}

type Comment struct {
	User struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
}

//go:generate faux --interface CommentGetter --output fakes/comment_getter.go
type CommentGetter interface {
	GetFirstReply(client Client, ignoredUsers ...string) (Comment, error)
	GetCreatedAt() string
	GetUserLogin() string
}

func (i *Issue) GetFirstReply(client Client, ignoredUsers ...string) (Comment, error) {
	if i.NumComments == 0 {
		return Comment{}, nil
	}

	commentsURL, err := url.Parse(i.CommentsURL)
	if err != nil {
		return Comment{}, fmt.Errorf("parsing comments url: %s", err)
	}

	//TODO: Figure out whether pagination is a thing we should worry about here
	body, err := client.Get(commentsURL.Path)
	if err != nil {
		return Comment{}, fmt.Errorf("getting issue comments: %s", err)
	}

	replies := []Comment{}
	err = json.Unmarshal(body, &replies)
	if err != nil {
		return Comment{}, fmt.Errorf("getting issue comments: could not unmarshal JSON '%s' : %s", string(body), err)
	}

	ignore := make(map[string]struct{})
	for i := range ignoredUsers {
		ignore[ignoredUsers[i]] = struct{}{}
	}

	// Comments are sorted by ascending ID. TODO:Does that correspond to recency of creation?
	for _, reply := range replies {
		if reply.User.Login == i.User.Login {
			continue
		}
		if _, skipUser := ignore[reply.User.Login]; skipUser {
			continue
		}
		if reply.User.Type == "Bot" {
			continue
		}
		return reply, nil
	}
	return Comment{}, nil
}

func (i *Issue) GetCreatedAt() string {
	return i.CreatedAt
}

func (i *Issue) GetUserLogin() string {
	return i.User.Login
}
