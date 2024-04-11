package forum

import (
	"encoding/json"
	"log"
)

type ReportedPost struct {
	OriginalPost Post
	Report
}

type GetReportOptions struct {
	Stale bool
}

func GetReports(options GetReportOptions) ([]ReportedPost, error) {
	result := make([]ReportedPost, 0, 10)
	var reports []string
	var err error

	if options.Stale {
		reports, err = db.reports.FindLastUpdated(0, 100)
	} else {
		reports, err = db.reports.FindLastUpdatedWhere(0, 100, cond("processed", "=", false))
	}
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
