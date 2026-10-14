package registry

import (
	"encoding/json"
	"net/http"
)

// Problem is an RFC 9457 problem details object. Every error response from
// the API uses it, with Content-Type application/problem+json.
type Problem struct {
	Type   string       `json:"type"`
	Title  string       `json:"title"`
	Status int          `json:"status"`
	Detail string       `json:"detail,omitempty"`
	Errors []FieldError `json:"errors,omitempty"`
}

func writeProblem(w http.ResponseWriter, status int, detail string, fields []FieldError) {
	p := Problem{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
		Errors: fields,
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}
