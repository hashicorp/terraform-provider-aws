// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
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

// @SDKResource("aws_kinesisanalyticsv2_application", name="Application")
// @Tags(identifierAttribute="arn")
func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationCreate,
		ReadWithoutTimeout:   resourceApplicationRead,
		UpdateWithoutTimeout: resourceApplicationUpdate,
		DeleteWithoutTimeout: resourceApplicationDelete,

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			customdiff.ForceNewIfChange("application_configuration.0.sql_application_configuration.0.input", func(_ context.Context, old, new, meta interface{}) bool {
				// An existing input configuration cannot be deleted.
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
		),

		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"application_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"application_code_configuration": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"code_content": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"s3_content_location": {
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
																"file_key": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringLenBetween(1, 1024),
																},
																"object_version": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
															},
														},
														ConflictsWith: []string{"application_configuration.0.application_code_configuration.0.code_content.0.text_content"},
													},
													"text_content": {
														Type:          schema.TypeString,
														Optional:      true,
														ValidateFunc:  validation.StringLenBetween(0, 102400),
														ConflictsWith: []string{"application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location"},
													},
												},
											},
										},
										"code_content_type": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.CodeContentType](),
										},
									},
								},
							},
							"application_snapshot_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"snapshots_enabled": {
											Type:     schema.TypeBool,
											Required: true,
										},
									},
								},
								ConflictsWith: []string{"application_configuration.0.sql_application_configuration"},
							},
							"environment_properties": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"property_group": {
											Type:     schema.TypeSet,
											Required: true,
											MaxItems: 50,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"property_group_id": {
														Type:     schema.TypeString,
														Required: true,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 50),
															validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
														),
													},
													"property_map": {
														Type:     schema.TypeMap,
														Required: true,
														Elem:     &schema.Schema{Type: schema.TypeString},
													},
												},
											},
										},
									},
								},
								ConflictsWith: []string{"application_configuration.0.sql_application_configuration"},
							},
							"flink_application_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"checkpoint_configuration": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"checkpointing_enabled": {
														Type:     schema.TypeBool,
														Optional: true,
														Computed: true,
													},
													"checkpoint_interval": {
														Type:         schema.TypeInt,
														Optional:     true,
														Computed:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"configuration_type": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.ConfigurationType](),
													},
													"min_pause_between_checkpoints": {
														Type:         schema.TypeInt,
														Optional:     true,
														Computed:     true,
														ValidateFunc: validation.IntAtLeast(0),
													},
												},
											},
										},
										"monitoring_configuration": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"configuration_type": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.ConfigurationType](),
													},
													"log_level": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[awstypes.LogLevel](),
													},
													"metrics_level": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[awstypes.MetricsLevel](),
													},
												},
											},
										},
										"parallelism_configuration": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"auto_scaling_enabled": {
														Type:     schema.TypeBool,
														Optional: true,
														Computed: true,
													},
													"configuration_type": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.ConfigurationType](),
													},
													"parallelism": {
														Type:         schema.TypeInt,
														Optional:     true,
														Computed:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"parallelism_per_kpu": {
														Type:         schema.TypeInt,
														Optional:     true,
														Computed:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
												},
											},
										},
									},
								},
								ConflictsWith: []string{"application_configuration.0.sql_application_configuration"},
							},
							"run_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								Computed: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"application_restore_configuration": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"application_restore_type": {
														Type:             schema.TypeString,
														Optional:         true,
														Computed:         true,
														ValidateDiagFunc: enum.Validate[awstypes.ApplicationRestoreType](),
													},
													"snapshot_name": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										"flink_run_configuration": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"allow_non_restored_state": {
														Type:     schema.TypeBool,
														Optional: true,
														Computed: true,
													},
												},
											},
										},
									},
								},
								ConflictsWith: []string{"application_configuration.0.sql_application_configuration"},
							},
							"sql_application_configuration": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"input": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"in_app_stream_names": {
														Type:     schema.TypeList,
														Computed: true,
														Elem:     &schema.Schema{Type: schema.TypeString},
													},
													"input_id": {
														Type:     schema.TypeString,
														Computed: true,
													},
													"input_parallelism": {
														Type:     schema.TypeList,
														Optional: true,
														Computed: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"count": {
																	Type:         schema.TypeInt,
																	Optional:     true,
																	Computed:     true,
																	ValidateFunc: validation.IntBetween(1, 64),
																},
															},
														},
													},
													"input_processing_configuration": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"input_lambda_processor": {
																	Type:     schema.TypeList,
																	Required: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			names.AttrResourceARN: {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: verify.ValidARN,
																			},
																		},
																	},
																},
															},
														},
													},
													"input_schema": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"record_column": {
																	Type:     schema.TypeList,
																	Required: true,
																	MaxItems: 1000,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"mapping": {
																				Type:     schema.TypeString,
																				Optional: true,
																			},
																			names.AttrName: {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[^-\s<>&]+$`), "must not include hyphen, whitespace, angle bracket, or ampersand characters"),
																			},
																			"sql_type": {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																		},
																	},
																},
																"record_encoding": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"record_format": {
																	Type:     schema.TypeList,
																	Required: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"mapping_parameters": {
																				Type:     schema.TypeList,
																				Required: true,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"csv_mapping_parameters": {
																							Type:     schema.TypeList,
																							Optional: true,
																							MaxItems: 1,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"record_column_delimiter": {
																										Type:     schema.TypeString,
																										Required: true,
																									},
																									"record_row_delimiter": {
																										Type:     schema.TypeString,
																										Required: true,
																									},
																								},
																							},
																							ExactlyOneOf: []string{
																								"application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters",
																								"application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters",
																							},
																						},
																						"json_mapping_parameters": {
																							Type:     schema.TypeList,
																							Optional: true,
																							MaxItems: 1,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"record_row_path": {
																										Type:     schema.TypeString,
																										Required: true,
																									},
																								},
																							},
																							ExactlyOneOf: []string{
																								"application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters",
																								"application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters",
																							},
																						},
																					},
																				},
																			},
																			"record_format_type": {
																				Type:             schema.TypeString,
																				Required:         true,
																				ValidateDiagFunc: enum.Validate[awstypes.RecordFormatType](),
																			},
																		},
																	},
																},
															},
														},
													},
													"input_starting_position_configuration": {
														Type:     schema.TypeList,
														Optional: true,
														Computed: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"input_starting_position": {
																	Type:             schema.TypeString,
																	Optional:         true,
																	Computed:         true,
																	ValidateDiagFunc: enum.Validate[awstypes.InputStartingPosition](),
																},
															},
														},
													},
													"kinesis_firehose_input": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrResourceARN: {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: verify.ValidARN,
																},
															},
														},
														ExactlyOneOf: []string{
															"application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input",
															"application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input",
														},
													},
													"kinesis_streams_input": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrResourceARN: {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: verify.ValidARN,
																},
															},
														},
														ExactlyOneOf: []string{
															"application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input",
															"application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input",
														},
													},
													names.AttrNamePrefix: {
														Type:     schema.TypeString,
														Required: true,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 32),
															validation.StringMatch(regexache.MustCompile(`^[^-\s<>&]+$`), "must not include hyphen, whitespace, angle bracket, or ampersand characters"),
														),
													},
												},
											},
										},
										"output": {
											Type:     schema.TypeSet,
											Optional: true,
											MaxItems: 3,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"destination_schema": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"record_format_type": {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[awstypes.RecordFormatType](),
																},
															},
														},
													},
													"kinesis_firehose_output": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrResourceARN: {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: verify.ValidARN,
																},
															},
														},
													},
													"kinesis_streams_output": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrResourceARN: {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: verify.ValidARN,
																},
															},
														},
													},
													"lambda_output": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrResourceARN: {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: verify.ValidARN,
																},
															},
														},
													},
													names.AttrName: {
														Type:     schema.TypeString,
														Required: true,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 32),
															validation.StringMatch(regexache.MustCompile(`^[^-\s<>&]+$`), "must not include hyphen, whitespace, angle bracket, or ampersand characters"),
														),
													},
													"output_id": {
														Type:     schema.TypeString,
														Computed: true,
													},
												},
											},
										},
										"reference_data_source": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"reference_id": {
														Type:     schema.TypeString,
														Computed: true,
													},
													"reference_schema": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"record_column": {
																	Type:     schema.TypeList,
																	Required: true,
																	MaxItems: 1000,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"mapping": {
																				Type:     schema.TypeString,
																				Optional: true,
																			},
																			names.AttrName: {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[^-\s<>&]+$`), "must not include hyphen, whitespace, angle bracket, or ampersand characters"),
																			},
																			"sql_type": {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																		},
																	},
																},
																"record_encoding": {
																	Type:     schema.TypeString,
																	Optional: true,
																},
																"record_format": {
																	Type:     schema.TypeList,
																	Required: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"mapping_parameters": {
																				Type:     schema.TypeList,
																				Required: true,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"csv_mapping_parameters": {
																							Type:     schema.TypeList,
																							Optional: true,
																							MaxItems: 1,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"record_column_delimiter": {
																										Type:     schema.TypeString,
																										Required: true,
																									},
																									"record_row_delimiter": {
																										Type:     schema.TypeString,
																										Required: true,
																									},
																								},
																							},
																							ExactlyOneOf: []string{
																								"application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters",
																								"application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters",
																							},
																						},
																						"json_mapping_parameters": {
																							Type:     schema.TypeList,
																							Optional: true,
																							MaxItems: 1,
																							Elem: &schema.Resource{
																								Schema: map[string]*schema.Schema{
																									"record_row_path": {
																										Type:     schema.TypeString,
																										Required: true,
																									},
																								},
																							},
																							ExactlyOneOf: []string{
																								"application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters",
																								"application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters",
																							},
																						},
																					},
																				},
																			},
																			"record_format_type": {
																				Type:             schema.TypeString,
																				Required:         true,
																				ValidateDiagFunc: enum.Validate[awstypes.RecordFormatType](),
																			},
																		},
																	},
																},
															},
														},
													},
													"s3_reference_data_source": {
														Type:     schema.TypeList,
														Required: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"bucket_arn": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: verify.ValidARN,
																},
																"file_key": {
																	Type:     schema.TypeString,
																	Required: true,
																},
															},
														},
													},
													names.AttrTableName: {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringLenBetween(1, 32),
													},
												},
											},
										},
									},
								},
								ConflictsWith: []string{
									"application_configuration.0.application_snapshot_configuration",
									"application_configuration.0.environment_properties",
									"application_configuration.0.flink_application_configuration",
									"application_configuration.0.run_configuration",
									"application_configuration.0.vpc_configuration",
								},
							},
							names.AttrVPCConfiguration: {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrSecurityGroupIDs: {
											Type:     schema.TypeSet,
											Required: true,
											MinItems: 1,
											MaxItems: 5,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrSubnetIDs: {
											Type:     schema.TypeSet,
											Required: true,
											MinItems: 1,
											MaxItems: 16,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										"vpc_configuration_id": {
											Type:     schema.TypeString,
											Computed: true,
										},
										names.AttrVPCID: {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
								ConflictsWith: []string{"application_configuration.0.sql_application_configuration"},
							},
						},
					},
				},
				"application_mode": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ApplicationMode](),
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"cloudwatch_logging_options": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cloudwatch_logging_option_id": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"log_stream_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
				"create_timestamp": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(0, 1024),
				},
				"force_stop": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"last_update_timestamp": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
					),
				},
				"runtime_environment": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.RuntimeEnvironment](),
				},
				"service_execution_role": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"start_application": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"version_id": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			}
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	applicationName := d.Get(names.AttrName).(string)
	input := &kinesisanalyticsv2.CreateApplicationInput{
		ApplicationConfiguration: expandApplicationConfiguration(d.Get("application_configuration").([]interface{})),
		ApplicationDescription:   aws.String(d.Get(names.AttrDescription).(string)),
		ApplicationName:          aws.String(applicationName),
		CloudWatchLoggingOptions: expandCloudWatchLoggingOptions(d.Get("cloudwatch_logging_options").([]interface{})),
		RuntimeEnvironment:       awstypes.RuntimeEnvironment(d.Get("runtime_environment").(string)),
		ServiceExecutionRole:     aws.String(d.Get("service_execution_role").(string)),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("application_mode"); ok {
		input.ApplicationMode = awstypes.ApplicationMode(v.(string))
	}

	output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.CreateApplicationOutput, error) {
		return conn.CreateApplication(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Analytics v2 Application (%s): %s", applicationName, err)
	}

	d.SetId(aws.ToString(output.ApplicationDetail.ApplicationARN))
	// CreateTimestamp is required for deletion, so persist to state now in case of subsequent errors and destroy being called without refresh.
	d.Set("create_timestamp", aws.ToTime(output.ApplicationDetail.CreateTimestamp).Format(time.RFC3339))

	if _, ok := d.GetOk("start_application"); ok {
		if err := startApplication(ctx, conn, expandStartApplicationInput(d), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	application, err := findApplicationDetailByName(ctx, conn, d.Get(names.AttrName).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Analytics v2 Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics v2 Application (%s): %s", d.Id(), err)
	}

	if err := d.Set("application_configuration", flattenApplicationConfigurationDescription(application.ApplicationConfigurationDescription)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting application_configuration: %s", err)
	}
	d.Set("application_mode", application.ApplicationMode)
	d.Set(names.AttrARN, application.ApplicationARN)
	if err := d.Set("cloudwatch_logging_options", flattenCloudWatchLoggingOptionDescriptions(application.CloudWatchLoggingOptionDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cloudwatch_logging_options: %s", err)
	}
	d.Set("create_timestamp", aws.ToTime(application.CreateTimestamp).Format(time.RFC3339))
	d.Set(names.AttrDescription, application.ApplicationDescription)
	d.Set("last_update_timestamp", aws.ToTime(application.LastUpdateTimestamp).Format(time.RFC3339))
	d.Set(names.AttrName, application.ApplicationName)
	d.Set("runtime_environment", application.RuntimeEnvironment)
	d.Set("service_execution_role", application.ServiceExecutionRole)
	d.Set(names.AttrStatus, application.ApplicationStatus)
	d.Set("version_id", application.ApplicationVersionId)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)
	applicationName := d.Get(names.AttrName).(string)

	if d.HasChanges("application_configuration", "cloudwatch_logging_options", "service_execution_role") {
		currentApplicationVersionID := int64(d.Get("version_id").(int))
		updateApplication := false

		input := &kinesisanalyticsv2.UpdateApplicationInput{
			ApplicationName: aws.String(applicationName),
		}

		if d.HasChange("application_configuration") {
			applicationConfigurationUpdate := &awstypes.ApplicationConfigurationUpdate{}

			if d.HasChange("application_configuration.0.application_code_configuration") {
				applicationConfigurationUpdate.ApplicationCodeConfigurationUpdate = expandApplicationCodeConfigurationUpdate(d.Get("application_configuration.0.application_code_configuration").([]interface{}))

				updateApplication = true
			}

			if d.HasChange("application_configuration.0.application_snapshot_configuration") {
				applicationConfigurationUpdate.ApplicationSnapshotConfigurationUpdate = expandApplicationSnapshotConfigurationUpdate(d.Get("application_configuration.0.application_snapshot_configuration").([]interface{}))

				updateApplication = true
			}

			if d.HasChange("application_configuration.0.environment_properties") {
				applicationConfigurationUpdate.EnvironmentPropertyUpdates = expandEnvironmentPropertyUpdates(d.Get("application_configuration.0.environment_properties").([]interface{}))

				updateApplication = true
			}

			if d.HasChange("application_configuration.0.flink_application_configuration") {
				applicationConfigurationUpdate.FlinkApplicationConfigurationUpdate = expandApplicationFlinkApplicationConfigurationUpdate(d.Get("application_configuration.0.flink_application_configuration").([]interface{}))

				updateApplication = true
			}

			if d.HasChange("application_configuration.0.sql_application_configuration") {
				sqlApplicationConfigurationUpdate := &awstypes.SqlApplicationConfigurationUpdate{}

				if d.HasChange("application_configuration.0.sql_application_configuration.0.input") {
					o, n := d.GetChange("application_configuration.0.sql_application_configuration.0.input")

					if len(o.([]interface{})) == 0 {
						// Add new input.
						input := &kinesisanalyticsv2.AddApplicationInputInput{
							ApplicationName:             aws.String(applicationName),
							CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
							Input:                       expandInput(n.([]interface{})),
						}

						output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.AddApplicationInputOutput, error) {
							return conn.AddApplicationInput(ctx, input)
						})

						if err != nil {
							return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics v2 Application (%s) input: %s", d.Id(), err)
						}

						if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
							return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
						}

						currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
					} else if len(n.([]interface{})) == 0 {
						// The existing input cannot be deleted.
						// This should be handled by the CustomizeDiff function above.
						return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application (%s) input", d.Id())
					} else {
						// Update existing input.
						inputUpdate := expandInputUpdate(n.([]interface{}))

						if d.HasChange("application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration") {
							o, n := d.GetChange("application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration")

							// Update of existing input processing configuration is handled via the updating of the existing input.

							if len(o.([]interface{})) == 0 {
								// Add new input processing configuration.
								input := &kinesisanalyticsv2.AddApplicationInputProcessingConfigurationInput{
									ApplicationName:              aws.String(applicationName),
									CurrentApplicationVersionId:  aws.Int64(currentApplicationVersionID),
									InputId:                      inputUpdate.InputId,
									InputProcessingConfiguration: expandInputProcessingConfiguration(n.([]interface{})),
								}

								output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.AddApplicationInputProcessingConfigurationOutput, error) {
									return conn.AddApplicationInputProcessingConfiguration(ctx, input)
								})

								if err != nil {
									return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics v2 Application (%s) input processing configuration: %s", d.Id(), err)
								}

								if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
									return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
								}

								currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
							} else if len(n.([]interface{})) == 0 {
								// Delete existing input processing configuration.
								input := &kinesisanalyticsv2.DeleteApplicationInputProcessingConfigurationInput{
									ApplicationName:             aws.String(applicationName),
									CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
									InputId:                     inputUpdate.InputId,
								}

								output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.DeleteApplicationInputProcessingConfigurationOutput, error) {
									return conn.DeleteApplicationInputProcessingConfiguration(ctx, input)
								})

								if err != nil {
									return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application (%s) input processing configuration: %s", d.Id(), err)
								}

								if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
									return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
								}

								currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
							}
						}

						sqlApplicationConfigurationUpdate.InputUpdates = []awstypes.InputUpdate{inputUpdate}

						updateApplication = true
					}
				}

				if d.HasChange("application_configuration.0.sql_application_configuration.0.output") {
					o, n := d.GetChange("application_configuration.0.sql_application_configuration.0.output")
					os, ns := o.(*schema.Set), n.(*schema.Set)

					additions := []interface{}{}
					deletions := []string{}

					// Additions.
					for _, vOutput := range ns.Difference(os).List() {
						if v, ok := vOutput.(map[string]interface{})["output_id"].(string); ok && v != "" {
							// Shouldn't be attempting to add an output with an ID.
							log.Printf("[WARN] Attempting to add invalid Kinesis Analytics v2 Application (%s) output: %#v", d.Id(), vOutput)
						} else {
							additions = append(additions, vOutput)
						}
					}

					// Deletions.
					for _, vOutput := range os.Difference(ns).List() {
						if v, ok := vOutput.(map[string]interface{})["output_id"].(string); ok && v != "" {
							deletions = append(deletions, v)
						} else {
							// Shouldn't be attempting to delete an output without an ID.
							log.Printf("[WARN] Attempting to delete invalid Kinesis Analytics v2 Application (%s) output: %#v", d.Id(), vOutput)
						}
					}

					// Delete existing outputs.
					for _, v := range deletions {
						input := &kinesisanalyticsv2.DeleteApplicationOutputInput{
							ApplicationName:             aws.String(applicationName),
							CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
							OutputId:                    aws.String(v),
						}

						output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.DeleteApplicationOutputOutput, error) {
							return conn.DeleteApplicationOutput(ctx, input)
						})

						if err != nil {
							return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application (%s) output: %s", d.Id(), err)
						}

						if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
							return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
						}

						currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
					}

					// Add new outputs.
					for _, v := range additions {
						input := &kinesisanalyticsv2.AddApplicationOutputInput{
							ApplicationName:             aws.String(applicationName),
							CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
							Output:                      expandOutput(v),
						}

						output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.AddApplicationOutputOutput, error) {
							return conn.AddApplicationOutput(ctx, input)
						})

						if err != nil {
							return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics v2 Application (%s) output: %s", d.Id(), err)
						}

						if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
							return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
						}

						currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
					}
				}

				if d.HasChange("application_configuration.0.sql_application_configuration.0.reference_data_source") {
					o, n := d.GetChange("application_configuration.0.sql_application_configuration.0.reference_data_source")

					if len(o.([]interface{})) == 0 {
						// Add new reference data source.
						input := &kinesisanalyticsv2.AddApplicationReferenceDataSourceInput{
							ApplicationName:             aws.String(applicationName),
							CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
							ReferenceDataSource:         expandReferenceDataSource(n.([]interface{})),
						}

						output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.AddApplicationReferenceDataSourceOutput, error) {
							return conn.AddApplicationReferenceDataSource(ctx, input)
						})

						if err != nil {
							return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics v2 Application (%s) reference data source: %s", d.Id(), err)
						}

						if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
							return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
						}

						currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
					} else if len(n.([]interface{})) == 0 {
						// Delete existing reference data source.
						mOldReferenceDataSource := o.([]interface{})[0].(map[string]interface{})

						input := &kinesisanalyticsv2.DeleteApplicationReferenceDataSourceInput{
							ApplicationName:             aws.String(applicationName),
							CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
							ReferenceId:                 aws.String(mOldReferenceDataSource["reference_id"].(string)),
						}

						output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.DeleteApplicationReferenceDataSourceOutput, error) {
							return conn.DeleteApplicationReferenceDataSource(ctx, input)
						})

						if err != nil {
							return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application (%s) reference data source: %s", d.Id(), err)
						}

						if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
							return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
						}

						currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
					} else {
						// Update existing reference data source.
						referenceDataSourceUpdate := expandReferenceDataSourceUpdate(n.([]interface{}))

						sqlApplicationConfigurationUpdate.ReferenceDataSourceUpdates = []awstypes.ReferenceDataSourceUpdate{referenceDataSourceUpdate}

						updateApplication = true
					}
				}

				applicationConfigurationUpdate.SqlApplicationConfigurationUpdate = sqlApplicationConfigurationUpdate
			}

			if d.HasChange("application_configuration.0.vpc_configuration") {
				o, n := d.GetChange("application_configuration.0.vpc_configuration")

				if len(o.([]interface{})) == 0 {
					// Add new VPC configuration.
					input := &kinesisanalyticsv2.AddApplicationVpcConfigurationInput{
						ApplicationName:             aws.String(applicationName),
						CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
						VpcConfiguration:            expandVPCConfiguration(n.([]interface{})),
					}

					output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.AddApplicationVpcConfigurationOutput, error) {
						return conn.AddApplicationVpcConfiguration(ctx, input)
					})

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics v2 Application (%s) VPC configuration: %s", d.Id(), err)
					}

					if operationID := aws.ToString(output.OperationId); operationID != "" {
						if _, err := waitApplicationOperationSucceeded(ctx, conn, applicationName, operationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
							return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) operation (%s) success: %s", applicationName, operationID, err)
						}
					}

					if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
					}

					currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
				} else if len(n.([]interface{})) == 0 {
					// Delete existing VPC configuration.
					mOldVpcConfiguration := o.([]interface{})[0].(map[string]interface{})

					input := &kinesisanalyticsv2.DeleteApplicationVpcConfigurationInput{
						ApplicationName:             aws.String(applicationName),
						CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
						VpcConfigurationId:          aws.String(mOldVpcConfiguration["vpc_configuration_id"].(string)),
					}

					output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.DeleteApplicationVpcConfigurationOutput, error) {
						return conn.DeleteApplicationVpcConfiguration(ctx, input)
					})

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application (%s) VPC configuration: %s", d.Id(), err)
					}

					if operationID := aws.ToString(output.OperationId); operationID != "" {
						if _, err := waitApplicationOperationSucceeded(ctx, conn, applicationName, operationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
							return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) operation (%s) success: %s", applicationName, operationID, err)
						}
					}

					if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
					}

					currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
				} else {
					// Update existing VPC configuration.
					vpcConfigurationUpdate := expandVPCConfigurationUpdate(n.([]interface{}))

					applicationConfigurationUpdate.VpcConfigurationUpdates = []awstypes.VpcConfigurationUpdate{vpcConfigurationUpdate}

					updateApplication = true
				}
			}

			if d.HasChange("application_configuration.0.run_configuration") {
				application, err := findApplicationDetailByName(ctx, conn, applicationName)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics v2 Application (%s): %s", applicationName, err)
				}

				if actual, expected := application.ApplicationStatus, awstypes.ApplicationStatusRunning; actual == expected {
					input.RunConfigurationUpdate = expandRunConfigurationUpdate(d.Get("application_configuration.0.run_configuration").([]interface{}))

					updateApplication = true
				}
			}

			input.ApplicationConfigurationUpdate = applicationConfigurationUpdate
		}

		if d.HasChange("cloudwatch_logging_options") {
			o, n := d.GetChange("cloudwatch_logging_options")

			if len(o.([]interface{})) == 0 {
				// Add new CloudWatch logging options.
				mNewCloudWatchLoggingOption := n.([]interface{})[0].(map[string]interface{})

				input := &kinesisanalyticsv2.AddApplicationCloudWatchLoggingOptionInput{
					ApplicationName: aws.String(applicationName),
					CloudWatchLoggingOption: &awstypes.CloudWatchLoggingOption{
						LogStreamARN: aws.String(mNewCloudWatchLoggingOption["log_stream_arn"].(string)),
					},
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
				}

				output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.AddApplicationCloudWatchLoggingOptionOutput, error) {
					return conn.AddApplicationCloudWatchLoggingOption(ctx, input)
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics v2 Application (%s) CloudWatch logging option: %s", d.Id(), err)
				}

				if operationID := aws.ToString(output.OperationId); operationID != "" {
					if _, err := waitApplicationOperationSucceeded(ctx, conn, applicationName, operationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) operation (%s) success: %s", applicationName, operationID, err)
					}
				}

				if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
				}

				currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
			} else if len(n.([]interface{})) == 0 {
				// Delete existing CloudWatch logging options.
				mOldCloudWatchLoggingOption := o.([]interface{})[0].(map[string]interface{})

				input := &kinesisanalyticsv2.DeleteApplicationCloudWatchLoggingOptionInput{
					ApplicationName:             aws.String(applicationName),
					CloudWatchLoggingOptionId:   aws.String(mOldCloudWatchLoggingOption["cloudwatch_logging_option_id"].(string)),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
				}

				output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.DeleteApplicationCloudWatchLoggingOptionOutput, error) {
					return conn.DeleteApplicationCloudWatchLoggingOption(ctx, input)
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application (%s) CloudWatch logging option: %s", d.Id(), err)
				}

				if operationID := aws.ToString(output.OperationId); operationID != "" {
					if _, err := waitApplicationOperationSucceeded(ctx, conn, applicationName, operationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) operation (%s) success: %s", applicationName, operationID, err)
					}
				}

				if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
				}

				currentApplicationVersionID = aws.ToInt64(output.ApplicationVersionId)
			} else {
				// Update existing CloudWatch logging options.
				mOldCloudWatchLoggingOption := o.([]interface{})[0].(map[string]interface{})
				mNewCloudWatchLoggingOption := n.([]interface{})[0].(map[string]interface{})

				input.CloudWatchLoggingOptionUpdates = []awstypes.CloudWatchLoggingOptionUpdate{
					{
						CloudWatchLoggingOptionId: aws.String(mOldCloudWatchLoggingOption["cloudwatch_logging_option_id"].(string)),
						LogStreamARNUpdate:        aws.String(mNewCloudWatchLoggingOption["log_stream_arn"].(string)),
					},
				}

				updateApplication = true
			}
		}

		if d.HasChange("service_execution_role") {
			input.ServiceExecutionRoleUpdate = aws.String(d.Get("service_execution_role").(string))

			updateApplication = true
		}

		if updateApplication {
			input.CurrentApplicationVersionId = aws.Int64(currentApplicationVersionID)

			output, err := waitIAMPropagation(ctx, func() (*kinesisanalyticsv2.UpdateApplicationOutput, error) {
				return conn.UpdateApplication(ctx, input)
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Kinesis Analytics v2 Application (%s): %s", d.Id(), err)
			}

			if operationID := aws.ToString(output.OperationId); operationID != "" {
				if _, err := waitApplicationOperationSucceeded(ctx, conn, applicationName, operationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) operation (%s) success: %s", applicationName, operationID, err)
				}
			}

			if _, err := waitApplicationUpdated(ctx, conn, applicationName, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) to update: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("start_application") {
		if _, ok := d.GetOk("start_application"); ok {
			if err := startApplication(ctx, conn, expandStartApplicationInput(d), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			if err := stopApplication(ctx, conn, expandStopApplicationInput(d), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	createTimestamp, err := time.Parse(time.RFC3339, d.Get("create_timestamp").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application (%s): parsing create_timestamp: %s", d.Id(), err)
	}

	applicationName := d.Get(names.AttrName).(string)

	log.Printf("[DEBUG] Deleting Kinesis Analytics v2 Application (%s)", d.Id())
	_, err = conn.DeleteApplication(ctx, &kinesisanalyticsv2.DeleteApplicationInput{
		ApplicationName: aws.String(applicationName),
		CreateTimestamp: aws.Time(createTimestamp),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application (%s): %s", d.Id(), err)
	}

	if _, err := waitApplicationDeleted(ctx, conn, applicationName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceApplicationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	arn, err := arn.Parse(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("parsing ARN %q: %w", d.Id(), err)
	}

	// application/<name>
	parts := strings.Split(arn.Resource, "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("unexpected ARN format: %q", d.Id())
	}

	d.Set(names.AttrName, parts[1])

	return []*schema.ResourceData{d}, nil
}

func startApplication(ctx context.Context, conn *kinesisanalyticsv2.Client, input *kinesisanalyticsv2.StartApplicationInput, timeout time.Duration) error {
	applicationName := aws.ToString(input.ApplicationName)

	application, err := findApplicationDetailByName(ctx, conn, applicationName)

	if err != nil {
		return fmt.Errorf("reading Kinesis Analytics v2 Application (%s): %w", applicationName, err)
	}

	applicationARN := aws.ToString(application.ApplicationARN)

	if actual, expected := application.ApplicationStatus, awstypes.ApplicationStatusReady; actual != expected {
		log.Printf("[DEBUG] Kinesis Analytics v2 Application (%s) has status %s. An application can only be started if it's in the %s state", applicationARN, actual, expected)
		return nil
	}

	if len(input.RunConfiguration.SqlRunConfigurations) > 0 {
		if application.ApplicationConfigurationDescription.SqlApplicationConfigurationDescription != nil &&
			len(application.ApplicationConfigurationDescription.SqlApplicationConfigurationDescription.InputDescriptions) == 0 {
			log.Printf("[DEBUG] Kinesis Analytics v2 Application (%s) has no input description", applicationARN)
			return nil
		}

		input.RunConfiguration.SqlRunConfigurations[0].InputId = application.ApplicationConfigurationDescription.SqlApplicationConfigurationDescription.InputDescriptions[0].InputId
	}

	output, err := conn.StartApplication(ctx, input)

	if err != nil {
		return fmt.Errorf("starting Kinesis Analytics v2 Application (%s):  %w", applicationARN, err)
	}

	if operationID := aws.ToString(output.OperationId); operationID != "" {
		if _, err := waitApplicationOperationSucceeded(ctx, conn, applicationName, operationID, timeout); err != nil {
			return fmt.Errorf("waiting for Kinesis Analytics v2 Application (%s) operation (%s) success: %w", applicationName, operationID, err)
		}
	}

	if _, err := waitApplicationStarted(ctx, conn, applicationName, timeout); err != nil {
		return fmt.Errorf("waiting for Kinesis Analytics v2 Application (%s) start: %w", applicationARN, err)
	}

	return nil
}

func stopApplication(ctx context.Context, conn *kinesisanalyticsv2.Client, input *kinesisanalyticsv2.StopApplicationInput, timeout time.Duration) error {
	applicationName := aws.ToString(input.ApplicationName)

	application, err := findApplicationDetailByName(ctx, conn, applicationName)

	if err != nil {
		return fmt.Errorf("reading Kinesis Analytics v2 Application (%s): %w", applicationName, err)
	}

	applicationARN := aws.ToString(application.ApplicationARN)

	if actual, expected := application.ApplicationStatus, awstypes.ApplicationStatusRunning; actual != expected {
		log.Printf("[DEBUG] Kinesis Analytics v2 Application (%s) has status %s. An application can only be stopped if it's in the %s state", applicationARN, actual, expected)
		return nil
	}

	output, err := conn.StopApplication(ctx, input)

	if err != nil {
		return fmt.Errorf("stopping Kinesis Analytics v2 Application (%s):  %w", applicationARN, err)
	}

	if operationID := aws.ToString(output.OperationId); operationID != "" {
		if _, err := waitApplicationOperationSucceeded(ctx, conn, applicationName, operationID, timeout); err != nil {
			return fmt.Errorf("waiting for Kinesis Analytics v2 Application (%s) operation (%s) success: %w", applicationName, operationID, err)
		}
	}

	if _, err := waitApplicationStopped(ctx, conn, applicationName, timeout); err != nil {
		return fmt.Errorf("waiting for Kinesis Analytics v2 Application (%s) stop: %w", applicationARN, err)
	}

	return nil
}

func findApplicationDetailByName(ctx context.Context, conn *kinesisanalyticsv2.Client, name string) (*awstypes.ApplicationDetail, error) {
	input := &kinesisanalyticsv2.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}

	return findApplicationDetail(ctx, conn, input)
}

func findApplicationDetail(ctx context.Context, conn *kinesisanalyticsv2.Client, input *kinesisanalyticsv2.DescribeApplicationInput) (*awstypes.ApplicationDetail, error) {
	output, err := conn.DescribeApplication(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationDetail == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApplicationDetail, nil
}

func statusApplication(ctx context.Context, conn *kinesisanalyticsv2.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findApplicationDetailByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ApplicationStatus), nil
	}
}

func waitApplicationStarted(ctx context.Context, conn *kinesisanalyticsv2.Client, name string, timeout time.Duration) (*awstypes.ApplicationDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusStarting),
		Target:  enum.Slice(awstypes.ApplicationStatusRunning),
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApplicationDetail); ok {
		return output, err
	}

	return nil, err
}

func waitApplicationStopped(ctx context.Context, conn *kinesisanalyticsv2.Client, name string, timeout time.Duration) (*awstypes.ApplicationDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusForceStopping, awstypes.ApplicationStatusStopping),
		Target:  enum.Slice(awstypes.ApplicationStatusReady),
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApplicationDetail); ok {
		return output, err
	}

	return nil, err
}

func waitApplicationUpdated(ctx context.Context, conn *kinesisanalyticsv2.Client, name string, timeout time.Duration) (*awstypes.ApplicationDetail, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusUpdating),
		Target:  enum.Slice(awstypes.ApplicationStatusReady, awstypes.ApplicationStatusRunning),
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApplicationDetail); ok {
		return output, err
	}

	return nil, err
}

func waitApplicationDeleted(ctx context.Context, conn *kinesisanalyticsv2.Client, name string, timeout time.Duration) (*awstypes.ApplicationDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationStatusDeleting),
		Target:  []string{},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApplicationDetail); ok {
		return output, err
	}

	return nil, err
}

func findApplicationOperationByTwoPartKey(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, operationID string) (*awstypes.ApplicationOperationInfoDetails, error) {
	input := &kinesisanalyticsv2.DescribeApplicationOperationInput{
		ApplicationName: aws.String(applicationName),
		OperationId:     aws.String(operationID),
	}

	return findApplicationOperation(ctx, conn, input)
}

func findApplicationOperation(ctx context.Context, conn *kinesisanalyticsv2.Client, input *kinesisanalyticsv2.DescribeApplicationOperationInput) (*awstypes.ApplicationOperationInfoDetails, error) {
	output, err := conn.DescribeApplicationOperation(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationOperationInfoDetails == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApplicationOperationInfoDetails, nil
}

func statusApplicationOperation(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, operationID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findApplicationOperationByTwoPartKey(ctx, conn, applicationName, operationID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.OperationStatus), nil
	}
}

func waitApplicationOperationSucceeded(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, operationID string, timeout time.Duration) (*awstypes.ApplicationOperationInfoDetails, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.OperationStatusInProgress),
		Target:  enum.Slice(awstypes.OperationStatusSuccessful),
		Refresh: statusApplicationOperation(ctx, conn, applicationName, operationID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApplicationOperationInfoDetails); ok {
		if failureDetails := output.OperationFailureDetails; failureDetails != nil && failureDetails.ErrorInfo != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(failureDetails.ErrorInfo.ErrorString)))
		}

		return output, err
	}

	return nil, err
}

func waitIAMPropagation[T any](ctx context.Context, f func() (*T, error)) (*T, error) {
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return f()
		},
		func(err error) (bool, error) {
			// Kinesis Stream: https://github.com/hashicorp/terraform-provider-aws/issues/7032
			if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "Kinesis Analytics service doesn't have sufficient privileges") {
				return true, err
			}

			// Kinesis Firehose: https://github.com/hashicorp/terraform-provider-aws/issues/7394
			if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "Kinesis Analytics doesn't have sufficient privileges") {
				return true, err
			}

			// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
			if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "does not provide Invoke permissions on the Lambda resource") {
				return true, err
			}

			// S3: https://github.com/hashicorp/terraform-provider-aws/issues/16104
			if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "Please check the role provided or validity of S3 location you provided") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*T), nil
}

func expandApplicationConfiguration(vApplicationConfiguration []interface{}) *awstypes.ApplicationConfiguration {
	if len(vApplicationConfiguration) == 0 || vApplicationConfiguration[0] == nil {
		return nil
	}

	applicationConfiguration := &awstypes.ApplicationConfiguration{}

	mApplicationConfiguration := vApplicationConfiguration[0].(map[string]interface{})

	if vApplicationCodeConfiguration, ok := mApplicationConfiguration["application_code_configuration"].([]interface{}); ok && len(vApplicationCodeConfiguration) > 0 && vApplicationCodeConfiguration[0] != nil {
		applicationCodeConfiguration := &awstypes.ApplicationCodeConfiguration{}

		mApplicationCodeConfiguration := vApplicationCodeConfiguration[0].(map[string]interface{})

		if vCodeContent, ok := mApplicationCodeConfiguration["code_content"].([]interface{}); ok && len(vCodeContent) > 0 && vCodeContent[0] != nil {
			codeContent := &awstypes.CodeContent{}

			mCodeContent := vCodeContent[0].(map[string]interface{})

			if vS3ContentLocation, ok := mCodeContent["s3_content_location"].([]interface{}); ok && len(vS3ContentLocation) > 0 && vS3ContentLocation[0] != nil {
				s3ContentLocation := &awstypes.S3ContentLocation{}

				mS3ContentLocation := vS3ContentLocation[0].(map[string]interface{})

				if vBucketArn, ok := mS3ContentLocation["bucket_arn"].(string); ok && vBucketArn != "" {
					s3ContentLocation.BucketARN = aws.String(vBucketArn)
				}
				if vFileKey, ok := mS3ContentLocation["file_key"].(string); ok && vFileKey != "" {
					s3ContentLocation.FileKey = aws.String(vFileKey)
				}
				if vObjectVersion, ok := mS3ContentLocation["object_version"].(string); ok && vObjectVersion != "" {
					s3ContentLocation.ObjectVersion = aws.String(vObjectVersion)
				}

				codeContent.S3ContentLocation = s3ContentLocation
			}

			if vTextContent, ok := mCodeContent["text_content"].(string); ok && vTextContent != "" {
				codeContent.TextContent = aws.String(vTextContent)
			}

			applicationCodeConfiguration.CodeContent = codeContent
		}

		if vCodeContentType, ok := mApplicationCodeConfiguration["code_content_type"].(string); ok && vCodeContentType != "" {
			applicationCodeConfiguration.CodeContentType = awstypes.CodeContentType(vCodeContentType)
		}

		applicationConfiguration.ApplicationCodeConfiguration = applicationCodeConfiguration
	}

	if vApplicationSnapshotConfiguration, ok := mApplicationConfiguration["application_snapshot_configuration"].([]interface{}); ok && len(vApplicationSnapshotConfiguration) > 0 && vApplicationSnapshotConfiguration[0] != nil {
		applicationSnapshotConfiguration := &awstypes.ApplicationSnapshotConfiguration{}

		mApplicationSnapshotConfiguration := vApplicationSnapshotConfiguration[0].(map[string]interface{})

		if vSnapshotsEnabled, ok := mApplicationSnapshotConfiguration["snapshots_enabled"].(bool); ok {
			applicationSnapshotConfiguration.SnapshotsEnabled = aws.Bool(vSnapshotsEnabled)
		}

		applicationConfiguration.ApplicationSnapshotConfiguration = applicationSnapshotConfiguration
	}

	if vEnvironmentProperties, ok := mApplicationConfiguration["environment_properties"].([]interface{}); ok && len(vEnvironmentProperties) > 0 && vEnvironmentProperties[0] != nil {
		environmentProperties := &awstypes.EnvironmentProperties{}

		mEnvironmentProperties := vEnvironmentProperties[0].(map[string]interface{})

		if vPropertyGroups, ok := mEnvironmentProperties["property_group"].(*schema.Set); ok && vPropertyGroups.Len() > 0 {
			environmentProperties.PropertyGroups = expandPropertyGroups(vPropertyGroups.List())
		}

		applicationConfiguration.EnvironmentProperties = environmentProperties
	}

	if vFlinkApplicationConfiguration, ok := mApplicationConfiguration["flink_application_configuration"].([]interface{}); ok && len(vFlinkApplicationConfiguration) > 0 && vFlinkApplicationConfiguration[0] != nil {
		flinkApplicationConfiguration := &awstypes.FlinkApplicationConfiguration{}

		mFlinkApplicationConfiguration := vFlinkApplicationConfiguration[0].(map[string]interface{})

		if vCheckpointConfiguration, ok := mFlinkApplicationConfiguration["checkpoint_configuration"].([]interface{}); ok && len(vCheckpointConfiguration) > 0 && vCheckpointConfiguration[0] != nil {
			checkpointConfiguration := &awstypes.CheckpointConfiguration{}

			mCheckpointConfiguration := vCheckpointConfiguration[0].(map[string]interface{})

			if vConfigurationType, ok := mCheckpointConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
				vConfigurationType := awstypes.ConfigurationType(vConfigurationType)
				checkpointConfiguration.ConfigurationType = vConfigurationType

				if vConfigurationType == awstypes.ConfigurationTypeCustom {
					if vCheckpointingEnabled, ok := mCheckpointConfiguration["checkpointing_enabled"].(bool); ok {
						checkpointConfiguration.CheckpointingEnabled = aws.Bool(vCheckpointingEnabled)
					}
					if vCheckpointInterval, ok := mCheckpointConfiguration["checkpoint_interval"].(int); ok {
						checkpointConfiguration.CheckpointInterval = aws.Int64(int64(vCheckpointInterval))
					}
					if vMinPauseBetweenCheckpoints, ok := mCheckpointConfiguration["min_pause_between_checkpoints"].(int); ok {
						checkpointConfiguration.MinPauseBetweenCheckpoints = aws.Int64(int64(vMinPauseBetweenCheckpoints))
					}
				}
			}

			flinkApplicationConfiguration.CheckpointConfiguration = checkpointConfiguration
		}

		if vMonitoringConfiguration, ok := mFlinkApplicationConfiguration["monitoring_configuration"].([]interface{}); ok && len(vMonitoringConfiguration) > 0 && vMonitoringConfiguration[0] != nil {
			monitoringConfiguration := &awstypes.MonitoringConfiguration{}

			mMonitoringConfiguration := vMonitoringConfiguration[0].(map[string]interface{})

			if vConfigurationType, ok := mMonitoringConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
				vConfigurationType := awstypes.ConfigurationType(vConfigurationType)
				monitoringConfiguration.ConfigurationType = vConfigurationType

				if vConfigurationType == awstypes.ConfigurationTypeCustom {
					if vLogLevel, ok := mMonitoringConfiguration["log_level"].(string); ok && vLogLevel != "" {
						monitoringConfiguration.LogLevel = awstypes.LogLevel(vLogLevel)
					}
					if vMetricsLevel, ok := mMonitoringConfiguration["metrics_level"].(string); ok && vMetricsLevel != "" {
						monitoringConfiguration.MetricsLevel = awstypes.MetricsLevel(vMetricsLevel)
					}
				}
			}

			flinkApplicationConfiguration.MonitoringConfiguration = monitoringConfiguration
		}

		if vParallelismConfiguration, ok := mFlinkApplicationConfiguration["parallelism_configuration"].([]interface{}); ok && len(vParallelismConfiguration) > 0 && vParallelismConfiguration[0] != nil {
			parallelismConfiguration := &awstypes.ParallelismConfiguration{}

			mParallelismConfiguration := vParallelismConfiguration[0].(map[string]interface{})

			if vConfigurationType, ok := mParallelismConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
				vConfigurationType := awstypes.ConfigurationType(vConfigurationType)
				parallelismConfiguration.ConfigurationType = vConfigurationType

				if vConfigurationType == awstypes.ConfigurationTypeCustom {
					if vAutoScalingEnabled, ok := mParallelismConfiguration["auto_scaling_enabled"].(bool); ok {
						parallelismConfiguration.AutoScalingEnabled = aws.Bool(vAutoScalingEnabled)
					}
					if vParallelism, ok := mParallelismConfiguration["parallelism"].(int); ok {
						parallelismConfiguration.Parallelism = aws.Int32(int32(vParallelism))
					}
					if vParallelismPerKPU, ok := mParallelismConfiguration["parallelism_per_kpu"].(int); ok {
						parallelismConfiguration.ParallelismPerKPU = aws.Int32(int32(vParallelismPerKPU))
					}
				}
			}

			flinkApplicationConfiguration.ParallelismConfiguration = parallelismConfiguration
		}

		applicationConfiguration.FlinkApplicationConfiguration = flinkApplicationConfiguration
	}

	if vSqlApplicationConfiguration, ok := mApplicationConfiguration["sql_application_configuration"].([]interface{}); ok && len(vSqlApplicationConfiguration) > 0 && vSqlApplicationConfiguration[0] != nil {
		sqlApplicationConfiguration := &awstypes.SqlApplicationConfiguration{}

		mSqlApplicationConfiguration := vSqlApplicationConfiguration[0].(map[string]interface{})

		if vInput, ok := mSqlApplicationConfiguration["input"].([]interface{}); ok && len(vInput) > 0 && vInput[0] != nil {
			sqlApplicationConfiguration.Inputs = []awstypes.Input{*expandInput(vInput)}
		}

		if vOutputs, ok := mSqlApplicationConfiguration["output"].(*schema.Set); ok {
			sqlApplicationConfiguration.Outputs = expandOutputs(vOutputs.List())
		}

		if vReferenceDataSource, ok := mSqlApplicationConfiguration["reference_data_source"].([]interface{}); ok && len(vReferenceDataSource) > 0 && vReferenceDataSource[0] != nil {
			sqlApplicationConfiguration.ReferenceDataSources = []awstypes.ReferenceDataSource{*expandReferenceDataSource(vReferenceDataSource)}
		}

		applicationConfiguration.SqlApplicationConfiguration = sqlApplicationConfiguration
	}

	if vVpcConfiguration, ok := mApplicationConfiguration[names.AttrVPCConfiguration].([]interface{}); ok && len(vVpcConfiguration) > 0 && vVpcConfiguration[0] != nil {
		applicationConfiguration.VpcConfigurations = []awstypes.VpcConfiguration{*expandVPCConfiguration(vVpcConfiguration)}
	}

	return applicationConfiguration
}

func expandApplicationCodeConfigurationUpdate(vApplicationCodeConfiguration []interface{}) *awstypes.ApplicationCodeConfigurationUpdate {
	if len(vApplicationCodeConfiguration) == 0 || vApplicationCodeConfiguration[0] == nil {
		return nil
	}

	applicationCodeConfigurationUpdate := &awstypes.ApplicationCodeConfigurationUpdate{}

	mApplicationCodeConfiguration := vApplicationCodeConfiguration[0].(map[string]interface{})

	if vCodeContent, ok := mApplicationCodeConfiguration["code_content"].([]interface{}); ok && len(vCodeContent) > 0 && vCodeContent[0] != nil {
		codeContentUpdate := &awstypes.CodeContentUpdate{}

		mCodeContent := vCodeContent[0].(map[string]interface{})

		if vS3ContentLocation, ok := mCodeContent["s3_content_location"].([]interface{}); ok && len(vS3ContentLocation) > 0 && vS3ContentLocation[0] != nil {
			s3ContentLocationUpdate := &awstypes.S3ContentLocationUpdate{}

			mS3ContentLocation := vS3ContentLocation[0].(map[string]interface{})

			if vBucketArn, ok := mS3ContentLocation["bucket_arn"].(string); ok && vBucketArn != "" {
				s3ContentLocationUpdate.BucketARNUpdate = aws.String(vBucketArn)
			}
			if vFileKey, ok := mS3ContentLocation["file_key"].(string); ok && vFileKey != "" {
				s3ContentLocationUpdate.FileKeyUpdate = aws.String(vFileKey)
			}
			if vObjectVersion, ok := mS3ContentLocation["object_version"].(string); ok && vObjectVersion != "" {
				s3ContentLocationUpdate.ObjectVersionUpdate = aws.String(vObjectVersion)
			}

			codeContentUpdate.S3ContentLocationUpdate = s3ContentLocationUpdate
		}

		if vTextContent, ok := mCodeContent["text_content"].(string); ok && vTextContent != "" {
			codeContentUpdate.TextContentUpdate = aws.String(vTextContent)
		}

		applicationCodeConfigurationUpdate.CodeContentUpdate = codeContentUpdate
	}

	if vCodeContentType, ok := mApplicationCodeConfiguration["code_content_type"].(string); ok && vCodeContentType != "" {
		applicationCodeConfigurationUpdate.CodeContentTypeUpdate = awstypes.CodeContentType(vCodeContentType)
	}

	return applicationCodeConfigurationUpdate
}

func expandApplicationFlinkApplicationConfigurationUpdate(vFlinkApplicationConfiguration []interface{}) *awstypes.FlinkApplicationConfigurationUpdate {
	if len(vFlinkApplicationConfiguration) == 0 || vFlinkApplicationConfiguration[0] == nil {
		return nil
	}

	flinkApplicationConfigurationUpdate := &awstypes.FlinkApplicationConfigurationUpdate{}

	mFlinkApplicationConfiguration := vFlinkApplicationConfiguration[0].(map[string]interface{})

	if vCheckpointConfiguration, ok := mFlinkApplicationConfiguration["checkpoint_configuration"].([]interface{}); ok && len(vCheckpointConfiguration) > 0 && vCheckpointConfiguration[0] != nil {
		checkpointConfigurationUpdate := &awstypes.CheckpointConfigurationUpdate{}

		mCheckpointConfiguration := vCheckpointConfiguration[0].(map[string]interface{})

		if vConfigurationType, ok := mCheckpointConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
			vConfigurationType := awstypes.ConfigurationType(vConfigurationType)
			checkpointConfigurationUpdate.ConfigurationTypeUpdate = vConfigurationType

			if vConfigurationType == awstypes.ConfigurationTypeCustom {
				if vCheckpointingEnabled, ok := mCheckpointConfiguration["checkpointing_enabled"].(bool); ok {
					checkpointConfigurationUpdate.CheckpointingEnabledUpdate = aws.Bool(vCheckpointingEnabled)
				}
				if vCheckpointInterval, ok := mCheckpointConfiguration["checkpoint_interval"].(int); ok {
					checkpointConfigurationUpdate.CheckpointIntervalUpdate = aws.Int64(int64(vCheckpointInterval))
				}
				if vMinPauseBetweenCheckpoints, ok := mCheckpointConfiguration["min_pause_between_checkpoints"].(int); ok {
					checkpointConfigurationUpdate.MinPauseBetweenCheckpointsUpdate = aws.Int64(int64(vMinPauseBetweenCheckpoints))
				}
			}
		}

		flinkApplicationConfigurationUpdate.CheckpointConfigurationUpdate = checkpointConfigurationUpdate
	}

	if vMonitoringConfiguration, ok := mFlinkApplicationConfiguration["monitoring_configuration"].([]interface{}); ok && len(vMonitoringConfiguration) > 0 && vMonitoringConfiguration[0] != nil {
		monitoringConfigurationUpdate := &awstypes.MonitoringConfigurationUpdate{}

		mMonitoringConfiguration := vMonitoringConfiguration[0].(map[string]interface{})

		if vConfigurationType, ok := mMonitoringConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
			vConfigurationType := awstypes.ConfigurationType(vConfigurationType)
			monitoringConfigurationUpdate.ConfigurationTypeUpdate = vConfigurationType

			if vConfigurationType == awstypes.ConfigurationTypeCustom {
				if vLogLevel, ok := mMonitoringConfiguration["log_level"].(string); ok && vLogLevel != "" {
					monitoringConfigurationUpdate.LogLevelUpdate = awstypes.LogLevel(vLogLevel)
				}
				if vMetricsLevel, ok := mMonitoringConfiguration["metrics_level"].(string); ok && vMetricsLevel != "" {
					monitoringConfigurationUpdate.MetricsLevelUpdate = awstypes.MetricsLevel(vMetricsLevel)
				}
			}
		}

		flinkApplicationConfigurationUpdate.MonitoringConfigurationUpdate = monitoringConfigurationUpdate
	}

	if vParallelismConfiguration, ok := mFlinkApplicationConfiguration["parallelism_configuration"].([]interface{}); ok && len(vParallelismConfiguration) > 0 && vParallelismConfiguration[0] != nil {
		parallelismConfigurationUpdate := &awstypes.ParallelismConfigurationUpdate{}

		mParallelismConfiguration := vParallelismConfiguration[0].(map[string]interface{})

		if vConfigurationType, ok := mParallelismConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
			vConfigurationType := awstypes.ConfigurationType(vConfigurationType)
			parallelismConfigurationUpdate.ConfigurationTypeUpdate = vConfigurationType

			if vConfigurationType == awstypes.ConfigurationTypeCustom {
				if vAutoScalingEnabled, ok := mParallelismConfiguration["auto_scaling_enabled"].(bool); ok {
					parallelismConfigurationUpdate.AutoScalingEnabledUpdate = aws.Bool(vAutoScalingEnabled)
				}
				if vParallelism, ok := mParallelismConfiguration["parallelism"].(int); ok {
					parallelismConfigurationUpdate.ParallelismUpdate = aws.Int32(int32(vParallelism))
				}
				if vParallelismPerKPU, ok := mParallelismConfiguration["parallelism_per_kpu"].(int); ok {
					parallelismConfigurationUpdate.ParallelismPerKPUUpdate = aws.Int32(int32(vParallelismPerKPU))
				}
			}
		}

		flinkApplicationConfigurationUpdate.ParallelismConfigurationUpdate = parallelismConfigurationUpdate
	}

	return flinkApplicationConfigurationUpdate
}

func expandApplicationSnapshotConfigurationUpdate(vApplicationSnapshotConfiguration []interface{}) *awstypes.ApplicationSnapshotConfigurationUpdate {
	if len(vApplicationSnapshotConfiguration) == 0 || vApplicationSnapshotConfiguration[0] == nil {
		return nil
	}

	applicationSnapshotConfigurationUpdate := &awstypes.ApplicationSnapshotConfigurationUpdate{}

	mApplicationSnapshotConfiguration := vApplicationSnapshotConfiguration[0].(map[string]interface{})

	if vSnapshotsEnabled, ok := mApplicationSnapshotConfiguration["snapshots_enabled"].(bool); ok {
		applicationSnapshotConfigurationUpdate.SnapshotsEnabledUpdate = aws.Bool(vSnapshotsEnabled)
	}

	return applicationSnapshotConfigurationUpdate
}

func expandCloudWatchLoggingOptions(vCloudWatchLoggingOptions []interface{}) []awstypes.CloudWatchLoggingOption {
	if len(vCloudWatchLoggingOptions) == 0 || vCloudWatchLoggingOptions[0] == nil {
		return nil
	}

	cloudWatchLoggingOption := awstypes.CloudWatchLoggingOption{}

	mCloudWatchLoggingOption := vCloudWatchLoggingOptions[0].(map[string]interface{})

	if vLogStreamArn, ok := mCloudWatchLoggingOption["log_stream_arn"].(string); ok && vLogStreamArn != "" {
		cloudWatchLoggingOption.LogStreamARN = aws.String(vLogStreamArn)
	}

	return []awstypes.CloudWatchLoggingOption{cloudWatchLoggingOption}
}

func expandEnvironmentPropertyUpdates(vEnvironmentProperties []interface{}) *awstypes.EnvironmentPropertyUpdates {
	if len(vEnvironmentProperties) == 0 || vEnvironmentProperties[0] == nil {
		// Return empty updates to remove all existing property groups.
		return &awstypes.EnvironmentPropertyUpdates{PropertyGroups: []awstypes.PropertyGroup{}}
	}

	environmentPropertyUpdates := &awstypes.EnvironmentPropertyUpdates{}

	mEnvironmentProperties := vEnvironmentProperties[0].(map[string]interface{})

	if vPropertyGroups, ok := mEnvironmentProperties["property_group"].(*schema.Set); ok && vPropertyGroups.Len() > 0 {
		environmentPropertyUpdates.PropertyGroups = expandPropertyGroups(vPropertyGroups.List())
	}

	return environmentPropertyUpdates
}

func expandInput(vInput []interface{}) *awstypes.Input {
	if len(vInput) == 0 || vInput[0] == nil {
		return nil
	}

	input := &awstypes.Input{}

	mInput := vInput[0].(map[string]interface{})

	if vInputParallelism, ok := mInput["input_parallelism"].([]interface{}); ok && len(vInputParallelism) > 0 && vInputParallelism[0] != nil {
		inputParallelism := &awstypes.InputParallelism{}

		mInputParallelism := vInputParallelism[0].(map[string]interface{})

		if vCount, ok := mInputParallelism["count"].(int); ok {
			inputParallelism.Count = aws.Int32(int32(vCount))
		}

		input.InputParallelism = inputParallelism
	}

	if vInputProcessingConfiguration, ok := mInput["input_processing_configuration"].([]interface{}); ok {
		input.InputProcessingConfiguration = expandInputProcessingConfiguration(vInputProcessingConfiguration)
	}

	if vInputSchema, ok := mInput["input_schema"].([]interface{}); ok {
		input.InputSchema = expandSourceSchema(vInputSchema)
	}

	if vKinesisFirehoseInput, ok := mInput["kinesis_firehose_input"].([]interface{}); ok && len(vKinesisFirehoseInput) > 0 && vKinesisFirehoseInput[0] != nil {
		kinesisFirehoseInput := &awstypes.KinesisFirehoseInput{}

		mKinesisFirehoseInput := vKinesisFirehoseInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseInput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisFirehoseInput.ResourceARN = aws.String(vResourceArn)
		}

		input.KinesisFirehoseInput = kinesisFirehoseInput
	}

	if vKinesisStreamsInput, ok := mInput["kinesis_streams_input"].([]interface{}); ok && len(vKinesisStreamsInput) > 0 && vKinesisStreamsInput[0] != nil {
		kinesisStreamsInput := &awstypes.KinesisStreamsInput{}

		mKinesisStreamsInput := vKinesisStreamsInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsInput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisStreamsInput.ResourceARN = aws.String(vResourceArn)
		}

		input.KinesisStreamsInput = kinesisStreamsInput
	}

	if vNamePrefix, ok := mInput[names.AttrNamePrefix].(string); ok && vNamePrefix != "" {
		input.NamePrefix = aws.String(vNamePrefix)
	}

	return input
}

func expandInputProcessingConfiguration(vInputProcessingConfiguration []interface{}) *awstypes.InputProcessingConfiguration {
	if len(vInputProcessingConfiguration) == 0 || vInputProcessingConfiguration[0] == nil {
		return nil
	}

	inputProcessingConfiguration := &awstypes.InputProcessingConfiguration{}

	mInputProcessingConfiguration := vInputProcessingConfiguration[0].(map[string]interface{})

	if vInputLambdaProcessor, ok := mInputProcessingConfiguration["input_lambda_processor"].([]interface{}); ok && len(vInputLambdaProcessor) > 0 && vInputLambdaProcessor[0] != nil {
		inputLambdaProcessor := &awstypes.InputLambdaProcessor{}

		mInputLambdaProcessor := vInputLambdaProcessor[0].(map[string]interface{})

		if vResourceArn, ok := mInputLambdaProcessor[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			inputLambdaProcessor.ResourceARN = aws.String(vResourceArn)
		}

		inputProcessingConfiguration.InputLambdaProcessor = inputLambdaProcessor
	}

	return inputProcessingConfiguration
}

func expandInputUpdate(vInput []interface{}) awstypes.InputUpdate {
	if len(vInput) == 0 || vInput[0] == nil {
		return awstypes.InputUpdate{}
	}

	inputUpdate := awstypes.InputUpdate{}

	mInput := vInput[0].(map[string]interface{})

	if vInputId, ok := mInput["input_id"].(string); ok && vInputId != "" {
		inputUpdate.InputId = aws.String(vInputId)
	}

	if vInputParallelism, ok := mInput["input_parallelism"].([]interface{}); ok && len(vInputParallelism) > 0 && vInputParallelism[0] != nil {
		inputParallelismUpdate := &awstypes.InputParallelismUpdate{}

		mInputParallelism := vInputParallelism[0].(map[string]interface{})

		if vCount, ok := mInputParallelism["count"].(int); ok {
			inputParallelismUpdate.CountUpdate = aws.Int32(int32(vCount))
		}

		inputUpdate.InputParallelismUpdate = inputParallelismUpdate
	}

	if vInputProcessingConfiguration, ok := mInput["input_processing_configuration"].([]interface{}); ok && len(vInputProcessingConfiguration) > 0 && vInputProcessingConfiguration[0] != nil {
		inputProcessingConfigurationUpdate := &awstypes.InputProcessingConfigurationUpdate{}

		mInputProcessingConfiguration := vInputProcessingConfiguration[0].(map[string]interface{})

		if vInputLambdaProcessor, ok := mInputProcessingConfiguration["input_lambda_processor"].([]interface{}); ok && len(vInputLambdaProcessor) > 0 && vInputLambdaProcessor[0] != nil {
			inputLambdaProcessorUpdate := &awstypes.InputLambdaProcessorUpdate{}

			mInputLambdaProcessor := vInputLambdaProcessor[0].(map[string]interface{})

			if vResourceArn, ok := mInputLambdaProcessor[names.AttrResourceARN].(string); ok && vResourceArn != "" {
				inputLambdaProcessorUpdate.ResourceARNUpdate = aws.String(vResourceArn)
			}

			inputProcessingConfigurationUpdate.InputLambdaProcessorUpdate = inputLambdaProcessorUpdate
		}

		inputUpdate.InputProcessingConfigurationUpdate = inputProcessingConfigurationUpdate
	}

	if vInputSchema, ok := mInput["input_schema"].([]interface{}); ok && len(vInputSchema) > 0 && vInputSchema[0] != nil {
		inputSchemaUpdate := &awstypes.InputSchemaUpdate{}

		mInputSchema := vInputSchema[0].(map[string]interface{})

		if vRecordColumns, ok := mInputSchema["record_column"].([]interface{}); ok {
			inputSchemaUpdate.RecordColumnUpdates = expandRecordColumns(vRecordColumns)
		}

		if vRecordEncoding, ok := mInputSchema["record_encoding"].(string); ok && vRecordEncoding != "" {
			inputSchemaUpdate.RecordEncodingUpdate = aws.String(vRecordEncoding)
		}

		if vRecordFormat, ok := mInputSchema["record_format"].([]interface{}); ok {
			inputSchemaUpdate.RecordFormatUpdate = expandRecordFormat(vRecordFormat)
		}

		inputUpdate.InputSchemaUpdate = inputSchemaUpdate
	}

	if vKinesisFirehoseInput, ok := mInput["kinesis_firehose_input"].([]interface{}); ok && len(vKinesisFirehoseInput) > 0 && vKinesisFirehoseInput[0] != nil {
		kinesisFirehoseInputUpdate := &awstypes.KinesisFirehoseInputUpdate{}

		mKinesisFirehoseInput := vKinesisFirehoseInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseInput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisFirehoseInputUpdate.ResourceARNUpdate = aws.String(vResourceArn)
		}

		inputUpdate.KinesisFirehoseInputUpdate = kinesisFirehoseInputUpdate
	}

	if vKinesisStreamsInput, ok := mInput["kinesis_streams_input"].([]interface{}); ok && len(vKinesisStreamsInput) > 0 && vKinesisStreamsInput[0] != nil {
		kinesisStreamsInputUpdate := &awstypes.KinesisStreamsInputUpdate{}

		mKinesisStreamsInput := vKinesisStreamsInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsInput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisStreamsInputUpdate.ResourceARNUpdate = aws.String(vResourceArn)
		}

		inputUpdate.KinesisStreamsInputUpdate = kinesisStreamsInputUpdate
	}

	if vNamePrefix, ok := mInput[names.AttrNamePrefix].(string); ok && vNamePrefix != "" {
		inputUpdate.NamePrefixUpdate = aws.String(vNamePrefix)
	}

	return inputUpdate
}

func expandOutput(vOutput interface{}) *awstypes.Output {
	if vOutput == nil {
		return nil
	}

	output := &awstypes.Output{}

	mOutput := vOutput.(map[string]interface{})

	if vDestinationSchema, ok := mOutput["destination_schema"].([]interface{}); ok && len(vDestinationSchema) > 0 && vDestinationSchema[0] != nil {
		destinationSchema := &awstypes.DestinationSchema{}

		mDestinationSchema := vDestinationSchema[0].(map[string]interface{})

		if vRecordFormatType, ok := mDestinationSchema["record_format_type"].(string); ok && vRecordFormatType != "" {
			destinationSchema.RecordFormatType = awstypes.RecordFormatType(vRecordFormatType)
		}

		output.DestinationSchema = destinationSchema
	}

	if vKinesisFirehoseOutput, ok := mOutput["kinesis_firehose_output"].([]interface{}); ok && len(vKinesisFirehoseOutput) > 0 && vKinesisFirehoseOutput[0] != nil {
		kinesisFirehoseOutput := &awstypes.KinesisFirehoseOutput{}

		mKinesisFirehoseOutput := vKinesisFirehoseOutput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseOutput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisFirehoseOutput.ResourceARN = aws.String(vResourceArn)
		}

		output.KinesisFirehoseOutput = kinesisFirehoseOutput
	}

	if vKinesisStreamsOutput, ok := mOutput["kinesis_streams_output"].([]interface{}); ok && len(vKinesisStreamsOutput) > 0 && vKinesisStreamsOutput[0] != nil {
		kinesisStreamsOutput := &awstypes.KinesisStreamsOutput{}

		mKinesisStreamsOutput := vKinesisStreamsOutput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsOutput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisStreamsOutput.ResourceARN = aws.String(vResourceArn)
		}

		output.KinesisStreamsOutput = kinesisStreamsOutput
	}

	if vLambdaOutput, ok := mOutput["lambda_output"].([]interface{}); ok && len(vLambdaOutput) > 0 && vLambdaOutput[0] != nil {
		lambdaOutput := &awstypes.LambdaOutput{}

		mLambdaOutput := vLambdaOutput[0].(map[string]interface{})

		if vResourceArn, ok := mLambdaOutput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			lambdaOutput.ResourceARN = aws.String(vResourceArn)
		}

		output.LambdaOutput = lambdaOutput
	}

	if vName, ok := mOutput[names.AttrName].(string); ok && vName != "" {
		output.Name = aws.String(vName)
	}

	return output
}

func expandOutputs(vOutputs []interface{}) []awstypes.Output {
	if len(vOutputs) == 0 {
		return nil
	}

	outputs := []awstypes.Output{}

	for _, vOutput := range vOutputs {
		output := expandOutput(vOutput)

		if output != nil {
			outputs = append(outputs, *expandOutput(vOutput))
		}
	}

	return outputs
}

func expandPropertyGroups(vPropertyGroups []interface{}) []awstypes.PropertyGroup {
	propertyGroups := []awstypes.PropertyGroup{}

	for _, vPropertyGroup := range vPropertyGroups {
		propertyGroup := awstypes.PropertyGroup{}

		mPropertyGroup := vPropertyGroup.(map[string]interface{})

		if vPropertyGroupID, ok := mPropertyGroup["property_group_id"].(string); ok && vPropertyGroupID != "" {
			propertyGroup.PropertyGroupId = aws.String(vPropertyGroupID)
		} else {
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			continue
		}

		if vPropertyMap, ok := mPropertyGroup["property_map"].(map[string]interface{}); ok && len(vPropertyMap) > 0 {
			propertyGroup.PropertyMap = flex.ExpandStringValueMap(vPropertyMap)
		}

		propertyGroups = append(propertyGroups, propertyGroup)
	}

	return propertyGroups
}

func expandRecordColumns(vRecordColumns []interface{}) []awstypes.RecordColumn {
	recordColumns := []awstypes.RecordColumn{}

	for _, vRecordColumn := range vRecordColumns {
		recordColumn := awstypes.RecordColumn{}

		mRecordColumn := vRecordColumn.(map[string]interface{})

		if vMapping, ok := mRecordColumn["mapping"].(string); ok && vMapping != "" {
			recordColumn.Mapping = aws.String(vMapping)
		}
		if vName, ok := mRecordColumn[names.AttrName].(string); ok && vName != "" {
			recordColumn.Name = aws.String(vName)
		}
		if vSqlType, ok := mRecordColumn["sql_type"].(string); ok && vSqlType != "" {
			recordColumn.SqlType = aws.String(vSqlType)
		}

		recordColumns = append(recordColumns, recordColumn)
	}

	return recordColumns
}

func expandRecordFormat(vRecordFormat []interface{}) *awstypes.RecordFormat {
	if len(vRecordFormat) == 0 || vRecordFormat[0] == nil {
		return nil
	}

	recordFormat := &awstypes.RecordFormat{}

	mRecordFormat := vRecordFormat[0].(map[string]interface{})

	if vMappingParameters, ok := mRecordFormat["mapping_parameters"].([]interface{}); ok && len(vMappingParameters) > 0 && vMappingParameters[0] != nil {
		mappingParameters := &awstypes.MappingParameters{}

		mMappingParameters := vMappingParameters[0].(map[string]interface{})

		if vCsvMappingParameters, ok := mMappingParameters["csv_mapping_parameters"].([]interface{}); ok && len(vCsvMappingParameters) > 0 && vCsvMappingParameters[0] != nil {
			csvMappingParameters := &awstypes.CSVMappingParameters{}

			mCsvMappingParameters := vCsvMappingParameters[0].(map[string]interface{})

			if vRecordColumnDelimiter, ok := mCsvMappingParameters["record_column_delimiter"].(string); ok && vRecordColumnDelimiter != "" {
				csvMappingParameters.RecordColumnDelimiter = aws.String(vRecordColumnDelimiter)
			}
			if vRecordRowDelimiter, ok := mCsvMappingParameters["record_row_delimiter"].(string); ok && vRecordRowDelimiter != "" {
				csvMappingParameters.RecordRowDelimiter = aws.String(vRecordRowDelimiter)
			}

			mappingParameters.CSVMappingParameters = csvMappingParameters
		}

		if vJsonMappingParameters, ok := mMappingParameters["json_mapping_parameters"].([]interface{}); ok && len(vJsonMappingParameters) > 0 && vJsonMappingParameters[0] != nil {
			jsonMappingParameters := &awstypes.JSONMappingParameters{}

			mJsonMappingParameters := vJsonMappingParameters[0].(map[string]interface{})

			if vRecordRowPath, ok := mJsonMappingParameters["record_row_path"].(string); ok && vRecordRowPath != "" {
				jsonMappingParameters.RecordRowPath = aws.String(vRecordRowPath)
			}

			mappingParameters.JSONMappingParameters = jsonMappingParameters
		}

		recordFormat.MappingParameters = mappingParameters
	}

	if vRecordFormatType, ok := mRecordFormat["record_format_type"].(string); ok && vRecordFormatType != "" {
		recordFormat.RecordFormatType = awstypes.RecordFormatType(vRecordFormatType)
	}

	return recordFormat
}

func expandReferenceDataSource(vReferenceDataSource []interface{}) *awstypes.ReferenceDataSource {
	if len(vReferenceDataSource) == 0 || vReferenceDataSource[0] == nil {
		return nil
	}

	referenceDataSource := &awstypes.ReferenceDataSource{}

	mReferenceDataSource := vReferenceDataSource[0].(map[string]interface{})

	if vReferenceSchema, ok := mReferenceDataSource["reference_schema"].([]interface{}); ok {
		referenceDataSource.ReferenceSchema = expandSourceSchema(vReferenceSchema)
	}

	if vS3ReferenceDataSource, ok := mReferenceDataSource["s3_reference_data_source"].([]interface{}); ok && len(vS3ReferenceDataSource) > 0 && vS3ReferenceDataSource[0] != nil {
		s3ReferenceDataSource := &awstypes.S3ReferenceDataSource{}

		mS3ReferenceDataSource := vS3ReferenceDataSource[0].(map[string]interface{})

		if vBucketArn, ok := mS3ReferenceDataSource["bucket_arn"].(string); ok && vBucketArn != "" {
			s3ReferenceDataSource.BucketARN = aws.String(vBucketArn)
		}
		if vFileKey, ok := mS3ReferenceDataSource["file_key"].(string); ok && vFileKey != "" {
			s3ReferenceDataSource.FileKey = aws.String(vFileKey)
		}

		referenceDataSource.S3ReferenceDataSource = s3ReferenceDataSource
	}

	if vTableName, ok := mReferenceDataSource[names.AttrTableName].(string); ok && vTableName != "" {
		referenceDataSource.TableName = aws.String(vTableName)
	}

	return referenceDataSource
}

func expandReferenceDataSourceUpdate(vReferenceDataSource []interface{}) awstypes.ReferenceDataSourceUpdate {
	if len(vReferenceDataSource) == 0 || vReferenceDataSource[0] == nil {
		return awstypes.ReferenceDataSourceUpdate{}
	}

	referenceDataSourceUpdate := awstypes.ReferenceDataSourceUpdate{}

	mReferenceDataSource := vReferenceDataSource[0].(map[string]interface{})

	if vReferenceId, ok := mReferenceDataSource["reference_id"].(string); ok && vReferenceId != "" {
		referenceDataSourceUpdate.ReferenceId = aws.String(vReferenceId)
	}

	if vReferenceSchema, ok := mReferenceDataSource["reference_schema"].([]interface{}); ok {
		referenceDataSourceUpdate.ReferenceSchemaUpdate = expandSourceSchema(vReferenceSchema)
	}

	if vS3ReferenceDataSource, ok := mReferenceDataSource["s3_reference_data_source"].([]interface{}); ok && len(vS3ReferenceDataSource) > 0 && vS3ReferenceDataSource[0] != nil {
		s3ReferenceDataSourceUpdate := &awstypes.S3ReferenceDataSourceUpdate{}

		mS3ReferenceDataSource := vS3ReferenceDataSource[0].(map[string]interface{})

		if vBucketArn, ok := mS3ReferenceDataSource["bucket_arn"].(string); ok && vBucketArn != "" {
			s3ReferenceDataSourceUpdate.BucketARNUpdate = aws.String(vBucketArn)
		}
		if vFileKey, ok := mS3ReferenceDataSource["file_key"].(string); ok && vFileKey != "" {
			s3ReferenceDataSourceUpdate.FileKeyUpdate = aws.String(vFileKey)
		}

		referenceDataSourceUpdate.S3ReferenceDataSourceUpdate = s3ReferenceDataSourceUpdate
	}

	if vTableName, ok := mReferenceDataSource[names.AttrTableName].(string); ok && vTableName != "" {
		referenceDataSourceUpdate.TableNameUpdate = aws.String(vTableName)
	}

	return referenceDataSourceUpdate
}

func expandSourceSchema(vSourceSchema []interface{}) *awstypes.SourceSchema {
	if len(vSourceSchema) == 0 || vSourceSchema[0] == nil {
		return nil
	}

	sourceSchema := &awstypes.SourceSchema{}

	mSourceSchema := vSourceSchema[0].(map[string]interface{})

	if vRecordColumns, ok := mSourceSchema["record_column"].([]interface{}); ok {
		sourceSchema.RecordColumns = expandRecordColumns(vRecordColumns)
	}

	if vRecordEncoding, ok := mSourceSchema["record_encoding"].(string); ok && vRecordEncoding != "" {
		sourceSchema.RecordEncoding = aws.String(vRecordEncoding)
	}

	if vRecordFormat, ok := mSourceSchema["record_format"].([]interface{}); ok && len(vRecordFormat) > 0 && vRecordFormat[0] != nil {
		sourceSchema.RecordFormat = expandRecordFormat(vRecordFormat)
	}

	return sourceSchema
}

func expandVPCConfiguration(vVpcConfiguration []interface{}) *awstypes.VpcConfiguration {
	if len(vVpcConfiguration) == 0 || vVpcConfiguration[0] == nil {
		return nil
	}

	vpcConfiguration := &awstypes.VpcConfiguration{}

	mVpcConfiguration := vVpcConfiguration[0].(map[string]interface{})

	if vSecurityGroupIds, ok := mVpcConfiguration[names.AttrSecurityGroupIDs].(*schema.Set); ok && vSecurityGroupIds.Len() > 0 {
		vpcConfiguration.SecurityGroupIds = flex.ExpandStringValueSet(vSecurityGroupIds)
	}

	if vSubnetIds, ok := mVpcConfiguration[names.AttrSubnetIDs].(*schema.Set); ok && vSubnetIds.Len() > 0 {
		vpcConfiguration.SubnetIds = flex.ExpandStringValueSet(vSubnetIds)
	}

	return vpcConfiguration
}

func expandVPCConfigurationUpdate(vVpcConfiguration []interface{}) awstypes.VpcConfigurationUpdate {
	if len(vVpcConfiguration) == 0 || vVpcConfiguration[0] == nil {
		return awstypes.VpcConfigurationUpdate{}
	}

	vpcConfigurationUpdate := awstypes.VpcConfigurationUpdate{}

	mVpcConfiguration := vVpcConfiguration[0].(map[string]interface{})

	if vSecurityGroupIds, ok := mVpcConfiguration[names.AttrSecurityGroupIDs].(*schema.Set); ok && vSecurityGroupIds.Len() > 0 {
		vpcConfigurationUpdate.SecurityGroupIdUpdates = flex.ExpandStringValueSet(vSecurityGroupIds)
	}

	if vSubnetIds, ok := mVpcConfiguration[names.AttrSubnetIDs].(*schema.Set); ok && vSubnetIds.Len() > 0 {
		vpcConfigurationUpdate.SubnetIdUpdates = flex.ExpandStringValueSet(vSubnetIds)
	}

	if vVpcConfigurationId, ok := mVpcConfiguration["vpc_configuration_id"].(string); ok && vVpcConfigurationId != "" {
		vpcConfigurationUpdate.VpcConfigurationId = aws.String(vVpcConfigurationId)
	}

	return vpcConfigurationUpdate
}

func expandRunConfigurationUpdate(vRunConfigurationUpdate []interface{}) *awstypes.RunConfigurationUpdate {
	if len(vRunConfigurationUpdate) == 0 || vRunConfigurationUpdate[0] == nil {
		return nil
	}

	runConfigurationUpdate := &awstypes.RunConfigurationUpdate{}

	mRunConfiguration := vRunConfigurationUpdate[0].(map[string]interface{})

	if vApplicationRestoreConfiguration, ok := mRunConfiguration["application_restore_configuration"].([]interface{}); ok && len(vApplicationRestoreConfiguration) > 0 && vApplicationRestoreConfiguration[0] != nil {
		applicationRestoreConfiguration := &awstypes.ApplicationRestoreConfiguration{}

		mApplicationRestoreConfiguration := vApplicationRestoreConfiguration[0].(map[string]interface{})

		if vApplicationRestoreType, ok := mApplicationRestoreConfiguration["application_restore_type"].(string); ok && vApplicationRestoreType != "" {
			applicationRestoreConfiguration.ApplicationRestoreType = awstypes.ApplicationRestoreType(vApplicationRestoreType)
		}

		if vSnapshotName, ok := mApplicationRestoreConfiguration["snapshot_name"].(string); ok && vSnapshotName != "" {
			applicationRestoreConfiguration.SnapshotName = aws.String(vSnapshotName)
		}

		runConfigurationUpdate.ApplicationRestoreConfiguration = applicationRestoreConfiguration
	}

	if vFlinkRunConfiguration, ok := mRunConfiguration["flink_run_configuration"].([]interface{}); ok && len(vFlinkRunConfiguration) > 0 && vFlinkRunConfiguration[0] != nil {
		flinkRunConfiguration := &awstypes.FlinkRunConfiguration{}

		mFlinkRunConfiguration := vFlinkRunConfiguration[0].(map[string]interface{})

		if vAllowNonRestoredState, ok := mFlinkRunConfiguration["allow_non_restored_state"].(bool); ok {
			flinkRunConfiguration.AllowNonRestoredState = aws.Bool(vAllowNonRestoredState)
		}

		runConfigurationUpdate.FlinkRunConfiguration = flinkRunConfiguration
	}

	return runConfigurationUpdate
}

func flattenApplicationConfigurationDescription(applicationConfigurationDescription *awstypes.ApplicationConfigurationDescription) []interface{} {
	if applicationConfigurationDescription == nil {
		return []interface{}{}
	}

	mApplicationConfiguration := map[string]interface{}{}

	if applicationCodeConfigurationDescription := applicationConfigurationDescription.ApplicationCodeConfigurationDescription; applicationCodeConfigurationDescription != nil {
		mApplicationCodeConfiguration := map[string]interface{}{
			"code_content_type": applicationCodeConfigurationDescription.CodeContentType,
		}

		if codeContentDescription := applicationCodeConfigurationDescription.CodeContentDescription; codeContentDescription != nil {
			mCodeContent := map[string]interface{}{
				"text_content": aws.ToString(codeContentDescription.TextContent),
			}

			if s3ApplicationCodeLocationDescription := codeContentDescription.S3ApplicationCodeLocationDescription; s3ApplicationCodeLocationDescription != nil {
				mS3ContentLocation := map[string]interface{}{
					"bucket_arn":     aws.ToString(s3ApplicationCodeLocationDescription.BucketARN),
					"file_key":       aws.ToString(s3ApplicationCodeLocationDescription.FileKey),
					"object_version": aws.ToString(s3ApplicationCodeLocationDescription.ObjectVersion),
				}

				mCodeContent["s3_content_location"] = []interface{}{mS3ContentLocation}
			}

			mApplicationCodeConfiguration["code_content"] = []interface{}{mCodeContent}
		}

		mApplicationConfiguration["application_code_configuration"] = []interface{}{mApplicationCodeConfiguration}
	}

	if applicationSnapshotConfigurationDescription := applicationConfigurationDescription.ApplicationSnapshotConfigurationDescription; applicationSnapshotConfigurationDescription != nil {
		mApplicationSnapshotConfiguration := map[string]interface{}{
			"snapshots_enabled": aws.ToBool(applicationSnapshotConfigurationDescription.SnapshotsEnabled),
		}

		mApplicationConfiguration["application_snapshot_configuration"] = []interface{}{mApplicationSnapshotConfiguration}
	}

	if environmentPropertyDescriptions := applicationConfigurationDescription.EnvironmentPropertyDescriptions; environmentPropertyDescriptions != nil && len(environmentPropertyDescriptions.PropertyGroupDescriptions) > 0 {
		mEnvironmentProperties := map[string]interface{}{}

		vPropertyGroups := []interface{}{}

		for _, propertyGroup := range environmentPropertyDescriptions.PropertyGroupDescriptions {
			mPropertyGroup := map[string]interface{}{
				"property_group_id": aws.ToString(propertyGroup.PropertyGroupId),
				"property_map":      flex.FlattenStringValueMap(propertyGroup.PropertyMap),
			}

			vPropertyGroups = append(vPropertyGroups, mPropertyGroup)
		}

		mEnvironmentProperties["property_group"] = vPropertyGroups

		mApplicationConfiguration["environment_properties"] = []interface{}{mEnvironmentProperties}
	}

	if flinkApplicationConfigurationDescription := applicationConfigurationDescription.FlinkApplicationConfigurationDescription; flinkApplicationConfigurationDescription != nil {
		mFlinkApplicationConfiguration := map[string]interface{}{}

		if checkpointConfigurationDescription := flinkApplicationConfigurationDescription.CheckpointConfigurationDescription; checkpointConfigurationDescription != nil {
			mCheckpointConfiguration := map[string]interface{}{
				"checkpointing_enabled":         aws.ToBool(checkpointConfigurationDescription.CheckpointingEnabled),
				"checkpoint_interval":           aws.ToInt64(checkpointConfigurationDescription.CheckpointInterval),
				"configuration_type":            checkpointConfigurationDescription.ConfigurationType,
				"min_pause_between_checkpoints": aws.ToInt64(checkpointConfigurationDescription.MinPauseBetweenCheckpoints),
			}

			mFlinkApplicationConfiguration["checkpoint_configuration"] = []interface{}{mCheckpointConfiguration}
		}

		if monitoringConfigurationDescription := flinkApplicationConfigurationDescription.MonitoringConfigurationDescription; monitoringConfigurationDescription != nil {
			mMonitoringConfiguration := map[string]interface{}{
				"configuration_type": monitoringConfigurationDescription.ConfigurationType,
				"log_level":          monitoringConfigurationDescription.LogLevel,
				"metrics_level":      monitoringConfigurationDescription.MetricsLevel,
			}

			mFlinkApplicationConfiguration["monitoring_configuration"] = []interface{}{mMonitoringConfiguration}
		}

		if parallelismConfigurationDescription := flinkApplicationConfigurationDescription.ParallelismConfigurationDescription; parallelismConfigurationDescription != nil {
			mParallelismConfiguration := map[string]interface{}{
				"auto_scaling_enabled": aws.ToBool(parallelismConfigurationDescription.AutoScalingEnabled),
				"configuration_type":   parallelismConfigurationDescription.ConfigurationType,
				"parallelism":          aws.ToInt32(parallelismConfigurationDescription.Parallelism),
				"parallelism_per_kpu":  aws.ToInt32(parallelismConfigurationDescription.ParallelismPerKPU),
			}

			mFlinkApplicationConfiguration["parallelism_configuration"] = []interface{}{mParallelismConfiguration}
		}

		mApplicationConfiguration["flink_application_configuration"] = []interface{}{mFlinkApplicationConfiguration}
	}

	if runConfigurationDescription := applicationConfigurationDescription.RunConfigurationDescription; runConfigurationDescription != nil {
		mRunConfiguration := map[string]interface{}{}

		if applicationRestoreConfigurationDescription := runConfigurationDescription.ApplicationRestoreConfigurationDescription; applicationRestoreConfigurationDescription != nil {
			mApplicationRestoreConfiguration := map[string]interface{}{
				"application_restore_type": applicationRestoreConfigurationDescription.ApplicationRestoreType,
				"snapshot_name":            aws.ToString(applicationRestoreConfigurationDescription.SnapshotName),
			}

			mRunConfiguration["application_restore_configuration"] = []interface{}{mApplicationRestoreConfiguration}
		}

		if flinkRunConfigurationDescription := runConfigurationDescription.FlinkRunConfigurationDescription; flinkRunConfigurationDescription != nil {
			mFlinkRunConfiguration := map[string]interface{}{
				"allow_non_restored_state": aws.ToBool(flinkRunConfigurationDescription.AllowNonRestoredState),
			}

			mRunConfiguration["flink_run_configuration"] = []interface{}{mFlinkRunConfiguration}
		}

		mApplicationConfiguration["run_configuration"] = []interface{}{mRunConfiguration}
	}

	if sqlApplicationConfigurationDescription := applicationConfigurationDescription.SqlApplicationConfigurationDescription; sqlApplicationConfigurationDescription != nil {
		mSqlApplicationConfiguration := map[string]interface{}{}

		if inputDescriptions := sqlApplicationConfigurationDescription.InputDescriptions; len(inputDescriptions) > 0 {
			inputDescription := inputDescriptions[0]

			mInput := map[string]interface{}{
				"in_app_stream_names": inputDescription.InAppStreamNames,
				"input_id":            aws.ToString(inputDescription.InputId),
				names.AttrNamePrefix:  aws.ToString(inputDescription.NamePrefix),
			}

			if inputParallelism := inputDescription.InputParallelism; inputParallelism != nil {
				mInputParallelism := map[string]interface{}{
					"count": aws.ToInt32(inputParallelism.Count),
				}

				mInput["input_parallelism"] = []interface{}{mInputParallelism}
			}

			if inputSchema := inputDescription.InputSchema; inputSchema != nil {
				mInput["input_schema"] = flattenSourceSchema(inputSchema)
			}

			if inputProcessingConfigurationDescription := inputDescription.InputProcessingConfigurationDescription; inputProcessingConfigurationDescription != nil {
				mInputProcessingConfiguration := map[string]interface{}{}

				if inputLambdaProcessorDescription := inputProcessingConfigurationDescription.InputLambdaProcessorDescription; inputLambdaProcessorDescription != nil {
					mInputLambdaProcessor := map[string]interface{}{
						names.AttrResourceARN: aws.ToString(inputLambdaProcessorDescription.ResourceARN),
					}

					mInputProcessingConfiguration["input_lambda_processor"] = []interface{}{mInputLambdaProcessor}
				}

				mInput["input_processing_configuration"] = []interface{}{mInputProcessingConfiguration}
			}

			if inputStartingPositionConfiguration := inputDescription.InputStartingPositionConfiguration; inputStartingPositionConfiguration != nil {
				mInputStartingPositionConfiguration := map[string]interface{}{
					"input_starting_position": inputStartingPositionConfiguration.InputStartingPosition,
				}

				mInput["input_starting_position_configuration"] = []interface{}{mInputStartingPositionConfiguration}
			}

			if kinesisFirehoseInputDescription := inputDescription.KinesisFirehoseInputDescription; kinesisFirehoseInputDescription != nil {
				mKinesisFirehoseInput := map[string]interface{}{
					names.AttrResourceARN: aws.ToString(kinesisFirehoseInputDescription.ResourceARN),
				}

				mInput["kinesis_firehose_input"] = []interface{}{mKinesisFirehoseInput}
			}

			if kinesisStreamsInputDescription := inputDescription.KinesisStreamsInputDescription; kinesisStreamsInputDescription != nil {
				mKinesisStreamsInput := map[string]interface{}{
					names.AttrResourceARN: aws.ToString(kinesisStreamsInputDescription.ResourceARN),
				}

				mInput["kinesis_streams_input"] = []interface{}{mKinesisStreamsInput}
			}

			mSqlApplicationConfiguration["input"] = []interface{}{mInput}
		}

		if outputDescriptions := sqlApplicationConfigurationDescription.OutputDescriptions; len(outputDescriptions) > 0 {
			vOutputs := []interface{}{}

			for _, outputDescription := range outputDescriptions {
				mOutput := map[string]interface{}{
					names.AttrName: aws.ToString(outputDescription.Name),
					"output_id":    aws.ToString(outputDescription.OutputId),
				}

				if destinationSchema := outputDescription.DestinationSchema; destinationSchema != nil {
					mDestinationSchema := map[string]interface{}{
						"record_format_type": destinationSchema.RecordFormatType,
					}

					mOutput["destination_schema"] = []interface{}{mDestinationSchema}
				}

				if kinesisFirehoseOutputDescription := outputDescription.KinesisFirehoseOutputDescription; kinesisFirehoseOutputDescription != nil {
					mKinesisFirehoseOutput := map[string]interface{}{
						names.AttrResourceARN: aws.ToString(kinesisFirehoseOutputDescription.ResourceARN),
					}

					mOutput["kinesis_firehose_output"] = []interface{}{mKinesisFirehoseOutput}
				}

				if kinesisStreamsOutputDescription := outputDescription.KinesisStreamsOutputDescription; kinesisStreamsOutputDescription != nil {
					mKinesisStreamsOutput := map[string]interface{}{
						names.AttrResourceARN: aws.ToString(kinesisStreamsOutputDescription.ResourceARN),
					}

					mOutput["kinesis_streams_output"] = []interface{}{mKinesisStreamsOutput}
				}

				if lambdaOutputDescription := outputDescription.LambdaOutputDescription; lambdaOutputDescription != nil {
					mLambdaOutput := map[string]interface{}{
						names.AttrResourceARN: aws.ToString(lambdaOutputDescription.ResourceARN),
					}

					mOutput["lambda_output"] = []interface{}{mLambdaOutput}
				}

				vOutputs = append(vOutputs, mOutput)
			}

			mSqlApplicationConfiguration["output"] = vOutputs
		}

		if referenceDataSourceDescriptions := sqlApplicationConfigurationDescription.ReferenceDataSourceDescriptions; len(referenceDataSourceDescriptions) > 0 {
			referenceDataSourceDescription := referenceDataSourceDescriptions[0]

			mReferenceDataSource := map[string]interface{}{
				"reference_id":      aws.ToString(referenceDataSourceDescription.ReferenceId),
				names.AttrTableName: aws.ToString(referenceDataSourceDescription.TableName),
			}

			if referenceSchema := referenceDataSourceDescription.ReferenceSchema; referenceSchema != nil {
				mReferenceDataSource["reference_schema"] = flattenSourceSchema(referenceSchema)
			}

			if s3ReferenceDataSource := referenceDataSourceDescription.S3ReferenceDataSourceDescription; s3ReferenceDataSource != nil {
				mS3ReferenceDataSource := map[string]interface{}{
					"bucket_arn": aws.ToString(s3ReferenceDataSource.BucketARN),
					"file_key":   aws.ToString(s3ReferenceDataSource.FileKey),
				}

				mReferenceDataSource["s3_reference_data_source"] = []interface{}{mS3ReferenceDataSource}
			}

			mSqlApplicationConfiguration["reference_data_source"] = []interface{}{mReferenceDataSource}
		}

		mApplicationConfiguration["sql_application_configuration"] = []interface{}{mSqlApplicationConfiguration}
	}

	if vpcConfigurationDescriptions := applicationConfigurationDescription.VpcConfigurationDescriptions; len(vpcConfigurationDescriptions) > 0 {
		vpcConfigurationDescription := vpcConfigurationDescriptions[0]

		mVpcConfiguration := map[string]interface{}{
			names.AttrSecurityGroupIDs: vpcConfigurationDescription.SecurityGroupIds,
			names.AttrSubnetIDs:        vpcConfigurationDescription.SubnetIds,
			"vpc_configuration_id":     aws.ToString(vpcConfigurationDescription.VpcConfigurationId),
			names.AttrVPCID:            aws.ToString(vpcConfigurationDescription.VpcId),
		}

		mApplicationConfiguration[names.AttrVPCConfiguration] = []interface{}{mVpcConfiguration}
	}

	return []interface{}{mApplicationConfiguration}
}

func flattenCloudWatchLoggingOptionDescriptions(cloudWatchLoggingOptionDescriptions []awstypes.CloudWatchLoggingOptionDescription) []interface{} {
	if len(cloudWatchLoggingOptionDescriptions) == 0 {
		return []interface{}{}
	}

	cloudWatchLoggingOptionDescription := cloudWatchLoggingOptionDescriptions[0]

	mCloudWatchLoggingOption := map[string]interface{}{
		"cloudwatch_logging_option_id": aws.ToString(cloudWatchLoggingOptionDescription.CloudWatchLoggingOptionId),
		"log_stream_arn":               aws.ToString(cloudWatchLoggingOptionDescription.LogStreamARN),
	}

	return []interface{}{mCloudWatchLoggingOption}
}

func flattenSourceSchema(sourceSchema *awstypes.SourceSchema) []interface{} {
	if sourceSchema == nil {
		return []interface{}{}
	}

	mSourceSchema := map[string]interface{}{
		"record_encoding": aws.ToString(sourceSchema.RecordEncoding),
	}

	if len(sourceSchema.RecordColumns) > 0 {
		vRecordColumns := []interface{}{}

		for _, recordColumn := range sourceSchema.RecordColumns {
			mRecordColumn := map[string]interface{}{
				"mapping":      aws.ToString(recordColumn.Mapping),
				names.AttrName: aws.ToString(recordColumn.Name),
				"sql_type":     aws.ToString(recordColumn.SqlType),
			}

			vRecordColumns = append(vRecordColumns, mRecordColumn)
		}

		mSourceSchema["record_column"] = vRecordColumns
	}

	if recordFormat := sourceSchema.RecordFormat; recordFormat != nil {
		mRecordFormat := map[string]interface{}{
			"record_format_type": recordFormat.RecordFormatType,
		}

		if mappingParameters := recordFormat.MappingParameters; mappingParameters != nil {
			mMappingParameters := map[string]interface{}{}

			if csvMappingParameters := mappingParameters.CSVMappingParameters; csvMappingParameters != nil {
				mCsvMappingParameters := map[string]interface{}{
					"record_column_delimiter": aws.ToString(csvMappingParameters.RecordColumnDelimiter),
					"record_row_delimiter":    aws.ToString(csvMappingParameters.RecordRowDelimiter),
				}

				mMappingParameters["csv_mapping_parameters"] = []interface{}{mCsvMappingParameters}
			}

			if jsonMappingParameters := mappingParameters.JSONMappingParameters; jsonMappingParameters != nil {
				mJsonMappingParameters := map[string]interface{}{
					"record_row_path": aws.ToString(jsonMappingParameters.RecordRowPath),
				}

				mMappingParameters["json_mapping_parameters"] = []interface{}{mJsonMappingParameters}
			}

			mRecordFormat["mapping_parameters"] = []interface{}{mMappingParameters}
		}

		mSourceSchema["record_format"] = []interface{}{mRecordFormat}
	}

	return []interface{}{mSourceSchema}
}

func expandStartApplicationInput(d *schema.ResourceData) *kinesisanalyticsv2.StartApplicationInput {
	apiObject := &kinesisanalyticsv2.StartApplicationInput{
		ApplicationName:  aws.String(d.Get(names.AttrName).(string)),
		RunConfiguration: &awstypes.RunConfiguration{},
	}

	if v, ok := d.GetOk("application_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		if v, ok := tfMap["sql_application_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})

			if v, ok := tfMap["input"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				tfMap := v[0].(map[string]interface{})

				if v, ok := tfMap["input_starting_position_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
					tfMap := v[0].(map[string]interface{})

					if v, ok := tfMap["input_starting_position"].(string); ok && v != "" {
						apiObject.RunConfiguration.SqlRunConfigurations = []awstypes.SqlRunConfiguration{{
							InputStartingPositionConfiguration: &awstypes.InputStartingPositionConfiguration{
								InputStartingPosition: awstypes.InputStartingPosition(v),
							},
						}}
					}
				}
			}
		}

		if v, ok := tfMap["run_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})

			if v, ok := tfMap["application_restore_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				tfMap := v[0].(map[string]interface{})

				apiObject.RunConfiguration.ApplicationRestoreConfiguration = &awstypes.ApplicationRestoreConfiguration{}

				if v, ok := tfMap["application_restore_type"].(string); ok && v != "" {
					apiObject.RunConfiguration.ApplicationRestoreConfiguration.ApplicationRestoreType = awstypes.ApplicationRestoreType(v)
				}

				if v, ok := tfMap["snapshot_name"].(string); ok && v != "" {
					apiObject.RunConfiguration.ApplicationRestoreConfiguration.SnapshotName = aws.String(v)
				}
			}

			if v, ok := tfMap["flink_run_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				tfMap := v[0].(map[string]interface{})

				if v, ok := tfMap["allow_non_restored_state"].(bool); ok {
					apiObject.RunConfiguration.FlinkRunConfiguration = &awstypes.FlinkRunConfiguration{
						AllowNonRestoredState: aws.Bool(v),
					}
				}
			}
		}
	}

	return apiObject
}

func expandStopApplicationInput(d *schema.ResourceData) *kinesisanalyticsv2.StopApplicationInput {
	apiObject := &kinesisanalyticsv2.StopApplicationInput{
		ApplicationName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("force_stop"); ok {
		apiObject.Force = aws.Bool(v.(bool))
	}

	return apiObject
}
