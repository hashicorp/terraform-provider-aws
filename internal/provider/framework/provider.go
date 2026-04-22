// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log"
	"reflect"
	"slices"
	"sync"
	"unique"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	aschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	empemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/metaschema"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tffunction "github.com/hashicorp/terraform-provider-aws/internal/function"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	tfunique "github.com/hashicorp/terraform-provider-aws/internal/unique"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	resourceSchemasValidated sync.Once
)

var (
	_ provider.Provider                       = &frameworkProvider{}
	_ provider.ProviderWithActions            = &frameworkProvider{}
	_ provider.ProviderWithFunctions          = &frameworkProvider{}
	_ provider.ProviderWithEphemeralResources = &frameworkProvider{}
	_ provider.ProviderWithListResources      = &frameworkProvider{}
	_ provider.ProviderWithMetaSchema         = &frameworkProvider{}
)

type frameworkProvider struct {
	actions            []func() action.Action
	dataSources        []func() datasource.DataSource
	ephemeralResources []func() ephemeral.EphemeralResource
	listResources      []func() list.ListResource
	primary            interface{ Meta() any }
	resources          []func() resource.Resource
	servicePackages    iter.Seq[conns.ServicePackage]
}

// NewProvider returns a new, initialized Terraform Plugin Framework-style provider instance.
// The provider instance is fully configured once the `Configure` method has been called.
func NewProvider(ctx context.Context, primary interface{ Meta() any }) (provider.Provider, error) {
	log.Printf("Creating Terraform AWS Provider (Framework-style)...")

	provider := &frameworkProvider{
		actions:            make([]func() action.Action, 0),
		dataSources:        make([]func() datasource.DataSource, 0),
		ephemeralResources: make([]func() ephemeral.EphemeralResource, 0),
		primary:            primary,
		resources:          make([]func() resource.Resource, 0),
		servicePackages:    primary.Meta().(*conns.AWSClient).ServicePackages(ctx),
	}

	// Because we try and share resource schemas as much as possible,
	// we need to ensure that we only validate the resource schemas once.
	var err error
	resourceSchemasValidated.Do(func() {
		err = provider.validateResourceSchemas(ctx)
	})
	if err != nil {
		return nil, err
	}

	provider.initialize(ctx)

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
			"tag_policy_compliance": schema.StringAttribute{
				Optional: true,
				Description: `The severity with which to enforce organizational tagging policies on resources managed by this provider instance. ` +
					`At this time this only includes compliance with required tag keys by resource type. ` +
					`Valid values are "error", "warning", and "disabled". ` +
					`When unset or "disabled", tag policy compliance will not be enforced by the provider. ` +
					`Can also be configured with the ` + tftags.TagPolicyComplianceEnvVar + ` environment variable.`,
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
			"user_agent": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Product details to append to the User-Agent string sent in all AWS API calls.",
			},
		},
		Blocks: map[string]schema.Block{
			"assume_role": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"duration": schema.StringAttribute{
							CustomType:  timetypes.GoDurationType{},
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
							CustomType:  timetypes.GoDurationType{},
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

func (p *frameworkProvider) MetaSchema(ctx context.Context, req provider.MetaSchemaRequest, resp *provider.MetaSchemaResponse) {
	resp.Schema = metaschema.Schema{
		Attributes: map[string]metaschema.Attribute{
			"user_agent": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Product details to append to the User-Agent string sent in all AWS API calls.",
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
	response.ActionData = v
	response.ListResourceData = v
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

// Actions returns a slice of functions to instantiate each Action
// implementation.
//
// All actions must have unique type names.
func (p *frameworkProvider) Actions(ctx context.Context) []func() action.Action {
	return slices.Clone(p.actions)
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
		tffunction.NewUserAgentFunction,
	}
}

func (p *frameworkProvider) ListResources(_ context.Context) []func() list.ListResource {
	return slices.Clone(p.listResources)
}

// initialize is called from `New` to perform any Terraform Framework-style initialization.
func (p *frameworkProvider) initialize(ctx context.Context) {
	log.Printf("Initializing Terraform AWS Provider (Framework-style)...")

	for sp := range p.servicePackages {
		servicePackageName := sp.ServicePackageName()

		for _, dataSourceSpec := range sp.FrameworkDataSources(ctx) {
			p.dataSources = append(p.dataSources, func() datasource.DataSource { //nolint:contextcheck // must be a func()
				return newWrappedDataSource(dataSourceSpec, servicePackageName)
			})
		}

		if v, ok := sp.(conns.ServicePackageWithEphemeralResources); ok {
			for _, ephemeralResourceSpec := range v.EphemeralResources(ctx) {
				p.ephemeralResources = append(p.ephemeralResources, func() ephemeral.EphemeralResource { //nolint:contextcheck // must be a func()
					return newWrappedEphemeralResource(ephemeralResourceSpec, servicePackageName)
				})
			}
		}

		if v, ok := sp.(conns.ServicePackageWithFrameworkListResources); ok {
			for listResourceSpec := range v.FrameworkListResources(ctx) {
				p.listResources = append(p.listResources, func() list.ListResource { //nolint:contextcheck // must be a func()
					return newWrappedListResourceFramework(listResourceSpec, servicePackageName)
				})
			}
		}
		if v, ok := sp.(conns.ServicePackageWithSDKListResources); ok {
			for listResourceSpec := range v.SDKListResources(ctx) {
				p.listResources = append(p.listResources, func() list.ListResource { //nolint:contextcheck // must be a func()
					return newWrappedListResourceSDK(listResourceSpec, servicePackageName)
				})
			}
		}

		for _, resourceSpec := range sp.FrameworkResources(ctx) {
			p.resources = append(p.resources, func() resource.Resource { //nolint:contextcheck // must be a func()
				return newWrappedResource(resourceSpec, servicePackageName)
			})
		}

		if v, ok := sp.(conns.ServicePackageWithActions); ok {
			for _, actionSpec := range v.Actions(ctx) {
				p.actions = append(p.actions, func() action.Action { //nolint:contextcheck // must be a func()
					return newWrappedAction(actionSpec, servicePackageName)
				})
			}
		}
	}
}

// validateResourceSchemas is called from `New` to validate Terraform Plugin Framework-style resource schemas.
func (p *frameworkProvider) validateResourceSchemas(ctx context.Context) error {
	var errs []error

	for sp := range p.servicePackages {
		for _, dataSourceSpec := range sp.FrameworkDataSources(ctx) {
			typeName := dataSourceSpec.TypeName
			inner, err := dataSourceSpec.Factory(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("creating data source type (%s): %w", typeName, err))
				continue
			}

			schemaResponse := datasource.SchemaResponse{}
			inner.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)

			if err := validateSchemaRegionForDataSource(dataSourceSpec.Region, schemaResponse.Schema); err != nil {
				errs = append(errs, fmt.Errorf("data source type %q: %w", typeName, err))
				continue
			}

			if err := validateSchemaTagsForDataSource(dataSourceSpec.Tags, schemaResponse.Schema); err != nil {
				errs = append(errs, fmt.Errorf("data source type %q: %w", typeName, err))
				continue
			}
		}

		if v, ok := sp.(conns.ServicePackageWithEphemeralResources); ok {
			for _, ephemeralResourceSpec := range v.EphemeralResources(ctx) {
				typeName := ephemeralResourceSpec.TypeName
				inner, err := ephemeralResourceSpec.Factory(ctx)

				if err != nil {
					errs = append(errs, fmt.Errorf("creating ephemeral resource type (%s): %w", typeName, err))
					continue
				}

				schemaResponse := ephemeral.SchemaResponse{}
				inner.Schema(ctx, ephemeral.SchemaRequest{}, &schemaResponse)

				if err := validateSchemaRegionForEphemeralResource(ephemeralResourceSpec.Region, schemaResponse.Schema); err != nil {
					errs = append(errs, fmt.Errorf("ephemeral resource type %q: %w", typeName, err))
					continue
				}
			}
		}

		if v, ok := sp.(conns.ServicePackageWithActions); ok {
			for _, actionSpec := range v.Actions(ctx) {
				typeName := actionSpec.TypeName
				inner, err := actionSpec.Factory(ctx)

				if err != nil {
					errs = append(errs, fmt.Errorf("creating action type (%s): %w", typeName, err))
					continue
				}

				schemaResponse := action.SchemaResponse{}
				inner.Schema(ctx, action.SchemaRequest{}, &schemaResponse)

				if err := validateSchemaRegionForAction(actionSpec.Region, schemaResponse.Schema); err != nil {
					errs = append(errs, fmt.Errorf("action type %q: %w", typeName, err))
					continue
				}
			}
		}

		for _, resourceSpec := range sp.FrameworkResources(ctx) {
			typeName := resourceSpec.TypeName
			inner, err := resourceSpec.Factory(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("creating resource type (%s): %w", typeName, err))
				continue
			}

			schemaResponse := resource.SchemaResponse{}
			inner.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

			if err := validateSchemaRegionForResource(resourceSpec.Region, schemaResponse.Schema); err != nil {
				errs = append(errs, fmt.Errorf("resource type %q: %w", typeName, err))
				continue
			}

			if err := validateSchemaTagsForResource(resourceSpec.Tags, schemaResponse.Schema); err != nil {
				errs = append(errs, fmt.Errorf("resource type %q: %w", typeName, err))
				continue
			}

			if resourceSpec.Import.WrappedImport {
				if resourceSpec.Import.SetIDAttr {
					if _, ok := resourceSpec.Import.ImportID.(inttypes.FrameworkImportIDCreator); !ok {
						errs = append(errs, fmt.Errorf("resource type %q: importer sets `%s` attribute, but creator isn't configured", resourceSpec.TypeName, names.AttrID))
						continue
					}
				}

				if _, ok := inner.(framework.ImportByIdentityer); !ok {
					errs = append(errs, fmt.Errorf("resource type %q: cannot configure importer, does not implement %q", resourceSpec.TypeName, reflect.TypeFor[framework.ImportByIdentityer]()))
					continue
				}
			}
		}
	}

	return errors.Join(errs...)
}

func validateSchemaRegionForDataSource(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion], schema datasourceschema.Schema) error {
	if !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		if _, ok := schema.Attributes[names.AttrRegion]; ok {
			return fmt.Errorf("configured for enhanced regions but defines `%s` attribute in schema", names.AttrRegion)
		}
	}
	return nil
}

func validateSchemaRegionForEphemeralResource(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion], schema empemeralschema.Schema) error {
	if !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		if _, ok := schema.Attributes[names.AttrRegion]; ok {
			return fmt.Errorf("configured for enhanced regions but defines `%s` attribute in schema", names.AttrRegion)
		}
	}
	return nil
}

func validateSchemaRegionForAction(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion], schemaIface any) error {
	if !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		if schema, ok := schemaIface.(aschema.Schema); ok {
			if _, ok := schema.Attributes[names.AttrRegion]; ok {
				return fmt.Errorf("configured for enhanced regions but defines `%s` attribute in schema", names.AttrRegion)
			}
		}
	}
	return nil
}

func validateSchemaRegionForResource(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion], schema resourceschema.Schema) error {
	if !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		if _, ok := schema.Attributes[names.AttrRegion]; ok {
			return fmt.Errorf("configured for enhanced regions but defines `%s` attribute in schema", names.AttrRegion)
		}
	}
	return nil
}

func validateSchemaTagsForDataSource(tagsSpec unique.Handle[inttypes.ServicePackageResourceTags], schema datasourceschema.Schema) error {
	if !tfunique.IsHandleNil(tagsSpec) {
		if v, ok := schema.Attributes[names.AttrTags]; ok {
			if !v.IsComputed() {
				return fmt.Errorf("`%s` attribute must be Computed", names.AttrTags)
			}
		} else {
			return fmt.Errorf("configured for tags but no `%s` attribute defined in schema", names.AttrTags)
		}
	}
	return nil
}

func validateSchemaTagsForResource(tagsSpec unique.Handle[inttypes.ServicePackageResourceTags], schema resourceschema.Schema) error {
	if !tfunique.IsHandleNil(tagsSpec) {
		if v, ok := schema.Attributes[names.AttrTags]; ok {
			if v.IsComputed() {
				return fmt.Errorf("`%s` attribute cannot be Computed", names.AttrTags)
			}
		} else {
			return fmt.Errorf("configured for tags but no `%s` attribute defined in schema", names.AttrTags)
		}
		if v, ok := schema.Attributes[names.AttrTagsAll]; ok {
			if !v.IsComputed() {
				return fmt.Errorf("`%s` attribute must be Computed", names.AttrTagsAll)
			}
		} else {
			return fmt.Errorf("configured for tags but no `%s` attribute defined in schema", names.AttrTagsAll)
		}
	}
	return nil
}
