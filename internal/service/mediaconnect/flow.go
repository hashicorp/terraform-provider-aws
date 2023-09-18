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
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect/types"
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
)

// @SDKResource("aws_mediaconnect_flow", name="Flow")
// @Tags(identifierAttribute="arn")
func ResourceFlow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFlowCreate,
		ReadWithoutTimeout:   resourceFlowRead,
		UpdateWithoutTimeout: resourceFlowUpdate,
		DeleteWithoutTimeout: resourceFlowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"availability_zone": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"description": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"egress_ip": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"entitlement": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"subscribers": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"data_transfer_subscriber_fee_percent": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(0, 100),
							},
							"description": {
								Type:     schema.TypeString,
								Required: true,
							},
							"encryption": func() *schema.Schema {
								return encryptionSchema()
							}(),
							"entitlement_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								Default:          types.EntitlementStatusEnabled,
								ValidateDiagFunc: enum.Validate[types.EntitlementStatus](),
							},
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"maintenance": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"maintenance_day": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.MaintenanceDay](),
							},
							"maintenance_deadline": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"maintenance_scheduled_date": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"maintenance_start_hour": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"media_stream": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"media_stream_id": {
								Type:     schema.TypeInt,
								Required: true,
							},
							"media_stream_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"media_stream_type": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.MediaStreamType](),
							},
							"attributes": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"fmtp": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"channel_order": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"colorimetry": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.Colorimetry](),
													},
													"exact_framerate": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"par": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"range": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.Range](),
													},
													"scan_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.ScanMode](),
													},
													"tcs": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.Tcs](),
													},
												},
											},
										},
										"lang": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
							"clock_rate": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"description": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"video_format": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"output": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"output_arn": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"bridge_ports": {
								Type:     schema.TypeSet,
								Optional: true,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeInt},
							},
							"data_transfer_subscriber_fee_percent": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"entitlement_arn": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"protocol": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.Protocol](),
							},
							"bridge_arn": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"cidr_allow_list": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.IsCIDRNetwork(0, 128),
								},
							},
							"description": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"destination": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"encryption": func() *schema.Schema {
								return encryptionSchema()
							}(),
							"listener_address": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"max_latency": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"media_live_input_arn": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"media_stream_output_configurations": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"encoding_name": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[types.EncodingName](),
										},
										"media_stream_name": {
											Type:     schema.TypeString,
											Required: true,
										},
										"destination_configurations": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"destination_ip": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.IsIPAddress,
													},
													"destination_port": {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IsPortNumber,
													},
													"interface": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:     schema.TypeString,
																	Required: true,
																},
															},
														},
													},
												},
											},
										},
										"encoding_parameters": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"compression_factor": {
														Type:     schema.TypeFloat,
														Required: true,
													},
													"encoder_profile": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[types.EncoderProfile](),
													},
												},
											},
										},
									},
								},
							},
							"min_latency": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"port": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"remote_id": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"sender_control_port": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"smoothing_latency": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"stream_id": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"transport": func() *schema.Schema {
								return transportSchema()
							}(),
							"vpc_interface_attachment": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"vpc_interface_name": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
				"source_failover_config": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"failover_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.FailoverMode](),
							},
							"recovery_window": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"source_priority": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"primary_source": {
											Type:     schema.TypeString,
											Required: true,
											Computed: true,
										},
									},
								},
							},
							"state": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.State](),
							},
						},
					},
				},
				"source": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"source_arn": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"data_transfer_subscriber_fee_percent": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"decryption": func() *schema.Schema {
								return encryptionSchema()
							}(),
							"description": {
								Type:     schema.TypeString,
								Required: true,
							},
							"entitlement_arn": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
							},
							"gateway_bridge_source": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"bridge_arn": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
										},
										"vpc_interface_attachment": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"vpc_interface_name": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
												},
											},
										},
									},
								},
							},
							"ingest_ip": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"ingest_port": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"max_bitrate": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"max_latency": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"max_sync_buffer": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"media_stream_source_configurations": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"encoding_name": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[types.EncodingName](),
										},
										"media_stream_name": {
											Type:     schema.TypeString,
											Required: true,
										},
										"input_configurations": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"input_port": {
														Type:     schema.TypeInt,
														Required: true,
													},
													"interface": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"name": {
																	Type:     schema.TypeString,
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
							"min_latency": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"protocol": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.Protocol](),
							},
							"sender_control_port": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"sender_ip_address": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IsIPAddress,
							},
							"source_listener_address": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"source_listener_port": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"stream_id": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"transport": func() *schema.Schema {
								return transportSchema()
							}(),
							"vpc_interface_name": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"whitelist_cidr": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.IsCIDRNetwork(0, 128),
							},
						},
					},
				},
				"start_flow": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"vpc_interface": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"role_arn": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
							},
							"security_group_ids": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"subnet_id": {
								Type:     schema.TypeString,
								Required: true,
							},
							"network_interface_ids": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"network_interface_type": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.NetworkInterfaceTypeEna,
								ValidateDiagFunc: enum.Validate[types.NetworkInterfaceType](),
							},
						},
					},
				},

				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func encryptionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"role_arn": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
				},
				"algorithm": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[types.Algorithm](),
				},
				"constant_initialization_vector": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"device_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"key_type": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.KeyType](),
				},
				"region": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"resource_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"secret_arn": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
				},
				"url": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

func transportSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,

		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"protocol": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[types.Protocol](),
				},
				"cidr_allow_list": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.IsCIDRNetwork(0, 128),
					},
				},
				"max_bitrate": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"max_latency": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"max_sync_buffer": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"min_latency": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"remote_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"sender_control_port": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"sender_ip_address": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.IsIPAddress,
				},
				"smoothing_latency": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"source_listener_address": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"source_listener_port": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"stream_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

const (
	ResNameFlow = "Flow"
)

func resourceFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaConnectClient(ctx)

	in := &mediaconnect.CreateFlowInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		in.AvailabilityZone = aws.String(v.(string))
	}
	if v, ok := d.GetOk("entitlement"); ok && v.(*schema.Set).Len() > 0 {
		in.Entitlements = expandFlowEntitlements(v.([]interface{}))
	}
	if v, ok := d.GetOk("maintenance"); ok && len(v.([]interface{})) > 0 {
		in.Maintenance = expandFlowMaintenanceCreate(v.([]interface{}))
	}
	if v, ok := d.GetOk("media_stream"); ok && v.(*schema.Set).Len() > 0 {
		in.MediaStreams = expandFlowMediaStreams(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("output"); ok && v.(*schema.Set).Len() > 0 {
		in.Outputs = expandFlowOutputs(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("source_failover_config"); ok && len(v.([]interface{})) > 0 {
		in.SourceFailoverConfig = expandFlowSourceFailoverConfigCreate(v.([]interface{}))
	}
	if v, ok := d.GetOk("source"); ok && v.(*schema.Set).Len() > 0 {
		sources := expandFlowSources(v.(*schema.Set).List())
		in.Source = &sources[0]
		in.Sources = sources[1:]
	}
	if v, ok := d.GetOk("vpc_interface"); ok && v.(*schema.Set).Len() > 0 {
		in.VpcInterfaces = expandFlowVPCInterfaces(v.(*schema.Set).List())
	}

	out, err := conn.CreateFlow(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionCreating, ResNameFlow, d.Get("name").(string), err)
	}

	if out == nil || out.Flow == nil {
		return create.DiagError(names.MediaConnect, create.ErrActionCreating, ResNameFlow, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Flow.FlowArn))

	if _, err := waitFlowCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionWaitingForCreation, ResNameFlow, d.Id(), err)
	}

	if d.Get("start_flow").(bool) {
		if err := startFlow(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
			return create.DiagError(names.MediaConnect, create.ErrActionCreating, ResNameFlow, d.Get("name").(string), err)
		}
	}

	return resourceFlowRead(ctx, d, meta)
}

func resourceFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaConnectClient(ctx)

	out, err := FindFlowByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaConnect Flow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionReading, ResNameFlow, d.Id(), err)
	}

	d.Set("arn", out.FlowArn)
	d.Set("name", out.Name)
	d.Set("availability_zone", out.AvailabilityZone)
	d.Set("description", out.Description)
	d.Set("egress_ip", out.EgressIp)

	if err := d.Set("entitlement", flattenFlowEntitlements(out.Entitlements)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameFlow, d.Id(), err)
	}
	if err := d.Set("output", flattenFlowOutputs(out.Outputs)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameFlow, d.Id(), err)
	}
	if err := d.Set("maintenance", flattenFlowMaintenance(out.Maintenance)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameFlow, d.Id(), err)
	}
	if err := d.Set("media_stream", flattenFlowMediaStreams(out.MediaStreams)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameFlow, d.Id(), err)
	}
	if err := d.Set("source_failover_config", flattenFlowSourceFailoverConfig(out.SourceFailoverConfig)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameFlow, d.Id(), err)
	}
	sources := []types.Source{*out.Source}
	sources = append(sources, out.Sources...)
	if err := d.Set("source", flattenFlowSources(sources)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameFlow, d.Id(), err)
	}
	if err := d.Set("vpc_interface", flattenFlowVPCInterfaces(out.VpcInterfaces)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameFlow, d.Id(), err)
	}

	return nil
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaConnectClient(ctx)

	if d.HasChangesExcept("tags", "tags_all", "start_flow") {
		in := &mediaconnect.UpdateFlowInput{
			FlowArn: aws.String(d.Id()),
		}

		if d.HasChange("maintenance") {
			in.Maintenance = expandFlowMaintenanceUpdate(d.Get("maintenance").([]interface{}))
		}

		if d.HasChange("source_failover_config") {
			in.SourceFailoverConfig = expandFlowSourceFailoverConfigUpdate(d.Get("source_failover_config").([]interface{}))
		}

		flow, err := FindFlowByARN(ctx, conn, d.Id())

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
		flow, err := FindFlowByARN(ctx, conn, d.Id())

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

	flow, err := FindFlowByARN(ctx, conn, d.Id())

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
		out, err := FindFlowByARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindFlowByARN(ctx context.Context, conn *mediaconnect.Client, arn string) (*types.Flow, error) {
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

func expandFlowEntitlements(tfList []interface{}) []types.GrantEntitlementRequest {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.GrantEntitlementRequest
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.GrantEntitlementRequest
		if v, ok := m["subscribers"].([]string); ok && len(v) > 0 {
			o.Subscribers = v
		}
		if v, ok := m["data_transfer_subscriber_fee_percent"].(int); ok {
			o.DataTransferSubscriberFeePercent = int32(v)
		}
		if v, ok := m["description"].(string); ok && v != "" {
			o.Description = aws.String(v)
		}
		if v, ok := m["encryption"].([]interface{}); ok && len(v) > 0 {
			o.Encryption = expandEncryption(v)
		}
		if v, ok := m["entitlement_status"].(string); ok && v != "" {
			o.EntitlementStatus = types.EntitlementStatus(v)
		}
		if v, ok := m["name"].(string); ok && v != "" {
			o.Name = aws.String(v)
		}

		out = append(out, o)
	}

	return out
}

func expandEncryption(tfList []interface{}) *types.Encryption {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.Encryption
	if v, ok := m["role_arn"].(string); ok && v != "" {
		out.RoleArn = aws.String(v)
	}
	if v, ok := m["algorithm"].(string); ok && v != "" {
		out.Algorithm = types.Algorithm(v)
	}
	if v, ok := m["constant_initialization_vector"].(string); ok && v != "" {
		out.ConstantInitializationVector = aws.String(v)
	}
	if v, ok := m["device_id"].(string); ok && v != "" {
		out.DeviceId = aws.String(v)
	}
	if v, ok := m["key_type"].(string); ok && v != "" {
		out.KeyType = types.KeyType(v)
	}
	if v, ok := m["region"].(string); ok && v != "" {
		out.Region = aws.String(v)
	}
	if v, ok := m["resource_id"].(string); ok && v != "" {
		out.ResourceId = aws.String(v)
	}
	if v, ok := m["secret_arn"].(string); ok && v != "" {
		out.SecretArn = aws.String(v)
	}
	if v, ok := m["url"].(string); ok && v != "" {
		out.Url = aws.String(v)
	}

	return &out
}

func expandFlowMediaStreams(tfList []interface{}) []types.AddMediaStreamRequest {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.AddMediaStreamRequest
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.AddMediaStreamRequest
		if v, ok := m["media_stream_id"].(int); ok {
			o.MediaStreamId = int32(v)
		}
		if v, ok := m["media_stream_name"].(string); ok {
			o.MediaStreamName = aws.String(v)
		}
		if v, ok := m["media_stream_type"].(string); ok {
			o.MediaStreamType = types.MediaStreamType(v)
		}
		if v, ok := m["attributes"].([]interface{}); ok && len(v) > 0 {
			o.Attributes = expandFlowMediaStreamsAttributes(v)
		}
		if v, ok := m["clock_rate"].(int); ok {
			o.ClockRate = int32(v)
		}
		if v, ok := m["description"].(string); ok {
			o.Description = aws.String(v)
		}
		if v, ok := m["video_format"].(string); ok {
			o.VideoFormat = aws.String(v)
		}

		out = append(out, o)
	}

	return out
}

func expandFlowMediaStreamsAttributes(tfList []interface{}) *types.MediaStreamAttributesRequest {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.MediaStreamAttributesRequest
	if v, ok := m["fmtp"].([]interface{}); ok && len(v) > 0 {
		out.Fmtp = expandFlowMediaStreamsAttributesFmtp(v)
	}
	if v, ok := m["lang"].(string); ok && v != "" {
		out.Lang = aws.String(v)
	}

	return &out
}

func expandFlowMediaStreamsAttributesFmtp(tfList []interface{}) *types.FmtpRequest {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.FmtpRequest
	if v, ok := m["channel_order"].(string); ok && v != "" {
		out.ChannelOrder = aws.String(v)
	}
	if v, ok := m["colorimetry"].(string); ok && v != "" {
		out.Colorimetry = types.Colorimetry(v)
	}
	if v, ok := m["exact_framerate"].(string); ok && v != "" {
		out.ExactFramerate = aws.String(v)
	}
	if v, ok := m["par"].(string); ok && v != "" {
		out.Par = aws.String(v)
	}
	if v, ok := m["range"].(string); ok && v != "" {
		out.Range = types.Range(v)
	}
	if v, ok := m["scan_mode"].(string); ok && v != "" {
		out.ScanMode = types.ScanMode(v)
	}
	if v, ok := m["tcs"].(string); ok && v != "" {
		out.Tcs = types.Tcs(v)
	}

	return &out
}

func expandFlowOutputs(tfList []interface{}) []types.AddOutputRequest {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.AddOutputRequest
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.AddOutputRequest
		if v, ok := m["protocol"].(string); ok && v != "" {
			o.Protocol = types.Protocol(v)
		}
		if v, ok := m["cidr_allow_list"].([]string); ok && len(v) > 0 {
			o.CidrAllowList = v
		}
		if v, ok := m["description"].(string); ok && v != "" {
			o.Description = aws.String(v)
		}
		if v, ok := m["destination"].(string); ok && v != "" {
			o.Destination = aws.String(v)
		}
		if v, ok := m["encryption"].([]interface{}); ok && len(v) > 0 {
			o.Encryption = expandEncryption(v)
		}
		if v, ok := m["max_latency"].(int); ok {
			o.MaxLatency = int32(v)
		}
		if v, ok := m["media_stream_output_configurations"].(*schema.Set); ok && v.Len() > 0 {
			o.MediaStreamOutputConfigurations = expandFlowOutputsMediaStreamOutputConfigurations(v.List())
		}
		if v, ok := m["min_latency"].(int); ok {
			o.MinLatency = int32(v)
		}
		if v, ok := m["name"].(string); ok && v != "" {
			o.Name = aws.String(v)
		}
		if v, ok := m["port"].(int); ok {
			o.Port = int32(v)
		}
		if v, ok := m["remote_id"].(string); ok && v != "" {
			o.RemoteId = aws.String(v)
		}
		if v, ok := m["sender_control_port"].(int); ok {
			o.SenderControlPort = int32(v)
		}
		if v, ok := m["smoothing_latency"].(int); ok {
			o.SmoothingLatency = int32(v)
		}
		if v, ok := m["stream_id"].(string); ok && v != "" {
			o.StreamId = aws.String(v)
		}
		if v, ok := m["vpc_interface_attachment"].([]interface{}); ok && len(v) > 0 {
			o.VpcInterfaceAttachment = expandFlowVPCInterfaceAttachment(v)
		}

		out = append(out, o)
	}

	return out
}

func expandFlowOutputsMediaStreamOutputConfigurations(tfList []interface{}) []types.MediaStreamOutputConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.MediaStreamOutputConfigurationRequest
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.MediaStreamOutputConfigurationRequest
		if v, ok := m["encoding_name"].(string); ok && v != "" {
			o.EncodingName = types.EncodingName(v)
		}
		if v, ok := m["media_stream_name"].(string); ok && v != "" {
			o.MediaStreamName = aws.String(v)
		}
		if v, ok := m["destination_configurations"].(*schema.Set); ok && v.Len() > 0 {
			o.DestinationConfigurations = expandFlowOutputsMediaStreamOutputConfigurationsDestinationConfigurations(v.List())
		}
		if v, ok := m["encoding_parameters"].([]interface{}); ok && len(v) > 0 {
			o.EncodingParameters = expandFlowOutputsMediaStreamOutputConfigurationsEncodingParameters(v)
		}

		out = append(out, o)
	}

	return out
}

func expandFlowOutputsMediaStreamOutputConfigurationsDestinationConfigurations(tfList []interface{}) []types.DestinationConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.DestinationConfigurationRequest
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.DestinationConfigurationRequest
		if v, ok := m["destination_ip"].(string); ok && v != "" {
			o.DestinationIp = aws.String(v)
		}
		if v, ok := m["destination_port"].(int); ok {
			o.DestinationPort = int32(v)
		}
		if v, ok := m["interface"].([]interface{}); ok && len(v) > 0 {
			o.Interface = expandFlowInterface(v)
		}

		out = append(out, o)
	}

	return out
}

func expandFlowInterface(tfList []interface{}) *types.InterfaceRequest {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.InterfaceRequest
	if v, ok := m["name"].(string); ok && v != "" {
		out.Name = aws.String(v)
	}

	return &out
}

func expandFlowOutputsMediaStreamOutputConfigurationsEncodingParameters(tfList []interface{}) *types.EncodingParametersRequest {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.EncodingParametersRequest
	if v, ok := m["compression_factor"].(float32); ok {
		out.CompressionFactor = float64(v)
	}
	if v, ok := m["encoder_profile"].(string); ok && v != "" {
		out.EncoderProfile = types.EncoderProfile(v)
	}

	return &out
}

func expandFlowVPCInterfaceAttachment(tfList []interface{}) *types.VpcInterfaceAttachment {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.VpcInterfaceAttachment
	if v, ok := m["vpc_interface_name"].(string); ok && v != "" {
		out.VpcInterfaceName = aws.String(v)
	}

	return &out
}

func expandFlowSources(tfList []interface{}) []types.SetSourceRequest {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.SetSourceRequest
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.SetSourceRequest
		if v, ok := m["decryption"].([]interface{}); ok && len(v) > 0 {
			o.Decryption = expandEncryption(v)
		}
		if v, ok := m["description"].(string); ok && v != "" {
			o.Description = aws.String(v)
		}
		if v, ok := m["entitlement_arn"].(string); ok && v != "" {
			o.EntitlementArn = aws.String(v)
		}
		if v, ok := m["gateway_bridge_source"].([]interface{}); ok && len(v) > 0 {
			o.GatewayBridgeSource = expandFlowSourcesGatewayBridgeSource(v)
		}
		if v, ok := m["ingest_port"].(int); ok {
			o.IngestPort = int32(v)
		}
		if v, ok := m["max_bitrate"].(int); ok {
			o.MaxBitrate = int32(v)
		}
		if v, ok := m["max_latency"].(int); ok {
			o.MaxLatency = int32(v)
		}
		if v, ok := m["max_sync_buffer"].(int); ok {
			o.MaxSyncBuffer = int32(v)
		}
		if v, ok := m["media_stream_source_configurations"].(*schema.Set); ok && v.Len() > 0 {
			o.MediaStreamSourceConfigurations = expandFlowSourcesMediaStreamSourceConfigurations(v.List())
		}
		if v, ok := m["min_latency"].(int); ok {
			o.MinLatency = int32(v)
		}
		if v, ok := m["name"].(string); ok && v != "" {
			o.Name = aws.String(v)
		}
		if v, ok := m["protocol"].(string); ok && v != "" {
			o.Protocol = types.Protocol(v)
		}
		if v, ok := m["sender_control_port"].(int); ok {
			o.SenderControlPort = int32(v)
		}
		if v, ok := m["sender_ip_address"].(string); ok && v != "" {
			o.SenderIpAddress = aws.String(v)
		}
		if v, ok := m["source_listener_address"].(string); ok && v != "" {
			o.SourceListenerAddress = aws.String(v)
		}
		if v, ok := m["source_listener_port"].(int); ok {
			o.SourceListenerPort = int32(v)
		}
		if v, ok := m["stream_id"].(string); ok && v != "" {
			o.StreamId = aws.String(v)
		}
		if v, ok := m["vpc_interface_name"].(string); ok && v != "" {
			o.VpcInterfaceName = aws.String(v)
		}
		if v, ok := m["whitelist_cidr"].(string); ok && v != "" {
			o.WhitelistCidr = aws.String(v)
		}

		out = append(out, o)
	}

	return out
}

func expandFlowSourcesGatewayBridgeSource(tfList []interface{}) *types.SetGatewayBridgeSourceRequest {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.SetGatewayBridgeSourceRequest
	if v, ok := m["bridge_arn"].(string); ok && v != "" {
		out.BridgeArn = aws.String(v)
	}
	if v, ok := m["vpc_interface_attachment"].([]interface{}); ok && len(v) > 0 {
		out.VpcInterfaceAttachment = expandFlowVPCInterfaceAttachment(v)
	}

	return &out
}

func expandFlowSourcesMediaStreamSourceConfigurations(tfList []interface{}) []types.MediaStreamSourceConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.MediaStreamSourceConfigurationRequest
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.MediaStreamSourceConfigurationRequest
		if v, ok := m["encoding_name"].(string); ok && v != "" {
			o.EncodingName = types.EncodingName(v)
		}
		if v, ok := m["media_stream_name"].(string); ok && v != "" {
			o.MediaStreamName = aws.String(v)
		}
		if v, ok := m["input_configurations"].(*schema.Set); ok && v.Len() > 0 {
			o.InputConfigurations = expandFlowSourcesMediaStreamSourceConfigurationsInputConfigurations(v.List())
		}

		out = append(out, o)
	}

	return out
}

func expandFlowSourcesMediaStreamSourceConfigurationsInputConfigurations(tfList []interface{}) []types.InputConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.InputConfigurationRequest
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.InputConfigurationRequest
		if v, ok := m["input_port"].(int); ok {
			o.InputPort = int32(v)
		}
		if v, ok := m["interface"].([]interface{}); ok && len(v) > 0 {
			o.Interface = expandFlowInterface(v)
		}

		out = append(out, o)
	}

	return out
}

func expandFlowVPCInterfaces(tfList []interface{}) []types.VpcInterfaceRequest {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.VpcInterfaceRequest
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.VpcInterfaceRequest
		if v, ok := m["name"].(string); ok {
			o.Name = aws.String(v)
		}
		if v, ok := m["role_arn"].(string); ok {
			o.RoleArn = aws.String(v)
		}
		if v, ok := m["security_group_ids"].([]string); ok && len(v) > 0 {
			o.SecurityGroupIds = v
		}
		if v, ok := m["subnet_id"].(string); ok && v != "" {
			o.SubnetId = aws.String(v)
		}
		if v, ok := m["network_interface_type"].(string); ok && v != "" {
			o.NetworkInterfaceType = types.NetworkInterfaceType(v)
		}

		out = append(out, o)
	}

	return out
}

func expandFlowMaintenanceCreate(tfList []interface{}) *types.AddMaintenance {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.AddMaintenance{}
	if v, ok := m["maintenance_day"].(string); ok && v != "" {
		settings.MaintenanceDay = types.MaintenanceDay(v)
	}
	if v, ok := m["maintenance_start_hour"].(string); ok && v != "" {
		settings.MaintenanceStartHour = aws.String(v)
	}

	return settings
}

func expandFlowMaintenanceUpdate(tfList []interface{}) *types.UpdateMaintenance {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.UpdateMaintenance{}
	if v, ok := m["maintenance_day"].(string); ok && v != "" {
		settings.MaintenanceDay = types.MaintenanceDay(v)
	}
	if v, ok := m["maintenance_start_hour"].(string); ok && v != "" {
		settings.MaintenanceStartHour = aws.String(v)
	}
	// NOTE: This field is only available in the update struct. To allow users to set a scheduled
	// date on update, it may be worth adding to the base schema.
	// if v, ok := m["maintenance_scheduled_date"].(string); ok && v != "" {
	//  settings.MaintenanceScheduledDate = aws.String(v)
	// }

	return settings
}

func expandFlowSourceFailoverConfigCreate(tfList []interface{}) *types.FailoverConfig {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.FailoverConfig{}
	if v, ok := m["failover_mode"].(string); ok && v != "" {
		settings.FailoverMode = types.FailoverMode(v)
	}
	if v, ok := m["recovery_window"].(int); ok {
		settings.RecoveryWindow = int32(v)
	}
	if v, ok := m["source_priority"].([]interface{}); ok && len(v) > 0 {
		settings.SourcePriority = expandFlowSourceFailoverConfigSourcePriority(v)
	}
	if v, ok := m["state"].(string); ok && v != "" {
		settings.State = types.State(v)
	}

	return settings
}

func expandFlowSourceFailoverConfigUpdate(tfList []interface{}) *types.UpdateFailoverConfig {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.UpdateFailoverConfig{}
	if v, ok := m["failover_mode"].(string); ok && v != "" {
		settings.FailoverMode = types.FailoverMode(v)
	}
	if v, ok := m["recovery_window"].(int); ok {
		settings.RecoveryWindow = int32(v)
	}
	if v, ok := m["source_priority"].([]interface{}); ok && len(v) > 0 {
		settings.SourcePriority = expandFlowSourceFailoverConfigSourcePriority(v)
	}
	if v, ok := m["state"].(string); ok && v != "" {
		settings.State = types.State(v)
	}

	return settings
}

func expandFlowSourceFailoverConfigSourcePriority(tfList []interface{}) *types.SourcePriority {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.SourcePriority{}
	if v, ok := m["primary_source"].(string); ok && v != "" {
		settings.PrimarySource = aws.String(v)
	}

	return settings
}

func flattenFlowEntitlements(tfList []types.Entitlement) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"entitlement_arn":                      aws.ToString(item.EntitlementArn),
			"name":                                 aws.ToString(item.Name),
			"subscribers":                          flex.FlattenStringValueList(item.Subscribers),
			"data_transfer_subscriber_fee_percent": int(item.DataTransferSubscriberFeePercent),
			"description":                          aws.ToString(item.Description),
			"encryption":                           flattenEncryption(item.Encryption),
			"entitlement_status":                   string(item.EntitlementStatus),
		}

		out = append(out, m)
	}

	return out
}

func flattenEncryption(in *types.Encryption) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"role_arn":                       aws.ToString(in.RoleArn),
		"algorithm":                      string(in.Algorithm),
		"constant_initialization_vector": aws.ToString(in.ConstantInitializationVector),
		"device_id":                      aws.ToString(in.DeviceId),
		"key_type":                       string(in.KeyType),
		"region":                         aws.ToString(in.Region),
		"resource_id":                    aws.ToString(in.ResourceId),
		"secret_arn":                     aws.ToString(in.SecretArn),
		"url":                            aws.ToString(in.Url),
	}

	return []interface{}{m}
}

func flattenFlowOutputs(tfList []types.Output) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"name":                                 aws.ToString(item.Name),
			"output_arn":                           aws.ToString(item.OutputArn),
			"bridge_arn":                           aws.ToString(item.BridgeArn),
			"bridge_ports":                         item.BridgePorts,
			"data_transfer_subscriber_fee_percent": int(item.DataTransferSubscriberFeePercent),
			"description":                          aws.ToString(item.Description),
			"destination":                          aws.ToString(item.Destination),
			"encryption":                           flattenEncryption(item.Encryption),
			"entitlement_arn":                      aws.ToString(item.EntitlementArn),
			"listener_address":                     aws.ToString(item.ListenerAddress),
			"media_live_input_arn":                 aws.ToString(item.MediaLiveInputArn),
			"media_stream_output_configurations":   flattenFlowOutputsMediaStreamOutputConfigurations(item.MediaStreamOutputConfigurations),
			"port":                                 int(item.Port),
			"transport":                            flattenFlowTransport(item.Transport),
			"vpc_interface_attachment":             flattenFlowVPCInterfaceAttachment(item.VpcInterfaceAttachment),
		}

		out = append(out, m)
	}

	return out
}

func flattenFlowOutputsMediaStreamOutputConfigurations(tfList []types.MediaStreamOutputConfiguration) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"encoding_name":              string(item.EncodingName),
			"media_stream_name":          aws.ToString(item.MediaStreamName),
			"destination_configurations": flattenFlowOutputsMediaStreamOutputConfigurationsDestinationConfigurations(item.DestinationConfigurations),
			"encoding_parameters":        flattenFlowOutputsMediaStreamOutputConfigurationsEncodingParameters(item.EncodingParameters),
		}

		out = append(out, m)
	}

	return out
}

func flattenFlowOutputsMediaStreamOutputConfigurationsDestinationConfigurations(tfList []types.DestinationConfiguration) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"destination_ip":   aws.ToString(item.DestinationIp),
			"destination_port": int(item.DestinationPort),
			"interface":        flattenFlowInterface(item.Interface),
			"outbound_ip":      aws.ToString(item.OutboundIp),
		}

		out = append(out, m)
	}

	return out
}

func flattenFlowOutputsMediaStreamOutputConfigurationsEncodingParameters(in *types.EncodingParameters) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"compression_factor": float32(in.CompressionFactor),
		"encoder_profile":    string(in.EncoderProfile),
	}

	return []interface{}{m}
}

func flattenFlowTransport(in *types.Transport) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"protocol":                string(in.Protocol),
		"cidr_allow_list":         flex.FlattenStringValueList(in.CidrAllowList),
		"max_bitrate":             int(in.MaxBitrate),
		"max_latency":             int(in.MaxLatency),
		"max_sync_buffer":         int(in.MaxSyncBuffer),
		"min_latency":             int(in.MinLatency),
		"remote_id":               aws.ToString(in.RemoteId),
		"sender_control_port":     int(in.SenderControlPort),
		"sender_ip_address":       aws.ToString(in.SenderIpAddress),
		"smoothing_latency":       int(in.SmoothingLatency),
		"source_listener_address": aws.ToString(in.SourceListenerAddress),
		"source_listener_port":    int(in.SourceListenerPort),
		"stream_id":               aws.ToString(in.StreamId),
	}

	return []interface{}{m}
}

func flattenFlowVPCInterfaces(tfList []types.VpcInterface) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"name":                   aws.ToString(item.Name),
			"network_interface_ids":  flex.FlattenStringValueList(item.NetworkInterfaceIds),
			"network_interface_type": string(item.NetworkInterfaceType),
			"role_arn":               aws.ToString(item.RoleArn),
			"security_group_ids":     flex.FlattenStringValueList(item.SecurityGroupIds),
			"subnet_id":              aws.ToString(item.SubnetId),
		}

		out = append(out, m)
	}

	return out
}

func flattenFlowVPCInterfaceAttachment(in *types.VpcInterfaceAttachment) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"vpc_interface_name": aws.ToString(in.VpcInterfaceName),
	}

	return []interface{}{m}
}

func flattenFlowMaintenance(in *types.Maintenance) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"maintenance_day":            string(in.MaintenanceDay),
		"maintenance_deadline":       aws.ToString(in.MaintenanceDeadline),
		"maintenance_scheduled_date": aws.ToString(in.MaintenanceScheduledDate),
		"maintenance_start_hour":     aws.ToString(in.MaintenanceStartHour),
	}

	return []interface{}{m}
}

func flattenFlowMediaStreams(tfList []types.MediaStream) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"fmt":               int(item.Fmt),
			"media_stream_id":   int(item.MediaStreamId),
			"media_stream_name": aws.ToString(item.MediaStreamName),
			"media_stream_type": string(item.MediaStreamType),
			"attributes":        flattenFlowMediaStreamsAttributes(item.Attributes),
			"clock_rate":        int(item.ClockRate),
			"description":       aws.ToString(item.Description),
			"video_format":      aws.ToString(item.VideoFormat),
		}

		out = append(out, m)
	}

	return out
}

func flattenFlowMediaStreamsAttributes(in *types.MediaStreamAttributes) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"fmtp": flattenFlowMediaStreamsAttributesFmtp(in.Fmtp),
		"lang": aws.ToString(in.Lang),
	}

	return []interface{}{m}
}

func flattenFlowMediaStreamsAttributesFmtp(in *types.Fmtp) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"channel_order":   aws.ToString(in.ChannelOrder),
		"colorimetry":     string(in.Colorimetry),
		"exact_framerate": aws.ToString(in.ExactFramerate),
		"par":             aws.ToString(in.Par),
		"range":           string(in.Range),
		"scan_mode":       string(in.ScanMode),
		"tcs":             string(in.Tcs),
	}

	return []interface{}{m}
}

func flattenFlowSourceFailoverConfig(in *types.FailoverConfig) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"failover_mode":   string(in.FailoverMode),
		"recovery_window": int(in.RecoveryWindow),
		"source_priority": flattenFlowSourceFailoverConfigSourcePriority(in.SourcePriority),
		"state":           string(in.State),
	}

	return []interface{}{m}
}

func flattenFlowSourceFailoverConfigSourcePriority(in *types.SourcePriority) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"primary_source": aws.ToString(in.PrimarySource),
	}

	return []interface{}{m}
}

func flattenFlowSources(tfList []types.Source) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"name":                                 aws.ToString(item.Name),
			"source_arn":                           aws.ToString(item.SourceArn),
			"data_transfer_subscriber_fee_percent": int(item.DataTransferSubscriberFeePercent),
			"decryption":                           flattenEncryption(item.Decryption),
			"description":                          aws.ToString(item.Description),
			"entitlement_arn":                      aws.ToString(item.EntitlementArn),
			"gateway_bridge_source":                flattenFlowSourcesGatewayBridgeSource(item.GatewayBridgeSource),
			"ingest_ip":                            aws.ToString(item.IngestIp),
			"ingest_port":                          int(item.IngestPort),
			"media_stream_source_configurations":   flattenFlowSourcesMediaStreamSourceConfigurations(item.MediaStreamSourceConfigurations),
			"sender_control_port":                  int(item.SenderControlPort),
			"sender_ip_address":                    aws.ToString(item.SenderIpAddress),
			"transport":                            flattenFlowTransport(item.Transport),
			"vpc_interface_name":                   aws.ToString(item.VpcInterfaceName),
			"whitelist_cidr":                       aws.ToString(item.WhitelistCidr),
		}

		out = append(out, m)
	}

	return out
}

func flattenFlowSourcesGatewayBridgeSource(in *types.GatewayBridgeSource) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"bridge_arn":               aws.ToString(in.BridgeArn),
		"vpc_interface_attachment": flattenFlowVPCInterfaceAttachment(in.VpcInterfaceAttachment),
	}

	return []interface{}{m}
}

func flattenFlowSourcesMediaStreamSourceConfigurations(tfList []types.MediaStreamSourceConfiguration) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"encoding_name":        string(item.EncodingName),
			"media_stream_name":    aws.ToString(item.MediaStreamName),
			"input_configurations": flattenFlowSourcesMediaStreamSourceConfigurationsInputConfigurations(item.InputConfigurations),
		}

		out = append(out, m)
	}

	return out
}

func flattenFlowSourcesMediaStreamSourceConfigurationsInputConfigurations(tfList []types.InputConfiguration) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"input_ip":   aws.ToString(item.InputIp),
			"input_port": int(item.InputPort),
			"interface":  flattenFlowInterface(item.Interface),
		}

		out = append(out, m)
	}

	return out
}

func flattenFlowInterface(in *types.Interface) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"name": aws.ToString(in.Name),
	}

	return []interface{}{m}
}
