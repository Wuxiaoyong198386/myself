package handler

import (
	"encoding/json"
	"net/http"
)

const (
	MsgHit     string = "hit"
	MsgMiss    string = "miss"
	MsgUnknown string = "unknown"
)

// SendResp sends response to http client with json body
func SendResp(w http.ResponseWriter, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bs)
	w.Write([]byte("\n"))
}

func SendString(w http.ResponseWriter, s string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte(s))
	w.Write([]byte("\n"))
}
