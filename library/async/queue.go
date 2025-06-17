package async

import (
	"sync/atomic"
	"unsafe"
)

type Node[T any] struct {
	next  *Node[T]
	value T
}

type Queue[T any] struct {
	head  *Node[T]
	tail  *Node[T]
	count int64
}

func NewQueue[T any]() *Queue[T] {
	node := new(Node[T])
	return &Queue[T]{head: node, tail: node}
}

func (d *Queue[T]) GetCount() int64 {
	return atomic.LoadInt64(&d.count)
}

func (d *Queue[T]) Push(val T) {
	addNode := new(Node[T])
	addNode.value = val
	// 将新增节点插入链表
	prevNode := (*Node[T])(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&d.tail)), unsafe.Pointer(addNode)))
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&prevNode.next)), unsafe.Pointer(addNode))
	atomic.AddInt64(&d.count, 1)
}

func (d *Queue[T]) Pop() (ret T) {
	if node := (*Node[T])(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&d.head.next)))); node != nil {
		atomic.AddInt64(&d.count, -1)
		ret = node.value
		d.head.next = nil
		d.head = node
	}
	return
}
