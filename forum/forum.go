package forum

import (
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
	posts            *doc.DocDB
	posts_by_parent  *doc.DocDB
	posts_by_creator *doc.DocDB

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

func Start() {
	db.posts = doc.Create("posts.db")

	db.posts.UsingIndexable(&Post{})

	db.posts_by_parent = doc.CreateIndex("posts.parent_id.db")
	db.posts_by_creator = doc.CreateIndex("posts.creator_id.db")

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

func GetTopics(page int, amount int) []Post {
	lista := db.posts_by_parent.List("root", page*amount, amount)

	resultado := make([]Post, 0, amount)
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

	err := db.posts.Save(id, t)
	if err != nil {
		return "", err
	}

	// indexa
	t.Id = id
	t.ParentId = "root"

	go t.WriteToIndex()

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

	lista := db.posts_by_parent.List(topic_id, int(fromPage)*TOPIC_PAGE_COUNT, TOPIC_PAGE_COUNT)

	for _, reply_id := range lista {
		mensagem := Post{}
		db.posts.Get(reply_id, &mensagem)
		mensagem.Id = reply_id
		topic.Replies = append(topic.Replies, mensagem)
	}

	return &topic, nil
}

func ReadUserPosts(userId string) ([]Post, error) {
	lista := db.posts_by_creator.List(userId, 0, 100)
	resultado := make([]Post, 0, 100)
	for _, id := range lista {
		mensagem := Post{}
		db.posts.Get(id, &mensagem)
		mensagem.Id = id
		resultado = append(resultado, mensagem)
	}
	return resultado, nil
}

func ReplyTopic(topic_id string, reply Post) (string, error) {
	// vê se o tópico existe
	conversa := Post{}
	if err := db.posts.Get(topic_id, &conversa); err != nil {
		return "", errors.New("no such topic")
	}

	reply.Id = doc.New()
	reply.ParentId = topic_id

	err := db.posts.Save(reply.Id, reply)
	go reply.WriteToIndex()

	if err != nil {
		log.Println("error saving post:", err)
		return "", errors.New("couldn't save the post")
	}

	conversa.ReplyCount += 1
	if err := db.posts.Save(topic_id, conversa); err != nil {
		log.Println("error incrementing reply count:", err)
	}

	status.PubPost(&reply)

	log.Printf("%s replied to %s\n", reply.Creator, reply.ParentId)

	return reply.Id, nil
}

func ReadPost(topicId string) (*Post, error) {
	resultado := Post{}
	if err := db.posts.Get(topicId, &resultado); err != nil {
		return nil, errors.New("could not read post " + topicId)
	}
	return &resultado, nil
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
