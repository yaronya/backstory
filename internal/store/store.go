package store

import (
	"database/sql"
	"time"

	"github.com/backstory-team/backstory/internal/decision"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func createSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS decisions (
			file_path    TEXT PRIMARY KEY,
			type         TEXT NOT NULL,
			date         TEXT NOT NULL,
			author       TEXT NOT NULL,
			anchor       TEXT NOT NULL,
			linear_issue TEXT,
			stale        INTEGER NOT NULL DEFAULT 0,
			title        TEXT NOT NULL,
			body         TEXT NOT NULL
		);

		CREATE VIRTUAL TABLE IF NOT EXISTS decisions_fts USING fts5(
			title, body, anchor,
			content='decisions',
			content_rowid='rowid'
		);

		CREATE TRIGGER IF NOT EXISTS decisions_ai AFTER INSERT ON decisions BEGIN
			INSERT INTO decisions_fts(rowid, title, body, anchor)
			VALUES (new.rowid, new.title, new.body, new.anchor);
		END;

		CREATE TRIGGER IF NOT EXISTS decisions_ad AFTER DELETE ON decisions BEGIN
			INSERT INTO decisions_fts(decisions_fts, rowid, title, body, anchor)
			VALUES ('delete', old.rowid, old.title, old.body, old.anchor);
		END;

		CREATE TRIGGER IF NOT EXISTS decisions_au AFTER UPDATE ON decisions BEGIN
			INSERT INTO decisions_fts(decisions_fts, rowid, title, body, anchor)
			VALUES ('delete', old.rowid, old.title, old.body, old.anchor);
			INSERT INTO decisions_fts(rowid, title, body, anchor)
			VALUES (new.rowid, new.title, new.body, new.anchor);
		END;
	`)
	return err
}

func (s *Store) Upsert(d *decision.Decision) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM decisions WHERE file_path = ?`, d.FilePath)
	if err != nil {
		return err
	}

	stale := 0
	if d.Stale {
		stale = 1
	}

	_, err = tx.Exec(`
		INSERT INTO decisions (file_path, type, date, author, anchor, linear_issue, stale, title, body)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.FilePath,
		d.Type,
		d.DateStr,
		d.Author,
		d.Anchor,
		d.LinearIssue,
		stale,
		d.Title,
		d.Body,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) QueryByAnchor(anchor string) ([]*decision.Decision, error) {
	rows, err := s.db.Query(`
		SELECT file_path, type, date, author, anchor, linear_issue, stale, title, body
		FROM decisions
		WHERE (anchor = ? OR anchor LIKE ? || '%' OR ? LIKE anchor || '%')
		  AND stale = 0
		ORDER BY date DESC`,
		anchor, anchor, anchor,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDecisions(rows)
}

func (s *Store) QueryByLinearIssue(issue string) ([]*decision.Decision, error) {
	rows, err := s.db.Query(`
		SELECT file_path, type, date, author, anchor, linear_issue, stale, title, body
		FROM decisions
		WHERE linear_issue = ?
		  AND stale = 0
		ORDER BY date DESC`,
		issue,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDecisions(rows)
}

func (s *Store) Search(query string) ([]*decision.Decision, error) {
	rows, err := s.db.Query(`
		SELECT d.file_path, d.type, d.date, d.author, d.anchor, d.linear_issue, d.stale, d.title, d.body
		FROM decisions_fts
		JOIN decisions d ON decisions_fts.rowid = d.rowid
		WHERE decisions_fts MATCH ?
		  AND d.stale = 0`,
		query,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDecisions(rows)
}

func scanDecisions(rows *sql.Rows) ([]*decision.Decision, error) {
	var results []*decision.Decision
	for rows.Next() {
		var d decision.Decision
		var stale int
		var linearIssue sql.NullString

		err := rows.Scan(
			&d.FilePath,
			&d.Type,
			&d.DateStr,
			&d.Author,
			&d.Anchor,
			&linearIssue,
			&stale,
			&d.Title,
			&d.Body,
		)
		if err != nil {
			return nil, err
		}

		if linearIssue.Valid {
			d.LinearIssue = linearIssue.String
		}
		d.Stale = stale != 0

		if d.DateStr != "" {
			parsed, err := time.Parse("2006-01-02", d.DateStr)
			if err == nil {
				d.Date = parsed
			}
		}

		results = append(results, &d)
	}
	return results, rows.Err()
}
