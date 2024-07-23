// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package firehose

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type destinationType string

const (
	destinationTypeElasticsearch        destinationType = "elasticsearch"
	destinationTypeExtendedS3           destinationType = "extended_s3"
	destinationTypeHTTPEndpoint         destinationType = "http_endpoint"
	destinationTypeOpenSearch           destinationType = "opensearch"
	destinationTypeOpenSearchServerless destinationType = "opensearchserverless"
	destinationTypeRedshift             destinationType = "redshift"
	destinationTypeSnowflake            destinationType = "snowflake"
	destinationTypeSplunk               destinationType = "splunk"
)

func (destinationType) Values() []destinationType {
	return []destinationType{
		destinationTypeElasticsearch,
		destinationTypeExtendedS3,
		destinationTypeHTTPEndpoint,
		destinationTypeOpenSearch,
		destinationTypeOpenSearchServerless,
		destinationTypeRedshift,
		destinationTypeSnowflake,
		destinationTypeSplunk,
	}
}

// @SDKResource("aws_kinesis_firehose_delivery_stream", name="Delivery Stream")
// @Tags(identifierAttribute="name")
func resourceDeliveryStream() *schema.Resource {
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
				d.Set(names.AttrName, resourceParts[1])
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
							names.AttrEnabled: {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							names.AttrLogGroupName: {
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
							names.AttrEnabled: {
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
							names.AttrEnabled: {
								Type:     schema.TypeBool,
								Optional: true,
							},
							"processors": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrParameters: {
											// See AWS::KinesisFirehose::DeliveryStream CloudFormation resource schema.
											// uniqueItems is true and insertionOrder is true.
											// However, IRL the order of the processors is not important.
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"parameter_name": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[types.ProcessorParameterName](),
													},
													"parameter_value": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringLenBetween(1, 5120),
													},
												},
											},
										},
										names.AttrType: {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[types.ProcessorType](),
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
										names.AttrName: {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrValue: {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"content_encoding": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.ContentEncodingNone,
								ValidateDiagFunc: enum.Validate[types.ContentEncoding](),
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
							Type:     schema.TypeInt,
							Optional: true,
							Default:  300,
						},
						"buffering_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
						"compression_format": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.CompressionFormatUncompressed,
							ValidateDiagFunc: enum.Validate[types.CompressionFormat](),
						},
						"error_output_prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						names.AttrKMSKeyARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrPrefix: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrRoleARN: {
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
			secretsManagerConfigurationSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrEnabled: {
								Type:     schema.TypeBool,
								Optional: true,
								Computed: true,
								ForceNew: true,
							},
							"secret_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				}
			}

			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				names.AttrDestination: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					StateFunc: func(v interface{}) string {
						value := v.(string)
						return strings.ToLower(value)
					},
					ValidateDiagFunc: enum.Validate[destinationType](),
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
								ValidateFunc: validation.IntBetween(0, 900),
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
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.ElasticsearchIndexRotationPeriodOneDay,
								ValidateDiagFunc: enum.Validate[types.ElasticsearchIndexRotationPeriod](),
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_mode": {
								Type:             schema.TypeString,
								ForceNew:         true,
								Optional:         true,
								Default:          types.ElasticsearchS3BackupModeFailedDocumentsOnly,
								ValidateDiagFunc: enum.Validate[types.ElasticsearchS3BackupMode](),
							},
							"s3_configuration": s3ConfigurationSchema(),
							"type_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 100),
							},
							names.AttrVPCConfig: {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
										names.AttrSecurityGroupIDs: {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrSubnetIDs: {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrVPCID: {
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
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.CompressionFormatUncompressed,
								ValidateDiagFunc: enum.Validate[types.CompressionFormat](),
							},
							"custom_time_zone": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      "UTC",
								ValidateFunc: validation.StringLenBetween(0, 50),
							},
							"data_format_conversion_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrEnabled: {
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
																				Type:             schema.TypeString,
																				Optional:         true,
																				Default:          types.OrcCompressionSnappy,
																				ValidateDiagFunc: enum.Validate[types.OrcCompression](),
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
																				Type:             schema.TypeString,
																				Optional:         true,
																				Default:          types.OrcFormatVersionV012,
																				ValidateDiagFunc: enum.Validate[types.OrcFormatVersion](),
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
																				Type:             schema.TypeString,
																				Optional:         true,
																				Default:          types.ParquetCompressionSnappy,
																				ValidateDiagFunc: enum.Validate[types.ParquetCompression](),
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
																				Type:             schema.TypeString,
																				Optional:         true,
																				Default:          types.ParquetWriterVersionV1,
																				ValidateDiagFunc: enum.Validate[types.ParquetWriterVersion](),
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
													names.AttrCatalogID: {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													names.AttrDatabaseName: {
														Type:     schema.TypeString,
														Required: true,
													},
													names.AttrRegion: {
														Type:     schema.TypeString,
														Optional: true,
														Computed: true,
													},
													names.AttrRoleARN: {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: verify.ValidARN,
													},
													names.AttrTableName: {
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
							"file_extension": {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(0, 50),
									validation.StringMatch(regexache.MustCompile(`^$|\.[0-9a-z!\-_.*'()]+`), ""),
								),
							},
							names.AttrKMSKeyARN: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
							names.AttrPrefix: {
								Type:     schema.TypeString,
								Optional: true,
							},
							"processing_configuration": processingConfigurationSchema(),
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_configuration": s3BackupConfigurationSchema(),
							"s3_backup_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.S3BackupModeDisabled,
								ValidateDiagFunc: enum.Validate[types.S3BackupMode](),
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
							names.AttrAccessKey: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 4096),
								Sensitive:    true,
							},
							"buffering_interval": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(0, 900),
							},
							"buffering_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      5,
								ValidateFunc: validation.IntBetween(1, 100),
							},
							"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
							names.AttrName: {
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
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.HttpEndpointS3BackupModeFailedDataOnly,
								ValidateDiagFunc: enum.Validate[types.HttpEndpointS3BackupMode](),
							},
							"s3_configuration":              s3ConfigurationSchema(),
							"secrets_manager_configuration": secretsManagerConfigurationSchema(),
							names.AttrURL: {
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
							names.AttrRoleARN: {
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
											Type:             schema.TypeString,
											Required:         true,
											ForceNew:         true,
											ValidateDiagFunc: enum.Validate[types.Connectivity](),
										},
										names.AttrRoleARN: {
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
				names.AttrName: {
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
								ValidateFunc: validation.IntBetween(0, 900),
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
							"document_id_options": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"default_document_id_format": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[types.DefaultDocumentIdFormat](),
										},
									},
								},
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
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.AmazonopensearchserviceIndexRotationPeriodOneDay,
								ValidateDiagFunc: enum.Validate[types.AmazonopensearchserviceIndexRotationPeriod](),
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      300,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_mode": {
								Type:             schema.TypeString,
								ForceNew:         true,
								Optional:         true,
								Default:          types.AmazonopensearchserviceS3BackupModeFailedDocumentsOnly,
								ValidateDiagFunc: enum.Validate[types.AmazonopensearchserviceS3BackupMode](),
							},
							"s3_configuration": s3ConfigurationSchema(),
							"type_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 100),
							},
							names.AttrVPCConfig: {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
										names.AttrSecurityGroupIDs: {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrSubnetIDs: {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrVPCID: {
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
								ValidateFunc: validation.IntBetween(0, 900),
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
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_mode": {
								Type:             schema.TypeString,
								ForceNew:         true,
								Optional:         true,
								Default:          types.AmazonOpenSearchServerlessS3BackupModeFailedDocumentsOnly,
								ValidateDiagFunc: enum.Validate[types.AmazonOpenSearchServerlessS3BackupMode](),
							},
							"s3_configuration": s3ConfigurationSchema(),
							names.AttrVPCConfig: {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
										names.AttrSecurityGroupIDs: {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrSubnetIDs: {
											Type:     schema.TypeSet,
											Required: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrVPCID: {
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
							names.AttrPassword: {
								Type:      schema.TypeString,
								Optional:  true,
								Sensitive: true,
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      3600,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_configuration": s3BackupConfigurationSchema(),
							"s3_backup_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.RedshiftS3BackupModeDisabled,
								ValidateDiagFunc: enum.Validate[types.RedshiftS3BackupMode](),
							},
							"s3_configuration":              s3ConfigurationSchema(),
							"secrets_manager_configuration": secretsManagerConfigurationSchema(),
							names.AttrUsername: {
								Type:     schema.TypeString,
								Optional: true,
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
							names.AttrEnabled: {
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
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.KeyTypeAwsOwnedCmk,
								ValidateDiagFunc: enum.Validate[types.KeyType](),
								RequiredWith:     []string{"server_side_encryption.0.enabled"},
							},
						},
					},
				},
				"snowflake_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"account_url": {
								Type:     schema.TypeString,
								Required: true,
							},
							"cloudwatch_logging_options": cloudWatchLoggingOptionsSchema(),
							"content_column_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							"data_loading_option": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.SnowflakeDataLoadingOptionJsonMapping,
								ValidateDiagFunc: enum.Validate[types.SnowflakeDataLoadingOption](),
							},
							names.AttrDatabase: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							"key_passphrase": {
								Type:         schema.TypeString,
								Optional:     true,
								Sensitive:    true,
								ValidateFunc: validation.StringLenBetween(7, 255),
							},
							"metadata_column_name": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							names.AttrPrivateKey: {
								Type:      schema.TypeString,
								Optional:  true,
								Sensitive: true,
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      60,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							names.AttrRoleARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"s3_backup_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.SnowflakeS3BackupModeFailedDataOnly,
								ValidateDiagFunc: enum.Validate[types.SnowflakeS3BackupMode](),
							},
							"s3_configuration": s3ConfigurationSchema(),
							names.AttrSchema: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							"secrets_manager_configuration": secretsManagerConfigurationSchema(),
							"snowflake_role_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrEnabled: {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  false,
										},
										"snowflake_role": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
									},
								},
								DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							},
							"snowflake_vpc_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"private_link_vpce_id": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"table": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							"user": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
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
							"buffering_interval": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      60,
								ValidateFunc: validation.IntBetween(0, 60),
							},
							"buffering_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      5,
								ValidateFunc: validation.IntBetween(1, 5),
							},
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
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.HECEndpointTypeRaw,
								ValidateDiagFunc: enum.Validate[types.HECEndpointType](),
							},
							"hec_token": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"processing_configuration": processingConfigurationSchema(),
							"retry_duration": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      3600,
								ValidateFunc: validation.IntBetween(0, 7200),
							},
							"s3_backup_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.SplunkS3BackupModeFailedEventsOnly,
								ValidateDiagFunc: enum.Validate[types.SplunkS3BackupMode](),
							},
							"s3_configuration":              s3ConfigurationSchema(),
							"secrets_manager_configuration": secretsManagerConfigurationSchema(),
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
				destination := destinationType(d.Get(names.AttrDestination).(string))
				requiredAttribute := map[destinationType]string{
					destinationTypeElasticsearch:        "elasticsearch_configuration",
					destinationTypeExtendedS3:           "extended_s3_configuration",
					destinationTypeHTTPEndpoint:         "http_endpoint_configuration",
					destinationTypeOpenSearch:           "opensearch_configuration",
					destinationTypeOpenSearchServerless: "opensearchserverless_configuration",
					destinationTypeRedshift:             "redshift_configuration",
					destinationTypeSnowflake:            "snowflake_configuration",
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
	conn := meta.(*conns.AWSClient).FirehoseClient(ctx)

	sn := d.Get(names.AttrName).(string)
	input := &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(sn),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("kinesis_source_configuration"); ok {
		input.DeliveryStreamType = types.DeliveryStreamTypeKinesisStreamAsSource
		input.KinesisStreamSourceConfiguration = expandKinesisStreamSourceConfiguration(v.([]interface{})[0].(map[string]interface{}))
	} else if v, ok := d.GetOk("msk_source_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeliveryStreamType = types.DeliveryStreamTypeMSKAsSource
		input.MSKSourceConfiguration = expandMSKSourceConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	switch v := destinationType(d.Get(names.AttrDestination).(string)); v {
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
	case destinationTypeSnowflake:
		if v, ok := d.GetOk("snowflake_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.SnowflakeDestinationConfiguration = expandSnowflakeDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	case destinationTypeSplunk:
		if v, ok := d.GetOk("splunk_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.SplunkDestinationConfiguration = expandSplunkDestinationConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	_, err := retryDeliveryStreamOp(ctx, func() (interface{}, error) {
		return conn.CreateDeliveryStream(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Firehose Delivery Stream (%s): %s", sn, err)
	}

	output, err := waitDeliveryStreamCreated(ctx, conn, sn, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Firehose Delivery Stream (%s) create: %s", sn, err)
	}

	d.SetId(aws.ToString(output.DeliveryStreamARN))

	if v, ok := d.GetOk("server_side_encryption"); ok && !isDeliveryStreamOptionDisabled(v) {
		input := &firehose.StartDeliveryStreamEncryptionInput{
			DeliveryStreamEncryptionConfigurationInput: expandDeliveryStreamEncryptionConfigurationInput(v.([]interface{})),
			DeliveryStreamName:                         aws.String(sn),
		}

		_, err := conn.StartDeliveryStreamEncryption(ctx, input)

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
	conn := meta.(*conns.AWSClient).FirehoseClient(ctx)

	sn := d.Get(names.AttrName).(string)
	s, err := findDeliveryStreamByName(ctx, conn, sn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Firehose Delivery Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Firehose Delivery Stream (%s): %s", sn, err)
	}

	d.Set(names.AttrARN, s.DeliveryStreamARN)
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
	d.Set(names.AttrName, s.DeliveryStreamName)
	d.Set("version_id", s.VersionId)

	sseOptions := map[string]interface{}{
		names.AttrEnabled: false,
		"key_type":        types.KeyTypeAwsOwnedCmk,
	}
	if s.DeliveryStreamEncryptionConfiguration != nil && s.DeliveryStreamEncryptionConfiguration.Status == types.DeliveryStreamEncryptionStatusEnabled {
		sseOptions[names.AttrEnabled] = true
		sseOptions["key_type"] = s.DeliveryStreamEncryptionConfiguration.KeyType

		if v := s.DeliveryStreamEncryptionConfiguration.KeyARN; v != nil {
			sseOptions["key_arn"] = aws.ToString(v)
		}
	}
	if err := d.Set("server_side_encryption", []map[string]interface{}{sseOptions}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting server_side_encryption: %s", err)
	}

	if len(s.Destinations) > 0 {
		destination := s.Destinations[0]
		switch {
		case destination.ElasticsearchDestinationDescription != nil:
			d.Set(names.AttrDestination, destinationTypeElasticsearch)
			if err := d.Set("elasticsearch_configuration", flattenElasticsearchDestinationDescription(destination.ElasticsearchDestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting elasticsearch_configuration: %s", err)
			}
		case destination.HttpEndpointDestinationDescription != nil:
			d.Set(names.AttrDestination, destinationTypeHTTPEndpoint)
			configuredAccessKey := d.Get("http_endpoint_configuration.0.access_key").(string)
			if err := d.Set("http_endpoint_configuration", flattenHTTPEndpointDestinationDescription(destination.HttpEndpointDestinationDescription, configuredAccessKey)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting http_endpoint_configuration: %s", err)
			}
		case destination.AmazonopensearchserviceDestinationDescription != nil:
			d.Set(names.AttrDestination, destinationTypeOpenSearch)
			if err := d.Set("opensearch_configuration", flattenAmazonopensearchserviceDestinationDescription(destination.AmazonopensearchserviceDestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting opensearch_configuration: %s", err)
			}
		case destination.AmazonOpenSearchServerlessDestinationDescription != nil:
			d.Set(names.AttrDestination, destinationTypeOpenSearchServerless)
			if err := d.Set("opensearchserverless_configuration", flattenAmazonOpenSearchServerlessDestinationDescription(destination.AmazonOpenSearchServerlessDestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting opensearchserverless_configuration: %s", err)
			}
		case destination.RedshiftDestinationDescription != nil:
			d.Set(names.AttrDestination, destinationTypeRedshift)
			configuredPassword := d.Get("redshift_configuration.0.password").(string)
			if err := d.Set("redshift_configuration", flattenRedshiftDestinationDescription(destination.RedshiftDestinationDescription, configuredPassword)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting redshift_configuration: %s", err)
			}
		case destination.SnowflakeDestinationDescription != nil:
			d.Set(names.AttrDestination, destinationTypeSnowflake)
			configuredKeyPassphrase := d.Get("snowflake_configuration.0.key_passphrase").(string)
			configuredPrivateKey := d.Get("snowflake_configuration.0.private_key").(string)
			if err := d.Set("snowflake_configuration", flattenSnowflakeDestinationDescription(destination.SnowflakeDestinationDescription, configuredKeyPassphrase, configuredPrivateKey)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting snowflake_configuration: %s", err)
			}
		case destination.SplunkDestinationDescription != nil:
			d.Set(names.AttrDestination, destinationTypeSplunk)
			if err := d.Set("splunk_configuration", flattenSplunkDestinationDescription(destination.SplunkDestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting splunk_configuration: %s", err)
			}
		default:
			d.Set(names.AttrDestination, destinationTypeExtendedS3)
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
	conn := meta.(*conns.AWSClient).FirehoseClient(ctx)

	sn := d.Get(names.AttrName).(string)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &firehose.UpdateDestinationInput{
			CurrentDeliveryStreamVersionId: aws.String(d.Get("version_id").(string)),
			DeliveryStreamName:             aws.String(sn),
			DestinationId:                  aws.String(d.Get("destination_id").(string)),
		}

		switch v := destinationType(d.Get(names.AttrDestination).(string)); v {
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
		case destinationTypeSnowflake:
			if v, ok := d.GetOk("snowflake_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.SnowflakeDestinationUpdate = expandSnowflakeDestinationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		case destinationTypeSplunk:
			if v, ok := d.GetOk("splunk_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.SplunkDestinationUpdate = expandSplunkDestinationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		_, err := retryDeliveryStreamOp(ctx, func() (interface{}, error) {
			return conn.UpdateDestination(ctx, input)
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

			_, err := conn.StopDeliveryStreamEncryption(ctx, input)

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

			_, err := conn.StartDeliveryStreamEncryption(ctx, input)

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
	conn := meta.(*conns.AWSClient).FirehoseClient(ctx)

	sn := d.Get(names.AttrName).(string)

	log.Printf("[DEBUG] Deleting Kinesis Firehose Delivery Stream: (%s)", sn)
	_, err := conn.DeleteDeliveryStream(ctx, &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(sn),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
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
			if errs.IsAErrorMessageContains[*types.InvalidArgumentException](err, "Access was denied") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*types.InvalidArgumentException](err, "is not authorized to") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*types.InvalidArgumentException](err, "Please make sure the role specified in VpcConfiguration has permissions") {
				return true, err
			}
			// InvalidArgumentException: Verify that the IAM role has access to the Elasticsearch domain.
			if errs.IsAErrorMessageContains[*types.InvalidArgumentException](err, "Verify that the IAM role has access") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*types.InvalidArgumentException](err, "Firehose is unable to assume role") {
				return true, err
			}
			return false, err
		},
	)
}

func findDeliveryStreamByName(ctx context.Context, conn *firehose.Client, name string) (*types.DeliveryStreamDescription, error) {
	input := &firehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: aws.String(name),
	}

	output, err := conn.DescribeDeliveryStream(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DeliveryStreamDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DeliveryStreamDescription, nil
}

func statusDeliveryStream(ctx context.Context, conn *firehose.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDeliveryStreamByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.DeliveryStreamStatus), nil
	}
}

func waitDeliveryStreamCreated(ctx context.Context, conn *firehose.Client, name string, timeout time.Duration) (*types.DeliveryStreamDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DeliveryStreamStatusCreating),
		Target:  enum.Slice(types.DeliveryStreamStatusActive),
		Refresh: statusDeliveryStream(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DeliveryStreamDescription); ok {
		if status, failureDescription := output.DeliveryStreamStatus, output.FailureDescription; status == types.DeliveryStreamStatusCreatingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", failureDescription.Type, aws.ToString(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func waitDeliveryStreamDeleted(ctx context.Context, conn *firehose.Client, name string, timeout time.Duration) (*types.DeliveryStreamDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DeliveryStreamStatusDeleting),
		Target:  []string{},
		Refresh: statusDeliveryStream(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DeliveryStreamDescription); ok {
		if status, failureDescription := output.DeliveryStreamStatus, output.FailureDescription; status == types.DeliveryStreamStatusDeletingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", failureDescription.Type, aws.ToString(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func findDeliveryStreamEncryptionConfigurationByName(ctx context.Context, conn *firehose.Client, name string) (*types.DeliveryStreamEncryptionConfiguration, error) {
	output, err := findDeliveryStreamByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.DeliveryStreamEncryptionConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output.DeliveryStreamEncryptionConfiguration, nil
}

func statusDeliveryStreamEncryptionConfiguration(ctx context.Context, conn *firehose.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDeliveryStreamEncryptionConfigurationByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDeliveryStreamEncryptionEnabled(ctx context.Context, conn *firehose.Client, name string, timeout time.Duration) (*types.DeliveryStreamEncryptionConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DeliveryStreamEncryptionStatusEnabling),
		Target:  enum.Slice(types.DeliveryStreamEncryptionStatusEnabled),
		Refresh: statusDeliveryStreamEncryptionConfiguration(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DeliveryStreamEncryptionConfiguration); ok {
		if status, failureDescription := output.Status, output.FailureDescription; status == types.DeliveryStreamEncryptionStatusEnablingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", failureDescription.Type, aws.ToString(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func waitDeliveryStreamEncryptionDisabled(ctx context.Context, conn *firehose.Client, name string, timeout time.Duration) (*types.DeliveryStreamEncryptionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DeliveryStreamEncryptionStatusDisabling),
		Target:  enum.Slice(types.DeliveryStreamEncryptionStatusDisabled),
		Refresh: statusDeliveryStreamEncryptionConfiguration(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DeliveryStreamEncryptionConfiguration); ok {
		if status, failureDescription := output.Status, output.FailureDescription; status == types.DeliveryStreamEncryptionStatusDisablingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", failureDescription.Type, aws.ToString(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func expandKinesisStreamSourceConfiguration(source map[string]interface{}) *types.KinesisStreamSourceConfiguration {
	configuration := &types.KinesisStreamSourceConfiguration{
		KinesisStreamARN: aws.String(source["kinesis_stream_arn"].(string)),
		RoleARN:          aws.String(source[names.AttrRoleARN].(string)),
	}

	return configuration
}

func expandS3DestinationConfiguration(tfList []interface{}) *types.S3DestinationConfiguration {
	s3 := tfList[0].(map[string]interface{})

	configuration := &types.S3DestinationConfiguration{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3[names.AttrRoleARN].(string)),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(s3["buffering_size"].(int))),
		},
		Prefix:                  expandPrefix(s3),
		CompressionFormat:       types.CompressionFormat(s3["compression_format"].(string)),
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

func expandS3DestinationConfigurationBackup(d map[string]interface{}) *types.S3DestinationConfiguration {
	config := d["s3_backup_configuration"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	s3 := config[0].(map[string]interface{})

	configuration := &types.S3DestinationConfiguration{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3[names.AttrRoleARN].(string)),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(s3["buffering_size"].(int))),
		},
		Prefix:                  expandPrefix(s3),
		CompressionFormat:       types.CompressionFormat(s3["compression_format"].(string)),
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

func expandExtendedS3DestinationConfiguration(s3 map[string]interface{}) *types.ExtendedS3DestinationConfiguration {
	roleARN := s3[names.AttrRoleARN].(string)
	configuration := &types.ExtendedS3DestinationConfiguration{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(roleARN),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(s3["buffering_size"].(int))),
		},
		Prefix:                            expandPrefix(s3),
		CompressionFormat:                 types.CompressionFormat(s3["compression_format"].(string)),
		CustomTimeZone:                    aws.String(s3["custom_time_zone"].(string)),
		DataFormatConversionConfiguration: expandDataFormatConversionConfiguration(s3["data_format_conversion_configuration"].([]interface{})),
		EncryptionConfiguration:           expandEncryptionConfiguration(s3),
		FileExtension:                     aws.String(s3["file_extension"].(string)),
	}

	if _, ok := s3["processing_configuration"]; ok {
		configuration.ProcessingConfiguration = expandProcessingConfiguration(s3, destinationTypeExtendedS3, roleARN)
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
		configuration.S3BackupMode = types.S3BackupMode(s3BackupMode.(string))
		configuration.S3BackupConfiguration = expandS3DestinationConfigurationBackup(s3)
	}

	return configuration
}

func expandS3DestinationUpdate(tfList []interface{}) *types.S3DestinationUpdate {
	s3 := tfList[0].(map[string]interface{})
	configuration := &types.S3DestinationUpdate{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3[names.AttrRoleARN].(string)),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(s3["buffering_size"].(int))),
		},
		ErrorOutputPrefix:        aws.String(s3["error_output_prefix"].(string)),
		Prefix:                   expandPrefix(s3),
		CompressionFormat:        types.CompressionFormat(s3["compression_format"].(string)),
		EncryptionConfiguration:  expandEncryptionConfiguration(s3),
		CloudWatchLoggingOptions: expandCloudWatchLoggingOptions(s3),
	}

	if _, ok := s3["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(s3)
	}

	return configuration
}

func expandS3DestinationUpdateBackup(d map[string]interface{}) *types.S3DestinationUpdate {
	config := d["s3_backup_configuration"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	s3 := config[0].(map[string]interface{})

	configuration := &types.S3DestinationUpdate{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(s3[names.AttrRoleARN].(string)),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(s3["buffering_size"].(int))),
		},
		ErrorOutputPrefix:        aws.String(s3["error_output_prefix"].(string)),
		Prefix:                   expandPrefix(s3),
		CompressionFormat:        types.CompressionFormat(s3["compression_format"].(string)),
		EncryptionConfiguration:  expandEncryptionConfiguration(s3),
		CloudWatchLoggingOptions: expandCloudWatchLoggingOptions(s3),
	}

	if _, ok := s3["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(s3)
	}

	return configuration
}

func expandExtendedS3DestinationUpdate(s3 map[string]interface{}) *types.ExtendedS3DestinationUpdate {
	roleARN := s3[names.AttrRoleARN].(string)
	configuration := &types.ExtendedS3DestinationUpdate{
		BucketARN: aws.String(s3["bucket_arn"].(string)),
		RoleARN:   aws.String(roleARN),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(s3["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(s3["buffering_size"].(int))),
		},
		CustomTimeZone:                    aws.String(s3["custom_time_zone"].(string)),
		ErrorOutputPrefix:                 aws.String(s3["error_output_prefix"].(string)),
		FileExtension:                     aws.String(s3["file_extension"].(string)),
		Prefix:                            expandPrefix(s3),
		CompressionFormat:                 types.CompressionFormat(s3["compression_format"].(string)),
		EncryptionConfiguration:           expandEncryptionConfiguration(s3),
		DataFormatConversionConfiguration: expandDataFormatConversionConfiguration(s3["data_format_conversion_configuration"].([]interface{})),
		CloudWatchLoggingOptions:          expandCloudWatchLoggingOptions(s3),
		ProcessingConfiguration:           expandProcessingConfiguration(s3, destinationTypeExtendedS3, roleARN),
	}

	if _, ok := s3["cloudwatch_logging_options"]; ok {
		configuration.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(s3)
	}

	if _, ok := s3["dynamic_partitioning_configuration"]; ok {
		configuration.DynamicPartitioningConfiguration = expandDynamicPartitioningConfiguration(s3)
	}

	if s3BackupMode, ok := s3["s3_backup_mode"]; ok {
		configuration.S3BackupMode = types.S3BackupMode(s3BackupMode.(string))
		configuration.S3BackupUpdate = expandS3DestinationUpdateBackup(s3)
	}

	return configuration
}

func expandDataFormatConversionConfiguration(l []interface{}) *types.DataFormatConversionConfiguration {
	if len(l) == 0 || l[0] == nil {
		// It is possible to just pass nil here, but this seems to be the
		// canonical form that AWS uses, and is less likely to produce diffs.
		return &types.DataFormatConversionConfiguration{
			Enabled: aws.Bool(false),
		}
	}

	m := l[0].(map[string]interface{})

	return &types.DataFormatConversionConfiguration{
		Enabled:                   aws.Bool(m[names.AttrEnabled].(bool)),
		InputFormatConfiguration:  expandInputFormatConfiguration(m["input_format_configuration"].([]interface{})),
		OutputFormatConfiguration: expandOutputFormatConfiguration(m["output_format_configuration"].([]interface{})),
		SchemaConfiguration:       expandSchemaConfiguration(m["schema_configuration"].([]interface{})),
	}
}

func expandInputFormatConfiguration(l []interface{}) *types.InputFormatConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &types.InputFormatConfiguration{
		Deserializer: expandDeserializer(m["deserializer"].([]interface{})),
	}
}

func expandDeserializer(l []interface{}) *types.Deserializer {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &types.Deserializer{
		HiveJsonSerDe:  expandHiveJSONSerDe(m["hive_json_ser_de"].([]interface{})),
		OpenXJsonSerDe: expandOpenXJSONSerDe(m["open_x_json_ser_de"].([]interface{})),
	}
}

func expandHiveJSONSerDe(l []interface{}) *types.HiveJsonSerDe {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &types.HiveJsonSerDe{}
	}

	m := l[0].(map[string]interface{})

	return &types.HiveJsonSerDe{
		TimestampFormats: flex.ExpandStringValueList(m["timestamp_formats"].([]interface{})),
	}
}

func expandOpenXJSONSerDe(l []interface{}) *types.OpenXJsonSerDe {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &types.OpenXJsonSerDe{}
	}

	m := l[0].(map[string]interface{})

	return &types.OpenXJsonSerDe{
		CaseInsensitive:                    aws.Bool(m["case_insensitive"].(bool)),
		ColumnToJsonKeyMappings:            flex.ExpandStringValueMap(m["column_to_json_key_mappings"].(map[string]interface{})),
		ConvertDotsInJsonKeysToUnderscores: aws.Bool(m["convert_dots_in_json_keys_to_underscores"].(bool)),
	}
}

func expandOutputFormatConfiguration(l []interface{}) *types.OutputFormatConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &types.OutputFormatConfiguration{
		Serializer: expandSerializer(m["serializer"].([]interface{})),
	}
}

func expandSerializer(l []interface{}) *types.Serializer {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &types.Serializer{
		OrcSerDe:     expandOrcSerDe(m["orc_ser_de"].([]interface{})),
		ParquetSerDe: expandParquetSerDe(m["parquet_ser_de"].([]interface{})),
	}
}

func expandOrcSerDe(l []interface{}) *types.OrcSerDe {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &types.OrcSerDe{}
	}

	m := l[0].(map[string]interface{})

	orcSerDe := &types.OrcSerDe{
		BlockSizeBytes:                      aws.Int32(int32(m["block_size_bytes"].(int))),
		BloomFilterFalsePositiveProbability: aws.Float64(m["bloom_filter_false_positive_probability"].(float64)),
		Compression:                         types.OrcCompression(m["compression"].(string)),
		DictionaryKeyThreshold:              aws.Float64(m["dictionary_key_threshold"].(float64)),
		EnablePadding:                       aws.Bool(m["enable_padding"].(bool)),
		FormatVersion:                       types.OrcFormatVersion(m["format_version"].(string)),
		PaddingTolerance:                    aws.Float64(m["padding_tolerance"].(float64)),
		RowIndexStride:                      aws.Int32(int32(m["row_index_stride"].(int))),
		StripeSizeBytes:                     aws.Int32(int32(m["stripe_size_bytes"].(int))),
	}

	if v, ok := m["bloom_filter_columns"].([]interface{}); ok && len(v) > 0 {
		orcSerDe.BloomFilterColumns = flex.ExpandStringValueList(v)
	}

	return orcSerDe
}

func expandParquetSerDe(l []interface{}) *types.ParquetSerDe {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &types.ParquetSerDe{}
	}

	m := l[0].(map[string]interface{})

	return &types.ParquetSerDe{
		BlockSizeBytes:              aws.Int32(int32(m["block_size_bytes"].(int))),
		Compression:                 types.ParquetCompression(m["compression"].(string)),
		EnableDictionaryCompression: aws.Bool(m["enable_dictionary_compression"].(bool)),
		MaxPaddingBytes:             aws.Int32(int32(m["max_padding_bytes"].(int))),
		PageSizeBytes:               aws.Int32(int32(m["page_size_bytes"].(int))),
		WriterVersion:               types.ParquetWriterVersion(m["writer_version"].(string)),
	}
}

func expandSchemaConfiguration(l []interface{}) *types.SchemaConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &types.SchemaConfiguration{
		DatabaseName: aws.String(m[names.AttrDatabaseName].(string)),
		RoleARN:      aws.String(m[names.AttrRoleARN].(string)),
		TableName:    aws.String(m[names.AttrTableName].(string)),
		VersionId:    aws.String(m["version_id"].(string)),
	}

	if v, ok := m[names.AttrCatalogID].(string); ok && v != "" {
		config.CatalogId = aws.String(v)
	}
	if v, ok := m[names.AttrRegion].(string); ok && v != "" {
		config.Region = aws.String(v)
	}

	return config
}

func expandDynamicPartitioningConfiguration(s3 map[string]interface{}) *types.DynamicPartitioningConfiguration {
	config := s3["dynamic_partitioning_configuration"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	dynamicPartitioningConfig := config[0].(map[string]interface{})
	DynamicPartitioningConfiguration := &types.DynamicPartitioningConfiguration{
		Enabled: aws.Bool(dynamicPartitioningConfig[names.AttrEnabled].(bool)),
	}

	if retryDuration, ok := dynamicPartitioningConfig["retry_duration"]; ok {
		DynamicPartitioningConfiguration.RetryOptions = &types.RetryOptions{
			DurationInSeconds: aws.Int32(int32(retryDuration.(int))),
		}
	}

	return DynamicPartitioningConfiguration
}

func expandProcessingConfiguration(tfMap map[string]interface{}, destinationType destinationType, roleARN string) *types.ProcessingConfiguration {
	config := tfMap["processing_configuration"].([]interface{})
	if len(config) == 0 || config[0] == nil {
		// It is possible to just pass nil here, but this seems to be the
		// canonical form that AWS uses, and is less likely to produce diffs.
		return &types.ProcessingConfiguration{
			Enabled:    aws.Bool(false),
			Processors: []types.Processor{},
		}
	}

	processingConfiguration := config[0].(map[string]interface{})

	return &types.ProcessingConfiguration{
		Enabled:    aws.Bool(processingConfiguration[names.AttrEnabled].(bool)),
		Processors: expandProcessors(processingConfiguration["processors"].([]interface{}), destinationType, roleARN),
	}
}

func expandProcessors(processingConfigurationProcessors []interface{}, destinationType destinationType, roleARN string) []types.Processor {
	processors := []types.Processor{}

	for _, processor := range processingConfigurationProcessors {
		extractedProcessor := expandProcessor(processor.(map[string]interface{}))
		if extractedProcessor != nil {
			// Merge in defaults.
			for name, value := range defaultProcessorParameters(destinationType, extractedProcessor.Type, roleARN) {
				if !slices.ContainsFunc(extractedProcessor.Parameters, func(param types.ProcessorParameter) bool { return name == param.ParameterName }) {
					extractedProcessor.Parameters = append(extractedProcessor.Parameters, types.ProcessorParameter{
						ParameterName:  name,
						ParameterValue: aws.String(value),
					})
				}
			}
			processors = append(processors, *extractedProcessor)
		}
	}

	return processors
}

func expandProcessor(processingConfigurationProcessor map[string]interface{}) *types.Processor {
	var processor *types.Processor
	processorType := processingConfigurationProcessor[names.AttrType].(string)
	if processorType != "" {
		processor = &types.Processor{
			Type:       types.ProcessorType(processorType),
			Parameters: expandProcessorParameters(processingConfigurationProcessor[names.AttrParameters].(*schema.Set).List()),
		}
	}
	return processor
}

func expandProcessorParameters(processorParameters []interface{}) []types.ProcessorParameter {
	parameters := []types.ProcessorParameter{}

	for _, attr := range processorParameters {
		parameters = append(parameters, expandProcessorParameter(attr.(map[string]interface{})))
	}

	return parameters
}

func expandProcessorParameter(processorParameter map[string]interface{}) types.ProcessorParameter {
	parameter := types.ProcessorParameter{
		ParameterName:  types.ProcessorParameterName(processorParameter["parameter_name"].(string)),
		ParameterValue: aws.String(processorParameter["parameter_value"].(string)),
	}

	return parameter
}

func expandSecretsManagerConfiguration(tfMap map[string]interface{}) *types.SecretsManagerConfiguration {
	config := tfMap["secrets_manager_configuration"].([]interface{})

	if len(config) == 0 || config[0] == nil {
		return nil
	}

	secretsManagerConfiguration := config[0].(map[string]interface{})
	configuration := &types.SecretsManagerConfiguration{
		Enabled: aws.Bool(secretsManagerConfiguration[names.AttrEnabled].(bool)),
	}

	if v, ok := secretsManagerConfiguration["secret_arn"]; ok && len(v.(string)) > 0 {
		configuration.SecretARN = aws.String(v.(string))
	}

	if v, ok := secretsManagerConfiguration[names.AttrRoleARN]; ok && len(v.(string)) > 0 {
		configuration.RoleARN = aws.String(v.(string))
	}

	return configuration
}

func expandEncryptionConfiguration(s3 map[string]interface{}) *types.EncryptionConfiguration {
	if key, ok := s3[names.AttrKMSKeyARN]; ok && len(key.(string)) > 0 {
		return &types.EncryptionConfiguration{
			KMSEncryptionConfig: &types.KMSEncryptionConfig{
				AWSKMSKeyARN: aws.String(key.(string)),
			},
		}
	}

	return &types.EncryptionConfiguration{
		NoEncryptionConfig: types.NoEncryptionConfigNoEncryption,
	}
}

func expandCloudWatchLoggingOptions(s3 map[string]interface{}) *types.CloudWatchLoggingOptions {
	config := s3["cloudwatch_logging_options"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	loggingConfig := config[0].(map[string]interface{})
	loggingOptions := &types.CloudWatchLoggingOptions{
		Enabled: aws.Bool(loggingConfig[names.AttrEnabled].(bool)),
	}

	if v, ok := loggingConfig[names.AttrLogGroupName]; ok {
		loggingOptions.LogGroupName = aws.String(v.(string))
	}

	if v, ok := loggingConfig["log_stream_name"]; ok {
		loggingOptions.LogStreamName = aws.String(v.(string))
	}

	return loggingOptions
}

func expandVPCConfiguration(es map[string]interface{}) *types.VpcConfiguration {
	config := es[names.AttrVPCConfig].([]interface{})
	if len(config) == 0 {
		return nil
	}

	vpcConfig := config[0].(map[string]interface{})

	return &types.VpcConfiguration{
		RoleARN:          aws.String(vpcConfig[names.AttrRoleARN].(string)),
		SubnetIds:        flex.ExpandStringValueSet(vpcConfig[names.AttrSubnetIDs].(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringValueSet(vpcConfig[names.AttrSecurityGroupIDs].(*schema.Set)),
	}
}

func expandPrefix(s3 map[string]interface{}) *string {
	if v, ok := s3[names.AttrPrefix]; ok {
		return aws.String(v.(string))
	}

	return nil
}

func expandRedshiftDestinationConfiguration(tfMap map[string]interface{}) *types.RedshiftDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.RedshiftDestinationConfiguration{
		ClusterJDBCURL:  aws.String(tfMap["cluster_jdbcurl"].(string)),
		CopyCommand:     expandCopyCommand(tfMap),
		RetryOptions:    expandRedshiftRetryOptions(tfMap),
		RoleARN:         aws.String(roleARN),
		S3Configuration: expandS3DestinationConfiguration(tfMap["s3_configuration"].([]interface{})),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap[names.AttrPassword]; ok && v.(string) != "" {
		apiObject.Password = aws.String(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeRedshift, roleARN)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.RedshiftS3BackupMode(v.(string))
		apiObject.S3BackupConfiguration = expandS3DestinationConfigurationBackup(tfMap)
	}

	if _, ok := tfMap["secrets_manager_configuration"]; ok {
		apiObject.SecretsManagerConfiguration = expandSecretsManagerConfiguration(tfMap)
	}

	if v, ok := tfMap[names.AttrUsername]; ok && v.(string) != "" {
		apiObject.Username = aws.String(v.(string))
	}

	return apiObject
}

func expandRedshiftDestinationUpdate(tfMap map[string]interface{}) *types.RedshiftDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.RedshiftDestinationUpdate{
		ClusterJDBCURL: aws.String(tfMap["cluster_jdbcurl"].(string)),
		CopyCommand:    expandCopyCommand(tfMap),
		RetryOptions:   expandRedshiftRetryOptions(tfMap),
		RoleARN:        aws.String(roleARN),
	}

	s3Config := expandS3DestinationUpdate(tfMap["s3_configuration"].([]interface{}))
	// Redshift does not currently support ErrorOutputPrefix,
	// which is set to the empty string within "updateS3Config",
	// thus we must remove it here to avoid an InvalidArgumentException.
	s3Config.ErrorOutputPrefix = nil
	apiObject.S3Update = s3Config

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap[names.AttrPassword]; ok && v.(string) != "" {
		apiObject.Password = aws.String(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeRedshift, roleARN)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.RedshiftS3BackupMode(v.(string))
		apiObject.S3BackupUpdate = expandS3DestinationUpdateBackup(tfMap)
		if apiObject.S3BackupUpdate != nil {
			// Redshift does not currently support ErrorOutputPrefix,
			// which is set to the empty string within "updateS3BackupConfig",
			// thus we must remove it here to avoid an InvalidArgumentException.
			apiObject.S3BackupUpdate.ErrorOutputPrefix = nil
		}
	}

	if _, ok := tfMap["secrets_manager_configuration"]; ok {
		apiObject.SecretsManagerConfiguration = expandSecretsManagerConfiguration(tfMap)
	}

	if v, ok := tfMap[names.AttrUsername]; ok && v.(string) != "" {
		apiObject.Username = aws.String(v.(string))
	}

	return apiObject
}

func expandElasticsearchDestinationConfiguration(es map[string]interface{}) *types.ElasticsearchDestinationConfiguration {
	roleARN := es[names.AttrRoleARN].(string)
	config := &types.ElasticsearchDestinationConfiguration{
		BufferingHints:  expandElasticsearchBufferingHints(es),
		IndexName:       aws.String(es["index_name"].(string)),
		RetryOptions:    expandElasticsearchRetryOptions(es),
		RoleARN:         aws.String(roleARN),
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
		config.ProcessingConfiguration = expandProcessingConfiguration(es, destinationTypeElasticsearch, roleARN)
	}

	if indexRotationPeriod, ok := es["index_rotation_period"]; ok {
		config.IndexRotationPeriod = types.ElasticsearchIndexRotationPeriod(indexRotationPeriod.(string))
	}
	if s3BackupMode, ok := es["s3_backup_mode"]; ok {
		config.S3BackupMode = types.ElasticsearchS3BackupMode(s3BackupMode.(string))
	}

	if _, ok := es[names.AttrVPCConfig]; ok {
		config.VpcConfiguration = expandVPCConfiguration(es)
	}

	return config
}

func expandElasticsearchDestinationUpdate(es map[string]interface{}) *types.ElasticsearchDestinationUpdate {
	roleARN := es[names.AttrRoleARN].(string)
	update := &types.ElasticsearchDestinationUpdate{
		BufferingHints: expandElasticsearchBufferingHints(es),
		IndexName:      aws.String(es["index_name"].(string)),
		RetryOptions:   expandElasticsearchRetryOptions(es),
		RoleARN:        aws.String(roleARN),
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
		update.ProcessingConfiguration = expandProcessingConfiguration(es, destinationTypeElasticsearch, roleARN)
	}

	if indexRotationPeriod, ok := es["index_rotation_period"]; ok {
		update.IndexRotationPeriod = types.ElasticsearchIndexRotationPeriod(indexRotationPeriod.(string))
	}

	return update
}

func expandAmazonopensearchserviceDestinationConfiguration(os map[string]interface{}) *types.AmazonopensearchserviceDestinationConfiguration {
	roleARN := os[names.AttrRoleARN].(string)
	config := &types.AmazonopensearchserviceDestinationConfiguration{
		BufferingHints:  expandAmazonopensearchserviceBufferingHints(os),
		IndexName:       aws.String(os["index_name"].(string)),
		RetryOptions:    expandAmazonopensearchserviceRetryOptions(os),
		RoleARN:         aws.String(roleARN),
		TypeName:        aws.String(os["type_name"].(string)),
		S3Configuration: expandS3DestinationConfiguration(os["s3_configuration"].([]interface{})),
	}

	if v, ok := os["domain_arn"]; ok && v.(string) != "" {
		config.DomainARN = aws.String(v.(string))
	}

	if v, ok := os["cluster_endpoint"]; ok && v.(string) != "" {
		config.ClusterEndpoint = aws.String(v.(string))
	}

	if _, ok := os["cloudwatch_logging_options"]; ok {
		config.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(os)
	}

	if _, ok := os["processing_configuration"]; ok {
		config.ProcessingConfiguration = expandProcessingConfiguration(os, destinationTypeOpenSearch, roleARN)
	}

	if indexRotationPeriod, ok := os["index_rotation_period"]; ok {
		config.IndexRotationPeriod = types.AmazonopensearchserviceIndexRotationPeriod(indexRotationPeriod.(string))
	}
	if s3BackupMode, ok := os["s3_backup_mode"]; ok {
		config.S3BackupMode = types.AmazonopensearchserviceS3BackupMode(s3BackupMode.(string))
	}

	if _, ok := os[names.AttrVPCConfig]; ok {
		config.VpcConfiguration = expandVPCConfiguration(os)
	}

	if v, ok := os["document_id_options"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		config.DocumentIdOptions = expandDocumentIDOptions(v[0].(map[string]interface{}))
	}

	return config
}

func expandAmazonopensearchserviceDestinationUpdate(os map[string]interface{}) *types.AmazonopensearchserviceDestinationUpdate {
	roleARN := os[names.AttrRoleARN].(string)
	update := &types.AmazonopensearchserviceDestinationUpdate{
		BufferingHints: expandAmazonopensearchserviceBufferingHints(os),
		IndexName:      aws.String(os["index_name"].(string)),
		RetryOptions:   expandAmazonopensearchserviceRetryOptions(os),
		RoleARN:        aws.String(roleARN),
		TypeName:       aws.String(os["type_name"].(string)),
		S3Update:       expandS3DestinationUpdate(os["s3_configuration"].([]interface{})),
	}

	if v, ok := os["domain_arn"]; ok && v.(string) != "" {
		update.DomainARN = aws.String(v.(string))
	}

	if v, ok := os["cluster_endpoint"]; ok && v.(string) != "" {
		update.ClusterEndpoint = aws.String(v.(string))
	}

	if _, ok := os["cloudwatch_logging_options"]; ok {
		update.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(os)
	}

	if _, ok := os["processing_configuration"]; ok {
		update.ProcessingConfiguration = expandProcessingConfiguration(os, destinationTypeOpenSearch, roleARN)
	}

	if indexRotationPeriod, ok := os["index_rotation_period"]; ok {
		update.IndexRotationPeriod = types.AmazonopensearchserviceIndexRotationPeriod(indexRotationPeriod.(string))
	}

	if v, ok := os["document_id_options"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		update.DocumentIdOptions = expandDocumentIDOptions(v[0].(map[string]interface{}))
	}

	return update
}

func expandAmazonOpenSearchServerlessDestinationConfiguration(oss map[string]interface{}) *types.AmazonOpenSearchServerlessDestinationConfiguration {
	roleARN := oss[names.AttrRoleARN].(string)
	config := &types.AmazonOpenSearchServerlessDestinationConfiguration{
		BufferingHints:  expandAmazonOpenSearchServerlessBufferingHints(oss),
		IndexName:       aws.String(oss["index_name"].(string)),
		RetryOptions:    expandAmazonOpenSearchServerlessRetryOptions(oss),
		RoleARN:         aws.String(roleARN),
		S3Configuration: expandS3DestinationConfiguration(oss["s3_configuration"].([]interface{})),
	}

	if v, ok := oss["collection_endpoint"]; ok && v.(string) != "" {
		config.CollectionEndpoint = aws.String(v.(string))
	}

	if _, ok := oss["cloudwatch_logging_options"]; ok {
		config.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(oss)
	}

	if _, ok := oss["processing_configuration"]; ok {
		config.ProcessingConfiguration = expandProcessingConfiguration(oss, destinationTypeOpenSearchServerless, roleARN)
	}

	if s3BackupMode, ok := oss["s3_backup_mode"]; ok {
		config.S3BackupMode = types.AmazonOpenSearchServerlessS3BackupMode(s3BackupMode.(string))
	}

	if _, ok := oss[names.AttrVPCConfig]; ok {
		config.VpcConfiguration = expandVPCConfiguration(oss)
	}

	return config
}

func expandAmazonOpenSearchServerlessDestinationUpdate(oss map[string]interface{}) *types.AmazonOpenSearchServerlessDestinationUpdate {
	roleARN := oss[names.AttrRoleARN].(string)
	update := &types.AmazonOpenSearchServerlessDestinationUpdate{
		BufferingHints: expandAmazonOpenSearchServerlessBufferingHints(oss),
		IndexName:      aws.String(oss["index_name"].(string)),
		RetryOptions:   expandAmazonOpenSearchServerlessRetryOptions(oss),
		RoleARN:        aws.String(roleARN),
		S3Update:       expandS3DestinationUpdate(oss["s3_configuration"].([]interface{})),
	}
	if v, ok := oss["collection_endpoint"]; ok && v.(string) != "" {
		update.CollectionEndpoint = aws.String(v.(string))
	}

	if _, ok := oss["cloudwatch_logging_options"]; ok {
		update.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(oss)
	}

	if _, ok := oss["processing_configuration"]; ok {
		update.ProcessingConfiguration = expandProcessingConfiguration(oss, destinationTypeOpenSearchServerless, roleARN)
	}

	return update
}

func expandSnowflakeDestinationConfiguration(tfMap map[string]interface{}) *types.SnowflakeDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.SnowflakeDestinationConfiguration{
		AccountUrl:                aws.String(tfMap["account_url"].(string)),
		Database:                  aws.String(tfMap[names.AttrDatabase].(string)),
		RetryOptions:              expandSnowflakeRetryOptions(tfMap),
		RoleARN:                   aws.String(roleARN),
		S3Configuration:           expandS3DestinationConfiguration(tfMap["s3_configuration"].([]interface{})),
		Schema:                    aws.String(tfMap[names.AttrSchema].(string)),
		SnowflakeVpcConfiguration: expandSnowflakeVPCConfiguration(tfMap),
		Table:                     aws.String(tfMap["table"].(string)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["content_column_name"]; ok && v.(string) != "" {
		apiObject.ContentColumnName = aws.String(v.(string))
	}

	if v, ok := tfMap["data_loading_option"]; ok && v.(string) != "" {
		apiObject.DataLoadingOption = types.SnowflakeDataLoadingOption(v.(string))
	}

	if v, ok := tfMap[names.AttrPrivateKey]; ok && v.(string) != "" {
		apiObject.PrivateKey = aws.String(v.(string))
	}

	if v, ok := tfMap["key_passphrase"]; ok && v.(string) != "" {
		apiObject.KeyPassphrase = aws.String(v.(string))
	}

	if v, ok := tfMap["metadata_column_name"]; ok && v.(string) != "" {
		apiObject.MetaDataColumnName = aws.String(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeSnowflake, roleARN)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.SnowflakeS3BackupMode(v.(string))
	}

	if _, ok := tfMap["secrets_manager_configuration"]; ok {
		apiObject.SecretsManagerConfiguration = expandSecretsManagerConfiguration(tfMap)
	}

	if _, ok := tfMap["snowflake_role_configuration"]; ok {
		apiObject.SnowflakeRoleConfiguration = expandSnowflakeRoleConfiguration(tfMap)
	}

	if _, ok := tfMap["snowflake_vpc_configuration"]; ok {
		apiObject.SnowflakeVpcConfiguration = expandSnowflakeVPCConfiguration(tfMap)
	}

	if v, ok := tfMap["user"]; ok && v.(string) != "" {
		apiObject.User = aws.String(v.(string))
	}

	return apiObject
}

func expandSnowflakeDestinationUpdate(tfMap map[string]interface{}) *types.SnowflakeDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.SnowflakeDestinationUpdate{
		AccountUrl:   aws.String(tfMap["account_url"].(string)),
		Database:     aws.String(tfMap[names.AttrDatabase].(string)),
		RetryOptions: expandSnowflakeRetryOptions(tfMap),
		RoleARN:      aws.String(roleARN),
		S3Update:     expandS3DestinationUpdate(tfMap["s3_configuration"].([]interface{})),
		Schema:       aws.String(tfMap[names.AttrSchema].(string)),
		Table:        aws.String(tfMap["table"].(string)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["content_column_name"]; ok && v.(string) != "" {
		apiObject.ContentColumnName = aws.String(v.(string))
	}

	if v, ok := tfMap["data_loading_option"]; ok && v.(string) != "" {
		apiObject.DataLoadingOption = types.SnowflakeDataLoadingOption(v.(string))
	}

	if v, ok := tfMap[names.AttrPrivateKey]; ok && v.(string) != "" {
		apiObject.PrivateKey = aws.String(v.(string))
	}

	if v, ok := tfMap["key_passphrase"]; ok && v.(string) != "" {
		apiObject.KeyPassphrase = aws.String(v.(string))
	}

	if v, ok := tfMap["metadata_column_name"]; ok && v.(string) != "" {
		apiObject.MetaDataColumnName = aws.String(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeSnowflake, roleARN)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.SnowflakeS3BackupMode(v.(string))
	}

	if _, ok := tfMap["secrets_manager_configuration"]; ok {
		apiObject.SecretsManagerConfiguration = expandSecretsManagerConfiguration(tfMap)
	}

	if _, ok := tfMap["snowflake_role_configuration"]; ok {
		apiObject.SnowflakeRoleConfiguration = expandSnowflakeRoleConfiguration(tfMap)
	}

	if v, ok := tfMap["user"]; ok && v.(string) != "" {
		apiObject.User = aws.String(v.(string))
	}

	return apiObject
}

func expandSplunkDestinationConfiguration(tfMap map[string]interface{}) *types.SplunkDestinationConfiguration {
	apiObject := &types.SplunkDestinationConfiguration{
		HECAcknowledgmentTimeoutInSeconds: aws.Int32(int32(tfMap["hec_acknowledgment_timeout"].(int))),
		HECEndpoint:                       aws.String(tfMap["hec_endpoint"].(string)),
		HECEndpointType:                   types.HECEndpointType(tfMap["hec_endpoint_type"].(string)),
		RetryOptions:                      expandSplunkRetryOptions(tfMap),
		S3Configuration:                   expandS3DestinationConfiguration(tfMap["s3_configuration"].([]interface{})),
	}

	bufferingHints := &types.SplunkBufferingHints{}
	if bufferingInterval, ok := tfMap["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int32(int32(bufferingInterval))
	}
	if bufferingSize, ok := tfMap["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int32(int32(bufferingSize))
	}
	apiObject.BufferingHints = bufferingHints

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["hec_token"]; ok && v.(string) != "" {
		apiObject.HECToken = aws.String(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeSplunk, "")
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.SplunkS3BackupMode(v.(string))
	}

	if _, ok := tfMap["secrets_manager_configuration"]; ok {
		apiObject.SecretsManagerConfiguration = expandSecretsManagerConfiguration(tfMap)
	}

	return apiObject
}

func expandSplunkDestinationUpdate(tfMap map[string]interface{}) *types.SplunkDestinationUpdate {
	apiObject := &types.SplunkDestinationUpdate{
		HECAcknowledgmentTimeoutInSeconds: aws.Int32(int32(tfMap["hec_acknowledgment_timeout"].(int))),
		HECEndpoint:                       aws.String(tfMap["hec_endpoint"].(string)),
		HECEndpointType:                   types.HECEndpointType(tfMap["hec_endpoint_type"].(string)),
		RetryOptions:                      expandSplunkRetryOptions(tfMap),
		S3Update:                          expandS3DestinationUpdate(tfMap["s3_configuration"].([]interface{})),
	}

	bufferingHints := &types.SplunkBufferingHints{}
	if bufferingInterval, ok := tfMap["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int32(int32(bufferingInterval))
	}
	if bufferingSize, ok := tfMap["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int32(int32(bufferingSize))
	}
	apiObject.BufferingHints = bufferingHints

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["hec_token"]; ok && v.(string) != "" {
		apiObject.HECToken = aws.String(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeSplunk, "")
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.SplunkS3BackupMode(v.(string))
	}

	if _, ok := tfMap["secrets_manager_configuration"]; ok {
		apiObject.SecretsManagerConfiguration = expandSecretsManagerConfiguration(tfMap)
	}

	return apiObject
}

func expandHTTPEndpointDestinationConfiguration(tfMap map[string]interface{}) *types.HttpEndpointDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.HttpEndpointDestinationConfiguration{
		EndpointConfiguration: expandHTTPEndpointConfiguration(tfMap),
		RetryOptions:          expandHTTPEndpointRetryOptions(tfMap),
		RoleARN:               aws.String(roleARN),
		S3Configuration:       expandS3DestinationConfiguration(tfMap["s3_configuration"].([]interface{})),
	}

	bufferingHints := &types.HttpEndpointBufferingHints{}
	if v, ok := tfMap["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int32(int32(v))
	}
	if v, ok := tfMap["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int32(int32(v))
	}
	apiObject.BufferingHints = bufferingHints

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeHTTPEndpoint, roleARN)
	}

	if _, ok := tfMap["request_configuration"]; ok {
		apiObject.RequestConfiguration = expandHTTPEndpointRequestConfiguration(tfMap)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.HttpEndpointS3BackupMode(v.(string))
	}

	if _, ok := tfMap["secrets_manager_configuration"]; ok {
		apiObject.SecretsManagerConfiguration = expandSecretsManagerConfiguration(tfMap)
	}

	return apiObject
}

func expandHTTPEndpointDestinationUpdate(tfMap map[string]interface{}) *types.HttpEndpointDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.HttpEndpointDestinationUpdate{
		EndpointConfiguration: expandHTTPEndpointConfiguration(tfMap),
		RetryOptions:          expandHTTPEndpointRetryOptions(tfMap),
		RoleARN:               aws.String(roleARN),
		S3Update:              expandS3DestinationUpdate(tfMap["s3_configuration"].([]interface{})),
	}

	bufferingHints := &types.HttpEndpointBufferingHints{}
	if v, ok := tfMap["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int32(int32(v))
	}
	if v, ok := tfMap["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int32(int32(v))
	}
	apiObject.BufferingHints = bufferingHints

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeHTTPEndpoint, roleARN)
	}

	if _, ok := tfMap["request_configuration"]; ok {
		apiObject.RequestConfiguration = expandHTTPEndpointRequestConfiguration(tfMap)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.HttpEndpointS3BackupMode(v.(string))
	}

	if _, ok := tfMap["secrets_manager_configuration"]; ok {
		apiObject.SecretsManagerConfiguration = expandSecretsManagerConfiguration(tfMap)
	}

	return apiObject
}

func expandHTTPEndpointCommonAttributes(ca []interface{}) []types.HttpEndpointCommonAttribute {
	commonAttributes := make([]types.HttpEndpointCommonAttribute, 0, len(ca))

	for _, raw := range ca {
		data := raw.(map[string]interface{})

		a := types.HttpEndpointCommonAttribute{
			AttributeName:  aws.String(data[names.AttrName].(string)),
			AttributeValue: aws.String(data[names.AttrValue].(string)),
		}
		commonAttributes = append(commonAttributes, a)
	}

	return commonAttributes
}

func expandHTTPEndpointRequestConfiguration(rc map[string]interface{}) *types.HttpEndpointRequestConfiguration {
	config := rc["request_configuration"].([]interface{})
	if len(config) == 0 {
		return nil
	}

	requestConfig := config[0].(map[string]interface{})
	RequestConfiguration := &types.HttpEndpointRequestConfiguration{}

	if contentEncoding, ok := requestConfig["content_encoding"]; ok {
		RequestConfiguration.ContentEncoding = types.ContentEncoding(contentEncoding.(string))
	}

	if commonAttributes, ok := requestConfig["common_attributes"]; ok {
		RequestConfiguration.CommonAttributes = expandHTTPEndpointCommonAttributes(commonAttributes.([]interface{}))
	}

	return RequestConfiguration
}

func expandHTTPEndpointConfiguration(ep map[string]interface{}) *types.HttpEndpointConfiguration {
	endpointConfiguration := &types.HttpEndpointConfiguration{
		Url: aws.String(ep[names.AttrURL].(string)),
	}

	if Name, ok := ep[names.AttrName]; ok {
		endpointConfiguration.Name = aws.String(Name.(string))
	}

	if AccessKey, ok := ep[names.AttrAccessKey]; ok {
		endpointConfiguration.AccessKey = aws.String(AccessKey.(string))
	}

	return endpointConfiguration
}

func expandElasticsearchBufferingHints(es map[string]interface{}) *types.ElasticsearchBufferingHints {
	bufferingHints := &types.ElasticsearchBufferingHints{}

	if bufferingInterval, ok := es["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int32(int32(bufferingInterval))
	}
	if bufferingSize, ok := es["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int32(int32(bufferingSize))
	}

	return bufferingHints
}

func expandAmazonopensearchserviceBufferingHints(es map[string]interface{}) *types.AmazonopensearchserviceBufferingHints {
	bufferingHints := &types.AmazonopensearchserviceBufferingHints{}

	if bufferingInterval, ok := es["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int32(int32(bufferingInterval))
	}
	if bufferingSize, ok := es["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int32(int32(bufferingSize))
	}

	return bufferingHints
}

func expandAmazonOpenSearchServerlessBufferingHints(es map[string]interface{}) *types.AmazonOpenSearchServerlessBufferingHints {
	bufferingHints := &types.AmazonOpenSearchServerlessBufferingHints{}

	if bufferingInterval, ok := es["buffering_interval"].(int); ok {
		bufferingHints.IntervalInSeconds = aws.Int32(int32(bufferingInterval))
	}
	if bufferingSize, ok := es["buffering_size"].(int); ok {
		bufferingHints.SizeInMBs = aws.Int32(int32(bufferingSize))
	}

	return bufferingHints
}

func expandElasticsearchRetryOptions(es map[string]interface{}) *types.ElasticsearchRetryOptions {
	retryOptions := &types.ElasticsearchRetryOptions{}

	if retryDuration, ok := es["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int32(int32(retryDuration))
	}

	return retryOptions
}

func expandAmazonopensearchserviceRetryOptions(es map[string]interface{}) *types.AmazonopensearchserviceRetryOptions {
	retryOptions := &types.AmazonopensearchserviceRetryOptions{}

	if retryDuration, ok := es["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int32(int32(retryDuration))
	}

	return retryOptions
}

func expandAmazonOpenSearchServerlessRetryOptions(es map[string]interface{}) *types.AmazonOpenSearchServerlessRetryOptions {
	retryOptions := &types.AmazonOpenSearchServerlessRetryOptions{}

	if retryDuration, ok := es["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int32(int32(retryDuration))
	}

	return retryOptions
}

func expandHTTPEndpointRetryOptions(tfMap map[string]interface{}) *types.HttpEndpointRetryOptions {
	retryOptions := &types.HttpEndpointRetryOptions{}

	if retryDuration, ok := tfMap["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int32(int32(retryDuration))
	}

	return retryOptions
}

func expandRedshiftRetryOptions(redshift map[string]interface{}) *types.RedshiftRetryOptions {
	retryOptions := &types.RedshiftRetryOptions{}

	if retryDuration, ok := redshift["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int32(int32(retryDuration))
	}

	return retryOptions
}

func expandSnowflakeRetryOptions(tfMap map[string]interface{}) *types.SnowflakeRetryOptions {
	apiObject := &types.SnowflakeRetryOptions{}

	if v, ok := tfMap["retry_duration"].(int); ok {
		apiObject.DurationInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandSnowflakeRoleConfiguration(tfMap map[string]interface{}) *types.SnowflakeRoleConfiguration {
	config := tfMap["snowflake_role_configuration"].([]interface{})
	if len(config) == 0 || config[0] == nil {
		// It is possible to just pass nil here, but this seems to be the
		// canonical form that AWS uses, and is less likely to produce diffs.
		return &types.SnowflakeRoleConfiguration{
			Enabled: aws.Bool(false),
		}
	}

	snowflakeRoleConfiguration := config[0].(map[string]interface{})
	apiObject := &types.SnowflakeRoleConfiguration{
		Enabled: aws.Bool(snowflakeRoleConfiguration[names.AttrEnabled].(bool)),
	}

	if v, ok := snowflakeRoleConfiguration["snowflake_role"]; ok && len(v.(string)) > 0 {
		apiObject.SnowflakeRole = aws.String(v.(string))
	}

	return apiObject
}

func expandSnowflakeVPCConfiguration(tfMap map[string]interface{}) *types.SnowflakeVpcConfiguration {
	tfList := tfMap["snowflake_vpc_configuration"].([]interface{})
	if len(tfList) == 0 {
		return nil
	}

	tfMap = tfList[0].(map[string]interface{})

	apiObject := &types.SnowflakeVpcConfiguration{
		PrivateLinkVpceId: aws.String(tfMap["private_link_vpce_id"].(string)),
	}

	return apiObject
}

func expandSplunkRetryOptions(splunk map[string]interface{}) *types.SplunkRetryOptions {
	retryOptions := &types.SplunkRetryOptions{}

	if retryDuration, ok := splunk["retry_duration"].(int); ok {
		retryOptions.DurationInSeconds = aws.Int32(int32(retryDuration))
	}

	return retryOptions
}

func expandCopyCommand(redshift map[string]interface{}) *types.CopyCommand {
	cmd := &types.CopyCommand{
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

func expandDeliveryStreamEncryptionConfigurationInput(tfList []interface{}) *types.DeliveryStreamEncryptionConfigurationInput {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &types.DeliveryStreamEncryptionConfigurationInput{}

	if v, ok := tfMap["key_arn"].(string); ok && v != "" {
		apiObject.KeyARN = aws.String(v)
	}

	if v, ok := tfMap["key_type"].(string); ok && v != "" {
		apiObject.KeyType = types.KeyType(v)
	}

	return apiObject
}

func expandMSKSourceConfiguration(tfMap map[string]interface{}) *types.MSKSourceConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.MSKSourceConfiguration{}

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

func expandAuthenticationConfiguration(tfMap map[string]interface{}) *types.AuthenticationConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AuthenticationConfiguration{}

	if v, ok := tfMap["connectivity"].(string); ok && v != "" {
		apiObject.Connectivity = types.Connectivity(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleARN = aws.String(v)
	}

	return apiObject
}

func flattenMSKSourceDescription(apiObject *types.MSKSourceDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AuthenticationConfiguration; v != nil {
		tfMap["authentication_configuration"] = []interface{}{flattenAuthenticationConfiguration(v)}
	}

	if v := apiObject.MSKClusterARN; v != nil {
		tfMap["msk_cluster_arn"] = aws.ToString(v)
	}

	if v := apiObject.TopicName; v != nil {
		tfMap["topic_name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenAuthenticationConfiguration(apiObject *types.AuthenticationConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"connectivity": apiObject.Connectivity,
	}

	if v := apiObject.RoleARN; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return tfMap
}

func flattenCloudWatchLoggingOptions(clo *types.CloudWatchLoggingOptions) []interface{} {
	if clo == nil {
		return []interface{}{}
	}

	cloudwatchLoggingOptions := map[string]interface{}{
		names.AttrEnabled: aws.ToBool(clo.Enabled),
	}
	if aws.ToBool(clo.Enabled) {
		cloudwatchLoggingOptions[names.AttrLogGroupName] = aws.ToString(clo.LogGroupName)
		cloudwatchLoggingOptions["log_stream_name"] = aws.ToString(clo.LogStreamName)
	}
	return []interface{}{cloudwatchLoggingOptions}
}

func flattenElasticsearchDestinationDescription(description *types.ElasticsearchDestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		names.AttrRoleARN:            aws.ToString(description.RoleARN),
		"type_name":                  aws.ToString(description.TypeName),
		"index_name":                 aws.ToString(description.IndexName),
		"s3_backup_mode":             description.S3BackupMode,
		"s3_configuration":           flattenS3DestinationDescription(description.S3DestinationDescription),
		"index_rotation_period":      description.IndexRotationPeriod,
		names.AttrVPCConfig:          flattenVPCConfigurationDescription(description.VpcConfigurationDescription),
		"processing_configuration":   flattenProcessingConfiguration(description.ProcessingConfiguration, destinationTypeElasticsearch, aws.ToString(description.RoleARN)),
	}

	if description.DomainARN != nil {
		m["domain_arn"] = aws.ToString(description.DomainARN)
	}

	if description.ClusterEndpoint != nil {
		m["cluster_endpoint"] = aws.ToString(description.ClusterEndpoint)
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.ToInt32(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.ToInt32(description.BufferingHints.SizeInMBs))
	}

	if description.RetryOptions != nil {
		m["retry_duration"] = int(aws.ToInt32(description.RetryOptions.DurationInSeconds))
	}

	return []map[string]interface{}{m}
}

func flattenAmazonopensearchserviceDestinationDescription(description *types.AmazonopensearchserviceDestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		names.AttrRoleARN:            aws.ToString(description.RoleARN),
		"type_name":                  aws.ToString(description.TypeName),
		"index_name":                 aws.ToString(description.IndexName),
		"s3_backup_mode":             description.S3BackupMode,
		"s3_configuration":           flattenS3DestinationDescription(description.S3DestinationDescription),
		"index_rotation_period":      description.IndexRotationPeriod,
		names.AttrVPCConfig:          flattenVPCConfigurationDescription(description.VpcConfigurationDescription),
		"processing_configuration":   flattenProcessingConfiguration(description.ProcessingConfiguration, destinationTypeOpenSearch, aws.ToString(description.RoleARN)),
	}

	if description.DomainARN != nil {
		m["domain_arn"] = aws.ToString(description.DomainARN)
	}

	if description.ClusterEndpoint != nil {
		m["cluster_endpoint"] = aws.ToString(description.ClusterEndpoint)
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.ToInt32(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.ToInt32(description.BufferingHints.SizeInMBs))
	}

	if description.RetryOptions != nil {
		m["retry_duration"] = int(aws.ToInt32(description.RetryOptions.DurationInSeconds))
	}

	if v := description.DocumentIdOptions; v != nil {
		m["document_id_options"] = []interface{}{flattenDocumentIDOptions(v)}
	}

	return []map[string]interface{}{m}
}

func flattenAmazonOpenSearchServerlessDestinationDescription(description *types.AmazonOpenSearchServerlessDestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		names.AttrRoleARN:            aws.ToString(description.RoleARN),
		"index_name":                 aws.ToString(description.IndexName),
		"s3_backup_mode":             description.S3BackupMode,
		"s3_configuration":           flattenS3DestinationDescription(description.S3DestinationDescription),
		names.AttrVPCConfig:          flattenVPCConfigurationDescription(description.VpcConfigurationDescription),
		"processing_configuration":   flattenProcessingConfiguration(description.ProcessingConfiguration, destinationTypeOpenSearchServerless, aws.ToString(description.RoleARN)),
	}

	if description.CollectionEndpoint != nil {
		m["collection_endpoint"] = aws.ToString(description.CollectionEndpoint)
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.ToInt32(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.ToInt32(description.BufferingHints.SizeInMBs))
	}

	if description.RetryOptions != nil {
		m["retry_duration"] = int(aws.ToInt32(description.RetryOptions.DurationInSeconds))
	}

	return []map[string]interface{}{m}
}

func flattenVPCConfigurationDescription(description *types.VpcConfigurationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrVPCID:            aws.ToString(description.VpcId),
		names.AttrSubnetIDs:        description.SubnetIds,
		names.AttrSecurityGroupIDs: description.SecurityGroupIds,
		names.AttrRoleARN:          aws.ToString(description.RoleARN),
	}

	return []map[string]interface{}{m}
}

func flattenExtendedS3DestinationDescription(description *types.ExtendedS3DestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"bucket_arn":                           aws.ToString(description.BucketARN),
		"cloudwatch_logging_options":           flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"compression_format":                   description.CompressionFormat,
		"custom_time_zone":                     aws.ToString(description.CustomTimeZone),
		"data_format_conversion_configuration": flattenDataFormatConversionConfiguration(description.DataFormatConversionConfiguration),
		"error_output_prefix":                  aws.ToString(description.ErrorOutputPrefix),
		"file_extension":                       aws.ToString(description.FileExtension),
		names.AttrPrefix:                       aws.ToString(description.Prefix),
		"processing_configuration":             flattenProcessingConfiguration(description.ProcessingConfiguration, destinationTypeExtendedS3, aws.ToString(description.RoleARN)),
		"dynamic_partitioning_configuration":   flattenDynamicPartitioningConfiguration(description.DynamicPartitioningConfiguration),
		names.AttrRoleARN:                      aws.ToString(description.RoleARN),
		"s3_backup_configuration":              flattenS3DestinationDescription(description.S3BackupDescription),
		"s3_backup_mode":                       description.S3BackupMode,
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.ToInt32(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.ToInt32(description.BufferingHints.SizeInMBs))
	}

	if description.EncryptionConfiguration != nil && description.EncryptionConfiguration.KMSEncryptionConfig != nil {
		m[names.AttrKMSKeyARN] = aws.ToString(description.EncryptionConfiguration.KMSEncryptionConfig.AWSKMSKeyARN)
	}

	return []map[string]interface{}{m}
}

func flattenRedshiftDestinationDescription(apiObject *types.RedshiftDestinationDescription, configuredPassword string) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"cloudwatch_logging_options":    flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		"cluster_jdbcurl":               aws.ToString(apiObject.ClusterJDBCURL),
		names.AttrPassword:              configuredPassword,
		"processing_configuration":      flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeRedshift, aws.ToString(apiObject.RoleARN)),
		names.AttrRoleARN:               aws.ToString(apiObject.RoleARN),
		"s3_backup_configuration":       flattenS3DestinationDescription(apiObject.S3BackupDescription),
		"s3_backup_mode":                apiObject.S3BackupMode,
		"s3_configuration":              flattenS3DestinationDescription(apiObject.S3DestinationDescription),
		"secrets_manager_configuration": flattenSecretsManagerConfiguration(apiObject.SecretsManagerConfiguration),
		names.AttrUsername:              aws.ToString(apiObject.Username),
	}

	if apiObject.CopyCommand != nil {
		tfMap["copy_options"] = aws.ToString(apiObject.CopyCommand.CopyOptions)
		tfMap["data_table_columns"] = aws.ToString(apiObject.CopyCommand.DataTableColumns)
		tfMap["data_table_name"] = aws.ToString(apiObject.CopyCommand.DataTableName)
	}

	if apiObject.RetryOptions != nil {
		tfMap["retry_duration"] = aws.ToInt32(apiObject.RetryOptions.DurationInSeconds)
	}

	return []interface{}{tfMap}
}

func flattenSnowflakeDestinationDescription(apiObject *types.SnowflakeDestinationDescription, configuredKeyPassphrase, configuredPrivateKey string) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	roleARN := aws.ToString(apiObject.RoleARN)
	tfMap := map[string]interface{}{
		"account_url":                   aws.ToString(apiObject.AccountUrl),
		"cloudwatch_logging_options":    flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		"content_column_name":           aws.ToString(apiObject.ContentColumnName),
		"data_loading_option":           apiObject.DataLoadingOption,
		names.AttrDatabase:              aws.ToString(apiObject.Database),
		"key_passphrase":                configuredKeyPassphrase,
		"metadata_column_name":          aws.ToString(apiObject.MetaDataColumnName),
		names.AttrPrivateKey:            configuredPrivateKey,
		"processing_configuration":      flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeSnowflake, roleARN),
		names.AttrRoleARN:               roleARN,
		"s3_backup_mode":                apiObject.S3BackupMode,
		"s3_configuration":              flattenS3DestinationDescription(apiObject.S3DestinationDescription),
		names.AttrSchema:                aws.ToString(apiObject.Schema),
		"secrets_manager_configuration": flattenSecretsManagerConfiguration(apiObject.SecretsManagerConfiguration),
		"snowflake_role_configuration":  flattenSnowflakeRoleConfiguration(apiObject.SnowflakeRoleConfiguration),
		"snowflake_vpc_configuration":   flattenSnowflakeVPCConfiguration(apiObject.SnowflakeVpcConfiguration),
		"table":                         aws.ToString(apiObject.Table),
		"user":                          aws.ToString(apiObject.User),
	}

	if apiObject.RetryOptions != nil {
		tfMap["retry_duration"] = int(aws.ToInt32(apiObject.RetryOptions.DurationInSeconds))
	}

	return []interface{}{tfMap}
}

func flattenSplunkDestinationDescription(apiObject *types.SplunkDestinationDescription) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"cloudwatch_logging_options":    flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		"hec_acknowledgment_timeout":    int(aws.ToInt32(apiObject.HECAcknowledgmentTimeoutInSeconds)),
		"hec_endpoint":                  aws.ToString(apiObject.HECEndpoint),
		"hec_endpoint_type":             apiObject.HECEndpointType,
		"hec_token":                     aws.ToString(apiObject.HECToken),
		"processing_configuration":      flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeSplunk, ""),
		"s3_backup_mode":                apiObject.S3BackupMode,
		"s3_configuration":              flattenS3DestinationDescription(apiObject.S3DestinationDescription),
		"secrets_manager_configuration": flattenSecretsManagerConfiguration(apiObject.SecretsManagerConfiguration),
	}

	if apiObject.BufferingHints != nil {
		tfMap["buffering_interval"] = int(aws.ToInt32(apiObject.BufferingHints.IntervalInSeconds))
		tfMap["buffering_size"] = int(aws.ToInt32(apiObject.BufferingHints.SizeInMBs))
	}

	if apiObject.RetryOptions != nil {
		tfMap["retry_duration"] = int(aws.ToInt32(apiObject.RetryOptions.DurationInSeconds))
	}

	return []interface{}{tfMap}
}

func flattenS3DestinationDescription(description *types.S3DestinationDescription) []map[string]interface{} {
	if description == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"bucket_arn":                 aws.ToString(description.BucketARN),
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(description.CloudWatchLoggingOptions),
		"compression_format":         description.CompressionFormat,
		"error_output_prefix":        aws.ToString(description.ErrorOutputPrefix),
		names.AttrPrefix:             aws.ToString(description.Prefix),
		names.AttrRoleARN:            aws.ToString(description.RoleARN),
	}

	if description.BufferingHints != nil {
		m["buffering_interval"] = int(aws.ToInt32(description.BufferingHints.IntervalInSeconds))
		m["buffering_size"] = int(aws.ToInt32(description.BufferingHints.SizeInMBs))
	}

	if description.EncryptionConfiguration != nil && description.EncryptionConfiguration.KMSEncryptionConfig != nil {
		m[names.AttrKMSKeyARN] = aws.ToString(description.EncryptionConfiguration.KMSEncryptionConfig.AWSKMSKeyARN)
	}

	return []map[string]interface{}{m}
}

func flattenDataFormatConversionConfiguration(dfcc *types.DataFormatConversionConfiguration) []map[string]interface{} {
	if dfcc == nil {
		return []map[string]interface{}{}
	}

	enabled := aws.ToBool(dfcc.Enabled)
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
		names.AttrEnabled:             enabled,
		"input_format_configuration":  ifc,
		"output_format_configuration": ofc,
		"schema_configuration":        sc,
	}

	return []map[string]interface{}{m}
}

func flattenInputFormatConfiguration(ifc *types.InputFormatConfiguration) []map[string]interface{} {
	if ifc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"deserializer": flattenDeserializer(ifc.Deserializer),
	}

	return []map[string]interface{}{m}
}

func flattenDeserializer(deserializer *types.Deserializer) []map[string]interface{} {
	if deserializer == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"hive_json_ser_de":   flattenHiveJSONSerDe(deserializer.HiveJsonSerDe),
		"open_x_json_ser_de": flattenOpenXJSONSerDe(deserializer.OpenXJsonSerDe),
	}

	return []map[string]interface{}{m}
}

func flattenHiveJSONSerDe(hjsd *types.HiveJsonSerDe) []map[string]interface{} {
	if hjsd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"timestamp_formats": hjsd.TimestampFormats,
	}

	return []map[string]interface{}{m}
}

func flattenOpenXJSONSerDe(oxjsd *types.OpenXJsonSerDe) []map[string]interface{} {
	if oxjsd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"column_to_json_key_mappings":              oxjsd.ColumnToJsonKeyMappings,
		"convert_dots_in_json_keys_to_underscores": aws.ToBool(oxjsd.ConvertDotsInJsonKeysToUnderscores),
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference

	m["case_insensitive"] = true
	if oxjsd.CaseInsensitive != nil {
		m["case_insensitive"] = aws.ToBool(oxjsd.CaseInsensitive)
	}

	return []map[string]interface{}{m}
}

func flattenOutputFormatConfiguration(ofc *types.OutputFormatConfiguration) []map[string]interface{} {
	if ofc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"serializer": flattenSerializer(ofc.Serializer),
	}

	return []map[string]interface{}{m}
}

func flattenSerializer(serializer *types.Serializer) []map[string]interface{} {
	if serializer == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"orc_ser_de":     flattenOrcSerDe(serializer.OrcSerDe),
		"parquet_ser_de": flattenParquetSerDe(serializer.ParquetSerDe),
	}

	return []map[string]interface{}{m}
}

func flattenOrcSerDe(osd *types.OrcSerDe) []map[string]interface{} {
	if osd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"bloom_filter_columns":     osd.BloomFilterColumns,
		"dictionary_key_threshold": aws.ToFloat64(osd.DictionaryKeyThreshold),
		"enable_padding":           aws.ToBool(osd.EnablePadding),
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference

	m["block_size_bytes"] = 268435456
	if osd.BlockSizeBytes != nil {
		m["block_size_bytes"] = int(aws.ToInt32(osd.BlockSizeBytes))
	}

	m["bloom_filter_false_positive_probability"] = 0.05
	if osd.BloomFilterFalsePositiveProbability != nil {
		m["bloom_filter_false_positive_probability"] = aws.ToFloat64(osd.BloomFilterFalsePositiveProbability)
	}

	m["compression"] = types.OrcCompressionSnappy
	if osd.Compression != "" {
		m["compression"] = osd.Compression
	}

	m["format_version"] = types.OrcFormatVersionV012
	if osd.FormatVersion != "" {
		m["format_version"] = osd.FormatVersion
	}

	m["padding_tolerance"] = 0.05
	if osd.PaddingTolerance != nil {
		m["padding_tolerance"] = aws.ToFloat64(osd.PaddingTolerance)
	}

	m["row_index_stride"] = 10000
	if osd.RowIndexStride != nil {
		m["row_index_stride"] = int(aws.ToInt32(osd.RowIndexStride))
	}

	m["stripe_size_bytes"] = 67108864
	if osd.StripeSizeBytes != nil {
		m["stripe_size_bytes"] = int(aws.ToInt32(osd.StripeSizeBytes))
	}

	return []map[string]interface{}{m}
}

func flattenParquetSerDe(psd *types.ParquetSerDe) []map[string]interface{} {
	if psd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enable_dictionary_compression": aws.ToBool(psd.EnableDictionaryCompression),
		"max_padding_bytes":             int(aws.ToInt32(psd.MaxPaddingBytes)),
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference

	m["block_size_bytes"] = 268435456
	if psd.BlockSizeBytes != nil {
		m["block_size_bytes"] = int(aws.ToInt32(psd.BlockSizeBytes))
	}

	m["compression"] = types.ParquetCompressionSnappy
	if psd.Compression != "" {
		m["compression"] = psd.Compression
	}

	m["page_size_bytes"] = 1048576
	if psd.PageSizeBytes != nil {
		m["page_size_bytes"] = int(aws.ToInt32(psd.PageSizeBytes))
	}

	m["writer_version"] = types.ParquetWriterVersionV1
	if psd.WriterVersion != "" {
		m["writer_version"] = psd.WriterVersion
	}

	return []map[string]interface{}{m}
}

func flattenSchemaConfiguration(sc *types.SchemaConfiguration) []map[string]interface{} {
	if sc == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrCatalogID:    aws.ToString(sc.CatalogId),
		names.AttrDatabaseName: aws.ToString(sc.DatabaseName),
		names.AttrRegion:       aws.ToString(sc.Region),
		names.AttrRoleARN:      aws.ToString(sc.RoleARN),
		names.AttrTableName:    aws.ToString(sc.TableName),
		"version_id":           aws.ToString(sc.VersionId),
	}

	return []map[string]interface{}{m}
}

func flattenHTTPEndpointRequestConfiguration(rc *types.HttpEndpointRequestConfiguration) []map[string]interface{} {
	if rc == nil {
		return []map[string]interface{}{}
	}

	requestConfiguration := make([]map[string]interface{}, 1)

	commonAttributes := make([]interface{}, 0)
	for _, params := range rc.CommonAttributes {
		name := aws.ToString(params.AttributeName)
		value := aws.ToString(params.AttributeValue)

		commonAttributes = append(commonAttributes, map[string]interface{}{
			names.AttrName:  name,
			names.AttrValue: value,
		})
	}

	requestConfiguration[0] = map[string]interface{}{
		"common_attributes": commonAttributes,
		"content_encoding":  rc.ContentEncoding,
	}

	return requestConfiguration
}

func flattenProcessingConfiguration(pc *types.ProcessingConfiguration, destinationType destinationType, roleARN string) []map[string]interface{} {
	if pc == nil {
		return []map[string]interface{}{}
	}

	processingConfiguration := make([]map[string]interface{}, 1)

	processors := make([]interface{}, len(pc.Processors))
	for i, p := range pc.Processors {
		t := p.Type
		parameters := make([]interface{}, 0)

		// It is necessary to explicitly filter this out
		// to prevent diffs during routine use and retain the ability
		// to show diffs if any field has drifted.
		defaultProcessorParameters := defaultProcessorParameters(destinationType, t, roleARN)

		for _, params := range p.Parameters {
			name := params.ParameterName
			value := aws.ToString(params.ParameterValue)

			// Ignore defaults.
			if v, ok := defaultProcessorParameters[name]; ok && v == value {
				continue
			}

			parameters = append(parameters, map[string]interface{}{
				"parameter_name":  name,
				"parameter_value": value,
			})
		}

		processors[i] = map[string]interface{}{
			names.AttrType:       t,
			names.AttrParameters: parameters,
		}
	}
	processingConfiguration[0] = map[string]interface{}{
		names.AttrEnabled: aws.ToBool(pc.Enabled),
		"processors":      processors,
	}
	return processingConfiguration
}

func flattenSecretsManagerConfiguration(smc *types.SecretsManagerConfiguration) []interface{} {
	if smc == nil {
		return []interface{}{}
	}

	secretsManagerConfiguration := map[string]interface{}{
		names.AttrEnabled: aws.ToBool(smc.Enabled),
	}
	if aws.ToBool(smc.Enabled) {
		secretsManagerConfiguration["secret_arn"] = aws.ToString(smc.SecretARN)
		if smc.RoleARN != nil {
			secretsManagerConfiguration[names.AttrRoleARN] = aws.ToString(smc.RoleARN)
		}
	}
	return []interface{}{secretsManagerConfiguration}
}

func flattenDynamicPartitioningConfiguration(dpc *types.DynamicPartitioningConfiguration) []map[string]interface{} {
	if dpc == nil {
		return []map[string]interface{}{}
	}

	dynamicPartitioningConfiguration := make([]map[string]interface{}, 1)

	dynamicPartitioningConfiguration[0] = map[string]interface{}{
		names.AttrEnabled: aws.ToBool(dpc.Enabled),
	}

	if dpc.RetryOptions != nil && dpc.RetryOptions.DurationInSeconds != nil {
		dynamicPartitioningConfiguration[0]["retry_duration"] = int(aws.ToInt32(dpc.RetryOptions.DurationInSeconds))
	}

	return dynamicPartitioningConfiguration
}

func flattenKinesisStreamSourceDescription(desc *types.KinesisStreamSourceDescription) []interface{} {
	if desc == nil {
		return []interface{}{}
	}

	mDesc := map[string]interface{}{
		"kinesis_stream_arn": aws.ToString(desc.KinesisStreamARN),
		names.AttrRoleARN:    aws.ToString(desc.RoleARN),
	}

	return []interface{}{mDesc}
}

func flattenHTTPEndpointDestinationDescription(apiObject *types.HttpEndpointDestinationDescription, configuredAccessKey string) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrAccessKey:             configuredAccessKey,
		"cloudwatch_logging_options":    flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		names.AttrName:                  aws.ToString(apiObject.EndpointConfiguration.Name),
		"processing_configuration":      flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeHTTPEndpoint, aws.ToString(apiObject.RoleARN)),
		"request_configuration":         flattenHTTPEndpointRequestConfiguration(apiObject.RequestConfiguration),
		names.AttrRoleARN:               aws.ToString(apiObject.RoleARN),
		"s3_backup_mode":                apiObject.S3BackupMode,
		"s3_configuration":              flattenS3DestinationDescription(apiObject.S3DestinationDescription),
		"secrets_manager_configuration": flattenSecretsManagerConfiguration(apiObject.SecretsManagerConfiguration),
		names.AttrURL:                   aws.ToString(apiObject.EndpointConfiguration.Url),
	}

	if apiObject.BufferingHints != nil {
		tfMap["buffering_interval"] = int(aws.ToInt32(apiObject.BufferingHints.IntervalInSeconds))
		tfMap["buffering_size"] = int(aws.ToInt32(apiObject.BufferingHints.SizeInMBs))
	}

	if apiObject.RetryOptions != nil {
		tfMap["retry_duration"] = int(aws.ToInt32(apiObject.RetryOptions.DurationInSeconds))
	}

	return []interface{}{tfMap}
}

func expandDocumentIDOptions(tfMap map[string]interface{}) *types.DocumentIdOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DocumentIdOptions{}

	if v, ok := tfMap["default_document_id_format"].(string); ok && v != "" {
		apiObject.DefaultDocumentIdFormat = types.DefaultDocumentIdFormat(v)
	}

	return apiObject
}

func flattenDocumentIDOptions(apiObject *types.DocumentIdOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"default_document_id_format": apiObject.DefaultDocumentIdFormat,
	}

	return tfMap
}

func flattenSnowflakeRoleConfiguration(apiObject *types.SnowflakeRoleConfiguration) []map[string]interface{} {
	if apiObject == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}
	if aws.ToBool(apiObject.Enabled) {
		m["snowflake_role"] = aws.ToString(apiObject.SnowflakeRole)
	}

	return []map[string]interface{}{m}
}

func flattenSnowflakeVPCConfiguration(apiObject *types.SnowflakeVpcConfiguration) []map[string]interface{} {
	if apiObject == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"private_link_vpce_id": aws.ToString(apiObject.PrivateLinkVpceId),
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

	if v, ok := tfMap[names.AttrEnabled]; ok {
		enabled = v.(bool)
	}

	return !enabled
}

// See https://docs.aws.amazon.com/firehose/latest/dev/data-transformation.html.
func defaultProcessorParameters(destinationType destinationType, processorType types.ProcessorType, roleARN string) map[types.ProcessorParameterName]string {
	switch processorType {
	case types.ProcessorTypeLambda:
		params := map[types.ProcessorParameterName]string{
			types.ProcessorParameterNameLambdaNumberOfRetries:   "3",
			types.ProcessorParameterNameBufferIntervalInSeconds: "60",
		}
		if roleARN != "" {
			params[types.ProcessorParameterNameRoleArn] = roleARN
		}
		switch destinationType {
		case destinationTypeSplunk:
			params[types.ProcessorParameterNameBufferSizeInMb] = "0.25"
		default:
			params[types.ProcessorParameterNameBufferSizeInMb] = "1"
		}
		return params
	default:
		return make(map[types.ProcessorParameterName]string)
	}
}
