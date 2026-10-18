package httpapi

import (
	"net/http"
	"strings"

	"diary/models"
)

func (a *API) listComments(w http.ResponseWriter, r *http.Request) {
	comments, err := a.store.ListComments(getID(r))

	if err != nil {
		writeStoreError(w, err)
		return
	}

	if comments == nil {
		comments = []*models.Comment{}
	}

	writeJSON(w, http.StatusOK, comments)
}

func (a *API) addComment(w http.ResponseWriter, r *http.Request) {
	var req models.CommentRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.Body) == "" {
		writeError(w, http.StatusBadRequest, "body не может быть пустым")
		return
	}

	c := &models.Comment{CaseID: getID(r), Body: req.Body}
	cid, err := a.store.AddComment(c)

	if err != nil {
		writeStoreError(w, err)
		return
	}

	c.ID = cid
	writeJSON(w, http.StatusCreated, c)
}
