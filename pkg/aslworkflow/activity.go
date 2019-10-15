package aslworkflow

import (
	"context"

	"go.uber.org/cadence/activity"
)

type Activity func(ctx context.Context, input interface{}) (interface{}, error)

func RegisterActivity(activityName string, activityFunc Activity) {
	activity.RegisterWithOptions(activityFunc, activity.RegisterOptions{Name: activityName})
}
