// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudfront_multitenant_distribution", name="Multi-tenant Distribution")
// @Tags(identifierAttribute="arn")
//
// Multi-tenant Distribution Limitations:
// The following fields are NOT supported for multi-tenant distributions and have been excluded from the schema:
// - CacheBehavior.DefaultTTL, MaxTTL, MinTTL (use cache policies instead)
// - CacheBehavior.SmoothStreaming
// - CacheBehavior.TrustedSigners (use TrustedKeyGroups instead)
// - DefaultCacheBehavior.DefaultTTL, MaxTTL, MinTTL (use cache policies instead)
// - DefaultCacheBehavior.SmoothStreaming
// - DefaultCacheBehavior.TrustedSigners (use TrustedKeyGroups instead)
// - DistributionConfig.Aliases (managed by connection groups)
// - DistributionConfig.AnycastIpListId (use connection group instead)
// - DistributionConfig.ContinuousDeploymentPolicyId
// - DistributionConfig.IsIPV6Enabled (use connection group instead)
// - DistributionConfig.PriceClass
// - DistributionConfig.Staging
// - CacheBehavior.ForwardedValues (deprecated and not supported)
// - DefaultCacheBehavior.ForwardedValues (deprecated and not supported)
// - ViewerCertificate.IAMCertificateId (use ACM certificates instead)
//
// Multi-tenant Distribution Requirements:
// - DistributionConfig.ConnectionMode is automatically set to "tenant-only"
// - DistributionConfig.TenantConfig must be specified (contains parameter definitions)
// - DistributionConfig.WebACLId must be a WAF V2 web ACL if specified
func newMultiTenantDistributionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &multiTenantDistributionResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type multiTenantDistributionResource struct {
	framework.ResourceWithModel[multiTenantDistributionResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (*multiTenantDistributionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cloudfront_multitenant_distribution"
}

func (r *multiTenantDistributionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"caller_reference": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrComment: schema.StringAttribute{
				Required: true,
			},
			"default_root_object": schema.StringAttribute{
				Optional: true,
			},
			names.AttrEnabled: schema.BoolAttribute{
				Required: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			"http_version": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.HttpVersion](),
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"web_acl_id": schema.StringAttribute{
				Optional: true,
				// Note: For multi-tenant distributions, this must be a WAF V2 web ACL if specified
			},
		},
		Blocks: map[string]schema.Block{
			"custom_error_response": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[customErrorResponseModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"error_caching_min_ttl": schema.Int64Attribute{
							Optional: true,
						},
						"error_code": schema.Int64Attribute{
							Required: true,
						},
						"response_code": schema.Int64Attribute{
							Optional: true,
						},
						"response_page_path": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"cache_behavior": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[cacheBehaviorModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"allowed_methods": schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringEnumType[awstypes.Method](),
						},
						"cache_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"cached_methods": schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringEnumType[awstypes.Method](),
						},
						"compress": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
						"field_level_encryption_id": schema.StringAttribute{
							Optional: true,
						},
						"origin_request_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"path_pattern": schema.StringAttribute{
							Required: true,
						},
						"realtime_log_config_arn": schema.StringAttribute{
							Optional: true,
						},
						"response_headers_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"target_origin_id": schema.StringAttribute{
							Required: true,
						},
						"viewer_protocol_policy": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.ViewerProtocolPolicy](),
						},
						// Note: smooth_streaming and trusted_signers removed - not supported for multi-tenant distributions
					},
					Blocks: map[string]schema.Block{
						"function_association": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[functionAssociationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"event_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.EventType](),
									},
									names.AttrFunctionARN: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"lambda_function_association": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[lambdaFunctionAssociationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"event_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.EventType](),
									},
									"include_body": schema.BoolAttribute{
										Optional: true,
										Computed: true,
									},
									"lambda_arn": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"trusted_key_groups": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[trustedKeyGroupsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"items": schema.ListAttribute{
										Optional:   true,
										CustomType: fwtypes.ListOfStringType,
									},
									names.AttrEnabled: schema.BoolAttribute{
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						// Note: trusted_signers removed - not supported for multi-tenant distributions
					},
				},
			},
			"default_cache_behavior": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[defaultCacheBehaviorModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"allowed_methods": schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringEnumType[awstypes.Method](),
						},
						"cache_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"cached_methods": schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringEnumType[awstypes.Method](),
						},
						"compress": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
						"field_level_encryption_id": schema.StringAttribute{
							Optional: true,
						},
						"origin_request_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"realtime_log_config_arn": schema.StringAttribute{
							Optional: true,
						},
						"response_headers_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"target_origin_id": schema.StringAttribute{
							Required: true,
						},
						// Note: smooth_streaming removed - not supported for multi-tenant distributions
						"viewer_protocol_policy": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.ViewerProtocolPolicy](),
						},
					},
					Blocks: map[string]schema.Block{
						"function_association": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[functionAssociationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"event_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.EventType](),
									},
									names.AttrFunctionARN: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"lambda_function_association": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[lambdaFunctionAssociationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"event_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.EventType](),
									},
									"include_body": schema.BoolAttribute{
										Optional: true,
										Computed: true,
									},
									"lambda_arn": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"trusted_key_groups": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[trustedKeyGroupsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"items": schema.ListAttribute{
										Optional:   true,
										CustomType: fwtypes.ListOfStringType,
									},
									names.AttrEnabled: schema.BoolAttribute{
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						// Note: trusted_signers removed - not supported for multi-tenant distributions
					},
				},
			},
			"origin": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[originModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"connection_attempts": schema.Int32Attribute{
							Optional: true,
							Computed: true,
						},
						"connection_timeout": schema.Int32Attribute{
							Optional: true,
							Computed: true,
						},
						names.AttrDomainName: schema.StringAttribute{
							Required: true,
						},
						"id": schema.StringAttribute{
							Required: true,
						},
						"origin_access_control_id": schema.StringAttribute{
							Optional: true,
						},
						"origin_path": schema.StringAttribute{
							Optional: true,
						},
						"response_completion_timeout": schema.Int32Attribute{
							Optional: true,
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"custom_header": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[customHeaderModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"custom_origin_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customOriginConfigModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"http_port": schema.Int32Attribute{
										Required: true,
									},
									"https_port": schema.Int32Attribute{
										Required: true,
									},
									names.AttrIPAddressType: schema.StringAttribute{
										Optional:   true,
										CustomType: fwtypes.StringEnumType[awstypes.IpAddressType](),
									},
									"origin_keepalive_timeout": schema.Int32Attribute{
										Optional: true,
										Computed: true,
									},
									"origin_read_timeout": schema.Int32Attribute{
										Optional: true,
										Computed: true,
									},
									"origin_protocol_policy": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.OriginProtocolPolicy](),
									},
									"origin_ssl_protocols": schema.SetAttribute{
										Required:   true,
										CustomType: fwtypes.SetOfStringEnumType[awstypes.SslProtocol](),
									},
								},
							},
						},
						"origin_shield": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[originShieldModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
									},
									"origin_shield_region": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"s3_origin_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3OriginConfigModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"origin_access_identity": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"origin_group": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[originGroupModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"origin_id": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"failover_criteria": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[failoverCriteriaModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"status_codes": schema.SetAttribute{
										Required:    true,
										ElementType: types.Int64Type,
									},
								},
							},
						},
						"member": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[memberModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(2),
								listvalidator.SizeAtMost(2),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"origin_id": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"restrictions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[restrictionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"geo_restriction": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[geoRestrictionModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"items": schema.SetAttribute{
										Optional:   true,
										CustomType: fwtypes.SetOfStringType,
									},
									"restriction_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.GeoRestrictionType](),
									},
								},
							},
						},
					},
				},
			},
			"tenant_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tenantConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"parameter_definition": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[parameterDefinitionModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"definition": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[parameterDefinitionSchemaModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"string_schema": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[stringSchemaConfigModel](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"required": schema.BoolAttribute{
																Required: true,
															},
															names.AttrComment: schema.StringAttribute{
																Optional: true,
															},
															"default_value": schema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Update: true,
			}),
			"viewer_certificate": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[viewerCertificateModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"acm_certificate_arn": schema.StringAttribute{
							Optional: true,
						},
						"cloudfront_default_certificate": schema.BoolAttribute{
							Optional: true,
						},
						"minimum_protocol_version": schema.StringAttribute{
							Optional:   true,
							Computed:   true,
							CustomType: fwtypes.StringEnumType[awstypes.MinimumProtocolVersion](),
						},
						"ssl_support_method": schema.StringAttribute{
							Optional:   true,
							CustomType: fwtypes.StringEnumType[awstypes.SSLSupportMethod](),
						},
					},
				},
			},
		},
	}
}

func (r *multiTenantDistributionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data multiTenantDistributionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := &cloudfront.CreateDistributionWithTagsInput{
		DistributionConfigWithTags: &awstypes.DistributionConfigWithTags{
			DistributionConfig: &awstypes.DistributionConfig{},
			Tags:               &awstypes.Tags{Items: []awstypes.Tag{}},
		},
	}

	expandDiags := fwflex.Expand(ctx, data, input.DistributionConfigWithTags.DistributionConfig)
	response.Diagnostics.Append(expandDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set required computed fields that AutoFlex can't handle
	input.DistributionConfigWithTags.DistributionConfig.CallerReference = aws.String(id.UniqueId())

	// Set ConnectionMode to "tenant-only" to create a multi-tenant distribution instead of standard distribution
	// This is the key field that distinguishes multi-tenant from standard distributions
	input.DistributionConfigWithTags.DistributionConfig.ConnectionMode = awstypes.ConnectionModeTenantOnly

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.DistributionConfigWithTags.Tags.Items = tags
	}

	output, err := conn.CreateDistributionWithTags(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("creating CloudFront Multi-tenant Distribution", err.Error())
		return
	}

	// Set the ID so Read can work
	data.ID = types.StringValue(aws.ToString(output.Distribution.Id))

	// Call Read to populate all computed fields
	readReq := resource.ReadRequest{State: response.State}
	readResp := &resource.ReadResponse{State: response.State}

	// Set the ID in state first
	response.Diagnostics.Append(response.State.Set(ctx, data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Now call Read to populate everything else
	r.Read(ctx, readReq, readResp)
	response.Diagnostics.Append(readResp.Diagnostics...)
	if response.Diagnostics.HasError() {
		return
	}

	// Copy the fully populated state
	response.State = readResp.State
}

func (r *multiTenantDistributionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	fmt.Printf("DEBUG Read: Entering Read method\n")
	var data multiTenantDistributionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	output, err := findDistributionByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.AddWarning(
			"CloudFront Multi-tenant Distribution not found",
			fmt.Sprintf("CloudFront Multi-tenant Distribution (%s) not found, removing from state", data.ID.ValueString()),
		)
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading CloudFront Multi-tenant Distribution", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Distribution.DistributionConfig, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *multiTenantDistributionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new multiTenantDistributionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	// For now, just set the new state - updates can be implemented later
	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

func (r *multiTenantDistributionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data multiTenantDistributionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	_, err := conn.DeleteDistribution(ctx, &cloudfront.DeleteDistributionInput{
		Id:      aws.String(data.ID.ValueString()),
		IfMatch: aws.String(data.ETag.ValueString()),
	})

	if errs.IsA[*awstypes.NoSuchDistribution](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError("deleting CloudFront Multi-tenant Distribution", err.Error())
		return
	}
}

type multiTenantDistributionResourceModel struct {
	ARN                  types.String                                               `tfsdk:"arn"`
	CacheBehavior        fwtypes.ListNestedObjectValueOf[cacheBehaviorModel]        `tfsdk:"cache_behavior"`
	CallerReference      types.String                                               `tfsdk:"caller_reference"`
	Comment              types.String                                               `tfsdk:"comment"`
	CustomErrorResponse  fwtypes.SetNestedObjectValueOf[customErrorResponseModel]   `tfsdk:"custom_error_response"`
	DefaultCacheBehavior fwtypes.ListNestedObjectValueOf[defaultCacheBehaviorModel] `tfsdk:"default_cache_behavior"`
	DefaultRootObject    types.String                                               `tfsdk:"default_root_object"`
	Enabled              types.Bool                                                 `tfsdk:"enabled"`
	ETag                 types.String                                               `tfsdk:"etag"`
	HTTPVersion          fwtypes.StringEnum[awstypes.HttpVersion]                   `tfsdk:"http_version"`
	ID                   types.String                                               `tfsdk:"id"`
	Origin               fwtypes.SetNestedObjectValueOf[originModel]                `tfsdk:"origin"`
	OriginGroup          fwtypes.SetNestedObjectValueOf[originGroupModel]           `tfsdk:"origin_group"`
	Restrictions         fwtypes.ListNestedObjectValueOf[restrictionsModel]         `tfsdk:"restrictions"`
	Status               types.String                                               `tfsdk:"status"`
	Tags                 tftags.Map                                                 `tfsdk:"tags"`
	TagsAll              tftags.Map                                                 `tfsdk:"tags_all"`
	TenantConfig         fwtypes.ListNestedObjectValueOf[tenantConfigModel]         `tfsdk:"tenant_config"`
	Timeouts             timeouts.Value                                             `tfsdk:"timeouts"`
	ViewerCertificate    fwtypes.ListNestedObjectValueOf[viewerCertificateModel]    `tfsdk:"viewer_certificate"`
	WebACLID             types.String                                               `tfsdk:"web_acl_id"`
}

type originModel struct {
	ConnectionAttempts        types.Int32                                              `tfsdk:"connection_attempts"`
	ConnectionTimeout         types.Int32                                              `tfsdk:"connection_timeout"`
	CustomHeader              fwtypes.SetNestedObjectValueOf[customHeaderModel]        `tfsdk:"custom_header"`
	CustomOriginConfig        fwtypes.ListNestedObjectValueOf[customOriginConfigModel] `tfsdk:"custom_origin_config"`
	DomainName                types.String                                             `tfsdk:"domain_name"`
	ID                        types.String                                             `tfsdk:"id"`
	OriginAccessControlID     types.String                                             `tfsdk:"origin_access_control_id"`
	OriginPath                types.String                                             `tfsdk:"origin_path"`
	OriginShield              fwtypes.ListNestedObjectValueOf[originShieldModel]       `tfsdk:"origin_shield"`
	ResponseCompletionTimeout types.Int32                                              `tfsdk:"response_completion_timeout"`
	S3OriginConfig            fwtypes.ListNestedObjectValueOf[s3OriginConfigModel]     `tfsdk:"s3_origin_config"`
}

type customHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type customOriginConfigModel struct {
	HTTPPort               types.Int32                                                  `tfsdk:"http_port"`
	HTTPSPort              types.Int32                                                  `tfsdk:"https_port"`
	IPAddressType          fwtypes.StringEnum[awstypes.IpAddressType]                   `tfsdk:"ip_address_type"`
	OriginKeepaliveTimeout types.Int32                                                  `tfsdk:"origin_keepalive_timeout"`
	OriginReadTimeout      types.Int32                                                  `tfsdk:"origin_read_timeout"`
	OriginProtocolPolicy   fwtypes.StringEnum[awstypes.OriginProtocolPolicy]            `tfsdk:"origin_protocol_policy"`
	OriginSSLProtocols     fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.SslProtocol]] `tfsdk:"origin_ssl_protocols"`
}

type originShieldModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	OriginShieldRegion types.String `tfsdk:"origin_shield_region"`
}

type s3OriginConfigModel struct {
	OriginAccessIdentity types.String `tfsdk:"origin_access_identity"`
}

type originGroupModel struct {
	FailoverCriteria fwtypes.ListNestedObjectValueOf[failoverCriteriaModel] `tfsdk:"failover_criteria"`
	Member           fwtypes.ListNestedObjectValueOf[memberModel]           `tfsdk:"member"`
	OriginID         types.String                                           `tfsdk:"origin_id"`
}

type failoverCriteriaModel struct {
	StatusCodes fwtypes.SetValueOf[types.Int64] `tfsdk:"status_codes"`
}

type memberModel struct {
	OriginID types.String `tfsdk:"origin_id"`
}

type defaultCacheBehaviorModel struct {
	AllowedMethods            fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]]        `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                   `tfsdk:"cache_policy_id"`
	CachedMethods             fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]]        `tfsdk:"cached_methods"`
	Compress                  types.Bool                                                     `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                   `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.SetNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association"`
	LambdaFunctionAssociation fwtypes.SetNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association"`
	OriginRequestPolicyID     types.String                                                   `tfsdk:"origin_request_policy_id"`
	RealtimeLogConfigARN      types.String                                                   `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                   `tfsdk:"response_headers_policy_id"`
	TargetOriginID            types.String                                                   `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListNestedObjectValueOf[trustedKeyGroupsModel]         `tfsdk:"trusted_key_groups"`
	ViewerProtocolPolicy      fwtypes.StringEnum[awstypes.ViewerProtocolPolicy]              `tfsdk:"viewer_protocol_policy"`
	// Note: SmoothStreaming and TrustedSigners removed - not supported for multi-tenant distributions
}

type cacheBehaviorModel struct {
	AllowedMethods            fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]]        `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                   `tfsdk:"cache_policy_id"`
	CachedMethods             fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]]        `tfsdk:"cached_methods"`
	Compress                  types.Bool                                                     `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                   `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.SetNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association"`
	LambdaFunctionAssociation fwtypes.SetNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association"`
	OriginRequestPolicyID     types.String                                                   `tfsdk:"origin_request_policy_id"`
	PathPattern               types.String                                                   `tfsdk:"path_pattern"`
	RealtimeLogConfigARN      types.String                                                   `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                   `tfsdk:"response_headers_policy_id"`
	TargetOriginID            types.String                                                   `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListNestedObjectValueOf[trustedKeyGroupsModel]         `tfsdk:"trusted_key_groups"`
	ViewerProtocolPolicy      fwtypes.StringEnum[awstypes.ViewerProtocolPolicy]              `tfsdk:"viewer_protocol_policy"`
	// Note: SmoothStreaming and TrustedSigners removed - not supported for multi-tenant distributions
}

type customErrorResponseModel struct {
	ErrorCachingMinTtl types.Int64  `tfsdk:"error_caching_min_ttl"`
	ErrorCode          types.Int64  `tfsdk:"error_code"`
	ResponseCode       types.Int64  `tfsdk:"response_code"`
	ResponsePagePath   types.String `tfsdk:"response_page_path"`
}

type restrictionsModel struct {
	GeoRestriction fwtypes.ListNestedObjectValueOf[geoRestrictionModel] `tfsdk:"geo_restriction"`
}

type geoRestrictionModel struct {
	Items           fwtypes.SetValueOf[types.String]                `tfsdk:"items"`
	RestrictionType fwtypes.StringEnum[awstypes.GeoRestrictionType] `tfsdk:"restriction_type"`
}

type viewerCertificateModel struct {
	ACMCertificateARN            types.String                                        `tfsdk:"acm_certificate_arn"`
	CloudfrontDefaultCertificate types.Bool                                          `tfsdk:"cloudfront_default_certificate"`
	MinimumProtocolVersion       fwtypes.StringEnum[awstypes.MinimumProtocolVersion] `tfsdk:"minimum_protocol_version"`
	SSLSupportMethod             fwtypes.StringEnum[awstypes.SSLSupportMethod]       `tfsdk:"ssl_support_method"`
}

type functionAssociationModel struct {
	EventType   fwtypes.StringEnum[awstypes.EventType] `tfsdk:"event_type"`
	FunctionARN types.String                           `tfsdk:"function_arn"`
}

type lambdaFunctionAssociationModel struct {
	EventType   fwtypes.StringEnum[awstypes.EventType] `tfsdk:"event_type"`
	IncludeBody types.Bool                             `tfsdk:"include_body"`
	LambdaARN   types.String                           `tfsdk:"lambda_arn"`
}

type tenantConfigModel struct {
	ParameterDefinition fwtypes.ListNestedObjectValueOf[parameterDefinitionModel] `tfsdk:"parameter_definition"`
}

type parameterDefinitionModel struct {
	Name       types.String                                                    `tfsdk:"name"`
	Definition fwtypes.ListNestedObjectValueOf[parameterDefinitionSchemaModel] `tfsdk:"definition"`
}

type parameterDefinitionSchemaModel struct {
	StringSchema fwtypes.ListNestedObjectValueOf[stringSchemaConfigModel] `tfsdk:"string_schema"`
}

type stringSchemaConfigModel struct {
	Required     types.Bool   `tfsdk:"required"`
	Comment      types.String `tfsdk:"comment"`
	DefaultValue types.String `tfsdk:"default_value"`
}

type trustedKeyGroupsModel struct {
	Items   fwtypes.ListValueOf[types.String] `tfsdk:"items"`
	Enabled types.Bool                        `tfsdk:"enabled"`
}
