package forum

import (
	"encoding/json"
	"errors"
	"hilos/doc"
	"log"
	"time"
)

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

func ReadPostsByUser(userId string, page int64, perPage int64) ([]Post, error) {
	lista, _ := db.posts.FindLast("creator_id", "=", userId, int(page), int(perPage))
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

func ReadPost(postId string) (*Post, error) {
	resultado := Post{}
	if err := db.posts.Get(postId, &resultado); err != nil {
		return nil, errors.New("could not read post " + postId)
	}
	return &resultado, nil
}

func WritePost(topic *Post, post Post) (string, error) {
	// vê se o tópico existe

	post.Id = doc.New()

	post.ParentId = topic.Id
	post.Time = time.Now()

	err := db.posts.Add(post.Id, &post)

	if err != nil {
		log.Println("error saving post:", err)
		return "", errors.New("couldn't save the post")
	}

	topic.ReplyCount += 1
	if err := db.posts.Save(topic.Id, topic); err != nil {
		log.Println("error incrementing reply count:", err)
	}
	status.PubPost(&post)
	log.Printf("%s (%s) replied to %s\n", post.Creator, post.IP, post.ParentId)
	return post.Id, nil
}

// não pode trocar o autor e o parent_id porque
// teria que reindexar e eu não quero fazer isso.
// na verdade agora faz meio automático
// idealmente é pra deixar poder fazer, mas não vai não por enquanto.
func RewritePost(id string, rewrite *Post) error {
	if err := db.posts.Save(id, rewrite); err != nil {
		log.Println("error editing post: ", err)
		return err
	}
	return nil
}

// os posts vão sumindo aos poucos, recursivamente
type PruneTask struct {
	PostID string `json:"post_id"`
}

func PrunePost(postId string) error {
	if !db.posts.Exists(postId) {
		return errors.New("post doesn't exist")
	}
	db.posts.Delete(postId)
	err := db.prunes.Save(postId, &PruneTask{
		PostID: postId,
	})
	if err != nil {
		log.Println("error issuing prune task: ", err)
	} else {
		log.Println("prune task has been issued")
	}
	return nil
}
