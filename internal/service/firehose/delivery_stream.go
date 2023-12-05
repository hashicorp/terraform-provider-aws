// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package firehose

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	destinationTypeElasticsearch        = "elasticsearch"
	destinationTypeExtendedS3           = "extended_s3"
	destinationTypeHTTPEndpoint         = "http_endpoint"
	destinationTypeOpenSearch           = "opensearch"
	destinationTypeOpenSearchServerless = "opensearchserverless"
	destinationTypeRedshift             = "redshift"
	destinationTypeSplunk               = "splunk"
)

func destinationType_Values() []string {
	return []string{
		destinationTypeElasticsearch,
		destinationTypeExtendedS3,
		destinationTypeHTTPEndpoint,
		destinationTypeOpenSearch,
		destinationTypeOpenSearchServerless,
		destinationTypeRedshift,
		destinationTypeSplunk,
	}
}

// @SDKResource("aws_kinesis_firehose_delivery_stream", name="Delivery Stream")
// @Tags(identifierAttribute="name")
func ResourceDeliveryStream() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeliveryStreamCreate,
		ReadWithoutTimeout:   resourceDeliveryStreamRead,
		UpdateWithoutTimeout: resourceDeliveryStreamUpdate,
		DeleteWithoutTimeout: resourceDeliveryStreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idErr := fmt.Errorf("Expected ID in format of arn:PARTITION:firehose:REGION:ACCOUNTID:deliverystream/NAME and provided: %s", d.Id())
				resARN, err := arn.Parse(d.Id())
				if err != nil {
					return nil, idErr
				}
				resourceParts := strings.Split(resARN.Resource, "/")
				if len(resourceParts) != 2 {
					return nil, idErr
				}
				d.Set("name", resourceParts[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  MigrateState,
		SchemaFunc: func() map[string]*schema.Schema {
			cloudWatchLoggingOptionsSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enabled": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							"log_group_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"log_stream_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				}
			}
			dynamicPartitioningConfigurationSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					ForceNew: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enabled": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
								ForceNew: true,
							},
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
						},
					},
				}
			}
			processingConfigurationSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:             schema.TypeList,
					Optional:         true,
					MaxItems:         1,
					DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enabled": {
								Type:     schema.TypeBool,
								Optional: true,
							},
							"processors": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"parameters": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"parameter_name": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringInSlice(firehose.ProcessorParameterName_Values(), false),
													},
													"parameter_value": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringLenBetween(1, 5120),
													},
												},
											},
										},
										"type": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(firehose.ProcessorType_Values(), false),
										},
									},
								},
							},
						},
					},
				}
			}
			requestConfigurationSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"common_attributes": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:     schema.TypeString,
											Required: true,
										},
										"value": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"content_encoding": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.ContentEncodingNone,
								ValidateFunc: validation.StringInSlice(firehose.ContentEncoding_Values(), false),
							},
						},
					},
				}
			}
			s3ConfigurationElem := func() *schema.Resource {
				return &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"buffering_interval": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      300,
							ValidateFunc: validation.IntAtLeast(60),
						},
						"buffering_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
						"compression_format": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      firehose.CompressionFormatUncompressed,
							ValidateFunc: validation.StringInSlice(firehose.CompressionFormat_Values(), false),
						},
						"error_output_prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"kms_key_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				}
			}
			s3BackupConfigurationSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem:     s3ConfigurationElem(),
				}
			}
			s3ConfigurationSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					MaxItems: 1,
					Required: true,
					Elem:     s3ConfigurationElem(),
				}
			}

			return map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"destination": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					StateFunc: func(v interface{}) string {
						value := v.(string)
						return strings.ToLower(value)
					},
					ValidateFunc: validation.StringInSlice(destinationType_Values(), false),
				},
				"destination_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"elasticsearch_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"buffering_interval": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(60, 900),
							},
							"buffering_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      5,
								ValidateFunc: validation.IntBetween(1, 100),
							},
							"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
							"cluster_endpoint": {
								Type:          schema.TypeString,
								Optional:      true,
								ConflictsWith: []string{"elasticsearch_configuration.0.domain_arn"},
							},
							"domain_arn": {
								Type:          schema.TypeString,
								Optional:      true,
								ValidateFunc:  verify.ValidARN,
								ConflictsWith: []string{"elasticsearch_configuration.0.cluster_endpoint"},
							},
							"index_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"index_rotation_period": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.ElasticsearchIndexRotationPeriodOneDay,
								ValidateFunc: validation.StringInSlice(firehose.ElasticsearchIndexRotationPeriod_Values(), false),
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							"role_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_mode": {
								Type:         schema.TypeString,
								ForceNew:     true,
								Optional:     true,
								Default:      firehose.ElasticsearchS3BackupModeFailedDocumentsOnly,
								ValidateFunc: validation.StringInSlice(firehose.ElasticsearchS3BackupMode_Values(), false),
							},
							"s3_configuration": s3ConfigurationSchema(),
							"type_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 100),
							},
							"vpc_config": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"role_arn": {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
										"security_group_ids": {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										"subnet_ids": {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										"vpc_id": {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
				"extended_s3_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"bucket_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"buffering_interval": {
								Type:     schema.TypeInt,
								Optional: true,
								Default:  300,
							},
							"buffering_size": {
								Type:     schema.TypeInt,
								Optional: true,
								Default:  5,
							},
							"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
							"compression_format": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.CompressionFormatUncompressed,
								ValidateFunc: validation.StringInSlice(firehose.CompressionFormat_Values(), false),
							},
							"data_format_conversion_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"enabled": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  true,
										},
										"input_format_configuration": {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"deserializer": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"hive_json_ser_de": {
																	Type:          schema.TypeList,
																	Optional:      true,
																	MaxItems:      1,
																	ConflictsWith: []string{"extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.open_x_json_ser_de"},
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"timestamp_formats": {
																				Type:     schema.TypeList,
																				Optional: true,
																				Elem:     &schema.Schema{Type: schema.TypeString},
																			},
																		},
																	},
																},
																"open_x_json_ser_de": {
																	Type:          schema.TypeList,
																	Optional:      true,
																	MaxItems:      1,
																	ConflictsWith: []string{"extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.hive_json_ser_de"},
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"case_insensitive": {
																				Type:     schema.TypeBool,
																				Optional: true,
																				Default:  true,
																			},
																			"column_to_json_key_mappings": {
																				Type:     schema.TypeMap,
																				Optional: true,
																				Elem:     &schema.Schema{Type: schema.TypeString},
																			},
																			"convert_dots_in_json_keys_to_underscores": {
																				Type:     schema.TypeBool,
																				Optional: true,
																				Default:  false,
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
										"output_format_configuration": {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"serializer": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"orc_ser_de": {
																	Type:          schema.TypeList,
																	Optional:      true,
																	MaxItems:      1,
																	ConflictsWith: []string{"extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.parquet_ser_de"},
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"block_size_bytes": {
																				Type:     schema.TypeInt,
																				Optional: true,
																				// 256 MiB
																				Default: 268435456,
																				// 64 MiB
																				ValidateFunc: validation.IntAtLeast(67108864),
																			},
																			"bloom_filter_columns": {
																				Type:     schema.TypeList,
																				Optional: true,
																				Elem:     &schema.Schema{Type: schema.TypeString},
																			},
																			"bloom_filter_false_positive_probability": {
																				Type:     schema.TypeFloat,
																				Optional: true,
																				Default:  0.05,
																			},
																			"compression": {
																				Type:         schema.TypeString,
																				Optional:     true,
																				Default:      firehose.OrcCompressionSnappy,
																				ValidateFunc: validation.StringInSlice(firehose.OrcCompression_Values(), false),
																			},
																			"dictionary_key_threshold": {
																				Type:     schema.TypeFloat,
																				Optional: true,
																				Default:  0.0,
																			},
																			"enable_padding": {
																				Type:     schema.TypeBool,
																				Optional: true,
																				Default:  false,
																			},
																			"format_version": {
																				Type:         schema.TypeString,
																				Optional:     true,
																				Default:      firehose.OrcFormatVersionV012,
																				ValidateFunc: validation.StringInSlice(firehose.OrcFormatVersion_Values(), false),
																			},
																			"padding_tolerance": {
																				Type:     schema.TypeFloat,
																				Optional: true,
																				Default:  0.05,
																			},
																			"row_index_stride": {
																				Type:         schema.TypeInt,
																				Optional:     true,
																				Default:      10000,
																				ValidateFunc: validation.IntAtLeast(1000),
																			},
																			"stripe_size_bytes": {
																				Type:     schema.TypeInt,
																				Optional: true,
																				// 64 MiB
																				Default: 67108864,
																				// 8 MiB
																				ValidateFunc: validation.IntAtLeast(8388608),
																			},
																		},
																	},
																},
																"parquet_ser_de": {
																	Type:          schema.TypeList,
																	Optional:      true,
																	MaxItems:      1,
																	ConflictsWith: []string{"extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.orc_ser_de"},
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"block_size_bytes": {
																				Type:     schema.TypeInt,
																				Optional: true,
																				// 256 MiB
																				Default: 268435456,
																				// 64 MiB
																				ValidateFunc: validation.IntAtLeast(67108864),
																			},
																			"compression": {
																				Type:         schema.TypeString,
																				Optional:     true,
																				Default:      firehose.ParquetCompressionSnappy,
																				ValidateFunc: validation.StringInSlice(firehose.ParquetCompression_Values(), false),
																			},
																			"enable_dictionary_compression": {
																				Type:     schema.TypeBool,
																				Optional: true,
																				Default:  false,
																			},
																			"max_padding_bytes": {
																				Type:     schema.TypeInt,
																				Optional: true,
																				Default:  0,
																			},
																			"page_size_bytes": {
																				Type:     schema.TypeInt,
																				Optional: true,
																				// 1 MiB
																				Default: 1048576,
																				// 64 KiB
																				ValidateFunc: validation.IntAtLeast(65536),
																			},
																			"writer_version": {
																				Type:         schema.TypeString,
																				Optional:     true,
																				Default:      firehose.ParquetWriterVersionV1,
																				ValidateFunc: validation.StringInSlice(firehose.ParquetWriterVersion_Values(), false),
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
										"schema_configuration": {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"catalog_id": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"database_name": {
														Type:     schema.TypeString,
														Required: true,
													},
													"region": {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													"role_arn": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: verify.ValidARN,
													},
													"table_name": {
														Type:     schema.TypeString,
														Required: true,
													},
													"version_id": {
														Type:     schema.TypeString,
														Optional: true,
														Default:  "LATEST",
													},
												},
											},
										},
									},
								},
							},
							"dynamic_partitioning_configuration": dynamicPartitioningConfigurationSchema(),
							"error_output_prefix": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							"kms_key_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
							"prefix": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"processing_configuration": processingConfigurationSchema(),
							"role_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_configuration": s3BackupConfigurationSchema(),
							"s3_backup_mode": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.S3BackupModeDisabled,
								ValidateFunc: validation.StringInSlice(firehose.S3BackupMode_Values(), false),
							},
						},
					},
				},
				"http_endpoint_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"access_key": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 4096),
								Sensitive:    true,
							},
							"buffering_interval": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(60, 900),
							},
							"buffering_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      5,
								ValidateFunc: validation.IntBetween(1, 100),
							},
							"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
							"name": {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 256),
								),
							},
							"processing_configuration": processingConfigurationSchema(),
							"request_configuration":    requestConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							"role_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_mode": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.HttpEndpointS3BackupModeFailedDataOnly,
								ValidateFunc: validation.StringInSlice(firehose.HttpEndpointS3BackupMode_Values(), false),
							},
							"s3_configuration": s3ConfigurationSchema(),
							"url": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 1000),
									validation.StringMatch(regexache.MustCompile(`^https://.*$`), ""),
								),
							},
						},
					},
				},
				"kinesis_source_configuration": {
					Type:          schema.TypeList,
					ForceNew:      true,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{"msk_source_configuration", "server_side_encryption"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"kinesis_stream_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ForceNew:     true,
								ValidateFunc: verify.ValidARN,
							},
							"role_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ForceNew:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"msk_source_configuration": {
					Type:          schema.TypeList,
					ForceNew:      true,
					Optional:      true,
					MaxItems:      1,
					ConflictsWith: []string{"kinesis_source_configuration", "server_side_encryption"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"authentication_configuration": {
								Type:     schema.TypeList,
								ForceNew: true,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"connectivity": {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: validation.StringInSlice(firehose.Connectivity_Values(), false),
										},
										"role_arn": {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
							"msk_cluster_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ForceNew:     true,
								ValidateFunc: verify.ValidARN,
							},
							"topic_name": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
						},
					},
				},
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 64),
				},
				"opensearch_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"buffering_interval": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(60, 900),
							},
							"buffering_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      5,
								ValidateFunc: validation.IntBetween(1, 100),
							},
							"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
							"cluster_endpoint": {
								Type:          schema.TypeString,
								Optional:      true,
								ConflictsWith: []string{"opensearch_configuration.0.domain_arn"},
							},
							"domain_arn": {
								Type:          schema.TypeString,
								Optional:      true,
								ValidateFunc:  verify.ValidARN,
								ConflictsWith: []string{"opensearch_configuration.0.cluster_endpoint"},
							},
							"index_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"index_rotation_period": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.AmazonopensearchserviceIndexRotationPeriodOneDay,
								ValidateFunc: validation.StringInSlice(firehose.AmazonopensearchserviceIndexRotationPeriod_Values(), false),
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							"role_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_mode": {
								Type:         schema.TypeString,
								ForceNew:     true,
								Optional:     true,
								Default:      firehose.AmazonopensearchserviceS3BackupModeFailedDocumentsOnly,
								ValidateFunc: validation.StringInSlice(firehose.AmazonopensearchserviceS3BackupMode_Values(), false),
							},
							"s3_configuration": s3ConfigurationSchema(),
							"type_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 100),
							},
							"vpc_config": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"role_arn": {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
										"security_group_ids": {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										"subnet_ids": {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										"vpc_id": {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
				"opensearchserverless_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"buffering_interval": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(60, 900),
							},
							"buffering_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      5,
								ValidateFunc: validation.IntBetween(1, 100),
							},
							"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
							"collection_endpoint": {
								Type:     schema.TypeString,
								Required: true,
							},
							"index_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							"role_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_mode": {
								Type:         schema.TypeString,
								ForceNew:     true,
								Optional:     true,
								Default:      firehose.AmazonOpenSearchServerlessS3BackupModeFailedDocumentsOnly,
								ValidateFunc: validation.StringInSlice(firehose.AmazonOpenSearchServerlessS3BackupMode_Values(), false),
							},
							"s3_configuration": s3ConfigurationSchema(),
							"vpc_config": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"role_arn": {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
										"security_group_ids": {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										"subnet_ids": {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										"vpc_id": {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},
				"redshift_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
							"cluster_jdbcurl": {
								Type:     schema.TypeString,
								Required: true,
							},
							"copy_options": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"data_table_columns": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"data_table_name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"password": {
								Type:      schema.TypeString,
								Required:  true,
								Sensitive: true,
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      3600,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							"role_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_configuration": s3BackupConfigurationSchema(),
							"s3_backup_mode": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.S3BackupModeDisabled,
								ValidateFunc: validation.StringInSlice(firehose.S3BackupMode_Values(), false),
							},
							"s3_configuration": s3ConfigurationSchema(),
							"username": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"server_side_encryption": {
					Type:             schema.TypeList,
					Optional:         true,
					MaxItems:         1,
					DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
					ConflictsWith:    []string{"kinesis_source_configuration", "msk_source_configuration"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enabled": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							"key_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
								RequiredWith: []string{"server_side_encryption.0.enabled", "server_side_encryption.0.key_type"},
							},
							"key_type": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.KeyTypeAwsOwnedCmk,
								ValidateFunc: validation.StringInSlice(firehose.KeyType_Values(), false),
								RequiredWith: []string{"server_side_encryption.0.enabled"},
							},
						},
					},
				},
				"splunk_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
							"hec_acknowledgment_timeout": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      180,
								ValidateFunc: validation.IntBetween(180, 600),
							},
							"hec_endpoint": {
								Type:     schema.TypeString,
								Required: true,
							},
							"hec_endpoint_type": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.HECEndpointTypeRaw,
								ValidateFunc: validation.StringInSlice(firehose.HECEndpointType_Values(), false),
							},
							"hec_token": {
								Type:     schema.TypeString,
								Required: true,
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      3600,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							"s3_backup_mode": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      firehose.SplunkS3BackupModeFailedEventsOnly,
								ValidateFunc: validation.StringInSlice(firehose.SplunkS3BackupMode_Values(), false),
							},
							"s3_configuration": s3ConfigurationSchema(),
						},
					},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"version_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
			}
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
				destination := d.Get("destination").(string)
				requiredAttribute := map[string]string{
					destinationTypeElasticsearch:        "elasticsearch_configuration",
					destinationTypeExtendedS3:           "extended_s3_configuration",
					destinationTypeHTTPEndpoint:         "http_endpoint_configuration",
					destinationTypeOpenSearch:           "opensearch_configuration",
					destinationTypeOpenSearchServerless: "opensearchserverless_configuration",
					destinationTypeRedshift:             "redshift_configuration",
					destinationTypeSplunk:               "splunk_configuration",
				}[destination]

				if _, ok := d.GetOk(requiredAttribute); !ok {
					return fmt.Errorf("when destination is '%s', %s is required", destination, requiredAttribute)
				}

				return nil
			},
		),
	}
}

func resourceDeliveryStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FirehoseConn(ctx)

	sn := d.Get("name").(string)
	input := &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(sn),
		DeliveryStreamType: aws.String(firehose.DeliveryStreamTypeDirectPut),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("kinesis_source_configuration"); ok {
		input.DeliveryStreamType = aws.String(firehose.DeliveryStreamTypeKinesisStreamAsSource)
		input.KinesisStreamSourceConfiguration = expandKinesisStreamSourceConfiguration(v.([]interface{})[0].(map[string]interface{}))
	} else if v, ok := d.GetOk("msk_source_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeliveryStreamType = aws.String(firehose.DeliveryStreamTypeMskasSource)
		input.MSKSourceConfiguration = expandMSKSourceConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	switch d.Get("destination").(string) {
	case destinationTypeElasticsearch:
		if v, ok := d.GetOk("elasticsearch_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ElasticsearchDestinationConfiguration = expandElasticsearchDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	case destinationTypeExtendedS3:
		if v, ok := d.GetOk("extended_s3_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ExtendedS3DestinationConfiguration = expandExtendedS3DestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	case destinationTypeHTTPEndpoint:
		if v, ok := d.GetOk("http_endpoint_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.HttpEndpointDestinationConfiguration = expandHTTPEndpointDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	case destinationTypeOpenSearch:
		if v, ok := d.GetOk("opensearch_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.AmazonopensearchserviceDestinationConfiguration = expandAmazonopensearchserviceDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	case destinationTypeOpenSearchServerless:
		if v, ok := d.GetOk("opensearchserverless_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.AmazonOpenSearchServerlessDestinationConfiguration = expandAmazonOpenSearchServerlessDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	case destinationTypeRedshift:
		if v, ok := d.GetOk("redshift_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.RedshiftDestinationConfiguration = expandRedshiftDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	case destinationTypeSplunk:
		if v, ok := d.GetOk("splunk_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.SplunkDestinationConfiguration = expandSplunkDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	_, err := retryDeliveryStreamOp(ctx, func() (interface{}, error) {
		return conn.CreateDeliveryStreamWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Firehose Delivery Stream (%s): %s", sn, err)
	}

	output, err := waitDeliveryStreamCreated(ctx, conn, sn, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Firehose Delivery Stream (%s) create: %s", sn, err)
	}

	d.SetId(aws.StringValue(output.DeliveryStreamARN))

	if v, ok := d.GetOk("server_side_encryption"); ok && !isDeliveryStreamOptionDisabled(v) {
		input := &firehose.StartDeliveryStreamEncryptionInput{
			DeliveryStreamEncryptionConfigurationInput: expandDeliveryStreamEncryptionConfigurationInput(v.([]interface{})),
			DeliveryStreamName:                         aws.String(sn),
		}

		_, err := conn.StartDeliveryStreamEncryptionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "starting Kinesis Firehose Delivery Stream (%s) encryption: %s", sn, err)
		}

		if _, err := waitDeliveryStreamEncryptionEnabled(ctx, conn, sn, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Firehose Delivery Stream (%s) encryption enable: %s", sn, err)
		}
	}

	return append(diags, resourceDeliveryStreamRead(ctx, d, meta)...)
}

func resourceDeliveryStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FirehoseConn(ctx)

	sn := d.Get("name").(string)
	s, err := FindDeliveryStreamByName(ctx, conn, sn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Firehose Delivery Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Firehose Delivery Stream (%s): %s", sn, err)
	}

	d.Set("arn", s.DeliveryStreamARN)
	if v := s.Source; v != nil {
		if v := v.KinesisStreamSourceDescription; v != nil {
			if err := d.Set("kinesis_source_configuration", flattenKinesisStreamSourceDescription(v)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting kinesis_source_configuration: %s", err)
			}
		}
		if v := v.MSKSourceDescription; v != nil {
			if err := d.Set("msk_source_configuration", []interface{}{flattenMSKSourceDescription(v)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting msk_source_configuration: %s", err)
			}
		}
	}
	d.Set("name", s.DeliveryStreamName)
	d.Set("version_id", s.VersionId)

	sseOptions := map[string]interface{}{
		"enabled":  false,
		"key_type": firehose.KeyTypeAwsOwnedCmk,
	}
	if s.DeliveryStreamEncryptionConfiguration != nil && aws.StringValue(s.DeliveryStreamEncryptionConfiguration.Status) == firehose.DeliveryStreamEncryptionStatusEnabled {
		sseOptions["enabled"] = true

		if v := s.DeliveryStreamEncryptionConfiguration.KeyARN; v != nil {
			sseOptions["key_arn"] = aws.StringValue(v)
		}
		if v := s.DeliveryStreamEncryptionConfiguration.KeyType; v != nil {
			sseOptions["key_type"] = aws.StringValue(v)
		}
	}
	if err := d.Set("server_side_encryption", []map[string]interface{}{sseOptions}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting server_side_encryption: %s", err)
	}

	if len(s.Destinations) > 0 {
		destination := s.Destinations[0]
		switch {
		case destination.ElasticsearchDestinationDescription != nil:
			d.Set("destination", destinationTypeElasticsearch)
			if err := d.Set("elasticsearch_configuration", flattenElasticsearchDestinationDescription(destination.ElasticsearchDestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting elasticsearch_configuration: %s", err)
			}
		case destination.HttpEndpointDestinationDescription != nil:
			d.Set("destination", destinationTypeHTTPEndpoint)
			configuredAccessKey := d.Get("http_endpoint_configuration.0.access_key").(string)
			if err := d.Set("http_endpoint_configuration", flattenHTTPEndpointDestinationDescription(destination.HttpEndpointDestinationDescription, configuredAccessKey)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting http_endpoint_configuration: %s", err)
			}
		case destination.AmazonopensearchserviceDestinationDescription != nil:
			d.Set("destination", destinationTypeOpenSearch)
			if err := d.Set("opensearch_configuration", flattenAmazonopensearchserviceDestinationDescription(destination.AmazonopensearchserviceDestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting opensearch_configuration: %s", err)
			}
		case destination.AmazonOpenSearchServerlessDestinationDescription != nil:
			d.Set("destination", destinationTypeOpenSearchServerless)
			if err := d.Set("opensearchserverless_configuration", flattenAmazonOpenSearchServerlessDestinationDescription(destination.AmazonOpenSearchServerlessDestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting opensearchserverless_configuration: %s", err)
			}
		case destination.RedshiftDestinationDescription != nil:
			d.Set("destination", destinationTypeRedshift)
			configuredPassword := d.Get("redshift_configuration.0.password").(string)
			if err := d.Set("redshift_configuration", flattenRedshiftDestinationDescription(destination.RedshiftDestinationDescription, configuredPassword)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting redshift_configuration: %s", err)
			}
		case destination.SplunkDestinationDescription != nil:
			d.Set("destination", destinationTypeSplunk)
			if err := d.Set("splunk_configuration", flattenSplunkDestinationDescription(destination.SplunkDestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting splunk_configuration: %s", err)
			}
		default:
			d.Set("destination", destinationTypeExtendedS3)
			if err := d.Set("extended_s3_configuration", flattenExtendedS3DestinationDescription(destination.ExtendedS3DestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting extended_s3_configuration: %s", err)
			}
		}
		d.Set("destination_id", destination.DestinationId)
	}

	return diags
}

func resourceDeliveryStreamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FirehoseConn(ctx)

	sn := d.Get("name").(string)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &firehose.UpdateDestinationInput{
			CurrentDeliveryStreamVersionId: aws.String(d.Get("version_id").(string)),
			DeliveryStreamName:             aws.String(sn),
			DestinationId:                  aws.String(d.Get("destination_id").(string)),
		}

		switch d.Get("destination").(string) {
		case destinationTypeElasticsearch:
			if v, ok := d.GetOk("elasticsearch_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ElasticsearchDestinationUpdate = expandElasticsearchDestinationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		case destinationTypeExtendedS3:
			if v, ok := d.GetOk("extended_s3_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ExtendedS3DestinationUpdate = expandExtendedS3DestinationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		case destinationTypeHTTPEndpoint:
			if v, ok := d.GetOk("http_endpoint_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.HttpEndpointDestinationUpdate = expandHTTPEndpointDestinationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		case destinationTypeOpenSearch:
			if v, ok := d.GetOk("opensearch_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.AmazonopensearchserviceDestinationUpdate = expandAmazonopensearchserviceDestinationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		case destinationTypeOpenSearchServerless:
			if v, ok := d.GetOk("opensearchserverless_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.AmazonOpenSearchServerlessDestinationUpdate = expandAmazonOpenSearchServerlessDestinationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		case destinationTypeRedshift:
			if v, ok := d.GetOk("redshift_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.RedshiftDestinationUpdate = expandRedshiftDestinationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		case destinationTypeSplunk:
			if v, ok := d.GetOk("splunk_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.SplunkDestinationUpdate = expandSplunkDestinationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		_, err := retryDeliveryStreamOp(ctx, func() (interface{}, error) {
			return conn.UpdateDestinationWithContext(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Firehose Delivery Stream (%s): %s", sn, err)
		}
	}

	if d.HasChange("server_side_encryption") {
		v := d.Get("server_side_encryption")
		if isDeliveryStreamOptionDisabled(v) {
			input := &firehose.StopDeliveryStreamEncryptionInput{
				DeliveryStreamName: aws.String(sn),
			}

			_, err := conn.StopDeliveryStreamEncryptionWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "stopping Kinesis Firehose Delivery Stream (%s) encryption: %s", sn, err)
			}

			if _, err := waitDeliveryStreamEncryptionDisabled(ctx, conn, sn, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Firehose Delivery Stream (%s) encryption disable: %s", sn, err)
			}
		} else {
			input := &firehose.StartDeliveryStreamEncryptionInput{
				DeliveryStreamEncryptionConfigurationInput: expandDeliveryStreamEncryptionConfigurationInput(v.([]interface{})),
				DeliveryStreamName:                         aws.String(sn),
			}

			_, err := conn.StartDeliveryStreamEncryptionWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "starting Kinesis Firehose Delivery Stream (%s) encryption: %s", sn, err)
			}

			if _, err := waitDeliveryStreamEncryptionEnabled(ctx, conn, sn, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Firehose Delivery Stream (%s) encryption enable: %s", sn, err)
			}
		}
	}

	return append(diags, resourceDeliveryStreamRead(ctx, d, meta)...)
}

func resourceDeliveryStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FirehoseConn(ctx)

	sn := d.Get("name").(string)

	log.Printf("[DEBUG] Deleting Kinesis Firehose Delivery Stream: (%s)", sn)
	_, err := conn.DeleteDeliveryStreamWithContext(ctx, &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(sn),
	})

	if tfawserr.ErrCodeEquals(err, firehose.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Firehose Delivery Stream (%s): %s", sn, err)
	}

	if _, err := waitDeliveryStreamDeleted(ctx, conn, sn, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Firehose Delivery Stream (%s) delete: %s", sn, err)
	}

	return diags
}

func retryDeliveryStreamOp(ctx context.Context, f func() (interface{}, error)) (interface{}, error) {
	return tfresource.RetryWhen(ctx, propagationTimeout,
		f,
		func(err error) (bool, error) {
			// Access was denied when calling Glue. Please ensure that the role specified in the data format conversion configuration has the necessary permissions.
			if tfawserr.ErrMessageContains(err, firehose.ErrCodeInvalidArgumentException, "Access was denied") {
				return true, err
			}
			if tfawserr.ErrMessageContains(err, firehose.ErrCodeInvalidArgumentException, "is not authorized to") {
				return true, err
			}
			if tfawserr.ErrMessageContains(err, firehose.ErrCodeInvalidArgumentException, "Please make sure the role specified in VpcConfiguration has permissions") {
				return true, err
			}
			// InvalidArgumentException: Verify that the IAM role has access to the Elasticsearch domain.
			if tfawserr.ErrMessageContains(err, firehose.ErrCodeInvalidArgumentException, "Verify that the IAM role has access") {
				return true, err
			}
			if tfawserr.ErrMessageContains(err, firehose.ErrCodeInvalidArgumentException, "Firehose is unable to assume role") {
				return true, err
			}
			return false, err
		},
	)
}

func expandKinesisStreamSourceConfiguration(source map[string]interface{}) *firehose.KinesisStreamSourceConfiguration {
	configuration := &firehose.KinesisStreamSourceConfiguration{
		KinesisStreamARN: aws.String(source["kinesis_stream_arn"].(string)),
		RoleARN:          aws.String(source["role_arn"].(string)),
	}

	return configuration
}

func expandS3DestinationConfiguration(tfList []interface{}) *firehose.S3DestinationConfiguration {
	s3 := tfList[0].(map[string]interface{})

	configuration := &firehose.S3DestinationConfiguration{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3["role_arn"].(string)),
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64(int64(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int64(int64(s3["buffering_size"].(int))),
		},
		Prefix:                  expandPrefix(s3),
		CompressionFormat:       aws.String(s3["compression_format"].(string)),
		EncryptionConfiguration: expandEncryptionConfiguration(s3),
	}

	if v, ok := s3["error_output_prefix"].(string); ok && v != "" {
		configuration.ErrorOutputPrefix = aws.String(v)
	}

	if _, ok := s3["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(s3)
	}

	return configuration
}

func expandS3DestinationConfigurationBackup(d map[string]interface{}) *firehose.S3DestinationConfiguration {
	config := d["s3_backup_configuration"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	s3 := config[0].(map[string]interface{})

	configuration := &firehose.S3DestinationConfiguration{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3["role_arn"].(string)),
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64(int64(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int64(int64(s3["buffering_size"].(int))),
		},
		Prefix:                  expandPrefix(s3),
		CompressionFormat:       aws.String(s3["compression_format"].(string)),
		EncryptionConfiguration: expandEncryptionConfiguration(s3),
	}

	if _, ok := s3["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(s3)
	}

	if v, ok := s3["error_output_prefix"].(string); ok && v != "" {
		configuration.ErrorOutputPrefix = aws.String(v)
	}

	return configuration
}

func expandExtendedS3DestinationConfiguration(s3 map[string]interface{}) *firehose.ExtendedS3DestinationConfiguration {
	configuration := &firehose.ExtendedS3DestinationConfiguration{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3["role_arn"].(string)),
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64(int64(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int64(int64(s3["buffering_size"].(int))),
		},
		Prefix:                            expandPrefix(s3),
		CompressionFormat:                 aws.String(s3["compression_format"].(string)),
		DataFormatConversionConfiguration: expandDataFormatConversionConfiguration(s3["data_format_conversion_configuration"].([]interface{})),
		EncryptionConfiguration:           expandEncryptionConfiguration(s3),
	}

	if _, ok := s3["processing_configuration"]; ok {
		configuration.ProcessingConfiguration = expandProcessingConfiguration(s3)
	}

	if _, ok := s3["dynamic_partitioning_configuration"]; ok {
		configuration.DynamicPartitioningConfiguration = expandDynamicPartitioningConfiguration(s3)
	}

	if _, ok := s3["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(s3)
	}

	if v, ok := s3["error_output_prefix"].(string); ok && v != "" {
		configuration.ErrorOutputPrefix = aws.String(v)
	}

	if s3BackupMode, ok := s3["s3_backup_mode"]; ok {
		configuration.S3BackupMode = aws.String(s3BackupMode.(string))
		configuration.S3BackupConfiguration = expandS3DestinationConfigurationBackup(s3)
	}

	return configuration
}

func expandS3DestinationUpdate(tfList []interface{}) *firehose.S3DestinationUpdate {
	s3 := tfList[0].(map[string]interface{})
	configuration := &firehose.S3DestinationUpdate{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3["role_arn"].(string)),
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64((int64)(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int64((int64)(s3["buffering_size"].(int))),
		},
		ErrorOutputPrefix:        aws.String(s3["error_output_prefix"].(string)),
		Prefix:                   expandPrefix(s3),
		CompressionFormat:        aws.String(s3["compression_format"].(string)),
		EncryptionConfiguration:  expandEncryptionConfiguration(s3),
		CloudWatchLoggingOptions: expandCloudWatchLoggingOptions(s3),
	}

	if _, ok := s3["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(s3)
	}

	return configuration
}

func expandS3DestinationUpdateBackup(d map[string]interface{}) *firehose.S3DestinationUpdate {
	config := d["s3_backup_configuration"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	s3 := config[0].(map[string]interface{})

	configuration := &firehose.S3DestinationUpdate{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3["role_arn"].(string)),
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64((int64)(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int64((int64)(s3["buffering_size"].(int))),
		},
		ErrorOutputPrefix:        aws.String(s3["error_output_prefix"].(string)),
		Prefix:                   expandPrefix(s3),
		CompressionFormat:        aws.String(s3["compression_format"].(string)),
		EncryptionConfiguration:  expandEncryptionConfiguration(s3),
		CloudWatchLoggingOptions: expandCloudWatchLoggingOptions(s3),
	}

	if _, ok := s3["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(s3)
	}

	return configuration
}

func expandExtendedS3DestinationUpdate(s3 map[string]interface{}) *firehose.ExtendedS3DestinationUpdate {
	configuration := &firehose.ExtendedS3DestinationUpdate{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3["role_arn"].(string)),
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64((int64)(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int64((int64)(s3["buffering_size"].(int))),
		},
		ErrorOutputPrefix:                 aws.String(s3["error_output_prefix"].(string)),
		Prefix:                            expandPrefix(s3),
		CompressionFormat:                 aws.String(s3["compression_format"].(string)),
		EncryptionConfiguration:           expandEncryptionConfiguration(s3),
		DataFormatConversionConfiguration: expandDataFormatConversionConfiguration(s3["data_format_conversion_configuration"].([]interface{})),
		CloudWatchLoggingOptions:          expandCloudWatchLoggingOptions(s3),
		ProcessingConfiguration:           expandProcessingConfiguration(s3),
	}

	if _, ok := s3["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(s3)
	}

	if _, ok := s3["dynamic_partitioning_configuration"]; ok {
		configuration.DynamicPartitioningConfiguration = expandDynamicPartitioningConfiguration(s3)
	}

	if s3BackupMode, ok := s3["s3_backup_mode"]; ok {
		configuration.S3BackupMode = aws.String(s3BackupMode.(string))
		configuration.S3BackupUpdate = expandS3DestinationUpdateBackup(s3)
	}

	return configuration
}

func expandDataFormatConversionConfiguration(l []interface{}) *firehose.DataFormatConversionConfiguration {
	if len(l) == 0 || l[0] == nil {
		// It is possible to just pass nil here, but this seems to be the
		// canonical form that AWS uses, and is less likely to produce diffs.
		return &firehose.DataFormatConversionConfiguration{
			Enabled: aws.Bool(false),
		}
	}

	m := l[0].(map[string]interface{})

	return &firehose.DataFormatConversionConfiguration{
		Enabled:                   aws.Bool(m["enabled"].(bool)),
		InputFormatConfiguration:  expandInputFormatConfiguration(m["input_format_configuration"].([]interface{})),
		OutputFormatConfiguration: expandOutputFormatConfiguration(m["output_format_configuration"].([]interface{})),
		SchemaConfiguration:       expandSchemaConfiguration(m["schema_configuration"].([]interface{})),
	}
}

func expandInputFormatConfiguration(l []interface{}) *firehose.InputFormatConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &firehose.InputFormatConfiguration{
		Deserializer: expandDeserializer(m["deserializer"].([]interface{})),
	}
}

func expandDeserializer(l []interface{}) *firehose.Deserializer {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &firehose.Deserializer{
		HiveJsonSerDe:  expandHiveJSONSerDe(m["hive_json_ser_de"].([]interface{})),
		OpenXJsonSerDe: expandOpenXJSONSerDe(m["open_x_json_ser_de"].([]interface{})),
	}
}

func expandHiveJSONSerDe(l []interface{}) *firehose.HiveJsonSerDe {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &firehose.HiveJsonSerDe{}
	}

	m := l[0].(map[string]interface{})

	return &firehose.HiveJsonSerDe{
		TimestampFormats: flex.ExpandStringList(m["timestamp_formats"].([]interface{})),
	}
}

func expandOpenXJSONSerDe(l []interface{}) *firehose.OpenXJsonSerDe {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &firehose.OpenXJsonSerDe{}
	}

	m := l[0].(map[string]interface{})

	return &firehose.OpenXJsonSerDe{
		CaseInsensitive:                    aws.Bool(m["case_insensitive"].(bool)),
		ColumnToJsonKeyMappings:            flex.ExpandStringMap(m["column_to_json_key_mappings"].(map[string]interface{})),
		ConvertDotsInJsonKeysToUnderscores: aws.Bool(m["convert_dots_in_json_keys_to_underscores"].(bool)),
	}
}

func expandOutputFormatConfiguration(l []interface{}) *firehose.OutputFormatConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &firehose.OutputFormatConfiguration{
		Serializer: expandSerializer(m["serializer"].([]interface{})),
	}
}

func expandSerializer(l []interface{}) *firehose.Serializer {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &firehose.Serializer{
		OrcSerDe:     expandOrcSerDe(m["orc_ser_de"].([]interface{})),
		ParquetSerDe: expandParquetSerDe(m["parquet_ser_de"].([]interface{})),
	}
}

func expandOrcSerDe(l []interface{}) *firehose.OrcSerDe {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &firehose.OrcSerDe{}
	}

	m := l[0].(map[string]interface{})

	orcSerDe := &firehose.OrcSerDe{
		BlockSizeBytes:                      aws.Int64(int64(m["block_size_bytes"].(int))),
		BloomFilterFalsePositiveProbability: aws.Float64(m["bloom_filter_false_positive_probability"].(float64)),
		Compression:                         aws.String(m["compression"].(string)),
		DictionaryKeyThreshold:              aws.Float64(m["dictionary_key_threshold"].(float64)),
		EnablePadding:                       aws.Bool(m["enable_padding"].(bool)),
		FormatVersion:                       aws.String(m["format_version"].(string)),
		PaddingTolerance:                    aws.Float64(m["padding_tolerance"].(float64)),
		RowIndexStride:                      aws.Int64(int64(m["row_index_stride"].(int))),
		StripeSizeBytes:                     aws.Int64(int64(m["stripe_size_bytes"].(int))),
	}

	if v, ok := m["bloom_filter_columns"].([]interface{}); ok && len(v) > 0 {
		orcSerDe.BloomFilterColumns = flex.ExpandStringList(v)
	}

	return orcSerDe
}

func expandParquetSerDe(l []interface{}) *firehose.ParquetSerDe {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &firehose.ParquetSerDe{}
	}

	m := l[0].(map[string]interface{})

	return &firehose.ParquetSerDe{
		BlockSizeBytes:              aws.Int64(int64(m["block_size_bytes"].(int))),
		Compression:                 aws.String(m["compression"].(string)),
		EnableDictionaryCompression: aws.Bool(m["enable_dictionary_compression"].(bool)),
		MaxPaddingBytes:             aws.Int64(int64(m["max_padding_bytes"].(int))),
		PageSizeBytes:               aws.Int64(int64(m["page_size_bytes"].(int))),
		WriterVersion:               aws.String(m["writer_version"].(string)),
	}
}

func expandSchemaConfiguration(l []interface{}) *firehose.SchemaConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &firehose.SchemaConfiguration{
		DatabaseName: aws.String(m["database_name"].(string)),
		RoleARN:      aws.String(m["role_arn"].(string)),
		TableName:    aws.String(m["table_name"].(string)),
		VersionId:    aws.String(m["version_id"].(string)),
	}

	if v, ok := m["catalog_id"].(string); ok && v != "" {
		config.CatalogId = aws.String(v)
	}
	if v, ok := m["region"].(string); ok && v != "" {
		config.Region = aws.String(v)
	}

	return config
}

func expandDynamicPartitioningConfiguration(s3 map[string]interface{}) *firehose.DynamicPartitioningConfiguration {
	config := s3["dynamic_partitioning_configuration"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	dynamicPartitioningConfig := config[0].(map[string]interface{})
	DynamicPartitioningConfiguration := &firehose.DynamicPartitioningConfiguration{
		Enabled: aws.Bool(dynamicPartitioningConfig["enabled"].(bool)),
	}

	if retryDuration, ok := dynamicPartitioningConfig["retry_duration"]; ok {
		DynamicPartitioningConfiguration.RetryOptions = &firehose.RetryOptions{
			DurationInSeconds: aws.Int64(int64(retryDuration.(int))),
		}
	}

	return DynamicPartitioningConfiguration
}

func expandProcessingConfiguration(s3 map[string]interface{}) *firehose.ProcessingConfiguration {
	config := s3["processing_configuration"].([]interface{})
	if len(config) == 0 || config[0] == nil {
		// It is possible to just pass nil here, but this seems to be the
		// canonical form that AWS uses, and is less likely to produce diffs.
		return &firehose.ProcessingConfiguration{
			Enabled:    aws.Bool(false),
			Processors: []*firehose.Processor{},
		}
	}

	processingConfiguration := config[0].(map[string]interface{})

	return &firehose.ProcessingConfiguration{
		Enabled:    aws.Bool(processingConfiguration["enabled"].(bool)),
		Processors: expandProcessors(processingConfiguration["processors"].([]interface{})),
	}
}

func expandProcessors(processingConfigurationProcessors []interface{}) []*firehose.Processor {
	processors := []*firehose.Processor{}

	for _, processor := range processingConfigurationProcessors {
		extractedProcessor := expandProcessor(processor.(map[string]interface{}))
		if extractedProcessor != nil {
			processors = append(processors, extractedProcessor)
		}
	}

	return processors
}

func expandProcessor(processingConfigurationProcessor map[string]interface{}) *firehose.Processor {
	var processor *firehose.Processor
	processorType := processingConfigurationProcessor["type"].(string)
	if processorType != "" {
		processor = &firehose.Processor{
			Type:       aws.String(processorType),
			Parameters: expandProcessorParameters(processingConfigurationProcessor["parameters"].([]interface{})),
		}
	}
	return processor
}

func expandProcessorParameters(processorParameters []interface{}) []*firehose.ProcessorParameter {
	parameters := []*firehose.ProcessorParameter{}

	for _, attr := range processorParameters {
		parameters = append(parameters, expandProcessorParameter(attr.(map[string]interface{})))
	}

	return parameters
}

func expandProcessorParameter(processorParameter map[string]interface{}) *firehose.ProcessorParameter {
	parameter := &firehose.ProcessorParameter{
		ParameterName:  aws.String(processorParameter["parameter_name"].(string)),
		ParameterValue: aws.String(processorParameter["parameter_value"].(string)),
	}

	return parameter
}

func expandEncryptionConfiguration(s3 map[string]interface{}) *firehose.EncryptionConfiguration {
	if key, ok := s3["kms_key_arn"]; ok && len(key.(string)) > 0 {
		return &firehose.EncryptionConfiguration{
			KMSEncryptionConfig: &firehose.KMSEncryptionConfig{
				AWSKMSKeyARN: aws.String(key.(string)),
			},
		}
	}

	return &firehose.EncryptionConfiguration{
		NoEncryptionConfig: aws.String(firehose.NoEncryptionConfigNoEncryption),
	}
}

func expandCloudWatchLoggingOptions(s3 map[string]interface{}) *firehose.CloudWatchLoggingOptions {
	config := s3["cloudwatch_logging_options"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	loggingConfig := config[0].(map[string]interface{})
	loggingOptions := &firehose.CloudWatchLoggingOptions{
		Enabled: aws.Bool(loggingConfig["enabled"].(bool)),
	}

	if v, ok := loggingConfig["log_group_name"]; ok {
		loggingOptions.LogGroupName = aws.String(v.(string))
	}

	if v, ok := loggingConfig["log_stream_name"]; ok {
		loggingOptions.LogStreamName = aws.String(v.(string))
	}

	return loggingOptions
}

func expandVPCConfiguration(es map[string]interface{}) *firehose.VpcConfiguration {
	config := es["vpc_config"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	vpcConfig := config[0].(map[string]interface{})

	return &firehose.VpcConfiguration{
		RoleARN:          aws.String(vpcConfig["role_arn"].(string)),
		SubnetIds:        flex.ExpandStringSet(vpcConfig["subnet_ids"].(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringSet(vpcConfig["security_group_ids"].(*schema.Set)),
	}
}

func expandPrefix(s3 map[string]interface{}) *string {
	if v, ok := s3["prefix"]; ok {
		return aws.String(v.(string))
	}

	return nil
}

func expandRedshiftDestinationConfiguration(redshift map[string]interface{}) *firehose.RedshiftDestinationConfiguration {
	configuration := &firehose.RedshiftDestinationConfiguration{
		ClusterJDBCURL:  aws.String(redshift["cluster_jdbcurl"].(string)),
		RetryOptions:    expandRedshiftRetryOptions(redshift),
		Password:        aws.String(redshift["password"].(string)),
		Username:        aws.String(redshift["username"].(string)),
		RoleARN:         aws.String(redshift["role_arn"].(string)),
		CopyCommand:     expandCopyCommand(redshift),
		S3Configuration: expandS3DestinationConfiguration(redshift["s3_configuration"].([]interface{})),
	}

	if _, ok := redshift["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(redshift)
	}
	if _, ok := redshift["processing_configuration"]; ok {
		configuration.ProcessingConfiguration = expandProcessingConfiguration(redshift)
	}
	if s3BackupMode, ok := redshift["s3_backup_mode"]; ok {
		configuration.S3BackupMode = aws.String(s3BackupMode.(string))
		configuration.S3BackupConfiguration = expandS3DestinationConfigurationBackup(redshift)
	}

	return configuration
}

func expandRedshiftDestinationUpdate(redshift map[string]interface{}) *firehose.RedshiftDestinationUpdate {
	configuration := &firehose.RedshiftDestinationUpdate{
		ClusterJDBCURL: aws.String(redshift["cluster_jdbcurl"].(string)),
		RetryOptions:   expandRedshiftRetryOptions(redshift),
		Password:       aws.String(redshift["password"].(string)),
		Username:       aws.String(redshift["username"].(string)),
		RoleARN:        aws.String(redshift["role_arn"].(string)),
		CopyCommand:    expandCopyCommand(redshift),
	}

	s3Config := expandS3DestinationUpdate(redshift["s3_configuration"].([]interface{}))
	// Redshift does not currently support ErrorOutputPrefix,
	// which is set to the empty string within "updateS3Config",
	// thus we must remove it here to avoid an InvalidArgumentException.
	s3Config.ErrorOutputPrefix = nil
	configuration.S3Update = s3Config

	if _, ok := redshift["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(redshift)
	}
	if _, ok := redshift["processing_configuration"]; ok {
		configuration.ProcessingConfiguration = expandProcessingConfiguration(redshift)
	}
	if s3BackupMode, ok := redshift["s3_backup_mode"]; ok {
		configuration.S3BackupMode = aws.String(s3BackupMode.(string))
		configuration.S3BackupUpdate = expandS3DestinationUpdateBackup(redshift)
		if configuration.S3BackupUpdate != nil {
			// Redshift does not currently support ErrorOutputPrefix,
			// which is set to the empty string within "updateS3BackupConfig",
			// thus we must remove it here to avoid an InvalidArgumentException.
			configuration.S3BackupUpdate.ErrorOutputPrefix = nil
		}
	}

	return configuration
}

func expandElasticsearchDestinationConfiguration(es map[string]interface{}) *firehose.ElasticsearchDestinationConfiguration {
	config := &firehose.ElasticsearchDestinationConfiguration{
		BufferingHints:  expandElasticsearchBufferingHints(es),
		IndexName:       aws.String(es["index_name"].(string)),
		RetryOptions:    expandElasticsearchRetryOptions(es),
		RoleARN:         aws.String(es["role_arn"].(string)),
		TypeName:        aws.String(es["type_name"].(string)),
		S3Configuration: expandS3DestinationConfiguration(es["s3_configuration"].([]interface{})),
	}

	if v, ok := es["domain_arn"]; ok && v.(string) != "" {
		config.DomainARN = aws.String(v.(string))
	}

	if v, ok := es["cluster_endpoint"]; ok && v.(string) != "" {
		config.ClusterEndpoint = aws.String(v.(string))
	}

	if _, ok := es["cloudwatch_logging_options"]; ok {
		config.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(es)
	}

	if _, ok := es["processing_configuration"]; ok {
		config.ProcessingConfiguration = expandProcessingConfiguration(es)
	}

	if indexRotationPeriod, ok := es["index_rotation_period"]; ok {
		config.IndexRotationPeriod = aws.String(indexRotationPeriod.(string))
	}
	if s3BackupMode, ok := es["s3_backup_mode"]; ok {
		config.S3BackupMode = aws.String(s3BackupMode.(string))
	}

	if _, ok := es["vpc_config"]; ok {
		config.VpcConfiguration = expandVPCConfiguration(es)
	}

	return config
}

func expandElasticsearchDestinationUpdate(es map[string]interface{}) *firehose.ElasticsearchDestinationUpdate {
	update := &firehose.ElasticsearchDestinationUpdate{
		BufferingHints: expandElasticsearchBufferingHints(es),
		IndexName:      aws.String(es["index_name"].(string)),
		RetryOptions:   expandElasticsearchRetryOptions(es),
		RoleARN:        aws.String(es["role_arn"].(string)),
		TypeName:       aws.String(es["type_name"].(string)),
		S3Update:       expandS3DestinationUpdate(es["s3_configuration"].([]interface{})),
	}

	if v, ok := es["domain_arn"]; ok && v.(string) != "" {
		update.DomainARN = aws.String(v.(string))
	}

	if v, ok := es["cluster_endpoint"]; ok && v.(string) != "" {
		update.ClusterEndpoint = aws.String(v.(string))
	}

	if _, ok := es["cloudwatch_logging_options"]; ok {
		update.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(es)
	}

	if _, ok := es["processing_configuration"]; ok {
		update.ProcessingConfiguration = expandProcessingConfiguration(es)
	}

	if indexRotationPeriod, ok := es["index_rotation_period"]; ok {
		update.IndexRotationPeriod = aws.String(indexRotationPeriod.(string))
	}

	return update
}

func expandAmazonopensearchserviceDestinationConfiguration(es map[string]interface{}) *firehose.AmazonopensearchserviceDestinationConfiguration {
	config := &firehose.AmazonopensearchserviceDestinationConfiguration{
		BufferingHints:  expandAmazonopensearchserviceBufferingHints(es),
		IndexName:       aws.String(es["index_name"].(string)),
		RetryOptions:    expandAmazonopensearchserviceRetryOptions(es),
		RoleARN:         aws.String(es["role_arn"].(string)),
		TypeName:        aws.String(es["type_name"].(string)),
		S3Configuration: expandS3DestinationConfiguration(es["s3_configuration"].([]interface{})),
	}

	if v, ok := es["domain_arn"]; ok && v.(string) != "" {
		config.DomainARN = aws.String(v.(string))
	}

	if v, ok := es["cluster_endpoint"]; ok && v.(string) != "" {
		config.ClusterEndpoint = aws.String(v.(string))
	}

	if _, ok := es["cloudwatch_logging_options"]; ok {
		config.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(es)
	}

	if _, ok := es["processing_configuration"]; ok {
		config.ProcessingConfiguration = expandProcessingConfiguration(es)
	}

	if indexRotationPeriod, ok := es["index_rotation_period"]; ok {
		config.IndexRotationPeriod = aws.String(indexRotationPeriod.(string))
	}
	if s3BackupMode, ok := es["s3_backup_mode"]; ok {
		config.S3BackupMode = aws.String(s3BackupMode.(string))
	}

	if _, ok := es["vpc_config"]; ok {
		config.VpcConfiguration = expandVPCConfiguration(es)
	}

	return config
}

func expandAmazonopensearchserviceDestinationUpdate(es map[string]interface{}) *firehose.AmazonopensearchserviceDestinationUpdate {
	update := &firehose.AmazonopensearchserviceDestinationUpdate{
		BufferingHints: expandAmazonopensearchserviceBufferingHints(es),
		IndexName:      aws.String(es["index_name"].(string)),
		RetryOptions:   expandAmazonopensearchserviceRetryOptions(es),
		RoleARN:        aws.String(es["role_arn"].(string)),
		TypeName:       aws.String(es["type_name"].(string)),
		S3Update:       expandS3DestinationUpdate(es["s3_configuration"].([]interface{})),
	}

	if v, ok := es["domain_arn"]; ok && v.(string) != "" {
		update.DomainARN = aws.String(v.(string))
	}

	if v, ok := es["cluster_endpoint"]; ok && v.(string) != "" {
		update.ClusterEndpoint = aws.String(v.(string))
	}

	if _, ok := es["cloudwatch_logging_options"]; ok {
		update.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(es)
	}

	if _, ok := es["processing_configuration"]; ok {
		update.ProcessingConfiguration = expandProcessingConfiguration(es)
	}

	if indexRotationPeriod, ok := es["index_rotation_period"]; ok {
		update.IndexRotationPeriod = aws.String(indexRotationPeriod.(string))
	}

	return update
}

func expandAmazonOpenSearchServerlessDestinationConfiguration(es map[string]interface{}) *firehose.AmazonOpenSearchServerlessDestinationConfiguration {
	config := &firehose.AmazonOpenSearchServerlessDestinationConfiguration{
		BufferingHints:  expandAmazonOpenSearchServerlessBufferingHints(es),
		IndexName:       aws.String(es["index_name"].(string)),
		RetryOptions:    expandAmazonOpenSearchServerlessRetryOptions(es),
		RoleARN:         aws.String(es["role_arn"].(string)),
		S3Configuration: expandS3DestinationConfiguration(es["s3_configuration"].([]interface{})),
	}

	if v, ok := es["collection_endpoint"]; ok && v.(string) != "" {
		config.CollectionEndpoint = aws.String(v.(string))
	}

	if _, ok := es["cloudwatch_logging_options"]; ok {
		config.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(es)
	}

	if _, ok := es["processing_configuration"]; ok {
		config.ProcessingConfiguration = expandProcessingConfiguration(es)
	}

	if s3BackupMode, ok := es["s3_backup_mode"]; ok {
		config.S3BackupMode = aws.String(s3BackupMode.(string))
	}

	if _, ok := es["vpc_config"]; ok {
		config.VpcConfiguration = expandVPCConfiguration(es)
	}

	return config
}

func expandAmazonOpenSearchServerlessDestinationUpdate(es map[string]interface{}) *firehose.AmazonOpenSearchServerlessDestinationUpdate {
	update := &firehose.AmazonOpenSearchServerlessDestinationUpdate{
		BufferingHints: expandAmazonOpenSearchServerlessBufferingHints(es),
		IndexName:      aws.String(es["index_name"].(string)),
		RetryOptions:   expandAmazonOpenSearchServerlessRetryOptions(es),
		RoleARN:        aws.String(es["role_arn"].(string)),
		S3Update:       expandS3DestinationUpdate(es["s3_configuration"].([]interface{})),
	}
	if v, ok := es["collection_endpoint"]; ok && v.(string) != "" {
		update.CollectionEndpoint = aws.String(v.(string))
	}

	if _, ok := es["cloudwatch_logging_options"]; ok {
		update.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(es)
	}

	if _, ok := es["processing_configuration"]; ok {
		update.ProcessingConfiguration = expandProcessingConfiguration(es)
	}

	return update
}

func expandSplunkDestinationConfiguration(splunk map[string]interface{}) *firehose.SplunkDestinationConfiguration {
	configuration := &firehose.SplunkDestinationConfiguration{
		HECToken:                          aws.String(splunk["hec_token"].(string)),
		HECEndpointType:                   aws.String(splunk["hec_endpoint_type"].(string)),
		HECEndpoint:                       aws.String(splunk["hec_endpoint"].(string)),
		HECAcknowledgmentTimeoutInSeconds: aws.Int64(int64(splunk["hec_acknowledgment_timeout"].(int))),
		RetryOptions:                      expandSplunkRetryOptions(splunk),
		S3Configuration:                   expandS3DestinationConfiguration(splunk["s3_configuration"].([]interface{})),
	}

	if _, ok := splunk["processing_configuration"]; ok {
		configuration.ProcessingConfiguration = expandProcessingConfiguration(splunk)
	}

	if _, ok := splunk["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(splunk)
	}
	if s3BackupMode, ok := splunk["s3_backup_mode"]; ok {
		configuration.S3BackupMode = aws.String(s3BackupMode.(string))
	}

	return configuration
}

func expandSplunkDestinationUpdate(splunk map[string]interface{}) *firehose.SplunkDestinationUpdate {
	configuration := &firehose.SplunkDestinationUpdate{
		HECToken:                          aws.String(splunk["hec_token"].(string)),
		HECEndpointType:                   aws.String(splunk["hec_endpoint_type"].(string)),
		HECEndpoint:                       aws.String(splunk["hec_endpoint"].(string)),
		HECAcknowledgmentTimeoutInSeconds: aws.Int64(int64(splunk["hec_acknowledgment_timeout"].(int))),
		RetryOptions:                      expandSplunkRetryOptions(splunk),
		S3Update:                          expandS3DestinationUpdate(splunk["s3_configuration"].([]interface{})),
	}

	if _, ok := splunk["processing_configuration"]; ok {
		configuration.ProcessingConfiguration = expandProcessingConfiguration(splunk)
	}

	if _, ok := splunk["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(splunk)
	}
	if s3BackupMode, ok := splunk["s3_backup_mode"]; ok {
		configuration.S3BackupMode = aws.String(s3BackupMode.(string))
	}

	return configuration
}

func expandHTTPEndpointDestinationConfiguration(HttpEndpoint map[string]interface{}) *firehose.HttpEndpointDestinationConfiguration {
	configuration := &firehose.HttpEndpointDestinationConfiguration{
		RetryOptions:    expandHTTPEndpointRetryOptions(HttpEndpoint),
		RoleARN:         aws.String(HttpEndpoint["role_arn"].(string)),
		S3Configuration: expandS3DestinationConfiguration(HttpEndpoint["s3_configuration"].([]interface{})),
	}

	configuration.EndpointConfiguration = expandHTTPEndpointConfiguration(HttpEndpoint)

	bufferingHints := &firehose.HttpEndpointBufferingHints{}

	if bufferingInterval, ok := HttpEndpoint["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int64(int64(bufferingInterval))
	}
	if bufferingSize, ok := HttpEndpoint["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int64(int64(bufferingSize))
	}
	configuration.BufferingHints = bufferingHints

	if _, ok := HttpEndpoint["processing_configuration"]; ok {
		configuration.ProcessingConfiguration = expandProcessingConfiguration(HttpEndpoint)
	}

	if _, ok := HttpEndpoint["request_configuration"]; ok {
		configuration.RequestConfiguration = expandHTTPEndpointRequestConfiguration(HttpEndpoint)
	}

	if _, ok := HttpEndpoint["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(HttpEndpoint)
	}
	if s3BackupMode, ok := HttpEndpoint["s3_backup_mode"]; ok {
		configuration.S3BackupMode = aws.String(s3BackupMode.(string))
	}

	return configuration
}

func expandHTTPEndpointDestinationUpdate(HttpEndpoint map[string]interface{}) *firehose.HttpEndpointDestinationUpdate {
	configuration := &firehose.HttpEndpointDestinationUpdate{
		RetryOptions: expandHTTPEndpointRetryOptions(HttpEndpoint),
		RoleARN:      aws.String(HttpEndpoint["role_arn"].(string)),
		S3Update:     expandS3DestinationUpdate(HttpEndpoint["s3_configuration"].([]interface{})),
	}

	configuration.EndpointConfiguration = expandHTTPEndpointConfiguration(HttpEndpoint)

	bufferingHints := &firehose.HttpEndpointBufferingHints{}

	if bufferingInterval, ok := HttpEndpoint["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int64(int64(bufferingInterval))
	}
	if bufferingSize, ok := HttpEndpoint["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int64(int64(bufferingSize))
	}
	configuration.BufferingHints = bufferingHints

	if _, ok := HttpEndpoint["processing_configuration"]; ok {
		configuration.ProcessingConfiguration = expandProcessingConfiguration(HttpEndpoint)
	}

	if _, ok := HttpEndpoint["request_configuration"]; ok {
		configuration.RequestConfiguration = expandHTTPEndpointRequestConfiguration(HttpEndpoint)
	}

	if _, ok := HttpEndpoint["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(HttpEndpoint)
	}

	if s3BackupMode, ok := HttpEndpoint["s3_backup_mode"]; ok {
		configuration.S3BackupMode = aws.String(s3BackupMode.(string))
	}

	return configuration
}

func expandHTTPEndpointCommonAttributes(ca []interface{}) []*firehose.HttpEndpointCommonAttribute {
	CommonAttributes := make([]*firehose.HttpEndpointCommonAttribute, 0, len(ca))

	for _, raw := range ca {
		data := raw.(map[string]interface{})

		a := &firehose.HttpEndpointCommonAttribute{
			AttributeName:  aws.String(data["name"].(string)),
			AttributeValue: aws.String(data["value"].(string)),
		}
		CommonAttributes = append(CommonAttributes, a)
	}

	return CommonAttributes
}

func expandHTTPEndpointRequestConfiguration(rc map[string]interface{}) *firehose.HttpEndpointRequestConfiguration {
	config := rc["request_configuration"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	requestConfig := config[0].(map[string]interface{})
	RequestConfiguration := &firehose.HttpEndpointRequestConfiguration{}

	if contentEncoding, ok := requestConfig["content_encoding"]; ok {
		RequestConfiguration.ContentEncoding = aws.String(contentEncoding.(string))
	}

	if commonAttributes, ok := requestConfig["common_attributes"]; ok {
		RequestConfiguration.CommonAttributes = expandHTTPEndpointCommonAttributes(commonAttributes.([]interface{}))
	}

	return RequestConfiguration
}

func expandHTTPEndpointConfiguration(ep map[string]interface{}) *firehose.HttpEndpointConfiguration {
	endpointConfiguration := &firehose.HttpEndpointConfiguration{
		Url: aws.String(ep["url"].(string)),
	}

	if Name, ok := ep["name"]; ok {
		endpointConfiguration.Name = aws.String(Name.(string))
	}

	if AccessKey, ok := ep["access_key"]; ok {
		endpointConfiguration.AccessKey = aws.String(AccessKey.(string))
	}

	return endpointConfiguration
}

func expandElasticsearchBufferingHints(es map[string]interface{}) *firehose.ElasticsearchBufferingHints {
	bufferingHints := &firehose.ElasticsearchBufferingHints{}

	if bufferingInterval, ok := es["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int64(int64(bufferingInterval))
	}
	if bufferingSize, ok := es["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int64(int64(bufferingSize))
	}

	return bufferingHints
}

func expandAmazonopensearchserviceBufferingHints(es map[string]interface{}) *firehose.AmazonopensearchserviceBufferingHints {
	bufferingHints := &firehose.AmazonopensearchserviceBufferingHints{}

	if bufferingInterval, ok := es["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int64(int64(bufferingInterval))
	}
	if bufferingSize, ok := es["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int64(int64(bufferingSize))
	}

	return bufferingHints
}

func expandAmazonOpenSearchServerlessBufferingHints(es map[string]interface{}) *firehose.AmazonOpenSearchServerlessBufferingHints {
	bufferingHints := &firehose.AmazonOpenSearchServerlessBufferingHints{}

	if bufferingInterval, ok := es["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int64(int64(bufferingInterval))
	}
	if bufferingSize, ok := es["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int64(int64(bufferingSize))
	}

	return bufferingHints
}

func expandElasticsearchRetryOptions(es map[string]interface{}) *firehose.ElasticsearchRetryOptions {
	retryOptions := &firehose.ElasticsearchRetryOptions{}

	if retryDuration, ok := es["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int64(int64(retryDuration))
	}

	return retryOptions
}

func expandAmazonopensearchserviceRetryOptions(es map[string]interface{}) *firehose.AmazonopensearchserviceRetryOptions {
	retryOptions := &firehose.AmazonopensearchserviceRetryOptions{}

	if retryDuration, ok := es["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int64(int64(retryDuration))
	}

	return retryOptions
}

func expandAmazonOpenSearchServerlessRetryOptions(es map[string]interface{}) *firehose.AmazonOpenSearchServerlessRetryOptions {
	retryOptions := &firehose.AmazonOpenSearchServerlessRetryOptions{}

	if retryDuration, ok := es["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int64(int64(retryDuration))
	}

	return retryOptions
}

func expandHTTPEndpointRetryOptions(tfMap map[string]interface{}) *firehose.HttpEndpointRetryOptions {
	retryOptions := &firehose.HttpEndpointRetryOptions{}

	if retryDuration, ok := tfMap["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int64(int64(retryDuration))
	}

	return retryOptions
}

func expandRedshiftRetryOptions(redshift map[string]interface{}) *firehose.RedshiftRetryOptions {
	retryOptions := &firehose.RedshiftRetryOptions{}

	if retryDuration, ok := redshift["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int64(int64(retryDuration))
	}

	return retryOptions
}

func expandSplunkRetryOptions(splunk map[string]interface{}) *firehose.SplunkRetryOptions {
	retryOptions := &firehose.SplunkRetryOptions{}

	if retryDuration, ok := splunk["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int64(int64(retryDuration))
	}

	return retryOptions
}

func expandCopyCommand(redshift map[string]interface{}) *firehose.CopyCommand {
	cmd := &firehose.CopyCommand{
		DataTableName: aws.String(redshift["data_table_name"].(string)),
	}
	if copyOptions, ok := redshift["copy_options"]; ok {
		cmd.CopyOptions = aws.String(copyOptions.(string))
	}
	if columns, ok := redshift["data_table_columns"]; ok {
		cmd.DataTableColumns = aws.String(columns.(string))
	}

	return cmd
}

func expandDeliveryStreamEncryptionConfigurationInput(tfList []interface{}) *firehose.DeliveryStreamEncryptionConfigurationInput {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &firehose.DeliveryStreamEncryptionConfigurationInput{}

	if v, ok := tfMap["key_arn"].(string); ok && v != "" {
		apiObject.KeyARN = aws.String(v)
	}

	if v, ok := tfMap["key_type"].(string); ok && v != "" {
		apiObject.KeyType = aws.String(v)
	}

	return apiObject
}

func expandMSKSourceConfiguration(tfMap map[string]interface{}) *firehose.MSKSourceConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &firehose.MSKSourceConfiguration{}

	if v, ok := tfMap["authentication_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AuthenticationConfiguration = expandAuthenticationConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["msk_cluster_arn"].(string); ok && v != "" {
		apiObject.MSKClusterARN = aws.String(v)
	}

	if v, ok := tfMap["topic_name"].(string); ok && v != "" {
		apiObject.TopicName = aws.String(v)
	}

	return apiObject
}

func expandAuthenticationConfiguration(tfMap map[string]interface{}) *firehose.AuthenticationConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &firehose.AuthenticationConfiguration{}

	if v, ok := tfMap["connectivity"].(string); ok && v != "" {
		apiObject.Connectivity = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleARN = aws.String(v)
	}

	return apiObject
}

func flattenMSKSourceDescription(apiObject *firehose.MSKSourceDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AuthenticationConfiguration; v != nil {
		tfMap["authentication_configuration"] = []interface{}{flattenAuthenticationConfiguration(v)}
	}

	if v := apiObject.MSKClusterARN; v != nil {
		tfMap["msk_cluster_arn"] = aws.StringValue(v)
	}

	if v := apiObject.TopicName; v != nil {
		tfMap["topic_name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenAuthenticationConfiguration(apiObject *firehose.AuthenticationConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Connectivity; v != nil {
		tfMap["connectivity"] = aws.StringValue(v)
	}

	if v := apiObject.RoleARN; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenCloudWatchLoggingOptions(clo *firehose.CloudWatchLoggingOptions) []interface{} {
	if clo == nil {
		return []interface{}{}
	}

	cloudwatchLoggingOptions := map[string]interface{}{
		"enabled": aws.BoolValue(clo.Enabled),
	}
	if aws.BoolValue(clo.Enabled) {
		cloudwatchLoggingOptions["log_group_name"] = aws.StringValue(clo.LogGroupName)
		cloudwatchLoggingOptions["log_stream_name"] = aws.StringValue(clo.LogStreamName)
	}
	return []interface{}{cloudwatchLoggingOptions}
}

func flattenElasticsearchDestinationDescription(description *firehose.ElasticsearchDestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"role_arn":                   aws.StringValue(description.RoleARN),
		"type_name":                  aws.StringValue(description.TypeName),
		"index_name":                 aws.StringValue(description.IndexName),
		"s3_backup_mode":             aws.StringValue(description.S3BackupMode),
		"s3_configuration":           flattenS3DestinationDescription(description.S3DestinationDescription),
		"index_rotation_period":      aws.StringValue(description.IndexRotationPeriod),
		"vpc_config":                 flattenVPCConfigurationDescription(description.VpcConfigurationDescription),
		"processing_configuration":   flattenProcessingConfiguration(description.ProcessingConfiguration, aws.StringValue(description.RoleARN)),
	}

	if description.DomainARN != nil {
		m["domain_arn"] = aws.StringValue(description.DomainARN)
	}

	if description.ClusterEndpoint != nil {
		m["cluster_endpoint"] = aws.StringValue(description.ClusterEndpoint)
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.Int64Value(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.Int64Value(description.BufferingHints.SizeInMBs))
	}

	if description.RetryOptions != nil {
		m["retry_duration"] = int(aws.Int64Value(description.RetryOptions.DurationInSeconds))
	}

	return []map[string]interface{}{m}
}

func flattenAmazonopensearchserviceDestinationDescription(description *firehose.AmazonopensearchserviceDestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"role_arn":                   aws.StringValue(description.RoleARN),
		"type_name":                  aws.StringValue(description.TypeName),
		"index_name":                 aws.StringValue(description.IndexName),
		"s3_backup_mode":             aws.StringValue(description.S3BackupMode),
		"s3_configuration":           flattenS3DestinationDescription(description.S3DestinationDescription),
		"index_rotation_period":      aws.StringValue(description.IndexRotationPeriod),
		"vpc_config":                 flattenVPCConfigurationDescription(description.VpcConfigurationDescription),
		"processing_configuration":   flattenProcessingConfiguration(description.ProcessingConfiguration, aws.StringValue(description.RoleARN)),
	}

	if description.DomainARN != nil {
		m["domain_arn"] = aws.StringValue(description.DomainARN)
	}

	if description.ClusterEndpoint != nil {
		m["cluster_endpoint"] = aws.StringValue(description.ClusterEndpoint)
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.Int64Value(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.Int64Value(description.BufferingHints.SizeInMBs))
	}

	if description.RetryOptions != nil {
		m["retry_duration"] = int(aws.Int64Value(description.RetryOptions.DurationInSeconds))
	}

	return []map[string]interface{}{m}
}

func flattenAmazonOpenSearchServerlessDestinationDescription(description *firehose.AmazonOpenSearchServerlessDestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"role_arn":                   aws.StringValue(description.RoleARN),
		"index_name":                 aws.StringValue(description.IndexName),
		"s3_backup_mode":             aws.StringValue(description.S3BackupMode),
		"s3_configuration":           flattenS3DestinationDescription(description.S3DestinationDescription),
		"vpc_config":                 flattenVPCConfigurationDescription(description.VpcConfigurationDescription),
		"processing_configuration":   flattenProcessingConfiguration(description.ProcessingConfiguration, aws.StringValue(description.RoleARN)),
	}

	if description.CollectionEndpoint != nil {
		m["collection_endpoint"] = aws.StringValue(description.CollectionEndpoint)
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.Int64Value(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.Int64Value(description.BufferingHints.SizeInMBs))
	}

	if description.RetryOptions != nil {
		m["retry_duration"] = int(aws.Int64Value(description.RetryOptions.DurationInSeconds))
	}

	return []map[string]interface{}{m}
}

func flattenVPCConfigurationDescription(description *firehose.VpcConfigurationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"vpc_id":             aws.StringValue(description.VpcId),
		"subnet_ids":         flex.FlattenStringSet(description.SubnetIds),
		"security_group_ids": flex.FlattenStringSet(description.SecurityGroupIds),
		"role_arn":           aws.StringValue(description.RoleARN),
	}

	return []map[string]interface{}{m}
}

func flattenExtendedS3DestinationDescription(description *firehose.ExtendedS3DestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"bucket_arn":                           aws.StringValue(description.BucketARN),
		"cloudwatch_logging_options":           flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"compression_format":                   aws.StringValue(description.CompressionFormat),
		"data_format_conversion_configuration": flattenDataFormatConversionConfiguration(description.DataFormatConversionConfiguration),
		"error_output_prefix":                  aws.StringValue(description.ErrorOutputPrefix),
		"prefix":                               aws.StringValue(description.Prefix),
		"processing_configuration":             flattenProcessingConfiguration(description.ProcessingConfiguration, aws.StringValue(description.RoleARN)),
		"dynamic_partitioning_configuration":   flattenDynamicPartitioningConfiguration(description.DynamicPartitioningConfiguration),
		"role_arn":                             aws.StringValue(description.RoleARN),
		"s3_backup_configuration":              flattenS3DestinationDescription(description.S3BackupDescription),
		"s3_backup_mode":                       aws.StringValue(description.S3BackupMode),
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.Int64Value(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.Int64Value(description.BufferingHints.SizeInMBs))
	}

	if description.EncryptionConfiguration != nil && description.EncryptionConfiguration.KMSEncryptionConfig != nil {
		m["kms_key_arn"] = aws.StringValue(description.EncryptionConfiguration.KMSEncryptionConfig.AWSKMSKeyARN)
	}

	return []map[string]interface{}{m}
}

func flattenRedshiftDestinationDescription(description *firehose.RedshiftDestinationDescription, configuredPassword string) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"cluster_jdbcurl":            aws.StringValue(description.ClusterJDBCURL),
		"password":                   configuredPassword,
		"processing_configuration":   flattenProcessingConfiguration(description.ProcessingConfiguration, aws.StringValue(description.RoleARN)),
		"role_arn":                   aws.StringValue(description.RoleARN),
		"s3_backup_configuration":    flattenS3DestinationDescription(description.S3BackupDescription),
		"s3_backup_mode":             aws.StringValue(description.S3BackupMode),
		"s3_configuration":           flattenS3DestinationDescription(description.S3DestinationDescription),
		"username":                   aws.StringValue(description.Username),
	}

	if description.CopyCommand != nil {
		m["copy_options"] = aws.StringValue(description.CopyCommand.CopyOptions)
		m["data_table_columns"] = aws.StringValue(description.CopyCommand.DataTableColumns)
		m["data_table_name"] = aws.StringValue(description.CopyCommand.DataTableName)
	}

	if description.RetryOptions != nil {
		m["retry_duration"] = int(aws.Int64Value(description.RetryOptions.DurationInSeconds))
	}

	return []map[string]interface{}{m}
}

func flattenSplunkDestinationDescription(description *firehose.SplunkDestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}
	m := map[string]interface{}{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"hec_acknowledgment_timeout": int(aws.Int64Value(description.HECAcknowledgmentTimeoutInSeconds)),
		"hec_endpoint_type":          aws.StringValue(description.HECEndpointType),
		"hec_endpoint":               aws.StringValue(description.HECEndpoint),
		"hec_token":                  aws.StringValue(description.HECToken),
		"processing_configuration":   flattenProcessingConfiguration(description.ProcessingConfiguration, ""),
		"s3_backup_mode":             aws.StringValue(description.S3BackupMode),
		"s3_configuration":           flattenS3DestinationDescription(description.S3DestinationDescription),
	}

	if description.RetryOptions != nil {
		m["retry_duration"] = int(aws.Int64Value(description.RetryOptions.DurationInSeconds))
	}

	return []map[string]interface{}{m}
}

func flattenS3DestinationDescription(description *firehose.S3DestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"bucket_arn":                 aws.StringValue(description.BucketARN),
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"compression_format":         aws.StringValue(description.CompressionFormat),
		"error_output_prefix":        aws.StringValue(description.ErrorOutputPrefix),
		"prefix":                     aws.StringValue(description.Prefix),
		"role_arn":                   aws.StringValue(description.RoleARN),
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.Int64Value(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.Int64Value(description.BufferingHints.SizeInMBs))
	}

	if description.EncryptionConfiguration != nil && description.EncryptionConfiguration.KMSEncryptionConfig != nil {
		m["kms_key_arn"] = aws.StringValue(description.EncryptionConfiguration.KMSEncryptionConfig.AWSKMSKeyARN)
	}

	return []map[string]interface{}{m}
}

func flattenDataFormatConversionConfiguration(dfcc *firehose.DataFormatConversionConfiguration) []map[string]interface{} {
	if dfcc == nil {
		return []map[string]interface{}{}
	}

	enabled := aws.BoolValue(dfcc.Enabled)
	ifc := flattenInputFormatConfiguration(dfcc.InputFormatConfiguration)
	ofc := flattenOutputFormatConfiguration(dfcc.OutputFormatConfiguration)
	sc := flattenSchemaConfiguration(dfcc.SchemaConfiguration)

	// The AWS SDK can represent "no data format conversion configuration" in two ways:
	// 1. With a nil value
	// 2. With enabled set to false and nil for ALL the config sections.
	// We normalize this with an empty configuration in the state due
	// to the existing Default: true on the enabled attribute.
	if !enabled && len(ifc) == 0 && len(ofc) == 0 && len(sc) == 0 {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled":                     enabled,
		"input_format_configuration":  ifc,
		"output_format_configuration": ofc,
		"schema_configuration":        sc,
	}

	return []map[string]interface{}{m}
}

func flattenInputFormatConfiguration(ifc *firehose.InputFormatConfiguration) []map[string]interface{} {
	if ifc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"deserializer": flattenDeserializer(ifc.Deserializer),
	}

	return []map[string]interface{}{m}
}

func flattenDeserializer(deserializer *firehose.Deserializer) []map[string]interface{} {
	if deserializer == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"hive_json_ser_de":   flattenHiveJSONSerDe(deserializer.HiveJsonSerDe),
		"open_x_json_ser_de": flattenOpenXJSONSerDe(deserializer.OpenXJsonSerDe),
	}

	return []map[string]interface{}{m}
}

func flattenHiveJSONSerDe(hjsd *firehose.HiveJsonSerDe) []map[string]interface{} {
	if hjsd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"timestamp_formats": flex.FlattenStringList(hjsd.TimestampFormats),
	}

	return []map[string]interface{}{m}
}

func flattenOpenXJSONSerDe(oxjsd *firehose.OpenXJsonSerDe) []map[string]interface{} {
	if oxjsd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"column_to_json_key_mappings":              aws.StringValueMap(oxjsd.ColumnToJsonKeyMappings),
		"convert_dots_in_json_keys_to_underscores": aws.BoolValue(oxjsd.ConvertDotsInJsonKeysToUnderscores),
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference

	m["case_insensitive"] = true
	if oxjsd.CaseInsensitive != nil {
		m["case_insensitive"] = aws.BoolValue(oxjsd.CaseInsensitive)
	}

	return []map[string]interface{}{m}
}

func flattenOutputFormatConfiguration(ofc *firehose.OutputFormatConfiguration) []map[string]interface{} {
	if ofc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"serializer": flattenSerializer(ofc.Serializer),
	}

	return []map[string]interface{}{m}
}

func flattenSerializer(serializer *firehose.Serializer) []map[string]interface{} {
	if serializer == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"orc_ser_de":     flattenOrcSerDe(serializer.OrcSerDe),
		"parquet_ser_de": flattenParquetSerDe(serializer.ParquetSerDe),
	}

	return []map[string]interface{}{m}
}

func flattenOrcSerDe(osd *firehose.OrcSerDe) []map[string]interface{} {
	if osd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"bloom_filter_columns":     aws.StringValueSlice(osd.BloomFilterColumns),
		"dictionary_key_threshold": aws.Float64Value(osd.DictionaryKeyThreshold),
		"enable_padding":           aws.BoolValue(osd.EnablePadding),
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference

	m["block_size_bytes"] = 268435456
	if osd.BlockSizeBytes != nil {
		m["block_size_bytes"] = int(aws.Int64Value(osd.BlockSizeBytes))
	}

	m["bloom_filter_false_positive_probability"] = 0.05
	if osd.BloomFilterFalsePositiveProbability != nil {
		m["bloom_filter_false_positive_probability"] = aws.Float64Value(osd.BloomFilterFalsePositiveProbability)
	}

	m["compression"] = firehose.OrcCompressionSnappy
	if osd.Compression != nil {
		m["compression"] = aws.StringValue(osd.Compression)
	}

	m["format_version"] = firehose.OrcFormatVersionV012
	if osd.FormatVersion != nil {
		m["format_version"] = aws.StringValue(osd.FormatVersion)
	}

	m["padding_tolerance"] = 0.05
	if osd.PaddingTolerance != nil {
		m["padding_tolerance"] = aws.Float64Value(osd.PaddingTolerance)
	}

	m["row_index_stride"] = 10000
	if osd.RowIndexStride != nil {
		m["row_index_stride"] = int(aws.Int64Value(osd.RowIndexStride))
	}

	m["stripe_size_bytes"] = 67108864
	if osd.StripeSizeBytes != nil {
		m["stripe_size_bytes"] = int(aws.Int64Value(osd.StripeSizeBytes))
	}

	return []map[string]interface{}{m}
}

func flattenParquetSerDe(psd *firehose.ParquetSerDe) []map[string]interface{} {
	if psd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enable_dictionary_compression": aws.BoolValue(psd.EnableDictionaryCompression),
		"max_padding_bytes":             int(aws.Int64Value(psd.MaxPaddingBytes)),
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference

	m["block_size_bytes"] = 268435456
	if psd.BlockSizeBytes != nil {
		m["block_size_bytes"] = int(aws.Int64Value(psd.BlockSizeBytes))
	}

	m["compression"] = firehose.ParquetCompressionSnappy
	if psd.Compression != nil {
		m["compression"] = aws.StringValue(psd.Compression)
	}

	m["page_size_bytes"] = 1048576
	if psd.PageSizeBytes != nil {
		m["page_size_bytes"] = int(aws.Int64Value(psd.PageSizeBytes))
	}

	m["writer_version"] = firehose.ParquetWriterVersionV1
	if psd.WriterVersion != nil {
		m["writer_version"] = aws.StringValue(psd.WriterVersion)
	}

	return []map[string]interface{}{m}
}

func flattenSchemaConfiguration(sc *firehose.SchemaConfiguration) []map[string]interface{} {
	if sc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"catalog_id":    aws.StringValue(sc.CatalogId),
		"database_name": aws.StringValue(sc.DatabaseName),
		"region":        aws.StringValue(sc.Region),
		"role_arn":      aws.StringValue(sc.RoleARN),
		"table_name":    aws.StringValue(sc.TableName),
		"version_id":    aws.StringValue(sc.VersionId),
	}

	return []map[string]interface{}{m}
}

func flattenHTTPEndpointRequestConfiguration(rc *firehose.HttpEndpointRequestConfiguration) []map[string]interface{} {
	if rc == nil {
		return []map[string]interface{}{}
	}

	requestConfiguration := make([]map[string]interface{}, 1)

	commonAttributes := make([]interface{}, 0)
	for _, params := range rc.CommonAttributes {
		name := aws.StringValue(params.AttributeName)
		value := aws.StringValue(params.AttributeValue)

		commonAttributes = append(commonAttributes, map[string]interface{}{
			"name":  name,
			"value": value,
		})
	}

	requestConfiguration[0] = map[string]interface{}{
		"common_attributes": commonAttributes,
		"content_encoding":  aws.StringValue(rc.ContentEncoding),
	}

	return requestConfiguration
}

func flattenProcessingConfiguration(pc *firehose.ProcessingConfiguration, roleArn string) []map[string]interface{} {
	if pc == nil {
		return []map[string]interface{}{}
	}

	processingConfiguration := make([]map[string]interface{}, 1)

	// It is necessary to explicitly filter this out
	// to prevent diffs during routine use and retain the ability
	// to show diffs if any field has drifted
	defaultLambdaParams := map[string]string{
		"NumberOfRetries":         "3",
		"RoleArn":                 roleArn,
		"BufferSizeInMBs":         "3",
		"BufferIntervalInSeconds": "60",
	}

	processors := make([]interface{}, len(pc.Processors))
	for i, p := range pc.Processors {
		t := aws.StringValue(p.Type)
		parameters := make([]interface{}, 0)

		for _, params := range p.Parameters {
			name := aws.StringValue(params.ParameterName)
			value := aws.StringValue(params.ParameterValue)

			if t == firehose.ProcessorTypeLambda {
				// Ignore defaults
				if v, ok := defaultLambdaParams[name]; ok && v == value {
					continue
				}
			}

			parameters = append(parameters, map[string]interface{}{
				"parameter_name":  name,
				"parameter_value": value,
			})
		}

		processors[i] = map[string]interface{}{
			"type":       t,
			"parameters": parameters,
		}
	}
	processingConfiguration[0] = map[string]interface{}{
		"enabled":    aws.BoolValue(pc.Enabled),
		"processors": processors,
	}
	return processingConfiguration
}

func flattenDynamicPartitioningConfiguration(dpc *firehose.DynamicPartitioningConfiguration) []map[string]interface{} {
	if dpc == nil {
		return []map[string]interface{}{}
	}

	dynamicPartitioningConfiguration := make([]map[string]interface{}, 1)

	dynamicPartitioningConfiguration[0] = map[string]interface{}{
		"enabled": aws.BoolValue(dpc.Enabled),
	}

	if dpc.RetryOptions != nil && dpc.RetryOptions.DurationInSeconds != nil {
		dynamicPartitioningConfiguration[0]["retry_duration"] = int(aws.Int64Value(dpc.RetryOptions.DurationInSeconds))
	}

	return dynamicPartitioningConfiguration
}

func flattenKinesisStreamSourceDescription(desc *firehose.KinesisStreamSourceDescription) []interface{} {
	if desc == nil {
		return []interface{}{}
	}

	mDesc := map[string]interface{}{
		"kinesis_stream_arn": aws.StringValue(desc.KinesisStreamARN),
		"role_arn":           aws.StringValue(desc.RoleARN),
	}

	return []interface{}{mDesc}
}

func flattenHTTPEndpointDestinationDescription(description *firehose.HttpEndpointDestinationDescription, configuredAccessKey string) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}
	m := map[string]interface{}{
		"access_key":                 configuredAccessKey,
		"url":                        aws.StringValue(description.EndpointConfiguration.Url),
		"name":                       aws.StringValue(description.EndpointConfiguration.Name),
		"role_arn":                   aws.StringValue(description.RoleARN),
		"s3_backup_mode":             aws.StringValue(description.S3BackupMode),
		"s3_configuration":           flattenS3DestinationDescription(description.S3DestinationDescription),
		"request_configuration":      flattenHTTPEndpointRequestConfiguration(description.RequestConfiguration),
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"processing_configuration":   flattenProcessingConfiguration(description.ProcessingConfiguration, aws.StringValue(description.RoleARN)),
	}

	if description.RetryOptions != nil {
		m["retry_duration"] = int(aws.Int64Value(description.RetryOptions.DurationInSeconds))
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.Int64Value(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.Int64Value(description.BufferingHints.SizeInMBs))
	}

	return []map[string]interface{}{m}
}

func isDeliveryStreamOptionDisabled(v interface{}) bool {
	tfList := v.([]interface{})
	if len(tfList) == 0 || tfList[0] == nil {
		return true
	}
	tfMap := tfList[0].(map[string]interface{})

	var enabled bool

	if v, ok := tfMap["enabled"]; ok {
		enabled = v.(bool)
	}

	return !enabled
}
