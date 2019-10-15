package aslworkflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMachineGetTasks(t *testing.T) {
	sm, err := FromJSON([]byte(`
		{
			"StartAt": "Example1",
			"States": {
				"Example1": {
					"Type": "Task",
					"Resource": "arn:aws:resource:example",
					"Next": "Example2"
				},
				"Example2": {
					"Type": "Task",
					"Resource": "arn:aws:resource:example",
					"End": true
				}
			}
		}
	`))
	if err != nil {
		return
	}

	tasks := sm.Tasks()

	assert.Equal(t, 2, len(tasks))
	assert.Equal(t, "arn:aws:resource:example", *tasks[0].Resource)
}
