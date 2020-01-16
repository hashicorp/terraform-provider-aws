package aws

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsKinesisAnalyticsApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisAnalyticsApplicationCreate,
		Read:   resourceAwsKinesisAnalyticsApplicationRead,
		Update: resourceAwsKinesisAnalyticsApplicationUpdate,
		Delete: resourceAwsKinesisAnalyticsApplicationDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				arns := strings.Split(d.Id(), ":")
				name := strings.Replace(arns[len(arns)-1], "application/", "", 1)
				d.Set("name", name)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service_execution_role": {
				Type:     schema.TypeString,
				Required: true,
			},

			"runtime": {
				Type:     schema.TypeString,
				Required: true,
			},

			"s3_bucket": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"s3_object": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"code": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"code_content_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"create_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"last_update_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"cloudwatch_logging_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
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

			// Flink only
			"checkpoint_configuration": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"checkpoint_interval": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60000,
						},
						"checkpointing_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"configuration_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  kinesisanalyticsv2.ConfigurationTypeDefault,
						},
						"min_pause_between_checkpoints": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  5000,
						},
					},
				},
			},
			// Flink only
			"monitoring_configuration": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configuration_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  kinesisanalyticsv2.ConfigurationTypeDefault,
						},
						"log_level": {
							Type:     schema.TypeString,
							Required: true,
						},
						"metrics_level": {
							Type:     schema.TypeString,
							Required: true,
							// TODO add validation
							// that this is one of
							// "APPLICATION",
							// "TASK", "OPERATOR",
							// or "PARALLELISM"
						},
					},
				},
			},
			// Flink only
			"parallelism_configuration": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_scaling": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"configuration_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  kinesisanalyticsv2.ConfigurationTypeDefault,
						},
						"parallelism": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"parallelism_per_kpu": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},

			// SQL only
			"inputs": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"kinesis_firehose": {
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

									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},

						"kinesis_stream": {
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

									"role_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},

						"name_prefix": {
							Type:     schema.TypeString,
							Required: true,
						},

						"parallelism": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"count": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},

						"processing_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"lambda": {
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

												"role_arn": {
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

						"schema": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"record_columns": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mapping": {
													Type:     schema.TypeString,
													Optional: true,
												},

												"name": {
													Type:     schema.TypeString,
													Required: true,
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
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"csv": {
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
															},

															"json": {
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
															},
														},
													},
												},

												"record_format_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},

						"starting_position_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"starting_position": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"stream_names": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"outputs": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"kinesis_firehose": {
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

									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},

						"kinesis_stream": {
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

									"role_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},

						"lambda": {
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

									"role_arn": {
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
						},

						"schema": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"record_format_type": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											kinesisanalyticsv2.RecordFormatTypeCsv,
											kinesisanalyticsv2.RecordFormatTypeJson,
										}, false),
									},
								},
							},
						},
					},
				},
			},

			"reference_data_sources": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"s3": {
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

									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},

						"schema": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"record_columns": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mapping": {
													Type:     schema.TypeString,
													Optional: true,
												},

												"name": {
													Type:     schema.TypeString,
													Required: true,
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
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"csv": {
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
															},

															"json": {
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
															},
														},
													},
												},

												"record_format_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},

						"table_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsKinesisAnalyticsApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn
	name := d.Get("name").(string)
	serviceExecutionRole := d.Get("service_execution_role").(string)
	runtime := d.Get("runtime").(string)
	s3Bucket := d.Get("s3_bucket").(string)
	s3Object := d.Get("s3_object").(string)
	codeContentType := d.Get("code_content_type").(string)

	var sqlApplicationConfiguration *kinesisanalyticsv2.SqlApplicationConfiguration
	var flinkApplicationConfiguration *kinesisanalyticsv2.FlinkApplicationConfiguration
	switch {
	case strings.HasPrefix(runtime, "SQL"):
		sqlApplicationConfiguration = &kinesisanalyticsv2.SqlApplicationConfiguration{}
		if v, ok := d.GetOk("inputs"); ok {
			i := v.([]interface{})[0].(map[string]interface{})
			inputs := expandKinesisAnalyticsInputs(i)
			sqlApplicationConfiguration.Inputs = []*kinesisanalyticsv2.Input{inputs}
		}
		if v := d.Get("outputs").([]interface{}); len(v) > 0 {
			outputs := make([]*kinesisanalyticsv2.Output, 0)
			for _, o := range v {
				output := expandKinesisAnalyticsOutputs(o.(map[string]interface{}))
				outputs = append(outputs, output)
			}
			sqlApplicationConfiguration.Outputs = outputs
		}
	case strings.HasPrefix(runtime, "FLINK"):
		flinkApplicationConfiguration = &kinesisanalyticsv2.FlinkApplicationConfiguration{}
		if v, ok := d.GetOk("checkpoint_configuration"); ok {
			m := v.(map[string]interface{})
			flinkApplicationConfiguration.CheckpointConfiguration = expandCheckpointConfiguration(m)
		}
		if v, ok := d.GetOk("monitoring_configuration"); ok {
			m := v.(map[string]interface{})
			flinkApplicationConfiguration.MonitoringConfiguration = expandMonitoringConfiguration(m)
		}
	}

	var contentType *string
	switch codeContentType {
	case "zip":
		contentType = aws.String(kinesisanalyticsv2.CodeContentTypeZipfile)
	case "plain_text":
		contentType = aws.String(kinesisanalyticsv2.CodeContentTypePlaintext)
	}

	var s3ContentLocation *kinesisanalyticsv2.S3ContentLocation
	if s3Bucket != "" && s3Object != "" {
		s3ContentLocation = &kinesisanalyticsv2.S3ContentLocation{
			BucketARN: aws.String(s3Bucket),
			FileKey:   aws.String(s3Object),
		}
	}

	createOpts := &kinesisanalyticsv2.CreateApplicationInput{
		RuntimeEnvironment:   aws.String(runtime),
		ApplicationName:      aws.String(name),
		ServiceExecutionRole: aws.String(serviceExecutionRole),
		ApplicationConfiguration: &kinesisanalyticsv2.ApplicationConfiguration{
			SqlApplicationConfiguration:   sqlApplicationConfiguration,
			FlinkApplicationConfiguration: flinkApplicationConfiguration,
			ApplicationCodeConfiguration: &kinesisanalyticsv2.ApplicationCodeConfiguration{
				CodeContent: &kinesisanalyticsv2.CodeContent{
					S3ContentLocation: s3ContentLocation,
				},
				CodeContentType: contentType,
			},
		},
	}

	if v, ok := d.GetOk("code"); ok && v.(string) != "" {
		createOpts.ApplicationConfiguration.ApplicationCodeConfiguration.CodeContent.TextContent = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
		clo := v.([]interface{})[0].(map[string]interface{})
		cloudwatchLoggingOption := expandKinesisAnalyticsCloudwatchLoggingOption(clo)
		createOpts.CloudWatchLoggingOptions = []*kinesisanalyticsv2.CloudWatchLoggingOption{cloudwatchLoggingOption}
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		createOpts.Tags = keyvaluetags.New(v).IgnoreAws().Kinesisanalyticsv2Tags()
	}

	// Retry for IAM eventual consistency
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		output, err := conn.CreateApplication(createOpts)
		if err != nil {
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
			return resource.NonRetryableError(err)
		}
		d.SetId(aws.StringValue(output.ApplicationDetail.ApplicationARN))
		return nil
	})

	if isResourceTimeoutError(err) {
		var output *kinesisanalyticsv2.CreateApplicationOutput
		output, err = conn.CreateApplication(createOpts)
		d.SetId(aws.StringValue(output.ApplicationDetail.ApplicationARN))
	}

	if err != nil {
		return fmt.Errorf("Unable to create Kinesis Analytics application: %s", err)
	}

	return resourceAwsKinesisAnalyticsApplicationUpdate(d, meta)
}

func resourceAwsKinesisAnalyticsApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn
	name := d.Get("name").(string)

	describeOpts := &kinesisanalyticsv2.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}
	resp, err := conn.DescribeApplication(describeOpts)
	if isAWSErr(err, kinesisanalyticsv2.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Kinesis Analytics Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Kinesis Analytics Application (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.ApplicationDetail.ApplicationARN)
	d.Set("name", aws.StringValue(resp.ApplicationDetail.ApplicationName))
	d.Set("arn", arn)
	d.Set("code", aws.StringValue(resp.ApplicationDetail.ApplicationConfigurationDescription.ApplicationCodeConfigurationDescription.CodeContentDescription.TextContent))
	d.Set("create_timestamp", aws.TimeValue(resp.ApplicationDetail.CreateTimestamp).Format(time.RFC3339))
	d.Set("description", aws.StringValue(resp.ApplicationDetail.ApplicationDescription))
	d.Set("last_update_timestamp", aws.TimeValue(resp.ApplicationDetail.LastUpdateTimestamp).Format(time.RFC3339))
	d.Set("status", aws.StringValue(resp.ApplicationDetail.ApplicationStatus))
	d.Set("version", int(aws.Int64Value(resp.ApplicationDetail.ApplicationVersionId)))

	if err := d.Set("cloudwatch_logging_options", flattenKinesisAnalyticsCloudwatchLoggingOptions(resp.ApplicationDetail.CloudWatchLoggingOptionDescriptions)); err != nil {
		return fmt.Errorf("error setting cloudwatch_logging_options: %s", err)
	}

	runtime := aws.StringValue(resp.ApplicationDetail.RuntimeEnvironment)
	if runtime == kinesisanalyticsv2.RuntimeEnvironmentSql10 {
		if err := d.Set("inputs", flattenKinesisAnalyticsInputs(resp.ApplicationDetail.ApplicationConfigurationDescription.SqlApplicationConfigurationDescription.InputDescriptions)); err != nil {
			return fmt.Errorf("error setting inputs: %s", err)
		}

		if err := d.Set("outputs", flattenKinesisAnalyticsOutputs(resp.ApplicationDetail.ApplicationConfigurationDescription.SqlApplicationConfigurationDescription.OutputDescriptions)); err != nil {
			return fmt.Errorf("error setting outputs: %s", err)
		}

		if err := d.Set("reference_data_sources", flattenKinesisAnalyticsReferenceDataSources(resp.ApplicationDetail.ApplicationConfigurationDescription.SqlApplicationConfigurationDescription.ReferenceDataSourceDescriptions)); err != nil {
			return fmt.Errorf("error setting reference_data_sources: %s", err)
		}
	}
	if runtime == kinesisanalyticsv2.RuntimeEnvironmentFlink16 ||
		runtime == kinesisanalyticsv2.RuntimeEnvironmentFlink18 {

		d.Set("checkpoint_configuration", map[string]interface{}{
			"checkpoint_interval":           resp.ApplicationDetail.ApplicationConfigurationDescription.FlinkApplicationConfigurationDescription.CheckpointConfigurationDescription.CheckpointInterval,
			"checkpointing_enabled":         resp.ApplicationDetail.ApplicationConfigurationDescription.FlinkApplicationConfigurationDescription.CheckpointConfigurationDescription.CheckpointingEnabled,
			"configuration_type":            resp.ApplicationDetail.ApplicationConfigurationDescription.FlinkApplicationConfigurationDescription.CheckpointConfigurationDescription.ConfigurationType,
			"min_pause_between_checkpoints": resp.ApplicationDetail.ApplicationConfigurationDescription.FlinkApplicationConfigurationDescription.CheckpointConfigurationDescription.MinPauseBetweenCheckpoints,
		})
	}

	tags, err := keyvaluetags.Kinesisanalyticsv2ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Kinesis Analytics Application (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsKinesisAnalyticsApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	var version int
	conn := meta.(*AWSClient).kinesisanalyticsv2conn
	name := d.Get("name").(string)

	if v, ok := d.GetOk("version"); ok {
		version = v.(int)
	} else {
		version = 1
	}

	if !d.IsNewResource() {
		updateApplicationOpts := &kinesisanalyticsv2.UpdateApplicationInput{
			ApplicationName:             aws.String(name),
			CurrentApplicationVersionId: aws.Int64(int64(version)),
		}

		applicationUpdate, err := createApplicationUpdateOpts(d)
		if err != nil {
			return err
		}

		if !reflect.DeepEqual(applicationUpdate, &kinesisanalyticsv2.UpdateApplicationInput{}) {
			updateApplicationOpts.SetApplicationConfigurationUpdate(applicationUpdate.ApplicationConfigurationUpdate)
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
				cloudwatchLoggingOption := expandKinesisAnalyticsCloudwatchLoggingOption(clo)
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

		if d.Get("runtime") == kinesisanalyticsv2.RuntimeEnvironmentSql10 {
			oldInputs, newInputs := d.GetChange("inputs")
			if len(oldInputs.([]interface{})) == 0 && len(newInputs.([]interface{})) > 0 {
				if v, ok := d.GetOk("inputs"); ok {
					i := v.([]interface{})[0].(map[string]interface{})
					input := expandKinesisAnalyticsInputs(i)
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
			}

			oldOutputs, newOutputs := d.GetChange("outputs")
			if len(oldOutputs.([]interface{})) == 0 && len(newOutputs.([]interface{})) > 0 {
				if v, ok := d.GetOk("outputs"); ok {
					o := v.([]interface{})[0].(map[string]interface{})
					output := expandKinesisAnalyticsOutputs(o)
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
					version = version + 1
				}
			}
			oldReferenceData, newReferenceData := d.GetChange("reference_data_sources")
			if len(oldReferenceData.([]interface{})) == 0 && len(newReferenceData.([]interface{})) > 0 {
				if v := d.Get("reference_data_sources").([]interface{}); len(v) > 0 {
					for _, r := range v {
						rd := r.(map[string]interface{})
						referenceData := expandKinesisAnalyticsReferenceData(rd)
						addOpts := &kinesisanalyticsv2.AddApplicationReferenceDataSourceInput{
							ApplicationName:             aws.String(name),
							CurrentApplicationVersionId: aws.Int64(int64(version)),
							ReferenceDataSource:         referenceData,
						}
						// Retry for IAM eventual consistency
						err := resource.Retry(1*time.Minute, func() *resource.RetryError {
							_, err := conn.AddApplicationReferenceDataSource(addOpts)
							if err != nil {
								if isAWSErr(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
									return resource.RetryableError(err)
								}
								return resource.NonRetryableError(err)
							}
							return nil
						})
						if isResourceTimeoutError(err) {
							_, err = conn.AddApplicationReferenceDataSource(addOpts)
						}
						if err != nil {
							return fmt.Errorf("Unable to add application reference data source: %s", err)
						}
						version = version + 1
					}
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
	}

	return resourceAwsKinesisAnalyticsApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn
	name := d.Get("name").(string)
	createTimestamp, parseErr := time.Parse(time.RFC3339, d.Get("create_timestamp").(string))
	if parseErr != nil {
		return parseErr
	}

	log.Printf("[DEBUG] Kinesis Analytics Application destroy: %v", d.Id())
	deleteOpts := &kinesisanalyticsv2.DeleteApplicationInput{
		ApplicationName: aws.String(name),
		CreateTimestamp: aws.Time(createTimestamp),
	}
	_, deleteErr := conn.DeleteApplication(deleteOpts)
	if isAWSErr(deleteErr, kinesisanalyticsv2.ErrCodeResourceNotFoundException, "") {
		return nil
	}
	deleteErr = waitForDeleteKinesisAnalyticsApplication(conn, d.Id(), d.Timeout(schema.TimeoutDelete))
	if deleteErr != nil {
		return fmt.Errorf("error waiting for deletion of Kinesis Analytics Application (%s): %s", d.Id(), deleteErr)
	}

	log.Printf("[DEBUG] Kinesis Analytics Application deleted: %v", d.Id())
	return nil
}

func expandKinesisAnalyticsCloudwatchLoggingOption(clo map[string]interface{}) *kinesisanalyticsv2.CloudWatchLoggingOption {
	cloudwatchLoggingOption := &kinesisanalyticsv2.CloudWatchLoggingOption{
		LogStreamARN: aws.String(clo["log_stream_arn"].(string)),
	}
	return cloudwatchLoggingOption
}

func expandKinesisAnalyticsInputs(i map[string]interface{}) *kinesisanalyticsv2.Input {
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
		ss := expandKinesisAnalyticsSourceSchema(vL)
		input.InputSchema = ss
	}

	return input
}

func expandKinesisAnalyticsSourceSchema(vL map[string]interface{}) *kinesisanalyticsv2.SourceSchema {
	ss := &kinesisanalyticsv2.SourceSchema{}
	if v := vL["record_columns"].([]interface{}); len(v) > 0 {
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

func expandKinesisAnalyticsOutputs(o map[string]interface{}) *kinesisanalyticsv2.Output {
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

func expandKinesisAnalyticsReferenceData(rd map[string]interface{}) *kinesisanalyticsv2.ReferenceDataSource {
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
		ss := expandKinesisAnalyticsSourceSchema(v[0].(map[string]interface{}))
		referenceData.ReferenceSchema = ss
	}

	return referenceData
}

func createApplicationUpdateOpts(d *schema.ResourceData) (*kinesisanalyticsv2.UpdateApplicationInput, error) {
	applicationUpdate := &kinesisanalyticsv2.UpdateApplicationInput{
		ApplicationConfigurationUpdate: &kinesisanalyticsv2.ApplicationConfigurationUpdate{},
	}

	if d.HasChange("code") {
		if v, ok := d.GetOk("code"); ok && v.(string) != "" {
			applicationUpdate.ApplicationConfigurationUpdate = &kinesisanalyticsv2.ApplicationConfigurationUpdate{
				ApplicationCodeConfigurationUpdate: &kinesisanalyticsv2.ApplicationCodeConfigurationUpdate{
					CodeContentUpdate: &kinesisanalyticsv2.CodeContentUpdate{
						TextContentUpdate: aws.String(v.(string)),
					},
				},
			}
		}
	}

	oldLoggingOptions, newLoggingOptions := d.GetChange("cloudwatch_logging_options")
	if len(oldLoggingOptions.([]interface{})) > 0 && len(newLoggingOptions.([]interface{})) > 0 {
		if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
			clo := v.([]interface{})[0].(map[string]interface{})
			cloudwatchLoggingOption := expandKinesisAnalyticsCloudwatchLoggingOptionUpdate(clo)
			applicationUpdate.CloudWatchLoggingOptionUpdates = []*kinesisanalyticsv2.CloudWatchLoggingOptionUpdate{cloudwatchLoggingOption}
		}
	}

	runtime := d.Get("runtime").(string)
	if runtime == kinesisanalyticsv2.RuntimeEnvironmentSql10 {
		sqlUpdate := &kinesisanalyticsv2.SqlApplicationConfigurationUpdate{}

		oldInputs, newInputs := d.GetChange("inputs")
		if len(oldInputs.([]interface{})) > 0 && len(newInputs.([]interface{})) > 0 {
			if v, ok := d.GetOk("inputs"); ok {
				vL := v.([]interface{})[0].(map[string]interface{})
				inputUpdate := expandKinesisAnalyticsInputUpdate(vL)
				sqlUpdate.InputUpdates = []*kinesisanalyticsv2.InputUpdate{inputUpdate}
			}
		}

		oldOutputs, newOutputs := d.GetChange("outputs")
		if len(oldOutputs.([]interface{})) > 0 && len(newOutputs.([]interface{})) > 0 {
			if v, ok := d.GetOk("outputs"); ok {
				vL := v.([]interface{})[0].(map[string]interface{})
				outputUpdate := expandKinesisAnalyticsOutputUpdate(vL)
				sqlUpdate.OutputUpdates = []*kinesisanalyticsv2.OutputUpdate{outputUpdate}
			}
		}

		oldReferenceData, newReferenceData := d.GetChange("reference_data_sources")
		if len(oldReferenceData.([]interface{})) > 0 && len(newReferenceData.([]interface{})) > 0 {
			if v := d.Get("reference_data_sources").([]interface{}); len(v) > 0 {
				var rdsus []*kinesisanalyticsv2.ReferenceDataSourceUpdate
				for _, rd := range v {
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
						ss := expandKinesisAnalyticsSourceSchema(vL)
						rdsu.ReferenceSchemaUpdate = ss
					}

					rdsus = append(rdsus, rdsu)
				}
				sqlUpdate.ReferenceDataSourceUpdates = rdsus
			}
		}
		applicationUpdate.ApplicationConfigurationUpdate.SqlApplicationConfigurationUpdate = sqlUpdate
	}
	if runtime == kinesisanalyticsv2.RuntimeEnvironmentFlink16 ||
		runtime == kinesisanalyticsv2.RuntimeEnvironmentFlink18 {

		flinkUpdate := &kinesisanalyticsv2.FlinkApplicationConfigurationUpdate{}

		oldConfig, newConfig := d.GetChange("checkpoint_configuration")
		if len(oldConfig.(map[string]interface{})) > 0 && len(newConfig.(map[string]interface{})) > 0 {
			if v := d.Get("checkpoint_configuration").(map[string]interface{}); len(v) > 0 {
				checkpointConfig := expandCheckpointConfiguration(v)
				flinkUpdate.CheckpointConfigurationUpdate = &kinesisanalyticsv2.CheckpointConfigurationUpdate{
					CheckpointIntervalUpdate:         checkpointConfig.CheckpointInterval,
					CheckpointingEnabledUpdate:       checkpointConfig.CheckpointingEnabled,
					ConfigurationTypeUpdate:          checkpointConfig.ConfigurationType,
					MinPauseBetweenCheckpointsUpdate: checkpointConfig.MinPauseBetweenCheckpoints,
				}
				monitoringConfig := expandMonitoringConfiguration(v)
				flinkUpdate.MonitoringConfigurationUpdate = &kinesisanalyticsv2.MonitoringConfigurationUpdate{
					ConfigurationTypeUpdate: monitoringConfig.ConfigurationType,
					LogLevelUpdate:          monitoringConfig.LogLevel,
					MetricsLevelUpdate:      monitoringConfig.MetricsLevel,
				}
				parallelismConfig := expandParallelismConfiguration(v)
				flinkUpdate.ParallelismConfigurationUpdate = &kinesisanalyticsv2.ParallelismConfigurationUpdate{
					AutoScalingEnabledUpdate: parallelismConfig.AutoScalingEnabled,
					ConfigurationTypeUpdate:  parallelismConfig.ConfigurationType,
					ParallelismUpdate:        parallelismConfig.Parallelism,
					ParallelismPerKPUUpdate:  parallelismConfig.ParallelismPerKPU,
				}
			}
		}
		applicationUpdate.ApplicationConfigurationUpdate.FlinkApplicationConfigurationUpdate = flinkUpdate
	}

	return applicationUpdate, nil
}

func expandParallelismConfiguration(v map[string]interface{}) *kinesisanalyticsv2.ParallelismConfiguration {
	var autoscalingEnabled *bool
	var configurationType *string
	var parallelism *int64
	var parallelismPerKPU *int64

	if aEnabled, ok := v["autoscaling_enabled"]; ok {
		v, _ := strconv.ParseBool(aEnabled.(string))
		autoscalingEnabled = aws.Bool(v)
	}
	if confType, ok := v["configuration_type"]; ok {
		configurationType = aws.String(confType.(string))
	}
	if p, ok := v["parallelism"]; ok {
		parallelism = aws.Int64(p.(int64))
	}
	if p, ok := v["parallelism_per_kpu"]; ok {
		parallelismPerKPU = aws.Int64(p.(int64))
	}
	return &kinesisanalyticsv2.ParallelismConfiguration{
		AutoScalingEnabled: autoscalingEnabled,
		ConfigurationType:  configurationType,
		Parallelism:        parallelism,
		ParallelismPerKPU:  parallelismPerKPU,
	}
}

func expandMonitoringConfiguration(v map[string]interface{}) *kinesisanalyticsv2.MonitoringConfiguration {
	var configurationType *string
	var logLevel *string
	var metricsLevel *string

	if confType, ok := v["configuration_type"]; ok {
		configurationType = aws.String(confType.(string))
	}
	if level, ok := v["log_level"]; ok {
		logLevel = aws.String(level.(string))
	}
	if level, ok := v["metrics_level"]; ok {
		metricsLevel = aws.String(level.(string))
	}
	return &kinesisanalyticsv2.MonitoringConfiguration{
		ConfigurationType: configurationType,
		LogLevel:          logLevel,
		MetricsLevel:      metricsLevel,
	}
}

func expandCheckpointConfiguration(v map[string]interface{}) *kinesisanalyticsv2.CheckpointConfiguration {
	var checkpointingEnabled *bool
	var checkpointInterval *int64
	var configurationType *string
	var configurationMinPause *int64

	if interval, ok := v["checkpoint_interval"]; ok {
		checkpointInterval = aws.Int64(interval.(int64))
	}
	if enabled, ok := v["checkpointing_enabled"]; ok {
		v, _ := strconv.ParseBool(enabled.(string))
		checkpointingEnabled = aws.Bool(v)
	}
	if confType, ok := v["configuration_type"]; ok {
		configurationType = aws.String(confType.(string))
	}
	if minPause, ok := v["min_pause_between_checkpoints"]; ok {
		configurationMinPause = aws.Int64(minPause.(int64))
	}
	return &kinesisanalyticsv2.CheckpointConfiguration{
		CheckpointInterval:         checkpointInterval,
		CheckpointingEnabled:       checkpointingEnabled,
		ConfigurationType:          configurationType,
		MinPauseBetweenCheckpoints: configurationMinPause,
	}
}

func expandKinesisAnalyticsInputUpdate(vL map[string]interface{}) *kinesisanalyticsv2.InputUpdate {
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

		if v := vL["record_columns"].([]interface{}); len(v) > 0 {
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

func expandKinesisAnalyticsOutputUpdate(vL map[string]interface{}) *kinesisanalyticsv2.OutputUpdate {
	outputUpdate := &kinesisanalyticsv2.OutputUpdate{
		OutputId:   aws.String(vL["id"].(string)),
		NameUpdate: aws.String(vL["name"].(string)),
	}

	if v := vL["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfou := &kinesisanalyticsv2.KinesisFirehoseOutputUpdate{
			ResourceARNUpdate: aws.String(kf["resource_arn"].(string)),
		}
		outputUpdate.KinesisFirehoseOutputUpdate = kfou
	}

	if v := vL["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		ksou := &kinesisanalyticsv2.KinesisStreamsOutputUpdate{
			ResourceARNUpdate: aws.String(ks["resource_arn"].(string)),
		}
		outputUpdate.KinesisStreamsOutputUpdate = ksou
	}

	if v := vL["lambda"].([]interface{}); len(v) > 0 {
		l := v[0].(map[string]interface{})
		lou := &kinesisanalyticsv2.LambdaOutputUpdate{
			ResourceARNUpdate: aws.String(l["resource_arn"].(string)),
		}
		outputUpdate.LambdaOutputUpdate = lou
	}

	if v := vL["schema"].([]interface{}); len(v) > 0 {
		ds := v[0].(map[string]interface{})
		dsu := &kinesisanalyticsv2.DestinationSchema{
			RecordFormatType: aws.String(ds["record_format_type"].(string)),
		}
		outputUpdate.DestinationSchemaUpdate = dsu
	}

	return outputUpdate
}

func expandKinesisAnalyticsCloudwatchLoggingOptionUpdate(clo map[string]interface{}) *kinesisanalyticsv2.CloudWatchLoggingOptionUpdate {
	cloudwatchLoggingOption := &kinesisanalyticsv2.CloudWatchLoggingOptionUpdate{
		CloudWatchLoggingOptionId: aws.String(clo["id"].(string)),
		LogStreamARNUpdate:        aws.String(clo["log_stream_arn"].(string)),
	}
	return cloudwatchLoggingOption
}

func flattenKinesisAnalyticsCloudwatchLoggingOptions(options []*kinesisanalyticsv2.CloudWatchLoggingOptionDescription) []interface{} {
	s := []interface{}{}
	for _, v := range options {
		option := map[string]interface{}{
			"id":             aws.StringValue(v.CloudWatchLoggingOptionId),
			"log_stream_arn": aws.StringValue(v.LogStreamARN),
		}
		s = append(s, option)
	}
	return s
}

func flattenKinesisAnalyticsInputs(inputs []*kinesisanalyticsv2.InputDescription) []interface{} {
	s := []interface{}{}

	if len(inputs) > 0 {
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
								"role_arn":     aws.StringValue(ipcd.InputLambdaProcessorDescription.RoleARN),
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
			ss["record_columns"] = rcs

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
					"role_arn":     aws.StringValue(id.KinesisFirehoseInputDescription.RoleARN),
				},
			}
		}

		if id.KinesisStreamsInputDescription != nil {
			input["kinesis_stream"] = []interface{}{
				map[string]interface{}{
					"resource_arn": aws.StringValue(id.KinesisStreamsInputDescription.ResourceARN),
					"role_arn":     aws.StringValue(id.KinesisStreamsInputDescription.RoleARN),
				},
			}
		}

		s = append(s, input)
	}
	return s
}

func flattenKinesisAnalyticsOutputs(outputs []*kinesisanalyticsv2.OutputDescription) []interface{} {
	s := []interface{}{}

	for _, o := range outputs {
		output := map[string]interface{}{
			"id":   aws.StringValue(o.OutputId),
			"name": aws.StringValue(o.Name),
		}

		if o.KinesisFirehoseOutputDescription != nil {
			output["kinesis_firehose"] = []interface{}{
				map[string]interface{}{
					"resource_arn": aws.StringValue(o.KinesisFirehoseOutputDescription.ResourceARN),
					"role_arn":     aws.StringValue(o.KinesisFirehoseOutputDescription.RoleARN),
				},
			}
		}

		if o.KinesisStreamsOutputDescription != nil {
			output["kinesis_stream"] = []interface{}{
				map[string]interface{}{
					"resource_arn": aws.StringValue(o.KinesisStreamsOutputDescription.ResourceARN),
					"role_arn":     aws.StringValue(o.KinesisStreamsOutputDescription.RoleARN),
				},
			}
		}

		if o.LambdaOutputDescription != nil {
			output["lambda"] = []interface{}{
				map[string]interface{}{
					"resource_arn": aws.StringValue(o.LambdaOutputDescription.ResourceARN),
					"role_arn":     aws.StringValue(o.LambdaOutputDescription.RoleARN),
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

	return s
}

func flattenKinesisAnalyticsReferenceDataSources(dataSources []*kinesisanalyticsv2.ReferenceDataSourceDescription) []interface{} {
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
						"role_arn":   aws.StringValue(ds.S3ReferenceDataSourceDescription.ReferenceRoleARN),
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
				ss["record_columns"] = rcs

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

func waitForDeleteKinesisAnalyticsApplication(conn *kinesisanalyticsv2.KinesisAnalyticsV2, applicationId string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			kinesisanalyticsv2.ApplicationStatusRunning,
			kinesisanalyticsv2.ApplicationStatusDeleting,
		},
		Target:  []string{""},
		Timeout: timeout,
		Refresh: refreshKinesisAnalyticsApplicationStatus(conn, applicationId),
	}
	application, err := stateConf.WaitForState()
	if err != nil {
		if isAWSErr(err, kinesisanalyticsv2.ErrCodeResourceNotFoundException, "") {
			return nil
		}
	}
	if application == nil {
		return nil
	}
	return err
}

func refreshKinesisAnalyticsApplicationStatus(conn *kinesisanalyticsv2.KinesisAnalyticsV2, applicationId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeApplication(&kinesisanalyticsv2.DescribeApplicationInput{
			ApplicationName: aws.String(applicationId),
		})
		if err != nil {
			return nil, "", err
		}
		application := output.ApplicationDetail
		if application == nil {
			return application, "", fmt.Errorf("Kinesis Analytics Application (%s) could not be found.", applicationId)
		}
		return application, aws.StringValue(application.ApplicationStatus), nil
	}
}
