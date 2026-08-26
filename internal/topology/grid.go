// Package topology 实现纹样单元的网格拓扑解析与比较。
//
// 纹样单元的网格用 '#'（浮色/经浮）与 '.'（底/纬浮）表达经向-纬向的
// 二值交错图案。本包负责：
//   - Parse：把编码字符串解析为 [][]bool 矩阵；
//   - DetectPeriod：检测平铺周期（最小平移使图案自匹配）；
//   - DetectSymmetry：检测水平/垂直/180° 对称；
//   - Compare：比较两个单元的拓扑相似度（网格重叠 + 结构一致性）。
//
// 拓扑相似度是工艺传承关系判定的第一证据：外观（网格）相似是候选前提，
// 但最终裁决必须叠加 technique 包的经纬交错与染料证据。
package topology

import (
	"fmt"
	"strings"

	"task266-textilerelation/internal/model"
)

// Grid 是纹样单元的二维二值矩阵：rows x cols，true 表示浮色格。
type Grid struct {
	Rows int
	Cols int
	Cells [][]bool
}

// Parse 把模型网格编码解析为 Grid；编码格式由 model.ValidateGridString 保证。
func Parse(encoded string) (*Grid, error) {
	rows := strings.Split(encoded, ",")
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: empty grid", model.ErrInvalid)
	}
	cols := len(rows[0])
	cells := make([][]bool, len(rows))
	for i, row := range rows {
		if len(row) != cols {
			return nil, fmt.Errorf("%w: ragged grid row %d", model.ErrInvalid, i)
		}
		cells[i] = make([]bool, cols)
		for j, ch := range row {
			cells[i][j] = ch == '#'
		}
	}
	return &Grid{Rows: len(rows), Cols: cols, Cells: cells}, nil
}

// Encode 把 Grid 转回模型编码字符串。
func (g *Grid) Encode() string {
	var sb strings.Builder
	for i := 0; i < g.Rows; i++ {
		for j := 0; j < g.Cols; j++ {
			if g.Cells[i][j] {
				sb.WriteByte('#')
			} else {
				sb.WriteByte('.')
			}
		}
		if i < g.Rows-1 {
			sb.WriteByte(',')
		}
	}
	return sb.String()
}

// Density 返回浮色格占比（0~1）。
func (g *Grid) Density() float64 {
	if g.Rows == 0 || g.Cols == 0 {
		return 0
	}
	on := 0
	for _, row := range g.Cells {
		for _, c := range row {
			if c {
				on++
			}
		}
	}
	return float64(on) / float64(g.Rows*g.Cols)
}

// rowMatchRate 比较两行（支持循环位移），返回相同比例。
func rowMatchRate(a, b []bool) float64 {
	n := len(a)
	if n == 0 {
		return 1
	}
	match := 0
	for i := 0; i < n; i++ {
		if a[i] == b[i] {
			match++
		}
	}
	return float64(match) / float64(n)
}

// colMatchRate 比较两列，返回相同比例。
func colMatchRate(g *Grid, ca, cb int) float64 {
	if g.Rows == 0 {
		return 1
	}
	match := 0
	for i := 0; i < g.Rows; i++ {
		if g.Cells[i][ca] == g.Cells[i][cb] {
			match++
		}
	}
	return float64(match) / float64(g.Rows)
}

// OverlapRate 计算两网格在共同区域上的逐格一致率（归一化尺寸后）。
func OverlapRate(a, b *Grid) float64 {
	if a.Rows == 0 || a.Cols == 0 || b.Rows == 0 || b.Cols == 0 {
		return 0
	}
	rows := a.Rows
	if b.Rows < rows {
		rows = b.Rows
	}
	cols := a.Cols
	if b.Cols < cols {
		cols = b.Cols
	}
	match := 0
	total := rows * cols
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if a.Cells[i][j] == b.Cells[i][j] {
				match++
			}
		}
	}
	return float64(match) / float64(total)
}
