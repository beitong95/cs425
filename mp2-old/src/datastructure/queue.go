package datastructure
import "sync"


type CommandQueue struct {
	mux sync.Mutex
	queue [][]string
}

func (cq *CommandQueue) Enqueue(cmd []string) {
	cq.mux.Lock()
	cq.queue = append(cq.queue, cmd)
	cq.mux.Unlock()
}

func (cq *CommandQueue) Dequeue() []string {
	cq.mux.Lock()
	ret := cq.queue[0]
	cq.queue = cq.queue[1:]
	cq.mux.Unlock()
	return ret
}

func (cq *CommandQueue) IsEmpty() bool {
	cq.mux.Lock()
	ret := (len(cq.queue) == 0)
	cq.mux.Unlock()
	return ret
}