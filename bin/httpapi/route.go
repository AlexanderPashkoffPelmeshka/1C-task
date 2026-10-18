package httpapi

import (
	"net/http"

	"diary/database"
)

type API struct {
	store *database.Store
}

func New(store *database.Store) *API {
	return &API{store: store}
}

func (a *API) Router(fs http.FileSystem) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("GET /api/cases", a.listCases)
	mux.HandleFunc("POST /api/cases", a.createCase)
	mux.HandleFunc("GET /api/cases/{id}", a.getCase)
	mux.HandleFunc("PATCH /api/cases/{id}", a.updateCase)
	mux.HandleFunc("POST /api/cases/{id}/close", a.closeCase)
	mux.HandleFunc("POST /api/cases/{id}/reopen", a.reopenCase)

	mux.HandleFunc("GET /api/cases/{id}/comments", a.listComments)
	mux.HandleFunc("POST /api/cases/{id}/comments", a.addComment)

	mux.Handle("/", http.FileServer(fs))

	return mux
}
