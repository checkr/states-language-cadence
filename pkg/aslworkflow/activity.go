package aslworkflow

import (
	"context"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

// Activity is the activity that Workflow runs
func Activity(ctx context.Context, resource string, input interface{}) (interface{}, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("executing activity", zap.Any("input", input), zap.Any("resource", resource))

	return input, nil
}
