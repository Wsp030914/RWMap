package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// // 基于 sync.RWMutex 和内置 map 实现的并发安全泛型 RWMap。
type RWMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewRWMap[K comparable, V any](n int) *RWMap[K, V] {
	return &RWMap[K, V]{
		m: make(map[K]V, n),
	}
}

func (m *RWMap[K, V]) Load(key K) (V, bool) {
	m.mu.RLock()
	v, ok := m.m[key]
	m.mu.RUnlock()
	return v, ok
}

func (m *RWMap[K, V]) Store(key K, val V) {
	m.mu.Lock()
	m.m[key] = val
	m.mu.Unlock()
}

func (m *RWMap[K, V]) Len() int {
	m.mu.RLock()
	l := len(m.m)
	m.mu.RUnlock()
	return l
}

func (m *RWMap[K, V]) Delete(k K) {
	m.mu.Lock()
	delete(m.m, k)
	m.mu.Unlock()
}

func (m *RWMap[K, V]) Each(fn func(K, V) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, v := range m.m {
		if !fn(k, v) {
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	m := NewRWMap[int, string](1)

	var wg sync.WaitGroup

	writerNum := 5
	readerNum := 5
	keyRange := 100

	wg.Add(writerNum)
	for wid := 0; wid < writerNum; wid++ {
		go func(id int) {
			defer wg.Done()
			for k := 0; k < keyRange; k++ {
				val := fmt.Sprintf("writer-%d-val-%d", id, k)
				m.Store(k, val)
			}
			fmt.Printf("[writer %d] done\n", id)
		}(wid)
	}

	wg.Add(readerNum)
	for rid := 0; rid < readerNum; rid++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < keyRange; i++ {
				k := rand.Intn(keyRange)
				if v, ok := m.Load(k); ok {
					fmt.Printf("[reader %d] load k=%d v=%s\n", id, k, v)
				}
			}
			fmt.Printf("[reader %d] done\n", id)
		}(rid)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < keyRange/2; i++ {
			k := rand.Intn(keyRange)
			m.Delete(k)
		}
		fmt.Println("[deleter] done")
	}()

	wg.Wait()

	fmt.Printf("final len = %d\n", m.Len())

	count := 0
	m.Each(func(k int, v string) bool {
		fmt.Printf("  k=%d v=%s\n", k, v)
		count++
		if count >= 10 {
			return false
		}
		return true
	})
}
