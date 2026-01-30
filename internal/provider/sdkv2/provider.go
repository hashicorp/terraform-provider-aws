// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	"github.com/hashicorp/aws-sdk-go-base/v2/useragent"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/internal/attribute"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	tfunique "github.com/hashicorp/terraform-provider-aws/internal/unique"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	resourceSchemasValidated bool
)

type sdkProvider struct {
	provider        *schema.Provider
	servicePackages iter.Seq2[int, conns.ServicePackage]
}

// providerMeta matches the shape of ProviderMetaSchema
type providerMeta struct {
	UserAgent []string `cty:"user_agent"`
}

// NewProvider returns a new, initialized Terraform Plugin SDK v2-style provider instance.
// The provider instance is fully configured once the `ConfigureContextFunc` has been called.
func NewProvider(ctx context.Context) (*schema.Provider, error) {
	log.Printf("Creating Terraform AWS Provider (SDKv2-style)...")

	sdkProvider := &sdkProvider{
		provider: &schema.Provider{
			// This schema must match exactly the Terraform Protocol v6 (Terraform Plugin Framework) provider's schema.
			// Notably the attributes can have no Default values.
			Schema: map[string]*schema.Schema{
				"access_key": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "The access key for API operations. You can retrieve this\n" +
						"from the 'Security & Credentials' section of the AWS console.",
				},
				"allowed_account_ids": {
					Type:          schema.TypeSet,
					Elem:          &schema.Schema{Type: schema.TypeString},
					Optional:      true,
					ConflictsWith: []string{"forbidden_account_ids"},
				},
				"assume_role":                   assumeRoleSchema(),
				"assume_role_with_web_identity": assumeRoleWithWebIdentitySchema(),
				"custom_ca_bundle": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "File containing custom root and intermediate certificates. " +
						"Can also be configured using the `AWS_CA_BUNDLE` environment variable. " +
						"(Setting `ca_bundle` in the shared config file is not supported.)",
				},
				"default_tags": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Configuration block with settings to default resource tags across all resources.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"tags": {
								Type:     schema.TypeMap,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								Description: "Resource tags to default across all resources. " +
									"Can also be configured with environment variables like `" + tftags.DefaultTagsEnvVarPrefix + "<tag_name>`.",
							},
						},
					},
				},
				"ec2_metadata_service_endpoint": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "Address of the EC2 metadata service endpoint to use. " +
						"Can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.",
				},
				"ec2_metadata_service_endpoint_mode": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "Protocol to use with EC2 metadata service endpoint." +
						"Valid values are `IPv4` and `IPv6`. Can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE` environment variable.",
				},
				"endpoints": endpointsSchema(),
				"forbidden_account_ids": {
					Type:          schema.TypeSet,
					Elem:          &schema.Schema{Type: schema.TypeString},
					Optional:      true,
					ConflictsWith: []string{"allowed_account_ids"},
				},
				"http_proxy": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "URL of a proxy to use for HTTP requests when accessing the AWS API. " +
						"Can also be set using the `HTTP_PROXY` or `http_proxy` environment variables.",
				},
				"https_proxy": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "URL of a proxy to use for HTTPS requests when accessing the AWS API. " +
						"Can also be set using the `HTTPS_PROXY` or `https_proxy` environment variables.",
				},
				"ignore_tags": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Configuration block with settings to ignore resource tags across all resources.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"keys": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								Description: "Resource tag keys to ignore across all resources. " +
									"Can also be configured with the " + tftags.IgnoreTagsKeysEnvVar + " environment variable.",
							},
							"key_prefixes": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								Description: "Resource tag key prefixes to ignore across all resources. " +
									"Can also be configured with the " + tftags.IgnoreTagsKeyPrefixesEnvVar + " environment variable.",
							},
						},
					},
				},
				"insecure": {
					Type:     schema.TypeBool,
					Optional: true,
					Description: "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted, " +
						"default value is `false`",
				},
				"max_retries": {
					Type:     schema.TypeInt,
					Optional: true,
					Description: "The maximum number of times an AWS API request is\n" +
						"being executed. If the API request still fails, an error is\n" +
						"thrown.",
				},
				"no_proxy": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "Comma-separated list of hosts that should not use HTTP or HTTPS proxies. " +
						"Can also be set using the `NO_PROXY` or `no_proxy` environment variables.",
				},
				"profile": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "The profile for API operations. If not set, the default profile\n" +
						"created with `aws configure` will be used.",
				},
				"region": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "The region where AWS operations will take place. Examples\n" +
						"are us-east-1, us-west-2, etc.", // lintignore:AWSAT003,
				},
				"retry_mode": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "Specifies how retries are attempted. Valid values are `standard` and `adaptive`. " +
						"Can also be configured using the `AWS_RETRY_MODE` environment variable.",
				},
				"s3_use_path_style": {
					Type:     schema.TypeBool,
					Optional: true,
					Description: "Set this to true to enable the request to use path-style addressing,\n" +
						"i.e., https://s3.amazonaws.com/BUCKET/KEY. By default, the S3 client will\n" +
						"use virtual hosted bucket addressing when possible\n" +
						"(https://BUCKET.s3.amazonaws.com/KEY). Specific to the Amazon S3 service.",
				},
				"s3_us_east_1_regional_endpoint": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "Specifies whether S3 API calls in the `us-east-1` region use the legacy global endpoint or a regional endpoint. " + //lintignore:AWSAT003
						"Valid values are `legacy` or `regional`. " +
						"Can also be configured using the `AWS_S3_US_EAST_1_REGIONAL_ENDPOINT` environment variable or the `s3_us_east_1_regional_endpoint` shared config file parameter",
				},
				"secret_key": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "The secret key for API operations. You can retrieve this\n" +
						"from the 'Security & Credentials' section of the AWS console.",
				},
				"shared_config_files": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "List of paths to shared config files. If not set, defaults to [~/.aws/config].",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"shared_credentials_files": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "List of paths to shared credentials files. If not set, defaults to [~/.aws/credentials].",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"skip_credentials_validation": {
					Type:     schema.TypeBool,
					Optional: true,
					Description: "Skip the credentials validation via STS API. " +
						"Used for AWS API implementations that do not have STS available/implemented.",
				},
				"skip_metadata_api_check": {
					Type:         nullable.TypeNullableBool,
					Optional:     true,
					ValidateFunc: nullable.ValidateTypeStringNullableBool,
					Description: "Skip the AWS Metadata API check. " +
						"Used for AWS API implementations that do not have a metadata api endpoint.",
				},
				"skip_region_validation": {
					Type:     schema.TypeBool,
					Optional: true,
					Description: "Skip static validation of region name. " +
						"Used by users of alternative AWS-like APIs or users w/ access to regions that are not public (yet).",
				},
				"skip_requesting_account_id": {
					Type:     schema.TypeBool,
					Optional: true,
					Description: "Skip requesting the account ID. " +
						"Used for AWS API implementations that do not have IAM/STS API and/or metadata API.",
				},
				"sts_region": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "The region where AWS STS operations will take place. Examples\n" +
						"are us-east-1 and us-west-2.", // lintignore:AWSAT003,
				},
				"tag_policy_compliance": {
					Type:     schema.TypeString,
					Optional: true,
					Description: `The severity with which to enforce organizational tagging policies on resources managed by this provider instance. ` +
						`At this time this only includes compliance with required tag keys by resource type. ` +
						`Valid values are "error", "warning", and "disabled". ` +
						`When unset or "disabled", tag policy compliance will not be enforced by the provider. ` +
						`Can also be configured with the ` + tftags.TagPolicyComplianceEnvVar + ` environment variable.`,
				},
				"token": {
					Type:     schema.TypeString,
					Optional: true,
					Description: "session token. A session token is only required if you are\n" +
						"using temporary security credentials.",
				},
				"token_bucket_rate_limiter_capacity": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "The capacity of the AWS SDK's token bucket rate limiter.",
				},
				"use_dualstack_endpoint": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Resolve an endpoint with DualStack capability",
				},
				"use_fips_endpoint": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Resolve an endpoint with FIPS capability",
				},
				"user_agent": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Product details to append to the User-Agent string sent in all AWS API calls.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},

			// ProviderMetaSchema enables module-scoped User-Agent modifications
			ProviderMetaSchema: map[string]*schema.Schema{
				"user_agent": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Product details to append to the User-Agent string sent in all AWS API calls.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},

			// Data sources and resources implemented using Terraform Plugin SDK
			// should use the @SDKDataSource and @SDKResource function-level annotations
			// rather than adding directly to these maps.
			DataSourcesMap: make(map[string]*schema.Resource),
			ResourcesMap:   make(map[string]*schema.Resource),
		},
		servicePackages: slices.All(servicePackages(ctx)),
	}

	sdkProvider.provider.ConfigureContextFunc = sdkProvider.configure

	// Acceptance tests call this function multiple times, potentially in parallel.
	// To avoid "fatal error: concurrent map writes", take a lock.
	const (
		mutexKVKey = "provider.New"
	)
	conns.GlobalMutexKV.Lock(mutexKVKey)
	defer conns.GlobalMutexKV.Unlock(mutexKVKey)

	// Because we try and share resource schemas as much as possible,
	// we need to ensure that we only validate the resource schemas once.
	if !resourceSchemasValidated {
		if err := sdkProvider.validateResourceSchemas(ctx); err != nil {
			return nil, err
		}

		resourceSchemasValidated = true
	}

	servicePackageMap, err := sdkProvider.initialize(ctx)

	if err != nil {
		return nil, err
	}

	// Set the provider Meta (instance data) here.
	// It will be overwritten by the result of the call to ConfigureContextFunc,
	// but can be used pre-configuration by other (non-primary) provider servers.
	var c *conns.AWSClient
	if v, ok := sdkProvider.provider.Meta().(*conns.AWSClient); ok {
		c = v
	} else {
		c = new(conns.AWSClient)
	}
	c.SetServicePackages(ctx, servicePackageMap)
	sdkProvider.provider.SetMeta(c)

	return sdkProvider.provider, nil
}

// configure ensures that the provider is fully configured.
func (p *sdkProvider) configure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	terraformVersion := p.provider.TerraformVersion
	if terraformVersion == "" {
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing it's 0.10 or 0.11
		terraformVersion = "0.11+compatible"
	}

	config := conns.Config{
		AccessKey:                      d.Get("access_key").(string),
		CustomCABundle:                 d.Get("custom_ca_bundle").(string),
		EC2MetadataServiceEndpoint:     d.Get("ec2_metadata_service_endpoint").(string),
		EC2MetadataServiceEndpointMode: d.Get("ec2_metadata_service_endpoint_mode").(string),
		Endpoints:                      make(map[string]string),
		Insecure:                       d.Get("insecure").(bool),
		MaxRetries:                     25, // Set default here, not in schema (muxing with v6 provider).
		Profile:                        d.Get("profile").(string),
		Region:                         d.Get("region").(string),
		S3UsePathStyle:                 d.Get("s3_use_path_style").(bool),
		SecretKey:                      d.Get("secret_key").(string),
		SkipCredsValidation:            d.Get("skip_credentials_validation").(bool),
		SkipRegionValidation:           d.Get("skip_region_validation").(bool),
		SkipRequestingAccountId:        d.Get("skip_requesting_account_id").(bool),
		STSRegion:                      d.Get("sts_region").(string),
		TerraformVersion:               terraformVersion,
		Token:                          d.Get("token").(string),
		TokenBucketRateLimiterCapacity: d.Get("token_bucket_rate_limiter_capacity").(int),
		UseDualStackEndpoint:           d.Get("use_dualstack_endpoint").(bool),
		UseFIPSEndpoint:                d.Get("use_fips_endpoint").(bool),
	}

	if v, ok := d.Get("retry_mode").(string); ok && v != "" {
		mode, err := aws.ParseRetryMode(v)
		if err != nil {
			return nil, sdkdiag.AppendFromErr(diags, err)
		}
		config.RetryMode = mode
	}

	if v, ok := d.Get("s3_us_east_1_regional_endpoint").(string); ok && v != "" {
		endpoint := conns.NormalizeS3USEast1RegionalEndpoint(v)
		if endpoint == "legacy" {
			diags = append(diags,
				errs.NewAttributeWarningDiagnostic(
					cty.GetAttrPath("s3_us_east_1_regional_endpoint"),
					"Global S3 Endpoint Support Deprecated",
					"Support for the global S3 endpoint is deprecated. The \"s3_us_east_1_regional_endpoint\" "+
						"argument will be removed in a future major version. Remove this argument from the "+
						"configuration, or set it to \"regional\" to verify connectivity with the regional "+
						"S3 endpoint instead.",
				),
			)
		}
		config.S3USEast1RegionalEndpoint = endpoint
	}

	if v, ok := d.GetOk("allowed_account_ids"); ok && v.(*schema.Set).Len() > 0 {
		config.AllowedAccountIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("assume_role"); ok {
		path := cty.GetAttrPath("assume_role")
		v := v.([]any)
		if len(v) == 1 {
			if v[0] == nil {
				diags = append(diags,
					errs.NewAttributeRequiredWillBeError(path.IndexInt(0), "role_arn"),
				)
			} else {
				l := v[0].(map[string]any)
				if s, ok := l["role_arn"]; !ok || s == "" {
					diags = append(diags,
						errs.NewAttributeRequiredWillBeError(path.IndexInt(0), "role_arn"),
					)
				} else {
					ar, dg := expandAssumeRoles(ctx, path, v)
					diags = append(diags, dg...)
					if dg.HasError() {
						return nil, diags
					}
					config.AssumeRole = ar
				}
			}
		} else if len(v) > 1 {
			ar, dg := expandAssumeRoles(ctx, path, v)
			diags = append(diags, dg...)
			if dg.HasError() {
				return nil, diags
			}
			config.AssumeRole = ar
		}
	}

	if v, ok := d.GetOk("assume_role_with_web_identity"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		config.AssumeRoleWithWebIdentity = expandAssumeRoleWithWebIdentity(ctx, v.([]any)[0].(map[string]any))
		tflog.Info(ctx, "assume_role_with_web_identity configuration set", map[string]any{
			"tf_aws.assume_role_with_web_identity.role_arn":     config.AssumeRoleWithWebIdentity.RoleARN,
			"tf_aws.assume_role_with_web_identity.session_name": config.AssumeRoleWithWebIdentity.SessionName,
		})
	}

	if v, ok := d.GetOk("default_tags"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		config.DefaultTagsConfig = expandDefaultTags(ctx, v.([]any)[0].(map[string]any))
	} else {
		config.DefaultTagsConfig = expandDefaultTags(ctx, nil)
	}

	v := d.Get("endpoints")
	endpoints, dx := expandEndpoints(ctx, v.(*schema.Set).List())
	diags = append(diags, dx...)
	if diags.HasError() {
		return nil, diags
	}
	config.Endpoints = endpoints

	if v, ok := d.GetOk("forbidden_account_ids"); ok && v.(*schema.Set).Len() > 0 {
		config.ForbiddenAccountIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOkExists("http_proxy"); ok {
		if s, sok := v.(string); sok {
			config.HTTPProxy = aws.String(s)
		}
	}
	if v, ok := d.GetOkExists("https_proxy"); ok {
		if s, sok := v.(string); sok {
			config.HTTPSProxy = aws.String(s)
		}
	}

	if v, ok := d.Get("no_proxy").(string); ok && v != "" {
		config.NoProxy = v
	}

	if v, ok := d.GetOk("ignore_tags"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		config.IgnoreTagsConfig = expandIgnoreTags(ctx, v.([]any)[0].(map[string]any))
	} else {
		config.IgnoreTagsConfig = expandIgnoreTags(ctx, nil)
	}

	tagCfg, dg := expandTagPolicyConfig(cty.GetAttrPath("tag_policy_compliance"), d.Get("tag_policy_compliance").(string))
	diags = append(diags, dg...)
	if dg.HasError() {
		return nil, diags
	}
	config.TagPolicyConfig = tagCfg

	if v, ok := d.GetOk("max_retries"); ok {
		config.MaxRetries = v.(int)
	}

	if v, ok := d.GetOk("shared_credentials_files"); ok && len(v.([]any)) > 0 {
		config.SharedCredentialsFiles = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk("shared_config_files"); ok && len(v.([]any)) > 0 {
		config.SharedConfigFiles = flex.ExpandStringValueList(v.([]any))
	}

	if v, null, _ := nullable.Bool(d.Get("skip_metadata_api_check").(string)).ValueBool(); !null {
		if v {
			config.EC2MetadataServiceEnableState = imds.ClientDisabled
		} else {
			config.EC2MetadataServiceEnableState = imds.ClientEnabled
		}
	}

	if v, ok := d.GetOk("user_agent"); ok && len(v.([]any)) > 0 {
		config.UserAgent = useragent.FromSlice(v.([]any))
	}

	var c *conns.AWSClient
	if v, ok := p.provider.Meta().(*conns.AWSClient); ok {
		c = v
	} else {
		c = new(conns.AWSClient)
	}
	c, ds := config.ConfigureProvider(ctx, c)
	diags = append(diags, ds...)

	if diags.HasError() {
		return nil, diags
	}

	return c, diags
}

// initialize is called from `New` to perform any Terraform Plugin SDK v2-style initialization.
func (p *sdkProvider) initialize(ctx context.Context) (map[string]conns.ServicePackage, error) {
	log.Printf("Initializing Terraform AWS Provider (SDKv2-style)...")

	var errs []error
	servicePackageMap := make(map[string]conns.ServicePackage)

	for _, sp := range p.servicePackages {
		servicePackageName := sp.ServicePackageName()
		servicePackageMap[servicePackageName] = sp

		for _, v := range sp.SDKDataSources(ctx) {
			typeName := v.TypeName

			if _, ok := p.provider.DataSourcesMap[typeName]; ok {
				errs = append(errs, fmt.Errorf("duplicate data source: %s", typeName))
				continue
			}

			r := v.Factory()

			// Ensure that the correct CRUD handler variants are used.
			if r.Read != nil || r.ReadContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Read handler variant: %s data source", typeName))
				continue
			}

			var isRegionOverrideEnabled bool
			if v := v.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
				isRegionOverrideEnabled = true
			}

			var interceptors interceptorInvocations

			if isRegionOverrideEnabled {
				v := v.Region.Value()
				s := r.SchemaMap()

				if _, ok := s[names.AttrRegion]; !ok {
					// Inject a top-level "region" attribute.
					regionSchema := attribute.Region()

					if f := r.SchemaFunc; f != nil {
						r.SchemaFunc = func() map[string]*schema.Schema {
							s := f()
							s[names.AttrRegion] = regionSchema
							return s
						}
					} else {
						r.Schema[names.AttrRegion] = regionSchema
					}
				}

				if v.IsValidateOverrideInPartition {
					interceptors = append(interceptors, interceptorInvocation{
						when:        Before,
						why:         Read,
						interceptor: dataSourceValidateRegion(),
					})
				}
				interceptors = append(interceptors, interceptorInvocation{
					when:        After,
					why:         Read,
					interceptor: setRegionInState(),
				})
			}

			if !tfunique.IsHandleNil(v.Tags) {
				interceptors = append(interceptors, interceptorInvocation{
					when:        Before | After,
					why:         Read,
					interceptor: dataSourceTransparentTagging(v.Tags),
				})
			}

			opts := wrappedDataSourceOptions{
				bootstrapContext: func(ctx context.Context, getAttribute getAttributeFunc, getProviderMeta getProviderMetaFunc, meta any) (context.Context, error) {
					var overrideRegion string

					if isRegionOverrideEnabled && getAttribute != nil {
						if region, ok := getAttribute(names.AttrRegion); ok {
							overrideRegion = region.(string)
						}
					}

					ctx = conns.NewResourceContext(ctx, servicePackageName, v.Name, v.TypeName, overrideRegion)
					if c, ok := meta.(*conns.AWSClient); ok {
						ctx = tftags.NewContext(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), c.TagPolicyConfig(ctx))
						ctx = c.RegisterLogger(ctx)
					}

					if getProviderMeta != nil {
						var metadata providerMeta
						if err := getProviderMeta(&metadata); err != nil {
							return ctx, fmt.Errorf("getting provider_meta: %w", err)
						}

						if len(metadata.UserAgent) > 0 {
							ctx = useragent.Context(ctx, useragent.FromSlice(metadata.UserAgent))
						}
					}

					return ctx, nil
				},
				interceptors: interceptors,
				typeName:     typeName,
			}
			wrapDataSource(r, opts)
			p.provider.DataSourcesMap[typeName] = r
		}

		for _, resource := range sp.SDKResources(ctx) {
			typeName := resource.TypeName

			if _, ok := p.provider.ResourcesMap[typeName]; ok {
				errs = append(errs, fmt.Errorf("duplicate resource: %s", typeName))
				continue
			}

			r := resource.Factory()

			// Ensure that the correct CRUD handler variants are used.
			if r.Create != nil || r.CreateContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Create handler variant: %s resource", typeName))
				continue
			}
			if r.Read != nil || r.ReadContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Read handler variant: %s resource", typeName))
				continue
			}
			if r.Update != nil || r.UpdateContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Update handler variant: %s resource", typeName))
				continue
			}
			if r.Delete != nil || r.DeleteContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Delete handler variant: %s resource", typeName))
				continue
			}

			var isRegionOverrideEnabled bool
			if v := resource.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
				isRegionOverrideEnabled = true
			}

			var interceptors interceptorInvocations

			if isRegionOverrideEnabled {
				v := resource.Region.Value()
				s := r.SchemaMap()

				if _, ok := s[names.AttrRegion]; !ok {
					// Inject a top-level "region" attribute.
					regionSchema := attribute.Region()

					// If the resource defines no Update handler then add a stub to fake out 'Provider.Validate'.
					if r.UpdateWithoutTimeout == nil {
						r.UpdateWithoutTimeout = schema.NoopContext
					}

					if f := r.SchemaFunc; f != nil {
						r.SchemaFunc = func() map[string]*schema.Schema {
							s := f()
							s[names.AttrRegion] = regionSchema
							return s
						}
					} else {
						r.Schema[names.AttrRegion] = regionSchema
					}
				}

				if v.IsValidateOverrideInPartition {
					interceptors = append(interceptors, interceptorInvocation{
						when:        Before,
						why:         CustomizeDiff,
						interceptor: resourceValidateRegion(),
					})
				}
				interceptors = append(interceptors, interceptorInvocation{
					when:        Before,
					why:         CustomizeDiff,
					interceptor: defaultRegion(),
				})
				interceptors = append(interceptors, interceptorInvocation{
					when:        After,
					why:         Read,
					interceptor: setRegionInState(),
				})
				// We can't just set the injected "region" attribute to ForceNew because if
				// a plan is run with '-refresh=false', then after provider v5 to v6 upgrade
				// the region attribute is not set in state and its value shows a change.
				interceptors = append(interceptors, interceptorInvocation{
					when:        Before,
					why:         CustomizeDiff,
					interceptor: forceNewIfRegionChanges(),
				})
				if resource.Identity.HasInherentRegion() {
					interceptors = append(interceptors, resourceImportRegionNoDefault())
				} else {
					interceptors = append(interceptors, resourceImportRegion())
				}
			}

			if !tfunique.IsHandleNil(resource.Tags) {
				interceptors = append(interceptors, interceptorInvocation{
					when:        Before | After | Finally,
					why:         Create | Read | Update,
					interceptor: resourceTransparentTagging(resource.Tags),
				})
				interceptors = append(interceptors, interceptorInvocation{
					when:        Before,
					why:         CustomizeDiff,
					interceptor: setTagsAll(),
				})
				interceptors = append(interceptors, interceptorInvocation{
					when:        Before,
					why:         CustomizeDiff,
					interceptor: validateRequiredTags(),
				})
			}

			if len(resource.Identity.Attributes) > 0 {
				r.Identity = newResourceIdentity(resource.Identity)

				if resource.Identity.IsMutable {
					r.ResourceBehavior.MutableIdentity = true
				}

				interceptors = append(interceptors, newIdentityInterceptor(&resource.Identity))
			}

			if resource.Import.CustomImport {
				if r.Importer == nil || r.Importer.StateContext == nil {
					errs = append(errs, fmt.Errorf("resource type %s: uses CustomImport but does not define an import function", typeName))
					continue
				}

				customResourceImporter(r, &resource.Identity, &resource.Import)
			}
			if resource.Import.WrappedImport {
				if r.Importer != nil && r.Importer.StateContext != nil {
					errs = append(errs, fmt.Errorf("resource type %s: uses WrappedImport but defines an import function", typeName))
					continue
				}

				if resource.Identity.IsARN {
					r.Importer = arnIdentityResourceImporter(resource.Identity)
				} else if resource.Identity.IsSingleton {
					r.Importer = singletonIdentityResourceImporter(resource.Identity)
				} else if resource.Identity.IsCustomInherentRegion {
					r.Importer = customInherentRegionResourceImporter(resource.Identity)
				} else {
					r.Importer = newParameterizedIdentityImporter(resource.Identity, resource.Import)
				}
			}

			opts := wrappedResourceOptions{
				// bootstrapContext is run on all wrapped methods before any interceptors.
				bootstrapContext: func(ctx context.Context, getAttribute getAttributeFunc, getProviderMeta getProviderMetaFunc, meta any) (context.Context, error) {
					var overrideRegion string

					if isRegionOverrideEnabled && getAttribute != nil {
						if region, ok := getAttribute(names.AttrRegion); ok && region != nil {
							overrideRegion = region.(string)
						}
					}

					ctx = conns.NewResourceContext(ctx, servicePackageName, resource.Name, resource.TypeName, overrideRegion)
					if c, ok := meta.(*conns.AWSClient); ok {
						ctx = tftags.NewContext(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), c.TagPolicyConfig(ctx))
						ctx = c.RegisterLogger(ctx)
						if s := c.RandomnessSource(); s != nil {
							ctx = vcr.NewContext(ctx, s)
						}
					}

					if getProviderMeta != nil {
						var metadata providerMeta
						if err := getProviderMeta(&metadata); err != nil {
							return ctx, fmt.Errorf("getting provider_meta: %w", err)
						}

						if len(metadata.UserAgent) > 0 {
							ctx = useragent.Context(ctx, useragent.FromSlice(metadata.UserAgent))
						}
					}

					return ctx, nil
				},
				interceptors: interceptors,
				typeName:     typeName,
			}
			wrapResource(r, opts)
			p.provider.ResourcesMap[typeName] = r
		}
	}

	return servicePackageMap, errors.Join(errs...)
}

// validateResourceSchemas is called from `New` to validate Terraform Plugin SDK v2-style resource schemas.
func (p *sdkProvider) validateResourceSchemas(ctx context.Context) error {
	var errs []error

	for _, sp := range p.servicePackages {
		for _, v := range sp.SDKDataSources(ctx) {
			typeName := v.TypeName
			r := v.Factory()
			s := r.SchemaMap()

			if v := v.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
				if _, ok := s[names.AttrRegion]; ok {
					errs = append(errs, fmt.Errorf("`%s` attribute is defined: %s data source", names.AttrRegion, typeName))
					continue
				}
			}

			if !tfunique.IsHandleNil(v.Tags) {
				// The data source has opted in to transparent tagging.
				// Ensure that the schema look OK.
				if v, ok := s[names.AttrTags]; ok {
					if !v.Computed {
						errs = append(errs, fmt.Errorf("`%s` attribute must be Computed: %s data source", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s data source", names.AttrTags, typeName))
					continue
				}
			}
		}

		for _, resource := range sp.SDKResources(ctx) {
			typeName := resource.TypeName
			r := resource.Factory()
			s := r.SchemaMap()

			if v := resource.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
				if _, ok := s[names.AttrRegion]; ok {
					errs = append(errs, fmt.Errorf("`%s` attribute is defined: %s resource", names.AttrRegion, typeName))
					continue
				}
			}

			if !tfunique.IsHandleNil(resource.Tags) {
				// The resource has opted in to transparent tagging.
				// Ensure that the schema look OK.
				if v, ok := s[names.AttrTags]; ok {
					if v.Computed {
						errs = append(errs, fmt.Errorf("`%s` attribute cannot be Computed: %s resource", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s resource", names.AttrTags, typeName))
					continue
				}
				if v, ok := s[names.AttrTagsAll]; ok {
					if !v.Computed {
						errs = append(errs, fmt.Errorf("`%s` attribute must be Computed: %s resource", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s resource", names.AttrTagsAll, typeName))
					continue
				}
			}

			if resource.Identity.IsCustomInherentRegion {
				if resource.Identity.IsGlobalResource {
					errs = append(errs, fmt.Errorf("`IsCustomInherentRegion` is not supported for Global resources: %s resource", typeName))
					continue
				}
			}
		}
	}

	return errors.Join(errs...)
}

func assumeRoleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"duration": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
					ValidateFunc: validAssumeRoleDuration,
				},
				"external_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A unique identifier that might be required when you assume a role in another account.",
					ValidateFunc: validation.All(
						validation.StringLenBetween(2, 1224),
						validation.StringMatch(regexache.MustCompile(`[\w+=,.@:\/\-]*`), ""),
					),
				},
				"policy": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
					ValidateFunc: validation.StringIsJSON,
				},
				"policy_arns": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: verify.ValidARN,
					},
				},
				"role_arn": {
					Type:         schema.TypeString,
					Optional:     true, // For historical reasons, we allow an empty `assume_role` block
					Description:  "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
					ValidateFunc: verify.ValidARN,
				},
				"session_name": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "An identifier for the assumed role session.",
					ValidateFunc: validAssumeRoleSessionName,
				},
				"source_identity": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "Source identity specified by the principal assuming the role.",
					ValidateFunc: validAssumeRoleSourceIdentity,
				},
				"tags": {
					Type:        schema.TypeMap,
					Optional:    true,
					Description: "Assume role session tags.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"transitive_tag_keys": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Assume role session tag keys to pass to any subsequent sessions.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func assumeRoleWithWebIdentitySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"duration": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
					ValidateFunc: validAssumeRoleDuration,
				},
				"policy": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
					ValidateFunc: validation.StringIsJSON,
				},
				"policy_arns": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: verify.ValidARN,
					},
				},
				"role_arn": {
					Type:         schema.TypeString,
					Optional:     true, // For historical reasons, we allow an empty `assume_role_with_web_identity` block
					Description:  "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
					ValidateFunc: verify.ValidARN,
				},
				"session_name": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "An identifier for the assumed role session.",
					ValidateFunc: validAssumeRoleSessionName,
				},
				"web_identity_token": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(4, 20000),
					ExactlyOneOf: []string{"assume_role_with_web_identity.0.web_identity_token", "assume_role_with_web_identity.0.web_identity_token_file"},
				},
				"web_identity_token_file": {
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{"assume_role_with_web_identity.0.web_identity_token", "assume_role_with_web_identity.0.web_identity_token_file"},
				},
			},
		},
	}
}

func expandAssumeRoles(ctx context.Context, path cty.Path, tfList []any) (result []awsbase.AssumeRole, diags diag.Diagnostics) {
	result = make([]awsbase.AssumeRole, len(tfList))

	for i, v := range tfList {
		path := path.IndexInt(i)
		if ar, ok := v.(map[string]any); ok {
			x, d := expandAssumeRole(ctx, path, ar)
			diags = append(diags, d...)
			if d.HasError() {
				return result, diags
			}
			result[i] = x
			tflog.Info(ctx, "assume_role configuration set", map[string]any{
				"tf_aws.assume_role.index":           i,
				"tf_aws.assume_role.role_arn":        result[i].RoleARN,
				"tf_aws.assume_role.session_name":    result[i].SessionName,
				"tf_aws.assume_role.external_id":     result[i].ExternalID,
				"tf_aws.assume_role.source_identity": result[i].SourceIdentity,
			})
		} else {
			return result, append(diags, errs.NewAttributeRequiredError(path, "role_arn"))
		}
	}

	return result, diags
}

func expandAssumeRole(_ context.Context, path cty.Path, tfMap map[string]any) (result awsbase.AssumeRole, diags diag.Diagnostics) {
	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		result.RoleARN = v
	} else {
		return result, append(diags, errs.NewAttributeRequiredError(path, "role_arn"))
	}

	if v, ok := tfMap["duration"].(string); ok && v != "" {
		duration, _ := time.ParseDuration(v)
		result.Duration = duration
	}

	if v, ok := tfMap["external_id"].(string); ok && v != "" {
		result.ExternalID = v
	}

	if v, ok := tfMap["policy"].(string); ok && v != "" {
		result.Policy = v
	}

	if v, ok := tfMap["policy_arns"].(*schema.Set); ok && v.Len() > 0 {
		result.PolicyARNs = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["session_name"].(string); ok && v != "" {
		result.SessionName = v
	}

	if v, ok := tfMap["source_identity"].(string); ok && v != "" {
		result.SourceIdentity = v
	}

	if v, ok := tfMap["tags"].(map[string]any); ok && len(v) > 0 {
		result.Tags = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["transitive_tag_keys"].(*schema.Set); ok && v.Len() > 0 {
		result.TransitiveTagKeys = flex.ExpandStringValueSet(v)
	}

	return result, diags
}

func expandAssumeRoleWithWebIdentity(_ context.Context, tfMap map[string]any) *awsbase.AssumeRoleWithWebIdentity {
	if tfMap == nil {
		return nil
	}

	assumeRole := awsbase.AssumeRoleWithWebIdentity{}

	if v, ok := tfMap["duration"].(string); ok && v != "" {
		duration, _ := time.ParseDuration(v)
		assumeRole.Duration = duration
	}

	if v, ok := tfMap["policy"].(string); ok && v != "" {
		assumeRole.Policy = v
	}

	if v, ok := tfMap["policy_arns"].(*schema.Set); ok && v.Len() > 0 {
		assumeRole.PolicyARNs = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		assumeRole.RoleARN = v
	}

	if v, ok := tfMap["session_name"].(string); ok && v != "" {
		assumeRole.SessionName = v
	}

	if v, ok := tfMap["web_identity_token"].(string); ok && v != "" {
		assumeRole.WebIdentityToken = v
	}

	if v, ok := tfMap["web_identity_token_file"].(string); ok && v != "" {
		assumeRole.WebIdentityTokenFile = v
	}

	return &assumeRole
}

func expandDefaultTags(ctx context.Context, tfMap map[string]any) *tftags.DefaultConfig {
	tags := make(map[string]any)
	for _, ev := range os.Environ() {
		k, v, _ := strings.Cut(ev, "=")
		if before, tk, ok := strings.Cut(k, tftags.DefaultTagsEnvVarPrefix); ok && before == "" {
			tags[tk] = v
		}
	}

	if cfgTags, ok := tfMap["tags"].(map[string]any); ok {
		maps.Copy(tags, cfgTags)
	}

	if len(tags) > 0 {
		return &tftags.DefaultConfig{
			Tags: tftags.New(ctx, tags),
		}
	}

	return nil
}

func expandIgnoreTags(ctx context.Context, tfMap map[string]any) *tftags.IgnoreConfig {
	var keys, keyPrefixes []any

	if tfMap != nil {
		if v, ok := tfMap["keys"].(*schema.Set); ok {
			keys = v.List()
		}
		if v, ok := tfMap["key_prefixes"].(*schema.Set); ok {
			keyPrefixes = v.List()
		}
	}

	if v := os.Getenv(tftags.IgnoreTagsKeysEnvVar); v != "" {
		for k := range strings.SplitSeq(v, ",") {
			if trimmed := strings.TrimSpace(k); trimmed != "" {
				keys = append(keys, trimmed)
			}
		}
	}

	if v := os.Getenv(tftags.IgnoreTagsKeyPrefixesEnvVar); v != "" {
		for kp := range strings.SplitSeq(v, ",") {
			if trimmed := strings.TrimSpace(kp); trimmed != "" {
				keyPrefixes = append(keyPrefixes, trimmed)
			}
		}
	}

	// To preseve behavior prior to supporting environment variables:
	//
	// - Return nil when no keys or prefixes are set
	// - For a non-nil return, `keys` or `key_prefixes` should be
	//   nil if empty (versus a zero-value `KeyValueTags` struct)
	if len(keys) == 0 && len(keyPrefixes) == 0 {
		return nil
	}

	ignoreConfig := &tftags.IgnoreConfig{}
	if len(keys) > 0 {
		ignoreConfig.Keys = tftags.New(ctx, keys)
	}
	if len(keyPrefixes) > 0 {
		ignoreConfig.KeyPrefixes = tftags.New(ctx, keyPrefixes)
	}

	return ignoreConfig
}

func expandTagPolicyConfig(path cty.Path, severity string) (*tftags.TagPolicyConfig, diag.Diagnostics) {
	envSeverity := os.Getenv(tftags.TagPolicyComplianceEnvVar)
	switch {
	case severity != "" && severity != "disabled":
		return &tftags.TagPolicyConfig{Severity: severity}, validateTagPolicySeverity(path, severity)
	case envSeverity != "" && severity != "disabled":
		return &tftags.TagPolicyConfig{Severity: envSeverity}, validateTagPolicySeverityEnvVar(envSeverity)
	}

	return nil, nil
}

func validateTagPolicySeverity(path cty.Path, s string) diag.Diagnostics {
	var diags diag.Diagnostics
	switch s {
	case "error", "warning", "disabled":
		return diags
	}
	return append(diags, errs.NewInvalidValueAttributeError(path, `Must be one of "error", "warning", or "disabled"`))
}

const (
	summaryInvalidEnvironmentVariableValue = "Invalid environment variable value"
)

func validateTagPolicySeverityEnvVar(s string) diag.Diagnostics {
	var diags diag.Diagnostics
	switch s {
	case "error", "warning", "disabled":
		return diags
	}
	return append(diags, errs.NewErrorDiagnostic(
		summaryInvalidEnvironmentVariableValue,
		fmt.Sprintf(`%s must be one of "error", "warning", or "disabled"`, tftags.TagPolicyComplianceEnvVar),
	))
}
