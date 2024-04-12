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

	prunes *doc.DocDB
	// TODO:
	// - lista de bans.
	// - lista de sessões.
}

type ForumStatus struct {
	somethingChanged bool
	PendingPrunes    int
	TotalPosts       int
	LastPosts        []*Post
}

var status ForumStatus

func Status() *ForumStatus {
	return &status
}

// trocar isso pelo ring buffer
func (s *ForumStatus) PubPost(p *Post) {
	s.somethingChanged = true
	s.TotalPosts++
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
	db.prunes.Clear()
	db.status.Clear()
	status.PendingPrunes = 0
	status.TotalPosts = 0
}
func Start() {
	db.posts = doc.Create("posts.db", &Post{})
	db.reports = doc.Create("reports.db", &Report{})
	db.status = doc.Create("status.db", nil)
	db.prunes = doc.Create("prunes.db", nil)
	db.status.Get("lastPosts", &status.LastPosts)

	status.PendingPrunes = int(db.prunes.Count())
	status.TotalPosts = int(db.posts.Count())
	// persiste periodicamente os últimos posts...
	go func() {
		for {
			if !status.somethingChanged {
				if status.PendingPrunes > 0 {
					log.Println("jannies are working ... ")
					RunPruneTask()
				}
				time.Sleep(time.Second * 15)
				continue
			}
			log.Println("writing status...")
			db.status.Save("lastPosts", &status.LastPosts)
			status.somethingChanged = false
			time.Sleep(time.Second * 300)
		}
	}()

	log.Println("forum component initialized")
}
