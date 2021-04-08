package fsml

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zain-bahsarat/fsml/internal/parser"
	"github.com/zain-bahsarat/fsml/internal/schema"
)

type dummyItem struct {
	id    int
	state string
}

func (o *dummyItem) SetState(s string) error {
	o.state = s
	return nil
}

func (o *dummyItem) GetState() string {
	return o.state
}

type dummyTask struct{}

func (t *dummyTask) Name() string {
	return "dummy"
}

func (t *dummyTask) Execute(entity interface{}) error {
	order, ok := entity.(*dummyItem)
	if !ok {
		return errors.New("wrong entity type")
	}
	order.id++

	return nil
}

func TestWrapper_Simple(t *testing.T) {
	input := `<Schema>
	<OnBeforeEvent>
		<Task>dummy</Task>
	</OnBeforeEvent>
		<States>
			<new>
				<OnBeforeEvent>
				<Task>dummy</Task>
			</OnBeforeEvent>
			<OnAfterEvent>
				<Task>dummy</Task>
			</OnAfterEvent>
			<OnStateSet>
				<Task>dummy</Task>
			</OnStateSet>
				<Events>
					<DummyEvent targetState="pending" errorState="error">
						<Task>dummy</Task>
					</DummyEvent>
				</Events>
			</new>
			<pending></pending>
			<error></error>
		</States>
	</Schema>`
	p := parser.New(parser.NewLexer(input))
	s, err := schema.New(p)
	assert.Nil(t, err)

	wrapper := newFSMWrapper(*s)
	err = wrapper.taskCollection.addTask(&dummyTask{})
	assert.Nil(t, err)

	order := &dummyItem{id: 1, state: "new"}
	fsm, err := wrapper.newFSM(order)

	assert.Nil(t, err)

	err = fsm.Event("DummyEvent")
	assert.Nil(t, err)
	assert.Nil(t, order.SetState(fsm.Current()))

	assert.Equal(t, &dummyItem{id: 5, state: "pending"}, order, "Order not equal")
}

func TestWrapper_errTaskNotFound(t *testing.T) {
	input := `<Schema>
	<OnBeforeEvent>
		<Task>dummy</Task>
	</OnBeforeEvent>
		<States>
			<new>
				<Events>
					<DummyEvent targetState="pending" errorState="error">
					
					</DummyEvent>
				</Events>
			</new>
		</States>
	</Schema>`
	p := parser.New(parser.NewLexer(input))
	s, err := schema.New(p)
	assert.Nil(t, err)

	wrapper := newFSMWrapper(*s)

	// check in collection
	_, err = wrapper.taskCollection.get("dummy2")
	assert.True(t, strings.Contains(err.Error(), errTaskNotFound.Error()))

	// add and then remove in collection
	err = wrapper.taskCollection.addTask(&dummyTask{})
	assert.Nil(t, err)

	// remove from collection
	err = wrapper.taskCollection.removeTask(&dummyTask{})
	assert.Nil(t, err)

	// again remove from collection
	err = wrapper.taskCollection.removeTask(&dummyTask{})
	assert.True(t, strings.Contains(err.Error(), errTaskNotFound.Error()))

	order := &dummyItem{id: 1, state: "new"}
	fsm, err := wrapper.newFSM(order)
	assert.Nil(t, err)

	err = fsm.Event("DummyEvent")
	assert.True(t, strings.Contains(err.Error(), errTaskNotFound.Error()))
}

func TestWrapper_TaskAlreayExists(t *testing.T) {
	input := `<Schema>
	<OnBeforeEvent>
		<Task>dummy</Task>
	</OnBeforeEvent>
	<OnAfterEvent>
		<Task>dummy</Task>
	</OnAfterEvent>
	<OnStateSet>
		<Task>dummy</Task>
	</OnStateSet>
		<States>
			<new>
				<OnBeforeEvent>
					<Task>dummy</Task>
				</OnBeforeEvent>
				<OnAfterEvent>
					<Task>dummy</Task>
				</OnAfterEvent>
				<OnStateSet>
					<Task>dummy</Task>
				</OnStateSet>
				<Events>
					<DummyEvent targetState="pending" errorState="error">
					</DummyEvent>
				</Events>
			</new>
		</States>
	</Schema>`
	p := parser.New(parser.NewLexer(input))
	s, err := schema.New(p)
	assert.Nil(t, err)

	wrapper := newFSMWrapper(*s)
	err = wrapper.taskCollection.addTask(&dummyTask{})
	assert.Nil(t, err)

	err = wrapper.taskCollection.addTask(&dummyTask{})
	assert.True(t, strings.Contains(err.Error(), errTaskAlreadyExists.Error()))
}

func TestWrapper_TaskLookupTable(t *testing.T) {
	input := `<Schema>
	<OnBeforeEvent>
		<Task>dummy1</Task>
	</OnBeforeEvent>
	<OnAfterEvent>
		<Task>dummy1</Task>
	</OnAfterEvent>
	<OnStateSet>
		<Task>dummy1</Task>
	</OnStateSet>
		<States>
			<new>
				<OnBeforeEvent>
					<Task>dummy2</Task>
					<Task>dummy22</Task>
				</OnBeforeEvent>
				<OnAfterEvent>
					<Task>dummy4</Task>
				</OnAfterEvent>
				<OnStateSet>
					<Task>dummy3</Task>
				</OnStateSet>
				<Events>
					<DummyEvent targetState="pending" errorState="error">
					</DummyEvent>
				</Events>
			</new>
		</States>
	</Schema>`
	p := parser.New(parser.NewLexer(input))
	s, err := schema.New(p)
	assert.Nil(t, err)

	wrapper := newFSMWrapper(*s)

	expected := map[string][]string{
		"after_DummyEvent":  {"dummy4"},
		"before_DummyEvent": {"dummy2", "dummy22"},
		"before_event":      {"dummy1"},
		"after_event":       {"dummy1"},
		"enter_new":         {"dummy3"},
		"enter_state":       {"dummy1"},
	}

	assert.Equal(t, expected, wrapper.taskLookupTable, "Lookup table is not equal")
}

type StatefulMock struct{ s string }

func (s *StatefulMock) GetState() string {
	return ""
}

func (s *StatefulMock) SetState(input string) error {
	s.s = input
	return nil
}

func TestWrapper_StatefulInterface(t *testing.T) {
	input := `<Schema>
		<States>
			<new>
			</new>
		</States>
	</Schema>`
	p := parser.New(parser.NewLexer(input))
	s, err := schema.New(p)
	assert.Nil(t, err)

	wrapper := newFSMWrapper(*s)
	_, err = wrapper.newFSM(struct{}{})

	assert.True(t, strings.Contains(err.Error(), errMissingStatefulInterface.Error()))
}
