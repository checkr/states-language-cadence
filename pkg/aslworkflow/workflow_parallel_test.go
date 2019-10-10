package aslworkflow

import (
	"github.com/checkr/states-language-cadence/pkg/aslmachine"
)

var parallelMachine = []byte(`
{
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Parallel",
			"End": true,
			"Branches": [
				{
					"StartAt": "Branch1",
					"States": {
						"Branch1": {
							"Type": "Pass",
							"Result": {
								"branch1": true
							},
							"End": true
						}
					}
				},
				{
					"StartAt": "Branch2",
					"States": {
						"Branch2": {
							"Type": "Pass",
							"Result": {
								"branch2": true
							},
							"End": true
						}
					}
				}
			  ]
		}
	}
}
`)

func (s *UnitTestSuite) Test_Workflow_Parallel_State() {
	workflowName := "TestParallelWorkflow"

	sm, err := aslmachine.FromJSON(parallelMachine)
	if err != nil {
		s.NoError(err)
		return
	}

	exampleInput := map[string]interface{}{"example": "example"}

	RegisterWorkflow(workflowName, *sm)
	s.env.ExecuteWorkflow(workflowName, exampleInput)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result []map[string]bool
	err = s.env.GetWorkflowResult(&result)
	s.NoError(err)

	s.True(result[0]["branch1"])
	s.True(result[1]["branch2"])
}
