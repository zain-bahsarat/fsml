package schema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zain-bahsarat/fsml/internal/parser"
	"github.com/zain-bahsarat/fsml/internal/queue"
)

const (
	// Predefined Nodes
	SchemaNodeName = "Schema"
	StatesNodeName = "States"
	TaskNodeName   = "Task"
	EventsNodeName = "Events"

	// Event Nodes
	OnBeforeEvent = "OnBeforeEvent"
	OnAfterEvent  = "OnAfterEvent"
	OnStateSet    = "OnStateSet"

	// Attributes
	TargetState = "targetState"
	ErrorState  = "errorState"
)

var defaultEvents = map[string]string{
	"OnBeforeEvent": OnBeforeEvent,
	"OnAfterEvent":  OnAfterEvent,
	"OnStateSet":    OnStateSet,
}

func isDefaultEventNode(str string) bool {
	if _, ok := defaultEvents[str]; !ok {
		return false
	}

	return true
}

type SchemaNode struct {
	N              parser.Node
	ParentNodeName string
	ParentNodeType parser.NodeType
}

type ConditionFn func(c Conditions) bool

type Conditions struct {
	ParentNodeName string
	ParentNodeType parser.NodeType
	NodeName       string
	NodeType       parser.NodeType
	CustomFn       ConditionFn
}

func (c *Conditions) suffice(c1 Conditions) bool {
	if len(c.ParentNodeName) > 0 && c1.ParentNodeName != c.ParentNodeName {
		return false
	}

	if len(c.ParentNodeType) > 0 && c1.ParentNodeType != c.ParentNodeType {
		return false
	}

	if len(c.NodeName) > 0 && c1.NodeName != c.NodeName {
		return false
	}

	if len(c.NodeType) > 0 && c1.NodeType != c.NodeType {
		return false
	}

	if c.CustomFn != nil {
		return c.CustomFn(c1)
	}

	return true
}

type Rule struct {
	Msg        string
	Criteria   Conditions
	Validation Conditions
}

func (r *Rule) Applicable(c Conditions) bool {
	return r.Criteria.suffice(c)
}

func (r *Rule) Validate(c Conditions) bool {
	return r.Validation.suffice(c)
}

type SchemaChecker struct {
	root         parser.Node
	states       map[string]bool
	visitedNodes map[string]int
}

func (sc *SchemaChecker) Validate() error {
	var errorList []string
	q := queue.New()

	q.Enqueue(SchemaNode{N: sc.root})
	for len(q.Items()) > 0 {

		qlen := len(q.Items())
		for i := 0; i < qlen; i++ {
			cur, ok := q.Dequeue().(SchemaNode)
			if !ok {
				return errors.New("type consversion error.")
			}

			sc.visitedNodes[cur.N.Name] += 1
			if err := sc.applyRules(cur); err != nil {
				errorList = append(errorList, err.Error())
			}

			for _, child := range cur.N.Children {
				// check if custom event
				if child.Type == parser.ElementNode && cur.N.Name == StatesNodeName {
					sc.states[child.Name] = true
				}

				q.Enqueue(SchemaNode{N: child, ParentNodeName: cur.N.Name, ParentNodeType: cur.N.Type})
			}
		}
	}

	// check required nodes
	for _, required := range sc.requiredNodes() {
		if _, ok := sc.visitedNodes[required]; !ok {
			errorList = append(errorList, fmt.Sprintf("Missing %s node", required))
		}
	}

	if len(errorList) > 0 {
		return fmt.Errorf("errors: \n--- %s", strings.Join(errorList, "\n--- "))
	}

	return nil
}

func (sc *SchemaChecker) requiredNodes() []string {
	return []string{
		StatesNodeName,
		SchemaNodeName,
	}
}

func (sc *SchemaChecker) applyRules(node SchemaNode) error {

	for _, rule := range sc.validationRules() {
		c := Conditions{
			ParentNodeName: node.ParentNodeName,
			ParentNodeType: node.ParentNodeType,
			NodeName:       node.N.Name,
			NodeType:       node.N.Type,
		}

		if rule.Applicable(c) && !rule.Validate(c) {
			return errors.New(rule.Msg)
		}
	}

	return nil
}

func (sc *SchemaChecker) validationRules() []Rule {
	return []Rule{
		{
			Msg:        "Root Node is not Schema",
			Criteria:   Conditions{NodeType: parser.RootNode},
			Validation: Conditions{NodeName: SchemaNodeName},
		},
		{
			Msg: "Default Events should be direct child of Schema or State node",
			Criteria: Conditions{NodeType: parser.ElementNode, CustomFn: func(c Conditions) bool {

				return isDefaultEventNode(c.NodeName)
			}},
			Validation: Conditions{CustomFn: func(c Conditions) bool {
				_, ok := sc.states[c.ParentNodeName]
				return c.ParentNodeType == parser.RootNode || ok
			}},
		},
		{
			Msg:      "Events node should be inside State node",
			Criteria: Conditions{NodeName: EventsNodeName},
			Validation: Conditions{ParentNodeType: parser.ElementNode, CustomFn: func(c Conditions) bool {
				_, ok := sc.states[c.ParentNodeName]
				return ok
			}},
		},
		// Extend the validation rules
	}
}

// ===============================================

type DefaultEvents struct {
	OnBeforeEvent Event
	OnAfterEvent  Event
	OnStateSet    Event
}

type Event struct {
	Tasks []string
}

func (e Event) Copy() Event {
	tasks := []string{}
	copy(tasks, e.Tasks)
	return Event{Tasks: tasks}
}

type CustomEvent struct {
	Name        string
	Tasks       []string
	TargetState string
	ErrorState  string
}

type Schema struct {
	DefaultEvents
	States []State
}

type State struct {
	DefaultEvents
	Name   string
	Events []CustomEvent
}

func New(p *parser.Parser) (*Schema, error) {

	ast := p.Parse()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("Parsing errors: \n%s ", strings.Join(p.Errors(), "\n"))
	}

	checker := SchemaChecker{root: *ast, states: make(map[string]bool), visitedNodes: make(map[string]int)}
	if err := checker.Validate(); err != nil {
		return nil, fmt.Errorf("Schema validation - %s", err.Error())
	}

	return buildFromAST(ast)
}

func buildFromAST(ast *parser.Node) (*Schema, error) {
	schema := Schema{}
	schema.DefaultEvents = buildDefaultEvents(ast)
	schema.States = buildStates(ast)
	return &schema, nil
}

func filterChildByName(ast *parser.Node, name string) *parser.Node {
	for _, child := range ast.Children {
		if child.Name == name {
			return &child
		}
	}

	return nil
}

func filterChildByNodeType(ast *parser.Node, nt string) *parser.Node {
	for _, child := range ast.Children {
		if child.Type == parser.NodeType(nt) {
			return &child
		}
	}

	return nil
}

func buildStates(ast *parser.Node) []State {
	states := make([]State, 0)

	if sts := filterChildByName(ast, "States"); sts != nil {
		for _, st := range sts.Children {
			states = append(states, buildState(&st))
		}
	}

	return states
}

func buildState(ast *parser.Node) State {
	state := State{}
	if events := filterChildByName(ast, "Events"); events != nil {
		state.Events = buildCustomEvents(events)
	}

	state.DefaultEvents = buildDefaultEvents(ast)
	state.Name = ast.Name

	return state
}

func buildDefaultEvents(ast *parser.Node) DefaultEvents {
	events := DefaultEvents{}
	for _, child := range ast.Children {
		switch child.Name {
		case OnBeforeEvent:
			events.OnBeforeEvent = Event{Tasks: buildTasks(&child)}
		case OnAfterEvent:
			events.OnAfterEvent = Event{Tasks: buildTasks(&child)}
		case OnStateSet:
			events.OnStateSet = Event{Tasks: buildTasks(&child)}
		}
	}
	return events
}

func buildCustomEvents(ast *parser.Node) []CustomEvent {

	events := make([]CustomEvent, 0)
	for _, child := range ast.Children {

		if isDefaultEventNode(child.Name) {
			continue
		}

		customEvt := CustomEvent{Name: child.Name, Tasks: buildTasks(&child)}
		for _, attr := range child.Attributes {
			switch attr.Name {
			case TargetState:
				customEvt.TargetState = attr.Value
			case ErrorState:
				customEvt.ErrorState = attr.Value
			}
		}

		events = append(events, customEvt)
	}
	return events
}

func buildTasks(ast *parser.Node) []string {
	tasks := make([]string, 0)
	for _, child := range ast.Children {
		if tn := filterChildByNodeType(&child, string(parser.TextNode)); tn != nil {
			tasks = append(tasks, tn.Name)
		}
	}
	return tasks
}
