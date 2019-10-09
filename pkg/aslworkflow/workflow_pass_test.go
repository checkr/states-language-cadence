package aslworkflow

import "github.com/checkr/states-language-cadence/pkg/aslmachine"

var machine = []byte(`
{
  "StartAt": "Example1",
  "States": {
    "Example1": {
      "Type": "Pass",
      "Next": "Example2"
    },
    "Example2": {
      "Type": "Pass",
      "Result": {
        "test": "example_output"
      },
      "End": true
    }
  }
}
`)

func (s *UnitTestSuite) Test_Workflow_Pass_State() {
	sm, err := aslmachine.FromJSON(machine)
	if err != nil {
		s.NoError(err)
		return
	}

	exampleInput := map[string]interface{}{"test": "example_input"}

	RegisterWorkflow("TestWorkflow", *sm)

	s.env.ExecuteWorkflow("TestWorkflow", exampleInput)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result map[string]interface{}
	err = s.env.GetWorkflowResult(&result)
	s.NoError(err)

	s.Equal("example_output", result["test"])
}
