// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func channelEncoderSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"audio_descriptions": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"audio_selector_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"audio_normalization_settings": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"algorithm": {
											Type:             schema.TypeString,
											Optional:         true,
											Computed:         true,
											ValidateDiagFunc: enum.Validate[types.AudioNormalizationAlgorithm](),
										},
										"algorithm_control": {
											Type:             schema.TypeString,
											Optional:         true,
											Computed:         true,
											ValidateDiagFunc: enum.Validate[types.AudioNormalizationAlgorithmControl](),
										},
										"target_lkfs": {
											Type:     schema.TypeFloat,
											Optional: true,
											Computed: true,
										},
									},
								},
							},
							"audio_type": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.AudioType](),
							},
							"audio_type_control": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.AudioDescriptionAudioTypeControl](),
							},
							"audio_watermark_settings": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"nielsen_watermarks_settings": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"nielsen_cbet_settings": {
														Type:     schema.TypeList,
														Optional: true,
														Computed: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"cbet_check_digit_string": {
																	Type:     schema.TypeString,
																	Required: true,
																},
																"cbet_stepaside": {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[types.NielsenWatermarksCbetStepaside](),
																},
																"csid": {
																	Type:     schema.TypeString,
																	Required: true,
																},
															},
														},
													},
													"nielsen_distribution_type": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.NielsenWatermarksDistributionTypes](),
													},
													"nielsen_naes_ii_nw_settings": {
														Type:     schema.TypeList,
														Optional: true,
														Computed: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"check_digit_string": {
																	Type:     schema.TypeString,
																	Required: true,
																},
																"sid": {
																	Type:     schema.TypeFloat,
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
							"codec_settings": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"aac_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"bitrate": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"coding_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AacCodingMode](),
													},
													"input_type": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AacInputType](),
													},
													names.AttrProfile: {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AacProfile](),
													},
													"rate_control_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AacRateControlMode](),
													},
													"raw_format": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AacRawFormat](),
													},
													"sample_rate": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"spec": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AacSpec](),
													},
													"vbr_quality": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AacVbrQuality](),
													},
												},
											},
										},
										"ac3_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"bitrate": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"bitstream_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Ac3BitstreamMode](),
													},
													"coding_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Ac3CodingMode](),
													},
													"dialnorm": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"drc_profile": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Ac3DrcProfile](),
													},
													"lfe_filter": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Ac3LfeFilter](),
													},
													"metadata_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Ac3MetadataControl](),
													},
												},
											},
										},
										"eac3_atmos_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"bitrate": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"coding_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3AtmosCodingMode](),
													},
													"dialnorm": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"drc_line": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3AtmosDrcLine](),
													},
													"drc_rf": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3AtmosDrcRf](),
													},
													"height_trim": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"surround_trim": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
												},
											},
										},
										"eac3_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"attenuation_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3AttenuationControl](),
													},
													"bitrate": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"bitstream_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3BitstreamMode](),
													},
													"coding_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3CodingMode](),
													},
													"dc_filter": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3DcFilter](),
													},
													"dialnorm": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"drc_line": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3DrcLine](),
													},
													"drc_rf": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3DrcRf](),
													},
													"lfe_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3LfeControl](),
													},
													"lfe_filter": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3LfeFilter](),
													},
													"lo_ro_center_mix_level": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"lo_ro_surround_mix_level": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"lt_rt_center_mix_level": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"lt_rt_surround_mix_level": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"metadata_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3MetadataControl](),
													},
													"passthrough_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3PassthroughControl](),
													},
													"phase_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3PhaseControl](),
													},
													"stereo_downmix": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3StereoDownmix](),
													},
													"surround_ex_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3SurroundExMode](),
													},
													"surround_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Eac3SurroundMode](),
													},
												},
											},
										},
										"mp2_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"bitrate": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"coding_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.Mp2CodingMode](),
													},
													"sample_rate": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
												},
											},
										},
										"pass_through_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"wav_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"bit_depth": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"coding_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.WavCodingMode](),
													},
													"sample_rate": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
												},
											},
										},
									},
								},
							},
							names.AttrLanguageCode: {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"language_code_control": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.AudioDescriptionLanguageCodeControl](),
							},
							"remix_settings": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"channel_mappings": {
											Type:     schema.TypeSet,
											Required: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"input_channel_levels": {
														Type:     schema.TypeSet,
														Required: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"gain": {
																	Type:     schema.TypeInt,
																	Required: true,
																},
																"input_channel": {
																	Type:     schema.TypeInt,
																	Required: true,
																},
															},
														},
													},
													"output_channel": {
														Type:     schema.TypeInt,
														Required: true,
													},
												},
											},
										},
										"channels_in": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
										"channels_out": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
									},
								},
							},
							"stream_name": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				"output_groups": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"output_group_settings": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"archive_group_settings": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrDestination: func() *schema.Schema {
														return destinationSchema()
													}(),
													"archive_cdn_settings": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"archive_s3_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"canned_acl": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.S3CannedAcl](),
																			},
																		},
																	},
																},
															},
														},
													},
													"rollover_interval": {
														Type:     schema.TypeInt,
														Optional: true,
													},
												},
											},
										},
										"frame_capture_group_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrDestination: func() *schema.Schema {
														return destinationSchema()
													}(),
													"frame_capture_cdn_settings": {
														Type:     schema.TypeList,
														Optional: true,
														Computed: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"frame_capture_s3_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	Computed: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"canned_acl": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.S3CannedAcl](),
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
										"hls_group_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrDestination: func() *schema.Schema {
														return destinationSchema()
													}(),
													"ad_markers": {
														Type:     schema.TypeList,
														Optional: true,
														Computed: true,
														Elem: &schema.Schema{
															Type:             schema.TypeString,
															ValidateDiagFunc: enum.Validate[types.HlsAdMarkers](),
														},
													},
													"base_url_content": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"base_url_content1": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"base_url_manifest": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"base_url_manifest1": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"caption_language_mappings": {
														Type:     schema.TypeSet,
														Optional: true,
														MaxItems: 4,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"caption_channel": {
																	Type:     schema.TypeInt,
																	Required: true,
																},
																names.AttrLanguageCode: {
																	Type:     schema.TypeString,
																	Required: true,
																},
																"language_description": {
																	Type:     schema.TypeString,
																	Required: true,
																},
															},
														},
													},
													"caption_language_setting": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsCaptionLanguageSetting](),
													},
													"client_cache": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsClientCache](),
													},
													"codec_specification": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsCodecSpecification](),
													},
													"constant_iv": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"directory_structure": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsDirectoryStructure](),
													},
													"discontinuity_tags": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsDiscontinuityTags](),
													},
													"encryption_type": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsEncryptionType](),
													},
													"hls_cdn_settings": {
														Type:     schema.TypeList,
														Optional: true,
														Computed: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"hls_akamai_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"connection_retry_interval": func() *schema.Schema {
																				return connectionRetryIntervalSchema()
																			}(),
																			"filecache_duration": func() *schema.Schema {
																				return filecacheDurationSchema()
																			}(),
																			"http_transfer_mode": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				Computed:         true,
																				ValidateDiagFunc: enum.Validate[types.HlsAkamaiHttpTransferMode](),
																			},
																			"num_retries": func() *schema.Schema {
																				return numRetriesSchema()
																			}(),
																			"restart_delay": func() *schema.Schema {
																				return restartDelaySchema()
																			}(),
																			"salt": {
																				Type:     schema.TypeString,
																				Optional: true,
																				Computed: true,
																			},
																			"token": {
																				Type:     schema.TypeString,
																				Optional: true,
																				Computed: true,
																			},
																		},
																	},
																},
																"hls_basic_put_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"connection_retry_interval": func() *schema.Schema {
																				return connectionRetryIntervalSchema()
																			}(),
																			"filecache_duration": func() *schema.Schema {
																				return filecacheDurationSchema()
																			}(),
																			"num_retries": func() *schema.Schema {
																				return numRetriesSchema()
																			}(),
																			"restart_delay": func() *schema.Schema {
																				return restartDelaySchema()
																			}(),
																		},
																	},
																},
																"hls_media_store_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"connection_retry_interval": func() *schema.Schema {
																				return connectionRetryIntervalSchema()
																			}(),
																			"filecache_duration": func() *schema.Schema {
																				return filecacheDurationSchema()
																			}(),
																			"media_store_storage_class": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				Computed:         true,
																				ValidateDiagFunc: enum.Validate[types.HlsMediaStoreStorageClass](),
																			},
																			"num_retries": func() *schema.Schema {
																				return numRetriesSchema()
																			}(),
																			"restart_delay": func() *schema.Schema {
																				return restartDelaySchema()
																			}(),
																		},
																	},
																},
																"hls_s3_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	Computed: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"canned_acl": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.S3CannedAcl](),
																			},
																		},
																	},
																},
																"hls_webdav_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	Computed: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"connection_retry_interval": func() *schema.Schema {
																				return connectionRetryIntervalSchema()
																			}(),
																			"filecache_duration": func() *schema.Schema {
																				return filecacheDurationSchema()
																			}(),
																			"http_transfer_mode": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				Computed:         true,
																				ValidateDiagFunc: enum.Validate[types.HlsWebdavHttpTransferMode](),
																			},
																			"num_retries": func() *schema.Schema {
																				return numRetriesSchema()
																			}(),
																			"restart_delay": func() *schema.Schema {
																				return restartDelaySchema()
																			}(),
																		},
																	},
																},
															},
														},
													},
													"hls_id3_segment_tagging": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsId3SegmentTaggingState](),
													},
													"iframe_only_playlists": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.IFrameOnlyPlaylistType](),
													},
													"incomplete_segment_behavior": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsIncompleteSegmentBehavior](),
													},
													"index_n_segments": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"input_loss_action": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.InputLossActionForHlsOut](),
													},
													"iv_in_manifest": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsIvInManifest](),
													},
													"iv_source": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsIvSource](),
													},
													"keep_segments": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"key_format": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"key_format_versions": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"key_provider_settings": {
														Type:     schema.TypeList,
														Optional: true,
														Computed: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"static_key_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"static_key_value": {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																			"key_provider_server": func() *schema.Schema {
																				return inputLocationSchema()
																			}(),
																		},
																	},
																},
															},
														},
													},
													"manifest_compression": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsManifestCompression](),
													},
													"manifest_duration_format": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsManifestDurationFormat](),
													},
													"min_segment_length": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													names.AttrMode: {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsMode](),
													},
													"output_selection": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsOutputSelection](),
													},
													"program_date_time": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsProgramDateTime](),
													},
													"program_date_time_clock": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsProgramDateTimeClock](),
													},
													"program_date_time_period": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"redundant_manifest": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsRedundantManifest](),
													},
													"segment_length": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"segments_per_subdirectory": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"stream_inf_resolution": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsStreamInfResolution](),
													},
													"timed_metadata_id3_frame": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsTimedMetadataId3Frame](),
													},
													"timed_metadata_id3_period": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"timestamp_delta_milliseconds": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"ts_file_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsTsFileMode](),
													},
												},
											},
										},
										"media_package_group_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrDestination: func() *schema.Schema {
														return destinationSchema()
													}(),
												},
											},
										},
										"multiplex_group_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{},
											},
										},
										"ms_smooth_group_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrDestination: func() *schema.Schema {
														return destinationSchema()
													}(),
													"acquisition_point_id": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"audio_only_timecode_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.SmoothGroupAudioOnlyTimecodeControl](),
													},
													"certificate_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.SmoothGroupCertificateMode](),
													},
													"connection_retry_interval": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"event_id": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"event_id_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.SmoothGroupEventIdMode](),
													},
													"event_stop_behavior": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.SmoothGroupEventStopBehavior](),
													},
													"filecache_duration": func() *schema.Schema {
														return filecacheDurationSchema()
													}(),
													"fragment_length": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"input_loss_action": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.InputLossActionForMsSmoothOut](),
													},
													"num_retries": func() *schema.Schema {
														return numRetriesSchema()
													}(),
													"restart_delay": func() *schema.Schema {
														return restartDelaySchema()
													}(),
													"segmentation_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.SmoothGroupSegmentationMode](),
													},
													"send_delay_ms": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"sparse_track_type": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.SmoothGroupSparseTrackType](),
													},
													"stream_manifest_behavior": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.SmoothGroupStreamManifestBehavior](),
													},
													"timestamp_offset": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"timestamp_offset_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.SmoothGroupTimestampOffsetMode](),
													},
												},
											},
										},
										"rtmp_group_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"ad_markers": {
														Type:     schema.TypeList,
														Optional: true,
														Elem: &schema.Schema{
															Type:             schema.TypeString,
															ValidateDiagFunc: enum.Validate[types.RtmpAdMarkers](),
														},
													},
													"authentication_scheme": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AuthenticationScheme](),
													},
													"cache_full_behavior": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.RtmpCacheFullBehavior](),
													},
													"cache_length": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"caption_data": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.RtmpCaptionData](),
													},
													"input_loss_action": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.InputLossActionForRtmpOut](),
													},
													"restart_delay": func() *schema.Schema {
														return restartDelaySchema()
													}(),
												},
											},
										},
										"udp_group_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"input_loss_action": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.InputLossActionForUdpOut](),
													},
													"timed_metadata_id3_frame": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.UdpTimedMetadataId3Frame](),
													},
													"timed_metadata_id3_period": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
												},
											},
										},
									},
								},
							},
							"outputs": {
								Type:     schema.TypeList,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"output_settings": func() *schema.Schema {
											return outputSettingsSchema()
										}(),
										"audio_description_names": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"caption_description_names": {
											Type:     schema.TypeSet,
											Optional: true,
											Computed: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"output_name": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"video_description_name": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"timecode_config": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrSource: {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.TimecodeConfigSource](),
							},
							"sync_threshold": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				"video_descriptions": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"codec_settings": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"frame_capture_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"capture_interval": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"capture_interval_units": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.FrameCaptureIntervalUnit](),
													},
												},
											},
										},
										"h264_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"adaptive_quantization": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264AdaptiveQuantization](),
													},
													"afd_signaling": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AfdSignaling](),
													},
													"bitrate": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"buf_fill_pct": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"buf_size": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"color_metadata": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264ColorMetadata](),
													},
													"entropy_encoding": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264EntropyEncoding](),
													},
													"filter_settings": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"temporal_filter_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"post_filter_sharpening": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.TemporalFilterPostFilterSharpening](),
																			},
																			"strength": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.TemporalFilterStrength](),
																			},
																		},
																	},
																},
															},
														},
													},
													"fixed_afd": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.FixedAfd](),
													},
													"flicker_aq": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264FlickerAq](),
													},
													"force_field_pictures": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264ForceFieldPictures](),
													},
													"framerate_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264FramerateControl](),
													},
													"framerate_denominator": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"framerate_numerator": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"gop_b_reference": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264GopBReference](),
													},
													"gop_closed_cadence": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"gop_num_b_frames": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"gop_size": {
														Type:     schema.TypeFloat,
														Optional: true,
														Computed: true,
													},
													"gop_size_units": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264GopSizeUnits](),
													},
													"level": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264Level](),
													},
													"look_ahead_rate_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264LookAheadRateControl](),
													},
													"max_bitrate": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"min_i_interval": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"num_ref_frames": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"par_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264ParControl](),
													},
													"par_denominator": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"par_numerator": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													names.AttrProfile: {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264Profile](),
													},
													"quality_level": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264QualityLevel](),
													},
													"qvbr_quality_level": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"rate_control_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264RateControlMode](),
													},
													"scan_type": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264ScanType](),
													},
													"scene_change_detect": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264SceneChangeDetect](),
													},
													"slices": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"softness": {
														Type:     schema.TypeInt,
														Optional: true,
														Computed: true,
													},
													"spatial_aq": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264SpatialAq](),
													},
													"subgop_length": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264SubGopLength](),
													},
													"syntax": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264Syntax](),
													},
													"temporal_aq": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264TemporalAq](),
													},
													"timecode_insertion": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H264TimecodeInsertionBehavior](),
													},
												},
											},
										},
										"h265_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"framerate_denominator": {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"framerate_numerator": {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"adaptive_quantization": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265AdaptiveQuantization](),
													},
													"afd_signaling": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.AfdSignaling](),
													},
													"alternative_transfer_function": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265AlternativeTransferFunction](),
													},
													"bitrate": {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"buf_size": {
														Type:         schema.TypeInt,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"color_metadata": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265ColorMetadata](),
													},
													"color_space_settings": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"color_space_passthrough_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{}, // no exported elements in this list
																	},
																},
																"dolby_vision81_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{}, // no exported elements in this list
																	},
																},
																"hdr10_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"max_cll": {
																				Type:         schema.TypeInt,
																				Default:      0,
																				Optional:     true,
																				ValidateFunc: validation.IntAtLeast(0),
																			},
																			"max_fall": {
																				Type:         schema.TypeInt,
																				Default:      0,
																				Optional:     true,
																				ValidateFunc: validation.IntAtLeast(0),
																			},
																		},
																	},
																},
																"rec601_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{}, // no exported elements in this list
																	},
																},
																"rec709_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{}, // no exported elements in this list
																	},
																},
															},
														},
													},
													"filter_settings": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"temporal_filter_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"post_filter_sharpening": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.TemporalFilterPostFilterSharpening](),
																			},
																			"strength": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				ValidateDiagFunc: enum.Validate[types.TemporalFilterStrength](),
																			},
																		},
																	},
																},
															},
														},
													},
													"fixed_afd": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.FixedAfd](),
													},
													"flicker_aq": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265FlickerAq](),
													},

													"gop_closed_cadence": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"gop_size": {
														Type:     schema.TypeFloat,
														Optional: true,
													},
													"gop_size_units": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265GopSizeUnits](),
													},
													"level": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265Level](),
													},
													"look_ahead_rate_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265LookAheadRateControl](),
													},
													"max_bitrate": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"min_i_interval": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"min_qp": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"mv_over_picture_boundaries": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265MvOverPictureBoundaries](),
													},
													"mv_temporal_predictor": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265MvTemporalPredictor](),
													},
													"par_denominator": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"par_numerator": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													names.AttrProfile: {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265Profile](),
													},
													"qvbr_quality_level": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"rate_control_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265RateControlMode](),
													},
													"scan_type": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265ScanType](),
													},
													"scene_change_detect": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265SceneChangeDetect](),
													},
													"slices": {
														Type:         schema.TypeInt,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"tier": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265Tier](),
													},
													"tile_height": {
														Type:         schema.TypeInt,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(64),
													},
													"tile_padding": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265TilePadding](),
													},
													"tile_width": {
														Type:         schema.TypeInt,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(256),
													},
													"timecode_burnin_settings": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"timecode_burnin_font_size": {
																	Type:             schema.TypeString,
																	Optional:         true,
																	Computed:         true,
																	ValidateDiagFunc: enum.Validate[types.TimecodeBurninFontSize](),
																},
																"timecode_burnin_position": {
																	Type:             schema.TypeString,
																	Optional:         true,
																	Computed:         true,
																	ValidateDiagFunc: enum.Validate[types.TimecodeBurninPosition](),
																},
																names.AttrPrefix: {
																	Type:     schema.TypeString,
																	Optional: true,
																	Computed: true,
																},
															},
														},
													},
													"timecode_insertion": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265TimecodeInsertionBehavior](),
													},
													"treeblock_size": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.H265TreeblockSize](),
													},
												},
											},
										},
										// TODO mgeg2_settings
									},
								},
							},
							"height": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"respond_to_afd": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.VideoDescriptionRespondToAfd](),
							},
							"scaling_behavior": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.VideoDescriptionScalingBehavior](),
							},
							"sharpness": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"width": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				"avail_blanking": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"avail_blanking_image": func() *schema.Schema {
								return inputLocationSchema()
							}(),
							names.AttrState: {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				// TODO avail_configuration
				// TODO blackout_slate
				"caption_descriptions": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"caption_selector_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"accessibility": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.AccessibilityType](),
							},

							"destination_settings": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"arib_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"burn_in_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"alignment": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.BurnInAlignment](),
													},
													"background_color": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.BurnInBackgroundColor](),
													},
													"background_opacity": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"font": func() *schema.Schema {
														return inputLocationSchema()
													}(),
													"font_color": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.BurnInFontColor](),
													},
													"font_opacity": {
														Type:         schema.TypeInt,
														Default:      0,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(0),
													},
													"font_resolution": {
														Type:         schema.TypeInt,
														Default:      96,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"font_size": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"outline_color": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[types.BurnInOutlineColor](),
													},
													"outline_size": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"shadow_color": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.BurnInShadowColor](),
													},
													"shadow_opacity": {
														Type:         schema.TypeInt,
														Default:      0,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(0),
													},
													"shadow_x_offset": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"shadow_y_offset": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"teletext_grid_control": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[types.BurnInTeletextGridControl](),
													},
													"x_position": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"y_position": {
														Type:     schema.TypeInt,
														Optional: true,
													},
												},
											},
										},
										"dvb_sub_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"alignment": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.DvbSubDestinationAlignment](),
													},
													"background_color": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.DvbSubDestinationBackgroundColor](),
													},
													"background_opacity": {
														Type:         schema.TypeInt,
														Default:      0,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(0),
													},
													"font": func() *schema.Schema {
														return inputLocationSchema()
													}(),
													"font_color": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.DvbSubDestinationFontColor](),
													},
													"font_opacity": {
														Type:         schema.TypeInt,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(0),
													},
													"font_resolution": {
														Type:         schema.TypeInt,
														Default:      96,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"font_size": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"outline_color": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.DvbSubDestinationOutlineColor](),
													},
													"outline_size": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"shadow_color": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.DvbSubDestinationShadowColor](),
													},
													"shadow_opacity": {
														Type:         schema.TypeInt,
														Default:      0,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(0),
													},
													"shadow_x_offset": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"shadow_y_offset": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"teletext_grid_control": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.DvbSubDestinationTeletextGridControl](),
													},
													"x_position": {
														Type:     schema.TypeInt,
														Optional: true,
													},
													"y_position": {
														Type:     schema.TypeInt,
														Optional: true,
													},
												},
											},
										},
										"ebu_tt_d_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"copyright_holder": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"fill_line_gap": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.EbuTtDFillLineGapControl](),
													},
													"font_family": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"style_control": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.EbuTtDDestinationStyleControl](),
													},
												},
											},
										},
										"embedded_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"embedded_plus_scte20_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"rtmp_caption_info_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"scte20_plus_embedded_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"scte27_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"smpte_tt_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"teletext_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"ttml_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"style_control": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[types.TtmlDestinationStyleControl](),
													},
												},
											},
										},
										"webvtt_destination_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"style_control": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[types.WebvttDestinationStyleControl](),
													},
												},
											},
										},
									},
								},
							},
							names.AttrLanguageCode: {
								Type:     schema.TypeString,
								Optional: true,
							},
							"language_description": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				// TODO feature_activations
				"global_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"initial_audio_gain": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"input_end_action": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.GlobalConfigurationInputEndAction](),
							},
							"input_loss_behavior": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"black_frame_msec": {
											Type:     schema.TypeInt,
											Optional: true,
										},
										"input_loss_image_color": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"input_loss_image_slate": func() *schema.Schema {
											return inputLocationSchema()
										}(),

										"input_loss_image_type": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.InputLossImageType](),
										},
										"repeat_frame_msec": {
											Type:     schema.TypeInt,
											Optional: true,
										},
									},
								},
							},
							"output_locking_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.GlobalConfigurationOutputLockingMode](),
							},
							"output_timing_source": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.GlobalConfigurationOutputTimingSource](),
							},
							"support_low_framerate_inputs": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.GlobalConfigurationLowFramerateInputs](),
							},
						},
					},
				},
				"motion_graphics_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"motion_graphics_settings": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"html_motion_graphics_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
									},
								},
							},
							"motion_graphics_insertion": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.MotionGraphicsInsertion](),
							},
						},
					},
				},
				"nielsen_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"distributor_id": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"nielsen_pcm_to_id3_tagging": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.NielsenPcmToId3TaggingState](),
							},
						},
					},
				},
			},
		},
	}
}
func outputSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"archive_output_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"container_settings": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"m2ts_settings": func() *schema.Schema {
											return m2tsSettingsSchema()
										}(),
										// This is in the API and Go SDK docs, but has no exported fields.
										"raw_settings": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{},
											},
										},
									},
								},
							},
							"extension": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"name_modifier": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"frame_capture_output_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name_modifier": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				"hls_output_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"hls_settings": func() *schema.Schema {
								return hlsSettingsSchema()
							}(),
							"h265_packaging_type": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"name_modifier": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"segment_modifier": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				// This is in the API and Go SDK docs, but has no exported fields.
				"media_package_output_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{},
					},
				},
				"ms_smooth_output_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"h265_packaging_type": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.MsSmoothH265PackagingType](),
							},
							"name_modifier": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				"multiplex_output_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDestination: destinationSchema(),
						},
					},
				},
				"rtmp_output_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDestination: destinationSchema(),
							"certificate_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.RtmpOutputCertificateMode](),
							},
							"connection_retry_interval": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"num_retries": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				"udp_output_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"container_settings": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"m2ts_settings": func() *schema.Schema {
											return m2tsSettingsSchema()
										}(),
									}},
							},
							names.AttrDestination: destinationSchema(),
							"buffer_msec": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"fec_output_settings": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"column_depth": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
										"include_fec": {
											Type:             schema.TypeString,
											Optional:         true,
											Computed:         true,
											ValidateDiagFunc: enum.Validate[types.FecOutputIncludeFec](),
										},
										"row_length": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func hlsSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"audio_only_hls_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"audio_group_id": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"audio_only_image": func() *schema.Schema {
								return inputLocationSchema()
							}(),
							"audio_track_type": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.AudioOnlyHlsTrackType](),
							},
							"segment_type": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.AudioOnlyHlsSegmentType](),
							},
						},
					},
				},
				"fmp4_hls_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"audio_rendition_sets": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
							"nielsen_id3_behavior": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.Fmp4NielsenId3Behavior](),
							},
							"timed_metadata_behavior": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.Fmp4TimedMetadataBehavior](),
							},
						},
					},
				},
				// This is in the API and Go SDK docs, but has no exported fields.
				"frame_capture_hls_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{},
					},
				},
				"standard_hls_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"m3u8_settings": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"audio_frames_per_pes": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
										"audio_pids": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
										"ecm_pid": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
										"nielsen_id3_behavior": {
											Type:             schema.TypeString,
											Optional:         true,
											Computed:         true,
											ValidateDiagFunc: enum.Validate[types.M3u8NielsenId3Behavior](),
										},
										"pat_interval": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
										"pcr_control": {
											Type:             schema.TypeString,
											Optional:         true,
											Computed:         true,
											ValidateDiagFunc: enum.Validate[types.M3u8PcrControl](),
										},
										"pcr_period": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
										"pcr_pid": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
										"pmt_interval": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
										"pmt_pid": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
										"program_num": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
										"scte35_behavior": {
											Type:             schema.TypeString,
											Optional:         true,
											Computed:         true,
											ValidateDiagFunc: enum.Validate[types.M3u8Scte35Behavior](),
										},
										"scte35_pid": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
										"timed_metadata_behavior": {
											Type:             schema.TypeString,
											Optional:         true,
											Computed:         true,
											ValidateDiagFunc: enum.Validate[types.M3u8TimedMetadataBehavior](),
										},
										"timed_metadata_pid": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
										"transport_stream_id": {
											Type:     schema.TypeInt,
											Optional: true,
											Computed: true,
										},
										"video_pid": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
									},
								},
							},
							"audio_rendition_sets": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
			},
		},
	}
}

func m2tsSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"absent_input_audio_behavior": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsAbsentInputAudioBehavior](),
				},
				"arib": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsArib](),
				},
				"arib_captions_pid": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"arib_captions_pid_control": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsAribCaptionsPidControl](),
				},
				"audio_buffer_model": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsAudioBufferModel](),
				},
				"audio_frames_per_pes": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"audio_pids": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"audio_stream_type": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsAudioStreamType](),
				},
				"bitrate": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"buffer_model": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsBufferModel](),
				},
				"cc_descriptor": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsCcDescriptor](),
				},
				"dvb_nit_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"network_id": {
								Type:     schema.TypeInt,
								Required: true,
							},
							"network_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"rep_interval": {
								Type:     schema.TypeInt,
								Optional: true,
							},
						},
					},
				},
				"dvb_sdt_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"output_sdt": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[types.DvbSdtOutputSdt](),
							},
							"rep_interval": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							names.AttrServiceName: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
							"service_provider_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
						},
					},
				},
				"dvb_sub_pids": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"dvb_tdt_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"rep_interval": {
								Type:     schema.TypeInt,
								Optional: true,
							},
						},
					},
				},
				"dvb_teletext_pid": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"ebif": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsEbifControl](),
				},
				"ebp_audio_interval": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsAudioInterval](),
				},
				"ebp_lookahead_ms": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"ebp_placement": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsEbpPlacement](),
				},
				"ecm_pid": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"es_rate_in_pes": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsEsRateInPes](),
				},
				"etv_platform_pid": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"etv_signal_pid": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"fragment_time": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
				"klv": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsKlv](),
				},
				"klv_data_pids": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"nielsen_id3_behavior": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsNielsenId3Behavior](),
				},
				"null_packet_bitrate": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
				"pat_interval": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"pcr_control": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsPcrControl](),
				},
				"pcr_period": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"pcr_pid": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"pmt_interval": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"pmt_pid": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"program_num": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"rate_mode": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsRateMode](),
				},
				"scte27_pids": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"scte35_control": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsScte35Control](),
				},
				"scte35_pid": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"segmentation_markers": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsSegmentationMarkers](),
				},
				"segmentation_style": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsSegmentationStyle](),
				},
				"segmentation_time": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
				"timed_metadata_behavior": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsTimedMetadataBehavior](),
				},
				"timed_metadata_pid": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"transport_stream_id": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"video_pid": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

func expandChannelEncoderSettings(tfList []any) *types.EncoderSettings {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]any)

	var settings types.EncoderSettings
	if v, ok := m["audio_descriptions"].(*schema.Set); ok {
		settings.AudioDescriptions = expandChannelEncoderSettingsAudioDescriptions(v.List())
	}
	if v, ok := m["output_groups"].([]any); ok && len(v) > 0 {
		settings.OutputGroups = expandChannelEncoderSettingsOutputGroups(v)
	}
	if v, ok := m["timecode_config"].([]any); ok && len(v) > 0 {
		settings.TimecodeConfig = expandChannelEncoderSettingsTimecodeConfig(v)
	}
	if v, ok := m["video_descriptions"].([]any); ok && len(v) > 0 {
		settings.VideoDescriptions = expandChannelEncoderSettingsVideoDescriptions(v)
	}
	if v, ok := m["avail_blanking"].([]any); ok && len(v) > 0 {
		settings.AvailBlanking = expandChannelEncoderSettingsAvailBlanking(v)
	}
	if v, ok := m["avail_configuration"].([]any); ok && len(v) > 0 {
		settings.AvailConfiguration = nil // TODO expandChannelEncoderSettingsAvailConfiguration(v)
	}
	if v, ok := m["blackout_slate"].([]any); ok && len(v) > 0 {
		settings.BlackoutSlate = nil // TODO expandChannelEncoderSettingsBlackoutSlate(v)
	}
	if v, ok := m["caption_descriptions"].([]any); ok && len(v) > 0 {
		settings.CaptionDescriptions = expandChannelEncoderSettingsCaptionDescriptions(v)
	}
	if v, ok := m["feature_activations"].([]any); ok && len(v) > 0 {
		settings.FeatureActivations = nil // TODO expandChannelEncoderSettingsFeatureActivations(v)
	}
	if v, ok := m["global_configuration"].([]any); ok && len(v) > 0 {
		settings.GlobalConfiguration = expandChannelEncoderSettingsGlobalConfiguration(v)
	}
	if v, ok := m["motion_graphics_configuration"].([]any); ok && len(v) > 0 {
		settings.MotionGraphicsConfiguration = expandChannelEncoderSettingsMotionGraphicsConfiguration(v)
	}
	if v, ok := m["nielsen_configuration"].([]any); ok && len(v) > 0 {
		settings.NielsenConfiguration = expandChannelEncoderSettingsNielsenConfiguration(v)
	}

	return &settings
}

func expandChannelEncoderSettingsAudioDescriptions(tfList []any) []types.AudioDescription {
	if tfList == nil {
		return nil
	}

	audioDesc := []types.AudioDescription{}
	for _, tfItem := range tfList {
		m, ok := tfItem.(map[string]any)
		if !ok {
			continue
		}

		var a types.AudioDescription
		if v, ok := m["audio_selector_name"].(string); ok && v != "" {
			a.AudioSelectorName = aws.String(v)
		}
		if v, ok := m[names.AttrName].(string); ok && v != "" {
			a.Name = aws.String(v)
		}
		if v, ok := m["audio_normalization_settings"].([]any); ok && len(v) > 0 {
			a.AudioNormalizationSettings = expandAudioDescriptionsAudioNormalizationSettings(v)
		}
		if v, ok := m["audio_type"].(string); ok && v != "" {
			a.AudioType = types.AudioType(v)
		}
		if v, ok := m["audio_type_control"].(string); ok && v != "" {
			a.AudioTypeControl = types.AudioDescriptionAudioTypeControl(v)
		}
		if v, ok := m["audio_watermark_settings"].([]any); ok && len(v) > 0 {
			a.AudioWatermarkingSettings = expandAudioWatermarkSettings(v)
		}
		if v, ok := m["codec_settings"].([]any); ok && len(v) > 0 {
			a.CodecSettings = expandChannelEncoderSettingsAudioDescriptionsCodecSettings(v)
		}
		if v, ok := m[names.AttrLanguageCode].(string); ok && v != "" {
			a.LanguageCode = aws.String(v)
		}
		if v, ok := m["language_code_control"].(string); ok && v != "" {
			a.LanguageCodeControl = types.AudioDescriptionLanguageCodeControl(v)
		}
		if v, ok := m["remix_settings"].([]any); ok && len(v) > 0 {
			a.RemixSettings = expandChannelEncoderSettingsAudioDescriptionsRemixSettings(v)
		}
		if v, ok := m["stream_name"].(string); ok && v != "" {
			a.StreamName = aws.String(v)
		}

		audioDesc = append(audioDesc, a)
	}

	return audioDesc
}

func expandChannelEncoderSettingsOutputGroups(tfList []any) []types.OutputGroup {
	if tfList == nil {
		return nil
	}

	var outputGroups []types.OutputGroup
	for _, tfItem := range tfList {
		m, ok := tfItem.(map[string]any)
		if !ok {
			continue
		}

		var o types.OutputGroup
		if v, ok := m["output_group_settings"].([]any); ok && len(v) > 0 {
			o.OutputGroupSettings = expandChannelEncoderSettingsOutputGroupsOutputGroupSettings(v)
		}
		if v, ok := m["outputs"].([]any); ok && len(v) > 0 {
			o.Outputs = expandChannelEncoderSettingsOutputGroupsOutputs(v)
		}
		if v, ok := m[names.AttrName].(string); ok && v != "" {
			o.Name = aws.String(v)
		}

		outputGroups = append(outputGroups, o)
	}

	return outputGroups
}

func expandAudioDescriptionsAudioNormalizationSettings(tfList []any) *types.AudioNormalizationSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.AudioNormalizationSettings
	if v, ok := m["algorithm"].(string); ok && v != "" {
		out.Algorithm = types.AudioNormalizationAlgorithm(v)
	}
	if v, ok := m["algorithm_control"].(string); ok && v != "" {
		out.AlgorithmControl = types.AudioNormalizationAlgorithmControl(v)
	}
	if v, ok := m["target_lkfs"].(float32); ok && v != 0.0 {
		out.TargetLkfs = aws.Float64(float64(v))
	}

	return &out
}

func expandChannelEncoderSettingsAudioDescriptionsCodecSettings(tfList []any) *types.AudioCodecSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.AudioCodecSettings
	if v, ok := m["aac_settings"].([]any); ok && len(v) > 0 {
		out.AacSettings = expandAudioDescriptionsCodecSettingsAacSettings(v)
	}
	if v, ok := m["ac3_settings"].([]any); ok && len(v) > 0 {
		out.Ac3Settings = expandAudioDescriptionsCodecSettingsAc3Settings(v)
	}
	if v, ok := m["eac3_atmos_settings"].([]any); ok && len(v) > 0 {
		out.Eac3AtmosSettings = expandAudioDescriptionsCodecSettingsEac3AtmosSettings(v)
	}
	if v, ok := m["eac3_settings"].([]any); ok && len(v) > 0 {
		out.Eac3Settings = expandAudioDescriptionsCodecSettingsEac3Settings(v)
	}
	if v, ok := m["vp2_settings"].([]any); ok && len(v) > 0 {
		out.Mp2Settings = expandAudioDescriptionsCodecSettingsMp2Settings(v)
	}
	if v, ok := m["pass_through_settings"].([]any); ok && len(v) > 0 {
		out.PassThroughSettings = &types.PassThroughSettings{} // no exported fields
	}
	if v, ok := m["wav_settings"].([]any); ok && len(v) > 0 {
		out.WavSettings = expandAudioDescriptionsCodecSettingsWavSettings(v)
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsAacSettings(tfList []any) *types.AacSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.AacSettings
	if v, ok := m["bitrate"].(float64); ok && v != 0.0 {
		out.Bitrate = aws.Float64(v)
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.AacCodingMode(v)
	}
	if v, ok := m["input_type"].(string); ok && v != "" {
		out.InputType = types.AacInputType(v)
	}
	if v, ok := m[names.AttrProfile].(string); ok && v != "" {
		out.Profile = types.AacProfile(v)
	}
	if v, ok := m["rate_control_mode"].(string); ok && v != "" {
		out.RateControlMode = types.AacRateControlMode(v)
	}
	if v, ok := m["raw_format"].(string); ok && v != "" {
		out.RawFormat = types.AacRawFormat(v)
	}
	if v, ok := m["sample_rate"].(float64); ok && v != 0.0 {
		out.SampleRate = aws.Float64(v)
	}
	if v, ok := m["spec"].(string); ok && v != "" {
		out.Spec = types.AacSpec(v)
	}
	if v, ok := m["vbr_quality"].(string); ok && v != "" {
		out.VbrQuality = types.AacVbrQuality(v)
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsAc3Settings(tfList []any) *types.Ac3Settings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.Ac3Settings
	if v, ok := m["bitrate"].(float64); ok && v != 0.0 {
		out.Bitrate = aws.Float64(v)
	}
	if v, ok := m["bitstream_mode"].(string); ok && v != "" {
		out.BitstreamMode = types.Ac3BitstreamMode(v)
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.Ac3CodingMode(v)
	}
	if v, ok := m["dialnorm"].(int); ok && v != 0 {
		out.Dialnorm = aws.Int32(int32(v))
	}
	if v, ok := m["drc_profile"].(string); ok && v != "" {
		out.DrcProfile = types.Ac3DrcProfile(v)
	}
	if v, ok := m["lfe_filter"].(string); ok && v != "" {
		out.LfeFilter = types.Ac3LfeFilter(v)
	}
	if v, ok := m["metadata_control"].(string); ok && v != "" {
		out.MetadataControl = types.Ac3MetadataControl(v)
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsEac3AtmosSettings(tfList []any) *types.Eac3AtmosSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.Eac3AtmosSettings
	if v, ok := m["bitrate"].(float32); ok && v != 0.0 {
		out.Bitrate = aws.Float64(float64(v))
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.Eac3AtmosCodingMode(v)
	}
	if v, ok := m["dialnorm"].(int); ok && v != 0 {
		out.Dialnorm = aws.Int32(int32(v))
	}
	if v, ok := m["drc_line"].(string); ok && v != "" {
		out.DrcLine = types.Eac3AtmosDrcLine(v)
	}
	if v, ok := m["drc_rf"].(string); ok && v != "" {
		out.DrcRf = types.Eac3AtmosDrcRf(v)
	}
	if v, ok := m["height_trim"].(float32); ok && v != 0.0 {
		out.HeightTrim = aws.Float64(float64(v))
	}
	if v, ok := m["surround_trim"].(float32); ok && v != 0.0 {
		out.SurroundTrim = aws.Float64(float64(v))
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsEac3Settings(tfList []any) *types.Eac3Settings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.Eac3Settings
	if v, ok := m["attenuation_control"].(string); ok && v != "" {
		out.AttenuationControl = types.Eac3AttenuationControl(v)
	}
	if v, ok := m["bitrate"].(float32); ok && v != 0.0 {
		out.Bitrate = aws.Float64(float64(v))
	}
	if v, ok := m["bitstream_mode"].(string); ok && v != "" {
		out.BitstreamMode = types.Eac3BitstreamMode(v)
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.Eac3CodingMode(v)
	}
	if v, ok := m["dc_filter"].(string); ok && v != "" {
		out.DcFilter = types.Eac3DcFilter(v)
	}
	if v, ok := m["dialnorm"].(int); ok && v != 0 {
		out.Dialnorm = aws.Int32(int32(v))
	}
	if v, ok := m["drc_line"].(string); ok && v != "" {
		out.DrcLine = types.Eac3DrcLine(v)
	}
	if v, ok := m["drc_rf"].(string); ok && v != "" {
		out.DrcRf = types.Eac3DrcRf(v)
	}
	if v, ok := m["lfe_control"].(string); ok && v != "" {
		out.LfeControl = types.Eac3LfeControl(v)
	}
	if v, ok := m["lfe_filter"].(string); ok && v != "" {
		out.LfeFilter = types.Eac3LfeFilter(v)
	}
	if v, ok := m["lo_ro_center_mix_level"].(float32); ok && v != 0.0 {
		out.LoRoCenterMixLevel = aws.Float64(float64(v))
	}
	if v, ok := m["lo_ro_surround_mix_level"].(float32); ok && v != 0.0 {
		out.LoRoSurroundMixLevel = aws.Float64(float64(v))
	}
	if v, ok := m["lt_rt_center_mix_level"].(float32); ok && v != 0.0 {
		out.LtRtCenterMixLevel = aws.Float64(float64(v))
	}
	if v, ok := m["lt_rt_surround_mix_level"].(float32); ok && v != 0.0 {
		out.LtRtSurroundMixLevel = aws.Float64(float64(v))
	}
	if v, ok := m["metadata_control"].(string); ok && v != "" {
		out.MetadataControl = types.Eac3MetadataControl(v)
	}
	if v, ok := m["phase_control"].(string); ok && v != "" {
		out.PhaseControl = types.Eac3PhaseControl(v)
	}
	if v, ok := m["stereo_downmix"].(string); ok && v != "" {
		out.StereoDownmix = types.Eac3StereoDownmix(v)
	}
	if v, ok := m["surround_ex_mode"].(string); ok && v != "" {
		out.SurroundExMode = types.Eac3SurroundExMode(v)
	}
	if v, ok := m["surround_mode"].(string); ok && v != "" {
		out.SurroundMode = types.Eac3SurroundMode(v)
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsMp2Settings(tfList []any) *types.Mp2Settings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.Mp2Settings
	if v, ok := m["bitrate"].(float32); ok && v != 0.0 {
		out.Bitrate = aws.Float64(float64(v))
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.Mp2CodingMode(v)
	}
	if v, ok := m["sample_rate"].(float32); ok && v != 0.0 {
		out.Bitrate = aws.Float64(float64(v))
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsWavSettings(tfList []any) *types.WavSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.WavSettings
	if v, ok := m["bit_depth"].(float32); ok && v != 0.0 {
		out.BitDepth = aws.Float64(float64(v))
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.WavCodingMode(v)
	}
	if v, ok := m["sample_rate"].(float32); ok && v != 0.0 {
		out.SampleRate = aws.Float64(float64(v))
	}

	return &out
}

func expandChannelEncoderSettingsAudioDescriptionsRemixSettings(tfList []any) *types.RemixSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.RemixSettings
	if v, ok := m["channel_mappings"].(*schema.Set); ok && v.Len() > 0 {
		out.ChannelMappings = expandChannelMappings(v.List())
	}
	if v, ok := m["channels_in"].(int); ok && v != 0 {
		out.ChannelsIn = aws.Int32(int32(v))
	}
	if v, ok := m["channels_out"].(int); ok && v != 0 {
		out.ChannelsOut = aws.Int32(int32(v))
	}

	return &out
}

func expandChannelMappings(tfList []any) []types.AudioChannelMapping {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.AudioChannelMapping
	for _, item := range tfList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		var o types.AudioChannelMapping
		if v, ok := m["input_channel_levels"].(*schema.Set); ok && v.Len() > 0 {
			o.InputChannelLevels = expandInputChannelLevels(v.List())
		}
		if v, ok := m["output_channel"].(int); ok && v != 0 {
			o.OutputChannel = aws.Int32(int32(v))
		}

		out = append(out, o)
	}

	return out
}

func expandInputChannelLevels(tfList []any) []types.InputChannelLevel {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.InputChannelLevel
	for _, item := range tfList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		var o types.InputChannelLevel
		if v, ok := m["gain"].(int); ok && v != 0 {
			o.Gain = aws.Int32(int32(v))
		}
		if v, ok := m["input_channel"].(int); ok && v != 0 {
			o.InputChannel = aws.Int32(int32(v))
		}

		out = append(out, o)
	}

	return out
}

func expandChannelEncoderSettingsOutputGroupsOutputGroupSettings(tfList []any) *types.OutputGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var o types.OutputGroupSettings

	if v, ok := m["archive_group_settings"].([]any); ok && len(v) > 0 {
		o.ArchiveGroupSettings = expandArchiveGroupSettings(v)
	}
	if v, ok := m["frame_capture_group_settings"].([]any); ok && len(v) > 0 {
		o.FrameCaptureGroupSettings = expandFrameCaptureGroupSettings(v)
	}
	if v, ok := m["hls_group_settings"].([]any); ok && len(v) > 0 {
		o.HlsGroupSettings = expandHLSGroupSettings(v)
	}
	if v, ok := m["ms_smooth_group_settings"].([]any); ok && len(v) > 0 {
		o.MsSmoothGroupSettings = expandMsSmoothGroupSettings(v)
	}
	if v, ok := m["media_package_group_settings"].([]any); ok && len(v) > 0 {
		o.MediaPackageGroupSettings = expandMediaPackageGroupSettings(v)
	}
	if v, ok := m["multiplex_group_settings"].([]any); ok && len(v) > 0 {
		o.MultiplexGroupSettings = &types.MultiplexGroupSettings{} // only unexported fields
	}
	if v, ok := m["rtmp_group_settings"].([]any); ok && len(v) > 0 {
		o.RtmpGroupSettings = expandRtmpGroupSettings(v)
	}
	if v, ok := m["udp_group_settings"].([]any); ok && len(v) > 0 {
		o.UdpGroupSettings = expandUdpGroupSettings(v)
	}

	return &o
}

func expandDestination(in []any) *types.OutputLocationRef {
	if len(in) == 0 {
		return nil
	}

	m := in[0].(map[string]any)

	var out types.OutputLocationRef
	if v, ok := m["destination_ref_id"].(string); ok && v != "" {
		out.DestinationRefId = aws.String(v)
	}

	return &out
}

func expandMediaPackageGroupSettings(tfList []any) *types.MediaPackageGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var o types.MediaPackageGroupSettings

	if v, ok := m[names.AttrDestination].([]any); ok && len(v) > 0 {
		o.Destination = expandDestination(v)
	}

	return &o
}

func expandArchiveGroupSettings(tfList []any) *types.ArchiveGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var o types.ArchiveGroupSettings

	if v, ok := m[names.AttrDestination].([]any); ok && len(v) > 0 {
		o.Destination = expandDestination(v)
	}
	if v, ok := m["archive_cdn_settings"].([]any); ok && len(v) > 0 {
		o.ArchiveCdnSettings = expandArchiveCDNSettings(v)
	}
	if v, ok := m["rollover_interval"].(int); ok && v != 0 {
		o.RolloverInterval = aws.Int32(int32(v))
	}

	return &o
}

func expandFrameCaptureGroupSettings(tfList []any) *types.FrameCaptureGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.FrameCaptureGroupSettings
	if v, ok := m[names.AttrDestination].([]any); ok && len(v) > 0 {
		out.Destination = expandDestination(v)
	}
	if v, ok := m["frame_capture_cdn_settings"].([]any); ok && len(v) > 0 {
		out.FrameCaptureCdnSettings = expandFrameCaptureCDNSettings(v)
	}
	return &out
}

func expandFrameCaptureCDNSettings(tfList []any) *types.FrameCaptureCdnSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.FrameCaptureCdnSettings
	if v, ok := m["frame_capture_s3_settings"].([]any); ok && len(v) > 0 {
		out.FrameCaptureS3Settings = expandFrameCaptureS3Settings(v)
	}

	return &out
}

func expandFrameCaptureS3Settings(tfList []any) *types.FrameCaptureS3Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.FrameCaptureS3Settings
	if v, ok := m["canned_acl"].(string); ok && v != "" {
		out.CannedAcl = types.S3CannedAcl(v)
	}

	return &out
}

func expandHLSGroupSettings(tfList []any) *types.HlsGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.HlsGroupSettings
	if v, ok := m[names.AttrDestination].([]any); ok && len(v) > 0 {
		out.Destination = expandDestination(v)
	}
	if v, ok := m["ad_markers"].([]any); ok && len(v) > 0 {
		out.AdMarkers = expandHLSAdMarkers(v)
	}
	if v, ok := m["base_url_content"].(string); ok && v != "" {
		out.BaseUrlContent = aws.String(v)
	}
	if v, ok := m["base_url_content1"].(string); ok && v != "" {
		out.BaseUrlContent1 = aws.String(v)
	}
	if v, ok := m["base_url_manifest"].(string); ok && v != "" {
		out.BaseUrlManifest = aws.String(v)
	}
	if v, ok := m["base_url_manifest1"].(string); ok && v != "" {
		out.BaseUrlManifest1 = aws.String(v)
	}
	if v, ok := m["caption_language_mappings"].(*schema.Set); ok && v.Len() > 0 {
		out.CaptionLanguageMappings = expandHSLGroupSettingsCaptionLanguageMappings(v.List())
	}
	if v, ok := m["caption_language_setting"].(string); ok && v != "" {
		out.CaptionLanguageSetting = types.HlsCaptionLanguageSetting(v)
	}
	if v, ok := m["client_cache"].(string); ok && v != "" {
		out.ClientCache = types.HlsClientCache(v)
	}
	if v, ok := m["codec_specification"].(string); ok && v != "" {
		out.CodecSpecification = types.HlsCodecSpecification(v)
	}
	if v, ok := m["constant_iv"].(string); ok && v != "" {
		out.ConstantIv = aws.String(v)
	}
	if v, ok := m["directory_structure"].(string); ok && v != "" {
		out.DirectoryStructure = types.HlsDirectoryStructure(v)
	}
	if v, ok := m["discontinuity_tags"].(string); ok && v != "" {
		out.DiscontinuityTags = types.HlsDiscontinuityTags(v)
	}
	if v, ok := m["encryption_type"].(string); ok && v != "" {
		out.EncryptionType = types.HlsEncryptionType(v)
	}
	if v, ok := m["hls_cdn_settings"].([]any); ok && len(v) > 0 {
		out.HlsCdnSettings = expandHLSCDNSettings(v)
	}
	if v, ok := m["hls_id3_segment_tagging"].(string); ok && v != "" {
		out.HlsId3SegmentTagging = types.HlsId3SegmentTaggingState(v)
	}
	if v, ok := m["iframe_only_playlists"].(string); ok && v != "" {
		out.IFrameOnlyPlaylists = types.IFrameOnlyPlaylistType(v)
	}
	if v, ok := m["incomplete_segment_behavior"].(string); ok && v != "" {
		out.IncompleteSegmentBehavior = types.HlsIncompleteSegmentBehavior(v)
	}
	if v, ok := m["index_n_segments"].(int); ok && v != 0 {
		out.IndexNSegments = aws.Int32(int32(v))
	}
	if v, ok := m["input_loss_action"].(string); ok && v != "" {
		out.InputLossAction = types.InputLossActionForHlsOut(v)
	}
	if v, ok := m["iv_in_manifest"].(string); ok && v != "" {
		out.IvInManifest = types.HlsIvInManifest(v)
	}
	if v, ok := m["iv_source"].(string); ok && v != "" {
		out.IvSource = types.HlsIvSource(v)
	}
	if v, ok := m["keep_segments"].(int); ok && v != 0 {
		out.KeepSegments = aws.Int32(int32(v))
	}
	if v, ok := m["key_format"].(string); ok && v != "" {
		out.KeyFormat = aws.String(v)
	}
	if v, ok := m["key_format_versions"].(string); ok && v != "" {
		out.KeyFormatVersions = aws.String(v)
	}
	if v, ok := m["key_provider_settings"].([]any); ok && len(v) > 0 {
		out.KeyProviderSettings = expandHLSGroupSettingsKeyProviderSettings(v)
	}
	if v, ok := m["manifest_compression"].(string); ok && v != "" {
		out.ManifestCompression = types.HlsManifestCompression(v)
	}
	if v, ok := m["manifest_duration_format"].(string); ok && v != "" {
		out.ManifestDurationFormat = types.HlsManifestDurationFormat(v)
	}
	if v, ok := m["min_segment_length"].(int); ok && v != 0 {
		out.MinSegmentLength = aws.Int32(int32(v))
	}
	if v, ok := m[names.AttrMode].(string); ok && v != "" {
		out.Mode = types.HlsMode(v)
	}
	if v, ok := m["output_selection"].(string); ok && v != "" {
		out.OutputSelection = types.HlsOutputSelection(v)
	}
	if v, ok := m["program_date_time"].(string); ok && v != "" {
		out.ProgramDateTime = types.HlsProgramDateTime(v)
	}
	if v, ok := m["program_date_time_clock"].(string); ok && v != "" {
		out.ProgramDateTimeClock = types.HlsProgramDateTimeClock(v)
	}
	if v, ok := m["program_date_time_period"].(int); ok && v != 0 {
		out.ProgramDateTimePeriod = aws.Int32(int32(v))
	}
	if v, ok := m["redundant_manifest"].(string); ok && v != "" {
		out.RedundantManifest = types.HlsRedundantManifest(v)
	}
	if v, ok := m["segment_length"].(int); ok && v != 0 {
		out.SegmentLength = aws.Int32(int32(v))
	}
	if v, ok := m["segments_per_subdirectory"].(int); ok && v != 0 {
		out.SegmentsPerSubdirectory = aws.Int32(int32(v))
	}
	if v, ok := m["stream_inf_resolution"].(string); ok && v != "" {
		out.StreamInfResolution = types.HlsStreamInfResolution(v)
	}
	if v, ok := m["timed_metadata_id3_frame"].(string); ok && v != "" {
		out.TimedMetadataId3Frame = types.HlsTimedMetadataId3Frame(v)
	}
	if v, ok := m["timed_metadata_id3_period"].(int); ok && v != 0 {
		out.TimedMetadataId3Period = aws.Int32(int32(v))
	}
	if v, ok := m["timestamp_delta_milliseconds"].(int); ok && v != 0 {
		out.TimestampDeltaMilliseconds = aws.Int32(int32(v))
	}
	if v, ok := m["ts_file_mode"].(string); ok && v != "" {
		out.TsFileMode = types.HlsTsFileMode(v)
	}

	return &out
}

func expandMsSmoothGroupSettings(tfList []any) *types.MsSmoothGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.MsSmoothGroupSettings
	if v, ok := m[names.AttrDestination].([]any); ok && len(v) > 0 {
		out.Destination = expandDestination(v)
	}
	if v, ok := m["acquisition_point_id"].(string); ok && v != "" {
		out.AcquisitionPointId = aws.String(v)
	}
	if v, ok := m["audio_only_timecode_control"].(string); ok && v != "" {
		out.AudioOnlyTimecodeControl = types.SmoothGroupAudioOnlyTimecodeControl(v)
	}
	if v, ok := m["certificate_mode"].(string); ok && v != "" {
		out.CertificateMode = types.SmoothGroupCertificateMode(v)
	}
	if v, ok := m["connection_retry_interval"].(int); ok && v != 0 {
		out.ConnectionRetryInterval = aws.Int32(int32(v))
	}
	if v, ok := m["event_id"].(string); ok && v != "" {
		out.EventId = aws.String(v)
	}
	if v, ok := m["event_id_mode"].(string); ok && v != "" {
		out.EventIdMode = types.SmoothGroupEventIdMode(v)
	}
	if v, ok := m["event_stop_behavior"].(string); ok && v != "" {
		out.EventStopBehavior = types.SmoothGroupEventStopBehavior(v)
	}
	if v, ok := m["filecache_duration"].(int); ok && v != 0 {
		out.FilecacheDuration = aws.Int32(int32(v))
	}
	if v, ok := m["fragment_length"].(int); ok && v != 0 {
		out.FragmentLength = aws.Int32(int32(v))
	}
	if v, ok := m["input_loss_action"].(string); ok && v != "" {
		out.InputLossAction = types.InputLossActionForMsSmoothOut(v)
	}
	if v, ok := m["num_retries"].(int); ok && v != 0 {
		out.NumRetries = aws.Int32(int32(v))
	}
	if v, ok := m["restart_delay"].(int); ok && v != 0 {
		out.RestartDelay = aws.Int32(int32(v))
	}
	if v, ok := m["segmentation_mode"].(string); ok && v != "" {
		out.SegmentationMode = types.SmoothGroupSegmentationMode(v)
	}
	if v, ok := m["send_delay_ms"].(int); ok && v != 0 {
		out.SendDelayMs = aws.Int32(int32(v))
	}
	if v, ok := m["sparse_track_type"].(string); ok && v != "" {
		out.SparseTrackType = types.SmoothGroupSparseTrackType(v)
	}
	if v, ok := m["stream_manifest_behavior"].(string); ok && v != "" {
		out.StreamManifestBehavior = types.SmoothGroupStreamManifestBehavior(v)
	}
	if v, ok := m["timestamp_offset"].(string); ok && v != "" {
		out.TimestampOffset = aws.String(v)
	}
	if v, ok := m["timestamp_offset_mode"].(string); ok && v != "" {
		out.TimestampOffsetMode = types.SmoothGroupTimestampOffsetMode(v)
	}

	return &out
}

func expandHLSCDNSettings(tfList []any) *types.HlsCdnSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.HlsCdnSettings
	if v, ok := m["hls_akamai_settings"].([]any); ok && len(v) > 0 {
		out.HlsAkamaiSettings = expandHSLAkamaiSettings(v)
	}
	if v, ok := m["hls_basic_put_settings"].([]any); ok && len(v) > 0 {
		out.HlsBasicPutSettings = expandHSLBasicPutSettings(v)
	}
	if v, ok := m["hls_media_store_settings"].([]any); ok && len(v) > 0 {
		out.HlsMediaStoreSettings = expandHLSMediaStoreSettings(v)
	}
	if v, ok := m["hls_s3_settings"].([]any); ok && len(v) > 0 {
		out.HlsS3Settings = expandHSLS3Settings(v)
	}
	if v, ok := m["hls_webdav_settings"].([]any); ok && len(v) > 0 {
		out.HlsWebdavSettings = expandHLSWebdavSettings(v)
	}
	return &out
}

func expandHSLAkamaiSettings(tfList []any) *types.HlsAkamaiSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.HlsAkamaiSettings
	if v, ok := m["connection_retry_interval"].(int); ok && v != 0 {
		out.ConnectionRetryInterval = aws.Int32(int32(v))
	}
	if v, ok := m["filecache_duration"].(int); ok && v != 0 {
		out.FilecacheDuration = aws.Int32(int32(v))
	}
	if v, ok := m["http_transfer_mode"].(string); ok && v != "" {
		out.HttpTransferMode = types.HlsAkamaiHttpTransferMode(v)
	}
	if v, ok := m["num_retries"].(int); ok && v != 0 {
		out.NumRetries = aws.Int32(int32(v))
	}
	if v, ok := m["restart_delay"].(int); ok && v != 0 {
		out.RestartDelay = aws.Int32(int32(v))
	}
	if v, ok := m["salt"].(string); ok && v != "" {
		out.Salt = aws.String(v)
	}
	if v, ok := m["token"].(string); ok && v != "" {
		out.Token = aws.String(v)
	}

	return &out
}

func expandHSLBasicPutSettings(tfList []any) *types.HlsBasicPutSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.HlsBasicPutSettings
	if v, ok := m["connection_retry_interval"].(int); ok && v != 0 {
		out.ConnectionRetryInterval = aws.Int32(int32(v))
	}
	if v, ok := m["filecache_duration"].(int); ok && v != 0 {
		out.FilecacheDuration = aws.Int32(int32(v))
	}
	if v, ok := m["num_retries"].(int); ok && v != 0 {
		out.NumRetries = aws.Int32(int32(v))
	}
	if v, ok := m["restart_delay"].(int); ok && v != 0 {
		out.RestartDelay = aws.Int32(int32(v))
	}

	return &out
}

func expandHLSMediaStoreSettings(tfList []any) *types.HlsMediaStoreSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.HlsMediaStoreSettings
	if v, ok := m["connection_retry_interval"].(int); ok && v != 0 {
		out.ConnectionRetryInterval = aws.Int32(int32(v))
	}
	if v, ok := m["filecache_duration"].(int); ok && v != 0 {
		out.FilecacheDuration = aws.Int32(int32(v))
	}
	if v, ok := m["media_store_storage_class"].(string); ok && v != "" {
		out.MediaStoreStorageClass = types.HlsMediaStoreStorageClass(v)
	}
	if v, ok := m["num_retries"].(int); ok && v != 0 {
		out.NumRetries = aws.Int32(int32(v))
	}
	if v, ok := m["restart_delay"].(int); ok && v != 0 {
		out.RestartDelay = aws.Int32(int32(v))
	}

	return &out
}

func expandHSLS3Settings(tfList []any) *types.HlsS3Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.HlsS3Settings
	if v, ok := m["canned_acl"].(string); ok && v != "" {
		out.CannedAcl = types.S3CannedAcl(v)
	}

	return &out
}

func expandHLSWebdavSettings(tfList []any) *types.HlsWebdavSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.HlsWebdavSettings
	if v, ok := m["connection_retry_interval"].(int); ok && v != 0 {
		out.ConnectionRetryInterval = aws.Int32(int32(v))
	}
	if v, ok := m["filecache_duration"].(int); ok && v != 0 {
		out.FilecacheDuration = aws.Int32(int32(v))
	}
	if v, ok := m["http_transfer_mode"].(string); ok && v != "" {
		out.HttpTransferMode = types.HlsWebdavHttpTransferMode(v)
	}
	if v, ok := m["num_retries"].(int); ok && v != 0 {
		out.NumRetries = aws.Int32(int32(v))
	}
	if v, ok := m["restart_delay"].(int); ok && v != 0 {
		out.RestartDelay = aws.Int32(int32(v))
	}
	return &out
}

func expandHSLGroupSettingsCaptionLanguageMappings(tfList []any) []types.CaptionLanguageMapping {
	if tfList == nil {
		return nil
	}

	var out []types.CaptionLanguageMapping
	for _, item := range tfList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		var o types.CaptionLanguageMapping
		if v, ok := m["caption_channel"].(int); ok && v != 0 {
			o.CaptionChannel = aws.Int32(int32(v))
		}
		if v, ok := m[names.AttrLanguageCode].(string); ok && v != "" {
			o.LanguageCode = aws.String(v)
		}
		if v, ok := m["language_description"].(string); ok && v != "" {
			o.LanguageDescription = aws.String(v)
		}

		out = append(out, o)
	}

	return out
}

func expandHLSGroupSettingsKeyProviderSettings(tfList []any) *types.KeyProviderSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.KeyProviderSettings
	if v, ok := m["static_key_settings"].([]any); ok && len(v) > 0 {
		out.StaticKeySettings = expandKeyProviderSettingsStaticKeySettings(v)
	}

	return &out
}

func expandKeyProviderSettingsStaticKeySettings(tfList []any) *types.StaticKeySettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.StaticKeySettings
	if v, ok := m["static_key_value"].(string); ok && v != "" {
		out.StaticKeyValue = aws.String(v)
	}
	if v, ok := m["key_provider_server"].([]any); ok && len(v) > 0 {
		out.KeyProviderServer = expandInputLocation(v)
	}

	return &out
}

func expandInputLocation(tfList []any) *types.InputLocation {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.InputLocation
	if v, ok := m[names.AttrURI].(string); ok && v != "" {
		out.Uri = aws.String(v)
	}
	if v, ok := m["password_param"].(string); ok && v != "" {
		out.PasswordParam = aws.String(v)
	}
	if v, ok := m[names.AttrUsername].(string); ok && v != "" {
		out.Username = aws.String(v)
	}

	return &out
}

func expandArchiveCDNSettings(tfList []any) *types.ArchiveCdnSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.ArchiveCdnSettings
	if v, ok := m["archive_s3_settings"].([]any); ok && len(v) > 0 {
		out.ArchiveS3Settings = func(in []any) *types.ArchiveS3Settings {
			if len(in) == 0 {
				return nil
			}

			m := in[0].(map[string]any)

			var o types.ArchiveS3Settings
			if v, ok := m["canned_acl"].(string); ok && v != "" {
				o.CannedAcl = types.S3CannedAcl(v)
			}

			return &o
		}(v)
	}

	return &out
}

func expandAudioWatermarkSettings(tfList []any) *types.AudioWatermarkSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var o types.AudioWatermarkSettings
	if v, ok := m["nielsen_watermark_settings"].([]any); ok && len(v) > 0 {
		o.NielsenWatermarksSettings = func(n []any) *types.NielsenWatermarksSettings {
			if len(n) == 0 {
				return nil
			}

			inner := n[0].(map[string]any)

			var ns types.NielsenWatermarksSettings
			if v, ok := inner["nielsen_distribution_type"].(string); ok && v != "" {
				ns.NielsenDistributionType = types.NielsenWatermarksDistributionTypes(v)
			}
			if v, ok := inner["nielsen_cbet_settings"].([]any); ok && len(v) > 0 {
				ns.NielsenCbetSettings = expandNielsenCbetSettings(v)
			}
			if v, ok := inner["nielsen_naes_ii_nw_settings"].([]any); ok && len(v) > 0 {
				ns.NielsenNaesIiNwSettings = expandNielsenNaseIiNwSettings(v)
			}

			return &ns
		}(v)
	}

	return &o
}

func expandRtmpGroupSettings(tfList []any) *types.RtmpGroupSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.RtmpGroupSettings
	if v, ok := m["ad_markers"].([]any); ok && len(v) > 0 {
		out.AdMarkers = expandRTMPAdMarkers(v)
	}
	if v, ok := m["authentication_scheme"].(string); ok && v != "" {
		out.AuthenticationScheme = types.AuthenticationScheme(v)
	}
	if v, ok := m["cache_full_behavior"].(string); ok && v != "" {
		out.CacheFullBehavior = types.RtmpCacheFullBehavior(v)
	}
	if v, ok := m["cache_length"].(int); ok && v != 0 {
		out.CacheLength = aws.Int32(int32(v))
	}
	if v, ok := m["caption_data"].(string); ok && v != "" {
		out.CaptionData = types.RtmpCaptionData(v)
	}
	if v, ok := m["input_loss_action"].(string); ok && v != "" {
		out.InputLossAction = types.InputLossActionForRtmpOut(v)
	}
	if v, ok := m["restart_delay"].(int); ok && v != 0 {
		out.RestartDelay = aws.Int32(int32(v))
	}

	return &out
}

func expandUdpGroupSettings(tfList []any) *types.UdpGroupSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.UdpGroupSettings
	if v, ok := m["input_loss_action"].(string); ok && v != "" {
		out.InputLossAction = types.InputLossActionForUdpOut(v)
	}
	if v, ok := m["timed_metadata_id3_frame"].(string); ok && v != "" {
		out.TimedMetadataId3Frame = types.UdpTimedMetadataId3Frame(v)
	}
	if v, ok := m["timed_metadata_id3_period"].(int); ok && v != 0 {
		out.TimedMetadataId3Period = aws.Int32(int32(v))
	}

	return &out
}

func expandRTMPAdMarkers(tfList []any) []types.RtmpAdMarkers {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.RtmpAdMarkers
	for _, v := range tfList {
		out = append(out, types.RtmpAdMarkers(v.(string)))
	}

	return out
}

func expandHLSAdMarkers(tfList []any) []types.HlsAdMarkers {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.HlsAdMarkers
	for _, v := range tfList {
		out = append(out, types.HlsAdMarkers(v.(string)))
	}

	return out
}

func expandChannelEncoderSettingsOutputGroupsOutputs(tfList []any) []types.Output {
	if tfList == nil {
		return nil
	}

	var outputs []types.Output
	for _, item := range tfList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		var o types.Output
		if v, ok := m["output_settings"].([]any); ok && len(v) > 0 {
			o.OutputSettings = expandOutputsOutputSettings(v)
		}
		if v, ok := m["audio_description_names"].(*schema.Set); ok && v.Len() > 0 {
			o.AudioDescriptionNames = flex.ExpandStringValueSet(v)
		}
		if v, ok := m["caption_description_names"].(*schema.Set); ok && v.Len() > 0 {
			o.CaptionDescriptionNames = flex.ExpandStringValueSet(v)
		}
		if v, ok := m["output_name"].(string); ok && v != "" {
			o.OutputName = aws.String(v)
		}
		if v, ok := m["video_description_name"].(string); ok && v != "" {
			o.VideoDescriptionName = aws.String(v)
		}
		outputs = append(outputs, o)
	}

	return outputs
}

func expandOutputsOutputSettings(tfList []any) *types.OutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var os types.OutputSettings
	if v, ok := m["archive_output_settings"].([]any); ok && len(v) > 0 {
		os.ArchiveOutputSettings = expandOutputsOutputSettingsArchiveOutputSettings(v)
	}
	if v, ok := m["frame_capture_output_settings"].([]any); ok && len(v) > 0 {
		os.FrameCaptureOutputSettings = expandOutputsOutSettingsFrameCaptureOutputSettings(v)
	}
	if v, ok := m["hls_output_settings"].([]any); ok && len(v) > 0 {
		os.HlsOutputSettings = expandOutputsOutSettingsHLSOutputSettings(v)
	}
	if v, ok := m["media_package_output_settings"].([]any); ok && len(v) > 0 {
		os.MediaPackageOutputSettings = &types.MediaPackageOutputSettings{} // no exported fields
	}
	if v, ok := m["ms_smooth_output_settings"].([]any); ok && len(v) > 0 {
		os.MsSmoothOutputSettings = expandOutputsOutSettingsMsSmoothOutputSettings(v)
	}
	if v, ok := m["multiplex_output_settings"].([]any); ok && len(v) > 0 {
		os.MultiplexOutputSettings = func(inner []any) *types.MultiplexOutputSettings {
			if len(inner) == 0 {
				return nil
			}

			data := inner[0].(map[string]any)
			var mos types.MultiplexOutputSettings
			if v, ok := data[names.AttrDestination].([]any); ok && len(v) > 0 {
				mos.Destination = expandDestination(v)
			}
			return &mos
		}(v)
	}

	if v, ok := m["rtmp_output_settings"].([]any); ok && len(v) > 0 {
		os.RtmpOutputSettings = expandOutputsOutputSettingsRtmpOutputSettings(v)
	}
	if v, ok := m["udp_output_settings"].([]any); ok && len(v) > 0 {
		os.UdpOutputSettings = expandOutputsOutputSettingsUdpOutputSettings(v)
	}

	return &os
}

func expandOutputsOutputSettingsArchiveOutputSettings(tfList []any) *types.ArchiveOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var settings types.ArchiveOutputSettings
	if v, ok := m["container_settings"].([]any); ok && len(v) > 0 {
		settings.ContainerSettings = expandOutputsOutputSettingsArchiveSettingsContainerSettings(v)
	}
	if v, ok := m["extension"].(string); ok && v != "" {
		settings.Extension = aws.String(v)
	}
	if v, ok := m["name_modifier"].(string); ok && v != "" {
		settings.NameModifier = aws.String(v)
	}
	return &settings
}

func expandOutputsOutSettingsFrameCaptureOutputSettings(tfList []any) *types.FrameCaptureOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.FrameCaptureOutputSettings
	if v, ok := m["name_modifier"].(string); ok && v != "" {
		out.NameModifier = aws.String(v)
	}

	return &out
}

func expandOutputsOutSettingsHLSOutputSettings(tfList []any) *types.HlsOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.HlsOutputSettings
	if v, ok := m["hls_settings"].([]any); ok && len(v) > 0 {
		out.HlsSettings = expandHLSOutputSettingsHLSSettings(v)
	}
	if v, ok := m["h265_packaging_type"].(string); ok && v != "" {
		out.H265PackagingType = types.HlsH265PackagingType(v)
	}
	if v, ok := m["name_modifier"].(string); ok && v != "" {
		out.NameModifier = aws.String(v)
	}
	if v, ok := m["segment_modifier"].(string); ok && v != "" {
		out.SegmentModifier = aws.String(v)
	}

	return &out
}

func expandOutputsOutSettingsMsSmoothOutputSettings(tfList []any) *types.MsSmoothOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.MsSmoothOutputSettings
	if v, ok := m["h265_packaging_type"].(string); ok && v != "" {
		out.H265PackagingType = types.MsSmoothH265PackagingType(v)
	}
	if v, ok := m["name_modifier"].(string); ok && v != "" {
		out.NameModifier = aws.String(v)
	}

	return &out
}

func expandHLSOutputSettingsHLSSettings(tfList []any) *types.HlsSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.HlsSettings
	if v, ok := m["audio_only_hls_settings"].([]any); ok && len(v) > 0 {
		out.AudioOnlyHlsSettings = expandHLSSettingsAudioOnlyHLSSettings(v)
	}
	if v, ok := m["fmp4_hls_settings"].([]any); ok && len(v) > 0 {
		out.Fmp4HlsSettings = expandHLSSettingsFmp4HLSSettings(v)
	}
	if v, ok := m["frame_capture_hls_settings"].([]any); ok && len(v) > 0 {
		out.FrameCaptureHlsSettings = &types.FrameCaptureHlsSettings{} // no exported types
	}
	if v, ok := m["standard_hls_settings"].([]any); ok && len(v) > 0 {
		out.StandardHlsSettings = expandHLSSettingsStandardHLSSettings(v)
	}

	return &out
}

func expandHLSSettingsAudioOnlyHLSSettings(tfList []any) *types.AudioOnlyHlsSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.AudioOnlyHlsSettings
	if v, ok := m["audio_group_id"].(string); ok && v != "" {
		out.AudioGroupId = aws.String(v)
	}
	if v, ok := m["audio_only_image"].([]any); ok && len(v) > 0 {
		out.AudioOnlyImage = expandInputLocation(v)
	}
	if v, ok := m["audio_track_type"].(string); ok && v != "" {
		out.AudioTrackType = types.AudioOnlyHlsTrackType(v)
	}
	if v, ok := m["segment_type"].(string); ok && v != "" {
		out.SegmentType = types.AudioOnlyHlsSegmentType(v)
	}

	return &out
}

func expandHLSSettingsFmp4HLSSettings(tfList []any) *types.Fmp4HlsSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.Fmp4HlsSettings
	if v, ok := m["audio_rendition_sets"].(string); ok && v != "" {
		out.AudioRenditionSets = aws.String(v)
	}
	if v, ok := m["segment_type"].(string); ok && v != "" {
		out.NielsenId3Behavior = types.Fmp4NielsenId3Behavior(v)
	}
	if v, ok := m["timed_metadata_behavior"].(string); ok && v != "" {
		out.TimedMetadataBehavior = types.Fmp4TimedMetadataBehavior(v)
	}

	return &out
}

func expandHLSSettingsStandardHLSSettings(tfList []any) *types.StandardHlsSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.StandardHlsSettings
	if v, ok := m["m3u8_settings"].([]any); ok && len(v) > 0 {
		out.M3u8Settings = expandStandardHLSSettingsH3u8Settings(v)
	}
	if v, ok := m["audio_rendition_sets"].(string); ok && v != "" {
		out.AudioRenditionSets = aws.String(v)
	}

	return &out
}

func expandStandardHLSSettingsH3u8Settings(tfList []any) *types.M3u8Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.M3u8Settings
	if v, ok := m["audio_frames_per_pes"].(int); ok && v != 0 {
		out.AudioFramesPerPes = aws.Int32(int32(v))
	}
	if v, ok := m["audio_pids"].(string); ok && v != "" {
		out.AudioPids = aws.String(v)
	}
	if v, ok := m["ecm_pid"].(string); ok && v != "" {
		out.EcmPid = aws.String(v)
	}
	if v, ok := m["nielsen_id3_behavior"].(string); ok && v != "" {
		out.NielsenId3Behavior = types.M3u8NielsenId3Behavior(v)
	}
	if v, ok := m["pat_interval"].(int); ok && v != 0 {
		out.PatInterval = aws.Int32(int32(v))
	}
	if v, ok := m["pcr_control"].(string); ok && v != "" {
		out.PcrControl = types.M3u8PcrControl(v)
	}
	if v, ok := m["pcr_period"].(int); ok && v != 0 {
		out.PcrPeriod = aws.Int32(int32(v))
	}
	if v, ok := m["pcr_pid"].(string); ok && v != "" {
		out.PcrPid = aws.String(v)
	}
	if v, ok := m["pmt_interval"].(int); ok && v != 0 {
		out.PmtInterval = aws.Int32(int32(v))
	}
	if v, ok := m["pmt_pid"].(string); ok && v != "" {
		out.PmtPid = aws.String(v)
	}
	if v, ok := m["program_num"].(int); ok && v != 0 {
		out.ProgramNum = aws.Int32(int32(v))
	}
	if v, ok := m["scte35_behavior"].(string); ok && v != "" {
		out.Scte35Behavior = types.M3u8Scte35Behavior(v)
	}
	if v, ok := m["scte35_pid"].(string); ok && v != "" {
		out.Scte35Pid = aws.String(v)
	}
	if v, ok := m["timed_metadata_behavior"].(string); ok && v != "" {
		out.TimedMetadataBehavior = types.M3u8TimedMetadataBehavior(v)
	}
	if v, ok := m["timed_metadata_pid"].(string); ok && v != "" {
		out.TimedMetadataPid = aws.String(v)
	}
	if v, ok := m["transport_stream_id"].(int); ok && v != 0 {
		out.TransportStreamId = aws.Int32(int32(v))
	}
	if v, ok := m["video_pid"].(string); ok && v != "" {
		out.VideoPid = aws.String(v)
	}

	return &out
}

func expandOutputsOutputSettingsRtmpOutputSettings(tfList []any) *types.RtmpOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var settings types.RtmpOutputSettings
	if v, ok := m[names.AttrDestination].([]any); ok && len(v) > 0 {
		settings.Destination = expandDestination(v)
	}
	if v, ok := m["certificate_mode"].(string); ok && v != "" {
		settings.CertificateMode = types.RtmpOutputCertificateMode(v)
	}
	if v, ok := m["connection_retry_interval"].(int); ok && v != 0 {
		settings.ConnectionRetryInterval = aws.Int32(int32(v))
	}
	if v, ok := m["num_retries"].(int); ok && v != 0 {
		settings.NumRetries = aws.Int32(int32(v))
	}

	return &settings
}

func expandOutputsOutputSettingsUdpOutputSettings(tfList []any) *types.UdpOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var settings types.UdpOutputSettings
	if v, ok := m["container_settings"].([]any); ok && len(v) > 0 {
		settings.ContainerSettings = expandOutputsOutputSettingsUdpSettingsContainerSettings(v)
	}
	if v, ok := m[names.AttrDestination].([]any); ok && len(v) > 0 {
		settings.Destination = expandDestination(v)
	}
	if v, ok := m["buffer_msec"].(int); ok && v != 0 {
		settings.BufferMsec = aws.Int32(int32(v))
	}
	if v, ok := m["fec_output_settings"].([]any); ok && len(v) > 0 {
		settings.FecOutputSettings = expandFecOutputSettings(v)
	}

	return &settings
}

func expandOutputsOutputSettingsArchiveSettingsContainerSettings(tfList []any) *types.ArchiveContainerSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var settings types.ArchiveContainerSettings
	if v, ok := m["m2ts_settings"].([]any); ok && len(v) > 0 {
		settings.M2tsSettings = expandM2tsSettings(v)
	}

	if v, ok := m["raw_settings"].([]any); ok && len(v) > 0 {
		settings.RawSettings = &types.RawSettings{}
	}
	return &settings
}

func expandOutputsOutputSettingsUdpSettingsContainerSettings(tfList []any) *types.UdpContainerSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var settings types.UdpContainerSettings
	if v, ok := m["m2ts_settings"].([]any); ok && len(v) > 0 {
		settings.M2tsSettings = expandM2tsSettings(v)
	}

	return &settings
}

func expandFecOutputSettings(tfList []any) *types.FecOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var settings types.FecOutputSettings
	if v, ok := m["column_depth"].(int); ok && v != 0 {
		settings.ColumnDepth = aws.Int32(int32(v))
	}
	if v, ok := m["include_fec"].(string); ok && v != "" {
		settings.IncludeFec = types.FecOutputIncludeFec(v)
	}
	if v, ok := m["row_length"].(int); ok && v != 0 {
		settings.RowLength = aws.Int32(int32(v))
	}

	return &settings
}

func expandM2tsSettings(tfList []any) *types.M2tsSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var s types.M2tsSettings
	if v, ok := m["absent_input_audio_behavior"].(string); ok && v != "" {
		s.AbsentInputAudioBehavior = types.M2tsAbsentInputAudioBehavior(v)
	}
	if v, ok := m["arib"].(string); ok && v != "" {
		s.Arib = types.M2tsArib(v)
	}
	if v, ok := m["arib_captions_pid"].(string); ok && v != "" {
		s.AribCaptionsPid = aws.String(v)
	}
	if v, ok := m["arib_captions_pid_control"].(string); ok && v != "" {
		s.AribCaptionsPidControl = types.M2tsAribCaptionsPidControl(v)
	}
	if v, ok := m["audio_buffer_model"].(string); ok && v != "" {
		s.AudioBufferModel = types.M2tsAudioBufferModel(v)
	}
	if v, ok := m["audio_frames_per_pes"].(int); ok && v != 0 {
		s.AudioFramesPerPes = aws.Int32(int32(v))
	}
	if v, ok := m["audio_pids"].(string); ok && v != "" {
		s.AudioPids = aws.String(v)
	}
	if v, ok := m["audio_stream_type"].(string); ok && v != "" {
		s.AudioStreamType = types.M2tsAudioStreamType(v)
	}
	if v, ok := m["bitrate"].(int); ok && v != 0 {
		s.Bitrate = aws.Int32(int32(v))
	}
	if v, ok := m["buffer_model"].(string); ok && v != "" {
		s.BufferModel = types.M2tsBufferModel(v)
	}
	if v, ok := m["cc_descriptor"].(string); ok && v != "" {
		s.CcDescriptor = types.M2tsCcDescriptor(v)
	}
	if v, ok := m["dvb_nit_settings"].([]any); ok && len(v) > 0 {
		s.DvbNitSettings = expandM2tsDvbNitSettings(v)
	}
	if v, ok := m["dvb_sdt_settings"].([]any); ok && len(v) > 0 {
		s.DvbSdtSettings = expandM2tsDvbSdtSettings(v)
	}
	if v, ok := m["dvb_sub_pids"].(string); ok && v != "" {
		s.DvbSubPids = aws.String(v)
	}
	if v, ok := m["dvb_tdt_settings"].([]any); ok && len(v) > 0 {
		s.DvbTdtSettings = func(tfList []any) *types.DvbTdtSettings {
			if tfList == nil {
				return nil
			}

			m := tfList[0].(map[string]any)

			var s types.DvbTdtSettings
			if v, ok := m["rep_interval"].(int); ok && v != 0 {
				s.RepInterval = aws.Int32(int32(v))
			}
			return &s
		}(v)
	}
	if v, ok := m["dvb_teletext_pid"].(string); ok && v != "" {
		s.DvbTeletextPid = aws.String(v)
	}
	if v, ok := m["ebif"].(string); ok && v != "" {
		s.Ebif = types.M2tsEbifControl(v)
	}
	if v, ok := m["ebp_audio_interval"].(string); ok && v != "" {
		s.EbpAudioInterval = types.M2tsAudioInterval(v)
	}
	if v, ok := m["ebp_lookahead_ms"].(int); ok && v != 0 {
		s.EbpLookaheadMs = aws.Int32(int32(v))
	}
	if v, ok := m["ebp_placement"].(string); ok && v != "" {
		s.EbpPlacement = types.M2tsEbpPlacement(v)
	}
	if v, ok := m["ecm_pid"].(string); ok && v != "" {
		s.EcmPid = aws.String(v)
	}
	if v, ok := m["es_rate_in_pes"].(string); ok && v != "" {
		s.EsRateInPes = types.M2tsEsRateInPes(v)
	}
	if v, ok := m["etv_platform_pid"].(string); ok && v != "" {
		s.EtvPlatformPid = aws.String(v)
	}
	if v, ok := m["etv_signal_pid"].(string); ok && v != "" {
		s.EtvSignalPid = aws.String(v)
	}
	if v, ok := m["fragment_time"].(float64); ok && v != 0.0 {
		s.FragmentTime = aws.Float64(v)
	}
	if v, ok := m["klv"].(string); ok && v != "" {
		s.Klv = types.M2tsKlv(v)
	}
	if v, ok := m["klv_data_pids"].(string); ok && v != "" {
		s.KlvDataPids = aws.String(v)
	}
	if v, ok := m["nielsen_id3_behavior"].(string); ok && v != "" {
		s.NielsenId3Behavior = types.M2tsNielsenId3Behavior(v)
	}
	if v, ok := m["null_packet_bitrate"].(float32); ok && v != 0.0 {
		s.NullPacketBitrate = aws.Float64(float64(v))
	}
	if v, ok := m["pat_interval"].(int); ok && v != 0 {
		s.PatInterval = aws.Int32(int32(v))
	}
	if v, ok := m["pcr_control"].(string); ok && v != "" {
		s.PcrControl = types.M2tsPcrControl(v)
	}
	if v, ok := m["pcr_period"].(int); ok && v != 0 {
		s.PcrPeriod = aws.Int32(int32(v))
	}
	if v, ok := m["pcr_pid"].(string); ok && v != "" {
		s.PcrPid = aws.String(v)
	}
	if v, ok := m["pmt_interval"].(int); ok && v != 0 {
		s.PmtInterval = aws.Int32(int32(v))
	}
	if v, ok := m["pmt_pid"].(string); ok && v != "" {
		s.PmtPid = aws.String(v)
	}
	if v, ok := m["program_num"].(int); ok && v != 0 {
		s.ProgramNum = aws.Int32(int32(v))
	}
	if v, ok := m["rate_mode"].(string); ok && v != "" {
		s.RateMode = types.M2tsRateMode(v)
	}
	if v, ok := m["scte27_pids"].(string); ok && v != "" {
		s.Scte27Pids = aws.String(v)
	}
	if v, ok := m["scte35_control"].(string); ok && v != "" {
		s.Scte35Control = types.M2tsScte35Control(v)
	}
	if v, ok := m["scte35_pid"].(string); ok && v != "" {
		s.Scte35Pid = aws.String(v)
	}
	if v, ok := m["segmentation_markers"].(string); ok && v != "" {
		s.SegmentationMarkers = types.M2tsSegmentationMarkers(v)
	}
	if v, ok := m["segmentation_style"].(string); ok && v != "" {
		s.SegmentationStyle = types.M2tsSegmentationStyle(v)
	}
	if v, ok := m["segmentation_time"].(float64); ok && v != 0.0 {
		s.SegmentationTime = aws.Float64(v)
	}
	if v, ok := m["timed_metadata_behavior"].(string); ok && v != "" {
		s.TimedMetadataBehavior = types.M2tsTimedMetadataBehavior(v)
	}
	if v, ok := m["timed_metadata_pid"].(string); ok && v != "" {
		s.TimedMetadataPid = aws.String(v)
	}
	if v, ok := m["transport_stream_id"].(int); ok && v != 0 {
		s.TransportStreamId = aws.Int32(int32(v))
	}
	if v, ok := m["video_pid"].(string); ok && v != "" {
		s.VideoPid = aws.String(v)
	}

	return &s
}

func expandM2tsDvbNitSettings(tfList []any) *types.DvbNitSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var s types.DvbNitSettings
	if v, ok := m["network_ids"].(int); ok && v != 0 {
		s.NetworkId = aws.Int32(int32(v))
	}
	if v, ok := m["network_name"].(string); ok && v != "" {
		s.NetworkName = aws.String(v)
	}
	if v, ok := m["network_ids"].(int); ok && v != 0 {
		s.RepInterval = aws.Int32(int32(v))
	}
	return &s
}

func expandM2tsDvbSdtSettings(tfList []any) *types.DvbSdtSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var s types.DvbSdtSettings
	if v, ok := m["output_sdt"].(string); ok && v != "" {
		s.OutputSdt = types.DvbSdtOutputSdt(v)
	}
	if v, ok := m["rep_interval"].(int); ok && v != 0 {
		s.RepInterval = aws.Int32(int32(v))
	}
	if v, ok := m[names.AttrServiceName].(string); ok && v != "" {
		s.ServiceName = aws.String(v)
	}
	if v, ok := m["service_provider_name"].(string); ok && v != "" {
		s.ServiceProviderName = aws.String(v)
	}

	return &s
}

func expandChannelEncoderSettingsTimecodeConfig(tfList []any) *types.TimecodeConfig {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]any)

	var config types.TimecodeConfig
	if v, ok := m[names.AttrSource].(string); ok && v != "" {
		config.Source = types.TimecodeConfigSource(v)
	}
	if v, ok := m["sync_threshold"].(int); ok && v != 0 {
		config.SyncThreshold = aws.Int32(int32(v))
	}

	return &config
}

func expandChannelEncoderSettingsVideoDescriptions(tfList []any) []types.VideoDescription {
	if tfList == nil {
		return nil
	}

	var videoDesc []types.VideoDescription
	for _, tfItem := range tfList {
		m, ok := tfItem.(map[string]any)
		if !ok {
			continue
		}

		var d types.VideoDescription
		if v, ok := m[names.AttrName].(string); ok && v != "" {
			d.Name = aws.String(v)
		}
		if v, ok := m["codec_settings"].([]any); ok && len(v) > 0 {
			d.CodecSettings = expandChannelEncoderSettingsVideoDescriptionsCodecSettings(v)
		}
		if v, ok := m["height"].(int); ok && v != 0 {
			d.Height = aws.Int32(int32(v))
		}
		if v, ok := m["respond_to_afd"].(string); ok && v != "" {
			d.RespondToAfd = types.VideoDescriptionRespondToAfd(v)
		}
		if v, ok := m["scaling_behavior"].(string); ok && v != "" {
			d.ScalingBehavior = types.VideoDescriptionScalingBehavior(v)
		}
		if v, ok := m["sharpness"].(int); ok && v != 0 {
			d.Sharpness = aws.Int32(int32(v))
		}
		if v, ok := m["width"].(int); ok && v != 0 {
			d.Width = aws.Int32(int32(v))
		}

		videoDesc = append(videoDesc, d)
	}

	return videoDesc
}

func expandChannelEncoderSettingsAvailBlanking(tfList []any) *types.AvailBlanking {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.AvailBlanking
	if v, ok := m["avail_blanking_image"].([]any); ok && len(v) > 0 {
		out.AvailBlankingImage = expandInputLocation(v)
	}
	if v, ok := m[names.AttrState].(string); ok && v != "" {
		out.State = types.AvailBlankingState(v)
	}

	return &out
}

func expandChannelEncoderSettingsCaptionDescriptions(tfList []any) []types.CaptionDescription {
	if tfList == nil {
		return nil
	}

	var captionDesc []types.CaptionDescription
	for _, tfItem := range tfList {
		m, ok := tfItem.(map[string]any)
		if !ok {
			continue
		}

		var d types.CaptionDescription
		if v, ok := m["caption_selector_name"].(string); ok && v != "" {
			d.CaptionSelectorName = aws.String(v)
		}
		if v, ok := m[names.AttrName].(string); ok && v != "" {
			d.Name = aws.String(v)
		}
		if v, ok := m["accessibility"].(string); ok && v != "" {
			d.Accessibility = types.AccessibilityType(v)
		}
		if v, ok := m["destination_settings"].([]any); ok && len(v) > 0 {
			d.DestinationSettings = expandChannelEncoderSettingsCaptionDescriptionsDestinationSettings(v)
		}
		if v, ok := m[names.AttrLanguageCode].(string); ok && v != "" {
			d.LanguageCode = aws.String(v)
		}
		if v, ok := m["language_description"].(string); ok && v != "" {
			d.LanguageDescription = aws.String(v)
		}

		captionDesc = append(captionDesc, d)
	}

	return captionDesc
}

func expandChannelEncoderSettingsCaptionDescriptionsDestinationSettings(tfList []any) *types.CaptionDestinationSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.CaptionDestinationSettings
	if v, ok := m["arib_destination_settings"].([]any); ok && len(v) > 0 {
		out.AribDestinationSettings = &types.AribDestinationSettings{} // only unexported fields
	}
	if v, ok := m["burn_in_destination_settings"].([]any); ok && len(v) > 0 {
		out.BurnInDestinationSettings = expandsCaptionDescriptionsDestinationSettingsBurnInDestinationSettings(v)
	}
	if v, ok := m["dvb_sub_destination_settings"].([]any); ok && len(v) > 0 {
		out.DvbSubDestinationSettings = expandsCaptionDescriptionsDestinationSettingsDvbSubDestinationSettings(v)
	}
	if v, ok := m["ebu_tt_d_destination_settings"].([]any); ok && len(v) > 0 {
		out.EbuTtDDestinationSettings = expandsCaptionDescriptionsDestinationSettingsEbuTtDDestinationSettings(v)
	}
	if v, ok := m["embedded_destination_settings"].([]any); ok && len(v) > 0 {
		out.EmbeddedDestinationSettings = &types.EmbeddedDestinationSettings{} // only unexported fields
	}
	if v, ok := m["embedded_plus_scte20_destination_settings"].([]any); ok && len(v) > 0 {
		out.EmbeddedPlusScte20DestinationSettings = &types.EmbeddedPlusScte20DestinationSettings{} // only unexported fields
	}
	if v, ok := m["rtmp_caption_info_destination_settings"].([]any); ok && len(v) > 0 {
		out.RtmpCaptionInfoDestinationSettings = &types.RtmpCaptionInfoDestinationSettings{} // only unexported fields
	}
	if v, ok := m["scte20_plus_embedded_destination_settings"].([]any); ok && len(v) > 0 {
		out.Scte20PlusEmbeddedDestinationSettings = &types.Scte20PlusEmbeddedDestinationSettings{} // only unexported fields
	}
	if v, ok := m["scte27_destination_settings"].([]any); ok && len(v) > 0 {
		out.Scte27DestinationSettings = &types.Scte27DestinationSettings{} // only unexported fields
	}
	if v, ok := m["smpte_tt_destination_settings"].([]any); ok && len(v) > 0 {
		out.SmpteTtDestinationSettings = &types.SmpteTtDestinationSettings{} // only unexported fields
	}
	if v, ok := m["teletext_destination_settings"].([]any); ok && len(v) > 0 {
		out.TeletextDestinationSettings = expandsCaptionDescriptionsDestinationSettingsTeletextDestinationSettings(v)
	}
	if v, ok := m["ttml_destination_settings"].([]any); ok && len(v) > 0 {
		out.TtmlDestinationSettings = expandsCaptionDescriptionsDestinationSettingsTtmlDestinationSettings(v)
	}
	if v, ok := m["webvtt_destination_settings"].([]any); ok && len(v) > 0 {
		out.WebvttDestinationSettings = expandsCaptionDescriptionsDestinationSettingsWebvttDestinationSettings(v)
	}

	return &out
}

func expandsCaptionDescriptionsDestinationSettingsBurnInDestinationSettings(tfList []any) *types.BurnInDestinationSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.BurnInDestinationSettings
	if v, ok := m["alignment"].(string); ok && len(v) > 0 {
		out.Alignment = types.BurnInAlignment(v)
	}
	if v, ok := m["background_color"].(string); ok && len(v) > 0 {
		out.BackgroundColor = types.BurnInBackgroundColor(v)
	}
	if v, ok := m["background_opacity"].(int); ok && v != 0 {
		out.BackgroundOpacity = aws.Int32(int32(v))
	}
	if v, ok := m["font"].([]any); ok && len(v) > 0 {
		out.Font = expandInputLocation(v)
	}
	if v, ok := m["font_color"].(string); ok && len(v) > 0 {
		out.FontColor = types.BurnInFontColor(v)
	}
	if v, ok := m["font_opacity"].(int); ok && v != 0 {
		out.FontOpacity = aws.Int32(int32(v))
	}
	if v, ok := m["font_resolution"].(int); ok && v != 0 {
		out.FontResolution = aws.Int32(int32(v))
	}
	if v, ok := m["font_size"].(string); ok && v != "" {
		out.FontSize = aws.String(v)
	}
	if v, ok := m["outline_color"].(string); ok && len(v) > 0 {
		out.OutlineColor = types.BurnInOutlineColor(v)
	}
	if v, ok := m["outline_size"].(int); ok && v != 0 {
		out.OutlineSize = aws.Int32(int32(v))
	}
	if v, ok := m["shadow_color"].(string); ok && len(v) > 0 {
		out.ShadowColor = types.BurnInShadowColor(v)
	}
	if v, ok := m["shadow_opacity"].(int); ok && v != 0 {
		out.ShadowOpacity = aws.Int32(int32(v))
	}
	if v, ok := m["shadow_x_offset"].(int); ok && v != 0 {
		out.ShadowXOffset = aws.Int32(int32(v))
	}
	if v, ok := m["shadow_y_offset"].(int); ok && v != 0 {
		out.ShadowYOffset = aws.Int32(int32(v))
	}
	if v, ok := m["teletext_grid_control"].(string); ok && len(v) > 0 {
		out.TeletextGridControl = types.BurnInTeletextGridControl(v)
	}
	if v, ok := m["x_position"].(int); ok && v != 0 {
		out.XPosition = aws.Int32(int32(v))
	}
	if v, ok := m["y_position"].(int); ok && v != 0 {
		out.YPosition = aws.Int32(int32(v))
	}

	return &out
}

func expandsCaptionDescriptionsDestinationSettingsDvbSubDestinationSettings(tfList []any) *types.DvbSubDestinationSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.DvbSubDestinationSettings
	if v, ok := m["alignment"].(string); ok && len(v) > 0 {
		out.Alignment = types.DvbSubDestinationAlignment(v)
	}
	if v, ok := m["background_color"].(string); ok && len(v) > 0 {
		out.BackgroundColor = types.DvbSubDestinationBackgroundColor(v)
	}
	if v, ok := m["background_opacity"].(int); ok && v != 0 {
		out.BackgroundOpacity = aws.Int32(int32(v))
	}
	if v, ok := m["font"].([]any); ok && len(v) > 0 {
		out.Font = expandInputLocation(v)
	}
	if v, ok := m["font_color"].(string); ok && len(v) > 0 {
		out.FontColor = types.DvbSubDestinationFontColor(v)
	}
	if v, ok := m["font_opacity"].(int); ok && v != 0 {
		out.FontOpacity = aws.Int32(int32(v))
	}
	if v, ok := m["font_resolution"].(int); ok && v != 0 {
		out.FontResolution = aws.Int32(int32(v))
	}
	if v, ok := m["font_size"].(string); ok && v != "" {
		out.FontSize = aws.String(v)
	}
	if v, ok := m["outline_color"].(string); ok && len(v) > 0 {
		out.OutlineColor = types.DvbSubDestinationOutlineColor(v)
	}
	if v, ok := m["outline_size"].(int); ok && v != 0 {
		out.OutlineSize = aws.Int32(int32(v))
	}
	if v, ok := m["shadow_color"].(string); ok && len(v) > 0 {
		out.ShadowColor = types.DvbSubDestinationShadowColor(v)
	}
	if v, ok := m["shadow_opacity"].(int); ok && v != 0 {
		out.ShadowOpacity = aws.Int32(int32(v))
	}
	if v, ok := m["shadow_x_offset"].(int); ok && v != 0 {
		out.ShadowXOffset = aws.Int32(int32(v))
	}
	if v, ok := m["shadow_y_offset"].(int); ok && v != 0 {
		out.ShadowYOffset = aws.Int32(int32(v))
	}
	if v, ok := m["teletext_grid_control"].(string); ok && len(v) > 0 {
		out.TeletextGridControl = types.DvbSubDestinationTeletextGridControl(v)
	}
	if v, ok := m["x_position"].(int); ok && v != 0 {
		out.XPosition = aws.Int32(int32(v))
	}
	if v, ok := m["y_position"].(int); ok && v != 0 {
		out.YPosition = aws.Int32(int32(v))
	}

	return &out
}

func expandsCaptionDescriptionsDestinationSettingsEbuTtDDestinationSettings(tfList []any) *types.EbuTtDDestinationSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.EbuTtDDestinationSettings
	if v, ok := m["copyright_holder"].(string); ok && v != "" {
		out.CopyrightHolder = aws.String(v)
	}

	if v, ok := m["fill_line_gap"].(string); ok && len(v) > 0 {
		out.FillLineGap = types.EbuTtDFillLineGapControl(v)
	}

	if v, ok := m["font_family"].(string); ok && v != "" {
		out.FontFamily = aws.String(v)
	}

	if v, ok := m["style_control"].(string); ok && len(v) > 0 {
		out.StyleControl = types.EbuTtDDestinationStyleControl(v)
	}

	return &out
}

func expandsCaptionDescriptionsDestinationSettingsTeletextDestinationSettings(tfList []any) *types.TeletextDestinationSettings {
	if tfList == nil {
		return nil
	}

	var out types.TeletextDestinationSettings

	return &out
}

func expandsCaptionDescriptionsDestinationSettingsTtmlDestinationSettings(tfList []any) *types.TtmlDestinationSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.TtmlDestinationSettings
	if v, ok := m["style_control"].(string); ok && len(v) > 0 {
		out.StyleControl = types.TtmlDestinationStyleControl(v)
	}

	return &out
}

func expandsCaptionDescriptionsDestinationSettingsWebvttDestinationSettings(tfList []any) *types.WebvttDestinationSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.WebvttDestinationSettings
	if v, ok := m["style_control"].(string); ok && len(v) > 0 {
		out.StyleControl = types.WebvttDestinationStyleControl(v)
	}
	return &out
}

func expandChannelEncoderSettingsGlobalConfiguration(tfList []any) *types.GlobalConfiguration {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.GlobalConfiguration

	if v, ok := m["initial_audio_gain"].(int); ok && v != 0 {
		out.InitialAudioGain = aws.Int32(int32(v))
	}

	if v, ok := m["input_end_action"].(string); ok && len(v) > 0 {
		out.InputEndAction = types.GlobalConfigurationInputEndAction(v)
	}

	if v, ok := m["input_loss_behavior"].([]any); ok && len(v) > 0 {
		out.InputLossBehavior = expandChannelEncoderSettingsGlobalConfigurationInputLossBehavior(v)
	}

	if v, ok := m["output_locking_mode"].(string); ok && len(v) > 0 {
		out.OutputLockingMode = types.GlobalConfigurationOutputLockingMode(v)
	}

	if v, ok := m["output_timing_source"].(string); ok && len(v) > 0 {
		out.OutputTimingSource = types.GlobalConfigurationOutputTimingSource(v)
	}

	if v, ok := m["support_low_framerate_inputs"].(string); ok && len(v) > 0 {
		out.SupportLowFramerateInputs = types.GlobalConfigurationLowFramerateInputs(v)
	}

	return &out
}

func expandChannelEncoderSettingsGlobalConfigurationInputLossBehavior(tfList []any) *types.InputLossBehavior {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.InputLossBehavior

	if v, ok := m["black_frame_msec"].(int); ok && v != 0 {
		out.BlackFrameMsec = aws.Int32(int32(v))
	}

	if v, ok := m["input_loss_image_color"].(string); ok && v != "" {
		out.InputLossImageColor = aws.String(v)
	}

	if v, ok := m["input_loss_image_slate"].([]any); ok && len(v) > 0 {
		out.InputLossImageSlate = expandInputLocation(v)
	}

	if v, ok := m["input_loss_image_type"].(string); ok && len(v) > 0 {
		out.InputLossImageType = types.InputLossImageType(v)
	}

	if v, ok := m["repeat_frame_msec"].(int); ok && v != 0 {
		out.RepeatFrameMsec = aws.Int32(int32(v))
	}

	return &out
}

func expandChannelEncoderSettingsMotionGraphicsConfiguration(tfList []any) *types.MotionGraphicsConfiguration {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.MotionGraphicsConfiguration

	if v, ok := m["motion_graphics_settings"].([]any); ok && len(v) > 0 {
		out.MotionGraphicsSettings = expandChannelEncoderSettingsMotionGraphicsConfigurationMotionGraphicsSettings(v)
	}

	if v, ok := m["motion_graphics_insertion"].(string); ok && len(v) > 0 {
		out.MotionGraphicsInsertion = types.MotionGraphicsInsertion(v)
	}

	return &out
}

func expandChannelEncoderSettingsMotionGraphicsConfigurationMotionGraphicsSettings(tfList []any) *types.MotionGraphicsSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.MotionGraphicsSettings
	if v, ok := m["html_motion_graphics_settings"].([]any); ok && len(v) > 0 {
		out.HtmlMotionGraphicsSettings = &types.HtmlMotionGraphicsSettings{} // no exported elements in this list
	}

	return &out
}

func expandChannelEncoderSettingsNielsenConfiguration(tfList []any) *types.NielsenConfiguration {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.NielsenConfiguration
	if v, ok := m["distributor_id"].(string); ok && v != "" {
		out.DistributorId = aws.String(v)
	}

	if v, ok := m["nielsen_pcm_to_id3_tagging"].(string); ok && len(v) > 0 {
		out.NielsenPcmToId3Tagging = types.NielsenPcmToId3TaggingState(v)
	}

	return &out
}

func expandChannelEncoderSettingsVideoDescriptionsCodecSettings(tfList []any) *types.VideoCodecSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.VideoCodecSettings
	if v, ok := m["frame_capture_settings"].([]any); ok && len(v) > 0 {
		out.FrameCaptureSettings = expandsVideoDescriptionsCodecSettingsFrameCaptureSettings(v)
	}
	if v, ok := m["h264_settings"].([]any); ok && len(v) > 0 {
		out.H264Settings = expandsVideoDescriptionsCodecSettingsH264Settings(v)
	}
	if v, ok := m["h265_settings"].([]any); ok && len(v) > 0 {
		out.H265Settings = expandsVideoDescriptionsCodecSettingsH265Settings(v)
	}

	return &out
}

func expandsVideoDescriptionsCodecSettingsFrameCaptureSettings(tfList []any) *types.FrameCaptureSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.FrameCaptureSettings
	if v, ok := m["capture_interval"].(int); ok && v != 0 {
		out.CaptureInterval = aws.Int32(int32(v))
	}
	if v, ok := m["capture_interval_units"].(string); ok && v != "" {
		out.CaptureIntervalUnits = types.FrameCaptureIntervalUnit(v)
	}

	return &out
}

func expandsVideoDescriptionsCodecSettingsH264Settings(tfList []any) *types.H264Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.H264Settings
	if v, ok := m["adaptive_quantization"].(string); ok && v != "" {
		out.AdaptiveQuantization = types.H264AdaptiveQuantization(v)
	}
	if v, ok := m["afd_signaling"].(string); ok && v != "" {
		out.AfdSignaling = types.AfdSignaling(v)
	}
	if v, ok := m["bitrate"].(int); ok && v != 0 {
		out.Bitrate = aws.Int32(int32(v))
	}
	if v, ok := m["buf_fill_pct"].(int); ok && v != 0 {
		out.BufFillPct = aws.Int32(int32(v))
	}
	if v, ok := m["buf_size"].(int); ok && v != 0 {
		out.BufSize = aws.Int32(int32(v))
	}
	if v, ok := m["color_metadata"].(string); ok && v != "" {
		out.ColorMetadata = types.H264ColorMetadata(v)
	}
	if v, ok := m["entropy_encoding"].(string); ok && v != "" {
		out.EntropyEncoding = types.H264EntropyEncoding(v)
	}
	if v, ok := m["filter_settings"].([]any); ok && len(v) > 0 {
		out.FilterSettings = expandH264SettingsFilterSettings(v)
	}
	if v, ok := m["fixed_afd"].(string); ok && v != "" {
		out.FixedAfd = types.FixedAfd(v)
	}
	if v, ok := m["flicker_aq"].(string); ok && v != "" {
		out.FlickerAq = types.H264FlickerAq(v)
	}
	if v, ok := m["force_field_pictures"].(string); ok && v != "" {
		out.ForceFieldPictures = types.H264ForceFieldPictures(v)
	}
	if v, ok := m["framerate_control"].(string); ok && v != "" {
		out.FramerateControl = types.H264FramerateControl(v)
	}
	if v, ok := m["framerate_denominator"].(int); ok && v != 0 {
		out.FramerateDenominator = aws.Int32(int32(v))
	}
	if v, ok := m["framerate_numerator"].(int); ok && v != 0 {
		out.FramerateNumerator = aws.Int32(int32(v))
	}
	if v, ok := m["gop_b_reference"].(string); ok && v != "" {
		out.GopBReference = types.H264GopBReference(v)
	}
	if v, ok := m["gop_closed_cadence"].(int); ok && v != 0 {
		out.GopClosedCadence = aws.Int32(int32(v))
	}
	if v, ok := m["gop_num_b_frames"].(int); ok && v != 0 {
		out.GopNumBFrames = aws.Int32(int32(v))
	}
	if v, ok := m["gop_size"].(float64); ok && v != 0.0 {
		out.GopSize = aws.Float64(v)
	}
	if v, ok := m["gop_size_units"].(string); ok && v != "" {
		out.GopSizeUnits = types.H264GopSizeUnits(v)
	}
	if v, ok := m["level"].(string); ok && v != "" {
		out.Level = types.H264Level(v)
	}
	if v, ok := m["look_ahead_rate_control"].(string); ok && v != "" {
		out.LookAheadRateControl = types.H264LookAheadRateControl(v)
	}
	if v, ok := m["max_bitrate"].(int); ok && v != 0 {
		out.MaxBitrate = aws.Int32(int32(v))
	}
	if v, ok := m["min_i_interval"].(int); ok && v != 0 {
		out.MinIInterval = aws.Int32(int32(v))
	}
	if v, ok := m["num_ref_frames"].(int); ok && v != 0 {
		out.NumRefFrames = aws.Int32(int32(v))
	}
	if v, ok := m["par_control"].(string); ok && v != "" {
		out.ParControl = types.H264ParControl(v)
	}
	if v, ok := m["par_denominator"].(int); ok && v != 0 {
		out.ParDenominator = aws.Int32(int32(v))
	}
	if v, ok := m["par_numerator"].(int); ok && v != 0 {
		out.ParNumerator = aws.Int32(int32(v))
	}
	if v, ok := m[names.AttrProfile].(string); ok && v != "" {
		out.Profile = types.H264Profile(v)
	}
	if v, ok := m["quality_level"].(string); ok && v != "" {
		out.QualityLevel = types.H264QualityLevel(v)
	}
	if v, ok := m["qvbr_quality_level"].(int); ok && v != 0 {
		out.QvbrQualityLevel = aws.Int32(int32(v))
	}
	if v, ok := m["rate_control_mode"].(string); ok && v != "" {
		out.RateControlMode = types.H264RateControlMode(v)
	}
	if v, ok := m["scan_type"].(string); ok && v != "" {
		out.ScanType = types.H264ScanType(v)
	}
	if v, ok := m["scene_change_detect"].(string); ok && v != "" {
		out.SceneChangeDetect = types.H264SceneChangeDetect(v)
	}
	if v, ok := m["slices"].(int); ok && v != 0 {
		out.Slices = aws.Int32(int32(v))
	}
	if v, ok := m["softness"].(int); ok && v != 0 {
		out.Softness = aws.Int32(int32(v))
	}
	if v, ok := m["spatial_aq"].(string); ok && v != "" {
		out.SpatialAq = types.H264SpatialAq(v)
	}
	if v, ok := m["subgop_length"].(string); ok && v != "" {
		out.SubgopLength = types.H264SubGopLength(v)
	}
	if v, ok := m["syntax"].(string); ok && v != "" {
		out.Syntax = types.H264Syntax(v)
	}
	if v, ok := m["temporal_aq"].(string); ok && v != "" {
		out.TemporalAq = types.H264TemporalAq(v)
	}
	if v, ok := m["timecode_insertion"].(string); ok && v != "" {
		out.TimecodeInsertion = types.H264TimecodeInsertionBehavior(v)
	}

	return &out
}

func expandH264SettingsFilterSettings(tfList []any) *types.H264FilterSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.H264FilterSettings
	if v, ok := m["temporal_filter_settings"].([]any); ok && len(v) > 0 {
		out.TemporalFilterSettings = expandH264FilterSettingsTemporalFilterSettings(v)
	}

	return &out
}

func expandH264FilterSettingsTemporalFilterSettings(tfList []any) *types.TemporalFilterSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.TemporalFilterSettings
	if v, ok := m["post_filter_sharpening"].(string); ok && v != "" {
		out.PostFilterSharpening = types.TemporalFilterPostFilterSharpening(v)
	}
	if v, ok := m["strength"].(string); ok && v != "" {
		out.Strength = types.TemporalFilterStrength(v)
	}

	return &out
}

func expandsVideoDescriptionsCodecSettingsH265Settings(tfList []any) *types.H265Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.H265Settings
	if v, ok := m["framerate_denominator"].(int); ok && v != 0 {
		out.FramerateDenominator = aws.Int32(int32(v))
	}
	if v, ok := m["framerate_numerator"].(int); ok && v != 0 {
		out.FramerateNumerator = aws.Int32(int32(v))
	}
	if v, ok := m["adaptive_quantization"].(string); ok && v != "" {
		out.AdaptiveQuantization = types.H265AdaptiveQuantization(v)
	}
	if v, ok := m["afd_signaling"].(string); ok && v != "" {
		out.AfdSignaling = types.AfdSignaling(v)
	}
	if v, ok := m["alternative_transfer_function"].(string); ok && v != "" {
		out.AlternativeTransferFunction = types.H265AlternativeTransferFunction(v)
	}
	if v, ok := m["bitrate"].(int); ok && v != 0 {
		out.Bitrate = aws.Int32(int32(v))
	}
	if v, ok := m["buf_size"].(int); ok && v != 0 {
		out.BufSize = aws.Int32(int32(v))
	}
	if v, ok := m["color_metadata"].(string); ok && v != "" {
		out.ColorMetadata = types.H265ColorMetadata(v)
	}
	if v, ok := m["color_space_settings"].([]any); ok && len(v) > 0 {
		out.ColorSpaceSettings = expandH265ColorSpaceSettings(v)
	}
	if v, ok := m["filter_settings"].([]any); ok && len(v) > 0 {
		out.FilterSettings = expandH265FilterSettings(v)
	}
	if v, ok := m["fixed_afd"].(string); ok && v != "" {
		out.FixedAfd = types.FixedAfd(v)
	}
	if v, ok := m["flicker_aq"].(string); ok && v != "" {
		out.FlickerAq = types.H265FlickerAq(v)
	}
	if v, ok := m["gop_closed_cadence"].(int); ok && v != 0 {
		out.GopClosedCadence = aws.Int32(int32(v))
	}
	if v, ok := m["gop_size"].(float64); ok && v != 0.0 {
		out.GopSize = aws.Float64(v)
	}
	if v, ok := m["gop_size_units"].(string); ok && v != "" {
		out.GopSizeUnits = types.H265GopSizeUnits(v)
	}
	if v, ok := m["level"].(string); ok && v != "" {
		out.Level = types.H265Level(v)
	}
	if v, ok := m["look_ahead_rate_control"].(string); ok && v != "" {
		out.LookAheadRateControl = types.H265LookAheadRateControl(v)
	}
	if v, ok := m["max_bitrate"].(int); ok && v != 0 {
		out.MaxBitrate = aws.Int32(int32(v))
	}
	if v, ok := m["min_i_interval"].(int); ok && v != 0 {
		out.MinIInterval = aws.Int32(int32(v))
	}
	if v, ok := m["min_qp"].(int); ok && v != 0 {
		out.MinQp = aws.Int32(int32(v))
	}
	if v, ok := m["mv_over_picture_boundaries"].(string); ok && v != "" {
		out.MvOverPictureBoundaries = types.H265MvOverPictureBoundaries(v)
	}
	if v, ok := m["mv_temporal_predictor"].(string); ok && v != "" {
		out.MvTemporalPredictor = types.H265MvTemporalPredictor(v)
	}
	if v, ok := m["par_denominator"].(int); ok && v != 0 {
		out.ParDenominator = aws.Int32(int32(v))
	}
	if v, ok := m["par_numerator"].(int); ok && v != 0 {
		out.ParNumerator = aws.Int32(int32(v))
	}
	if v, ok := m[names.AttrProfile].(string); ok && v != "" {
		out.Profile = types.H265Profile(v)
	}
	if v, ok := m["qvbr_quality_level"].(int); ok && v != 0 {
		out.QvbrQualityLevel = aws.Int32(int32(v))
	}
	if v, ok := m["rate_control_mode"].(string); ok && v != "" {
		out.RateControlMode = types.H265RateControlMode(v)
	}
	if v, ok := m["scan_type"].(string); ok && v != "" {
		out.ScanType = types.H265ScanType(v)
	}
	if v, ok := m["scene_change_detect"].(string); ok && v != "" {
		out.SceneChangeDetect = types.H265SceneChangeDetect(v)
	}
	if v, ok := m["slices"].(int); ok && v != 0 {
		out.Slices = aws.Int32(int32(v))
	}
	if v, ok := m["tier"].(string); ok && v != "" {
		out.Tier = types.H265Tier(v)
	}
	if v, ok := m["tile_height"].(int); ok && v != 0 {
		out.TileHeight = aws.Int32(int32(v))
	}
	if v, ok := m["tile_padding"].(string); ok && v != "" {
		out.TilePadding = types.H265TilePadding(v)
	}
	if v, ok := m["tile_width"].(int); ok && v != 0 {
		out.TileWidth = aws.Int32(int32(v))
	}
	if v, ok := m["timecode_burnin_settings"].([]any); ok && len(v) > 0 {
		out.TimecodeBurninSettings = expandH265TimecodeBurninSettings(v)
	}
	if v, ok := m["timecode_insertion"].(string); ok && v != "" {
		out.TimecodeInsertion = types.H265TimecodeInsertionBehavior(v)
	}
	if v, ok := m["treeblock_size"].(string); ok && v != "" {
		out.TreeblockSize = types.H265TreeblockSize(v)
	}

	return &out
}

func expandH265ColorSpaceSettings(tfList []any) *types.H265ColorSpaceSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.H265ColorSpaceSettings
	if v, ok := m["color_space_passthrough_settings"].([]any); ok && len(v) > 0 {
		out.ColorSpacePassthroughSettings = &types.ColorSpacePassthroughSettings{} // no exported elements in this list
	}
	if v, ok := m["dolby_vision81_settings"].([]any); ok && len(v) > 0 {
		out.DolbyVision81Settings = &types.DolbyVision81Settings{} // no exported elements in this list
	}
	if v, ok := m["hdr10_settings"].([]any); ok && len(v) > 0 {
		out.Hdr10Settings = expandH265Hdr10Settings(v)
	}
	if v, ok := m["rec601_settings"].([]any); ok && len(v) > 0 {
		out.Rec601Settings = &types.Rec601Settings{} // no exported elements in this list
	}
	if v, ok := m["rec709_settings"].([]any); ok && len(v) > 0 {
		out.Rec709Settings = &types.Rec709Settings{} // no exported elements in this list
	}

	return &out
}

func expandH265Hdr10Settings(tfList []any) *types.Hdr10Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.Hdr10Settings
	if v, ok := m["max_cll"].(int); ok && v != 0 {
		out.MaxCll = aws.Int32(int32(v))
	}
	if v, ok := m["max_fall"].(int); ok && v != 0 {
		out.MaxFall = aws.Int32(int32(v))
	}

	return &out
}

func expandH265FilterSettings(tfList []any) *types.H265FilterSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.H265FilterSettings
	if v, ok := m["temporal_filter_settings"].([]any); ok && len(v) > 0 {
		out.TemporalFilterSettings = expandH265FilterSettingsTemporalFilterSettings(v)
	}

	return &out
}

func expandH265FilterSettingsTemporalFilterSettings(tfList []any) *types.TemporalFilterSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.TemporalFilterSettings
	if v, ok := m["post_filter_sharpening"].(string); ok && v != "" {
		out.PostFilterSharpening = types.TemporalFilterPostFilterSharpening(v)
	}
	if v, ok := m["strength"].(string); ok && v != "" {
		out.Strength = types.TemporalFilterStrength(v)
	}

	return &out
}

func expandH265TimecodeBurninSettings(tfList []any) *types.TimecodeBurninSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.TimecodeBurninSettings
	if v, ok := m["timecode_burnin_font_size"].(string); ok && v != "" {
		out.FontSize = types.TimecodeBurninFontSize(v)
	}
	if v, ok := m["timecode_burnin_position"].(string); ok && v != "" {
		out.Position = types.TimecodeBurninPosition(v)
	}
	if v, ok := m[names.AttrPrefix].(string); ok && v != "" {
		out.Prefix = &v
	}

	return &out
}

func expandNielsenCbetSettings(tfList []any) *types.NielsenCBET {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.NielsenCBET
	if v, ok := m["cbet_check_digit_string"].(string); ok && v != "" {
		out.CbetCheckDigitString = aws.String(v)
	}
	if v, ok := m["cbet_stepaside"].(string); ok && v != "" {
		out.CbetStepaside = types.NielsenWatermarksCbetStepaside(v)
	}
	if v, ok := m["csid"].(string); ok && v != "" {
		out.Csid = aws.String(v)
	}

	return &out
}

func expandNielsenNaseIiNwSettings(tfList []any) *types.NielsenNaesIiNw {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	var out types.NielsenNaesIiNw
	if v, ok := m["check_digit_string"].(string); ok && v != "" {
		out.CheckDigitString = aws.String(v)
	}
	if v, ok := m["sid"].(float32); ok && v != 0.0 {
		out.Sid = aws.Float64(float64(v))
	}

	return &out
}

func flattenChannelEncoderSettings(apiObject *types.EncoderSettings) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"audio_descriptions": flattenAudioDescriptions(apiObject.AudioDescriptions),
		"output_groups":      flattenOutputGroups(apiObject.OutputGroups),
		"timecode_config":    flattenTimecodeConfig(apiObject.TimecodeConfig),
		"video_descriptions": flattenVideoDescriptions(apiObject.VideoDescriptions),
		"avail_blanking":     flattenAvailBlanking(apiObject.AvailBlanking),
		// TODO avail_configuration
		// TODO blackout_slate
		"caption_descriptions": flattenCaptionDescriptions(apiObject.CaptionDescriptions),
		// TODO feature_activations
		"global_configuration":          flattenGlobalConfiguration(apiObject.GlobalConfiguration),
		"motion_graphics_configuration": flattenMotionGraphicsConfiguration(apiObject.MotionGraphicsConfiguration),
		"nielsen_configuration":         flattenNielsenConfiguration(apiObject.NielsenConfiguration),
	}

	return []any{m}
}

func flattenAudioDescriptions(od []types.AudioDescription) []any {
	if len(od) == 0 {
		return nil
	}

	var ml []any

	for _, v := range od {
		m := map[string]any{
			"audio_selector_name":          aws.ToString(v.AudioSelectorName),
			names.AttrName:                 aws.ToString(v.Name),
			"audio_normalization_settings": flattenAudioNormalization(v.AudioNormalizationSettings),
			"audio_type":                   v.AudioType,
			"audio_type_control":           v.AudioTypeControl,
			"audio_watermark_settings":     flattenAudioWatermarkSettings(v.AudioWatermarkingSettings),
			"codec_settings":               flattenAudioDescriptionsCodecSettings(v.CodecSettings),
			names.AttrLanguageCode:         aws.ToString(v.LanguageCode),
			"language_code_control":        string(v.LanguageCodeControl),
			"remix_settings":               flattenAudioDescriptionsRemixSettings(v.RemixSettings),
			"stream_name":                  aws.ToString(v.StreamName),
		}

		ml = append(ml, m)
	}

	return ml
}

func flattenOutputGroups(op []types.OutputGroup) []any {
	if len(op) == 0 {
		return nil
	}

	var ol []any

	for _, v := range op {
		m := map[string]any{
			"output_group_settings": flattenOutputGroupSettings(v.OutputGroupSettings),
			"outputs":               flattenOutputs(v.Outputs),
			names.AttrName:          aws.ToString(v.Name),
		}

		ol = append(ol, m)
	}

	return ol
}

func flattenOutputGroupSettings(os *types.OutputGroupSettings) []any {
	if os == nil {
		return nil
	}

	m := map[string]any{
		"archive_group_settings":       flattenOutputGroupSettingsArchiveGroupSettings(os.ArchiveGroupSettings),
		"frame_capture_group_settings": flattenOutputGroupSettingsFrameCaptureGroupSettings(os.FrameCaptureGroupSettings),
		"hls_group_settings":           flattenOutputGroupSettingsHLSGroupSettings(os.HlsGroupSettings),
		"ms_smooth_group_settings":     flattenOutputGroupSettingsMsSmoothGroupSettings(os.MsSmoothGroupSettings),
		"media_package_group_settings": flattenOutputGroupSettingsMediaPackageGroupSettings(os.MediaPackageGroupSettings),
		"multiplex_group_settings": func(inner *types.MultiplexGroupSettings) []any {
			if inner == nil {
				return nil
			}
			return []any{} // no exported attributes
		}(os.MultiplexGroupSettings),
		"rtmp_group_settings": flattenOutputGroupSettingsRtmpGroupSettings(os.RtmpGroupSettings),
		"udp_group_settings":  flattenOutputGroupSettingsUdpGroupSettings(os.UdpGroupSettings),
	}

	return []any{m}
}

func flattenOutputs(os []types.Output) []any {
	if len(os) == 0 {
		return nil
	}

	var outputs []any

	for _, item := range os {
		m := map[string]any{
			"audio_description_names":   flex.FlattenStringValueSet(item.AudioDescriptionNames),
			"caption_description_names": flex.FlattenStringValueSet(item.CaptionDescriptionNames),
			"output_name":               aws.ToString(item.OutputName),
			"output_settings":           flattenOutputsOutputSettings(item.OutputSettings),
			"video_description_name":    aws.ToString(item.VideoDescriptionName),
		}

		outputs = append(outputs, m)
	}

	return outputs
}

func flattenOutputsOutputSettings(in *types.OutputSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"archive_output_settings":       flattenOutputsOutputSettingsArchiveOutputSettings(in.ArchiveOutputSettings),
		"frame_capture_output_settings": flattenOutputsOutputSettingsFrameCaptureOutputSettings(in.FrameCaptureOutputSettings),
		"hls_output_settings":           flattenOutputsOutputSettingsHLSOutputSettings(in.HlsOutputSettings),
		"media_package_output_settings": func(inner *types.MediaPackageOutputSettings) []any {
			if inner == nil {
				return nil
			}
			return []any{} // no exported attributes
		}(in.MediaPackageOutputSettings),
		"ms_smooth_output_settings": flattenOutputsOutputSettingsMsSmoothOutputSettings(in.MsSmoothOutputSettings),
		"multiplex_output_settings": func(inner *types.MultiplexOutputSettings) []any {
			if inner == nil {
				return nil
			}
			data := map[string]any{
				names.AttrDestination: flattenDestination(inner.Destination),
			}

			return []any{data}
		}(in.MultiplexOutputSettings),
		"rtmp_output_settings": flattenOutputsOutputSettingsRtmpOutputSettings(in.RtmpOutputSettings),
		"udp_output_settings":  flattenOutputsOutputSettingsUdpOutputSettings(in.UdpOutputSettings),
	}

	return []any{m}
}

func flattenOutputsOutputSettingsArchiveOutputSettings(in *types.ArchiveOutputSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"container_settings": flattenOutputsOutputSettingsArchiveOutputSettingsContainerSettings(in.ContainerSettings),
		"extension":          aws.ToString(in.Extension),
		"name_modifier":      aws.ToString(in.NameModifier),
	}

	return []any{m}
}

func flattenOutputsOutputSettingsFrameCaptureOutputSettings(in *types.FrameCaptureOutputSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"name_modifier": aws.ToString(in.NameModifier),
	}

	return []any{m}
}

func flattenOutputsOutputSettingsHLSOutputSettings(in *types.HlsOutputSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"hls_settings":        flattenHLSOutputSettingsHLSSettings(in.HlsSettings),
		"h265_packaging_type": string(in.H265PackagingType),
		"name_modifier":       aws.ToString(in.NameModifier),
		"segment_modifier":    aws.ToString(in.SegmentModifier),
	}

	return []any{m}
}

func flattenOutputsOutputSettingsMsSmoothOutputSettings(in *types.MsSmoothOutputSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"h265_packaging_type": string(in.H265PackagingType),
		"name_modifier":       aws.ToString(in.NameModifier),
	}

	return []any{m}
}

func flattenHLSOutputSettingsHLSSettings(in *types.HlsSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"audio_only_hls_settings": flattenHLSSettingsAudioOnlyHLSSettings(in.AudioOnlyHlsSettings),
		"fmp4_hls_settings":       flattenHLSSettingsFmp4HLSSettings(in.Fmp4HlsSettings),
		"frame_capture_hls_settings": func(inner *types.FrameCaptureHlsSettings) []any {
			if inner == nil {
				return nil
			}
			return []any{} // no exported fields
		}(in.FrameCaptureHlsSettings),
		"standard_hls_settings": flattenHLSSettingsStandardHLSSettings(in.StandardHlsSettings),
	}

	return []any{m}
}

func flattenHLSSettingsAudioOnlyHLSSettings(in *types.AudioOnlyHlsSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"audio_group_id":   aws.ToString(in.AudioGroupId),
		"audio_only_image": flattenInputLocation(in.AudioOnlyImage),
		"audio_track_type": string(in.AudioTrackType),
		"segment_type":     string(in.AudioTrackType),
	}

	return []any{m}
}

func flattenHLSSettingsFmp4HLSSettings(in *types.Fmp4HlsSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"audio_rendition_sets":    aws.ToString(in.AudioRenditionSets),
		"nielsen_id3_behavior":    string(in.NielsenId3Behavior),
		"timed_metadata_behavior": string(in.TimedMetadataBehavior),
	}

	return []any{m}
}

func flattenHLSSettingsStandardHLSSettings(in *types.StandardHlsSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"m3u8_settings":        flattenStandardHLSSettingsM3u8Settings(in.M3u8Settings),
		"audio_rendition_sets": aws.ToString(in.AudioRenditionSets),
	}

	return []any{m}
}

func flattenStandardHLSSettingsM3u8Settings(in *types.M3u8Settings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"audio_frames_per_pes":    int(aws.ToInt32(in.AudioFramesPerPes)),
		"audio_pids":              aws.ToString(in.AudioPids),
		"ecm_pid":                 aws.ToString(in.EcmPid),
		"nielsen_id3_behavior":    string(in.NielsenId3Behavior),
		"pat_interval":            int(aws.ToInt32(in.PatInterval)),
		"pcr_control":             string(in.PcrControl),
		"pcr_period":              int(aws.ToInt32(in.PcrPeriod)),
		"pcr_pid":                 aws.ToString(in.PcrPid),
		"pmt_interval":            int(aws.ToInt32(in.PmtInterval)),
		"pmt_pid":                 aws.ToString(in.PmtPid),
		"program_num":             int(aws.ToInt32(in.ProgramNum)),
		"scte35_behavior":         string(in.Scte35Behavior),
		"scte35_pid":              aws.ToString(in.Scte35Pid),
		"timed_metadata_behavior": string(in.TimedMetadataBehavior),
		"timed_metadata_pid":      aws.ToString(in.TimedMetadataPid),
		"transport_stream_id":     int(aws.ToInt32(in.TransportStreamId)),
		"video_pid":               aws.ToString(in.VideoPid),
	}

	return []any{m}
}

func flattenOutputsOutputSettingsRtmpOutputSettings(in *types.RtmpOutputSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		names.AttrDestination:       flattenDestination(in.Destination),
		"certificate_mode":          string(in.CertificateMode),
		"connection_retry_interval": int(aws.ToInt32(in.ConnectionRetryInterval)),
		"num_retries":               int(aws.ToInt32(in.NumRetries)),
	}

	return []any{m}
}

func flattenOutputsOutputSettingsUdpOutputSettings(in *types.UdpOutputSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"container_settings":  flattenOutputsOutputSettingsUdpOutputSettingsContainerSettings(in.ContainerSettings),
		names.AttrDestination: flattenDestination(in.Destination),
		"buffer_msec":         int(aws.ToInt32(in.BufferMsec)),
		"fec_output_settings": flattenFecOutputSettings(in.FecOutputSettings),
	}

	return []any{m}
}

func flattenOutputsOutputSettingsArchiveOutputSettingsContainerSettings(in *types.ArchiveContainerSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"m2ts_settings": flattenM2tsSettings(in.M2tsSettings),
		"raw_settings":  []any{}, // attribute has no exported fields
	}

	return []any{m}
}

func flattenOutputsOutputSettingsUdpOutputSettingsContainerSettings(in *types.UdpContainerSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"m2ts_settings": flattenM2tsSettings(in.M2tsSettings),
	}

	return []any{m}
}

func flattenFecOutputSettings(in *types.FecOutputSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"column_depth": int(aws.ToInt32(in.ColumnDepth)),
		"include_fec":  string(in.IncludeFec),
		"row_length":   int(aws.ToInt32(in.RowLength)),
	}

	return []any{m}
}

func flattenM2tsSettings(in *types.M2tsSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"absent_input_audio_behavior": string(in.AbsentInputAudioBehavior),
		"arib":                        string(in.Arib),
		"arib_captions_pid":           aws.ToString(in.AribCaptionsPid),
		"arib_captions_pid_control":   string(in.AribCaptionsPidControl),
		"audio_buffer_model":          string(in.AudioBufferModel),
		"audio_frames_per_pes":        int(aws.ToInt32(in.AudioFramesPerPes)),
		"audio_pids":                  aws.ToString(in.AudioPids),
		"audio_stream_type":           string(in.AudioStreamType),
		"bitrate":                     int(aws.ToInt32(in.Bitrate)),
		"buffer_model":                string(in.BufferModel),
		"cc_descriptor":               string(in.CcDescriptor),
		"dvb_nit_settings":            flattenDvbNitSettings(in.DvbNitSettings),
		"dvb_sdt_settings":            flattenDvbSdtSettings(in.DvbSdtSettings),
		"dvb_sub_pids":                aws.ToString(in.DvbSubPids),
		"dvb_tdt_settings":            flattenDvbTdtSettings(in.DvbTdtSettings),
		"dvb_teletext_pid":            aws.ToString(in.DvbTeletextPid),
		"ebif":                        string(in.Ebif),
		"ebp_audio_interval":          string(in.EbpAudioInterval),
		"ebp_lookahead_ms":            int(aws.ToInt32(in.EbpLookaheadMs)),
		"ebp_placement":               string(in.EbpPlacement),
		"ecm_pid":                     aws.ToString(in.EcmPid),
		"es_rate_in_pes":              string(in.EsRateInPes),
		"etv_platform_pid":            aws.ToString(in.EtvPlatformPid),
		"etv_signal_pid":              aws.ToString(in.EtvSignalPid),
		"fragment_time":               in.FragmentTime,
		"klv":                         string(in.Klv),
		"klv_data_pids":               aws.ToString(in.KlvDataPids),
		"nielsen_id3_behavior":        string(in.NielsenId3Behavior),
		"null_packet_bitrate":         float32(aws.ToFloat64(in.NullPacketBitrate)),
		"pat_interval":                int(aws.ToInt32(in.PatInterval)),
		"pcr_control":                 string(in.PcrControl),
		"pcr_period":                  int(aws.ToInt32(in.PcrPeriod)),
		"pcr_pid":                     aws.ToString(in.PcrPid),
		"pmt_interval":                int(aws.ToInt32(in.PmtInterval)),
		"pmt_pid":                     aws.ToString(in.PmtPid),
		"program_num":                 int(aws.ToInt32(in.ProgramNum)),
		"rate_mode":                   string(in.RateMode),
		"scte27_pids":                 aws.ToString(in.Scte27Pids),
		"scte35_control":              string(in.Scte35Control),
		"scte35_pid":                  aws.ToString(in.Scte35Pid),
		"segmentation_markers":        string(in.SegmentationMarkers),
		"segmentation_style":          string(in.SegmentationStyle),
		"segmentation_time":           in.SegmentationTime,
		"timed_metadata_behavior":     string(in.TimedMetadataBehavior),
		"timed_metadata_pid":          aws.ToString(in.TimedMetadataPid),
		"transport_stream_id":         int(aws.ToInt32(in.TransportStreamId)),
		"video_pid":                   aws.ToString(in.VideoPid),
	}

	return []any{m}
}

func flattenDvbNitSettings(in *types.DvbNitSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"network_id":   int(aws.ToInt32(in.NetworkId)),
		"network_name": aws.ToString(in.NetworkName),
		"rep_interval": int(aws.ToInt32(in.RepInterval)),
	}

	return []any{m}
}

func flattenDvbSdtSettings(in *types.DvbSdtSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"output_sdt":            string(in.OutputSdt),
		"rep_interval":          int(aws.ToInt32(in.RepInterval)),
		names.AttrServiceName:   aws.ToString(in.ServiceName),
		"service_provider_name": aws.ToString(in.ServiceProviderName),
	}

	return []any{m}
}

func flattenDvbTdtSettings(in *types.DvbTdtSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"rep_interval": int(aws.ToInt32(in.RepInterval)),
	}

	return []any{m}
}

func flattenOutputGroupSettingsArchiveGroupSettings(in *types.ArchiveGroupSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		names.AttrDestination:  flattenDestination(in.Destination),
		"archive_cdn_settings": flattenOutputGroupSettingsArchiveCDNSettings(in.ArchiveCdnSettings),
		"rollover_interval":    int(aws.ToInt32(in.RolloverInterval)),
	}

	return []any{m}
}

func flattenOutputGroupSettingsFrameCaptureGroupSettings(in *types.FrameCaptureGroupSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		names.AttrDestination:        flattenDestination(in.Destination),
		"frame_capture_cdn_settings": flattenFrameCaptureCDNSettings(in.FrameCaptureCdnSettings),
	}

	return []any{m}
}

func flattenOutputGroupSettingsHLSGroupSettings(in *types.HlsGroupSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		names.AttrDestination:          flattenDestination(in.Destination),
		"ad_markers":                   flattenHLSAdMarkers(in.AdMarkers),
		"base_url_content":             aws.ToString(in.BaseUrlContent),
		"base_url_content1":            aws.ToString(in.BaseUrlContent1),
		"base_url_manifest":            aws.ToString(in.BaseUrlManifest),
		"base_url_manifest1":           aws.ToString(in.BaseUrlManifest1),
		"caption_language_mappings":    flattenHLSCaptionLanguageMappings(in.CaptionLanguageMappings),
		"caption_language_setting":     string(in.CaptionLanguageSetting),
		"client_cache":                 string(in.ClientCache),
		"codec_specification":          string(in.CodecSpecification),
		"constant_iv":                  aws.ToString(in.ConstantIv),
		"directory_structure":          string(in.DirectoryStructure),
		"discontinuity_tags":           string(in.DiscontinuityTags),
		"encryption_type":              string(in.EncryptionType),
		"hls_cdn_settings":             flattenHLSCDNSettings(in.HlsCdnSettings),
		"hls_id3_segment_tagging":      string(in.HlsId3SegmentTagging),
		"iframe_only_playlists":        string(in.IFrameOnlyPlaylists),
		"incomplete_segment_behavior":  string(in.IncompleteSegmentBehavior),
		"index_n_segments":             int(aws.ToInt32(in.IndexNSegments)),
		"input_loss_action":            string(in.InputLossAction),
		"iv_in_manifest":               string(in.IvInManifest),
		"iv_source":                    string(in.IvSource),
		"keep_segments":                int(aws.ToInt32(in.KeepSegments)),
		"key_format":                   aws.ToString(in.KeyFormat),
		"key_format_versions":          aws.ToString(in.KeyFormatVersions),
		"key_provider_settings":        flattenHLSKeyProviderSettings(in.KeyProviderSettings),
		"manifest_compression":         string(in.ManifestCompression),
		"manifest_duration_format":     string(in.ManifestDurationFormat),
		"min_segment_length":           int(aws.ToInt32(in.MinSegmentLength)),
		names.AttrMode:                 string(in.Mode),
		"output_selection":             string(in.OutputSelection),
		"program_date_time":            string(in.ProgramDateTime),
		"program_date_time_clock":      string(in.ProgramDateTimeClock),
		"program_date_time_period":     int(aws.ToInt32(in.ProgramDateTimePeriod)),
		"redundant_manifest":           string(in.RedundantManifest),
		"segment_length":               int(aws.ToInt32(in.SegmentLength)),
		"segments_per_subdirectory":    int(aws.ToInt32(in.SegmentsPerSubdirectory)),
		"stream_inf_resolution":        string(in.StreamInfResolution),
		"timed_metadata_id3_frame":     string(in.TimedMetadataId3Frame),
		"timed_metadata_id3_period":    int(aws.ToInt32(in.TimedMetadataId3Period)),
		"timestamp_delta_milliseconds": int(aws.ToInt32(in.TimestampDeltaMilliseconds)),
		"ts_file_mode":                 string(in.TsFileMode),
	}

	return []any{m}
}

func flattenOutputGroupSettingsMsSmoothGroupSettings(in *types.MsSmoothGroupSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		names.AttrDestination:         flattenDestination(in.Destination),
		"acquisition_point_id":        aws.ToString(in.AcquisitionPointId),
		"audio_only_timecode_control": string(in.AudioOnlyTimecodeControl),
		"certificate_mode":            string(in.CertificateMode),
		"connection_retry_interval":   int(aws.ToInt32(in.ConnectionRetryInterval)),
		"event_id":                    aws.ToString(in.EventId),
		"event_id_mode":               string(in.EventIdMode),
		"event_stop_behavior":         string(in.EventStopBehavior),
		"filecache_duration":          int(aws.ToInt32(in.FilecacheDuration)),
		"fragment_length":             int(aws.ToInt32(in.FragmentLength)),
		"input_loss_action":           string(in.InputLossAction),
		"num_retries":                 int(aws.ToInt32(in.NumRetries)),
		"restart_delay":               int(aws.ToInt32(in.RestartDelay)),
		"segmentation_mode":           string(in.SegmentationMode),
		"send_delay_ms":               int(aws.ToInt32(in.SendDelayMs)),
		"sparse_track_type":           string(in.SparseTrackType),
		"stream_manifest_behavior":    string(in.StreamManifestBehavior),
		"timestamp_offset":            aws.ToString(in.TimestampOffset),
		"timestamp_offset_mode":       string(in.TimestampOffsetMode),
	}

	return []any{m}
}

func flattenHLSAdMarkers(in []types.HlsAdMarkers) []any {
	if len(in) == 0 {
		return nil
	}

	var out []any
	for _, item := range in {
		out = append(out, string(item))
	}

	return out
}

func flattenHLSCaptionLanguageMappings(in []types.CaptionLanguageMapping) []any {
	if len(in) == 0 {
		return nil
	}

	var out []any
	for _, item := range in {
		m := map[string]any{
			"caption_channel":      int(aws.ToInt32(item.CaptionChannel)),
			names.AttrLanguageCode: aws.ToString(item.LanguageCode),
			"language_description": aws.ToString(item.LanguageDescription),
		}

		out = append(out, m)
	}

	return out
}

func flattenHLSCDNSettings(in *types.HlsCdnSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"hls_akamai_settings":      flattenHLSAkamaiSettings(in.HlsAkamaiSettings),
		"hls_basic_put_settings":   flattenHLSBasicPutSettings(in.HlsBasicPutSettings),
		"hls_media_store_settings": flattenHLSMediaStoreSettings(in.HlsMediaStoreSettings),
		"hls_s3_settings":          flattenHLSS3Settings(in.HlsS3Settings),
		"hls_webdav_settings":      flattenHLSWebdavSettings(in.HlsWebdavSettings),
	}

	return []any{m}
}

func flattenHLSAkamaiSettings(in *types.HlsAkamaiSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"connection_retry_interval": int(aws.ToInt32(in.ConnectionRetryInterval)),
		"filecache_duration":        int(aws.ToInt32(in.FilecacheDuration)),
		"http_transfer_mode":        string(in.HttpTransferMode),
		"num_retries":               int(aws.ToInt32(in.NumRetries)),
		"restart_delay":             int(aws.ToInt32(in.RestartDelay)),
		"salt":                      aws.ToString(in.Salt),
		"token":                     aws.ToString(in.Token),
	}

	return []any{m}
}

func flattenHLSBasicPutSettings(in *types.HlsBasicPutSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"connection_retry_interval": int(aws.ToInt32(in.ConnectionRetryInterval)),
		"filecache_duration":        int(aws.ToInt32(in.FilecacheDuration)),
		"num_retries":               int(aws.ToInt32(in.NumRetries)),
		"restart_delay":             int(aws.ToInt32(in.RestartDelay)),
	}

	return []any{m}
}

func flattenHLSMediaStoreSettings(in *types.HlsMediaStoreSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"connection_retry_interval": int(aws.ToInt32(in.ConnectionRetryInterval)),
		"filecache_duration":        int(aws.ToInt32(in.FilecacheDuration)),
		"media_store_storage_class": string(in.MediaStoreStorageClass),
		"num_retries":               int(aws.ToInt32(in.NumRetries)),
		"restart_delay":             int(aws.ToInt32(in.RestartDelay)),
	}

	return []any{m}
}

func flattenHLSS3Settings(in *types.HlsS3Settings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"canned_acl": string(in.CannedAcl),
	}

	return []any{m}
}

func flattenFrameCaptureCDNSettings(in *types.FrameCaptureCdnSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"frame_capture_s3_settings": flattenFrameCaptureS3Settings(in.FrameCaptureS3Settings),
	}

	return []any{m}
}

func flattenHLSWebdavSettings(in *types.HlsWebdavSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"connection_retry_interval": int(aws.ToInt32(in.ConnectionRetryInterval)),
		"filecache_duration":        int(aws.ToInt32(in.FilecacheDuration)),
		"http_transfer_mode":        string(in.HttpTransferMode),
		"num_retries":               int(aws.ToInt32(in.NumRetries)),
		"restart_delay":             int(aws.ToInt32(in.RestartDelay)),
	}

	return []any{m}
}

func flattenHLSKeyProviderSettings(in *types.KeyProviderSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"static_key_settings": flattenKeyProviderSettingsStaticKeySettings(in.StaticKeySettings),
	}

	return []any{m}
}

func flattenKeyProviderSettingsStaticKeySettings(in *types.StaticKeySettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"static_key_value":    aws.ToString(in.StaticKeyValue),
		"key_provider_server": flattenInputLocation(in.KeyProviderServer),
	}

	return []any{m}
}

func flattenInputLocation(in *types.InputLocation) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		names.AttrURI:      aws.ToString(in.Uri),
		"password_param":   aws.ToString(in.PasswordParam),
		names.AttrUsername: aws.ToString(in.Username),
	}

	return []any{m}
}

func flattenFrameCaptureS3Settings(in *types.FrameCaptureS3Settings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"canned_acl": string(in.CannedAcl),
	}

	return []any{m}
}

func flattenOutputGroupSettingsMediaPackageGroupSettings(mp *types.MediaPackageGroupSettings) []any {
	if mp == nil {
		return nil
	}

	m := map[string]any{
		names.AttrDestination: flattenDestination(mp.Destination),
	}

	return []any{m}
}

func flattenOutputGroupSettingsRtmpGroupSettings(in *types.RtmpGroupSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"ad_markers":            flattenAdMakers(in.AdMarkers),
		"authentication_scheme": string(in.AuthenticationScheme),
		"cache_full_behavior":   string(in.CacheFullBehavior),
		"cache_length":          int(aws.ToInt32(in.CacheLength)),
		"caption_data":          string(in.CaptionData),
		"input_loss_action":     string(in.InputLossAction),
		"restart_delay":         int(aws.ToInt32(in.RestartDelay)),
	}

	return []any{m}
}

func flattenOutputGroupSettingsUdpGroupSettings(in *types.UdpGroupSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"input_loss_action":         string(in.InputLossAction),
		"timed_metadata_id3_frame":  string(in.TimedMetadataId3Frame),
		"timed_metadata_id3_period": int(aws.ToInt32(in.TimedMetadataId3Period)),
	}

	return []any{m}
}

func flattenAdMakers(l []types.RtmpAdMarkers) []string {
	if len(l) == 0 {
		return nil
	}

	var out []string
	for _, v := range l {
		out = append(out, string(v))
	}

	return out
}

func flattenDestination(des *types.OutputLocationRef) []any {
	if des == nil {
		return nil
	}

	m := map[string]any{
		"destination_ref_id": aws.ToString(des.DestinationRefId),
	}

	return []any{m}
}

func flattenOutputGroupSettingsArchiveCDNSettings(as *types.ArchiveCdnSettings) []any {
	if as == nil {
		return nil
	}

	m := map[string]any{
		"archive_s3_settings": func(in *types.ArchiveS3Settings) []any {
			if in == nil {
				return nil
			}

			inner := map[string]any{
				"canned_acl": string(in.CannedAcl),
			}

			return []any{inner}
		}(as.ArchiveS3Settings),
	}

	return []any{m}
}

func flattenTimecodeConfig(in *types.TimecodeConfig) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		names.AttrSource: string(in.Source),
		"sync_threshold": int(aws.ToInt32(in.SyncThreshold)),
	}

	return []any{m}
}

func flattenVideoDescriptions(tfList []types.VideoDescription) []any {
	if len(tfList) == 0 {
		return nil
	}

	var out []any

	for _, item := range tfList {
		m := map[string]any{
			names.AttrName:     aws.ToString(item.Name),
			"codec_settings":   flattenVideoDescriptionsCodecSettings(item.CodecSettings),
			"height":           int(aws.ToInt32(item.Height)),
			"respond_to_afd":   string(item.RespondToAfd),
			"scaling_behavior": string(item.ScalingBehavior),
			"sharpness":        int(aws.ToInt32(item.Sharpness)),
			"width":            int(aws.ToInt32(item.Width)),
		}

		out = append(out, m)
	}
	return out
}

func flattenAvailBlanking(in *types.AvailBlanking) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"avail_blanking_image": flattenInputLocation(in.AvailBlankingImage),
		names.AttrState:        string(in.State),
	}

	return []any{m}
}

func flattenCaptionDescriptions(tfList []types.CaptionDescription) []any {
	if len(tfList) == 0 {
		return nil
	}

	var out []any

	for _, item := range tfList {
		m := map[string]any{
			"caption_selector_name": aws.ToString(item.CaptionSelectorName),
			names.AttrName:          aws.ToString(item.Name),
			"accessibility":         string(item.Accessibility),
			"destination_settings":  flattenCaptionDescriptionsCaptionDestinationSettings(item.DestinationSettings),
			names.AttrLanguageCode:  aws.ToString(item.LanguageCode),
			"language_description":  aws.ToString(item.LanguageDescription),
		}

		out = append(out, m)
	}
	return out
}

func flattenCaptionDescriptionsCaptionDestinationSettings(in *types.CaptionDestinationSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"arib_destination_settings":                 []any{}, // attribute has no exported fields
		"burn_in_destination_settings":              flattenCaptionDescriptionsCaptionDestinationSettingsBurnInDestinationSettings(in.BurnInDestinationSettings),
		"dvb_sub_destination_settings":              flattenCaptionDescriptionsCaptionDestinationSettingsDvbSubDestinationSettings(in.DvbSubDestinationSettings),
		"ebu_tt_d_destination_settings":             flattenCaptionDescriptionsCaptionDestinationSettingsEbuTtDDestinationSettings(in.EbuTtDDestinationSettings),
		"embedded_destination_settings":             []any{}, // attribute has no exported fields
		"embedded_plus_scte20_destination_settings": []any{}, // attribute has no exported fields
		"rtmp_caption_info_destination_settings":    []any{}, // attribute has no exported fields
		"scte20_plus_embedded_destination_settings": []any{}, // attribute has no exported fields
		"scte27_destination_settings":               []any{}, // attribute has no exported fields
		"smpte_tt_destination_settings":             []any{}, // attribute has no exported fields
		"teletext_destination_settings":             flattenCaptionDescriptionsCaptionDestinationSettingsTeletextDestinationSettings(in.TeletextDestinationSettings),
		"ttml_destination_settings":                 flattenCaptionDescriptionsCaptionDestinationSettingsTtmlDestinationSettings(in.TtmlDestinationSettings),
		"webvtt_destination_settings":               flattenCaptionDescriptionsCaptionDestinationSettingsWebvttDestinationSettings(in.WebvttDestinationSettings),
	}

	return []any{m}
}

func flattenCaptionDescriptionsCaptionDestinationSettingsBurnInDestinationSettings(in *types.BurnInDestinationSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"alignment":             string(in.Alignment),
		"background_color":      string(in.BackgroundColor),
		"background_opacity":    int(aws.ToInt32(in.BackgroundOpacity)),
		"font":                  flattenInputLocation(in.Font),
		"font_color":            string(in.FontColor),
		"font_opacity":          int(aws.ToInt32(in.FontOpacity)),
		"font_resolution":       int(aws.ToInt32(in.FontResolution)),
		"font_size":             aws.ToString(in.FontSize),
		"outline_color":         string(in.OutlineColor),
		"outline_size":          int(aws.ToInt32(in.OutlineSize)),
		"shadow_color":          string(in.ShadowColor),
		"shadow_opacity":        int(aws.ToInt32(in.ShadowOpacity)),
		"shadow_x_offset":       int(aws.ToInt32(in.ShadowXOffset)),
		"shadow_y_offset":       int(aws.ToInt32(in.ShadowYOffset)),
		"teletext_grid_control": string(in.TeletextGridControl),
		"x_position":            int(aws.ToInt32(in.XPosition)),
		"y_position":            int(aws.ToInt32(in.YPosition)),
	}

	return []any{m}
}

func flattenCaptionDescriptionsCaptionDestinationSettingsDvbSubDestinationSettings(in *types.DvbSubDestinationSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"alignment":             string(in.Alignment),
		"background_color":      string(in.BackgroundColor),
		"background_opacity":    int(aws.ToInt32(in.BackgroundOpacity)),
		"font":                  flattenInputLocation(in.Font),
		"font_color":            string(in.FontColor),
		"font_opacity":          int(aws.ToInt32(in.FontOpacity)),
		"font_resolution":       int(aws.ToInt32(in.FontResolution)),
		"font_size":             aws.ToString(in.FontSize),
		"outline_color":         string(in.OutlineColor),
		"outline_size":          int(aws.ToInt32(in.OutlineSize)),
		"shadow_color":          string(in.ShadowColor),
		"shadow_opacity":        int(aws.ToInt32(in.ShadowOpacity)),
		"shadow_x_offset":       int(aws.ToInt32(in.ShadowXOffset)),
		"shadow_y_offset":       int(aws.ToInt32(in.ShadowYOffset)),
		"teletext_grid_control": string(in.TeletextGridControl),
		"x_position":            int(aws.ToInt32(in.XPosition)),
		"y_position":            int(aws.ToInt32(in.YPosition)),
	}

	return []any{m}
}

func flattenCaptionDescriptionsCaptionDestinationSettingsEbuTtDDestinationSettings(in *types.EbuTtDDestinationSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"copyright_holder": aws.ToString(in.CopyrightHolder),
		"fill_line_gap":    string(in.FillLineGap),
		"font_family":      aws.ToString(in.FontFamily),
		"style_control":    string(in.StyleControl),
	}

	return []any{m}
}

func flattenCaptionDescriptionsCaptionDestinationSettingsTeletextDestinationSettings(in *types.TeletextDestinationSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{}

	return []any{m}
}

func flattenCaptionDescriptionsCaptionDestinationSettingsTtmlDestinationSettings(in *types.TtmlDestinationSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"style_control": string(in.StyleControl),
	}

	return []any{m}
}

func flattenCaptionDescriptionsCaptionDestinationSettingsWebvttDestinationSettings(in *types.WebvttDestinationSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"style_control": string(in.StyleControl),
	}

	return []any{m}
}

func flattenGlobalConfiguration(in *types.GlobalConfiguration) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"initial_audio_gain":           int(aws.ToInt32(in.InitialAudioGain)),
		"input_end_action":             string(in.InputEndAction),
		"input_loss_behavior":          flattenGlobalConfigurationInputLossBehavior(in.InputLossBehavior),
		"output_locking_mode":          string(in.OutputLockingMode),
		"output_timing_source":         string(in.OutputTimingSource),
		"support_low_framerate_inputs": string(in.SupportLowFramerateInputs),
	}

	return []any{m}
}

func flattenGlobalConfigurationInputLossBehavior(in *types.InputLossBehavior) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"black_frame_msec":       int(aws.ToInt32(in.BlackFrameMsec)),
		"input_loss_image_color": aws.ToString(in.InputLossImageColor),
		"input_loss_image_slate": flattenInputLocation(in.InputLossImageSlate),
		"input_loss_image_type":  string(in.InputLossImageType),
		"repeat_frame_msec":      int(aws.ToInt32(in.RepeatFrameMsec)),
	}

	return []any{m}
}

func flattenMotionGraphicsConfiguration(apiObject *types.MotionGraphicsConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"motion_graphics_settings":  flattenMotionGraphicsConfigurationMotionGraphicsSettings(apiObject.MotionGraphicsSettings),
		"motion_graphics_insertion": string(apiObject.MotionGraphicsInsertion),
	}

	return []any{m}
}

func flattenMotionGraphicsConfigurationMotionGraphicsSettings(in *types.MotionGraphicsSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"html_motion_graphics_settings": []any{}, // attribute has no exported fields
	}

	return []any{m}
}

func flattenNielsenConfiguration(apiObject *types.NielsenConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"distributor_id":             aws.ToString(apiObject.DistributorId),
		"nielsen_pcm_to_id3_tagging": string(apiObject.NielsenPcmToId3Tagging),
	}

	return []any{m}
}

func flattenVideoDescriptionsCodecSettings(in *types.VideoCodecSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"frame_capture_settings": flattenCodecSettingsFrameCaptureSettings(in.FrameCaptureSettings),
		"h264_settings":          flattenCodecSettingsH264Settings(in.H264Settings),
		"h265_settings":          flattenCodecSettingsH265Settings(in.H265Settings),
	}

	return []any{m}
}

func flattenCodecSettingsFrameCaptureSettings(in *types.FrameCaptureSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"capture_interval":       int(aws.ToInt32(in.CaptureInterval)),
		"capture_interval_units": string(in.CaptureIntervalUnits),
	}

	return []any{m}
}

func flattenCodecSettingsH264Settings(in *types.H264Settings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"adaptive_quantization":   string(in.AdaptiveQuantization),
		"afd_signaling":           string(in.AfdSignaling),
		"bitrate":                 int(aws.ToInt32(in.Bitrate)),
		"buf_fill_pct":            int(aws.ToInt32(in.BufFillPct)),
		"buf_size":                int(aws.ToInt32(in.BufSize)),
		"color_metadata":          string(in.ColorMetadata),
		"entropy_encoding":        string(in.EntropyEncoding),
		"filter_settings":         flattenH264SettingsFilterSettings(in.FilterSettings),
		"fixed_afd":               string(in.FixedAfd),
		"flicker_aq":              string(in.FlickerAq),
		"force_field_pictures":    string(in.ForceFieldPictures),
		"framerate_control":       string(in.FramerateControl),
		"framerate_denominator":   int(aws.ToInt32(in.FramerateDenominator)),
		"framerate_numerator":     int(aws.ToInt32(in.FramerateNumerator)),
		"gop_b_reference":         string(in.GopBReference),
		"gop_closed_cadence":      int(aws.ToInt32(in.GopClosedCadence)),
		"gop_num_b_frames":        int(aws.ToInt32(in.GopNumBFrames)),
		"gop_size":                in.GopSize,
		"gop_size_units":          string(in.GopSizeUnits),
		"level":                   string(in.Level),
		"look_ahead_rate_control": string(in.LookAheadRateControl),
		"max_bitrate":             int(aws.ToInt32(in.MaxBitrate)),
		"min_i_interval":          int(aws.ToInt32(in.MinIInterval)),
		"num_ref_frames":          int(aws.ToInt32(in.NumRefFrames)),
		"par_control":             string(in.ParControl),
		"par_denominator":         int(aws.ToInt32(in.ParDenominator)),
		"par_numerator":           int(aws.ToInt32(in.ParNumerator)),
		names.AttrProfile:         string(in.Profile),
		"quality_level":           string(in.QualityLevel),
		"qvbr_quality_level":      int(aws.ToInt32(in.QvbrQualityLevel)),
		"rate_control_mode":       string(in.RateControlMode),
		"scan_type":               string(in.ScanType),
		"scene_change_detect":     string(in.SceneChangeDetect),
		"slices":                  int(aws.ToInt32(in.Slices)),
		"spatial_aq":              string(in.SpatialAq),
		"subgop_length":           string(in.SubgopLength),
		"syntax":                  string(in.Syntax),
		"temporal_aq":             string(in.TemporalAq),
		"timecode_insertion":      string(in.TimecodeInsertion),
	}

	return []any{m}
}

func flattenH264SettingsFilterSettings(in *types.H264FilterSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"temporal_filter_settings": flattenFilterSettingsTemporalFilterSettings(in.TemporalFilterSettings),
	}

	return []any{m}
}

func flattenFilterSettingsTemporalFilterSettings(in *types.TemporalFilterSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"post_filter_sharpening": string(in.PostFilterSharpening),
		"strength":               string(in.Strength),
	}

	return []any{m}
}

func flattenCodecSettingsH265Settings(in *types.H265Settings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"framerate_denominator":         int(aws.ToInt32(in.FramerateDenominator)),
		"framerate_numerator":           int(aws.ToInt32(in.FramerateNumerator)),
		"adaptive_quantization":         string(in.AdaptiveQuantization),
		"afd_signaling":                 string(in.AfdSignaling),
		"alternative_transfer_function": string(in.AlternativeTransferFunction),
		"bitrate":                       int(aws.ToInt32(in.Bitrate)),
		"buf_size":                      int(aws.ToInt32(in.BufSize)),
		"color_metadata":                string(in.ColorMetadata),
		"color_space_settings":          flattenH265ColorSpaceSettings(in.ColorSpaceSettings),
		"filter_settings":               flattenH265FilterSettings(in.FilterSettings),
		"fixed_afd":                     string(in.FixedAfd),
		"flicker_aq":                    string(in.FlickerAq),
		"gop_closed_cadence":            int(aws.ToInt32(in.GopClosedCadence)),
		"gop_size":                      in.GopSize,
		"gop_size_units":                string(in.GopSizeUnits),
		"level":                         string(in.Level),
		"look_ahead_rate_control":       string(in.LookAheadRateControl),
		"max_bitrate":                   int(aws.ToInt32(in.MaxBitrate)),
		"min_i_interval":                int(aws.ToInt32(in.MinIInterval)),
		"min_qp":                        int(aws.ToInt32(in.MinQp)),
		"mv_over_picture_boundaries":    string(in.MvOverPictureBoundaries),
		"mv_temporal_predictor":         string(in.MvTemporalPredictor),
		"par_denominator":               int(aws.ToInt32(in.ParDenominator)),
		"par_numerator":                 int(aws.ToInt32(in.ParNumerator)),
		names.AttrProfile:               string(in.Profile),
		"qvbr_quality_level":            int(aws.ToInt32(in.QvbrQualityLevel)),
		"rate_control_mode":             string(in.RateControlMode),
		"scan_type":                     string(in.ScanType),
		"scene_change_detect":           string(in.SceneChangeDetect),
		"slices":                        int(aws.ToInt32(in.Slices)),
		"tier":                          string(in.Tier),
		"tile_height":                   int(aws.ToInt32(in.TileHeight)),
		"tile_padding":                  string(in.TilePadding),
		"tile_width":                    int(aws.ToInt32(in.TileWidth)),
		"timecode_burnin_settings":      flattenH265TimecodeBurninSettings(in.TimecodeBurninSettings),
		"timecode_insertion":            string(in.TimecodeInsertion),
		"treeblock_size":                string(in.TreeblockSize),
	}
	return []any{m}
}

func flattenH265ColorSpaceSettings(in *types.H265ColorSpaceSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{}
	if in.ColorSpacePassthroughSettings != nil {
		m["color_space_passthrough_settings"] = []any{} // no exported fields
	}
	if in.DolbyVision81Settings != nil {
		m["dolby_vision81_settings"] = []any{} // no exported fields
	}
	if in.Hdr10Settings != nil {
		m["hdr10_settings"] = flattenH265Hdr10Settings(in.Hdr10Settings)
	}
	if in.Rec601Settings != nil {
		m["rec601_settings"] = []any{} // no exported fields
	}
	if in.Rec709Settings != nil {
		m["rec709_settings"] = []any{} // no exported fields
	}

	return []any{m}
}

func flattenH265Hdr10Settings(in *types.Hdr10Settings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"max_cll":  int(aws.ToInt32(in.MaxCll)),
		"max_fall": int(aws.ToInt32(in.MaxFall)),
	}

	return []any{m}
}

func flattenH265FilterSettings(in *types.H265FilterSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"temporal_filter_settings": flattenH265FilterSettingsTemporalFilterSettings(in.TemporalFilterSettings),
	}

	return []any{m}
}

func flattenH265FilterSettingsTemporalFilterSettings(in *types.TemporalFilterSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"post_filter_sharpening": in.PostFilterSharpening,
		"strength":               string(in.Strength),
	}

	return []any{m}
}

func flattenH265TimecodeBurninSettings(in *types.TimecodeBurninSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"timecode_burnin_font_size": string(in.FontSize),
		"timecode_burnin_position":  string(in.Position),
		names.AttrPrefix:            in.Prefix,
	}

	return []any{m}
}

func flattenAudioNormalization(ns *types.AudioNormalizationSettings) []any {
	if ns == nil {
		return nil
	}

	m := map[string]any{
		"algorithm":         ns.Algorithm,
		"algorithm_control": ns.AlgorithmControl,
		"target_lkfs":       ns.TargetLkfs,
	}

	return []any{m}
}

func flattenAudioWatermarkSettings(ns *types.AudioWatermarkSettings) []any {
	if ns == nil {
		return nil
	}

	m := map[string]any{
		"nielsen_watermark_settings": func(n *types.NielsenWatermarksSettings) []any {
			if n == nil {
				return nil
			}

			m := map[string]any{
				"nielsen_distribution_type":   string(n.NielsenDistributionType),
				"nielsen_cbet_settings":       flattenNielsenCbetSettings(n.NielsenCbetSettings),
				"nielsen_naes_ii_nw_settings": flattenNielsenNaesIiNwSettings(n.NielsenNaesIiNwSettings),
			}

			return []any{m}
		}(ns.NielsenWatermarksSettings),
	}

	return []any{m}
}

func flattenAudioDescriptionsCodecSettings(in *types.AudioCodecSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"aac_settings":        flattenCodecSettingsAacSettings(in.AacSettings),
		"ac3_settings":        flattenCodecSettingsAc3Settings(in.Ac3Settings),
		"eac3_atmos_settings": flattenCodecSettingsEac3AtmosSettings(in.Eac3AtmosSettings),
		"eac3_settings":       flattenCodecSettingsEac3Settings(in.Eac3Settings),
		"mp2_settings":        flattenCodecSettingsMp2Settings(in.Mp2Settings),
		"wav_settings":        flattenCodecSettingsWavSettings(in.WavSettings),
	}

	if in.PassThroughSettings != nil {
		m["pass_through_settings"] = []any{} // no exported fields
	}

	return []any{m}
}

func flattenCodecSettingsAacSettings(in *types.AacSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"bitrate":           in.Bitrate,
		"coding_mode":       string(in.CodingMode),
		"input_type":        string(in.InputType),
		names.AttrProfile:   string(in.Profile),
		"rate_control_mode": string(in.RateControlMode),
		"raw_format":        string(in.RawFormat),
		"sample_rate":       in.SampleRate,
		"spec":              string(in.Spec),
		"vbr_quality":       string(in.VbrQuality),
	}

	return []any{m}
}

func flattenCodecSettingsAc3Settings(in *types.Ac3Settings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"bitrate":          in.Bitrate,
		"bitstream_mode":   string(in.BitstreamMode),
		"coding_mode":      string(in.CodingMode),
		"dialnorm":         int(aws.ToInt32(in.Dialnorm)),
		"drc_profile":      string(in.DrcProfile),
		"lfe_filter":       string(in.LfeFilter),
		"metadata_control": string(in.MetadataControl),
	}

	return []any{m}
}

func flattenCodecSettingsEac3AtmosSettings(in *types.Eac3AtmosSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"bitrate":       float32(aws.ToFloat64(in.Bitrate)),
		"coding_mode":   string(in.CodingMode),
		"dialnorm":      int(aws.ToInt32(in.Dialnorm)),
		"drc_line":      string(in.DrcLine),
		"drc_rf":        string(in.DrcRf),
		"height_trim":   float32(aws.ToFloat64(in.HeightTrim)),
		"surround_trim": float32(aws.ToFloat64(in.SurroundTrim)),
	}

	return []any{m}
}

func flattenCodecSettingsEac3Settings(in *types.Eac3Settings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"attenuation_control":      string(in.AttenuationControl),
		"bitrate":                  float32(aws.ToFloat64(in.Bitrate)),
		"bitstream_mode":           string(in.BitstreamMode),
		"coding_mode":              string(in.CodingMode),
		"dc_filter":                string(in.DcFilter),
		"dialnorm":                 int(aws.ToInt32(in.Dialnorm)),
		"drc_line":                 string(in.DrcLine),
		"drc_rf":                   string(in.DrcRf),
		"lfe_control":              string(in.LfeControl),
		"lfe_filter":               string(in.LfeFilter),
		"lo_ro_center_mix_level":   float32(aws.ToFloat64(in.LoRoCenterMixLevel)),
		"lo_ro_surround_mix_level": float32(aws.ToFloat64(in.LoRoSurroundMixLevel)),
		"lt_rt_center_mix_level":   float32(aws.ToFloat64(in.LtRtCenterMixLevel)),
		"lt_rt_surround_mix_level": float32(aws.ToFloat64(in.LtRtSurroundMixLevel)),
		"metadata_control":         string(in.MetadataControl),
		"passthrough_control":      string(in.PassthroughControl),
		"phase_control":            string(in.PhaseControl),
		"stereo_downmix":           string(in.StereoDownmix),
		"surround_ex_mode":         string(in.SurroundExMode),
		"surround_mode":            string(in.SurroundMode),
	}

	return []any{m}
}

func flattenCodecSettingsMp2Settings(in *types.Mp2Settings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"bitrate":     float32(aws.ToFloat64(in.Bitrate)),
		"coding_mode": string(in.CodingMode),
		"sample_rate": float32(aws.ToFloat64(in.SampleRate)),
	}

	return []any{m}
}

func flattenCodecSettingsWavSettings(in *types.WavSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"bit_depth":   float32(aws.ToFloat64(in.BitDepth)),
		"coding_mode": string(in.CodingMode),
		"sample_rate": float32(aws.ToFloat64(in.SampleRate)),
	}

	return []any{m}
}

func flattenAudioDescriptionsRemixSettings(in *types.RemixSettings) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"channel_mappings": flattenChannelMappings(in.ChannelMappings),
		"channels_in":      int(aws.ToInt32(in.ChannelsIn)),
		"channels_out":     int(aws.ToInt32(in.ChannelsOut)),
	}

	return []any{m}
}

func flattenChannelMappings(in []types.AudioChannelMapping) []any {
	if len(in) == 0 {
		return nil
	}

	var out []any
	for _, item := range in {
		m := map[string]any{
			"input_channel_levels": flattenInputChannelLevels(item.InputChannelLevels),
			"output_channel":       int(aws.ToInt32(item.OutputChannel)),
		}

		out = append(out, m)
	}

	return out
}

func flattenInputChannelLevels(in []types.InputChannelLevel) []any {
	if len(in) == 0 {
		return nil
	}

	var out []any
	for _, item := range in {
		m := map[string]any{
			"gain":          int(aws.ToInt32(item.Gain)),
			"input_channel": int(aws.ToInt32(item.InputChannel)),
		}

		out = append(out, m)
	}

	return out
}

func flattenNielsenCbetSettings(in *types.NielsenCBET) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"cbet_check_digit_string": aws.ToString(in.CbetCheckDigitString),
		"cbet_stepaside":          string(in.CbetStepaside),
		"csid":                    aws.ToString(in.Csid),
	}

	return []any{m}
}

func flattenNielsenNaesIiNwSettings(in *types.NielsenNaesIiNw) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"check_digit_string": aws.ToString(in.CheckDigitString),
		"sid":                float32(aws.ToFloat64(in.Sid)),
	}

	return []any{m}
}
