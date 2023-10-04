// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconnect

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediaconnect/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/exp/slices"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Flow")
// @Tags(identifierAttribute="id")
func newResourceFlow(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceFlow{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(60 * time.Minute)

	return r, nil
}

const (
	ResNameFlow          = "Flow"
	subnetIdRegex        = "^subnet-[0-9a-z]*$"
	securityGroupIdRegex = "^sg-[0-9a-z]*$"
)

type resourceFlow struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceFlow) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_mediaconnect_flow"
}

func (r *resourceFlow) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"availability_zone": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"egress_ip": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"start_flow": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"entitlement": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"arn": framework.ARNAttributeComputedOnly(),
						"data_transfer_subscriber_fee_percent": schema.Int64Attribute{
							Optional: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int64{
								int64validator.Between(0, 100),
							},
						},
						"description": schema.StringAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"status": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(string(awstypes.EntitlementStatusEnabled)),
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.EntitlementStatus](),
							},
						},
						"subscriber": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"encryption": encryptionBlock(),
					},
				},
			},
			"maintenance": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"day": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.MaintenanceDay](),
							},
						},
						"deadline": schema.StringAttribute{
							//CustomType: fwtypes.TimestampType,
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
			"media_stream": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"clock_rate": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.OneOf(48000, 90000, 96000),
							},
						},
						"description": schema.StringAttribute{
							Optional: true,
						},
						"fmt": schema.Int64Attribute{
							Computed: true,
						},
						"id": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"video_format": schema.StringAttribute{
							Optional: true,
						},
						"type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.MediaStreamType](),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"attributes": schema.ListNestedBlock{
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
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"channel_order": schema.StringAttribute{
													Optional: true,
												},
												"colorimetry": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														enum.FrameworkValidate[awstypes.Colorimetry](),
													},
												},
												"exact_framerate": schema.StringAttribute{
													Optional: true,
												},
												"par": schema.StringAttribute{
													Optional: true,
												},
												"range": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														enum.FrameworkValidate[awstypes.Range](),
													},
												},
												"scan_mode": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														enum.FrameworkValidate[awstypes.ScanMode](),
													},
												},
												"tcs": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														enum.FrameworkValidate[awstypes.Tcs](),
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
			"output": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"arn":        framework.ARNAttributeComputedOnly(),
						"bridge_arn": framework.ARNAttributeComputedOnly(),
						"bridge_ports": schema.ListAttribute{
							ElementType: types.Int64Type,
							Computed:    true,
						},
						"cidr_allow_list": schema.ListAttribute{
							ElementType: fwtypes.CIDRBlockType,
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"data_transfer_subscriber_fee_percent": schema.Int64Attribute{
							Computed: true,
							Validators: []validator.Int64{
								int64validator.Between(0, 100),
							},
						},
						"description": schema.StringAttribute{
							Optional: true,
						},
						"destination": schema.StringAttribute{
							Optional: true,
						},
						"entitlement_arn": framework.ARNAttributeComputedOnly(),
						"listener_address": schema.StringAttribute{
							Computed: true,
						},
						"max_latency": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"media_live_input_arn": framework.ARNAttributeComputedOnly(),
						"min_latency": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"port": schema.Int64Attribute{
							Optional: true,
						},
						"protocol": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.Protocol](),
							},
						},
						"remote_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"sender_control_port": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"sender_ip_address": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"smoothing_latency": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"stream_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"encryption": encryptionBlock(),
						"media_stream_output_configurations": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"encoding_name": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.EncodingName](),
										},
									},
									"name": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"destination_configurations": schema.SetNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"ip": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														fwvalidators.IPv4Address(),
													},
												},
												"outbound_ip": schema.StringAttribute{
													Computed: true,
												},
												"port": schema.Int64Attribute{
													Required: true,
													Validators: []validator.Int64{
														int64validator.Between(1, 65535),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"interface": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"name": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"encoding_parameters": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"compression_factor": schema.Float64Attribute{
													Required: true,
												},
												"encoder_profile": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														enum.FrameworkValidate[awstypes.EncoderProfile](),
													},
												},
											},
										},
									},
								},
							},
						},
						"vpc_interface_attachment": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"source": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"arn": framework.ARNAttributeComputedOnly(),
						"data_transfer_subscriber_fee_percent": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int64{
								int64validator.Between(0, 100),
							},
						},
						"description": schema.StringAttribute{
							Required: true,
						},
						"entitlement_arn": schema.StringAttribute{
							Optional: true,
						},
						"ingest_ip": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"ingest_port": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"listener_address": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								fwvalidators.IPv4Address(),
							},
						},
						"listener_port": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"max_bitrate": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"max_latency": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"max_sync_buffer": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"min_latency": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"protocol": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.Protocol](),
							},
						},
						"sender_control_port": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"sender_ip_address": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								fwvalidators.IPv4Address(),
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
						"decryption": encryptionBlock(),
						"gateway_bridge_source": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"vpc_interface_attachment": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Optional: true,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"media_stream_source_configurations": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"encoding_name": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.EncodingName](),
										},
									},
									"name": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"input_configurations": schema.SetNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"ip": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														fwvalidators.IPv4Address(),
													},
												},
												"port": schema.Int64Attribute{
													Required: true,
													Validators: []validator.Int64{
														int64validator.Between(1, 65535),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"interface": schema.ListNestedBlock{
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"name": schema.StringAttribute{
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
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"failover_mode": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.FailoverMode](),
							},
						},
						"recovery_window": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"state": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.State](),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"source_priority": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"primary_source": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"vpc_interface": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"network_interface_ids": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
						"network_interface_type": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(string(awstypes.NetworkInterfaceTypeEna)),
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.NetworkInterfaceType](),
							},
						},
						"role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						"security_group_ids": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
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
						"subnet_id": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(subnetIdRegex), "Subnet ID must match regex: "+subnetIdRegex),
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
		},
	}
}

func encryptionBlock() datasourceschema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"role_arn": schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Required:   true,
				},
				"algorithm": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						enum.FrameworkValidate[awstypes.Algorithm](),
					},
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
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						enum.FrameworkValidate[awstypes.KeyType](),
					},
				},
				"region": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"resource_id": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"secret_arn": schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Optional:   true,
					Computed:   true,
				},
				"url": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

func (r *resourceFlow) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().MediaConnectClient(ctx)

	var plan resourceFlowData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &mediaconnect.CreateFlowInput{
		Name: flex.StringFromFramework(ctx, plan.Name),
	}

	if !plan.AvailabilityZone.IsNull() {
		in.AvailabilityZone = flex.StringFromFramework(ctx, plan.AvailabilityZone)
	}
	if !plan.Entitlements.IsNull() {
		in.Entitlements = flex.ExpandFrameworkListNestedBlock(ctx, plan.Entitlements, r.expandEntitlementCreate)
	}
	if !plan.Maintenance.IsNull() {
		in.Maintenance = flex.ExpandFrameworkListNestedBlockPtr(ctx, plan.Maintenance, r.expandMaintenanceCreate)
	}
	if !plan.MediaStreams.IsNull() {
		in.MediaStreams = flex.ExpandFrameworkSetNestedBlock(ctx, plan.MediaStreams, r.expandMediaStreamCreate)
	}
	if !plan.Outputs.IsNull() {
		in.Outputs = flex.ExpandFrameworkSetNestedBlock(ctx, plan.Outputs, r.expandOutputCreate)
	}
	if !plan.Sources.IsNull() {
		in.Sources = flex.ExpandFrameworkListNestedBlock(ctx, plan.Sources, r.expandSourceCreate)
	}
	if !plan.SourceFailoverConfig.IsNull() {
		in.SourceFailoverConfig = flex.ExpandFrameworkListNestedBlockPtr(ctx, plan.SourceFailoverConfig, r.expandSourceFailoverConfigCreate)
	}
	if !plan.VpcInterfaces.IsNull() {
		in.VpcInterfaces = flex.ExpandFrameworkSetNestedBlock(ctx, plan.VpcInterfaces, r.expandVPCInterfaceCreate)
	}

	out, err := conn.CreateFlow(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionCreating, ResNameFlow, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Flow == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionCreating, ResNameFlow, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.FlowArn = flex.StringToFramework(ctx, out.Flow.FlowArn)
	plan.ID = flex.StringToFramework(ctx, out.Flow.FlowArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitFlowCreated(ctx, conn, plan.FlowArn.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForCreation, ResNameFlow, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	plan.AvailabilityZone = flex.StringToFramework(ctx, out.Flow.AvailabilityZone)
	plan.Description = flex.StringToFramework(ctx, out.Flow.Description)
	plan.EgressIp = flex.StringToFramework(ctx, out.Flow.EgressIp)
	plan.Status = flex.StringValueToFramework(ctx, out.Flow.Status)
	if len(out.Flow.Entitlements) > 0 {
		entitlements, d := r.flattenEntitlements(ctx, out.Flow.Entitlements)
		resp.Diagnostics.Append(d...)
		plan.Entitlements = entitlements
	}
	if out.Flow.Maintenance != nil {
		maintenance, d := r.flattenMaintenance(ctx, out.Flow.Maintenance)
		resp.Diagnostics.Append(d...)
		plan.Maintenance = maintenance
	}
	if len(out.Flow.MediaStreams) > 0 {
		mediaStreams, d := r.flattenMediaStreams(ctx, out.Flow.MediaStreams)
		resp.Diagnostics.Append(d...)
		plan.MediaStreams = mediaStreams
	}
	if len(out.Flow.Outputs) > 0 {
		outputs, d := r.flattenOutputs(ctx, out.Flow.Outputs)
		resp.Diagnostics.Append(d...)
		plan.Outputs = outputs
	}
	if len(out.Flow.Sources) > 0 {
		sources, d := r.flattenSources(ctx, out.Flow.Sources)
		resp.Diagnostics.Append(d...)
		plan.Sources = sources
	} else if out.Flow.Source != nil {
		sources, d := r.flattenSources(ctx, []awstypes.Source{*out.Flow.Source})
		resp.Diagnostics.Append(d...)
		plan.Sources = sources
	}
	if out.Flow.SourceFailoverConfig != nil {
		sourceFailoverConfig, d := r.flattenSourceFailoverConfig(ctx, out.Flow.SourceFailoverConfig)
		resp.Diagnostics.Append(d...)
		plan.SourceFailoverConfig = sourceFailoverConfig
	}
	if len(out.Flow.VpcInterfaces) > 0 {
		vpcInterfaces, d := r.flattenVPCInterfaces(ctx, out.Flow.VpcInterfaces)
		resp.Diagnostics.Append(d...)
		plan.VpcInterfaces = vpcInterfaces
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	pending := enum.Slice(awstypes.StatusActive, awstypes.StatusUpdating)
	if plan.StartFlow.ValueBool() {
		err = startFlow(ctx, conn, createTimeout, plan.FlowArn.ValueString())
	} else if slices.Contains(pending, plan.Status.ValueString()) {
		err = stopFlow(ctx, conn, createTimeout, plan.FlowArn.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForCreation, ResNameFlow, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceFlow) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().MediaConnectClient(ctx)

	var state resourceFlowData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFlowByARN(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionSetting, ResNameFlow, state.FlowArn.String(), err),
			err.Error(),
		)
		return
	}

	state.FlowArn = flex.StringToFramework(ctx, out.FlowArn)
	state.ID = flex.StringToFramework(ctx, out.FlowArn)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.AvailabilityZone = flex.StringToFramework(ctx, out.AvailabilityZone)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.EgressIp = flex.StringToFramework(ctx, out.EgressIp)
	entitlements, d := r.flattenEntitlements(ctx, out.Entitlements)
	resp.Diagnostics.Append(d...)
	state.Entitlements = entitlements
	maintenance, d := r.flattenMaintenance(ctx, out.Maintenance)
	resp.Diagnostics.Append(d...)
	state.Maintenance = maintenance
	mediaStreams, d := r.flattenMediaStreams(ctx, out.MediaStreams)
	resp.Diagnostics.Append(d...)
	state.MediaStreams = mediaStreams
	outputs, d := r.flattenOutputs(ctx, out.Outputs)
	resp.Diagnostics.Append(d...)
	state.Outputs = outputs
	state.Status = flex.StringValueToFramework(ctx, out.Status)
	sources, d := r.flattenSources(ctx, out.Sources)
	resp.Diagnostics.Append(d...)
	state.Sources = sources
	sourceFailoverConfig, d := r.flattenSourceFailoverConfig(ctx, out.SourceFailoverConfig)
	resp.Diagnostics.Append(d...)
	state.SourceFailoverConfig = sourceFailoverConfig
	vpcInterfaces, d := r.flattenVPCInterfaces(ctx, out.VpcInterfaces)
	resp.Diagnostics.Append(d...)
	state.VpcInterfaces = vpcInterfaces

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceFlow) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().MediaConnectClient(ctx)

	var plan, state resourceFlowData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Maintenance.Equal(state.Maintenance) || !plan.SourceFailoverConfig.Equal(state.SourceFailoverConfig) {
		in := &mediaconnect.UpdateFlowInput{
			FlowArn: aws.String(plan.FlowArn.ValueString()),
		}

		if !plan.Maintenance.IsNull() {
			in.Maintenance = flex.ExpandFrameworkListNestedBlockPtr(ctx, plan.Maintenance, r.expandMaintenanceUpdate)
		}

		if !plan.SourceFailoverConfig.IsNull() {
			in.SourceFailoverConfig = flex.ExpandFrameworkListNestedBlockPtr(ctx, plan.SourceFailoverConfig, r.expandSourceFailoverConfigUpdate)
		}

		out, err := conn.UpdateFlow(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, plan.FlowArn.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Flow == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, plan.FlowArn.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.FlowArn = flex.StringToFramework(ctx, out.Flow.FlowArn)
		plan.ID = flex.StringToFramework(ctx, out.Flow.FlowArn)
	}
	if !plan.Entitlements.Equal(state.Entitlements) {
		entitlementIns := flex.ExpandFrameworkListNestedBlock(ctx, plan.Entitlements, r.expandEntitlementUpdate)
		for _, entitlementIn := range entitlementIns {
			entitlementIn.FlowArn = aws.String(plan.FlowArn.ValueString())
			_, err := conn.UpdateFlowEntitlement(ctx, &entitlementIn)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, plan.FlowArn.String(), err),
					err.Error(),
				)
				return
			}
		}
	}
	if !plan.MediaStreams.Equal(state.MediaStreams) {
		mediaStreamIns := flex.ExpandFrameworkListNestedBlock(ctx, plan.Entitlements, r.expandMediaStreamUpdate)
		for _, mediaStreamIn := range mediaStreamIns {
			mediaStreamIn.FlowArn = aws.String(plan.FlowArn.ValueString())
			_, err := conn.UpdateFlowMediaStream(ctx, &mediaStreamIn)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, plan.FlowArn.String(), err),
					err.Error(),
				)
				return
			}
		}
	}
	if !plan.Outputs.Equal(state.Outputs) {
		outputIns := flex.ExpandFrameworkSetNestedBlock(ctx, plan.Outputs, r.expandOutputUpdate)
		for _, outputIn := range outputIns {
			outputIn.FlowArn = aws.String(plan.FlowArn.ValueString())
			_, err := conn.UpdateFlowOutput(ctx, &outputIn)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, plan.FlowArn.String(), err),
					err.Error(),
				)
				return
			}
		}
	}
	if !plan.Sources.Equal(state.Sources) {
		sourceIns := flex.ExpandFrameworkListNestedBlock(ctx, plan.Sources, r.expandSourceUpdate)
		for _, sourceIn := range sourceIns {
			sourceIn.FlowArn = aws.String(plan.FlowArn.ValueString())
			_, err := conn.UpdateFlowSource(ctx, &sourceIn)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, plan.FlowArn.String(), err),
					err.Error(),
				)
				return
			}
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitFlowUpdated(ctx, conn, plan.FlowArn.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, plan.FlowArn.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if !plan.StartFlow.Equal(state.StartFlow) {
		pending := enum.Slice(awstypes.StatusActive, awstypes.StatusUpdating)
		if plan.StartFlow.ValueBool() {
			err = startFlow(ctx, conn, updateTimeout, plan.FlowArn.ValueString())
		} else if slices.Contains(pending, state.Status.ValueString()) {
			err = stopFlow(ctx, conn, updateTimeout, plan.FlowArn.ValueString())
		}
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, plan.FlowArn.String(), err),
				err.Error(),
			)
			return
		}
	}
}

func (r *resourceFlow) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().MediaConnectClient(ctx)

	var state resourceFlowData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	pending := enum.Slice(awstypes.StatusActive, awstypes.StatusUpdating)
	if slices.Contains(pending, state.Status.ValueString()) {
		if err := stopFlow(ctx, conn, deleteTimeout, state.FlowArn.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaConnect, create.ErrActionDeleting, ResNameFlow, state.FlowArn.String(), err),
				err.Error(),
			)
			return
		}
	}

	in := &mediaconnect.DeleteFlowInput{
		FlowArn: aws.String(state.FlowArn.ValueString()),
	}

	_, err := conn.DeleteFlow(ctx, in)
	if err != nil {
		var nfe *awstypes.NotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionDeleting, ResNameFlow, state.FlowArn.String(), err),
			err.Error(),
		)
		return
	}

	_, err = waitFlowDeleted(ctx, conn, state.FlowArn.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForDeletion, ResNameFlow, state.FlowArn.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceFlow) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func startFlow(ctx context.Context, conn *mediaconnect.Client, timeout time.Duration, arn string) error {
	_, err := conn.StartFlow(ctx, &mediaconnect.StartFlowInput{
		FlowArn: aws.String(arn),
	})

	if err != nil {
		return fmt.Errorf("starting MediaConnect Flow (%s): %s", arn, err)
	}

	_, err = waitFlowStarted(ctx, conn, arn, timeout)

	if err != nil {
		return fmt.Errorf("waiting for MediaConnect Flow (%s) start: %s", arn, err)
	}

	return nil
}

func stopFlow(ctx context.Context, conn *mediaconnect.Client, timeout time.Duration, arn string) error {
	_, err := conn.StopFlow(ctx, &mediaconnect.StopFlowInput{
		FlowArn: aws.String(arn),
	})

	if err != nil {
		return fmt.Errorf("stopping MediaConnect Flow (%s): %s", arn, err)
	}

	_, err = waitFlowStopped(ctx, conn, arn, timeout)

	if err != nil {
		return fmt.Errorf("waiting for MediaConnect Flow (%s) stop: %s", arn, err)
	}

	return nil
}

func waitFlowCreated(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusUpdating),
		Target:                    enum.Slice(awstypes.StatusStandby, awstypes.StatusActive),
		Refresh:                   statusFlow(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.DescribeFlowOutput); ok {
		return out, err
	}

	return nil, err
}

func waitFlowUpdated(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusUpdating),
		Target:                    enum.Slice(awstypes.StatusStandby, awstypes.StatusActive),
		Refresh:                   statusFlow(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.DescribeFlowOutput); ok {
		return out, err
	}

	return nil, err
}

func waitFlowDeleted(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusDeleting),
		Target:  []string{},
		Refresh: statusFlow(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.DescribeFlowOutput); ok {
		return out, err
	}

	return nil, err
}

func waitFlowStarted(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusStarting),
		Target:  enum.Slice(awstypes.StatusActive),
		Refresh: statusFlow(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.DescribeFlowOutput); ok {
		return out, err
	}

	return nil, err
}

func waitFlowStopped(ctx context.Context, conn *mediaconnect.Client, arn string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusStopping),
		Target:  enum.Slice(awstypes.StatusStandby),
		Refresh: statusFlow(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.DescribeFlowOutput); ok {
		return out, err
	}

	return nil, err
}

func statusFlow(ctx context.Context, conn *mediaconnect.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findFlowByARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findFlowByARN(ctx context.Context, conn *mediaconnect.Client, arn string) (*awstypes.Flow, error) {
	in := &mediaconnect.DescribeFlowInput{
		FlowArn: aws.String(arn),
	}
	out, err := conn.DescribeFlow(ctx, in)
	if err != nil {
		var nfe *awstypes.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Flow, nil
}

func (r *resourceFlow) expandEntitlementCreate(ctx context.Context, data entitlementData) awstypes.GrantEntitlementRequest {
	var out awstypes.GrantEntitlementRequest

	if !data.Name.IsNull() {
		out.Name = flex.StringFromFramework(ctx, data.Name)
	}
	if !data.Subscribers.IsNull() {
		out.Subscribers = flex.ExpandFrameworkStringValueList(ctx, data.Subscribers)
	}
	if !data.DataTransferSubscriberFeePercent.IsNull() {
		out.DataTransferSubscriberFeePercent = int32(data.DataTransferSubscriberFeePercent.ValueInt64())
	}
	if !data.Description.IsNull() {
		out.Description = flex.StringFromFramework(ctx, data.Description)
	}
	if !data.Encryption.IsNull() {
		out.Encryption = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Encryption, r.expandEncryptionCreate)
	}
	if !data.EntitlementStatus.IsNull() {
		out.EntitlementStatus = awstypes.EntitlementStatus(data.EntitlementStatus.ValueString())
	}

	return out
}

func (r *resourceFlow) expandEntitlementUpdate(ctx context.Context, data entitlementData) mediaconnect.UpdateFlowEntitlementInput {
	var out mediaconnect.UpdateFlowEntitlementInput

	out.EntitlementArn = flex.StringFromFramework(ctx, data.EntitlementArn)
	if !data.Subscribers.IsNull() {
		out.Subscribers = flex.ExpandFrameworkStringValueList(ctx, data.Subscribers)
	}
	if !data.Description.IsNull() {
		out.Description = flex.StringFromFramework(ctx, data.Description)
	}
	if !data.Encryption.IsNull() {
		out.Encryption = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Encryption, r.expandEncryptionUpdate)
	}
	if !data.EntitlementStatus.IsNull() {
		out.EntitlementStatus = awstypes.EntitlementStatus(data.EntitlementStatus.ValueString())
	}

	return out
}

func (r *resourceFlow) expandEncryptionCreate(ctx context.Context, data encryptionData) *awstypes.Encryption {
	return &awstypes.Encryption{
		RoleArn:                      flex.ARNStringFromFramework(ctx, data.RoleArn),
		Algorithm:                    awstypes.Algorithm(data.Algorithm.ValueString()),
		ConstantInitializationVector: flex.StringFromFramework(ctx, data.ConstantInitializationVector),
		DeviceId:                     flex.StringFromFramework(ctx, data.DeviceId),
		KeyType:                      awstypes.KeyType(data.KeyType.ValueString()),
		Region:                       flex.StringFromFramework(ctx, data.Region),
		ResourceId:                   flex.StringFromFramework(ctx, data.ResourceId),
		SecretArn:                    flex.ARNStringFromFramework(ctx, data.SecretArn),
		Url:                          flex.StringFromFramework(ctx, data.Url),
	}
}

func (r *resourceFlow) expandEncryptionUpdate(ctx context.Context, data encryptionData) *awstypes.UpdateEncryption {
	return &awstypes.UpdateEncryption{
		RoleArn:                      flex.ARNStringFromFramework(ctx, data.RoleArn),
		Algorithm:                    awstypes.Algorithm(data.Algorithm.ValueString()),
		ConstantInitializationVector: flex.StringFromFramework(ctx, data.ConstantInitializationVector),
		DeviceId:                     flex.StringFromFramework(ctx, data.DeviceId),
		KeyType:                      awstypes.KeyType(data.KeyType.ValueString()),
		Region:                       flex.StringFromFramework(ctx, data.Region),
		ResourceId:                   flex.StringFromFramework(ctx, data.ResourceId),
		SecretArn:                    flex.ARNStringFromFramework(ctx, data.SecretArn),
		Url:                          flex.StringFromFramework(ctx, data.Url),
	}
}

func (r *resourceFlow) expandMaintenanceCreate(ctx context.Context, data maintenanceData) *awstypes.AddMaintenance {
	return &awstypes.AddMaintenance{
		MaintenanceDay:       awstypes.MaintenanceDay(data.MaintenanceDay.ValueString()),
		MaintenanceStartHour: flex.StringFromFramework(ctx, data.MaintenanceStartHour),
	}
}

func (r *resourceFlow) expandMaintenanceUpdate(ctx context.Context, data maintenanceData) *awstypes.UpdateMaintenance {
	return &awstypes.UpdateMaintenance{
		MaintenanceDay:       awstypes.MaintenanceDay(data.MaintenanceDay.ValueString()),
		MaintenanceStartHour: flex.StringFromFramework(ctx, data.MaintenanceStartHour),
		//MaintenanceScheduledDate: flex.StringFromFramework(ctx, data.MaintenanceScheduledDate),
	}
}

func (r *resourceFlow) expandMediaStreamCreate(ctx context.Context, data mediaStreamData) awstypes.AddMediaStreamRequest {
	return awstypes.AddMediaStreamRequest{
		MediaStreamId:   int32(data.MediaStreamId.ValueInt64()),
		MediaStreamName: flex.StringFromFramework(ctx, data.MediaStreamName),
		MediaStreamType: awstypes.MediaStreamType(data.MediaStreamType.ValueString()),
		Attributes:      flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Attributes, r.expandMediaStreamAttribute),
		ClockRate:       int32(data.ClockRate.ValueInt64()),
		Description:     flex.StringFromFramework(ctx, data.Description),
		VideoFormat:     flex.StringFromFramework(ctx, data.VideoFormat),
	}
}

func (r *resourceFlow) expandMediaStreamUpdate(ctx context.Context, data mediaStreamData) mediaconnect.UpdateFlowMediaStreamInput {
	return mediaconnect.UpdateFlowMediaStreamInput{
		MediaStreamName: flex.StringFromFramework(ctx, data.MediaStreamName),
		MediaStreamType: awstypes.MediaStreamType(data.MediaStreamType.ValueString()),
		Attributes:      flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Attributes, r.expandMediaStreamAttribute),
		ClockRate:       int32(data.ClockRate.ValueInt64()),
		Description:     flex.StringFromFramework(ctx, data.Description),
		VideoFormat:     flex.StringFromFramework(ctx, data.VideoFormat),
	}
}

func (r *resourceFlow) expandMediaStreamAttribute(ctx context.Context, data mediaStreamAttributeData) *awstypes.MediaStreamAttributesRequest {
	return &awstypes.MediaStreamAttributesRequest{
		Fmtp: flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Fmtp, r.expandMediaStreamAttributeFmtpCreate),
		Lang: flex.StringFromFramework(ctx, data.Lang),
	}
}

func (r *resourceFlow) expandMediaStreamAttributeFmtpCreate(ctx context.Context, data mediaStreamAttributeFmtpData) *awstypes.FmtpRequest {
	return &awstypes.FmtpRequest{
		ChannelOrder:   flex.StringFromFramework(ctx, data.ChannelOrder),
		Colorimetry:    awstypes.Colorimetry(data.Colorimetry.ValueString()),
		ExactFramerate: flex.StringFromFramework(ctx, data.ExactFramerate),
		Par:            flex.StringFromFramework(ctx, data.Par),
		Range:          awstypes.Range(data.Range.ValueString()),
		ScanMode:       awstypes.ScanMode(data.ScanMode.ValueString()),
		Tcs:            awstypes.Tcs(data.Tcs.ValueString()),
	}
}

func (r *resourceFlow) expandOutputCreate(ctx context.Context, data outputData) awstypes.AddOutputRequest {
	out := awstypes.AddOutputRequest{
		Name:     flex.StringFromFramework(ctx, data.Name),
		Protocol: awstypes.Protocol(data.Protocol.ValueString()),
	}

	if !data.CidrAllowList.IsNull() {
		out.CidrAllowList = flex.ExpandFrameworkStringValueList(ctx, data.CidrAllowList)
	}
	if !data.Description.IsNull() {
		out.Description = flex.StringFromFramework(ctx, data.Description)
	}
	if !data.Destination.IsNull() {
		out.Destination = flex.StringFromFramework(ctx, data.Destination)
	}
	if !data.Encryption.IsNull() {
		out.Encryption = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Encryption, r.expandEncryptionCreate)
	}
	if !data.MaxLatency.IsNull() {
		out.MaxLatency = int32(data.MaxLatency.ValueInt64())
	}
	if !data.MediaStreamOutputConfigurations.IsNull() {
		out.MediaStreamOutputConfigurations = flex.ExpandFrameworkListNestedBlock(ctx, data.MediaStreamOutputConfigurations, r.expandMediaStreamOutputConfiguration)
	}
	if !data.MinLatency.IsNull() {
		out.MinLatency = int32(data.MinLatency.ValueInt64())
	}
	if !data.Port.IsNull() {
		out.Port = int32(data.Port.ValueInt64())
	}
	if !data.RemoteId.IsNull() {
		out.RemoteId = flex.StringFromFramework(ctx, data.RemoteId)
	}
	if !data.SenderControlPort.IsNull() {
		out.SenderControlPort = int32(data.SenderControlPort.ValueInt64())
	}
	if !data.SmoothingLatency.IsNull() {
		out.SmoothingLatency = int32(data.SmoothingLatency.ValueInt64())
	}
	if !data.StreamId.IsNull() {
		out.StreamId = flex.StringFromFramework(ctx, data.StreamId)
	}
	if !data.VpcInterfaceAttachment.IsNull() {
		out.VpcInterfaceAttachment = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.VpcInterfaceAttachment, r.expandVPCInterfaceAttachment)
	}

	return out
}

func (r *resourceFlow) expandOutputUpdate(ctx context.Context, data outputData) mediaconnect.UpdateFlowOutputInput {
	out := mediaconnect.UpdateFlowOutputInput{
		OutputArn: flex.StringFromFramework(ctx, data.OutputArn),
	}

	if !data.CidrAllowList.IsNull() {
		out.CidrAllowList = flex.ExpandFrameworkStringValueList(ctx, data.CidrAllowList)
	}
	if !data.Description.IsNull() {
		out.Description = flex.StringFromFramework(ctx, data.Description)
	}
	if !data.Destination.IsNull() {
		out.Destination = flex.StringFromFramework(ctx, data.Destination)
	}
	if !data.Encryption.IsNull() {
		out.Encryption = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Encryption, r.expandEncryptionUpdate)
	}
	if !data.MaxLatency.IsNull() {
		out.MaxLatency = int32(data.MaxLatency.ValueInt64())
	}
	if !data.MediaStreamOutputConfigurations.IsNull() {
		out.MediaStreamOutputConfigurations = flex.ExpandFrameworkListNestedBlock(ctx, data.MediaStreamOutputConfigurations, r.expandMediaStreamOutputConfiguration)
	}
	if !data.MinLatency.IsNull() {
		out.MinLatency = int32(data.MinLatency.ValueInt64())
	}
	if !data.Port.IsNull() {
		out.Port = int32(data.Port.ValueInt64())
	}
	if !data.Protocol.IsNull() {
		out.Protocol = awstypes.Protocol(data.Protocol.ValueString())
	}
	if !data.RemoteId.IsNull() {
		out.RemoteId = flex.StringFromFramework(ctx, data.RemoteId)
	}
	if !data.SenderControlPort.IsNull() {
		out.SenderControlPort = int32(data.SenderControlPort.ValueInt64())
	}
	if !data.SenderIpAddress.IsNull() {
		out.SenderIpAddress = flex.StringFromFramework(ctx, data.SenderIpAddress)
	}
	if !data.SmoothingLatency.IsNull() {
		out.SmoothingLatency = int32(data.SmoothingLatency.ValueInt64())
	}
	if !data.StreamId.IsNull() {
		out.StreamId = flex.StringFromFramework(ctx, data.StreamId)
	}
	if !data.VpcInterfaceAttachment.IsNull() {
		out.VpcInterfaceAttachment = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.VpcInterfaceAttachment, r.expandVPCInterfaceAttachment)
	}

	return out
}

func (r *resourceFlow) expandMediaStreamOutputConfiguration(ctx context.Context, data mediaStreamOutputConfigurationData) awstypes.MediaStreamOutputConfigurationRequest {
	out := awstypes.MediaStreamOutputConfigurationRequest{
		EncodingName:    awstypes.EncodingName(data.EncodingName.ValueString()),
		MediaStreamName: flex.StringFromFramework(ctx, data.MediaStreamName),
	}

	if !data.DestinationConfigurations.IsNull() {
		out.DestinationConfigurations = flex.ExpandFrameworkSetNestedBlock(ctx, data.DestinationConfigurations, r.expandMediaStreamOutputConfigurationDestinationConfigurationCreate)
	}
	if !data.EncodingParameters.IsNull() {
		out.EncodingParameters = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.EncodingParameters, r.expandMediaStreamOutputConfigurationEncodingParametersCreate)
	}

	return out
}

func (r *resourceFlow) expandMediaStreamOutputConfigurationDestinationConfigurationCreate(ctx context.Context, data mediaStreamOutputConfigurationDestinationConfigurationData) awstypes.DestinationConfigurationRequest {
	return awstypes.DestinationConfigurationRequest{
		DestinationIp:   flex.StringFromFramework(ctx, data.DestinationIp),
		DestinationPort: int32(data.DestinationPort.ValueInt64()),
		Interface:       flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Interface, r.expandInterfaceCreate),
	}
}

func (r *resourceFlow) expandInterfaceCreate(ctx context.Context, data vpcInterfaceData) *awstypes.InterfaceRequest {
	return &awstypes.InterfaceRequest{
		Name: flex.StringFromFramework(ctx, data.Name),
	}
}

func (r *resourceFlow) expandMediaStreamOutputConfigurationEncodingParametersCreate(ctx context.Context, data mediaStreamOutputConfigurationEncodingParametersData) *awstypes.EncodingParametersRequest {
	return &awstypes.EncodingParametersRequest{
		CompressionFactor: data.CompressionFactor.ValueFloat64(),
		EncoderProfile:    awstypes.EncoderProfile(data.EncoderProfile.ValueString()),
	}
}

func (r *resourceFlow) expandVPCInterfaceAttachment(ctx context.Context, data vpcInterfaceAttachmentData) *awstypes.VpcInterfaceAttachment {
	return &awstypes.VpcInterfaceAttachment{
		VpcInterfaceName: flex.StringFromFramework(ctx, data.VpcInterfaceName),
	}
}

func (r *resourceFlow) expandSourceCreate(ctx context.Context, data sourceData) awstypes.SetSourceRequest {
	var out awstypes.SetSourceRequest

	if !data.Decryption.IsNull() {
		out.Decryption = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Decryption, r.expandEncryptionCreate)
	}
	if !data.Description.IsNull() {
		out.Description = flex.StringFromFramework(ctx, data.Description)
	}
	if !data.EntitlementArn.IsNull() {
		out.EntitlementArn = flex.StringFromFramework(ctx, data.EntitlementArn)
	}
	if !data.GatewayBridgeSource.IsNull() {
		out.GatewayBridgeSource = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.GatewayBridgeSource, r.expandFlowGatewayBridgeSourceCreate)
	}
	if !data.IngestPort.IsNull() {
		out.IngestPort = int32(data.IngestPort.ValueInt64())
	}
	if !data.MaxBitrate.IsNull() {
		out.MaxBitrate = int32(data.MaxBitrate.ValueInt64())
	}
	if !data.MaxLatency.IsNull() {
		out.MaxLatency = int32(data.MaxLatency.ValueInt64())
	}
	if !data.MaxSyncBuffer.IsNull() {
		out.MaxSyncBuffer = int32(data.MaxSyncBuffer.ValueInt64())
	}
	if !data.MediaStreamSourceConfigurations.IsNull() {
		out.MediaStreamSourceConfigurations = flex.ExpandFrameworkSetNestedBlock(ctx, data.MediaStreamSourceConfigurations, r.expandMediaStreamSourceConfiguration)
	}
	if !data.MinLatency.IsNull() {
		out.MinLatency = int32(data.MinLatency.ValueInt64())
	}
	if !data.Name.IsNull() {
		out.Name = flex.StringFromFramework(ctx, data.Name)
	}
	if !data.Protocol.IsNull() {
		out.Protocol = awstypes.Protocol(data.Protocol.ValueString())
	}
	if !data.SenderControlPort.IsNull() {
		out.SenderControlPort = int32(data.SenderControlPort.ValueInt64())
	}
	if !data.SenderIpAddress.IsNull() {
		out.SenderIpAddress = flex.StringFromFramework(ctx, data.SenderIpAddress)
	}
	if !data.SourceListenerAddress.IsNull() {
		out.SourceListenerAddress = flex.StringFromFramework(ctx, data.SourceListenerAddress)
	}
	if !data.SourceListenerPort.IsNull() {
		out.SourceListenerPort = int32(data.SourceListenerPort.ValueInt64())
	}
	if !data.StreamId.IsNull() {
		out.StreamId = flex.StringFromFramework(ctx, data.StreamId)
	}
	if !data.VpcInterfaceName.IsNull() {
		out.VpcInterfaceName = flex.StringFromFramework(ctx, data.VpcInterfaceName)
	}
	if !data.WhitelistCidr.IsNull() {
		out.WhitelistCidr = flex.StringFromFramework(ctx, data.WhitelistCidr)
	}

	return out
}

func (r *resourceFlow) expandSourceUpdate(ctx context.Context, data sourceData) mediaconnect.UpdateFlowSourceInput {
	var out mediaconnect.UpdateFlowSourceInput

	if !data.Decryption.IsNull() {
		out.Decryption = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Decryption, r.expandEncryptionUpdate)
	}
	if !data.Description.IsNull() {
		out.Description = flex.StringFromFramework(ctx, data.Description)
	}
	if !data.EntitlementArn.IsNull() {
		out.EntitlementArn = flex.StringFromFramework(ctx, data.EntitlementArn)
	}
	if !data.GatewayBridgeSource.IsNull() {
		out.GatewayBridgeSource = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.GatewayBridgeSource, r.expandFlowGatewayBridgeSourceUpdate)
	}
	if !data.IngestPort.IsNull() {
		out.IngestPort = int32(data.IngestPort.ValueInt64())
	}
	if !data.MaxBitrate.IsNull() {
		out.MaxBitrate = int32(data.MaxBitrate.ValueInt64())
	}
	if !data.MaxLatency.IsNull() {
		out.MaxLatency = int32(data.MaxLatency.ValueInt64())
	}
	if !data.MaxSyncBuffer.IsNull() {
		out.MaxSyncBuffer = int32(data.MaxSyncBuffer.ValueInt64())
	}
	if !data.MediaStreamSourceConfigurations.IsNull() {
		out.MediaStreamSourceConfigurations = flex.ExpandFrameworkSetNestedBlock(ctx, data.MediaStreamSourceConfigurations, r.expandMediaStreamSourceConfiguration)
	}
	if !data.MinLatency.IsNull() {
		out.MinLatency = int32(data.MinLatency.ValueInt64())
	}
	if !data.Protocol.IsNull() {
		out.Protocol = awstypes.Protocol(data.Protocol.ValueString())
	}
	if !data.SenderControlPort.IsNull() {
		out.SenderControlPort = int32(data.SenderControlPort.ValueInt64())
	}
	if !data.SenderIpAddress.IsNull() {
		out.SenderIpAddress = flex.StringFromFramework(ctx, data.SenderIpAddress)
	}
	if !data.SourceListenerAddress.IsNull() {
		out.SourceListenerAddress = flex.StringFromFramework(ctx, data.SourceListenerAddress)
	}
	if !data.SourceListenerPort.IsNull() {
		out.SourceListenerPort = int32(data.SourceListenerPort.ValueInt64())
	}
	if !data.StreamId.IsNull() {
		out.StreamId = flex.StringFromFramework(ctx, data.StreamId)
	}
	if !data.VpcInterfaceName.IsNull() {
		out.VpcInterfaceName = flex.StringFromFramework(ctx, data.VpcInterfaceName)
	}
	if !data.WhitelistCidr.IsNull() {
		out.WhitelistCidr = flex.StringFromFramework(ctx, data.WhitelistCidr)
	}

	return out
}

func (r *resourceFlow) expandFlowGatewayBridgeSourceCreate(ctx context.Context, data flowGatewayBridgeSourceData) *awstypes.SetGatewayBridgeSourceRequest {
	out := awstypes.SetGatewayBridgeSourceRequest{
		BridgeArn: flex.ARNStringFromFramework(ctx, data.BridgeArn),
	}

	if !data.VpcInterfaceAttachment.IsNull() {
		out.VpcInterfaceAttachment = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.VpcInterfaceAttachment, r.expandVPCInterfaceAttachment)
	}

	return &out
}

func (r *resourceFlow) expandFlowGatewayBridgeSourceUpdate(ctx context.Context, data flowGatewayBridgeSourceData) *awstypes.UpdateGatewayBridgeSourceRequest {
	out := awstypes.UpdateGatewayBridgeSourceRequest{
		BridgeArn: flex.ARNStringFromFramework(ctx, data.BridgeArn),
	}

	if !data.VpcInterfaceAttachment.IsNull() {
		out.VpcInterfaceAttachment = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.VpcInterfaceAttachment, r.expandVPCInterfaceAttachment)
	}

	return &out
}

func (r *resourceFlow) expandMediaStreamSourceConfiguration(ctx context.Context, data mediaStreamSourceConfigurationData) awstypes.MediaStreamSourceConfigurationRequest {
	out := awstypes.MediaStreamSourceConfigurationRequest{
		EncodingName:    awstypes.EncodingName(data.EncodingName.ValueString()),
		MediaStreamName: flex.StringFromFramework(ctx, data.MediaStreamName),
	}

	if !data.InputConfigurations.IsNull() {
		out.InputConfigurations = flex.ExpandFrameworkSetNestedBlock(ctx, data.InputConfigurations, r.expandMediaStreamSourceConfigurationInputConfigurationCreate)
	}

	return out
}

func (r *resourceFlow) expandMediaStreamSourceConfigurationInputConfigurationCreate(ctx context.Context, data mediaStreamSourceConfigurationInputConfigurationData) awstypes.InputConfigurationRequest {
	out := awstypes.InputConfigurationRequest{
		InputPort: int32(data.InputPort.ValueInt64()),
	}

	if !data.Interface.IsNull() {
		out.Interface = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Interface, r.expandInterfaceCreate)
	}

	return out
}

func (r *resourceFlow) expandSourceFailoverConfigCreate(ctx context.Context, data sourceFailoverConfigData) *awstypes.FailoverConfig {
	var out awstypes.FailoverConfig

	if !data.FailoverMode.IsNull() {
		out.FailoverMode = awstypes.FailoverMode(data.FailoverMode.ValueString())
	}
	if !data.RecoveryWindow.IsNull() {
		out.RecoveryWindow = int32(data.RecoveryWindow.ValueInt64())
	}
	if !data.SourcePriority.IsNull() {
		out.SourcePriority = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.SourcePriority, r.expandSourceFailoverConfigSourcePriority)
	}
	if !data.State.IsNull() {
		out.State = awstypes.State(data.State.ValueString())
	}

	return &out
}

func (r *resourceFlow) expandSourceFailoverConfigUpdate(ctx context.Context, data sourceFailoverConfigData) *awstypes.UpdateFailoverConfig {
	var out awstypes.UpdateFailoverConfig

	if !data.FailoverMode.IsNull() {
		out.FailoverMode = awstypes.FailoverMode(data.FailoverMode.ValueString())
	}
	if !data.RecoveryWindow.IsNull() {
		out.RecoveryWindow = int32(data.RecoveryWindow.ValueInt64())
	}
	if !data.SourcePriority.IsNull() {
		out.SourcePriority = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.SourcePriority, r.expandSourceFailoverConfigSourcePriority)
	}
	if !data.State.IsNull() {
		out.State = awstypes.State(data.State.ValueString())
	}

	return &out
}

func (r *resourceFlow) expandSourceFailoverConfigSourcePriority(ctx context.Context, data sourceFailoverConfigSourcePriorityData) *awstypes.SourcePriority {
	return &awstypes.SourcePriority{
		PrimarySource: flex.StringFromFramework(ctx, data.PrimarySource),
	}
}

func (r *resourceFlow) expandVPCInterfaceCreate(ctx context.Context, data vpcInterfaceData) awstypes.VpcInterfaceRequest {
	out := awstypes.VpcInterfaceRequest{
		Name:             flex.StringFromFramework(ctx, data.Name),
		RoleArn:          flex.ARNStringFromFramework(ctx, data.RoleArn),
		SecurityGroupIds: flex.ExpandFrameworkStringValueSet(ctx, data.SecurityGroupIds),
		SubnetId:         flex.StringFromFramework(ctx, data.SubnetId),
	}

	if !data.NetworkInterfaceType.IsNull() {
		out.NetworkInterfaceType = awstypes.NetworkInterfaceType(data.NetworkInterfaceType.ValueString())
	}

	return out
}

func (r *resourceFlow) flattenEntitlements(ctx context.Context, apiObjects []awstypes.Entitlement) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrTypes = map[string]attr.Type{
		"arn":                                  types.StringType,
		"data_transfer_subscriber_fee_percent": types.Int64Type,
		"description":                          types.StringType,
		"encryption":                           types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[encryptionData](ctx)}},

		"name":       types.StringType,
		"subscriber": types.ListType{ElemType: types.StringType},
		"status":     types.StringType,
	}

	elemType := types.ObjectType{AttrTypes: attrTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"arn":                                  flex.StringToFramework(ctx, apiObject.EntitlementArn),
			"data_transfer_subscriber_fee_percent": types.Int64Value(int64(apiObject.DataTransferSubscriberFeePercent)),
			"description":                          flex.StringToFramework(ctx, apiObject.Description),
			"name":                                 flex.StringToFramework(ctx, apiObject.Name),
			"subscriber":                           flex.FlattenFrameworkStringValueList(ctx, apiObject.Subscribers),
			"status":                               flex.StringValueToFramework(ctx, apiObject.EntitlementStatus),
		}
		encryption, d := r.flattenEncryption(ctx, apiObject.Encryption)
		diags.Append(d...)
		obj["encryption"] = encryption

		objVal, d := types.ObjectValue(attrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenEncryption(ctx context.Context, apiObject *awstypes.Encryption) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := flex.AttributeTypesMust[encryptionData](ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"algorithm":                      flex.StringValueToFramework(ctx, apiObject.Algorithm),
		"constant_initialization_vector": flex.StringToFramework(ctx, apiObject.ConstantInitializationVector),
		"device_id":                      flex.StringToFramework(ctx, apiObject.DeviceId),
		"key_type":                       flex.StringValueToFramework(ctx, apiObject.KeyType),
		"region":                         flex.StringToFramework(ctx, apiObject.Region),
		"resource_id":                    flex.StringToFramework(ctx, apiObject.ResourceId),
		"role_arn":                       flex.StringToFrameworkARN(ctx, apiObject.RoleArn, &diags),
		"secret_arn":                     flex.StringToFrameworkARN(ctx, apiObject.SecretArn, &diags),
		"url":                            flex.StringToFramework(ctx, apiObject.Url),
	}

	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenMaintenance(ctx context.Context, apiObject *awstypes.Maintenance) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := flex.AttributeTypesMust[maintenanceData](ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"day":            flex.StringValueToFramework(ctx, apiObject.MaintenanceDay),
		"deadline":       flex.StringToFrameworkLegacy(ctx, apiObject.MaintenanceDeadline),
		"scheduled_date": flex.StringToFrameworkLegacy(ctx, apiObject.MaintenanceScheduledDate),
		"start_hour":     flex.StringToFramework(ctx, apiObject.MaintenanceStartHour),
	}

	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenMediaStreams(ctx context.Context, apiObjects []awstypes.MediaStream) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrTypes = map[string]attr.Type{
		"attributes":   types.ListType{ElemType: types.ObjectType{AttrTypes: r.mediaStreamAttributesAttrTypes(ctx)}},
		"clock_rate":   types.Int64Type,
		"description":  types.StringType,
		"fmt":          types.Int64Type,
		"id":           types.Int64Type,
		"name":         types.StringType,
		"type":         types.StringType,
		"video_format": types.StringType,
	}
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if len(apiObjects) == 0 {
		return types.SetNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"clock_rate":   types.Int64Value(int64(apiObject.ClockRate)),
			"description":  flex.StringToFramework(ctx, apiObject.Description),
			"fmt":          types.Int64Value(int64(apiObject.Fmt)),
			"id":           types.Int64Value(int64(apiObject.MediaStreamId)),
			"name":         flex.StringToFramework(ctx, apiObject.MediaStreamName),
			"type":         flex.StringValueToFramework(ctx, apiObject.MediaStreamType),
			"video_format": flex.StringToFramework(ctx, apiObject.VideoFormat),
		}
		attributes, d := r.flattenMediaStreamAttributes(ctx, apiObject.Attributes)
		diags.Append(d...)
		obj["attributes"] = attributes

		objVal, d := types.ObjectValue(attrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func (r *resourceFlow) mediaStreamAttributesAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"fmtp": types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[mediaStreamAttributeFmtpData](ctx)}},
		"lang": types.StringType,
	}
}

func (r *resourceFlow) flattenMediaStreamAttributes(ctx context.Context, apiObject *awstypes.MediaStreamAttributes) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := r.mediaStreamAttributesAttrTypes(ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"lang": flex.StringToFramework(ctx, apiObject.Lang),
	}
	fmtp, d := r.flattenMediaStreamAttributeFmtp(ctx, apiObject.Fmtp)
	diags.Append(d...)
	obj["fmtp"] = fmtp

	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenMediaStreamAttributeFmtp(ctx context.Context, apiObject *awstypes.Fmtp) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := flex.AttributeTypesMust[mediaStreamAttributeFmtpData](ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"channel_order":   flex.StringToFramework(ctx, apiObject.ChannelOrder),
		"colorimetry":     flex.StringValueToFramework(ctx, apiObject.Colorimetry),
		"exact_framerate": flex.StringToFramework(ctx, apiObject.ExactFramerate),
		"par":             flex.StringToFramework(ctx, apiObject.Par),
		"range":           flex.StringValueToFramework(ctx, apiObject.Range),
		"scan_mode":       flex.StringValueToFramework(ctx, apiObject.ScanMode),
		"tcs":             flex.StringValueToFramework(ctx, apiObject.Tcs),
	}
	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenOutputBridgePorts(ctx context.Context, apiObjects []int32) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	listVal, d := types.ListValueFrom(ctx, types.Int64Type, apiObjects)
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenCIDRAllowList(ctx context.Context, apiObjects []string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	listVal, d := types.ListValueFrom(ctx, fwtypes.CIDRBlockType, apiObjects)
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenOutputs(ctx context.Context, apiObjects []awstypes.Output) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrTypes = map[string]attr.Type{
		"arn":                                  types.StringType,
		"bridge_arn":                           fwtypes.ARNType,
		"bridge_ports":                         types.ListType{ElemType: types.Int64Type},
		"cidr_allow_list":                      types.ListType{ElemType: fwtypes.CIDRBlockType},
		"data_transfer_subscriber_fee_percent": types.Int64Type,
		"description":                          types.StringType,
		"destination":                          types.StringType,
		"encryption":                           types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[encryptionData](ctx)}},
		"entitlement_arn":                      types.StringType,
		"listener_address":                     types.StringType,
		"max_latency":                          types.Int64Type,
		"media_live_input_arn":                 types.StringType,
		"media_stream_output_configurations":   types.ListType{ElemType: types.ObjectType{AttrTypes: r.mediaStreamOutputConfigurationsAttrTypes(ctx)}},
		"min_latency":                          types.Int64Type,
		"name":                                 types.StringType,
		"port":                                 types.Int64Type,
		"protocol":                             types.StringType,
		"remote_id":                            types.StringType,
		"sender_control_port":                  types.Int64Type,
		"sender_ip_address":                    types.StringType,
		"smoothing_latency":                    types.Int64Type,
		"stream_id":                            types.StringType,
		"vpc_interface_attachment":             types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[vpcInterfaceAttachmentData](ctx)}},
	}

	elemType := types.ObjectType{AttrTypes: attrTypes}

	if len(apiObjects) == 0 {
		return types.SetNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"arn":                                  flex.StringToFramework(ctx, apiObject.OutputArn),
			"bridge_arn":                           flex.StringToFrameworkARN(ctx, apiObject.BridgeArn, &diags),
			"data_transfer_subscriber_fee_percent": types.Int64Value(int64(apiObject.DataTransferSubscriberFeePercent)),
			"description":                          flex.StringToFramework(ctx, apiObject.Description),
			"destination":                          flex.StringToFramework(ctx, apiObject.Destination),
			"entitlement_arn":                      flex.StringToFramework(ctx, apiObject.EntitlementArn),
			"listener_address":                     flex.StringToFramework(ctx, apiObject.ListenerAddress),
			"media_live_input_arn":                 flex.StringToFramework(ctx, apiObject.MediaLiveInputArn),
			"name":                                 flex.StringToFramework(ctx, apiObject.Name),
			"port":                                 types.Int64Value(int64(apiObject.Port)),
		}
		bridgePorts, d := r.flattenOutputBridgePorts(ctx, apiObject.BridgePorts)
		diags.Append(d...)
		obj["bridge_ports"] = bridgePorts
		encryption, d := r.flattenEncryption(ctx, apiObject.Encryption)
		diags.Append(d...)
		obj["encryption"] = encryption
		mediaStreamOutputConfigurations, d := r.flattenMediaStreamOutputConfigurations(ctx, apiObject.MediaStreamOutputConfigurations)
		diags.Append(d...)
		obj["media_stream_output_configurations"] = mediaStreamOutputConfigurations
		vpcInterfaceAttachment, d := r.flattenVPCInterfaceAttachment(ctx, apiObject.VpcInterfaceAttachment)
		diags.Append(d...)
		obj["vpc_interface_attachment"] = vpcInterfaceAttachment

		// Set unknowns
		if apiObject.Transport != nil {
			cidrAllowList, d := r.flattenCIDRAllowList(ctx, apiObject.Transport.CidrAllowList)
			diags.Append(d...)
			obj["cidr_allow_list"] = cidrAllowList
			//obj["cidr_allow_list"] = flex.FlattenFrameworkStringValueList(ctx, apiObject.Transport.CidrAllowList)
			obj["max_latency"] = types.Int64Value(int64(apiObject.Transport.MaxLatency))
			obj["min_latency"] = types.Int64Value(int64(apiObject.Transport.MinLatency))
			obj["protocol"] = flex.StringValueToFramework(ctx, apiObject.Transport.Protocol)
			obj["remote_id"] = flex.StringToFramework(ctx, apiObject.Transport.RemoteId)
			obj["sender_control_port"] = types.Int64Value(int64(apiObject.Transport.SenderControlPort))
			obj["sender_ip_address"] = flex.StringToFramework(ctx, apiObject.Transport.SenderIpAddress)
			obj["smoothing_latency"] = types.Int64Value(int64(apiObject.Transport.SmoothingLatency))
			obj["stream_id"] = flex.StringToFramework(ctx, apiObject.Transport.StreamId)
		} else {
			obj["cidr_allow_list"] = types.ListNull(fwtypes.CIDRBlockType)
			obj["max_latency"] = types.Int64Null()
			obj["min_latency"] = types.Int64Null()
			obj["protocol"] = types.StringNull()
			obj["remote_id"] = types.StringNull()
			obj["sender_control_port"] = types.Int64Null()
			obj["sender_ip_address"] = types.StringNull()
			obj["smoothing_latency"] = types.Int64Null()
			obj["stream_id"] = types.StringNull()
		}

		objVal, d := types.ObjectValue(attrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func (r *resourceFlow) mediaStreamOutputConfigurationsAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"encoding_name":              types.StringType,
		"name":                       types.StringType,
		"destination_configurations": types.SetType{ElemType: types.ObjectType{AttrTypes: r.mediaStreamOutputConfigurationDestinationConfigurationsAttrTypes(ctx)}},
		"encoding_parameters":        types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[mediaStreamOutputConfigurationEncodingParametersData](ctx)}},
	}
}

func (r *resourceFlow) flattenMediaStreamOutputConfigurations(ctx context.Context, apiObjects []awstypes.MediaStreamOutputConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := r.mediaStreamOutputConfigurationsAttrTypes(ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"encoding_name": flex.StringValueToFramework(ctx, apiObject.EncodingName),
			"name":          flex.StringToFramework(ctx, apiObject.MediaStreamName),
		}
		destinationConfigurations, d := r.flattenMediaStreamOutputConfigurationDestinationConfigurations(ctx, apiObject.DestinationConfigurations)
		diags.Append(d...)
		obj["destination_configurations"] = destinationConfigurations
		encodingParameters, d := r.flattenMediaStreamOutputConfigurationEncodingParameters(ctx, apiObject.EncodingParameters)
		diags.Append(d...)
		obj["encoding_parameters"] = encodingParameters

		objVal, d := types.ObjectValue(attrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) mediaStreamOutputConfigurationDestinationConfigurationsAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"ip":          types.StringType,
		"outbound_ip": types.StringType,
		"port":        types.Int64Type,
		"interface":   types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[interfaceData](ctx)}},
	}
}

func (r *resourceFlow) flattenMediaStreamOutputConfigurationDestinationConfigurations(ctx context.Context, apiObjects []awstypes.DestinationConfiguration) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := r.mediaStreamOutputConfigurationDestinationConfigurationsAttrTypes(ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if len(apiObjects) == 0 {
		return types.SetNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"ip":          flex.StringToFramework(ctx, apiObject.DestinationIp),
			"port":        types.Int64Value(int64(apiObject.DestinationPort)),
			"outbound_ip": flex.StringToFramework(ctx, apiObject.OutboundIp),
		}
		vpcInterface, d := r.flattenInterface(ctx, apiObject.Interface)
		diags.Append(d...)
		obj["interface"] = vpcInterface

		objVal, d := types.ObjectValue(attrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func (r *resourceFlow) flattenInterface(ctx context.Context, apiObject *awstypes.Interface) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := flex.AttributeTypesMust[interfaceData](ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"name": flex.StringToFramework(ctx, apiObject.Name),
	}
	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

var vpcInterfacesAttrTypes = map[string]attr.Type{
	"name":                   types.StringType,
	"network_interface_ids":  types.ListType{ElemType: types.StringType},
	"network_interface_type": types.StringType,
	"role_arn":               fwtypes.ARNType,
	"security_group_ids":     types.SetType{ElemType: types.StringType},
	"subnet_id":              types.StringType,
}

func (r *resourceFlow) flattenVPCInterfaces(ctx context.Context, apiObjects []awstypes.VpcInterface) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := vpcInterfacesAttrTypes
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if len(apiObjects) == 0 {
		return types.SetNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"name":                   flex.StringToFramework(ctx, apiObject.Name),
			"network_interface_ids":  flex.FlattenFrameworkStringValueList(ctx, apiObject.NetworkInterfaceIds),
			"network_interface_type": flex.StringValueToFramework(ctx, apiObject.NetworkInterfaceType),
			"role_arn":               flex.StringToFrameworkARN(ctx, apiObject.RoleArn, &diags),
			"security_group_ids":     flex.FlattenFrameworkStringValueSet(ctx, apiObject.SecurityGroupIds),
			"subnet_id":              flex.StringToFramework(ctx, apiObject.SubnetId),
		}

		objVal, d := types.ObjectValue(attrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func (r *resourceFlow) flattenMediaStreamOutputConfigurationEncodingParameters(ctx context.Context, apiObject *awstypes.EncodingParameters) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := flex.AttributeTypesMust[mediaStreamOutputConfigurationEncodingParametersData](ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"compression_factor": flex.Float64ToFramework(ctx, &apiObject.CompressionFactor),
		"encoder_profile":    flex.StringValueToFramework(ctx, apiObject.EncoderProfile),
	}
	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenVPCInterfaceAttachment(ctx context.Context, apiObject *awstypes.VpcInterfaceAttachment) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := flex.AttributeTypesMust[vpcInterfaceAttachmentData](ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"name": flex.StringToFramework(ctx, apiObject.VpcInterfaceName),
	}
	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenSources(ctx context.Context, apiObjects []awstypes.Source) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrTypes = map[string]attr.Type{
		"arn":                                  types.StringType,
		"data_transfer_subscriber_fee_percent": types.Int64Type,
		"decryption":                           types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[encryptionData](ctx)}},
		"description":                          types.StringType,
		"entitlement_arn":                      types.StringType,
		"gateway_bridge_source":                types.ListType{ElemType: types.ObjectType{AttrTypes: r.flowGatewayBridgeSourceAttrTypes(ctx)}},
		"ingest_ip":                            types.StringType,
		"ingest_port":                          types.Int64Type,
		"listener_address":                     types.StringType,
		"listener_port":                        types.Int64Type,
		"max_bitrate":                          types.Int64Type,
		"max_latency":                          types.Int64Type,
		"max_sync_buffer":                      types.Int64Type,
		"media_stream_source_configurations":   types.SetType{ElemType: types.ObjectType{AttrTypes: r.mediaStreamSourceConfigurationsAttrTypes(ctx)}},
		"min_latency":                          types.Int64Type,
		"name":                                 types.StringType,
		"protocol":                             types.StringType,
		"sender_control_port":                  types.Int64Type,
		"sender_ip_address":                    types.StringType,
		"stream_id":                            types.StringType,
		"vpc_interface_name":                   types.StringType,
		"whitelist_cidr":                       types.StringType,
	}
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"arn":                                  flex.StringToFramework(ctx, apiObject.SourceArn),
			"data_transfer_subscriber_fee_percent": types.Int64Value(int64(apiObject.DataTransferSubscriberFeePercent)),
			"description":                          flex.StringToFramework(ctx, apiObject.Description),
			"entitlement_arn":                      flex.StringToFramework(ctx, apiObject.EntitlementArn),
			"ingest_ip":                            flex.StringToFramework(ctx, apiObject.IngestIp),
			"ingest_port":                          types.Int64Value(int64(apiObject.IngestPort)),
			"name":                                 flex.StringToFramework(ctx, apiObject.Name),
			"sender_control_port":                  types.Int64Value(int64(apiObject.SenderControlPort)),
			"sender_ip_address":                    flex.StringToFramework(ctx, apiObject.SenderIpAddress),
			"vpc_interface_name":                   flex.StringToFramework(ctx, apiObject.VpcInterfaceName),
			"whitelist_cidr":                       flex.StringToFramework(ctx, apiObject.WhitelistCidr),
		}
		decryption, d := r.flattenEncryption(ctx, apiObject.Decryption)
		diags.Append(d...)
		obj["decryption"] = decryption
		gatewayBridgeSource, d := r.flattenFlowGatewayBridgeSource(ctx, apiObject.GatewayBridgeSource)
		diags.Append(d...)
		obj["gateway_bridge_source"] = gatewayBridgeSource
		mediaStreamSourceConfigurations, d := r.flattenMediaStreamSourceConfigurations(ctx, apiObject.MediaStreamSourceConfigurations)
		diags.Append(d...)
		obj["media_stream_source_configurations"] = mediaStreamSourceConfigurations

		// Set unknowns
		obj["listener_address"] = flex.StringToFramework(ctx, apiObject.Transport.SourceListenerAddress)
		obj["listener_port"] = types.Int64Value(int64(apiObject.Transport.SourceListenerPort))
		obj["max_bitrate"] = types.Int64Value(int64(apiObject.Transport.MaxBitrate))
		obj["max_latency"] = types.Int64Value(int64(apiObject.Transport.MaxLatency))
		obj["max_sync_buffer"] = types.Int64Value(int64(apiObject.Transport.MaxSyncBuffer))
		obj["min_latency"] = types.Int64Value(int64(apiObject.Transport.MinLatency))
		obj["protocol"] = flex.StringValueToFramework(ctx, apiObject.Transport.Protocol)
		obj["stream_id"] = flex.StringToFramework(ctx, apiObject.Transport.StreamId)

		objVal, d := types.ObjectValue(attrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flowGatewayBridgeSourceAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"arn":                      fwtypes.ARNType,
		"vpc_interface_attachment": types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[vpcInterfaceAttachmentData](ctx)}},
	}
}

func (r *resourceFlow) flattenFlowGatewayBridgeSource(ctx context.Context, apiObject *awstypes.GatewayBridgeSource) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := r.flowGatewayBridgeSourceAttrTypes(ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"arn": flex.StringToFrameworkARN(ctx, apiObject.BridgeArn, &diags),
	}
	vpcInterfaceAttachment, d := r.flattenVPCInterfaceAttachment(ctx, apiObject.VpcInterfaceAttachment)
	diags.Append(d...)
	obj["vpc_interface_attachment"] = vpcInterfaceAttachment

	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) mediaStreamSourceConfigurationsAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"input_configurations": types.SetType{ElemType: types.ObjectType{AttrTypes: r.mediaStreamSourceConfigurationInputConfigurationsAttrTypes(ctx)}},
		"encoding_name":        types.StringType,
		"name":                 types.StringType,
	}
}

func (r *resourceFlow) flattenMediaStreamSourceConfigurations(ctx context.Context, apiObjects []awstypes.MediaStreamSourceConfiguration) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := r.mediaStreamSourceConfigurationsAttrTypes(ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if len(apiObjects) == 0 {
		return types.SetNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"encoding_name": flex.StringValueToFramework(ctx, apiObject.EncodingName),
			"name":          flex.StringToFramework(ctx, apiObject.MediaStreamName),
		}
		inputConfigurations, d := r.flattenMediaStreamSourceConfigurationInputConfigurations(ctx, apiObject.InputConfigurations)
		diags.Append(d...)
		obj["input_configurations"] = inputConfigurations

		objVal, d := types.ObjectValue(attrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func (r *resourceFlow) mediaStreamSourceConfigurationInputConfigurationsAttrTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"ip":        types.StringType,
		"port":      types.Int64Type,
		"interface": types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[interfaceData](ctx)}},
	}
}

func (r *resourceFlow) flattenMediaStreamSourceConfigurationInputConfigurations(ctx context.Context, apiObjects []awstypes.InputConfiguration) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := r.mediaStreamSourceConfigurationInputConfigurationsAttrTypes(ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if len(apiObjects) == 0 {
		return types.SetNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		obj := map[string]attr.Value{
			"ip":   flex.StringToFramework(ctx, apiObject.InputIp),
			"port": types.Int64Value(int64(apiObject.InputPort)),
		}
		vpcInterface, d := r.flattenInterface(ctx, apiObject.Interface)
		diags.Append(d...)
		obj["interface"] = vpcInterface

		objVal, d := types.ObjectValue(attrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	setVal, d := types.SetValue(elemType, elems)
	diags.Append(d...)

	return setVal, diags
}

func (r *resourceFlow) flattenSourceFailoverConfig(ctx context.Context, apiObject *awstypes.FailoverConfig) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var attrTypes = map[string]attr.Type{
		"failover_mode":   types.StringType,
		"recovery_window": types.Int64Type,
		"state":           types.StringType,
		"source_priority": types.ListType{ElemType: types.ObjectType{AttrTypes: flex.AttributeTypesMust[sourceFailoverConfigSourcePriorityData](ctx)}},
	}
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"failover_mode":   flex.StringValueToFramework(ctx, apiObject.FailoverMode),
		"recovery_window": types.Int64Value(int64(apiObject.RecoveryWindow)),
		"state":           flex.StringValueToFramework(ctx, apiObject.State),
	}
	source_priority, d := r.flattenSourceFailoverConfigSourcePriority(ctx, apiObject.SourcePriority)
	diags.Append(d...)
	obj["source_priority"] = source_priority
	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) flattenSourceFailoverConfigSourcePriority(ctx context.Context, apiObject *awstypes.SourcePriority) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := flex.AttributeTypesMust[sourceFailoverConfigSourcePriorityData](ctx)
	elemType := types.ObjectType{AttrTypes: attrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"primary_source": flex.StringToFramework(ctx, apiObject.PrimarySource),
	}
	objVal, d := types.ObjectValue(attrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func (r *resourceFlow) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

type resourceFlowData struct {
	FlowArn              types.String   `tfsdk:"arn"`
	ID                   types.String   `tfsdk:"id"`
	Name                 types.String   `tfsdk:"name"`
	Description          types.String   `tfsdk:"description"`
	AvailabilityZone     types.String   `tfsdk:"availability_zone"`
	Entitlements         types.List     `tfsdk:"entitlement"`
	Maintenance          types.List     `tfsdk:"maintenance"`
	MediaStreams         types.Set      `tfsdk:"media_stream"`
	Outputs              types.Set      `tfsdk:"output"`
	Sources              types.List     `tfsdk:"source"`
	SourceFailoverConfig types.List     `tfsdk:"source_failover_config"`
	VpcInterfaces        types.Set      `tfsdk:"vpc_interface"`
	EgressIp             types.String   `tfsdk:"egress_ip"`
	StartFlow            types.Bool     `tfsdk:"start_flow"`
	Status               types.String   `tfsdk:"status"`
	Tags                 types.Map      `tfsdk:"tags"`
	TagsAll              types.Map      `tfsdk:"tags_all"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}

type entitlementData struct {
	EntitlementArn                   types.String `tfsdk:"arn"`
	Name                             types.String `tfsdk:"name"`
	Subscribers                      types.List   `tfsdk:"subscriber"`
	DataTransferSubscriberFeePercent types.Int64  `tfsdk:"data_transfer_subscriber_fee_percent"`
	Description                      types.String `tfsdk:"description"`
	Encryption                       types.List   `tfsdk:"encryption"`
	EntitlementStatus                types.String `tfsdk:"status"`
}

type encryptionData struct {
	RoleArn                      fwtypes.ARN  `tfsdk:"role_arn"`
	Algorithm                    types.String `tfsdk:"algorithm"`
	ConstantInitializationVector types.String `tfsdk:"constant_initialization_vector"`
	DeviceId                     types.String `tfsdk:"device_id"`
	KeyType                      types.String `tfsdk:"key_type"`
	Region                       types.String `tfsdk:"region"`
	ResourceId                   types.String `tfsdk:"resource_id"`
	SecretArn                    fwtypes.ARN  `tfsdk:"secret_arn"`
	Url                          types.String `tfsdk:"url"`
}

type maintenanceData struct {
	MaintenanceDay           types.String `tfsdk:"day"`
	MaintenanceDeadline      types.String `tfsdk:"deadline"`
	MaintenanceScheduledDate types.String `tfsdk:"scheduled_date"`
	MaintenanceStartHour     types.String `tfsdk:"start_hour"`
}

type mediaStreamData struct {
	Fmt             types.Int64  `tfsdk:"fmt"`
	MediaStreamId   types.Int64  `tfsdk:"id"`
	MediaStreamName types.String `tfsdk:"name"`
	MediaStreamType types.String `tfsdk:"type"`
	Attributes      types.List   `tfsdk:"attributes"`
	ClockRate       types.Int64  `tfsdk:"clock_rate"`
	Description     types.String `tfsdk:"desription"`
	VideoFormat     types.String `tfsdk:"video_format"`
}

type mediaStreamAttributeData struct {
	Fmtp types.List   `tfsdk:"fmtp"`
	Lang types.String `tfsdk:"lang"`
}

type mediaStreamAttributeFmtpData struct {
	ChannelOrder   types.String `tfsdk:"channel_order"`
	Colorimetry    types.String `tfsdk:"colorimetry"`
	ExactFramerate types.String `tfsdk:"exact_framerate"`
	Par            types.String `tfsdk:"par"`
	Range          types.String `tfsdk:"range"`
	ScanMode       types.String `tfsdk:"scan_mode"`
	Tcs            types.String `tfsdk:"tcs"`
}

type outputData struct {
	BridgeArn                        fwtypes.ARN  `tfsdk:"bridge_arn"`
	BridgePorts                      types.List   `tfsdk:"bridge_ports"`
	CidrAllowList                    types.List   `tfsdk:"cidr_allow_list"`
	DataTransferSubscriberFeePercent types.Int64  `tfsdk:"data_transfer_subscriber_fee_percent"`
	Description                      types.String `tfsdk:"description"`
	Destination                      types.String `tfsdk:"destination"`
	Encryption                       types.List   `tfsdk:"encryption"`
	EntitlementArn                   types.String `tfsdk:"entitlement_arn"`
	ListenerAddress                  types.String `tfsdk:"listener_address"`
	MaxLatency                       types.Int64  `tfsdk:"max_latency"`
	MediaLiveInputArn                types.String `tfsdk:"media_live_input_arn"`
	MediaStreamOutputConfigurations  types.List   `tfsdk:"media_stream_output_configurations"`
	MinLatency                       types.Int64  `tfsdk:"min_latency"`
	Name                             types.String `tfsdk:"name"`
	OutputArn                        types.String `tfsdk:"arn"`
	Port                             types.Int64  `tfsdk:"port"`
	Protocol                         types.String `tfsdk:"protocol"`
	RemoteId                         types.String `tfsdk:"remote_id"`
	SenderControlPort                types.Int64  `tfsdk:"sender_control_port"`
	SenderIpAddress                  types.String `tfsdk:"sender_ip_address"`
	SmoothingLatency                 types.Int64  `tfsdk:"smoothing_latency"`
	StreamId                         types.String `tfsdk:"stream_id"`
	VpcInterfaceAttachment           types.List   `tfsdk:"vpc_interface_attachment"`
}

type mediaStreamOutputConfigurationData struct {
	DestinationConfigurations types.Set    `tfsdk:"destination_configurations"`
	EncodingName              types.String `tfsdk:"encoding_name"`
	EncodingParameters        types.List   `tfsdk:"encoding_parameters"`
	MediaStreamName           types.String `tfsdk:"name"`
}

type mediaStreamOutputConfigurationDestinationConfigurationData struct {
	DestinationIp   types.String `tfsdk:"ip"`
	DestinationPort types.Int64  `tfsdk:"port"`
	Interface       types.List   `tfsdk:"interface"`
	OutboundIp      types.String `tfsdk:"outbound_ip"`
}

type mediaStreamOutputConfigurationEncodingParametersData struct {
	CompressionFactor types.Float64 `tfsdk:"compression_factor"`
	EncoderProfile    types.String  `tfsdk:"encoder_profile"`
}

type vpcInterfaceAttachmentData struct {
	VpcInterfaceName types.String `tfsdk:"name"`
}

type sourceData struct {
	DataTransferSubscriberFeePercent types.Int64  `tfsdk:"data_transfer_subscriber_fee_percent"`
	Decryption                       types.List   `tfsdk:"decryption"`
	Description                      types.String `tfsdk:"description"`
	EntitlementArn                   types.String `tfsdk:"entitlement_arn"`
	GatewayBridgeSource              types.List   `tfsdk:"gateway_bridge_source"`
	IngestIp                         types.String `tfsdk:"ingest_ip"`
	IngestPort                       types.Int64  `tfsdk:"ingest_port"`
	MaxBitrate                       types.Int64  `tfsdk:"max_bitrate"`
	MaxLatency                       types.Int64  `tfsdk:"max_latency"`
	MaxSyncBuffer                    types.Int64  `tfsdk:"max_sync_buffer"`
	MediaStreamSourceConfigurations  types.Set    `tfsdk:"media_stream_source_configurations"`
	MinLatency                       types.Int64  `tfsdk:"min_latency"`
	Name                             types.String `tfsdk:"name"`
	Protocol                         types.String `tfsdk:"protocol"`
	SenderControlPort                types.Int64  `tfsdk:"sender_control_port"`
	SenderIpAddress                  types.String `tfsdk:"sender_ip_address"`
	SourceArn                        types.String `tfsdk:"arn"`
	SourceListenerAddress            types.String `tfsdk:"listener_address"`
	SourceListenerPort               types.Int64  `tfsdk:"listener_port"`
	StreamId                         types.String `tfsdk:"stream_id"`
	VpcInterfaceName                 types.String `tfsdk:"vpc_interface_name"`
	WhitelistCidr                    types.String `tfsdk:"whitelist_cidr"`
}

type mediaStreamSourceConfigurationData struct {
	EncodingName        types.String `tfsdk:"encoding_name"`
	MediaStreamName     types.String `tfsdk:"name"`
	InputConfigurations types.Set    `tfsdk:"input_configurations"`
}

type mediaStreamSourceConfigurationInputConfigurationData struct {
	InputIp   types.String `tfsdk:"ip"`
	InputPort types.Int64  `tfsdk:"port"`
	Interface types.List   `tfsdk:"interface"`
}

type flowGatewayBridgeSourceData struct {
	BridgeArn              fwtypes.ARN `tfsdk:"arn"`
	VpcInterfaceAttachment types.List  `tfsdk:"vpc_interface_attachment"`
}

type sourceFailoverConfigData struct {
	FailoverMode   types.String `tfsdk:"failover_mode"`
	RecoveryWindow types.Int64  `tfsdk:"recovery_window"`
	SourcePriority types.List   `tfsdk:"source_priority"`
	State          types.String `tfsdk:"state"`
}

type sourceFailoverConfigSourcePriorityData struct {
	PrimarySource types.String `tfsdk:"primary_source"`
}

type interfaceData struct {
	Name types.String `tfsdk:"name"`
}

type vpcInterfaceData struct {
	Name                 types.String `tfsdk:"name"`
	RoleArn              fwtypes.ARN  `tfsdk:"role_arn"`
	SecurityGroupIds     types.Set    `tfsdk:"security_group_ids"`
	SubnetId             types.String `tfsdk:"subnet_id"`
	NetworkInterfaceType types.String `tfsdk:"network_interface_type"`
	NetworkInterfaceIds  types.List   `tfsdk:"network_interface_ids"`
}
