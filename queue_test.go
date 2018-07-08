package mtque

import (
	"testing"
	"time"
)

func TestNewEmptyQueue(t *testing.T) {
	queue := NewQueue()
	if queue == nil {
		t.Fatal("Construct a new empty queue fail!")
	}

	if len(queueList) > 0 {
		t.Fatal("Initailize the queue list wrong!")
	}
}

func TestSettingQueue(t *testing.T) {
	queue := NewQueue()

	t.Run("SetPersistencePeriod", func(t *testing.T) {
		queue.SetPersistencePeriod(5 * time.Minute)
		period := queue.GetPersistencePeriod()

		if period != 5*time.Minute {
			t.Fatal("queue: SetPersistencePeriod error!", period)
		}
	})
	t.Run("SetPersistenceControl", func(t *testing.T) {
		queue.SetPersistenceControl(true)
		ctl := queue.GetPersistenceControl()

		if ctl != true {
			t.Fatal("queue: SetPersistenceControl error!", ctl)
		}
	})
	t.Run("SetRecoveryControl", func(t *testing.T) {
		queue.SetRecoveryControl(true)
		ctl := queue.GetRecoveryControl()

		if ctl != true {
			t.Fatal("queue: SetRecoveryControl error!", ctl)
		}
	})
	t.Run("SetFile", func(t *testing.T) {
		err := queue.SetFile("./queuetest")
		if err != nil {
			t.Fatal("queue: SetFile error: ", err)
		}

		file := queue.GetFile()
		if file != "./queuetest" {
			t.Fatal("queue: SetFile error:", err)
		}
	})
}

func TestEnQueue(t *testing.T) {
	queue := NewQueue()

	queue.EnQueue(1)
	queue.EnQueue(2)
	queue.EnQueue(3)
	queue.EnQueue(4)
	len1 := queue.Len()

	if len1 != 4 {
		t.Fatal("Push queue error!")
	}
}

func TestGetQueueHead(t *testing.T) {
	queue := NewQueue()

	queue.EnQueue(1)
	queue.EnQueue(2)
	queue.EnQueue(3)
	queue.EnQueue(4)

	len1 := queue.Len()
	val1, err := queue.GetHead()
	if err != nil {
		t.Fatal("GetHead error:", err)
	}

	len2 := queue.Len()
	val2, err := queue.GetHead()
	if err != nil {
		t.Fatal("GetHead error:", err)
	}

	len3 := queue.Len()

	if len1 != len2 && len2 != len3 && len3 != 4 {
		t.Fatal("GetHead cause the length of queue wrong!")
	}

	if val1 != val2 && val2 != 1 {
		t.Fatal("GetHead got 2 different values in twice fetching!")

	}
}

func TestDeQueue(t *testing.T) {
	queue := NewQueue()

	queue.EnQueue(1)
	queue.EnQueue(2)
	queue.EnQueue(3)
	queue.EnQueue(4)

	len1 := queue.Len()
	val1, err := queue.DeQueue()
	if err != nil {
		t.Fatal("dequeue error:", err)
	}

	len2 := queue.Len()
	val2, err := queue.DeQueue()
	if err != nil {
		t.Fatal("dequeue error:", err)
	}

	len3 := queue.Len()

	if len1 != len2+1 && len2 != len3+1 && len3 != 2 {
		t.Fatal("dequeue length error!", len3, len2, len1)
	}

	if val1 != 1 && val2 != 2 {
		t.Fatal("dequeue value error!")
	}
}
