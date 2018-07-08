package mtque

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

const BUFFER_INFO_SIZE = 202
const DEFAULT_PERIOD_PERSISTENCE_TIME = time.Minute * 5

// DataNode should be encoded as |len|value....|
type DataNode struct {
	Value    interface{}
	ValueLen int64
	Next     *DataNode
	Previous *DataNode
}

func NewDataNode(value interface{}) *DataNode {
	return &DataNode{Value: value}
}

func (d *DataNode) Bytes() ([]byte, error) {
	binBuf := new(bytes.Buffer)
	err := gob.NewEncoder(binBuf).Encode(d.Value)
	if err != nil {
		return []byte{}, err
	}

	d.ValueLen = int64(binBuf.Len())

	headgob := new(bytes.Buffer)
	err = gob.NewEncoder(headgob).Encode(d.ValueLen)
	if err != nil {
		return []byte{}, err
	}

	// size after gob is not fixed 8, it depends on the value.
	head := make([]byte, 8)
	if headgob.Len() > 8 {
		return []byte{}, fmt.Errorf("data size is exceed the range of int64")
	}
	copy(head, headgob.Bytes())

	return append(head, binBuf.Bytes()...), nil
}

type DataLink struct {
	Head            *DataNode
	Tail            *DataNode
	LastPersistence *DataNode
}

func NewDataLink() *DataLink {
	return new(DataLink)
}

func (dl *DataLink) AddNodeAtHead(data *DataNode) {
	if dl.Head == nil {
		dl.Head = data
		dl.Tail = data
	} else {
		data.Next = dl.Head
		dl.Head.Previous = data
		dl.Head = data
	}
}

func (dl *DataLink) AddNodeAtTail(data *DataNode) {
	if dl.Tail == nil {
		dl.Head = data
		dl.Tail = data
	} else {
		data.Previous = dl.Tail
		dl.Tail.Next = data
		dl.Tail = data
	}
}

func (dl *DataLink) GetHeadValue() (interface{}, error) {
	if dl.Head == nil {
		return nil, fmt.Errorf("link is empty")
	}

	return dl.Head.Value, nil
}

func (dl *DataLink) DeleteNodeAtHead() *DataNode {
	if dl.Head == nil {
		return nil
	}

	node := dl.Head
	dl.Head.Next.Previous = nil
	dl.Head = dl.Head.Next
	node.Next = nil

	if dl.LastPersistence == node {
		dl.LastPersistence = dl.Head
	}

	return node
}

func (dl *DataLink) GetTailValue() (interface{}, error) {
	if dl.Tail == nil {
		return nil, fmt.Errorf("link is empty")
	}

	return dl.Tail.Value, nil
}

func (dl *DataLink) DeleteNodeAtTail() *DataNode {
	if dl.Tail == nil {
		return nil
	}

	node := dl.Tail
	dl.Tail.Previous.Next = nil
	dl.Tail = dl.Tail.Previous
	node.Previous = nil

	if dl.LastPersistence == node {
		dl.LastPersistence = dl.Tail
	}

	return node
}

type BufferInfo struct {
	Id     string
	Length int64

	//Control weather set up the buffer through recoverying from file
	RecoveryControl bool

	PersistenceControl bool
	PersistencePeriod  time.Duration

	FileStartSeek int64 //start position in file of the persistence
	FileEndSeek   int64 //last position in file for the persistence
}

func SetBufferInfoPersistenceControl(ctl bool) func(*BufferInfo) {
	return func(info *BufferInfo) {
		info.PersistenceControl = ctl
	}
}

func SetBufferInfoPersistencePeriod(period time.Duration) func(*BufferInfo) {
	return func(info *BufferInfo) {
		info.PersistencePeriod = period
	}
}

func SetBufferInfoRecoveryControl(ctl bool) func(*BufferInfo) {
	return func(info *BufferInfo) {
		info.RecoveryControl = ctl
	}
}

func NewBufferInfo(opts ...func(*BufferInfo)) *BufferInfo {
	info := new(BufferInfo)

	info.Id = uuid.Must(uuid.NewV4()).String()
	info.PersistencePeriod = 5 * time.Minute

	for _, opt := range opts {
		opt(info)
	}

	return info
}

func (b *BufferInfo) Bytes() ([]byte, error) {
	binBuf := new(bytes.Buffer)
	err := gob.NewEncoder(binBuf).Encode(*b)
	if err != nil {
		return []byte{}, err
	}

	return binBuf.Bytes(), nil
}

type Buffer struct {
	BufferInfo
	File string

	Mutex sync.RWMutex

	Datas *DataLink

	//User should register the data origin type to recovery data
	Register interface{}
}

func SetBufferFile(file string) func(*Buffer) {
	return func(buf *Buffer) {
		buf.File = file
	}
}

func SetBufferPersistenceControl(ctl bool) func(*Buffer) {
	return func(buf *Buffer) {
		buf.PersistenceControl = ctl
	}
}

func SetBufferPersistencePeriod(period time.Duration) func(*Buffer) {
	return func(buf *Buffer) {
		buf.PersistencePeriod = period
	}
}

func SetBufferRecoveryControl(ctl bool) func(*Buffer) {
	return func(buf *Buffer) {
		buf.RecoveryControl = ctl
	}
}

func SetBufferRegister(datatype interface{}) func(*Buffer) {
	return func(buf *Buffer) {
		buf.Register = datatype
	}
}

func NewBuffer(opts ...func(*Buffer)) *Buffer {
	buffer := new(Buffer)
	buffer.BufferInfo = *NewBufferInfo()
	buffer.Mutex = sync.RWMutex{}
	buffer.Datas = NewDataLink()

	for _, opt := range opts {
		opt(buffer)
	}

	return buffer
}

func (b *Buffer) Clear() {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	b.Datas = NewDataLink()
	b.Length = 0
}

func (b *Buffer) Len() int64 {
	b.Mutex.RLock()
	defer b.Mutex.RUnlock()

	return b.Length
}

func (b *Buffer) ID() string {
	b.Mutex.RLock()
	defer b.Mutex.RUnlock()

	return b.Id
}

func (b *Buffer) AddDataAtHead(value interface{}) {
	node := NewDataNode(value)
	b.Datas.AddNodeAtHead(node)
}

func (b *Buffer) AddDataAtTail(value interface{}) {
	node := NewDataNode(value)
	b.Datas.AddNodeAtTail(node)
}

func (b *Buffer) GetHeadValue() (interface{}, error) {
	return b.Datas.GetHeadValue()
}

func (b *Buffer) DeleteNodeAtHead() {
	node := b.Datas.DeleteNodeAtHead()
	if node != nil && node.ValueLen > 0 {
		b.decrementPersistentAtHead(node)
	}
}

func (b *Buffer) GetTailValue() (interface{}, error) {
	return b.Datas.GetTailValue()
}

func (b *Buffer) DeleteNodeAtTail() {
	node := b.Datas.DeleteNodeAtTail()
	if node != nil && node.ValueLen > 0 {
		b.decrementPersistentAtTail(node)
	}
}

func (b *Buffer) SetRegister(datatype interface{}) {
	b.Register = datatype
}

//IncrementPersistent will Persistent datas from last persistence at
//the end of the file.
func (b *Buffer) incrementPersistent() error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	if !b.PersistenceControl {
		return fmt.Errorf("presistence is not enabled")
	}
	if b.File == "" {
		return fmt.Errorf("the file to persistent datas is not specified")
	}

	file, err := os.OpenFile(b.File, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if b.FileStartSeek == 0 {
		b.FileStartSeek = BUFFER_INFO_SIZE
		b.FileEndSeek = BUFFER_INFO_SIZE
	}

	_, err = file.Seek(b.FileEndSeek, 0)
	if err != nil {
		return err
	}

	if b.Datas.LastPersistence == nil {
		b.Datas.LastPersistence = b.Datas.Head
	} else {
		b.Datas.LastPersistence = b.Datas.LastPersistence.Next
	}

	for node := b.Datas.LastPersistence; node != nil; node = node.Next {
		content, err := node.Bytes()
		if err != nil {
			return err
		}

		_, err = file.Write(content)
		if err != nil {
			return err
		}

		b.FileEndSeek += int64(len(content))
	}

	info, err := b.Bytes()
	if err != nil {
		return err
	}

	_, err = file.WriteAt(info, 0)
	if err != nil {
		return err
	}

	file.Seek(b.FileEndSeek, 0)

	return nil
}

//DecrementPersistentAtHead will delete datas which already be persistented
//in the file when deleting data from the header of buffer.
//It is trigged by deleting data from the header of buffer, and should satisfied
//the condition that buffer.head is before buffer.LastPersistence
func (b *Buffer) decrementPersistentAtHead(node *DataNode) error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	if node == nil {
		return fmt.Errorf("there is no node to remove from persistence")
	}
	if node.ValueLen <= 0 {
		return fmt.Errorf("the node is not persistent before")
	}

	if b.FileStartSeek < BUFFER_INFO_SIZE {
		return fmt.Errorf("the node was persistented by covering the buffer info")
	}

	if b.FileStartSeek+int64(node.ValueLen) >= b.FileEndSeek {
		b.FileStartSeek = BUFFER_INFO_SIZE
		b.FileEndSeek = BUFFER_INFO_SIZE
	} else {
		b.FileStartSeek += int64(node.ValueLen)
	}

	return nil
}

//DecrementPersistentAtTail will delete datas which already be persistented
//in the file when deleting data from the tail of buffer.
//It is trigged by deleting data from the tail of buffer, and should satisfied
//the condition that buffer.tail is after buffer.LastPersistence
func (b *Buffer) decrementPersistentAtTail(node *DataNode) error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	if node == nil {
		return fmt.Errorf("there is no node to remove from persistence")
	}
	if node.ValueLen <= 0 {
		return fmt.Errorf("the node is not persistent before")
	}

	if b.FileEndSeek < BUFFER_INFO_SIZE+int64(node.ValueLen) {
		return fmt.Errorf("the node was persistented by covering the buffer info")
	}

	if b.FileStartSeek <= b.FileEndSeek-int64(node.ValueLen) {
		b.FileStartSeek = BUFFER_INFO_SIZE
		b.FileEndSeek = BUFFER_INFO_SIZE
	} else {
		b.FileEndSeek -= int64(node.ValueLen)
	}

	return nil
}

func (b *Buffer) Persistent() error {
	return b.incrementPersistent()
}

func (b *Buffer) recoveryInfo(file *os.File) error {
	if file == nil {
		return fmt.Errorf("should provide a file hanler")
	}

	err := gob.NewDecoder(file).Decode(b)
	if err != nil {
		return err
	}

	b.RecoveryControl = true
	b.Datas = NewDataLink()

	return nil
}

func (b *Buffer) recoveryData(file *os.File, fileseek int64) (*DataNode, int64, error) {
	var size int64

	indexbyte := make([]byte, 8)
	start := fileseek
	_, err := file.ReadAt(indexbyte, start)
	if err != nil {
		return nil, start, err
	}

	sizebuf := bytes.NewBuffer(indexbyte)
	err = gob.NewDecoder(sizebuf).Decode(&size)
	if err != nil {
		return nil, start, err
	}

	if start+8 > b.FileEndSeek || start+8+size > b.FileEndSeek {
		return nil, start, fmt.Errorf("data seek is exceed the end, file maybe destroyed!")
	}

	start += 8
	databyte := make([]byte, size)
	_, err = file.ReadAt(databyte, start)
	if err != nil {
		return nil, start, err
	}

	databuf := bytes.NewBuffer(databyte)
	value := reflect.New(reflect.TypeOf(b.Register))
	err = gob.NewDecoder(databuf).Decode(value.Interface())
	if err != nil {
		return nil, start, err
	}

	datanode := NewDataNode(value.Interface())
	datanode.ValueLen = size
	start += size

	return datanode, start, nil
}

func (b *Buffer) recoveryDataLink(file *os.File) error {
	if file == nil {
		return fmt.Errorf("should provide a file hanler")
	}

	if b.Register == nil {
		return fmt.Errorf("should register data type to recover datas")
	}

	var currentnode *DataNode
	fileseek := b.BufferInfo.FileStartSeek
	for fileseek < b.BufferInfo.FileEndSeek {
		datanode, seek, err := b.recoveryData(file, fileseek)
		if err != nil {
			return err
		}

		if b.Datas.Head == nil {
			b.Datas.Head = datanode
		}
		b.Datas.Tail = datanode
		b.Datas.LastPersistence = datanode

		if currentnode == nil {
			currentnode = datanode
		} else {
			currentnode.Next = datanode
			datanode.Previous = currentnode
			currentnode = datanode
		}

		fileseek = seek
	}

	return nil
}

func (b *Buffer) Recovery() error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	if !b.RecoveryControl {
		return fmt.Errorf("presistence is not enabled")
	}
	if b.File == "" {
		return fmt.Errorf("the file which persistence datas is not specified")
	}

	file, err := os.Open(b.File)
	if err != nil {
		return err
	}
	defer file.Close()

	err = b.recoveryInfo(file)
	if err != nil {
		return err
	}

	return b.recoveryDataLink(file)
}
