package post

import (
	"redditclone/pkg/user"
	"time"
)

type Vote struct {
	User string `json:"user"`
	Vote int32  `json:"vote"`
}

type Comment struct {
	Created time.Time  `json:"created"`
	Author  *user.User `json:"author"`
	Body    string     `json:"body"`
	ID      string     `json:"id"`
}

type Post struct {
	Score            int32      `json:"score"`
	Views            int32      `json:"views"`
	Type             string     `json:"type"`
	Title            string     `json:"title"`
	Text             string     `json:"text,omitempty"`
	Url              string     `json:"url,omitempty"`
	Author           *user.User `json:"author"`
	Category         string     `json:"category"`
	Votes            []Vote     `json:"votes"`
	Comments         []Comment  `json:"comments"`
	Created          time.Time  `json:"created"`
	UpvotePercentage uint32     `json:"upvotePercentage"`
	ID               string     `json:"id"`
}
