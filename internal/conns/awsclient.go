// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	apigatewayv2_types "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	baselogging "github.com/hashicorp/aws-sdk-go-base/v2/logging"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/dns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type AWSClient struct {
	accountID                 string
	awsConfig                 *aws.Config
	clients                   map[string]any
	defaultTagsConfig         *tftags.DefaultConfig
	endpoints                 map[string]string // From provider configuration.
	httpClient                *http.Client
	ignoreTagsConfig          *tftags.IgnoreConfig
	lock                      sync.Mutex
	logger                    baselogging.Logger
	partition                 endpoints.Partition
	region                    string
	servicePackages           map[string]ServicePackage
	session                   *session_sdkv1.Session
	s3ExpressClient           *s3.Client
	s3UsePathStyle            bool   // From provider configuration.
	s3USEast1RegionalEndpoint string // From provider configuration.
	stsRegion                 string // From provider configuration.
}

func (c *AWSClient) SetServicePackages(_ context.Context, servicePackages map[string]ServicePackage) {
	c.servicePackages = maps.Clone(servicePackages)
}

func (c *AWSClient) ServicePackage(_ context.Context, name string) ServicePackage {
	sp, ok := c.servicePackages[name]
	if !ok {
		return nil
	}
	return sp
}

func (c *AWSClient) ServicePackages(context.Context) iter.Seq2[string, ServicePackage] {
	return maps.All(c.servicePackages)
}

// CredentialsProvider returns the AWS SDK for Go v2 credentials provider.
func (c *AWSClient) CredentialsProvider(context.Context) aws.CredentialsProvider {
	if c.awsConfig == nil {
		return nil
	}
	return c.awsConfig.Credentials
}

func (c *AWSClient) DefaultTagsConfig(context.Context) *tftags.DefaultConfig {
	return c.defaultTagsConfig
}

func (c *AWSClient) IgnoreTagsConfig(context.Context) *tftags.IgnoreConfig {
	return c.ignoreTagsConfig
}

func (c *AWSClient) AwsConfig(context.Context) aws.Config { // nosemgrep:ci.aws-in-func-name
	return c.awsConfig.Copy()
}

// AwsSession and Endpoints can be removed once the simpledb service is removed.
func (c *AWSClient) AwsSession(context.Context) *session_sdkv1.Session { // nosemgrep:ci.aws-in-func-name
	return c.session
}

func (c *AWSClient) Endpoints(context.Context) map[string]string {
	return maps.Clone(c.endpoints)
}

// AccountID returns the configured AWS account ID.
func (c *AWSClient) AccountID(context.Context) string {
	return c.accountID
}

// Partition returns the ID of the configured AWS partition.
func (c *AWSClient) Partition(context.Context) string {
	return c.partition.ID()
}

// Region returns the ID of the configured AWS Region.
func (c *AWSClient) Region(context.Context) string {
	return c.region
}

// PartitionHostname returns a hostname with the provider domain suffix for the partition
// e.g. PREFIX.amazonaws.com
// The prefix should not contain a trailing period.
func (c *AWSClient) PartitionHostname(ctx context.Context, prefix string) string {
	return fmt.Sprintf("%s.%s", prefix, c.DNSSuffix(ctx))
}

// GlobalARN returns a global (no Region) ARN for the specified service namespace and resource.
func (c *AWSClient) GlobalARN(ctx context.Context, service, resource string) string {
	return c.GlobalARNWithAccount(ctx, service, c.AccountID(ctx), resource)
}

// GlobalARNNoAccount returns a global (no Region) ARN for the specified service namespace and resource without AWS account ID.
func (c *AWSClient) GlobalARNNoAccount(ctx context.Context, service, resource string) string {
	return c.GlobalARNWithAccount(ctx, service, "", resource)
}

// GlobalARNWithAccount returns a global (no Region) ARN for the specified service namespace, resource and account ID.
func (c *AWSClient) GlobalARNWithAccount(ctx context.Context, service, accountID, resource string) string {
	return arn.ARN{
		Partition: c.Partition(ctx),
		Service:   service,
		AccountID: accountID,
		Resource:  resource,
	}.String()
}

// RegionalARN returns a regional ARN for the specified service namespace and resource.
func (c *AWSClient) RegionalARN(ctx context.Context, service, resource string) string {
	return c.RegionalARNWithAccount(ctx, service, c.AccountID(ctx), resource)
}

// RegionalARNNoAccount returns a regional ARN for the specified service namespace and resource without AWS account ID.
func (c *AWSClient) RegionalARNNoAccount(ctx context.Context, service, resource string) string {
	return c.RegionalARNWithAccount(ctx, service, "", resource)
}

// RegionalARNWithAccount returns a regional ARN for the specified service namespace, resource and account ID.
func (c *AWSClient) RegionalARNWithAccount(ctx context.Context, service, accountID, resource string) string {
	return arn.ARN{
		Partition: c.Partition(ctx),
		Service:   service,
		Region:    c.Region(ctx),
		AccountID: accountID,
		Resource:  resource,
	}.String()
}

// RegionalHostname returns a hostname with the provider domain suffix for the region and partition
// e.g. PREFIX.us-west-2.amazonaws.com
// The prefix should not contain a trailing period.
func (c *AWSClient) RegionalHostname(ctx context.Context, prefix string) string {
	return fmt.Sprintf("%s.%s.%s", prefix, c.Region(ctx), c.DNSSuffix(ctx))
}

// S3ExpressClient returns an AWS SDK for Go v2 S3 API client suitable for use with S3 Express (directory buckets).
// This client differs from the standard S3 API client only in us-east-1 if the global S3 endpoint is used.
// In that case the returned client uses the regional S3 endpoint.
func (c *AWSClient) S3ExpressClient(ctx context.Context) *s3.Client {
	s3Client := c.S3Client(ctx)

	c.lock.Lock() // OK since a non-default client is created.
	defer c.lock.Unlock()

	if c.s3ExpressClient == nil {
		if s3Client.Options().Region == endpoints.AwsGlobalRegionID {
			// No global endpoint for S3 Express.
			c.s3ExpressClient = errs.Must(client[*s3.Client](ctx, c, names.S3, map[string]any{
				"s3_us_east_1_regional_endpoint": "regional",
			}))
		} else {
			c.s3ExpressClient = s3Client
		}
	}

	return c.s3ExpressClient
}

// S3UsePathStyle returns the s3_force_path_style provider configuration value.
func (c *AWSClient) S3UsePathStyle(context.Context) bool {
	return c.s3UsePathStyle
}

// SetHTTPClient sets the http.Client used for AWS API calls.
// To have effect it must be called before the AWS SDK v1 Session is created.
func (c *AWSClient) SetHTTPClient(_ context.Context, httpClient *http.Client) {
	if c.session == nil {
		c.httpClient = httpClient
	}
}

// HTTPClient returns the http.Client used for AWS API calls.
func (c *AWSClient) HTTPClient(context.Context) *http.Client {
	return c.httpClient
}

// RegisterLogger places the configured logger into Context so it can be used via `tflog`.
func (c *AWSClient) RegisterLogger(ctx context.Context) context.Context {
	return baselogging.RegisterLogger(ctx, c.logger)
}

// APIGatewayInvokeURL returns the Amazon API Gateway (REST APIs) invoke URL for the configured AWS Region.
// See https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-call-api.html.
func (c *AWSClient) APIGatewayInvokeURL(ctx context.Context, restAPIID, stageName string) string {
	return fmt.Sprintf("https://%s/%s", c.RegionalHostname(ctx, fmt.Sprintf("%s.execute-api", restAPIID)), stageName)
}

// APIGatewayV2InvokeURL returns the Amazon API Gateway v2 (WebSocket & HTTP APIs) invoke URL for the configured AWS Region.
// See https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-publish.html and
// https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-set-up-websocket-deployment.html.
func (c *AWSClient) APIGatewayV2InvokeURL(ctx context.Context, protocolType apigatewayv2_types.ProtocolType, apiID, stageName string) string {
	if protocolType == apigatewayv2_types.ProtocolTypeWebsocket {
		return fmt.Sprintf("wss://%s/%s", c.RegionalHostname(ctx, fmt.Sprintf("%s.execute-api", apiID)), stageName)
	}

	if stageName == "$default" {
		return fmt.Sprintf("https://%s/", c.RegionalHostname(ctx, fmt.Sprintf("%s.execute-api", apiID)))
	}

	return fmt.Sprintf("https://%s/%s", c.RegionalHostname(ctx, fmt.Sprintf("%s.execute-api", apiID)), stageName)
}

// CloudFrontDistributionHostedZoneID returns the Route 53 hosted zone ID
// for Amazon CloudFront distributions in the configured AWS partition.
func (c *AWSClient) CloudFrontDistributionHostedZoneID(ctx context.Context) string {
	if c.Partition(ctx) == endpoints.AwsCnPartitionID {
		return "Z3RFFRIM2A3IF5" // See https://docs.amazonaws.cn/en_us/aws/latest/userguide/route53.html
	}
	return "Z2FDTNDATAQYW2" // See https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html#Route53-Type-AliasTarget-HostedZoneId
}

// DefaultKMSKeyPolicy returns the default policy for KMS keys in the configured AWS partition.
func (c *AWSClient) DefaultKMSKeyPolicy(ctx context.Context) string {
	return fmt.Sprintf(`
{
	"Id": "default",
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "Enable IAM User Permissions",
			"Effect": "Allow",
			"Principal": {
				"AWS": "arn:%[1]s:iam::%[2]s:root"
			},
			"Action": "kms:*",
			"Resource": "*"
		}
	]
}	
`, c.Partition(ctx), c.AccountID(ctx))
}

// GlobalAcceleratorHostedZoneID returns the Route 53 hosted zone ID
// for AWS Global Accelerator accelerators in the configured AWS partition.
func (c *AWSClient) GlobalAcceleratorHostedZoneID(context.Context) string {
	return "Z2BJ6XQ5FK7U4H" // See https://docs.aws.amazon.com/general/latest/gr/global_accelerator.html#global_accelerator_region
}

// DNSSuffix returns the domain suffix for the configured AWS partition.
func (c *AWSClient) DNSSuffix(context.Context) string {
	dnsSuffix := c.partition.DNSSuffix()
	if dnsSuffix == "" {
		dnsSuffix = "amazonaws.com"
	}

	return dnsSuffix
}

// ReverseDNSPrefix returns the reverse DNS prefix for the configured AWS partition.
func (c *AWSClient) ReverseDNSPrefix(ctx context.Context) string {
	return dns.Reverse(c.DNSSuffix(ctx))
}

// EC2RegionalPrivateDNSSuffix returns the EC2 private DNS suffix for the configured AWS Region.
func (c *AWSClient) EC2RegionalPrivateDNSSuffix(ctx context.Context) string {
	region := c.Region(ctx)
	if region == endpoints.UsEast1RegionID {
		return "ec2.internal"
	}

	return fmt.Sprintf("%s.compute.internal", region)
}

// EC2RegionalPublicDNSSuffix returns the EC2 public DNS suffix for the configured AWS Region.
func (c *AWSClient) EC2RegionalPublicDNSSuffix(ctx context.Context) string {
	region := c.Region(ctx)
	if region == endpoints.UsEast1RegionID {
		return "compute-1"
	}

	return fmt.Sprintf("%s.compute", region)
}

// EC2PrivateDNSNameForIP returns a EC2 private DNS name in the configured AWS Region.
func (c *AWSClient) EC2PrivateDNSNameForIP(ctx context.Context, ip string) string {
	return fmt.Sprintf("ip-%s.%s", convertIPToDashIP(ip), c.EC2RegionalPrivateDNSSuffix(ctx))
}

// EC2PublicDNSNameForIP returns a EC2 public DNS name in the configured AWS Region.
func (c *AWSClient) EC2PublicDNSNameForIP(ctx context.Context, ip string) string {
	return c.PartitionHostname(ctx, fmt.Sprintf("ec2-%s.%s", convertIPToDashIP(ip), c.EC2RegionalPublicDNSSuffix(ctx)))
}

func convertIPToDashIP(ip string) string {
	return strings.Replace(ip, ".", "-", -1)
}

// apiClientConfig returns the AWS API client configuration parameters for the specified service.
func (c *AWSClient) apiClientConfig(ctx context.Context, servicePackageName string) map[string]any {
	m := map[string]any{
		"aws_sdkv2_config": c.awsConfig,
		"endpoint":         c.endpoints[servicePackageName],
		"partition":        c.Partition(ctx),
	}
	switch servicePackageName {
	case names.S3:
		m["s3_use_path_style"] = c.s3UsePathStyle
		// AWS SDK for Go v2 does not use the AWS_S3_US_EAST_1_REGIONAL_ENDPOINT environment variable during configuration.
		// For compatibility, read it now.
		if c.s3USEast1RegionalEndpoint == "" {
			c.s3USEast1RegionalEndpoint = NormalizeS3USEast1RegionalEndpoint(os.Getenv("AWS_S3_US_EAST_1_REGIONAL_ENDPOINT"))
		}
		m["s3_us_east_1_regional_endpoint"] = c.s3USEast1RegionalEndpoint
	case names.STS:
		m["sts_region"] = c.stsRegion
	}

	return m
}

// client returns the AWS SDK for Go v2 API client for the specified service.
// The default service client (`extra` is empty) is cached. In this case the AWSClient lock is held.
// This function is not a method on `AWSClient` as methods can't be parameterized (https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#no-parameterized-methods).
func client[T any](ctx context.Context, c *AWSClient, servicePackageName string, extra map[string]any) (T, error) {
	ctx = tflog.SetField(ctx, "tf_aws.service_package", servicePackageName)

	isDefault := len(extra) == 0
	// Default service client is cached.
	if isDefault {
		c.lock.Lock()
		defer c.lock.Unlock() // Runs at function exit, NOT block.

		if raw, ok := c.clients[servicePackageName]; ok {
			if client, ok := raw.(T); ok {
				return client, nil
			} else {
				var zero T
				return zero, fmt.Errorf("AWS SDK v2 API client (%s): %T, want %T", servicePackageName, raw, zero)
			}
		}
	}

	sp := c.ServicePackage(ctx, servicePackageName)
	if sp == nil {
		var zero T
		return zero, fmt.Errorf("unknown service package: %s", servicePackageName)
	}

	v, ok := sp.(interface {
		NewClient(context.Context, map[string]any) (T, error)
	})
	if !ok {
		var zero T
		return zero, fmt.Errorf("no AWS SDK v2 API client factory: %s", servicePackageName)
	}

	config := c.apiClientConfig(ctx, servicePackageName)
	maps.Copy(config, extra) // Extras overwrite per-service defaults.
	client, err := v.NewClient(ctx, config)
	if err != nil {
		var zero T
		return zero, err
	}

	// All customization for AWS SDK for Go v2 API clients must be done during construction.

	if isDefault {
		c.clients[servicePackageName] = client
	}

	return client, nil
}
