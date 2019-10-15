package aslworkflow

import (
	"fmt"

	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
)

type FailState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	Error *string `json:",omitempty"`
	Cause *string `json:",omitempty"`
}

func (s *FailState) Execute(_ workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	return nil, nil, cadence.NewCustomError(
		"Fail",
		errorOutput(s.Error, s.Cause))
}

func (s *FailState) Validate() error {
	s.SetType(to.Strp("Fail"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	if is.EmptyStr(s.Error) {
		return fmt.Errorf("%v %v", errorPrefix(s), "must contain Error")
	}

	return nil
}

func (s *FailState) SetType(t *string) {
	s.Type = t
}

func (s *FailState) GetType() *string {
	return s.Type
}
