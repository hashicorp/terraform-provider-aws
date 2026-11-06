// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconvert

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// sdk object : https://github.com/aws/aws-sdk-go-v2/blob/main/service/mediaconvert/types/types.go
// types https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/mediaconvert@v1.53.1/types

// @SDKResource("aws_media_convert_job_template", name="Job Template")
// @Tags(identifierAttribute="arn")
func resourceJobTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceJobTemplateCreate,
		ReadWithoutTimeout:   resourceJobTemplateRead,
		UpdateWithoutTimeout: resourceJobTemplateUpdate,
		DeleteWithoutTimeout: resourceJobTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"acceleration_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Schema: map[string]*schema.Schema{
					"mode": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: enum.Validate[types.AccelerationMode](),
					},
				},
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hop_destination": {
				Type:     schema.TypeList,
				Optional: true,
				Schema: map[string]*schema.Schema{
					"priority": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"queue": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"wait_minutes": {
						Type:     schema.TypeInt,
						Required: true,
					},
				},
			},
			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"queue": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"settings": { // or "job_template_settings" ?
				Type:     schema.TypeList,
				Computed: true,
				Required: true,
				MaxItems: 1,
				Schema: map[string]*schema.Schema{
					"ad_avail_offset": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"avail_blanking": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Schema: map[string]*schema.Schema{
							"AvailBlankingImage": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
					"color_conversion_3D_LUT_settings": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Schema: map[string]*schema.Schema{
							"file_input": {
								Type:     schema.TypeString,
								Required: true,
							},
							"input_color_space": {
								Type:             schema.TypeList,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.ColorSpace](),
							},
							"input_mastering_luminance": {
								Type:     schema.TypeInt,
								Required: true,
							},
							"output_color_space": {
								Type:             schema.TypeList,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.ColorSpace](),
							},
							"output_mastering_luminance": {
								Type:     schema.TypeInt,
								Required: true,
							},
						},
					},
					"esam": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Schema: map[string]*schema.Schema{
							"manifest_confirm_condition_notification": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Schema: map[string]*schema.Schema{
									"mcc_xml": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							"response_signal_preroll": {
								Type:     schema.TypeInt,
								Required: true,
							},
							"signal_processing_notification": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Schema: map[string]*schema.Schema{
									"scc_xml": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
					"extended_data_services": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Schema: map[string]*schema.Schema{
							"copy_protection_action": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.CopyProtectionAction](),
							},
							"vchip_action": {
								Type:             schema.TypeList,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.VchipAction](),
							},
						},
					},
					"follow_source": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"inputs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Schema: map[string]*schema.Schema{
							"advanced_input_filter": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.AdvancedInputFilter](),
							},
							"advanced_input_filter_settings": {
								Type:     schema.TypeString,
								Optional: true,
								Schema: map[string]*schema.Schema{
									"crop": {
										Type:     schema.TypeString,
										Optional: true,
										MaxItems: 1,
										Schema: map[string]*schema.Schema{
											"add_texture": {
												Type:             schema.TypeString,
												Optional:         true,
												ValidateDiagFunc: enum.Validate[types.AdvancedInputFilterAddTexture](),
											},
											"sharpening": {
												Type:             schema.TypeString,
												Optional:         true,
												ValidateDiagFunc: enum.Validate[types.AdvancedInputFilterSharpen](),
											},
										},
									},
								},
							},
							"audio_selector_group": {
								Type:     schema.TypeMap,
								Optional: true,
								Schema: map[string]*schema.Schema{
									"audio_selector_names": {
										Type:     schema.TypeList,
										Optional: true,
									},
								},
							},
							"audio_selectors": {
								Type:     schema.TypeMap,
								Optional: true,
								Schema: map[string]*schema.Schema{
									"audio_duration_correction": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.AudioDurationCorrection](),
									},
									"custom_language_code": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"default_selection": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.AudioDefaultSelection](),
									},
									"external_audio_file_input": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"hls_rendition_group_settings": {
										Type:     schema.TypeMap,
										Optional: true,
										Schema: map[string]*schema.Schema{
											"rendition_group_id": {
												Type:     schema.TypeString,
												Optional: true,
											},
											"rendition_language_code": {
												Type:             schema.TypeInt,
												Optional:         true,
												ValidateDiagFunc: enum.Validate[types.LanguageCode](),
											},
											"rendition_name": {
												Type:     schema.TypeString,
												Optional: true,
											},
										},
									},
									"language_code": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.LanguageCode](),
									},
									"offset": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"pids": {
										Type:     schema.TypeList[TypeInt],
										Optional: true,
									},
									"program_selection": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"remix_settings": {
										Type:     schema.TypeInt,
										Optional: true,
										Schema: map[string]*schema.Schema{
											"audio_description_audio_channel": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"audio_description_data_channel": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"channel_mapping": {
												Type:     schema.TypeInt,
												Optional: true,
												Schema: map[string]*schema.Schema{
													"output_channels": {
														Type:     schema.TypeList,
														Optional: true,
														Schema: map[string]*schema.Schema{
															"input_channels": {
																Type:     schema.TypeList[TypeInt],
																Optional: true,
															},
															"input_channels_fine_tune": {
																Type:     schema.TypeList[TypeFloat],
																Optional: true,
															},
														},
													},
													"source_608_channel_number": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"source_608-track_number": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"terminateCaptions": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.EmbeddedTerminateCaptions](),
													},
												},
											},
											"channels_in": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"channels_out": {
												Type:     schema.TypeInt,
												Optional: true,
											},
										},
									},
									"selector_type": {
										Type:             schema.TypeInt,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.AudioSelectorType](),
									},
									"tracks": {
										Type:     schema.TypeList[TypeInt],
										Optional: true,
									},
								},
							},
							"caption_selectors": {
								Type:     schema.TypeMap,
								Optional: true,
								Schema: map[string]*schema.Schema{
									"custom_language_code": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"language_code": {
										Type:             schema.TypeInt,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.LanguageCode](),
									},
									"source_settings": {
										Type:     schema.TypeInt,
										Optional: true,
										Schema: map[string]*schema.Schema{
											"ancillary_source_settings": {
												Type:     schema.TypeString,
												Optional: true,
												Schema: map[string]*schema.Schema{
													"convert_608_to_708": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.AncillaryConvert608To708](),
													},
													"source_ancillary_channel_number": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"terminate_captions": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.AncillaryTerminateCaptions](),
													},
												},
											},
											"dvb_sub_source_settings": {
												Type:     schema.TypeInt,
												Optional: true,
												Schema: map[string]*schema.Schema{
													"Pid": {
														Type:     schema.TypeInt,
														Optional: true,
													}},
												"embedded_source_settings": {
													Type:     schema.TypeInt,
													Optional: true,
													Schema: map[string]*schema.Schema{
														"convert_608_to_708": {
															Type:             schema.TypeString,
															Optional:         true,
															ValidateDiagFunc: enum.Validate[types.EmbeddedConvert608To708](),
														},
														"_source_608_channel_number": {
															Type:     schema.TypeInt,
															Optional: true,
														},
														"source_608_track_number": {
															Type:     schema.TypeInt,
															Optional: true,
														},
														"terminate_captions": {
															Type:             schema.TypeString,
															Optional:         true,
															ValidateDiagFunc: enum.Validate[types.EmbeddedTerminateCaptions](),
														},
													},
												},
												"file_source_settings": {
													Type:     schema.TypeInt,
													Optional: true,
													Schema: map[string]*schema.Schema{
														"convert_608_to_708": {
															Type:             schema.TypeString,
															Optional:         true,
															ValidateDiagFunc: enum.Validate[types.FileSourceConvert608To708](),
														},
														"convert_paint_to_pop": {
															Type:             schema.TypeString,
															Optional:         true,
															ValidateDiagFunc: enum.Validate[types.CaptionSourceConvertPaintOnToPopOn](),
														},
														"framerate": {
															Type:     schema.TypeList,
															Optional: true,
															Schema: map[string]*schema.Schema{
																"framerate_denominator": {
																	Type:     schema.TypeInt,
																	Optional: true,
																},
																"framerate_numerator": {
																	Type:     schema.TypeInt,
																	Optional: true,
																},
															},
														},
														"source_file": {
															Type:     schema.TypeString,
															Optional: true,
														},
														"time_delta": {
															Type:     schema.TypeInt,
															Optional: true,
														},
														"time_delta_units": {
															Type:             schema.TypeString,
															Optional:         true,
															ValidateDiagFunc: enum.Validate[types.FileSourceTimeDeltaUnits](),
														},
													},
												},
												"source_type": {
													Type:             schema.TypeInt,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.CaptionSourceType](),
												},
												"teletext_source_settings": {
													Type:     schema.TypeInt,
													Optional: true,
													Schema: map[string]*schema.Schema{
														"page_number": {
															Type:     schema.TypeString,
															Optional: true,
														},
													},
												},
												"track_source_settings": {
													Type:     schema.TypeInt,
													Optional: true,
													Schema: map[string]*schema.Schema{
														"track_number": {
															Type:     schema.TypeString,
															Optional: true,
														},
													},
												},
												"webvtt_hls_source_settings": {
													Type:     schema.TypeInt,
													Optional: true,
													Schema: map[string]*schema.Schema{
														"rendition_group_id": {
															Type:     schema.TypeString,
															Optional: true,
														},
														"rendition_language_code": {
															Type:             schema.TypeString,
															Optional:         true,
															ValidateDiagFunc: enum.Validate[types.LanguageCode](),
														},
														"rendition_name": {
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
							"crop": {
								Type:     schema.TypeString,
								Optional: true,
								MaxItems: 1,
								Schema: map[string]*schema.Schema{
									"height": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"width": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"x": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"y": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
							"deblock_filter": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.InputDeblockFilter](),
							},
							"decryption_settings": {
								Type:     schema.TypeString,
								Optional: true,
								MaxItems: 1,
								Schema: map[string]*schema.Schema{
									"decryption_mode": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.DecryptionMode](),
									},
									"encrypted_decryption_key": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"initialization_vector": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"kms_key_region": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							"denoise_filter": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.InputDenoiseFilter](),
							},
							"dolby_vision_metadata_xml": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"file_input": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"filter_enable": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.InputFilterEnable](),
							},
							"filter_strength": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"image_inserter": {
								Type:     schema.TypeString,
								Optional: true,
								Schema: map[string]*schema.Schema{
									"insertable_images": {
										Type:     schema.TypeList,
										Optional: true,
										Schema: map[string]*schema.Schema{
											"duration": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"fade_in": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"fade_out": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"height": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"image_inserter_input": {
												Type:     schema.TypeString,
												Optional: true,
											},
											"image_x": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"image_y": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"layer": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"opacity": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"start_time": {
												Type:     schema.TypeString,
												Optional: true,
											},
											"width": {
												Type:     schema.TypeInt,
												Optional: true,
											},
										},
									},
									"sdr_reference_white_level": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
							"input_clippings": {
								Type:     schema.TypeString,
								Optional: true,
								Schema: map[string]*schema.Schema{
									"end_timecode": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"start_timecode": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							"input_scan_type": {
								Type:             schema.TypeList,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.InputScanType](),
							},
							"position": {
								Type:     schema.TypeString,
								Optional: true,
								MaxItems: 1,
								Schema: map[string]*schema.Schema{
									"height": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"width": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"x": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"y": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
							"program_number": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"psi_control": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.PsiControl](),
							},
							"timecode_source": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.InputTimecodeSource](),
							},
							"timecode_start": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"video_overlays": {
								Type:     schema.TypeList,
								Optional: true,
								Schema: map[string]*schema.Schema{
									"end_timecode": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"input": {
										Type:     schema.TypeString,
										Optional: true,
										Schema: map[string]*schema.Schema{
											"file_input": {
												Type:     schema.TypeString,
												Optional: true,
											},
											"input_clippings": {
												Type:     schema.TypeList,
												Optional: true,
												Schema: map[string]*schema.Schema{
													"end_timecode": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"start_timecode": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
											"timecode_source": {
												Type:             schema.TypeString,
												Optional:         true,
												ValidateDiagFunc: enum.Validate[types.InputTimecodeSource](),
											},
											"timecode_start": {
												Type:     schema.TypeString,
												Optional: true,
											},
										},
									},
									"start_timecode": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							"video_selector": {
								Type:     schema.TypeInt,
								Optional: true,
								MaxItems: 1,
								Schema: map[string]*schema.Schema{
									"alpha_behavior": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.AlphaBehavior](),
									},
									"color_space": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.ColorSpace](),
									},
									"color_space_usage": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.ColorSpaceUsage](),
									},
									"embedded_timecode_override": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.EmbeddedTimecodeOverride](),
									},
									"hdr10_metadata": {
										Type:     schema.TypeString,
										Optional: true,
										Schema: map[string]*schema.Schema{
											"blue_primary_x": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"blue_primary_y": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"green_primary_x": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"green_primary_y": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"MaxContentLightLevel": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"max_fram_average_light_level": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"min_luminance": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"red_primary_x": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"red_primary_y": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"white_point_x": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"white_point_y": {
												Type:     schema.TypeInt,
												Optional: true,
											},
										},
									},
									"max_luminance": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"pad_video": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.PadVideo](),
									},
									"pid": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"program_number": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"rotate": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.InputRotate](),
									},
									"sample_range": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.InputSampleRange](),
									},
								},
							},
						},
					},
					"kantar_watermark": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Schema: map[string]*schema.Schema{
							"channel_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"content_reference": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"credentials_secret_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"file_offset": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
							"kantar_license_id": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"kantar_server_url": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"log_destination": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"metadata_3": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"metadata_4": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"metadata_5": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"metadata_6": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"metadata_7": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"metadata_8": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
					"motion_image_inserter": {
						Type:     schema.TypeList,
						Optional: true,
					},
					"nielsen_configuration": {
						Type:     schema.TypeList,
						Optional: true,
						Schema: map[string]*schema.Schema{
							"breakout_code": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"distributor_id": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
					"nielsen_non_linear_watermark": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Schema: map[string]*schema.Schema{
							"active_watermark_process": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.NielsenActiveWatermarkProcessType](),
							},
							"adi_filename": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"asset_id": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"asset_name": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
							"cbe_source_id": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"episode_id": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"metadata_destination": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"source_id": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"source_watermark_status": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.NielsenSourceWatermarkStatusType](),
							},
							"tic_server_url": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"unique_tic_per_audio_track": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.NielsenUniqueTicPerAudioTrackType](),
							},
						},
					},
					"output_groups": {
						Type:     schema.TypeList,
						Required: true,
						Schema: map[string]*schema.Schema{
							"automated_encoding_settings": {
								Type:     schema.TypeString,
								Optional: true,
								Schema: map[string]*schema.Schema{
									"abr_settings": {
										Type:     schema.TypeString,
										Optional: true,
										Schema: map[string]*schema.Schema{
											"max_abr_bitrate": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"max_renditions": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"min_abr_bitrate": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"rules": {
												Type:     schema.TypeString,
												Optional: true,
												Schema: map[string]*schema.Schema{
													"allowed_renditions": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"force_include_renditions": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"min_bottom_rendition_size": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"min_top_rendition_size": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"type ": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
							"custom_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"output_group_settings": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
							"outputs": {
								Type:     schema.TypeList,
								Optional: true,
							},
						},
					},
					"timecode_config": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
					},
					"timed_metadata_insertion": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
					},
				},
			},
			"status_update_interval": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.StatusUpdateInterval](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceJobTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	name := d.Get("name").(string)
	input := &mediaconvert.CreateJobTemplateInput{
		Name:     aws.String(name),
		Settings: types.JobTemplateSettings(d.Get("settings").(string)),
		Tags:     getTagsIn(ctx),
		Queue:    aws.String(v.(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("category"); ok {
		input.Category = aws.String(v.(string))
	}

	if v, ok := d.Get("acceleration_settings").([]interface{}); ok && len(v) > 0 && v[0] != nil {
		input.AccelerationSettings = expandAccelerationSettings(v[0].(map[string]interface{}))
	}

	output, err := conn.CreateJobTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Media Convert JobTemplate (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.JobTemplate.Name))

	return append(diags, resourceJobTemplateRead(ctx, d, meta)...)
}

func resourceJobTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	job_template, err := findJobTemplateByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Media Convert JobTemplate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Media Convert JobTemplate (%s): %s", d.Id(), err)
	}

	d.Set("arn", job_template.Arn)
	d.Set("description", job_template.Description)
	d.Set("name", job_template.Name)
	d.Set("settings", job_template.Settings)
	if job_template.Acceleration_settings != nil {
		if err := d.Set("acceleration_settings", []interface{}{flattenAccelerationSettings(job_template.AccelerationSettings)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting acceleration_settings: %s", err)
		}
	} else {
		d.Set("acceleration_settings", nil)
	}
	d.Set("status", job_template.Status)

	return diags
}

func resourceJobTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &mediaconvert.UpdateJobTemplateInput{
			Name:   aws.String(d.Id()),
			Status: types.JobTemplateStatus(d.Get("status").(string)),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.Get("acceleration_settings").([]interface{}); ok && len(v) > 0 && v[0] != nil {
			input.AccelerationSettings = expandAccelerationSettings(v[0].(map[string]interface{}))
		}

		_, err := conn.UpdateJobTemplate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Media Convert JobTemplate (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceJobTemplateRead(ctx, d, meta)...)
}

func resourceJobTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	log.Printf("[DEBUG] Deleting Media Convert JobTemplate: %s", d.Id())
	_, err := conn.DeleteJobTemplate(ctx, &mediaconvert.DeleteJobTemplateInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Media Convert JobTemplate (%s): %s", d.Id(), err)
	}

	return diags
}

func findJobTemplateByName(ctx context.Context, conn *mediaconvert.Client, name string) (*types.JobTemplate, error) {
	input := &mediaconvert.GetJobTemplateInput{
		Name: aws.String(name),
	}

	output, err := conn.GetJobTemplate(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.JobTemplate == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.JobTemplate, nil
}

func expandAccelerationSettings(tfMap map[string]interface{}) *types.AccelerationSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AccelerationSettings{}

	if v, ok := tfMap["commitment"]; ok {
		apiObject.Commitment = types.Commitment(v.(string))
	}

	if v, ok := tfMap["renewal_type"]; ok {
		apiObject.RenewalType = types.RenewalType(v.(string))
	}

	if v, ok := tfMap["reserved_slots"]; ok {
		apiObject.ReservedSlots = aws.Int32(int32(v.(int)))
	}

	return apiObject
}

func flattenAccelerationSettings(apiObject *types.AccelerationSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"Mode":           apiObject.Commitment,
		"renewal_type":   apiObject.RenewalType,
		"reserved_slots": aws.ToInt32(apiObject.ReservedSlots),
	}

	return tfMap
}
