// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func (p *servicePackage) withExtraOptions(_ context.Context, config map[string]any) []func(*kinesis.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*kinesis.Options){
		func(o *kinesis.Options) {
			o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
				if errs.IsAErrorMessageContains[*types.LimitExceededException](err, "simultaneously be in CREATING or DELETING") ||
					errs.IsAErrorMessageContains[*types.LimitExceededException](err, "Rate exceeded for stream") {
					return aws.TrueTernary
				}
				return aws.UnknownTernary // Delegate to configured Retryer.
			}))
		},
	}
}
