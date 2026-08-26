package store

import (
	"database/sql"
	"time"

	"task266-textilerelation/internal/model"
)

// MotifStore 持久化纹样单元。
type MotifStore struct{ db *sql.DB }

// Create 插入纹样单元（初始 pending）。
func (m *MotifStore) Create(u *model.MotifUnit) (*model.MotifUnit, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := m.db.Exec(
		`INSERT INTO motifs(sample_id, name, origin_x, origin_y, width, height,
		 grid, period_x, period_y, symmetry, status, created_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		u.SampleID, u.Name, u.OriginX, u.OriginY, u.Width, u.Height,
		u.Grid, u.PeriodX, u.PeriodY, u.Symmetry, u.Status, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return m.Get(id)
}

// Get 按 ID 查询纹样单元。
func (m *MotifStore) Get(id int64) (*model.MotifUnit, error) {
	row := m.db.QueryRow(
		`SELECT id, sample_id, name, origin_x, origin_y, width, height, grid,
		 period_x, period_y, symmetry, status, created_at FROM motifs WHERE id=?`, id)
	var u model.MotifUnit
	var created string
	if err := row.Scan(&u.ID, &u.SampleID, &u.Name, &u.OriginX, &u.OriginY,
		&u.Width, &u.Height, &u.Grid, &u.PeriodX, &u.PeriodY, &u.Symmetry,
		&u.Status, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &u, nil
}

// ListBySample 列出样本的全部纹样单元。
func (m *MotifStore) ListBySample(sampleID int64) ([]*model.MotifUnit, error) {
	rows, err := m.db.Query(
		`SELECT id, sample_id, name, origin_x, origin_y, width, height, grid,
		 period_x, period_y, symmetry, status, created_at
		 FROM motifs WHERE sample_id=? ORDER BY id`, sampleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMotifs(rows)
}

// List 列出全部纹样单元。
func (m *MotifStore) List() ([]*model.MotifUnit, error) {
	rows, err := m.db.Query(
		`SELECT id, sample_id, name, origin_x, origin_y, width, height, grid,
		 period_x, period_y, symmetry, status, created_at FROM motifs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMotifs(rows)
}

// UpdateParsed 回填拓扑解析结果（周期、对称），并标记状态。
func (m *MotifStore) UpdateParsed(id int64, periodX, periodY int, symmetry, status string) error {
	res, err := m.db.Exec(
		`UPDATE motifs SET period_x=?, period_y=?, symmetry=?, status=? WHERE id=?`,
		periodX, periodY, symmetry, status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// UpdateStatus 更新单元状态（valid/patch/excluded）。
func (m *MotifStore) UpdateStatus(id int64, status string) error {
	res, err := m.db.Exec(`UPDATE motifs SET status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func scanMotifs(rows *sql.Rows) ([]*model.MotifUnit, error) {
	var out []*model.MotifUnit
	for rows.Next() {
		var u model.MotifUnit
		var created string
		if err := rows.Scan(&u.ID, &u.SampleID, &u.Name, &u.OriginX, &u.OriginY,
			&u.Width, &u.Height, &u.Grid, &u.PeriodX, &u.PeriodY, &u.Symmetry,
			&u.Status, &created); err != nil {
			return nil, err
		}
		u.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, &u)
	}
	return out, rows.Err()
}
