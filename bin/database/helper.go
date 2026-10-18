package database

import (
	"time"
)

func toDB(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func fromDB(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
