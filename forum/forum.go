package forum

import (
	"log"
	"time"

	"hilos/doc"
)

const (
	TOPIC_PAGE_COUNT = 10
)

var db struct {
	posts   *doc.DocDB
	reports *doc.DocDB
	status  *doc.DocDB

	// TODO:
	// - lista de tópicos a extinguir.
	// - lista de bans.
	// - lista de sessões.
}

type ForumStatus struct {
	somethingChanged bool
	LastPosts        []*Post
}

var status ForumStatus

func Status() *ForumStatus {
	return &status
}

// trocar isso pelo ring buffer
func (s *ForumStatus) PubPost(p *Post) {
	s.somethingChanged = true
	s.LastPosts = append(s.LastPosts, p)
	if len(s.LastPosts) > 4 {
		s.LastPosts = s.LastPosts[1:]
	}
}
func RebuildIndex() {
	db.posts.RebuildIndex()
	db.reports.RebuildIndex()
}
func Nuke() {
	db.posts.Clear()
	db.reports.Clear()
	db.posts.RebuildIndex()
	db.reports.RebuildIndex()
}
func Start() {
	db.posts = doc.Create("posts.db", &Post{})
	db.reports = doc.Create("reports.db", &Report{})
	db.status = doc.Create("status.db", nil)
	db.status.Get("lastPosts", &status.LastPosts)
	// persiste periodicamente os últimos posts...
	go func() {
		for {
			if !status.somethingChanged {
				log.Println("⏲️")
				time.Sleep(time.Second * 240)
				continue
			}
			time.Sleep(time.Second * 15)
			db.status.Save("lastPosts", &status.LastPosts)
			status.somethingChanged = false
		}
	}()

	log.Println("forum component initialized")
}
