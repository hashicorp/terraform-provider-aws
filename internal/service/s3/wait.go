// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	lifecycleConfigurationExtraRetryDelay         = 5 * time.Second
	lifecycleConfigurationRulesPropagationTimeout = 3 * time.Minute
	lifecycleConfigurationRulesSteadyTimeout      = 2 * time.Minute

	// General timeout for S3 bucket changes to propagate.
	// See https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html#ConsistencyModel.
	s3BucketPropagationTimeout = 2 * time.Minute // nosemgrep:ci.s3-in-const-name, ci.s3-in-var-name

	// LifecycleConfigurationRulesStatusReady occurs when all configured rules reach their desired state (Enabled or Disabled)
	LifecycleConfigurationRulesStatusReady = "READY"
	// LifecycleConfigurationRulesStatusNotReady occurs when all configured rules have not reached their desired state (Enabled or Disabled)
	LifecycleConfigurationRulesStatusNotReady = "NOT_READY"
)

func waitForLifecycleConfigurationRulesStatus(ctx context.Context, conn *s3.S3, bucket, expectedBucketOwner string, rules []*s3.LifecycleRule) error {
	stateConf := &retry.StateChangeConf{
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
