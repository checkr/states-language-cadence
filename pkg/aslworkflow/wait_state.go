package aslworkflow

import (
	"fmt"
	"time"

	"github.com/checkr/states-language-cadence/pkg/jsonpath"
	"github.com/coinbase/step/utils/to"
	"go.uber.org/cadence/workflow"
)

type WaitState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`

	Seconds     *float64       `json:",omitempty"`
	SecondsPath *jsonpath.Path `json:",omitempty"`

	Timestamp     *time.Time     `json:",omitempty"`
	TimestampPath *jsonpath.Path `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

func (s *WaitState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	var duration time.Duration

	if s.SecondsPath != nil {
		// Validate the path exists
		secs, err := s.SecondsPath.GetNumber(input)
		if err != nil {
			return nil, nil, err
		}
		duration = time.Duration(*secs) * time.Second

	} else if s.Seconds != nil {
		duration = time.Duration(*s.Seconds) * time.Second

	} else if s.TimestampPath != nil {
		// Validate the path exists
		ts, err := s.TimestampPath.GetTime(input)
		if err != nil {
			return nil, nil, err
		}
		now := workflow.Now(ctx).UTC()
		duration = ts.Sub(now)

	} else if s.Timestamp != nil {
		now := workflow.Now(ctx).UTC()
		duration = s.Timestamp.Sub(now)
	}

	err := workflow.Sleep(ctx, duration)
	if err != nil {
		return nil, nil, err
	}

	return input, nextState(s.Next, s.End), nil
}

func (s *WaitState) Execute(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(s,
		processInputOutput(
			s.InputPath,
			s.OutputPath,
			s.process,
		),
	)(ctx, input)
}

func (s *WaitState) Validate() error {
	s.SetType(to.Strp("Wait"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	// Next xor End
	if err := isEndValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	exactlyOne := []bool{
		s.Seconds != nil,
		s.SecondsPath != nil,
		s.Timestamp != nil,
		s.TimestampPath != nil,
	}

	count := 0
	for _, c := range exactlyOne {
		if c {
			count++
		}
	}

	if count != 1 {
		return fmt.Errorf("%v Exactly One (Seconds,SecondsPath,TimeStamp,TimeStampPath)", errorPrefix(s))
	}

	return nil
}

func (s *WaitState) SetType(t *string) {
	s.Type = t
}

func (s *WaitState) GetType() *string {
	return s.Type
}
