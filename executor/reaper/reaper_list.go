package reaper

import "sync"

type PIDNode struct {
	PID  int
	Next *PIDNode
	Previous *PIDNode
}

var headNode, tailNode *PIDNode
var mutex sync.Mutex

func initialize() {
	mutex.Lock()
	defer mutex.Unlock()

	headNode = &PIDNode{
		PID: 0,
		Next: nil,
		Previous: nil,
	}

	tailNode = &PIDNode{
		PID: 0,
		Next: nil,
		Previous: headNode,
	}

	headNode.Next = tailNode
}

func RecordToReap(pid int) {
	if headNode == nil {
		initialize()
	}

	newNode := &PIDNode{
		PID: pid,
	}

	mutex.Lock()
	defer mutex.Unlock()

	if headNode.Next == tailNode {
		newNode.Next = tailNode
		newNode.Previous = headNode
		headNode.Next = newNode
		tailNode.Previous = newNode
	} else {
		lastNode := tailNode.Previous
		newNode.Next = tailNode
		newNode.Previous = lastNode
		lastNode.Next = newNode
	}
}

func MarkAsReaped(node *PIDNode) {
	node.Previous.Next = node.Next
	node.Next.Previous = node.Previous
}