// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// New returns a new, initialized Terraform Plugin SDK v2-style provider instance.
// The provider instance is fully configured once the `ConfigureContextFunc` has been called.
func New(ctx context.Context) (*schema.Provider, error) {
	provider := &schema.Provider{
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
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Resource tags to default across all resources",
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
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Resource tag keys to ignore across all resources.",
						},
						"key_prefixes": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Resource tag key prefixes to ignore across all resources.",
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
		},

		// Data sources and resources implemented using Terraform Plugin SDK
		// should use the @SDKDataSource and @SDKResource function-level annotations
		// rather than adding directly to these maps.
		DataSourcesMap: make(map[string]*schema.Resource),
		ResourcesMap:   make(map[string]*schema.Resource),
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return configure(ctx, provider, d)
	}

	var errs []error
	servicePackageMap := make(map[string]conns.ServicePackage)

	for _, sp := range servicePackages(ctx) {
		servicePackageName := sp.ServicePackageName()
		servicePackageMap[servicePackageName] = sp

		for _, v := range sp.SDKDataSources(ctx) {
			v := v
			typeName := v.TypeName

			if _, ok := provider.DataSourcesMap[typeName]; ok {
				errs = append(errs, fmt.Errorf("duplicate data source: %s", typeName))
				continue
			}

			r := v.Factory()

			// Ensure that the correct CRUD handler variants are used.
			if r.Read != nil || r.ReadContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Read handler variant: %s", typeName))
				continue
			}

			// bootstrapContext is run on all wrapped methods before any interceptors.
			bootstrapContext := func(ctx context.Context, meta any) context.Context {
				ctx = conns.NewDataSourceContext(ctx, servicePackageName, v.Name)
				if v, ok := meta.(*conns.AWSClient); ok {
					ctx = tftags.NewContext(ctx, v.DefaultTagsConfig, v.IgnoreTagsConfig)
					ctx = v.RegisterLogger(ctx)
				}

				return ctx
			}
			interceptors := interceptorItems{}

			if v.Tags != nil {
				schema := r.SchemaMap()

				// The data source has opted in to transparent tagging.
				// Ensure that the schema look OK.
				if v, ok := schema[names.AttrTags]; ok {
					if !v.Computed {
						errs = append(errs, fmt.Errorf("`%s` attribute must be Computed: %s", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s", names.AttrTags, typeName))
					continue
				}

				interceptors = append(interceptors, interceptorItem{
					when: Before | After,
					why:  Read,
					interceptor: tagsDataSourceInterceptor{
						tags: v.Tags,
					},
				})
			}

			ds := &wrappedDataSource{
				bootstrapContext: bootstrapContext,
				interceptors:     interceptors,
			}

			if v := r.ReadWithoutTimeout; v != nil {
				r.ReadWithoutTimeout = ds.Read(v)
			}

			provider.DataSourcesMap[typeName] = r
		}

		for _, v := range sp.SDKResources(ctx) {
			v := v
			typeName := v.TypeName

			if _, ok := provider.ResourcesMap[typeName]; ok {
				errs = append(errs, fmt.Errorf("duplicate resource: %s", typeName))
				continue
			}

			r := v.Factory()

			// Ensure that the correct CRUD handler variants are used.
			if r.Create != nil || r.CreateContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Create handler variant: %s", typeName))
				continue
			}
			if r.Read != nil || r.ReadContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Read handler variant: %s", typeName))
				continue
			}
			if r.Update != nil || r.UpdateContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Update handler variant: %s", typeName))
				continue
			}
			if r.Delete != nil || r.DeleteContext != nil {
				errs = append(errs, fmt.Errorf("incorrect Delete handler variant: %s", typeName))
				continue
			}

			// bootstrapContext is run on all wrapped methods before any interceptors.
			bootstrapContext := func(ctx context.Context, meta any) context.Context {
				ctx = conns.NewResourceContext(ctx, servicePackageName, v.Name)
				if v, ok := meta.(*conns.AWSClient); ok {
					ctx = tftags.NewContext(ctx, v.DefaultTagsConfig, v.IgnoreTagsConfig)
					ctx = v.RegisterLogger(ctx)
				}

				return ctx
			}
			interceptors := interceptorItems{}

			if v.Tags != nil {
				schema := r.SchemaMap()

				// The resource has opted in to transparent tagging.
				// Ensure that the schema look OK.
				if v, ok := schema[names.AttrTags]; ok {
					if v.Computed {
						errs = append(errs, fmt.Errorf("`%s` attribute cannot be Computed: %s", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s", names.AttrTags, typeName))
					continue
				}
				if v, ok := schema[names.AttrTagsAll]; ok {
					if !v.Computed {
						errs = append(errs, fmt.Errorf("`%s` attribute must be Computed: %s", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s", names.AttrTagsAll, typeName))
					continue
				}

				interceptors = append(interceptors, interceptorItem{
					when: Before | After | Finally,
					why:  Create | Read | Update,
					interceptor: tagsResourceInterceptor{
						tags:       v.Tags,
						updateFunc: tagsUpdateFunc,
						readFunc:   tagsReadFunc,
					},
				})
			}

			rs := &wrappedResource{
				bootstrapContext: bootstrapContext,
				interceptors:     interceptors,
			}

			if v := r.CreateWithoutTimeout; v != nil {
				r.CreateWithoutTimeout = rs.Create(v)
			}
			if v := r.ReadWithoutTimeout; v != nil {
				r.ReadWithoutTimeout = rs.Read(v)
			}
			if v := r.UpdateWithoutTimeout; v != nil {
				r.UpdateWithoutTimeout = rs.Update(v)
			}
			if v := r.DeleteWithoutTimeout; v != nil {
				r.DeleteWithoutTimeout = rs.Delete(v)
			}
			if v := r.Importer; v != nil {
				if v := v.StateContext; v != nil {
					r.Importer.StateContext = rs.State(v)
				}
			}
			if v := r.CustomizeDiff; v != nil {
				r.CustomizeDiff = rs.CustomizeDiff(v)
			}
			for _, stateUpgrader := range r.StateUpgraders {
				if v := stateUpgrader.Upgrade; v != nil {
					stateUpgrader.Upgrade = rs.StateUpgrade(v)
				}
			}

			provider.ResourcesMap[typeName] = r
		}
	}

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	// Set the provider Meta (instance data) here.
	// It will be overwritten by the result of the call to ConfigureContextFunc,
	// but can be used pre-configuration by other (non-primary) provider servers.
	var meta *conns.AWSClient
	if v, ok := provider.Meta().(*conns.AWSClient); ok {
		meta = v
	} else {
		meta = new(conns.AWSClient)
	}
	meta.ServicePackages = servicePackageMap
	provider.SetMeta(meta)

	return provider, nil
}

// configure ensures that the provider is fully configured.
func configure(ctx context.Context, provider *schema.Provider, d *schema.ResourceData) (*conns.AWSClient, diag.Diagnostics) {
	var diags diag.Diagnostics

	terraformVersion := provider.TerraformVersion
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
		config.S3USEast1RegionalEndpoint = conns.NormalizeS3USEast1RegionalEndpoint(v)
	}

	if v, ok := d.GetOk("allowed_account_ids"); ok && v.(*schema.Set).Len() > 0 {
		config.AllowedAccountIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("assume_role"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.AssumeRole = expandAssumeRole(ctx, v.([]interface{})[0].(map[string]interface{}))
		tflog.Info(ctx, "assume_role configuration set", map[string]any{
			"tf_aws.assume_role.role_arn":        config.AssumeRole.RoleARN,
			"tf_aws.assume_role.session_name":    config.AssumeRole.SessionName,
			"tf_aws.assume_role.external_id":     config.AssumeRole.ExternalID,
			"tf_aws.assume_role.source_identity": config.AssumeRole.SourceIdentity,
		})
	}

	if v, ok := d.GetOk("assume_role_with_web_identity"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.AssumeRoleWithWebIdentity = expandAssumeRoleWithWebIdentity(ctx, v.([]interface{})[0].(map[string]interface{}))
		tflog.Info(ctx, "assume_role_with_web_identity configuration set", map[string]any{
			"tf_aws.assume_role_with_web_identity.role_arn":     config.AssumeRoleWithWebIdentity.RoleARN,
			"tf_aws.assume_role_with_web_identity.session_name": config.AssumeRoleWithWebIdentity.SessionName,
		})
	}

	if v, ok := d.GetOk("default_tags"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.DefaultTagsConfig = expandDefaultTags(ctx, v.([]interface{})[0].(map[string]interface{}))
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

	if v, ok := d.GetOk("ignore_tags"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.IgnoreTagsConfig = expandIgnoreTags(ctx, v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("max_retries"); ok {
		config.MaxRetries = v.(int)
	}

	if v, ok := d.GetOk("shared_credentials_files"); ok && len(v.([]interface{})) > 0 {
		config.SharedCredentialsFiles = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("shared_config_files"); ok && len(v.([]interface{})) > 0 {
		config.SharedConfigFiles = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, null, _ := nullable.Bool(d.Get("skip_metadata_api_check").(string)).ValueBool(); !null {
		if v {
			config.EC2MetadataServiceEnableState = imds.ClientDisabled
		} else {
			config.EC2MetadataServiceEnableState = imds.ClientEnabled
		}
	}

	var meta *conns.AWSClient
	if v, ok := provider.Meta().(*conns.AWSClient); ok {
		meta = v
	} else {
		meta = new(conns.AWSClient)
	}
	meta, ds := config.ConfigureProvider(ctx, meta)
	diags = append(diags, ds...)

	if diags.HasError() {
		return nil, diags
	}

	return meta, diags
}

func assumeRoleSchema() *schema.Schema {
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
					Optional:     true,
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
					Optional:     true,
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

func endpointsSchema() *schema.Schema {
	endpointsAttributes := make(map[string]*schema.Schema)

	for _, serviceKey := range names.Aliases() {
		endpointsAttributes[serviceKey] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: "Use this to override the default service endpoint URL",
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: endpointsAttributes,
		},
	}
}

func expandAssumeRole(_ context.Context, tfMap map[string]interface{}) *awsbase.AssumeRole {
	if tfMap == nil {
		return nil
	}

	assumeRole := awsbase.AssumeRole{}

	if v, ok := tfMap["duration"].(string); ok && v != "" {
		duration, _ := time.ParseDuration(v)
		assumeRole.Duration = duration
	}

	if v, ok := tfMap["external_id"].(string); ok && v != "" {
		assumeRole.ExternalID = v
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

	if v, ok := tfMap["source_identity"].(string); ok && v != "" {
		assumeRole.SourceIdentity = v
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		assumeRole.Tags = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["transitive_tag_keys"].(*schema.Set); ok && v.Len() > 0 {
		assumeRole.TransitiveTagKeys = flex.ExpandStringValueSet(v)
	}

	return &assumeRole
}

func expandAssumeRoleWithWebIdentity(_ context.Context, tfMap map[string]interface{}) *awsbase.AssumeRoleWithWebIdentity {
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

func expandDefaultTags(ctx context.Context, tfMap map[string]interface{}) *tftags.DefaultConfig {
	if tfMap == nil {
		return nil
	}

	defaultConfig := &tftags.DefaultConfig{}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok {
		defaultConfig.Tags = tftags.New(ctx, v)
	}

	return defaultConfig
}

func expandIgnoreTags(ctx context.Context, tfMap map[string]interface{}) *tftags.IgnoreConfig {
	if tfMap == nil {
		return nil
	}

	ignoreConfig := &tftags.IgnoreConfig{}

	if v, ok := tfMap["keys"].(*schema.Set); ok {
		ignoreConfig.Keys = tftags.New(ctx, v.List())
	}

	if v, ok := tfMap["key_prefixes"].(*schema.Set); ok {
		ignoreConfig.KeyPrefixes = tftags.New(ctx, v.List())
	}

	return ignoreConfig
}

func expandEndpoints(_ context.Context, tfList []interface{}) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	endpointsPath := cty.GetAttrPath("endpoints")

	if l := len(tfList); l > 1 {
		diags = append(diags, errs.NewAttributeWarningDiagnostic(
			endpointsPath,
			"Invalid Attribute Value",
			fmt.Sprintf("Attribute %q should have at most 1 element, got %d."+
				"\n\nThis will be an error in a future release.",
				errs.PathString(endpointsPath), l),
		))
	}

	endpoints := make(map[string]string)

	for i, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		elementPath := endpointsPath.IndexInt(i)

		for _, endpoint := range names.Endpoints() {
			pkg := endpoint.ProviderPackage

			if len(endpoint.Aliases) > 0 {
				count := 0
				attrs := []string{pkg}
				if tfMap[pkg] != "" {
					count++
				}
				for _, alias := range endpoint.Aliases {
					attrs = append(attrs, alias)
					if tfMap[alias] != "" {
						count++
					}
				}

				if count > 1 {
					diags = append(diags, ConflictingEndpointsWarningDiag(elementPath, attrs...))
				}
			}

			if endpoints[pkg] == "" {
				if v := tfMap[pkg].(string); v != "" {
					endpoints[pkg] = v
				} else {
					for _, alias := range endpoint.Aliases {
						if v := tfMap[alias].(string); v != "" {
							endpoints[pkg] = v
							break
						}
					}
				}
			}
		}
	}
	if diags.HasError() {
		return nil, diags
	}

	for _, pkg := range names.ProviderPackages() {
		if endpoints[pkg] != "" {
			continue
		}

		// We only need to handle the services with custom envvars here before we hand off to `aws-sdk-go-base`
		tfAwsEnvVar := names.TFAWSEnvVar(pkg)
		deprecatedEnvVar := names.DeprecatedEnvVar(pkg)
		if tfAwsEnvVar == "" && deprecatedEnvVar == "" {
			continue
		}

		awsEnvVar := names.AWSServiceEnvVar(pkg)
		if awsEnvVar != "" {
			if v := os.Getenv(awsEnvVar); v != "" {
				endpoints[pkg] = v
				continue
			}
		}

		if tfAwsEnvVar != "" {
			if v := os.Getenv(tfAwsEnvVar); v != "" {
				diags = append(diags, DeprecatedEnvVarDiag(tfAwsEnvVar, awsEnvVar))
				endpoints[pkg] = v
				continue
			}
		}

		if deprecatedEnvVar != "" {
			if v := os.Getenv(deprecatedEnvVar); v != "" {
				diags = append(diags, DeprecatedEnvVarDiag(deprecatedEnvVar, awsEnvVar))
				endpoints[pkg] = v
				continue
			}
		}
	}

	return endpoints, diags
}

func DeprecatedEnvVarDiag(envvar, replacement string) diag.Diagnostic {
	return errs.NewWarningDiagnostic(
		"Deprecated Environment Variable",
		fmt.Sprintf(`The environment variable "%s" is deprecated. Use environment variable "%s" instead.`, envvar, replacement),
	)
}

func ConflictingEndpointsWarningDiag(elementPath cty.Path, attrs ...string) diag.Diagnostic {
	attrPaths := make([]string, len(attrs))
	for i, attr := range attrs {
		path := elementPath.GetAttr(attr)
		attrPaths[i] = `"` + errs.PathString(path) + `"`
	}
	return errs.NewAttributeWarningDiagnostic(
		elementPath,
		"Invalid Attribute Combination",
		fmt.Sprintf("Only one of the following attributes should be set: %s"+
			"\n\nThis will be an error in a future release.",
			strings.Join(attrPaths, ", ")),
	)
}
