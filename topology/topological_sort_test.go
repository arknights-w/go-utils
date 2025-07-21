package topology_test

import (
	"testing"

	topo "github.com/arknights-w/go-utils/topology"
)

func TestSuccess(t *testing.T) {
	edges := map[int][]int{
		1: {2, 3},
		2: {4, 5},
		3: {5},
		4: {6},
		5: {6},
		6: {},
	}
	sorted, cycle := topo.TopologicalSort(edges)
	t.Log(sorted)
	t.Log(cycle)
}

func TestCircular(t *testing.T) {
	edges := map[int][]int{
		0: {1},
		1: {2},
		2: {3},
		3: {1},
	}
	sorted, cycle := topo.TopologicalSort(edges)
	t.Log(sorted)
	t.Log(cycle)
}
