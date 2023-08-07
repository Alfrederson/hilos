package forum

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"hilos/doc"
)

type Entry struct {
	Id string
}

const (
	TOPIC_PAGE_COUNT = 10
)

var db struct {
	posts   *doc.DocDB
	reports *doc.DocDB

	status *doc.DocDB
}

type ForumStatus struct {
	somethingChanged bool
	LastPosts        []*Post
}

var status ForumStatus

func Status() *ForumStatus {
	return &status
}

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
	db.posts = doc.Create("posts.db")
	db.posts.UsingIndexable(&Post{})

	db.reports = doc.Create("reports.db")
	db.reports.UsingIndexable(&Report{})

	db.status = doc.Create("status.db")

	db.status.Get("lastPosts", &status.LastPosts)

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

func GetTopics(page int, count int) []Post {
	lista, err := db.posts.Find("parent_id", "=", "", page, count)
	resultado := make([]Post, 0, count)

	if err != nil {
		resultado = append(resultado, Post{
			Subject: err.Error(),
		})
		return resultado
	}

	for _, id := range lista {
		conversa := Post{}
		db.posts.Get(id, &conversa)
		conversa.Id = id
		resultado = append(resultado, conversa)
	}
	return resultado
}

func CreateTopic(t Post) (string, error) {
	// salva
	id := doc.New()
	t.Time = time.Now()

	t.Id = id
	t.ParentId = ""

	err := db.posts.Add(id, t)
	if err != nil {
		return "", err
	}

	return id, err
}

func ReadTopic(topic_id string, fromPage int64) (*Post, error) {
	topic := Post{}

	if err := db.posts.Get(topic_id, &topic); err != nil {
		log.Println("error reading topic ", err)
		return nil, errors.New("error reading topic. database dead or topic doesn't exist.")
	}
	topic.Id = topic_id

	topic.Replies = make([]Post, 0, TOPIC_PAGE_COUNT)

	//	lista, err := db.posts.Find("parent_id", "=", topic_id, int(fromPage), TOPIC_PAGE_COUNT)
	//	if err != nil {
	//		log.Println(err)
	//		return nil, errors.New("could not read this topic. dunno why.")
	//	}

	// TODO: deixar a pessoa escolher se quer ver os primeiros, os últimos, os últimos que tiveram
	//       respostas.
	lista, _ := db.posts.FindLastUpdated("parent_id", "=", topic_id, int(fromPage), TOPIC_PAGE_COUNT)
	for _, data := range lista {
		mensagem := Post{}
		err := json.Unmarshal([]byte(data), &mensagem)
		if err != nil {
			log.Println(err)
		}
		topic.Replies = append(topic.Replies, mensagem)
	}

	return &topic, nil
}

func ReadUserPosts(userId string, fromPage int64) ([]Post, error) {
	lista, _ := db.posts.FindLast("creator_id", "=", userId, int(fromPage), TOPIC_PAGE_COUNT)
	posts := make([]Post, 0, TOPIC_PAGE_COUNT)
	for _, data := range lista {
		mensagem := Post{}
		err := json.Unmarshal([]byte(data), &mensagem)
		if err != nil {
			log.Println(err)
		}
		posts = append(posts, mensagem)
	}
	return posts, nil
}

// duas funções com exatamente a mesma implementação................
func GetTopic(topic_id string) (*Post, error) {
	conversa := Post{}
	if err := db.posts.Get(topic_id, &conversa); err != nil {
		return nil, errors.New("no such topic")
	}
	return &conversa, nil
}
func ReadPost(topicId string) (*Post, error) {
	resultado := Post{}
	if err := db.posts.Get(topicId, &resultado); err != nil {
		return nil, errors.New("could not read post " + topicId)
	}
	return &resultado, nil
}

func ReplyTopic(topic *Post, reply Post) (string, error) {
	// vê se o tópico existe

	reply.Id = doc.New()

	reply.ParentId = topic.Id
	reply.Time = time.Now()

	err := db.posts.Add(reply.Id, reply)

	if err != nil {
		log.Println("error saving post:", err)
		return "", errors.New("couldn't save the post")
	}

	topic.ReplyCount += 1
	if err := db.posts.Save(topic.Id, topic); err != nil {
		log.Println("error incrementing reply count:", err)
	}

	status.PubPost(&reply)

	log.Printf("%s (%s) replied to %s\n", reply.Creator, reply.IP, reply.ParentId)

	return reply.Id, nil
}

// não pode trocar o autor e o parent_id porque
// teria que reindexar e eu não quero fazer isso.
// idealmente é pra deixar poder fazer, mas não vai não por enquanto.

func RewritePost(id string, rewrite *Post) error {
	if err := db.posts.Save(id, rewrite); err != nil {
		log.Println("error editing post: ", err)
		return err
	}
	return nil
}
