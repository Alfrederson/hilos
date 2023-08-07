package forum

import (
	"errors"
	"time"
)

type Report struct {
	PostID    string    `json:"post_id,omitempty"`
	Message   string    `json:"message,omitempty"`
	CreatorID string    `json:"creator_id"`
	IP        string    `json:"ip"`
	Time      time.Time `json:"time,omitempty"`
}

func (r Report) Indices() []string {
	return []string{
		"post_id",
		"creator_id",
		"ip",
	}
}

func (r Report) ReadField(field string) (string, error) {
	switch field {
	case "post_id":
		return r.PostID, nil
	case "creator_id":
		return r.CreatorID, nil
	case "ip":
		return r.IP, nil
	default:
		return "", errors.New("invalid field " + field)
	}
}

// Path é o caminho do post.
// Creator é a identidade do criador.
type Post struct {
	Id          string    `json:"id,omitempty"`
	Time        time.Time `json:"time,omitempty"`
	ParentId    string    `json:"parent_id,omitempty"`
	Creator     string    `json:"creator" form:"creator"`
	CreatorId   string    `json:"creator_id" form:"creator_id"`
	Subject     string    `json:"subject" form:"subject"`
	Content     string    `json:"content" form:"content"`
	ReplyCount  int       `json:"replies_count"`
	Replies     []Post    `json:"replies,omitempty"`
	ReportCount int       `json:"report_count,omitempty"`
	IP          string    `json:"-"`
}

func (p Post) Indices() []string {
	return []string{
		"parent_id",
		"creator_id",
		"ip",
	}
}

func (p Post) ReadField(field string) (string, error) {
	switch field {
	case "parent_id":
		return p.ParentId, nil
	case "creator_id":
		return p.CreatorId, nil
	case "ip":
		return p.IP, nil
	default:
		return "", errors.New("invalid field " + field)
	}
}
