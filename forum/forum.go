package forum

import (
	"errors"
	"log"

	"hilos/doc"
)

var posts *doc.DocDB

var posts_by_parent *doc.DocDB
var posts_by_creator *doc.DocDB

type Entry struct {
	Id string
}

const (
	TOPIC_PAGE_COUNT = 10
)

// Path é o caminho do post.
// Creator é a identidade do criador.
type Post struct {
	Id         string `json:"id,omitempty"`
	ParentId   string `json:"parent_id,omitempty"`
	Creator    string `json:"creator" form:"creator"`
	CreatorId  string `json:"creator_id" form:"creator_id"`
	Subject    string `json:"subject" form:"subject"`
	Content    string `json:"content" form:"content"`
	ReplyCount int    `json:"replies_count"`
	Replies    []Post `json:"replies,omitempty"`
}

// global exportada e estou 300% nem aí.
var LastPost *Post

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
	err := posts_by_parent.Add(p.ParentId, p.Id)
	if err != nil {
		log.Println("ERROR indexing post ", p.Id, ":", err)
	}
	// creator id
	err = posts_by_creator.Add(p.CreatorId, p.Id)
	if err != nil {
		log.Println("ERROR indexing post by parent", p.Id, ":", err)
	}
}

func (p *Post) RemoveFromIndex() {
	posts_by_parent.Delete(p.ParentId)
	posts_by_creator.Delete(p.CreatorId)
}

func Start() {
	posts = doc.Create("posts.db")

	posts.UsingIndexable(&Post{})

	posts_by_parent = doc.CreateIndex("posts.parent_id.db")
	posts_by_creator = doc.CreateIndex("posts.creator_id.db")

	log.Println("forum component initialized")
}

func GetTopics(page int, amount int) []Post {
	lista := posts_by_parent.List("root", page*amount, amount)

	resultado := make([]Post, 0, amount)
	for _, id := range lista {
		conversa := Post{}
		posts.Get(id, &conversa)
		conversa.Id = id
		resultado = append(resultado, conversa)
	}
	return resultado
}

func CreateTopic(t Post) (string, error) {
	// salva
	id := doc.New()

	err := posts.Save(id, t)
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
	// pega o tópico
	topic := Post{}

	if err := posts.Get(topic_id, &topic); err != nil {
		log.Println("error reading topic ", err)
		return nil, errors.New("error reading topic. database dead or topic doesn't exist.")
	}
	topic.Id = topic_id

	topic.Replies = make([]Post, 0, TOPIC_PAGE_COUNT)

	lista := posts_by_parent.List(topic_id, int(fromPage)*TOPIC_PAGE_COUNT, TOPIC_PAGE_COUNT)

	for _, reply_id := range lista {
		mensagem := Post{}
		posts.Get(reply_id, &mensagem)
		mensagem.Id = reply_id
		topic.Replies = append(topic.Replies, mensagem)
	}

	return &topic, nil
}

func ReadUserPosts(userId string) ([]Post, error) {
	lista := posts_by_creator.List(userId, 0, 100)
	resultado := make([]Post, 0, 100)
	for _, id := range lista {
		mensagem := Post{}
		posts.Get(id, &mensagem)
		mensagem.Id = id
		resultado = append(resultado, mensagem)
	}
	return resultado, nil
}

func ReplyTopic(topic_id string, reply Post) (string, error) {
	// vê se o tópico existe
	conversa := Post{}
	if err := posts.Get(topic_id, &conversa); err != nil {
		return "", errors.New("no such topic")
	}

	reply.Id = doc.New()
	reply.ParentId = topic_id

	err := posts.Save(reply.Id, reply)
	go reply.WriteToIndex()

	if err != nil {
		log.Println("error saving post:", err)
		return "", errors.New("couldn't save the post")
	}

	conversa.ReplyCount += 1
	if err := posts.Save(topic_id, conversa); err != nil {
		log.Println("error incrementing reply count:", err)
	}

	LastPost = &reply

	log.Printf("%s replied to %s\n", reply.Creator, reply.ParentId)

	return reply.Id, nil
}

func ReadPost(topicId string) (*Post, error) {
	resultado := Post{}
	if err := posts.Get(topicId, &resultado); err != nil {
		return nil, errors.New("could not read post " + topicId)
	}
	return &resultado, nil
}

// não pode trocar o autor e o parent_id porque
// teria que reindexar e eu não quero fazer isso.
// idealmente é pra deixar poder fazer, mas não vai não por enquanto.

func RewritePost(id string, rewrite *Post) error {
	if err := posts.Save(id, rewrite); err != nil {
		log.Println("error editing post: ", err)
		return err
	}
	return nil
}
