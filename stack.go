package mtque

import (
	"fmt"
	"sync"
	"time"
)

type Stack struct {
	Buffer
}

var (
	stackList  map[string]*Stack
	stackMutex sync.RWMutex
)

func init() {
	stackList = make(map[string]*Stack)
	stackMutex = sync.RWMutex{}
}

func newStack() *Stack {
	buffer := NewBuffer()
	stack := Stack{Buffer: *buffer}

	return &stack
}

// SetStackFile set a file for stack persistence
// This function should be called at constructing
// the stack if you want to persitent the stack at backend
func SetStackFile(file string) func(*Stack) {
	return func(stack *Stack) {
		stack.File = file
	}
}

// SetStackPersistenceControl set if persistent the stack
// This function should be called at constructing
// the stack if you want to persitent the stack at backend
func SetStackPersistenceControl(ctl bool) func(*Stack) {
	return func(stack *Stack) {
		stack.PersistenceControl = ctl
	}
}

// SetStackPersistencePeriod set period to persistent stack
// This function should be called at constructing
// the stack if you want to persitent the stack at backend
func SetStackPersistencePeriod(period time.Duration) func(*Stack) {
	return func(stack *Stack) {
		stack.PersistencePeriod = period
	}
}

// SetStackRecoveryControl set if reocovery the stack from file
// This function should be called at constructing
// the stack if you want to persitent the stack at backend
// If set to recovery from file, the stack constructor will
// try to recovery stack from the file. if recovery failed,
// constructor will setup a new stack and clear the file.
func SetStackRecoveryControl(ctl bool) func(*Stack) {
	return func(stack *Stack) {
		stack.RecoveryControl = ctl
	}
}

// NewStack is the constructor of Stack.
// When use NewStack to construct a stack, you can
// use option functions to set the options of stack.
func NewStack(opts ...func(*Stack)) *Stack {
	stack := newStack()
	for _, opt := range opts {
		opt(stack)
	}

	if stack.File != "" {
		stackMutex.Lock()
		defer stackMutex.Unlock()

		if stk, ok := stackList[stack.File]; ok {
			stack = stk
		} else {
			stackList[stack.File] = stack
		}
	}

	return stack
}

// GetStack will try to findout an already existed stack through the file.
// If there is not a stack in the memroy, it will first try to recovery
// one from the file. If recovery failed, it will construct a new one,
// and set persistence and recovery control as true.
func GetStack(file string) *Stack {
	listMutex.RLock()
	if stack, ok := stackList[file]; ok {
		return stack
	}
	defer listMutex.RUnlock()

	return NewStack(
		SetStackFile(file),
		SetStackPersistenceControl(true),
		SetStackRecoveryControl(true),
	)
}

// DestroyStack will destroy the stack in the stack list
// And then delete the persistence file for the stack.
func DestroyStack(file string) {
	listMutex.Lock()
	defer listMutex.Unlock()

	delete(stackList, file)
	//TODO：delete file
}

// SetPersistencePeriod set persistence period for stack.
func (s *Stack) SetPersistencePeriod(p time.Duration) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.PersistencePeriod = p
}

// SetFile will set the persistence file for stack.
func (s *Stack) SetFile(file string) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	return fmt.Errorf("Not implemented")
}

// SetPersistenceControl enable/disable the persistence control for stack.
func (s *Stack) SetPersistenceControl(ctl bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.PersistenceControl = ctl
}

// SetRecoveryControl enable/disable the recovery control for stack.
func (s *Stack) SetRecoveryControl(ctl bool) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.RecoveryControl = ctl
}

// GetPersistencePeriod returns the persistence period of the stack
func (s *Stack) GetPersistencePeriod() time.Duration {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.PersistencePeriod
}

// GetPersistenceControl retruns the Persistence control setting of stack
func (s *Stack) GetPersistenceControl() bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.PersistenceControl
}

// GetRecoveryControl returns the recovery control setting of stack
func (s *Stack) GetRecoveryControl() bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.RecoveryControl
}

// GetFile returns the presistence file path of stack if setting
func (s *Stack) GetFile() string {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	return s.File
}

// GetTail will return the value at stack tail without
// delete the value from stack.
// If the stack is empty, it will return an error.
func (s *Stack) GetTail() (interface{}, error) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	if s.Length == 0 {
		return nil, fmt.Errorf("stack is empty")
	}

	return s.Datas.GetTailValue()
}

// Push will push a value at tail of stack
func (s *Stack) Push(value interface{}) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	node := NewDataNode(value)
	s.Datas.AddNodeAtTail(node)
	s.Length++
}

// Pop will pop the value at the tail of stack out.
// It will also delete it from the stack.
func (s *Stack) Pop() (interface{}, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Length == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	value, err := s.Datas.GetTailValue()
	if err == nil {
		s.Datas.DeleteNodeAtTail()
		s.Length--
	}

	return value, err
}
