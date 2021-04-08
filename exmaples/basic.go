// +build ignore

package main

import (
	"fmt"
	"strings"

	"github.com/zain-bahsarat/fsml"
)

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
