package internal

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"time"
)

type TimeContainer struct {
	Time  float64
	Error error
}

type Issue struct {
	CreatedAt   string `json:"created_at"`
	NumComments int    `json:"comments"`
	CommentsURL string `json:"comments_url"`
	Number      int    `json:"number"`
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
	GetNumber() int
	GetFirstResponseTime(client Client, clock Clock, ignoredUsers ...string) (float64, error)
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
func (i *Issue) GetNumber() int {
	return i.Number
}

func (i *Issue) GetFirstResponseTime(client Client, clock Clock, ignoredUsers ...string) (float64, string, error) {

	// // TODO: pass a set of ignored users here
	comment, err := i.GetFirstReply(client, ignoredUsers...)

	if err != nil {
		return -1, "", fmt.Errorf("could not get first reply: %s", err)
	}

	// // TODO: decide whether to actually include issues without comments on them
	var replyCreated time.Time
	var replyUser string
	if comment.CreatedAt == "" {
		replyCreated = clock.Now().UTC()
	} else {
		replyCreated, err = time.Parse(time.RFC3339, comment.CreatedAt)

		if err != nil {
			return -1, "", fmt.Errorf("could not parse first reply time: %s", err)
		}
		replyUser = comment.User.Login
	}

	issueCreated, err := time.Parse(time.RFC3339, i.GetCreatedAt())
	if err != nil {
		return -1, "", fmt.Errorf("could not parse issue creation time: %s", err)
	}
	replyTime := math.Round(replyCreated.Sub(issueCreated).Minutes())
	return replyTime, replyUser, nil
}
