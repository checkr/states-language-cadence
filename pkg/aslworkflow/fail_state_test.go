package aslworkflow

import (
	"go.uber.org/cadence"
)

var failMachine = []byte(`
{
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Fail",
			"Error": "ExampleError",
			"Cause": "This is an example error",
			"End": true
		}
	}
}
`)

func (s *UnitTestSuite) Test_Workflow_Fail_State() {
	sm, err := FromJSON(failMachine)
	if err != nil {
		s.NoError(err)
		return
	}

	exampleInput := map[string]interface{}{"test": "example_input"}

	RegisterWorkflow("TestFailWorkflow", *sm)

	s.env.ExecuteWorkflow("TestFailWorkflow", exampleInput)
	s.True(s.env.IsWorkflowCompleted())
	err = s.env.GetWorkflowError()
	s.Error(err)

	var details map[string]interface{}

	switch err := err.(type) {
	case *cadence.CustomError:
		detailsErr := err.Details(&details)

		if detailsErr != nil {
			s.NoError(detailsErr)
		}
	}

	s.Equal("ExampleError", details["Error"])
	s.Equal("This is an example error", details["Cause"])
}
