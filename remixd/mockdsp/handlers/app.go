package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type App struct {
}

// json helpers

// JSONErr err
type JSONErr struct {
	Err string `json:"err,omitempty"`
}

// JSONError renders json with error
func JSONError(w http.ResponseWriter, errStr string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	writeJSON(w, &JSONErr{Err: errStr})
}

// JSONOk renders json with 200 ok
func JSONOk(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, v)
}

// NoContent renders empty with 204 k
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
	_, err := w.Write(nil)
	if err != nil {
		fmt.Printf("nocontent write err: %v", err)
	}
}

// writeJSON to response body
func writeJSON(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, fmt.Sprintf("json encoding error: %v", err), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		http.Error(w, fmt.Sprintf("write error: %v", err), http.StatusInternalServerError)
		return
	}
}

// readJSON from request body
func readJSON(r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("invalid JSON input")
	}

	return nil
}

// CtxKey context key
type CtxKey string
