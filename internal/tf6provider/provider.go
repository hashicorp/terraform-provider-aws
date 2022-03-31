package tf6provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func New() tfsdk.Provider {
	return &provider{}
}

type config struct {
	AccessKey                      types.String `tfsdk:"access_key"`
	AssumeRole                     types.List   `tfsdk:"assume_role"`
	AllowedAccountIDs              types.Set    `tfsdk:"allowed_account_ids"`
	CustomCABundle                 types.String `tfsdk:"custom_ca_bundle"`
	DefaultTags                    types.List   `tfsdk:"default_tags"`
	EC2MetadataServiceEndpoint     types.String `tfsdk:"ec2_metadata_service_endpoint"`
	EC2MetadataServiceEndpointMode types.String `tfsdk:"ec2_metadata_service_endpoint_mode"`
	Endpoints                      types.Set    `tfsdk:"endpoints"`
	ForbiddenAccountIDs            types.Set    `tfsdk:"forbidden_account_ids"`
	HTTPProxy                      types.String `tfsdk:"http_proxy"`
	IgnoreTags                     types.List   `tfsdk:"ignore_tags"`
	Insecure                       types.Bool   `tfsdk:"insecure"`
	MaxRetries                     types.Int64  `tfsdk:"max_retries"`
	Profile                        types.String `tfsdk:"profile"`
	Region                         types.String `tfsdk:"region"`
	S3ForcePathStyle               types.Bool   `tfsdk:"s3_force_path_style"`
	S3UsePathStyle                 types.Bool   `tfsdk:"s3_use_path_style"`
	SecretKey                      types.String `tfsdk:"secret_key"`
	SharedConfigFiles              types.List   `tfsdk:"shared_config_files"`
	SharedCredentialsFile          types.String `tfsdk:"shared_credentials_file"`
	SharedCredentialsFiles         types.List   `tfsdk:"shared_credentials_files"`
	SkipCredentialsValidation      types.Bool   `tfsdk:"skip_credentials_validation"`
	SkipGetEC2Platforms            types.Bool   `tfsdk:"skip_get_ec2_platforms"`
	SkipMetadataAPICheck           types.Bool   `tfsdk:"skip_metadata_api_check"`
	SkipRegionValidation           types.Bool   `tfsdk:"skip_region_validation"`
	SkipRequestingAccountID        types.Bool   `tfsdk:"skip_requesting_account_id"`
	STSRegion                      types.String `tfsdk:"sts_region"`
	Token                          types.String `tfsdk:"token"`
	UseDualStackEndpoint           types.Bool   `tfsdk:"use_dualstack_endpoint"`
	UseFIPSEndpoint                types.Bool   `tfsdk:"use_fips_endpoint"`

	// TODO
	// AssumeRole                     *assumeRoleConfig  `tfsdk:"assume_role"`
	// DefaultTags                    *defaultTagsConfig `tfsdk:"default_tags"`
	// IgnoreTags                     *ignoreTagsConfig `tfsdk:"ignore_tags"`

	terraformVersion string
}

type assumeRoleConfig struct {
	Duration          Duration     `tfsdk:"duration"`
	DurationSeconds   types.Int64  `tfsdk:"duration_seconds"`
	ExternalID        types.String `tfsdk:"external_id"`
	Policy            types.String `tfsdk:"policy"`
	PolicyARNs        types.Set    `tfsdk:"policy_arns"`
	RoleARN           types.String `tfsdk:"role_arn"`
	SessionName       types.String `tfsdk:"session_name"`
	Tags              types.Map    `tfsdk:"tags"`
	TransitiveTagKeys types.Set    `tfsdk:"transitive_tag_keys"`
}

type defaultTagsConfig struct {
	Tags types.Map `tfsdk:"tags"`
}

type ignoreTagsConfig struct {
	Keys        types.Set `tfsdk:"keys"`
	KeyPrefixes types.Set `tfsdk:"key_prefixes"`
}

// provider represents a Terraform Protocol v6 provider.
type provider struct {
	config *config
}

func (p *provider) Configure(ctx context.Context, request tfsdk.ConfigureProviderRequest, response *tfsdk.ConfigureProviderResponse) {
	var config config

	diags := request.Config.Get(ctx, &config)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)

		return
	}

	if !request.Config.Raw.IsFullyKnown() {
		response.Diagnostics.AddError("Unknown Value", "An attribute value is not yet known")
	}

	config.terraformVersion = request.TerraformVersion

	p.config = &config
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	var diags diag.Diagnostics
	dataSources := make(map[string]tfsdk.DataSourceType)

	return dataSources, diags
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	var diags diag.Diagnostics
	resources := make(map[string]tfsdk.ResourceType)

	return resources, diags
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	var diags diag.Diagnostics
	// This schema must match exactly the Terraform Protocol v5 provider's schema.
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"access_key": {
				Type:     types.StringType,
				Optional: true,
				Description: "The access key for API operations. You can retrieve this\n" +
					"from the 'Security & Credentials' section of the AWS console.",
			},
			"allowed_account_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Optional: true,
			},
			"custom_ca_bundle": {
				Type:     types.StringType,
				Optional: true,
				Description: "File containing custom root and intermediate certificates. " +
					"Can also be configured using the `AWS_CA_BUNDLE` environment variable. " +
					"(Setting `ca_bundle` in the shared config file is not supported.)",
			},
			"ec2_metadata_service_endpoint": {
				Type:     types.StringType,
				Optional: true,
				Description: "Address of the EC2 metadata service endpoint to use. " +
					"Can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.",
			},
			"ec2_metadata_service_endpoint_mode": {
				Type:     types.StringType,
				Optional: true,
				Description: "Protocol to use with EC2 metadata service endpoint." +
					"Valid values are `IPv4` and `IPv6`. Can also be configured using the `AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE` environment variable.",
			},
			"forbidden_account_ids": {
				Type:     types.SetType{ElemType: types.StringType},
				Optional: true,
			},
			"http_proxy": {
				Type:     types.StringType,
				Optional: true,
				Description: "The address of an HTTP proxy to use when accessing the AWS API. " +
					"Can also be configured using the `HTTP_PROXY` or `HTTPS_PROXY` environment variables.",
			},
			"insecure": {
				Type:     types.BoolType,
				Optional: true,
				Description: "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted, " +
					"default value is `false`",
			},
			"max_retries": {
				Type:     types.Int64Type,
				Optional: true,
				Description: "The maximum number of times an AWS API request is\n" +
					"being executed. If the API request still fails, an error is\n" +
					"thrown.",
			},
			"profile": {
				Type:     types.StringType,
				Optional: true,
				Description: "The profile for API operations. If not set, the default profile\n" +
					"created with `aws configure` will be used.",
			},
			"region": {
				Type:     types.StringType,
				Optional: true,
				Description: "The region where AWS operations will take place. Examples\n" +
					"are us-east-1, us-west-2, etc.", // lintignore:AWSAT003,
			},
			"s3_force_path_style": {
				Type:               types.BoolType,
				Optional:           true,
				DeprecationMessage: "Use s3_use_path_style instead.",
				Description: "Set this to true to enable the request to use path-style addressing,\n" +
					"i.e., https://s3.amazonaws.com/BUCKET/KEY. By default, the S3 client will\n" +
					"use virtual hosted bucket addressing when possible\n" +
					"(https://BUCKET.s3.amazonaws.com/KEY). Specific to the Amazon S3 service.",
			},
			"s3_use_path_style": {
				Type:     types.BoolType,
				Optional: true,
				Description: "Set this to true to enable the request to use path-style addressing,\n" +
					"i.e., https://s3.amazonaws.com/BUCKET/KEY. By default, the S3 client will\n" +
					"use virtual hosted bucket addressing when possible\n" +
					"(https://BUCKET.s3.amazonaws.com/KEY). Specific to the Amazon S3 service.",
			},
			"secret_key": {
				Type:     types.StringType,
				Optional: true,
				Description: "The secret key for API operations. You can retrieve this\n" +
					"from the 'Security & Credentials' section of the AWS console.",
			},
			"shared_config_files": {
				Type:        types.ListType{ElemType: types.StringType},
				Optional:    true,
				Description: "List of paths to shared config files. If not set, defaults to [~/.aws/config].",
			},
			"shared_credentials_file": {
				Type:               types.StringType,
				Optional:           true,
				DeprecationMessage: "Use shared_credentials_files instead.",
				Description:        "The path to the shared credentials file. If not set, defaults to ~/.aws/credentials.",
			},
			"shared_credentials_files": {
				Type:        types.ListType{ElemType: types.StringType},
				Optional:    true,
				Description: "List of paths to shared credentials files. If not set, defaults to [~/.aws/credentials].",
			},
			"skip_credentials_validation": {
				Type:     types.BoolType,
				Optional: true,
				Description: "Skip the credentials validation via STS API. " +
					"Used for AWS API implementations that do not have STS available/implemented.",
			},
			"skip_get_ec2_platforms": {
				Type:     types.BoolType,
				Optional: true,
				Description: "Skip getting the supported EC2 platforms. " +
					"Used by users that don't have ec2:DescribeAccountAttributes permissions.",
			},
			"skip_metadata_api_check": {
				Type:     types.BoolType,
				Optional: true,
				Description: "Skip the AWS Metadata API check. " +
					"Used for AWS API implementations that do not have a metadata api endpoint.",
			},
			"skip_region_validation": {
				Type:     types.BoolType,
				Optional: true,
				Description: "Skip static validation of region name. " +
					"Used by users of alternative AWS-like APIs or users w/ access to regions that are not public (yet).",
			},
			"skip_requesting_account_id": {
				Type:     types.BoolType,
				Optional: true,
				Description: "Skip requesting the account ID. " +
					"Used for AWS API implementations that do not have IAM/STS API and/or metadata API.",
			},
			"sts_region": {
				Type:     types.StringType,
				Optional: true,
				Description: "The region where AWS STS operations will take place. Examples\n" +
					"are us-east-1 and us-west-2.", // lintignore:AWSAT003,
			},
			"token": {
				Type:     types.StringType,
				Optional: true,
				Description: "session token. A session token is only required if you are\n" +
					"using temporary security credentials.",
			},
			"use_dualstack_endpoint": {
				Type:        types.BoolType,
				Optional:    true,
				Description: "Resolve an endpoint with DualStack capability",
			},
			"use_fips_endpoint": {
				Type:        types.BoolType,
				Optional:    true,
				Description: "Resolve an endpoint with FIPS capability",
			},
		},
		Blocks: map[string]tfsdk.Block{
			"assume_role": {
				Attributes: map[string]tfsdk.Attribute{
					"duration": {
						Type:        types.StringType,
						Optional:    true,
						Description: "The duration, between 15 minutes and 12 hours, of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
					},
					"duration_seconds": {
						Type:               types.Int64Type,
						Optional:           true,
						DeprecationMessage: "Use assume_role.0.duration instead",
						Description:        "The duration, in seconds, of the role session.",
					},
					"external_id": {
						Type:        types.StringType,
						Optional:    true,
						Description: "A unique identifier that might be required when you assume a role in another account.",
					},
					"policy": {
						Type:        types.StringType,
						Optional:    true,
						Description: "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
					},
					"policy_arns": {
						Type:        types.SetType{ElemType: types.StringType},
						Optional:    true,
						Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
					},
					"role_arn": {
						Type:        types.StringType,
						Optional:    true,
						Description: "Amazon Resource Name of an IAM Role to assume prior to making API calls.",
					},
					"session_name": {
						Type:        types.StringType,
						Optional:    true,
						Description: "An identifier for the assumed role session.",
					},
					"tags": {
						Type:        types.MapType{ElemType: types.StringType},
						Optional:    true,
						Description: "Assume role session tags.",
					},
					"transitive_tag_keys": {
						Type:        types.SetType{ElemType: types.StringType},
						Optional:    true,
						Description: "Assume role session tag keys to pass to any subsequent sessions.",
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    1,
			},
			"default_tags": {
				Attributes: map[string]tfsdk.Attribute{
					"tags": {
						Type:        types.MapType{ElemType: types.StringType},
						Optional:    true,
						Description: "Resource tags to default across all resources",
					},
				},
				Description: "Configuration block with settings to default resource tags across all resources.",
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    1,
			},
			"endpoints": endpointsBlock(),
			"ignore_tags": {
				Attributes: map[string]tfsdk.Attribute{
					"keys": {
						Type:        types.SetType{ElemType: types.StringType},
						Optional:    true,
						Description: "Resource tag keys to ignore across all resources.",
					},
					"key_prefixes": {
						Type:        types.SetType{ElemType: types.StringType},
						Optional:    true,
						Description: "Resource tag key prefixes to ignore across all resources.",
					},
				},
				Description: "Configuration block with settings to ignore resource tags across all resources.",
				NestingMode: tfsdk.BlockNestingModeList,
				MaxItems:    1,
			},
		},
	}

	return schema, diags
}

func endpointsBlock() tfsdk.Block {
	endpointsAttributes := make(map[string]tfsdk.Attribute)

	for _, serviceKey := range names.HCLKeys() {
		endpointsAttributes[serviceKey] = tfsdk.Attribute{
			Type:        types.StringType,
			Optional:    true,
			Description: "Use this to override the default service endpoint URL",
		}
	}

	return tfsdk.Block{
		Attributes:  endpointsAttributes,
		NestingMode: tfsdk.BlockNestingModeSet,
	}
}
