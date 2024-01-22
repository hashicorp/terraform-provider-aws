// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	ssoadmin_sdkv2 "github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	ssoadmin_sdkv2_types "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*ssoadmin_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return ssoadmin_sdkv2.NewFromConfig(cfg, func(o *ssoadmin_sdkv2.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)
		}
		o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws_sdkv2.RetryerV2), retry_sdkv2.IsErrorRetryableFunc(func(err error) aws_sdkv2.Ternary {
			if errs.IsA[*ssoadmin_sdkv2_types.ConflictException](err) ||
				errs.IsA[*ssoadmin_sdkv2_types.ThrottlingException](err) {
				return aws_sdkv2.TrueTernary
			}
			return aws_sdkv2.UnknownTernary // Delegate to configured Retryer.
		}))
	}), nil
}
