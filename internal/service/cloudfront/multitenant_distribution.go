// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	defaultConnectionAttempts        = 3
	defaultConnectionTimeout         = 10
	defaultResponseCompletionTimeout = 30
	defaultOriginKeepaliveTimeout    = 5
	defaultOriginReadTimeout         = 30
)

// @FrameworkResource("aws_cloudfront_multitenant_distribution", name="Multi-tenant Distribution")
// @Tags(identifierAttribute="arn")
//
// Multi-tenant Distribution Limitations:
// The following fields are NOT supported for multi-tenant distributions and have been excluded from the schema:
// - ActiveTrustedSigners (use ActiveTrustedKeyGroups instead)
// - AliasICPRecordals (managed by connection groups)
// - CacheBehavior.DefaultTTL, MaxTTL, MinTTL (use cache policies instead)
// - CacheBehavior.ForwardedValues (deprecated and not supported)
// - CacheBehavior.SmoothStreaming
// - CacheBehavior.TrustedSigners (use TrustedKeyGroups instead)
// - DefaultCacheBehavior.DefaultTTL, MaxTTL, MinTTL (use cache policies instead)
// - DefaultCacheBehavior.ForwardedValues (deprecated and not supported)
// - DefaultCacheBehavior.SmoothStreaming
// - DefaultCacheBehavior.TrustedSigners (use TrustedKeyGroups instead)
// - DistributionConfig.Aliases (managed by connection groups)
// - DistributionConfig.AnycastIpListId (use connection group instead)
// - DistributionConfig.ContinuousDeploymentPolicyId
// - DistributionConfig.IsIPV6Enabled (use connection group instead)
// - DistributionConfig.PriceClass
// - DistributionConfig.Staging
// - S3OriginConfig (use origin access control instead)
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
	framework.WithTimeouts
}

func (r *multiTenantDistributionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:                      framework.ARNAttributeComputedOnly(),
			names.AttrDomainName:               schema.StringAttribute{Computed: true},
			"etag":                             schema.StringAttribute{Computed: true},
			names.AttrID:                       framework.IDAttribute(),
			"in_progress_invalidation_batches": schema.Int32Attribute{Computed: true},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrStatus: schema.StringAttribute{Computed: true},
			"caller_reference": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ConnectionMode](),
				Computed:   true,
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
			"http_version": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.HttpVersion](),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"web_acl_id": schema.StringAttribute{
				Optional: true,
				// Note: For multi-tenant distributions, this must be a WAF V2 web ACL if specified
			},
		},
		Blocks: map[string]schema.Block{
			"active_trusted_key_groups": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[activeTrustedKeyGroupsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEnabled: schema.BoolAttribute{Computed: true},
					},
					Blocks: map[string]schema.Block{
						"items": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[kgKeyPairIDsModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"key_group_id": schema.StringAttribute{Computed: true},
									"key_pair_ids": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
			},
			"custom_error_response": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customErrorResponseModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"error_caching_min_ttl": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Default:  int64default.StaticInt64(0),
						},
						"error_code": schema.Int64Attribute{
							Required: true,
						},
						"response_code": schema.StringAttribute{
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
						"cache_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"compress": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"field_level_encryption_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
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
					},
					Blocks: map[string]schema.Block{
						"allowed_methods": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[allowedMethodsModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"items": schema.SetAttribute{
										Required:   true,
										CustomType: fwtypes.SetOfStringEnumType[awstypes.Method](),
									},
									"cached_methods": schema.SetAttribute{
										Required:   true,
										CustomType: fwtypes.SetOfStringEnumType[awstypes.Method](),
									},
								},
							},
						},
						"function_association": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[functionAssociationModel](ctx),
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
						"lambda_function_association": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaFunctionAssociationModel](ctx),
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
									"lambda_function_arn": schema.StringAttribute{
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
						"cache_policy_id": schema.StringAttribute{
							Optional: true,
						},
						"compress": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"field_level_encryption_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
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
						"viewer_protocol_policy": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.ViewerProtocolPolicy](),
						},
					},
					Blocks: map[string]schema.Block{
						"allowed_methods": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[allowedMethodsModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"items": schema.SetAttribute{
										Required:   true,
										CustomType: fwtypes.SetOfStringEnumType[awstypes.Method](),
									},
									"cached_methods": schema.SetAttribute{
										Required:   true,
										CustomType: fwtypes.SetOfStringEnumType[awstypes.Method](),
									},
								},
							},
						},
						"function_association": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[functionAssociationModel](ctx),
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
						"lambda_function_association": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaFunctionAssociationModel](ctx),
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
									"lambda_function_arn": schema.StringAttribute{
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
					},
				},
			},
			"origin": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[originModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"connection_attempts": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							Default:  int32default.StaticInt32(defaultConnectionAttempts),
						},
						"connection_timeout": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							Default:  int32default.StaticInt32(defaultConnectionTimeout),
						},
						names.AttrDomainName: schema.StringAttribute{
							Required: true,
						},
						names.AttrID: schema.StringAttribute{
							Required: true,
						},
						"origin_access_control_id": schema.StringAttribute{
							Optional: true,
						},
						"origin_path": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
						"response_completion_timeout": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							Default:  int32default.StaticInt32(defaultResponseCompletionTimeout),
						},
					},
					Blocks: map[string]schema.Block{
						"custom_header": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customHeaderModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"header_name": schema.StringAttribute{
										Required: true,
									},
									"header_value": schema.StringAttribute{
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
										Default:  int32default.StaticInt32(defaultOriginKeepaliveTimeout),
									},
									"origin_read_timeout": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										Default:  int32default.StaticInt32(defaultOriginReadTimeout),
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

						"vpc_origin_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vpcOriginConfigModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"origin_keepalive_timeout": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										Default:  int32default.StaticInt32(defaultOriginKeepaliveTimeout),
									},
									"origin_read_timeout": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										Default:  int32default.StaticInt32(defaultOriginReadTimeout),
									},
									"vpc_origin_id": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"origin_group": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[originGroupModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: schema.StringAttribute{
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
															names.AttrDefaultValue: schema.StringAttribute{
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
							Computed: true,
						},
						"minimum_protocol_version": schema.StringAttribute{
							Optional:   true,
							Computed:   true,
							Default:    stringdefault.StaticString(string(awstypes.MinimumProtocolVersionTLSv1)),
							CustomType: fwtypes.StringEnumType[awstypes.MinimumProtocolVersion](),
						},
						"ssl_support_method": schema.StringAttribute{
							Optional:   true,
							Computed:   true,
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
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input.DistributionConfigWithTags.DistributionConfig)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Fix origins: CloudFront requires S3OriginConfig to be set (even if empty) when no custom or VPC origin config is specified
	// This is needed for S3 origins using Origin Access Control (OAC)
	fixOriginConfigs(input.DistributionConfigWithTags.DistributionConfig.Origins)

	// Set required computed fields that AutoFlex can't handle
	input.DistributionConfigWithTags.DistributionConfig.CallerReference = aws.String(sdkid.UniqueId())

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

	// Set ID immediately
	data.ID = types.StringValue(aws.ToString(output.Distribution.Id))
	data.ARN = types.StringValue(aws.ToString(output.Distribution.ARN))

	// Wait for distribution to be deployed
	distro, err := waitDistributionDeployed(ctx, conn, data.ID.ValueString())
	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudFront Multi-tenant Distribution (%s) create", data.ID.ValueString()), err.Error())
		return
	}

	// Read the distribution to get consistent state
	data.ETag = fwflex.StringToFramework(ctx, distro.ETag)
	response.Diagnostics.Append(fwflex.Flatten(ctx, distro.Distribution, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(fwflex.Flatten(ctx, distro.Distribution.DistributionConfig, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *multiTenantDistributionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data multiTenantDistributionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	output, err := findDistributionByID(ctx, conn, id)
	if retry.NotFound(err) {
		response.Diagnostics.AddWarning(
			"CloudFront Multi-tenant Distribution not found",
			fmt.Sprintf("CloudFront Multi-tenant Distribution (%s) not found, removing from state", id),
		)
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("reading CloudFront Multi-tenant Distribution", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Distribution, &data)...)
	if response.Diagnostics.HasError() {
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
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	// Handle tag updates first
	if !new.Tags.Equal(old.Tags) {
		if err := updateTags(ctx, conn, new.ARN.ValueString(), old.Tags, new.Tags); err != nil {
			response.Diagnostics.AddError("updating CloudFront Multi-tenant Distribution tags", err.Error())
			return
		}
	}

	// Check if distribution config needs updating (anything other than tags)
	if mtDistributionHasChanges(old, new) {
		// Get current distribution to get ETag for update
		output, err := findDistributionByID(ctx, conn, new.ID.ValueString())
		if err != nil {
			response.Diagnostics.AddError("reading CloudFront Multi-tenant Distribution for update", err.Error())
			return
		}

		// Prepare update input - start with existing config to preserve all fields
		input := cloudfront.UpdateDistributionInput{
			Id:                 new.ID.ValueStringPointer(),
			IfMatch:            output.ETag,
			DistributionConfig: output.Distribution.DistributionConfig,
		}

		// Expand the new configuration over the existing config
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input.DistributionConfig)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Fix origins: CloudFront requires S3OriginConfig to be set (even if empty) when no custom or VPC origin config is specified
		// This is needed for S3 origins using Origin Access Control (OAC)
		fixOriginConfigs(input.DistributionConfig.Origins)

		// Ensure ConnectionMode remains tenant-only
		input.DistributionConfig.ConnectionMode = awstypes.ConnectionModeTenantOnly

		// Update the distribution
		updateOutput, err := conn.UpdateDistribution(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError("updating CloudFront Multi-tenant Distribution", err.Error())
			return
		}

		// Wait for deployment if enabled
		if new.Enabled.ValueBool() {
			_, err = waitDistributionDeployed(ctx, conn, new.ID.ValueString())
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudFront Multi-tenant Distribution (%s) update", new.ID.ValueString()), err.Error())
				return
			}
		}

		// Update ETag from response
		new.ETag = fwflex.StringToFramework(ctx, updateOutput.ETag)
	}

	// Read back the updated distribution to ensure state consistency
	output, err := findDistributionByID(ctx, conn, new.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("reading CloudFront Multi-tenant Distribution after update", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Distribution, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Distribution.DistributionConfig, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

// mtDistributionHasChanges checks if distribution configuration has changed (excluding tags)
func mtDistributionHasChanges(old, new multiTenantDistributionResourceModel) bool {
	return !old.Comment.Equal(new.Comment) ||
		!old.DefaultRootObject.Equal(new.DefaultRootObject) ||
		!old.Enabled.Equal(new.Enabled) ||
		!old.HTTPVersion.Equal(new.HTTPVersion) ||
		!old.WebACLID.Equal(new.WebACLID) ||
		!old.CacheBehavior.Equal(new.CacheBehavior) ||
		!old.CustomErrorResponse.Equal(new.CustomErrorResponse) ||
		!old.DefaultCacheBehavior.Equal(new.DefaultCacheBehavior) ||
		!old.Origin.Equal(new.Origin) ||
		!old.OriginGroup.Equal(new.OriginGroup) ||
		!old.Restrictions.Equal(new.Restrictions) ||
		!old.TenantConfig.Equal(new.TenantConfig) ||
		!old.ViewerCertificate.Equal(new.ViewerCertificate)
}

func (r *multiTenantDistributionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data multiTenantDistributionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)
	id := fwflex.StringValueFromFramework(ctx, data.ID)

	// 1. Start by waiting for deployment (returns immediate if already deployed)
	if _, err := waitDistributionDeployed(ctx, conn, id); err != nil && !retry.NotFound(err) && !errs.IsA[*awstypes.NoSuchDistribution](err) {
		response.Diagnostics.AddError("waiting for CloudFront Multi-tenant Distribution deploy before delete", err.Error())
		return
	}

	// 2. Try delete
	err := deleteMultiTenantDistribution(ctx, conn, id)

	// 3. Return early for success
	if err == nil || retry.NotFound(err) || errs.IsA[*awstypes.NoSuchDistribution](err) {
		return
	}

	// 4. If not disabled error, disable, wait for deploy, delete
	if errs.IsA[*awstypes.DistributionNotDisabled](err) {
		disableErr := disableMultiTenantDistribution(ctx, conn, id)
		if retry.NotFound(disableErr) || errs.IsA[*awstypes.NoSuchDistribution](disableErr) {
			return
		}

		if disableErr != nil {
			response.Diagnostics.AddError("disabling CloudFront Multi-tenant Distribution", disableErr.Error())
			return
		}

		err = deleteMultiTenantDistribution(ctx, conn, id)
	}

	// 5. If precondition/invalidifmatchversion, retry delete
	if errs.IsA[*awstypes.PreconditionFailed](err) || errs.IsA[*awstypes.InvalidIfMatchVersion](err) {
		const timeout = 1 * time.Minute
		_, err = tfresource.RetryWhenIsOneOf2[any, *awstypes.PreconditionFailed, *awstypes.InvalidIfMatchVersion](ctx, timeout, func(ctx context.Context) (any, error) {
			return nil, deleteMultiTenantDistribution(ctx, conn, id)
		})
	}

	// 6. Return early for success
	if err == nil || retry.NotFound(err) || errs.IsA[*awstypes.NoSuchDistribution](err) {
		return
	}

	// 7. If err != nil, add error
	response.Diagnostics.AddError("deleting CloudFront Multi-tenant Distribution", err.Error())
}

func (r *multiTenantDistributionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	output, err := findDistributionByID(ctx, conn, request.ID)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Multi-tenant Distribution (%s)", request.ID), err.Error())
		return
	}

	if connectionMode := output.Distribution.DistributionConfig.ConnectionMode; connectionMode != awstypes.ConnectionModeTenantOnly {
		response.Diagnostics.AddError(fmt.Sprintf("distribution (%s) has incorrect connection mode: %s. Use the aws_cloudfront_distribution resource instead", request.ID, connectionMode), "")
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

func deleteMultiTenantDistribution(ctx context.Context, conn *cloudfront.Client, id string) error {
	etag, err := distroETag(ctx, conn, id)

	if err != nil {
		return err
	}

	input := cloudfront.DeleteDistributionInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err = conn.DeleteDistribution(ctx, &input)

	if err != nil {
		return fmt.Errorf("deleting CloudFront Multi-tenant Distribution (%s): %w", id, err)
	}

	if err := waitDistributionDeleted(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Multi-tenant Distribution (%s) delete: %w", id, err)
	}

	return nil
}

func disableMultiTenantDistribution(ctx context.Context, conn *cloudfront.Client, id string) error {
	output, err := findDistributionByID(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("reading CloudFront Multi-tenant Distribution (%s): %w", id, err)
	}

	if aws.ToString(output.Distribution.Status) == distributionStatusInProgress {
		output, err = waitDistributionDeployed(ctx, conn, id)

		if err != nil {
			return fmt.Errorf("waiting for CloudFront Multi-tenant Distribution (%s) deploy: %w", id, err)
		}
	}

	if !aws.ToBool(output.Distribution.DistributionConfig.Enabled) {
		return nil
	}

	input := cloudfront.UpdateDistributionInput{
		DistributionConfig: output.Distribution.DistributionConfig,
		Id:                 aws.String(id),
		IfMatch:            output.ETag,
	}
	input.DistributionConfig.Enabled = aws.Bool(false)

	_, err = conn.UpdateDistribution(ctx, &input)

	if err != nil {
		return fmt.Errorf("updating CloudFront Multi-tenant Distribution (%s): %w", id, err)
	}

	if _, err := waitDistributionDeployed(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Multi-tenant Distribution (%s) deploy: %w", id, err)
	}

	return nil
}

type multiTenantDistributionResourceModel struct {
	ActiveTrustedKeyGroups        fwtypes.ListNestedObjectValueOf[activeTrustedKeyGroupsModel] `tfsdk:"active_trusted_key_groups" autoflex:",xmlwrapper=Items,omitempty"`
	ARN                           types.String                                                 `tfsdk:"arn"`
	CacheBehavior                 fwtypes.ListNestedObjectValueOf[cacheBehaviorModel]          `tfsdk:"cache_behavior" autoflex:",xmlwrapper=Items,omitempty"`
	CallerReference               types.String                                                 `tfsdk:"caller_reference"`
	ConnectionMode                fwtypes.StringEnum[awstypes.ConnectionMode]                  `tfsdk:"connection_mode"`
	Comment                       types.String                                                 `tfsdk:"comment"`
	CustomErrorResponse           fwtypes.ListNestedObjectValueOf[customErrorResponseModel]    `tfsdk:"custom_error_response" autoflex:",xmlwrapper=Items,omitempty"`
	DefaultCacheBehavior          fwtypes.ListNestedObjectValueOf[defaultCacheBehaviorModel]   `tfsdk:"default_cache_behavior"`
	DefaultRootObject             types.String                                                 `tfsdk:"default_root_object" autoflex:",omitempty"`
	DomainName                    types.String                                                 `tfsdk:"domain_name"`
	Enabled                       types.Bool                                                   `tfsdk:"enabled"`
	ETag                          types.String                                                 `tfsdk:"etag"`
	HTTPVersion                   fwtypes.StringEnum[awstypes.HttpVersion]                     `tfsdk:"http_version"`
	ID                            types.String                                                 `tfsdk:"id"`
	InProgressInvalidationBatches types.Int32                                                  `tfsdk:"in_progress_invalidation_batches"`
	LastModifiedTime              timetypes.RFC3339                                            `tfsdk:"last_modified_time"`
	Origin                        fwtypes.ListNestedObjectValueOf[originModel]                 `tfsdk:"origin" autoflex:",xmlwrapper=Items"`
	OriginGroup                   fwtypes.ListNestedObjectValueOf[originGroupModel]            `tfsdk:"origin_group" autoflex:",xmlwrapper=Items,omitempty"`
	Restrictions                  fwtypes.ListNestedObjectValueOf[restrictionsModel]           `tfsdk:"restrictions"`
	Status                        types.String                                                 `tfsdk:"status"`
	Tags                          tftags.Map                                                   `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                   `tfsdk:"tags_all"`
	TenantConfig                  fwtypes.ListNestedObjectValueOf[tenantConfigModel]           `tfsdk:"tenant_config"`
	Timeouts                      timeouts.Value                                               `tfsdk:"timeouts"`
	ViewerCertificate             fwtypes.ListNestedObjectValueOf[viewerCertificateModel]      `tfsdk:"viewer_certificate"`
	WebACLID                      types.String                                                 `tfsdk:"web_acl_id" autoflex:",omitempty"`
}

type originModel struct {
	ConnectionAttempts        types.Int32                                              `tfsdk:"connection_attempts"`
	ConnectionTimeout         types.Int32                                              `tfsdk:"connection_timeout"`
	CustomHeader              fwtypes.ListNestedObjectValueOf[customHeaderModel]       `tfsdk:"custom_header" autoflex:",xmlwrapper=Items"`
	CustomOriginConfig        fwtypes.ListNestedObjectValueOf[customOriginConfigModel] `tfsdk:"custom_origin_config" autoflex:",omitempty"`
	DomainName                types.String                                             `tfsdk:"domain_name"`
	ID                        types.String                                             `tfsdk:"id"`
	OriginAccessControlID     types.String                                             `tfsdk:"origin_access_control_id" autoflex:",omitempty"`
	OriginPath                types.String                                             `tfsdk:"origin_path"`
	OriginShield              fwtypes.ListNestedObjectValueOf[originShieldModel]       `tfsdk:"origin_shield" autoflex:",omitempty"`
	ResponseCompletionTimeout types.Int32                                              `tfsdk:"response_completion_timeout"`
	VpcOriginConfig           fwtypes.ListNestedObjectValueOf[vpcOriginConfigModel]    `tfsdk:"vpc_origin_config" autoflex:",omitempty"`
}

type customHeaderModel struct {
	HeaderName  types.String `tfsdk:"header_name"`
	HeaderValue types.String `tfsdk:"header_value"`
}

type customOriginConfigModel struct {
	HTTPPort               types.Int32                                                  `tfsdk:"http_port"`
	HTTPSPort              types.Int32                                                  `tfsdk:"https_port"`
	IPAddressType          fwtypes.StringEnum[awstypes.IpAddressType]                   `tfsdk:"ip_address_type"`
	OriginKeepaliveTimeout types.Int32                                                  `tfsdk:"origin_keepalive_timeout"`
	OriginReadTimeout      types.Int32                                                  `tfsdk:"origin_read_timeout"`
	OriginProtocolPolicy   fwtypes.StringEnum[awstypes.OriginProtocolPolicy]            `tfsdk:"origin_protocol_policy"`
	OriginSSLProtocols     fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.SslProtocol]] `tfsdk:"origin_ssl_protocols" autoflex:",xmlwrapper=Items"`
}

type originShieldModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	OriginShieldRegion types.String `tfsdk:"origin_shield_region"`
}

type vpcOriginConfigModel struct {
	OriginKeepaliveTimeout types.Int32  `tfsdk:"origin_keepalive_timeout"`
	OriginReadTimeout      types.Int32  `tfsdk:"origin_read_timeout"`
	VpcOriginID            types.String `tfsdk:"vpc_origin_id"`
}

type originGroupModel struct {
	FailoverCriteria fwtypes.ListNestedObjectValueOf[failoverCriteriaModel] `tfsdk:"failover_criteria"`
	Member           fwtypes.ListNestedObjectValueOf[memberModel]           `tfsdk:"member" autoflex:",xmlwrapper=Items"`
	ID               types.String                                           `tfsdk:"id"`
}

type failoverCriteriaModel struct {
	StatusCodes fwtypes.SetValueOf[types.Int64] `tfsdk:"status_codes"`
}

type memberModel struct {
	OriginID types.String `tfsdk:"origin_id"`
}

type defaultCacheBehaviorModel struct {
	AllowedMethods            fwtypes.ListNestedObjectValueOf[allowedMethodsModel]            `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                    `tfsdk:"cache_policy_id"`
	Compress                  types.Bool                                                      `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                    `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.ListNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association" autoflex:",xmlwrapper=Items,omitempty"`
	LambdaFunctionAssociation fwtypes.ListNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association" autoflex:",xmlwrapper=Items"`
	OriginRequestPolicyID     types.String                                                    `tfsdk:"origin_request_policy_id"`
	RealtimeLogConfigARN      types.String                                                    `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                    `tfsdk:"response_headers_policy_id"`
	TargetOriginID            types.String                                                    `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListNestedObjectValueOf[trustedKeyGroupsModel]          `tfsdk:"trusted_key_groups" autoflex:",omitempty"`
	ViewerProtocolPolicy      fwtypes.StringEnum[awstypes.ViewerProtocolPolicy]               `tfsdk:"viewer_protocol_policy"`
	// Note: SmoothStreaming and TrustedSigners removed - not supported for multi-tenant distributions
}

type allowedMethodsModel struct {
	Items         fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
	CachedMethods fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.Method]] `tfsdk:"cached_methods" autoflex:",xmlwrapper=Items"`
}

type cacheBehaviorModel struct {
	AllowedMethods            fwtypes.ListNestedObjectValueOf[allowedMethodsModel]            `tfsdk:"allowed_methods"`
	CachePolicyID             types.String                                                    `tfsdk:"cache_policy_id"`
	Compress                  types.Bool                                                      `tfsdk:"compress"`
	FieldLevelEncryptionID    types.String                                                    `tfsdk:"field_level_encryption_id"`
	FunctionAssociation       fwtypes.ListNestedObjectValueOf[functionAssociationModel]       `tfsdk:"function_association" autoflex:",xmlwrapper=Items,omitempty"`
	LambdaFunctionAssociation fwtypes.ListNestedObjectValueOf[lambdaFunctionAssociationModel] `tfsdk:"lambda_function_association" autoflex:",xmlwrapper=Items"`
	OriginRequestPolicyID     types.String                                                    `tfsdk:"origin_request_policy_id"`
	PathPattern               types.String                                                    `tfsdk:"path_pattern"`
	RealtimeLogConfigARN      types.String                                                    `tfsdk:"realtime_log_config_arn"`
	ResponseHeadersPolicyID   types.String                                                    `tfsdk:"response_headers_policy_id"`
	TargetOriginID            types.String                                                    `tfsdk:"target_origin_id"`
	TrustedKeyGroups          fwtypes.ListNestedObjectValueOf[trustedKeyGroupsModel]          `tfsdk:"trusted_key_groups" autoflex:",omitempty"`
	ViewerProtocolPolicy      fwtypes.StringEnum[awstypes.ViewerProtocolPolicy]               `tfsdk:"viewer_protocol_policy"`
	// Note: SmoothStreaming and TrustedSigners removed - not supported for multi-tenant distributions
}

type customErrorResponseModel struct {
	ErrorCachingMinTtl types.Int64  `tfsdk:"error_caching_min_ttl"`
	ErrorCode          types.Int64  `tfsdk:"error_code"`
	ResponseCode       types.String `tfsdk:"response_code" autoflex:",omitempty"`
	ResponsePagePath   types.String `tfsdk:"response_page_path" autoflex:",omitempty"`
}

type restrictionsModel struct {
	GeoRestriction fwtypes.ListNestedObjectValueOf[geoRestrictionModel] `tfsdk:"geo_restriction"`
}

type geoRestrictionModel struct {
	Items           fwtypes.SetValueOf[types.String]                `tfsdk:"items" autoflex:",omitempty"`
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
	EventType         fwtypes.StringEnum[awstypes.EventType] `tfsdk:"event_type"`
	IncludeBody       types.Bool                             `tfsdk:"include_body"`
	LambdaFunctionARN types.String                           `tfsdk:"lambda_function_arn"`
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
	Items   fwtypes.ListValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=Items,omitempty"`
	Enabled types.Bool                        `tfsdk:"enabled"`
}

type activeTrustedKeyGroupsModel struct {
	Enabled types.Bool                                         `tfsdk:"enabled"`
	Items   fwtypes.ListNestedObjectValueOf[kgKeyPairIDsModel] `tfsdk:"items"`
}

type kgKeyPairIDsModel struct {
	KeyGroupID types.String                      `tfsdk:"key_group_id"`
	KeyPairIDs fwtypes.ListValueOf[types.String] `tfsdk:"key_pair_ids" autoflex:",xmlwrapper=Items"`
}

// fixOriginConfigs ensures that each origin has the required S3OriginConfig when no custom or VPC origin config is specified.
// CloudFront requires S3OriginConfig to be set (even if empty) for S3 origins using Origin Access Control (OAC).
func fixOriginConfigs(origins *awstypes.Origins) {
	if origins == nil || origins.Items == nil {
		return
	}

	for i := range origins.Items {
		origin := &origins.Items[i]
		// If custom, S3, and VPC origin configs are all missing, add an empty S3 origin config
		// One or the other must be specified, but the S3 origin can be "empty"
		if origin.CustomOriginConfig == nil && origin.S3OriginConfig == nil && origin.VpcOriginConfig == nil {
			origin.S3OriginConfig = &awstypes.S3OriginConfig{
				OriginAccessIdentity: aws.String(""),
			}
		}
	}
}
