package mtque

import (
	"testing"
	"time"
)

func TestNewEmptyStack(t *testing.T) {
	stack := NewStack()
	if stack == nil {
		t.Fatal("Construct a new empty stack fail!")
	}

	if len(stackList) > 0 {
		t.Fatal("Initailize the stack list wrong!")
	}
}

func TestSetting(t *testing.T) {
	stack := NewStack()

	t.Run("SetPersistencePeriod", func(t *testing.T) {
		stack.SetPersistencePeriod(5 * time.Minute)
		period := stack.GetPersistencePeriod()

		if period != 5*time.Minute {
			t.Fatal("Stack: SetPersistencePeriod error!", period)
		}
	})
	t.Run("SetPersistenceControl", func(t *testing.T) {
		stack.SetPersistenceControl(true)
		ctl := stack.GetPersistenceControl()

		if ctl != true {
			t.Fatal("Stack: SetPersistenceControl error!", ctl)
		}
	})
	t.Run("SetRecoveryControl", func(t *testing.T) {
		stack.SetRecoveryControl(true)
		ctl := stack.GetRecoveryControl()

		if ctl != true {
			t.Fatal("Stack: SetRecoveryControl error!", ctl)
		}
	})
	t.Run("SetFile", func(t *testing.T) {
		err := stack.SetFile("./stacktest")
		if err != nil {
			t.Fatal("Stack: SetFile error: ", err)
		}

		file := stack.GetFile()
		if file != "./stacktest" {
			t.Fatal("Stack: SetFile error:", err)
		}
	})
}

func TestPushStack(t *testing.T) {
	stack := NewStack()

	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	stack.Push(4)
	len1 := stack.Len()

	if len1 != 4 {
		t.Fatal("Push stack error!")
	}
}

func TestGetStackTail(t *testing.T) {
	stack := NewStack()

	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	stack.Push(4)

	len1 := stack.Len()
	val1, err := stack.GetTail()
	if err != nil {
		t.Fatal("GetTail error:", err)
	}

	len2 := stack.Len()
	val2, err := stack.GetTail()
	if err != nil {
		t.Fatal("GetTail error:", err)
	}

	len3 := stack.Len()

	if len1 != len2 && len2 != len3 && len3 != 4 {
		t.Fatal("Getail cause the length of stack wrong!")
	}

	if val1 != val2 && val2 != 4 {
		t.Fatal("Getail got 2 different value in twice poping!")
	}
}

func TestPopStack(t *testing.T) {
	stack := NewStack()

	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	stack.Push(4)

	len1 := stack.Len()
	val1, err := stack.Pop()
	if err != nil {
		t.Fatal("Stack pop error:", err)
	}

	len2 := stack.Len()
	val2, err := stack.Pop()
	if err != nil {
		t.Fatal("Stack pop error:", err)
	}

	len3 := stack.Len()

	if len1 != len2+1 && len2 != len3+1 && len3 != 2 {
		t.Fatal("Stack pop length error!", len3, len2, len1)
	}

	if val1 != 4 && val2 != 3 {
		t.Fatal("Stack pop value error!")
	}
}

func TestPeriodicallyPersistence(t *testing.T) {
	stack := NewStack(
		SetStackFile("./periodically"),
		SetStackPersistencePeriod(10*time.Millisecond),
		SetStackPersistenceControl(true),
	)

	stack.Push(1)
	time.Sleep(5 * time.Millisecond)
	stack.Push(3)
	time.Sleep(5 * time.Millisecond)
	stack.Push(5)
	time.Sleep(5 * time.Millisecond)
	stack.Pop()
	time.Sleep(5 * time.Millisecond)
	stack.Push(7)
	time.Sleep(5 * time.Millisecond)
	stack.Push(9)
	stack.Pop()
	time.Sleep(10 * time.Millisecond)
}

func TestRecovery(t *testing.T) {
	stack := NewStack(
		SetStackFile("./periodically"),
		SetStackRecoveryControl(true),
	)

	if stack.Len() != 3 {
		t.Fatal("Recovery error:", stack.Len())
	}

	v, err := stack.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if v != 7 {
		t.Fatal("Wrong value:", v)
	}

	v, err = stack.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if v != 3 {
		t.Fatal("Wrong value:", v)
	}

	v, err = stack.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 {
		t.Fatal("Wrong value:", v)
	}
}
