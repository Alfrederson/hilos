package forum

import "github.com/Alfrederson/hilos/doc"

func cond(field string, op string, value any) doc.Condition {
	return doc.Condition{
		Field: field,
		Op:    op,
		Value: value,
	}
}
