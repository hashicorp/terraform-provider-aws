// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tffunction "github.com/hashicorp/terraform-provider-aws/internal/function"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	tfunique "github.com/hashicorp/terraform-provider-aws/internal/unique"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	resourceSchemasValidated bool
)

var (
	_ provider.Provider                       = &frameworkProvider{}
	_ provider.ProviderWithFunctions          = &frameworkProvider{}
	_ provider.ProviderWithEphemeralResources = &frameworkProvider{}
)

type frameworkProvider struct {
	dataSources        []func() datasource.DataSource
	ephemeralResources []func() ephemeral.EphemeralResource
	primary            interface{ Meta() any }
	resources          []func() resource.Resource
	servicePackages    iter.Seq[conns.ServicePackage]
}

// NewProvider returns a new, initialized Terraform Plugin Framework-style provider instance.
// The provider instance is fully configured once the `Configure` method has been called.
func NewProvider(ctx context.Context, primary interface{ Meta() any }) (provider.Provider, error) {
	log.Printf("Creating Terraform AWS Provider (Framework-style)...")

	provider := &frameworkProvider{
		dataSources:        make([]func() datasource.DataSource, 0),
		ephemeralResources: make([]func() ephemeral.EphemeralResource, 0),
		primary:            primary,
		resources:          make([]func() resource.Resource, 0),
		servicePackages:    primary.Meta().(*conns.AWSClient).ServicePackages(ctx),
	}

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
		if err := provider.validateResourceSchemas(ctx); err != nil {
			return nil, err
		}

		resourceSchemasValidated = true
	}

	if err := provider.initialize(ctx); err != nil {
		return nil, err
	}

	return provider, nil
}

func (*frameworkProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "aws"
}

// Schema returns the schema for this provider's configuration.
func (*frameworkProvider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	// This schema must match exactly the Terraform Protocol v5 (Terraform Plugin SDK v2) provider's schema.
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				Optional:    true,
				Description: "The access key for API operations. You can retrieve this\nfrom the 'Security & Credentials' section of the AWS console.",
			},
			"allowed_account_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"custom_ca_bundle": schema.StringAttribute{
				Optional:    true,
				Description: "File containing custom root and intermediate certificates. Can also be configured using the `AWS_CA_BUNDLE` environment variable. (Setting `ca_bundle` in the shared config file is not supported.)",
			},
			"ec2_metadata_service_endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Address of the EC2 metadata service endpoint to use. Can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.",
			},
			"ec2_metadata_service_endpoint_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Protocol to use with EC2 metadata service endpoint.Valid values are `IPv4` and `IPv6`. Can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE` environment variable.",
			},
			"forbidden_account_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"http_proxy": schema.StringAttribute{
				Optional:    true,
				Description: "URL of a proxy to use for HTTP requests when accessing the AWS API. Can also be set using the `HTTP_PROXY` or `http_proxy` environment variables.",
			},
			"https_proxy": schema.StringAttribute{
				Optional:    true,
				Description: "URL of a proxy to use for HTTPS requests when accessing the AWS API. Can also be set using the `HTTPS_PROXY` or `https_proxy` environment variables.",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted, default value is `false`",
			},
			"max_retries": schema.Int64Attribute{
				Optional:    true,
				Description: "The maximum number of times an AWS API request is\nbeing executed. If the API request still fails, an error is\nthrown.",
			},
			"no_proxy": schema.StringAttribute{
				Optional:    true,
				Description: "Comma-separated list of hosts that should not use HTTP or HTTPS proxies. Can also be set using the `NO_PROXY` or `no_proxy` environment variables.",
			},
			"profile": schema.StringAttribute{
				Optional:    true,
				Description: "The profile for API operations. If not set, the default profile\ncreated with `aws configure` will be used.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "The region where AWS operations will take place. Examples\nare us-east-1, us-west-2, etc.", // lintignore:AWSAT003
			},
			"retry_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Specifies how retries are attempted. Valid values are `standard` and `adaptive`. Can also be configured using the `AWS_RETRY_MODE` environment variable.",
			},
			"s3_use_path_style": schema.BoolAttribute{
				Optional:    true,
				Description: "Set this to true to enable the request to use path-style addressing,\ni.e., https://s3.amazonaws.com/BUCKET/KEY. By default, the S3 client will\nuse virtual hosted bucket addressing when possible\n(https://BUCKET.s3.amazonaws.com/KEY). Specific to the Amazon S3 service.",
			},
			"s3_us_east_1_regional_endpoint": schema.StringAttribute{
				Optional: true,
				Description: "Specifies whether S3 API calls in the `us-east-1` region use the legacy global endpoint or a regional endpoint. " + //lintignore:AWSAT003
					"Valid values are `legacy` or `regional`. " +
					"Can also be configured using the `AWS_S3_US_EAST_1_REGIONAL_ENDPOINT` environment variable or the `s3_us_east_1_regional_endpoint` shared config file parameter",
			},
			"secret_key": schema.StringAttribute{
				Optional:    true,
				Description: "The secret key for API operations. You can retrieve this\nfrom the 'Security & Credentials' section of the AWS console.",
			},
			"shared_config_files": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of paths to shared config files. If not set, defaults to [~/.aws/config].",
			},
			"shared_credentials_files": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of paths to shared credentials files. If not set, defaults to [~/.aws/credentials].",
			},
			"skip_credentials_validation": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip the credentials validation via STS API. Used for AWS API implementations that do not have STS available/implemented.",
			},
			"skip_metadata_api_check": schema.StringAttribute{
				Optional:    true,
				Description: "Skip the AWS Metadata API check. Used for AWS API implementations that do not have a metadata api endpoint.",
			},
			"skip_region_validation": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip static validation of region name. Used by users of alternative AWS-like APIs or users w/ access to regions that are not public (yet).",
			},
			"skip_requesting_account_id": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip requesting the account ID. Used for AWS API implementations that do not have IAM/STS API and/or metadata API.",
			},
			"sts_region": schema.StringAttribute{
				Optional:    true,
				Description: "The region where AWS STS operations will take place. Examples\nare us-east-1 and us-west-2.", // lintignore:AWSAT003
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Description: "session token. A session token is only required if you are\nusing temporary security credentials.",
			},
			"token_bucket_rate_limiter_capacity": schema.Int64Attribute{
				Optional:    true,
				Description: "The capacity of the AWS SDK's token bucket rate limiter.",
			},
			"use_dualstack_endpoint": schema.BoolAttribute{
				Optional:    true,
				Description: "Resolve an endpoint with DualStack capability",
			},
			"use_fips_endpoint": schema.BoolAttribute{
				Optional:    true,
				Description: "Resolve an endpoint with FIPS capability",
			},
		},
		Blocks: map[string]schema.Block{
			"assume_role": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"duration": schema.StringAttribute{
							CustomType:  fwtypes.DurationType,
							Optional:    true,
							Description: "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
						},
						"external_id": schema.StringAttribute{
							Optional:    true,
							Description: "A unique identifier that might be required when you assume a role in another account.",
						},
						"policy": schema.StringAttribute{
							Optional:    true,
							Description: "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
						},
						"policy_arns": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
						},
						"role_arn": schema.StringAttribute{
							Optional:    true, // For historical reasons, we allow an empty `assume_role` block
							Description: "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
						},
						"session_name": schema.StringAttribute{
							Optional:    true,
							Description: "An identifier for the assumed role session.",
						},
						"source_identity": schema.StringAttribute{
							Optional:    true,
							Description: "Source identity specified by the principal assuming the role.",
						},
						"tags": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Assume role session tags.",
						},
						"transitive_tag_keys": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Assume role session tag keys to pass to any subsequent sessions.",
						},
					},
				},
			},
			"assume_role_with_web_identity": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"duration": schema.StringAttribute{
							CustomType:  fwtypes.DurationType,
							Optional:    true,
							Description: "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
						},
						"policy": schema.StringAttribute{
							Optional:    true,
							Description: "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
						},
						"policy_arns": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
						},
						"role_arn": schema.StringAttribute{
							Optional:    true, // For historical reasons, we allow an empty `assume_role_with_web_identity` block
							Description: "Amazon Resource Name (ARN) of an IAM Role to assume prior to making API calls.",
						},
						"session_name": schema.StringAttribute{
							Optional:    true,
							Description: "An identifier for the assumed role session.",
						},
						"web_identity_token": schema.StringAttribute{
							Optional: true,
						},
						"web_identity_token_file": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"default_tags": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: "Configuration block with settings to default resource tags across all resources.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tags": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Resource tags to default across all resources. " +
								"Can also be configured with environment variables like `" + tftags.DefaultTagsEnvVarPrefix + "<tag_name>`.",
						},
					},
				},
			},
			"endpoints": endpointsBlock(),
			"ignore_tags": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: "Configuration block with settings to ignore resource tags across all resources.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key_prefixes": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Resource tag key prefixes to ignore across all resources. " +
								"Can also be configured with the " + tftags.IgnoreTagsKeyPrefixesEnvVar + " environment variable.",
						},
						"keys": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Resource tag keys to ignore across all resources. " +
								"Can also be configured with the " + tftags.IgnoreTagsKeysEnvVar + " environment variable.",
						},
					},
				},
			},
		},
	}
}

// Configure is called at the beginning of the provider lifecycle, when
// Terraform sends to the provider the values the user specified in the
// provider configuration block.
func (p *frameworkProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	// Provider's parsed configuration (its instance state) is available through the primary provider's Meta() method.
	v := p.primary.Meta()
	response.DataSourceData = v
	response.ResourceData = v
	response.EphemeralResourceData = v
}

// DataSources returns a slice of functions to instantiate each DataSource
// implementation.
//
// All data sources must have unique type names.
func (p *frameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return slices.Clone(p.dataSources)
}

// Resources returns a slice of functions to instantiate each Resource
// implementation.
//
// All resources must have unique type names.
func (p *frameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return slices.Clone(p.resources)
}

// EphemeralResources returns a slice of functions to instantiate each Ephemeral Resource
// implementation.
//
// All ephemeral resources must have unique type names.
func (p *frameworkProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return slices.Clone(p.ephemeralResources)
}

// Functions returns a slice of functions to instantiate each Function
// implementation.
//
// The function type name is determined by the Function implementing
// the Metadata method. All functions must have unique names.
func (p *frameworkProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		tffunction.NewARNBuildFunction,
		tffunction.NewARNParseFunction,
		tffunction.NewTrimIAMRolePathFunction,
	}
}

// initialize is called from `New` to perform any Terraform Framework-style initialization.
func (p *frameworkProvider) initialize(ctx context.Context) error {
	log.Printf("Initializing Terraform AWS Provider (Framework-style)...")

	var errs []error

	for sp := range p.servicePackages {
		servicePackageName := sp.ServicePackageName()

		for _, v := range sp.FrameworkDataSources(ctx) {
			typeName := v.TypeName
			inner, err := v.Factory(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("creating data source (%s): %w", typeName, err))
				continue
			}

			var isRegionOverrideEnabled bool
			if v := v.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
				isRegionOverrideEnabled = true
			}

			var interceptors interceptorInvocations

			if isRegionOverrideEnabled {
				v := v.Region.Value()

				interceptors = append(interceptors, dataSourceInjectRegionAttribute())
				if v.IsValidateOverrideInPartition {
					interceptors = append(interceptors, dataSourceValidateRegion())
				}
				interceptors = append(interceptors, dataSourceSetRegionInState())
			}

			if !tfunique.IsHandleNil(v.Tags) {
				interceptors = append(interceptors, dataSourceTransparentTagging(v.Tags))
			}

			opts := wrappedDataSourceOptions{
				// bootstrapContext is run on all wrapped methods before any interceptors.
				bootstrapContext: func(ctx context.Context, getAttribute getAttributeFunc, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
					var diags diag.Diagnostics
					var overrideRegion string

					if isRegionOverrideEnabled && getAttribute != nil {
						var target types.String
						diags.Append(getAttribute(ctx, path.Root(names.AttrRegion), &target)...)
						if diags.HasError() {
							return ctx, diags
						}

						overrideRegion = target.ValueString()
					}

					ctx = conns.NewResourceContext(ctx, servicePackageName, v.Name, overrideRegion)
					if c != nil {
						ctx = tftags.NewContext(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx))
						ctx = c.RegisterLogger(ctx)
						ctx = fwflex.RegisterLogger(ctx)
					}

					return ctx, diags
				},
				interceptors: interceptors,
				typeName:     typeName,
			}
			p.dataSources = append(p.dataSources, func() datasource.DataSource {
				return newWrappedDataSource(inner, opts)
			})
		}

		if v, ok := sp.(conns.ServicePackageWithEphemeralResources); ok {
			for _, v := range v.EphemeralResources(ctx) {
				typeName := v.TypeName
				inner, err := v.Factory(ctx)

				if err != nil {
					errs = append(errs, fmt.Errorf("creating ephemeral resource (%s): %w", typeName, err))
					continue
				}

				var isRegionOverrideEnabled bool
				if v := v.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
					isRegionOverrideEnabled = true
				}

				var interceptors interceptorInvocations

				if isRegionOverrideEnabled {
					v := v.Region.Value()

					interceptors = append(interceptors, ephemeralResourceInjectRegionAttribute())
					if v.IsValidateOverrideInPartition {
						interceptors = append(interceptors, ephemeralResourceValidateRegion())
					}
					interceptors = append(interceptors, ephemeralResourceSetRegionInResult())
				}

				opts := wrappedEphemeralResourceOptions{
					// bootstrapContext is run on all wrapped methods before any interceptors.
					bootstrapContext: func(ctx context.Context, getAttribute getAttributeFunc, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
						var diags diag.Diagnostics
						var overrideRegion string

						if isRegionOverrideEnabled && getAttribute != nil {
							var target types.String
							diags.Append(getAttribute(ctx, path.Root(names.AttrRegion), &target)...)
							if diags.HasError() {
								return ctx, diags
							}

							overrideRegion = target.ValueString()
						}

						ctx = conns.NewResourceContext(ctx, servicePackageName, v.Name, overrideRegion)
						if c != nil {
							ctx = c.RegisterLogger(ctx)
							ctx = fwflex.RegisterLogger(ctx)
							ctx = logging.MaskSensitiveValuesByKey(ctx, logging.HTTPKeyRequestBody, logging.HTTPKeyResponseBody)
						}
						return ctx, diags
					},
					interceptors: interceptors,
					typeName:     v.TypeName,
				}
				p.ephemeralResources = append(p.ephemeralResources, func() ephemeral.EphemeralResource {
					return newWrappedEphemeralResource(inner, opts)
				})
			}
		}

		for _, res := range sp.FrameworkResources(ctx) {
			typeName := res.TypeName
			inner, err := res.Factory(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("creating resource (%s): %w", typeName, err))
				continue
			}

			var isRegionOverrideEnabled bool
			if v := res.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
				isRegionOverrideEnabled = true
			}

			var interceptors interceptorInvocations

			if isRegionOverrideEnabled {
				v := res.Region.Value()

				interceptors = append(interceptors, resourceInjectRegionAttribute())
				if v.IsValidateOverrideInPartition {
					interceptors = append(interceptors, resourceValidateRegion())
				}
				interceptors = append(interceptors, resourceDefaultRegion())
				interceptors = append(interceptors, resourceForceNewIfRegionChanges())
				interceptors = append(interceptors, resourceSetRegionInState())
				if res.Identity.HasInherentRegion() {
					interceptors = append(interceptors, resourceImportRegionNoDefault())
				} else {
					interceptors = append(interceptors, resourceImportRegion())
				}
			}

			if !tfunique.IsHandleNil(res.Tags) {
				interceptors = append(interceptors, resourceTransparentTagging(res.Tags))
			}

			if res.Import.WrappedImport {
				switch v := inner.(type) {
				case framework.ImportByIdentityer:
					v.SetIdentitySpec(res.Identity)

				default:
					errs = append(errs, fmt.Errorf("resource type %s: cannot configure importer", typeName))
					continue
				}
			}

			opts := wrappedResourceOptions{
				// bootstrapContext is run on all wrapped methods before any interceptors.
				bootstrapContext: func(ctx context.Context, getAttribute getAttributeFunc, c *conns.AWSClient) (context.Context, diag.Diagnostics) {
					var diags diag.Diagnostics
					var overrideRegion string

					if isRegionOverrideEnabled && getAttribute != nil {
						var target types.String
						diags.Append(getAttribute(ctx, path.Root(names.AttrRegion), &target)...)
						if diags.HasError() {
							return ctx, diags
						}

						overrideRegion = target.ValueString()
					}

					ctx = conns.NewResourceContext(ctx, servicePackageName, res.Name, overrideRegion)
					if c != nil {
						ctx = tftags.NewContext(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx))
						ctx = c.RegisterLogger(ctx)
						ctx = fwflex.RegisterLogger(ctx)
					}

					return ctx, diags
				},
				interceptors: interceptors,
				typeName:     typeName,
			}
			if len(res.Identity.Attributes) > 0 {
				opts.identity = res.Identity
				opts.interceptors = append(opts.interceptors, newIdentityInterceptor(res.Identity.Attributes))
			}

			p.resources = append(p.resources, func() resource.Resource {
				return newWrappedResource(inner, opts)
			})
		}
	}

	return errors.Join(errs...)
}

// validateResourceSchemas is called from `New` to validate Terraform Plugin Framework-style resource schemas.
func (p *frameworkProvider) validateResourceSchemas(ctx context.Context) error {
	var errs []error

	for sp := range p.servicePackages {
		for _, v := range sp.FrameworkDataSources(ctx) {
			typeName := v.TypeName
			ds, err := v.Factory(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("creating data source (%s): %w", typeName, err))
				continue
			}

			schemaResponse := datasource.SchemaResponse{}
			ds.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)

			if v := v.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
				if _, ok := schemaResponse.Schema.Attributes[names.AttrRegion]; ok {
					errs = append(errs, fmt.Errorf("`%s` attribute is defined: %s data source", names.AttrRegion, typeName))
					continue
				}
			}

			if !tfunique.IsHandleNil(v.Tags) {
				// The data source has opted in to transparent tagging.
				// Ensure that the schema look OK.
				if v, ok := schemaResponse.Schema.Attributes[names.AttrTags]; ok {
					if !v.IsComputed() {
						errs = append(errs, fmt.Errorf("`%s` attribute must be Computed: %s data source", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s data source", names.AttrTags, typeName))
					continue
				}
			}
		}

		if v, ok := sp.(conns.ServicePackageWithEphemeralResources); ok {
			for _, v := range v.EphemeralResources(ctx) {
				typeName := v.TypeName
				er, err := v.Factory(ctx)

				if err != nil {
					errs = append(errs, fmt.Errorf("creating ephemeral resource (%s): %w", typeName, err))
					continue
				}

				schemaResponse := ephemeral.SchemaResponse{}
				er.Schema(ctx, ephemeral.SchemaRequest{}, &schemaResponse)

				if v := v.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
					if _, ok := schemaResponse.Schema.Attributes[names.AttrRegion]; ok {
						errs = append(errs, fmt.Errorf("`%s` attribute is defined: %s ephemeral resource", names.AttrRegion, typeName))
						continue
					}
				}
			}
		}

		for _, v := range sp.FrameworkResources(ctx) {
			typeName := v.TypeName
			r, err := v.Factory(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("creating resource (%s): %w", typeName, err))
				continue
			}

			schemaResponse := resource.SchemaResponse{}
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

			if v := v.Region; !tfunique.IsHandleNil(v) && v.Value().IsOverrideEnabled {
				if _, ok := schemaResponse.Schema.Attributes[names.AttrRegion]; ok {
					errs = append(errs, fmt.Errorf("`%s` attribute is defined: %s resource", names.AttrRegion, typeName))
					continue
				}
			}

			if !tfunique.IsHandleNil(v.Tags) {
				// The resource has opted in to transparent tagging.
				// Ensure that the schema look OK.
				if v, ok := schemaResponse.Schema.Attributes[names.AttrTags]; ok {
					if v.IsComputed() {
						errs = append(errs, fmt.Errorf("`%s` attribute cannot be Computed: %s resource", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s resource", names.AttrTags, typeName))
					continue
				}
				if v, ok := schemaResponse.Schema.Attributes[names.AttrTagsAll]; ok {
					if !v.IsComputed() {
						errs = append(errs, fmt.Errorf("`%s` attribute must be Computed: %s resource", names.AttrTagsAll, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s resource", names.AttrTagsAll, typeName))
					continue
				}
			}
		}
	}

	return errors.Join(errs...)
}
