package sample

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"task266-textilerelation/internal/model"
	"task266-textilerelation/internal/store"
)

// IngestInput 是样本导入请求。
type IngestInput struct {
	BatchID     int64
	Name        string
	Provenance  string
	WarpCount   int
	WeftCount   int
	WarpDensity int
	WeftDensity int
	// MotifGrids 附带导入的纹样单元网格（可空）。
	MotifGrids map[string]string // name -> grid 编码
}

// Ingest 导入织物样本：校验 → 计算内容哈希 → 幂等写入 → 附带单元。
//
// 返回 (样本, created, error)：created=false 表示命中了既有哈希。
func Ingest(samples *store.SampleStore, motifs *store.MotifStore, in IngestInput) (*model.FabricSample, bool, error) {
	if err := model.ValidateSample(in.BatchID, in.Name, in.Provenance,
		in.WarpCount, in.WeftCount, in.WarpDensity, in.WeftDensity); err != nil {
		return nil, false, err
	}
	hash := contentHash(in)
	existing, err := samples.ByHash(hash)
	if err == nil {
		return existing, false, nil
	}
	if err != model.ErrNotFound {
		return nil, false, err
	}

	m := &model.FabricSample{
		BatchID:     in.BatchID,
		Name:        in.Name,
		Provenance:  in.Provenance,
		WarpCount:   in.WarpCount,
		WeftCount:   in.WeftCount,
		WarpDensity: in.WarpDensity,
		WeftDensity: in.WeftDensity,
		SHA256:      hash,
		Status:      model.MotifPending,
	}
	created, err := samples.Create(m)
	if err != nil {
		return nil, false, err
	}

	// 附带导入的单元（按名称排序保证确定性）。
	names := make([]string, 0, len(in.MotifGrids))
	for n := range in.MotifGrids {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		grid := in.MotifGrids[n]
		rows := strings.Split(grid, ",")
		if len(rows) == 0 {
			continue
		}
		_, _ = motifs.Create(&model.MotifUnit{
			SampleID: created.ID,
			Name:     n,
			OriginX:  0,
			OriginY:  0,
			Width:    len(rows[0]),
			Height:   len(rows),
			Grid:     grid,
			Status:   model.MotifPending,
		})
	}
	return created, true, nil
}

// contentHash 计算样本内容哈希：名称+出处+经纬参数+附带单元网格。
//
// 只对"结构特征"做摘要，不含 ID/时间，保证幂等。
func contentHash(in IngestInput) string {
	names := make([]string, 0, len(in.MotifGrids))
	for n := range in.MotifGrids {
		names = append(names, n)
	}
	sort.Strings(names)
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s|%s|%d|%d|%d|%d", in.Name, in.Provenance,
		in.WarpCount, in.WeftCount, in.WarpDensity, in.WeftDensity)
	for _, n := range names {
		fmt.Fprintf(&sb, "|%s=%s", n, in.MotifGrids[n])
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:])
}
