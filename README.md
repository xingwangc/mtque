# Install

    go get -u github.com/xingwangc/mtque

# Example

## Use the Queue

```
    import (
        "github.com/xingwangc/mtque"
    )

    func main() {
        queue := mtque.NewQueue()
        queue.EnQueue(10)
        value, _ := queue.DeQueue()
    }
```

## Use Queue with Persistence

```
    import (
        "time"

        "github.com/xingwang/mtqueue"
    }

    func main() {
        queue := mtque.NewQueue(
            mtque.SetQueueFile("./test"),
            mtque.SetQueuePersistenceControl(true),
            mtque.SetQueuePersistencePeriod(10 * time.Second))
        queue.EnQueue(10)
        value, _ := queue.DeQueue()
    }
```

## Init a Queue by recovering from a file

```
    import (
        "github.com/xingwang/mtqueue"
    }

    func main() {
        queue := mtque.NewQueue(
            mtque.SetQueueFile("./test"),
            mtque.SetQueueRecoveryControl(true))
        queue.EnQueue(10)
        value, _ := queue.DeQueue()
    }
```

## Force recover the queue from another file


```
    import (
        "github.com/xingwang/mtqueue"
    }

    func main() {
        queue := mtque.NewQueue()
        queue.EnQueue(10)
        value, _ := queue.DeQueue()

        queue.SetRecoveryControl(true)
        queue.ForceSetFile("./test")
    }
```

## Use the Stack

```
    import (
        "github.com/xingwangc/mtque"
    )

    func main() {
        stack := mtque.NewStack()
        stack.Push(10)
        value, _ := stack.Pop()
    }
```

## Use Stack with Persistence

```
    import (
        "time"

        "github.com/xingwang/mtqueue"
    }

    func main() {
        stack := mtque.NewStack(
            SetStackFile("./test"),
            SetStackPersistenceControl(true),
            SetStackPersistencePeriod(10 * time.Second))
        stack.Push(10)
        value, _ := stack.Pop()
    }
```

## Init a Stack by recovering from a file

```
    import (
        "github.com/xingwang/mtqueue"
    }

    func main() {
        stack := NewStack(
            SetStackFile("./test"),
            SetStackRecoveryControl(true))
        stack.Push(10)
        value, _ := stack.Pop()
    }
```

## Force recover the Stack from another file


```
    import (
        "github.com/xingwang/mtqueue"
    }

    func main() {
        stack := mtque.NewStack()
        stack.Push(10)

        stack.SetRecoveryControl(true)
        stack.ForceSetFile("./test")
    }
```
