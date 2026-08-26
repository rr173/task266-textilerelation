package topology

import (
	"fmt"

	"task266-textilerelation/internal/model"
)

// Comparison 是一次单元拓扑比较的完整结果。
type Comparison struct {
	// OverlapRate 归一化网格逐格一致率（0~1）。
	OverlapRate float64 `json:"overlap_rate"`
	// StructureScore 结构一致性：周期（若两者都有周期则比较周期组合）+ 对称族。
	StructureScore float64 `json:"structure_score"`
	// TopoSimilarity 综合拓扑相似度 = 0.7*OverlapRate + 0.3*StructureScore。
	TopoSimilarity float64 `json:"topo_similarity"`
	// Reason 人类可读的判定说明。
	Reason string `json:"reason"`
}

// CompareMotifs 比较两个纹样单元的拓扑，返回综合相似度。
func CompareMotifs(a, b *model.MotifUnit) (*Comparison, error) {
	ga, err := Parse(a.Grid)
	if err != nil {
		return nil, fmt.Errorf("parse motif %d: %w", a.ID, err)
	}
	gb, err := Parse(b.Grid)
	if err != nil {
		return nil, fmt.Errorf("parse motif %d: %w", b.ID, err)
	}
	if a.Status != model.MotifValid || b.Status != model.MotifValid {
		return nil, fmt.Errorf("%w: motifs must be valid (a=%s b=%s)",
			model.ErrState, a.Status, b.Status)
	}

	overlap := OverlapRate(ga, gb)
	structure := structureScore(a, ga, b, gb)
	sim := 0.7*overlap + 0.3*structure

	reason := fmt.Sprintf(
		"网格一致率 %.2f，结构一致性 %.2f，综合拓扑相似度 %.2f", overlap, structure, sim)
	return &Comparison{
		OverlapRate:    overlap,
		StructureScore: structure,
		TopoSimilarity: sim,
		Reason:         reason,
	}, nil
}

// structureScore 结构一致性：周期组合一致 + 对称族一致 + 尺寸比接近。
func structureScore(a *model.MotifUnit, ga *Grid, b *model.MotifUnit, gb *Grid) float64 {
	score := 0.0

	// 周期一致性：两者都无周期，或周期组合相同。
	aPeriodic := a.PeriodX > 0 && a.PeriodY > 0
	bPeriodic := b.PeriodX > 0 && b.PeriodY > 0
	if aPeriodic == bPeriodic {
		if aPeriodic {
			if a.PeriodX == b.PeriodX && a.PeriodY == b.PeriodY {
				score += 0.5
			} else {
				score += 0.2 // 周期不同但都是平铺纹样
			}
		} else {
			score += 0.5 // 都无周期
		}
	} else {
		score += 0.1
	}

	// 对称族一致性。
	if SymmetryCompatible(a.Symmetry, b.Symmetry) {
		score += 0.3
	}

	// 尺寸比接近（长宽比差 < 40%）。
	arA := float64(ga.Cols) / float64(ga.Rows)
	arB := float64(gb.Cols) / float64(gb.Rows)
	if arA > arB {
		if arA <= arB*1.4 {
			score += 0.2
		}
	} else if arB <= arA*1.4 {
		score += 0.2
	}

	return score
}
