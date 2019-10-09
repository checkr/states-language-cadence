package aslmachine

import (
	"encoding/json"
	"fmt"

	"go.uber.org/cadence/workflow"
)

type StateMachine struct {
	States         States
	StartAt        string
	Comment        string
	Version        string
	TimeoutSeconds int32
}

// States is the collection of states
type States map[string]State

func FromJSON(raw []byte) (*StateMachine, error) {
	var sm StateMachine
	err := json.Unmarshal(raw, &sm)
	return &sm, err
}

func (sm *States) UnmarshalJSON(b []byte) error {
	// States
	var rawStates map[string]*json.RawMessage
	err := json.Unmarshal(b, &rawStates)

	if err != nil {
		return err
	}

	newStates := States{}
	for name, raw := range rawStates {
		states, err := unmarshallState(name, raw)
		if err != nil {
			return err
		}

		for _, s := range states {
			newStates[*s.Name()] = s
		}
	}

	*sm = newStates
	return nil
}

// Default State Methods

func (s *stateStr) GetType() *string {
	return s.Type
}

func (s *stateStr) SetType(t *string) {
	s.Type = t
}

type stateType struct {
	Type string
}

func unmarshallState(name string, rawJSON *json.RawMessage) ([]State, error) {
	var err error

	// extract type (safer than regex)
	var stateType stateType
	if err = json.Unmarshal(*rawJSON, &stateType); err != nil {
		return nil, err
	}

	var newState State

	switch stateType.Type {
	case "Pass":
		var s PassState
		err = json.Unmarshal(*rawJSON, &s)
		newState = &s
	case "Task":
		var s TaskState
		err = json.Unmarshal(*rawJSON, &s)
		newState = &s
	case "Choice":
		var s ChoiceState
		err = json.Unmarshal(*rawJSON, &s)
		newState = &s
	case "Wait":
		var s WaitState
		err = json.Unmarshal(*rawJSON, &s)
		newState = &s
	case "Succeed":
		var s SucceedState
		err = json.Unmarshal(*rawJSON, &s)
		newState = &s
	case "Fail":
		var s FailState
		err = json.Unmarshal(*rawJSON, &s)
		newState = &s
	case "Parallel":
		var s ParallelState
		err = json.Unmarshal(*rawJSON, &s)
		newState = &s
	default:
		err = fmt.Errorf("unknown state %q", stateType.Type)
	}

	// End of loop return error
	if err != nil {
		return nil, err
	}

	// Set Name and Defaults
	newName := name
	newState.SetName(&newName) // Require New Variable Pointer

	return []State{newState}, nil
}

func (m *StateMachine) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	nextState := &m.StartAt

	for {
		output, next, err := m.States[*nextState].Execute(ctx, input)

		if err != nil {
			return nil, err
		}

		if next == nil {
			return output, nil
		}

		nextState = next
		input = output
	}
}
