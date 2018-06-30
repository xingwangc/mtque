package mtque

import (
	"testing"
)

type TestData struct {
	Name string
	Age  int
}

func TestBufferPersistent(t *testing.T) {
	buf := NewBuffer(
		SetBufferFile("./buffer_persistent"),
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
		SetBufferFile("./buffer_persistent"),
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
