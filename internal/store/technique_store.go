package store

import (
	"database/sql"
	"time"

	"task266-textilerelation/internal/model"
)

// TechniqueStore 持久化样本工艺特征（每样本一条）。
type TechniqueStore struct{ db *sql.DB }

// Upsert 写入/更新样本工艺特征（sample_id 唯一）。
func (t *TechniqueStore) Upsert(f *model.TechniqueFeature) (*model.TechniqueFeature, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := t.db.Exec(
		`INSERT INTO techniques(sample_id, weave_class, interlacing, twill_direction,
		 yarn_twist, dye_class, dye_pigment, colorfastness, carbon_ratio, updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?)
		 ON CONFLICT(sample_id) DO UPDATE SET
		   weave_class=excluded.weave_class,
		   interlacing=excluded.interlacing,
		   twill_direction=excluded.twill_direction,
		   yarn_twist=excluded.yarn_twist,
		   dye_class=excluded.dye_class,
		   dye_pigment=excluded.dye_pigment,
		   colorfastness=excluded.colorfastness,
		   carbon_ratio=excluded.carbon_ratio,
		   updated_at=excluded.updated_at`,
		f.SampleID, f.WeaveClass, f.Interlacing, f.TwillDirection,
		f.YarnTwist, f.DyeClass, f.DyePigment, f.Colorfastness, f.CarbonRatio, now)
	if err != nil {
		return nil, err
	}
	return t.BySample(f.SampleID)
}

// BySample 按样本查询工艺特征。
func (t *TechniqueStore) BySample(sampleID int64) (*model.TechniqueFeature, error) {
	row := t.db.QueryRow(
		`SELECT id, sample_id, weave_class, interlacing, twill_direction, yarn_twist,
		 dye_class, dye_pigment, colorfastness, carbon_ratio, updated_at
		 FROM techniques WHERE sample_id=?`, sampleID)
	var f model.TechniqueFeature
	var updated string
	if err := row.Scan(&f.ID, &f.SampleID, &f.WeaveClass, &f.Interlacing,
		&f.TwillDirection, &f.YarnTwist, &f.DyeClass, &f.DyePigment,
		&f.Colorfastness, &f.CarbonRatio, &updated); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	f.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	return &f, nil
}

// ListBySamples 批量查询样本工艺特征，缺失项返回 ErrTechnique。
func (t *TechniqueStore) ListBySamples(sampleIDs []int64) (map[int64]*model.TechniqueFeature, error) {
	out := make(map[int64]*model.TechniqueFeature, len(sampleIDs))
	for _, sid := range sampleIDs {
		f, err := t.BySample(sid)
		if err != nil {
			if err == model.ErrNotFound {
				continue // 技法缺失项在验证层显式处理
			}
			return nil, err
		}
		out[sid] = f
	}
	return out, nil
}

// List 列出全部工艺特征。
func (t *TechniqueStore) List() ([]*model.TechniqueFeature, error) {
	rows, err := t.db.Query(
		`SELECT id, sample_id, weave_class, interlacing, twill_direction, yarn_twist,
		 dye_class, dye_pigment, colorfastness, carbon_ratio, updated_at
		 FROM techniques ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.TechniqueFeature
	for rows.Next() {
		var f model.TechniqueFeature
		var updated string
		if err := rows.Scan(&f.ID, &f.SampleID, &f.WeaveClass, &f.Interlacing,
			&f.TwillDirection, &f.YarnTwist, &f.DyeClass, &f.DyePigment,
			&f.Colorfastness, &f.CarbonRatio, &updated); err != nil {
			return nil, err
		}
		f.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, &f)
	}
	return out, rows.Err()
}
