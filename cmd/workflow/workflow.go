package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/checkr/states-language-cadence/internal/pkg/common"
	"github.com/checkr/states-language-cadence/pkg/aslworkflow"
	"github.com/joho/godotenv"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

var ApplicationName = "WorkflowDemo"

type Example struct {
	name string
	json string
}

func getExamples() []Example {
	return []Example{
		{
			name: "example:workflow:ExampleWorkflow",
			json: `{
				"StartAt": "Example1",
				"States": {
					"Example1": {
						"Type": "Task",
						"Resource": "example:activity:ExampleActivity",
						"Next": "Example2"
					},
					"Example2": {
						"Type": "Task",
						"Resource": "example:workflow:ExampleSubworkflow",
						"End": true
					}
				}
			}`,
		},
		{
			name: "example:workflow:ExampleSubworkflow",
			json: `{
				"StartAt": "Example1",
				"States": {
					"Example1": {
						"Type": "Task",
						"Resource": "example:activity:ExampleActivity",
						"Next": "Wait"
					},
					"Wait": {
						"Type": "Wait",
						"Seconds": 3,
						"Next": "Result"
					},
					"Result": {
						"Type": "Pass",
						"Result": {
							"subworkflow": "example has been completed!"
						},
						"End": true
					}
				}
			}`,
		},
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	for _, example := range getExamples() {
		sm, err := aslworkflow.FromJSON([]byte(example.json))
		if err != nil {
			panic(fmt.Errorf("error loading state machine %w", err))
		}

		sm.RegisterWorkflow(example.name)
		sm.RegisterActivities(ExampleActivity)
	}

	// Register the Global Task Handler
	aslworkflow.RegisterHandler(ExampleTaskHandler)

	var h common.CadenceHelper
	h.SetupServiceConfig()

	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)

	select {}
}

// ActivityPrefix is the prefix used for specifying activities to run (could also be something like `arn:aws:...`
var ActivityPrefix = "example:activity:"

// WorkflowPrefix is the prefix used for specifying subworkflow to run (could also be something like `arn:aws:...`
var WorkflowPrefix = "example:workflow:"

var ErrUnknownResource = errors.New("unknown resource")

// ExampleTaskHandler is called for each task, it decides what to do. In this example it will execute and activity or subworkflow
func ExampleTaskHandler(ctx workflow.Context, resource string, input interface{}) (interface{}, error) {
	var result interface{}
	var err error

	if strings.HasPrefix(resource, ActivityPrefix) {
		err = workflow.ExecuteActivity(ctx, resource, input).Get(ctx, &result)
	} else if strings.HasPrefix(resource, WorkflowPrefix) {
		err = workflow.ExecuteChildWorkflow(ctx, resource, input).Get(ctx, &result)
	} else {
		return nil, ErrUnknownResource
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExampleActivity is the actual activity that is run by the handler, it could calla lambda function, http request, grpc, anything you'd like.
// Noop for now, just passing input as output
func ExampleActivity(ctx context.Context, input interface{}) (interface{}, error) {
	logger := activity.GetLogger(ctx)

	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.Info("activity executed", zap.Any("input", input), zap.Any("taskToken", taskToken), zap.Any("activityName", activityName))
	return input, nil
}
