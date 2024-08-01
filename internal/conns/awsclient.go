// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"os"
	"strings"
	"sync"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	config_sdkv2 "github.com/aws/aws-sdk-go-v2/config"
	apigatewayv2_types "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	s3_sdkv2 "github.com/aws/aws-sdk-go-v2/service/s3"
	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	opsworks_sdkv1 "github.com/aws/aws-sdk-go/service/opsworks"
	rds_sdkv1 "github.com/aws/aws-sdk-go/service/rds"
	baselogging "github.com/hashicorp/aws-sdk-go-base/v2/logging"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type AWSClient struct {
	AccountID         string
	DefaultTagsConfig *tftags.DefaultConfig
	IgnoreTagsConfig  *tftags.IgnoreConfig
	Partition         string
	Region            string
	ServicePackages   map[string]ServicePackage

	awsConfig                 *aws_sdkv2.Config
	clients                   map[string]any
	conns                     map[string]any
	dnsSuffix                 string
	endpoints                 map[string]string // From provider configuration.
	httpClient                *http.Client
	lock                      sync.Mutex
	logger                    baselogging.Logger
	session                   *session_sdkv1.Session
	s3ExpressClient           *s3_sdkv2.Client
	s3UsePathStyle            bool   // From provider configuration.
	s3USEast1RegionalEndpoint string // From provider configuration.
	stsRegion                 string // From provider configuration.
}

// CredentialsProvider returns the AWS SDK for Go v2 credentials provider.
func (c *AWSClient) CredentialsProvider(context.Context) aws_sdkv2.CredentialsProvider {
	if c.awsConfig == nil {
		return nil
	}
	return c.awsConfig.Credentials
}

func (c *AWSClient) AwsConfig(context.Context) aws_sdkv2.Config { // nosemgrep:ci.aws-in-func-name
	return c.awsConfig.Copy()
}

// OpsWorksConnForRegion returns an AWS SDK For Go v1 OpsWorks API client for the specified AWS Region.
// If the specified region is not the default a new "simple" client is created.
// This new client does not use any configured endpoint override.
func (c *AWSClient) OpsWorksConnForRegion(ctx context.Context, region string) *opsworks_sdkv1.OpsWorks {
	if region == c.Region {
		return c.OpsWorksConn(ctx)
	}
	return opsworks_sdkv1.New(c.session, aws_sdkv1.NewConfig().WithRegion(region))
}

// PartitionHostname returns a hostname with the provider domain suffix for the partition
// e.g. PREFIX.amazonaws.com
// The prefix should not contain a trailing period.
func (c *AWSClient) PartitionHostname(ctx context.Context, prefix string) string {
	return fmt.Sprintf("%s.%s", prefix, c.DNSSuffix(ctx))
}

// RegionalHostname returns a hostname with the provider domain suffix for the region and partition
// e.g. PREFIX.us-west-2.amazonaws.com
// The prefix should not contain a trailing period.
func (c *AWSClient) RegionalHostname(ctx context.Context, prefix string) string {
	return fmt.Sprintf("%s.%s.%s", prefix, c.Region, c.DNSSuffix(ctx))
}

// RDSConnForRegion returns an AWS SDK For Go v1 RDS API client for the specified AWS Region.
// If the specified region is not the default a new "simple" client is created.
// This new client does not use any configured endpoint override.
func (c *AWSClient) RDSConnForRegion(ctx context.Context, region string) *rds_sdkv1.RDS {
	if region == c.Region {
		return c.RDSConn(ctx)
	}
	return rds_sdkv1.New(c.session, aws_sdkv1.NewConfig().WithRegion(region))
}

// S3ExpressClient returns an AWS SDK for Go v2 S3 API client suitable for use with S3 Express (directory buckets).
// This client differs from the standard S3 API client only in us-east-1 if the global S3 endpoint is used.
// In that case the returned client uses the regional S3 endpoint.
func (c *AWSClient) S3ExpressClient(ctx context.Context) *s3_sdkv2.Client {
	s3Client := c.S3Client(ctx)

	c.lock.Lock() // OK since a non-default client is created.
	defer c.lock.Unlock()

	if c.s3ExpressClient == nil {
		if s3Client.Options().Region == names.GlobalRegionID {
			c.s3ExpressClient = errs.Must(client[*s3_sdkv2.Client](ctx, c, names.S3, map[string]any{
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
func (c *AWSClient) CloudFrontDistributionHostedZoneID(context.Context) string {
	if c.Partition == names.ChinaPartitionID {
		return "Z3RFFRIM2A3IF5" // See https://docs.amazonaws.cn/en_us/aws/latest/userguide/route53.html
	}
	return "Z2FDTNDATAQYW2" // See https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html#Route53-Type-AliasTarget-HostedZoneId
}

// DefaultKMSKeyPolicy returns the default policy for KMS keys in the configured AWS partition.
func (c *AWSClient) DefaultKMSKeyPolicy(context.Context) string {
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
`, c.Partition, c.AccountID)
}

// GlobalAcceleratorHostedZoneID returns the Route 53 hosted zone ID
// for AWS Global Accelerator accelerators in the configured AWS partition.
func (c *AWSClient) GlobalAcceleratorHostedZoneID(context.Context) string {
	return "Z2BJ6XQ5FK7U4H" // See https://docs.aws.amazon.com/general/latest/gr/global_accelerator.html#global_accelerator_region
}

// DNSSuffix returns the domain suffix for the configured AWS partition.
func (c *AWSClient) DNSSuffix(context.Context) string {
	return c.dnsSuffix
}

// ReverseDNSPrefix returns the reverse DNS prefix for the configured AWS partition.
func (c *AWSClient) ReverseDNSPrefix(ctx context.Context) string {
	return names.ReverseDNS(c.DNSSuffix(ctx))
}

// EC2RegionalPrivateDNSSuffix returns the EC2 private DNS suffix for the configured AWS Region.
func (c *AWSClient) EC2RegionalPrivateDNSSuffix(context.Context) string {
	region := c.Region
	if region == names.USEast1RegionID {
		return "ec2.internal"
	}

	return fmt.Sprintf("%s.compute.internal", region)
}

// EC2RegionalPublicDNSSuffix returns the EC2 public DNS suffix for the configured AWS Region.
func (c *AWSClient) EC2RegionalPublicDNSSuffix(context.Context) string {
	region := c.Region
	if region == names.USEast1RegionID {
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
		"endpoint":         c.resolveEndpoint(ctx, servicePackageName),
		"partition":        c.Partition,
		"session":          c.session,
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

func (c *AWSClient) resolveEndpoint(ctx context.Context, servicePackageName string) string {
	endpoint := c.endpoints[servicePackageName]
	if endpoint != "" {
		return endpoint
	}

	// Only continue if there is an SDK v1 package. SDK v2 supports envvars and config file
	if names.ClientSDKV1(servicePackageName) {
		endpoint = aws_sdkv2.ToString(c.awsConfig.BaseEndpoint)

		envvar := names.AWSServiceEnvVar(servicePackageName)
		svc := os.Getenv(envvar)
		if svc != "" {
			return svc
		}

		if base := os.Getenv("AWS_ENDPOINT_URL"); base != "" {
			return base
		}

		sdkId := names.SDKID(servicePackageName)
		endpoint, found, err := resolveServiceBaseEndpoint(ctx, sdkId, c.awsConfig.ConfigSources)
		if found && err == nil {
			return endpoint
		}
	}
	return endpoint
}

// serviceBaseEndpointProvider is needed to search for all providers
// that provide a configured service endpoint
type serviceBaseEndpointProvider interface {
	GetServiceBaseEndpoint(ctx context.Context, sdkID string) (string, bool, error)
}

// resolveServiceBaseEndpoint is used to retrieve service endpoints from configured sources
// while allowing for configured endpoints to be disabled
func resolveServiceBaseEndpoint(ctx context.Context, sdkID string, configs []any) (value string, found bool, err error) {
	if val, found, _ := config_sdkv2.GetIgnoreConfiguredEndpoints(ctx, configs); found && val {
		return "", false, nil
	}

	for _, cs := range configs {
		if p, ok := cs.(serviceBaseEndpointProvider); ok {
			value, found, err = p.GetServiceBaseEndpoint(ctx, sdkID)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// conn returns the AWS SDK for Go v1 API client for the specified service.
// The default service client (`extra` is empty) is cached. In this case the AWSClient lock is held.
// This function is not a method on `AWSClient` as methods can't be parameterized (https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#no-parameterized-methods).
func conn[T any](ctx context.Context, c *AWSClient, servicePackageName string, extra map[string]any) (T, error) {
	ctx = tflog.SetField(ctx, "tf_aws.service_package", servicePackageName)

	isDefault := len(extra) == 0
	// Default service client is cached.
	if isDefault {
		c.lock.Lock()
		defer c.lock.Unlock() // Runs at function exit, NOT block.

		if raw, ok := c.conns[servicePackageName]; ok {
			if conn, ok := raw.(T); ok {
				return conn, nil
			} else {
				var zero T
				return zero, fmt.Errorf("AWS SDK v1 API client (%s): %T, want %T", servicePackageName, raw, zero)
			}
		}
	}

	sp, ok := c.ServicePackages[servicePackageName]
	if !ok {
		var zero T
		return zero, fmt.Errorf("unknown service package: %s", servicePackageName)
	}

	v, ok := sp.(interface {
		NewConn(context.Context, map[string]any) (T, error)
	})
	if !ok {
		var zero T
		return zero, fmt.Errorf("no AWS SDK v1 API client factory: %s", servicePackageName)
	}

	config := c.apiClientConfig(ctx, servicePackageName)
	maps.Copy(config, extra) // Extras overwrite per-service defaults.
	conn, err := v.NewConn(ctx, config)
	if err != nil {
		var zero T
		return zero, err
	}

	if v, ok := sp.(interface {
		CustomizeConn(context.Context, T) (T, error)
	}); ok {
		conn, err = v.CustomizeConn(ctx, conn)
		if err != nil {
			var zero T
			return zero, err
		}
	}

	// Default service client is cached.
	if isDefault {
		c.conns[servicePackageName] = conn
	}

	return conn, nil
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

	sp, ok := c.ServicePackages[servicePackageName]
	if !ok {
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
