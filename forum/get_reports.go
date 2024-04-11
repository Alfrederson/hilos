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
	Processed bool
	Page      int
	PerPage   int
}

func GetReports(options GetReportOptions) ([]ReportedPost, error) {
	result := make([]ReportedPost, options.Page, options.PerPage)
	var reports []string
	var err error
	if options.Processed {
		reports, err = db.reports.FindLastUpdated(options.Page, options.PerPage)
	} else {
		reports, err = db.reports.FindLastUpdatedWhere(options.Page, options.PerPage, cond("processed", "=", false))
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
