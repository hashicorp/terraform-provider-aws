// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// New returns a new, initialized Terraform Plugin Framework-style provider instance.
// The provider instance is fully configured once the `Configure` method has been called.
func New(primary interface{ Meta() interface{} }) provider.Provider {
	return &fwprovider{
		Primary: primary,
	}
}

type fwprovider struct {
	Primary interface{ Meta() interface{} }
}

func (p *fwprovider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "aws"
}

// Schema returns the schema for this provider's configuration.
func (p *fwprovider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	// This schema must match exactly the Terraform Protocol v5 (Terraform Plugin SDK v2) provider's schema.
	resp.Schema = schema.Schema{
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
							Optional:    true,
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
							Optional:    true,
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
							Description: "Resource tags to default across all resources",
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
							Description: "Resource tag key prefixes to ignore across all resources.",
						},
						"keys": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "Resource tag keys to ignore across all resources.",
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
func (p *fwprovider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	// Provider's parsed configuration (its instance state) is available through the primary provider's Meta() method.
	v := p.Primary.Meta()
	response.DataSourceData = v
	response.ResourceData = v
}

// DataSources returns a slice of functions to instantiate each DataSource
// implementation.
//
// The data source type name is determined by the DataSource implementing
// the Metadata method. All data sources must have unique names.
func (p *fwprovider) DataSources(ctx context.Context) []func() datasource.DataSource {
	var errs []error
	var dataSources []func() datasource.DataSource

	for n, sp := range p.Primary.Meta().(*conns.AWSClient).ServicePackages {
		servicePackageName := sp.ServicePackageName()

		for _, v := range sp.FrameworkDataSources(ctx) {
			v := v
			inner, err := v.Factory(ctx)

			if err != nil {
				tflog.Warn(ctx, "creating data source", map[string]interface{}{
					"service_package_name": n,
					"error":                err.Error(),
				})

				continue
			}

			metadataResponse := datasource.MetadataResponse{}
			inner.Metadata(ctx, datasource.MetadataRequest{}, &metadataResponse)
			typeName := metadataResponse.TypeName

			// bootstrapContext is run on all wrapped methods before any interceptors.
			bootstrapContext := func(ctx context.Context, meta *conns.AWSClient) context.Context {
				ctx = conns.NewDataSourceContext(ctx, servicePackageName, v.Name)
				if meta != nil {
					ctx = tftags.NewContext(ctx, meta.DefaultTagsConfig, meta.IgnoreTagsConfig)
					ctx = meta.RegisterLogger(ctx)
				}

				return ctx
			}
			interceptors := dataSourceInterceptors{}

			if v.Tags != nil {
				// The data source has opted in to transparent tagging.
				// Ensure that the schema look OK.
				schemaResponse := datasource.SchemaResponse{}
				inner.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)

				if v, ok := schemaResponse.Schema.Attributes[names.AttrTags]; ok {
					if !v.IsComputed() {
						errs = append(errs, fmt.Errorf("`%s` attribute must be Computed: %s", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s", names.AttrTags, typeName))
					continue
				}

				interceptors = append(interceptors, tagsDataSourceInterceptor{tags: v.Tags})
			}

			dataSources = append(dataSources, func() datasource.DataSource {
				return newWrappedDataSource(bootstrapContext, inner, interceptors)
			})
		}
	}

	if err := errors.Join(errs...); err != nil {
		tflog.Warn(ctx, "registering data sources", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return dataSources
}

// Resources returns a slice of functions to instantiate each Resource
// implementation.
//
// The resource type name is determined by the Resource implementing
// the Metadata method. All resources must have unique names.
func (p *fwprovider) Resources(ctx context.Context) []func() resource.Resource {
	var errs []error
	var resources []func() resource.Resource

	for _, sp := range p.Primary.Meta().(*conns.AWSClient).ServicePackages {
		servicePackageName := sp.ServicePackageName()

		for _, v := range sp.FrameworkResources(ctx) {
			v := v
			inner, err := v.Factory(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("creating resource: %w", err))
				continue
			}

			metadataResponse := resource.MetadataResponse{}
			inner.Metadata(ctx, resource.MetadataRequest{}, &metadataResponse)
			typeName := metadataResponse.TypeName

			// bootstrapContext is run on all wrapped methods before any interceptors.
			bootstrapContext := func(ctx context.Context, meta *conns.AWSClient) context.Context {
				ctx = conns.NewResourceContext(ctx, servicePackageName, v.Name)
				if meta != nil {
					ctx = tftags.NewContext(ctx, meta.DefaultTagsConfig, meta.IgnoreTagsConfig)
					ctx = meta.RegisterLogger(ctx)
				}

				return ctx
			}
			interceptors := resourceInterceptors{}

			if v.Tags != nil {
				// The resource has opted in to transparent tagging.
				// Ensure that the schema look OK.
				schemaResponse := resource.SchemaResponse{}
				inner.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

				if v, ok := schemaResponse.Schema.Attributes[names.AttrTags]; ok {
					if v.IsComputed() {
						errs = append(errs, fmt.Errorf("`%s` attribute cannot be Computed: %s", names.AttrTags, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s", names.AttrTags, typeName))
					continue
				}
				if v, ok := schemaResponse.Schema.Attributes[names.AttrTagsAll]; ok {
					if !v.IsComputed() {
						errs = append(errs, fmt.Errorf("`%s` attribute must be Computed: %s", names.AttrTagsAll, typeName))
						continue
					}
				} else {
					errs = append(errs, fmt.Errorf("no `%s` attribute defined in schema: %s", names.AttrTagsAll, typeName))
					continue
				}

				interceptors = append(interceptors, tagsResourceInterceptor{tags: v.Tags})
			}

			resources = append(resources, func() resource.Resource {
				return newWrappedResource(bootstrapContext, inner, interceptors)
			})
		}
	}

	if err := errors.Join(errs...); err != nil {
		tflog.Warn(ctx, "registering resources", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return resources
}

func endpointsBlock() schema.SetNestedBlock {
	endpointsAttributes := make(map[string]schema.Attribute)

	for _, serviceKey := range names.Aliases() {
		endpointsAttributes[serviceKey] = schema.StringAttribute{
			Optional:    true,
			Description: "Use this to override the default service endpoint URL",
		}
	}

	return schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: endpointsAttributes,
		},
	}
}
