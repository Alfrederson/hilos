package forum

import (
	"encoding/json"
	"log"
)

type ReportedPost struct {
	OriginalPost Post
	Report
}

func GetReports() ([]ReportedPost, error) {
	result := make([]ReportedPost, 0, 10)

	/*
		lista, _ := db.posts.FindLastUpdated("parent_id", "=", topic_id, int(fromPage), TOPIC_PAGE_COUNT)
		for _, data := range lista {
			mensagem := Post{}
			err := json.Unmarshal([]byte(data), &mensagem)
			if err != nil {
				log.Println(err)
			}
			topic.Replies = append(topic.Replies, mensagem)
		}

	*/
	reports, err := db.reports.FindLastUpdated("processed", "=", false, 0, 100)
	if err != nil {
		return result, err
	}
	for _, report := range reports {
		r := Report{}
		if err := json.Unmarshal([]byte(report), &r); err != nil {
			log.Println(err)
		}
		p := Post{}
		if err := db.posts.Get(r.PostID, &p); err != nil {
			p.Subject = err.Error()
			p.Content = "(this post was already deleted)"
		}
		result = append(result, ReportedPost{
			Report:       r,
			OriginalPost: p,
		})
	}
	return result, nil
}
