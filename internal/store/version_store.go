package store

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"task266-textilerelation/internal/model"
)

// VersionStore 持久化关系版本。
type VersionStore struct{ db *sql.DB }

// Create 新建草稿版本。
func (v *VersionStore) Create(name, summary, relationIDs, provenanceNote string) (*model.RelationVersion, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := v.db.Exec(
		`INSERT INTO relation_versions(name, status, summary, relation_ids,
		 provenance_note, created_at, frozen_at, superseded_by)
		 VALUES(?,?,?,?,?,?,?,?)`,
		name, model.VersionDraft, summary, relationIDs, provenanceNote, now, "", 0)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return v.Get(id)
}

// Get 按 ID 查询版本。
func (v *VersionStore) Get(id int64) (*model.RelationVersion, error) {
	row := v.db.QueryRow(
		`SELECT id, name, status, summary, relation_ids, provenance_note,
		 created_at, frozen_at, superseded_by FROM relation_versions WHERE id=?`, id)
	var m model.RelationVersion
	var created, frozen string
	if err := row.Scan(&m.ID, &m.Name, &m.Status, &m.Summary, &m.RelationIDs,
		&m.ProvenanceNote, &created, &frozen, &m.SupersededBy); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, created)
	if frozen != "" {
		m.FrozenAt, _ = time.Parse(time.RFC3339, frozen)
	}
	return &m, nil
}

// List 列出全部版本（新→旧）。
func (v *VersionStore) List() ([]*model.RelationVersion, error) {
	rows, err := v.db.Query(
		`SELECT id, name, status, summary, relation_ids, provenance_note,
		 created_at, frozen_at, superseded_by FROM relation_versions ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.RelationVersion
	for rows.Next() {
		var m model.RelationVersion
		var created, frozen string
		if err := rows.Scan(&m.ID, &m.Name, &m.Status, &m.Summary, &m.RelationIDs,
			&m.ProvenanceNote, &created, &frozen, &m.SupersededBy); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, created)
		if frozen != "" {
			m.FrozenAt, _ = time.Parse(time.RFC3339, frozen)
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

// UpdateStatus 流转版本状态。
func (v *VersionStore) UpdateStatus(id int64, status string, frozenAt string) error {
	res, err := v.db.Exec(
		`UPDATE relation_versions SET status=?, frozen_at=? WHERE id=?`,
		status, frozenAt, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// MarkSuperseded 将版本标记为被 newID 替代。
func (v *VersionStore) MarkSuperseded(id, newID int64) error {
	res, err := v.db.Exec(
		`UPDATE relation_versions SET status=?, superseded_by=? WHERE id=?`,
		model.VersionSuperseded, newID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// ParseRelationIDs 将逗号分隔的关系 ID 字符串解析为 int64 切片。
func ParseRelationIDs(s string) []int64 {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		if id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
			out = append(out, id)
		}
	}
	return out
}
