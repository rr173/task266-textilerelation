package store

import (
	"database/sql"
	"fmt"
	"time"

	"task266-textilerelation/internal/model"
)

// StateMismatchError 表示并发裁决时的版本校验失败：
// 期望的旧裁决已被其他请求改写。
type StateMismatchError struct {
	ID      int64
	Current string
	Want    string
}

func (e *StateMismatchError) Error() string {
	return fmt.Sprintf("state mismatch: relation %d current=%s want=%s", e.ID, e.Current, e.Want)
}

// EvidenceStore 持久化反证条目。
type EvidenceStore struct{ db *sql.DB }

// Create 追加一条反证。
func (e *EvidenceStore) Create(m *model.CounterEvidence) (*model.CounterEvidence, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := e.db.Exec(
		`INSERT INTO counter_evidence(relation_id, kind, description, ref, created_at)
		 VALUES(?,?,?,?,?)`,
		m.RelationID, m.Kind, m.Description, m.Ref, now)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return e.Get(id)
}

// Get 按 ID 查询反证。
func (e *EvidenceStore) Get(id int64) (*model.CounterEvidence, error) {
	row := e.db.QueryRow(
		`SELECT id, relation_id, kind, description, ref, created_at
		 FROM counter_evidence WHERE id=?`, id)
	var m model.CounterEvidence
	var created string
	if err := row.Scan(&m.ID, &m.RelationID, &m.Kind, &m.Description, &m.Ref, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	m.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &m, nil
}

// ListByRelation 列出关系的全部反证。
func (e *EvidenceStore) ListByRelation(relationID int64) ([]*model.CounterEvidence, error) {
	rows, err := e.db.Query(
		`SELECT id, relation_id, kind, description, ref, created_at
		 FROM counter_evidence WHERE relation_id=? ORDER BY id`, relationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.CounterEvidence
	for rows.Next() {
		var m model.CounterEvidence
		var created string
		if err := rows.Scan(&m.ID, &m.RelationID, &m.Kind, &m.Description, &m.Ref, &created); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, &m)
	}
	return out, rows.Err()
}

// CountByRelation 统计某关系的反证数量（用于证据强度评估）。
func (e *EvidenceStore) CountByRelation(relationID int64) (int, error) {
	var n int
	err := e.db.QueryRow(
		`SELECT COUNT(*) FROM counter_evidence WHERE relation_id=?`, relationID).Scan(&n)
	return n, err
}

// ListByRelations 批量取反证：返回 relationID -> []evidence。
func (e *EvidenceStore) ListByRelations(relationIDs []int64) (map[int64][]*model.CounterEvidence, error) {
	out := make(map[int64][]*model.CounterEvidence)
	for _, rid := range relationIDs {
		list, err := e.ListByRelation(rid)
		if err != nil {
			return nil, err
		}
		if len(list) > 0 {
			out[rid] = list
		}
	}
	return out, nil
}
