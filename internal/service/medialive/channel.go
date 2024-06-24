// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_medialive_channel", name="Channel")
// @Tags(identifierAttribute="arn")
func ResourceChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceChannelCreate,
		ReadWithoutTimeout:   resourceChannelRead,
		UpdateWithoutTimeout: resourceChannelUpdate,
		DeleteWithoutTimeout: resourceChannelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"cdi_input_specification": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"resolution": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.CdiInputResolution](),
							},
						},
					},
				},
				"channel_class": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[types.ChannelClass](),
				},
				"channel_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"destinations": {
					Type:     schema.TypeSet,
					Required: true,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrID: {
								Type:     schema.TypeString,
								Required: true,
							},
							"media_package_settings": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"channel_id": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"multiplex_settings": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"multiplex_id": {
											Type:     schema.TypeString,
											Required: true,
										},
										"program_name": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"settings": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"password_param": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"stream_name": {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrURL: {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrUsername: {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
				"encoder_settings": func() *schema.Schema {
					return channelEncoderSettingsSchema()
				}(),
				"input_attachments": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"automatic_input_failover_settings": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"secondary_input_id": {
											Type:     schema.TypeString,
											Required: true,
										},
										"error_clear_time_msec": {
											Type:     schema.TypeInt,
											Optional: true,
										},
										"failover_condition": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"failover_condition_settings": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"audio_silence_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"audio_selector_name": {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																			"audio_silence_threshold_msec": {
																				Type:     schema.TypeInt,
																				Optional: true,
																			},
																		},
																	},
																},
																"input_loss_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"input_loss_threshold_msec": {
																				Type:     schema.TypeInt,
																				Optional: true,
																			},
																		},
																	},
																},
																"video_black_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"black_detect_threshold": {
																				Type:     schema.TypeFloat,
																				Optional: true,
																			},
																			"video_black_threshold_msec": {
																				Type:     schema.TypeInt,
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
										"input_preference": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.InputPreference](),
										},
									},
								},
							},
							"input_attachment_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"input_id": {
								Type:     schema.TypeString,
								Required: true,
							},
							"input_settings": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"audio_selector": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrName: {
														Type:     schema.TypeString,
														Required: true,
													},
													"selector_settings": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"audio_hls_rendition_selection": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"group_id": {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																			names.AttrName: {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																		},
																	},
																},
																"audio_language_selection": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			names.AttrLanguageCode: {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																			"language_selection_policy": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.AudioLanguageSelectionPolicy](),
																			},
																		},
																	},
																},
																"audio_pid_selection": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"pid": {
																				Type:     schema.TypeInt,
																				Required: true,
																			},
																		},
																	},
																},
																"audio_track_selection": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"dolby_e_decode": {
																				Type:     schema.TypeList,
																				Optional: true,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"program_selection": {
																							Type:             schema.TypeString,
																							Required:         true,
																							ValidateDiagFunc: enum.Validate[types.DolbyEProgramSelection](),
																						},
																					},
																				},
																			},
																			"tracks": {
																				Type:     schema.TypeSet,
																				Required: true,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"track": {
																							Type:     schema.TypeInt,
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
										"caption_selector": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrName: {
														Type:     schema.TypeString,
														Required: true,
													},
													names.AttrLanguageCode: {
														Type:     schema.TypeString,
														Optional: true,
													},
													"selector_settings": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"ancillary_source_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"source_ancillary_channel_number": {
																				Type:     schema.TypeInt,
																				Optional: true,
																			},
																		},
																	},
																},
																"arib_source_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{}, // no exported elements in this list
																	},
																},
																"dvb_sub_source_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"ocr_language": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.DvbSubOcrLanguage](),
																			},
																			"pid": {
																				Type:         schema.TypeInt,
																				Optional:     true,
																				ValidateFunc: validation.IntAtLeast(1),
																			},
																		},
																	},
																},
																"embedded_source_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"convert_608_to_708": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.EmbeddedConvert608To708](),
																			},
																			"scte20_detection": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.EmbeddedScte20Detection](),
																			},
																			"source_608_channel_number": {
																				Type:     schema.TypeInt,
																				Optional: true,
																			},
																		},
																	},
																},
																"scte20_source_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"convert_608_to_708": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.Scte20Convert608To708](),
																			},
																			"source_608_channel_number": {
																				Type:     schema.TypeInt,
																				Optional: true,
																			},
																		},
																	},
																},
																"scte27_source_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"ocr_language": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.Scte27OcrLanguage](),
																			},
																			"pid": {
																				Type:     schema.TypeInt,
																				Optional: true,
																			},
																		},
																	},
																},
																"teletext_source_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"output_rectangle": {
																				Type:     schema.TypeList,
																				Optional: true,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"height": {
																							Type:     schema.TypeFloat,
																							Required: true,
																						},
																						"left_offset": {
																							Type:     schema.TypeFloat,
																							Required: true,
																						},
																						"top_offset": {
																							Type:     schema.TypeFloat,
																							Required: true,
																						},
																						"width": {
																							Type:     schema.TypeFloat,
																							Required: true,
																						},
																					},
																				},
																			},
																			"page_number": {
																				Type:     schema.TypeString,
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
										"deblock_filter": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.InputDeblockFilter](),
										},
										"denoise_filter": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.InputDenoiseFilter](),
										},
										"filter_strength": {
											Type:             schema.TypeInt,
											Optional:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 5)),
										},
										"input_filter": {
											Type:             schema.TypeString,
											Optional:         true,
											Computed:         true,
											ValidateDiagFunc: enum.Validate[types.InputFilter](),
										},
										"network_input_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"hls_input_settings": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"bandwidth": {
																	Type:     schema.TypeInt,
																	Optional: true,
																},
																"buffer_segments": {
																	Type:     schema.TypeInt,
																	Optional: true,
																},
																"retries": {
																	Type:     schema.TypeInt,
																	Optional: true,
																},
																"retry_interval": {
																	Type:     schema.TypeInt,
																	Optional: true,
																},
																"scte35_source": {
																	Type:             schema.TypeString,
																	Optional:         true,
																	ValidateDiagFunc: enum.Validate[types.HlsScte35SourceType](),
																},
															},
														},
													},
													"server_validation": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.NetworkInputServerValidation](),
													},
												},
											},
										},
										"scte35_pid": {
											Type:     schema.TypeInt,
											Optional: true,
										},
										"smpte2038_data_preference": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.Smpte2038DataPreference](),
										},
										"source_end_behavior": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.InputSourceEndBehavior](),
										},
										"video_selector": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"color_space": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.VideoSelectorColorSpace](),
													},
													// TODO implement color_space_settings
													"color_space_usage": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.VideoSelectorColorSpaceUsage](),
													},
													// TODO implement selector_settings
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"input_specification": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"codec": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.InputCodec](),
							},
							"maximum_bitrate": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.InputMaximumBitrate](),
							},
							"input_resolution": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.InputResolution](),
							},
						},
					},
				},
				"log_level": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.LogLevel](),
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
							"maintenance_start_time": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrRoleARN: {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
				},
				"start_channel": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"vpc": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					ForceNew: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrAvailabilityZones: {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"network_interface_ids": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"public_address_allocation_ids": {
								Type:     schema.TypeList,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							names.AttrSecurityGroupIDs: {
								Type:     schema.TypeSet,
								Optional: true,
								Computed: true,
								MaxItems: 5,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							names.AttrSubnetIDs: {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
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

const (
	ResNameChannel = "Channel"
)

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	in := &medialive.CreateChannelInput{
		Name:      aws.String(d.Get(names.AttrName).(string)),
		RequestId: aws.String(id.UniqueId()),
		Tags:      getTagsIn(ctx),
	}

	if v, ok := d.GetOk("cdi_input_specification"); ok && len(v.([]interface{})) > 0 {
		in.CdiInputSpecification = expandChannelCdiInputSpecification(v.([]interface{}))
	}
	if v, ok := d.GetOk("channel_class"); ok {
		in.ChannelClass = types.ChannelClass(v.(string))
	}
	if v, ok := d.GetOk("destinations"); ok && v.(*schema.Set).Len() > 0 {
		in.Destinations = expandChannelDestinations(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("encoder_settings"); ok && len(v.([]interface{})) > 0 {
		in.EncoderSettings = expandChannelEncoderSettings(v.([]interface{}))
	}
	if v, ok := d.GetOk("input_attachments"); ok && v.(*schema.Set).Len() > 0 {
		in.InputAttachments = expandChannelInputAttachments(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("input_specification"); ok && len(v.([]interface{})) > 0 {
		in.InputSpecification = expandChannelInputSpecification(v.([]interface{}))
	}
	if v, ok := d.GetOk("maintenance"); ok && len(v.([]interface{})) > 0 {
		in.Maintenance = expandChannelMaintenanceCreate(v.([]interface{}))
	}
	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		in.RoleArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("vpc"); ok && len(v.([]interface{})) > 0 {
		in.Vpc = expandChannelVPC(v.([]interface{}))
	}

	out, err := conn.CreateChannel(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionCreating, ResNameChannel, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.Channel == nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionCreating, ResNameChannel, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Channel.Id))

	if _, err := waitChannelCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionWaitingForCreation, ResNameChannel, d.Id(), err)
	}

	if d.Get("start_channel").(bool) {
		if err := startChannel(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionCreating, ResNameChannel, d.Get(names.AttrName).(string), err)
		}
	}

	return append(diags, resourceChannelRead(ctx, d, meta)...)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	out, err := FindChannelByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaLive Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionReading, ResNameChannel, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrName, out.Name)
	d.Set("channel_class", out.ChannelClass)
	d.Set("channel_id", out.Id)
	d.Set("log_level", out.LogLevel)
	d.Set(names.AttrRoleARN, out.RoleArn)

	if err := d.Set("cdi_input_specification", flattenChannelCdiInputSpecification(out.CdiInputSpecification)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("input_attachments", flattenChannelInputAttachments(out.InputAttachments)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("destinations", flattenChannelDestinations(out.Destinations)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("encoder_settings", flattenChannelEncoderSettings(out.EncoderSettings)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("input_specification", flattenChannelInputSpecification(out.InputSpecification)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("maintenance", flattenChannelMaintenance(out.Maintenance)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("vpc", flattenChannelVPC(out.Vpc)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}

	return diags
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "start_channel") {
		in := &medialive.UpdateChannelInput{
			ChannelId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			in.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("cdi_input_specification") {
			in.CdiInputSpecification = expandChannelCdiInputSpecification(d.Get("cdi_input_specification").([]interface{}))
		}

		if d.HasChange("destinations") {
			in.Destinations = expandChannelDestinations(d.Get("destinations").(*schema.Set).List())
		}

		if d.HasChange("encoder_settings") {
			in.EncoderSettings = expandChannelEncoderSettings(d.Get("encoder_settings").([]interface{}))
		}

		if d.HasChange("input_attachments") {
			in.InputAttachments = expandChannelInputAttachments(d.Get("input_attachments").(*schema.Set).List())
		}

		if d.HasChange("input_specification") {
			in.InputSpecification = expandChannelInputSpecification(d.Get("input_specification").([]interface{}))
		}

		if d.HasChange("log_level") {
			in.LogLevel = types.LogLevel(d.Get("log_level").(string))
		}

		if d.HasChange("maintenance") {
			in.Maintenance = expandChannelMaintenanceUpdate(d.Get("maintenance").([]interface{}))
		}

		if d.HasChange(names.AttrRoleARN) {
			in.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
		}

		channel, err := FindChannelByID(ctx, conn, d.Id())

		if err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameChannel, d.Id(), err)
		}

		if channel.State == types.ChannelStateRunning {
			if err := stopChannel(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
				return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameChannel, d.Id(), err)
			}
		}

		out, err := conn.UpdateChannel(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameChannel, d.Id(), err)
		}

		if _, err := waitChannelUpdated(ctx, conn, aws.ToString(out.Channel.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionWaitingForUpdate, ResNameChannel, d.Id(), err)
		}
	}

	if d.Get("start_channel").(bool) {
		if err := startChannel(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameChannel, d.Get(names.AttrName).(string), err)
		}
	}

	if d.HasChange("start_channel") {
		channel, err := FindChannelByID(ctx, conn, d.Id())

		if err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameChannel, d.Id(), err)
		}

		switch d.Get("start_channel").(bool) {
		case true:
			if channel.State == types.ChannelStateIdle {
				if err := startChannel(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
					return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameChannel, d.Id(), err)
				}
			}
		default:
			if channel.State == types.ChannelStateRunning {
				if err := stopChannel(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
					return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameChannel, d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceChannelRead(ctx, d, meta)...)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	log.Printf("[INFO] Deleting MediaLive Channel %s", d.Id())

	channel, err := FindChannelByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionDeleting, ResNameChannel, d.Id(), err)
	}

	if channel.State == types.ChannelStateRunning {
		if err := stopChannel(ctx, conn, d.Timeout(schema.TimeoutDelete), d.Id()); err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionDeleting, ResNameChannel, d.Id(), err)
		}
	}

	_, err = conn.DeleteChannel(ctx, &medialive.DeleteChannelInput{
		ChannelId: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionDeleting, ResNameChannel, d.Id(), err)
	}

	if _, err := waitChannelDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionWaitingForDeletion, ResNameChannel, d.Id(), err)
	}

	return diags
}

func startChannel(ctx context.Context, conn *medialive.Client, timeout time.Duration, id string) error {
	_, err := conn.StartChannel(ctx, &medialive.StartChannelInput{
		ChannelId: aws.String(id),
	})

	if err != nil {
		return fmt.Errorf("starting Medialive Channel (%s): %s", id, err)
	}

	_, err = waitChannelStarted(ctx, conn, id, timeout)

	if err != nil {
		return fmt.Errorf("waiting for Medialive Channel (%s) start: %s", id, err)
	}

	return nil
}

func stopChannel(ctx context.Context, conn *medialive.Client, timeout time.Duration, id string) error {
	_, err := conn.StopChannel(ctx, &medialive.StopChannelInput{
		ChannelId: aws.String(id),
	})

	if err != nil {
		return fmt.Errorf("stopping Medialive Channel (%s): %s", id, err)
	}

	_, err = waitChannelStopped(ctx, conn, id, timeout)

	if err != nil {
		return fmt.Errorf("waiting for Medialive Channel (%s) stop: %s", id, err)
	}

	return nil
}

func waitChannelCreated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeChannelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ChannelStateCreating),
		Target:                    enum.Slice(types.ChannelStateIdle),
		Refresh:                   statusChannel(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func waitChannelUpdated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeChannelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ChannelStateUpdating),
		Target:                    enum.Slice(types.ChannelStateIdle),
		Refresh:                   statusChannel(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func waitChannelDeleted(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeChannelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ChannelStateDeleting),
		Target:  []string{},
		Refresh: statusChannel(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func waitChannelStarted(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeChannelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ChannelStateStarting),
		Target:  enum.Slice(types.ChannelStateRunning),
		Refresh: statusChannel(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func waitChannelStopped(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeChannelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ChannelStateStopping),
		Target:  enum.Slice(types.ChannelStateIdle),
		Refresh: statusChannel(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func statusChannel(ctx context.Context, conn *medialive.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindChannelByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindChannelByID(ctx context.Context, conn *medialive.Client, id string) (*medialive.DescribeChannelOutput, error) {
	in := &medialive.DescribeChannelInput{
		ChannelId: aws.String(id),
	}
	out, err := conn.DescribeChannel(ctx, in)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	// Channel can still be found with a state of DELETED.
	// Set result as not found when the state is deleted.
	if out.State == types.ChannelStateDeleted {
		return nil, &retry.NotFoundError{
			LastResponse: string(types.ChannelStateDeleted),
			LastRequest:  in,
		}
	}

	return out, nil
}

func expandChannelInputAttachments(tfList []interface{}) []types.InputAttachment {
	var attachments []types.InputAttachment
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var a types.InputAttachment
		if v, ok := m["input_attachment_name"].(string); ok {
			a.InputAttachmentName = aws.String(v)
		}
		if v, ok := m["input_id"].(string); ok {
			a.InputId = aws.String(v)
		}
		if v, ok := m["input_settings"].([]interface{}); ok && len(v) > 0 {
			a.InputSettings = expandInputAttachmentInputSettings(v)
		}
		if v, ok := m["automatic_input_failover_settings"].([]interface{}); ok && len(v) > 0 {
			a.AutomaticInputFailoverSettings = expandInputAttachmentAutomaticInputFailoverSettings(v)
		}

		attachments = append(attachments, a)
	}

	return attachments
}

func expandInputAttachmentInputSettings(tfList []interface{}) *types.InputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.InputSettings
	if v, ok := m["audio_selector"].([]interface{}); ok && len(v) > 0 {
		out.AudioSelectors = expandInputAttachmentInputSettingsAudioSelectors(v)
	}
	if v, ok := m["caption_selector"].([]interface{}); ok && len(v) > 0 {
		out.CaptionSelectors = expandInputAttachmentInputSettingsCaptionSelectors(v)
	}
	if v, ok := m["deblock_filter"].(string); ok && v != "" {
		out.DeblockFilter = types.InputDeblockFilter(v)
	}
	if v, ok := m["denoise_filter"].(string); ok && v != "" {
		out.DenoiseFilter = types.InputDenoiseFilter(v)
	}
	if v, ok := m["filter_strength"].(int); ok && v != 0 {
		out.FilterStrength = aws.Int32(int32(v))
	}
	if v, ok := m["input_filter"].(string); ok && v != "" {
		out.InputFilter = types.InputFilter(v)
	}
	if v, ok := m["network_input_settings"].([]interface{}); ok && len(v) > 0 {
		out.NetworkInputSettings = expandInputAttachmentInputSettingsNetworkInputSettings(v)
	}
	if v, ok := m["scte35_pid"].(int); ok && v != 0 {
		out.Scte35Pid = aws.Int32(int32(v))
	}
	if v, ok := m["smpte2038_data_preference"].(string); ok && v != "" {
		out.Smpte2038DataPreference = types.Smpte2038DataPreference(v)
	}
	if v, ok := m["source_end_behavior"].(string); ok && v != "" {
		out.SourceEndBehavior = types.InputSourceEndBehavior(v)
	}

	return &out
}

func expandInputAttachmentInputSettingsAudioSelectors(tfList []interface{}) []types.AudioSelector {
	var as []types.AudioSelector
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var a types.AudioSelector
		if v, ok := m[names.AttrName].(string); ok && v != "" {
			a.Name = aws.String(v)
		}
		if v, ok := m["selector_settings"].([]interface{}); ok && len(v) > 0 {
			a.SelectorSettings = expandInputAttachmentInputSettingsAudioSelectorsSelectorSettings(v)
		}

		as = append(as, a)
	}

	return as
}

func expandInputAttachmentInputSettingsAudioSelectorsSelectorSettings(tfList []interface{}) *types.AudioSelectorSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioSelectorSettings
	if v, ok := m["audio_hls_rendition_selection"].([]interface{}); ok && len(v) > 0 {
		out.AudioHlsRenditionSelection = expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioHlsRenditionSelection(v)
	}
	if v, ok := m["audio_language_selection"].([]interface{}); ok && len(v) > 0 {
		out.AudioLanguageSelection = expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioLanguageSelection(v)
	}
	if v, ok := m["audio_pid_selection"].([]interface{}); ok && len(v) > 0 {
		out.AudioPidSelection = expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioPidSelection(v)
	}
	if v, ok := m["audio_track_selection"].([]interface{}); ok && len(v) > 0 {
		out.AudioTrackSelection = expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelection(v)
	}

	return &out
}

func expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioHlsRenditionSelection(tfList []interface{}) *types.AudioHlsRenditionSelection {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioHlsRenditionSelection
	if v, ok := m["group_id"].(string); ok && len(v) > 0 {
		out.GroupId = aws.String(v)
	}
	if v, ok := m[names.AttrName].(string); ok && len(v) > 0 {
		out.Name = aws.String(v)
	}

	return &out
}

func expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioLanguageSelection(tfList []interface{}) *types.AudioLanguageSelection {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioLanguageSelection
	if v, ok := m[names.AttrLanguageCode].(string); ok && len(v) > 0 {
		out.LanguageCode = aws.String(v)
	}
	if v, ok := m["language_selection_policy"].(string); ok && len(v) > 0 {
		out.LanguageSelectionPolicy = types.AudioLanguageSelectionPolicy(v)
	}

	return &out
}

func expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioPidSelection(tfList []interface{}) *types.AudioPidSelection {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioPidSelection
	if v, ok := m["pid"].(int); ok && v != 0 {
		out.Pid = aws.Int32(int32(v))
	}

	return &out
}

func expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelection(tfList []interface{}) *types.AudioTrackSelection {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioTrackSelection
	if v, ok := m["tracks"].(*schema.Set); ok && v.Len() > 0 {
		out.Tracks = expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelectionTracks(v.List())
	}
	if v, ok := m["dolby_e_decode"].([]interface{}); ok && len(v) > 0 {
		out.DolbyEDecode = expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelectionDolbyEDecode(v)
	}

	return &out
}

func expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelectionTracks(tfList []interface{}) []types.AudioTrack {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.AudioTrack
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.AudioTrack
		if v, ok := m["track"].(int); ok && v != 0 {
			o.Track = aws.Int32(int32(v))
		}

		out = append(out, o)
	}

	return out
}

func expandInputAttachmentInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelectionDolbyEDecode(tfList []interface{}) *types.AudioDolbyEDecode {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioDolbyEDecode
	if v, ok := m["program_selection"].(string); ok && v != "" {
		out.ProgramSelection = types.DolbyEProgramSelection(v)
	}

	return &out
}

func expandInputAttachmentInputSettingsCaptionSelectors(tfList []interface{}) []types.CaptionSelector {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.CaptionSelector
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.CaptionSelector
		if v, ok := m[names.AttrName].(string); ok && v != "" {
			o.Name = aws.String(v)
		}
		if v, ok := m[names.AttrLanguageCode].(string); ok && v != "" {
			o.LanguageCode = aws.String(v)
		}
		if v, ok := m["selector_settings"].([]interface{}); ok && len(v) > 0 {
			o.SelectorSettings = expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettings(v)
		}

		out = append(out, o)
	}

	return out
}

func expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettings(tfList []interface{}) *types.CaptionSelectorSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.CaptionSelectorSettings
	if v, ok := m["ancillary_source_settings"].([]interface{}); ok && len(v) > 0 {
		out.AncillarySourceSettings = expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsAncillarySourceSettings(v)
	}
	if v, ok := m["arib_source_settings"].([]interface{}); ok && len(v) > 0 {
		out.AribSourceSettings = &types.AribSourceSettings{} // no exported fields
	}
	if v, ok := m["dvb_sub_source_settings"].([]interface{}); ok && len(v) > 0 {
		out.DvbSubSourceSettings = expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsDvbSubSourceSettings(v)
	}
	if v, ok := m["embedded_source_settings"].([]interface{}); ok && len(v) > 0 {
		out.EmbeddedSourceSettings = expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsEmbeddedSourceSettings(v)
	}
	if v, ok := m["scte20_source_settings"].([]interface{}); ok && len(v) > 0 {
		out.Scte20SourceSettings = expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsScte20SourceSettings(v)
	}
	if v, ok := m["scte27_source_settings"].([]interface{}); ok && len(v) > 0 {
		out.Scte27SourceSettings = expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsScte27SourceSettings(v)
	}
	if v, ok := m["teletext_source_settings"].([]interface{}); ok && len(v) > 0 {
		out.TeletextSourceSettings = expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsTeletextSourceSettings(v)
	}

	return &out
}

func expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsAncillarySourceSettings(tfList []interface{}) *types.AncillarySourceSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AncillarySourceSettings
	if v, ok := m["source_ancillary_channel_number"].(int); ok && v != 0 {
		out.SourceAncillaryChannelNumber = aws.Int32(int32(v))
	}

	return &out
}

func expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsDvbSubSourceSettings(tfList []interface{}) *types.DvbSubSourceSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.DvbSubSourceSettings
	if v, ok := m["ocr_language"].(string); ok && v != "" {
		out.OcrLanguage = types.DvbSubOcrLanguage(v)
	}
	if v, ok := m["pid"].(int); ok {
		out.Pid = aws.Int32(int32(v))
	}

	return &out
}

func expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsEmbeddedSourceSettings(tfList []interface{}) *types.EmbeddedSourceSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.EmbeddedSourceSettings
	if v, ok := m["convert_608_to_708"].(string); ok && v != "" {
		out.Convert608To708 = types.EmbeddedConvert608To708(v)
	}
	if v, ok := m["scte20_detection"].(string); ok && v != "" {
		out.Scte20Detection = types.EmbeddedScte20Detection(v)
	}
	if v, ok := m["source_608_channel_number"].(int); ok && v != 0 {
		out.Source608ChannelNumber = aws.Int32(int32(v))
	}

	return &out
}

func expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsScte20SourceSettings(tfList []interface{}) *types.Scte20SourceSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.Scte20SourceSettings
	if v, ok := m["convert_608_to_708"].(string); ok && v != "" {
		out.Convert608To708 = types.Scte20Convert608To708(v)
	}
	if v, ok := m["source_608_channel_number"].(int); ok && v != 0 {
		out.Source608ChannelNumber = aws.Int32(int32(v))
	}

	return &out
}

func expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsScte27SourceSettings(tfList []interface{}) *types.Scte27SourceSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.Scte27SourceSettings
	if v, ok := m["ocr_language"].(string); ok && v != "" {
		out.OcrLanguage = types.Scte27OcrLanguage(v)
	}
	if v, ok := m["pid"].(int); ok && v != 0 {
		out.Pid = aws.Int32(int32(v))
	}

	return &out
}

func expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsTeletextSourceSettings(tfList []interface{}) *types.TeletextSourceSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.TeletextSourceSettings
	if v, ok := m["output_rectangle"].([]interface{}); ok && len(v) > 0 {
		out.OutputRectangle = expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsTeletextSourceSettingsOutputRectangle(v)
	}
	if v, ok := m["page_number"].(string); ok && v != "" {
		out.PageNumber = aws.String(v)
	}

	return &out
}

func expandInputAttachmentInputSettingsCaptionSelectorsSelectorSettingsTeletextSourceSettingsOutputRectangle(tfList []interface{}) *types.CaptionRectangle {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.CaptionRectangle
	if v, ok := m["height"].(float32); ok && v != 0.0 {
		out.Height = aws.Float64(float64(v))
	}
	if v, ok := m["left_offset"].(float32); ok && v != 0.0 {
		out.LeftOffset = aws.Float64(float64(v))
	}
	if v, ok := m["top_offset"].(float32); ok && v != 0.0 {
		out.TopOffset = aws.Float64(float64(v))
	}
	if v, ok := m["width"].(float32); ok && v != 0.0 {
		out.Width = aws.Float64(float64(v))
	}

	return &out
}

func expandInputAttachmentInputSettingsNetworkInputSettings(tfList []interface{}) *types.NetworkInputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.NetworkInputSettings
	if v, ok := m["hls_input_settings"].([]interface{}); ok && len(v) > 0 {
		out.HlsInputSettings = expandNetworkInputSettingsHLSInputSettings(v)
	}
	if v, ok := m["server_validation"].(string); ok && v != "" {
		out.ServerValidation = types.NetworkInputServerValidation(v)
	}

	return &out
}

func expandNetworkInputSettingsHLSInputSettings(tfList []interface{}) *types.HlsInputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsInputSettings
	if v, ok := m["bandwidth"].(int); ok && v != 0 {
		out.Bandwidth = aws.Int32(int32(v))
	}
	if v, ok := m["buffer_segments"].(int); ok && v != 0 {
		out.BufferSegments = aws.Int32(int32(v))
	}
	if v, ok := m["retries"].(int); ok && v != 0 {
		out.Retries = aws.Int32(int32(v))
	}
	if v, ok := m["retry_interval"].(int); ok && v != 0 {
		out.RetryInterval = aws.Int32(int32(v))
	}
	if v, ok := m["scte35_source"].(string); ok && v != "" {
		out.Scte35Source = types.HlsScte35SourceType(v)
	}

	return &out
}

func expandInputAttachmentAutomaticInputFailoverSettings(tfList []interface{}) *types.AutomaticInputFailoverSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AutomaticInputFailoverSettings
	if v, ok := m["secondary_input_id"].(string); ok && v != "" {
		out.SecondaryInputId = aws.String(v)
	}
	if v, ok := m["error_clear_time_msec"].(int); ok && v != 0 {
		out.ErrorClearTimeMsec = aws.Int32(int32(v))
	}
	if v, ok := m["failover_condition"].(*schema.Set); ok && v.Len() > 0 {
		out.FailoverConditions = expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditions(v.List())
	}
	if v, ok := m["input_preference"].(string); ok && v != "" {
		out.InputPreference = types.InputPreference(v)
	}

	return &out
}

func expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditions(tfList []interface{}) []types.FailoverCondition {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.FailoverCondition
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.FailoverCondition
		if v, ok := m["failover_condition_settings"].([]interface{}); ok && len(v) > 0 {
			o.FailoverConditionSettings = expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettings(v)
		}

		out = append(out, o)
	}

	return out
}

func expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettings(tfList []interface{}) *types.FailoverConditionSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.FailoverConditionSettings
	if v, ok := m["audio_silence_settings"].([]interface{}); ok && len(v) > 0 {
		out.AudioSilenceSettings = expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsAudioSilenceSettings(v)
	}
	if v, ok := m["input_loss_settings"].([]interface{}); ok && len(v) > 0 {
		out.InputLossSettings = expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsInputLossSettings(v)
	}
	if v, ok := m["video_black_settings"].([]interface{}); ok && len(v) > 0 {
		out.VideoBlackSettings = expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsVideoBlackSettings(v)
	}

	return &out
}

func expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsAudioSilenceSettings(tfList []interface{}) *types.AudioSilenceFailoverSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioSilenceFailoverSettings
	if v, ok := m["audio_selector_name"].(string); ok && v != "" {
		out.AudioSelectorName = aws.String(v)
	}
	if v, ok := m["audio_silence_threshold_msec"].(int); ok && v != 0 {
		out.AudioSilenceThresholdMsec = aws.Int32(int32(v))
	}

	return &out
}

func expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsInputLossSettings(tfList []interface{}) *types.InputLossFailoverSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.InputLossFailoverSettings
	if v, ok := m["input_loss_threshold_msec"].(int); ok && v != 0 {
		out.InputLossThresholdMsec = aws.Int32(int32(v))
	}

	return &out
}

func expandInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsVideoBlackSettings(tfList []interface{}) *types.VideoBlackFailoverSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.VideoBlackFailoverSettings
	if v, ok := m["black_detect_threshold"].(float32); ok && v != 0.0 {
		out.BlackDetectThreshold = aws.Float64(float64(v))
	}
	if v, ok := m["video_black_threshold_msec"].(int); ok && v != 0 {
		out.VideoBlackThresholdMsec = aws.Int32(int32(v))
	}

	return &out
}

func flattenChannelInputAttachments(tfList []types.InputAttachment) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"input_id":                          aws.ToString(item.InputId),
			"input_attachment_name":             aws.ToString(item.InputAttachmentName),
			"input_settings":                    flattenInputAttachmentsInputSettings(item.InputSettings),
			"automatic_input_failover_settings": flattenInputAttachmentAutomaticInputFailoverSettings(item.AutomaticInputFailoverSettings),
		}

		out = append(out, m)
	}

	return out
}

func flattenInputAttachmentsInputSettings(in *types.InputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_selector":            flattenInputAttachmentsInputSettingsAudioSelectors(in.AudioSelectors),
		"caption_selector":          flattenInputAttachmentsInputSettingsCaptionSelectors(in.CaptionSelectors),
		"deblock_filter":            string(in.DeblockFilter),
		"denoise_filter":            string(in.DenoiseFilter),
		"filter_strength":           int(aws.ToInt32(in.FilterStrength)),
		"input_filter":              string(in.InputFilter),
		"network_input_settings":    flattenInputAttachmentsInputSettingsNetworkInputSettings(in.NetworkInputSettings),
		"scte35_pid":                int(aws.ToInt32(in.Scte35Pid)),
		"smpte2038_data_preference": string(in.Smpte2038DataPreference),
		"source_end_behavior":       string(in.SourceEndBehavior),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsAudioSelectors(tfList []types.AudioSelector) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, v := range tfList {
		m := map[string]interface{}{
			names.AttrName:      aws.ToString(v.Name),
			"selector_settings": flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettings(v.SelectorSettings),
		}

		out = append(out, m)
	}

	return out
}

func flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettings(in *types.AudioSelectorSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_hls_rendition_selection": flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioHlsRenditionSelection(in.AudioHlsRenditionSelection),
		"audio_language_selection":      flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioLanguageSelection(in.AudioLanguageSelection),
		"audio_pid_selection":           flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioPidSelection(in.AudioPidSelection),
		"audio_track_selection":         flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelection(in.AudioTrackSelection),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioHlsRenditionSelection(in *types.AudioHlsRenditionSelection) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"group_id":     aws.ToString(in.GroupId),
		names.AttrName: aws.ToString(in.Name),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioLanguageSelection(in *types.AudioLanguageSelection) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrLanguageCode:      aws.ToString(in.LanguageCode),
		"language_selection_policy": string(in.LanguageSelectionPolicy),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioPidSelection(in *types.AudioPidSelection) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"pid": int(aws.ToInt32(in.Pid)),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelection(in *types.AudioTrackSelection) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"dolby_e_decode": flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelectionDolbyEDecode(in.DolbyEDecode),
		"tracks":         flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelectionTracks(in.Tracks),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelectionDolbyEDecode(in *types.AudioDolbyEDecode) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"program_selection": string(in.ProgramSelection),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsAudioSelectorsSelectorSettingsAudioTrackSelectionTracks(tfList []types.AudioTrack) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, v := range tfList {
		m := map[string]interface{}{
			"track": int(aws.ToInt32(v.Track)),
		}

		out = append(out, m)
	}

	return out
}

func flattenInputAttachmentsInputSettingsCaptionSelectors(tfList []types.CaptionSelector) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, v := range tfList {
		m := map[string]interface{}{
			names.AttrName:         aws.ToString(v.Name),
			names.AttrLanguageCode: aws.ToString(v.LanguageCode),
			"selector_settings":    flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettings(v.SelectorSettings),
		}

		out = append(out, m)
	}

	return out
}

func flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettings(in *types.CaptionSelectorSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"ancillary_source_settings": flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsAncillarySourceSettings(in.AncillarySourceSettings),
		"arib_source_settings":      []interface{}{}, // attribute has no exported fields
		"dvb_sub_source_settings":   flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsDvbSubSourceSettings(in.DvbSubSourceSettings),
		"embedded_source_settings":  flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsEmbeddedSourceSettings(in.EmbeddedSourceSettings),
		"scte20_source_settings":    flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsScte20SourceSettings(in.Scte20SourceSettings),
		"scte27_source_settings":    flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsScte27SourceSettings(in.Scte27SourceSettings),
		"teletext_source_settings":  flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsTeletextSourceSettings(in.TeletextSourceSettings),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsAncillarySourceSettings(in *types.AncillarySourceSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"source_ancillary_channel_number": int(aws.ToInt32(in.SourceAncillaryChannelNumber)),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsDvbSubSourceSettings(in *types.DvbSubSourceSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"ocr_language": string(in.OcrLanguage),
		"pid":          int(aws.ToInt32(in.Pid)),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsEmbeddedSourceSettings(in *types.EmbeddedSourceSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"convert_608_to_708":        string(in.Convert608To708),
		"scte20_detection":          string(in.Scte20Detection),
		"source_608_channel_number": int(aws.ToInt32(in.Source608ChannelNumber)),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsScte20SourceSettings(in *types.Scte20SourceSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"convert_608_to_708":        string(in.Convert608To708),
		"source_608_channel_number": int(aws.ToInt32(in.Source608ChannelNumber)),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsScte27SourceSettings(in *types.Scte27SourceSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"ocr_language": string(in.OcrLanguage),
		"pid":          int(aws.ToInt32(in.Pid)),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsTeletextSourceSettings(in *types.TeletextSourceSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"output_rectangle": flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsTeletextSourceSettingsOutputRectangle(in.OutputRectangle),
		"page_number":      aws.ToString(in.PageNumber),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsCaptionSelectorsSelectorSettingsTeletextSourceSettingsOutputRectangle(in *types.CaptionRectangle) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"height":      float32(aws.ToFloat64(in.Height)),
		"left_offset": float32(aws.ToFloat64(in.LeftOffset)),
		"top_offset":  float32(aws.ToFloat64(in.TopOffset)),
		"width":       float32(aws.ToFloat64(in.Width)),
	}

	return []interface{}{m}
}

func flattenInputAttachmentsInputSettingsNetworkInputSettings(in *types.NetworkInputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"hls_input_settings": flattenNetworkInputSettingsHLSInputSettings(in.HlsInputSettings),
		"server_validation":  string(in.ServerValidation),
	}

	return []interface{}{m}
}

func flattenNetworkInputSettingsHLSInputSettings(in *types.HlsInputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"bandwidth":       int(aws.ToInt32(in.Bandwidth)),
		"buffer_segments": int(aws.ToInt32(in.BufferSegments)),
		"retries":         int(aws.ToInt32(in.Retries)),
		"retry_interval":  int(aws.ToInt32(in.RetryInterval)),
		"scte35_source":   string(in.Scte35Source),
	}

	return []interface{}{m}
}

func flattenInputAttachmentAutomaticInputFailoverSettings(in *types.AutomaticInputFailoverSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"secondary_input_id":    aws.ToString(in.SecondaryInputId),
		"error_clear_time_msec": int(aws.ToInt32(in.ErrorClearTimeMsec)),
		"failover_condition":    flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditions(in.FailoverConditions),
		"input_preference":      string(in.InputPreference),
	}

	return []interface{}{m}
}

func flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditions(tfList []types.FailoverCondition) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"failover_condition_settings": flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettings(item.FailoverConditionSettings),
		}

		out = append(out, m)
	}
	return out
}

func flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettings(in *types.FailoverConditionSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_silence_settings": flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsAudioSilenceSettings(in.AudioSilenceSettings),
		"input_loss_settings":    flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsInputLossSettings(in.InputLossSettings),
		"video_black_settings":   flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsVideoBlackSettings(in.VideoBlackSettings),
	}

	return []interface{}{m}
}

func flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsAudioSilenceSettings(in *types.AudioSilenceFailoverSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_selector_name":          aws.ToString(in.AudioSelectorName),
		"audio_silence_threshold_msec": int(aws.ToInt32(in.AudioSilenceThresholdMsec)),
	}

	return []interface{}{m}
}

func flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsInputLossSettings(in *types.InputLossFailoverSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"input_loss_threshold_msec": int(aws.ToInt32(in.InputLossThresholdMsec)),
	}

	return []interface{}{m}
}

func flattenInputAttachmentAutomaticInputFailoverSettingsFailoverConditionsFailoverConditionSettingsVideoBlackSettings(in *types.VideoBlackFailoverSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"black_detect_threshold":     float32(aws.ToFloat64(in.BlackDetectThreshold)),
		"video_black_threshold_msec": int(aws.ToInt32(in.VideoBlackThresholdMsec)),
	}

	return []interface{}{m}
}

func expandChannelCdiInputSpecification(tfList []interface{}) *types.CdiInputSpecification {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	spec := &types.CdiInputSpecification{}
	if v, ok := m["resolution"].(string); ok && v != "" {
		spec.Resolution = types.CdiInputResolution(v)
	}

	return spec
}

func flattenChannelCdiInputSpecification(apiObject *types.CdiInputSpecification) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"resolution": string(apiObject.Resolution),
	}

	return []interface{}{m}
}

func expandChannelDestinations(tfList []interface{}) []types.OutputDestination {
	if tfList == nil {
		return nil
	}

	var destinations []types.OutputDestination
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var d types.OutputDestination
		if v, ok := m[names.AttrID].(string); ok {
			d.Id = aws.String(v)
		}
		if v, ok := m["media_package_settings"].(*schema.Set); ok && v.Len() > 0 {
			d.MediaPackageSettings = expandChannelDestinationsMediaPackageSettings(v.List())
		}
		if v, ok := m["multiplex_settings"].([]interface{}); ok && len(v) > 0 {
			d.MultiplexSettings = expandChannelDestinationsMultiplexSettings(v)
		}
		if v, ok := m["settings"].(*schema.Set); ok && v.Len() > 0 {
			d.Settings = expandChannelDestinationsSettings(v.List())
		}

		destinations = append(destinations, d)
	}

	return destinations
}

func expandChannelDestinationsMediaPackageSettings(tfList []interface{}) []types.MediaPackageOutputDestinationSettings {
	if tfList == nil {
		return nil
	}

	var settings []types.MediaPackageOutputDestinationSettings
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var s types.MediaPackageOutputDestinationSettings
		if v, ok := m["channel_id"].(string); ok {
			s.ChannelId = aws.String(v)
		}

		settings = append(settings, s)
	}

	return settings
}

func expandChannelDestinationsMultiplexSettings(tfList []interface{}) *types.MultiplexProgramChannelDestinationSettings {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.MultiplexProgramChannelDestinationSettings{}
	if v, ok := m["multiplex_id"].(string); ok && v != "" {
		settings.MultiplexId = aws.String(v)
	}
	if v, ok := m["program_name"].(string); ok && v != "" {
		settings.ProgramName = aws.String(v)
	}

	return settings
}

func expandChannelDestinationsSettings(tfList []interface{}) []types.OutputDestinationSettings {
	if tfList == nil {
		return nil
	}

	var settings []types.OutputDestinationSettings
	for _, v := range tfList {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var s types.OutputDestinationSettings
		if v, ok := m["password_param"].(string); ok {
			s.PasswordParam = aws.String(v)
		}
		if v, ok := m["stream_name"].(string); ok {
			s.StreamName = aws.String(v)
		}
		if v, ok := m[names.AttrURL].(string); ok {
			s.Url = aws.String(v)
		}
		if v, ok := m[names.AttrUsername].(string); ok {
			s.Username = aws.String(v)
		}

		settings = append(settings, s)
	}

	return settings
}

func flattenChannelDestinations(apiObject []types.OutputDestination) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}
	for _, v := range apiObject {
		m := map[string]interface{}{
			names.AttrID:             aws.ToString(v.Id),
			"media_package_settings": flattenChannelDestinationsMediaPackageSettings(v.MediaPackageSettings),
			"multiplex_settings":     flattenChannelDestinationsMultiplexSettings(v.MultiplexSettings),
			"settings":               flattenChannelDestinationsSettings(v.Settings),
		}

		tfList = append(tfList, m)
	}

	return tfList
}

func flattenChannelDestinationsMediaPackageSettings(apiObject []types.MediaPackageOutputDestinationSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}
	for _, v := range apiObject {
		m := map[string]interface{}{
			"channel_id": aws.ToString(v.ChannelId),
		}

		tfList = append(tfList, m)
	}

	return tfList
}

func flattenChannelDestinationsMultiplexSettings(apiObject *types.MultiplexProgramChannelDestinationSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"multiplex_id": aws.ToString(apiObject.MultiplexId),
		"program_name": aws.ToString(apiObject.ProgramName),
	}

	return []interface{}{m}
}

func flattenChannelDestinationsSettings(apiObject []types.OutputDestinationSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}
	for _, v := range apiObject {
		m := map[string]interface{}{
			"password_param":   aws.ToString(v.PasswordParam),
			"stream_name":      aws.ToString(v.StreamName),
			names.AttrURL:      aws.ToString(v.Url),
			names.AttrUsername: aws.ToString(v.Username),
		}

		tfList = append(tfList, m)
	}

	return tfList
}

func expandChannelInputSpecification(tfList []interface{}) *types.InputSpecification {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	spec := &types.InputSpecification{}
	if v, ok := m["codec"].(string); ok && v != "" {
		spec.Codec = types.InputCodec(v)
	}
	if v, ok := m["maximum_bitrate"].(string); ok && v != "" {
		spec.MaximumBitrate = types.InputMaximumBitrate(v)
	}
	if v, ok := m["input_resolution"].(string); ok && v != "" {
		spec.Resolution = types.InputResolution(v)
	}

	return spec
}

func flattenChannelInputSpecification(apiObject *types.InputSpecification) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"codec":            string(apiObject.Codec),
		"maximum_bitrate":  string(apiObject.MaximumBitrate),
		"input_resolution": string(apiObject.Resolution),
	}

	return []interface{}{m}
}

func expandChannelMaintenanceCreate(tfList []interface{}) *types.MaintenanceCreateSettings {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.MaintenanceCreateSettings{}
	if v, ok := m["maintenance_day"].(string); ok && v != "" {
		settings.MaintenanceDay = types.MaintenanceDay(v)
	}
	if v, ok := m["maintenance_start_time"].(string); ok && v != "" {
		settings.MaintenanceStartTime = aws.String(v)
	}

	return settings
}

func expandChannelMaintenanceUpdate(tfList []interface{}) *types.MaintenanceUpdateSettings {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.MaintenanceUpdateSettings{}
	if v, ok := m["maintenance_day"].(string); ok && v != "" {
		settings.MaintenanceDay = types.MaintenanceDay(v)
	}
	if v, ok := m["maintenance_start_time"].(string); ok && v != "" {
		settings.MaintenanceStartTime = aws.String(v)
	}
	// NOTE: This field is only available in the update struct. To allow users to set a scheduled
	// date on update, it may be worth adding to the base schema.
	// if v, ok := m["maintenance_scheduled_date"].(string); ok && v != "" {
	// 	settings.MaintenanceScheduledDate = aws.String(v)
	// }

	return settings
}

func flattenChannelMaintenance(apiObject *types.MaintenanceStatus) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"maintenance_day":        string(apiObject.MaintenanceDay),
		"maintenance_start_time": aws.ToString(apiObject.MaintenanceStartTime),
	}

	return []interface{}{m}
}

func expandChannelVPC(tfList []interface{}) *types.VpcOutputSettings {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.VpcOutputSettings{}
	if v, ok := m[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		settings.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}
	if v, ok := m[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		settings.SubnetIds = flex.ExpandStringValueSet(v)
	}
	if v, ok := m["public_address_allocation_ids"].(*schema.Set); ok && v.Len() > 0 {
		settings.PublicAddressAllocationIds = flex.ExpandStringValueSet(v)
	}

	return settings
}

func flattenChannelVPC(apiObject *types.VpcOutputSettingsDescription) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrAvailabilityZones: flex.FlattenStringValueSet(apiObject.AvailabilityZones),
		"network_interface_ids":     flex.FlattenStringValueSet(apiObject.NetworkInterfaceIds),
		names.AttrSecurityGroupIDs:  flex.FlattenStringValueSet(apiObject.SecurityGroupIds),
		names.AttrSubnetIDs:         flex.FlattenStringValueSet(apiObject.SubnetIds),
		// public_address_allocation_ids is not included in the output struct
	}

	return []interface{}{m}
}
