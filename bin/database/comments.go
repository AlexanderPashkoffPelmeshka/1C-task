package database

import (
	"fmt"
	"time"

	"diary/models"
)

func (s *Store) AddComment(c *models.Comment) (int64, error) {
	res, err := s.db.Exec(`
		insert into comments (case_id, body, created_at) 
		values (?, ?, ?)`,
		c.CaseID, c.Body, toDB(time.Now().UTC()),
	)

	if err != nil {
		return 0, fmt.Errorf("ошибка добавления комментария: %w", err)
	}

	return res.LastInsertId()
}

func (s *Store) ListComments(caseID int64) ([]*models.Comment, error) {
	rows, err := s.db.Query(`
		select id, case_id, body, created_at 
		from comments where case_id = ? order by created_at asc`, caseID)

	if err != nil {
		return nil, fmt.Errorf("ошибка списка комментариев: %w", err)
	}

	defer rows.Close()

	var result []*models.Comment

	for rows.Next() {
		var c models.Comment
		var createdStr string

		err := rows.Scan(&c.ID, &c.CaseID, &c.Body, &createdStr)

		if err != nil {
			return nil, err
		}

		c.CreatedAt, _ = fromDB(createdStr)
		result = append(result, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении комментариев: %w", err)
	}

	return result, nil
}
