package forum

import (
	"time"
)

type Report struct {
	PostID      string    `json:"post_id,omitempty"`
	Message     string    `json:"message,omitempty"`
	CreatorID   string    `json:"creator_id"`
	CreatorName string    `json:"creator_name"`
	IP          string    `json:"ip,omitempty"`
	Time        time.Time `json:"time,omitempty"`
	Processed   bool      `json:"processed"`
}

func (r *Report) ObjectIndex() []string {
	return []string{
		"post_id",
		"creator_id",
		"processed",
	}
}

func (r *Report) IndexTable() interface{} {
	type Indices struct {
		PostID    string `json:"post_id" gorm:"index:idx_post_id"`
		CreatorID string `json:"creator_id" gorm:"index:idx_creator_id"`
		IP        string `json:"ip" gorm:"index:idx_ip"`
		Processed bool   `json:"processed" gorm:"index:idx_processed"`
	}
	return Indices{}
}

func (r *Report) IndexedFields() interface{} {
	type Fields struct {
		PostID    string `json:"post_id" gorm:"index:idx_post_id"`
		CreatorID string `json:"creator_id" gorm:"index:idx_creator_id"`
		IP        string `json:"ip" gorm:"index:idx_ip"`
		Processed bool   `json:"processed" gorm:"index:idx_processed"`
	}
	return Fields{
		PostID:    r.PostID,
		CreatorID: r.CreatorID,
		IP:        r.IP,
		Processed: r.Processed,
	}
}

// Path é o caminho do post.
// Creator é a identidade do criador.
type Post struct {
	Id          string    `json:"id,omitempty"`
	Time        time.Time `json:"time,omitempty"`
	ParentId    string    `json:"parent_id"`
	Creator     string    `json:"creator" form:"creator"`
	CreatorId   string    `json:"creator_id" form:"creator_id"`
	Subject     string    `json:"subject" form:"subject"`
	Content     string    `json:"content" form:"content"`
	ReplyCount  int       `json:"replies_count"`
	Replies     []Post    `json:"replies,omitempty"`
	ReportCount int       `json:"report_count,omitempty"`
	IP          string    `json:"ip,omitempty"`
	Frozen      bool      `json:"frozen"`
}

func (p *Post) ObjectIndex() []string {
	return []string{
		"creator_id",
		"ip",
		"parent_id",
	}
}

func (p *Post) IndexTable() interface{} {
	type Indices struct {
		ParentId  string `json:"parent_id" gorm:"index:idx_parent_id"`
		CreatorId string `json:"creator_id" gorm:"index:idx_creator_id"`
		IP        string `json:"ip" gorm:"index:idx_ip"`
	}
	return Indices{}
}
func (p *Post) IndexedFields() interface{} {
	type Fields struct {
		ParentId  string `json:"parent_id" gorm:"index:idx_parent_id"`
		CreatorId string `json:"creator_id" gorm:"index:idx_creator_id"`
		IP        string `json:"ip" gorm:"index:idx_ip"`
	}
	return Fields{
		ParentId:  p.ParentId,
		CreatorId: p.CreatorId,
		IP:        p.IP,
	}
}
