package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const StandartPath = "data/diary.db"

//go:embed schema.sql
var schemaSQL string

var (
	ErrNotFound      = errors.New("не найдено")
	ErrAlreadyClosed = errors.New("дело уже закрыто")
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("создание папки: %w", err)
		}
	}

	dsn := "file:" + path + "?" +
		"_pragma=journal_mode(WAL)&" +
		"_pragma=foreign_keys(ON)&" +
		"_pragma=busy_timeout(5000)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("открытие БД: %w", err)
	}

	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.applySchema(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) applySchema() error {
	tx, err := s.db.Begin()

	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(schemaSQL); err != nil {
		return fmt.Errorf("применение схемы: %w", err)
	}

	return tx.Commit()
}

func (s *Store) Truncate() error {
	tx, err := s.db.Begin()

	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec("delete from comments")
	tx.Exec("delete from cases")

	return tx.Commit()
}
