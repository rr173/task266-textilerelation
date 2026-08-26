package topology

// 周期检测：找最小水平/垂直平移，使网格与该平移后的自身高度自匹配。
//
// 纺织纹样的单元在织物表面沿经纬方向重复平铺；真实周期必然同时满足：
//   - 逐行比较：row[i] 与 row[i+shift] 的匹配率 ≥ matchThreshold；
//   - 平移后网格密度基本不变。
//
// 返回 0 表示未检测到周期（非平铺纹样），调用方按 PeriodX/PeriodY=0 处理。

const (
	// periodMatchThreshold 自匹配判定阈值。
	periodMatchThreshold = 0.85
	// minPeriod 允许的最小周期，排除 1（全同格无意义）。
	minPeriod = 2
)

// DetectPeriod 检测水平周期（经向）与垂直周期（纬向）。
func DetectPeriod(g *Grid) (periodX, periodY int) {
	periodX = detectHorizontal(g)
	periodY = detectVertical(g)
	return
}

// detectHorizontal 检测水平周期：平移 shift 列后逐行比较。
func detectHorizontal(g *Grid) int {
	for shift := minPeriod; shift <= g.Cols/2; shift++ {
		score := 0.0
		for i := 0; i < g.Rows; i++ {
			score += shiftMatchRate(g.Cells[i], shift)
		}
		if score/float64(g.Rows) >= periodMatchThreshold {
			return shift
		}
	}
	return 0
}

// detectVertical 检测垂直周期：平移 shift 行后逐列比较。
func detectVertical(g *Grid) int {
	for shift := minPeriod; shift <= g.Rows/2; shift++ {
		score := 0.0
		for j := 0; j < g.Cols; j++ {
			match := 0
			for i := shift; i < g.Rows; i++ {
				if g.Cells[i][j] == g.Cells[i-shift][j] {
					match++
				}
			}
			score += float64(match) / float64(g.Rows-shift)
		}
		if score/float64(g.Cols) >= periodMatchThreshold {
			return shift
		}
	}
	return 0
}

// shiftMatchRate 计算一行与自身平移 shift 位后的匹配率（循环移位）。
func shiftMatchRate(row []bool, shift int) float64 {
	n := len(row)
	if shift <= 0 || shift >= n {
		return 0
	}
	match := 0
	for i := 0; i < n; i++ {
		if row[i] == row[(i+shift)%n] {
			match++
		}
	}
	return float64(match) / float64(n)
}
