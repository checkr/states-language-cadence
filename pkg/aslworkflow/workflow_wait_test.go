package aslworkflow

import (
	"fmt"
	"time"

	"github.com/checkr/states-language-cadence/pkg/aslmachine"
)

var waitMachine = []byte(`
{
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Wait",
			"Seconds" : 60,
			"Next": "Example2"
		},
		"Example2": {
			"Type": "Wait",
			"SecondsPath" : "$.input_seconds",
			"Next": "Example3"
		},
		"Example3": {
			"Type": "Wait",
			"Timestamp" : "%s",
			"Next": "Example4"
		},
		"Example4": {
			"Type": "Wait",
			"TimestampPath" : "$.input_timestamp",
			"End": true
		}
	}
}
`)

func (s *UnitTestSuite) Test_Workflow_Wait_State() {
	workflowName := "TestWaitWorkflow"

	// Set the time to nearest minute. Timestamps only have seconds precision, so if you start the workflow not on a
	// minute interval the durations set in wait will chop off milliseconds. Probably fine in production, but makes
	// testing harder.
	s.env.SetStartTime(time.Now().UTC().Truncate(time.Minute))

	startTime := s.env.Now().UTC()

	inlineTime := startTime.Add(3 * time.Minute).Format(time.RFC3339)
	inputTime := startTime.Add(4 * time.Minute).Format(time.RFC3339)

	waitMachine = []byte(fmt.Sprintf(string(waitMachine), inlineTime))

	sm, err := aslmachine.FromJSON(waitMachine)
	if err != nil {
		s.NoError(err)
		return
	}

	exampleInput := map[string]interface{}{"input_seconds": 60, "input_timestamp": inputTime}

	// Expected timers
	timerCount := 0
	s.env.SetOnTimerScheduledListener(func(timerID string, d time.Duration) {
		secs := time.Minute
		diff := secs - d
		s.Equal(time.Duration(0), diff)

		timerCount++
	})

	RegisterWorkflow(workflowName, *sm)
	s.env.ExecuteWorkflow(workflowName, exampleInput)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	endTime := s.env.Now().UTC()

	// Check that the workflow took 4 minutes
	s.Equal(4*time.Minute, endTime.Sub(startTime))

	// Check that the timer ran 4 times
	s.Equal(4, timerCount)

	var result map[string]interface{}
	err = s.env.GetWorkflowResult(&result)
	s.NoError(err)

	s.Equal(float64(60), result["input_seconds"])
	s.Equal(inputTime, result["input_timestamp"])
}
