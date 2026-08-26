package store

import (
	"database/sql"
	"strings"
	"time"

	"task266-textilerelation/internal/model"
)

// SampleStore 持久化织物样本。
type SampleStore struct{ db *sql.DB }

// Create 插入样本，违反唯一哈希时返回 ErrConflict。
func (s *SampleStore) Create(m *model.FabricSample) (*model.FabricSample, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO samples(batch_id, name, provenance, warp_count, weft_count,
		 warp_density, weft_density, sha256, status, created_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		m.BatchID, m.Name, m.Provenance, m.WarpCount, m.WeftCount,
		m.WarpDensity, m.WeftDensity, m.SHA256, m.Status, now)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrConflict
		}
		return nil, err
	}
	return s.ByHash(m.SHA256)
}

// Get 按 ID 查询样本。
func (s *SampleStore) Get(id int64) (*model.FabricSample, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, name, provenance, warp_count, weft_count,
		 warp_density, weft_density, sha256, status, created_at FROM samples WHERE id=?`, id)
	var m model.FabricSample
	var created string
	if err := row.Scan(&m.ID, &m.BatchID, &m.Name, &m.Provenance, &m.WarpCount,
		&m.WeftCount, &m.WarpDensity, &m.WeftDensity, &m.SHA256, &m.Status, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &m, nil
}

// ByHash 按内容哈希幂等查询。
func (s *SampleStore) ByHash(sha string) (*model.FabricSample, error) {
	row := s.db.QueryRow(
		`SELECT id, batch_id, name, provenance, warp_count, weft_count,
		 warp_density, weft_density, sha256, status, created_at FROM samples WHERE sha256=?`, sha)
	var m model.FabricSample
	var created string
	if err := row.Scan(&m.ID, &m.BatchID, &m.Name, &m.Provenance, &m.WarpCount,
		&m.WeftCount, &m.WarpDensity, &m.WeftDensity, &m.SHA256, &m.Status, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &m, nil
}

// ListByBatch 列出批次内样本。
func (s *SampleStore) ListByBatch(batchID int64) ([]*model.FabricSample, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, name, provenance, warp_count, weft_count,
		 warp_density, weft_density, sha256, status, created_at
		 FROM samples WHERE batch_id=? ORDER BY id`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.FabricSample
	for rows.Next() {
		var m model.FabricSample
		var created string
		if err := rows.Scan(&m.ID, &m.BatchID, &m.Name, &m.Provenance, &m.WarpCount,
			&m.WeftCount, &m.WarpDensity, &m.WeftDensity, &m.SHA256, &m.Status, &created); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, &m)
	}
	return out, rows.Err()
}

// UpdateStatus 更新样本状态。
func (s *SampleStore) UpdateStatus(id int64, status string) error {
	res, err := s.db.Exec(`UPDATE samples SET status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// List 列出全部样本（跨批次）。
func (s *SampleStore) List() ([]*model.FabricSample, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, name, provenance, warp_count, weft_count,
		 warp_density, weft_density, sha256, status, created_at FROM samples ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.FabricSample
	for rows.Next() {
		var m model.FabricSample
		var created string
		if err := rows.Scan(&m.ID, &m.BatchID, &m.Name, &m.Provenance, &m.WarpCount,
			&m.WeftCount, &m.WarpDensity, &m.WeftDensity, &m.SHA256, &m.Status, &created); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, &m)
	}
	return out, rows.Err()
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, key := range []string{"UNIQUE constraint failed", "constraint failed"} {
		if strings.Contains(msg, key) {
			return true
		}
	}
	return false
}
