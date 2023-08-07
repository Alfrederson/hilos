package forum

import (
	"errors"
	"time"
)

// Path é o caminho do post.
// Creator é a identidade do criador.
type Post struct {
	Id         string    `json:"id,omitempty"`
	Time       time.Time `json:"time,omitempty"`
	ParentId   string    `json:"parent_id,omitempty"`
	Creator    string    `json:"creator" form:"creator"`
	CreatorId  string    `json:"creator_id" form:"creator_id"`
	Subject    string    `json:"subject" form:"subject"`
	Content    string    `json:"content" form:"content"`
	ReplyCount int       `json:"replies_count"`
	Replies    []Post    `json:"replies,omitempty"`
}

func (p Post) Indices() []string {
	return []string{
		"parent_id",
		"creator_id",
	}
}

func (p Post) ReadField(field string) (string, error) {
	switch field {
	case "parent_id":
		return p.ParentId, nil
	case "creator_id":
		return p.CreatorId, nil
	default:
		return "", errors.New("invalid field " + field)
	}
}
