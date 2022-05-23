package kendra

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/kendra"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	kendraIndexCreatedTimeout = 60 * time.Minute
	kendraIndexUpdatedTimeout = 60 * time.Minute
	kendraIndexDeletedTimeout = 60 * time.Minute
)

func waitIndexCreated(ctx context.Context, conn *kendra.Kendra, timeout time.Duration, Id string) (*kendra.DescribeIndexOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kendra.IndexStatusCreating},
		Target:  []string{kendra.IndexStatusActive, kendra.IndexStatusFailed},
		Refresh: statusIndex(ctx, conn, Id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kendra.DescribeIndexOutput); ok {
		return v, err
	}

	return nil, err
}

func waitIndexUpdated(ctx context.Context, conn *kendra.Kendra, timeout time.Duration, Id string) (*kendra.DescribeIndexOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kendra.IndexStatusUpdating},
		Target:  []string{kendra.IndexStatusActive, kendra.IndexStatusFailed},
		Refresh: statusIndex(ctx, conn, Id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kendra.DescribeIndexOutput); ok {
		return v, err
	}

	return nil, err
}

func waitIndexDeleted(ctx context.Context, conn *kendra.Kendra, timeout time.Duration, Id string) (*kendra.DescribeIndexOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kendra.IndexStatusDeleting},
		Target:  []string{kendra.ErrCodeResourceNotFoundException},
		Refresh: statusIndex(ctx, conn, Id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kendra.DescribeIndexOutput); ok {
		return v, err
	}

	return nil, err
}
