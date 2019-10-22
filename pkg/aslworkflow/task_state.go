package aslworkflow

import (
	"errors"
	"fmt"

	"github.com/checkr/states-language-cadence/pkg/jsonpath"
	"github.com/coinbase/step/utils/to"
	"go.uber.org/cadence/workflow"
)

type TaskState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`
	Parameters interface{}    `json:",omitempty"`

	Resource *string `json:",omitempty"`

	Catch []*Catcher `json:",omitempty"`
	Retry []*Retrier `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`

	TimeoutSeconds   int `json:",omitempty"`
	HeartbeatSeconds int `json:",omitempty"`
}

var ErrTaskHandlerNotRegistered = errors.New("handler has not been registered")

func (s *TaskState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	if globalTaskHandler != nil {
		result, err := globalTaskHandler(ctx, *s.Resource, input)
		if err != nil {
			return nil, nil, err
		}
		return result, nextState(s.Next, s.End), nil
	}

	return nil, nil, ErrTaskHandlerNotRegistered
}

// Input must include the Task name in $.Task
func (s *TaskState) Execute(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(s,
		processCatcher(s.Catch,
			processRetrier(s.Name(), s.Retry,
				processInputOutput(
					s.InputPath,
					s.OutputPath,
					processParams(
						s.Parameters,
						processResult(s.ResultPath, s.process),
					),
				),
			),
		),
	)(ctx, input)
}

func (s *TaskState) Validate() error {
	s.SetType(to.Strp("Task"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %w", errorPrefix(s), err)
	}

	if err := isEndValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %w", errorPrefix(s), err)
	}

	if s.Resource == nil {
		return fmt.Errorf("%v Requires Resource", errorPrefix(s))
	}

	// TODO: implement custom handlers
	//if s.taskHandler != nil {
	//}

	if err := isCatchValid(s.Catch); err != nil {
		return err
	}

	if err := isRetryValid(s.Retry); err != nil {
		return err
	}

	return nil
}

func (s *TaskState) SetType(t *string) {
	s.Type = t
}

func (s *TaskState) GetType() *string {
	return s.Type
}

type TaskHandler func(ctx workflow.Context, resource string, input interface{}) (interface{}, error)

var globalTaskHandler TaskHandler

// RegisterHandler registers a global handler for tasks. The implementation is left up to importers of this package.
func RegisterHandler(taskHandlerFunc TaskHandler) {
	globalTaskHandler = taskHandlerFunc
}

func DeregisterHandler() {
	globalTaskHandler = nil
}
