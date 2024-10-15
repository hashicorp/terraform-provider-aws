// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*apigateway.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return apigateway.NewFromConfig(cfg,
		apigateway.WithEndpointResolverV2(newEndpointResolverSDKv2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *apigateway.Options) {
			o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
				// Many operations can return an error such as:
				//   ConflictException: Unable to complete operation due to concurrent modification. Please try again later.
				// Handle them all globally for the service client.
				if errs.IsAErrorMessageContains[*types.ConflictException](err, "try again later") {
					return aws.TrueTernary
				}
				return aws.UnknownTernary // Delegate to configured Retryer.
			}))
		},
	), nil
}
