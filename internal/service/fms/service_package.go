// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/fms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fms/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func (p *servicePackage) withExtraOptions(_ context.Context, config map[string]any) []func(*fms.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*fms.Options){
		func(o *fms.Options) {
			o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
				// Acceptance testing creates and deletes resources in quick succession.
				// The FMS onboarding process into Organizations is opaque to consumers.
				// Since we cannot reasonably check this status before receiving the error,
				// set the operation as retryable.
				if errs.IsAErrorMessageContains[*awstypes.InvalidOperationException](err, "Your AWS Organization is currently onboarding with AWS Firewall Manager and cannot be offboarded") ||
					errs.IsAErrorMessageContains[*awstypes.InvalidOperationException](err, "Your AWS Organization is currently offboarding with AWS Firewall Manager. Please submit onboard request after offboarded") {
					return aws.TrueTernary
				}
				return aws.UnknownTernary // Delegate to configured Retryer.
			}))
		},
	}
}
