package fsml

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test item
type testItem struct {
	count int
	state string
}

func (i *testItem) SetState(state string) error {
	i.state = state
	return nil
}

func (i *testItem) GetState() string {
	return i.state
}

// Test task
type testTask struct {
	name      string
	executeFn func(entity interface{}) error
}

func (t *testTask) Name() string {
	return t.name
}

func (t *testTask) Execute(entity interface{}) error {
	return t.executeFn(entity)
}

func TestStatemachine_Simple(t *testing.T) {
	input := ``

	_, err := New(strings.NewReader(input))
	assert.NotNil(t, err)
}

func TestStatemachine_Simple_Trigger(t *testing.T) {
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

	sm, err := New(strings.NewReader(input))
	assert.Nil(t, err)

	task1 := &testTask{name: "dummy", executeFn: func(entity interface{}) error {
		item := entity.(*testItem)
		item.count++
		return nil
	}}
	err = sm.AddTask(task1)
	assert.Nil(t, err)

	item := &testItem{state: "new"}
	err = sm.Trigger("DummyEvent", item)
	assert.Nil(t, err)
	assert.Equal(t, item.count, 4, "Item count is not equal")
	assert.Equal(t, item.GetState(), "pending", "Item state is not equal")
}

func TestStatemachine_Error_Trigger(t *testing.T) {
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
						<Task>dummy2</Task>
					</DummyEvent>
				</Events>
			</new>
			<pending></pending>
			<error></error>
		</States>
	</Schema>`

	sm, err := New(strings.NewReader(input))
	if err != nil {
		assert.Nil(t, err)
		return
	}

	task1 := &testTask{name: "dummy", executeFn: func(entity interface{}) error {
		item := entity.(*testItem)
		item.count++
		return nil
	}}
	err = sm.AddTask(task1)
	assert.Nil(t, err)

	item := &testItem{state: "new"}

	task2 := &testTask{name: "dummy2", executeFn: func(entity interface{}) error {
		return errors.New("test error")
	}}

	err = sm.AddTask(task2)
	assert.Nil(t, err)

	can := sm.Can("DummyEvent", item)
	assert.True(t, can)

	err = sm.Trigger("DummyEvent", item)

	assert.Nil(t, err)
	assert.NotEqual(t, item.count, 4, "Item count is not expected")
	assert.Equal(t, "error", item.GetState(), "Item state is not equal")

	err = sm.RemoveTask(task2)
	assert.Nil(t, err)
}

func TestStatemachine_Invalid_Schema(t *testing.T) {
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
			<Events>
					<DummyEvent targetState="pending" errorState="error">
						<Task>dummy</Task>
						<Task>dummy2</Task>
					</DummyEvent>
				</Events>
			</new>
			<pending></pending>
			<error></error>
		</States
	</Schema>`

	_, err := New(strings.NewReader(input))
	assert.NotNil(t, err)
}

func TestStatemachine_Invalid_Event(t *testing.T) {
	input := `<Schema>
		<States>
			<new>
			<Events>
					<DummyEvent targetState="pending" errorState="error">
						<Task>dummy</Task>
					</DummyEvent>
				</Events>
			</new>
			<pending>
			</pending>
			<error>
			<OnStateSet>
				<Task>dummy1</Task>
			</OnStateSet>
			</error>
		</States>
	</Schema>`

	sm, err := New(strings.NewReader(input))
	assert.Nil(t, err)

	item := &testItem{state: "new"}
	err = sm.Trigger("Event", item)
	assert.NotNil(t, err)

	task := &testTask{name: "dummy", executeFn: func(entity interface{}) error {
		return errors.New("error")
	}}
	err = sm.AddTask(task)
	assert.Nil(t, err)

	err = sm.Trigger("DummyEvent", item)
	assert.NotNil(t, err)
}

func TestStatemachine_Invalid_Entity(t *testing.T) {
	input := `<Schema>
		<States>
			<new>
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

	sm, err := New(strings.NewReader(input))
	assert.Nil(t, err)

	err = sm.Trigger("Event", nil)
	assert.NotNil(t, err)
}

func TestStatemachine_Can(t *testing.T) {
	input := `<Schema>
		<States>
			<new>
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

	sm, err := New(strings.NewReader(input))
	assert.Nil(t, err)

	item := &testItem{state: "new"}
	can := sm.Can("Event", item)
	assert.False(t, can)

	can = sm.Can("Event", nil)
	assert.False(t, can)
}
