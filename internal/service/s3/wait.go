package s3

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	bucketCreatedTimeout                          = 2 * time.Minute
	propagationTimeout                            = 1 * time.Minute
	lifecycleConfigurationRulesPropagationTimeout = 2 * time.Minute

	// LifecycleConfigurationRulesStatusReady occurs when all configured rules reach their desired state (Enabled or Disabled)
	LifecycleConfigurationRulesStatusReady = "READY"
	// LifecycleConfigurationRulesStatusNotReady occurs when all configured rules have not reached their desired state (Enabled or Disabled)
	LifecycleConfigurationRulesStatusNotReady = "NOT_READY"
)

func retryWhenBucketNotFound(f func() (interface{}, error)) (interface{}, error) {
	return tfresource.RetryWhenAWSErrCodeEquals(propagationTimeout, f, s3.ErrCodeNoSuchBucket)
}

func waitForLifecycleConfigurationRulesStatus(ctx context.Context, conn *s3.S3, bucket, expectedBucketOwner string, rules []*s3.LifecycleRule) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"", LifecycleConfigurationRulesStatusNotReady},
		Target:  []string{LifecycleConfigurationRulesStatusReady},
		Refresh: lifecycleConfigurationRulesStatus(ctx, conn, bucket, expectedBucketOwner, rules),
		Timeout: lifecycleConfigurationRulesPropagationTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}
