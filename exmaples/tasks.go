// +build ignore

package main

import (
	"fmt"
	"strings"

	"github.com/zain-bahsarat/fsml"
)

// create an entity which implements stateful interface
type item struct {
	state string
	count int
}

func (i *item) SetState(state string) error {
	i.state = state
	return nil
}

func (i *item) GetState() string {
	return i.state
}

type task struct {
	name string
}

func (t *task) Name() string {
	return t.name
}

func (t *task) Execute(i interface{}) error {
	entity := i.(*item)
	entity.count++

	return nil
}

func main() {
	statemachineDef := getTasksStatemachineDef()

	reader := strings.NewReader(statemachineDef)
	sm, err := fsml.New(reader)
	if err != nil {
		fmt.Printf("error= %+v\n", sm)
	}

	// create task
	t := &task{name: "increment"}
	sm.AddTask(t)

	// setup item
	o := &item{count: 0}
	o.SetState("new")

	if err := sm.Trigger("DummyEvent", o); err != nil {
		fmt.Printf("error= %+v\n", err)
	}

	fmt.Printf("Count= %d\n", o.count)
	fmt.Printf("State= %s\n", o.GetState())
}

func getTasksStatemachineDef() string {
	return `<Schema>
		<States>
			<new>
				<Events>
					<DummyEvent targetState="pending" errorState="error">
						<Task>increment</Task>
						<Task>increment</Task>
					</DummyEvent>
				</Events>
			</new>
			<pending>
				<OnStateSet>
					<Task>increment</Task>
				</OnStateSet>
			</pending>
			<error></error>
		</States>
	</Schema>`
}
