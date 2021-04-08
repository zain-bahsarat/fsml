package fsml

import (
	"fmt"

	"github.com/looplab/fsm"
	"github.com/pkg/errors"
	S "github.com/zain-bahsarat/fsml/internal/schema"
)

var (
	errTaskNotFound             = errors.New("task not found")
	errTaskAlreadyExists        = errors.New("task already exists")
	errMissingStatefulInterface = errors.New("must implement stateful interface")
)

// Stateful ...
type Stateful interface {
	GetState() string
	SetState(state string) error
}

// Task ...
type Task interface {
	Name() string
	Execute(entity interface{}) error
}

func buildTasksLookup(schema S.Schema) map[string][]string {

	lookupTable := make(map[string][]string)

	addToLookup := func(key string, data []string) {
		if _, ok := lookupTable[key]; !ok {
			lookupTable[key] = []string{}
		}
		lookupTable[key] = append(lookupTable[key], data...)
	}

	setDefaultEvents := func(eventName string, de S.DefaultEvents) {
		// @TODO make copies
		if len(de.OnAfterEvent.Tasks) > 0 {
			addToLookup("after_"+eventName, de.OnAfterEvent.Tasks)
		}

		if len(de.OnBeforeEvent.Tasks) > 0 {
			addToLookup("before_"+eventName, de.OnBeforeEvent.Tasks)
		}
	}

	setStateEvents := func(stateName string, de S.DefaultEvents) {
		// @TODO make copies
		if len(de.OnStateSet.Tasks) > 0 {
			addToLookup("enter_"+stateName, de.OnStateSet.Tasks)
		}
	}

	setDefaultEvents("event", schema.DefaultEvents)
	setStateEvents("state", schema.DefaultEvents)

	for _, s := range schema.States {

		setStateEvents(s.Name, s.DefaultEvents)

		for _, e := range s.Events {
			setDefaultEvents(e.Name, s.DefaultEvents)
			addToLookup("before_"+e.Name, e.Tasks)
		}
	}

	return lookupTable
}

func buildFSMEvents(schema S.Schema) []fsm.EventDesc {
	events := make([]fsm.EventDesc, 0)
	for _, s := range schema.States {
		for _, e := range s.Events {
			events = append(events, fsm.EventDesc{Name: e.Name, Src: []string{s.Name}, Dst: e.TargetState})
			if len(e.ErrorState) > 0 {
				failedName := createFailedStateEvent(e.Name)
				events = append(events, fsm.EventDesc{Name: failedName, Src: []string{s.Name}, Dst: e.ErrorState})
			}
		}
	}

	return events
}

func buildFSMCallbacks(schema S.Schema, cb func(trigger string, event *fsm.Event)) fsm.Callbacks {
	callbacks := fsm.Callbacks{}

	setDefaultEvents := func(eventName string, de S.DefaultEvents) {
		// @TODO make copies
		if len(de.OnAfterEvent.Tasks) > 0 {
			callbacks["after_"+eventName] = func(event *fsm.Event) {
				cb("after_"+eventName, event)
			}
		}

		if len(de.OnBeforeEvent.Tasks) > 0 {
			callbacks["before_"+eventName] = func(event *fsm.Event) {
				cb("before_"+eventName, event)
			}
		}
	}

	setStateEvents := func(stateName string, de S.DefaultEvents) {
		// @TODO make copies
		if len(de.OnStateSet.Tasks) > 0 {
			callbacks["enter_"+stateName] = func(event *fsm.Event) {
				cb("enter_"+stateName, event)
			}
		}
	}

	setDefaultEvents("event", schema.DefaultEvents)
	setStateEvents("state", schema.DefaultEvents)

	for _, s := range schema.States {

		setStateEvents(s.Name, s.DefaultEvents)

		for _, e := range s.Events {
			setDefaultEvents(e.Name, s.DefaultEvents)
			callbacks["before_"+e.Name] = func(event *fsm.Event) {
				cb("before_"+e.Name, event)
			}
		}
	}

	return callbacks
}

func createFailedStateEvent(eventName string) string {
	return fmt.Sprintf("%s_failed", eventName)
}

type fsmWrapper struct {
	schema          S.Schema
	events          []fsm.EventDesc
	taskCollection  taskCollection
	taskLookupTable map[string][]string
}

func newFSMWrapper(schema S.Schema) *fsmWrapper {
	tCollection := taskCollection{tasks: make(map[string]Task)}
	events := buildFSMEvents(schema)
	lookupTable := buildTasksLookup(schema)

	return &fsmWrapper{
		schema:          schema,
		events:          events,
		taskCollection:  tCollection,
		taskLookupTable: lookupTable,
	}
}

func (wrapper *fsmWrapper) newFSM(entity interface{}) (*fsm.FSM, error) {
	stateful, ok := entity.(Stateful)
	if !ok {
		return nil, errors.Wrap(errMissingStatefulInterface, fmt.Sprintf("%+v: ", entity))
	}

	callbacks := buildFSMCallbacks(wrapper.schema, func(trigger string, event *fsm.Event) {
		if taskNames, ok := wrapper.taskLookupTable[trigger]; ok {
			for _, taskName := range taskNames {
				task, err := wrapper.taskCollection.get(taskName)
				if err != nil {
					event.Cancel(err)
					return
				}

				if err := task.Execute(entity); err != nil {
					event.Cancel(err)
					return
				}
			}
		}
	})

	return fsm.NewFSM(
		stateful.GetState(),
		wrapper.events,
		callbacks,
	), nil
}

type taskCollection struct {
	tasks map[string]Task
}

func (collection *taskCollection) addTask(t Task) error {
	if _, ok := collection.tasks[t.Name()]; ok {
		return errors.Wrap(errTaskAlreadyExists, t.Name())
	}

	collection.tasks[t.Name()] = t
	return nil
}

func (collection *taskCollection) removeTask(t Task) error {
	if _, ok := collection.tasks[t.Name()]; !ok {
		return errors.Wrap(errTaskNotFound, t.Name())
	}

	delete(collection.tasks, t.Name())
	return nil
}

func (collection *taskCollection) get(taskName string) (Task, error) {

	task, ok := collection.tasks[taskName]
	if !ok {
		return task, errors.Wrap(errTaskNotFound, taskName)
	}

	return task, nil
}
