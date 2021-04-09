# fsml ![Bulid Status](https://github.com/zain-bahsarat/fsml/actions/workflows/test.yml/badge.svg) [![Coverage Status](https://img.shields.io/coveralls/zain-bahsarat/fsml.svg)](https://coveralls.io/r/zain-bahsarat/fsml) [![Go Report Card](https://goreportcard.com/badge/zain-bahsarat/fsml)](https://goreportcard.com/report/zain-bahsarat/fsml)

**FSML** is a XML based wrapper on top of https://github.com/looplab/fsm.<br>
It provides the capabilities to define state machine of an entity in XML format and also supports features to handle error states and tasks execution on events and successful transition.
Here is an example of state machine definition in xml format.

```xml
<Schema>
    <States>
        <new>
            <Events>
                <DummyEvent targetState="pending" errorState="error"></DummyEvent>
            </Events>
        </new>
        <pending></pending>
        <error></error>
    </States>
</Schema>
```

Above statemachine has three states `new`, `pending` and `error` and event named `DummyEvent`

## Schema Definition

### Nodes

- Schema `Root Node`
- States `Container Node`
  - All the state definitions will be inside this node
- Events `Container Node`
- Task
- OnStateSet `Default Event`
- OnAfterEvent `Default Event`
- OnBeforeEvent `Default Event`

### Default Events

Default events can used inside each state to define default behaviors when state is updated. you can also define global default events.

```xml
 <Schema>
    <OnStateSet></OnStateSet>
    <States>
        <new>
            <OnStateSet></OnStateSet>
            <Events>
                <DummyEvent targetState="pending" errorState="error"></DummyEvent>
            </Events>
        </new>
        <pending></pending>
        <error></error>
    </States>
</Schema>
```

### Custom Events

Custom events can be deined inside `Events` Node. There is an option to define `targetState`(required) and `errorState` which will take effect based on transition result

### Tasks

`Task` Node is defined inside Custom Event or Default Event when we want to execute some task on them. If all tasks defined inside event are executed successfully then state will be changed to `targetState` otherwise it will be `errorState`

Every task needs to implement `fsml.Task` interface to be accessible by statemachine. check `examples/tasks.go`

```go
    type task struct {}
    func (t *task) Name() string {
        return "increment"
    }

    func (t *task) Execute(i interface{}) error {
        entity := i.(*item)
        entity.count++

        return nil
    }

    ......

    statemachine.AddTask(&task{})
```

```xml
    <new>
        <OnStateSet></OnStateSet>
        <Events>
            <DummyEvent targetState="pending" errorState="error">
                <Task>increment</Task>
                <Task>increment</Task>
            </DummyEvent>
        </Events>
    </new>
```

---

<br>

### Code Example `examples/basic.go`ßß

```go
    import (
        "fmt"
        "strings"

        "github.com/zain-bahsarat/fsml"
    )

    // Every object which is passed to the statemachine has to implement
    // `Stateful` interface otherwise it will throw error
    // create an entity which implements stateful interface
    type order struct {
        state string
        fsml.Stateful
    }

    func (o *order) SetState(state string) error {
        o.state = state
        return nil
    }

    func (o *order) GetState() string {
        return o.state
    }

    func main() {
        statemachineDef := getSimpleStatemachineDef()

        reader := strings.NewReader(statemachineDef)
        sm, err := fsml.New(reader)
        if err != nil {
            fmt.Printf("error= %+v\n", sm)
        }

        o := &order{}
        o.SetState("new")

        if err := sm.Trigger("DummyEvent", o); err != nil {
            fmt.Printf("error= %+v\n", err)
        }

        if !sm.Can("UndefinedEvent", o) {
            fmt.Printf("cannot trigger UndefinedEvent on %+v\n", o)
        }

    }

    // state machine definition
    func getSimpleStatemachineDef() string {
        return `<Schema>
                <States>
                    <new>
                        <Events>
                            <DummyEvent targetState="pending" errorState="error"></DummyEvent>
                        </Events>
                    </new>
                    <pending></pending>
                    <error></error>
                </States>
            </Schema>`
    }
```

# License

FSML is licensed under Apache License 2.0

http://www.apache.org/licenses/LICENSE-2.0
