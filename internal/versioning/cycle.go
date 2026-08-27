package versioning

import (
	"fmt"

	"task266-textilerelation/internal/model"
)

// cycleDetector 用 DFS 检测关系图是否有向环。
type cycleDetector struct {
	adj    map[int64][]int64 // 单元 -> 可达单元
	state  map[int64]int     // 0=未访问 1=访问中 2=已完成
	cycle  []int64
	has    bool
}

// DetectCycle 检测单元关系图是否存在出处循环。
//
// 入参为有向边 (fromUnitID -> toUnitID)。成环表示"单元 A 传承自 B，
// 且 B 又传承自 A"，自指矛盾，发布前必须拒绝。
func DetectCycle(edges [][2]int64) error {
	d := &cycleDetector{
		adj:   make(map[int64][]int64),
		state: make(map[int64]int),
	}
	for _, e := range edges {
		d.adj[e[0]] = append(d.adj[e[0]], e[1])
		d.state[e[0]] = 0
		d.state[e[1]] = 0
	}
	nodes := make([]int64, 0, len(d.state))
	for n := range d.state {
		nodes = append(nodes, n)
	}
	for _, n := range nodes {
		if d.state[n] == 0 {
			d.visit(n)
			if d.has {
				break
			}
		}
	}
	if !d.has {
		return nil
	}
	path := ""
	for i, n := range d.cycle {
		if i > 0 {
			path += " -> "
		}
		path += fmt.Sprintf("%d", n)
	}
	return fmt.Errorf("%w: %s", model.ErrCycle, path)
}

// visit DFS 三色标记。
func (d *cycleDetector) visit(n int64) {
	d.state[n] = 1
	d.cycle = append(d.cycle, n)
	for _, next := range d.adj[n] {
		switch d.state[next] {
		case 1:
			// 命中当前递归栈中的节点即成环：自环(1)、双向边(2)、长链(>=3)
			// 一律拒绝。截取从该节点起的路径作为成环证据。
			d.has = true
			i := 0
			for i < len(d.cycle) && d.cycle[i] != next {
				i++
			}
			d.cycle = d.cycle[i:]
			d.cycle = append(d.cycle, next)
			return
		case 0:
			d.visit(next)
			if d.has {
				return
			}
		}
	}
	d.state[n] = 2
	d.cycle = d.cycle[:len(d.cycle)-1]
}
