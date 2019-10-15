package aslworkflow

var succeedMachine = []byte(`
{
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Succeed"
		}
	}
}
`)

func (s *UnitTestSuite) Test_Workflow_Succeed_State() {
	sm, err := FromJSON(succeedMachine)
	if err != nil {
		s.NoError(err)
		return
	}

	exampleInput := map[string]interface{}{"test": "example_input"}

	RegisterWorkflow("TestSucceedWorkflow", *sm)

	s.env.ExecuteWorkflow("TestSucceedWorkflow", exampleInput)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result map[string]interface{}
	err = s.env.GetWorkflowResult(&result)
	s.NoError(err)

	s.Equal("example_input", result["test"])
}
