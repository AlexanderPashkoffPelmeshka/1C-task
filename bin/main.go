package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"diary/database"
	"diary/httpapi"
	"diary/models"
)

//go:embed web/*
var webFS embed.FS

func main() {
	store, err := database.Open(database.StandartPath)

	if err != nil {
		log.Fatal("открытие БД: ", err)
	}

	defer store.Close()

	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal("web: ", err)
	}

	api := httpapi.New(store)
	handler := api.Router(http.FS(staticFS))

	log.Printf("сервер запущен: http://%s", models.Localhost)

	if err := http.ListenAndServe(models.Localhost, handler); err != nil {
		log.Fatal(err)
	}
}
