// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*cloudformation.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return cloudformation.NewFromConfig(cfg, func(o *cloudformation.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}

		o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
			if errs.IsAErrorMessageContains[*types.OperationInProgressException](err, "Another Operation on StackSet") {
				return aws.TrueTernary
			}
			return aws.UnknownTernary // Delegate to configured Retryer.
		}))
	}), nil
}
