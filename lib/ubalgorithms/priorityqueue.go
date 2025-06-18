package ubalgorithms

import (
	"container/heap"
	"errors"
	"fmt"
	"strings"
)

type priorityQueueNode[T comparable] struct {
	Priority int
	Data     T
	Index    int
}

type PriorityQueue[T comparable] struct {
	elements []*priorityQueueNode[T]
	compFunc func(a, b int) bool
}

func NewMinCostPriorityQueue[T comparable](capacity int) *PriorityQueue[T] {
	return newPriorityQueue[T](capacity, false)
}

func NewMaxCostPriorityQueue[T comparable](capacity int) *PriorityQueue[T] {
	return newPriorityQueue[T](capacity, true)
}

// Create a new priority queue.
func newPriorityQueue[T comparable](capacity int, maximize bool) *PriorityQueue[T] {
	q := PriorityQueue[T]{
		elements: make([]*priorityQueueNode[T], 0, capacity),
	}

	if !maximize {
		q.compFunc = func(i, j int) bool {
			return q.elements[i].Priority < q.elements[j].Priority
		}
	} else {
		q.compFunc = func(i, j int) bool {
			return q.elements[i].Priority > q.elements[j].Priority
		}
	}

	return &q
}

func (q *PriorityQueue[T]) Less(i, j int) bool {
	return q.compFunc(i, j)
}

func (q *PriorityQueue[T]) Push(x any) {
	item := x.(*priorityQueueNode[T])
	item.Index = len(q.elements)
	q.elements = append(q.elements, item)
}

// PushItem a new item onto the priority queue
func (q *PriorityQueue[T]) PushItem(value T, priority int) {
	node := priorityQueueNode[T]{
		Priority: priority,
		Data:     value,
	}
	heap.Push(q, &node)
}

// Get the current size of the priority queue
func (q *PriorityQueue[T]) Len() int {
	return len(q.elements)
}

// Retrieve the next item without popping it
func (q *PriorityQueue[T]) Peek() (*T, error) {
	if len(q.elements) == 0 {
		return nil, errors.New("No items are in the queue.")
	}
	return &q.elements[0].Data, nil
}

func (q *PriorityQueue[T]) String() string {
	parts := make([]string, len(q.elements))
	for _, node := range q.elements {
		parts = append(parts, fmt.Sprintf("%v:%v", node.Priority, node.Data))
	}
	str := strings.Join(parts, ",")
	return fmt.Sprintf("PriorityQueue{%v}", str)
}

func (q *PriorityQueue[T]) Pop() any {
	old := q.elements
	n := len(old)
	item := old[n-1]
	item.Index = -1 // Mark as removed.
	q.elements = old[0 : n-1]
	return item
}

func (q PriorityQueue[T]) Swap(i, j int) {
	q.elements[i], q.elements[j] = q.elements[j], q.elements[i]
	q.elements[i].Index = i
	q.elements[j].Index = j
}

// Pop the next item
func (pq *PriorityQueue[T]) PopItem() T {
	item := heap.Pop(pq).(*priorityQueueNode[T])
	return item.Data
}

// Heapify the current index up
func (q *PriorityQueue[T]) heapifyUp(index int) {
	if index == 0 {
		return
	}
	currentNode := q.elements[index]
	parentIndex := q.parent(index)
	parentNode := q.elements[parentIndex]

	if parentNode.Priority <= currentNode.Priority {
		return
	}
	q.elements[index] = parentNode
	q.elements[parentIndex] = currentNode
	q.heapifyUp(parentIndex)
}

// Heapify the current index down
func (q *PriorityQueue[T]) heapifyDown(index int) {

	if index >= len(q.elements) {
		return
	}

	leftChildIndex := q.leftChild(index)
	if leftChildIndex > len(q.elements)-1 {
		return
	}

	rightChildIndex := q.rightChild(index)
	currentNode := q.elements[index]
	leftNode := q.elements[leftChildIndex]

	if rightChildIndex < len(q.elements) {
		if leftNode.Priority < currentNode.Priority {
			q.elements[index] = leftNode
			q.elements[leftChildIndex] = currentNode
			q.heapifyDown(leftChildIndex)
		}
	}

	var rightNode *priorityQueueNode[T] = nil
	if rightChildIndex < len(q.elements) {
		rightNode = q.elements[rightChildIndex]
	}

	if (rightNode == nil || leftNode.Priority < rightNode.Priority) &&
		leftNode.Priority < currentNode.Priority {
		q.elements[index] = leftNode
		q.elements[leftChildIndex] = currentNode
		q.heapifyDown(leftChildIndex)
	} else if rightNode != nil && rightNode.Priority < currentNode.Priority {
		q.elements[index] = rightNode
		q.elements[rightChildIndex] = currentNode
		q.heapifyDown(rightChildIndex)
	}
}

func (q *PriorityQueue[T]) parent(index int) int {
	return (index - 1) / 2
}

func (q *PriorityQueue[T]) leftChild(index int) int {
	return 2*index + 1
}

func (q *PriorityQueue[T]) rightChild(index int) int {
	return 2*index + 2
}
