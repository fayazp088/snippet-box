package models

import (
	"database/sql"
	"log"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title, content string, expires int) (int, error) {
	query := `INSERT INTO snippets (title, content, created, expires) 
           VALUES (?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(query, title, content, expires)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		log.Print("2nd err", err)
		return 0, err
	}

	return int(id), nil
}

func (m *SnippetModel) Get(id int) (Snippet, error) {
	return Snippet{}, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	return nil, nil
}
