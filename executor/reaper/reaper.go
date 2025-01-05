package reaper

import (
	"sync"
	"syscall"
	"time"
)

var wakerReaperOnce sync.Once

func reap(pid_node *PIDNode) {
	pid := pid_node.PID
	var status syscall.WaitStatus

	wpid, err := syscall.Wait4(pid, &status, syscall.WNOHANG, nil)
	if err != nil {
		return
	}

	if wpid == 0 {
		return
	}

	MarkAsReaped(pid_node)
}

func goAndReap() {
	for {
		time.Sleep(5 * time.Second)

		if headNode == nil {
			continue
		}

		// mutex declared in reaper_list.go
		// Preavent new PIDs being added to reap while reaping
		mutex.Lock()

		for node := headNode.Next; node != tailNode; node = node.Next {
			reap(node)
		}

		mutex.Unlock()
	}
}

func WakeUpReaper() {
	wakerReaperOnce.Do(func() {
		go goAndReap()
		// go func ()  {
		// 	for {
		// 		fmt.Println("Waking up reaper")
		// 		time.Sleep(5 * time.Second)
		// 	}
		// }()
	})

}