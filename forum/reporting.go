package forum

import (
	"errors"
	"log"
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

// func (r *Report) IndexTable() interface{} {
// 	type Indices struct {
// 		PostID    string `json:"post_id" gorm:"index:idx_post_id"`
// 		CreatorID string `json:"creator_id" gorm:"index:idx_creator_id"`
// 		IP        string `json:"ip" gorm:"index:idx_ip"`
// 		Processed bool   `json:"processed" gorm:"index:idx_processed"`
// 	}
// 	return Indices{}
// }

// func (r *Report) IndexedFields() interface{} {
// 	type Fields struct {
// 		PostID    string `json:"post_id" gorm:"index:idx_post_id"`
// 		CreatorID string `json:"creator_id" gorm:"index:idx_creator_id"`
// 		IP        string `json:"ip" gorm:"index:idx_ip"`
// 		Processed bool   `json:"processed" gorm:"index:idx_processed"`
// 	}
// 	return Fields{
// 		PostID:    r.PostID,
// 		CreatorID: r.CreatorID,
// 		IP:        r.IP,
// 		Processed: r.Processed,
// 	}
// }

// reportar um post.
func ReportPost(id string, report *Report) error {
	// post já foi reportado?
	if db.reports.Exists(id) {
		return errors.New("post was already reported, sir")
	}
	if !db.posts.Exists(id) {
		return errors.New("no such post, sir")
	}
	if err := db.reports.Add(id, report); err != nil {
		log.Println("error reporting post: ", err)
		return err
	}
	return nil
}

// pega um report.
func GetReport(reportId string) (*Report, error) {
	r := Report{}
	if err := db.reports.Get(reportId, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// marca o report como concluído.
func DismissReport(r *Report) error {
	r.Processed = true
	return db.reports.Save(r.PostID, r)
}
