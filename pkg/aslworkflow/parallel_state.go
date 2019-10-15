package aslworkflow

import (
	"fmt"

	"github.com/checkr/states-language-cadence/pkg/jsonpath"
	"github.com/coinbase/step/utils/to"
	"go.uber.org/cadence/workflow"
)

type ParallelState struct {
	stateStr

	Branches []Branch

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`
	Parameters interface{}    `json:",omitempty"`

	Catch []*Catcher `json:",omitempty"`
	Retry []*Retrier `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

type Branch struct {
	States  States
	StartAt string
}

func (s *ParallelState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	// You can use the context passed in to activity as a way to cancel the activity like standard GO way.
	// Cancelling a parent context will cancel all the derived contexts as well.
	// In the parallel block, we want to execute all of them in parallel and wait for all of them.
	// if one activity fails then we want to cancel all the rest of them as well.
	childCtx, cancelHandler := workflow.WithCancel(ctx)
	selector := workflow.NewSelector(ctx)
	var activityErr error

	var resp []interface{}

	for _, branch := range s.Branches {
		f := executeAsync(s, branch, childCtx, input)
		selector.AddFuture(f, func(f workflow.Future) {
			var r interface{}
			err := f.Get(ctx, &r)
			if err != nil {
				// cancel all pending activities
				cancelHandler()
				activityErr = err
			}
			resp = append(resp, r)
		})
	}

	for i := 0; i < len(s.Branches); i++ {
		selector.Select(ctx) // this will wait for one branch
		if activityErr != nil {
			return nil, nil, activityErr
		}
	}

	return interface{}(resp), nextState(s.Next, s.End), nil
}

func (s *ParallelState) Execute(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
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

func executeAsync(p *ParallelState, b Branch, ctx workflow.Context, input interface{}) workflow.Future {
	future, settable := workflow.NewFuture(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		output, _, err := b.Execute(ctx, *p, input)
		settable.Set(output, err)
	})
	return future
}

func (m *Branch) Execute(ctx workflow.Context, s ParallelState, input interface{}) (interface{}, *string, error) {
	nextState := &m.StartAt

	for {
		s := m.States[*nextState]
		if s == nil {
			return nil, nil, fmt.Errorf("next state invalid (%v)", *nextState)
		}

		output, next, err := s.Execute(ctx, input)

		if err != nil {
			return nil, nil, err
		}

		if next == nil || *next == "" {
			return output, nil, nil
		}

		nextState = next
		input = output
	}
}

func (m *Branch) Tasks() []*TaskState {
	var tasks []*TaskState
	tasks = append(tasks, tasksFromStates(m.States)...)
	return tasks
}

func (s *ParallelState) Validate() error {
	s.SetType(to.Strp("Parallel"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	return nil
}
