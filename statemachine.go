package fsml

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/zain-bahsarat/fsml/internal/parser"
	"github.com/zain-bahsarat/fsml/internal/schema"

	"github.com/pkg/errors"
)

// Statemachine ...
type Statemachine struct {
	fsmWrapper *fsmWrapper
}

// New ...
func New(input io.Reader) (*Statemachine, error) {

	buf, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, err
	}

	p := parser.New(parser.NewLexer(string(buf)))
	schma, err := schema.New(p)
	if err != nil {
		return nil, err
	}

	return &Statemachine{fsmWrapper: newFSMWrapper(*schma)}, nil
}

// Trigger ...
func (s *Statemachine) Trigger(eventName string, entity interface{}) error {
	fsm, err := s.fsmWrapper.newFSM(entity)
	if err != nil {
		return err
	}

	err = fsm.Event(eventName)
	if err != nil {
		failedState := createFailedStateEvent(eventName)
		if fsm.Can(failedState) {
			if err := fsm.Event(failedState); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	stateful, ok := entity.(Stateful)
	if !ok {
		return errors.Wrap(errMissingStatefulInterface, fmt.Sprintf("%+v: ", entity))
	}

	return stateful.SetState(fsm.Current())
}

// Can ...
func (s *Statemachine) Can(eventName string, entity interface{}) bool {
	fsm, err := s.fsmWrapper.newFSM(entity)
	if err != nil {
		return false
	}

	return fsm.Can(eventName)
}

// AddTask ...
func (s *Statemachine) AddTask(task Task) error {
	return s.fsmWrapper.taskCollection.addTask(task)
}

// RemoveTask ...
func (s *Statemachine) RemoveTask(task Task) error {
	return s.fsmWrapper.taskCollection.removeTask(task)
}
