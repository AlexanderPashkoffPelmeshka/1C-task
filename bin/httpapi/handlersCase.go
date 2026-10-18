package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"diary/models"
)

func validateCase(req *models.CaseRequest, status string) error {
	if status == "draft" {
		return nil
	}

	if strings.TrimSpace(req.Title) == "" {
		return errors.New("title не может быть пустым")
	}

	if req.DueAt.Before(req.PromisedAt) {
		return errors.New("due_at не может быть раньше promised_at")
	}

	return nil
}

func (a *API) createCase(w http.ResponseWriter, r *http.Request) {
	var req models.CaseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := req.Status
	if status == "" {
		status = models.StatusOpen
	}

	if err := validateCase(&req, status); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := a.store.CreateCase(&models.Case{
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		PromisedAt:  req.PromisedAt,
		DueAt:       req.DueAt,
	})

	if err != nil {
		writeStoreError(w, err)
		return
	}

	c, _ := a.store.GetCase(id)
	writeJSON(w, http.StatusCreated, c)
}

func (a *API) listCases(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cases, err := a.store.ListCases(q.Get("status"), q.Get("q"), q.Get("from"), q.Get("to"))

	if err != nil {
		writeStoreError(w, err)
		return
	}

	if cases == nil {
		cases = []*models.Case{}
	}

	writeJSON(w, http.StatusOK, cases)
}

func (a *API) getCase(w http.ResponseWriter, r *http.Request) {
	c, err := a.store.GetCase(getID(r))

	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (a *API) updateCase(w http.ResponseWriter, r *http.Request) {
	id := getID(r)
	existing, err := a.store.GetCase(id)

	if err != nil {
		writeStoreError(w, err)
		return
	}

	var req models.CaseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	newStatus := existing.Status
	if req.Status != "" {
		newStatus = req.Status
	}

	if err := validateCase(&req, newStatus); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	existing.Title = req.Title
	existing.Description = req.Description
	existing.PromisedAt = req.PromisedAt
	existing.DueAt = req.DueAt
	existing.Status = newStatus

	if err := a.store.UpdateCase(existing); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (a *API) closeCase(w http.ResponseWriter, r *http.Request) {
	var req models.CloseCaseRequest

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := a.store.CloseCase(getID(r), req.Status, time.Now().UTC()); err != nil {
		writeStoreError(w, err)
		return
	}

	c, _ := a.store.GetCase(getID(r))
	writeJSON(w, http.StatusOK, c)
}

func (a *API) reopenCase(w http.ResponseWriter, r *http.Request) {
	id := getID(r)

	if err := a.store.ReopenCase(id); err != nil {
		writeStoreError(w, err)
		return
	}

	c, _ := a.store.GetCase(id)
	writeJSON(w, http.StatusOK, c)
}
