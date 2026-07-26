package problem

import (
	"encoding/json"
	"net/http"

	mw "github.com/opencorex-org/openrevenue/pkg/middleware"
)

type Detail struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Status        int    `json:"status"`
	Detail        string `json:"detail,omitempty"`
	Instance      string `json:"instance,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
}

func Write(w http.ResponseWriter, r *http.Request, status int, title string, err error) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	d := Detail{
		Type: "about:blank", Title: title, Status: status, Instance: r.URL.Path,
		CorrelationID: mw.CorrelationID(r.Context()),
	}
	if err != nil {
		d.Detail = err.Error()
	}
	_ = json.NewEncoder(w).Encode(d)
}
