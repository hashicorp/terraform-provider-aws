// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconnect

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediaconnect/types"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Flow")
// @Tags(identifierAttribute="arn")
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
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_zone": schema.StringAttribute{
				Computed: true,
			},
			"egress_ip": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"start_flow": schema.BoolAttribute{
				Optional: true,
				Default:  false,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"entitlement": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"arn": framework.ARNAttributeComputedOnly(),
						"subscriber": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"data_transfer_subscriber_fee_percent": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.Between(0, 100),
							},
						},
						"description": schema.StringAttribute{
							Required: true,
						},
						"status": schema.StringAttribute{
							Optional: true,
							Default:  awstypes.EntitlementStatusEnabled,
							Validators: []validator.String{
								stringvalidator.OneOf(mediaconnect.EntitlementStatus_Values()...),
							},
						},
						"name": schema.StringAttribute{
							Required: true,
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
								stringvalidator.OneOf(mediaconnect.MaintenanceDay_Values()...),
							},
						},
						"deadline": schema.StringAttribute{
							Computed: true,
						},
						"scheduled_date": schema.StringAttribute{
							Optional: true,
							Computed: true,
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
						"id": schema.Int64Attribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(mediaconnect.MediaStreamType_Values()...),
							},
						},
						"clock_rate": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.OneOf(48000, 90000, 96000),
							},
						},
						"description": schema.StringAttribute{
							Optional: true,
						},
						"video_format": schema.StringAttribute{
							Optional: true,
						},
						"fmt": schema.Int64Attribute{
							Computed: true,
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
														stringvalidator.OneOf(mediaconnect.Colorimetry_Values()...),
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
														stringvalidator.OneOf(mediaconnect.Range_Values()...),
													},
												},
												"scan_mode": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.OneOf(mediaconnect.ScanMode_Values()...),
													},
												},
												"tcs": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.OneOf(mediaconnect.Tcs_Values()...),
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
						"name": schema.StringAttribute{
							Optional: true,
						},
						"arn": framework.ARNAttributeComputedOnly(),
						"bridge_ports": schema.SetAttribute{
							ElementType: types.Int64Type,
							Computed:    true,
						},
						"data_transfer_subscriber_fee_percent": schema.Int64Attribute{
							Computed: true,
							Validators: []validator.Int64{
								int64validator.Between(0, 100),
							},
						},
						"entitlement_arn": framework.ARNAttributeComputedOnly(),
						"protocol": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(mediaconnect.Protocol_Values()...),
							},
						},
						"bridge_arn": framework.ARNAttributeComputedOnly(),
						"cidr_allow_list": schema.SetAttribute{
							ElementType: fwtypes.CIDRBlockType,
							Optional:    true,
						},
						"description": schema.StringAttribute{
							Optional: true,
						},
						"destination": schema.StringAttribute{
							Optional: true,
						},
						"listener_address": schema.StringAttribute{
							Computed: true,
						},
						"max_latency": schema.Int64Attribute{
							Optional: true,
						},
						"media_live_input_arn": framework.ARNAttributeComputedOnly(),
						"min_latency": schema.Int64Attribute{
							Optional: true,
						},
						"port": schema.Int64Attribute{
							Optional: true,
						},
						"remote_id": schema.StringAttribute{
							Optional: true,
						},
						"sender_control_port": schema.Int64Attribute{
							Optional: true,
						},
						"smoothing_latency": schema.Int64Attribute{
							Optional: true,
						},
						"stream_id": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"encryption": encryptionBlock(),
						"media_stream_output_configurations": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"encoding_name": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.OneOf(mediaconnect.EncodingName_Values()...),
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
													Validators: []validator.Validator{
														fwvalidators.IPv4Address(),
													},
												},
												"port": schema.Int64Attribute{
													Required: true,
													Validators: []validator.Int64{
														int64validator.Between(1, 65535),
													},
												},
												"outbound_ip": schema.StringAttribute{
													Computed: true,
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
														stringvalidator.OneOf(mediaconnect.EncoderProfile_Values()...),
													},
												},
											},
										},
									},
								},
							},
						},
						"transport": transportBlock(),
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
								stringvalidator.OneOf(mediaconnect.FailoverMode_Values()...),
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
								stringvalidator.OneOf(mediaconnect.State_Values()...),
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
							Validators: []validator.Int64{
								int64validator.Between(0, 100),
							},
						},
						"description": schema.StringAttribute{
							Required: true,
						},
						"entitlement_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Computed:   true,
						},
						"ingest_ip": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"ingest_port": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"max_bitrate": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"max_latency": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"max_sync_buffer": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"min_latency": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"protocol": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								stringvalidator.OneOf(mediaconnect.Protocol_Values()...),
							},
						},
						"sender_control_port": schema.Int64Attribute{
							Optional: true,
						},
						"sender_ip_address": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Validator{
								fwvalidators.IPv4Address(),
							},
						},
						"source_listener_address": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"source_listener_port": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"stream_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"vpc_interface_name": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"whitelist_cidr": schema.StringAttribute{
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsCIDRNetwork(0, 128),
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
											stringvalidator.OneOf(mediaconnect.EncodingName_Values()...),
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
													Validators: []validator.Validator{
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
						"transport": transportBlock(),
					},
				},
			},
			"vpc_interface": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
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
						"network_interface_ids": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
						"network_interface_type": schema.StringAttribute{
							Optional: true,
							Default:  types.NetworkInterfaceTypeEna,
							Validators: []validator.String{
								stringvalidator.OneOf(mediaconnect.NetworkInterfaceType_Values()...),
							},
						},
					},
				},
			},
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
						stringvalidator.OneOf(mediaconnect.Algorithm_Values()...),
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
						stringvalidator.OneOf(mediaconnect.KeyType_Values()...),
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

func transportBlock() datasourceschema.Block {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"protocol": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf(mediaconnect.Protocol_Values()...),
					},
				},
				"cidr_allow_list": schema.SetAttribute{
					ElementType: fwtypes.CIDRBlockType,
					Optional:    true,
					Computed:    true,
				},
				"max_bitrate": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"max_latency": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"max_sync_buffer": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"min_latency": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"remote_id": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"sender_control_port": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"sender_ip_address": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Validator{
						fwvalidators.IPv4Address(),
					},
				},
				"smoothing_latency": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"source_listener_address": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"source_listener_port": schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
				"stream_id": schema.StringAttribute{
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
		in.Entitlements = flex.ExpandFrameworkListNestedBlock(ctx, plan.Entitlements, r.expandEntitlement)
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
		sources := flex.ExpandFrameworkListNestedBlock(ctx, plan.Sources, r.expandSourceCreate)
		in.Source = &sources[0]
		in.Sources = sources[1:]
	}
	if !plan.SourceFailoverConfig() {
		in.SourceFailoverConfig = flex.ExpandFrameworkListNestedBlockPtr(ctx, plan.SourceFailoverConfig, r.expandSourceFailoverConfigCreate)
	}
	if !plan.VpcInterfaces() {
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

	plan.ARN = flex.StringToFramework(ctx, out.Flow.FlowArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitFlowCreated(ctx, conn, plan.ARN.String(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForCreation, ResNameFlow, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceFlow) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().MediaConnectClient(ctx)

	var state resourceFlowData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFlowByARN(ctx, conn, state.FlowArn.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionSetting, ResNameFlow, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.FlowArn = flex.StringToFramework(ctx, out.FlowArn)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.AvailabilityZone = flex.StringToFramework(ctx, out.AvailabilityZone)
	state.Entitlements = flex.FlattenFrameworkListNestedBlock[entitlementData](ctx, out.Entitlements, r.flattenEntitlement)
	state.Maintenance = flex.FlattenFrameworkListNestedBlock[maintenanceData](ctx, out.Maintenance, r.flattenMaintenance)
	state.MediaStreams = flex.FlattenFrameworkListNestedBlock[mediaStreamData](ctx, out.Entitlements, r.flattenMediaStream)
	state.Outputs = flex.FlattenFrameworkListNestedBlock[outputData](ctx, out.Outputs, r.flattenOutput)
	sources := []awstypes.Source{*out.Source}
	sources = append(sources, out.Sources...)
	state.Sources = flex.FlattenFrameworkListNestedBlock[sourceData](ctx, out.Sources, r.flattenSource)
	state.SourceFailoverConfig = flex.FlattenFrameworkListNestedBlock[sourceFailoverConfigData](ctx, out.SourceFailoverConfig, r.flattenSourceFailoverConfig)
	state.VpcInterfaces = flex.FlattenFrameworkListNestedBlock[vpcInterfaceData](ctx, out.VpcInterfaces, r.flattenVPCInterface)
	state.EgressIp = flex.StringToFramework(ctx, out.EgressIp)
	state.Status = flex.StringToFramework(ctx, out.Status)

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

	if !plan.Maintenance.Equal(state.Maintenance) ||
		!plan.SourceFailoverConfig.Equal(state.SourceFailoverConfig) {

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
				create.ProblemStandardMessage(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, plan.ARN.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Flow == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, plan.ARN.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.Flow.FlowArn)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitFlowUpdated(ctx, conn, plan.ARN.String(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaConnectClient(ctx)

	if d.HasChangesExcept("tags", "tags_all", "start_flow") {
		in := &mediaconnect.UpdateFlowInput{
			FlowArn: aws.String(d.Id()),
		}

		if d.HasChange("maintenance") {
			in.Maintenance = expandFlowMaintenanceCreate(v.([]interface{}))
			in.Maintenance = expandFlowMaintenanceUpdate(d.Get("maintenance").([]interface{}))
		}

		if d.HasChange("source_failover_config") {
			in.SourceFailoverConfig = expandFlowSourceFailoverConfigUpdate(d.Get("source_failover_config").([]interface{}))
		}

		flow, err := findFlowByARN(ctx, conn, d.Id())

		if err != nil {
			return create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, d.Id(), err)
		}

		if flow.Status == types.StatusActive {
			if err := stopFlow(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
				return create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, d.Id(), err)
			}
		}

		out, err := conn.UpdateFlow(ctx, in)
		if err != nil {
			return create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, d.Id(), err)
		}

		if _, err := waitFlowUpdated(ctx, conn, aws.ToString(out.Flow.FlowArn), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, d.Id(), err)
		}
	}

	if d.Get("start_flow").(bool) {
		if err := startFlow(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
			return create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, d.Get("name").(string), err)
		}
	}

	if d.HasChange("start_flow") {
		flow, err := findFlowByARN(ctx, conn, d.Id())

		if err != nil {
			return create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, d.Id(), err)
		}

		switch d.Get("start_flow").(bool) {
		case true:
			if flow.Status == types.StatusStandby {
				if err := startFlow(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
					return create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, d.Id(), err)
				}
			}
		default:
			if flow.Status == types.StatusActive {
				if err := stopFlow(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
					return create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, d.Id(), err)
				}
			}
		}
	}

	return resourceFlowRead(ctx, d, meta)
}

func resourceFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaConnectClient(ctx)

	log.Printf("[INFO] Deleting MediaConnect Flow %s", d.Id())

	flow, err := findFlowByARN(ctx, conn, d.Id())

	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionDeleting, ResNameFlow, d.Id(), err)
	}

	if flow.Status == types.StatusActive {
		if err := stopFlow(ctx, conn, d.Timeout(schema.TimeoutDelete), d.Id()); err != nil {
			return create.DiagError(names.MediaConnect, create.ErrActionDeleting, ResNameFlow, d.Id(), err)
		}
	}

	_, err = conn.DeleteFlow(ctx, &mediaconnect.DeleteFlowInput{
		FlowArn: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.MediaConnect, create.ErrActionDeleting, ResNameFlow, d.Id(), err)
	}

	if _, err := waitFlowDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionWaitingForDeletion, ResNameFlow, d.Id(), err)
	}

	return nil
}

func startFlow(ctx context.Context, conn *mediaconnect.Client, timeout time.Duration, id string) error {
	_, err := conn.StartFlow(ctx, &mediaconnect.StartFlowInput{
		FlowArn: aws.String(id),
	})

	if err != nil {
		return fmt.Errorf("starting MediaConnect Flow (%s): %s", id, err)
	}

	_, err = waitFlowStarted(ctx, conn, id, timeout)

	if err != nil {
		return fmt.Errorf("waiting for MediaConnect Flow (%s) start: %s", id, err)
	}

	return nil
}

func stopFlow(ctx context.Context, conn *mediaconnect.Client, timeout time.Duration, id string) error {
	_, err := conn.StopFlow(ctx, &mediaconnect.StopFlowInput{
		FlowArn: aws.String(id),
	})

	if err != nil {
		return fmt.Errorf("stopping MediaConnect Flow (%s): %s", id, err)
	}

	_, err = waitFlowStopped(ctx, conn, id, timeout)

	if err != nil {
		return fmt.Errorf("waiting for MediaConnect Flow (%s) stop: %s", id, err)
	}

	return nil
}

func waitFlowCreated(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(types.StatusStandby, types.StatusActive),
		Refresh:                   statusFlow(ctx, conn, id),
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

func waitFlowUpdated(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.StatusUpdating),
		Target:                    enum.Slice(types.StatusStandby, types.StatusActive),
		Refresh:                   statusFlow(ctx, conn, id),
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

func waitFlowDeleted(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.StatusDeleting),
		Target:  []string{},
		Refresh: statusFlow(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.DescribeFlowOutput); ok {
		return out, err
	}

	return nil, err
}

func waitFlowStarted(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.DescribeFlowOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.StatusStarting),
		Target:  enum.Slice(types.StatusActive),
		Refresh: statusFlow(ctx, conn, id),
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
		Pending: enum.Slice(types.StatusStopping),
		Target:  enum.Slice(types.StatusStandby),
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

func findFlowByARN(ctx context.Context, conn *mediaconnect.Client, arn string) (*mediaconnect.Flow, error) {
	in := &mediaconnect.DescribeFlowInput{
		FlowArn: aws.String(arn),
	}
	out, err := conn.DescribeFlow(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
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

func (r *resourceFlow) expandEntitlement(ctx context.Context, data entitlementData) awstypes.GrantEntitlementRequest {
	var out awstypes.GrantEntitlementRequest

	if !data.Name.IsNull() {
		out.Name = flex.StringFromFramework(ctx, data.Name)
	}
	if !data.Subscribers.IsNull() {
		out.Subscribers = flex.ExpandFrameworkStringValueList(ctx, data.Subscribers)
	}
	if !data.DataTransferSubscriberFeePercent.IsNull() {
		out.DataTransferSubscriberFeePercent = flex.Int64FromFramework(ctx, data.DataTransferSubscriberFeePercent)
	}
	if !data.Description.IsNull() {
		out.Description = flex.StringFromFramework(ctx, data.Description)
	}
	if !data.Encryption.IsNull() {
		out.Encryption = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Encryption, r.expandEncryption)
	}
	if !data.EntitlementStatus.IsNull() {
		out.EntitlementStatus = awstypes.EntitlementStatus(data.EntitlementStatus.ValueString())
	}

	return out
}

func (r *resourceFlow) expandEncryption(ctx context.Context, data encryptionData) *awstypes.Encryption {
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

func (r *resourceFlow) expandMaintenanceCreate(ctx context.Context, data maintenanceData) *awstypes.AddMaintenance {
	return &awstypes.AddMaintenance{
		MaintenanceDay:       awstypes.MaintenanceDay(m.MaintenanceDay.ValueString()),
		MaintenanceStartHour: flex.StringFromFramework(ctx, m.MaintenanceStartHour),
	}
}

func (r *resourceFlow) expandMaintenanceUpdate(ctx context.Context, data maintenanceData) *awstypes.UpdateMaintenance {
	return &awstypes.UpdateMaintenance{
		MaintenanceDay:           awstypes.MaintenanceDay(m.MaintenanceDay.ValueString()),
		MaintenanceStartHour:     flex.StringFromFramework(ctx, m.MaintenanceStartHour),
		MaintenanceScheduledDate: flex.StringFromFramework(ctx, m.MaintenanceScheduledDate),
	}
}

func (r *resourceFlow) expandMediaStreamCreate(ctx context.Context, data mediaStreamData) awstypes.AddMediaStreamRequest {
	return awstypes.AddMediaStreamRequest{
		MediaStreamId:   flex.Int64FromFramework(ctx, data.MediaStreamId),
		MediaStreamName: flex.StringFromFramework(ctx, data.MediaStreamName),
		MediaStreamType: awstypes.MediaStreamType(data.MediaStreamType.ValueString()),
		Attributes:      flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Attributes, r.expandMediaStreamAttributeCreate),
		ClockRate:       flex.Int64FromFramework(ctx, data.ClockRate),
		Description:     flex.StringFromFramework(ctx, data.Description),
		VideoFormat:     flex.StringFromFramework(ctx, data.VideoFormat),
	}
}

func (r *resourceFlow) expandMediaStreamAttributeCreate(ctx context.Context, data mediaStreamAttributeData) *awstypes.MediaStreamAttributesRequest {
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
		out.Encryption = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Encryption, r.expandEncryption)
	}
	if !data.MaxLatency.IsNull() {
		out.MaxLatency = flex.Int64FromFramework(ctx, data.MaxLatency)
	}
	if !data.MediaStreamOutputConfigurations.IsNull() {
		out.MediaStreamOutputConfigurations = flex.ExpandFrameworkSetNestedBlock(ctx, data.MediaStreamOutputConfigurations, r.expandMediaStreamOutputConfigurationCreate)
	}
	if !data.MinLatency.IsNull() {
		out.MinLatency = flex.Int64FromFramework(ctx, data.MinLatency)
	}
	if !data.Port.IsNull() {
		out.Port = flex.Int64FromFramework(ctx, data.Port)
	}
	if !data.SenderControlPort.IsNull() {
		out.SenderControlPort = flex.Int64FromFramework(ctx, data.SenderControlPort)
	}
	if !data.SmoothingLatency.IsNull() {
		out.SmoothingLatency = flex.Int64FromFramework(ctx, data.SmoothingLatency)
	}
	if !data.StreamId.IsNull() {
		out.StreamId = flex.StringFromFramework(ctx, data.StreamId)
	}
	if !data.VpcInterfaceAttachment.IsNull() {
		out.VpcInterfaceAttachment = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.VpcInterfaceAttachment, r.expandVPCInterfaceAttachment)
	}

	return out
}

func (r *resourceFlow) expandMediaStreamOutputConfigurationCreate(ctx context.Context, data mediaStreamOutputConfigurationData) awstypes.MediaStreamOutputConfigurationRequest {
	out := awstypes.MediaStreamOutputConfigurationRequest{
		EncodingName:    flex.StringFromFramework(ctx, data.EncodingName),
		MediaStreamName: flex.StringFromFramework(ctx, data.MediaStreamName),
	}

	if !data.DestinationConfigurations.IsNull() {
		out.DestinationConfigurations = flex.ExpandFrameworkSetNestedBlock(ctx, data.DestinationConfigurations, r.expandMediaStreamOutputConfigurationDestinationConfigurationCreate)
	}
	if !data.EncodingParameters.IsNull() {
		out.EncodingParameters = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.EncodingParameters, r.expandMediaStreamOutputConfigurationEncodingParametersCreate)
	}

	return &out
}

func (r *resourceFlow) expandMediaStreamOutputConfigurationDestinationConfigurationCreate(ctx context.Context, data mediaStreamOutputConfigurationDestinationConfigurationData) awstypes.DestinationConfigurationRequest {
	return awstypes.DestinationConfigurationRequest{
		DestinationIp:   flex.StringFromFramework(ctx, data.DestinationIp),
		DestinationPort: flex.Int64FromFramework(ctx, data.DestinationPort),
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
		CompressionFactor: flex.Float64FromFramework(ctx, data.CompressionFactor),
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
		out.Decryption = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Decryption, r.expandEncryption)
	}
	if !data.Description.IsNull() {
		out.Description = flex.StringFromFramework(ctx, data.Description)
	}
	if !data.EntitlementArn.IsNull() {
		out.EntitlementArn = flex.ARNStringFromFramework(ctx, data.EntitlementArn)
	}
	if !data.GatewayBridgeSource.IsNull() {
		out.GatewayBridgeSource = flex.ExpandFrameworkListNestedBlockPtr(ctx, data.GatewayBridgeSource, r.expandFlowGatewayBridgeSourceCreate)
	}
	if !data.IngestPort.IsNull() {
		out.IngestPort = flex.Int64FromFramework(ctx, data.IngestPort)
	}
	if !data.MaxBitrate.IsNull() {
		out.MaxBitrate = flex.Int64FromFramework(ctx, data.MaxBitrate)
	}
	if !data.MaxLatency.IsNull() {
		out.MaxLatency = flex.Int64FromFramework(ctx, data.MaxLatency)
	}
	if !data.MaxSyncBuffer.IsNull() {
		out.MaxSyncBuffer = flex.Int64FromFramework(ctx, data.MaxSyncBuffer)
	}
	if !data.MediaStreamSourceConfigurations.IsNull() {
		out.MediaStreamSourceConfigurations = flex.ExpandFrameworkSetNestedBlock(ctx, data.MediaStreamSourceConfigurations, r.expandMediaStreamSourceConfigurationCreate)
	}
	if !data.MinLatency.IsNull() {
		out.MinLatency = flex.Int64FromFramework(ctx, data.MinLatency)
	}
	if !data.Name.IsNull() {
		out.Name = flex.StringFromFramework(ctx, data.Name)
	}
	if !data.Protocol.IsNull() {
		out.Protocol = awstypes.Protocol(data.Protocol.ValueString())
	}
	if !data.SenderControlPort.IsNull() {
		out.SenderControlPort = flex.Int64FromFramework(ctx, data.SenderControlPort)
	}
	if !data.SenderIpAddress.IsNull() {
		out.SenderIpAddress = flex.StringFromFramework(ctx, data.SenderIpAddress)
	}
	if !data.SourceListenerAddress.IsNull() {
		out.SourceListenerAddress = flex.StringFromFramework(ctx, data.SourceListenerAddress)
	}
	if !data.SourceListenerPort.IsNull() {
		out.SourceListenerPort = flex.Int64FromFramework(ctx, data.SourceListenerPort)
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

func (r *resourceFlow) expandMediaStreamSourceConfigurationCreate(ctx context.Context, data mediaStreamSourceConfigurationData) awstypes.MediaStreamSourceConfigurationRequest {
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
		InputPort: flex.Int64FromFramework(ctx, data.InputPort),
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
		out.RecoveryWindow = flex.Int64FromFramework(ctx, data.RecoveryWindow)
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
		out.RecoveryWindow = flex.Int64FromFramework(ctx, data.RecoveryWindow)
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

func (r *resourceFlow) flattenEntitlement(ctx context.Context, apiObject awstypes.Entitlement) types.List {
	attributeTypes := flex.AttributeTypesMust[entitlementData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"arn":                                  flex.StringToFramework(ctx, apiObject.EntitlementArn),
			"name":                                 flex.StringToFramework(ctx, apiObject.Name),
			"subscriber":                           flex.FlattenFrameworkStringValueList(ctx, apiObject.Subscribers),
			"data_transfer_subscriber_fee_percent": flex.Int64ToFramework(ctx, &apiObject.DataTransferSubscriberFeePercent),
			"description":                          flex.StringToFramework(ctx, apiObject.Description),
			"encryption":                           flex.FlattenFrameworkListNestedBlock[encryptionData](ctx, apiObject.Encryption, r.flattenEncryption),
			"status":                               flex.StringValueToFramework(ctx, apiObject.EntitlementStatus),
		}),
	})
}

func (r *resourceFlow) flattenEncryption(ctx context.Context, apiObject *awstypes.Encryption) types.List {
	attributeTypes := flex.AttributeTypesMust[encryptionData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"role_arn":                       flex.StringToFramework(ctx, apiObject.RoleArn),
			"algorithm":                      flex.StringValueToFramework(ctx, apiObject.Algorithm),
			"constant_initialization_vector": flex.StringToFramework(ctx, apiObject.ConstantInitializationVector),
			"device_id":                      flex.StringToFramework(ctx, apiObject.DeviceId),
			"key_type":                       flex.StringValueToFramework(ctx, apiObject.KeyType),
			"region":                         flex.StringToFramework(ctx, apiObject.Region),
			"resource_id":                    flex.StringToFramework(ctx, apiObject.ResourceId),
			"secret_arn":                     flex.StringToFramework(ctx, apiObject.SecretArn),
			"url":                            flex.StringToFramework(ctx, apiObject.Url),
		}),
	})
}

func (r *resourceFlow) flattenMaintenance(ctx context.Context, apiObject *awstypes.Maintenance) types.List {
	attributeTypes := flex.AttributeTypesMust[maintenanceData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"day":            flex.StringValueToFramework(ctx, apiObject.MaintenanceDay),
			"deadline":       flex.StringToFramework(ctx, apiObject.MaintenanceDeadline),
			"scheduled_date": flex.StringToFramework(ctx, apiObject.MaintenanceScheduledDate),
			"start_hour":     flex.StringToFramework(ctx, apiObject.MaintenanceStartHour),
		}),
	})
}

func (r *resourceFlow) flattenMediaStream(ctx context.Context, apiObject awstypes.MediaStream) types.List {
	attributeTypes := flex.AttributeTypesMust[mediaStreamData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"fmt":          flex.Int64ToFramework(ctx, &apiObject.Fmt),
			"id":           flex.Int64ToFramework(ctx, &apiObject.MediaStreamId),
			"name":         flex.StringToFramework(ctx, apiObject.MediaStreamName),
			"type":         flex.StringValueToFramework(ctx, apiObject.MediaStreamType),
			"attributes":   flex.FlattenFrameworkListNestedBlock[encryptionData](ctx, apiObject.Attributes, r.flattenMediaStreamAttribute),
			"clock_rate":   flex.Int64ToFramework(ctx, &apiObject.ClockRate),
			"description":  flex.StringToFramework(ctx, apiObject.Description),
			"video_format": flex.StringToFramework(ctx, apiObject.VideoFormat),
		}),
	})
}

func (r *resourceFlow) flattenMediaStreamAttribute(ctx context.Context, apiObject *awstypes.MediaStreamAttributes) types.List {
	attributeTypes := flex.AttributeTypesMust[mediaStreamAttributeData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"fmtp": flex.FlattenFrameworkListNestedBlock[mediaStreamAttributeFmtpData](ctx, apiObject.Fmtp, r.flattenMediaStreamAttributeFmtp),
			"lang": flex.StringToFramework(ctx, apiObject.Lang),
		}),
	})
}

func (r *resourceFlow) flattenMediaStreamAttributeFmtp(ctx context.Context, apiObject *awstypes.Fmtp) types.List {
	attributeTypes := flex.AttributeTypesMust[mediaStreamAttributeFmtpData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"channel_order":   flex.StringToFramework(ctx, apiObject.ChannelOrder),
			"colorimetry":     flex.StringValueToFramework(ctx, apiObject.Colorimetry),
			"exact_framerate": flex.StringToFramework(ctx, apiObject.ExactFramerate),
			"par":             flex.StringToFramework(ctx, apiObject.Par),
			"range":           flex.StringValueToFramework(ctx, apiObject.Range),
			"scan_mode":       flex.StringValueToFramework(ctx, apiObject.ScanMode),
			"tcs":             flex.StringValueToFramework(ctx, apiObject.Tcs),
		}),
	})
}

func (r *resourceFlow) flattenOutput(ctx context.Context, apiObject awstypes.Output) types.List {
	attributeTypes := flex.AttributeTypesMust[outputData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"name":                                 flex.StringToFramework(ctx, apiObject.Name),
			"arn":                                  flex.StringToFramework(ctx, apiObject.OutputArn),
			"bridge_arn":                           flex.StringToFramework(ctx, apiObject.BridgeArn),
			"data_transfer_subscriber_fee_percent": flex.Int64ToFramework(ctx, apiObject.DataTransferSubscriberFeePercent),
			"description":                          flex.StringToFramework(ctx, apiObject.Description),
			"destination":                          flex.StringToFramework(ctx, apiObject.Destination),
			"encryption":                           flex.FlattenFrameworkListNestedBlock[encryptionData](ctx, apiObject.Encryption, r.flattenEncryption),
			"entitlement_arn":                      flex.StringToFramework(ctx, apiObject.EntitlementArn),
			"listener_address":                     flex.StringToFramework(ctx, apiObject.ListenerAddress),
			"media_lvie_input_arn":                 flex.StringToFramework(ctx, apiObject.MediaLiveInputArn),
			"media_stream_output_configurations":   flex.FlattenFrameworkListNestedBlock[mediaStreamOutputConfigurationData](ctx, apiObject.MediaStreamOutputConfigurations, r.flattenMediaStreamOutputConfiguration),
			"port":                                 flex.Int64ToFramework(ctx, apiObject.Port),
			"transport":                            flex.FlattenFrameworkListNestedBlock[transportData](ctx, apiObject.Transport, r.flattenTransport),
			"vpc_interface_attachment":             flex.FlattenFrameworkListNestedBlock[vpcInterfaceAttachmentData](ctx, apiObject.VpcInterfaceAttachment, r.flattenVPCInterfaceAttachment),
			"channel_order":                        flex.StringToFramework(ctx, apiObject.ChannelOrder),
			"colorimetry":                          flex.StringValueToFramework(ctx, apiObject.Colorimetry),
			"exact_framerate":                      flex.StringToFramework(ctx, apiObject.ExactFramerate),
			"par":                                  flex.StringToFramework(ctx, apiObject.Par),
			"range":                                flex.StringValueToFramework(ctx, apiObject.Range),
			"scan_mode":                            flex.StringValueToFramework(ctx, apiObject.ScanMode),
			"tcs":                                  flex.StringValueToFramework(ctx, apiObject.Tcs),
		}),
	})
}

func (r *resourceFlow) flattenMediaStreamOutputConfiguration(ctx context.Context, apiObject awstypes.MediaStreamOutputConfiguration) types.List {
	attributeTypes := flex.AttributeTypesMust[mediaStreamOutputConfigurationData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"encoding_name":             flex.StringValueToFramework(ctx, apiObject.EncodingName),
			"name":                      flex.StringToFramework(ctx, apiObject.MediaStreamName),
			"destination_configuration": flex.FlattenFrameworkListNestedBlock[mediaStreamOutputConfigurationDestinationConfigurationData](ctx, apiObject.DestinationConfigurations, r.flattenMediaStreamOutputConfigurationDestinationConfiguration),
			"encoding_parameters":       flex.FlattenFrameworkListNestedBlock[mediaStreamOutputConfigurationEncodingParametersData](ctx, apiObject.EncodingParameters, r.flattenMediaStreamOutputConfigurationEncodingParameters),
		}),
	})
}

func (r *resourceFlow) flattenMediaStreamOutputConfigurationDestinationConfiguration(ctx context.Context, apiObject awstypes.DestinationConfiguration) types.List {
	attributeTypes := flex.AttributeTypesMust[mediaStreamOutputConfigurationDestinationConfigurationData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"ip":          flex.StringToFramework(ctx, apiObject.DestinationIp),
			"port":        flex.Int64ToFramework(ctx, &apiObject.DestinationPort),
			"interface":   flex.FlattenFrameworkListNestedBlock[vpcInterfaceData](ctx, apiObject.Interface, r.flattenVPCInterface),
			"outbound_ip": flex.StringToFramework(ctx, apiObject.OutboundIp),
		}),
	})
}

func (r *resourceFlow) flattenVPCInterface(ctx context.Context, apiObject *awstypes.Interface) types.List {
	attributeTypes := flex.AttributeTypesMust[vpcInterfaceData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"name": flex.StringToFramework(ctx, apiObject.DestinationIp),
		}),
	})
}

func (r *resourceFlow) flattenMediaStreamOutputConfigurationEncodingParameters(ctx context.Context, apiObject *awstypes.EncodingParameters) types.List {
	attributeTypes := flex.AttributeTypesMust[mediaStreamOutputConfigurationEncodingParametersData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"compression_factor": flex.Float64ToFramework(ctx, &apiObject.CompressionFactor),
			"encoder_profile":    flex.StringValueToFramework(ctx, apiObject.EncoderProfile),
		}),
	})
}

func (r *resourceFlow) flattenTransport(ctx context.Context, apiObject *awstypes.Transport) types.List {
	attributeTypes := flex.AttributeTypesMust[transportData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"protocol":                flex.StringValueToFramework(ctx, apiObject.Protocol),
			"cidr_allow_list":         flex.ExpandFrameworkStringValueList(ctx, data.CidrAllowList),
			"max_bitrate":             flex.Int64ToFramework(ctx, &apiObject.MaxBitrate),
			"max_latency":             flex.Int64ToFramework(ctx, &apiObject.MaxLatency),
			"max_sync_buffer":         flex.Int64ToFramework(ctx, &apiObject.MaxSyncBuffer),
			"min_latency":             flex.Int64ToFramework(ctx, &apiObject.MinLatency),
			"remote_id":               flex.StringToFramework(ctx, apiObject.RemoteId),
			"sender_control_port":     flex.Int64ToFramework(ctx, &apiObject.SenderControlPort),
			"sender_ip_address":       flex.StringToFramework(ctx, apiObject.SenderIpAddress),
			"smoothing_latency":       flex.Int64ToFramework(ctx, &apiObject.SmoothingLatency),
			"source_listener_address": flex.StringToFramework(ctx, apiObject.SourceListenerAddress),
			"source_listener_port":    flex.Int64ToFramework(ctx, &apiObject.SourceListenerPort),
			"stream_id":               flex.StringToFramework(ctx, apiObject.StreamId),
		}),
	})
}

func (r *resourceFlow) flattenVPCInterfaceAttachment(ctx context.Context, apiObject *awstypes.VpcInterfaceAttachment) types.List {
	attributeTypes := flex.AttributeTypesMust[vpcInterfaceAttachmentData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"vpc_interface_name": flex.StringToFramework(ctx, apiObject.VpcInterfaceName),
		}),
	})
}

func (r *resourceFlow) flattenSource(ctx context.Context, apiObject awstypes.Source) types.List {
	attributeTypes := flex.AttributeTypesMust[sourceData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"name":                                 flex.StringToFramework(ctx, apiObject.Name),
			"arn":                                  flex.StringToFramework(ctx, apiObject.SourceArn),
			"data_transfer_subscriber_fee_percent": flex.Int64ToFramework(ctx, &apiObject.DataTransferSubscriberFeePercent),
			"decryption":                           flex.FlattenFrameworkListNestedBlock[encryptionData](ctx, apiObject.Decryption, r.flattenEncryption),
			"description":                          flex.StringToFramework(ctx, apiObject.Description),
			"entitlement_arn":                      flex.StringToFramework(ctx, apiObject.EntitlementArn),
			"gateway_bridge_source":                flex.FlattenFrameworkListNestedBlock[flowGatewayBridgeSourceData](ctx, apiObject.GatewayBridgeSource, r.flattenFlowGatewayBridgeSource),
			"ingest_ip":                            flex.StringToFramework(ctx, apiObject.IngestIp),
			"ingest_port":                          flex.Int64ToFramework(ctx, &apiObject.IngestPort),
			"media_stream_source_configurations":   flex.FlattenFrameworkListNestedBlock[mediaStreamSourceConfigurationData](ctx, apiObject.MediaStreamSourceConfigurations, r.flattenMediaStreamSourceConfiguration),
			"sender_control_port":                  flex.StringToFramework(ctx, &apiObject.SenderControlPort),
			"sender_ip_address":                    flex.StringToFramework(ctx, apiObject.SenderIpAddress),
			"transport":                            flex.FlattenFrameworkListNestedBlock[transportData](ctx, apiObject.Transport, r.flattenTransport),
			"vpc_interface_name":                   flex.StringToFramework(ctx, apiObject.VpcInterfaceName),
			"whitelist_cidr":                       flex.StringToFramework(ctx, apiObject.WhitelistCidr),
		}),
	})
}

func (r *resourceFlow) flattenFlowGatewayBridgeSource(ctx context.Context, apiObject *awstypes.GatewayBridgeSource) types.List {
	attributeTypes := flex.AttributeTypesMust[flowGatewayBridgeSourceData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"bridge_arn":               flex.StringToFramework(ctx, apiObject.BridgeArn),
			"vpc_interface_attachment": flex.FlattenFrameworkListNestedBlock[vpcInterfaceAttachmentData](ctx, apiObject.VpcInterfaceAttachment, r.flattenVPCInterfaceAttachment),
		}),
	})
}

func (r *resourceFlow) flattenMediaStreamSourceConfiguration(ctx context.Context, apiObject awstypes.MediaStreamSourceConfiguration) types.List {
	attributeTypes := flex.AttributeTypesMust[mediaStreamSourceConfigurationData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"encoding_name":        flex.StringValueToFramework(ctx, apiObject.EncodingName),
			"name":                 flex.StringToFramework(ctx, apiObject.MediaStreamName),
			"input_configurations": flex.FlattenFrameworkListNestedBlock[mediaStreamSourceConfigurationInputConfigurationData](ctx, apiObject.InputConfigurations, r.flattenMediaStreamSourceConfigurationInputConfiguration),
		}),
	})
}

func (r *resourceFlow) flattenMediaStreamSourceConfigurationInputConfiguration(ctx context.Context, apiObject awstypes.InputConfiguration) types.List {
	attributeTypes := flex.AttributeTypesMust[mediaStreamSourceConfigurationInputConfigurationData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"ip":        flex.StringToFramework(ctx, apiObject.InputIp),
			"port":      flex.Int64ToFramework(ctx, &apiObject.InputPort),
			"interface": flex.FlattenFrameworkListNestedBlock[vpcInterfaceData](ctx, apiObject.Interface, r.flattenVPCInterface),
		}),
	})
}

func (r *resourceFlow) flattenSourceFailoverConfig(ctx context.Context, apiObject *awstypes.FailoverConfig) types.List {
	attributeTypes := flex.AttributeTypesMust[sourceFailoverConfigData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"failover_mode":   flex.StringValueToFramework(ctx, apiObject.FailoverMode),
			"recovery_window": flex.Int64ToFramework(ctx, &apiObject.RecoveryWindow),
			"source_priority": flex.FlattenFrameworkListNestedBlock[sourceFailoverConfigSourcePriorityData](ctx, apiObject.SourcePriority, r.flattenSourceFailoverConfigSourcePriority),
			"state":           flex.StringValueToFramework(ctx, apiObject.State),
			"interface":       flex.FlattenFrameworkListNestedBlock[vpcInterfaceData](ctx, apiObject.Interface, r.flattenVPCInterface),
		}),
	})
}

func (r *resourceFlow) flattenSourceFailoverConfigSourcePriority(ctx context.Context, apiObject *awstypes.SourcePriority) types.List {
	attributeTypes := flex.AttributeTypesMust[sourceFailoverConfigData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"primary_source": flex.StringToFramework(ctx, apiObject.PrimarySource),
		}),
	})
}

type resourceFlowData struct {
	FlowArn              fwtypes.ARN  `tfsdk:"arn"`
	Name                 types.String `tfsdk:"name"`
	AvailabilityZone     types.String `tfsdk:"availability_zone"`
	Entitlements         types.List   `tfsdk:"entitlement"`
	Maintenance          types.List   `tfsdk:"maintenance"`
	MediaStreams         types.List   `tfsdk:"media_stream"`
	Outputs              types.List   `tfsdk:"output"`
	Sources              types.List   `tfsdk:"source"`
	SourceFailoverConfig types.List   `tfsdk:"source_failover_config"`
	VpcInterfaces        types.Set    `tfsdk:"vpc_interface"`
	EgressIp             types.String `tfsdk:"egress_ip"`
	Status               types.String `tfsdk:"status"`
	Tags                 types.Map    `tfsdk:"tags"`
	TagsAll              types.Map    `tfsdk:"tags_all"`
}

type entitlementData struct {
	EntitlementArn                   fwtypes.ARN  `tfsdk:"arn"`
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
	Name                             types.String `tfsdk:"name"`
	OutputArn                        fwtypes.ARN  `tfsdk:"arn"`
	BridgeArn                        fwtypes.ARN  `tfsdk:"bridge_arn"`
	DataTransferSubscriberFeePercent types.Int64  `tfsdk:"data_transfer_subscriber_fee_percent"`
	Description                      types.String `tfsdk:"description"`
	Destination                      types.String `tfsdk:"destination"`
	Encryption                       types.List   `tfsdk:"encryption"`
	EntitlementArn                   fwtypes.ARN  `tfsdk:"entitlement_arn"`
	ListenerAddress                  types.String `tfsdk:"listener_address"`
	MediaLiveInputArn                fwtypes.ARN  `tfsdk:"media_live_input_arn"`
	MediaStreamOutputConfigurations  types.Set    `tfsdk:"media_stream_output_configurations"`
	Port                             types.Int64  `tfsdk:"port"`
	Transport                        types.List   `tfsdk:"transport"`
	VpcInterfaceAttachment           types.List   `tfsdk:"vpc_interface_attachment"`
}

type mediaStreamOutputConfigurationData struct {
	EncodingName              types.String `tfsdk:"encoding_name"`
	MediaStreamName           types.String `tfsdk:"name"`
	DestinationConfigurations types.Set    `tfsdk:"destination_configurations"`
	EncodingParameters        types.List   `tfsdk:"encoding_parameters"`
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
	Name                             types.String `tfsdk:"name"`
	SourceArn                        fwtypes.ARN  `tfsdk:"arn"`
	DataTransferSubscriberFeePercent types.Int64  `tfsdk:"data_transfer_subscriber_fee_percent"`
	Decryption                       types.List   `tfsdk:"decryption"`
	Description                      types.String `tfsdk:"description"`
	EntitlementArn                   fwtypes.ARN  `tfsdk:"entitlement_arn"`
	GatewayBridgeSource              types.List   `tfsdk:"gateway_bridge_source"`
	IngestIp                         types.String `tfsdk:"ingest_ip"`
	IngestPort                       types.Int64  `tfsdk:"ingest_port"`
	MaxBitrate                       types.Int64  `tfsdk:"max_bitrate"`
	MaxLatency                       types.Int64  `tfsdk:"max_latency"`
	MinLatency                       types.Int64  `tfsdk:"min_latency"`
	MaxSyncBuffer                    types.Int64  `tfsdk:"max_sync_buffer"`
	MediaStreamSourceConfigurations  types.Set    `tfsdk:"media_stream_source_configurations"`
	Protocol                         types.String `tfsdk:"protocol"`
	SenderControlPort                types.Int64  `tfsdk:"sender_control_port"`
	SenderIpAddress                  types.String `tfsdk:"sender_ip_address"`
	SourceListenerAddress            types.String `tfsdk:"listener_address"`
	SourceListenerPort               types.Int64  `tfsdk:"listener_port"`
	StreamId                         types.Int64  `tfsdk:"stream_id"`
	Transport                        types.List   `tfsdk:"transport"`
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
	InputPort int32        `tfsdk:"port"`
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

type vpcInterfaceData struct {
	Name                 types.String `tfsdk:"name"`
	RoleArn              fwtypes.ARN  `tfsdk:"role_arn"`
	SecurityGroupIds     []string
	SubnetId             types.String `tfsdk:"subnet_id"`
	NetworkInterfaceType types.String `tfsdk:"network_interface_type"`
	NetworkInterfaceIds  types.List   `tfsdk:"network_interface_ids"`
}

type complexArgumentData struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}

var complexArgumentAttrTypes = map[string]attr.Type{
	"nested_required": types.StringType,
	"nested_optional": types.StringType,
}
