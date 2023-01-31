package s3

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	bucketVersioningStableTimeout                 = 1 * time.Minute
	lifecycleConfigurationExtraRetryDelay         = 5 * time.Second
	lifecycleConfigurationRulesPropagationTimeout = 3 * time.Minute
	lifecycleConfigurationRulesSteadyTimeout      = 2 * time.Minute
	propagationTimeout                            = 1 * time.Minute

	// LifecycleConfigurationRulesStatusReady occurs when all configured rules reach their desired state (Enabled or Disabled)
	LifecycleConfigurationRulesStatusReady = "READY"
	// LifecycleConfigurationRulesStatusNotReady occurs when all configured rules have not reached their desired state (Enabled or Disabled)
	LifecycleConfigurationRulesStatusNotReady = "NOT_READY"
)

func retryWhenBucketNotFound(ctx context.Context, f func() (interface{}, error)) (interface{}, error) {
	return tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, f, s3.ErrCodeNoSuchBucket)
}

func waitForLifecycleConfigurationRulesStatus(ctx context.Context, conn *s3.S3, bucket, expectedBucketOwner string, rules []*s3.LifecycleRule) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{"", LifecycleConfigurationRulesStatusNotReady},
		Target:                    []string{LifecycleConfigurationRulesStatusReady},
		Refresh:                   lifecycleConfigurationRulesStatus(ctx, conn, bucket, expectedBucketOwner, rules),
		Timeout:                   lifecycleConfigurationRulesPropagationTimeout,
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 3,
		NotFoundChecks:            20,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitForBucketVersioningStatus(ctx context.Context, conn *s3.S3, bucket, expectedBucketOwner string) (*s3.GetBucketVersioningOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{""},
		Target:                    []string{s3.BucketVersioningStatusEnabled, s3.BucketVersioningStatusSuspended, BucketVersioningStatusDisabled},
		Refresh:                   bucketVersioningStatus(ctx, conn, bucket, expectedBucketOwner),
		Timeout:                   bucketVersioningStableTimeout,
		ContinuousTargetOccurence: 3,
		NotFoundChecks:            3,
		Delay:                     1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3.GetBucketVersioningOutput); ok {
		return output, err
	}

	return nil, err
}
