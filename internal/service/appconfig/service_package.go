// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"errors"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	appconfig_sdkv2 "github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/aws/smithy-go"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*appconfig_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return appconfig_sdkv2.NewFromConfig(cfg, func(o *appconfig_sdkv2.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)
		}

		// StartDeployment operations can return a ConflictException
		// if ongoing deployments are in-progress, thus we handle them
		// here for the service client.
		o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws_sdkv2.RetryerV2), retry_sdkv2.IsErrorRetryableFunc(func(err error) aws_sdkv2.Ternary {
			if err != nil {
				var oe *smithy.OperationError
				if errors.As(err, &oe) {
					if oe.OperationName == "StartDeployment" {
						if errs.IsA[*awstypes.ConflictException](err) {
							return aws_sdkv2.TrueTernary
						}
					}
				}
			}
			return aws_sdkv2.UnknownTernary
		}))
	}), nil
}
