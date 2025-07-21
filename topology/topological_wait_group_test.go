package topology_test

import (
	"testing"
	"time"

	"github.com/arknights-w/go-utils/topology"
)

func TestTopologicalWaitGroup(t *testing.T) {
	inDegMap := map[string][]string{
		"1": nil,
		"2": {"1"},
		"3": {"1"},
		"4": {"2"},
		"5": {"3"},
		"6": {"4", "5"},
	}
	execMap := map[string]func(){
		"1": func() { time.Sleep(1 * time.Second); println("1") },
		"2": func() { time.Sleep(2 * time.Second); println("2") },
		"3": func() { time.Sleep(3 * time.Second); println("3") },
		"4": func() { time.Sleep(2 * time.Second); println("4") },
		"5": func() { time.Sleep(1 * time.Second); println("5") },
		"6": func() { time.Sleep(1 * time.Second); println("6") },
	}
	wg := topology.NewWaitGroup(inDegMap)
	for name, exec := range execMap {
		node, err := wg.GetNode(name)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			defer node.Done()
			node.Wait()
			exec()
		}()
	}
	wg.Wait()
}
