// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
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
	log.Printf("Initializing Terraform AWS Provider...")

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

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		return configure(ctx, provider, d)
	}

	var errs []error
	servicePackageMap := make(map[string]conns.ServicePackage)

	for _, sp := range servicePackages(ctx) {
		servicePackageName := sp.ServicePackageName()
		servicePackageMap[servicePackageName] = sp

		for _, v := range sp.SDKDataSources(ctx) {
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
					when:        Before | After,
					why:         Read,
					interceptor: newTagsDataSourceInterceptor(v.Tags),
				})
			}

			opts := wrappedDataSourceOptions{
				bootstrapContext: func(ctx context.Context, _ getAttributeFunc, meta any) (context.Context, diag.Diagnostics) {
					var diags diag.Diagnostics

					ctx = conns.NewDataSourceContext(ctx, servicePackageName, v.Name)
					if v, ok := meta.(*conns.AWSClient); ok {
						ctx = tftags.NewContext(ctx, v.DefaultTagsConfig(ctx), v.IgnoreTagsConfig(ctx))
						ctx = v.RegisterLogger(ctx)
					}

					return ctx, diags
				},
				interceptors: interceptors,
				typeName:     typeName,
			}
			wrapDataSource(r, opts)
			provider.DataSourcesMap[typeName] = r
		}

		for _, v := range sp.SDKResources(ctx) {
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

			var customizeDiffFuncs []schema.CustomizeDiffFunc
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

				customizeDiffFuncs = append(customizeDiffFuncs, setTagsAll)
				interceptors = append(interceptors, interceptorItem{
					when:        Before | After | Finally,
					why:         Create | Read | Update,
					interceptor: newTagsResourceInterceptor(v.Tags),
				})
			}

			opts := wrappedResourceOptions{
				// bootstrapContext is run on all wrapped methods before any interceptors.
				bootstrapContext: func(ctx context.Context, _ getAttributeFunc, meta any) (context.Context, diag.Diagnostics) {
					var diags diag.Diagnostics

					ctx = conns.NewResourceContext(ctx, servicePackageName, v.Name)
					if v, ok := meta.(*conns.AWSClient); ok {
						ctx = tftags.NewContext(ctx, v.DefaultTagsConfig(ctx), v.IgnoreTagsConfig(ctx))
						ctx = v.RegisterLogger(ctx)
					}

					return ctx, diags
				},
				customizeDiffFuncs: customizeDiffFuncs,
				interceptors:       interceptors,
				typeName:           typeName,
			}
			wrapResource(r, opts)
			provider.ResourcesMap[typeName] = r
		}
	}

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	// Set the provider Meta (instance data) here.
	// It will be overwritten by the result of the call to ConfigureContextFunc,
	// but can be used pre-configuration by other (non-primary) provider servers.
	var c *conns.AWSClient
	if v, ok := provider.Meta().(*conns.AWSClient); ok {
		c = v
	} else {
		c = new(conns.AWSClient)
	}
	c.SetServicePackages(ctx, servicePackageMap)
	provider.SetMeta(c)

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

	var c *conns.AWSClient
	if v, ok := provider.Meta().(*conns.AWSClient); ok {
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
		for _, k := range strings.Split(v, ",") {
			if trimmed := strings.TrimSpace(k); trimmed != "" {
				keys = append(keys, trimmed)
			}
		}
	}

	if v := os.Getenv(tftags.IgnoreTagsKeyPrefixesEnvVar); v != "" {
		for _, kp := range strings.Split(v, ",") {
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
