// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"strings"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	lightsail_sdkv2 "github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*lightsail_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return lightsail_sdkv2.NewFromConfig(cfg, func(o *lightsail_sdkv2.Options) {
		if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)
		}

		o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws_sdkv2.RetryerV2), retry_sdkv2.IsErrorRetryableFunc(func(err error) aws_sdkv2.Ternary {
			if errs.IsAErrorMessageContains[*types.InvalidInputException](err, "Please try again in a few minutes") ||
				strings.Contains(err.Error(), "Please wait for it to complete before trying again") {
				return aws_sdkv2.TrueTernary
			}
			return aws_sdkv2.UnknownTernary
		}))
	}), nil
}
