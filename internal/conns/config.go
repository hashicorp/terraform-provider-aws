// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"
	"fmt"
	"strings"
	"time"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	imds_sdkv2 "github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	awsbasev1 "github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2"
	basediag "github.com/hashicorp/aws-sdk-go-base/v2/diag"
	"github.com/hashicorp/aws-sdk-go-base/v2/logging"
	basevalidation "github.com/hashicorp/aws-sdk-go-base/v2/validation"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/hashicorp/terraform-provider-aws/version"
)

type Config struct {
	AccessKey                      string
	AllowedAccountIds              []string
	AssumeRole                     *awsbase.AssumeRole
	AssumeRoleWithWebIdentity      *awsbase.AssumeRoleWithWebIdentity
	CustomCABundle                 string
	DefaultTagsConfig              *tftags.DefaultConfig
	EC2MetadataServiceEnableState  imds_sdkv2.ClientEnableState
	EC2MetadataServiceEndpoint     string
	EC2MetadataServiceEndpointMode string
	Endpoints                      map[string]string
	ForbiddenAccountIds            []string
	HTTPProxy                      *string
	HTTPSProxy                     *string
	IgnoreTagsConfig               *tftags.IgnoreConfig
	Insecure                       bool
	MaxRetries                     int
	NoProxy                        string
	Profile                        string
	Region                         string
	RetryMode                      aws_sdkv2.RetryMode
	S3UsePathStyle                 bool
	S3USEast1RegionalEndpoint      string
	SecretKey                      string
	SharedConfigFiles              []string
	SharedCredentialsFiles         []string
	SkipCredsValidation            bool
	SkipRegionValidation           bool
	SkipRequestingAccountId        bool
	STSRegion                      string
	SuppressDebugLog               bool
	TerraformVersion               string
	Token                          string
	TokenBucketRateLimiterCapacity int
	UseDualStackEndpoint           bool
	UseFIPSEndpoint                bool
}

// ConfigureProvider configures the provided provider Meta (instance data).
func (c *Config) ConfigureProvider(ctx context.Context, client *AWSClient) (*AWSClient, diag.Diagnostics) {
	var diags diag.Diagnostics

	ctx, logger := logging.NewTfLogger(ctx)

	const (
		maxBackoff = 300 * time.Second // AWS SDK for Go v1 DefaultRetryerMaxRetryDelay: https://github.com/aws/aws-sdk-go/blob/9f6e3bb9f523aef97fa1cd5c5f8ba8ecf212e44e/aws/client/default_retryer.go#L48-L49.
	)
	awsbaseConfig := awsbase.Config{
		AccessKey:         c.AccessKey,
		AllowedAccountIds: c.AllowedAccountIds,
		APNInfo: &awsbase.APNInfo{
			PartnerName: "HashiCorp",
			Products: []awsbase.UserAgentProduct{
				{Name: "Terraform", Version: c.TerraformVersion, Comment: "+https://www.terraform.io"},
				{Name: "terraform-provider-aws", Version: version.ProviderVersion, Comment: "+https://registry.terraform.io/providers/hashicorp/aws"},
			},
		},
		AssumeRoleWithWebIdentity:      c.AssumeRoleWithWebIdentity,
		Backoff:                        &v1CompatibleBackoff{maxRetryDelay: maxBackoff},
		CallerDocumentationURL:         "https://registry.terraform.io/providers/hashicorp/aws",
		CallerName:                     "Terraform AWS Provider",
		EC2MetadataServiceEnableState:  c.EC2MetadataServiceEnableState,
		ForbiddenAccountIds:            c.ForbiddenAccountIds,
		IamEndpoint:                    c.Endpoints[names.IAM],
		Insecure:                       c.Insecure,
		HTTPClient:                     client.HTTPClient(ctx),
		HTTPProxy:                      c.HTTPProxy,
		HTTPSProxy:                     c.HTTPSProxy,
		HTTPProxyMode:                  awsbase.HTTPProxyModeLegacy,
		Logger:                         logger,
		MaxBackoff:                     maxBackoff,
		MaxRetries:                     c.MaxRetries,
		NoProxy:                        c.NoProxy,
		Profile:                        c.Profile,
		Region:                         c.Region,
		RetryMode:                      c.RetryMode,
		SecretKey:                      c.SecretKey,
		SkipCredsValidation:            c.SkipCredsValidation,
		SkipRequestingAccountId:        c.SkipRequestingAccountId,
		SsoEndpoint:                    c.Endpoints[names.SSO],
		StsEndpoint:                    c.Endpoints[names.STS],
		SuppressDebugLog:               c.SuppressDebugLog,
		Token:                          c.Token,
		TokenBucketRateLimiterCapacity: c.TokenBucketRateLimiterCapacity,
		UseDualStackEndpoint:           c.UseDualStackEndpoint,
		UseFIPSEndpoint:                c.UseFIPSEndpoint,
	}

	if c.AssumeRole != nil && c.AssumeRole.RoleARN != "" {
		awsbaseConfig.AssumeRole = c.AssumeRole
	}

	if c.CustomCABundle != "" {
		awsbaseConfig.CustomCABundle = c.CustomCABundle
	}

	if c.EC2MetadataServiceEndpoint != "" {
		awsbaseConfig.EC2MetadataServiceEndpoint = c.EC2MetadataServiceEndpoint
		awsbaseConfig.EC2MetadataServiceEndpointMode = c.EC2MetadataServiceEndpointMode
	}

	if len(c.SharedConfigFiles) != 0 {
		awsbaseConfig.SharedConfigFiles = c.SharedConfigFiles
	}

	if len(c.SharedCredentialsFiles) != 0 {
		awsbaseConfig.SharedCredentialsFiles = c.SharedCredentialsFiles
	}

	if c.STSRegion != "" {
		awsbaseConfig.StsRegion = c.STSRegion
	}

	// Avoid duplicate calls to STS by enabling SkipCredsValidation for the call to GetAwsConfig
	// and then restoring the configured value for the call to GetAwsAccountIDAndPartition.
	skipCredsValidation := awsbaseConfig.SkipCredsValidation
	awsbaseConfig.SkipCredsValidation = true

	tflog.Debug(ctx, "Configuring Terraform AWS Provider")
	ctx, cfg, awsDiags := awsbase.GetAwsConfig(ctx, &awsbaseConfig)

	for _, d := range awsDiags {
		diags = append(diags, diag.Diagnostic{
			Severity: baseSeverityToSDKSeverity(d.Severity()),
			Summary:  d.Summary(),
			Detail:   d.Detail(),
		})
	}

	if diags.HasError() {
		return nil, diags
	}

	if !c.SkipRegionValidation {
		if err := basevalidation.SupportedRegion(cfg.Region); err != nil {
			return nil, sdkdiag.AppendFromErr(diags, err)
		}
	}
	c.Region = cfg.Region

	awsbaseConfig.SkipCredsValidation = skipCredsValidation

	tflog.Debug(ctx, "Creating AWS SDK v1 session")
	session, awsDiags := awsbasev1.GetSession(ctx, &cfg, &awsbaseConfig)

	for _, d := range awsDiags {
		diags = append(diags, diag.Diagnostic{
			Severity: baseSeverityToSDKSeverity(d.Severity()),
			Summary:  fmt.Sprintf("creating AWS SDK v1 session: %s", d.Summary()),
			Detail:   d.Detail(),
		})
	}

	if diags.HasError() {
		return nil, diags
	}

	tflog.Debug(ctx, "Retrieving AWS account details")
	accountID, partition, awsDiags := awsbase.GetAwsAccountIDAndPartition(ctx, cfg, &awsbaseConfig)
	for _, d := range awsDiags {
		diags = append(diags, diag.Diagnostic{
			Severity: baseSeverityToSDKSeverity(d.Severity()),
			Summary:  fmt.Sprintf("Retrieving AWS account details: %s", d.Summary()),
			Detail:   d.Detail(),
		})
	}

	if accountID == "" {
		diags = append(diags, errs.NewWarningDiagnostic(
			"AWS account ID not found for provider",
			"See https://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id for implications."))
	}

	err := awsbaseConfig.VerifyAccountIDAllowed(accountID)
	if err != nil {
		return nil, sdkdiag.AppendErrorf(diags, err.Error())
	}

	dnsSuffix := "amazonaws.com"
	if p, ok := endpoints_sdkv1.PartitionForRegion(endpoints_sdkv1.DefaultPartitions(), c.Region); ok {
		dnsSuffix = p.DNSSuffix()
	}

	client.AccountID = accountID
	client.DefaultTagsConfig = c.DefaultTagsConfig
	client.dnsSuffix = dnsSuffix
	client.IgnoreTagsConfig = c.IgnoreTagsConfig
	client.Partition = partition
	client.Region = c.Region
	client.SetHTTPClient(ctx, session.Config.HTTPClient) // Must be called while client.Session is nil.
	client.session = session

	// Used for lazy-loading AWS API clients.
	client.awsConfig = &cfg
	client.clients = make(map[string]any, 0)
	client.conns = make(map[string]any, 0)
	client.endpoints = c.Endpoints
	client.logger = logger
	client.s3UsePathStyle = c.S3UsePathStyle
	client.s3USEast1RegionalEndpoint = c.S3USEast1RegionalEndpoint
	client.stsRegion = c.STSRegion

	return client, diags
}

func baseSeverityToSDKSeverity(s basediag.Severity) diag.Severity {
	switch s {
	case basediag.SeverityWarning:
		return diag.Warning
	case basediag.SeverityError:
		return diag.Error
	default:
		return -1
	}
}

func NormalizeS3USEast1RegionalEndpoint(v string) string {
	switch v := strings.ToLower(v); v {
	case "legacy", "regional":
		return v
	default:
		return ""
	}
}
