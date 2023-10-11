// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"
	"fmt"

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
	HTTPProxy                      string
	IgnoreTagsConfig               *tftags.IgnoreConfig
	Insecure                       bool
	MaxRetries                     int
	Profile                        string
	Region                         string
	RetryMode                      aws_sdkv2.RetryMode
	S3UsePathStyle                 bool
	S3UsEast1RegionalEndpoint      endpoints_sdkv1.S3UsEast1RegionalEndpoint
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
	UseDualStackEndpoint           bool
	UseFIPSEndpoint                bool
}

// ConfigureProvider configures the provided provider Meta (instance data).
func (c *Config) ConfigureProvider(ctx context.Context, client *AWSClient) (*AWSClient, diag.Diagnostics) {
	var diags diag.Diagnostics

	ctx, logger := logging.NewTfLogger(ctx)

	awsbaseConfig := awsbase.Config{
		AccessKey:                     c.AccessKey,
		AllowedAccountIds:             c.AllowedAccountIds,
		APNInfo:                       StdUserAgentProducts(c.TerraformVersion),
		AssumeRoleWithWebIdentity:     c.AssumeRoleWithWebIdentity,
		CallerDocumentationURL:        "https://registry.terraform.io/providers/hashicorp/aws",
		CallerName:                    "Terraform AWS Provider",
		EC2MetadataServiceEnableState: c.EC2MetadataServiceEnableState,
		ForbiddenAccountIds:           c.ForbiddenAccountIds,
		IamEndpoint:                   c.Endpoints[names.IAM],
		Insecure:                      c.Insecure,
		HTTPClient:                    client.HTTPClient(),
		HTTPProxy:                     c.HTTPProxy,
		Logger:                        logger,
		MaxRetries:                    c.MaxRetries,
		Profile:                       c.Profile,
		Region:                        c.Region,
		RetryMode:                     c.RetryMode,
		SecretKey:                     c.SecretKey,
		SkipCredsValidation:           c.SkipCredsValidation,
		SkipRequestingAccountId:       c.SkipRequestingAccountId,
		StsEndpoint:                   c.Endpoints[names.STS],
		SuppressDebugLog:              c.SuppressDebugLog,
		Token:                         c.Token,
		UseDualStackEndpoint:          c.UseDualStackEndpoint,
		UseFIPSEndpoint:               c.UseFIPSEndpoint,
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

	tflog.Debug(ctx, "Configuring Terraform AWS Provider")
	ctx, cfg, awsDiags := awsbase.GetAwsConfig(ctx, &awsbaseConfig)

	for _, d := range awsDiags {
		diags = append(diags, diag.Diagnostic{
			Severity: baseSeverityToSdkSeverity(d.Severity()),
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

	tflog.Debug(ctx, "Creating AWS SDK v1 session")
	sess, awsDiags := awsbasev1.GetSession(ctx, &cfg, &awsbaseConfig)

	for _, d := range awsDiags {
		diags = append(diags, diag.Diagnostic{
			Severity: baseSeverityToSdkSeverity(d.Severity()),
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
			Severity: baseSeverityToSdkSeverity(d.Severity()),
			Summary:  fmt.Sprintf("retrieving AWS account details: %s", d.Summary()),
			Detail:   d.Detail(),
		})
	}

	if accountID == "" {
		diags = append(diags, errs.NewWarningDiagnostic(
			"AWS account ID not found for provider",
			"See https://www.terraform.io/docs/providers/aws/index.html#skip_requesting_account_id for implications."))
	}

	err := awsbaseConfig.VerifyAccountIDAllowed(accountID)
	if err != nil {
		return nil, sdkdiag.AppendErrorf(diags, err.Error())
	}

	DNSSuffix := "amazonaws.com"
	if p, ok := endpoints_sdkv1.PartitionForRegion(endpoints_sdkv1.DefaultPartitions(), c.Region); ok {
		DNSSuffix = p.DNSSuffix()
	}

	client.AccountID = accountID
	client.DefaultTagsConfig = c.DefaultTagsConfig
	client.DNSSuffix = DNSSuffix
	client.IgnoreTagsConfig = c.IgnoreTagsConfig
	client.Partition = partition
	client.Region = c.Region
	client.ReverseDNSPrefix = ReverseDNS(DNSSuffix)
	client.SetHTTPClient(sess.Config.HTTPClient) // Must be called while client.Session is nil.
	client.Session = sess
	client.TerraformVersion = c.TerraformVersion

	// Used for lazy-loading AWS API clients.
	client.awsConfig = &cfg
	client.clients = make(map[string]any, 0)
	client.conns = make(map[string]any, 0)
	client.endpoints = c.Endpoints
	client.s3UsePathStyle = c.S3UsePathStyle
	client.s3UsEast1RegionalEndpoint = c.S3UsEast1RegionalEndpoint
	client.stsRegion = c.STSRegion

	return client, diags
}

func baseSeverityToSdkSeverity(s basediag.Severity) diag.Severity {
	switch s {
	case basediag.SeverityWarning:
		return diag.Warning
	case basediag.SeverityError:
		return diag.Error
	default:
		return -1
	}
}
