package topology

// 对称检测：判断纹样单元在水平翻转（镜像）、垂直翻转、180° 旋转下的自一致性。
//
// 返回值取值：none / h / v / hv / rot180。
// 对称性是织物纹样单元的重要结构指纹——同一织造传统的单元常共享对称类型。

// DetectSymmetry 检测单元对称类型。
func DetectSymmetry(g *Grid) string {
	h := horizontalSymmetric(g)
	v := verticalSymmetric(g)
	r := rot180Symmetric(g)

	switch {
	case h && v && r:
		return "hv"
	case h && r:
		return "hv"
	case h:
		return "h"
	case v:
		return "v"
	case r:
		return "rot180"
	default:
		return "none"
	}
}

// horizontalSymmetric 水平镜像对称：grid[i][j] == grid[i][cols-1-j]。
func horizontalSymmetric(g *Grid) bool {
	rows, cols := g.Rows, g.Cols
	if rows == 0 || cols == 0 {
		return false
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols/2; j++ {
			if g.Cells[i][j] != g.Cells[i][cols-1-j] {
				return false
			}
		}
	}
	return true
}

// verticalSymmetric 垂直镜像对称：grid[i][j] == grid[rows-1-i][j]。
func verticalSymmetric(g *Grid) bool {
	rows, cols := g.Rows, g.Cols
	if rows == 0 || cols == 0 {
		return false
	}
	for i := 0; i < rows/2; i++ {
		for j := 0; j < cols; j++ {
			if g.Cells[i][j] != g.Cells[rows-1-i][j] {
				return false
			}
		}
	}
	return true
}

// rot180Symmetric 180° 旋转对称：grid[i][j] == grid[rows-1-i][cols-1-j]。
func rot180Symmetric(g *Grid) bool {
	rows, cols := g.Rows, g.Cols
	if rows == 0 || cols == 0 {
		return false
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if g.Cells[i][j] != g.Cells[rows-1-i][cols-1-j] {
				return false
			}
		}
	}
	return true
}

// SymmetryCompatible 判断两个单元的对称类型是否属于同一结构族：
// 有对称（h/v/hv/rot180 任意）与无对称（none）分属两族；
// 对称族内部任意子类型视为可兼容（单元可以朝向不同）。
func SymmetryCompatible(a, b string) bool {
	if a == "none" || b == "none" {
		return a == b
	}
	return true
}
