package mtque

import (
	"fmt"
	"sync"
	"time"
)

var (
	queueList  map[string]*Queue
	queueMutex sync.RWMutex
)

func init() {
	queueList = make(map[string]*Queue)
	queueMutex = sync.RWMutex{}

	go func() {
		for {
			for _, queue := range queueList {
				go queue.PeriodicallyPersistent()
			}
		}
	}()
}

type Queue struct {
	Buffer
}

func newQueue() *Queue {
	buffer := NewBuffer()
	queue := Queue{Buffer: *buffer}

	return &queue
}

// SetQueueFile set a file for queue persistence
// This function should be called at constructing
// the queue if you want to persitent queue at backend
func SetQueueFile(file string) func(*Queue) {
	return func(queue *Queue) {
		queue.File = file
	}
}

// SetQueuePersistenceControl set if persistent the queue
// This function should be called at constructing
// the queue if you want to persitent the queue at backend
func SetQueuePersistenceControl(ctl bool) func(*Queue) {
	return func(queue *Queue) {
		queue.PersistenceControl = ctl
	}
}

// SetQueuePersistencePeriod set period to persistent queue
// This function should be called at constructing
// the queue if you want to persitent the queue at backend
func SetQueuePersistencePeriod(period time.Duration) func(*Queue) {
	return func(queue *Queue) {
		queue.PersistencePeriod = period
	}
}

// SetQueueRecoveryControl set if reocovery the queue from file
// This function should be called at constructing
// the queue if you want to persitent the queue at backend
// If set to recovery from file, the queue constructor will
// try to recovery queue from the file. if recovery failed,
// constructor will setup a new queue and clear the file.
func SetQueueRecoveryControl(ctl bool) func(*Queue) {
	return func(queue *Queue) {
		queue.RecoveryControl = ctl
	}
}

// SetPersistencePeriod set persistence period for queue.
func (q *Queue) SetPersistencePeriod(p time.Duration) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	q.PersistencePeriod = p
}

// SetFile will set the persistence file for queue.
func (q *Queue) SetFile(file string) error {
	if q.File != "" && file == q.File {
		return nil
	} else if q.File != "" && file != q.File {
		return fmt.Errorf("the new file:[%s] != the exist one[%s], you should use the ForceSetFile method to reset it. And should notice that if the recovery mode is enabled, the stack will be recoverd from the new file", file, q.File)
	}

	if _, ok := queueList[file]; ok {
		return fmt.Errorf("there is already a stack with file [%s] in stacklist, use the ForceSetFile method to reset current one", file)
	}

	q.File = file
	queueList[file] = q

	if q.RecoveryControl {
		q.Recovery()
	}

	q.PersistenceControl = true
	if q.PersistencePeriod == 0 {
		q.PersistencePeriod = DEFAULT_PERIOD_PERSISTENCE_TIME
	}

	return nil
}

func (q *Queue) ForceSetFile(file string) error {
	if queue, ok := queueList[file]; ok {
		q = queue
		return nil
	}

	if q.File != "" {
		delete(queueList, q.File)
	}

	q.File = file
	queueList[file] = q

	if q.RecoveryControl {
		q.Recovery()
	}

	q.PersistenceControl = true
	if q.PersistencePeriod == 0 {
		q.PersistencePeriod = DEFAULT_PERIOD_PERSISTENCE_TIME
	}

	return nil
}

// SetPersistenceControl enable/disable the persistence control for queue.
func (q *Queue) SetPersistenceControl(ctl bool) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	q.PersistenceControl = ctl
}

// SetRecoveryControl enable/disable the recovery control for Queue.
func (q *Queue) SetRecoveryControl(ctl bool) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	q.RecoveryControl = ctl
}

// GetPersistencePeriod returns the persistence period of the queue
func (q *Queue) GetPersistencePeriod() time.Duration {
	q.Mutex.RLock()
	defer q.Mutex.RUnlock()

	return q.PersistencePeriod
}

// GetPersistenceControl retruns the Persistence control setting of queue
func (q *Queue) GetPersistenceControl() bool {
	q.Mutex.RLock()
	defer q.Mutex.RUnlock()

	return q.PersistenceControl
}

// GetRecoveryControl returns the recovery control setting of queue
func (q *Queue) GetRecoveryControl() bool {
	q.Mutex.RLock()
	defer q.Mutex.RUnlock()

	return q.RecoveryControl
}

// GetFile returns the presistence file path of queue if setting
func (q *Queue) GetFile() string {
	q.Mutex.RLock()
	defer q.Mutex.RUnlock()

	return q.File
}

// NewQueue is the constructor of Queue.
// When use NewQueue to construct a queue, you can
// use option functions to set the options of queue.
func NewQueue(opts ...func(*Queue)) *Queue {
	queue := newQueue()
	for _, opt := range opts {
		opt(queue)
	}

	if queue.File != "" {
		queueMutex.Lock()
		defer queueMutex.Unlock()

		if q, ok := queueList[queue.File]; ok {
			queue = q
		} else {
			queueList[queue.File] = queue
		}
	}

	return queue
}

func GetQueue(file string) *Queue {
	queueMutex.RLock()
	if queue, ok := queueList[file]; ok {
		return queue
	}
	defer queueMutex.RUnlock()

	return NewQueue(
		SetQueueFile(file),
		SetQueuePersistenceControl(true),
		SetQueueRecoveryControl(true),
	)
}

func DestroyQueue(file string) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if _, ok := queueList[file]; ok {
		delete(queueList, file)
	}
}

func (q *Queue) Len() int64 {
	return q.Buffer.Len()
}

func (q *Queue) ID() string {
	return q.Buffer.ID()
}

func (q *Queue) Clear() {
	q.Buffer.Clear()
}

func (q *Queue) GetHead() (interface{}, error) {
	return q.Buffer.GetHeadValue()
}

func (q *Queue) EnQueue(value interface{}) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	node := NewDataNode(value)
	q.Datas.AddNodeAtTail(node)

	q.Length++

	if q.Length == 1 {
		q.SetRegister(value)
	}
}

func (q *Queue) DeQueue() (interface{}, error) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	if q.Length == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	value, err := q.Datas.GetHeadValue()
	if err == nil {
		q.Datas.DeleteNodeAtHead()
		q.Length--
	}

	return value, err
}

func (q *Queue) Persistent() error {
	return q.Buffer.Persistent()
}

func (q *Queue) PeriodicallyPersistent() {
	if q.PersistenceControl && !q.persRunning {
		q.persRunning = true
		select {
		case <-time.After(q.PersistencePeriod):
			q.Persistent()
		}
		q.persRunning = false
	}
}
