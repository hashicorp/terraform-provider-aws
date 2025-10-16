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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
				Optional: true,
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
							CustomType: fwtypes.SetOfStringType,
						},
						"cache_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"cached_methods": schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringType,
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
						"smooth_streaming": schema.BoolAttribute{
							Optional: true,
						},
						"target_origin_id": schema.StringAttribute{
							Required: true,
						},
						"trusted_key_groups": schema.ListAttribute{
							Optional:   true,
							Computed:   true,
							CustomType: fwtypes.ListOfStringType,
						},
						"trusted_signers": schema.ListAttribute{
							Optional:   true,
							Computed:   true,
							CustomType: fwtypes.ListOfStringType,
						},
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
					},
				},
			},
			"default_cache_behavior": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[defaultCacheBehaviorModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"allowed_methods": schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringType,
						},
						"cache_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"cached_methods": schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringType,
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
						"smooth_streaming": schema.BoolAttribute{
							Optional: true,
						},
						"target_origin_id": schema.StringAttribute{
							Required: true,
						},
						"trusted_key_groups": schema.ListAttribute{
							Optional:   true,
							Computed:   true,
							CustomType: fwtypes.ListOfStringType,
						},
						"trusted_signers": schema.ListAttribute{
							Optional:   true,
							Computed:   true,
							CustomType: fwtypes.ListOfStringType,
						},
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
										CustomType: fwtypes.SetOfStringType,
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
									"locations": schema.SetAttribute{
										Optional:   true,
										Computed:   true,
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

	var distributionConfig awstypes.DistributionConfig
	expandedConfig, d := data.Expand(ctx)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}
	distributionConfig = expandedConfig.(awstypes.DistributionConfig)

	// Debug: Log what AutoFlex populated
	fmt.Printf("DEBUG: After AutoFlex - CallerReference: %v\n", distributionConfig.CallerReference)
	fmt.Printf("DEBUG: After AutoFlex - Origins: %+v\n", distributionConfig.Origins)
	fmt.Printf("DEBUG: After AutoFlex - DefaultCacheBehavior: %+v\n", distributionConfig.DefaultCacheBehavior)
	if distributionConfig.DefaultCacheBehavior != nil {
		fmt.Printf("DEBUG: AllowedMethods: %+v\n", distributionConfig.DefaultCacheBehavior.AllowedMethods)
	}

	// Set required computed fields that AutoFlex can't handle
	distributionConfig.CallerReference = aws.String(id.UniqueId())

	input := cloudfront.CreateDistributionWithTagsInput{
		DistributionConfigWithTags: &awstypes.DistributionConfigWithTags{
			DistributionConfig: &distributionConfig,
			Tags:               &awstypes.Tags{Items: []awstypes.Tag{}},
		},
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.DistributionConfigWithTags.Tags.Items = tags
	}

	output, err := conn.CreateDistributionWithTags(ctx, &input)
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

	// Use custom flattener instead of AutoFlex
	response.Diagnostics.Append(r.Flatten(ctx, output, &data)...)
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
	HTTPPort               types.Int32                                       `tfsdk:"http_port"`
	HTTPSPort              types.Int32                                       `tfsdk:"https_port"`
	IPAddressType          fwtypes.StringEnum[awstypes.IpAddressType]        `tfsdk:"ip_address_type"`
	OriginKeepaliveTimeout types.Int32                                       `tfsdk:"origin_keepalive_timeout"`
	OriginReadTimeout      types.Int32                                       `tfsdk:"origin_read_timeout"`
	OriginProtocolPolicy   fwtypes.StringEnum[awstypes.OriginProtocolPolicy] `tfsdk:"origin_protocol_policy"`
	OriginSSLProtocols     fwtypes.SetValueOf[types.String]                  `tfsdk:"origin_ssl_protocols"`
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
	AllowedMethods            fwtypes.SetValueOf[types.String]                               `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                   `tfsdk:"cache_policy_id"`
	CachedMethods             fwtypes.SetValueOf[types.String]                               `tfsdk:"cached_methods"`
	Compress                  types.Bool                                                     `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                   `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.SetNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association"`
	LambdaFunctionAssociation fwtypes.SetNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association"`
	OriginRequestPolicyID     types.String                                                   `tfsdk:"origin_request_policy_id"`
	RealtimeLogConfigARN      types.String                                                   `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                   `tfsdk:"response_headers_policy_id"`
	SmoothStreaming           types.Bool                                                     `tfsdk:"smooth_streaming"`
	TargetOriginID            types.String                                                   `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListValueOf[types.String]                              `tfsdk:"trusted_key_groups"`
	TrustedSigners            fwtypes.ListValueOf[types.String]                              `tfsdk:"trusted_signers"`
	ViewerProtocolPolicy      fwtypes.StringEnum[awstypes.ViewerProtocolPolicy]              `tfsdk:"viewer_protocol_policy"`
}

type cacheBehaviorModel struct {
	AllowedMethods            fwtypes.SetValueOf[types.String]                               `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                   `tfsdk:"cache_policy_id"`
	CachedMethods             fwtypes.SetValueOf[types.String]                               `tfsdk:"cached_methods"`
	Compress                  types.Bool                                                     `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                   `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.SetNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association"`
	LambdaFunctionAssociation fwtypes.SetNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association"`
	OriginRequestPolicyID     types.String                                                   `tfsdk:"origin_request_policy_id"`
	PathPattern               types.String                                                   `tfsdk:"path_pattern"`
	RealtimeLogConfigARN      types.String                                                   `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                   `tfsdk:"response_headers_policy_id"`
	SmoothStreaming           types.Bool                                                     `tfsdk:"smooth_streaming"`
	TargetOriginID            types.String                                                   `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListValueOf[types.String]                              `tfsdk:"trusted_key_groups"`
	TrustedSigners            fwtypes.ListValueOf[types.String]                              `tfsdk:"trusted_signers"`
	ViewerProtocolPolicy      fwtypes.StringEnum[awstypes.ViewerProtocolPolicy]              `tfsdk:"viewer_protocol_policy"`
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
	Locations       fwtypes.SetValueOf[types.String]                `tfsdk:"locations"`
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

func (m multiTenantDistributionResourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var distributionConfig awstypes.DistributionConfig

	// DEBUGGING: Add a stack trace detection to catch infinite recursion
	// This line is causing infinite recursion because AutoFlex will call this Expand method again
	// when it encounters the multiTenantDistributionResourceModel which implements Expander interface
	// REMOVED: diags.Append(fwflex.Expand(ctx, m, &distributionConfig)...)

	// Manually set basic fields instead of using AutoFlex on the entire struct
	distributionConfig.CallerReference = aws.String(m.CallerReference.ValueString())
	distributionConfig.Comment = aws.String(m.Comment.ValueString())
	distributionConfig.DefaultRootObject = aws.String(m.DefaultRootObject.ValueString())
	distributionConfig.Enabled = aws.Bool(m.Enabled.ValueBool())
	if !m.HTTPVersion.IsNull() {
		distributionConfig.HttpVersion = m.HTTPVersion.ValueEnum()
	}
	if !m.WebACLID.IsNull() {
		distributionConfig.WebACLId = aws.String(m.WebACLID.ValueString())
	}

	// Set multi-tenant specific fields
	distributionConfig.ConnectionMode = awstypes.ConnectionModeTenantOnly

	// Handle CustomErrorResponse field
	if !m.CustomErrorResponse.IsNull() {
		customErrorResponses, d := m.CustomErrorResponse.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var items []awstypes.CustomErrorResponse
		for _, errorResponse := range customErrorResponses {
			var item awstypes.CustomErrorResponse
			diags.Append(fwflex.Expand(ctx, errorResponse, &item)...)
			if diags.HasError() {
				return nil, diags
			}
			items = append(items, item)
		}

		distributionConfig.CustomErrorResponses = &awstypes.CustomErrorResponses{
			Items:    items,
			Quantity: aws.Int32(int32(len(items))),
		}
	}

	// Handle Restrictions field
	if !m.Restrictions.IsNull() {
		restrictions, d := m.Restrictions.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		if len(restrictions) > 0 {
			var restrictionsItem awstypes.Restrictions
			diags.Append(fwflex.Expand(ctx, restrictions[0], &restrictionsItem)...)
			if diags.HasError() {
				return nil, diags
			}

			// Handle GeoRestriction Items and Quantity
			if restrictionsItem.GeoRestriction != nil {
				geoRestrictions, d := restrictions[0].GeoRestriction.ToSlice(ctx)
				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}

				if len(geoRestrictions) > 0 {
					if !geoRestrictions[0].Locations.IsNull() {
						locations := geoRestrictions[0].Locations.Elements()

						var items []string
						for _, loc := range locations {
							items = append(items, loc.(types.String).ValueString())
						}

						restrictionsItem.GeoRestriction.Items = items
						restrictionsItem.GeoRestriction.Quantity = aws.Int32(int32(len(items)))
					}
				}
			}

			distributionConfig.Restrictions = &restrictionsItem
		}
	}

	// Handle ViewerCertificate field
	if !m.ViewerCertificate.IsNull() {
		viewerCerts, d := m.ViewerCertificate.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		if len(viewerCerts) > 0 {
			var viewerCert awstypes.ViewerCertificate
			diags.Append(fwflex.Expand(ctx, viewerCerts[0], &viewerCert)...)
			if diags.HasError() {
				return nil, diags
			}
			distributionConfig.ViewerCertificate = &viewerCert
		}
	}

	// Handle TenantConfig field (this might be a custom field that needs special handling)
	// Note: This field might not exist in the AWS DistributionConfig, so we may need to handle it differently

	// Manually handle Origins (AWS expects Origins struct with Items + Quantity)
	if !m.Origin.IsNull() {
		origins, d := m.Origin.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var items []awstypes.Origin
		for _, origin := range origins {
			var item awstypes.Origin

			// Use AutoFlex with field prefix to avoid nested CustomOriginConfig issues
			diags.Append(fwflex.Expand(ctx, origin, &item, fwflex.WithFieldNamePrefix("Origin"))...)
			if diags.HasError() {
				return nil, diags
			}

			// Manually handle CustomOriginConfig.OriginSSLProtocols
			if !origin.CustomOriginConfig.IsNull() {
				customConfigs, d := origin.CustomOriginConfig.ToSlice(ctx)
				diags.Append(d...)
				if diags.HasError() {
					return nil, diags
				}

				if len(customConfigs) > 0 && !customConfigs[0].OriginSSLProtocols.IsNull() {
					protocolElements := customConfigs[0].OriginSSLProtocols.Elements()
					sslProtocols := make([]awstypes.SslProtocol, 0, len(protocolElements))
					for _, elem := range protocolElements {
						if strVal, ok := elem.(types.String); ok {
							sslProtocols = append(sslProtocols, awstypes.SslProtocol(strVal.ValueString()))
						}
					}
					if item.CustomOriginConfig == nil {
						item.CustomOriginConfig = &awstypes.CustomOriginConfig{}
					}
					item.CustomOriginConfig.OriginSslProtocols = &awstypes.OriginSslProtocols{
						Items:    sslProtocols,
						Quantity: aws.Int32(int32(len(sslProtocols))),
					}
				}
			}

			items = append(items, item)
		}

		distributionConfig.Origins = &awstypes.Origins{
			Items:    items,
			Quantity: aws.Int32(int32(len(items))),
		}
	}

	// Manually handle DefaultCacheBehavior (AWS expects AllowedMethods/CachedMethods structs)
	if !m.DefaultCacheBehavior.IsNull() {
		behaviors, d := m.DefaultCacheBehavior.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		if len(behaviors) > 0 {
			behavior := behaviors[0]
			var defaultBehavior awstypes.DefaultCacheBehavior

			// Manually set basic fields instead of using AutoFlex
			if !behavior.CachePolicyID.IsNull() {
				defaultBehavior.CachePolicyId = aws.String(behavior.CachePolicyID.ValueString())
			}
			if !behavior.Compress.IsNull() {
				defaultBehavior.Compress = aws.Bool(behavior.Compress.ValueBool())
			}
			if !behavior.FieldLevelEncryptionID.IsNull() {
				defaultBehavior.FieldLevelEncryptionId = aws.String(behavior.FieldLevelEncryptionID.ValueString())
			}
			if !behavior.OriginRequestPolicyID.IsNull() {
				defaultBehavior.OriginRequestPolicyId = aws.String(behavior.OriginRequestPolicyID.ValueString())
			}
			if !behavior.RealtimeLogConfigARN.IsNull() {
				defaultBehavior.RealtimeLogConfigArn = aws.String(behavior.RealtimeLogConfigARN.ValueString())
			}
			if !behavior.ResponseHeadersPolicyID.IsNull() {
				defaultBehavior.ResponseHeadersPolicyId = aws.String(behavior.ResponseHeadersPolicyID.ValueString())
			}
			if !behavior.SmoothStreaming.IsNull() {
				defaultBehavior.SmoothStreaming = aws.Bool(behavior.SmoothStreaming.ValueBool())
			}
			if !behavior.TargetOriginID.IsNull() {
				defaultBehavior.TargetOriginId = aws.String(behavior.TargetOriginID.ValueString())
			}
			if !behavior.ViewerProtocolPolicy.IsNull() {
				defaultBehavior.ViewerProtocolPolicy = behavior.ViewerProtocolPolicy.ValueEnum()
			}

			// Manually handle AllowedMethods structure
			if !behavior.AllowedMethods.IsNull() {
				allowedMethods := behavior.AllowedMethods.Elements()

				var methods []awstypes.Method
				for _, method := range allowedMethods {
					methods = append(methods, awstypes.Method(method.(types.String).ValueString()))
				}

				defaultBehavior.AllowedMethods = &awstypes.AllowedMethods{
					Items:    methods,
					Quantity: aws.Int32(int32(len(methods))),
				}

				// Handle CachedMethods if present
				if !behavior.CachedMethods.IsNull() {
					cachedMethods := behavior.CachedMethods.Elements()

					var cached []awstypes.Method
					for _, method := range cachedMethods {
						cached = append(cached, awstypes.Method(method.(types.String).ValueString()))
					}

					defaultBehavior.AllowedMethods.CachedMethods = &awstypes.CachedMethods{
						Items:    cached,
						Quantity: aws.Int32(int32(len(cached))),
					}
				}
			}

			distributionConfig.DefaultCacheBehavior = &defaultBehavior
		}
	}

	// Manually handle CacheBehaviors (AWS expects CacheBehaviors struct with Items + Quantity)
	if !m.CacheBehavior.IsNull() {
		behaviors, d := m.CacheBehavior.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var items []awstypes.CacheBehavior
		for _, behavior := range behaviors {
			var cacheBehavior awstypes.CacheBehavior

			// Use AutoFlex for most fields
			diags.Append(fwflex.Expand(ctx, behavior, &cacheBehavior)...)
			if diags.HasError() {
				return nil, diags
			}

			// Manually handle AllowedMethods structure
			if !behavior.AllowedMethods.IsNull() {
				allowedMethods := behavior.AllowedMethods.Elements()

				var methods []awstypes.Method
				for _, method := range allowedMethods {
					methods = append(methods, awstypes.Method(method.(types.String).ValueString()))
				}

				cacheBehavior.AllowedMethods = &awstypes.AllowedMethods{
					Items:    methods,
					Quantity: aws.Int32(int32(len(methods))),
				}

				// Handle CachedMethods if present
				if !behavior.CachedMethods.IsNull() {
					cachedMethods := behavior.CachedMethods.Elements()

					var cached []awstypes.Method
					for _, method := range cachedMethods {
						cached = append(cached, awstypes.Method(method.(types.String).ValueString()))
					}

					cacheBehavior.AllowedMethods.CachedMethods = &awstypes.CachedMethods{
						Items:    cached,
						Quantity: aws.Int32(int32(len(cached))),
					}
				}
			}

			items = append(items, cacheBehavior)
		}

		distributionConfig.CacheBehaviors = &awstypes.CacheBehaviors{
			Items:    items,
			Quantity: aws.Int32(int32(len(items))),
		}
	}

	// Manually handle OriginGroups (AWS expects OriginGroups struct with Items + Quantity)
	if !m.OriginGroup.IsNull() {
		originGroups, d := m.OriginGroup.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var items []awstypes.OriginGroup
		for _, originGroup := range originGroups {
			var item awstypes.OriginGroup
			diags.Append(fwflex.Expand(ctx, originGroup, &item)...)
			if diags.HasError() {
				return nil, diags
			}
			items = append(items, item)
		}

		distributionConfig.OriginGroups = &awstypes.OriginGroups{
			Items:    items,
			Quantity: aws.Int32(int32(len(items))),
		}
	}

	return distributionConfig, diags
}

func (r *multiTenantDistributionResource) Flatten(ctx context.Context, output *cloudfront.GetDistributionOutput, data *multiTenantDistributionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if output == nil || output.Distribution == nil {
		return diags
	}

	distribution := output.Distribution
	config := distribution.DistributionConfig

	// DEBUGGING: Add a stack trace detection to catch infinite recursion
	// This line is causing infinite recursion because AutoFlex will call this Flatten method again
	// when it encounters the multiTenantDistributionResourceModel which implements Flattener interface
	// REMOVED: diags.Append(fwflex.Flatten(ctx, output, data)...)

	// Manually set basic fields that AutoFlex would have handled
	if distribution.Id != nil {
		data.ID = types.StringValue(aws.ToString(distribution.Id))
	}
	if distribution.ARN != nil {
		data.ARN = types.StringValue(aws.ToString(distribution.ARN))
	}
	if distribution.Status != nil {
		data.Status = types.StringValue(aws.ToString(distribution.Status))
	}
	if output.ETag != nil {
		data.ETag = types.StringValue(aws.ToString(output.ETag))
	}
	if config != nil {
		if config.CallerReference != nil {
			data.CallerReference = types.StringValue(aws.ToString(config.CallerReference))
		}
		if config.Comment != nil {
			data.Comment = types.StringValue(aws.ToString(config.Comment))
		}
		if config.DefaultRootObject != nil {
			data.DefaultRootObject = types.StringValue(aws.ToString(config.DefaultRootObject))
		}
		if config.Enabled != nil {
			data.Enabled = types.BoolValue(aws.ToBool(config.Enabled))
		}
		if config.HttpVersion != "" {
			data.HTTPVersion = fwtypes.StringEnumValue(config.HttpVersion)
		}
		if config.WebACLId != nil {
			data.WebACLID = types.StringValue(aws.ToString(config.WebACLId))
		}
	}

	if config != nil && config.Origins != nil && config.Origins.Items != nil {
		var origins []*originModel
		for _, awsOrigin := range config.Origins.Items {
			var origin originModel
			diags.Append(fwflex.Flatten(ctx, awsOrigin, &origin)...)
			if diags.HasError() {
				return diags
			}

			// Handle CustomHeaders Items
			if awsOrigin.CustomHeaders != nil && awsOrigin.CustomHeaders.Items != nil {
				var customHeaders []*customHeaderModel
				for _, item := range awsOrigin.CustomHeaders.Items {
					var header customHeaderModel
					diags.Append(fwflex.Flatten(ctx, item, &header)...)
					if diags.HasError() {
						return diags
					}
					customHeaders = append(customHeaders, &header)
				}
				setVal, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, customHeaders, nil)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				origin.CustomHeader = setVal
			}

			origins = append(origins, &origin)
		}

		originsSet, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, origins, nil)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Origin = originsSet
	}

	if config != nil && config.DefaultCacheBehavior != nil {
		var behavior defaultCacheBehaviorModel
		diags.Append(fwflex.Flatten(ctx, config.DefaultCacheBehavior, &behavior)...)
		if diags.HasError() {
			return diags
		}

		// Handle AllowedMethods and CachedMethods Items
		if config.DefaultCacheBehavior.AllowedMethods != nil {
			if config.DefaultCacheBehavior.AllowedMethods.Items != nil {
				var methods []attr.Value
				for _, method := range config.DefaultCacheBehavior.AllowedMethods.Items {
					methods = append(methods, types.StringValue(string(method)))
				}
				setVal, d := fwtypes.NewSetValueOf[types.String](ctx, methods)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				behavior.AllowedMethods = setVal
			}
			if config.DefaultCacheBehavior.AllowedMethods.CachedMethods != nil && config.DefaultCacheBehavior.AllowedMethods.CachedMethods.Items != nil {
				var methods []attr.Value
				for _, method := range config.DefaultCacheBehavior.AllowedMethods.CachedMethods.Items {
					methods = append(methods, types.StringValue(string(method)))
				}
				setVal, d := fwtypes.NewSetValueOf[types.String](ctx, methods)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				behavior.CachedMethods = setVal
			}
		}

		// Handle FunctionAssociations Items
		if config.DefaultCacheBehavior.FunctionAssociations != nil && config.DefaultCacheBehavior.FunctionAssociations.Items != nil {
			var funcAssocs []*functionAssociationModel
			for _, item := range config.DefaultCacheBehavior.FunctionAssociations.Items {
				var assoc functionAssociationModel
				diags.Append(fwflex.Flatten(ctx, item, &assoc)...)
				if diags.HasError() {
					return diags
				}
				funcAssocs = append(funcAssocs, &assoc)
			}
			setVal, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, funcAssocs, nil)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			behavior.FunctionAssociation = setVal
		}

		// Handle LambdaFunctionAssociations Items
		if config.DefaultCacheBehavior.LambdaFunctionAssociations != nil && config.DefaultCacheBehavior.LambdaFunctionAssociations.Items != nil {
			var lambdaAssocs []*lambdaFunctionAssociationModel
			for _, item := range config.DefaultCacheBehavior.LambdaFunctionAssociations.Items {
				var assoc lambdaFunctionAssociationModel
				diags.Append(fwflex.Flatten(ctx, item, &assoc)...)
				if diags.HasError() {
					return diags
				}
				lambdaAssocs = append(lambdaAssocs, &assoc)
			}
			setVal, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, lambdaAssocs, nil)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			behavior.LambdaFunctionAssociation = setVal
		}

		// Handle TrustedKeyGroups Items
		if config.DefaultCacheBehavior.TrustedKeyGroups != nil && config.DefaultCacheBehavior.TrustedKeyGroups.Items != nil {
			var items []attr.Value
			for _, item := range config.DefaultCacheBehavior.TrustedKeyGroups.Items {
				items = append(items, types.StringValue(item))
			}
			listVal, d := fwtypes.NewListValueOf[types.String](ctx, items)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			behavior.TrustedKeyGroups = listVal
		}

		// Handle TrustedSigners Items
		if config.DefaultCacheBehavior.TrustedSigners != nil && config.DefaultCacheBehavior.TrustedSigners.Items != nil {
			var items []attr.Value
			for _, item := range config.DefaultCacheBehavior.TrustedSigners.Items {
				items = append(items, types.StringValue(item))
			}
			listVal, d := fwtypes.NewListValueOf[types.String](ctx, items)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			behavior.TrustedSigners = listVal
		}

		behaviorList, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, []*defaultCacheBehaviorModel{&behavior}, nil)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.DefaultCacheBehavior = behaviorList
	}

	if config != nil && config.CacheBehaviors != nil && config.CacheBehaviors.Items != nil {
		var behaviors []*cacheBehaviorModel
		for _, awsBehavior := range config.CacheBehaviors.Items {
			var behavior cacheBehaviorModel
			diags.Append(fwflex.Flatten(ctx, awsBehavior, &behavior)...)
			if diags.HasError() {
				return diags
			}

			// Handle AllowedMethods and CachedMethods Items
			if awsBehavior.AllowedMethods != nil {
				if awsBehavior.AllowedMethods.Items != nil {
					var methods []attr.Value
					for _, method := range awsBehavior.AllowedMethods.Items {
						methods = append(methods, types.StringValue(string(method)))
					}
					setVal, d := fwtypes.NewSetValueOf[types.String](ctx, methods)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}
					behavior.AllowedMethods = setVal
				}
				if awsBehavior.AllowedMethods.CachedMethods != nil && awsBehavior.AllowedMethods.CachedMethods.Items != nil {
					var methods []attr.Value
					for _, method := range awsBehavior.AllowedMethods.CachedMethods.Items {
						methods = append(methods, types.StringValue(string(method)))
					}
					setVal, d := fwtypes.NewSetValueOf[types.String](ctx, methods)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}
					behavior.CachedMethods = setVal
				}
			}

			// Handle FunctionAssociations Items
			if awsBehavior.FunctionAssociations != nil && awsBehavior.FunctionAssociations.Items != nil {
				var funcAssocs []*functionAssociationModel
				for _, item := range awsBehavior.FunctionAssociations.Items {
					var assoc functionAssociationModel
					diags.Append(fwflex.Flatten(ctx, item, &assoc)...)
					if diags.HasError() {
						return diags
					}
					funcAssocs = append(funcAssocs, &assoc)
				}
				setVal, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, funcAssocs, nil)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				behavior.FunctionAssociation = setVal
			}

			// Handle LambdaFunctionAssociations Items
			if awsBehavior.LambdaFunctionAssociations != nil && awsBehavior.LambdaFunctionAssociations.Items != nil {
				var lambdaAssocs []*lambdaFunctionAssociationModel
				for _, item := range awsBehavior.LambdaFunctionAssociations.Items {
					var assoc lambdaFunctionAssociationModel
					diags.Append(fwflex.Flatten(ctx, item, &assoc)...)
					if diags.HasError() {
						return diags
					}
					lambdaAssocs = append(lambdaAssocs, &assoc)
				}
				setVal, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, lambdaAssocs, nil)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				behavior.LambdaFunctionAssociation = setVal
			}

			// Handle TrustedKeyGroups Items
			if awsBehavior.TrustedKeyGroups != nil && awsBehavior.TrustedKeyGroups.Items != nil {
				var items []attr.Value
				for _, item := range awsBehavior.TrustedKeyGroups.Items {
					items = append(items, types.StringValue(item))
				}
				listVal, d := fwtypes.NewListValueOf[types.String](ctx, items)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				behavior.TrustedKeyGroups = listVal
			}

			// Handle TrustedSigners Items
			if awsBehavior.TrustedSigners != nil && awsBehavior.TrustedSigners.Items != nil {
				var items []attr.Value
				for _, item := range awsBehavior.TrustedSigners.Items {
					items = append(items, types.StringValue(item))
				}
				listVal, d := fwtypes.NewListValueOf[types.String](ctx, items)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				behavior.TrustedSigners = listVal
			}

			behaviors = append(behaviors, &behavior)
		}

		behaviorsList, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, behaviors, nil)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.CacheBehavior = behaviorsList
	}

	if config != nil && config.OriginGroups != nil && config.OriginGroups.Items != nil {
		var originGroups []*originGroupModel
		for _, awsOriginGroup := range config.OriginGroups.Items {
			var originGroup originGroupModel
			diags.Append(fwflex.Flatten(ctx, awsOriginGroup, &originGroup)...)
			if diags.HasError() {
				return diags
			}

			// Handle FailoverCriteria StatusCodes Items
			if awsOriginGroup.FailoverCriteria != nil && awsOriginGroup.FailoverCriteria.StatusCodes != nil && awsOriginGroup.FailoverCriteria.StatusCodes.Items != nil {
				var statusCodes []attr.Value
				for _, code := range awsOriginGroup.FailoverCriteria.StatusCodes.Items {
					statusCodes = append(statusCodes, types.Int64Value(int64(code)))
				}
				setVal, d := fwtypes.NewSetValueOf[types.Int64](ctx, statusCodes)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				var failoverCriteria failoverCriteriaModel
				failoverCriteria.StatusCodes = setVal
				listVal, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, []*failoverCriteriaModel{&failoverCriteria}, nil)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				originGroup.FailoverCriteria = listVal
			}

			// Handle Members Items
			if awsOriginGroup.Members != nil && awsOriginGroup.Members.Items != nil {
				var members []*memberModel
				for _, item := range awsOriginGroup.Members.Items {
					var member memberModel
					diags.Append(fwflex.Flatten(ctx, item, &member)...)
					if diags.HasError() {
						return diags
					}
					members = append(members, &member)
				}
				listVal, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, members, nil)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				originGroup.Member = listVal
			}

			originGroups = append(originGroups, &originGroup)
		}

		originGroupsSet, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, originGroups, nil)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.OriginGroup = originGroupsSet
	}

	if config != nil && config.Restrictions != nil {
		var restrictions restrictionsModel
		diags.Append(fwflex.Flatten(ctx, config.Restrictions, &restrictions)...)
		if diags.HasError() {
			return diags
		}

		// Handle GeoRestriction Items
		if config.Restrictions.GeoRestriction != nil && config.Restrictions.GeoRestriction.Items != nil {
			var geoRestriction geoRestrictionModel
			diags.Append(fwflex.Flatten(ctx, config.Restrictions.GeoRestriction, &geoRestriction)...)
			if diags.HasError() {
				return diags
			}

			var items []attr.Value
			for _, item := range config.Restrictions.GeoRestriction.Items {
				items = append(items, types.StringValue(item))
			}
			setVal, d := fwtypes.NewSetValueOf[types.String](ctx, items)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			geoRestriction.Locations = setVal

			listVal, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, []*geoRestrictionModel{&geoRestriction}, nil)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			restrictions.GeoRestriction = listVal
		}

		restrictionsList, d := fwtypes.NewListNestedObjectValueOfSlice(ctx, []*restrictionsModel{&restrictions}, nil)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Restrictions = restrictionsList
	}

	return diags
}
