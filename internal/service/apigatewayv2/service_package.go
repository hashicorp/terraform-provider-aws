// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func (p *servicePackage) withExtraOptions(_ context.Context, config map[string]any) []func(*apigatewayv2.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*apigatewayv2.Options){
		func(o *apigatewayv2.Options) {
			o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
				if errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "try again later") {
					return aws.TrueTernary
				}
				// In some instances, ConflictException error responses have been observed as
				// a *smithy.OperationError type (not an *awstypes.ConflictException), which
				// can't be handled via errs.IsAErrorMessageContains. Instead we fall back
				// to a simple match on the message contents.
				if errs.Contains(err, "Unable to complete operation due to concurrent modification. Please try again later.") {
					return aws.TrueTernary
				}
				return aws.UnknownTernary // Delegate to configured Retryer.
			}))
		},
	}
}
