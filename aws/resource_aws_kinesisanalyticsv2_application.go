package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsKinesisAnalyticsV2Application() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisAnalyticsV2ApplicationCreate,
		Read:   resourceAwsKinesisAnalyticsV2ApplicationRead,
		Update: resourceAwsKinesisAnalyticsV2ApplicationUpdate,
		Delete: resourceAwsKinesisAnalyticsV2ApplicationDelete,

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("application_configuration.0.sql_application_configuration.0.input", func(_ context.Context, old, new, meta interface{}) bool {
				// An existing input configuration cannot be deleted.
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
		),

		Importer: &schema.ResourceImporter{
			State: resourceAwsKinesisAnalyticsV2ApplicationImport,
		},

		Schema: map[string]*schema.Schema{
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
												"text_content": {
													Type:          schema.TypeString,
													Optional:      true,
													ValidateFunc:  validation.StringLenBetween(0, 102400),
													ConflictsWith: []string{"application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location"},
												},

												"s3_content_location": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validateArn,
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
											},
										},
									},

									"code_content_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.CodeContentType_Values(), false),
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
														validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.ConfigurationType_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.ConfigurationType_Values(), false),
												},

												"log_level": {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.LogLevel_Values(), false),
												},

												"metrics_level": {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.MetricsLevel_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.ConfigurationType_Values(), false),
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
																		"resource_arn": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validateArn,
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
																Type:     schema.TypeSet,
																Required: true,
																MaxItems: 1000,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"mapping": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},

																		"name": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[^-\s<>&]+$`), "must not include hyphen, whitespace, angle bracket, or ampersand characters"),
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
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.RecordFormatType_Values(), false),
																		},
																	},
																},
															},
														},
													},
												},

												"input_starting_position_configuration": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"input_starting_position": {
																Type:     schema.TypeString,
																Computed: true,
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
															"resource_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validateArn,
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
															"resource_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validateArn,
															},
														},
													},
													ExactlyOneOf: []string{
														"application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input",
														"application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input",
													},
												},

												"name_prefix": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 32),
														validation.StringMatch(regexp.MustCompile(`^[^-\s<>&]+$`), "must not include hyphen, whitespace, angle bracket, or ampersand characters"),
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
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.RecordFormatType_Values(), false),
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
															"resource_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validateArn,
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
															"resource_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validateArn,
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
															"resource_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validateArn,
															},
														},
													},
												},

												"name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 32),
														validation.StringMatch(regexp.MustCompile(`^[^-\s<>&]+$`), "must not include hyphen, whitespace, angle bracket, or ampersand characters"),
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
																Type:     schema.TypeSet,
																Required: true,
																MaxItems: 1000,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"mapping": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},

																		"name": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[^-\s<>&]+$`), "must not include hyphen, whitespace, angle bracket, or ampersand characters"),
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
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.RecordFormatType_Values(), false),
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
																ValidateFunc: validateArn,
															},

															"file_key": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},

												"table_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 32),
												},
											},
										},
									},
								},
							},
							ConflictsWith: []string{"application_configuration.0.flink_application_configuration"},
						},
					},
				},
			},

			"arn": {
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
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"create_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},

			"last_update_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},

			"runtime_environment": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(kinesisanalyticsv2.RuntimeEnvironment_Values(), false),
			},

			"service_execution_role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),

			"version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsKinesisAnalyticsV2ApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn

	input := &kinesisanalyticsv2.CreateApplicationInput{
		ApplicationConfiguration: expandKinesisAnalyticsV2ApplicationConfiguration(d.Get("application_configuration").([]interface{})),
		ApplicationDescription:   aws.String(d.Get("description").(string)),
		ApplicationName:          aws.String(d.Get("name").(string)),
		CloudWatchLoggingOptions: expandKinesisAnalyticsV2CloudWatchLoggingOptions(d.Get("cloudwatch_logging_options").([]interface{})),
		RuntimeEnvironment:       aws.String(d.Get("runtime_environment").(string)),
		ServiceExecutionRole:     aws.String(d.Get("service_execution_role").(string)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().Kinesisanalyticsv2Tags()
	}

	log.Printf("[DEBUG] Creating Kinesis Analytics v2 Application: %s", input)

	outputRaw, err := kinesisAnalyticsV2RetryIAMEventualConsistency(func() (interface{}, error) {
		return conn.CreateApplication(input)
	})

	if err != nil {
		return fmt.Errorf("error creating Kinesis Analytics v2 Application: %w", err)
	}

	output := outputRaw.(*kinesisanalyticsv2.CreateApplicationOutput)

	d.SetId(aws.StringValue(output.ApplicationDetail.ApplicationARN))

	return resourceAwsKinesisAnalyticsV2ApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsV2ApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	application, err := finder.ApplicationByName(conn, d.Get("name").(string))

	if isAWSErr(err, kinesisanalyticsv2.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Kinesis Analytics v2 Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Kinesis Analytics v2 Application (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(application.ApplicationARN)
	d.Set("arn", arn)
	d.Set("create_timestamp", aws.TimeValue(application.CreateTimestamp).Format(time.RFC3339))
	d.Set("description", aws.StringValue(application.ApplicationDescription))
	d.Set("last_update_timestamp", aws.TimeValue(application.LastUpdateTimestamp).Format(time.RFC3339))
	d.Set("name", application.ApplicationName)
	d.Set("runtime_environment", application.RuntimeEnvironment)
	d.Set("service_execution_role", aws.StringValue(application.ServiceExecutionRole))
	d.Set("status", aws.StringValue(application.ApplicationStatus))
	d.Set("version_id", int(aws.Int64Value(application.ApplicationVersionId)))

	if err := d.Set("application_configuration", flattenKinesisAnalyticsV2ApplicationConfigurationDescription(application.ApplicationConfigurationDescription)); err != nil {
		return fmt.Errorf("error setting application_configuration: %w", err)
	}

	if err := d.Set("cloudwatch_logging_options", flattenKinesisAnalyticsV2CloudWatchLoggingOptionDescriptions(application.CloudWatchLoggingOptionDescriptions)); err != nil {
		return fmt.Errorf("error setting cloudwatch_logging_options: %w", err)
	}

	tags, err := keyvaluetags.Kinesisanalyticsv2ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Kinesis Analytics v2 Application (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsKinesisAnalyticsV2ApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn

	if d.HasChanges("application_configuration", "cloudwatch_logging_options", "service_execution_role") {
		applicationName := d.Get("name").(string)
		currentApplicationVersionId := int64(d.Get("version_id").(int))
		updateApplication := false

		input := &kinesisanalyticsv2.UpdateApplicationInput{
			ApplicationName: aws.String(applicationName),
		}

		if d.HasChange("cloudwatch_logging_options") {
			o, n := d.GetChange("cloudwatch_logging_options")

			if len(o.([]interface{})) == 0 {
				// Add new CloudWatch logging options.
				input := &kinesisanalyticsv2.AddApplicationCloudWatchLoggingOptionInput{
					ApplicationName: aws.String(applicationName),
					CloudWatchLoggingOption: &kinesisanalyticsv2.CloudWatchLoggingOption{
						LogStreamARN: aws.String(n.([]interface{})[0].(map[string]interface{})["log_stream_arn"].(string)),
					},
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionId),
				}

				log.Printf("[DEBUG] Adding Kinesis Analytics v2 Application (%s) CloudWatch logging option: %s", d.Id(), input)

				outputRaw, err := kinesisAnalyticsV2RetryIAMEventualConsistency(func() (interface{}, error) {
					return conn.AddApplicationCloudWatchLoggingOption(input)
				})

				if err != nil {
					return fmt.Errorf("error adding Kinesis Analytics v2 Application (%s) CloudWatch logging option: %w", d.Id(), err)
				}

				output := outputRaw.(*kinesisanalyticsv2.AddApplicationCloudWatchLoggingOptionOutput)

				currentApplicationVersionId = aws.Int64Value(output.ApplicationVersionId)
			} else if len(n.([]interface{})) == 0 {
				// Delete existing CloudWatch logging options.
				input := &kinesisanalyticsv2.DeleteApplicationCloudWatchLoggingOptionInput{
					ApplicationName:             aws.String(applicationName),
					CloudWatchLoggingOptionId:   aws.String(o.([]interface{})[0].(map[string]interface{})["cloudwatch_logging_option_id"].(string)),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionId),
				}

				log.Printf("[DEBUG] Deleting Kinesis Analytics v2 Application (%s) CloudWatch logging option: %s", d.Id(), input)

				outputRaw, err := kinesisAnalyticsV2RetryIAMEventualConsistency(func() (interface{}, error) {
					return conn.DeleteApplicationCloudWatchLoggingOption(input)
				})

				if err != nil {
					return fmt.Errorf("error deleting Kinesis Analytics v2 Application (%s) CloudWatch logging option: %w", d.Id(), err)
				}

				output := outputRaw.(*kinesisanalyticsv2.DeleteApplicationCloudWatchLoggingOptionOutput)

				currentApplicationVersionId = aws.Int64Value(output.ApplicationVersionId)
			} else {
				// Update existing CloudWatch logging options.
				input.CloudWatchLoggingOptionUpdates = []*kinesisanalyticsv2.CloudWatchLoggingOptionUpdate{
					{
						CloudWatchLoggingOptionId: aws.String(o.([]interface{})[0].(map[string]interface{})["cloudwatch_logging_option_id"].(string)),
						LogStreamARNUpdate:        aws.String(n.([]interface{})[0].(map[string]interface{})["log_stream_arn"].(string)),
					},
				}

				updateApplication = true
			}
		}

		if d.HasChange("service_execution_role") {
			input.ServiceExecutionRoleUpdate = aws.String(d.Get("service_execution_role").(string))

			updateApplication = true
		}

		if d.HasChange("application_configuration") {
			applicationConfigurationUpdate := &kinesisanalyticsv2.ApplicationConfigurationUpdate{}

			if d.HasChange("application_configuration.0.application_code_configuration") {
				applicationConfigurationUpdate.ApplicationCodeConfigurationUpdate = expandKinesisAnalyticsV2ApplicationCodeConfigurationUpdate(d.Get("application_configuration.0.application_code_configuration").([]interface{}))

				updateApplication = true
			}

			if d.HasChange("application_configuration.0.application_snapshot_configuration") {
				applicationConfigurationUpdate.ApplicationSnapshotConfigurationUpdate = expandKinesisAnalyticsV2ApplicationSnapshotConfigurationUpdate(d.Get("application_configuration.0.application_snapshot_configuration").([]interface{}))

				updateApplication = true
			}

			if d.HasChange("application_configuration.0.environment_properties") {
				applicationConfigurationUpdate.EnvironmentPropertyUpdates = expandKinesisAnalyticsV2EnvironmentPropertyUpdates(d.Get("application_configuration.0.environment_properties").([]interface{}))

				updateApplication = true
			}

			if d.HasChange("application_configuration.0.flink_application_configuration") {
				applicationConfigurationUpdate.FlinkApplicationConfigurationUpdate = expandKinesisAnalyticsV2ApplicationFlinkApplicationConfigurationUpdate(d.Get("application_configuration.0.flink_application_configuration").([]interface{}))

				updateApplication = true
			}

			if d.HasChange("application_configuration.0.sql_application_configuration") {
				sqlApplicationConfigurationUpdate := &kinesisanalyticsv2.SqlApplicationConfigurationUpdate{}

				if d.HasChange("application_configuration.0.sql_application_configuration.0.input") {
					o, n := d.GetChange("application_configuration.0.sql_application_configuration.0.input")

					if len(o.([]interface{})) == 0 {
						// Add new input.
						input := &kinesisanalyticsv2.AddApplicationInputInput{
							ApplicationName:             aws.String(applicationName),
							CurrentApplicationVersionId: aws.Int64(currentApplicationVersionId),
							Input:                       expandKinesisAnalyticsV2Input(n.([]interface{})),
						}

						log.Printf("[DEBUG] Adding Kinesis Analytics v2 Application (%s) input: %s", d.Id(), input)

						outputRaw, err := kinesisAnalyticsV2RetryIAMEventualConsistency(func() (interface{}, error) {
							return conn.AddApplicationInput(input)
						})

						if err != nil {
							return fmt.Errorf("error adding Kinesis Analytics v2 Application (%s) input: %w", d.Id(), err)
						}

						output := outputRaw.(*kinesisanalyticsv2.AddApplicationInputOutput)

						currentApplicationVersionId = aws.Int64Value(output.ApplicationVersionId)
					} else if len(n.([]interface{})) == 0 {
						// The existing input cannot be deleted.
						// This should be handled by the CustomizeDiff function above.
						return fmt.Errorf("error deleting Kinesis Analytics v2 Application (%s) input", d.Id())
					} else {
						// Update existing input.
						inputUpdate := expandKinesisAnalyticsV2InputUpdate(n.([]interface{}))

						if d.HasChange("application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration") {
							o, n := d.GetChange("application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration")

							if len(o.([]interface{})) == 0 {
								// Add new input processing configuration.
								input := &kinesisanalyticsv2.AddApplicationInputProcessingConfigurationInput{
									ApplicationName:              aws.String(applicationName),
									CurrentApplicationVersionId:  aws.Int64(currentApplicationVersionId),
									InputId:                      inputUpdate.InputId,
									InputProcessingConfiguration: expandKinesisAnalyticsV2InputProcessingConfiguration(n.([]interface{})),
								}

								log.Printf("[DEBUG] Adding Kinesis Analytics v2 Application (%s) input processing configuration: %s", d.Id(), input)

								outputRaw, err := kinesisAnalyticsV2RetryIAMEventualConsistency(func() (interface{}, error) {
									return conn.AddApplicationInputProcessingConfiguration(input)
								})

								if err != nil {
									return fmt.Errorf("error adding Kinesis Analytics v2 Application (%s) input processing configuration: %w", d.Id(), err)
								}

								output := outputRaw.(*kinesisanalyticsv2.AddApplicationInputProcessingConfigurationOutput)

								currentApplicationVersionId = aws.Int64Value(output.ApplicationVersionId)
							} else if len(n.([]interface{})) == 0 {
								// Delete existing input processing configuration.
								input := &kinesisanalyticsv2.DeleteApplicationInputProcessingConfigurationInput{
									ApplicationName:             aws.String(applicationName),
									CurrentApplicationVersionId: aws.Int64(currentApplicationVersionId),
									InputId:                     inputUpdate.InputId,
								}

								log.Printf("[DEBUG] Deleting Kinesis Analytics v2 Application (%s) input processing configuration: %s", d.Id(), input)

								outputRaw, err := kinesisAnalyticsV2RetryIAMEventualConsistency(func() (interface{}, error) {
									return conn.DeleteApplicationInputProcessingConfiguration(input)
								})

								if err != nil {
									return fmt.Errorf("error deleting Kinesis Analytics v2 Application (%s) input processing configuration: %w", d.Id(), err)
								}

								output := outputRaw.(*kinesisanalyticsv2.DeleteApplicationInputProcessingConfigurationOutput)

								currentApplicationVersionId = aws.Int64Value(output.ApplicationVersionId)
							} else {
								// Update existing input processing configuration.
								// This is handled via the updating of the existing input.
							}
						}

						sqlApplicationConfigurationUpdate.InputUpdates = []*kinesisanalyticsv2.InputUpdate{inputUpdate}

						updateApplication = true
					}
				}

				if updateApplication {
					applicationConfigurationUpdate.SqlApplicationConfigurationUpdate = sqlApplicationConfigurationUpdate
				}
			}

			input.ApplicationConfigurationUpdate = applicationConfigurationUpdate
		}

		if updateApplication {
			input.CurrentApplicationVersionId = aws.Int64(currentApplicationVersionId)

			log.Printf("[DEBUG] Updating Kinesis Analytics v2 Application (%s): %s", d.Id(), input)

			outputRaw, err := kinesisAnalyticsV2RetryIAMEventualConsistency(func() (interface{}, error) {
				return conn.UpdateApplication(input)
			})

			if err != nil {
				return fmt.Errorf("error updating Kinesis Analytics v2 Application (%s): %w", d.Id(), err)
			}

			output := outputRaw.(*kinesisanalyticsv2.UpdateApplicationOutput)

			currentApplicationVersionId = aws.Int64Value(output.ApplicationDetail.ApplicationVersionId)
		}
	}

	if d.HasChange("tags") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Kinesisanalyticsv2UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Kinesis Analytics Application (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsKinesisAnalyticsV2ApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsV2ApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn

	createTimestamp, err := time.Parse(time.RFC3339, d.Get("create_timestamp").(string))
	if err != nil {
		return fmt.Errorf("error parsing create_timestamp: %w", err)
	}

	applicationName := d.Get("name").(string)

	log.Printf("[DEBUG] Deleting Kinesis Analytics v2 Application (%s)", d.Id())
	_, err = conn.DeleteApplication(&kinesisanalyticsv2.DeleteApplicationInput{
		ApplicationName: aws.String(applicationName),
		CreateTimestamp: aws.Time(createTimestamp),
	})

	if isAWSErr(err, kinesisanalyticsv2.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Kinesis Analytics v2 Application (%s): %w", d.Id(), err)
	}

	_, err = waiter.ApplicationDeleted(conn, applicationName, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for Kinesis Analytics v2 Application (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsKinesisAnalyticsV2ApplicationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	arn, err := arn.Parse(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Error parsing ARN %q: %w", d.Id(), err)
	}

	// application/<name>
	parts := strings.Split(arn.Resource, "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Unexpected ARN format: %q", d.Id())
	}

	d.Set("name", parts[1])

	return []*schema.ResourceData{d}, nil
}

func expandKinesisAnalyticsV2ApplicationConfiguration(vApplicationConfiguration []interface{}) *kinesisanalyticsv2.ApplicationConfiguration {
	if len(vApplicationConfiguration) == 0 || vApplicationConfiguration[0] == nil {
		return nil
	}

	applicationConfiguration := &kinesisanalyticsv2.ApplicationConfiguration{}

	mApplicationConfiguration := vApplicationConfiguration[0].(map[string]interface{})

	if vApplicationCodeConfiguration, ok := mApplicationConfiguration["application_code_configuration"].([]interface{}); ok && len(vApplicationCodeConfiguration) > 0 && vApplicationCodeConfiguration[0] != nil {
		applicationCodeConfiguration := &kinesisanalyticsv2.ApplicationCodeConfiguration{}

		mApplicationCodeConfiguration := vApplicationCodeConfiguration[0].(map[string]interface{})

		if vCodeContent, ok := mApplicationCodeConfiguration["code_content"].([]interface{}); ok && len(vCodeContent) > 0 && vCodeContent[0] != nil {
			codeContent := &kinesisanalyticsv2.CodeContent{}

			mCodeContent := vCodeContent[0].(map[string]interface{})

			if vS3ContentLocation, ok := mCodeContent["s3_content_location"].([]interface{}); ok && len(vS3ContentLocation) > 0 && vS3ContentLocation[0] != nil {
				s3ContentLocation := &kinesisanalyticsv2.S3ContentLocation{}

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
			applicationCodeConfiguration.CodeContentType = aws.String(vCodeContentType)
		}

		applicationConfiguration.ApplicationCodeConfiguration = applicationCodeConfiguration
	}

	if vApplicationSnapshotConfiguration, ok := mApplicationConfiguration["application_snapshot_configuration"].([]interface{}); ok && len(vApplicationSnapshotConfiguration) > 0 && vApplicationSnapshotConfiguration[0] != nil {
		applicationSnapshotConfiguration := &kinesisanalyticsv2.ApplicationSnapshotConfiguration{}

		mApplicationSnapshotConfiguration := vApplicationSnapshotConfiguration[0].(map[string]interface{})

		if vSnapshotsEnabled, ok := mApplicationSnapshotConfiguration["snapshots_enabled"].(bool); ok {
			applicationSnapshotConfiguration.SnapshotsEnabled = aws.Bool(vSnapshotsEnabled)
		}

		applicationConfiguration.ApplicationSnapshotConfiguration = applicationSnapshotConfiguration
	}

	if vEnvironmentProperties, ok := mApplicationConfiguration["environment_properties"].([]interface{}); ok && len(vEnvironmentProperties) > 0 && vEnvironmentProperties[0] != nil {
		environmentProperties := &kinesisanalyticsv2.EnvironmentProperties{}

		mEnvironmentProperties := vEnvironmentProperties[0].(map[string]interface{})

		if vPropertyGroups, ok := mEnvironmentProperties["property_group"].(*schema.Set); ok && vPropertyGroups.Len() > 0 {
			environmentProperties.PropertyGroups = expandKinesisAnalyticsV2PropertyGroups(vPropertyGroups.List())
		}

		applicationConfiguration.EnvironmentProperties = environmentProperties
	}

	if vFlinkApplicationConfiguration, ok := mApplicationConfiguration["flink_application_configuration"].([]interface{}); ok && len(vFlinkApplicationConfiguration) > 0 && vFlinkApplicationConfiguration[0] != nil {
		flinkApplicationConfiguration := &kinesisanalyticsv2.FlinkApplicationConfiguration{}

		mFlinkApplicationConfiguration := vFlinkApplicationConfiguration[0].(map[string]interface{})

		if vCheckpointConfiguration, ok := mFlinkApplicationConfiguration["checkpoint_configuration"].([]interface{}); ok && len(vCheckpointConfiguration) > 0 && vCheckpointConfiguration[0] != nil {
			checkpointConfiguration := &kinesisanalyticsv2.CheckpointConfiguration{}

			mCheckpointConfiguration := vCheckpointConfiguration[0].(map[string]interface{})

			if vConfigurationType, ok := mCheckpointConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
				checkpointConfiguration.ConfigurationType = aws.String(vConfigurationType)

				if vConfigurationType == kinesisanalyticsv2.ConfigurationTypeCustom {
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
			monitoringConfiguration := &kinesisanalyticsv2.MonitoringConfiguration{}

			mMonitoringConfiguration := vMonitoringConfiguration[0].(map[string]interface{})

			if vConfigurationType, ok := mMonitoringConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
				monitoringConfiguration.ConfigurationType = aws.String(vConfigurationType)

				if vConfigurationType == kinesisanalyticsv2.ConfigurationTypeCustom {
					if vLogLevel, ok := mMonitoringConfiguration["log_level"].(string); ok && vLogLevel != "" {
						monitoringConfiguration.LogLevel = aws.String(vLogLevel)
					}
					if vMetricsLevel, ok := mMonitoringConfiguration["metrics_level"].(string); ok && vMetricsLevel != "" {
						monitoringConfiguration.MetricsLevel = aws.String(vMetricsLevel)
					}
				}
			}

			flinkApplicationConfiguration.MonitoringConfiguration = monitoringConfiguration
		}

		if vParallelismConfiguration, ok := mFlinkApplicationConfiguration["parallelism_configuration"].([]interface{}); ok && len(vParallelismConfiguration) > 0 && vParallelismConfiguration[0] != nil {
			parallelismConfiguration := &kinesisanalyticsv2.ParallelismConfiguration{}

			mParallelismConfiguration := vParallelismConfiguration[0].(map[string]interface{})

			if vConfigurationType, ok := mParallelismConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
				parallelismConfiguration.ConfigurationType = aws.String(vConfigurationType)

				if vConfigurationType == kinesisanalyticsv2.ConfigurationTypeCustom {
					if vAutoScalingEnabled, ok := mParallelismConfiguration["auto_scaling_enabled"].(bool); ok {
						parallelismConfiguration.AutoScalingEnabled = aws.Bool(vAutoScalingEnabled)
					}
					if vParallelism, ok := mParallelismConfiguration["parallelism"].(int); ok {
						parallelismConfiguration.Parallelism = aws.Int64(int64(vParallelism))
					}
					if vParallelismPerKPU, ok := mParallelismConfiguration["parallelism_per_kpu"].(int); ok {
						parallelismConfiguration.ParallelismPerKPU = aws.Int64(int64(vParallelismPerKPU))
					}
				}
			}

			flinkApplicationConfiguration.ParallelismConfiguration = parallelismConfiguration
		}

		applicationConfiguration.FlinkApplicationConfiguration = flinkApplicationConfiguration
	}

	if vSqlApplicationConfiguration, ok := mApplicationConfiguration["sql_application_configuration"].([]interface{}); ok && len(vSqlApplicationConfiguration) > 0 && vSqlApplicationConfiguration[0] != nil {
		sqlApplicationConfiguration := &kinesisanalyticsv2.SqlApplicationConfiguration{}

		mSqlApplicationConfiguration := vSqlApplicationConfiguration[0].(map[string]interface{})

		if vInput, ok := mSqlApplicationConfiguration["input"].([]interface{}); ok && len(vInput) > 0 && vInput[0] != nil {
			sqlApplicationConfiguration.Inputs = []*kinesisanalyticsv2.Input{expandKinesisAnalyticsV2Input(vInput)}
		}

		if vOutputs, ok := mSqlApplicationConfiguration["output"].(*schema.Set); ok && vOutputs.Len() > 0 {
			sqlApplicationConfiguration.Outputs = expandKinesisAnalyticsV2Outputs(vOutputs.List())
		}

		if vReferenceDataSource, ok := mSqlApplicationConfiguration["reference_data_source"].([]interface{}); ok && len(vReferenceDataSource) > 0 && vReferenceDataSource[0] != nil {
			sqlApplicationConfiguration.ReferenceDataSources = []*kinesisanalyticsv2.ReferenceDataSource{expandKinesisAnalyticsV2ReferenceDataSource(vReferenceDataSource)}
		}

		applicationConfiguration.SqlApplicationConfiguration = sqlApplicationConfiguration
	}

	return applicationConfiguration
}

func flattenKinesisAnalyticsV2ApplicationConfigurationDescription(applicationConfigurationDescription *kinesisanalyticsv2.ApplicationConfigurationDescription) []interface{} {
	if applicationConfigurationDescription == nil {
		return []interface{}{}
	}

	mApplicationConfiguration := map[string]interface{}{}

	if applicationCodeConfigurationDescription := applicationConfigurationDescription.ApplicationCodeConfigurationDescription; applicationCodeConfigurationDescription != nil {
		mApplicationCodeConfiguration := map[string]interface{}{
			"code_content_type": aws.StringValue(applicationCodeConfigurationDescription.CodeContentType),
		}

		if codeContentDescription := applicationCodeConfigurationDescription.CodeContentDescription; codeContentDescription != nil {
			mCodeContent := map[string]interface{}{
				"text_content": aws.StringValue(codeContentDescription.TextContent),
			}

			if s3ApplicationCodeLocationDescription := codeContentDescription.S3ApplicationCodeLocationDescription; s3ApplicationCodeLocationDescription != nil {
				mS3ContentLocation := map[string]interface{}{
					"bucket_arn":     aws.StringValue(s3ApplicationCodeLocationDescription.BucketARN),
					"file_key":       aws.StringValue(s3ApplicationCodeLocationDescription.FileKey),
					"object_version": aws.StringValue(s3ApplicationCodeLocationDescription.ObjectVersion),
				}

				mCodeContent["s3_content_location"] = []interface{}{mS3ContentLocation}
			}

			mApplicationCodeConfiguration["code_content"] = []interface{}{mCodeContent}
		}

		mApplicationConfiguration["application_code_configuration"] = []interface{}{mApplicationCodeConfiguration}
	}

	if applicationSnapshotConfigurationDescription := applicationConfigurationDescription.ApplicationSnapshotConfigurationDescription; applicationSnapshotConfigurationDescription != nil {
		mApplicationSnapshotConfiguration := map[string]interface{}{
			"snapshots_enabled": aws.BoolValue(applicationSnapshotConfigurationDescription.SnapshotsEnabled),
		}

		mApplicationConfiguration["application_snapshot_configuration"] = []interface{}{mApplicationSnapshotConfiguration}
	}

	if environmentPropertyDescriptions := applicationConfigurationDescription.EnvironmentPropertyDescriptions; environmentPropertyDescriptions != nil && len(environmentPropertyDescriptions.PropertyGroupDescriptions) > 0 {
		mEnvironmentProperties := map[string]interface{}{}

		vPropertyGroups := []interface{}{}

		for _, propertyGroup := range environmentPropertyDescriptions.PropertyGroupDescriptions {
			if propertyGroup != nil {
				mPropertyGroup := map[string]interface{}{
					"property_group_id": aws.StringValue(propertyGroup.PropertyGroupId),
					"property_map":      pointersMapToStringList(propertyGroup.PropertyMap),
				}

				vPropertyGroups = append(vPropertyGroups, mPropertyGroup)
			}
		}

		mEnvironmentProperties["property_group"] = vPropertyGroups

		mApplicationConfiguration["environment_properties"] = []interface{}{mEnvironmentProperties}
	}

	if flinkApplicationConfigurationDescription := applicationConfigurationDescription.FlinkApplicationConfigurationDescription; flinkApplicationConfigurationDescription != nil {
		mFlinkApplicationConfiguration := map[string]interface{}{}

		if checkpointConfigurationDescription := flinkApplicationConfigurationDescription.CheckpointConfigurationDescription; checkpointConfigurationDescription != nil {
			mCheckpointConfiguration := map[string]interface{}{
				"checkpointing_enabled":         aws.BoolValue(checkpointConfigurationDescription.CheckpointingEnabled),
				"checkpoint_interval":           int(aws.Int64Value(checkpointConfigurationDescription.CheckpointInterval)),
				"configuration_type":            aws.StringValue(checkpointConfigurationDescription.ConfigurationType),
				"min_pause_between_checkpoints": int(aws.Int64Value(checkpointConfigurationDescription.MinPauseBetweenCheckpoints)),
			}

			mFlinkApplicationConfiguration["checkpoint_configuration"] = []interface{}{mCheckpointConfiguration}
		}

		if monitoringConfigurationDescription := flinkApplicationConfigurationDescription.MonitoringConfigurationDescription; monitoringConfigurationDescription != nil {
			mMonitoringConfiguration := map[string]interface{}{
				"configuration_type": aws.StringValue(monitoringConfigurationDescription.ConfigurationType),
				"log_level":          aws.StringValue(monitoringConfigurationDescription.LogLevel),
				"metrics_level":      aws.StringValue(monitoringConfigurationDescription.MetricsLevel),
			}

			mFlinkApplicationConfiguration["monitoring_configuration"] = []interface{}{mMonitoringConfiguration}
		}

		if parallelismConfigurationDescription := flinkApplicationConfigurationDescription.ParallelismConfigurationDescription; parallelismConfigurationDescription != nil {
			mParallelismConfiguration := map[string]interface{}{
				"auto_scaling_enabled": aws.BoolValue(parallelismConfigurationDescription.AutoScalingEnabled),
				"configuration_type":   aws.StringValue(parallelismConfigurationDescription.ConfigurationType),
				"parallelism":          int(aws.Int64Value(parallelismConfigurationDescription.Parallelism)),
				"parallelism_per_kpu":  int(aws.Int64Value(parallelismConfigurationDescription.ParallelismPerKPU)),
			}

			mFlinkApplicationConfiguration["parallelism_configuration"] = []interface{}{mParallelismConfiguration}
		}

		mApplicationConfiguration["flink_application_configuration"] = []interface{}{mFlinkApplicationConfiguration}
	}

	if sqlApplicationConfigurationDescription := applicationConfigurationDescription.SqlApplicationConfigurationDescription; sqlApplicationConfigurationDescription != nil {
		mSqlApplicationConfiguration := map[string]interface{}{}

		if inputDescriptions := sqlApplicationConfigurationDescription.InputDescriptions; len(inputDescriptions) > 0 && inputDescriptions[0] != nil {
			inputDescription := inputDescriptions[0]

			mInput := map[string]interface{}{
				"in_app_stream_names": flattenStringList(inputDescription.InAppStreamNames),
				"input_id":            aws.StringValue(inputDescription.InputId),
				"name_prefix":         aws.StringValue(inputDescription.NamePrefix),
			}

			if inputParallelism := inputDescription.InputParallelism; inputParallelism != nil {
				mInputParallelism := map[string]interface{}{
					"count": int(aws.Int64Value(inputParallelism.Count)),
				}

				mInput["input_parallelism"] = []interface{}{mInputParallelism}
			}

			if inputSchema := inputDescription.InputSchema; inputSchema != nil {
				mInput["input_schema"] = flattenKinesisAnalyticsV2SourceSchema(inputSchema)
			}

			if inputProcessingConfigurationDescription := inputDescription.InputProcessingConfigurationDescription; inputProcessingConfigurationDescription != nil {
				mInputProcessingConfiguration := map[string]interface{}{}

				if inputLambdaProcessorDescription := inputProcessingConfigurationDescription.InputLambdaProcessorDescription; inputLambdaProcessorDescription != nil {
					mInputLambdaProcessor := map[string]interface{}{
						"resource_arn": aws.StringValue(inputLambdaProcessorDescription.ResourceARN),
					}

					mInputProcessingConfiguration["input_lambda_processor"] = []interface{}{mInputLambdaProcessor}
				}

				mInput["input_processing_configuration"] = []interface{}{mInputProcessingConfiguration}
			}

			if kinesisFirehoseInputDescription := inputDescription.KinesisFirehoseInputDescription; kinesisFirehoseInputDescription != nil {
				mKinesisFirehoseInput := map[string]interface{}{
					"resource_arn": aws.StringValue(kinesisFirehoseInputDescription.ResourceARN),
				}

				mInput["kinesis_firehose_input"] = []interface{}{mKinesisFirehoseInput}
			}

			if kinesisStreamsInputDescription := inputDescription.KinesisStreamsInputDescription; kinesisStreamsInputDescription != nil {
				mKinesisStreamsInput := map[string]interface{}{
					"resource_arn": aws.StringValue(kinesisStreamsInputDescription.ResourceARN),
				}

				mInput["kinesis_streams_input"] = []interface{}{mKinesisStreamsInput}
			}

			mSqlApplicationConfiguration["input"] = []interface{}{mInput}
		}

		if outputDescriptions := sqlApplicationConfigurationDescription.OutputDescriptions; len(outputDescriptions) > 0 {
			vOutputs := []interface{}{}

			for _, outputDescription := range outputDescriptions {
				if outputDescription != nil {
					mOutput := map[string]interface{}{
						"name":      aws.StringValue(outputDescription.Name),
						"output_id": aws.StringValue(outputDescription.OutputId),
					}

					if destinationSchema := outputDescription.DestinationSchema; destinationSchema != nil {
						mDestinationSchema := map[string]interface{}{
							"record_format_type": aws.StringValue(destinationSchema.RecordFormatType),
						}

						mOutput["destination_schema"] = []interface{}{mDestinationSchema}
					}

					if kinesisFirehoseOutputDescription := outputDescription.KinesisFirehoseOutputDescription; kinesisFirehoseOutputDescription != nil {
						mKinesisFirehoseOutput := map[string]interface{}{
							"resource_arn": aws.StringValue(kinesisFirehoseOutputDescription.ResourceARN),
						}

						mOutput["kinesis_firehose_output"] = []interface{}{mKinesisFirehoseOutput}
					}

					if kinesisStreamsOutputDescription := outputDescription.KinesisStreamsOutputDescription; kinesisStreamsOutputDescription != nil {
						mKinesisStreamsOutput := map[string]interface{}{
							"resource_arn": aws.StringValue(kinesisStreamsOutputDescription.ResourceARN),
						}

						mOutput["kinesis_streams_output"] = []interface{}{mKinesisStreamsOutput}
					}

					if lambdaOutputDescription := outputDescription.LambdaOutputDescription; lambdaOutputDescription != nil {
						mLambdaOutput := map[string]interface{}{
							"resource_arn": aws.StringValue(lambdaOutputDescription.ResourceARN),
						}

						mOutput["lambda_output"] = []interface{}{mLambdaOutput}
					}

					vOutputs = append(vOutputs, mOutput)
				}
			}

			mSqlApplicationConfiguration["output"] = vOutputs
		}

		if referenceDataSourceDescriptions := sqlApplicationConfigurationDescription.ReferenceDataSourceDescriptions; len(referenceDataSourceDescriptions) > 0 && referenceDataSourceDescriptions[0] != nil {
			referenceDataSourceDescription := referenceDataSourceDescriptions[0]

			mReferenceDataSource := map[string]interface{}{
				"reference_id": aws.StringValue(referenceDataSourceDescription.ReferenceId),
				"table_name":   aws.StringValue(referenceDataSourceDescription.TableName),
			}

			if referenceSchema := referenceDataSourceDescription.ReferenceSchema; referenceSchema != nil {
				mReferenceDataSource["reference_schema"] = flattenKinesisAnalyticsV2SourceSchema(referenceSchema)
			}

			if s3ReferenceDataSource := referenceDataSourceDescription.S3ReferenceDataSourceDescription; s3ReferenceDataSource != nil {
				mS3ReferenceDataSource := map[string]interface{}{
					"bucket_arn": aws.StringValue(s3ReferenceDataSource.BucketARN),
					"file_key":   aws.StringValue(s3ReferenceDataSource.FileKey),
				}

				mReferenceDataSource["s3_reference_data_source"] = []interface{}{mS3ReferenceDataSource}
			}

			mSqlApplicationConfiguration["reference_data_source"] = []interface{}{mReferenceDataSource}
		}

		mApplicationConfiguration["sql_application_configuration"] = []interface{}{mSqlApplicationConfiguration}
	}

	return []interface{}{mApplicationConfiguration}
}

func expandKinesisAnalyticsV2ApplicationCodeConfigurationUpdate(vApplicationCodeConfiguration []interface{}) *kinesisanalyticsv2.ApplicationCodeConfigurationUpdate {
	if len(vApplicationCodeConfiguration) == 0 || vApplicationCodeConfiguration[0] == nil {
		return nil
	}

	applicationCodeConfigurationUpdate := &kinesisanalyticsv2.ApplicationCodeConfigurationUpdate{}

	mApplicationCodeConfiguration := vApplicationCodeConfiguration[0].(map[string]interface{})

	if vCodeContent, ok := mApplicationCodeConfiguration["code_content"].([]interface{}); ok && len(vCodeContent) > 0 && vCodeContent[0] != nil {
		codeContentUpdate := &kinesisanalyticsv2.CodeContentUpdate{}

		mCodeContent := vCodeContent[0].(map[string]interface{})

		if vS3ContentLocation, ok := mCodeContent["s3_content_location"].([]interface{}); ok && len(vS3ContentLocation) > 0 && vS3ContentLocation[0] != nil {
			s3ContentLocationUpdate := &kinesisanalyticsv2.S3ContentLocationUpdate{}

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
		applicationCodeConfigurationUpdate.CodeContentTypeUpdate = aws.String(vCodeContentType)
	}

	return applicationCodeConfigurationUpdate
}

func expandKinesisAnalyticsV2ApplicationSnapshotConfigurationUpdate(vApplicationSnapshotConfiguration []interface{}) *kinesisanalyticsv2.ApplicationSnapshotConfigurationUpdate {
	if len(vApplicationSnapshotConfiguration) == 0 || vApplicationSnapshotConfiguration[0] == nil {
		return nil
	}

	applicationSnapshotConfigurationUpdate := &kinesisanalyticsv2.ApplicationSnapshotConfigurationUpdate{}

	mApplicationSnapshotConfiguration := vApplicationSnapshotConfiguration[0].(map[string]interface{})

	if vSnapshotsEnabled, ok := mApplicationSnapshotConfiguration["snapshots_enabled"].(bool); ok {
		applicationSnapshotConfigurationUpdate.SnapshotsEnabledUpdate = aws.Bool(vSnapshotsEnabled)
	}

	return applicationSnapshotConfigurationUpdate
}

func expandKinesisAnalyticsV2EnvironmentPropertyUpdates(vEnvironmentProperties []interface{}) *kinesisanalyticsv2.EnvironmentPropertyUpdates {
	if len(vEnvironmentProperties) == 0 || vEnvironmentProperties[0] == nil {
		// Return empty updates to remove all existing property groups.
		return &kinesisanalyticsv2.EnvironmentPropertyUpdates{PropertyGroups: []*kinesisanalyticsv2.PropertyGroup{}}
	}

	environmentPropertyUpdates := &kinesisanalyticsv2.EnvironmentPropertyUpdates{}

	mEnvironmentProperties := vEnvironmentProperties[0].(map[string]interface{})

	if vPropertyGroups, ok := mEnvironmentProperties["property_group"].(*schema.Set); ok && vPropertyGroups.Len() > 0 {
		environmentPropertyUpdates.PropertyGroups = expandKinesisAnalyticsV2PropertyGroups(vPropertyGroups.List())
	}

	return environmentPropertyUpdates
}

func expandKinesisAnalyticsV2Input(vInput []interface{}) *kinesisanalyticsv2.Input {
	if len(vInput) == 0 || vInput[0] == nil {
		return nil
	}

	input := &kinesisanalyticsv2.Input{}

	mInput := vInput[0].(map[string]interface{})

	if vInputParallelism, ok := mInput["input_parallelism"].([]interface{}); ok && len(vInputParallelism) > 0 && vInputParallelism[0] != nil {
		inputParallelism := &kinesisanalyticsv2.InputParallelism{}

		mInputParallelism := vInputParallelism[0].(map[string]interface{})

		if vCount, ok := mInputParallelism["count"].(int); ok {
			inputParallelism.Count = aws.Int64(int64(vCount))
		}

		input.InputParallelism = inputParallelism
	}

	if vInputProcessingConfiguration, ok := mInput["input_processing_configuration"].([]interface{}); ok {
		input.InputProcessingConfiguration = expandKinesisAnalyticsV2InputProcessingConfiguration(vInputProcessingConfiguration)
	}

	if vInputSchema, ok := mInput["input_schema"].([]interface{}); ok {
		input.InputSchema = expandKinesisAnalyticsV2SourceSchema(vInputSchema)
	}

	if vKinesisFirehoseInput, ok := mInput["kinesis_firehose_input"].([]interface{}); ok && len(vKinesisFirehoseInput) > 0 && vKinesisFirehoseInput[0] != nil {
		kinesisFirehoseInput := &kinesisanalyticsv2.KinesisFirehoseInput{}

		mKinesisFirehoseInput := vKinesisFirehoseInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseInput["resource_arn"].(string); ok && vResourceArn != "" {
			kinesisFirehoseInput.ResourceARN = aws.String(vResourceArn)
		}

		input.KinesisFirehoseInput = kinesisFirehoseInput
	}

	if vKinesisStreamsInput, ok := mInput["kinesis_streams_input"].([]interface{}); ok && len(vKinesisStreamsInput) > 0 && vKinesisStreamsInput[0] != nil {
		kinesisStreamsInput := &kinesisanalyticsv2.KinesisStreamsInput{}

		mKinesisStreamsInput := vKinesisStreamsInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsInput["resource_arn"].(string); ok && vResourceArn != "" {
			kinesisStreamsInput.ResourceARN = aws.String(vResourceArn)
		}

		input.KinesisStreamsInput = kinesisStreamsInput
	}

	if vNamePrefix, ok := mInput["name_prefix"].(string); ok && vNamePrefix != "" {
		input.NamePrefix = aws.String(vNamePrefix)
	}

	return input
}

func expandKinesisAnalyticsV2InputProcessingConfiguration(vInputProcessingConfiguration []interface{}) *kinesisanalyticsv2.InputProcessingConfiguration {
	if len(vInputProcessingConfiguration) == 0 || vInputProcessingConfiguration[0] == nil {
		return nil
	}

	inputProcessingConfiguration := &kinesisanalyticsv2.InputProcessingConfiguration{}

	mInputProcessingConfiguration := vInputProcessingConfiguration[0].(map[string]interface{})

	if vInputLambdaProcessor, ok := mInputProcessingConfiguration["input_lambda_processor"].([]interface{}); ok && len(vInputLambdaProcessor) > 0 && vInputLambdaProcessor[0] != nil {
		inputLambdaProcessor := &kinesisanalyticsv2.InputLambdaProcessor{}

		mInputLambdaProcessor := vInputLambdaProcessor[0].(map[string]interface{})

		if vResourceArn, ok := mInputLambdaProcessor["resource_arn"].(string); ok && vResourceArn != "" {
			inputLambdaProcessor.ResourceARN = aws.String(vResourceArn)
		}

		inputProcessingConfiguration.InputLambdaProcessor = inputLambdaProcessor
	}

	return inputProcessingConfiguration
}

func expandKinesisAnalyticsV2Outputs(vOutputs []interface{}) []*kinesisanalyticsv2.Output {
	outputs := []*kinesisanalyticsv2.Output{}

	for _, vOutput := range vOutputs {
		output := &kinesisanalyticsv2.Output{}

		mOutput := vOutput.(map[string]interface{})

		if vDestinationSchema, ok := mOutput["destination_schema"].([]interface{}); ok && len(vDestinationSchema) > 0 && vDestinationSchema[0] != nil {
			destinationSchema := &kinesisanalyticsv2.DestinationSchema{}

			mDestinationSchema := vDestinationSchema[0].(map[string]interface{})

			if vRecordFormatType, ok := mDestinationSchema["record_format_type"].(string); ok && vRecordFormatType != "" {
				destinationSchema.RecordFormatType = aws.String(vRecordFormatType)
			}

			output.DestinationSchema = destinationSchema
		}

		if vKinesisFirehoseOutput, ok := mOutput["kinesis_firehose_output"].([]interface{}); ok && len(vKinesisFirehoseOutput) > 0 && vKinesisFirehoseOutput[0] != nil {
			kinesisFirehoseOutput := &kinesisanalyticsv2.KinesisFirehoseOutput{}

			mKinesisFirehoseOutput := vKinesisFirehoseOutput[0].(map[string]interface{})

			if vResourceArn, ok := mKinesisFirehoseOutput["resource_arn"].(string); ok && vResourceArn != "" {
				kinesisFirehoseOutput.ResourceARN = aws.String(vResourceArn)
			}

			output.KinesisFirehoseOutput = kinesisFirehoseOutput
		}

		if vKinesisStreamsOutput, ok := mOutput["kinesis_streams_output"].([]interface{}); ok && len(vKinesisStreamsOutput) > 0 && vKinesisStreamsOutput[0] != nil {
			kinesisStreamsOutput := &kinesisanalyticsv2.KinesisStreamsOutput{}

			mKinesisStreamsOutput := vKinesisStreamsOutput[0].(map[string]interface{})

			if vResourceArn, ok := mKinesisStreamsOutput["resource_arn"].(string); ok && vResourceArn != "" {
				kinesisStreamsOutput.ResourceARN = aws.String(vResourceArn)
			}

			output.KinesisStreamsOutput = kinesisStreamsOutput
		}

		if vLambdaOutput, ok := mOutput["lambda_output"].([]interface{}); ok && len(vLambdaOutput) > 0 && vLambdaOutput[0] != nil {
			lambdaOutput := &kinesisanalyticsv2.LambdaOutput{}

			mLambdaOutput := vLambdaOutput[0].(map[string]interface{})

			if vResourceArn, ok := mLambdaOutput["resource_arn"].(string); ok && vResourceArn != "" {
				lambdaOutput.ResourceARN = aws.String(vResourceArn)
			}

			output.LambdaOutput = lambdaOutput
		}

		if vName, ok := mOutput["name"].(string); ok && vName != "" {
			output.Name = aws.String(vName)
		}

		outputs = append(outputs, output)
	}

	return outputs
}

func expandKinesisAnalyticsV2PropertyGroups(vPropertyGroups []interface{}) []*kinesisanalyticsv2.PropertyGroup {
	propertyGroups := []*kinesisanalyticsv2.PropertyGroup{}

	for _, vPropertyGroup := range vPropertyGroups {
		propertyGroup := &kinesisanalyticsv2.PropertyGroup{}

		mPropertyGroup := vPropertyGroup.(map[string]interface{})

		if vPropertyGroupID, ok := mPropertyGroup["property_group_id"].(string); ok && vPropertyGroupID != "" {
			propertyGroup.PropertyGroupId = aws.String(vPropertyGroupID)
		} else {
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			continue
		}

		if vPropertyMap, ok := mPropertyGroup["property_map"].(map[string]interface{}); ok && len(vPropertyMap) > 0 {
			propertyGroup.PropertyMap = stringMapToPointers(vPropertyMap)
		}

		propertyGroups = append(propertyGroups, propertyGroup)
	}

	return propertyGroups
}

func expandKinesisAnalyticsV2RecordColumns(vRecordColumns []interface{}) []*kinesisanalyticsv2.RecordColumn {
	recordColumns := []*kinesisanalyticsv2.RecordColumn{}

	for _, vRecordColumn := range vRecordColumns {
		recordColumn := &kinesisanalyticsv2.RecordColumn{}

		mRecordColumn := vRecordColumn.(map[string]interface{})

		if vMapping, ok := mRecordColumn["mapping"].(string); ok && vMapping != "" {
			recordColumn.Mapping = aws.String(vMapping)
		}
		if vName, ok := mRecordColumn["name"].(string); ok && vName != "" {
			recordColumn.Name = aws.String(vName)
		}
		if vSqlType, ok := mRecordColumn["sql_type"].(string); ok && vSqlType != "" {
			recordColumn.SqlType = aws.String(vSqlType)
		}

		recordColumns = append(recordColumns, recordColumn)
	}

	return recordColumns
}

func expandKinesisAnalyticsV2RecordFormat(vRecordFormat []interface{}) *kinesisanalyticsv2.RecordFormat {
	if len(vRecordFormat) == 0 || vRecordFormat[0] == nil {
		return nil
	}

	recordFormat := &kinesisanalyticsv2.RecordFormat{}

	mRecordFormat := vRecordFormat[0].(map[string]interface{})

	if vMappingParameters, ok := mRecordFormat["mapping_parameters"].([]interface{}); ok && len(vMappingParameters) > 0 && vMappingParameters[0] != nil {
		mappingParameters := &kinesisanalyticsv2.MappingParameters{}

		mMappingParameters := vMappingParameters[0].(map[string]interface{})

		if vCsvMappingParameters, ok := mMappingParameters["csv_mapping_parameters"].([]interface{}); ok && len(vCsvMappingParameters) > 0 && vCsvMappingParameters[0] != nil {
			csvMappingParameters := &kinesisanalyticsv2.CSVMappingParameters{}

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
			jsonMappingParameters := &kinesisanalyticsv2.JSONMappingParameters{}

			mJsonMappingParameters := vJsonMappingParameters[0].(map[string]interface{})

			if vRecordRowPath, ok := mJsonMappingParameters["record_row_path"].(string); ok && vRecordRowPath != "" {
				jsonMappingParameters.RecordRowPath = aws.String(vRecordRowPath)
			}

			mappingParameters.JSONMappingParameters = jsonMappingParameters
		}

		recordFormat.MappingParameters = mappingParameters
	}

	if vRecordFormatType, ok := mRecordFormat["record_format_type"].(string); ok && vRecordFormatType != "" {
		recordFormat.RecordFormatType = aws.String(vRecordFormatType)
	}

	return recordFormat
}

func expandKinesisAnalyticsV2ReferenceDataSource(vReferenceDataSource []interface{}) *kinesisanalyticsv2.ReferenceDataSource {
	if len(vReferenceDataSource) == 0 || vReferenceDataSource[0] == nil {
		return nil
	}

	referenceDataSource := &kinesisanalyticsv2.ReferenceDataSource{}

	mReferenceDataSource := vReferenceDataSource[0].(map[string]interface{})

	if vReferenceSchema, ok := mReferenceDataSource["reference_schema"].([]interface{}); ok {
		referenceDataSource.ReferenceSchema = expandKinesisAnalyticsV2SourceSchema(vReferenceSchema)
	}

	if vS3ReferenceDataSource, ok := mReferenceDataSource["s3_reference_data_source"].([]interface{}); ok && len(vS3ReferenceDataSource) > 0 && vS3ReferenceDataSource[0] != nil {
		s3ReferenceDataSource := &kinesisanalyticsv2.S3ReferenceDataSource{}

		mS3ReferenceDataSource := vS3ReferenceDataSource[0].(map[string]interface{})

		if vBucketArn, ok := mS3ReferenceDataSource["bucket_arn"].(string); ok && vBucketArn != "" {
			s3ReferenceDataSource.BucketARN = aws.String(vBucketArn)
		}
		if vFileKey, ok := mS3ReferenceDataSource["file_key"].(string); ok && vFileKey != "" {
			s3ReferenceDataSource.FileKey = aws.String(vFileKey)
		}

		referenceDataSource.S3ReferenceDataSource = s3ReferenceDataSource
	}

	if vTableName, ok := mReferenceDataSource["table_name"].(string); ok && vTableName != "" {
		referenceDataSource.TableName = aws.String(vTableName)
	}

	return referenceDataSource
}

func expandKinesisAnalyticsV2SourceSchema(vSourceSchema []interface{}) *kinesisanalyticsv2.SourceSchema {
	if len(vSourceSchema) == 0 || vSourceSchema[0] == nil {
		return nil
	}

	sourceSchema := &kinesisanalyticsv2.SourceSchema{}

	mSourceSchema := vSourceSchema[0].(map[string]interface{})

	if vRecordColumns, ok := mSourceSchema["record_column"].(*schema.Set); ok && vRecordColumns.Len() > 0 {
		sourceSchema.RecordColumns = expandKinesisAnalyticsV2RecordColumns(vRecordColumns.List())
	}

	if vRecordEncoding, ok := mSourceSchema["record_encoding"].(string); ok && vRecordEncoding != "" {
		sourceSchema.RecordEncoding = aws.String(vRecordEncoding)
	}

	if vRecordFormat, ok := mSourceSchema["record_format"].([]interface{}); ok && len(vRecordFormat) > 0 && vRecordFormat[0] != nil {
		sourceSchema.RecordFormat = expandKinesisAnalyticsV2RecordFormat(vRecordFormat)
	}

	return sourceSchema
}

func flattenKinesisAnalyticsV2SourceSchema(sourceSchema *kinesisanalyticsv2.SourceSchema) []interface{} {
	if sourceSchema == nil {
		return []interface{}{}
	}

	mSourceSchema := map[string]interface{}{
		"record_encoding": aws.StringValue(sourceSchema.RecordEncoding),
	}

	if len(sourceSchema.RecordColumns) > 0 {
		vRecordColumns := []interface{}{}

		for _, recordColumn := range sourceSchema.RecordColumns {
			if recordColumn != nil {
				mRecordColumn := map[string]interface{}{
					"mapping":  aws.StringValue(recordColumn.Mapping),
					"name":     aws.StringValue(recordColumn.Name),
					"sql_type": aws.StringValue(recordColumn.SqlType),
				}

				vRecordColumns = append(vRecordColumns, mRecordColumn)
			}
		}

		mSourceSchema["record_column"] = vRecordColumns
	}

	if recordFormat := sourceSchema.RecordFormat; recordFormat != nil {
		mRecordFormat := map[string]interface{}{
			"record_format_type": aws.StringValue(recordFormat.RecordFormatType),
		}

		if mappingParameters := recordFormat.MappingParameters; mappingParameters != nil {
			mMappingParameters := map[string]interface{}{}

			if csvMappingParameters := mappingParameters.CSVMappingParameters; csvMappingParameters != nil {
				mCsvMappingParameters := map[string]interface{}{
					"record_column_delimiter": aws.StringValue(csvMappingParameters.RecordColumnDelimiter),
					"record_row_delimiter":    aws.StringValue(csvMappingParameters.RecordRowDelimiter),
				}

				mMappingParameters["csv_mapping_parameters"] = []interface{}{mCsvMappingParameters}
			}

			if jsonMappingParameters := mappingParameters.JSONMappingParameters; jsonMappingParameters != nil {
				mJsonMappingParameters := map[string]interface{}{
					"record_row_path": aws.StringValue(jsonMappingParameters.RecordRowPath),
				}

				mMappingParameters["json_mapping_parameters"] = []interface{}{mJsonMappingParameters}
			}

			mRecordFormat["mapping_parameters"] = []interface{}{mMappingParameters}
		}

		mSourceSchema["record_format"] = []interface{}{mRecordFormat}
	}

	return []interface{}{mSourceSchema}
}

func expandKinesisAnalyticsV2ApplicationFlinkApplicationConfigurationUpdate(vFlinkApplicationConfiguration []interface{}) *kinesisanalyticsv2.FlinkApplicationConfigurationUpdate {
	if len(vFlinkApplicationConfiguration) == 0 || vFlinkApplicationConfiguration[0] == nil {
		return nil
	}

	flinkApplicationConfigurationUpdate := &kinesisanalyticsv2.FlinkApplicationConfigurationUpdate{}

	mFlinkApplicationConfiguration := vFlinkApplicationConfiguration[0].(map[string]interface{})

	if vCheckpointConfiguration, ok := mFlinkApplicationConfiguration["checkpoint_configuration"].([]interface{}); ok && len(vCheckpointConfiguration) > 0 && vCheckpointConfiguration[0] != nil {
		checkpointConfigurationUpdate := &kinesisanalyticsv2.CheckpointConfigurationUpdate{}

		mCheckpointConfiguration := vCheckpointConfiguration[0].(map[string]interface{})

		if vConfigurationType, ok := mCheckpointConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
			checkpointConfigurationUpdate.ConfigurationTypeUpdate = aws.String(vConfigurationType)

			if vConfigurationType == kinesisanalyticsv2.ConfigurationTypeCustom {
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
		monitoringConfigurationUpdate := &kinesisanalyticsv2.MonitoringConfigurationUpdate{}

		mMonitoringConfiguration := vMonitoringConfiguration[0].(map[string]interface{})

		if vConfigurationType, ok := mMonitoringConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
			monitoringConfigurationUpdate.ConfigurationTypeUpdate = aws.String(vConfigurationType)

			if vConfigurationType == kinesisanalyticsv2.ConfigurationTypeCustom {
				if vLogLevel, ok := mMonitoringConfiguration["log_level"].(string); ok && vLogLevel != "" {
					monitoringConfigurationUpdate.LogLevelUpdate = aws.String(vLogLevel)
				}
				if vMetricsLevel, ok := mMonitoringConfiguration["metrics_level"].(string); ok && vMetricsLevel != "" {
					monitoringConfigurationUpdate.MetricsLevelUpdate = aws.String(vMetricsLevel)
				}
			}
		}

		flinkApplicationConfigurationUpdate.MonitoringConfigurationUpdate = monitoringConfigurationUpdate
	}

	if vParallelismConfiguration, ok := mFlinkApplicationConfiguration["parallelism_configuration"].([]interface{}); ok && len(vParallelismConfiguration) > 0 && vParallelismConfiguration[0] != nil {
		parallelismConfigurationUpdate := &kinesisanalyticsv2.ParallelismConfigurationUpdate{}

		mParallelismConfiguration := vParallelismConfiguration[0].(map[string]interface{})

		if vConfigurationType, ok := mParallelismConfiguration["configuration_type"].(string); ok && vConfigurationType != "" {
			parallelismConfigurationUpdate.ConfigurationTypeUpdate = aws.String(vConfigurationType)

			if vConfigurationType == kinesisanalyticsv2.ConfigurationTypeCustom {
				if vAutoScalingEnabled, ok := mParallelismConfiguration["auto_scaling_enabled"].(bool); ok {
					parallelismConfigurationUpdate.AutoScalingEnabledUpdate = aws.Bool(vAutoScalingEnabled)
				}
				if vParallelism, ok := mParallelismConfiguration["parallelism"].(int); ok {
					parallelismConfigurationUpdate.ParallelismUpdate = aws.Int64(int64(vParallelism))
				}
				if vParallelismPerKPU, ok := mParallelismConfiguration["parallelism_per_kpu"].(int); ok {
					parallelismConfigurationUpdate.ParallelismPerKPUUpdate = aws.Int64(int64(vParallelismPerKPU))
				}
			}
		}

		flinkApplicationConfigurationUpdate.ParallelismConfigurationUpdate = parallelismConfigurationUpdate
	}

	return flinkApplicationConfigurationUpdate
}

func expandKinesisAnalyticsV2InputUpdate(vInput []interface{}) *kinesisanalyticsv2.InputUpdate {
	if len(vInput) == 0 || vInput[0] == nil {
		return nil
	}

	inputUpdate := &kinesisanalyticsv2.InputUpdate{}

	mInput := vInput[0].(map[string]interface{})

	if vInputId, ok := mInput["input_id"].(string); ok && vInputId != "" {
		inputUpdate.InputId = aws.String(vInputId)
	}

	if vInputParallelism, ok := mInput["input_parallelism"].([]interface{}); ok && len(vInputParallelism) > 0 && vInputParallelism[0] != nil {
		inputParallelismUpdate := &kinesisanalyticsv2.InputParallelismUpdate{}

		mInputParallelism := vInputParallelism[0].(map[string]interface{})

		if vCount, ok := mInputParallelism["count"].(int); ok {
			inputParallelismUpdate.CountUpdate = aws.Int64(int64(vCount))
		}

		inputUpdate.InputParallelismUpdate = inputParallelismUpdate
	}

	if vInputProcessingConfiguration, ok := mInput["input_processing_configuration"].([]interface{}); ok && len(vInputProcessingConfiguration) > 0 && vInputProcessingConfiguration[0] != nil {
		inputProcessingConfigurationUpdate := &kinesisanalyticsv2.InputProcessingConfigurationUpdate{}

		mInputProcessingConfiguration := vInputProcessingConfiguration[0].(map[string]interface{})

		if vInputLambdaProcessor, ok := mInputProcessingConfiguration["input_lambda_processor"].([]interface{}); ok && len(vInputLambdaProcessor) > 0 && vInputLambdaProcessor[0] != nil {
			inputLambdaProcessorUpdate := &kinesisanalyticsv2.InputLambdaProcessorUpdate{}

			mInputLambdaProcessor := vInputLambdaProcessor[0].(map[string]interface{})

			if vResourceArn, ok := mInputLambdaProcessor["resource_arn"].(string); ok && vResourceArn != "" {
				inputLambdaProcessorUpdate.ResourceARNUpdate = aws.String(vResourceArn)
			}

			inputProcessingConfigurationUpdate.InputLambdaProcessorUpdate = inputLambdaProcessorUpdate
		}

		inputUpdate.InputProcessingConfigurationUpdate = inputProcessingConfigurationUpdate
	}

	if vInputSchema, ok := mInput["input_schema"].([]interface{}); ok {
		inputUpdate.InputSchemaUpdate = expandKinesisAnalyticsV2InputSchemaUpdate(vInputSchema)
	}

	if vKinesisFirehoseInput, ok := mInput["kinesis_firehose_input"].([]interface{}); ok && len(vKinesisFirehoseInput) > 0 && vKinesisFirehoseInput[0] != nil {
		kinesisFirehoseInputUpdate := &kinesisanalyticsv2.KinesisFirehoseInputUpdate{}

		mKinesisFirehoseInput := vKinesisFirehoseInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseInput["resource_arn"].(string); ok && vResourceArn != "" {
			kinesisFirehoseInputUpdate.ResourceARNUpdate = aws.String(vResourceArn)
		}

		inputUpdate.KinesisFirehoseInputUpdate = kinesisFirehoseInputUpdate
	}

	if vKinesisStreamsInput, ok := mInput["kinesis_streams_input"].([]interface{}); ok && len(vKinesisStreamsInput) > 0 && vKinesisStreamsInput[0] != nil {
		kinesisStreamsInputUpdate := &kinesisanalyticsv2.KinesisStreamsInputUpdate{}

		mKinesisStreamsInput := vKinesisStreamsInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsInput["resource_arn"].(string); ok && vResourceArn != "" {
			kinesisStreamsInputUpdate.ResourceARNUpdate = aws.String(vResourceArn)
		}

		inputUpdate.KinesisStreamsInputUpdate = kinesisStreamsInputUpdate
	}

	if vNamePrefix, ok := mInput["name_prefix"].(string); ok && vNamePrefix != "" {
		inputUpdate.NamePrefixUpdate = aws.String(vNamePrefix)
	}

	return inputUpdate
}

func expandKinesisAnalyticsV2InputSchemaUpdate(vInputSchema []interface{}) *kinesisanalyticsv2.InputSchemaUpdate {
	if len(vInputSchema) == 0 || vInputSchema[0] == nil {
		return nil
	}

	inputSchemaUpdate := &kinesisanalyticsv2.InputSchemaUpdate{}

	mInputSchema := vInputSchema[0].(map[string]interface{})

	if vRecordColumns, ok := mInputSchema["record_column"].(*schema.Set); ok && vRecordColumns.Len() > 0 {
		inputSchemaUpdate.RecordColumnUpdates = expandKinesisAnalyticsV2RecordColumns(vRecordColumns.List())
	}

	if vRecordEncoding, ok := mInputSchema["record_encoding"].(string); ok && vRecordEncoding != "" {
		inputSchemaUpdate.RecordEncodingUpdate = aws.String(vRecordEncoding)
	}

	if vRecordFormat, ok := mInputSchema["record_format"].([]interface{}); ok {
		inputSchemaUpdate.RecordFormatUpdate = expandKinesisAnalyticsV2RecordFormat(vRecordFormat)
	}

	return inputSchemaUpdate
}

func expandKinesisAnalyticsV2OutputUpdate(vOutput []interface{}) *kinesisanalyticsv2.OutputUpdate {
	if len(vOutput) == 0 || vOutput[0] == nil {
		return nil
	}

	outputUpdate := &kinesisanalyticsv2.OutputUpdate{}

	return outputUpdate
}

func expandKinesisAnalyticsV2ReferenceDataSourceUpdate(vReferenceDataSource []interface{}) *kinesisanalyticsv2.ReferenceDataSourceUpdate {
	if len(vReferenceDataSource) == 0 || vReferenceDataSource[0] == nil {
		return nil
	}

	referenceDataSourceUpdate := &kinesisanalyticsv2.ReferenceDataSourceUpdate{}

	return referenceDataSourceUpdate
}

func expandKinesisAnalyticsV2CloudWatchLoggingOptions(vCloudWatchLoggingOptions []interface{}) []*kinesisanalyticsv2.CloudWatchLoggingOption {
	if len(vCloudWatchLoggingOptions) == 0 || vCloudWatchLoggingOptions[0] == nil {
		return nil
	}

	cloudWatchLoggingOption := &kinesisanalyticsv2.CloudWatchLoggingOption{}

	mCloudWatchLoggingOption := vCloudWatchLoggingOptions[0].(map[string]interface{})

	if vLogStreamArn, ok := mCloudWatchLoggingOption["log_stream_arn"].(string); ok && vLogStreamArn != "" {
		cloudWatchLoggingOption.LogStreamARN = aws.String(vLogStreamArn)
	}

	return []*kinesisanalyticsv2.CloudWatchLoggingOption{cloudWatchLoggingOption}
}

func flattenKinesisAnalyticsV2CloudWatchLoggingOptionDescriptions(cloudWatchLoggingOptionDescriptions []*kinesisanalyticsv2.CloudWatchLoggingOptionDescription) []interface{} {
	if len(cloudWatchLoggingOptionDescriptions) == 0 || cloudWatchLoggingOptionDescriptions[0] == nil {
		return []interface{}{}
	}

	cloudWatchLoggingOptionDescription := cloudWatchLoggingOptionDescriptions[0]

	mCloudWatchLoggingOption := map[string]interface{}{
		"cloudwatch_logging_option_id": aws.StringValue(cloudWatchLoggingOptionDescription.CloudWatchLoggingOptionId),
		"log_stream_arn":               aws.StringValue(cloudWatchLoggingOptionDescription.LogStreamARN),
	}

	return []interface{}{mCloudWatchLoggingOption}
}

// kinesisAnalyticsV2RetryIAMEventualConsistency retries the specified function for 1 minute
// if the returned error indicates an IAM eventual consistency issue. If the retries time out
// the specified function is called one more time.
func kinesisAnalyticsV2RetryIAMEventualConsistency(f func() (interface{}, error)) (interface{}, error) {
	var output interface{}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		output, err = f()

		// Kinesis Stream: https://github.com/terraform-providers/terraform-provider-aws/issues/7032
		if isAWSErr(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
			return resource.RetryableError(err)
		}

		// Kinesis Firehose: https://github.com/terraform-providers/terraform-provider-aws/issues/7394
		if isAWSErr(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics doesn't have sufficient privileges") {
			return resource.RetryableError(err)
		}

		// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
		if isAWSErr(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = f()
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// TODO Delete below...
/*
func resourceAwsKinesisAnalyticsV2ApplicationUpdateOld(d *schema.ResourceData, meta interface{}) error {
	var version int
	conn := meta.(*AWSClient).kinesisanalyticsv2conn
	name := d.Get("name").(string)

	if v, ok := d.GetOk("version_id"); ok {
		version = v.(int)
	} else {
		version = 1
	}

	updateApplicationOpts := &kinesisanalyticsv2.UpdateApplicationInput{
		ApplicationName:             aws.String(name),
		CurrentApplicationVersionId: aws.Int64(int64(version)),
	}

	applicationUpdate := createApplicationV2UpdateOpts(d)

	if !reflect.DeepEqual(applicationUpdate, &kinesisanalyticsv2.UpdateApplicationInput{}) {
		updateApplicationOpts.SetApplicationConfigurationUpdate(applicationUpdate.ApplicationConfigurationUpdate)
		updateApplicationOpts.SetCloudWatchLoggingOptionUpdates(applicationUpdate.CloudWatchLoggingOptionUpdates)
		_, updateErr := conn.UpdateApplication(updateApplicationOpts)
		if updateErr != nil {
			return updateErr
		}
		version = version + 1
	}

	oldLoggingOptions, newLoggingOptions := d.GetChange("cloudwatch_logging_options")
	if len(oldLoggingOptions.([]interface{})) == 0 && len(newLoggingOptions.([]interface{})) > 0 {
		if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
			clo := v.([]interface{})[0].(map[string]interface{})
			cloudwatchLoggingOption := expandKinesisAnalyticsV2CloudwatchLoggingOptionOld(clo)
			addOpts := &kinesisanalyticsv2.AddApplicationCloudWatchLoggingOptionInput{
				ApplicationName:             aws.String(name),
				CurrentApplicationVersionId: aws.Int64(int64(version)),
				CloudWatchLoggingOption:     cloudwatchLoggingOption,
			}
			// Retry for IAM eventual consistency
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := conn.AddApplicationCloudWatchLoggingOption(addOpts)
				if err != nil {
					if isAWSErr(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if isResourceTimeoutError(err) {
				_, err = conn.AddApplicationCloudWatchLoggingOption(addOpts)
			}

			if err != nil {
				return fmt.Errorf("Unable to add CloudWatch logging options: %s", err)
			}
			version = version + 1
		}
	}
	if d.HasChange("application_configuration.0.sql_application_configuration") {
		oldConf, newConf := d.GetChange("application_configuration.0.sql_application_configuration")
		o := oldConf.([]interface{})[0].(map[string]interface{})
		n := newConf.([]interface{})[0].(map[string]interface{})
		oldInputs := o["input"].([]interface{})
		oldOutputs := o["output"].(*schema.Set).List()
		newInputs := n["input"].([]interface{})
		newOutputs := n["output"].(*schema.Set).List()

		if len(oldInputs) == 0 && len(newInputs) > 0 {
			i := newInputs[0].(map[string]interface{})
			input := expandKinesisAnalyticsV2Input(i)
			addOpts := &kinesisanalyticsv2.AddApplicationInputInput{
				ApplicationName:             aws.String(name),
				CurrentApplicationVersionId: aws.Int64(int64(version)),
				Input:                       input,
			}
			// Retry for IAM eventual consistency
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := conn.AddApplicationInput(addOpts)
				if err != nil {
					if isAWSErr(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
						return resource.RetryableError(err)
					}
					// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
					if isAWSErr(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if isResourceTimeoutError(err) {
				_, err = conn.AddApplicationInput(addOpts)
			}

			if err != nil {
				return fmt.Errorf("Unable to add application inputs: %s", err)
			}
			version = version + 1
		}
		if len(oldOutputs) == 0 && len(newOutputs) > 0 {
			o := newOutputs[0].(map[string]interface{})
			output := expandKinesisAnalyticsV2Output(o)
			addOpts := &kinesisanalyticsv2.AddApplicationOutputInput{
				ApplicationName:             aws.String(name),
				CurrentApplicationVersionId: aws.Int64(int64(version)),
				Output:                      output,
			}
			// Retry for IAM eventual consistency
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := conn.AddApplicationOutput(addOpts)
				if err != nil {
					if isAWSErr(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
						return resource.RetryableError(err)
					}
					// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
					if isAWSErr(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if isResourceTimeoutError(err) {
				_, err = conn.AddApplicationOutput(addOpts)
			}
			if err != nil {
				return fmt.Errorf("Unable to add application outputs: %s", err)
			}
		}
	}
	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Kinesisanalyticsv2UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Kinesis Analytics Application (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsKinesisAnalyticsV2ApplicationRead(d, meta)
}

func expandKinesisAnalyticsV2ApplicationConfigurationOld(runtime string, conf []interface{}) *kinesisanalyticsv2.ApplicationConfiguration {
	if len(conf) == 0 {
		return nil
	}
	c := conf[0].(map[string]interface{})

	var sqlApplicationConfiguration *kinesisanalyticsv2.SqlApplicationConfiguration
	var flinkApplicationConfiguration *kinesisanalyticsv2.FlinkApplicationConfiguration
	switch {
	case strings.HasPrefix(runtime, "SQL"):
		sqlConfig, ok := c["sql_application_configuration"]
		if !ok {
			break
		}
		sc := sqlConfig.([]interface{})[0]
		if sc != nil {
			sqlApplicationConfiguration = expandKinesisAnalayticsV2SqlApplicationConfiguration(sc.(map[string]interface{}))
		}

	case runtimeIsFlink(runtime):
		flinkConfig, ok := c["flink_application_configuration"]
		if !ok {
			break
		}
		fc := flinkConfig.([]interface{})[0]
		if fc != nil {
			flinkApplicationConfiguration = expandKinesisAnalyticsV2FlinkApplicationConfiguration(fc.(map[string]interface{}))
		}
	}
	var snapshotConfig *kinesisanalyticsv2.ApplicationSnapshotConfiguration
	if v, ok := c["application_snapshot_configuration"]; ok {
		snapshotConfig = expandKinesisAnalyticsV2ApplicationSnapshotConfiguration(v.(*schema.Set))
	}

	var codeConfig *kinesisanalyticsv2.ApplicationCodeConfiguration
	if v, ok := c["application_code_configuration"]; ok {
		codeConfig = expandKinesisAnalyticsV2ApplicationCodeConfiguration(v.([]interface{}))
	}

	var environmentProperties *kinesisanalyticsv2.EnvironmentProperties
	if v, ok := c["environment_properties"]; ok {
		var propertyGroups []*kinesisanalyticsv2.PropertyGroup
		lst := v.([]interface{})
		if len(lst) > 0 {
			m := lst[0].(map[string]interface{})
			propertyGroups = expandPropertyGroups(m["property_group"].(*schema.Set).List())
			environmentProperties = &kinesisanalyticsv2.EnvironmentProperties{
				PropertyGroups: propertyGroups,
			}
		}
	}
	return &kinesisanalyticsv2.ApplicationConfiguration{
		SqlApplicationConfiguration:      sqlApplicationConfiguration,
		FlinkApplicationConfiguration:    flinkApplicationConfiguration,
		EnvironmentProperties:            environmentProperties,
		ApplicationSnapshotConfiguration: snapshotConfig,
		ApplicationCodeConfiguration:     codeConfig,
	}
}

func expandKinesisAnalyticsV2CloudwatchLoggingOptionOld(clo map[string]interface{}) *kinesisanalyticsv2.CloudWatchLoggingOption {
	cloudwatchLoggingOption := &kinesisanalyticsv2.CloudWatchLoggingOption{
		LogStreamARN: aws.String(clo["log_stream_arn"].(string)),
	}
	return cloudwatchLoggingOption
}

func expandPropertyGroups(i []interface{}) []*kinesisanalyticsv2.PropertyGroup {
	propertyGroups := []*kinesisanalyticsv2.PropertyGroup{}
	for _, v := range i {
		pg := v.(map[string]interface{})
		id := pg["property_group_id"].(string)
		pm := pg["property_map"].(map[string]interface{})

		awsMap := make(map[string]*string, len(pm))
		for k, v := range pm {
			awsMap[k] = aws.String(v.(string))
		}

		propertyGroups = append(propertyGroups, &kinesisanalyticsv2.PropertyGroup{
			PropertyGroupId: aws.String(id),
			PropertyMap:     awsMap,
		})
	}
	return propertyGroups
}

func expandKinesisAnalyticsV2Input(i map[string]interface{}) *kinesisanalyticsv2.Input {
	input := &kinesisanalyticsv2.Input{
		NamePrefix: aws.String(i["name_prefix"].(string)),
	}

	if v := i["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfi := &kinesisanalyticsv2.KinesisFirehoseInput{
			ResourceARN: aws.String(kf["resource_arn"].(string)),
		}
		input.KinesisFirehoseInput = kfi
	}

	if v := i["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		ksi := &kinesisanalyticsv2.KinesisStreamsInput{
			ResourceARN: aws.String(ks["resource_arn"].(string)),
		}
		input.KinesisStreamsInput = ksi
	}

	if v := i["parallelism"].([]interface{}); len(v) > 0 {
		p := v[0].(map[string]interface{})

		if c, ok := p["count"]; ok {
			ip := &kinesisanalyticsv2.InputParallelism{
				Count: aws.Int64(int64(c.(int))),
			}
			input.InputParallelism = ip
		}
	}

	if v := i["processing_configuration"].([]interface{}); len(v) > 0 {
		pc := v[0].(map[string]interface{})

		if l := pc["lambda"].([]interface{}); len(l) > 0 {
			lp := l[0].(map[string]interface{})
			ipc := &kinesisanalyticsv2.InputProcessingConfiguration{
				InputLambdaProcessor: &kinesisanalyticsv2.InputLambdaProcessor{
					ResourceARN: aws.String(lp["resource_arn"].(string)),
				},
			}
			input.InputProcessingConfiguration = ipc
		}
	}

	if v := i["schema"].([]interface{}); len(v) > 0 {
		vL := v[0].(map[string]interface{})
		ss := expandKinesisAnalyticsV2SourceSchemaOld(vL)
		input.InputSchema = ss
	}

	return input
}

func expandKinesisAnalyticsV2Output(o map[string]interface{}) *kinesisanalyticsv2.Output {
	output := &kinesisanalyticsv2.Output{
		Name: aws.String(o["name"].(string)),
	}

	if v := o["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfo := &kinesisanalyticsv2.KinesisFirehoseOutput{
			ResourceARN: aws.String(kf["resource_arn"].(string)),
		}
		output.KinesisFirehoseOutput = kfo
	}

	if v := o["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		kso := &kinesisanalyticsv2.KinesisStreamsOutput{
			ResourceARN: aws.String(ks["resource_arn"].(string)),
		}
		output.KinesisStreamsOutput = kso
	}

	if v := o["lambda"].([]interface{}); len(v) > 0 {
		l := v[0].(map[string]interface{})
		lo := &kinesisanalyticsv2.LambdaOutput{
			ResourceARN: aws.String(l["resource_arn"].(string)),
		}
		output.LambdaOutput = lo
	}

	if v := o["schema"].([]interface{}); len(v) > 0 {
		ds := v[0].(map[string]interface{})
		dso := &kinesisanalyticsv2.DestinationSchema{
			RecordFormatType: aws.String(ds["record_format_type"].(string)),
		}
		output.DestinationSchema = dso
	}

	return output
}

func expandKinesisAnalyticsV2SourceSchemaOld(vL map[string]interface{}) *kinesisanalyticsv2.SourceSchema {
	ss := &kinesisanalyticsv2.SourceSchema{}
	if v := vL["record_column"].(*schema.Set).List(); len(v) > 0 {
		var rcs []*kinesisanalyticsv2.RecordColumn

		for _, rc := range v {
			rcD := rc.(map[string]interface{})
			rc := &kinesisanalyticsv2.RecordColumn{
				Name:    aws.String(rcD["name"].(string)),
				SqlType: aws.String(rcD["sql_type"].(string)),
			}

			if v, ok := rcD["mapping"]; ok {
				rc.Mapping = aws.String(v.(string))
			}

			rcs = append(rcs, rc)
		}

		ss.RecordColumns = rcs
	}

	if v, ok := vL["record_encoding"]; ok && v.(string) != "" {
		ss.RecordEncoding = aws.String(v.(string))
	}

	if v := vL["record_format"].([]interface{}); len(v) > 0 {
		vL := v[0].(map[string]interface{})
		rf := &kinesisanalyticsv2.RecordFormat{}

		if v := vL["mapping_parameters"].([]interface{}); len(v) > 0 {
			vL := v[0].(map[string]interface{})
			mp := &kinesisanalyticsv2.MappingParameters{}

			if v := vL["csv"].([]interface{}); len(v) > 0 {
				cL := v[0].(map[string]interface{})
				cmp := &kinesisanalyticsv2.CSVMappingParameters{
					RecordColumnDelimiter: aws.String(cL["record_column_delimiter"].(string)),
					RecordRowDelimiter:    aws.String(cL["record_row_delimiter"].(string)),
				}
				mp.CSVMappingParameters = cmp
				rf.RecordFormatType = aws.String("CSV")
			}

			if v := vL["json"].([]interface{}); len(v) > 0 {
				jL := v[0].(map[string]interface{})
				jmp := &kinesisanalyticsv2.JSONMappingParameters{
					RecordRowPath: aws.String(jL["record_row_path"].(string)),
				}
				mp.JSONMappingParameters = jmp
				rf.RecordFormatType = aws.String("JSON")
			}
			rf.MappingParameters = mp
		}

		ss.RecordFormat = rf
	}
	return ss
}

func expandKinesisAnalyticsV2ReferenceData(rd map[string]interface{}) *kinesisanalyticsv2.ReferenceDataSource {
	referenceData := &kinesisanalyticsv2.ReferenceDataSource{
		TableName: aws.String(rd["table_name"].(string)),
	}

	if v := rd["s3"].([]interface{}); len(v) > 0 {
		s3 := v[0].(map[string]interface{})
		s3rds := &kinesisanalyticsv2.S3ReferenceDataSource{
			BucketARN: aws.String(s3["bucket_arn"].(string)),
			FileKey:   aws.String(s3["file_key"].(string)),
		}
		referenceData.S3ReferenceDataSource = s3rds
	}

	if v := rd["schema"].([]interface{}); len(v) > 0 {
		ss := expandKinesisAnalyticsV2SourceSchema(v[0].(map[string]interface{}))
		referenceData.ReferenceSchema = ss
	}

	return referenceData
}

func createApplicationV2UpdateOpts(d *schema.ResourceData) *kinesisanalyticsv2.UpdateApplicationInput {
	var cloudwatchLoggingUpdates []*kinesisanalyticsv2.CloudWatchLoggingOptionUpdate
	oldLoggingOptions, newLoggingOptions := d.GetChange("cloudwatch_logging_options")
	if len(oldLoggingOptions.([]interface{})) > 0 && len(newLoggingOptions.([]interface{})) > 0 {
		if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
			clo := v.([]interface{})[0].(map[string]interface{})
			cloudwatchLoggingOption := expandKinesisAnalyticsV2CloudwatchLoggingOptionUpdate(clo)
			cloudwatchLoggingUpdates = []*kinesisanalyticsv2.CloudWatchLoggingOptionUpdate{cloudwatchLoggingOption}
		}
	}

	runtime := d.Get("runtime_environment").(string)

	return &kinesisanalyticsv2.UpdateApplicationInput{
		ApplicationConfigurationUpdate: createKinesisAnalyticsV2ApplicationUpdateOpts(
			runtime, d),
		CloudWatchLoggingOptionUpdates: cloudwatchLoggingUpdates,
	}
}

func createKinesisAnalyticsV2SqlUpdateOpts(d *schema.ResourceData) *kinesisanalyticsv2.SqlApplicationConfigurationUpdate {
	var sqlUpdate *kinesisanalyticsv2.SqlApplicationConfigurationUpdate
	var inputsUpdate []*kinesisanalyticsv2.InputUpdate
	var outputsUpdate []*kinesisanalyticsv2.OutputUpdate
	var referenceDataUpdate []*kinesisanalyticsv2.ReferenceDataSourceUpdate

	sc := d.Get("application_configuration.0.sql_application_configuration").([]interface{})[0].(map[string]interface{})
	oldConfigIfc, _ := d.GetChange("application_configuration.0.sql_application_configuration")
	oldConfig := oldConfigIfc.([]interface{})
	var hasOldInputs, hasOldOutputs bool

	if len(oldConfig) > 0 {
		hasOldInputs = len(oldConfig[0].(map[string]interface{})["input"].([]interface{})) > 0
		hasOldOutputs = len(oldConfig[0].(map[string]interface{})["output"].(*schema.Set).List()) > 0
	}
	if hasOldInputs {
		if iConf := sc["input"].([]interface{}); len(iConf) > 0 {
			inputsUpdate = []*kinesisanalyticsv2.InputUpdate{expandKinesisAnalyticsV2InputUpdate(iConf[0].(map[string]interface{}))}
		}
	}
	if hasOldOutputs {
		if oConf := sc["output"].(*schema.Set).List(); len(oConf) > 0 {
			outputsUpdate = []*kinesisanalyticsv2.OutputUpdate{
				expandKinesisAnalyticsV2OutputUpdate(oldConfig[0].(map[string]interface{})["output"].(*schema.Set).List()[0].(map[string]interface{}), oConf[0].(map[string]interface{}))}
		}
	}
	rConf := sc["reference_data_sources"].([]interface{})
	for _, rd := range rConf {
		rdL := rd.(map[string]interface{})
		rdsu := &kinesisanalyticsv2.ReferenceDataSourceUpdate{
			ReferenceId:     aws.String(rdL["id"].(string)),
			TableNameUpdate: aws.String(rdL["table_name"].(string)),
		}

		if v := rdL["s3"].([]interface{}); len(v) > 0 {
			vL := v[0].(map[string]interface{})
			s3rdsu := &kinesisanalyticsv2.S3ReferenceDataSourceUpdate{
				BucketARNUpdate: aws.String(vL["bucket_arn"].(string)),
				FileKeyUpdate:   aws.String(vL["file_key"].(string)),
			}
			rdsu.S3ReferenceDataSourceUpdate = s3rdsu
		}

		if v := rdL["schema"].([]interface{}); len(v) > 0 {
			vL := v[0].(map[string]interface{})
			ss := expandKinesisAnalyticsV2SourceSchema(vL)
			rdsu.ReferenceSchemaUpdate = ss
		}

		referenceDataUpdate = append(referenceDataUpdate, rdsu)
	}
	if inputsUpdate != nil || outputsUpdate != nil || referenceDataUpdate != nil {
		sqlUpdate = &kinesisanalyticsv2.SqlApplicationConfigurationUpdate{
			InputUpdates:               inputsUpdate,
			OutputUpdates:              outputsUpdate,
			ReferenceDataSourceUpdates: referenceDataUpdate,
		}
	}
	return sqlUpdate
}

func createKinesisAnalyticsV2ApplicationUpdateOpts(runtime string, d *schema.ResourceData) *kinesisanalyticsv2.ApplicationConfigurationUpdate {

	codeConfigUpdate := createKinesisAnalyticsV2ApplicationCodeConfigurationUpdateOpts(d)

	var flinkUpdate *kinesisanalyticsv2.FlinkApplicationConfigurationUpdate
	if runtimeIsFlink(runtime) {
		flinkUpdate = createKinesisAnalyticsFlinkUpdateOpts(d)
	}

	var propertyGroupsUpdate *kinesisanalyticsv2.EnvironmentPropertyUpdates
	if d.HasChange("application_configuration.0.environment_properties") {
		lst := d.Get("application_configuration.0.environment_properties").([]interface{})
		if len(lst) > 0 {
			m := lst[0].(map[string]interface{})
			propertyGroupsUpdate = &kinesisanalyticsv2.EnvironmentPropertyUpdates{
				PropertyGroups: expandPropertyGroups(m["property_group"].(*schema.Set).List()),
			}
		}
	}

	var snapshotUpdate *kinesisanalyticsv2.ApplicationSnapshotConfigurationUpdate
	if d.HasChange("application_configuration.0.application_snapshot_configuration") {
		snapshotConfig := expandKinesisAnalyticsV2ApplicationSnapshotConfiguration(d.Get(
			"application_configuration.0.application_snapshot_configuration").(*schema.Set))
		snapshotUpdate = &kinesisanalyticsv2.ApplicationSnapshotConfigurationUpdate{
			SnapshotsEnabledUpdate: snapshotConfig.SnapshotsEnabled,
		}
	}

	var sqlUpdate *kinesisanalyticsv2.SqlApplicationConfigurationUpdate
	if runtime == kinesisanalyticsv2.RuntimeEnvironmentSql10 {
		sqlUpdate = createKinesisAnalyticsV2SqlUpdateOpts(d)
	}

	var applicationUpdate *kinesisanalyticsv2.ApplicationConfigurationUpdate
	if codeConfigUpdate != nil || snapshotUpdate != nil || propertyGroupsUpdate != nil || flinkUpdate != nil || sqlUpdate != nil {
		applicationUpdate = &kinesisanalyticsv2.ApplicationConfigurationUpdate{
			ApplicationCodeConfigurationUpdate:     codeConfigUpdate,
			ApplicationSnapshotConfigurationUpdate: snapshotUpdate,
			EnvironmentPropertyUpdates:             propertyGroupsUpdate,
			FlinkApplicationConfigurationUpdate:    flinkUpdate,
			SqlApplicationConfigurationUpdate:      sqlUpdate,
		}
	}
	return applicationUpdate
}

func createKinesisAnalyticsV2ApplicationCodeConfigurationUpdateOpts(d *schema.ResourceData) *kinesisanalyticsv2.ApplicationCodeConfigurationUpdate {
	if !d.HasChange("application_configuration.0.application_code_configuration.0.code_content") &&
		!d.HasChange("application_configuration.0.application_code_configuration.0.code_content_type") {
		return nil
	}

	codeConfigUpdate := &kinesisanalyticsv2.ApplicationCodeConfigurationUpdate{}
	codeConfig := d.Get("application_configuration.0.application_code_configuration").([]interface{})
	if len(codeConfig) == 0 {
		return codeConfigUpdate
	}
	cc := expandKinesisAnalyticsV2ApplicationCodeConfiguration(codeConfig)

	var s3ContentLocationUpdate *kinesisanalyticsv2.S3ContentLocationUpdate
	if cc.CodeContent.S3ContentLocation != nil {
		s3ContentLocationUpdate = &kinesisanalyticsv2.S3ContentLocationUpdate{
			BucketARNUpdate:     cc.CodeContent.S3ContentLocation.BucketARN,
			FileKeyUpdate:       cc.CodeContent.S3ContentLocation.FileKey,
			ObjectVersionUpdate: cc.CodeContent.S3ContentLocation.ObjectVersion,
		}
	}

	return &kinesisanalyticsv2.ApplicationCodeConfigurationUpdate{
		CodeContentTypeUpdate: cc.CodeContentType,
		CodeContentUpdate: &kinesisanalyticsv2.CodeContentUpdate{
			TextContentUpdate:       cc.CodeContent.TextContent,
			S3ContentLocationUpdate: s3ContentLocationUpdate,
			// TODO: add zip file contents
		},
	}
}

func createKinesisAnalyticsFlinkUpdateOpts(d *schema.ResourceData) *kinesisanalyticsv2.FlinkApplicationConfigurationUpdate {
	var flinkUpdate *kinesisanalyticsv2.FlinkApplicationConfigurationUpdate
	var checkpointUpdate *kinesisanalyticsv2.CheckpointConfigurationUpdate
	var monitoringUpdate *kinesisanalyticsv2.MonitoringConfigurationUpdate
	var parallelismUpdate *kinesisanalyticsv2.ParallelismConfigurationUpdate
	if d.HasChange("application_configuration.0.flink_application_configuration") {
		fc := d.Get("application_configuration.0.flink_application_configuration").([]interface{})[0].(map[string]interface{})
		cpConf := fc["checkpoint_configuration"].(*schema.Set)
		checkpointConfig := expandCheckpointConfiguration(cpConf)
		checkpointUpdate = &kinesisanalyticsv2.CheckpointConfigurationUpdate{
			CheckpointIntervalUpdate:         checkpointConfig.CheckpointInterval,
			CheckpointingEnabledUpdate:       checkpointConfig.CheckpointingEnabled,
			ConfigurationTypeUpdate:          checkpointConfig.ConfigurationType,
			MinPauseBetweenCheckpointsUpdate: checkpointConfig.MinPauseBetweenCheckpoints,
		}
		montConf := fc["monitoring_configuration"].(*schema.Set)
		monitoringConfig := expandMonitoringConfiguration(montConf)
		monitoringUpdate = &kinesisanalyticsv2.MonitoringConfigurationUpdate{
			ConfigurationTypeUpdate: monitoringConfig.ConfigurationType,
			LogLevelUpdate:          monitoringConfig.LogLevel,
			MetricsLevelUpdate:      monitoringConfig.MetricsLevel,
		}
		paraConf := fc["parallelism_configuration"].(*schema.Set)
		parallelismConfig := expandParallelismConfiguration(paraConf)
		parallelismUpdate = &kinesisanalyticsv2.ParallelismConfigurationUpdate{
			AutoScalingEnabledUpdate: parallelismConfig.AutoScalingEnabled,
			ConfigurationTypeUpdate:  parallelismConfig.ConfigurationType,
			ParallelismUpdate:        parallelismConfig.Parallelism,
			ParallelismPerKPUUpdate:  parallelismConfig.ParallelismPerKPU,
		}
	}

	if checkpointUpdate != nil || monitoringUpdate != nil || parallelismUpdate != nil {
		flinkUpdate = &kinesisanalyticsv2.FlinkApplicationConfigurationUpdate{
			CheckpointConfigurationUpdate:  checkpointUpdate,
			MonitoringConfigurationUpdate:  monitoringUpdate,
			ParallelismConfigurationUpdate: parallelismUpdate,
		}
	}
	return flinkUpdate
}

func expandCheckpointConfiguration(config *schema.Set) *kinesisanalyticsv2.CheckpointConfiguration {
	var checkpointingEnabled *bool
	var checkpointInterval *int64
	var configurationType *string
	var checkpointMinPause *int64

	for _, v := range config.List() {
		m := v.(map[string]interface{})
		if interval, ok := m["checkpoint_interval"]; ok {
			checkpointInterval = aws.Int64(int64(interval.(int)))
		}
		if enabled, ok := m["checkpointing_enabled"]; ok {
			checkpointingEnabled = aws.Bool(enabled.(bool))
		}
		if confType, ok := m["configuration_type"]; ok {
			configurationType = aws.String(confType.(string))
		}
		if minPause, ok := m["min_pause_between_checkpoints"]; ok {
			checkpointMinPause = aws.Int64(int64(minPause.(int)))
		}
	}
	return &kinesisanalyticsv2.CheckpointConfiguration{
		CheckpointInterval:         checkpointInterval,
		CheckpointingEnabled:       checkpointingEnabled,
		ConfigurationType:          configurationType,
		MinPauseBetweenCheckpoints: checkpointMinPause,
	}
}

func expandMonitoringConfiguration(config *schema.Set) *kinesisanalyticsv2.MonitoringConfiguration {
	var configurationType *string
	var logLevel *string
	var metricsLevel *string

	for _, v := range config.List() {
		m := v.(map[string]interface{})
		if confType, ok := m["configuration_type"]; ok {
			configurationType = aws.String(confType.(string))
		}
		if level, ok := m["log_level"]; ok {
			logLevel = aws.String(level.(string))
		}
		if level, ok := m["metrics_level"]; ok {
			metricsLevel = aws.String(level.(string))
		}
	}
	return &kinesisanalyticsv2.MonitoringConfiguration{
		ConfigurationType: configurationType,
		LogLevel:          logLevel,
		MetricsLevel:      metricsLevel,
	}
}

func expandParallelismConfiguration(config *schema.Set) *kinesisanalyticsv2.ParallelismConfiguration {
	var autoscalingEnabled *bool
	var configurationType *string
	var parallelism *int64
	var parallelismPerKPU *int64

	for _, v := range config.List() {
		m := v.(map[string]interface{})
		if aEnabled, ok := m["autoscaling_enabled"]; ok {
			autoscalingEnabled = aws.Bool(aEnabled.(bool))
		}
		if confType, ok := m["configuration_type"]; ok {
			configurationType = aws.String(confType.(string))
		}
		if p, ok := m["parallelism"]; ok {
			parallelism = aws.Int64(int64(p.(int)))
		}
		if p, ok := m["parallelism_per_kpu"]; ok {
			parallelismPerKPU = aws.Int64(int64(p.(int)))
		}
	}
	return &kinesisanalyticsv2.ParallelismConfiguration{
		AutoScalingEnabled: autoscalingEnabled,
		ConfigurationType:  configurationType,
		Parallelism:        parallelism,
		ParallelismPerKPU:  parallelismPerKPU,
	}
}

func expandKinesisAnalyticsV2InputUpdate(vL map[string]interface{}) *kinesisanalyticsv2.InputUpdate {
	inputUpdate := &kinesisanalyticsv2.InputUpdate{
		InputId:          aws.String(vL["id"].(string)),
		NamePrefixUpdate: aws.String(vL["name_prefix"].(string)),
	}

	if v := vL["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfiu := &kinesisanalyticsv2.KinesisFirehoseInputUpdate{
			ResourceARNUpdate: aws.String(kf["resource_arn"].(string)),
		}
		inputUpdate.KinesisFirehoseInputUpdate = kfiu
	}

	if v := vL["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		ksiu := &kinesisanalyticsv2.KinesisStreamsInputUpdate{
			ResourceARNUpdate: aws.String(ks["resource_arn"].(string)),
		}
		inputUpdate.KinesisStreamsInputUpdate = ksiu
	}

	if v := vL["parallelism"].([]interface{}); len(v) > 0 {
		p := v[0].(map[string]interface{})

		if c, ok := p["count"]; ok {
			ipu := &kinesisanalyticsv2.InputParallelismUpdate{
				CountUpdate: aws.Int64(int64(c.(int))),
			}
			inputUpdate.InputParallelismUpdate = ipu
		}
	}

	if v := vL["processing_configuration"].([]interface{}); len(v) > 0 {
		pc := v[0].(map[string]interface{})

		if l := pc["lambda"].([]interface{}); len(l) > 0 {
			lp := l[0].(map[string]interface{})
			ipc := &kinesisanalyticsv2.InputProcessingConfigurationUpdate{
				InputLambdaProcessorUpdate: &kinesisanalyticsv2.InputLambdaProcessorUpdate{
					ResourceARNUpdate: aws.String(lp["resource_arn"].(string)),
				},
			}
			inputUpdate.InputProcessingConfigurationUpdate = ipc
		}
	}

	if v := vL["schema"].([]interface{}); len(v) > 0 {
		ss := &kinesisanalyticsv2.InputSchemaUpdate{}
		vL := v[0].(map[string]interface{})

		if v := vL["record_column"].([]interface{}); len(v) > 0 {
			var rcs []*kinesisanalyticsv2.RecordColumn

			for _, rc := range v {
				rcD := rc.(map[string]interface{})
				rc := &kinesisanalyticsv2.RecordColumn{
					Name:    aws.String(rcD["name"].(string)),
					SqlType: aws.String(rcD["sql_type"].(string)),
				}

				if v, ok := rcD["mapping"]; ok {
					rc.Mapping = aws.String(v.(string))
				}

				rcs = append(rcs, rc)
			}

			ss.RecordColumnUpdates = rcs
		}

		if v, ok := vL["record_encoding"]; ok && v.(string) != "" {
			ss.RecordEncodingUpdate = aws.String(v.(string))
		}

		if v := vL["record_format"].([]interface{}); len(v) > 0 {
			vL := v[0].(map[string]interface{})
			rf := &kinesisanalyticsv2.RecordFormat{}

			if v := vL["mapping_parameters"].([]interface{}); len(v) > 0 {
				vL := v[0].(map[string]interface{})
				mp := &kinesisanalyticsv2.MappingParameters{}

				if v := vL["csv"].([]interface{}); len(v) > 0 {
					cL := v[0].(map[string]interface{})
					cmp := &kinesisanalyticsv2.CSVMappingParameters{
						RecordColumnDelimiter: aws.String(cL["record_column_delimiter"].(string)),
						RecordRowDelimiter:    aws.String(cL["record_row_delimiter"].(string)),
					}
					mp.CSVMappingParameters = cmp
					rf.RecordFormatType = aws.String("CSV")
				}

				if v := vL["json"].([]interface{}); len(v) > 0 {
					jL := v[0].(map[string]interface{})
					jmp := &kinesisanalyticsv2.JSONMappingParameters{
						RecordRowPath: aws.String(jL["record_row_path"].(string)),
					}
					mp.JSONMappingParameters = jmp
					rf.RecordFormatType = aws.String("JSON")
				}
				rf.MappingParameters = mp
			}
			ss.RecordFormatUpdate = rf
		}
		inputUpdate.InputSchemaUpdate = ss
	}

	return inputUpdate
}

func expandKinesisAnalyticsV2OutputUpdate(oldOutput map[string]interface{}, newOutput map[string]interface{}) *kinesisanalyticsv2.OutputUpdate {
	outputUpdate := &kinesisanalyticsv2.OutputUpdate{
		OutputId:   aws.String(oldOutput["id"].(string)),
		NameUpdate: aws.String(newOutput["name"].(string)),
	}

	if v := newOutput["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfou := &kinesisanalyticsv2.KinesisFirehoseOutputUpdate{
			ResourceARNUpdate: aws.String(kf["resource_arn"].(string)),
		}
		outputUpdate.KinesisFirehoseOutputUpdate = kfou
	}

	if v := newOutput["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		ksou := &kinesisanalyticsv2.KinesisStreamsOutputUpdate{
			ResourceARNUpdate: aws.String(ks["resource_arn"].(string)),
		}
		outputUpdate.KinesisStreamsOutputUpdate = ksou
	}

	if v := newOutput["lambda"].([]interface{}); len(v) > 0 {
		l := v[0].(map[string]interface{})
		lou := &kinesisanalyticsv2.LambdaOutputUpdate{
			ResourceARNUpdate: aws.String(l["resource_arn"].(string)),
		}
		outputUpdate.LambdaOutputUpdate = lou
	}

	if v := newOutput["schema"].([]interface{}); len(v) > 0 {
		ds := v[0].(map[string]interface{})
		dsu := &kinesisanalyticsv2.DestinationSchema{
			RecordFormatType: aws.String(ds["record_format_type"].(string)),
		}
		outputUpdate.DestinationSchemaUpdate = dsu
	}

	return outputUpdate
}

func expandKinesisAnalyticsV2CloudwatchLoggingOptionUpdate(clo map[string]interface{}) *kinesisanalyticsv2.CloudWatchLoggingOptionUpdate {
	cloudwatchLoggingOption := &kinesisanalyticsv2.CloudWatchLoggingOptionUpdate{
		CloudWatchLoggingOptionId: aws.String(clo["cloudwatch_logging_option_id"].(string)),
		LogStreamARNUpdate:        aws.String(clo["log_stream_arn"].(string)),
	}
	return cloudwatchLoggingOption
}

func expandKinesisAnalayticsV2SqlApplicationConfiguration(appConfig map[string]interface{}) *kinesisanalyticsv2.SqlApplicationConfiguration {
	sqlApplicationConfiguration := &kinesisanalyticsv2.SqlApplicationConfiguration{}
	if appConfig["input"] != nil {
		if v := appConfig["input"].([]interface{}); len(v) > 0 {
			i := v[0].(map[string]interface{})
			inputs := expandKinesisAnalyticsV2Input(i)
			sqlApplicationConfiguration.Inputs = []*kinesisanalyticsv2.Input{inputs}
		}
	}
	if appConfig["output"] != nil {
		if v := appConfig["output"].(*schema.Set).List(); len(v) > 0 {
			outputs := make([]*kinesisanalyticsv2.Output, 0)
			for _, o := range v {
				output := expandKinesisAnalyticsV2Output(o.(map[string]interface{}))
				outputs = append(outputs, output)
			}
			sqlApplicationConfiguration.Outputs = outputs
		}
	}
	if appConfig["reference_data_sources"] != nil {
		if v := appConfig["reference_data_sources"].([]interface{}); len(v) > 0 {
			references := make([]*kinesisanalyticsv2.ReferenceDataSource, 0)
			for _, r := range v {
				references = append(references, expandKinesisAnalyticsV2ReferenceData(r.(map[string]interface{})))
			}
			sqlApplicationConfiguration.ReferenceDataSources = references
		}
	}
	return sqlApplicationConfiguration
}

func expandKinesisAnalyticsV2FlinkApplicationConfiguration(appConfig map[string]interface{}) *kinesisanalyticsv2.FlinkApplicationConfiguration {
	flinkApplicationConfiguration := &kinesisanalyticsv2.FlinkApplicationConfiguration{}

	flinkApplicationConfiguration.CheckpointConfiguration = expandCheckpointConfiguration(appConfig["checkpoint_configuration"].(*schema.Set))
	flinkApplicationConfiguration.MonitoringConfiguration = expandMonitoringConfiguration(appConfig["monitoring_configuration"].(*schema.Set))
	flinkApplicationConfiguration.ParallelismConfiguration = expandParallelismConfiguration(appConfig["parallelism_configuration"].(*schema.Set))

	return flinkApplicationConfiguration
}

func expandKinesisAnalyticsV2ApplicationCodeConfiguration(conf []interface{}) *kinesisanalyticsv2.ApplicationCodeConfiguration {
	if len(conf) < 1 {
		return nil
	}
	codeConfig := conf[0].(map[string]interface{})
	contentType := aws.String(codeConfig["code_content_type"].(string))

	var codeContent *kinesisanalyticsv2.CodeContent
	if v, ok := codeConfig["code_content"]; ok && v != nil {
		codeContent = expandKinesisAnalyticsV2CodeContent(v.([]interface{}))
	}

	return &kinesisanalyticsv2.ApplicationCodeConfiguration{
		CodeContentType: contentType,
		CodeContent:     codeContent,
	}
}

func expandKinesisAnalyticsV2CodeContent(cc []interface{}) *kinesisanalyticsv2.CodeContent {
	var s3ContentLocation *kinesisanalyticsv2.S3ContentLocation
	var textContent *string
	for _, v := range cc {
		m := v.(map[string]interface{})
		if tc, ok := m["text_content"]; ok && tc != "" {
			textContent = aws.String(tc.(string))
		}
		if loc, ok := m["s3_content_location"].([]interface{}); ok && len(loc) > 0 {
			locMap := loc[0].(map[string]interface{})
			if len(locMap) > 0 {
				var objectVersion *string
				// Object version is optional
				if v, ok := locMap["object_version"]; ok {
					objectVersion = aws.String(v.(string))
				}
				s3ContentLocation = &kinesisanalyticsv2.S3ContentLocation{
					BucketARN:     aws.String(locMap["bucket_arn"].(string)),
					FileKey:       aws.String(locMap["file_key"].(string)),
					ObjectVersion: objectVersion,
				}
			}
		}
	}
	return &kinesisanalyticsv2.CodeContent{
		TextContent:       textContent,
		S3ContentLocation: s3ContentLocation,
	}
}

func expandKinesisAnalyticsV2ApplicationSnapshotConfiguration(s *schema.Set) *kinesisanalyticsv2.ApplicationSnapshotConfiguration {
	v := s.List()
	if len(v) == 0 {
		return nil
	}
	var snapshotsEnabled bool
	m := v[0].(map[string]interface{})
	if enabled, ok := m["snapshots_enabled"]; ok {
		snapshotsEnabled = enabled.(bool)
	}
	return &kinesisanalyticsv2.ApplicationSnapshotConfiguration{
		SnapshotsEnabled: aws.Bool(snapshotsEnabled),
	}
}

func flattenKinesisAnalyticsV2ApplicationConfiguration(runtime string, appConfig *kinesisanalyticsv2.ApplicationConfigurationDescription) []interface{} {

	ret := map[string]interface{}{}
	if appConfig == nil {
		return []interface{}{ret}
	}

	// ApplicationCodeConfiguration is required
	ret["application_code_configuration"] = flattenKinesisAnalyticsV2ApplicationCodeConfiguration(appConfig.ApplicationCodeConfigurationDescription)

	if runtime == kinesisanalyticsv2.RuntimeEnvironmentSql10 {
		ret["sql_application_configuration"] = flattenSqlApplicationConfigurationDescription(appConfig.SqlApplicationConfigurationDescription)
	} else if runtimeIsFlink(runtime) {
		ret["flink_application_configuration"] = flattenFlinkApplicationConfigurationDescription(appConfig.FlinkApplicationConfigurationDescription)
	}
	if appConfig.EnvironmentPropertyDescriptions != nil {
		ret["environment_properties"] = flattenKinesisAnalyticsV2EnvironmentProperties(appConfig.EnvironmentPropertyDescriptions)
	}
	if appConfig.ApplicationSnapshotConfigurationDescription != nil {
		ret["application_snapshot_configuration"] = flattenKinesisAnalyticsV2SnapshotConfiguration(appConfig.ApplicationSnapshotConfigurationDescription)
	}

	return []interface{}{ret}
}

func flattenKinesisAnalyticsV2SnapshotConfiguration(snapshotConfig *kinesisanalyticsv2.ApplicationSnapshotConfigurationDescription) *schema.Set {
	return schema.NewSet(resourceKinesisAnalyticsV2ApplicationSnapshotConfigurationHash, []interface{}{map[string]interface{}{
		"snapshots_enabled": aws.BoolValue(snapshotConfig.SnapshotsEnabled),
	}})
}

func flattenKinesisAnalyticsV2EnvironmentProperties(envProps *kinesisanalyticsv2.EnvironmentPropertyDescriptions) []interface{} {
	items := []interface{}{}
	for _, group := range envProps.PropertyGroupDescriptions {
		items = append(items, map[string]interface{}{
			"property_group_id": aws.StringValue(group.PropertyGroupId),
			"property_map":      flattenKinesisAnalyticsV2ApplicationPropertyMap(group.PropertyMap),
		})
	}
	set := schema.NewSet(resourceKinesisAnalyticsV2ApplicationPropertyGroupHash, items)
	return []interface{}{map[string]interface{}{"property_group": set}}
}

func flattenKinesisAnalyticsV2ApplicationPropertyMap(m map[string]*string) map[string]string {
	flattened := make(map[string]string, len(m))
	for k, v := range m {
		flattened[k] = aws.StringValue(v)
	}
	return flattened
}

func flattenKinesisAnalyticsV2ApplicationCodeConfiguration(codeConfig *kinesisanalyticsv2.ApplicationCodeConfigurationDescription) []interface{} {

	appCodeConfig := make(map[string]interface{})
	if codeConfig == nil {
		return []interface{}{appCodeConfig}
	}
	codeContent := make(map[string]interface{})
	appCodeConfig["code_content_type"] = *codeConfig.CodeContentType
	if contentDesc := codeConfig.CodeContentDescription; contentDesc != nil {
		if contentDesc.TextContent != nil {
			codeContent["text_content"] = *contentDesc.TextContent
		}
		if locDesc := contentDesc.S3ApplicationCodeLocationDescription; locDesc != nil {
			locationDescription := make(map[string]interface{})
			locationDescription["bucket_arn"] = *locDesc.BucketARN
			locationDescription["file_key"] = *locDesc.FileKey
			if locDesc.ObjectVersion != nil {
				locationDescription["object_version"] = *locDesc.ObjectVersion
			}
			codeContent["s3_content_location"] = []interface{}{locationDescription}
		}
		appCodeConfig["code_content"] = []interface{}{codeContent}
	}

	return []interface{}{appCodeConfig}
}

func flattenSqlApplicationConfigurationDescription(sqlApplicationConfig *kinesisanalyticsv2.SqlApplicationConfigurationDescription) []interface{} {
	ret := map[string]interface{}{}

	if sqlApplicationConfig == nil {
		return []interface{}{ret}
	}

	ret["input"] = flattenKinesisAnalyticsV2Inputs(sqlApplicationConfig.InputDescriptions)
	ret["output"] = flattenKinesisAnalyticsV2Outputs(sqlApplicationConfig.OutputDescriptions)
	ret["reference_data_sources"] = flattenKinesisAnalyticsV2ReferenceDataSources(sqlApplicationConfig.ReferenceDataSourceDescriptions)
	return []interface{}{ret}
}

func flattenFlinkApplicationConfigurationDescription(flinkApplicationConfig *kinesisanalyticsv2.FlinkApplicationConfigurationDescription) []interface{} {
	if flinkApplicationConfig == nil {
		return []interface{}{}
	}
	return []interface{}{map[string]interface{}{
		"checkpoint_configuration":  flattenCheckpointConfiguration(flinkApplicationConfig.CheckpointConfigurationDescription),
		"monitoring_configuration":  flattenMonitoringConfiguration(flinkApplicationConfig.MonitoringConfigurationDescription),
		"parallelism_configuration": flattenParallelismConfiguration(flinkApplicationConfig.ParallelismConfigurationDescription),
	},
	}
}

func flattenCheckpointConfiguration(checkpointConfiguration *kinesisanalyticsv2.CheckpointConfigurationDescription) *schema.Set {
	return schema.NewSet(resourceKinesisAnalyticsV2ApplicationCheckpointConfigurationHash, []interface{}{map[string]interface{}{
		"checkpoint_interval":           aws.Int64Value(checkpointConfiguration.CheckpointInterval),
		"checkpointing_enabled":         aws.BoolValue(checkpointConfiguration.CheckpointingEnabled),
		"configuration_type":            aws.StringValue(checkpointConfiguration.ConfigurationType),
		"min_pause_between_checkpoints": aws.Int64Value(checkpointConfiguration.MinPauseBetweenCheckpoints),
	}})
}

func flattenMonitoringConfiguration(monitoringConfiguration *kinesisanalyticsv2.MonitoringConfigurationDescription) *schema.Set {
	return schema.NewSet(resourceKinesisAnalyticsV2ApplicationMonitoringConfigurationHash, []interface{}{map[string]interface{}{
		"configuration_type": aws.StringValue(monitoringConfiguration.ConfigurationType),
		"log_level":          aws.StringValue(monitoringConfiguration.LogLevel),
		"metrics_level":      aws.StringValue(monitoringConfiguration.MetricsLevel),
	}})
}

func flattenParallelismConfiguration(parallelismConfiguration *kinesisanalyticsv2.ParallelismConfigurationDescription) *schema.Set {
	return schema.NewSet(resourceKinesisAnalyticsV2ApplicationParallelismConfigurationHash, []interface{}{map[string]interface{}{
		"autoscaling_enabled": aws.BoolValue(parallelismConfiguration.AutoScalingEnabled),
		"configuration_type":  aws.StringValue(parallelismConfiguration.ConfigurationType),
		"parallelism":         aws.Int64Value(parallelismConfiguration.Parallelism),
		"parallelism_per_kpu": aws.Int64Value(parallelismConfiguration.ParallelismPerKPU),
	}})
}

func flattenKinesisAnalyticsV2CloudwatchLoggingOptions(options []*kinesisanalyticsv2.CloudWatchLoggingOptionDescription) []interface{} {
	s := []interface{}{}
	for _, v := range options {
		option := map[string]interface{}{
			"cloudwatch_logging_option_id": aws.StringValue(v.CloudWatchLoggingOptionId),
			"log_stream_arn":               aws.StringValue(v.LogStreamARN),
		}
		s = append(s, option)
	}
	return s
}

func flattenKinesisAnalyticsV2Inputs(inputs []*kinesisanalyticsv2.InputDescription) []interface{} {
	if len(inputs) == 0 {
		return []interface{}{}
	}
	id := inputs[0]

	input := map[string]interface{}{
		"id":          aws.StringValue(id.InputId),
		"name_prefix": aws.StringValue(id.NamePrefix),
	}

	list := schema.NewSet(schema.HashString, nil)
	for _, sn := range id.InAppStreamNames {
		list.Add(aws.StringValue(sn))
	}
	input["stream_names"] = list

	if id.InputParallelism != nil {
		input["parallelism"] = []interface{}{
			map[string]interface{}{
				"count": int(aws.Int64Value(id.InputParallelism.Count)),
			},
		}
	}

	if id.InputProcessingConfigurationDescription != nil {
		ipcd := id.InputProcessingConfigurationDescription

		if ipcd.InputLambdaProcessorDescription != nil {
			input["processing_configuration"] = []interface{}{
				map[string]interface{}{
					"lambda": []interface{}{
						map[string]interface{}{
							"resource_arn": aws.StringValue(ipcd.InputLambdaProcessorDescription.ResourceARN),
						},
					},
				},
			}
		}
	}

	if id.InputSchema != nil {
		inputSchema := id.InputSchema
		is := []interface{}{}
		rcs := []interface{}{}
		ss := map[string]interface{}{
			"record_encoding": aws.StringValue(inputSchema.RecordEncoding),
		}

		for _, rc := range inputSchema.RecordColumns {
			rcM := map[string]interface{}{
				"mapping":  aws.StringValue(rc.Mapping),
				"name":     aws.StringValue(rc.Name),
				"sql_type": aws.StringValue(rc.SqlType),
			}
			rcs = append(rcs, rcM)
		}
		ss["record_column"] = schema.NewSet(resourceKinesisAnalyticsV2ApplicationRecordColumnHash, rcs)

		if inputSchema.RecordFormat != nil {
			rf := inputSchema.RecordFormat
			rfM := map[string]interface{}{
				"record_format_type": aws.StringValue(rf.RecordFormatType),
			}

			if rf.MappingParameters != nil {
				mps := []interface{}{}
				if rf.MappingParameters.CSVMappingParameters != nil {
					cmp := map[string]interface{}{
						"csv": []interface{}{
							map[string]interface{}{
								"record_column_delimiter": aws.StringValue(rf.MappingParameters.CSVMappingParameters.RecordColumnDelimiter),
								"record_row_delimiter":    aws.StringValue(rf.MappingParameters.CSVMappingParameters.RecordRowDelimiter),
							},
						},
					}
					mps = append(mps, cmp)
				}

				if rf.MappingParameters.JSONMappingParameters != nil {
					jmp := map[string]interface{}{
						"json": []interface{}{
							map[string]interface{}{
								"record_row_path": aws.StringValue(rf.MappingParameters.JSONMappingParameters.RecordRowPath),
							},
						},
					}
					mps = append(mps, jmp)
				}

				rfM["mapping_parameters"] = mps
			}
			ss["record_format"] = []interface{}{rfM}
		}

		is = append(is, ss)
		input["schema"] = is
	}

	if id.InputStartingPositionConfiguration != nil && id.InputStartingPositionConfiguration.InputStartingPosition != nil {
		input["starting_position_configuration"] = []interface{}{
			map[string]interface{}{
				"starting_position": aws.StringValue(id.InputStartingPositionConfiguration.InputStartingPosition),
			},
		}
	}

	if id.KinesisFirehoseInputDescription != nil {
		input["kinesis_firehose"] = []interface{}{
			map[string]interface{}{
				"resource_arn": aws.StringValue(id.KinesisFirehoseInputDescription.ResourceARN),
			},
		}
	}

	if id.KinesisStreamsInputDescription != nil {
		input["kinesis_stream"] = []interface{}{
			map[string]interface{}{
				"resource_arn": aws.StringValue(id.KinesisStreamsInputDescription.ResourceARN),
			},
		}
	}

	return []interface{}{input}
}

func flattenKinesisAnalyticsV2Outputs(outputs []*kinesisanalyticsv2.OutputDescription) *schema.Set {
	s := []interface{}{}
	{
	}
	for _, o := range outputs {
		output := map[string]interface{}{
			"id":   aws.StringValue(o.OutputId),
			"name": aws.StringValue(o.Name),
		}

		if o.KinesisFirehoseOutputDescription != nil {
			output["kinesis_firehose"] = []interface{}{
				map[string]interface{}{
					"resource_arn": aws.StringValue(o.KinesisFirehoseOutputDescription.ResourceARN),
				},
			}
		}

		if o.KinesisStreamsOutputDescription != nil {
			output["kinesis_stream"] = []interface{}{
				map[string]interface{}{
					"resource_arn": aws.StringValue(o.KinesisStreamsOutputDescription.ResourceARN),
				},
			}
		}

		if o.LambdaOutputDescription != nil {
			output["lambda"] = []interface{}{
				map[string]interface{}{
					"resource_arn": aws.StringValue(o.LambdaOutputDescription.ResourceARN),
				},
			}
		}

		if o.DestinationSchema != nil {
			output["schema"] = []interface{}{
				map[string]interface{}{
					"record_format_type": aws.StringValue(o.DestinationSchema.RecordFormatType),
				},
			}
		}

		s = append(s, output)
	}

	return schema.NewSet(resourceKinesisAnalyticsV2ApplicationOutputHash, s)
}

func flattenKinesisAnalyticsV2ReferenceDataSources(dataSources []*kinesisanalyticsv2.ReferenceDataSourceDescription) []interface{} {
	s := []interface{}{}

	if len(dataSources) > 0 {
		for _, ds := range dataSources {
			dataSource := map[string]interface{}{
				"id":         aws.StringValue(ds.ReferenceId),
				"table_name": aws.StringValue(ds.TableName),
			}

			if ds.S3ReferenceDataSourceDescription != nil {
				dataSource["s3"] = []interface{}{
					map[string]interface{}{
						"bucket_arn": aws.StringValue(ds.S3ReferenceDataSourceDescription.BucketARN),
						"file_key":   aws.StringValue(ds.S3ReferenceDataSourceDescription.FileKey),
					},
				}
			}

			if ds.ReferenceSchema != nil {
				rs := ds.ReferenceSchema
				rcs := []interface{}{}
				ss := map[string]interface{}{
					"record_encoding": aws.StringValue(rs.RecordEncoding),
				}

				for _, rc := range rs.RecordColumns {
					rcM := map[string]interface{}{
						"mapping":  aws.StringValue(rc.Mapping),
						"name":     aws.StringValue(rc.Name),
						"sql_type": aws.StringValue(rc.SqlType),
					}
					rcs = append(rcs, rcM)
				}
				ss["record_column"] = schema.NewSet(resourceKinesisAnalyticsV2ApplicationRecordColumnHash, rcs)

				if rs.RecordFormat != nil {
					rf := rs.RecordFormat
					rfM := map[string]interface{}{
						"record_format_type": aws.StringValue(rf.RecordFormatType),
					}

					if rf.MappingParameters != nil {
						mps := []interface{}{}
						if rf.MappingParameters.CSVMappingParameters != nil {
							cmp := map[string]interface{}{
								"csv": []interface{}{
									map[string]interface{}{
										"record_column_delimiter": aws.StringValue(rf.MappingParameters.CSVMappingParameters.RecordColumnDelimiter),
										"record_row_delimiter":    aws.StringValue(rf.MappingParameters.CSVMappingParameters.RecordRowDelimiter),
									},
								},
							}
							mps = append(mps, cmp)
						}

						if rf.MappingParameters.JSONMappingParameters != nil {
							jmp := map[string]interface{}{
								"json": []interface{}{
									map[string]interface{}{
										"record_row_path": aws.StringValue(rf.MappingParameters.JSONMappingParameters.RecordRowPath),
									},
								},
							}
							mps = append(mps, jmp)
						}

						rfM["mapping_parameters"] = mps
					}
					ss["record_format"] = []interface{}{rfM}
				}

				dataSource["schema"] = []interface{}{ss}
			}

			s = append(s, dataSource)
		}
	}

	return s
}

func runtimeIsFlink(runtime string) bool {
	return runtime == kinesisanalyticsv2.RuntimeEnvironmentFlink16 ||
		runtime == kinesisanalyticsv2.RuntimeEnvironmentFlink18
}

func resourceKinesisAnalyticsV2ApplicationSnapshotConfigurationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["snapshots_enabled"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}
	return hashcode.String(buf.String())
}

func resourceKinesisAnalyticsV2ApplicationPropertyGroupHash(v interface{}) int {
	var buf bytes.Buffer

	m := v.(map[string]interface{})

	if v, ok := m["property_group_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	// Sort properties before writing them
	if v, ok := m["property_map"]; ok {
		properties := []string{}
		pm := v.(map[string]string)
		for k := range pm {
			properties = append(properties, k)
		}
		sort.Strings(properties)
		for _, k := range properties {
			buf.WriteString(fmt.Sprintf("%s-%s-", k, pm[k]))
		}
	}

	return hashcode.String(buf.String())
}

func resourceKinesisAnalyticsV2ApplicationCheckpointConfigurationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["checkpoint_interval"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int64)))
	}
	if v, ok := m["checkpointing_enabled"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}
	if v, ok := m["configuration_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["min_pause_between_checkpoints"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int64)))
	}

	return hashcode.String(buf.String())
}

func resourceKinesisAnalyticsV2ApplicationMonitoringConfigurationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["log_level"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["configuration_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["metrics_level"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}

func resourceKinesisAnalyticsV2ApplicationParallelismConfigurationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["autoscaling_enabled"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}
	if v, ok := m["parallelism"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int64)))
	}
	if v, ok := m["parallelism_per_kpu"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int64)))
	}
	if v, ok := m["configuration_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}

func resourceKinesisAnalyticsV2ApplicationRecordColumnHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["mapping"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["sql_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}

func resourceKinesisAnalyticsV2ApplicationOutputHash(v interface{}) int {
	var buf bytes.Buffer

	m := v.(map[string]interface{})

	if kf, ok := m["kinesis_firehose"]; ok {
		val := kf.([]interface{})[0].(map[string]interface{})
		buf.WriteString(fmt.Sprintf("%s-", val["resource_arn"].(string)))
	}

	if ks, ok := m["kinesis_stream"]; ok {
		val := ks.([]interface{})[0].(map[string]interface{})
		buf.WriteString(fmt.Sprintf("%s-", val["resource_arn"].(string)))
	}

	if l, ok := m["lambda"]; ok {
		val := l.([]interface{})[0].(map[string]interface{})
		buf.WriteString(fmt.Sprintf("%s-", val["resource_arn"].(string)))
	}

	if s, ok := m["schema"]; ok {
		val := s.([]interface{})[0].(map[string]interface{})
		if rft, ok := val["record_format_type"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", rft.(string)))
		}
	}

	return hashcode.String(buf.String())
}
*/
