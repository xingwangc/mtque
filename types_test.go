package mtque

import (
	"testing"
)

type TestData struct {
	Name string
	Age  int
}

type TestData2 struct {
	Name    string
	Address string
	Number  int
	People  map[string]string
	Lists   []int
}

func TestBufferPersistent2(t *testing.T) {
	buf := NewBuffer(
		//SetBufferFile("./buffer_persistent"),
		SetBufferFile("./buffer_struct2"),
		SetBufferPersistenceControl(true))

	buf.AddDataAtHead(TestData2{"a-a", "address a", 1,
		map[string]string{"a-a": "address a"}, []int{1, 1, 1, 1}})
	buf.AddDataAtHead(TestData2{"b-b", "address b", 2,
		map[string]string{"b-b": "address b"}, []int{2, 2, 2, 2}})
	buf.AddDataAtHead(TestData2{"c-c", "address c", 3,
		map[string]string{"c-c": "address c"}, []int{3, 3, 3, 3}})
	buf.AddDataAtHead(TestData2{"d-d", "address d", 4,
		map[string]string{"d-d": "address d"}, []int{4, 4, 4, 4}})

	err := buf.Persistent()
	if err != nil {
		t.Fatal(err)
	}
}

func TestBufferPersistent(t *testing.T) {
	buf := NewBuffer(
		//SetBufferFile("./buffer_persistent"),
		SetBufferFile("./buffer_ptt"),
		SetBufferPersistenceControl(true))

	buf.AddDataAtHead(TestData{"AA", 1})
	buf.AddDataAtHead(TestData{"BB", 2})
	buf.AddDataAtHead(TestData{"CC", 3})
	buf.AddDataAtHead(TestData{"DD", 4})
	buf.AddDataAtHead(TestData{"EE", 5})
	buf.AddDataAtHead(TestData{"FF", 6})
	buf.AddDataAtHead(TestData{"GG", 7})
	buf.AddDataAtHead(TestData{"HH", 8})
	buf.AddDataAtHead(TestData{"II", 9})
	buf.AddDataAtHead(TestData{"JJ", 10})

	err := buf.Persistent()
	if err != nil {
		t.Fatal(err)
	}
}

func TestBufferRecovery(t *testing.T) {
	buf := NewBuffer(
		//SetBufferFile("./buffer_persistent"),
		SetBufferFile("./buffer_ptt"),
		SetBufferRecoveryControl(true),
		SetBufferRegister(TestData{}))

	err := buf.Recovery()
	if err != nil {
		t.Fatal(err)
	}

	for node := buf.Datas.Head; node != nil; node = node.Next {
		t.Log("Value:", node.Value)
	}
}

func TestBufferRecovery2(t *testing.T) {
	buf := NewBuffer(
		//SetBufferFile("./buffer_persistent"),
		SetBufferFile("./buffer_struct2"),
		SetBufferRecoveryControl(true),
		SetBufferRegister(TestData2{}))

	err := buf.Recovery()
	if err != nil {
		t.Fatal(err)
	}

	for node := buf.Datas.Head; node != nil; node = node.Next {
		t.Log("Value:", node.Value)
	}
}
