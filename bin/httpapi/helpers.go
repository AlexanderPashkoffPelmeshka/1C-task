package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"diary/database"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, errorResponse{Error: msg})
}

func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, database.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())

	} else if errors.Is(err, database.ErrAlreadyClosed) {
		writeError(w, http.StatusConflict, err.Error())

	} else {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("пустое тело запроса")
	}
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("некорректный JSON: %w", err)
	}
	return nil
}

func getID(r *http.Request) int64 {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id
}
