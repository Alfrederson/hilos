package forum

import (
	"encoding/json"
	"errors"
	"hilos/doc"
	"log"
	"time"
)

func GetRootTopics(page int, count int) []Post {
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

func CreateRootTopic(t Post) (string, error) {
	id := doc.New()
	t.Time = time.Now()

	t.Id = id
	t.ParentId = ""

	err := db.posts.Add(id, &t)
	if err != nil {
		return "", err
	}

	status.TotalPosts++

	return id, err
}

func ReadTopic(topic_id string, fromPage int64) (*Post, error) {
	topic := Post{}
	// pega a raiz...
	if err := db.posts.Get(topic_id, &topic); err != nil {
		return nil, errors.New("error reading topic. database dead or topic doesn't exist")
	}
	topic.Id = topic_id
	topic.Replies = make([]Post, 0, TOPIC_PAGE_COUNT)

	// TODO: deixar a pessoa escolher se quer ver os primeiros, os últimos, os últimos que tiveram
	//       respostas.
	lista, _ := db.posts.FindLastUpdatedWhere(int(fromPage), TOPIC_PAGE_COUNT, cond("parent_id", "=", topic_id))
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

// duas funções com exatamente a mesma implementação................
func GetTopic(topic_id string) (*Post, error) {
	conversa := Post{}
	if err := db.posts.Get(topic_id, &conversa); err != nil {
		return nil, errors.New("no such topic")
	}
	return &conversa, nil
}
