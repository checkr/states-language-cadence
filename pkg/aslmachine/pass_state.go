package aslmachine

import (
	"fmt"

	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/to"
	"go.uber.org/cadence/workflow"
)

type PassState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`
	Parameters interface{}    `json:",omitempty"` // TODO: Create a struct for Parameters?

	Result interface{} `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

func (s *PassState) Execute(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(s,
		processInputOutput(
			s.InputPath,
			s.OutputPath,
			processParams(
				s.Parameters,
				processResult(s.ResultPath, s.process),
			),
		),
	)(ctx, input)
}

func (s *PassState) process(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	output = input
	if s.Result != nil {
		output = s.Result
	}
	return output, nextState(s.Next, s.End), nil
}

func (s *PassState) Validate() error {
	s.SetType(to.Strp("Pass"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	// Next xor End
	if err := isEndValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	return nil
}

func (s *PassState) SetType(t *string) {
	s.Type = t
}

func (s *PassState) GetType() *string {
	return s.Type
}
