// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mediaconnect

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediaconnect/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameFlow          = "Flow"
	subnetIdRegex        = "^subnet-[0-9a-z]*$"
	securityGroupIdRegex = "^sg-[0-9a-z]*$"
)

// @FrameworkResource("aws_mediaconnect_flow", name="Flow")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/mediaconnect;mediaconnect.DescribeFlowOutput")
func newFlowResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &flowResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type flowResource struct {
	framework.ResourceWithModel[flowResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func encryptionSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"algorithm": schema.StringAttribute{
			CustomType: fwtypes.StringEnumType[awstypes.Algorithm](),
			Required:   true,
		},
		"constant_initialization_vector": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"device_id": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"key_type": schema.StringAttribute{
			CustomType: fwtypes.StringEnumType[awstypes.KeyType](),
			Optional:   true,
			Computed:   true,
		},
		names.AttrRegion: schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		names.AttrResourceID: schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		names.AttrRoleARN: schema.StringAttribute{
			CustomType: fwtypes.ARNType,
			Required:   true,
		},
		"secret_arn": schema.StringAttribute{
			CustomType: fwtypes.ARNType,
			Optional:   true,
			Computed:   true,
		},
		names.AttrURL: schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
	}
}

func (r *flowResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAvailabilityZone: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"egress_ip": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"flow_size": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.FlowSize](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"start_flow": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"entitlement": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[entitlementModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"data_transfer_subscriber_fee_percent": schema.Int32Attribute{
							Optional: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int32{
								int32validator.Between(0, 100),
							},
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						names.AttrARN: framework.ARNAttributeComputedOnly(),
						names.AttrStatus: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.EntitlementStatus](),
							Optional:   true,
							Computed:   true,
							Default:    stringdefault.StaticString(string(awstypes.EntitlementStatusEnabled)),
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"subscriber": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
					Blocks: map[string]schema.Block{
						"encryption": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: encryptionSchemaAttributes(),
							},
						},
					},
				},
			},
			"maintenance": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[maintenanceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"day": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MaintenanceDay](),
							Required:   true,
						},
						"deadline": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"scheduled_date": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"start_hour": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"media_stream": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[mediaStreamModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"clock_rate": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int32{
								int32validator.OneOf(48000, 90000, 96000),
							},
						},
						"fmt": schema.Int32Attribute{
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						names.AttrID: schema.Int32Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MediaStreamType](),
							Required:   true,
						},
						"video_format": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrAttributes: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mediaStreamAttributesModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"lang": schema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"fmtp": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[fmtpModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"channel_order": schema.StringAttribute{
													Optional: true,
												},
												"colorimetry": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.Colorimetry](),
													Optional:   true,
												},
												"exact_framerate": schema.StringAttribute{
													Optional: true,
												},
												"par": schema.StringAttribute{
													Optional: true,
												},
												"range": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.Range](),
													Optional:   true,
												},
												"scan_mode": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.ScanMode](),
													Optional:   true,
												},
												"tcs": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.Tcs](),
													Optional:   true,
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
			"output": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[outputModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cidr_allow_list": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: fwtypes.CIDRBlockType,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						names.AttrDestination: schema.StringAttribute{
							Optional: true,
						},
						"max_latency": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"min_latency": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						names.AttrName: schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrARN: framework.ARNAttributeComputedOnly(),
						names.AttrStatus: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.OutputStatus](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrPort: schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						names.AttrProtocol: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.Protocol](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"remote_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"sender_control_port": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"smoothing_latency": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"stream_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"bridge_arn": framework.ARNAttributeComputedOnly(),
						"bridge_ports": schema.ListAttribute{
							Computed:    true,
							ElementType: types.Int32Type,
						},
						"data_transfer_subscriber_fee_percent": schema.Int32Attribute{
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"entitlement_arn": framework.ARNAttributeComputedOnly(),
						"listener_address": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"peer_ip_address": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"media_live_input_arn": framework.ARNAttributeComputedOnly(),
						"sender_ip_address": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"encryption": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: encryptionSchemaAttributes(),
							},
						},
						"media_stream_output_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mediaStreamOutputConfigModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"encoding_name": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.EncodingName](),
										Required:   true,
									},
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"destination_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[destinationConfigModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"ip": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														fwvalidators.IPv4Address(),
													},
												},
												names.AttrPort: schema.Int32Attribute{
													Required: true,
													Validators: []validator.Int32{
														int32validator.Between(1, 65535),
													},
												},
												"outbound_ip": schema.StringAttribute{
													Computed: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"interface": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[interfaceModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeBetween(1, 1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrName: schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"encoding_parameters": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[encodingParametersModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"compression_factor": schema.Float64Attribute{
													Required: true,
												},
												"encoder_profile": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.EncoderProfile](),
													Required:   true,
												},
											},
										},
									},
								},
							},
						},
						"vpc_interface_attachment": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[interfaceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrSource: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[sourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"data_transfer_subscriber_fee_percent": schema.Int32Attribute{
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						"entitlement_arn": schema.StringAttribute{
							Optional: true,
						},
						"ingest_ip": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"peer_ip_address": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"ingest_port": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"max_bitrate": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"max_latency": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"max_sync_buffer": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"min_latency": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrProtocol: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.Protocol](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"sender_control_port": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"sender_ip_address": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								fwvalidators.IPv4Address(),
							},
						},
						names.AttrARN: framework.ARNAttributeComputedOnly(),
						"listener_address": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								fwvalidators.IPv4Address(),
							},
						},
						"listener_port": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.UseStateForUnknown(),
							},
						},
						"stream_id": schema.StringAttribute{
							Optional: true,
						},
						"vpc_interface_name": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"whitelist_cidr": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								fwvalidators.IPv4CIDRNetworkAddress(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"decryption": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: encryptionSchemaAttributes(),
							},
						},
						"gateway_bridge_source": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayBridgeSourceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"vpc_interface_attachment": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[interfaceModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrName: schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"media_stream_source_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mediaStreamSourceConfigModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"encoding_name": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.EncodingName](),
										Required:   true,
									},
									names.AttrName: schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"input_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[inputConfigModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"ip": schema.StringAttribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(),
													},
													Validators: []validator.String{
														fwvalidators.IPv4Address(),
													},
												},
												names.AttrPort: schema.Int32Attribute{
													Required: true,
													Validators: []validator.Int32{
														int32validator.Between(1, 65535),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"interface": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[interfaceModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeBetween(1, 1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrName: schema.StringAttribute{
																Required: true,
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
			"source_failover_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[failoverConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"failover_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.FailoverMode](),
							Optional:   true,
							Computed:   true,
						},
						"recovery_window": schema.Int32Attribute{
							Optional: true,
							Computed: true,
						},
						names.AttrState: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.State](),
							Optional:   true,
							Computed:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"source_priority": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[sourcePriorityModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"primary_source": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"source_monitoring_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[sourceMonitoringConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"content_quality_analysis_state": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ContentQualityAnalysisState](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"thumbnail_state": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ThumbnailState](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"audio_monitoring_setting": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[audioMonitoringSettingModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"silent_audio": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[silentAudioModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrState: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.State](),
													Optional:   true,
												},
												"threshold_seconds": schema.Int32Attribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"video_monitoring_setting": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[videoMonitoringSettingModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"black_frames": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[blackFramesModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrState: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.State](),
													Optional:   true,
												},
												"threshold_seconds": schema.Int32Attribute{
													Optional: true,
												},
											},
										},
									},
									"frozen_frames": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[frozenFramesModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrState: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.State](),
													Optional:   true,
												},
												"threshold_seconds": schema.Int32Attribute{
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"vpc_interface": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcInterfaceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						"network_interface_ids": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
						"network_interface_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.NetworkInterfaceType](),
							Optional:   true,
							Computed:   true,
							Default:    stringdefault.StaticString(string(awstypes.NetworkInterfaceTypeEna)),
						},
						names.AttrRoleARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 16),
								setvalidator.ValueStringsAre(
									stringvalidator.All(
										stringvalidator.LengthAtMost(255),
										stringvalidator.RegexMatches(regexache.MustCompile(securityGroupIdRegex), "Security group ID must match regex: "+securityGroupIdRegex),
									),
								),
							},
						},
						names.AttrSubnetID: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(subnetIdRegex), "Subnet ID must match regex: "+subnetIdRegex),
							},
						},
					},
				},
			},
		},
	}
}

func (r *flowResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data flowResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MediaConnectClient(ctx)

	input := mediaconnect.CreateFlowInput{
		Name: fwflex.StringFromFramework(ctx, data.Name),
	}

	if !data.AvailabilityZone.IsNull() && !data.AvailabilityZone.IsUnknown() {
		input.AvailabilityZone = fwflex.StringFromFramework(ctx, data.AvailabilityZone)
	}

	if !data.FlowSize.IsNull() && !data.FlowSize.IsUnknown() {
		input.FlowSize = awstypes.FlowSize(data.FlowSize.ValueString())
	}

	// Expand sources.
	sourcesData, d := data.Source.ToSlice(ctx)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	var additionalSources []awstypes.SetSourceRequest
	if len(sourcesData) > 0 {
		// Determine primary source. If source_failover_config.source_priority is set,
		// use that name to find the primary; otherwise use the first source.
		primaryIdx := 0
		if !data.SourceFailoverConfig.IsNull() && !data.SourceFailoverConfig.IsUnknown() {
			fcData, d := data.SourceFailoverConfig.ToPtr(ctx)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
			if fcData != nil && !fcData.SourcePriority.IsNull() && !fcData.SourcePriority.IsUnknown() {
				spData, _ := fcData.SourcePriority.ToPtr(ctx)
				if spData != nil && !spData.PrimarySource.IsNull() {
					primaryName := spData.PrimarySource.ValueString()
					for i, s := range sourcesData {
						if s != nil && s.Name.ValueString() == primaryName {
							primaryIdx = i
							break
						}
					}
				}
			}
		}

		for i, s := range sourcesData {
			if s == nil {
				continue
			}
			var src awstypes.SetSourceRequest
			response.Diagnostics.Append(fwflex.Expand(ctx, s, &src)...)
			if response.Diagnostics.HasError() {
				return
			}
			if i == primaryIdx {
				input.Source = &src
			} else {
				additionalSources = append(additionalSources, src)
			}
		}
	}

	// Expand source failover config.
	if !data.SourceFailoverConfig.IsNull() && !data.SourceFailoverConfig.IsUnknown() {
		failoverData, d := data.SourceFailoverConfig.ToPtr(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		if failoverData != nil {
			var fc awstypes.FailoverConfig
			response.Diagnostics.Append(fwflex.Expand(ctx, failoverData, &fc)...)
			if response.Diagnostics.HasError() {
				return
			}
			input.SourceFailoverConfig = &fc
		}
	}

	// Expand maintenance.
	if !data.Maintenance.IsNull() && !data.Maintenance.IsUnknown() {
		maintData, d := data.Maintenance.ToPtr(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		if maintData != nil {
			var maint awstypes.AddMaintenance
			response.Diagnostics.Append(fwflex.Expand(ctx, maintData, &maint)...)
			if response.Diagnostics.HasError() {
				return
			}
			input.Maintenance = &maint
		}
	}

	// Expand VPC interfaces.
	if !data.VpcInterfaces.IsNull() && !data.VpcInterfaces.IsUnknown() {
		vpcData, d := data.VpcInterfaces.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		var vpcRequests []awstypes.VpcInterfaceRequest
		for _, v := range vpcData {
			if v == nil {
				continue
			}
			var req awstypes.VpcInterfaceRequest
			response.Diagnostics.Append(fwflex.Expand(ctx, v, &req)...)
			vpcRequests = append(vpcRequests, req)
		}
		if response.Diagnostics.HasError() {
			return
		}
		input.VpcInterfaces = vpcRequests
	}

	// Expand entitlements.
	if !data.Entitlements.IsNull() && !data.Entitlements.IsUnknown() {
		entData, d := data.Entitlements.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		var entRequests []awstypes.GrantEntitlementRequest
		for _, e := range entData {
			if e == nil {
				continue
			}
			var req awstypes.GrantEntitlementRequest
			response.Diagnostics.Append(fwflex.Expand(ctx, e, &req)...)
			entRequests = append(entRequests, req)
		}
		if response.Diagnostics.HasError() {
			return
		}
		input.Entitlements = entRequests
	}

	// Expand outputs.
	if !data.Outputs.IsNull() && !data.Outputs.IsUnknown() {
		outData, d := data.Outputs.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		var outRequests []awstypes.AddOutputRequest
		for _, o := range outData {
			if o == nil {
				continue
			}
			var req awstypes.AddOutputRequest
			response.Diagnostics.Append(fwflex.Expand(ctx, o, &req)...)
			outRequests = append(outRequests, req)
		}
		if response.Diagnostics.HasError() {
			return
		}
		input.Outputs = outRequests
	}

	// Expand media streams.
	if !data.MediaStreams.IsNull() && !data.MediaStreams.IsUnknown() {
		msData, d := data.MediaStreams.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		var msRequests []awstypes.AddMediaStreamRequest
		for _, ms := range msData {
			if ms == nil {
				continue
			}
			var req awstypes.AddMediaStreamRequest
			response.Diagnostics.Append(fwflex.Expand(ctx, ms, &req)...)
			msRequests = append(msRequests, req)
		}
		if response.Diagnostics.HasError() {
			return
		}
		input.MediaStreams = msRequests
	}

	// Expand source monitoring config.
	if !data.SourceMonitoringConfig.IsNull() && !data.SourceMonitoringConfig.IsUnknown() {
		smcData, d := data.SourceMonitoringConfig.ToPtr(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		if smcData != nil {
			var smc awstypes.MonitoringConfig
			response.Diagnostics.Append(fwflex.Expand(ctx, smcData, &smc)...)
			if response.Diagnostics.HasError() {
				return
			}
			input.SourceMonitoringConfig = &smc
		}
	}

	// Tags.
	input.FlowTags = getTagsIn(ctx)

	output, err := conn.CreateFlow(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating MediaConnect Flow (%s)", data.Name.ValueString()), err.Error())
		return
	}

	if output == nil || output.Flow == nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating MediaConnect Flow (%s)", data.Name.ValueString()), "empty output")
		return
	}

	// Add additional sources for multi-source flows.
	if len(additionalSources) > 0 {
		addSourcesInput := mediaconnect.AddFlowSourcesInput{
			FlowArn: output.Flow.FlowArn,
			Sources: additionalSources,
		}
		_, err = conn.AddFlowSources(ctx, &addSourcesInput)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("adding sources to MediaConnect Flow (%s)", data.Name.ValueString()), err.Error())
			return
		}
	}

	// Set ID to the flow ARN.
	data.ID = fwflex.StringToFramework(ctx, output.Flow.FlowArn)
	data.ARN = data.ID

	// Wait for STANDBY state.
	flowOutput, err := waitFlowStandby(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID)
		response.Diagnostics.AddError(fmt.Sprintf("waiting for MediaConnect Flow (%s) create", data.ID.ValueString()), err.Error())
		return
	}

	// Flatten the response.
	response.Diagnostics.Append(flattenFlow(ctx, flowOutput.Flow, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Start or stop the flow based on start_flow.
	if data.StartFlow.ValueBool() {
		if err := startFlow(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("starting MediaConnect Flow (%s)", data.ID.ValueString()), err.Error())
			return
		}
	} else if data.Status.ValueString() == string(awstypes.StatusActive) || data.Status.ValueString() == string(awstypes.StatusUpdating) {
		if err := stopFlow(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("stopping MediaConnect Flow (%s)", data.ID.ValueString()), err.Error())
			return
		}
	}
}

func (r *flowResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data flowResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MediaConnectClient(ctx)

	output, err := findFlowByARN(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading MediaConnect Flow (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(flattenFlow(ctx, output.Flow, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *flowResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var newData, oldData flowResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &newData)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &oldData)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MediaConnectClient(ctx)

	// Update flow-level attributes (flow_size, maintenance, source_failover_config, source_monitoring_config).
	if !newData.FlowSize.Equal(oldData.FlowSize) ||
		!newData.Maintenance.Equal(oldData.Maintenance) ||
		!newData.SourceFailoverConfig.Equal(oldData.SourceFailoverConfig) ||
		!newData.SourceMonitoringConfig.Equal(oldData.SourceMonitoringConfig) {
		input := mediaconnect.UpdateFlowInput{
			FlowArn: fwflex.StringFromFramework(ctx, newData.ID),
		}

		if !newData.FlowSize.IsNull() && !newData.FlowSize.IsUnknown() {
			input.FlowSize = awstypes.FlowSize(newData.FlowSize.ValueString())
		}

		if !newData.Maintenance.IsNull() && !newData.Maintenance.IsUnknown() {
			maintData, d := newData.Maintenance.ToPtr(ctx)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
			if maintData != nil {
				var maint awstypes.UpdateMaintenance
				response.Diagnostics.Append(fwflex.Expand(ctx, maintData, &maint)...)
				if response.Diagnostics.HasError() {
					return
				}
				input.Maintenance = &maint
			}
		}

		if !newData.SourceFailoverConfig.IsNull() && !newData.SourceFailoverConfig.IsUnknown() {
			failoverData, d := newData.SourceFailoverConfig.ToPtr(ctx)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
			if failoverData != nil {
				var fc awstypes.UpdateFailoverConfig
				response.Diagnostics.Append(fwflex.Expand(ctx, failoverData, &fc)...)
				if response.Diagnostics.HasError() {
					return
				}
				input.SourceFailoverConfig = &fc
			}
		}

		if !newData.SourceMonitoringConfig.IsNull() && !newData.SourceMonitoringConfig.IsUnknown() {
			smcData, d := newData.SourceMonitoringConfig.ToPtr(ctx)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
			if smcData != nil {
				var smc awstypes.MonitoringConfig
				response.Diagnostics.Append(fwflex.Expand(ctx, smcData, &smc)...)
				if response.Diagnostics.HasError() {
					return
				}
				input.SourceMonitoringConfig = &smc
			}
		}

		_, err := conn.UpdateFlow(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating MediaConnect Flow (%s)", newData.ID.ValueString()), err.Error())
			return
		}
	}

	// Update sources.
	if !newData.Source.Equal(oldData.Source) {
		newSources, d := newData.Source.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		oldSources, d := oldData.Source.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		for i, src := range newSources {
			if src == nil || i >= len(oldSources) || oldSources[i] == nil {
				continue
			}
			var srcInput mediaconnect.UpdateFlowSourceInput
			response.Diagnostics.Append(fwflex.Expand(ctx, src, &srcInput)...)
			if response.Diagnostics.HasError() {
				return
			}
			srcInput.FlowArn = fwflex.StringFromFramework(ctx, newData.ID)
			srcInput.SourceArn = fwflex.StringFromFramework(ctx, oldSources[i].SourceARN)

			_, err := conn.UpdateFlowSource(ctx, &srcInput)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("updating MediaConnect Flow (%s) source", newData.ID.ValueString()), err.Error())
				return
			}
		}
	}

	// Update entitlements.
	if !newData.Entitlements.Equal(oldData.Entitlements) {
		newEnts, d := newData.Entitlements.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		oldEnts, d := oldData.Entitlements.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		for i, ent := range newEnts {
			if ent == nil || i >= len(oldEnts) || oldEnts[i] == nil {
				continue
			}
			entInput := mediaconnect.UpdateFlowEntitlementInput{
				FlowArn:        fwflex.StringFromFramework(ctx, newData.ID),
				EntitlementArn: fwflex.StringFromFramework(ctx, oldEnts[i].EntitlementARN),
			}
			if !ent.Description.IsNull() && !ent.Description.IsUnknown() {
				entInput.Description = fwflex.StringFromFramework(ctx, ent.Description)
			}
			if !ent.EntitlementStatus.IsNull() && !ent.EntitlementStatus.IsUnknown() {
				entInput.EntitlementStatus = awstypes.EntitlementStatus(ent.EntitlementStatus.ValueString())
			}
			if !ent.Subscribers.IsNull() && !ent.Subscribers.IsUnknown() {
				entInput.Subscribers = fwflex.ExpandFrameworkStringValueList(ctx, ent.Subscribers)
			}
			if !ent.Encryption.IsNull() && !ent.Encryption.IsUnknown() {
				encData, _ := ent.Encryption.ToPtr(ctx)
				if encData != nil {
					var enc awstypes.UpdateEncryption
					response.Diagnostics.Append(fwflex.Expand(ctx, encData, &enc)...)
					if response.Diagnostics.HasError() {
						return
					}
					entInput.Encryption = &enc
				}
			}
			_, err := conn.UpdateFlowEntitlement(ctx, &entInput)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("updating MediaConnect Flow (%s) entitlement", newData.ID.ValueString()), err.Error())
				return
			}
		}
	}

	// Update outputs.
	if !newData.Outputs.Equal(oldData.Outputs) {
		newOuts, d := newData.Outputs.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		oldOuts, d := oldData.Outputs.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		for i, out := range newOuts {
			if out == nil || i >= len(oldOuts) || oldOuts[i] == nil {
				continue
			}
			outInput := mediaconnect.UpdateFlowOutputInput{
				FlowArn:   fwflex.StringFromFramework(ctx, newData.ID),
				OutputArn: fwflex.StringFromFramework(ctx, oldOuts[i].OutputARN),
			}
			if !out.Description.IsNull() && !out.Description.IsUnknown() {
				outInput.Description = fwflex.StringFromFramework(ctx, out.Description)
			}
			if !out.Destination.IsNull() && !out.Destination.IsUnknown() {
				outInput.Destination = fwflex.StringFromFramework(ctx, out.Destination)
			}
			if !out.Protocol.IsNull() && !out.Protocol.IsUnknown() {
				outInput.Protocol = awstypes.Protocol(out.Protocol.ValueString())
			}
			if !out.Port.IsNull() && !out.Port.IsUnknown() {
				outInput.Port = fwflex.Int32FromFramework(ctx, out.Port)
			}
			if !out.MaxLatency.IsNull() && !out.MaxLatency.IsUnknown() {
				outInput.MaxLatency = fwflex.Int32FromFramework(ctx, out.MaxLatency)
			}
			if !out.MinLatency.IsNull() && !out.MinLatency.IsUnknown() {
				outInput.MinLatency = fwflex.Int32FromFramework(ctx, out.MinLatency)
			}
			if !out.SmoothingLatency.IsNull() && !out.SmoothingLatency.IsUnknown() {
				outInput.SmoothingLatency = fwflex.Int32FromFramework(ctx, out.SmoothingLatency)
			}
			if !out.StreamID.IsNull() && !out.StreamID.IsUnknown() {
				outInput.StreamId = fwflex.StringFromFramework(ctx, out.StreamID)
			}
			if !out.RemoteID.IsNull() && !out.RemoteID.IsUnknown() {
				outInput.RemoteId = fwflex.StringFromFramework(ctx, out.RemoteID)
			}
			if !out.SenderControlPort.IsNull() && !out.SenderControlPort.IsUnknown() {
				outInput.SenderControlPort = fwflex.Int32FromFramework(ctx, out.SenderControlPort)
			}
			if !out.OutputStatus.IsNull() && !out.OutputStatus.IsUnknown() {
				outInput.OutputStatus = awstypes.OutputStatus(out.OutputStatus.ValueString())
			}
			if !out.SenderIPAddress.IsNull() && !out.SenderIPAddress.IsUnknown() {
				outInput.SenderIpAddress = fwflex.StringFromFramework(ctx, out.SenderIPAddress)
			}
			if !out.CIDRAllowList.IsNull() && !out.CIDRAllowList.IsUnknown() {
				outInput.CidrAllowList = fwflex.ExpandFrameworkStringValueList(ctx, out.CIDRAllowList)
			}
			if !out.MediaStreamOutputConfigurations.IsNull() && !out.MediaStreamOutputConfigurations.IsUnknown() {
				msocData, _ := out.MediaStreamOutputConfigurations.ToSlice(ctx)
				var msocRequests []awstypes.MediaStreamOutputConfigurationRequest
				for _, msoc := range msocData {
					if msoc == nil {
						continue
					}
					var req awstypes.MediaStreamOutputConfigurationRequest
					response.Diagnostics.Append(fwflex.Expand(ctx, msoc, &req)...)
					msocRequests = append(msocRequests, req)
				}
				if response.Diagnostics.HasError() {
					return
				}
				outInput.MediaStreamOutputConfigurations = msocRequests
			}
			if !out.Encryption.IsNull() && !out.Encryption.IsUnknown() {
				encData, _ := out.Encryption.ToPtr(ctx)
				if encData != nil {
					var enc awstypes.UpdateEncryption
					response.Diagnostics.Append(fwflex.Expand(ctx, encData, &enc)...)
					if response.Diagnostics.HasError() {
						return
					}
					outInput.Encryption = &enc
				}
			}
			if !out.VpcInterfaceAttachment.IsNull() && !out.VpcInterfaceAttachment.IsUnknown() {
				viaData, _ := out.VpcInterfaceAttachment.ToPtr(ctx)
				if viaData != nil && !viaData.Name.IsNull() {
					outInput.VpcInterfaceAttachment = &awstypes.VpcInterfaceAttachment{
						VpcInterfaceName: fwflex.StringFromFramework(ctx, viaData.Name),
					}
				}
			}
			_, err := conn.UpdateFlowOutput(ctx, &outInput)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("updating MediaConnect Flow (%s) output", newData.ID.ValueString()), err.Error())
				return
			}
		}
	}

	// Update media streams.
	if !newData.MediaStreams.Equal(oldData.MediaStreams) {
		newMS, d := newData.MediaStreams.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		for _, ms := range newMS {
			if ms == nil {
				continue
			}
			msInput := mediaconnect.UpdateFlowMediaStreamInput{
				FlowArn:         fwflex.StringFromFramework(ctx, newData.ID),
				MediaStreamName: fwflex.StringFromFramework(ctx, ms.MediaStreamName),
				MediaStreamType: awstypes.MediaStreamType(ms.MediaStreamType.ValueString()),
			}
			if !ms.ClockRate.IsNull() && !ms.ClockRate.IsUnknown() {
				msInput.ClockRate = fwflex.Int32FromFramework(ctx, ms.ClockRate)
			}
			if !ms.Description.IsNull() && !ms.Description.IsUnknown() {
				msInput.Description = fwflex.StringFromFramework(ctx, ms.Description)
			}
			if !ms.VideoFormat.IsNull() && !ms.VideoFormat.IsUnknown() {
				msInput.VideoFormat = fwflex.StringFromFramework(ctx, ms.VideoFormat)
			}
			if !ms.Attributes.IsNull() && !ms.Attributes.IsUnknown() {
				attrData, _ := ms.Attributes.ToPtr(ctx)
				if attrData != nil {
					var attr awstypes.MediaStreamAttributesRequest
					response.Diagnostics.Append(fwflex.Expand(ctx, attrData, &attr)...)
					if response.Diagnostics.HasError() {
						return
					}
					msInput.Attributes = &attr
				}
			}
			_, err := conn.UpdateFlowMediaStream(ctx, &msInput)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("updating MediaConnect Flow (%s) media stream", newData.ID.ValueString()), err.Error())
				return
			}
		}
	}

	// Update VPC interfaces.
	if !newData.VpcInterfaces.Equal(oldData.VpcInterfaces) {
		newVpcs, d := newData.VpcInterfaces.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		oldVpcs, d := oldData.VpcInterfaces.ToSlice(ctx)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}

		// Build maps by name.
		oldByName := make(map[string]*vpcInterfaceModel)
		for _, v := range oldVpcs {
			if v != nil {
				oldByName[v.Name.ValueString()] = v
			}
		}
		newByName := make(map[string]*vpcInterfaceModel)
		for _, v := range newVpcs {
			if v != nil {
				newByName[v.Name.ValueString()] = v
			}
		}

		// Remove interfaces that are deleted or changed.
		for name, oldV := range oldByName {
			newV, exists := newByName[name]
			if !exists || oldV.RoleARN.ValueString() != newV.RoleARN.ValueString() || oldV.SubnetID.ValueString() != newV.SubnetID.ValueString() || !oldV.NetworkInterfaceType.Equal(newV.NetworkInterfaceType) || !oldV.SecurityGroupIDs.Equal(newV.SecurityGroupIDs) {
				removeInput := mediaconnect.RemoveFlowVpcInterfaceInput{
					FlowArn:          fwflex.StringFromFramework(ctx, newData.ID),
					VpcInterfaceName: aws.String(name),
				}
				_, err := conn.RemoveFlowVpcInterface(ctx, &removeInput)
				if err != nil {
					response.Diagnostics.AddError(fmt.Sprintf("removing VPC interface (%s) from MediaConnect Flow (%s)", name, newData.ID.ValueString()), err.Error())
					return
				}
			}
		}

		// Add interfaces that are new or were removed for re-creation.
		var toAdd []awstypes.VpcInterfaceRequest
		for name, newV := range newByName {
			oldV, exists := oldByName[name]
			if !exists || oldV.RoleARN.ValueString() != newV.RoleARN.ValueString() || oldV.SubnetID.ValueString() != newV.SubnetID.ValueString() || !oldV.NetworkInterfaceType.Equal(newV.NetworkInterfaceType) || !oldV.SecurityGroupIDs.Equal(newV.SecurityGroupIDs) {
				var req awstypes.VpcInterfaceRequest
				response.Diagnostics.Append(fwflex.Expand(ctx, newV, &req)...)
				toAdd = append(toAdd, req)
			}
		}
		if response.Diagnostics.HasError() {
			return
		}
		if len(toAdd) > 0 {
			addVpcInput := mediaconnect.AddFlowVpcInterfacesInput{
				FlowArn:       fwflex.StringFromFramework(ctx, newData.ID),
				VpcInterfaces: toAdd,
			}
			_, err := conn.AddFlowVpcInterfaces(ctx, &addVpcInput)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("adding VPC interfaces to MediaConnect Flow (%s)", newData.ID.ValueString()), err.Error())
				return
			}
		}
	}

	// Wait for updates to settle.
	if err := waitFlowUpdated(ctx, conn, newData.ID.ValueString(), r.UpdateTimeout(ctx, newData.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for MediaConnect Flow (%s) update", newData.ID.ValueString()), err.Error())
		return
	}

	// Read the updated flow.
	output, err := findFlowByARN(ctx, conn, newData.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading MediaConnect Flow (%s) after update", newData.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(flattenFlow(ctx, output.Flow, &newData)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newData)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Handle start_flow changes.
	if !newData.StartFlow.Equal(oldData.StartFlow) {
		if newData.StartFlow.ValueBool() {
			if err := startFlow(ctx, conn, newData.ID.ValueString(), r.UpdateTimeout(ctx, newData.Timeouts)); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("starting MediaConnect Flow (%s)", newData.ID.ValueString()), err.Error())
				return
			}
		} else if newData.Status.ValueString() == string(awstypes.StatusActive) || newData.Status.ValueString() == string(awstypes.StatusUpdating) {
			if err := stopFlow(ctx, conn, newData.ID.ValueString(), r.UpdateTimeout(ctx, newData.Timeouts)); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("stopping MediaConnect Flow (%s)", newData.ID.ValueString()), err.Error())
				return
			}
		}
	}
}

func (r *flowResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data flowResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MediaConnectClient(ctx)
	flowARN := data.ID.ValueString()

	// Check flow status - must stop if active before deleting.
	describeOutput, err := findFlowByARN(ctx, conn, flowARN)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading MediaConnect Flow (%s) before delete", flowARN), err.Error())
		return
	}

	if describeOutput.Flow != nil && describeOutput.Flow.Status != awstypes.StatusStandby {
		switch describeOutput.Flow.Status {
		case awstypes.StatusActive:
			// Stop the flow before deleting.
			if err := stopFlow(ctx, conn, flowARN, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("stopping MediaConnect Flow (%s) before delete", flowARN), err.Error())
				return
			}
		case awstypes.StatusStarting, awstypes.StatusUpdating:
			// Wait for the transitional state to settle, then stop if needed.
			if err := waitFlowUpdated(ctx, conn, flowARN, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("waiting for MediaConnect Flow (%s) to settle before delete", flowARN), err.Error())
				return
			}
			describeOutput, err = findFlowByARN(ctx, conn, flowARN)
			if errs.IsA[*awstypes.NotFoundException](err) {
				return
			}
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("reading MediaConnect Flow (%s) before delete", flowARN), err.Error())
				return
			}
			if describeOutput.Flow != nil && describeOutput.Flow.Status == awstypes.StatusActive {
				if err := stopFlow(ctx, conn, flowARN, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
					response.Diagnostics.AddError(fmt.Sprintf("stopping MediaConnect Flow (%s) before delete", flowARN), err.Error())
					return
				}
			}
		case awstypes.StatusStopping:
			// Already stopping — just wait for STANDBY.
			if _, err := waitFlowStandby(ctx, conn, flowARN, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("waiting for MediaConnect Flow (%s) to stop before delete", flowARN), err.Error())
				return
			}
		case awstypes.StatusError:
			// Flow is in error state. Attempt delete directly — the API may
			// accept it, or return an error. There's nothing else we can do.
		}
	}

	deleteInput := mediaconnect.DeleteFlowInput{
		FlowArn: aws.String(flowARN),
	}
	_, err = conn.DeleteFlow(ctx, &deleteInput)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting MediaConnect Flow (%s)", flowARN), err.Error())
		return
	}

	if _, err := waitFlowDeleted(ctx, conn, flowARN, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for MediaConnect Flow (%s) delete", flowARN), err.Error())
		return
	}
}

// Models.

type flowResourceModel struct {
	framework.WithRegionModel
	ARN                    types.String                                                 `tfsdk:"arn"`
	AvailabilityZone       types.String                                                 `tfsdk:"availability_zone"`
	Description            types.String                                                 `tfsdk:"description"`
	EgressIP               types.String                                                 `tfsdk:"egress_ip"`
	Entitlements           fwtypes.ListNestedObjectValueOf[entitlementModel]            `tfsdk:"entitlement"`
	FlowSize               fwtypes.StringEnum[awstypes.FlowSize]                        `tfsdk:"flow_size"`
	ID                     types.String                                                 `tfsdk:"id"`
	Maintenance            fwtypes.ListNestedObjectValueOf[maintenanceModel]            `tfsdk:"maintenance"`
	MediaStreams           fwtypes.ListNestedObjectValueOf[mediaStreamModel]            `tfsdk:"media_stream"`
	Name                   types.String                                                 `tfsdk:"name"`
	Outputs                fwtypes.ListNestedObjectValueOf[outputModel]                 `tfsdk:"output"`
	Source                 fwtypes.ListNestedObjectValueOf[sourceModel]                 `tfsdk:"source"`
	SourceFailoverConfig   fwtypes.ListNestedObjectValueOf[failoverConfigModel]         `tfsdk:"source_failover_config"`
	SourceMonitoringConfig fwtypes.ListNestedObjectValueOf[sourceMonitoringConfigModel] `tfsdk:"source_monitoring_config"`
	StartFlow              types.Bool                                                   `tfsdk:"start_flow"`
	Status                 types.String                                                 `tfsdk:"status"`
	Tags                   tftags.Map                                                   `tfsdk:"tags"`
	TagsAll                tftags.Map                                                   `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                               `tfsdk:"timeouts"`
	VpcInterfaces          fwtypes.ListNestedObjectValueOf[vpcInterfaceModel]           `tfsdk:"vpc_interface"`
}

type sourceModel struct {
	DataTransferSubscriberFeePercent types.Int32                                                   `tfsdk:"data_transfer_subscriber_fee_percent"`
	Decryption                       fwtypes.ListNestedObjectValueOf[encryptionModel]              `tfsdk:"decryption"`
	Description                      types.String                                                  `tfsdk:"description"`
	EntitlementARN                   types.String                                                  `tfsdk:"entitlement_arn"`
	GatewayBridgeSource              fwtypes.ListNestedObjectValueOf[gatewayBridgeSourceModel]     `tfsdk:"gateway_bridge_source"`
	IngestIP                         types.String                                                  `tfsdk:"ingest_ip"`
	PeerIpAddress                    types.String                                                  `tfsdk:"peer_ip_address"`
	IngestPort                       types.Int32                                                   `tfsdk:"ingest_port"`
	MaxBitrate                       types.Int32                                                   `tfsdk:"max_bitrate"`
	MaxLatency                       types.Int32                                                   `tfsdk:"max_latency"`
	MaxSyncBuffer                    types.Int32                                                   `tfsdk:"max_sync_buffer"`
	MediaStreamSourceConfigurations  fwtypes.ListNestedObjectValueOf[mediaStreamSourceConfigModel] `tfsdk:"media_stream_source_configuration"`
	MinLatency                       types.Int32                                                   `tfsdk:"min_latency"`
	Name                             types.String                                                  `tfsdk:"name"`
	Protocol                         fwtypes.StringEnum[awstypes.Protocol]                         `tfsdk:"protocol"`
	SenderControlPort                types.Int32                                                   `tfsdk:"sender_control_port"`
	SenderIPAddress                  types.String                                                  `tfsdk:"sender_ip_address"`
	SourceARN                        types.String                                                  `tfsdk:"arn"`
	SourceListenerAddress            types.String                                                  `tfsdk:"listener_address"`
	SourceListenerPort               types.Int32                                                   `tfsdk:"listener_port"`
	StreamID                         types.String                                                  `tfsdk:"stream_id"`
	VpcInterfaceName                 types.String                                                  `tfsdk:"vpc_interface_name"`
	WhitelistCIDR                    types.String                                                  `tfsdk:"whitelist_cidr"`
}

type encryptionModel struct {
	Algorithm                    fwtypes.StringEnum[awstypes.Algorithm] `tfsdk:"algorithm"`
	ConstantInitializationVector types.String                           `tfsdk:"constant_initialization_vector"`
	DeviceID                     types.String                           `tfsdk:"device_id"`
	KeyType                      fwtypes.StringEnum[awstypes.KeyType]   `tfsdk:"key_type"`
	Region                       types.String                           `tfsdk:"region"`
	ResourceID                   types.String                           `tfsdk:"resource_id"`
	RoleARN                      fwtypes.ARN                            `tfsdk:"role_arn"`
	SecretARN                    fwtypes.ARN                            `tfsdk:"secret_arn"`
	URL                          types.String                           `tfsdk:"url"`
}

type failoverConfigModel struct {
	FailoverMode   fwtypes.StringEnum[awstypes.FailoverMode]            `tfsdk:"failover_mode"`
	RecoveryWindow types.Int32                                          `tfsdk:"recovery_window"`
	SourcePriority fwtypes.ListNestedObjectValueOf[sourcePriorityModel] `tfsdk:"source_priority"`
	State          fwtypes.StringEnum[awstypes.State]                   `tfsdk:"state"`
}

type sourcePriorityModel struct {
	PrimarySource types.String `tfsdk:"primary_source"`
}

type maintenanceModel struct {
	MaintenanceDay           fwtypes.StringEnum[awstypes.MaintenanceDay] `tfsdk:"day"`
	MaintenanceDeadline      types.String                                `tfsdk:"deadline"`
	MaintenanceScheduledDate types.String                                `tfsdk:"scheduled_date"`
	MaintenanceStartHour     types.String                                `tfsdk:"start_hour"`
}

type vpcInterfaceModel struct {
	Name                 types.String                                      `tfsdk:"name"`
	NetworkInterfaceIDs  types.List                                        `tfsdk:"network_interface_ids"`
	NetworkInterfaceType fwtypes.StringEnum[awstypes.NetworkInterfaceType] `tfsdk:"network_interface_type"`
	RoleARN              fwtypes.ARN                                       `tfsdk:"role_arn"`
	SecurityGroupIDs     types.Set                                         `tfsdk:"security_group_ids"`
	SubnetID             types.String                                      `tfsdk:"subnet_id"`
}

type entitlementModel struct {
	DataTransferSubscriberFeePercent types.Int32                                      `tfsdk:"data_transfer_subscriber_fee_percent"`
	Description                      types.String                                     `tfsdk:"description"`
	Encryption                       fwtypes.ListNestedObjectValueOf[encryptionModel] `tfsdk:"encryption"`
	EntitlementARN                   types.String                                     `tfsdk:"arn"`
	EntitlementStatus                fwtypes.StringEnum[awstypes.EntitlementStatus]   `tfsdk:"status"`
	Name                             types.String                                     `tfsdk:"name"`
	Subscribers                      types.List                                       `tfsdk:"subscriber"`
}

type outputModel struct {
	BridgeARN                        types.String                                                  `tfsdk:"bridge_arn"`
	BridgePorts                      types.List                                                    `tfsdk:"bridge_ports"`
	CIDRAllowList                    types.List                                                    `tfsdk:"cidr_allow_list"`
	DataTransferSubscriberFeePercent types.Int32                                                   `tfsdk:"data_transfer_subscriber_fee_percent"`
	Description                      types.String                                                  `tfsdk:"description"`
	Destination                      types.String                                                  `tfsdk:"destination"`
	Encryption                       fwtypes.ListNestedObjectValueOf[encryptionModel]              `tfsdk:"encryption"`
	EntitlementARN                   types.String                                                  `tfsdk:"entitlement_arn"`
	ListenerAddress                  types.String                                                  `tfsdk:"listener_address"`
	PeerIpAddress                    types.String                                                  `tfsdk:"peer_ip_address"`
	MaxLatency                       types.Int32                                                   `tfsdk:"max_latency"`
	MediaLiveInputARN                types.String                                                  `tfsdk:"media_live_input_arn"`
	MediaStreamOutputConfigurations  fwtypes.ListNestedObjectValueOf[mediaStreamOutputConfigModel] `tfsdk:"media_stream_output_configuration"`
	MinLatency                       types.Int32                                                   `tfsdk:"min_latency"`
	Name                             types.String                                                  `tfsdk:"name"`
	OutputARN                        types.String                                                  `tfsdk:"arn"`
	OutputStatus                     fwtypes.StringEnum[awstypes.OutputStatus]                     `tfsdk:"status"`
	Port                             types.Int32                                                   `tfsdk:"port"`
	Protocol                         fwtypes.StringEnum[awstypes.Protocol]                         `tfsdk:"protocol"`
	RemoteID                         types.String                                                  `tfsdk:"remote_id"`
	SenderControlPort                types.Int32                                                   `tfsdk:"sender_control_port"`
	SenderIPAddress                  types.String                                                  `tfsdk:"sender_ip_address"`
	SmoothingLatency                 types.Int32                                                   `tfsdk:"smoothing_latency"`
	StreamID                         types.String                                                  `tfsdk:"stream_id"`
	VpcInterfaceAttachment           fwtypes.ListNestedObjectValueOf[interfaceModel]               `tfsdk:"vpc_interface_attachment"`
}

type mediaStreamModel struct {
	Attributes      fwtypes.ListNestedObjectValueOf[mediaStreamAttributesModel] `tfsdk:"attributes"`
	ClockRate       types.Int32                                                 `tfsdk:"clock_rate"`
	Description     types.String                                                `tfsdk:"description"`
	Fmt             types.Int32                                                 `tfsdk:"fmt"`
	MediaStreamID   types.Int32                                                 `tfsdk:"id"`
	MediaStreamName types.String                                                `tfsdk:"name"`
	MediaStreamType fwtypes.StringEnum[awstypes.MediaStreamType]                `tfsdk:"type"`
	VideoFormat     types.String                                                `tfsdk:"video_format"`
}

type mediaStreamAttributesModel struct {
	Fmtp fwtypes.ListNestedObjectValueOf[fmtpModel] `tfsdk:"fmtp"`
	Lang types.String                               `tfsdk:"lang"`
}

type fmtpModel struct {
	ChannelOrder   types.String                             `tfsdk:"channel_order"`
	Colorimetry    fwtypes.StringEnum[awstypes.Colorimetry] `tfsdk:"colorimetry"`
	ExactFramerate types.String                             `tfsdk:"exact_framerate"`
	Par            types.String                             `tfsdk:"par"`
	Range          fwtypes.StringEnum[awstypes.Range]       `tfsdk:"range"`
	ScanMode       fwtypes.StringEnum[awstypes.ScanMode]    `tfsdk:"scan_mode"`
	Tcs            fwtypes.StringEnum[awstypes.Tcs]         `tfsdk:"tcs"`
}

type gatewayBridgeSourceModel struct {
	BridgeARN              fwtypes.ARN                                     `tfsdk:"arn"`
	VpcInterfaceAttachment fwtypes.ListNestedObjectValueOf[interfaceModel] `tfsdk:"vpc_interface_attachment"`
}

type interfaceModel struct {
	Name types.String `tfsdk:"name"`
}

// Compile-time assertions for TypedExpander implementations.
var (
	_ fwflex.TypedExpander = encryptionModel{}
	_ fwflex.TypedExpander = failoverConfigModel{}
	_ fwflex.TypedExpander = maintenanceModel{}
	_ fwflex.TypedExpander = gatewayBridgeSourceModel{}
	_ fwflex.TypedExpander = sourceModel{}
)

func (m encryptionModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	type noExpand encryptionModel
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.Encryption]():
		var result awstypes.Encryption
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		return &result, diags
	case reflect.TypeFor[awstypes.UpdateEncryption]():
		var result awstypes.UpdateEncryption
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		return &result, diags
	}
	return nil, diags
}

func (m failoverConfigModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	type noExpand failoverConfigModel
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.FailoverConfig]():
		var result awstypes.FailoverConfig
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		return &result, diags
	case reflect.TypeFor[awstypes.UpdateFailoverConfig]():
		var result awstypes.UpdateFailoverConfig
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		return &result, diags
	}
	return nil, diags
}

func (m maintenanceModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	type noExpand maintenanceModel
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.AddMaintenance]():
		var result awstypes.AddMaintenance
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		return &result, diags
	case reflect.TypeFor[awstypes.UpdateMaintenance]():
		var result awstypes.UpdateMaintenance
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		return &result, diags
	}
	return nil, diags
}

func (m gatewayBridgeSourceModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	type noExpand gatewayBridgeSourceModel
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.SetGatewayBridgeSourceRequest]():
		var result awstypes.SetGatewayBridgeSourceRequest
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		// Handle VpcInterfaceAttachment manually: interfaceModel.Name -> VpcInterfaceAttachment.VpcInterfaceName
		if !m.VpcInterfaceAttachment.IsNull() && !m.VpcInterfaceAttachment.IsUnknown() {
			viaData, d := m.VpcInterfaceAttachment.ToPtr(ctx)
			diags.Append(d...)
			if viaData != nil && !viaData.Name.IsNull() {
				result.VpcInterfaceAttachment = &awstypes.VpcInterfaceAttachment{
					VpcInterfaceName: fwflex.StringFromFramework(ctx, viaData.Name),
				}
			}
		}
		return &result, diags
	case reflect.TypeFor[awstypes.UpdateGatewayBridgeSourceRequest]():
		var result awstypes.UpdateGatewayBridgeSourceRequest
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		// Handle VpcInterfaceAttachment manually: interfaceModel.Name -> VpcInterfaceAttachment.VpcInterfaceName
		if !m.VpcInterfaceAttachment.IsNull() && !m.VpcInterfaceAttachment.IsUnknown() {
			viaData, d := m.VpcInterfaceAttachment.ToPtr(ctx)
			diags.Append(d...)
			if viaData != nil && !viaData.Name.IsNull() {
				result.VpcInterfaceAttachment = &awstypes.VpcInterfaceAttachment{
					VpcInterfaceName: fwflex.StringFromFramework(ctx, viaData.Name),
				}
			}
		}
		return &result, diags
	}
	return nil, diags
}

func (m sourceModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	type noExpand sourceModel
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.SetSourceRequest]():
		var result awstypes.SetSourceRequest
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		return &result, diags
	case reflect.TypeFor[mediaconnect.UpdateFlowSourceInput]():
		var result mediaconnect.UpdateFlowSourceInput
		diags.Append(fwflex.Expand(ctx, noExpand(m), &result)...)
		return &result, diags
	}
	return nil, diags
}

type mediaStreamOutputConfigModel struct {
	DestinationConfigurations fwtypes.ListNestedObjectValueOf[destinationConfigModel]  `tfsdk:"destination_configuration"`
	EncodingName              fwtypes.StringEnum[awstypes.EncodingName]                `tfsdk:"encoding_name"`
	EncodingParameters        fwtypes.ListNestedObjectValueOf[encodingParametersModel] `tfsdk:"encoding_parameters"`
	MediaStreamName           types.String                                             `tfsdk:"name"`
}

type destinationConfigModel struct {
	DestinationIP   types.String                                    `tfsdk:"ip"`
	DestinationPort types.Int32                                     `tfsdk:"port"`
	Interface       fwtypes.ListNestedObjectValueOf[interfaceModel] `tfsdk:"interface"`
	OutboundIP      types.String                                    `tfsdk:"outbound_ip"`
}

type encodingParametersModel struct {
	CompressionFactor types.Float64                               `tfsdk:"compression_factor"`
	EncoderProfile    fwtypes.StringEnum[awstypes.EncoderProfile] `tfsdk:"encoder_profile"`
}

type mediaStreamSourceConfigModel struct {
	EncodingName        fwtypes.StringEnum[awstypes.EncodingName]         `tfsdk:"encoding_name"`
	InputConfigurations fwtypes.ListNestedObjectValueOf[inputConfigModel] `tfsdk:"input_configuration"`
	MediaStreamName     types.String                                      `tfsdk:"name"`
}

type inputConfigModel struct {
	InputIP   types.String                                    `tfsdk:"ip"`
	InputPort types.Int32                                     `tfsdk:"port"`
	Interface fwtypes.ListNestedObjectValueOf[interfaceModel] `tfsdk:"interface"`
}

type sourceMonitoringConfigModel struct {
	AudioMonitoringSettings     fwtypes.ListNestedObjectValueOf[audioMonitoringSettingModel] `tfsdk:"audio_monitoring_setting"`
	ContentQualityAnalysisState fwtypes.StringEnum[awstypes.ContentQualityAnalysisState]     `tfsdk:"content_quality_analysis_state"`
	ThumbnailState              fwtypes.StringEnum[awstypes.ThumbnailState]                  `tfsdk:"thumbnail_state"`
	VideoMonitoringSettings     fwtypes.ListNestedObjectValueOf[videoMonitoringSettingModel] `tfsdk:"video_monitoring_setting"`
}

type audioMonitoringSettingModel struct {
	SilentAudio fwtypes.ListNestedObjectValueOf[silentAudioModel] `tfsdk:"silent_audio"`
}

type silentAudioModel struct {
	State            fwtypes.StringEnum[awstypes.State] `tfsdk:"state"`
	ThresholdSeconds types.Int32                        `tfsdk:"threshold_seconds"`
}

type videoMonitoringSettingModel struct {
	BlackFrames  fwtypes.ListNestedObjectValueOf[blackFramesModel]  `tfsdk:"black_frames"`
	FrozenFrames fwtypes.ListNestedObjectValueOf[frozenFramesModel] `tfsdk:"frozen_frames"`
}

type blackFramesModel struct {
	State            fwtypes.StringEnum[awstypes.State] `tfsdk:"state"`
	ThresholdSeconds types.Int32                        `tfsdk:"threshold_seconds"`
}

type frozenFramesModel struct {
	State            fwtypes.StringEnum[awstypes.State] `tfsdk:"state"`
	ThresholdSeconds types.Int32                        `tfsdk:"threshold_seconds"`
}

// Finder.

func findFlowByARN(ctx context.Context, conn *mediaconnect.Client, arn string) (*mediaconnect.DescribeFlowOutput, error) {
	input := mediaconnect.DescribeFlowInput{
		FlowArn: aws.String(arn),
	}

	output, err := conn.DescribeFlow(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Flow == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

// Status.

func statusFlow(conn *mediaconnect.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFlowByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Flow.Status), nil
	}
}

// Waiters.

func waitFlowStandby(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.StatusStarting,
			awstypes.StatusStopping,
			awstypes.StatusUpdating,
			awstypes.StatusActive,
		),
		Target: enum.Slice(
			awstypes.StatusStandby,
		),
		Refresh:                   statusFlow(conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*mediaconnect.DescribeFlowOutput); ok {
		return output, err
	}

	return nil, err
}

func waitFlowDeleted(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.StatusDeleting,
			awstypes.StatusStandby,
		),
		Target:  []string{},
		Refresh: statusFlow(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*mediaconnect.DescribeFlowOutput); ok {
		return output, err
	}

	return nil, err
}

func waitFlowActive(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.StatusStarting,
			awstypes.StatusStandby,
			awstypes.StatusUpdating,
		),
		Target: enum.Slice(
			awstypes.StatusActive,
		),
		Refresh:                   statusFlow(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*mediaconnect.DescribeFlowOutput); ok {
		return output, err
	}

	return nil, err
}

func waitFlowUpdated(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.StatusUpdating,
			awstypes.StatusStarting,
		),
		Target: enum.Slice(
			awstypes.StatusStandby,
			awstypes.StatusActive,
			awstypes.StatusError,
		),
		Refresh:                   statusFlow(conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func startFlow(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) error {
	startInput := mediaconnect.StartFlowInput{
		FlowArn: aws.String(arn),
	}
	_, err := conn.StartFlow(ctx, &startInput)
	if err != nil {
		return err
	}

	_, err = waitFlowActive(ctx, conn, arn, timeout)
	return err
}

func stopFlow(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) error {
	stopInput := mediaconnect.StopFlowInput{
		FlowArn: aws.String(arn),
	}
	_, err := conn.StopFlow(ctx, &stopInput)
	if err != nil {
		return err
	}

	_, err = waitFlowStandby(ctx, conn, arn, timeout)
	return err
}

// Flatten functions.

func flattenFlow(ctx context.Context, flow *awstypes.Flow, data *flowResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if flow == nil {
		return diags
	}

	data.ARN = fwflex.StringToFramework(ctx, flow.FlowArn)
	data.ID = data.ARN
	data.AvailabilityZone = fwflex.StringToFramework(ctx, flow.AvailabilityZone)
	data.Description = fwflex.StringToFramework(ctx, flow.Description)
	data.EgressIP = fwflex.StringToFramework(ctx, flow.EgressIp)
	data.Name = fwflex.StringToFramework(ctx, flow.Name)
	data.Status = types.StringValue(string(flow.Status))

	if flow.FlowSize != "" {
		data.FlowSize = fwtypes.StringEnumValue(flow.FlowSize)
	}

	// Flatten sources.
	if len(flow.Sources) > 0 {
		var srcModels []*sourceModel
		for _, s := range flow.Sources {
			srcModel, d := flattenSource(ctx, &s)
			diags.Append(d...)
			srcModels = append(srcModels, srcModel)
		}
		data.Source = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, srcModels)
	} else if flow.Source != nil {
		srcModel, d := flattenSource(ctx, flow.Source)
		diags.Append(d...)
		data.Source = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, srcModel)
	}

	// Flatten source failover config.
	if flow.SourceFailoverConfig != nil {
		var fcModel failoverConfigModel
		diags.Append(fwflex.Flatten(ctx, flow.SourceFailoverConfig, &fcModel)...)
		data.SourceFailoverConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &fcModel)
	} else if !data.SourceFailoverConfig.IsNull() {
		data.SourceFailoverConfig = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*failoverConfigModel{})
	}

	// Flatten maintenance.
	if flow.Maintenance != nil {
		var maintModel maintenanceModel
		diags.Append(fwflex.Flatten(ctx, flow.Maintenance, &maintModel)...)
		data.Maintenance = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &maintModel)
	} else if !data.Maintenance.IsNull() {
		data.Maintenance = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*maintenanceModel{})
	}

	// Flatten VPC interfaces.
	if len(flow.VpcInterfaces) > 0 {
		var vpcModels []*vpcInterfaceModel
		for i := range flow.VpcInterfaces {
			var m vpcInterfaceModel
			diags.Append(fwflex.Flatten(ctx, &flow.VpcInterfaces[i], &m)...)
			vpcModels = append(vpcModels, &m)
		}
		data.VpcInterfaces = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, vpcModels)
	} else if !data.VpcInterfaces.IsNull() {
		data.VpcInterfaces = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*vpcInterfaceModel{})
	}

	// Flatten entitlements.
	if len(flow.Entitlements) > 0 {
		var entModels []*entitlementModel
		for i := range flow.Entitlements {
			var m entitlementModel
			diags.Append(fwflex.Flatten(ctx, &flow.Entitlements[i], &m)...)
			entModels = append(entModels, &m)
		}
		data.Entitlements = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, entModels)
	} else if !data.Entitlements.IsNull() {
		data.Entitlements = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*entitlementModel{})
	}

	// Flatten outputs.
	if len(flow.Outputs) > 0 {
		outModels, d := flattenOutputs(ctx, flow.Outputs)
		diags.Append(d...)
		data.Outputs = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, outModels)
	} else if !data.Outputs.IsNull() {
		data.Outputs = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*outputModel{})
	}

	// Flatten media streams.
	if len(flow.MediaStreams) > 0 {
		var msModels []*mediaStreamModel
		for i := range flow.MediaStreams {
			var m mediaStreamModel
			diags.Append(fwflex.Flatten(ctx, &flow.MediaStreams[i], &m)...)
			msModels = append(msModels, &m)
		}
		data.MediaStreams = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, msModels)
	} else if !data.MediaStreams.IsNull() {
		data.MediaStreams = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*mediaStreamModel{})
	}

	// Flatten source monitoring config.
	if flow.SourceMonitoringConfig != nil {
		var smcModel sourceMonitoringConfigModel
		diags.Append(fwflex.Flatten(ctx, flow.SourceMonitoringConfig, &smcModel)...)
		data.SourceMonitoringConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &smcModel)
	} else if !data.SourceMonitoringConfig.IsNull() {
		data.SourceMonitoringConfig = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*sourceMonitoringConfigModel{})
	}

	return diags
}

func flattenSource(ctx context.Context, source *awstypes.Source) (*sourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if source == nil {
		return nil, diags
	}

	model := &sourceModel{
		DataTransferSubscriberFeePercent: fwflex.Int32ToFramework(ctx, source.DataTransferSubscriberFeePercent),
		Description:                      fwflex.StringToFramework(ctx, source.Description),
		EntitlementARN:                   fwflex.StringToFramework(ctx, source.EntitlementArn),
		IngestIP:                         fwflex.StringToFramework(ctx, source.IngestIp),
		PeerIpAddress:                    fwflex.StringToFramework(ctx, source.PeerIpAddress),
		IngestPort:                       fwflex.Int32ToFramework(ctx, source.IngestPort),
		Name:                             fwflex.StringToFramework(ctx, source.Name),
		SenderControlPort:                fwflex.Int32ToFramework(ctx, source.SenderControlPort),
		SenderIPAddress:                  fwflex.StringToFramework(ctx, source.SenderIpAddress),
		SourceARN:                        fwflex.StringToFramework(ctx, source.SourceArn),
		VpcInterfaceName:                 fwflex.StringToFramework(ctx, source.VpcInterfaceName),
		WhitelistCIDR:                    fwflex.StringToFramework(ctx, source.WhitelistCidr),
	}

	// Transport fields.
	if source.Transport != nil {
		model.Protocol = fwtypes.StringEnumValue(source.Transport.Protocol)
		model.MaxBitrate = fwflex.Int32ToFramework(ctx, source.Transport.MaxBitrate)
		model.MaxLatency = fwflex.Int32ToFramework(ctx, source.Transport.MaxLatency)
		model.MinLatency = fwflex.Int32ToFramework(ctx, source.Transport.MinLatency)
		model.MaxSyncBuffer = fwflex.Int32ToFramework(ctx, source.Transport.MaxSyncBuffer)
		model.StreamID = fwflex.StringToFramework(ctx, source.Transport.StreamId)
		model.SourceListenerAddress = fwflex.StringToFramework(ctx, source.Transport.SourceListenerAddress)
		model.SourceListenerPort = fwflex.Int32ToFramework(ctx, source.Transport.SourceListenerPort)
	}

	// Flatten decryption.
	if source.Decryption != nil {
		var encModel encryptionModel
		diags.Append(fwflex.Flatten(ctx, source.Decryption, &encModel)...)
		model.Decryption = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &encModel)
	} else {
		model.Decryption = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*encryptionModel{})
	}

	// Flatten gateway bridge source.
	if source.GatewayBridgeSource != nil {
		gbs := source.GatewayBridgeSource
		gbsModel := &gatewayBridgeSourceModel{
			BridgeARN: fwflex.StringToFrameworkARN(ctx, gbs.BridgeArn),
		}
		if gbs.VpcInterfaceAttachment != nil {
			viaModel := &interfaceModel{Name: fwflex.StringToFramework(ctx, gbs.VpcInterfaceAttachment.VpcInterfaceName)}
			gbsModel.VpcInterfaceAttachment = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, viaModel)
		} else {
			gbsModel.VpcInterfaceAttachment = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*interfaceModel{})
		}
		model.GatewayBridgeSource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, gbsModel)
	} else {
		model.GatewayBridgeSource = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*gatewayBridgeSourceModel{})
	}

	// Flatten media stream source configurations.
	if len(source.MediaStreamSourceConfigurations) > 0 {
		var mssModels []*mediaStreamSourceConfigModel
		for i := range source.MediaStreamSourceConfigurations {
			var m mediaStreamSourceConfigModel
			diags.Append(fwflex.Flatten(ctx, &source.MediaStreamSourceConfigurations[i], &m)...)
			mssModels = append(mssModels, &m)
		}
		model.MediaStreamSourceConfigurations = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, mssModels)
	} else {
		model.MediaStreamSourceConfigurations = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*mediaStreamSourceConfigModel{})
	}

	return model, diags
}

func flattenOutputs(ctx context.Context, outputs []awstypes.Output) ([]*outputModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(outputs) == 0 {
		return nil, diags
	}

	var result []*outputModel

	for _, out := range outputs {
		model := &outputModel{
			BridgeARN:                        fwflex.StringToFramework(ctx, out.BridgeArn),
			DataTransferSubscriberFeePercent: fwflex.Int32ToFramework(ctx, out.DataTransferSubscriberFeePercent),
			Description:                      fwflex.StringToFramework(ctx, out.Description),
			Destination:                      fwflex.StringToFramework(ctx, out.Destination),
			EntitlementARN:                   fwflex.StringToFramework(ctx, out.EntitlementArn),
			ListenerAddress:                  fwflex.StringToFramework(ctx, out.ListenerAddress),
			PeerIpAddress:                    fwflex.StringToFramework(ctx, out.PeerIpAddress),
			MediaLiveInputARN:                fwflex.StringToFramework(ctx, out.MediaLiveInputArn),
			Name:                             fwflex.StringToFramework(ctx, out.Name),
			OutputARN:                        fwflex.StringToFramework(ctx, out.OutputArn),
			Port:                             fwflex.Int32ToFramework(ctx, out.Port),
		}

		if len(out.BridgePorts) > 0 {
			model.BridgePorts, _ = types.ListValueFrom(ctx, types.Int32Type, out.BridgePorts)
		} else {
			model.BridgePorts, _ = types.ListValueFrom(ctx, types.Int32Type, []int32{})
		}

		if out.OutputStatus != "" {
			model.OutputStatus = fwtypes.StringEnumValue(out.OutputStatus)
		}

		// Transport fields.
		if out.Transport != nil {
			model.Protocol = fwtypes.StringEnumValue(out.Transport.Protocol)
			model.MaxLatency = fwflex.Int32ToFramework(ctx, out.Transport.MaxLatency)
			model.MinLatency = fwflex.Int32ToFramework(ctx, out.Transport.MinLatency)
			model.SmoothingLatency = fwflex.Int32ToFramework(ctx, out.Transport.SmoothingLatency)
			model.StreamID = fwflex.StringToFramework(ctx, out.Transport.StreamId)
			model.RemoteID = fwflex.StringToFramework(ctx, out.Transport.RemoteId)
			model.SenderControlPort = fwflex.Int32ToFramework(ctx, out.Transport.SenderControlPort)
			model.SenderIPAddress = fwflex.StringToFramework(ctx, out.Transport.SenderIpAddress)
			model.CIDRAllowList, _ = types.ListValueFrom(ctx, fwtypes.CIDRBlockType, out.Transport.CidrAllowList)
		} else {
			model.CIDRAllowList, _ = types.ListValueFrom(ctx, fwtypes.CIDRBlockType, []string{})
		}

		if out.VpcInterfaceAttachment != nil {
			viaModel := &interfaceModel{Name: fwflex.StringToFramework(ctx, out.VpcInterfaceAttachment.VpcInterfaceName)}
			model.VpcInterfaceAttachment = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, viaModel)
		} else {
			model.VpcInterfaceAttachment = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*interfaceModel{})
		}

		if len(out.MediaStreamOutputConfigurations) > 0 {
			var msocModels []*mediaStreamOutputConfigModel
			for i := range out.MediaStreamOutputConfigurations {
				var m mediaStreamOutputConfigModel
				diags.Append(fwflex.Flatten(ctx, &out.MediaStreamOutputConfigurations[i], &m)...)
				msocModels = append(msocModels, &m)
			}
			model.MediaStreamOutputConfigurations = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, msocModels)
		} else {
			model.MediaStreamOutputConfigurations = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*mediaStreamOutputConfigModel{})
		}

		if out.Encryption != nil {
			var encModel encryptionModel
			diags.Append(fwflex.Flatten(ctx, out.Encryption, &encModel)...)
			model.Encryption = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &encModel)
		} else {
			model.Encryption = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*encryptionModel{})
		}

		result = append(result, model)
	}

	return result, diags
}
