package medialive

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cdi_input_specification": {
				Type:     schema.TypeList,
				Required: true,
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
				ValidateDiagFunc: enum.Validate[types.ChannelClass](),
			},
			"destinations": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
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
										Required: true,
									},
									"stream_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"url": {
										Type:     schema.TypeString,
										Required: true,
									},
									"username": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"encoder_settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audio_description": {
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
																Type:             schema.TypeList,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[types.HlsAdMarkers](),
																Elem:             &schema.Schema{Type: schema.TypeString},
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
																Type:             schema.TypeList,
																Optional:         true,
																Elem:             &schema.Schema{Type: schema.TypeString},
																ValidateDiagFunc: enum.Validate[types.RtmpAdMarkers](),
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
										Type:     schema.TypeSet,
										Required: true,
									},
									"name": {
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
			},
			"vpc": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"public_address_allocation_ids": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"security_group_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameChannel = "Channel"
)

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	in := &medialive.CreateChannelInput{
		Name:      aws.String(d.Get("name").(string)),
		RequestId: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("cdi_input_specification"); ok && len(v.([]interface{})) > 0 {
		in.CdiInputSpecification = expandChannelCdiInputSpecification(v.([]interface{}))
	}
	if v, ok := d.GetOk("destinations"); ok && len(v.([]interface{})) > 0 {
		in.Destinations = expandChannelDestinations(v.([]interface{}))
	}
	if v, ok := d.GetOk("input_specification"); ok && len(v.([]interface{})) > 0 {
		in.InputSpecification = expandChannelInputSpecification(v.([]interface{}))
	}
	if v, ok := d.GetOk("maintenance"); ok && len(v.([]interface{})) > 0 {
		in.Maintenance = expandChannelMaintenanceCreate(v.([]interface{}))
	}
	if v, ok := d.GetOk("vpc"); ok && len(v.([]interface{})) > 0 {
		in.Vpc = expandChannelVpc(v.([]interface{}))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateChannel(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameChannel, d.Get("name").(string), err)
	}

	if out == nil || out.Channel == nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameChannel, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Channel.Id))

	if _, err := waitChannelCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForCreation, ResNameChannel, d.Id(), err)
	}

	return resourceChannelRead(ctx, d, meta)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	out, err := FindChannelByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaLive Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionReading, ResNameChannel, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	if err := d.Set("cdi_input_specification", flattenChannelCdiInputSpecification(out.CdiInputSpecification)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("destinations", flattenChannelDestinations(out.Destinations)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("input_specification", flattenChannelInputSpecification(out.InputSpecification)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("maintenance", flattenChannelMaintenance(out.Maintenance)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}
	if err := d.Set("vpc", flattenChannelVpc(out.Vpc)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, aws.ToString(out.Arn))
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionReading, ResNameChannel, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}

	return nil
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	update := false

	in := &medialive.UpdateChannelInput{
		ChannelId: aws.String(d.Id()),
	}

	if d.HasChanges(
		"name",
		"cdi_input_specification",
		"destinations",
		"input_specification",
		"maintenance",
	) {
		update = true

		in.Name = aws.String(d.Get("name").(string))
		if v, ok := d.GetOk("cdi_input_specification"); ok {
			in.CdiInputSpecification = expandChannelCdiInputSpecification(v.([]interface{}))
		}
		if v, ok := d.GetOk("destinations"); ok {
			in.Destinations = expandChannelDestinations(v.([]interface{}))
		}
		if v, ok := d.GetOk("input_specification"); ok {
			in.InputSpecification = expandChannelInputSpecification(v.([]interface{}))
		}
		if v, ok := d.GetOk("maintenance"); ok {
			in.Maintenance = expandChannelMaintenanceUpdate(v.([]interface{}))
		}
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating MediaLive Channel (%s): %#v", d.Id(), in)
	out, err := conn.UpdateChannel(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionUpdating, ResNameChannel, d.Id(), err)
	}

	if _, err := waitChannelUpdated(ctx, conn, aws.ToString(out.Channel.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForUpdate, ResNameChannel, d.Id(), err)
	}

	return resourceChannelRead(ctx, d, meta)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	log.Printf("[INFO] Deleting MediaLive Channel %s", d.Id())

	_, err := conn.DeleteChannel(ctx, &medialive.DeleteChannelInput{
		ChannelId: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.MediaLive, create.ErrActionDeleting, ResNameChannel, d.Id(), err)
	}

	if _, err := waitChannelDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForDeletion, ResNameChannel, d.Id(), err)
	}

	return nil
}

func waitChannelCreated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeChannelOutput, error) {
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
		Pending: enum.Slice(types.ChannelStateDeleting),
		Target:  enum.Slice(types.ChannelStateDeleted),
		Refresh: statusChannel(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func statusChannel(ctx context.Context, conn *medialive.Client, id string) resource.StateRefreshFunc {
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
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
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
		if v, ok := m["id"].(string); ok {
			d.Id = aws.String(v)
		}
		if v, ok := m["media_package_settings"].([]interface{}); ok {
			d.MediaPackageSettings = expandChannelDestinationsMediaPackageSettings(v)
		}
		if v, ok := m["multiplex_settings"].([]interface{}); ok {
			d.MultiplexSettings = expandChannelDestinationsMultiplexSettings(v)
		}
		if v, ok := m["settings"].([]interface{}); ok {
			d.Settings = expandChannelDestinationsSettings(v)
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
		if v, ok := m["url"].(string); ok {
			s.Url = aws.String(v)
		}
		if v, ok := m["username"].(string); ok {
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
			"id":                     aws.ToString(v.Id),
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
			"password_param": aws.ToString(v.PasswordParam),
			"stream_name":    aws.ToString(v.StreamName),
			"url":            aws.ToString(v.Url),
			"username":       aws.ToString(v.Username),
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
		// "maintenance_scheduled_date": "",
	}

	return []interface{}{m}
}

func expandChannelVpc(tfList []interface{}) *types.VpcOutputSettings {
	if tfList == nil {
		return nil
	}
	m := tfList[0].(map[string]interface{})

	settings := &types.VpcOutputSettings{}
	if v, ok := m["security_group_ids"].([]string); ok && len(v) > 0 {
		settings.SecurityGroupIds = v
	}
	if v, ok := m["subnet_ids"].([]string); ok && len(v) > 0 {
		settings.SubnetIds = v
	}
	if v, ok := m["public_address_allocation_ids"].([]string); ok && len(v) > 0 {
		settings.PublicAddressAllocationIds = v
	}

	return settings
}

func flattenChannelVpc(apiObject *types.VpcOutputSettingsDescription) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"security_group_ids": flex.FlattenStringValueList(apiObject.SecurityGroupIds),
		"subnet_ids":         flex.FlattenStringValueList(apiObject.SubnetIds),
		// public_address_allocation_ids is not included in the output struct
	}

	return []interface{}{m}
}
