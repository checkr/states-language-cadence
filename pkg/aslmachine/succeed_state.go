package aslmachine

import (
	"fmt"

	"github.com/checkr/states-language-cadence/pkg/jsonpath"
	"github.com/coinbase/step/utils/to"
	"go.uber.org/cadence/workflow"
)

type SucceedState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
}

func (s *SucceedState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	return input, nil, nil
}

func (s *SucceedState) Execute(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(s,
		processInputOutput(
			s.InputPath,
			s.OutputPath,
			s.process,
		),
	)(ctx, input)
}

func (s *SucceedState) Validate() error {
	s.SetType(to.Strp("Succeed"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	return nil
}

func (s *SucceedState) SetType(t *string) {
	s.Type = t
}

func (s *SucceedState) GetType() *string {
	return s.Type
}
