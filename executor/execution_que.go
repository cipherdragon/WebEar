package executor

import (
	"log"
	"sync"
	"time"
)

var scriptLock map[string]bool
var scriptLockMutex sync.Mutex
var headNode *ExecutionNode
var execMutex sync.Mutex

var executorRoutineRunner sync.Once

type ExecutionNode struct {
	next       *ExecutionNode
	payload     string
	name        string
	scriptPath  string
	username    string
}

func init() {
	scriptLock = make(map[string]bool)
	headNode = &ExecutionNode{}

	executorRoutineRunner.Do(func() {
		go callExecution()
	})
}

func ExecuteScript(payload string, name string, scriptPath string, username string) {
	execMutex.Lock()
	defer execMutex.Unlock()

	newNode := &ExecutionNode{
		next:       headNode.next,
		payload:    payload,
		name:       name,
		scriptPath: scriptPath,
		username:   username,
	}
	
	headNode.next = newNode
}

func callExecution() {
	for {
		time.Sleep(5 * time.Second)

		execMutex.Lock()

		prevNode := headNode
		for node := headNode.next; node != nil; node = node.next {
			scriptLockMutex.Lock()
			if scriptLock[node.scriptPath] {
				scriptLockMutex.Unlock()
				continue
			}

			scriptLock[node.scriptPath] = true
			scriptLockMutex.Unlock()

			err := executeScript(node.payload, node.name, node.scriptPath, node.username)
			if err != nil {
				scriptLockMutex.Lock()
				delete(scriptLock, node.scriptPath)
				scriptLockMutex.Unlock()

				log.Println(err.Error())
			}

			prevNode.next = node.next
			prevNode = node
		}

		execMutex.Unlock()
	}
}