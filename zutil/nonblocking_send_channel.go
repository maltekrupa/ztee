package zutil

import (
	"sync"
)

type nonblockingSendChannel struct {
	in chan string
	out chan string
	queue UnboundedQueue
	cond *sync.Cond
	done bool
}

func addToQueue(nbsc *nonblockingSendChannel) {
	for elt := range nbsc.in {
		nbsc.cond.L.Lock()
		nbsc.queue.Push(elt)
		nbsc.cond.Signal()
		nbsc.cond.L.Unlock()
	}
	nbsc.cond.L.Lock()
	nbsc.done = true
	nbsc.cond.Signal()
	nbsc.cond.L.Unlock()
}

func sendToChannel(nbsc *nonblockingSendChannel) {
	for !nbsc.done {
		nbsc.cond.L.Lock()
		for !(nbsc.queue.Size() > 0 || nbsc.done) {
			nbsc.cond.Wait()
		}
		if (nbsc.done) {
			nbsc.cond.L.Unlock()
		} else {
			elt := nbsc.queue.Pop()
			nbsc.cond.L.Unlock()
			nbsc.out <- elt
		}
	}
	for nbsc.queue.Size() > 0 {
		nbsc.out <- nbsc.queue.Pop()
	}
	close(nbsc.out)
}

func NewNonblockingSendPair() (chan string, chan string) {
	nbsc := nonblockingSendChannel{in: make(chan string, 1000), 
		out: make(chan string, 1000), 
		queue: NewUnboundedRing(1000),
		cond: sync.NewCond(new(sync.Mutex))}
	go addToQueue(&nbsc)
	go sendToChannel(&nbsc)
	return nbsc.in, nbsc.out
}
