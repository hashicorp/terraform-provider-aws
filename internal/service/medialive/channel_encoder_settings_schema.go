package medialive

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"audio_normalization_settings": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
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
											Computed: true,
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
													"profile": {
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
											Computed: true,
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
											Computed: true,
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
											Computed: true,
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
											Computed: true,
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
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{}, // no exported elements in this list
											},
										},
										"wav_settings": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
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
							"language_code": {
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
								Computed: true,
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
													"destination": func() *schema.Schema {
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
													"destination": func() *schema.Schema {
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
													"destination": func() *schema.Schema {
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
																"language_code": {
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
													"mode": {
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
													"destination": func() *schema.Schema {
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
													"destination": func() *schema.Schema {
														return destinationSchema()
													}(),
													"acquisition_point_id": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"audio_only_timecodec_control": {
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
														Type:     schema.TypeInt,
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
							"name": {
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
							"source": {
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
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
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
													"profile": {
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
										// TODO h265_settings
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
							"state": {
								Type:     schema.TypeString,
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				// TODO avail_configuration
				// TODO blackout_slate
				// TODO caption_descriptions
				// TODO feature_activations
				// TODO global_configuration
				// TODO motion_graphics_configuration
				// TODO nielsen_configuration
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
							"destination": destinationSchema(),
						},
					},
				},
				"rtmp_output_settings": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"destination": destinationSchema(),
							"certficate_mode": {
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
							"destination": destinationSchema(),
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
							"service_name": {
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

func expandChannelEncoderSettings(tfList []interface{}) *types.EncoderSettings {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	var settings types.EncoderSettings
	if v, ok := m["audio_descriptions"].(*schema.Set); ok && v.Len() > 0 {
		settings.AudioDescriptions = expandChannelEncoderSettingsAudioDescriptions(v.List())
	}
	if v, ok := m["output_groups"].([]interface{}); ok && len(v) > 0 {
		settings.OutputGroups = expandChannelEncoderSettingsOutputGroups(v)
	}
	if v, ok := m["timecode_config"].([]interface{}); ok && len(v) > 0 {
		settings.TimecodeConfig = expandChannelEncoderSettingsTimecodeConfig(v)
	}
	if v, ok := m["video_descriptions"].(*schema.Set); ok && v.Len() > 0 {
		settings.VideoDescriptions = expandChannelEncoderSettingsVideoDescriptions(v.List())
	}
	if v, ok := m["avail_blanking"].([]interface{}); ok && len(v) > 0 {
		settings.AvailBlanking = expandChannelEncoderSettingsAvailBlanking(v)
	}
	if v, ok := m["avail_configuration"].([]interface{}); ok && len(v) > 0 {
		settings.AvailConfiguration = nil // TODO expandChannelEncoderSettingsAvailConfiguration(v)
	}
	if v, ok := m["blackout_slate"].([]interface{}); ok && len(v) > 0 {
		settings.BlackoutSlate = nil // TODO expandChannelEncoderSettingsBlackoutSlate(v)
	}
	if v, ok := m["caption_descriptions"].([]interface{}); ok && len(v) > 0 {
		settings.CaptionDescriptions = nil // TODO expandChannelEncoderSettingsCaptionDescriptions(v)
	}
	if v, ok := m["feature_activations"].([]interface{}); ok && len(v) > 0 {
		settings.FeatureActivations = nil // TODO expandChannelEncoderSettingsFeatureActivations(v)
	}
	if v, ok := m["global_configuration"].([]interface{}); ok && len(v) > 0 {
		settings.GlobalConfiguration = nil // TODO expandChannelEncoderSettingsGlobalConfiguration(v)
	}
	if v, ok := m["motion_graphics_configuration"].([]interface{}); ok && len(v) > 0 {
		settings.MotionGraphicsConfiguration = nil // TODO expandChannelEncoderSettingsMotionGraphicsConfiguration(v)
	}
	if v, ok := m["nielsen_configuration"].([]interface{}); ok && len(v) > 0 {
		settings.NielsenConfiguration = nil // TODO expandChannelEncoderSettingsNielsenConfiguration(v)
	}

	return &settings
}

func expandChannelEncoderSettingsAudioDescriptions(tfList []interface{}) []types.AudioDescription {
	if tfList == nil {
		return nil
	}

	var audioDesc []types.AudioDescription
	for _, tfItem := range tfList {
		m, ok := tfItem.(map[string]interface{})
		if !ok {
			continue
		}

		var a types.AudioDescription
		if v, ok := m["audio_selector_name"].(string); ok && v != "" {
			a.AudioSelectorName = aws.String(v)
		}
		if v, ok := m["name"].(string); ok && v != "" {
			a.Name = aws.String(v)
		}
		if v, ok := m["audio_normalization_settings"].([]interface{}); ok && len(v) > 0 {
			a.AudioNormalizationSettings = expandAudioDescriptionsAudioNormalizationSettings(v)
		}
		if v, ok := m["audio_type"].(string); ok && v != "" {
			a.AudioType = types.AudioType(v)
		}
		if v, ok := m["audio_type_control"].(string); ok && v != "" {
			a.AudioTypeControl = types.AudioDescriptionAudioTypeControl(v)
		}
		if v, ok := m["audio_watermark_settings"].([]interface{}); ok && len(v) > 0 {
			a.AudioWatermarkingSettings = expandAudioWatermarkSettings(v)
		}
		if v, ok := m["codec_settings"].([]interface{}); ok && len(v) > 0 {
			a.CodecSettings = expandChannelEncoderSettingsAudioDescriptionsCodecSettings(v)
		}
		if v, ok := m["language_code"].(string); ok && v != "" {
			a.LanguageCode = aws.String(v)
		}
		if v, ok := m["language_code_control"].(string); ok && v != "" {
			a.LanguageCodeControl = types.AudioDescriptionLanguageCodeControl(v)
		}
		if v, ok := m["remix_settings"].([]interface{}); ok && len(v) > 0 {
			a.RemixSettings = expandChannelEncoderSettingsAudioDescriptionsRemixSettings(v)
		}
		if v, ok := m["stream_name"].(string); ok && v != "" {
			a.StreamName = aws.String(v)
		}

		audioDesc = append(audioDesc, a)
	}

	return audioDesc
}

func expandChannelEncoderSettingsOutputGroups(tfList []interface{}) []types.OutputGroup {
	if tfList == nil {
		return nil
	}

	var outputGroups []types.OutputGroup
	for _, tfItem := range tfList {
		m, ok := tfItem.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.OutputGroup
		if v, ok := m["output_group_settings"].([]interface{}); ok && len(v) > 0 {
			o.OutputGroupSettings = expandChannelEncoderSettingsOutputGroupsOutputGroupSettings(v)
		}
		if v, ok := m["outputs"].([]interface{}); ok && len(v) > 0 {
			o.Outputs = expandChannelEncoderSettingsOutputGroupsOutputs(v)
		}
		if v, ok := m["name"].(string); ok && v != "" {
			o.Name = aws.String(v)
		}

		outputGroups = append(outputGroups, o)
	}

	return outputGroups
}

func expandAudioDescriptionsAudioNormalizationSettings(tfList []interface{}) *types.AudioNormalizationSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioNormalizationSettings
	if v, ok := m["algorithm"].(string); ok && v != "" {
		out.Algorithm = types.AudioNormalizationAlgorithm(v)
	}
	if v, ok := m["algorithm_control"].(string); ok && v != "" {
		out.AlgorithmControl = types.AudioNormalizationAlgorithmControl(v)
	}
	if v, ok := m["target_lkfs"].(float32); ok {
		out.TargetLkfs = float64(v)
	}

	return &out
}

func expandChannelEncoderSettingsAudioDescriptionsCodecSettings(tfList []interface{}) *types.AudioCodecSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioCodecSettings
	if v, ok := m["aac_settings"].([]interface{}); ok && len(v) > 0 {
		out.AacSettings = expandAudioDescriptionsCodecSettingsAacSettings(v)
	}
	if v, ok := m["ac3_settings"].([]interface{}); ok && len(v) > 0 {
		out.Ac3Settings = expandAudioDescriptionsCodecSettingsAc3Settings(v)
	}
	if v, ok := m["eac3_atmos_settings"].([]interface{}); ok && len(v) > 0 {
		out.Eac3AtmosSettings = expandAudioDescriptionsCodecSettingsEac3AtmosSettings(v)
	}
	if v, ok := m["eac3_settings"].([]interface{}); ok && len(v) > 0 {
		out.Eac3Settings = expandAudioDescriptionsCodecSettingsEac3Settings(v)
	}
	if v, ok := m["vp2_settings"].([]interface{}); ok && len(v) > 0 {
		out.Mp2Settings = expandAudioDescriptionsCodecSettingsMp2Settings(v)
	}
	if v, ok := m["pass_through_settings"].([]interface{}); ok && len(v) > 0 {
		out.PassThroughSettings = &types.PassThroughSettings{} // no exported fields
	}
	if v, ok := m["wav_settings"].([]interface{}); ok && len(v) > 0 {
		out.WavSettings = expandAudioDescriptionsCodecSettingsWavSettings(v)
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsAacSettings(tfList []interface{}) *types.AacSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AacSettings
	if v, ok := m["bitrate"].(float32); ok {
		out.Bitrate = float64(v)
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.AacCodingMode(v)
	}
	if v, ok := m["input_type"].(string); ok && v != "" {
		out.InputType = types.AacInputType(v)
	}
	if v, ok := m["profile"].(string); ok && v != "" {
		out.Profile = types.AacProfile(v)
	}
	if v, ok := m["rate_control_mode"].(string); ok && v != "" {
		out.RateControlMode = types.AacRateControlMode(v)
	}
	if v, ok := m["raw_format"].(string); ok && v != "" {
		out.RawFormat = types.AacRawFormat(v)
	}
	if v, ok := m["sample_rate"].(float32); ok {
		out.SampleRate = float64(v)
	}
	if v, ok := m["spec"].(string); ok && v != "" {
		out.Spec = types.AacSpec(v)
	}
	if v, ok := m["vbr_quality"].(string); ok && v != "" {
		out.VbrQuality = types.AacVbrQuality(v)
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsAc3Settings(tfList []interface{}) *types.Ac3Settings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.Ac3Settings
	if v, ok := m["bitrate"].(float32); ok {
		out.Bitrate = float64(v)
	}
	if v, ok := m["bitstream_mode"].(string); ok && v != "" {
		out.BitstreamMode = types.Ac3BitstreamMode(v)
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.Ac3CodingMode(v)
	}
	if v, ok := m["dialnorm"].(int); ok {
		out.Dialnorm = int32(v)
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

func expandAudioDescriptionsCodecSettingsEac3AtmosSettings(tfList []interface{}) *types.Eac3AtmosSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.Eac3AtmosSettings
	if v, ok := m["bitrate"].(float32); ok {
		out.Bitrate = float64(v)
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.Eac3AtmosCodingMode(v)
	}
	if v, ok := m["dialnorm"].(int); ok {
		out.Dialnorm = int32(v)
	}
	if v, ok := m["drc_line"].(string); ok && v != "" {
		out.DrcLine = types.Eac3AtmosDrcLine(v)
	}
	if v, ok := m["drc_rf"].(string); ok && v != "" {
		out.DrcRf = types.Eac3AtmosDrcRf(v)
	}
	if v, ok := m["height_trim"].(float32); ok {
		out.HeightTrim = float64(v)
	}
	if v, ok := m["surround_trim"].(float32); ok {
		out.SurroundTrim = float64(v)
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsEac3Settings(tfList []interface{}) *types.Eac3Settings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.Eac3Settings
	if v, ok := m["attenuation_control"].(string); ok && v != "" {
		out.AttenuationControl = types.Eac3AttenuationControl(v)
	}
	if v, ok := m["bitrate"].(float32); ok {
		out.Bitrate = float64(v)
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
	if v, ok := m["dialnorm"].(int); ok {
		out.Dialnorm = int32(v)
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
	if v, ok := m["lo_ro_center_mix_level"].(float32); ok {
		out.LoRoCenterMixLevel = float64(v)
	}
	if v, ok := m["lo_ro_surround_mix_level"].(float32); ok {
		out.LoRoSurroundMixLevel = float64(v)
	}
	if v, ok := m["lt_rt_center_mix_level"].(float32); ok {
		out.LtRtCenterMixLevel = float64(v)
	}
	if v, ok := m["lt_rt_surround_mix_level"].(float32); ok {
		out.LtRtSurroundMixLevel = float64(v)
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

func expandAudioDescriptionsCodecSettingsMp2Settings(tfList []interface{}) *types.Mp2Settings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.Mp2Settings
	if v, ok := m["bitrate"].(float32); ok {
		out.Bitrate = float64(v)
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.Mp2CodingMode(v)
	}
	if v, ok := m["sample_rate"].(float32); ok {
		out.Bitrate = float64(v)
	}

	return &out
}

func expandAudioDescriptionsCodecSettingsWavSettings(tfList []interface{}) *types.WavSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.WavSettings
	if v, ok := m["bit_depth"].(float32); ok {
		out.BitDepth = float64(v)
	}
	if v, ok := m["coding_mode"].(string); ok && v != "" {
		out.CodingMode = types.WavCodingMode(v)
	}
	if v, ok := m["sample_rate"].(float32); ok {
		out.SampleRate = float64(v)
	}

	return &out
}

func expandChannelEncoderSettingsAudioDescriptionsRemixSettings(tfList []interface{}) *types.RemixSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.RemixSettings
	if v, ok := m["channel_mappings"].(*schema.Set); ok && v.Len() > 0 {
		out.ChannelMappings = expandChannelMappings(v.List())
	}
	if v, ok := m["channels_in"].(int); ok {
		out.ChannelsIn = int32(v)
	}
	if v, ok := m["channels_out"].(int); ok {
		out.ChannelsOut = int32(v)
	}

	return &out
}

func expandChannelMappings(tfList []interface{}) []types.AudioChannelMapping {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.AudioChannelMapping
	for _, item := range tfList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.AudioChannelMapping
		if v, ok := m["input_channel_levels"].(*schema.Set); ok && v.Len() > 0 {
			o.InputChannelLevels = expandInputChannelLevels(v.List())
		}
		if v, ok := m["output_channel"].(int); ok {
			o.OutputChannel = int32(v)
		}

		out = append(out, o)
	}

	return out
}

func expandInputChannelLevels(tfList []interface{}) []types.InputChannelLevel {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.InputChannelLevel
	for _, item := range tfList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.InputChannelLevel
		if v, ok := m["gain"].(int); ok {
			o.Gain = int32(v)
		}
		if v, ok := m["input_channel"].(int); ok {
			o.InputChannel = int32(v)
		}

		out = append(out, o)
	}

	return out
}

func expandChannelEncoderSettingsOutputGroupsOutputGroupSettings(tfList []interface{}) *types.OutputGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var o types.OutputGroupSettings

	if v, ok := m["archive_group_settings"].([]interface{}); ok && len(v) > 0 {
		o.ArchiveGroupSettings = expandArchiveGroupSettings(v)
	}
	if v, ok := m["frame_capture_group_settings"].([]interface{}); ok && len(v) > 0 {
		o.FrameCaptureGroupSettings = expandFrameCaptureGroupSettings(v)
	}
	if v, ok := m["hls_group_settings"].([]interface{}); ok && len(v) > 0 {
		o.HlsGroupSettings = expandHLSGroupSettings(v)
	}
	if v, ok := m["ms_smooth_group_settings"].([]interface{}); ok && len(v) > 0 {
		o.MsSmoothGroupSettings = expandMsSmoothGroupSettings(v)
	}
	if v, ok := m["media_package_group_settings"].([]interface{}); ok && len(v) > 0 {
		o.MediaPackageGroupSettings = expandMediaPackageGroupSettings(v)
	}
	if v, ok := m["multiplex_group_settings"].([]interface{}); ok && len(v) > 0 {
		o.MultiplexGroupSettings = &types.MultiplexGroupSettings{} // only unexported fields
	}
	if v, ok := m["rtmp_group_settings"].([]interface{}); ok && len(v) > 0 {
		o.RtmpGroupSettings = expandRtmpGroupSettings(v)
	}
	if v, ok := m["udp_group_settings"].([]interface{}); ok && len(v) > 0 {
		o.UdpGroupSettings = expandUdpGroupSettings(v)
	}

	return &o
}

func expandDestination(in []interface{}) *types.OutputLocationRef {
	if len(in) == 0 {
		return nil
	}

	m := in[0].(map[string]interface{})

	var out types.OutputLocationRef
	if v, ok := m["destination_ref_id"].(string); ok && v != "" {
		out.DestinationRefId = aws.String(v)
	}

	return &out
}

func expandMediaPackageGroupSettings(tfList []interface{}) *types.MediaPackageGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var o types.MediaPackageGroupSettings

	if v, ok := m["destination"].([]interface{}); ok && len(v) > 0 {
		o.Destination = expandDestination(v)
	}

	return &o
}

func expandArchiveGroupSettings(tfList []interface{}) *types.ArchiveGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var o types.ArchiveGroupSettings

	if v, ok := m["destination"].([]interface{}); ok && len(v) > 0 {
		o.Destination = expandDestination(v)
	}
	if v, ok := m["archive_cdn_settings"].([]interface{}); ok && len(v) > 0 {
		o.ArchiveCdnSettings = expandArchiveCDNSettings(v)
	}
	if v, ok := m["rollover_interval"].(int); ok {
		o.RolloverInterval = int32(v)
	}

	return &o
}

func expandFrameCaptureGroupSettings(tfList []interface{}) *types.FrameCaptureGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.FrameCaptureGroupSettings
	if v, ok := m["destination"].([]interface{}); ok && len(v) > 0 {
		out.Destination = expandDestination(v)
	}
	if v, ok := m["frame_capture_cdn_settings"].([]interface{}); ok && len(v) > 0 {
		out.FrameCaptureCdnSettings = expandFrameCaptureCDNSettings(v)
	}
	return &out
}

func expandFrameCaptureCDNSettings(tfList []interface{}) *types.FrameCaptureCdnSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.FrameCaptureCdnSettings
	if v, ok := m["frame_capture_s3_settings"].([]interface{}); ok && len(v) > 0 {
		out.FrameCaptureS3Settings = expandFrameCaptureS3Settings(v)
	}

	return &out
}

func expandFrameCaptureS3Settings(tfList []interface{}) *types.FrameCaptureS3Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.FrameCaptureS3Settings
	if v, ok := m["canned_acl"].(string); ok && v != "" {
		out.CannedAcl = types.S3CannedAcl(v)
	}

	return &out
}

func expandHLSGroupSettings(tfList []interface{}) *types.HlsGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsGroupSettings
	if v, ok := m["destination"].([]interface{}); ok && len(v) > 0 {
		out.Destination = expandDestination(v)
	}
	if v, ok := m["ad_markers"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := m["hls_cdn_setting"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := m["index_n_segments"].(int); ok {
		out.IndexNSegments = int32(v)
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
	if v, ok := m["keep_segments"].(int); ok {
		out.KeepSegments = int32(v)
	}
	if v, ok := m["key_format"].(string); ok && v != "" {
		out.KeyFormat = aws.String(v)
	}
	if v, ok := m["key_format_versions"].(string); ok && v != "" {
		out.KeyFormatVersions = aws.String(v)
	}
	if v, ok := m["key_provider_settings"].([]interface{}); ok && len(v) > 0 {
		out.KeyProviderSettings = expandHLSGroupSettingsKeyProviderSettings(v)
	}
	if v, ok := m["manifest_compression"].(string); ok && v != "" {
		out.ManifestCompression = types.HlsManifestCompression(v)
	}
	if v, ok := m["manifest_duration_format"].(string); ok && v != "" {
		out.ManifestDurationFormat = types.HlsManifestDurationFormat(v)
	}
	if v, ok := m["min_segment_length"].(int); ok {
		out.MinSegmentLength = int32(v)
	}
	if v, ok := m["mode"].(string); ok && v != "" {
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
	if v, ok := m["program_date_time_period"].(int); ok {
		out.ProgramDateTimePeriod = int32(v)
	}
	if v, ok := m["redundant_manifest"].(string); ok && v != "" {
		out.RedundantManifest = types.HlsRedundantManifest(v)
	}
	if v, ok := m["segment_length"].(int); ok {
		out.SegmentLength = int32(v)
	}
	if v, ok := m["segments_per_subdirectory"].(int); ok {
		out.SegmentsPerSubdirectory = int32(v)
	}
	if v, ok := m["stream_inf_resolution"].(string); ok && v != "" {
		out.StreamInfResolution = types.HlsStreamInfResolution(v)
	}
	if v, ok := m["timed_metadata_id3_frame"].(string); ok && v != "" {
		out.TimedMetadataId3Frame = types.HlsTimedMetadataId3Frame(v)
	}
	if v, ok := m["timed_metadata_id3_period"].(int); ok {
		out.TimedMetadataId3Period = int32(v)
	}
	if v, ok := m["timestamp_delta_milliseconds"].(int); ok {
		out.TimestampDeltaMilliseconds = int32(v)
	}
	if v, ok := m["ts_file_mode"].(string); ok && v != "" {
		out.TsFileMode = types.HlsTsFileMode(v)
	}

	return &out
}

func expandMsSmoothGroupSettings(tfList []interface{}) *types.MsSmoothGroupSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.MsSmoothGroupSettings
	if v, ok := m["destination"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := m["connection_retry_interval"].(int); ok {
		out.ConnectionRetryInterval = int32(v)
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
	if v, ok := m["filecache_duration"].(int); ok {
		out.FilecacheDuration = int32(v)
	}
	if v, ok := m["fragment_length"].(int); ok {
		out.FragmentLength = int32(v)
	}
	if v, ok := m["input_loss_action"].(string); ok && v != "" {
		out.InputLossAction = types.InputLossActionForMsSmoothOut(v)
	}
	if v, ok := m["num_retries"].(int); ok {
		out.NumRetries = int32(v)
	}
	if v, ok := m["restart_delay"].(int); ok {
		out.RestartDelay = int32(v)
	}
	if v, ok := m["segmentation_mode"].(string); ok && v != "" {
		out.SegmentationMode = types.SmoothGroupSegmentationMode(v)
	}
	if v, ok := m["send_delay_ms"].(int); ok {
		out.SendDelayMs = int32(v)
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

func expandHLSCDNSettings(tfList []interface{}) *types.HlsCdnSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsCdnSettings
	if v, ok := m["hls_akamai_setting"].([]interface{}); ok && len(v) > 0 {
		out.HlsAkamaiSettings = expandHSLAkamaiSettings(v)
	}
	if v, ok := m["hls_basic_put_setting"].([]interface{}); ok && len(v) > 0 {
		out.HlsBasicPutSettings = expandHSLBasicPutSettings(v)
	}
	if v, ok := m["hls_media_store_setting"].([]interface{}); ok && len(v) > 0 {
		out.HlsMediaStoreSettings = expandHLSMediaStoreSettings(v)
	}
	if v, ok := m["hls_s3_setting"].([]interface{}); ok && len(v) > 0 {
		out.HlsS3Settings = expandHSLS3Settings(v)
	}
	if v, ok := m["hls_webdav_setting"].([]interface{}); ok && len(v) > 0 {
		out.HlsWebdavSettings = expandHLSWebdavSettings(v)
	}
	return &out
}

func expandHSLAkamaiSettings(tfList []interface{}) *types.HlsAkamaiSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsAkamaiSettings
	if v, ok := m["connection_retry_interval"].(int); ok {
		out.ConnectionRetryInterval = int32(v)
	}
	if v, ok := m["filecache_duration"].(int); ok {
		out.FilecacheDuration = int32(v)
	}
	if v, ok := m["http_transfer_mode"].(string); ok && v != "" {
		out.HttpTransferMode = types.HlsAkamaiHttpTransferMode(v)
	}
	if v, ok := m["num_retries"].(int); ok {
		out.NumRetries = int32(v)
	}
	if v, ok := m["restart_delay"].(int); ok {
		out.RestartDelay = int32(v)
	}
	if v, ok := m["salt"].(string); ok && v != "" {
		out.Salt = aws.String(v)
	}
	if v, ok := m["token"].(string); ok && v != "" {
		out.Token = aws.String(v)
	}

	return &out
}

func expandHSLBasicPutSettings(tfList []interface{}) *types.HlsBasicPutSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsBasicPutSettings
	if v, ok := m["connection_retry_interval"].(int); ok {
		out.ConnectionRetryInterval = int32(v)
	}
	if v, ok := m["filecache_duration"].(int); ok {
		out.FilecacheDuration = int32(v)
	}
	if v, ok := m["num_retries"].(int); ok {
		out.NumRetries = int32(v)
	}
	if v, ok := m["restart_delay"].(int); ok {
		out.RestartDelay = int32(v)
	}

	return &out
}

func expandHLSMediaStoreSettings(tfList []interface{}) *types.HlsMediaStoreSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsMediaStoreSettings
	if v, ok := m["connection_retry_interval"].(int); ok {
		out.ConnectionRetryInterval = int32(v)
	}
	if v, ok := m["filecache_duration"].(int); ok {
		out.FilecacheDuration = int32(v)
	}
	if v, ok := m["media_store_storage_class"].(string); ok && v != "" {
		out.MediaStoreStorageClass = types.HlsMediaStoreStorageClass(v)
	}
	if v, ok := m["num_retries"].(int); ok {
		out.NumRetries = int32(v)
	}
	if v, ok := m["restart_delay"].(int); ok {
		out.RestartDelay = int32(v)
	}

	return &out
}

func expandHSLS3Settings(tfList []interface{}) *types.HlsS3Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsS3Settings
	if v, ok := m["canned_acl"].(string); ok && v != "" {
		out.CannedAcl = types.S3CannedAcl(v)
	}

	return &out
}

func expandHLSWebdavSettings(tfList []interface{}) *types.HlsWebdavSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsWebdavSettings
	if v, ok := m["connection_retry_interval"].(int); ok {
		out.ConnectionRetryInterval = int32(v)
	}
	if v, ok := m["filecache_duration"].(int); ok {
		out.FilecacheDuration = int32(v)
	}
	if v, ok := m["http_transfer_mode"].(string); ok && v != "" {
		out.HttpTransferMode = types.HlsWebdavHttpTransferMode(v)
	}
	if v, ok := m["num_retries"].(int); ok {
		out.NumRetries = int32(v)
	}
	if v, ok := m["restart_delay"].(int); ok {
		out.RestartDelay = int32(v)
	}
	return &out
}

func expandHSLGroupSettingsCaptionLanguageMappings(tfList []interface{}) []types.CaptionLanguageMapping {
	if tfList == nil {
		return nil
	}

	var out []types.CaptionLanguageMapping
	for _, item := range tfList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.CaptionLanguageMapping
		if v, ok := m["caption_channel"].(int); ok {
			o.CaptionChannel = int32(v)
		}
		if v, ok := m["language_code"].(string); ok && v != "" {
			o.LanguageCode = aws.String(v)
		}
		if v, ok := m["language_description"].(string); ok && v != "" {
			o.LanguageDescription = aws.String(v)
		}

		out = append(out, o)
	}

	return out
}

func expandHLSGroupSettingsKeyProviderSettings(tfList []interface{}) *types.KeyProviderSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.KeyProviderSettings
	if v, ok := m["static_key_settings"].([]interface{}); ok && len(v) > 0 {
		out.StaticKeySettings = expandKeyProviderSettingsStaticKeySettings(v)
	}

	return &out
}

func expandKeyProviderSettingsStaticKeySettings(tfList []interface{}) *types.StaticKeySettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.StaticKeySettings
	if v, ok := m["static_key_value"].(string); ok && v != "" {
		out.StaticKeyValue = aws.String(v)
	}
	if v, ok := m["key_provider_server"].([]interface{}); ok && len(v) > 0 {
		out.KeyProviderServer = expandInputLocation(v)
	}

	return &out
}

func expandInputLocation(tfList []interface{}) *types.InputLocation {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.InputLocation
	if v, ok := m["uri"].(string); ok && v != "" {
		out.Uri = aws.String(v)
	}
	if v, ok := m["password_param"].(string); ok && v != "" {
		out.PasswordParam = aws.String(v)
	}
	if v, ok := m["username"].(string); ok && v != "" {
		out.Username = aws.String(v)
	}

	return &out
}

func expandArchiveCDNSettings(tfList []interface{}) *types.ArchiveCdnSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.ArchiveCdnSettings
	if v, ok := m["archive_s3_settings"].([]interface{}); ok && len(v) > 0 {
		out.ArchiveS3Settings = func(in []interface{}) *types.ArchiveS3Settings {
			if len(in) == 0 {
				return nil
			}

			m := in[0].(map[string]interface{})

			var o types.ArchiveS3Settings
			if v, ok := m["canned_acl"].(string); ok && v != "" {
				o.CannedAcl = types.S3CannedAcl(v)
			}

			return &o
		}(v)
	}

	return &out
}

func expandAudioWatermarkSettings(tfList []interface{}) *types.AudioWatermarkSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var o types.AudioWatermarkSettings
	if v, ok := m["nielsen_watermark_settings"].([]interface{}); ok && len(v) > 0 {
		o.NielsenWatermarksSettings = func(n []interface{}) *types.NielsenWatermarksSettings {
			if len(n) == 0 {
				return nil
			}

			inner := n[0].(map[string]interface{})

			var ns types.NielsenWatermarksSettings
			if v, ok := inner["nielsen_distribution_type"].(string); ok && v != "" {
				ns.NielsenDistributionType = types.NielsenWatermarksDistributionTypes(v)
			}
			if v, ok := inner["nielsen_cbet_settings"].([]interface{}); ok && len(v) > 0 {
				ns.NielsenCbetSettings = expandNielsenCbetSettings(v)
			}
			if v, ok := inner["nielsen_naes_ii_nw_settings"].([]interface{}); ok && len(v) > 0 {
				ns.NielsenNaesIiNwSettings = expandNielsenNaseIiNwSettings(v)
			}

			return &ns
		}(v)
	}

	return &o
}

func expandRtmpGroupSettings(tfList []interface{}) *types.RtmpGroupSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.RtmpGroupSettings
	if v, ok := m["ad_markers"].([]interface{}); ok && len(v) > 0 {
		out.AdMarkers = expandRTMPAdMarkers(v)
	}
	if v, ok := m["authentication_scheme"].(string); ok && v != "" {
		out.AuthenticationScheme = types.AuthenticationScheme(v)
	}
	if v, ok := m["cache_full_behavior"].(string); ok && v != "" {
		out.CacheFullBehavior = types.RtmpCacheFullBehavior(v)
	}
	if v, ok := m["cache_length"].(int); ok {
		out.CacheLength = int32(v)
	}
	if v, ok := m["caption_data"].(string); ok && v != "" {
		out.CaptionData = types.RtmpCaptionData(v)
	}
	if v, ok := m["input_loss_action"].(string); ok && v != "" {
		out.InputLossAction = types.InputLossActionForRtmpOut(v)
	}
	if v, ok := m["restart_delay"].(int); ok {
		out.RestartDelay = int32(v)
	}

	return &out
}

func expandUdpGroupSettings(tfList []interface{}) *types.UdpGroupSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.UdpGroupSettings
	if v, ok := m["input_loss_action"].(string); ok && v != "" {
		out.InputLossAction = types.InputLossActionForUdpOut(v)
	}
	if v, ok := m["timed_metadata_id3_frame"].(string); ok && v != "" {
		out.TimedMetadataId3Frame = types.UdpTimedMetadataId3Frame(v)
	}
	if v, ok := m["timed_metadata_id3_period"].(int); ok {
		out.TimedMetadataId3Period = int32(v)
	}

	return &out
}

func expandRTMPAdMarkers(tfList []interface{}) []types.RtmpAdMarkers {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.RtmpAdMarkers
	for _, v := range tfList {
		out = append(out, types.RtmpAdMarkers(v.(string)))
	}

	return out
}

func expandHLSAdMarkers(tfList []interface{}) []types.HlsAdMarkers {
	if len(tfList) == 0 {
		return nil
	}

	var out []types.HlsAdMarkers
	for _, v := range tfList {
		out = append(out, types.HlsAdMarkers(v.(string)))
	}

	return out
}

func expandChannelEncoderSettingsOutputGroupsOutputs(tfList []interface{}) []types.Output {
	if tfList == nil {
		return nil
	}

	var outputs []types.Output
	for _, item := range tfList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var o types.Output
		if v, ok := m["output_settings"].([]interface{}); ok && len(v) > 0 {
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

func expandOutputsOutputSettings(tfList []interface{}) *types.OutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var os types.OutputSettings
	if v, ok := m["archive_output_settings"].([]interface{}); ok && len(v) > 0 {
		os.ArchiveOutputSettings = expandOutputsOutputSettingsArchiveOutputSettings(v)
	}
	if v, ok := m["frame_capture_output_settings"].([]interface{}); ok && len(v) > 0 {
		os.FrameCaptureOutputSettings = expandOutputsOutSettingsFrameCaptureOutputSettings(v)
	}
	if v, ok := m["hls_output_settings"].([]interface{}); ok && len(v) > 0 {
		os.HlsOutputSettings = expandOutputsOutSettingsHLSOutputSettings(v)
	}
	if v, ok := m["media_package_output_settings"].([]interface{}); ok && len(v) > 0 {
		os.MediaPackageOutputSettings = &types.MediaPackageOutputSettings{} // no exported fields
	}
	if v, ok := m["ms_smooth_output_settings"].([]interface{}); ok && len(v) > 0 {
		os.MsSmoothOutputSettings = expandOutputsOutSettingsMsSmoothOutputSettings(v)
	}
	if v, ok := m["multiplex_output_settings"].([]interface{}); ok && len(v) > 0 {
		os.MultiplexOutputSettings = func(inner []interface{}) *types.MultiplexOutputSettings {
			if len(inner) == 0 {
				return nil
			}

			data := inner[0].(map[string]interface{})
			var mos types.MultiplexOutputSettings
			if v, ok := data["destination"].([]interface{}); ok && len(v) > 0 {
				mos.Destination = expandDestination(v)
			}
			return &mos
		}(v)
	}

	if v, ok := m["rtmp_output_settings"].([]interface{}); ok && len(v) > 0 {
		os.RtmpOutputSettings = expandOutputsOutputSettingsRtmpOutputSettings(v)
	}
	if v, ok := m["udp_output_settings"].([]interface{}); ok && len(v) > 0 {
		os.UdpOutputSettings = expandOutputsOutputSettingsUdpOutputSettings(v)
	}

	return &os
}

func expandOutputsOutputSettingsArchiveOutputSettings(tfList []interface{}) *types.ArchiveOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var settings types.ArchiveOutputSettings
	if v, ok := m["container_settings"].([]interface{}); ok && len(v) > 0 {
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

func expandOutputsOutSettingsFrameCaptureOutputSettings(tfList []interface{}) *types.FrameCaptureOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.FrameCaptureOutputSettings
	if v, ok := m["name_modifier"].(string); ok && v != "" {
		out.NameModifier = aws.String(v)
	}

	return &out
}

func expandOutputsOutSettingsHLSOutputSettings(tfList []interface{}) *types.HlsOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsOutputSettings
	if v, ok := m["hls_settings"].([]interface{}); ok && len(v) > 0 {
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

func expandOutputsOutSettingsMsSmoothOutputSettings(tfList []interface{}) *types.MsSmoothOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.MsSmoothOutputSettings
	if v, ok := m["h265_packaging_type"].(string); ok && v != "" {
		out.H265PackagingType = types.MsSmoothH265PackagingType(v)
	}
	if v, ok := m["name_modifier"].(string); ok && v != "" {
		out.NameModifier = aws.String(v)
	}

	return &out
}

func expandHLSOutputSettingsHLSSettings(tfList []interface{}) *types.HlsSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.HlsSettings
	if v, ok := m["audio_only_hls_settings"].([]interface{}); ok && len(v) > 0 {
		out.AudioOnlyHlsSettings = expandHLSSettingsAudioOnlyHLSSettings(v)
	}
	if v, ok := m["fmp4_hls_settings"].([]interface{}); ok && len(v) > 0 {
		out.Fmp4HlsSettings = expandHLSSettingsFmp4HLSSettings(v)
	}
	if v, ok := m["frame_capture_hls_settings"].([]interface{}); ok && len(v) > 0 {
		out.FrameCaptureHlsSettings = &types.FrameCaptureHlsSettings{} // no exported types
	}
	if v, ok := m["standard_hls_settings"].([]interface{}); ok && len(v) > 0 {
		out.StandardHlsSettings = expandHLSSettingsStandardHLSSettings(v)
	}

	return &out
}

func expandHLSSettingsAudioOnlyHLSSettings(tfList []interface{}) *types.AudioOnlyHlsSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AudioOnlyHlsSettings
	if v, ok := m["audio_group_id"].(string); ok && v != "" {
		out.AudioGroupId = aws.String(v)
	}
	if v, ok := m["audio_only_image"].([]interface{}); ok && len(v) > 0 {
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

func expandHLSSettingsFmp4HLSSettings(tfList []interface{}) *types.Fmp4HlsSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

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

func expandHLSSettingsStandardHLSSettings(tfList []interface{}) *types.StandardHlsSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.StandardHlsSettings
	if v, ok := m["m3u8_settings"].([]interface{}); ok && len(v) > 0 {
		out.M3u8Settings = expandStandardHLSSettingsH3u8Settings(v)
	}
	if v, ok := m["audio_rendition_sets"].(string); ok && v != "" {
		out.AudioRenditionSets = aws.String(v)
	}

	return &out
}

func expandStandardHLSSettingsH3u8Settings(tfList []interface{}) *types.M3u8Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.M3u8Settings
	if v, ok := m["audio_frames_per_pes"].(int); ok {
		out.AudioFramesPerPes = int32(v)
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
	if v, ok := m["pat_interval"].(int); ok {
		out.PatInterval = int32(v)
	}
	if v, ok := m["pcr_control"].(string); ok && v != "" {
		out.PcrControl = types.M3u8PcrControl(v)
	}
	if v, ok := m["pcr_period"].(int); ok {
		out.PcrPeriod = int32(v)
	}
	if v, ok := m["pcr_pid"].(string); ok && v != "" {
		out.PcrPid = aws.String(v)
	}
	if v, ok := m["pmt_interval"].(int); ok {
		out.PmtInterval = int32(v)
	}
	if v, ok := m["pmt_pid"].(string); ok && v != "" {
		out.PmtPid = aws.String(v)
	}
	if v, ok := m["program_num"].(int); ok {
		out.ProgramNum = int32(v)
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
	if v, ok := m["transport_stream_id"].(int); ok {
		out.TransportStreamId = int32(v)
	}
	if v, ok := m["video_pid"].(string); ok && v != "" {
		out.VideoPid = aws.String(v)
	}

	return &out
}

func expandOutputsOutputSettingsRtmpOutputSettings(tfList []interface{}) *types.RtmpOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var settings types.RtmpOutputSettings
	if v, ok := m["destination"].([]interface{}); ok && len(v) > 0 {
		settings.Destination = expandDestination(v)
	}
	if v, ok := m["certificate_mode"].(string); ok && v != "" {
		settings.CertificateMode = types.RtmpOutputCertificateMode(v)
	}
	if v, ok := m["connection_retry_interval"].(int); ok {
		settings.ConnectionRetryInterval = int32(v)
	}
	if v, ok := m["num_retries"].(int); ok {
		settings.NumRetries = int32(v)
	}

	return &settings
}

func expandOutputsOutputSettingsUdpOutputSettings(tfList []interface{}) *types.UdpOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var settings types.UdpOutputSettings
	if v, ok := m["container_settings"].([]interface{}); ok && len(v) > 0 {
		settings.ContainerSettings = expandOutputsOutputSettingsUdpSettingsContainerSettings(v)
	}
	if v, ok := m["destination"].([]interface{}); ok && len(v) > 0 {
		settings.Destination = expandDestination(v)
	}
	if v, ok := m["buffer_msec"].(int); ok {
		settings.BufferMsec = int32(v)
	}
	if v, ok := m["fec_output_settings"].([]interface{}); ok && len(v) > 0 {
		settings.FecOutputSettings = expandFecOutputSettings(v)
	}

	return &settings
}

func expandOutputsOutputSettingsArchiveSettingsContainerSettings(tfList []interface{}) *types.ArchiveContainerSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var settings types.ArchiveContainerSettings
	if v, ok := m["m2ts_settings"].([]interface{}); ok && len(v) > 0 {
		settings.M2tsSettings = expandM2tsSettings(v)
	}

	if v, ok := m["raw_settings"].([]interface{}); ok && len(v) > 0 {
		settings.RawSettings = &types.RawSettings{}
	}
	return &settings
}

func expandOutputsOutputSettingsUdpSettingsContainerSettings(tfList []interface{}) *types.UdpContainerSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var settings types.UdpContainerSettings
	if v, ok := m["m2ts_settings"].([]interface{}); ok && len(v) > 0 {
		settings.M2tsSettings = expandM2tsSettings(v)
	}

	return &settings
}

func expandFecOutputSettings(tfList []interface{}) *types.FecOutputSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var settings types.FecOutputSettings
	if v, ok := m["column_depth"].(int); ok {
		settings.ColumnDepth = int32(v)
	}
	if v, ok := m["column_depth"].(string); ok && v != "" {
		settings.IncludeFec = types.FecOutputIncludeFec(v)
	}
	if v, ok := m["row_length"].(int); ok {
		settings.RowLength = int32(v)
	}

	return &settings
}

func expandM2tsSettings(tfList []interface{}) *types.M2tsSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var s types.M2tsSettings
	if v, ok := m["absent_input_audio_behavior"].(string); ok && v != "" {
		s.AbsentInputAudioBehavior = types.M2tsAbsentInputAudioBehavior(v)
	}
	if v, ok := m["arib"].(string); ok && v != "" {
		s.Arib = types.M2tsArib(v)
	}
	if v, ok := m["arib_caption_pid"].(string); ok && v != "" {
		s.AribCaptionsPid = aws.String(v)
	}
	if v, ok := m["arib_caption_pid_control"].(string); ok && v != "" {
		s.AribCaptionsPidControl = types.M2tsAribCaptionsPidControl(v)
	}
	if v, ok := m["audio_buffer_model"].(string); ok && v != "" {
		s.AudioBufferModel = types.M2tsAudioBufferModel(v)
	}
	if v, ok := m["audio_frames_per_pes"].(int); ok {
		s.AudioFramesPerPes = int32(v)
	}
	if v, ok := m["audi_pids"].(string); ok && v != "" {
		s.AudioPids = aws.String(v)
	}
	if v, ok := m["audio_stream_type"].(string); ok && v != "" {
		s.AudioStreamType = types.M2tsAudioStreamType(v)
	}
	if v, ok := m["bitrate"].(int); ok {
		s.Bitrate = int32(v)
	}
	if v, ok := m["buffer_model"].(string); ok && v != "" {
		s.BufferModel = types.M2tsBufferModel(v)
	}
	if v, ok := m["cc_descriptor"].(string); ok && v != "" {
		s.CcDescriptor = types.M2tsCcDescriptor(v)
	}
	if v, ok := m["dvb_nit_settings"].([]interface{}); ok && len(v) > 0 {
		s.DvbNitSettings = expandM2tsDvbNitSettings(v)
	}
	if v, ok := m["dvb_sdt_settings"].([]interface{}); ok && len(v) > 0 {
		s.DvbSdtSettings = expandM2tsDvbSdtSettings(v)
	}
	if v, ok := m["dvb_sub_pids"].(string); ok && v != "" {
		s.AudioPids = aws.String(v)
	}
	if v, ok := m["dvb_tdt_settings"].([]interface{}); ok && len(v) > 0 {
		s.DvbTdtSettings = func(tfList []interface{}) *types.DvbTdtSettings {
			if tfList == nil {
				return nil
			}

			m := tfList[0].(map[string]interface{})

			var s types.DvbTdtSettings
			if v, ok := m["rep_interval"].(int); ok {
				s.RepInterval = int32(v)
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
	if v, ok := m["ebp_lookahead_ms"].(int); ok {
		s.EbpLookaheadMs = int32(v)
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
	if v, ok := m["fragment_time"].(float32); ok {
		s.FragmentTime = float64(v)
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
	if v, ok := m["null_packet_bitrate"].(float32); ok {
		s.NullPacketBitrate = float64(v)
	}
	if v, ok := m["pat_interval"].(int); ok {
		s.PatInterval = int32(v)
	}
	if v, ok := m["pcr_control"].(string); ok && v != "" {
		s.PcrControl = types.M2tsPcrControl(v)
	}
	if v, ok := m["pcr_period"].(int); ok {
		s.PcrPeriod = int32(v)
	}
	if v, ok := m["pcr_pid"].(string); ok && v != "" {
		s.PcrPid = aws.String(v)
	}
	if v, ok := m["pmt_interval"].(int); ok {
		s.PmtInterval = int32(v)
	}
	if v, ok := m["pmt_pid"].(string); ok && v != "" {
		s.PmtPid = aws.String(v)
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
	if v, ok := m["segmentation_time"].(float32); ok {
		s.SegmentationTime = float64(v)
	}
	if v, ok := m["timed_metadata_behavior"].(string); ok && v != "" {
		s.TimedMetadataBehavior = types.M2tsTimedMetadataBehavior(v)
	}
	if v, ok := m["timed_metadata_pid"].(string); ok && v != "" {
		s.TimedMetadataPid = aws.String(v)
	}
	if v, ok := m["transport_stream_id"].(int); ok {
		s.TransportStreamId = int32(v)
	}
	if v, ok := m["video_pid"].(string); ok && v != "" {
		s.TimedMetadataPid = aws.String(v)
	}

	return &s
}

func expandM2tsDvbNitSettings(tfList []interface{}) *types.DvbNitSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var s types.DvbNitSettings
	if v, ok := m["network_ids"].(int); ok {
		s.NetworkId = int32(v)
	}
	if v, ok := m["network_name"].(string); ok && v != "" {
		s.NetworkName = aws.String(v)
	}
	if v, ok := m["network_ids"].(int); ok {
		s.RepInterval = int32(v)
	}
	return &s
}

func expandM2tsDvbSdtSettings(tfList []interface{}) *types.DvbSdtSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var s types.DvbSdtSettings
	if v, ok := m["output_sdt"].(string); ok && v != "" {
		s.OutputSdt = types.DvbSdtOutputSdt(v)
	}
	if v, ok := m["rep_interval"].(int); ok {
		s.RepInterval = int32(v)
	}
	if v, ok := m["service_name"].(string); ok && v != "" {
		s.ServiceName = aws.String(v)
	}
	if v, ok := m["service_provider_name"].(string); ok && v != "" {
		s.ServiceProviderName = aws.String(v)
	}

	return &s
}

func expandChannelEncoderSettingsTimecodeConfig(tfList []interface{}) *types.TimecodeConfig {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	var config types.TimecodeConfig
	if v, ok := m["source"].(string); ok && v != "" {
		config.Source = types.TimecodeConfigSource(v)
	}
	if v, ok := m["sync_threshold"].(int32); ok {
		config.SyncThreshold = v
	}

	return &config
}

func expandChannelEncoderSettingsVideoDescriptions(tfList []interface{}) []types.VideoDescription {
	if tfList == nil {
		return nil
	}

	var videoDesc []types.VideoDescription
	for _, tfItem := range tfList {
		m, ok := tfItem.(map[string]interface{})
		if !ok {
			continue
		}

		var d types.VideoDescription
		if v, ok := m["name"].(string); ok && v != "" {
			d.Name = aws.String(v)
		}
		if v, ok := m["codec_settings"].([]interface{}); ok && len(v) > 0 {
			d.CodecSettings = expandChannelEncoderSettingsVideoDescriptionsCodecSettings(v)
		}
		if v, ok := m["height"].(int); ok {
			d.Height = int32(v)
		}
		if v, ok := m["respond_to_afd"].(string); ok && v != "" {
			d.RespondToAfd = types.VideoDescriptionRespondToAfd(v)
		}
		if v, ok := m["scaling_behavior"].(string); ok && v != "" {
			d.ScalingBehavior = types.VideoDescriptionScalingBehavior(v)
		}
		if v, ok := m["sharpness"].(int); ok {
			d.Sharpness = int32(v)
		}
		if v, ok := m["width"].(int); ok {
			d.Width = int32(v)
		}

		videoDesc = append(videoDesc, d)
	}

	return videoDesc
}

func expandChannelEncoderSettingsAvailBlanking(tfList []interface{}) *types.AvailBlanking {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.AvailBlanking
	if v, ok := m["avail_blanking_image"].([]interface{}); ok && len(v) > 0 {
		out.AvailBlankingImage = expandInputLocation(v)
	}
	if v, ok := m["state"].(string); ok && v != "" {
		out.State = types.AvailBlankingState(v)
	}

	return &out
}

func expandChannelEncoderSettingsVideoDescriptionsCodecSettings(tfList []interface{}) *types.VideoCodecSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.VideoCodecSettings
	if v, ok := m["frame_capture_settings"].([]interface{}); ok && len(v) > 0 {
		out.FrameCaptureSettings = expandsVideoDescriptionsCodecSettingsFrameCaptureSettings(v)
	}
	if v, ok := m["h264_settings"].([]interface{}); ok && len(v) > 0 {
		out.H264Settings = expandsVideoDescriptionsCodecSettingsH264Settings(v)
	}

	return &out
}

func expandsVideoDescriptionsCodecSettingsFrameCaptureSettings(tfList []interface{}) *types.FrameCaptureSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.FrameCaptureSettings
	if v, ok := m["capture_interval"].(int); ok {
		out.CaptureInterval = int32(v)
	}
	if v, ok := m["capture_interval_units"].(string); ok && v != "" {
		out.CaptureIntervalUnits = types.FrameCaptureIntervalUnit(v)
	}

	return &out
}

func expandsVideoDescriptionsCodecSettingsH264Settings(tfList []interface{}) *types.H264Settings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.H264Settings
	if v, ok := m["adaptive_quantization"].(string); ok && v != "" {
		out.AdaptiveQuantization = types.H264AdaptiveQuantization(v)
	}
	if v, ok := m["afd_signaling"].(string); ok && v != "" {
		out.AfdSignaling = types.AfdSignaling(v)
	}
	if v, ok := m["bitrate"].(int); ok {
		out.Bitrate = int32(v)
	}
	if v, ok := m["buf_fill_pct"].(int); ok {
		out.BufFillPct = int32(v)
	}
	if v, ok := m["buf_size"].(int); ok {
		out.BufSize = int32(v)
	}
	if v, ok := m["color_metadata"].(string); ok && v != "" {
		out.ColorMetadata = types.H264ColorMetadata(v)
	}
	if v, ok := m["entropy_encoding"].(string); ok && v != "" {
		out.EntropyEncoding = types.H264EntropyEncoding(v)
	}
	if v, ok := m["filter_settings"].([]interface{}); ok && len(v) > 0 {
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
	if v, ok := m["framerate_denominator"].(int); ok {
		out.FramerateDenominator = int32(v)
	}
	if v, ok := m["framerate_numerator"].(int); ok {
		out.FramerateNumerator = int32(v)
	}
	if v, ok := m["gop_b_reference"].(string); ok && v != "" {
		out.GopBReference = types.H264GopBReference(v)
	}
	if v, ok := m["gop_closed_cadence"].(int); ok {
		out.GopClosedCadence = int32(v)
	}
	if v, ok := m["gop_num_b_frames"].(int); ok {
		out.GopNumBFrames = int32(v)
	}
	if v, ok := m["gop_size"].(float32); ok {
		out.GopSize = float64(v)
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
	if v, ok := m["max_bitrate"].(int); ok {
		out.MaxBitrate = int32(v)
	}
	if v, ok := m["min_i_interval"].(int); ok {
		out.MinIInterval = int32(v)
	}
	if v, ok := m["num_ref_frames"].(int); ok {
		out.NumRefFrames = int32(v)
	}
	if v, ok := m["par_control"].(string); ok && v != "" {
		out.ParControl = types.H264ParControl(v)
	}
	if v, ok := m["par_denominator"].(int); ok {
		out.ParDenominator = int32(v)
	}
	if v, ok := m["par_numerator"].(int); ok {
		out.ParNumerator = int32(v)
	}
	if v, ok := m["profile"].(string); ok && v != "" {
		out.Profile = types.H264Profile(v)
	}
	if v, ok := m["quality_level"].(string); ok && v != "" {
		out.QualityLevel = types.H264QualityLevel(v)
	}
	if v, ok := m["qvbr_quality_level"].(int); ok {
		out.QvbrQualityLevel = int32(v)
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
	if v, ok := m["slices"].(int); ok {
		out.Slices = int32(v)
	}
	if v, ok := m["softness"].(int); ok {
		out.Softness = int32(v)
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

func expandH264SettingsFilterSettings(tfList []interface{}) *types.H264FilterSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.H264FilterSettings
	if v, ok := m["temporal_filter_settings"].([]interface{}); ok && len(v) > 0 {
		out.TemporalFilterSettings = expandH264FilterSettingsTemporalFilterSettings(v)
	}

	return &out
}

func expandH264FilterSettingsTemporalFilterSettings(tfList []interface{}) *types.TemporalFilterSettings {
	if tfList == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.TemporalFilterSettings
	if v, ok := m["post_filter_sharpening"].(string); ok && v != "" {
		out.PostFilterSharpening = types.TemporalFilterPostFilterSharpening(v)
	}
	if v, ok := m["strength"].(string); ok && v != "" {
		out.Strength = types.TemporalFilterStrength(v)
	}

	return &out
}

func expandNielsenCbetSettings(tfList []interface{}) *types.NielsenCBET {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

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

func expandNielsenNaseIiNwSettings(tfList []interface{}) *types.NielsenNaesIiNw {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	var out types.NielsenNaesIiNw
	if v, ok := m["check_digit_string"].(string); ok && v != "" {
		out.CheckDigitString = aws.String(v)
	}
	if v, ok := m["sid"].(float32); ok {
		out.Sid = float64(v)
	}

	return &out
}

func flattenChannelEncoderSettings(apiObject *types.EncoderSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_descriptions": flattenAudioDescriptions(apiObject.AudioDescriptions),
		"output_groups":      flattenOutputGroups(apiObject.OutputGroups),
		"timecode_config":    flattenTimecodeConfig(apiObject.TimecodeConfig),
		"video_descriptions": flattenVideoDescriptions(apiObject.VideoDescriptions),
		"avail_blanking":     flattenAvailBlanking(apiObject.AvailBlanking),
		// TODO avail_configuration
		// TODO blackout_slate
		// TODO caption_descriptions
		// TODO feature_activations
		// TODO global_configuration
		// TODO motion_graphics_configuration
		// TODO nielsen_configuration
	}

	return []interface{}{m}
}

func flattenAudioDescriptions(od []types.AudioDescription) []interface{} {
	if len(od) == 0 {
		return nil
	}

	var ml []interface{}

	for _, v := range od {
		m := map[string]interface{}{
			"audio_selector_name":          aws.ToString(v.AudioSelectorName),
			"name":                         aws.ToString(v.Name),
			"audio_normalization_settings": flattenAudioNormalization(v.AudioNormalizationSettings),
			"audio_type":                   v.AudioType,
			"audio_type_control":           v.AudioTypeControl,
			"audio_watermark_settings":     flattenAudioWatermarkSettings(v.AudioWatermarkingSettings),
			"codec_settings":               flattenAudioDescriptionsCodecSettings(v.CodecSettings),
			"language_code":                aws.ToString(v.LanguageCode),
			"language_code_control":        string(v.LanguageCodeControl),
			"remix_settings":               flattenAudioDescriptionsRemixSettings(v.RemixSettings),
			"stream_name":                  aws.ToString(v.StreamName),
		}

		ml = append(ml, m)
	}

	return ml
}

func flattenOutputGroups(op []types.OutputGroup) []interface{} {
	if len(op) == 0 {
		return nil
	}

	var ol []interface{}

	for _, v := range op {
		m := map[string]interface{}{
			"output_group_settings": flattenOutputGroupSettings(v.OutputGroupSettings),
			"outputs":               flattenOutputs(v.Outputs),
			"name":                  aws.ToString(v.Name),
		}

		ol = append(ol, m)
	}

	return ol
}

func flattenOutputGroupSettings(os *types.OutputGroupSettings) []interface{} {
	if os == nil {
		return nil
	}

	m := map[string]interface{}{
		"archive_group_settings":       flattenOutputGroupSettingsArchiveGroupSettings(os.ArchiveGroupSettings),
		"frame_capture_group_settings": flattenOutputGroupSettingsFrameCaptureGroupSettings(os.FrameCaptureGroupSettings),
		"hls_group_settings":           flattenOutputGroupSettingsHLSGroupSettings(os.HlsGroupSettings),
		"ms_smooth_group_settings":     flattenOutputGroupSettingsMsSmoothGroupSettings(os.MsSmoothGroupSettings),
		"media_package_group_settings": flattenOutputGroupSettingsMediaPackageGroupSettings(os.MediaPackageGroupSettings),
		"multiplex_group_settings": func(inner *types.MultiplexGroupSettings) []interface{} {
			if inner == nil {
				return nil
			}
			return []interface{}{} // no exported attributes
		}(os.MultiplexGroupSettings),
		"rtmp_group_settings": flattenOutputGroupSettingsRtmpGroupSettings(os.RtmpGroupSettings),
		"udp_group_settings":  flattenOutputGroupSettingsUdpGroupSettings(os.UdpGroupSettings),
	}

	return []interface{}{m}
}

func flattenOutputs(os []types.Output) []interface{} {
	if len(os) == 0 {
		return nil
	}

	var outputs []interface{}

	for _, item := range os {
		m := map[string]interface{}{
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

func flattenOutputsOutputSettings(in *types.OutputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"archive_output_settings":       flattenOutputsOutputSettingsArchiveOutputSettings(in.ArchiveOutputSettings),
		"frame_capture_output_settings": flattenOutputsOutputSettingsFrameCaptureOutputSettings(in.FrameCaptureOutputSettings),
		"hls_output_settings":           flattenOutputsOutputSettingsHLSOutputSettings(in.HlsOutputSettings),
		"media_package_output_settings": func(inner *types.MediaPackageOutputSettings) []interface{} {
			if inner == nil {
				return nil
			}
			return []interface{}{} // no exported attributes
		}(in.MediaPackageOutputSettings),
		"ms_smooth_output_settings": flattenOutputsOutputSettingsMsSmoothOutputSettings(in.MsSmoothOutputSettings),
		"multiplex_output_settings": func(inner *types.MultiplexOutputSettings) []interface{} {
			if inner == nil {
				return nil
			}
			data := map[string]interface{}{
				"destination": flattenDestination(inner.Destination),
			}

			return []interface{}{data}
		}(in.MultiplexOutputSettings),
		"rtmp_output_settings": flattenOutputsOutputSettingsRtmpOutputSettings(in.RtmpOutputSettings),
		"udp_output_settings":  flattenOutputsOutputSettingsUdpOutputSettings(in.UdpOutputSettings),
	}

	return []interface{}{m}
}

func flattenOutputsOutputSettingsArchiveOutputSettings(in *types.ArchiveOutputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"container_settings": flattenOutputsOutputSettingsArchiveOutputSettingsContainerSettings(in.ContainerSettings),
		"extension":          aws.ToString(in.Extension),
		"name_modifier":      aws.ToString(in.NameModifier),
	}

	return []interface{}{m}
}

func flattenOutputsOutputSettingsFrameCaptureOutputSettings(in *types.FrameCaptureOutputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"name_modifier": aws.ToString(in.NameModifier),
	}

	return []interface{}{m}
}

func flattenOutputsOutputSettingsHLSOutputSettings(in *types.HlsOutputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"hls_settings":        flattenHLSOutputSettingsHLSSettings(in.HlsSettings),
		"h265_packaging_type": string(in.H265PackagingType),
		"name_modifier":       aws.ToString(in.NameModifier),
		"segment_modifier":    aws.ToString(in.SegmentModifier),
	}

	return []interface{}{m}
}

func flattenOutputsOutputSettingsMsSmoothOutputSettings(in *types.MsSmoothOutputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"h265_packaging_type": string(in.H265PackagingType),
		"name_modifier":       aws.ToString(in.NameModifier),
	}

	return []interface{}{m}
}

func flattenHLSOutputSettingsHLSSettings(in *types.HlsSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_only_hls_settings": flattenHLSSettingsAudioOnlyHLSSettings(in.AudioOnlyHlsSettings),
		"fmp4_hls_settings":       flattenHLSSettingsFmp4HLSSettings(in.Fmp4HlsSettings),
		"frame_capture_hls_settings": func(inner *types.FrameCaptureHlsSettings) []interface{} {
			if inner == nil {
				return nil
			}
			return []interface{}{} // no exported fields
		}(in.FrameCaptureHlsSettings),
		"standard_hls_settings": flattenHLSSettingsStandardHLSSettings(in.StandardHlsSettings),
	}

	return []interface{}{m}
}

func flattenHLSSettingsAudioOnlyHLSSettings(in *types.AudioOnlyHlsSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_group_id":   aws.ToString(in.AudioGroupId),
		"audio_only_image": flattenInputLocation(in.AudioOnlyImage),
		"audio_track_type": string(in.AudioTrackType),
		"segment_type":     string(in.AudioTrackType),
	}

	return []interface{}{m}
}

func flattenHLSSettingsFmp4HLSSettings(in *types.Fmp4HlsSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_rendition_sets":    aws.ToString(in.AudioRenditionSets),
		"nielsen_id3_behavior":    string(in.NielsenId3Behavior),
		"timed_metadata_behavior": string(in.TimedMetadataBehavior),
	}

	return []interface{}{m}
}

func flattenHLSSettingsStandardHLSSettings(in *types.StandardHlsSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"m3u8_settings":        flattenStandardHLSSettingsM3u8Settings(in.M3u8Settings),
		"audio_rendition_sets": aws.ToString(in.AudioRenditionSets),
	}

	return []interface{}{m}
}

func flattenStandardHLSSettingsM3u8Settings(in *types.M3u8Settings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_frames_per_pes":    int(in.AudioFramesPerPes),
		"audio_pids":              aws.ToString(in.AudioPids),
		"ecm_pid":                 aws.ToString(in.EcmPid),
		"nielsen_id3_behavior":    string(in.NielsenId3Behavior),
		"pat_interval":            int(in.PatInterval),
		"pcr_control":             string(in.PcrControl),
		"pcr_period":              int(in.PcrPeriod),
		"pcr_pid":                 aws.ToString(in.PcrPid),
		"pmt_interval":            int(in.PmtInterval),
		"pmt_pid":                 aws.ToString(in.PmtPid),
		"program_num":             int(in.ProgramNum),
		"scte35_behavior":         string(in.Scte35Behavior),
		"scte35_pid":              aws.ToString(in.Scte35Pid),
		"timed_metadata_behavior": string(in.TimedMetadataBehavior),
		"timed_metadata_pid":      aws.ToString(in.TimedMetadataPid),
		"transport_stream_id":     int(in.TransportStreamId),
		"video_pid":               aws.ToString(in.VideoPid),
	}

	return []interface{}{m}
}

func flattenOutputsOutputSettingsRtmpOutputSettings(in *types.RtmpOutputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"destination":               flattenDestination(in.Destination),
		"certificate_mode":          string(in.CertificateMode),
		"connection_retry_interval": int(in.ConnectionRetryInterval),
		"num_retries":               int(in.NumRetries),
	}

	return []interface{}{m}
}

func flattenOutputsOutputSettingsUdpOutputSettings(in *types.UdpOutputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"container_settings":  flattenOutputsOutputSettingsUdpOutputSettingsContainerSettings(in.ContainerSettings),
		"destination":         flattenDestination(in.Destination),
		"buffer_msec":         int(in.BufferMsec),
		"fec_output_settings": flattenFecOutputSettings(in.FecOutputSettings),
	}

	return []interface{}{m}
}

func flattenOutputsOutputSettingsArchiveOutputSettingsContainerSettings(in *types.ArchiveContainerSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"m2ts_settings": flattenM2tsSettings(in.M2tsSettings),
		"raw_settings":  []interface{}{}, // attribute has no exported fields
	}

	return []interface{}{m}
}

func flattenOutputsOutputSettingsUdpOutputSettingsContainerSettings(in *types.UdpContainerSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"m2ts_settings": flattenM2tsSettings(in.M2tsSettings),
	}

	return []interface{}{m}
}

func flattenFecOutputSettings(in *types.FecOutputSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"column_depth": int(in.ColumnDepth),
		"include_fec":  string(in.IncludeFec),
		"row_length":   int(in.RowLength),
	}

	return []interface{}{m}
}

func flattenM2tsSettings(in *types.M2tsSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"absent_input_audio_behavior": string(in.AbsentInputAudioBehavior),
		"arib":                        string(in.Arib),
		"arib_captions_pid":           aws.ToString(in.AribCaptionsPid),
		"arib_captions_pid_control":   string(in.AribCaptionsPidControl),
		"audio_buffer_model":          string(in.AudioBufferModel),
		"audio_frames_per_pes":        int(in.AudioFramesPerPes),
		"audio_pids":                  aws.ToString(in.AudioPids),
		"audio_stream_type":           string(in.AudioStreamType),
		"bitrate":                     int(in.Bitrate),
		"buffer_model":                string(in.BufferModel),
		"cc_descriptor":               string(in.CcDescriptor),
		"dvb_nit_settings":            flattenDvbNitSettings(in.DvbNitSettings),
		"dvb_sdt_settings":            flattenDvbSdtSettings(in.DvbSdtSettings),
		"dvb_sub_pids":                aws.ToString(in.DvbSubPids),
		"dvb_tdt_settings":            flattenDvbTdtSettings(in.DvbTdtSettings),
		"dvb_teletext_pid":            aws.ToString(in.DvbTeletextPid),
		"ebif":                        string(in.Ebif),
		"ebp_audio_interval":          string(in.EbpAudioInterval),
		"ebp_lookahead_ms":            int(in.EbpLookaheadMs),
		"ebp_placement":               string(in.EbpPlacement),
		"ecm_pid":                     aws.ToString(in.EcmPid),
		"es_rate_in_pes":              string(in.EsRateInPes),
		"etv_platform_pid":            aws.ToString(in.EtvPlatformPid),
		"etv_signal_pid":              aws.ToString(in.EtvSignalPid),
		"fragment_time":               float32(in.FragmentTime),
		"klv":                         string(in.Klv),
		"klv_data_pids":               aws.ToString(in.KlvDataPids),
		"nielsen_id3_behavior":        string(in.NielsenId3Behavior),
		"null_packet_bitrate":         float32(in.NullPacketBitrate),
		"pat_interval":                int(in.PatInterval),
		"pcr_control":                 string(in.PcrControl),
		"pcr_period":                  int(in.PcrPeriod),
		"pcr_pid":                     aws.ToString(in.PcrPid),
		"pmt_interval":                int(in.PmtInterval),
		"pmt_pid":                     aws.ToString(in.PmtPid),
		"program_num":                 int(in.ProgramNum),
		"rate_mode":                   string(in.RateMode),
		"scte27_pids":                 aws.ToString(in.Scte27Pids),
		"scte35_control":              string(in.Scte35Control),
		"scte35_pid":                  aws.ToString(in.Scte35Pid),
		"segmentation_markers":        string(in.SegmentationMarkers),
		"segmentation_style":          string(in.SegmentationStyle),
		"segmentation_time":           float32(in.SegmentationTime),
		"timed_metadata_behavior":     string(in.TimedMetadataBehavior),
		"timed_metadata_pid":          aws.ToString(in.TimedMetadataPid),
		"transport_stream_id":         int(in.TransportStreamId),
		"video_pid":                   aws.ToString(in.VideoPid),
	}

	return []interface{}{m}
}

func flattenDvbNitSettings(in *types.DvbNitSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"network_id":   int(in.NetworkId),
		"network_name": aws.ToString(in.NetworkName),
		"rep_interval": int(in.RepInterval),
	}

	return []interface{}{m}
}

func flattenDvbSdtSettings(in *types.DvbSdtSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"output_sdt":            string(in.OutputSdt),
		"rep_interval":          int(in.RepInterval),
		"service_name":          aws.ToString(in.ServiceName),
		"service_provider_name": aws.ToString(in.ServiceProviderName),
	}

	return []interface{}{m}
}

func flattenDvbTdtSettings(in *types.DvbTdtSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"rep_interval": int(in.RepInterval),
	}

	return []interface{}{m}
}

func flattenOutputGroupSettingsArchiveGroupSettings(as *types.ArchiveGroupSettings) []interface{} {
	if as == nil {
		return nil
	}

	m := map[string]interface{}{
		"destination":          flattenDestination(as.Destination),
		"archive_cdn_settings": flattenOutputGroupSettingsArchiveCDNSettings(as.ArchiveCdnSettings),
		"rollover_interval":    int(as.RolloverInterval),
	}

	return []interface{}{m}
}

func flattenOutputGroupSettingsFrameCaptureGroupSettings(in *types.FrameCaptureGroupSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"destination":                flattenDestination(in.Destination),
		"frame_capture_cdn_settings": flattenFrameCaptureCDNSettings(in.FrameCaptureCdnSettings),
	}

	return []interface{}{m}
}

func flattenOutputGroupSettingsHLSGroupSettings(in *types.HlsGroupSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"destination":                  flattenDestination(in.Destination),
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
		"index_n_segments":             int(in.IndexNSegments),
		"input_loss_action":            string(in.InputLossAction),
		"iv_in_manifest":               string(in.IvInManifest),
		"iv_source":                    string(in.IvSource),
		"keep_segments":                int(in.KeepSegments),
		"key_format":                   aws.ToString(in.KeyFormat),
		"key_format_versions":          aws.ToString(in.KeyFormatVersions),
		"key_provider_settings":        flattenHLSKeyProviderSettings(in.KeyProviderSettings),
		"manifest_compression":         string(in.ManifestCompression),
		"manifest_duration_format":     string(in.ManifestDurationFormat),
		"min_segment_length":           int(in.MinSegmentLength),
		"mode":                         string(in.Mode),
		"output_selection":             string(in.OutputSelection),
		"program_date_time":            string(in.ProgramDateTime),
		"program_date_time_clock":      string(in.ProgramDateTimeClock),
		"program_date_time_period":     int(in.ProgramDateTimePeriod),
		"redundant_manifest":           string(in.RedundantManifest),
		"segment_length":               int(in.SegmentLength),
		"segments_per_subdirectory":    int(in.SegmentsPerSubdirectory),
		"stream_inf_resolution":        string(in.StreamInfResolution),
		"timed_metadata_id3_frame":     string(in.TimedMetadataId3Frame),
		"timed_metadata_id3_period":    int(in.TimedMetadataId3Period),
		"timestamp_delta_milliseconds": int(in.TimestampDeltaMilliseconds),
		"ts_file_mode":                 string(in.TsFileMode),
	}

	return []interface{}{m}
}

func flattenOutputGroupSettingsMsSmoothGroupSettings(in *types.MsSmoothGroupSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"destination":                 flattenDestination(in.Destination),
		"acquisition_point_id":        aws.ToString(in.AcquisitionPointId),
		"audio_only_timecode_control": string(in.AudioOnlyTimecodeControl),
		"certificate_mode":            string(in.CertificateMode),
		"connection_retry_interval":   int(in.ConnectionRetryInterval),
		"event_id":                    aws.ToString(in.EventId),
		"event_id_mode":               string(in.EventIdMode),
		"event_stop_behavior":         string(in.EventStopBehavior),
		"filecache_duration":          int(in.FilecacheDuration),
		"fragment_length":             int(in.FragmentLength),
		"input_loss_action":           string(in.InputLossAction),
		"num_retries":                 int(in.NumRetries),
		"restart_delay":               int(in.RestartDelay),
		"segmentation_mode":           string(in.SegmentationMode),
		"send_delay_ms":               int(in.SendDelayMs),
		"sparse_track_type":           string(in.SparseTrackType),
		"stream_manifest_behavior":    string(in.StreamManifestBehavior),
		"timestamp_offset":            aws.ToString(in.TimestampOffset),
		"timestamp_offset_mode":       string(in.TimestampOffsetMode),
	}

	return []interface{}{m}
}

func flattenHLSAdMarkers(in []types.HlsAdMarkers) []interface{} {
	if len(in) == 0 {
		return nil
	}

	var out []interface{}
	for _, item := range in {
		out = append(out, string(item))
	}

	return out
}

func flattenHLSCaptionLanguageMappings(in []types.CaptionLanguageMapping) []interface{} {
	if len(in) == 0 {
		return nil
	}

	var out []interface{}
	for _, item := range in {
		m := map[string]interface{}{
			"caption_channel":      int(item.CaptionChannel),
			"language_code":        aws.ToString(item.LanguageCode),
			"language_description": aws.ToString(item.LanguageDescription),
		}

		out = append(out, m)
	}

	return out
}

func flattenHLSCDNSettings(in *types.HlsCdnSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"hls_akamai_settings":      flattenHLSAkamaiSettings(in.HlsAkamaiSettings),
		"hls_basic_put_settings":   flattenHLSBasicPutSettings(in.HlsBasicPutSettings),
		"hls_media_store_settings": flattenHLSMediaStoreSettings(in.HlsMediaStoreSettings),
		"hls_s3_settings":          flattenHLSS3Settings(in.HlsS3Settings),
		"hls_webdav_settings":      flattenHLSWebdavSettings(in.HlsWebdavSettings),
	}

	return []interface{}{m}
}

func flattenHLSAkamaiSettings(in *types.HlsAkamaiSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"connection_retry_interval": int(in.ConnectionRetryInterval),
		"filecache_duration":        int(in.FilecacheDuration),
		"http_transfer_mode":        string(in.HttpTransferMode),
		"num_retries":               int(in.NumRetries),
		"restart_delay":             int(in.RestartDelay),
		"salt":                      aws.ToString(in.Salt),
		"token":                     aws.ToString(in.Token),
	}

	return []interface{}{m}
}

func flattenHLSBasicPutSettings(in *types.HlsBasicPutSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"connection_retry_interval": int(in.ConnectionRetryInterval),
		"filecache_duration":        int(in.FilecacheDuration),
		"num_retries":               int(in.NumRetries),
		"restart_delay":             int(in.RestartDelay),
	}

	return []interface{}{m}
}

func flattenHLSMediaStoreSettings(in *types.HlsMediaStoreSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"connection_retry_interval": int(in.ConnectionRetryInterval),
		"filecache_duration":        int(in.FilecacheDuration),
		"media_store_storage_class": string(in.MediaStoreStorageClass),
		"num_retries":               int(in.NumRetries),
		"restart_delay":             int(in.RestartDelay),
	}

	return []interface{}{m}
}

func flattenHLSS3Settings(in *types.HlsS3Settings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"canned_acl": string(in.CannedAcl),
	}

	return []interface{}{m}
}

func flattenFrameCaptureCDNSettings(in *types.FrameCaptureCdnSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"frame_capture_s3_settings": flattenFrameCaptureS3Settings(in.FrameCaptureS3Settings),
	}

	return []interface{}{m}
}

func flattenHLSWebdavSettings(in *types.HlsWebdavSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"connection_retry_interval": int(in.ConnectionRetryInterval),
		"filecache_duration":        int(in.FilecacheDuration),
		"http_transfer_mode":        string(in.HttpTransferMode),
		"num_retries":               int(in.NumRetries),
		"restart_delay":             int(in.RestartDelay),
	}

	return []interface{}{m}
}

func flattenHLSKeyProviderSettings(in *types.KeyProviderSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"static_key_settings": flattenKeyProviderSettingsStaticKeySettings(in.StaticKeySettings),
	}

	return []interface{}{m}
}

func flattenKeyProviderSettingsStaticKeySettings(in *types.StaticKeySettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"static_key_value":    aws.ToString(in.StaticKeyValue),
		"key_provider_server": flattenInputLocation(in.KeyProviderServer),
	}

	return []interface{}{m}
}

func flattenInputLocation(in *types.InputLocation) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"uri":            aws.ToString(in.Uri),
		"password_param": aws.ToString(in.PasswordParam),
		"username":       aws.ToString(in.Username),
	}

	return []interface{}{m}
}

func flattenFrameCaptureS3Settings(in *types.FrameCaptureS3Settings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"canned_acl": string(in.CannedAcl),
	}

	return []interface{}{m}
}

func flattenOutputGroupSettingsMediaPackageGroupSettings(mp *types.MediaPackageGroupSettings) []interface{} {
	if mp == nil {
		return nil
	}

	m := map[string]interface{}{
		"destination": flattenDestination(mp.Destination),
	}

	return []interface{}{m}
}

func flattenOutputGroupSettingsRtmpGroupSettings(rt *types.RtmpGroupSettings) []interface{} {
	if rt == nil {
		return nil
	}

	m := map[string]interface{}{
		"ad_markers":            flattenAdMakers(rt.AdMarkers),
		"authentication_scheme": string(rt.AuthenticationScheme),
		"cache_full_behavior":   string(rt.CacheFullBehavior),
		"cache_length":          int(rt.CacheLength),
		"caption_data":          string(rt.CaptionData),
		"input_loss_action":     string(rt.InputLossAction),
		"restart_delay":         int(rt.RestartDelay),
	}

	return []interface{}{m}
}

func flattenOutputGroupSettingsUdpGroupSettings(in *types.UdpGroupSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"input_loss_action":         string(in.InputLossAction),
		"timed_metadata_id3_frame":  string(in.TimedMetadataId3Frame),
		"timed_metadata_id3_period": int(in.TimedMetadataId3Period),
	}

	return []interface{}{m}
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

func flattenDestination(des *types.OutputLocationRef) []interface{} {
	if des == nil {
		return nil
	}

	m := map[string]interface{}{
		"destination_ref_id": aws.ToString(des.DestinationRefId),
	}

	return []interface{}{m}
}

func flattenOutputGroupSettingsArchiveCDNSettings(as *types.ArchiveCdnSettings) []interface{} {
	if as == nil {
		return nil
	}

	m := map[string]interface{}{
		"archive_s3_settings": func(in *types.ArchiveS3Settings) []interface{} {
			if in == nil {
				return nil
			}

			inner := map[string]interface{}{
				"canned_acl": string(in.CannedAcl),
			}

			return []interface{}{inner}
		}(as.ArchiveS3Settings),
	}

	return []interface{}{m}
}

func flattenTimecodeConfig(in *types.TimecodeConfig) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"source":         string(in.Source),
		"sync_threshold": int(in.SyncThreshold),
	}

	return []interface{}{m}
}

func flattenVideoDescriptions(tfList []types.VideoDescription) []interface{} {
	if len(tfList) == 0 {
		return nil
	}

	var out []interface{}

	for _, item := range tfList {
		m := map[string]interface{}{
			"name":             aws.ToString(item.Name),
			"codec_settings":   flattenVideoDescriptionsCodecSettings(item.CodecSettings),
			"height":           int(item.Height),
			"respond_to_afd":   string(item.RespondToAfd),
			"scaling_behavior": string(item.ScalingBehavior),
			"sharpness":        int(item.Sharpness),
			"width":            int(item.Width),
		}

		out = append(out, m)
	}
	return out
}

func flattenAvailBlanking(in *types.AvailBlanking) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"avail_blanking_image": flattenInputLocation(in.AvailBlankingImage),
		"state":                string(in.State),
	}

	return []interface{}{m}
}

func flattenVideoDescriptionsCodecSettings(in *types.VideoCodecSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"frame_capture_settings": flattenCodecSettingsFrameCaptureSettings(in.FrameCaptureSettings),
		"h264_settings":          flattenCodecSettingsH264Settings(in.H264Settings),
	}

	return []interface{}{m}
}

func flattenCodecSettingsFrameCaptureSettings(in *types.FrameCaptureSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"capture_interval":       int(in.CaptureInterval),
		"capture_interval_units": string(in.CaptureIntervalUnits),
	}

	return []interface{}{m}
}

func flattenCodecSettingsH264Settings(in *types.H264Settings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"adaptive_quantization":   string(in.AdaptiveQuantization),
		"afd_signaling":           string(in.AfdSignaling),
		"bitrate":                 int(in.Bitrate),
		"buf_size":                int(in.BufSize),
		"color_metadata":          string(in.ColorMetadata),
		"entropy_encoding":        string(in.EntropyEncoding),
		"filter_settings":         flattenH264SettingsFilterSettings(in.FilterSettings),
		"fixed_afd":               string(in.FixedAfd),
		"flicker_aq":              string(in.FlickerAq),
		"force_field_pictures":    string(in.ForceFieldPictures),
		"framerate_control":       string(in.FramerateControl),
		"framerate_denominator":   int(in.FramerateDenominator),
		"framerate_numerator":     int(in.FramerateNumerator),
		"gop_b_reference":         string(in.GopBReference),
		"gop_closed_cadence":      int(in.GopClosedCadence),
		"gop_num_b_frames":        int(in.GopNumBFrames),
		"gop_size":                float32(in.GopSize),
		"gop_size_units":          string(in.GopSizeUnits),
		"level":                   string(in.Level),
		"look_ahead_rate_control": string(in.LookAheadRateControl),
		"max_bitrate":             int(in.MaxBitrate),
		"min_i_interval":          int(in.MinIInterval),
		"num_ref_frames":          int(in.NumRefFrames),
		"par_control":             string(in.ParControl),
		"par_denominator":         int(in.ParDenominator),
		"par_numerator":           int(in.ParNumerator),
		"profile":                 string(in.Profile),
		"quality_level":           string(in.QualityLevel),
		"qvbr_quality_level":      int(in.QvbrQualityLevel),
		"rate_control_mode":       string(in.RateControlMode),
		"scan_type":               string(in.ScanType),
		"scene_change_detect":     string(in.SceneChangeDetect),
		"slices":                  int(in.Slices),
		"spatial_aq":              string(in.SpatialAq),
		"subgop_length":           string(in.SubgopLength),
		"syntax":                  string(in.Syntax),
		"temporal_aq":             string(in.TemporalAq),
		"timecode_insertion":      string(in.TimecodeInsertion),
	}

	return []interface{}{m}
}

func flattenH264SettingsFilterSettings(in *types.H264FilterSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"temporal_filter_settings": flattenFilterSettingsTemporalFilterSettings(in.TemporalFilterSettings),
	}

	return []interface{}{m}
}

func flattenFilterSettingsTemporalFilterSettings(in *types.TemporalFilterSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"post_filter_sharpening": string(in.PostFilterSharpening),
		"strength":               string(in.Strength),
	}

	return []interface{}{m}
}

func flattenAudioNormalization(ns *types.AudioNormalizationSettings) []interface{} {
	if ns == nil {
		return nil
	}

	m := map[string]interface{}{
		"algorithm":         ns.Algorithm,
		"algorithm_control": ns.AlgorithmControl,
		"target_lkfs":       ns.TargetLkfs,
	}

	return []interface{}{m}
}

func flattenAudioWatermarkSettings(ns *types.AudioWatermarkSettings) []interface{} {
	if ns == nil {
		return nil
	}

	m := map[string]interface{}{
		"nielsen_watermark_settings": func(n *types.NielsenWatermarksSettings) []interface{} {
			if n == nil {
				return nil
			}

			m := map[string]interface{}{
				"nielsen_distribution_type":   string(n.NielsenDistributionType),
				"nielsen_cbet_settings":       flattenNielsenCbetSettings(n.NielsenCbetSettings),
				"nielsen_naes_ii_nw_settings": flattenNielsenNaesIiNwSettings(n.NielsenNaesIiNwSettings),
			}

			return []interface{}{m}
		}(ns.NielsenWatermarksSettings),
	}

	return []interface{}{m}
}

func flattenAudioDescriptionsCodecSettings(in *types.AudioCodecSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"aac_settings":        flattenCodecSettingsAacSettings(in.AacSettings),
		"ac3_settings":        flattenCodecSettingsAc3Settings(in.Ac3Settings),
		"eac3_atmos_settings": flattenCodecSettingsEac3AtmosSettings(in.Eac3AtmosSettings),
		"eac3_settings":       flattenCodecSettingsEac3Settings(in.Eac3Settings),
		"mp2_settings":        flattenCodecSettingsMp2Settings(in.Mp2Settings),
		"wav_settings":        flattenCodecSettingsWavSettings(in.WavSettings),
	}

	if in.PassThroughSettings != nil {
		m["pass_through_settings"] = []interface{}{} // no exported fields
	}

	return []interface{}{m}
}

func flattenCodecSettingsAacSettings(in *types.AacSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"bitrate":           float32(in.Bitrate),
		"coding_mode":       string(in.CodingMode),
		"input_type":        string(in.InputType),
		"profile":           string(in.Profile),
		"rate_control_mode": string(in.RateControlMode),
		"raw_format":        string(in.RawFormat),
		"sample_rate":       float32(in.SampleRate),
		"spec":              string(in.Spec),
		"vbr_quality":       string(in.VbrQuality),
	}

	return []interface{}{m}
}

func flattenCodecSettingsAc3Settings(in *types.Ac3Settings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"bitrate":          float32(in.Bitrate),
		"bitstream_mode":   string(in.BitstreamMode),
		"coding_mode":      string(in.CodingMode),
		"dialnorm":         int(in.Dialnorm),
		"drc_profile":      string(in.DrcProfile),
		"lfe_filter":       string(in.LfeFilter),
		"metadata_control": string(in.MetadataControl),
	}

	return []interface{}{m}
}

func flattenCodecSettingsEac3AtmosSettings(in *types.Eac3AtmosSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"bitrate":       float32(in.Bitrate),
		"coding_mode":   string(in.CodingMode),
		"dialnorm":      int(in.Dialnorm),
		"drc_line":      string(in.DrcLine),
		"drc_rf":        string(in.DrcRf),
		"height_trim":   float32(in.HeightTrim),
		"surround_trim": float32(in.SurroundTrim),
	}

	return []interface{}{m}
}

func flattenCodecSettingsEac3Settings(in *types.Eac3Settings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"attenuation_control":      string(in.AttenuationControl),
		"bitrate":                  float32(in.Bitrate),
		"bitstream_mode":           string(in.BitstreamMode),
		"coding_mode":              string(in.CodingMode),
		"dc_filter":                string(in.DcFilter),
		"dialnorm":                 int(in.Dialnorm),
		"drc_line":                 string(in.DrcLine),
		"drc_rf":                   string(in.DrcRf),
		"lfe_control":              string(in.LfeControl),
		"lfe_filter":               string(in.LfeFilter),
		"lo_ro_center_mix_level":   float32(in.LoRoCenterMixLevel),
		"lo_ro_surround_mix_level": float32(in.LoRoSurroundMixLevel),
		"lt_rt_center_mix_level":   float32(in.LtRtCenterMixLevel),
		"lt_rt_surround_mix_level": float32(in.LtRtSurroundMixLevel),
		"metadata_control":         string(in.MetadataControl),
		"passthrough_control":      string(in.PassthroughControl),
		"phase_control":            string(in.PhaseControl),
		"stereo_downmix":           string(in.StereoDownmix),
		"surround_ex_mode":         string(in.SurroundExMode),
		"surround_mode":            string(in.SurroundMode),
	}

	return []interface{}{m}
}

func flattenCodecSettingsMp2Settings(in *types.Mp2Settings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"bitrate":     float32(in.Bitrate),
		"coding_mode": string(in.CodingMode),
		"sample_rate": float32(in.SampleRate),
	}

	return []interface{}{m}
}

func flattenCodecSettingsWavSettings(in *types.WavSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"bit_depth":   float32(in.BitDepth),
		"coding_mode": string(in.CodingMode),
		"sample_rate": float32(in.SampleRate),
	}

	return []interface{}{m}
}

func flattenAudioDescriptionsRemixSettings(in *types.RemixSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"channel_mappings": flattenChannelMappings(in.ChannelMappings),
		"channels_in":      int(in.ChannelsIn),
		"channels_out":     int(in.ChannelsOut),
	}

	return []interface{}{m}
}

func flattenChannelMappings(in []types.AudioChannelMapping) []interface{} {
	if len(in) == 0 {
		return nil
	}

	var out []interface{}
	for _, item := range in {
		m := map[string]interface{}{
			"input_channel_levels": flattenInputChannelLevels(item.InputChannelLevels),
			"output_channel":       int(item.OutputChannel),
		}

		out = append(out, m)
	}

	return out
}

func flattenInputChannelLevels(in []types.InputChannelLevel) []interface{} {
	if len(in) == 0 {
		return nil
	}

	var out []interface{}
	for _, item := range in {
		m := map[string]interface{}{
			"gain":          int(item.Gain),
			"input_channel": int(item.InputChannel),
		}

		out = append(out, m)
	}

	return out
}

func flattenNielsenCbetSettings(in *types.NielsenCBET) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"cbet_check_digit_string": aws.ToString(in.CbetCheckDigitString),
		"cbet_stepaside":          string(in.CbetStepaside),
		"csid":                    aws.ToString(in.Csid),
	}

	return []interface{}{m}
}

func flattenNielsenNaesIiNwSettings(in *types.NielsenNaesIiNw) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"check_digit_string": aws.ToString(in.CheckDigitString),
		"sid":                float32(in.Sid),
	}

	return []interface{}{m}
}
