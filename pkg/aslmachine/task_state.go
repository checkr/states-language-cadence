package aslmachine

import (
	"fmt"

	"github.com/coinbase/step/jsonpath"
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

	// Maps a Lambda Handler Function
	TaskHandler interface{} `json:"-"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`

	TimeoutSeconds   int `json:",omitempty"`
	HeartbeatSeconds int `json:",omitempty"`
}

//func (s *TaskState) SetTaskHandler(resourcefn interface{}) {
//	s.TaskHandler = resourcefn
//}

func (s *TaskState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	var result interface{}
	err := workflow.ExecuteActivity(ctx, "ResourceActivity", *s.Resource, input).Get(ctx, &result)
	if err != nil {
		return nil, nil, err
	}

	return result, nextState(s.Next, s.End), nil
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
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	if err := isEndValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	if s.Resource == nil {
		return fmt.Errorf("%v Requires Resource", errorPrefix(s))
	}

	// TODO: implement custom handlers
	//if s.TaskHandler != nil {
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
