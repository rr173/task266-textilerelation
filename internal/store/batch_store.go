package store

import (
	"database/sql"
	"fmt"
	"time"

	"task266-textilerelation/internal/model"
)

// BatchStore 持久化样本批次。
type BatchStore struct{ db *sql.DB }

// Create 新建批次（pending 初始态）。
func (b *BatchStore) Create(name, note string) (*model.Batch, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := b.db.Exec(
		`INSERT INTO batches(name, note, status, created_at) VALUES(?,?,?,?)`,
		name, note, model.BatchPending, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return b.Get(id)
}

// Get 按 ID 查询批次。
func (b *BatchStore) Get(id int64) (*model.Batch, error) {
	row := b.db.QueryRow(
		`SELECT id, name, note, status, created_at FROM batches WHERE id=?`, id)
	var m model.Batch
	var created string
	if err := row.Scan(&m.ID, &m.Name, &m.Note, &m.Status, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &m, nil
}

// List 按创建时间列出全部批次。
func (b *BatchStore) List() ([]*model.Batch, error) {
	rows, err := b.db.Query(
		`SELECT id, name, note, status, created_at FROM batches ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Batch
	for rows.Next() {
		var m model.Batch
		var created string
		if err := rows.Scan(&m.ID, &m.Name, &m.Note, &m.Status, &created); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, &m)
	}
	return out, rows.Err()
}

// UpdateStatus 流转批次状态；status 非空时校验为合法状态。
func (b *BatchStore) UpdateStatus(id int64, status string) error {
	if status == "" {
		return fmt.Errorf("%w: status is required", model.ErrInvalid)
	}
	res, err := b.db.Exec(`UPDATE batches SET status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// CountSamples 统计批次内样本数（用于状态机流转校验）。
func (b *BatchStore) CountSamples(id int64) (int, error) {
	var n int
	err := b.db.QueryRow(`SELECT COUNT(*) FROM samples WHERE batch_id=?`, id).Scan(&n)
	return n, err
}
