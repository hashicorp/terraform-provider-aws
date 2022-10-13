package medialive

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
					Required: true,
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
					Type:     schema.TypeSet,
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
														Computed: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"archive_s3_settings": {
																	Type:     schema.TypeList,
																	Optional: true,
																	Computed: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"canned_acl": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				Computed:         true,
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
														Computed: true,
													},
												},
											},
										},
										"frame_capture_group_settings": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
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
																				Computed:         true,
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
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"destination": func() *schema.Schema {
														return destinationSchema()
													}(),
													"ad_markers": {
														Type:     schema.TypeList,
														Optional: true,
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
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"canned_acl": {
																				Type:             schema.TypeString,
																				Optional:         true,
																				Computed:         true,
																				ValidateDiagFunc: enum.Validate[types.S3CannedAcl](),
																			},
																		},
																	},
																},
																"hls_webdav_settings": {
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
													"keep_segment": {
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
													"time_metadata_id3_frame": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[types.HlsTimedMetadataId3Frame](),
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
										"output_settings": {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"archive_output_settings": {
														Type:     schema.TypeList,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"container_settings": {
																	Type:     schema.TypeList,
																	Required: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"m2ts_settings": m2tsSettingsSchema(),
																			// This is in the API and Go SDK docs, but has no exported fields.
																			// "raw_settings": {
																			// 	Type:     schema.TypeList,
																			// 	MaxItems: 1,
																			// 	Elem: &schema.Resource{
																			// 		Schema: map[string]*schema.Schema{},
																			// 	},
																			// },
																		},
																	},
																},
																"extension": {
																	Type:     schema.TypeString,
																	Optional: true,
																	Computed: true,
																},
																"name_modifier": {
																	Type:     schema.TypeString,
																	Optional: true,
																	Computed: true,
																},
															},
														},
													},
													"frame_capture_output_settings": {
														Type:     schema.TypeList,
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
													"hls_output_settings":           {}, // TODO
													"media_package_output_settings": {}, // TODO
													"ms_smooth_output_settings":     {}, // TODO
													"multiplex_output_settings":     {}, // TODO
													"rtmp_output_settings":          {}, // TODO
													"udp_output_settings":           {}, // TODO
												},
											},
										},
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
								Computed: true,
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
							},
						},
					},
				},
				"video_descriptions": {
					Type:     schema.TypeSet,
					Required: true,
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
										"h_264_settings": {
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
										// TODO h_265_settings
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
				// TODO avail_blanking
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

func m2tsSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
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
					Computed:         true,
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
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsAribCaptionsPidControl](),
				},
				"audio_buffer_model": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsAudioBufferModel](),
				},
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
				"audio_stream_type": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsAudioStreamType](),
				},
				"bitrate": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"buffer_model": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsBufferModel](),
				},
				"cc_descriptor": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
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
								Computed: true,
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
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.DvbSdtOutputSdt](),
							},
							"rep_interval": {
								Type:     schema.TypeInt,
								Optional: true,
								Computed: true,
							},
							"service_name": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
							"service_provider_name": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
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
								Computed: true,
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
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsEbifControl](),
				},
				"ebp_audio_interval": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsAudioInterval](),
				},
				"ebp_lookahead_ms": {
					Type:     schema.TypeInt,
					Optional: true,
					Computed: true,
				},
				"ebp_placement": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsEbpPlacement](),
				},
				"ecm_pid": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"es_rate_in_pes": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
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
					Computed: true,
				},
				"klv": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
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
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsNielsenId3Behavior](),
				},
				"null_packet_bitrate": {
					Type:     schema.TypeFloat,
					Optional: true,
					Computed: true,
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
					ValidateDiagFunc: enum.Validate[types.M2tsPcrControl](),
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
				"rate_mode": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
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
					Computed:         true,
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
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsSegmentationMarkers](),
				},
				"segmentation_style": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[types.M2tsSegmentationStyle](),
				},
				"segmentation_time": {
					Type:     schema.TypeFloat,
					Optional: true,
					Computed: true,
				},
				"timed_metadata_behavior": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
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
					Computed: true,
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
	if v, ok := m["audio_descriptions"].([]interface{}); ok && len(v) > 0 {
		settings.AudioDescriptions = expandChannelEncoderSettingsAudioDescriptions(v)
	}
	if v, ok := m["output_groups"].([]interface{}); ok && len(v) > 0 {
		settings.OutputGroups = expandChannelEncoderSettingsOutputGroups(v)
	}
	if v, ok := m["timecode_config"].([]interface{}); ok && len(v) > 0 {
		settings.TimecodeConfig = expandChannelEncoderSettingsTimecodeConfig(v)
	}
	if v, ok := m["video_descriptions"].([]interface{}); ok && len(v) > 0 {
		settings.VideoDescriptions = expandChannelEncoderSettingsVideoDescriptions(v)
	}
	if v, ok := m["avail_blanking"].([]interface{}); ok && len(v) > 0 {
		settings.AvailBlanking = nil // TODO expandChannelEncoderSettingsAvailBlanking(v)
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
			a.AudioNormalizationSettings = nil // TODO expandChannelEncoderSettingsAudioDescriptionsAudioNormalizationSettings(v)
		}
		if v, ok := m["audio_type"].(string); ok && v != "" {
			a.AudioType = types.AudioType(v)
		}
		if v, ok := m["audio_type_control"].(string); ok && v != "" {
			a.AudioTypeControl = types.AudioDescriptionAudioTypeControl(v)
		}
		if v, ok := m["audio_watermark_settings"].([]interface{}); ok && len(v) > 0 {
			a.AudioWatermarkingSettings = nil // TODO expandChannelEncoderSettingsAudioDescriptionsAudioWatermarkSettings(v)
		}
		if v, ok := m["codec_settings"].([]interface{}); ok && len(v) > 0 {
			a.CodecSettings = nil // TODO expandChannelEncoderSettingsAudioDescriptionsCodecSettings(v)
		}
		if v, ok := m["language_code"].(string); ok && v != "" {
			a.LanguageCode = aws.String(v)
		}
		if v, ok := m["language_code_control"].(string); ok && v != "" {
			a.LanguageCodeControl = types.AudioDescriptionLanguageCodeControl(v)
		}
		if v, ok := m["remix_settings"].([]interface{}); ok && len(v) > 0 {
			a.RemixSettings = nil // TODO expandChannelEncoderSettingsAudioDescriptionsRemixSettings(v)
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
			o.OutputGroupSettings = nil // TODO expandChannelEncoderSettingsOutputGroupsOutputGroupSettings(v)
		}
		if v, ok := m["outputs"].([]interface{}); ok && len(v) > 0 {
			o.Outputs = nil // TODO expandChannelEncoderSettingsOutputGroupsOutputs(v)
		}
		if v, ok := m["name"].(string); ok && v != "" {
			o.Name = aws.String(v)
		}

		outputGroups = append(outputGroups, o)
	}

	return outputGroups
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
			d.CodecSettings = nil // TODO expandChannelEncoderSettingsVideoDescriptionsCodecSettings(v)
		}
		if v, ok := m["height"].(int32); ok {
			d.Height = v
		}
		if v, ok := m["respond_to_afd"].(string); ok && v != "" {
			d.RespondToAfd = types.VideoDescriptionRespondToAfd(v)
		}
		if v, ok := m["scaling_behavior"].(string); ok && v != "" {
			d.ScalingBehavior = types.VideoDescriptionScalingBehavior(v)
		}
		if v, ok := m["sharpness"].(int32); ok {
			d.Sharpness = v
		}
		if v, ok := m["width"].(int32); ok {
			d.Width = v
		}

		videoDesc = append(videoDesc, d)
	}

	return videoDesc
}

func flattenChannelEncoderSettings(apiObject *types.EncoderSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"audio_descriptions": nil, // TODO
		"output_groups":      nil, // TODO
		"timecode_config":    nil, // TODO
		"video_descriptions": nil, // TODO
		// TODO avail_blanking
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
