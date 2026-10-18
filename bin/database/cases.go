package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"diary/models"
)

func scanCase(rows *sql.Rows) (*models.Case, error) {
	var c models.Case
	var p, d, cr, up string
	var cl sql.NullString

	err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Status, &p, &d, &cr, &up, &cl)
	if err != nil {
		return nil, err
	}

	c.PromisedAt, _ = fromDB(p)
	c.DueAt, _ = fromDB(d)
	c.CreatedAt, _ = fromDB(cr)
	c.UpdatedAt, _ = fromDB(up)
	if cl.Valid {
		t, _ := fromDB(cl.String)
		c.ClosedAt = &t
	}

	return &c, nil
}

func (s *Store) CreateCase(c *models.Case) (int64, error) {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	c.UpdatedAt = c.CreatedAt

	var closedAt sql.NullString

	if c.ClosedAt != nil {
		closedAt = sql.NullString{String: toDB(*c.ClosedAt), Valid: true}
	}

	res, err := s.db.Exec(`
		insert into cases (title, description, status, promised_at, due_at, created_at, updated_at, closed_at)
		values (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Title, c.Description, c.Status,
		toDB(c.PromisedAt),
		toDB(c.DueAt),
		toDB(c.CreatedAt),
		toDB(c.UpdatedAt),
		closedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления: %w", err)
	}
	return res.LastInsertId()
}

func (s *Store) GetCase(id int64) (*models.Case, error) {
	rows, err := s.db.Query(`
		select id, title, description, status, promised_at, due_at, created_at, updated_at, closed_at 
		from cases where id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrNotFound
	}

	return scanCase(rows)
}

func (s *Store) ListCases(status, search, fromDate, toDate string) ([]*models.Case, error) {
	query := `
		select id, title, description, status, promised_at, due_at, created_at, updated_at, closed_at 
		from cases 
		where (status = ? or ? = '')
		  AND (title LIKE ? or description LIKE ? or ? = '')
		  AND (due_at >= ? or ? = '')
		  AND (due_at <= ? or ? = '')
		order by due_at asc`

	searchPattern := "%" + search + "%"

	rows, err := s.db.Query(query,
		status, status,
		searchPattern, searchPattern, search,
		fromDate, fromDate,
		toDate, toDate,
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка списка: %w", err)
	}

	defer rows.Close()

	var result []*models.Case

	for rows.Next() {
		c, err := scanCase(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении списка: %w", err)
	}

	return result, nil
}

func (s *Store) UpdateCase(c *models.Case) error {
	c.UpdatedAt = time.Now().UTC()

	var closedAt sql.NullString
	if c.ClosedAt != nil {
		closedAt = sql.NullString{String: toDB(*c.ClosedAt), Valid: true}
	}

	res, err := s.db.Exec(`
		update cases set title=?, description=?, status=?, promised_at=?, due_at=?, updated_at=?, closed_at=? 
		where id = ?`,
		c.Title, c.Description, c.Status,
		toDB(c.PromisedAt),
		toDB(c.DueAt),
		toDB(c.UpdatedAt),
		closedAt, c.ID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления: %w", err)
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CloseCase(id int64, status string, closedAt time.Time) error {
	res, err := s.db.Exec(`
		update cases set status=?, closed_at=?, updated_at=? 
		where id = ? AND status = 'open'`,
		status, toDB(closedAt), toDB(time.Now().UTC()), id,
	)
	if err != nil {
		return fmt.Errorf("ошибка закрытия: %w", err)
	}

	n, _ := res.RowsAffected()

	if n == 0 {
		var exists int
		err := s.db.QueryRow(`select 1 from cases where id = ?`, id).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return ErrAlreadyClosed
	}
	return nil
}

func (s *Store) ReopenCase(id int64) error {
	res, err := s.db.Exec(`
		update cases set status='open', closed_at=NULL, updated_at=? 
		where id = ?`, toDB(time.Now().UTC()), id)

	if err != nil {
		return fmt.Errorf("ошибка переоткрытия: %w", err)
	}

	n, _ := res.RowsAffected()

	if n == 0 {
		return ErrNotFound
	}

	return nil
}
