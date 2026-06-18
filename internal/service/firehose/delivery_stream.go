// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
	destinationTypeIceberg              destinationType = "iceberg"
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
		destinationTypeIceberg,
		destinationTypeOpenSearch,
		destinationTypeOpenSearchServerless,
		destinationTypeRedshift,
		destinationTypeSnowflake,
		destinationTypeSplunk,
	}
}

const (
	defaultBucketPrefixTimeZone = "UTC"
)

// @SDKResource("aws_kinesis_firehose_delivery_stream", name="Delivery Stream")
// @Tags(identifierAttribute="name")
// @ArnIdentity
// @CustomImport
// @Testing(preIdentityVersion="v6.47.0")
// @Testing(tagsTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/firehose/types;awstypes;awstypes.DeliveryStreamDescription")
func resourceDeliveryStream() *schema.Resource {
	// lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeliveryStreamCreate,
		ReadWithoutTimeout:   resourceDeliveryStreamRead,
		UpdateWithoutTimeout: resourceDeliveryStreamUpdate,
		DeleteWithoutTimeout: resourceDeliveryStreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.Import(ctx, d, meta); err != nil {
					return nil, err
				}

				id := d.Id()
				idErr := fmt.Errorf("unexpected format for ID (%[1]s), expected  arn:<partition>:firehose:<region>:<account-id>:deliverystream/<stream-name>", id)
				resARN, err := arn.Parse(id)
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
			destinationTableConfigurationSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					ForceNew: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDatabaseName: {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrTableName: {
								Type:     schema.TypeString,
								Required: true,
							},
							"s3_error_output_prefix": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							"unique_keys": {
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
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
					SchemaFunc: func() map[string]*schema.Schema {
						return map[string]*schema.Schema{
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
						}
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
					StateFunc: func(v any) string {
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
								Default:      defaultBucketPrefixTimeZone,
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
				"iceberg_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"append_only": {
								Type:     schema.TypeBool,
								Optional: true,
								Computed: true,
								ForceNew: true,
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
							"catalog_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ForceNew:     true,
								ValidateFunc: verify.ValidARN,
							},
							"cloudwatch_logging_options":      cloudWatchLoggingOptionsSchema(),
							"destination_table_configuration": destinationTableConfigurationSchema(),
							"processing_configuration":        processingConfigurationSchema(),
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
								Optional:         true,
								Default:          types.IcebergS3BackupModeFailedDataOnly,
								ValidateDiagFunc: enum.Validate[types.IcebergS3BackupMode](),
							},
							"s3_configuration": s3ConfigurationSchema(),
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
							"read_from_timestamp": {
								Type:         schema.TypeString,
								Optional:     true,
								ForceNew:     true,
								ValidateFunc: validation.IsRFC3339Time,
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
							"buffering_interval": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      0,
								ValidateFunc: validation.IntBetween(0, 900),
							},
							"buffering_size": {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      1,
								ValidateFunc: validation.IntBetween(1, 128),
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
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				destination := destinationType(d.Get(names.AttrDestination).(string))
				requiredAttribute := map[destinationType]string{
					destinationTypeElasticsearch:        "elasticsearch_configuration",
					destinationTypeExtendedS3:           "extended_s3_configuration",
					destinationTypeHTTPEndpoint:         "http_endpoint_configuration",
					destinationTypeIceberg:              "iceberg_configuration",
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

func resourceDeliveryStreamCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	partition := c.Partition(ctx)
	conn := c.FirehoseClient(ctx)

	sn := d.Get(names.AttrName).(string)
	input := firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String(sn),
		DeliveryStreamType: types.DeliveryStreamTypeDirectPut,
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("kinesis_source_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DeliveryStreamType = types.DeliveryStreamTypeKinesisStreamAsSource
		input.KinesisStreamSourceConfiguration = expandKinesisStreamSourceConfiguration(v.([]any)[0].(map[string]any))
	} else if v, ok := d.GetOk("msk_source_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DeliveryStreamType = types.DeliveryStreamTypeMSKAsSource
		input.MSKSourceConfiguration = expandMSKSourceConfiguration(v.([]any)[0].(map[string]any))
	}

	switch v := destinationType(d.Get(names.AttrDestination).(string)); v {
	case destinationTypeElasticsearch:
		if v, ok := d.GetOk("elasticsearch_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ElasticsearchDestinationConfiguration = expandElasticsearchDestinationConfiguration(v.([]any)[0].(map[string]any))
		}
	case destinationTypeExtendedS3:
		if v, ok := d.GetOk("extended_s3_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ExtendedS3DestinationConfiguration = expandExtendedS3DestinationConfiguration(v.([]any)[0].(map[string]any))
		}
	case destinationTypeHTTPEndpoint:
		if v, ok := d.GetOk("http_endpoint_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.HttpEndpointDestinationConfiguration = expandHTTPEndpointDestinationConfiguration(v.([]any)[0].(map[string]any))
		}
	case destinationTypeIceberg:
		if v, ok := d.GetOk("iceberg_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.IcebergDestinationConfiguration = expandIcebergDestinationConfiguration(v.([]any)[0].(map[string]any))
		}
	case destinationTypeOpenSearch:
		if v, ok := d.GetOk("opensearch_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.AmazonopensearchserviceDestinationConfiguration = expandAmazonopensearchserviceDestinationConfiguration(v.([]any)[0].(map[string]any))
		}
	case destinationTypeOpenSearchServerless:
		if v, ok := d.GetOk("opensearchserverless_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.AmazonOpenSearchServerlessDestinationConfiguration = expandAmazonOpenSearchServerlessDestinationConfiguration(v.([]any)[0].(map[string]any))
		}
	case destinationTypeRedshift:
		if v, ok := d.GetOk("redshift_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.RedshiftDestinationConfiguration = expandRedshiftDestinationConfiguration(v.([]any)[0].(map[string]any))
		}
	case destinationTypeSnowflake:
		if v, ok := d.GetOk("snowflake_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.SnowflakeDestinationConfiguration = expandSnowflakeDestinationConfiguration(v.([]any)[0].(map[string]any))
		}
	case destinationTypeSplunk:
		if v, ok := d.GetOk("splunk_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.SplunkDestinationConfiguration = expandSplunkDestinationConfiguration(v.([]any)[0].(map[string]any))
		}
	}

	_, err := retryDeliveryStreamOp(ctx, func(ctx context.Context) (any, error) {
		return conn.CreateDeliveryStream(ctx, &input)
	})

	// Some partitions (e.g. ISO) reject unsupported ExtendedS3DestinationConfiguration
	// fields with a ValidationException or an InvalidArgumentException. Outside the
	// standard partition, strip the fields and retry.
	if partition != endpoints.AwsPartitionID && input.ExtendedS3DestinationConfiguration != nil &&
		(tfawserr.ErrMessageContains(err, errCodeValidationException, "ExtendedS3DestinationConfiguration") ||
			tfawserr.ErrMessageContains(err, errCodeInvalidArgumentException, "S3 File Extension")) {
		input.ExtendedS3DestinationConfiguration.CustomTimeZone = nil
		input.ExtendedS3DestinationConfiguration.FileExtension = nil
		_, err = retryDeliveryStreamOp(ctx, func(ctx context.Context) (any, error) {
			return conn.CreateDeliveryStream(ctx, &input)
		})
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Firehose Delivery Stream (%s): %s", sn, err)
	}

	output, err := waitDeliveryStreamCreated(ctx, conn, sn, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Firehose Delivery Stream (%s) create: %s", sn, err)
	}

	d.SetId(aws.ToString(output.DeliveryStreamARN))

	if v, ok := d.GetOk("server_side_encryption"); ok && !isDeliveryStreamOptionDisabled(v) {
		input := firehose.StartDeliveryStreamEncryptionInput{
			DeliveryStreamEncryptionConfigurationInput: expandDeliveryStreamEncryptionConfigurationInput(v.([]any)),
			DeliveryStreamName:                         aws.String(sn),
		}

		_, err := conn.StartDeliveryStreamEncryption(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "starting Kinesis Firehose Delivery Stream (%s) encryption: %s", sn, err)
		}

		if _, err := waitDeliveryStreamEncryptionEnabled(ctx, conn, sn, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Firehose Delivery Stream (%s) encryption enable: %s", sn, err)
		}
	}

	return append(diags, resourceDeliveryStreamRead(ctx, d, meta)...)
}

func resourceDeliveryStreamRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FirehoseClient(ctx)

	sn := d.Get(names.AttrName).(string)
	s, err := findDeliveryStreamByName(ctx, conn, sn)

	if !d.IsNewResource() && retry.NotFound(err) {
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
			if err := d.Set("msk_source_configuration", []any{flattenMSKSourceDescription(v)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting msk_source_configuration: %s", err)
			}
		}
	}
	d.Set(names.AttrName, s.DeliveryStreamName)
	d.Set("version_id", s.VersionId)

	tfMapSSEOptions := map[string]any{
		names.AttrEnabled: false,
		"key_type":        types.KeyTypeAwsOwnedCmk,
	}
	if v := s.DeliveryStreamEncryptionConfiguration; v != nil && v.Status == types.DeliveryStreamEncryptionStatusEnabled {
		tfMapSSEOptions[names.AttrEnabled] = true
		tfMapSSEOptions["key_type"] = v.KeyType

		if v := v.KeyARN; v != nil {
			tfMapSSEOptions["key_arn"] = aws.ToString(v)
		}
	}
	if err := d.Set("server_side_encryption", []any{tfMapSSEOptions}); err != nil {
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
		case destination.IcebergDestinationDescription != nil:
			d.Set(names.AttrDestination, destinationTypeIceberg)
			if err := d.Set("iceberg_configuration", flattenIcebergDestinationDescription(destination.IcebergDestinationDescription)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting iceberg_configuration: %s", err)
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

func resourceDeliveryStreamUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	partition := c.Partition(ctx)
	conn := c.FirehoseClient(ctx)

	sn := d.Get(names.AttrName).(string)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := firehose.UpdateDestinationInput{
			CurrentDeliveryStreamVersionId: aws.String(d.Get("version_id").(string)),
			DeliveryStreamName:             aws.String(sn),
			DestinationId:                  aws.String(d.Get("destination_id").(string)),
		}

		switch v := destinationType(d.Get(names.AttrDestination).(string)); v {
		case destinationTypeElasticsearch:
			if v, ok := d.GetOk("elasticsearch_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.ElasticsearchDestinationUpdate = expandElasticsearchDestinationUpdate(v.([]any)[0].(map[string]any))
			}
		case destinationTypeExtendedS3:
			if v, ok := d.GetOk("extended_s3_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.ExtendedS3DestinationUpdate = expandExtendedS3DestinationUpdate(v.([]any)[0].(map[string]any))
			}
		case destinationTypeHTTPEndpoint:
			if v, ok := d.GetOk("http_endpoint_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.HttpEndpointDestinationUpdate = expandHTTPEndpointDestinationUpdate(v.([]any)[0].(map[string]any))
			}
		case destinationTypeIceberg:
			if v, ok := d.GetOk("iceberg_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.IcebergDestinationUpdate = expandIcebergDestinationUpdate(v.([]any)[0].(map[string]any))
			}
		case destinationTypeOpenSearch:
			if v, ok := d.GetOk("opensearch_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.AmazonopensearchserviceDestinationUpdate = expandAmazonopensearchserviceDestinationUpdate(v.([]any)[0].(map[string]any))
			}
		case destinationTypeOpenSearchServerless:
			if v, ok := d.GetOk("opensearchserverless_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.AmazonOpenSearchServerlessDestinationUpdate = expandAmazonOpenSearchServerlessDestinationUpdate(v.([]any)[0].(map[string]any))
			}
		case destinationTypeRedshift:
			if v, ok := d.GetOk("redshift_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.RedshiftDestinationUpdate = expandRedshiftDestinationUpdate(v.([]any)[0].(map[string]any))
			}
		case destinationTypeSnowflake:
			if v, ok := d.GetOk("snowflake_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.SnowflakeDestinationUpdate = expandSnowflakeDestinationUpdate(v.([]any)[0].(map[string]any))
			}
		case destinationTypeSplunk:
			if v, ok := d.GetOk("splunk_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.SplunkDestinationUpdate = expandSplunkDestinationUpdate(v.([]any)[0].(map[string]any))
			}
		}

		_, err := retryDeliveryStreamOp(ctx, func(ctx context.Context) (any, error) {
			return conn.UpdateDestination(ctx, &input)
		})

		// Some partitions (e.g. ISO) reject unsupported ExtendedS3DestinationUpdate
		// fields with a ValidationException or an InvalidArgumentException. Outside the
		// standard partition, strip the fields and retry.
		if partition != endpoints.AwsPartitionID && input.ExtendedS3DestinationUpdate != nil &&
			(tfawserr.ErrMessageContains(err, errCodeValidationException, "ExtendedS3DestinationUpdate") ||
				tfawserr.ErrMessageContains(err, errCodeInvalidArgumentException, "S3 File Extension")) {
			input.ExtendedS3DestinationUpdate.CustomTimeZone = nil
			input.ExtendedS3DestinationUpdate.FileExtension = nil
			_, err = retryDeliveryStreamOp(ctx, func(ctx context.Context) (any, error) {
				return conn.UpdateDestination(ctx, &input)
			})
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kinesis Firehose Delivery Stream (%s): %s", sn, err)
		}
	}

	if d.HasChange("server_side_encryption") {
		v := d.Get("server_side_encryption")
		if isDeliveryStreamOptionDisabled(v) {
			input := firehose.StopDeliveryStreamEncryptionInput{
				DeliveryStreamName: aws.String(sn),
			}

			_, err := conn.StopDeliveryStreamEncryption(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "stopping Kinesis Firehose Delivery Stream (%s) encryption: %s", sn, err)
			}

			if _, err := waitDeliveryStreamEncryptionDisabled(ctx, conn, sn, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Firehose Delivery Stream (%s) encryption disable: %s", sn, err)
			}
		} else {
			input := firehose.StartDeliveryStreamEncryptionInput{
				DeliveryStreamEncryptionConfigurationInput: expandDeliveryStreamEncryptionConfigurationInput(v.([]any)),
				DeliveryStreamName:                         aws.String(sn),
			}

			_, err := conn.StartDeliveryStreamEncryption(ctx, &input)

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

func resourceDeliveryStreamDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FirehoseClient(ctx)

	sn := d.Get(names.AttrName).(string)

	log.Printf("[DEBUG] Deleting Kinesis Firehose Delivery Stream: (%s)", sn)
	input := firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(sn),
	}
	_, err := conn.DeleteDeliveryStream(ctx, &input)

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

func retryDeliveryStreamOp(ctx context.Context, f func(context.Context) (any, error)) (any, error) {
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
	input := firehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: aws.String(name),
	}

	return findDeliveryStreamDescription(ctx, conn, &input)
}

func findDeliveryStreamDescription(ctx context.Context, conn *firehose.Client, input *firehose.DescribeDeliveryStreamInput) (*types.DeliveryStreamDescription, error) {
	output, err := conn.DescribeDeliveryStream(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DeliveryStreamDescription == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.DeliveryStreamDescription, nil
}

func statusDeliveryStream(conn *firehose.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDeliveryStreamByName(ctx, conn, name)

		if retry.NotFound(err) {
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
		Refresh: statusDeliveryStream(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DeliveryStreamDescription); ok {
		if status, failureDescription := output.DeliveryStreamStatus, output.FailureDescription; status == types.DeliveryStreamStatusCreatingFailed && failureDescription != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", failureDescription.Type, aws.ToString(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func waitDeliveryStreamDeleted(ctx context.Context, conn *firehose.Client, name string, timeout time.Duration) (*types.DeliveryStreamDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DeliveryStreamStatusDeleting),
		Target:  []string{},
		Refresh: statusDeliveryStream(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DeliveryStreamDescription); ok {
		if status, failureDescription := output.DeliveryStreamStatus, output.FailureDescription; status == types.DeliveryStreamStatusDeletingFailed && failureDescription != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", failureDescription.Type, aws.ToString(failureDescription.Details)))
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
		return nil, tfresource.NewEmptyResultError()
	}

	return output.DeliveryStreamEncryptionConfiguration, nil
}

func statusDeliveryStreamEncryptionConfiguration(conn *firehose.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDeliveryStreamEncryptionConfigurationByName(ctx, conn, name)

		if retry.NotFound(err) {
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
		Refresh: statusDeliveryStreamEncryptionConfiguration(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DeliveryStreamEncryptionConfiguration); ok {
		if status, failureDescription := output.Status, output.FailureDescription; status == types.DeliveryStreamEncryptionStatusEnablingFailed && failureDescription != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", failureDescription.Type, aws.ToString(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func waitDeliveryStreamEncryptionDisabled(ctx context.Context, conn *firehose.Client, name string, timeout time.Duration) (*types.DeliveryStreamEncryptionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DeliveryStreamEncryptionStatusDisabling),
		Target:  enum.Slice(types.DeliveryStreamEncryptionStatusDisabled),
		Refresh: statusDeliveryStreamEncryptionConfiguration(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DeliveryStreamEncryptionConfiguration); ok {
		if status, failureDescription := output.Status, output.FailureDescription; status == types.DeliveryStreamEncryptionStatusDisablingFailed && failureDescription != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", failureDescription.Type, aws.ToString(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func expandKinesisStreamSourceConfiguration(tfMap map[string]any) *types.KinesisStreamSourceConfiguration {
	apiObject := &types.KinesisStreamSourceConfiguration{
		KinesisStreamARN: aws.String(tfMap["kinesis_stream_arn"].(string)),
		RoleARN:          aws.String(tfMap[names.AttrRoleARN].(string)),
	}

	return apiObject
}

func expandS3DestinationConfiguration(tfList []any) *types.S3DestinationConfiguration {
	tfMap := tfList[0].(map[string]any)

	apiObject := &types.S3DestinationConfiguration{
		BucketARN: aws.String(tfMap["bucket_arn"].(string)),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(tfMap["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(tfMap["buffering_size"].(int))),
		},
		CompressionFormat:       types.CompressionFormat(tfMap["compression_format"].(string)),
		EncryptionConfiguration: expandEncryptionConfiguration(tfMap),
		Prefix:                  expandPrefix(tfMap),
		RoleARN:                 aws.String(tfMap[names.AttrRoleARN].(string)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["error_output_prefix"].(string); ok && v != "" {
		apiObject.ErrorOutputPrefix = aws.String(v)
	}

	return apiObject
}

func expandS3DestinationConfigurationBackup(tfMap map[string]any) *types.S3DestinationConfiguration {
	tfList := tfMap["s3_backup_configuration"].([]any)
	if len(tfList) == 0 {
		return nil
	}

	return expandS3DestinationConfiguration(tfList)
}

func expandExtendedS3DestinationConfiguration(tfMap map[string]any) *types.ExtendedS3DestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.ExtendedS3DestinationConfiguration{
		BucketARN: aws.String(tfMap["bucket_arn"].(string)),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(tfMap["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(tfMap["buffering_size"].(int))),
		},
		CompressionFormat:                 types.CompressionFormat(tfMap["compression_format"].(string)),
		CustomTimeZone:                    aws.String(tfMap["custom_time_zone"].(string)),
		DataFormatConversionConfiguration: expandDataFormatConversionConfiguration(tfMap["data_format_conversion_configuration"].([]any)),
		EncryptionConfiguration:           expandEncryptionConfiguration(tfMap),
		FileExtension:                     aws.String(tfMap["file_extension"].(string)),
		Prefix:                            expandPrefix(tfMap),
		RoleARN:                           aws.String(roleARN),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if _, ok := tfMap["dynamic_partitioning_configuration"]; ok {
		apiObject.DynamicPartitioningConfiguration = expandDynamicPartitioningConfiguration(tfMap)
	}

	if v, ok := tfMap["error_output_prefix"].(string); ok && v != "" {
		apiObject.ErrorOutputPrefix = aws.String(v)
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeExtendedS3, roleARN)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.S3BackupMode(v.(string))
		apiObject.S3BackupConfiguration = expandS3DestinationConfigurationBackup(tfMap)
	}

	return apiObject
}

func expandS3DestinationUpdate(tfList []any) *types.S3DestinationUpdate {
	tfMap := tfList[0].(map[string]any)
	apiObject := &types.S3DestinationUpdate{
		BucketARN: aws.String(tfMap["bucket_arn"].(string)),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(tfMap["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(tfMap["buffering_size"].(int))),
		},
		CompressionFormat:       types.CompressionFormat(tfMap["compression_format"].(string)),
		EncryptionConfiguration: expandEncryptionConfiguration(tfMap),
		ErrorOutputPrefix:       aws.String(tfMap["error_output_prefix"].(string)),
		Prefix:                  expandPrefix(tfMap),
		RoleARN:                 aws.String(tfMap[names.AttrRoleARN].(string)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	return apiObject
}

func expandS3DestinationUpdateBackup(tfMap map[string]any) *types.S3DestinationUpdate {
	tfList := tfMap["s3_backup_configuration"].([]any)
	if len(tfList) == 0 {
		return nil
	}

	return expandS3DestinationUpdate(tfList)
}

func expandExtendedS3DestinationUpdate(tfMap map[string]any) *types.ExtendedS3DestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.ExtendedS3DestinationUpdate{
		BucketARN: aws.String(tfMap["bucket_arn"].(string)),
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(tfMap["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(tfMap["buffering_size"].(int))),
		},
		CompressionFormat:                 types.CompressionFormat(tfMap["compression_format"].(string)),
		CustomTimeZone:                    aws.String(tfMap["custom_time_zone"].(string)),
		DataFormatConversionConfiguration: expandDataFormatConversionConfiguration(tfMap["data_format_conversion_configuration"].([]any)),
		EncryptionConfiguration:           expandEncryptionConfiguration(tfMap),
		ErrorOutputPrefix:                 aws.String(tfMap["error_output_prefix"].(string)),
		FileExtension:                     aws.String(tfMap["file_extension"].(string)),
		Prefix:                            expandPrefix(tfMap),
		ProcessingConfiguration:           expandProcessingConfiguration(tfMap, destinationTypeExtendedS3, roleARN),
		RoleARN:                           aws.String(roleARN),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if _, ok := tfMap["dynamic_partitioning_configuration"]; ok {
		apiObject.DynamicPartitioningConfiguration = expandDynamicPartitioningConfiguration(tfMap)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.S3BackupMode(v.(string))
		apiObject.S3BackupUpdate = expandS3DestinationUpdateBackup(tfMap)
	}

	return apiObject
}

func expandDataFormatConversionConfiguration(tfList []any) *types.DataFormatConversionConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		// It is possible to just pass nil here, but this seems to be the
		// canonical form that AWS uses, and is less likely to produce diffs.
		return &types.DataFormatConversionConfiguration{
			Enabled: aws.Bool(false),
		}
	}

	tfMap := tfList[0].(map[string]any)

	return &types.DataFormatConversionConfiguration{
		Enabled:                   aws.Bool(tfMap[names.AttrEnabled].(bool)),
		InputFormatConfiguration:  expandInputFormatConfiguration(tfMap["input_format_configuration"].([]any)),
		OutputFormatConfiguration: expandOutputFormatConfiguration(tfMap["output_format_configuration"].([]any)),
		SchemaConfiguration:       expandSchemaConfiguration(tfMap["schema_configuration"].([]any)),
	}
}

func expandInputFormatConfiguration(tfList []any) *types.InputFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	return &types.InputFormatConfiguration{
		Deserializer: expandDeserializer(tfMap["deserializer"].([]any)),
	}
}

func expandDeserializer(tfList []any) *types.Deserializer {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	return &types.Deserializer{
		HiveJsonSerDe:  expandHiveJSONSerDe(tfMap["hive_json_ser_de"].([]any)),
		OpenXJsonSerDe: expandOpenXJSONSerDe(tfMap["open_x_json_ser_de"].([]any)),
	}
}

func expandHiveJSONSerDe(tfList []any) *types.HiveJsonSerDe {
	if len(tfList) == 0 {
		return nil
	}

	if tfList[0] == nil {
		return &types.HiveJsonSerDe{}
	}

	tfMap := tfList[0].(map[string]any)

	return &types.HiveJsonSerDe{
		TimestampFormats: flex.ExpandStringValueList(tfMap["timestamp_formats"].([]any)),
	}
}

func expandOpenXJSONSerDe(tfList []any) *types.OpenXJsonSerDe {
	if len(tfList) == 0 {
		return nil
	}

	if tfList[0] == nil {
		return &types.OpenXJsonSerDe{}
	}

	tfMap := tfList[0].(map[string]any)

	return &types.OpenXJsonSerDe{
		CaseInsensitive:                    aws.Bool(tfMap["case_insensitive"].(bool)),
		ColumnToJsonKeyMappings:            flex.ExpandStringValueMap(tfMap["column_to_json_key_mappings"].(map[string]any)),
		ConvertDotsInJsonKeysToUnderscores: aws.Bool(tfMap["convert_dots_in_json_keys_to_underscores"].(bool)),
	}
}

func expandOutputFormatConfiguration(tfList []any) *types.OutputFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	return &types.OutputFormatConfiguration{
		Serializer: expandSerializer(tfMap["serializer"].([]any)),
	}
}

func expandSerializer(tfList []any) *types.Serializer {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	return &types.Serializer{
		OrcSerDe:     expandOrcSerDe(tfMap["orc_ser_de"].([]any)),
		ParquetSerDe: expandParquetSerDe(tfMap["parquet_ser_de"].([]any)),
	}
}

func expandOrcSerDe(tfList []any) *types.OrcSerDe {
	if len(tfList) == 0 {
		return nil
	}

	if tfList[0] == nil {
		return &types.OrcSerDe{}
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &types.OrcSerDe{
		BlockSizeBytes:                      aws.Int32(int32(tfMap["block_size_bytes"].(int))),
		BloomFilterFalsePositiveProbability: aws.Float64(tfMap["bloom_filter_false_positive_probability"].(float64)),
		Compression:                         types.OrcCompression(tfMap["compression"].(string)),
		DictionaryKeyThreshold:              aws.Float64(tfMap["dictionary_key_threshold"].(float64)),
		EnablePadding:                       aws.Bool(tfMap["enable_padding"].(bool)),
		FormatVersion:                       types.OrcFormatVersion(tfMap["format_version"].(string)),
		PaddingTolerance:                    aws.Float64(tfMap["padding_tolerance"].(float64)),
		RowIndexStride:                      aws.Int32(int32(tfMap["row_index_stride"].(int))),
		StripeSizeBytes:                     aws.Int32(int32(tfMap["stripe_size_bytes"].(int))),
	}

	if v, ok := tfMap["bloom_filter_columns"].([]any); ok && len(v) > 0 {
		apiObject.BloomFilterColumns = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandParquetSerDe(tfList []any) *types.ParquetSerDe {
	if len(tfList) == 0 {
		return nil
	}

	if tfList[0] == nil {
		return &types.ParquetSerDe{}
	}

	tfMap := tfList[0].(map[string]any)

	return &types.ParquetSerDe{
		BlockSizeBytes:              aws.Int32(int32(tfMap["block_size_bytes"].(int))),
		Compression:                 types.ParquetCompression(tfMap["compression"].(string)),
		EnableDictionaryCompression: aws.Bool(tfMap["enable_dictionary_compression"].(bool)),
		MaxPaddingBytes:             aws.Int32(int32(tfMap["max_padding_bytes"].(int))),
		PageSizeBytes:               aws.Int32(int32(tfMap["page_size_bytes"].(int))),
		WriterVersion:               types.ParquetWriterVersion(tfMap["writer_version"].(string)),
	}
}

func expandSchemaConfiguration(tfList []any) *types.SchemaConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &types.SchemaConfiguration{
		DatabaseName: aws.String(tfMap[names.AttrDatabaseName].(string)),
		RoleARN:      aws.String(tfMap[names.AttrRoleARN].(string)),
		TableName:    aws.String(tfMap[names.AttrTableName].(string)),
		VersionId:    aws.String(tfMap["version_id"].(string)),
	}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func expandDynamicPartitioningConfiguration(tfMap map[string]any) *types.DynamicPartitioningConfiguration {
	tfList := tfMap["dynamic_partitioning_configuration"].([]any)
	if len(tfList) == 0 {
		return nil
	}

	tfMap = tfList[0].(map[string]any)
	apiObject := &types.DynamicPartitioningConfiguration{
		Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
	}

	if v, ok := tfMap["retry_duration"]; ok {
		apiObject.RetryOptions = &types.RetryOptions{
			DurationInSeconds: aws.Int32(int32(v.(int))),
		}
	}

	return apiObject
}

func expandProcessingConfiguration(tfMap map[string]any, destinationType destinationType, roleARN string) *types.ProcessingConfiguration {
	tfList := tfMap["processing_configuration"].([]any)
	if len(tfList) == 0 || tfList[0] == nil {
		// It is possible to just pass nil here, but this seems to be the
		// canonical form that AWS uses, and is less likely to produce diffs.
		return &types.ProcessingConfiguration{
			Enabled:    aws.Bool(false),
			Processors: []types.Processor{},
		}
	}

	tfMap = tfList[0].(map[string]any)

	return &types.ProcessingConfiguration{
		Enabled:    aws.Bool(tfMap[names.AttrEnabled].(bool)),
		Processors: expandProcessors(tfMap["processors"].([]any), destinationType, roleARN),
	}
}

func expandProcessors(tfList []any, destinationType destinationType, roleARN string) []types.Processor {
	apiObjects := []types.Processor{}

	for _, tfMapRaw := range tfList {
		apiObject := expandProcessor(tfMapRaw.(map[string]any))
		if apiObject != nil {
			// Merge in defaults.
			for name, value := range defaultProcessorParameters(destinationType, apiObject.Type, roleARN) {
				if !slices.ContainsFunc(apiObject.Parameters, func(v types.ProcessorParameter) bool { return name == v.ParameterName }) {
					apiObject.Parameters = append(apiObject.Parameters, types.ProcessorParameter{
						ParameterName:  name,
						ParameterValue: aws.String(value),
					})
				}
			}
			apiObjects = append(apiObjects, *apiObject)
		}
	}

	return apiObjects
}

func expandProcessor(tfMap map[string]any) *types.Processor {
	var apiObject *types.Processor
	if v := tfMap[names.AttrType].(string); v != "" {
		apiObject = &types.Processor{
			Type:       types.ProcessorType(v),
			Parameters: expandProcessorParameters(tfMap[names.AttrParameters].(*schema.Set).List()),
		}
	}
	return apiObject
}

func expandProcessorParameters(tfList []any) []types.ProcessorParameter {
	apiObjects := []types.ProcessorParameter{}

	for _, tfMapRaw := range tfList {
		apiObjects = append(apiObjects, expandProcessorParameter(tfMapRaw.(map[string]any)))
	}

	return apiObjects
}

func expandProcessorParameter(tfMap map[string]any) types.ProcessorParameter {
	apiObject := types.ProcessorParameter{
		ParameterName:  types.ProcessorParameterName(tfMap["parameter_name"].(string)),
		ParameterValue: aws.String(tfMap["parameter_value"].(string)),
	}

	return apiObject
}

func expandSecretsManagerConfiguration(tfMap map[string]any) *types.SecretsManagerConfiguration {
	tfList := tfMap["secrets_manager_configuration"].([]any)
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap = tfList[0].(map[string]any)
	apiObject := &types.SecretsManagerConfiguration{
		Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
	}

	if v, ok := tfMap[names.AttrRoleARN]; ok && len(v.(string)) > 0 {
		apiObject.RoleARN = aws.String(v.(string))
	}

	if v, ok := tfMap["secret_arn"]; ok && len(v.(string)) > 0 {
		apiObject.SecretARN = aws.String(v.(string))
	}

	return apiObject
}

func expandEncryptionConfiguration(tfMap map[string]any) *types.EncryptionConfiguration {
	if v, ok := tfMap[names.AttrKMSKeyARN]; ok && len(v.(string)) > 0 {
		return &types.EncryptionConfiguration{
			KMSEncryptionConfig: &types.KMSEncryptionConfig{
				AWSKMSKeyARN: aws.String(v.(string)),
			},
		}
	}

	return &types.EncryptionConfiguration{
		NoEncryptionConfig: types.NoEncryptionConfigNoEncryption,
	}
}

func expandCloudWatchLoggingOptions(tfMap map[string]any) *types.CloudWatchLoggingOptions {
	tfList := tfMap["cloudwatch_logging_options"].([]any)
	if len(tfList) == 0 {
		return nil
	}

	tfMap = tfList[0].(map[string]any)
	apiObject := &types.CloudWatchLoggingOptions{
		Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
	}

	if v, ok := tfMap[names.AttrLogGroupName]; ok {
		apiObject.LogGroupName = aws.String(v.(string))
	}

	if v, ok := tfMap["log_stream_name"]; ok {
		apiObject.LogStreamName = aws.String(v.(string))
	}

	return apiObject
}

func expandVPCConfiguration(tfMap map[string]any) *types.VpcConfiguration {
	tfList := tfMap[names.AttrVPCConfig].([]any)
	if len(tfList) == 0 {
		return nil
	}

	tfMap = tfList[0].(map[string]any)

	return &types.VpcConfiguration{
		RoleARN:          aws.String(tfMap[names.AttrRoleARN].(string)),
		SubnetIds:        flex.ExpandStringValueSet(tfMap[names.AttrSubnetIDs].(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringValueSet(tfMap[names.AttrSecurityGroupIDs].(*schema.Set)),
	}
}

func expandPrefix(tfMap map[string]any) *string {
	if v, ok := tfMap[names.AttrPrefix]; ok {
		return aws.String(v.(string))
	}

	return nil
}

func expandIcebergDestinationConfiguration(tfMap map[string]any) *types.IcebergDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.IcebergDestinationConfiguration{
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(tfMap["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(tfMap["buffering_size"].(int))),
		},
		CatalogConfiguration: &types.CatalogConfiguration{
			CatalogARN: aws.String(tfMap["catalog_arn"].(string)),
		},
		RoleARN:         aws.String(roleARN),
		S3Configuration: expandS3DestinationConfiguration(tfMap["s3_configuration"].([]any)),
	}

	if v, ok := tfMap["append_only"].(bool); ok && v {
		apiObject.AppendOnly = aws.Bool(v)
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if _, ok := tfMap["destination_table_configuration"]; ok {
		apiObject.DestinationTableConfigurationList = expandDestinationTableConfigurationList(tfMap)
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeIceberg, roleARN)
	}

	if _, ok := tfMap["retry_duration"]; ok {
		apiObject.RetryOptions = expandIcebergRetryOptions(tfMap)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.IcebergS3BackupMode(v.(string))
	}

	return apiObject
}

func expandIcebergDestinationUpdate(tfMap map[string]any) *types.IcebergDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.IcebergDestinationUpdate{
		BufferingHints: &types.BufferingHints{
			IntervalInSeconds: aws.Int32(int32(tfMap["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(tfMap["buffering_size"].(int))),
		},
		RoleARN: aws.String(roleARN),
	}

	if v, ok := tfMap["append_only"].(bool); ok && v {
		apiObject.AppendOnly = aws.Bool(v)
	}

	if catalogARN, ok := tfMap["catalog_arn"].(string); ok {
		apiObject.CatalogConfiguration = &types.CatalogConfiguration{
			CatalogARN: aws.String(catalogARN),
		}
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if _, ok := tfMap["destination_table_configuration"]; ok {
		apiObject.DestinationTableConfigurationList = expandDestinationTableConfigurationList(tfMap)
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeIceberg, roleARN)
	}

	if _, ok := tfMap["retry_duration"]; ok {
		apiObject.RetryOptions = expandIcebergRetryOptions(tfMap)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.IcebergS3BackupMode(v.(string))
	}

	if v, ok := tfMap["s3_configuration"]; ok {
		apiObject.S3Configuration = expandS3DestinationConfiguration(v.([]any))
	}

	return apiObject
}

func expandRedshiftDestinationConfiguration(tfMap map[string]any) *types.RedshiftDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.RedshiftDestinationConfiguration{
		ClusterJDBCURL:  aws.String(tfMap["cluster_jdbcurl"].(string)),
		CopyCommand:     expandCopyCommand(tfMap),
		RetryOptions:    expandRedshiftRetryOptions(tfMap),
		RoleARN:         aws.String(roleARN),
		S3Configuration: expandS3DestinationConfiguration(tfMap["s3_configuration"].([]any)),
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

func expandRedshiftDestinationUpdate(tfMap map[string]any) *types.RedshiftDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.RedshiftDestinationUpdate{
		ClusterJDBCURL: aws.String(tfMap["cluster_jdbcurl"].(string)),
		CopyCommand:    expandCopyCommand(tfMap),
		RetryOptions:   expandRedshiftRetryOptions(tfMap),
		RoleARN:        aws.String(roleARN),
	}

	s3Config := expandS3DestinationUpdate(tfMap["s3_configuration"].([]any))
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

func expandElasticsearchDestinationConfiguration(tfMap map[string]any) *types.ElasticsearchDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.ElasticsearchDestinationConfiguration{
		BufferingHints:  expandElasticsearchBufferingHints(tfMap),
		IndexName:       aws.String(tfMap["index_name"].(string)),
		RetryOptions:    expandElasticsearchRetryOptions(tfMap),
		RoleARN:         aws.String(roleARN),
		S3Configuration: expandS3DestinationConfiguration(tfMap["s3_configuration"].([]any)),
		TypeName:        aws.String(tfMap["type_name"].(string)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["cluster_endpoint"]; ok && v.(string) != "" {
		apiObject.ClusterEndpoint = aws.String(v.(string))
	}

	if v, ok := tfMap["domain_arn"]; ok && v.(string) != "" {
		apiObject.DomainARN = aws.String(v.(string))
	}

	if v, ok := tfMap["index_rotation_period"]; ok {
		apiObject.IndexRotationPeriod = types.ElasticsearchIndexRotationPeriod(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeElasticsearch, roleARN)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.ElasticsearchS3BackupMode(v.(string))
	}

	if _, ok := tfMap[names.AttrVPCConfig]; ok {
		apiObject.VpcConfiguration = expandVPCConfiguration(tfMap)
	}

	return apiObject
}

func expandElasticsearchDestinationUpdate(tfMap map[string]any) *types.ElasticsearchDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.ElasticsearchDestinationUpdate{
		BufferingHints: expandElasticsearchBufferingHints(tfMap),
		IndexName:      aws.String(tfMap["index_name"].(string)),
		RetryOptions:   expandElasticsearchRetryOptions(tfMap),
		RoleARN:        aws.String(roleARN),
		S3Update:       expandS3DestinationUpdate(tfMap["s3_configuration"].([]any)),
		TypeName:       aws.String(tfMap["type_name"].(string)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["cluster_endpoint"]; ok && v.(string) != "" {
		apiObject.ClusterEndpoint = aws.String(v.(string))
	}

	if v, ok := tfMap["domain_arn"]; ok && v.(string) != "" {
		apiObject.DomainARN = aws.String(v.(string))
	}

	if v, ok := tfMap["index_rotation_period"]; ok {
		apiObject.IndexRotationPeriod = types.ElasticsearchIndexRotationPeriod(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeElasticsearch, roleARN)
	}

	return apiObject
}

func expandAmazonopensearchserviceDestinationConfiguration(tfMap map[string]any) *types.AmazonopensearchserviceDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.AmazonopensearchserviceDestinationConfiguration{
		BufferingHints:  expandAmazonopensearchserviceBufferingHints(tfMap),
		IndexName:       aws.String(tfMap["index_name"].(string)),
		RetryOptions:    expandAmazonopensearchserviceRetryOptions(tfMap),
		RoleARN:         aws.String(roleARN),
		S3Configuration: expandS3DestinationConfiguration(tfMap["s3_configuration"].([]any)),
		TypeName:        aws.String(tfMap["type_name"].(string)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["cluster_endpoint"]; ok && v.(string) != "" {
		apiObject.ClusterEndpoint = aws.String(v.(string))
	}

	if v, ok := tfMap["document_id_options"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.DocumentIdOptions = expandDocumentIDOptions(v[0].(map[string]any))
	}

	if v, ok := tfMap["domain_arn"]; ok && v.(string) != "" {
		apiObject.DomainARN = aws.String(v.(string))
	}

	if v, ok := tfMap["index_rotation_period"]; ok {
		apiObject.IndexRotationPeriod = types.AmazonopensearchserviceIndexRotationPeriod(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeOpenSearch, roleARN)
	}

	if s3BackupMode, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.AmazonopensearchserviceS3BackupMode(s3BackupMode.(string))
	}

	if _, ok := tfMap[names.AttrVPCConfig]; ok {
		apiObject.VpcConfiguration = expandVPCConfiguration(tfMap)
	}

	return apiObject
}

func expandAmazonopensearchserviceDestinationUpdate(tfMap map[string]any) *types.AmazonopensearchserviceDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.AmazonopensearchserviceDestinationUpdate{
		BufferingHints: expandAmazonopensearchserviceBufferingHints(tfMap),
		IndexName:      aws.String(tfMap["index_name"].(string)),
		RetryOptions:   expandAmazonopensearchserviceRetryOptions(tfMap),
		RoleARN:        aws.String(roleARN),
		S3Update:       expandS3DestinationUpdate(tfMap["s3_configuration"].([]any)),
		TypeName:       aws.String(tfMap["type_name"].(string)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["cluster_endpoint"]; ok && v.(string) != "" {
		apiObject.ClusterEndpoint = aws.String(v.(string))
	}

	if v, ok := tfMap["document_id_options"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.DocumentIdOptions = expandDocumentIDOptions(v[0].(map[string]any))
	}

	if v, ok := tfMap["domain_arn"]; ok && v.(string) != "" {
		apiObject.DomainARN = aws.String(v.(string))
	}

	if v, ok := tfMap["index_rotation_period"]; ok {
		apiObject.IndexRotationPeriod = types.AmazonopensearchserviceIndexRotationPeriod(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeOpenSearch, roleARN)
	}

	return apiObject
}

func expandAmazonOpenSearchServerlessDestinationConfiguration(tfMap map[string]any) *types.AmazonOpenSearchServerlessDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.AmazonOpenSearchServerlessDestinationConfiguration{
		BufferingHints:  expandAmazonOpenSearchServerlessBufferingHints(tfMap),
		IndexName:       aws.String(tfMap["index_name"].(string)),
		RetryOptions:    expandAmazonOpenSearchServerlessRetryOptions(tfMap),
		RoleARN:         aws.String(roleARN),
		S3Configuration: expandS3DestinationConfiguration(tfMap["s3_configuration"].([]any)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["collection_endpoint"]; ok && v.(string) != "" {
		apiObject.CollectionEndpoint = aws.String(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeOpenSearchServerless, roleARN)
	}

	if v, ok := tfMap["s3_backup_mode"]; ok {
		apiObject.S3BackupMode = types.AmazonOpenSearchServerlessS3BackupMode(v.(string))
	}

	if _, ok := tfMap[names.AttrVPCConfig]; ok {
		apiObject.VpcConfiguration = expandVPCConfiguration(tfMap)
	}

	return apiObject
}

func expandAmazonOpenSearchServerlessDestinationUpdate(tfMap map[string]any) *types.AmazonOpenSearchServerlessDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.AmazonOpenSearchServerlessDestinationUpdate{
		BufferingHints: expandAmazonOpenSearchServerlessBufferingHints(tfMap),
		IndexName:      aws.String(tfMap["index_name"].(string)),
		RetryOptions:   expandAmazonOpenSearchServerlessRetryOptions(tfMap),
		RoleARN:        aws.String(roleARN),
		S3Update:       expandS3DestinationUpdate(tfMap["s3_configuration"].([]any)),
	}

	if _, ok := tfMap["cloudwatch_logging_options"]; ok {
		apiObject.CloudWatchLoggingOptions = expandCloudWatchLoggingOptions(tfMap)
	}

	if v, ok := tfMap["collection_endpoint"]; ok && v.(string) != "" {
		apiObject.CollectionEndpoint = aws.String(v.(string))
	}

	if _, ok := tfMap["processing_configuration"]; ok {
		apiObject.ProcessingConfiguration = expandProcessingConfiguration(tfMap, destinationTypeOpenSearchServerless, roleARN)
	}

	return apiObject
}

func expandSnowflakeDestinationConfiguration(tfMap map[string]any) *types.SnowflakeDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.SnowflakeDestinationConfiguration{
		AccountUrl: aws.String(tfMap["account_url"].(string)),
		BufferingHints: &types.SnowflakeBufferingHints{
			IntervalInSeconds: aws.Int32(int32(tfMap["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(tfMap["buffering_size"].(int))),
		},
		Database:                  aws.String(tfMap[names.AttrDatabase].(string)),
		RetryOptions:              expandSnowflakeRetryOptions(tfMap),
		RoleARN:                   aws.String(roleARN),
		S3Configuration:           expandS3DestinationConfiguration(tfMap["s3_configuration"].([]any)),
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

	if v, ok := tfMap["key_passphrase"]; ok && v.(string) != "" {
		apiObject.KeyPassphrase = aws.String(v.(string))
	}

	if v, ok := tfMap["metadata_column_name"]; ok && v.(string) != "" {
		apiObject.MetaDataColumnName = aws.String(v.(string))
	}

	if v, ok := tfMap[names.AttrPrivateKey]; ok && v.(string) != "" {
		apiObject.PrivateKey = aws.String(v.(string))
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

func expandSnowflakeDestinationUpdate(tfMap map[string]any) *types.SnowflakeDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.SnowflakeDestinationUpdate{
		AccountUrl: aws.String(tfMap["account_url"].(string)),
		BufferingHints: &types.SnowflakeBufferingHints{
			IntervalInSeconds: aws.Int32(int32(tfMap["buffering_interval"].(int))),
			SizeInMBs:         aws.Int32(int32(tfMap["buffering_size"].(int))),
		},
		Database:     aws.String(tfMap[names.AttrDatabase].(string)),
		RetryOptions: expandSnowflakeRetryOptions(tfMap),
		RoleARN:      aws.String(roleARN),
		S3Update:     expandS3DestinationUpdate(tfMap["s3_configuration"].([]any)),
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

	if v, ok := tfMap["key_passphrase"]; ok && v.(string) != "" {
		apiObject.KeyPassphrase = aws.String(v.(string))
	}

	if v, ok := tfMap["metadata_column_name"]; ok && v.(string) != "" {
		apiObject.MetaDataColumnName = aws.String(v.(string))
	}

	if v, ok := tfMap[names.AttrPrivateKey]; ok && v.(string) != "" {
		apiObject.PrivateKey = aws.String(v.(string))
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

func expandSplunkDestinationConfiguration(tfMap map[string]any) *types.SplunkDestinationConfiguration {
	apiObject := &types.SplunkDestinationConfiguration{
		HECAcknowledgmentTimeoutInSeconds: aws.Int32(int32(tfMap["hec_acknowledgment_timeout"].(int))),
		HECEndpoint:                       aws.String(tfMap["hec_endpoint"].(string)),
		HECEndpointType:                   types.HECEndpointType(tfMap["hec_endpoint_type"].(string)),
		RetryOptions:                      expandSplunkRetryOptions(tfMap),
		S3Configuration:                   expandS3DestinationConfiguration(tfMap["s3_configuration"].([]any)),
	}

	bufferingHints := &types.SplunkBufferingHints{}
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

func expandSplunkDestinationUpdate(tfMap map[string]any) *types.SplunkDestinationUpdate {
	apiObject := &types.SplunkDestinationUpdate{
		HECAcknowledgmentTimeoutInSeconds: aws.Int32(int32(tfMap["hec_acknowledgment_timeout"].(int))),
		HECEndpoint:                       aws.String(tfMap["hec_endpoint"].(string)),
		HECEndpointType:                   types.HECEndpointType(tfMap["hec_endpoint_type"].(string)),
		RetryOptions:                      expandSplunkRetryOptions(tfMap),
		S3Update:                          expandS3DestinationUpdate(tfMap["s3_configuration"].([]any)),
	}

	bufferingHints := &types.SplunkBufferingHints{}
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

func expandHTTPEndpointDestinationConfiguration(tfMap map[string]any) *types.HttpEndpointDestinationConfiguration {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.HttpEndpointDestinationConfiguration{
		EndpointConfiguration: expandHTTPEndpointConfiguration(tfMap),
		RetryOptions:          expandHTTPEndpointRetryOptions(tfMap),
		RoleARN:               aws.String(roleARN),
		S3Configuration:       expandS3DestinationConfiguration(tfMap["s3_configuration"].([]any)),
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

func expandHTTPEndpointDestinationUpdate(tfMap map[string]any) *types.HttpEndpointDestinationUpdate {
	roleARN := tfMap[names.AttrRoleARN].(string)
	apiObject := &types.HttpEndpointDestinationUpdate{
		EndpointConfiguration: expandHTTPEndpointConfiguration(tfMap),
		RetryOptions:          expandHTTPEndpointRetryOptions(tfMap),
		RoleARN:               aws.String(roleARN),
		S3Update:              expandS3DestinationUpdate(tfMap["s3_configuration"].([]any)),
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

func expandHTTPEndpointCommonAttributes(tfList []any) []types.HttpEndpointCommonAttribute {
	apiObjects := make([]types.HttpEndpointCommonAttribute, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := types.HttpEndpointCommonAttribute{
			AttributeName:  aws.String(tfMap[names.AttrName].(string)),
			AttributeValue: aws.String(tfMap[names.AttrValue].(string)),
		}
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandHTTPEndpointRequestConfiguration(tfMap map[string]any) *types.HttpEndpointRequestConfiguration {
	tfList := tfMap["request_configuration"].([]any)
	if len(tfList) == 0 {
		return nil
	}

	tfMap = tfList[0].(map[string]any)
	apiObject := &types.HttpEndpointRequestConfiguration{}

	if commonAttributes, ok := tfMap["common_attributes"]; ok {
		apiObject.CommonAttributes = expandHTTPEndpointCommonAttributes(commonAttributes.([]any))
	}

	if contentEncoding, ok := tfMap["content_encoding"]; ok {
		apiObject.ContentEncoding = types.ContentEncoding(contentEncoding.(string))
	}

	return apiObject
}

func expandHTTPEndpointConfiguration(tfMap map[string]any) *types.HttpEndpointConfiguration {
	apiObject := &types.HttpEndpointConfiguration{
		Url: aws.String(tfMap[names.AttrURL].(string)),
	}

	if v, ok := tfMap[names.AttrAccessKey]; ok {
		apiObject.AccessKey = aws.String(v.(string))
	}

	if v, ok := tfMap[names.AttrName]; ok {
		apiObject.Name = aws.String(v.(string))
	}

	return apiObject
}

func expandElasticsearchBufferingHints(tfMap map[string]any) *types.ElasticsearchBufferingHints {
	apiObject := &types.ElasticsearchBufferingHints{}

	if v, ok := tfMap["buffering_interval"].(int); ok {
		apiObject.IntervalInSeconds = aws.Int32(int32(v))
	}
	if v, ok := tfMap["buffering_size"].(int); ok {
		apiObject.SizeInMBs = aws.Int32(int32(v))
	}

	return apiObject
}

func expandAmazonopensearchserviceBufferingHints(tfMap map[string]any) *types.AmazonopensearchserviceBufferingHints {
	apiObject := &types.AmazonopensearchserviceBufferingHints{}

	if v, ok := tfMap["buffering_interval"].(int); ok {
		apiObject.IntervalInSeconds = aws.Int32(int32(v))
	}
	if v, ok := tfMap["buffering_size"].(int); ok {
		apiObject.SizeInMBs = aws.Int32(int32(v))
	}

	return apiObject
}

func expandAmazonOpenSearchServerlessBufferingHints(tfMap map[string]any) *types.AmazonOpenSearchServerlessBufferingHints {
	apiObject := &types.AmazonOpenSearchServerlessBufferingHints{}

	if v, ok := tfMap["buffering_interval"].(int); ok {
		apiObject.IntervalInSeconds = aws.Int32(int32(v))
	}
	if v, ok := tfMap["buffering_size"].(int); ok {
		apiObject.SizeInMBs = aws.Int32(int32(v))
	}

	return apiObject
}

func expandIcebergRetryOptions(tfMap map[string]any) *types.RetryOptions {
	apiObject := &types.RetryOptions{}

	if v, ok := tfMap["retry_duration"].(int); ok {
		apiObject.DurationInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandElasticsearchRetryOptions(tfMap map[string]any) *types.ElasticsearchRetryOptions {
	apiObject := &types.ElasticsearchRetryOptions{}

	if v, ok := tfMap["retry_duration"].(int); ok {
		apiObject.DurationInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandAmazonopensearchserviceRetryOptions(tfMap map[string]any) *types.AmazonopensearchserviceRetryOptions {
	apiObject := &types.AmazonopensearchserviceRetryOptions{}

	if retryDuration, ok := tfMap["retry_duration"].(int); ok {
		apiObject.DurationInSeconds = aws.Int32(int32(retryDuration))
	}

	return apiObject
}

func expandAmazonOpenSearchServerlessRetryOptions(tfMap map[string]any) *types.AmazonOpenSearchServerlessRetryOptions {
	apiObject := &types.AmazonOpenSearchServerlessRetryOptions{}

	if retryDuration, ok := tfMap["retry_duration"].(int); ok {
		apiObject.DurationInSeconds = aws.Int32(int32(retryDuration))
	}

	return apiObject
}

func expandHTTPEndpointRetryOptions(tfMap map[string]any) *types.HttpEndpointRetryOptions {
	apiObject := &types.HttpEndpointRetryOptions{}

	if v, ok := tfMap["retry_duration"].(int); ok {
		apiObject.DurationInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandRedshiftRetryOptions(tfMap map[string]any) *types.RedshiftRetryOptions {
	apiObject := &types.RedshiftRetryOptions{}

	if v, ok := tfMap["retry_duration"].(int); ok {
		apiObject.DurationInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandSnowflakeRetryOptions(tfMap map[string]any) *types.SnowflakeRetryOptions {
	apiObject := &types.SnowflakeRetryOptions{}

	if v, ok := tfMap["retry_duration"].(int); ok {
		apiObject.DurationInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandSnowflakeRoleConfiguration(tfMap map[string]any) *types.SnowflakeRoleConfiguration {
	tfList := tfMap["snowflake_role_configuration"].([]any)
	if len(tfList) == 0 || tfList[0] == nil {
		// It is possible to just pass nil here, but this seems to be the
		// canonical form that AWS uses, and is less likely to produce diffs.
		return &types.SnowflakeRoleConfiguration{
			Enabled: aws.Bool(false),
		}
	}

	tfMap = tfList[0].(map[string]any)
	apiObject := &types.SnowflakeRoleConfiguration{
		Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
	}

	if v, ok := tfMap["snowflake_role"]; ok && len(v.(string)) > 0 {
		apiObject.SnowflakeRole = aws.String(v.(string))
	}

	return apiObject
}

func expandSnowflakeVPCConfiguration(tfMap map[string]any) *types.SnowflakeVpcConfiguration {
	tfList := tfMap["snowflake_vpc_configuration"].([]any)
	if len(tfList) == 0 {
		return nil
	}

	tfMap = tfList[0].(map[string]any)

	apiObject := &types.SnowflakeVpcConfiguration{
		PrivateLinkVpceId: aws.String(tfMap["private_link_vpce_id"].(string)),
	}

	return apiObject
}

func expandSplunkRetryOptions(tfMap map[string]any) *types.SplunkRetryOptions {
	apiObject := &types.SplunkRetryOptions{}

	if v, ok := tfMap["retry_duration"].(int); ok {
		apiObject.DurationInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func expandDestinationTableConfigurationList(tfMap map[string]any) []types.DestinationTableConfiguration {
	tfList := tfMap["destination_table_configuration"].([]any)
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]types.DestinationTableConfiguration, 0, len(tfList))
	for _, tfMapRaw := range tfList {
		apiObjects = append(apiObjects, expandDestinationTableConfiguration(tfMapRaw.(map[string]any)))
	}

	return apiObjects
}

func expandDestinationTableConfiguration(tfMap map[string]any) types.DestinationTableConfiguration {
	apiObject := types.DestinationTableConfiguration{
		DestinationDatabaseName: aws.String(tfMap[names.AttrDatabaseName].(string)),
		DestinationTableName:    aws.String(tfMap[names.AttrTableName].(string)),
	}

	if v, ok := tfMap["s3_error_output_prefix"].(string); ok {
		apiObject.S3ErrorOutputPrefix = aws.String(v)
	}

	if v, ok := tfMap["unique_keys"].([]any); ok {
		apiObject.UniqueKeys = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandCopyCommand(tfMap map[string]any) *types.CopyCommand {
	apiObject := &types.CopyCommand{
		DataTableName: aws.String(tfMap["data_table_name"].(string)),
	}

	if v, ok := tfMap["copy_options"]; ok {
		apiObject.CopyOptions = aws.String(v.(string))
	}
	if v, ok := tfMap["data_table_columns"]; ok {
		apiObject.DataTableColumns = aws.String(v.(string))
	}

	return apiObject
}

func expandDeliveryStreamEncryptionConfigurationInput(tfList []any) *types.DeliveryStreamEncryptionConfigurationInput {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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

func expandMSKSourceConfiguration(tfMap map[string]any) *types.MSKSourceConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.MSKSourceConfiguration{}

	if v, ok := tfMap["authentication_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.AuthenticationConfiguration = expandAuthenticationConfiguration(v[0].(map[string]any))
	}

	if v, ok := tfMap["msk_cluster_arn"].(string); ok && v != "" {
		apiObject.MSKClusterARN = aws.String(v)
	}

	if v, ok := tfMap["read_from_timestamp"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)
		apiObject.ReadFromTimestamp = aws.Time(v)
	}

	if v, ok := tfMap["topic_name"].(string); ok && v != "" {
		apiObject.TopicName = aws.String(v)
	}

	return apiObject
}

func expandAuthenticationConfiguration(tfMap map[string]any) *types.AuthenticationConfiguration {
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

func flattenMSKSourceDescription(apiObject *types.MSKSourceDescription) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AuthenticationConfiguration; v != nil {
		tfMap["authentication_configuration"] = []any{flattenAuthenticationConfiguration(v)}
	}

	if v := apiObject.MSKClusterARN; v != nil {
		tfMap["msk_cluster_arn"] = aws.ToString(v)
	}

	if v := apiObject.ReadFromTimestamp; v != nil {
		tfMap["read_from_timestamp"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.TopicName; v != nil {
		tfMap["topic_name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenAuthenticationConfiguration(apiObject *types.AuthenticationConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"connectivity": apiObject.Connectivity,
	}

	if v := apiObject.RoleARN; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return tfMap
}

func flattenCloudWatchLoggingOptions(apiObject *types.CloudWatchLoggingOptions) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}

	if aws.ToBool(apiObject.Enabled) {
		tfMap[names.AttrLogGroupName] = aws.ToString(apiObject.LogGroupName)
		tfMap["log_stream_name"] = aws.ToString(apiObject.LogStreamName)
	}

	return []any{tfMap}
}

func flattenElasticsearchDestinationDescription(apiObject *types.ElasticsearchDestinationDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	roleARN := aws.ToString(apiObject.RoleARN)
	tfMap := map[string]any{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		"index_name":                 aws.ToString(apiObject.IndexName),
		"index_rotation_period":      apiObject.IndexRotationPeriod,
		"processing_configuration":   flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeElasticsearch, roleARN),
		names.AttrRoleARN:            roleARN,
		"s3_backup_mode":             apiObject.S3BackupMode,
		"s3_configuration":           flattenS3DestinationDescription(apiObject.S3DestinationDescription),
		"type_name":                  aws.ToString(apiObject.TypeName),
		names.AttrVPCConfig:          flattenVPCConfigurationDescription(apiObject.VpcConfigurationDescription),
	}

	if v := apiObject.BufferingHints; v != nil {
		if v.IntervalInSeconds != nil {
			tfMap["buffering_interval"] = aws.ToInt32(v.IntervalInSeconds)
		}
		if v.SizeInMBs != nil {
			tfMap["buffering_size"] = aws.ToInt32(v.SizeInMBs)
		}
	}

	if apiObject.ClusterEndpoint != nil {
		tfMap["cluster_endpoint"] = aws.ToString(apiObject.ClusterEndpoint)
	}

	if apiObject.DomainARN != nil {
		tfMap["domain_arn"] = aws.ToString(apiObject.DomainARN)
	}

	if v := apiObject.RetryOptions; v != nil {
		tfMap["retry_duration"] = aws.ToInt32(v.DurationInSeconds)
	}

	return []any{tfMap}
}

func flattenAmazonopensearchserviceDestinationDescription(apiObject *types.AmazonopensearchserviceDestinationDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	roleARN := aws.ToString(apiObject.RoleARN)
	tfMap := map[string]any{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		"index_name":                 aws.ToString(apiObject.IndexName),
		"index_rotation_period":      apiObject.IndexRotationPeriod,
		"processing_configuration":   flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeOpenSearch, roleARN),
		names.AttrRoleARN:            roleARN,
		"s3_backup_mode":             apiObject.S3BackupMode,
		"s3_configuration":           flattenS3DestinationDescription(apiObject.S3DestinationDescription),
		"type_name":                  aws.ToString(apiObject.TypeName),
		names.AttrVPCConfig:          flattenVPCConfigurationDescription(apiObject.VpcConfigurationDescription),
	}

	if v := apiObject.BufferingHints; v != nil {
		if v.IntervalInSeconds != nil {
			tfMap["buffering_interval"] = aws.ToInt32(v.IntervalInSeconds)
		}
		if v.SizeInMBs != nil {
			tfMap["buffering_size"] = aws.ToInt32(v.SizeInMBs)
		}
	}

	if apiObject.ClusterEndpoint != nil {
		tfMap["cluster_endpoint"] = aws.ToString(apiObject.ClusterEndpoint)
	}

	if v := apiObject.DocumentIdOptions; v != nil {
		tfMap["document_id_options"] = []any{flattenDocumentIDOptions(v)}
	}

	if apiObject.DomainARN != nil {
		tfMap["domain_arn"] = aws.ToString(apiObject.DomainARN)
	}

	if v := apiObject.RetryOptions; v != nil {
		tfMap["retry_duration"] = aws.ToInt32(v.DurationInSeconds)
	}

	return []any{tfMap}
}

func flattenAmazonOpenSearchServerlessDestinationDescription(apiObject *types.AmazonOpenSearchServerlessDestinationDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	roleARN := aws.ToString(apiObject.RoleARN)
	tfMap := map[string]any{
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		"index_name":                 aws.ToString(apiObject.IndexName),
		"processing_configuration":   flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeOpenSearchServerless, roleARN),
		names.AttrRoleARN:            roleARN,
		"s3_backup_mode":             apiObject.S3BackupMode,
		"s3_configuration":           flattenS3DestinationDescription(apiObject.S3DestinationDescription),
		names.AttrVPCConfig:          flattenVPCConfigurationDescription(apiObject.VpcConfigurationDescription),
	}

	if v := apiObject.BufferingHints; v != nil {
		if v.IntervalInSeconds != nil {
			tfMap["buffering_interval"] = aws.ToInt32(v.IntervalInSeconds)
		}
		if v.SizeInMBs != nil {
			tfMap["buffering_size"] = aws.ToInt32(v.SizeInMBs)
		}
	}

	if apiObject.CollectionEndpoint != nil {
		tfMap["collection_endpoint"] = aws.ToString(apiObject.CollectionEndpoint)
	}

	if v := apiObject.RetryOptions; v != nil {
		tfMap["retry_duration"] = aws.ToInt32(v.DurationInSeconds)
	}

	return []any{tfMap}
}

func flattenVPCConfigurationDescription(apiObject *types.VpcConfigurationDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrRoleARN:          aws.ToString(apiObject.RoleARN),
		names.AttrSecurityGroupIDs: apiObject.SecurityGroupIds,
		names.AttrSubnetIDs:        apiObject.SubnetIds,
		names.AttrVPCID:            aws.ToString(apiObject.VpcId),
	}

	return []any{tfMap}
}

func flattenExtendedS3DestinationDescription(apiObject *types.ExtendedS3DestinationDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	roleARN := aws.ToString(apiObject.RoleARN)
	tfMap := map[string]any{
		"bucket_arn":                           aws.ToString(apiObject.BucketARN),
		"cloudwatch_logging_options":           flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		"compression_format":                   apiObject.CompressionFormat,
		"data_format_conversion_configuration": flattenDataFormatConversionConfiguration(apiObject.DataFormatConversionConfiguration),
		"dynamic_partitioning_configuration":   flattenDynamicPartitioningConfiguration(apiObject.DynamicPartitioningConfiguration),
		"error_output_prefix":                  aws.ToString(apiObject.ErrorOutputPrefix),
		"file_extension":                       aws.ToString(apiObject.FileExtension),
		names.AttrPrefix:                       aws.ToString(apiObject.Prefix),
		"processing_configuration":             flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeExtendedS3, roleARN),
		names.AttrRoleARN:                      roleARN,
		"s3_backup_configuration":              flattenS3DestinationDescription(apiObject.S3BackupDescription),
		"s3_backup_mode":                       apiObject.S3BackupMode,
	}

	if v := apiObject.BufferingHints; v != nil {
		if v.IntervalInSeconds != nil {
			tfMap["buffering_interval"] = aws.ToInt32(v.IntervalInSeconds)
		}
		if v.SizeInMBs != nil {
			tfMap["buffering_size"] = aws.ToInt32(v.SizeInMBs)
		}
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference.

	tfMap["custom_time_zone"] = defaultBucketPrefixTimeZone
	if apiObject.CustomTimeZone != nil {
		tfMap["custom_time_zone"] = aws.ToString(apiObject.CustomTimeZone)
	}

	if v := apiObject.EncryptionConfiguration; v != nil && v.KMSEncryptionConfig != nil {
		tfMap[names.AttrKMSKeyARN] = aws.ToString(v.KMSEncryptionConfig.AWSKMSKeyARN)
	}

	return []any{tfMap}
}

func flattenRedshiftDestinationDescription(apiObject *types.RedshiftDestinationDescription, configuredPassword string) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
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

	if v := apiObject.RetryOptions; v != nil {
		tfMap["retry_duration"] = aws.ToInt32(v.DurationInSeconds)
	}

	return []any{tfMap}
}

func flattenSnowflakeDestinationDescription(apiObject *types.SnowflakeDestinationDescription, configuredKeyPassphrase, configuredPrivateKey string) []any {
	if apiObject == nil {
		return []any{}
	}

	roleARN := aws.ToString(apiObject.RoleARN)
	tfMap := map[string]any{
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

	if v := apiObject.BufferingHints; v != nil {
		if v.IntervalInSeconds != nil {
			tfMap["buffering_interval"] = aws.ToInt32(v.IntervalInSeconds)
		}
		if v.SizeInMBs != nil {
			tfMap["buffering_size"] = aws.ToInt32(v.SizeInMBs)
		}
	}

	if v := apiObject.RetryOptions; v != nil {
		tfMap["retry_duration"] = aws.ToInt32(v.DurationInSeconds)
	}

	return []any{tfMap}
}

func flattenSplunkDestinationDescription(apiObject *types.SplunkDestinationDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"cloudwatch_logging_options":    flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		"hec_acknowledgment_timeout":    aws.ToInt32(apiObject.HECAcknowledgmentTimeoutInSeconds),
		"hec_endpoint":                  aws.ToString(apiObject.HECEndpoint),
		"hec_endpoint_type":             apiObject.HECEndpointType,
		"hec_token":                     aws.ToString(apiObject.HECToken),
		"processing_configuration":      flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeSplunk, ""),
		"s3_backup_mode":                apiObject.S3BackupMode,
		"s3_configuration":              flattenS3DestinationDescription(apiObject.S3DestinationDescription),
		"secrets_manager_configuration": flattenSecretsManagerConfiguration(apiObject.SecretsManagerConfiguration),
	}

	if v := apiObject.BufferingHints; v != nil {
		if v.IntervalInSeconds != nil {
			tfMap["buffering_interval"] = aws.ToInt32(v.IntervalInSeconds)
		}
		if v.SizeInMBs != nil {
			tfMap["buffering_size"] = aws.ToInt32(v.SizeInMBs)
		}
	}

	if v := apiObject.RetryOptions; v != nil {
		tfMap["retry_duration"] = aws.ToInt32(v.DurationInSeconds)
	}

	return []any{tfMap}
}

func flattenS3DestinationDescription(apiObject *types.S3DestinationDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"bucket_arn":                 aws.ToString(apiObject.BucketARN),
		"cloudwatch_logging_options": flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		"compression_format":         apiObject.CompressionFormat,
		"error_output_prefix":        aws.ToString(apiObject.ErrorOutputPrefix),
		names.AttrPrefix:             aws.ToString(apiObject.Prefix),
		names.AttrRoleARN:            aws.ToString(apiObject.RoleARN),
	}

	if v := apiObject.BufferingHints; v != nil {
		if v.IntervalInSeconds != nil {
			tfMap["buffering_interval"] = aws.ToInt32(v.IntervalInSeconds)
		}
		if v.SizeInMBs != nil {
			tfMap["buffering_size"] = aws.ToInt32(v.SizeInMBs)
		}
	}

	if v := apiObject.EncryptionConfiguration; v != nil && v.KMSEncryptionConfig != nil {
		tfMap[names.AttrKMSKeyARN] = aws.ToString(v.KMSEncryptionConfig.AWSKMSKeyARN)
	}

	return []any{tfMap}
}

func flattenDataFormatConversionConfiguration(apiObject *types.DataFormatConversionConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	enabled := aws.ToBool(apiObject.Enabled)
	tfListIFC := flattenInputFormatConfiguration(apiObject.InputFormatConfiguration)
	tfListOFC := flattenOutputFormatConfiguration(apiObject.OutputFormatConfiguration)
	tfListSC := flattenSchemaConfiguration(apiObject.SchemaConfiguration)

	// The AWS SDK can represent "no data format conversion configuration" in two ways:
	// 1. With a nil value
	// 2. With enabled set to false and nil for ALL the config sections.
	// We normalize this with an empty configuration in the state due
	// to the existing Default: true on the enabled attribute.
	if !enabled && len(tfListIFC) == 0 && len(tfListOFC) == 0 && len(tfListSC) == 0 {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrEnabled:             enabled,
		"input_format_configuration":  tfListIFC,
		"output_format_configuration": tfListOFC,
		"schema_configuration":        tfListSC,
	}

	return []any{tfMap}
}

func flattenInputFormatConfiguration(apiObject *types.InputFormatConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"deserializer": flattenDeserializer(apiObject.Deserializer),
	}

	return []any{tfMap}
}

func flattenDeserializer(apiObject *types.Deserializer) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"hive_json_ser_de":   flattenHiveJSONSerDe(apiObject.HiveJsonSerDe),
		"open_x_json_ser_de": flattenOpenXJSONSerDe(apiObject.OpenXJsonSerDe),
	}

	return []any{tfMap}
}

func flattenHiveJSONSerDe(apiObject *types.HiveJsonSerDe) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"timestamp_formats": apiObject.TimestampFormats,
	}

	return []any{tfMap}
}

func flattenOpenXJSONSerDe(apiObject *types.OpenXJsonSerDe) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"column_to_json_key_mappings":              apiObject.ColumnToJsonKeyMappings,
		"convert_dots_in_json_keys_to_underscores": aws.ToBool(apiObject.ConvertDotsInJsonKeysToUnderscores),
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference.

	tfMap["case_insensitive"] = true
	if apiObject.CaseInsensitive != nil {
		tfMap["case_insensitive"] = aws.ToBool(apiObject.CaseInsensitive)
	}

	return []any{tfMap}
}

func flattenOutputFormatConfiguration(apiObject *types.OutputFormatConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"serializer": flattenSerializer(apiObject.Serializer),
	}

	return []any{tfMap}
}

func flattenSerializer(apiObject *types.Serializer) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"orc_ser_de":     flattenOrcSerDe(apiObject.OrcSerDe),
		"parquet_ser_de": flattenParquetSerDe(apiObject.ParquetSerDe),
	}

	return []any{tfMap}
}

func flattenOrcSerDe(apiObject *types.OrcSerDe) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"bloom_filter_columns":     apiObject.BloomFilterColumns,
		"dictionary_key_threshold": aws.ToFloat64(apiObject.DictionaryKeyThreshold),
		"enable_padding":           aws.ToBool(apiObject.EnablePadding),
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference.

	tfMap["block_size_bytes"] = 268435456
	if apiObject.BlockSizeBytes != nil {
		tfMap["block_size_bytes"] = aws.ToInt32(apiObject.BlockSizeBytes)
	}

	tfMap["bloom_filter_false_positive_probability"] = 0.05
	if apiObject.BloomFilterFalsePositiveProbability != nil {
		tfMap["bloom_filter_false_positive_probability"] = aws.ToFloat64(apiObject.BloomFilterFalsePositiveProbability)
	}

	tfMap["compression"] = types.OrcCompressionSnappy
	if apiObject.Compression != "" {
		tfMap["compression"] = apiObject.Compression
	}

	tfMap["format_version"] = types.OrcFormatVersionV012
	if apiObject.FormatVersion != "" {
		tfMap["format_version"] = apiObject.FormatVersion
	}

	tfMap["padding_tolerance"] = 0.05
	if apiObject.PaddingTolerance != nil {
		tfMap["padding_tolerance"] = aws.ToFloat64(apiObject.PaddingTolerance)
	}

	tfMap["row_index_stride"] = 10000
	if apiObject.RowIndexStride != nil {
		tfMap["row_index_stride"] = aws.ToInt32(apiObject.RowIndexStride)
	}

	tfMap["stripe_size_bytes"] = 67108864
	if apiObject.StripeSizeBytes != nil {
		tfMap["stripe_size_bytes"] = aws.ToInt32(apiObject.StripeSizeBytes)
	}

	return []any{tfMap}
}

func flattenParquetSerDe(apiObject *types.ParquetSerDe) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"enable_dictionary_compression": aws.ToBool(apiObject.EnableDictionaryCompression),
		"max_padding_bytes":             aws.ToInt32(apiObject.MaxPaddingBytes),
	}

	// API omits default values
	// Return defaults that are not type zero values to prevent extraneous difference.

	tfMap["block_size_bytes"] = 268435456
	if apiObject.BlockSizeBytes != nil {
		tfMap["block_size_bytes"] = aws.ToInt32(apiObject.BlockSizeBytes)
	}

	tfMap["compression"] = types.ParquetCompressionSnappy
	if apiObject.Compression != "" {
		tfMap["compression"] = apiObject.Compression
	}

	tfMap["page_size_bytes"] = 1048576
	if apiObject.PageSizeBytes != nil {
		tfMap["page_size_bytes"] = aws.ToInt32(apiObject.PageSizeBytes)
	}

	tfMap["writer_version"] = types.ParquetWriterVersionV1
	if apiObject.WriterVersion != "" {
		tfMap["writer_version"] = apiObject.WriterVersion
	}

	return []any{tfMap}
}

func flattenSchemaConfiguration(apiObject *types.SchemaConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrCatalogID:    aws.ToString(apiObject.CatalogId),
		names.AttrDatabaseName: aws.ToString(apiObject.DatabaseName),
		names.AttrRegion:       aws.ToString(apiObject.Region),
		names.AttrRoleARN:      aws.ToString(apiObject.RoleARN),
		names.AttrTableName:    aws.ToString(apiObject.TableName),
		"version_id":           aws.ToString(apiObject.VersionId),
	}

	return []any{tfMap}
}

func flattenHTTPEndpointRequestConfiguration(apiObject *types.HttpEndpointRequestConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfListCommonAttributes := make([]any, 0)
	for _, apiObject := range apiObject.CommonAttributes {
		name := aws.ToString(apiObject.AttributeName)
		value := aws.ToString(apiObject.AttributeValue)

		tfListCommonAttributes = append(tfListCommonAttributes, map[string]any{
			names.AttrName:  name,
			names.AttrValue: value,
		})
	}

	tfMap := map[string]any{
		"common_attributes": tfListCommonAttributes,
		"content_encoding":  apiObject.ContentEncoding,
	}

	return []any{tfMap}
}

func flattenProcessingConfiguration(apiObject *types.ProcessingConfiguration, destinationType destinationType, roleARN string) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}

	tfListProcessors := make([]any, len(apiObject.Processors))
	for i, apiObject := range apiObject.Processors {
		typ := apiObject.Type
		tfListParameters := make([]any, 0)

		// It is necessary to explicitly filter this out
		// to prevent diffs during routine use and retain the ability
		// to show diffs if any field has drifted.
		defaultProcessorParameters := defaultProcessorParameters(destinationType, typ, roleARN)

		for _, apiObject := range apiObject.Parameters {
			name := apiObject.ParameterName
			value := aws.ToString(apiObject.ParameterValue)

			// Ignore defaults.
			if v, ok := defaultProcessorParameters[name]; ok && v == value {
				continue
			}

			tfListParameters = append(tfListParameters, map[string]any{
				"parameter_name":  name,
				"parameter_value": value,
			})
		}

		tfListProcessors[i] = map[string]any{
			names.AttrParameters: tfListParameters,
			names.AttrType:       typ,
		}
	}
	tfMap["processors"] = tfListProcessors

	return []any{tfMap}
}

func flattenSecretsManagerConfiguration(apiObject *types.SecretsManagerConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}

	if aws.ToBool(apiObject.Enabled) {
		tfMap["secret_arn"] = aws.ToString(apiObject.SecretARN)
		if apiObject.RoleARN != nil {
			tfMap[names.AttrRoleARN] = aws.ToString(apiObject.RoleARN)
		}
	}

	return []any{tfMap}
}

func flattenDynamicPartitioningConfiguration(apiObject *types.DynamicPartitioningConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}

	if v := apiObject.RetryOptions; v != nil {
		tfMap["retry_duration"] = aws.ToInt32(v.DurationInSeconds)
	}

	return []any{tfMap}
}

func flattenKinesisStreamSourceDescription(apiObject *types.KinesisStreamSourceDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"kinesis_stream_arn": aws.ToString(apiObject.KinesisStreamARN),
		names.AttrRoleARN:    aws.ToString(apiObject.RoleARN),
	}

	return []any{tfMap}
}

func flattenHTTPEndpointDestinationDescription(apiObject *types.HttpEndpointDestinationDescription, configuredAccessKey string) []any {
	if apiObject == nil {
		return []any{}
	}

	roleARN := aws.ToString(apiObject.RoleARN)
	tfMap := map[string]any{
		names.AttrAccessKey:             configuredAccessKey,
		"cloudwatch_logging_options":    flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions),
		names.AttrName:                  aws.ToString(apiObject.EndpointConfiguration.Name),
		"processing_configuration":      flattenProcessingConfiguration(apiObject.ProcessingConfiguration, destinationTypeHTTPEndpoint, roleARN),
		"request_configuration":         flattenHTTPEndpointRequestConfiguration(apiObject.RequestConfiguration),
		names.AttrRoleARN:               roleARN,
		"s3_backup_mode":                apiObject.S3BackupMode,
		"s3_configuration":              flattenS3DestinationDescription(apiObject.S3DestinationDescription),
		"secrets_manager_configuration": flattenSecretsManagerConfiguration(apiObject.SecretsManagerConfiguration),
		names.AttrURL:                   aws.ToString(apiObject.EndpointConfiguration.Url),
	}

	if v := apiObject.BufferingHints; v != nil {
		if v.IntervalInSeconds != nil {
			tfMap["buffering_interval"] = aws.ToInt32(v.IntervalInSeconds)
		}
		if v.SizeInMBs != nil {
			tfMap["buffering_size"] = aws.ToInt32(v.SizeInMBs)
		}
	}

	if v := apiObject.RetryOptions; v != nil {
		tfMap["retry_duration"] = aws.ToInt32(v.DurationInSeconds)
	}

	return []any{tfMap}
}

func flattenIcebergDestinationDescription(apiObject *types.IcebergDestinationDescription) []any {
	if apiObject == nil {
		return []any{}
	}

	roleARN := aws.ToString(apiObject.RoleARN)
	tfMap := map[string]any{
		"append_only":      aws.ToBool(apiObject.AppendOnly),
		"catalog_arn":      aws.ToString(apiObject.CatalogConfiguration.CatalogARN),
		names.AttrRoleARN:  roleARN,
		"s3_configuration": flattenS3DestinationDescription(apiObject.S3DestinationDescription),
	}

	if v := apiObject.BufferingHints; v != nil {
		if v.IntervalInSeconds != nil {
			tfMap["buffering_interval"] = aws.ToInt32(v.IntervalInSeconds)
		}
		if v.SizeInMBs != nil {
			tfMap["buffering_size"] = aws.ToInt32(v.SizeInMBs)
		}
	}

	if apiObject.CloudWatchLoggingOptions != nil {
		tfMap["cloudwatch_logging_options"] = flattenCloudWatchLoggingOptions(apiObject.CloudWatchLoggingOptions)
	}

	if v := apiObject.DestinationTableConfigurationList; v != nil {
		tfListTableConfigurations := make([]any, 0, len(v))
		for _, tableConfiguration := range v {
			tfListTableConfigurations = append(tfListTableConfigurations, map[string]any{
				names.AttrDatabaseName:   aws.ToString(tableConfiguration.DestinationDatabaseName),
				"s3_error_output_prefix": tableConfiguration.S3ErrorOutputPrefix,
				names.AttrTableName:      aws.ToString(tableConfiguration.DestinationTableName),
				"unique_keys":            tableConfiguration.UniqueKeys,
			})
		}
		tfMap["destination_table_configuration"] = tfListTableConfigurations
	}

	if v := apiObject.ProcessingConfiguration; v != nil {
		tfMap["processing_configuration"] = flattenProcessingConfiguration(v, destinationTypeIceberg, roleARN)
	}

	if v := apiObject.RetryOptions; v != nil {
		tfMap["retry_duration"] = aws.ToInt32(v.DurationInSeconds)
	}

	if apiObject.S3BackupMode != "" {
		tfMap["s3_backup_mode"] = apiObject.S3BackupMode
	}

	return []any{tfMap}
}

func expandDocumentIDOptions(tfMap map[string]any) *types.DocumentIdOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DocumentIdOptions{}

	if v, ok := tfMap["default_document_id_format"].(string); ok && v != "" {
		apiObject.DefaultDocumentIdFormat = types.DefaultDocumentIdFormat(v)
	}

	return apiObject
}

func flattenDocumentIDOptions(apiObject *types.DocumentIdOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"default_document_id_format": apiObject.DefaultDocumentIdFormat,
	}

	return tfMap
}

func flattenSnowflakeRoleConfiguration(apiObject *types.SnowflakeRoleConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}

	if aws.ToBool(apiObject.Enabled) {
		tfMap["snowflake_role"] = aws.ToString(apiObject.SnowflakeRole)
	}

	return []any{tfMap}
}

func flattenSnowflakeVPCConfiguration(apiObject *types.SnowflakeVpcConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"private_link_vpce_id": aws.ToString(apiObject.PrivateLinkVpceId),
	}

	return []any{tfMap}
}

func isDeliveryStreamOptionDisabled(v any) bool {
	tfList := v.([]any)
	if len(tfList) == 0 || tfList[0] == nil {
		return true
	}
	tfMap := tfList[0].(map[string]any)

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
