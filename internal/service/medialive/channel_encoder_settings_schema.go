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
							"hls_settings": hlsSettingsSchema(),
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
							"audio_only_image": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"uri": {
											Type:     schema.TypeString,
											Required: true,
										},
										"password_param": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
										"username": {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
									},
								},
							},
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
	// TODO implement rest of output group settings

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
	if v, ok := m["media_package_output_settings"].([]interface{}); ok && len(v) > 0 {
		os.MediaPackageOutputSettings = &types.MediaPackageOutputSettings{} // no exported fields
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
			d.CodecSettings = nil // TODO expandChannelEncoderSettingsVideoDescriptionsCodecSettings(v)
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
			"language_control_mode":        string(v.LanguageCodeControl),
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
		"archive_output_settings": flattenOutputsOutputSettingsArchiveOutputSettings(in.ArchiveOutputSettings),
		"media_package_output_settings": func(inner *types.MediaPackageOutputSettings) []interface{} {
			if inner == nil {
				return nil
			}
			return []interface{}{} // no exported attributes
		}(in.MediaPackageOutputSettings),
		"multiplex_output_settings": func(inner *types.MultiplexOutputSettings) []interface{} {
			if inner == nil {
				return nil
			}
			data := map[string]interface{}{
				"destination": flattenDestination(inner.Destination),
			}

			return []interface{}{data} // no exported attributes
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

func flattenFrameCaptureCDNSettings(in *types.FrameCaptureCdnSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"frame_capture_s3_settings": flattenFrameCaptureS3Settings(in.FrameCaptureS3Settings),
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

func flattenVideoDescriptionsCodecSettings(in *types.VideoCodecSettings) []interface{} {
	if in == nil {
		return nil
	}

	m := map[string]interface{}{
		"frame_capture_settings": flattenCodecSettingsFrameCaptureSettings(in.FrameCaptureSettings),
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
