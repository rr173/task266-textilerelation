package store

import (
	"database/sql"
	"time"

	"task266-textilerelation/internal/model"
)

// RelationStore 持久化工艺关系候选。
type RelationStore struct{ db *sql.DB }

// Create 插入候选（(from,to) 唯一，重复时返回 ErrConflict）。
func (r *RelationStore) Create(m *model.RelationCandidate) (*model.RelationCandidate, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.Exec(
		`INSERT INTO relations(from_unit_id, to_unit_id, topo_similarity,
		 weave_compat, dye_compat, verdict, summary, version_id, created_at, updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		m.FromUnitID, m.ToUnitID, m.TopoSimilarity, m.WeaveCompat, m.DyeCompat,
		m.Verdict, m.Summary, m.VersionID, now, now)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrConflict
		}
		return nil, err
	}
	return r.GetByUnits(m.FromUnitID, m.ToUnitID)
}

// Get 按 ID 查询候选。
func (r *RelationStore) Get(id int64) (*model.RelationCandidate, error) {
	row := r.db.QueryRow(
		`SELECT id, from_unit_id, to_unit_id, topo_similarity, weave_compat,
		 dye_compat, verdict, summary, version_id, created_at, updated_at
		 FROM relations WHERE id=?`, id)
	return scanRelation(row)
}

// GetByUnits 按单元对查询候选（幂等定位）。
func (r *RelationStore) GetByUnits(fromID, toID int64) (*model.RelationCandidate, error) {
	row := r.db.QueryRow(
		`SELECT id, from_unit_id, to_unit_id, topo_similarity, weave_compat,
		 dye_compat, verdict, summary, version_id, created_at, updated_at
		 FROM relations WHERE from_unit_id=? AND to_unit_id=?`, fromID, toID)
	return scanRelation(row)
}

// List 列出全部候选（可按裁决过滤）。
func (r *RelationStore) List(verdict string) ([]*model.RelationCandidate, error) {
	var rows *sql.Rows
	var err error
	if verdict == "" {
		rows, err = r.db.Query(
			`SELECT id, from_unit_id, to_unit_id, topo_similarity, weave_compat,
			 dye_compat, verdict, summary, version_id, created_at, updated_at
			 FROM relations ORDER BY topo_similarity DESC`)
	} else {
		rows, err = r.db.Query(
			`SELECT id, from_unit_id, to_unit_id, topo_similarity, weave_compat,
			 dye_compat, verdict, summary, version_id, created_at, updated_at
			 FROM relations WHERE verdict=? ORDER BY topo_similarity DESC`, verdict)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.RelationCandidate
	for rows.Next() {
		var m model.RelationCandidate
		var created, updated string
		if err := rows.Scan(&m.ID, &m.FromUnitID, &m.ToUnitID, &m.TopoSimilarity,
			&m.WeaveCompat, &m.DyeCompat, &m.Verdict, &m.Summary, &m.VersionID,
			&created, &updated); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, created)
		m.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, &m)
	}
	return out, rows.Err()
}

// ListByVersion 列出版本内全部候选。
func (r *RelationStore) ListByVersion(versionID int64) ([]*model.RelationCandidate, error) {
	rows, err := r.db.Query(
		`SELECT id, from_unit_id, to_unit_id, topo_similarity, weave_compat,
		 dye_compat, verdict, summary, version_id, created_at, updated_at
		 FROM relations WHERE version_id=? ORDER BY id`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.RelationCandidate
	for rows.Next() {
		var m model.RelationCandidate
		var created, updated string
		if err := rows.Scan(&m.ID, &m.FromUnitID, &m.ToUnitID, &m.TopoSimilarity,
			&m.WeaveCompat, &m.DyeCompat, &m.Verdict, &m.Summary, &m.VersionID,
			&created, &updated); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, created)
		m.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, &m)
	}
	return out, rows.Err()
}

// AssignVersion 将候选收录进版本（版本冻结后禁止调用，由 service 层把关）。
func (r *RelationStore) AssignVersion(id, versionID int64) error {
	res, err := r.db.Exec(`UPDATE relations SET version_id=? WHERE id=?`, versionID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// UpdateVerdict 条件更新裁决：仅当当前裁决与 expect 一致时才写入（版本校验，防并发覆盖）。
func (r *RelationStore) UpdateVerdict(id int64, expect, verdict, summary string) error {
	res, err := r.db.Exec(
		`UPDATE relations SET verdict=?, summary=?, updated_at=?
		 WHERE id=? AND verdict=?`,
		verdict, summary, time.Now().UTC().Format(time.RFC3339), id, expect)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		cur, gerr := r.Get(id)
		if gerr != nil {
			return gerr
		}
		return &StateMismatchError{ID: id, Current: cur.Verdict, Want: expect}
	}
	return nil
}

// UpdateWeaveDye 回写验证结果（topo/weave/dye 三要素）。
func (r *RelationStore) UpdateWeaveDye(id int64, weaveCompat, dyeCompat, verdict string) error {
	res, err := r.db.Exec(
		`UPDATE relations SET weave_compat=?, dye_compat=?, verdict=?, updated_at=?
		 WHERE id=?`,
		weaveCompat, dyeCompat, verdict, time.Now().UTC().Format(time.RFC3339), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// Stats 返回候选统计（用于 /api/stats）。
func (r *RelationStore) Stats() (map[string]int, error) {
	out := map[string]int{}
	rows, err := r.db.Query(`SELECT verdict, COUNT(*) FROM relations GROUP BY verdict`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		var n int
		if err := rows.Scan(&v, &n); err != nil {
			return nil, err
		}
		out[v] = n
	}
	return out, rows.Err()
}

func scanRelation(row *sql.Row) (*model.RelationCandidate, error) {
	var m model.RelationCandidate
	var created, updated string
	if err := row.Scan(&m.ID, &m.FromUnitID, &m.ToUnitID, &m.TopoSimilarity,
		&m.WeaveCompat, &m.DyeCompat, &m.Verdict, &m.Summary, &m.VersionID,
		&created, &updated); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, created)
	m.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &m, nil
}
