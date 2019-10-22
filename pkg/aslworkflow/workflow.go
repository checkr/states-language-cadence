package aslworkflow

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
)

func Workflow(ctx workflow.Context, sm StateMachine, input interface{}) (interface{}, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	output, err := sm.Execute(ctx, input)
	return output, err
}

func RegisterWorkflow(workflowName string, initStateMachine StateMachine) {
	workflowFunc := func(ctx workflow.Context, input interface{}) (interface{}, error) {
		// Create a local instance of state machine. Workflows that were started with a given machine
		// will continue using that machine.
		encodedStateMachine := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
			return initStateMachine
		})

		var sm StateMachine
		err := encodedStateMachine.Get(&sm)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize the state machine: %w", err)
		}

		return Workflow(ctx, sm, input)
	}
	workflow.RegisterWithOptions(workflowFunc, workflow.RegisterOptions{Name: workflowName})
}
