// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"strings"
	"time"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	lightsail_sdkv2 "github.com/aws/aws-sdk-go-v2/service/lightsail"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*lightsail_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return lightsail_sdkv2.NewFromConfig(cfg, func(o *lightsail_sdkv2.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.EndpointResolver = lightsail_sdkv2.EndpointResolverFromURL(endpoint)
		}

		retryable := retry_sdkv2.IsErrorRetryableFunc(func(e error) aws_sdkv2.Ternary {
			if strings.Contains(e.Error(), "Please try again in a few minutes") || strings.Contains(e.Error(), "Please wait for it to complete before trying again") {
				return aws_sdkv2.TrueTernary
			}
			return aws_sdkv2.UnknownTernary
		})
		const (
			backoff = 10 * time.Second
		)
		o.Retryer = retry_sdkv2.NewStandard(func(o *retry_sdkv2.StandardOptions) {
			o.Retryables = append(o.Retryables, retryable)
			o.MaxAttempts = 18
			o.Backoff = retry_sdkv2.NewExponentialJitterBackoff(backoff)
		})
	}), nil
}
