package api_common

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// writeResponse is a helper method that allows to write and HTTP status & response
func WriteResponse(w http.ResponseWriter, status int, data interface{}, err error) {
	resp := response{
		Data: data,
	}
	if err != nil {
		resp.Error = fmt.Sprint(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	err = json.NewEncoder(w).Encode(data)
	if err := err; err != nil {
		fmt.Fprintf(w, "error encoding resp %v:%s", resp, err)
	}
}
