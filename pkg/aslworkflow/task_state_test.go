package aslworkflow

import (
	"errors"
	"strings"

	"go.uber.org/cadence/workflow"
)

var taskMachine = []byte(`
{
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Task",
			"Resource": "arn:aws:resource:example",
			"End": true
		}
	}
}
`)

func (s *UnitTestSuite) Test_Workflow_Task_State() {
	sm, err := FromJSON(taskMachine)
	if err != nil {
		s.NoError(err)
		return
	}

	handler := func(ctx workflow.Context, resource string, input interface{}) (interface{}, error) {
		s.Equal("arn:aws:resource:example", resource)
		output := map[string]interface{}{"test": "example_output"}
		return output, nil
	}
	RegisterHandler(handler)

	exampleInput := map[string]interface{}{"test": "example_input"}

	RegisterWorkflow("TestTaskWorkflow", *sm)

	s.env.ExecuteWorkflow("TestTaskWorkflow", exampleInput)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result map[string]interface{}
	err = s.env.GetWorkflowResult(&result)
	s.NoError(err)

	s.Equal("example_output", result["test"])
}

func (s *UnitTestSuite) Test_Workflow_Task_State_Error() {
	sm, err := FromJSON(taskMachine)
	if err != nil {
		s.NoError(err)
		return
	}

	handler := func(ctx workflow.Context, resource string, input interface{}) (interface{}, error) {
		s.Equal("arn:aws:resource:example", resource)
		return nil, errors.New("task error")
	}
	RegisterHandler(handler)

	exampleInput := map[string]interface{}{"test": "example_input"}

	RegisterWorkflow("TestTaskErrorWorkflow", *sm)

	s.env.ExecuteWorkflow("TestTaskWorkflow", exampleInput)

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_Workflow_Task_State_Missing() {
	sm, err := FromJSON(taskMachine)
	if err != nil {
		s.NoError(err)
		return
	}

	DeregisterHandler()

	exampleInput := map[string]interface{}{"test": "example_input"}

	RegisterWorkflow("TestTaskMissingWorkflow", *sm)

	s.env.ExecuteWorkflow("TestTaskWorkflow", exampleInput)

	s.True(s.env.IsWorkflowCompleted())

	err = s.env.GetWorkflowError()
	if s.Error(err) {
		s.True(strings.Contains(err.Error(), ErrTaskHandlerNotRegistered.Error()))
	}

}
