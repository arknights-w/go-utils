package topology

import (
	"fmt"
	"sync"
)

type WaitGroup[T comparable] struct {
	inDegWaitMap  map[T]*sync.WaitGroup
	outDegDoneMap map[T][]*sync.WaitGroup
	globalWg      *sync.WaitGroup
}

func (wg *WaitGroup[T]) Wait() {
	wg.globalWg.Wait()
}

func (wg *WaitGroup[T]) GetNode(name T) (*node[T], error) {
	inDegWaitGroup, ok := wg.inDegWaitMap[name]
	if !ok {
		return nil, fmt.Errorf("inDegWaitMap cannot find node: %v", name)
	}
	outDegDoneGroup, ok := wg.outDegDoneMap[name]
	if !ok {
		return nil, fmt.Errorf("outDegDoneMap cannot find node: %v", name)
	}
	globalWg := wg.globalWg
	return &node[T]{
		inDegWaitGroup:  inDegWaitGroup,
		outDegDoneGroup: outDegDoneGroup,
		globalWg:        globalWg,
	}, nil
}

func NewWaitGroup[T comparable](inDegMap map[T][]T) *WaitGroup[T] {
	inDegWaitMap, outDegDoneMap := buildDegMap(inDegMap)
	globalWg := &sync.WaitGroup{}
	globalWg.Add(len(inDegWaitMap))
	return &WaitGroup[T]{
		inDegWaitMap:  inDegWaitMap,
		outDegDoneMap: outDegDoneMap,
		globalWg:      globalWg,
	}
}

type node[T comparable] struct {
	inDegWaitGroup  *sync.WaitGroup
	outDegDoneGroup []*sync.WaitGroup
	globalWg        *sync.WaitGroup
}

func (nw *node[T]) Wait() {
	nw.inDegWaitGroup.Wait()
}

func (nw *node[T]) Done() {
	for _, wg := range nw.outDegDoneGroup {
		wg.Done()
	}
	nw.globalWg.Done()
}

func buildDegMap[T comparable](inDegMap map[T][]T) (
	inDegWaitMap map[T]*sync.WaitGroup,
	outDegDoneMap map[T][]*sync.WaitGroup,
) {
	inDegWaitMap = make(map[T]*sync.WaitGroup)
	outDegDoneMap = make(map[T][]*sync.WaitGroup)

	for key, values := range inDegMap {
		if _, ok := outDegDoneMap[key]; !ok {
			outDegDoneMap[key] = make([]*sync.WaitGroup, 0)
		}
		inDegWaitMap[key] = &sync.WaitGroup{}
		inDegWaitMap[key].Add(len(values))
		for _, value := range values {
			if _, ok := outDegDoneMap[value]; !ok {
				outDegDoneMap[value] = make([]*sync.WaitGroup, 0)
			}
			outDegDoneMap[value] = append(outDegDoneMap[value], inDegWaitMap[key])
		}
	}

	return inDegWaitMap, outDegDoneMap
}
