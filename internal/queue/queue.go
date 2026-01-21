package queue

import (
	"sync"
)

// queue struct definition with front index for O(1) dequeue
type queue struct {
	mtx   *sync.Mutex
	items []interface{}
	front int // index of the front element
}

func New() *queue {
	return &queue{
		mtx:   &sync.Mutex{},
		items: make([]interface{}, 0),
		front: 0,
	}
}

// Enqueue adds the item into the queue
func (q *queue) Enqueue(item interface{}) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	q.items = append(q.items, item)
}

// Dequeue removes the item from the queue and RETURNS item - O(1) operation
func (q *queue) Dequeue() interface{} {
	q.mtx.Lock()
	defer q.mtx.Unlock()

	item := q.items[q.front]
	q.items[q.front] = nil // allow GC to reclaim the item
	q.front++

	// Compact the slice when front index is too large to avoid memory leak
	if q.front > len(q.items)/2 && q.front > 100 {
		q.items = q.items[q.front:]
		q.front = 0
	}

	return item
}

// Items returns the queue items (only the valid portion)
func (q *queue) Items() []interface{} {
	q.mtx.Lock()
	defer q.mtx.Unlock()

	return q.items[q.front:]
}

// Length returns the queue length
func (q *queue) Length() int {
	q.mtx.Lock()
	defer q.mtx.Unlock()

	return len(q.items) - q.front
}
