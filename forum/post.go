package forum

import (
	"errors"
	"log"
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

func (p *Post) ReadField(field string) (string, error) {
	switch field {
	case "parent_id":
		return p.ParentId, nil
	case "creator_id":
		return p.CreatorId, nil
	default:
		return "", errors.New("invalid field " + field)
	}
}

func (p *Post) WriteToIndex() {
	// parent id
	err := db.posts_by_parent.Add(p.ParentId, p.Id)
	if err != nil {
		log.Println("ERROR indexing post ", p.Id, ":", err)
	}
	// creator id
	err = db.posts_by_creator.Add(p.CreatorId, p.Id)
	if err != nil {
		log.Println("ERROR indexing post by parent", p.Id, ":", err)
	}
}

func (p *Post) RemoveFromIndex() {
	db.posts_by_parent.Delete(p.ParentId)
	db.posts_by_creator.Delete(p.CreatorId)
}
