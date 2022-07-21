package queue

import (
	"sync"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

var (
	_ SchedulingQueue = (*PrioritySchedulingQueue)(nil)
)

// Implements a priority queue version of the SchedulingQueue.
// All methods are thread-safe.
type PrioritySchedulingQueue struct {
	heapMap  util.HeapMap[string, *pipeline.QueuedPodInfo]
	isClosed bool
	mutex    sync.RWMutex
	cond     *sync.Cond
}

func NewPrioritySchedulingQueue(lessFn util.LessFunc[*pipeline.QueuedPodInfo]) *PrioritySchedulingQueue {
	priorityQueue := PrioritySchedulingQueue{
		heapMap:  util.NewMinHeapMap[string](lessFn, nil),
		isClosed: false,
		mutex:    sync.RWMutex{},
	}
	priorityQueue.cond = sync.NewCond(&priorityQueue.mutex)
	return &priorityQueue
}

func (pq *PrioritySchedulingQueue) Enqueue(podInfo *pipeline.QueuedPodInfo) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if pq.isClosed {
		return
	}

	pq.heapMap.AddOrReplace(podInfo.GetKey(), podInfo)
	pq.cond.Broadcast()
}

func (pq *PrioritySchedulingQueue) Dequeue() *pipeline.QueuedPodInfo {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if pq.isClosed {
		return nil
	}

	// If the queue is empty, we wait on the condition variable until
	// there is either something in the queue or the queue is closed.
	for pq.heapMap.Len() == 0 {
		pq.cond.Wait()

		if pq.isClosed {
			return nil
		}
	}

	if _, podInfo, ok := pq.heapMap.Pop(); ok {
		return podInfo
	}
	return nil
}

func (pq *PrioritySchedulingQueue) Close() {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()
	pq.isClosed = true
	pq.cond.Broadcast()
}

func (pq *PrioritySchedulingQueue) IsClosed() bool {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()
	return pq.isClosed
}
