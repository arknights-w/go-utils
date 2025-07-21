package topology

import (
	"slices"
)

/**
 * TopologicalSort 拓扑排序
 *
 * @type_param T 节点类型
 * @param outDegEdge(outDegreeEdges) 出度边表，key为当前节点，value为当前节点的所有出度节点
 * @return sorted 拓扑排序后的节点序列，同一个层级的节点可并发执行
 * @return cycle 环的节点序列，如果没有环，则返回空切片
 */
func TopologicalSort[T comparable](outDegEdge map[T][]T) (sorted [][]T, cycle []T) {
	// 1. 环检测
	if len(sorted) != len(outDegEdge) {
		cycle = CheckCircular(outDegEdge)
	}
	if len(cycle) > 0 {
		return
	}
	// 2. 构建入度数量表
	inDegCont := buildInDegreeCount(outDegEdge)
	// 3. 构建拓扑排序
	sorted = buildSorted(outDegEdge, inDegCont)
	return
}

/**
 * buildInDegreeCount 构建入度数量表，key为节点，value为入度数量
 *
 * @type_param T 节点类型
 * @param outDegEdge(outDegreeEdges) 出度边表，key为当前节点，value为当前节点的所有出度节点
 * @return inDegCnt(inDegreeCountMap) 入度数量表
 */
func buildInDegreeCount[T comparable](outDegEdge map[T][]T) (inDegCnt map[T]int) {
	inDegCnt = make(map[T]int)
	for cur, edges := range outDegEdge {
		if _, ok := inDegCnt[cur]; !ok {
			inDegCnt[cur] = 0
		}
		for _, edge := range edges {
			inDegCnt[edge]++
		}
	}
	return
}

/**
 * buildSorted 构建拓扑排序
 *
 * @type_param T 节点类型
 * @param outDegEdge(outDegreeEdges) 出度边表，key为当前节点，value为当前节点的所有出度节点
 * @param inDegCnt(inDegreeCountMap) 入度数量表
 * @return sorted 拓扑排序后的节点序列
 */
func buildSorted[T comparable](outDegEdge map[T][]T, inDegCnt map[T]int) (sorted [][]T) {
	var (
		oInDegQue     []T = make([]T, 0, len(outDegEdge)) // (zeroInDegreeQueue)0入度队列
		old_oInDegQue []T = make([]T, 0, len(outDegEdge)) // (oldZeroInDegreeQueue)旧的0入度队列
	)
	for name, degree := range inDegCnt {
		if degree == 0 {
			oInDegQue = append(oInDegQue, name)
		}
	}
	// 3. 构建拓扑排序
	for len(oInDegQue) > 0 {
		// 如果当前层级的拓扑排序不为空，则添加到sorted中

		sorted = append(sorted, slices.Clone(oInDegQue))
		// 3.1 清空0入度队列
		old_oInDegQue, oInDegQue = oInDegQue, old_oInDegQue[:0]
		// 3.2 更新入度数量表和0入度队列
		for _, curr := range old_oInDegQue {
			for _, dep := range outDegEdge[curr] {
				inDegCnt[dep]--
				if inDegCnt[dep] == 0 {
					oInDegQue = append(oInDegQue, dep)
				}
			}
		}
	}
	return
}

/**
 * CheckCircular 环检测
 *
 * @type_param T 节点类型
 * @param outDegEdge(outDegreeEdges) 出度边表，key为当前节点，value为当前节点的所有出度节点
 * @return cycle 环的节点序列，如果没有环，则返回空切片
 */
func CheckCircular[T comparable](outDegEdge map[T][]T) (cycle []T) {
	visited := make(map[T]bool)
	recStack := make(map[T]bool)

	var dfs func(node T) bool
	var path []T

	dfs = func(node T) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range outDegEdge[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// 发现环，提取环的信息
				cycleStartIndex := slices.Index(path, neighbor)
				if cycleStartIndex != -1 {
					cycle = slices.Clone(path[cycleStartIndex:])
					cycle = append(cycle, neighbor)
				}
				return true
			}
		}

		recStack[node] = false
		path = path[:len(path)-1]
		return false
	}

	// 从所有未访问的节点开始DFS
	for node := range outDegEdge {
		if !visited[node] {
			path = []T{}
			if dfs(node) {
				break
			}
		}
	}

	return cycle
}
