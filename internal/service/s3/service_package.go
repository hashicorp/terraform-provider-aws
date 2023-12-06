// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	s3_sdkv2 "github.com/aws/aws-sdk-go-v2/service/s3"
	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	s3_sdkv1 "github.com/aws/aws-sdk-go/service/s3"
	tfawserr_sdkv1 "github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, m map[string]any) (*s3_sdkv1.S3, error) {
	sess := m["session"].(*session_sdkv1.Session)
	config := &aws_sdkv1.Config{
		Endpoint:         aws_sdkv1.String(m["endpoint"].(string)),
		S3ForcePathStyle: aws_sdkv1.Bool(m["s3_use_path_style"].(bool)),
	}

	if v, ok := m["s3_us_east_1_regional_endpoint"]; ok {
		config.S3UsEast1RegionalEndpoint = v.(endpoints_sdkv1.S3UsEast1RegionalEndpoint)
	}

	return s3_sdkv1.New(sess.Copy(config)), nil
}

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *s3_sdkv1.S3) (*s3_sdkv1.S3, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if tfawserr_sdkv1.ErrMessageContains(r.Error, errCodeOperationAborted, "A conflicting conditional operation is currently in progress against this resource. Please try again.") {
			r.Retryable = aws_sdkv1.Bool(true)
		}
	})

	return conn, nil
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*s3_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return s3_sdkv2.NewFromConfig(cfg, func(o *s3_sdkv2.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)
		} else if o.Region == names.USEast1RegionID && config["s3_us_east_1_regional_endpoint"].(endpoints_sdkv1.S3UsEast1RegionalEndpoint) != endpoints_sdkv1.RegionalS3UsEast1Endpoint {
			// Maintain the AWS SDK for Go v1 default of using the global endpoint in us-east-1.
			// See https://github.com/hashicorp/terraform-provider-aws/issues/33028.
			o.Region = names.GlobalRegionID
		}
		o.UsePathStyle = config["s3_use_path_style"].(bool)

		o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws_sdkv2.RetryerV2), retry_sdkv2.IsErrorRetryableFunc(func(err error) aws_sdkv2.Ternary {
			if tfawserr_sdkv2.ErrMessageContains(err, errCodeOperationAborted, "A conflicting conditional operation is currently in progress against this resource. Please try again.") {
				return aws_sdkv2.TrueTernary
			}
			return aws_sdkv2.UnknownTernary // Delegate to configured Retryer.
		}))
	}), nil
}

// Functional options to force the regional endpoint in us-east-1 if the client is configured to use the global endpoint.
func useRegionalEndpointInUSEast1(o *s3_sdkv2.Options) {
	if o.Region == names.GlobalRegionID {
		o.Region = names.USEast1RegionID
	}
}
