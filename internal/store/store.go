// Package store 提供 SQLite 持久化：建表迁移与各实体仓库。
//
// 数据库为单文件 SQLite（modernc.org/sqlite 纯 Go 驱动，CGO 无关），
// 通过 PRAGMA foreign_keys、journal_mode=WAL 与 busy_timeout 保证并发安全；
// 关闭后重开同一路径即可恢复全部数据，是 --smoke-test 重启恢复验证的基础。
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store 聚合全部实体仓库，是 service 层唯一的持久化入口。
type Store struct {
	DB         *sql.DB
	Batches    *BatchStore
	Samples    *SampleStore
	Motifs     *MotifStore
	Techniques *TechniqueStore
	Relations  *RelationStore
	Evidence   *EvidenceStore
	Versions   *VersionStore
}

// Open 打开（必要时创建）SQLite 数据库并执行迁移。
func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir db dir: %w", err)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite 单写者，串行化事务
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		db.Close()
		return nil, fmt.Errorf("set wal: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		db.Close()
		return nil, fmt.Errorf("set fk: %w", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout=5000`); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	s := &Store{DB: db}
	s.Batches = &BatchStore{db: db}
	s.Samples = &SampleStore{db: db}
	s.Motifs = &MotifStore{db: db}
	s.Techniques = &TechniqueStore{db: db}
	s.Relations = &RelationStore{db: db}
	s.Evidence = &EvidenceStore{db: db}
	s.Versions = &VersionStore{db: db}
	return s, nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error {
	return s.DB.Close()
}

// migrate 执行幂等建表迁移（CREATE TABLE IF NOT EXISTS）。
func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS batches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS samples (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL REFERENCES batches(id),
			name TEXT NOT NULL,
			provenance TEXT NOT NULL DEFAULT '',
			warp_count INTEGER NOT NULL,
			weft_count INTEGER NOT NULL,
			warp_density INTEGER NOT NULL,
			weft_density INTEGER NOT NULL,
			sha256 TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(sha256)
		)`,
		`CREATE TABLE IF NOT EXISTS motifs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sample_id INTEGER NOT NULL REFERENCES samples(id),
			name TEXT NOT NULL,
			origin_x INTEGER NOT NULL,
			origin_y INTEGER NOT NULL,
			width INTEGER NOT NULL,
			height INTEGER NOT NULL,
			grid TEXT NOT NULL,
			period_x INTEGER NOT NULL DEFAULT 0,
			period_y INTEGER NOT NULL DEFAULT 0,
			symmetry TEXT NOT NULL DEFAULT 'none',
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS techniques (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sample_id INTEGER NOT NULL UNIQUE REFERENCES samples(id),
			weave_class TEXT NOT NULL,
			interlacing TEXT NOT NULL,
			twill_direction TEXT NOT NULL DEFAULT 'left',
			yarn_twist TEXT NOT NULL DEFAULT '',
			dye_class TEXT NOT NULL,
			dye_pigment TEXT NOT NULL DEFAULT '',
			colorfastness INTEGER NOT NULL DEFAULT 3,
			carbon_ratio REAL NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS relations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_unit_id INTEGER NOT NULL REFERENCES motifs(id),
			to_unit_id INTEGER NOT NULL REFERENCES motifs(id),
			topo_similarity REAL NOT NULL,
			weave_compat TEXT NOT NULL,
			dye_compat TEXT NOT NULL,
			verdict TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			version_id INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			UNIQUE(from_unit_id, to_unit_id)
		)`,
		`CREATE TABLE IF NOT EXISTS counter_evidence (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			relation_id INTEGER NOT NULL REFERENCES relations(id),
			kind TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			ref TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS relation_versions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			relation_ids TEXT NOT NULL DEFAULT '',
			provenance_note TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			frozen_at TEXT NOT NULL DEFAULT '',
			superseded_by INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_samples_batch ON samples(batch_id)`,
		`CREATE INDEX IF NOT EXISTS idx_motifs_sample ON motifs(sample_id)`,
		`CREATE INDEX IF NOT EXISTS idx_relations_version ON relations(version_id)`,
		`CREATE INDEX IF NOT EXISTS idx_evidence_relation ON counter_evidence(relation_id)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
