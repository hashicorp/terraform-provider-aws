package aws

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalytics/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalytics/waiter"
)

func resourceAwsKinesisAnalyticsApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisAnalyticsApplicationCreate,
		Read:   resourceAwsKinesisAnalyticsApplicationRead,
		Update: resourceAwsKinesisAnalyticsApplicationUpdate,
		Delete: resourceAwsKinesisAnalyticsApplicationDelete,

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("inputs", func(_ context.Context, old, new, meta interface{}) bool {
				// An existing input configuration cannot be deleted.
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
		),

		Importer: &schema.ResourceImporter{
			State: resourceAwsKinesisAnalyticsApplicationImport,
		},

		Schema: map[string]*schema.Schema{
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
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"log_stream_arn": {
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

			"code": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 102400),
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
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
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

						"parallelism": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"count": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 64),
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
																ExactlyOneOf: []string{
																	"inputs.0.schema.0.record_format.0.mapping_parameters.0.csv",
																	"inputs.0.schema.0.record_format.0.mapping_parameters.0.json",
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
																ExactlyOneOf: []string{
																	"inputs.0.schema.0.record_format.0.mapping_parameters.0.csv",
																	"inputs.0.schema.0.record_format.0.mapping_parameters.0.json",
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
				Type:     schema.TypeSet,
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
										Required:     true,
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
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 32),
								validation.StringMatch(regexp.MustCompile(`^[^-\s<>&]+$`), "must not include hyphen, whitespace, angle bracket, or ampersand characters"),
							),
						},

						"schema": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"record_format_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(kinesisanalytics.RecordFormatType_Values(), false),
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
																ExactlyOneOf: []string{
																	"reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv",
																	"reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json",
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
																ExactlyOneOf: []string{
																	"reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv",
																	"reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json",
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 32),
						},
					},
				},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),

			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsKinesisAnalyticsApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn

	input := &kinesisanalytics.CreateApplicationInput{
		ApplicationCode:          aws.String(d.Get("code").(string)),
		ApplicationDescription:   aws.String(d.Get("description").(string)),
		ApplicationName:          aws.String(d.Get("name").(string)),
		CloudWatchLoggingOptions: expandKinesisAnalyticsV1CloudWatchLoggingOptions(d.Get("cloudwatch_logging_options").([]interface{})),
		Inputs:                   expandKinesisAnalyticsV1Inputs(d.Get("inputs").([]interface{})),
		Outputs:                  expandKinesisAnalyticsV1Outputs(d.Get("outputs").(*schema.Set).List()),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().KinesisanalyticsTags()
	}

	log.Printf("[DEBUG] Creating Kinesis Analytics Application: %s", input)

	outputRaw, err := waiter.IAMPropagation(func() (interface{}, error) {
		return conn.CreateApplication(input)
	})

	if err != nil {
		return fmt.Errorf("error creating Kinesis Analytics Application: %w", err)
	}

	applicationSummary := outputRaw.(*kinesisanalytics.CreateApplicationOutput).ApplicationSummary

	d.SetId(aws.StringValue(applicationSummary.ApplicationARN))

	if v := d.Get("reference_data_sources").([]interface{}); len(v) > 0 && v[0] != nil {
		// Add new reference data source.
		input := &kinesisanalytics.AddApplicationReferenceDataSourceInput{
			ApplicationName:             applicationSummary.ApplicationName,
			CurrentApplicationVersionId: aws.Int64(1), // Newly created application version.
			ReferenceDataSource:         expandKinesisAnalyticsV1ReferenceDataSource(v),
		}

		log.Printf("[DEBUG] Adding Kinesis Analytics Application (%s) reference data source: %s", d.Id(), input)

		_, err := waiter.IAMPropagation(func() (interface{}, error) {
			return conn.AddApplicationReferenceDataSource(input)
		})

		if err != nil {
			return fmt.Errorf("error adding Kinesis Analytics Application (%s) reference data source: %w", d.Id(), err)
		}
	}

	return resourceAwsKinesisAnalyticsApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	application, err := finder.ApplicationByName(conn, d.Get("name").(string))

	if isAWSErr(err, kinesisanalytics.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Kinesis Analytics Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Kinesis Analytics Application (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(application.ApplicationARN)
	d.Set("arn", arn)
	d.Set("code", application.ApplicationCode)
	d.Set("create_timestamp", aws.TimeValue(application.CreateTimestamp).Format(time.RFC3339))
	d.Set("description", application.ApplicationDescription)
	d.Set("last_update_timestamp", aws.TimeValue(application.LastUpdateTimestamp).Format(time.RFC3339))
	d.Set("name", application.ApplicationName)
	d.Set("status", application.ApplicationStatus)
	d.Set("version", int(aws.Int64Value(application.ApplicationVersionId)))

	if err := d.Set("cloudwatch_logging_options", flattenKinesisAnalyticsCloudwatchLoggingOptions(application.CloudWatchLoggingOptionDescriptions)); err != nil {
		return fmt.Errorf("error setting cloudwatch_logging_options: %w", err)
	}

	if err := d.Set("inputs", flattenKinesisAnalyticsInputs(application.InputDescriptions)); err != nil {
		return fmt.Errorf("error setting inputs: %w", err)
	}

	if err := d.Set("outputs", flattenKinesisAnalyticsOutputs(application.OutputDescriptions)); err != nil {
		return fmt.Errorf("error setting outputs: %w", err)
	}

	if err := d.Set("reference_data_sources", flattenKinesisAnalyticsReferenceDataSources(application.ReferenceDataSourceDescriptions)); err != nil {
		return fmt.Errorf("error setting reference_data_sources: %w", err)
	}

	tags, err := keyvaluetags.KinesisanalyticsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Kinesis Analytics Application (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsKinesisAnalyticsApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn

	if d.HasChanges("cloudwatch_logging_options", "code", "inputs", "outputs", "reference_data_sources") {
		applicationName := d.Get("name").(string)
		currentApplicationVersionId := int64(d.Get("version").(int))
		updateApplication := false

		input := &kinesisanalytics.UpdateApplicationInput{
			ApplicationName:   aws.String(applicationName),
			ApplicationUpdate: &kinesisanalytics.ApplicationUpdate{},
		}

		if d.HasChange("cloudwatch_logging_options") {
			o, n := d.GetChange("cloudwatch_logging_options")

			if len(o.([]interface{})) == 0 {
				// Add new CloudWatch logging options.
				mNewCloudWatchLoggingOption := n.([]interface{})[0].(map[string]interface{})

				input := &kinesisanalytics.AddApplicationCloudWatchLoggingOptionInput{
					ApplicationName: aws.String(applicationName),
					CloudWatchLoggingOption: &kinesisanalytics.CloudWatchLoggingOption{
						LogStreamARN: aws.String(mNewCloudWatchLoggingOption["log_stream_arn"].(string)),
						RoleARN:      aws.String(mNewCloudWatchLoggingOption["role_arn"].(string)),
					},
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionId),
				}

				log.Printf("[DEBUG] Adding Kinesis Analytics Application (%s) CloudWatch logging option: %s", d.Id(), input)

				_, err := waiter.IAMPropagation(func() (interface{}, error) {
					return conn.AddApplicationCloudWatchLoggingOption(input)
				})

				if err != nil {
					return fmt.Errorf("error adding Kinesis Analytics Application (%s) CloudWatch logging option: %w", d.Id(), err)
				}

				currentApplicationVersionId += 1
			} else if len(n.([]interface{})) == 0 {
				// Delete existing CloudWatch logging options.
				mOldCloudWatchLoggingOption := o.([]interface{})[0].(map[string]interface{})

				input := &kinesisanalytics.DeleteApplicationCloudWatchLoggingOptionInput{
					ApplicationName:             aws.String(applicationName),
					CloudWatchLoggingOptionId:   aws.String(mOldCloudWatchLoggingOption["id"].(string)),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionId),
				}

				log.Printf("[DEBUG] Deleting Kinesis Analytics Application (%s) CloudWatch logging option: %s", d.Id(), input)

				_, err := waiter.IAMPropagation(func() (interface{}, error) {
					return conn.DeleteApplicationCloudWatchLoggingOption(input)
				})

				if err != nil {
					return fmt.Errorf("error deleting Kinesis Analytics Application (%s) CloudWatch logging option: %w", d.Id(), err)
				}

				currentApplicationVersionId += 1
			} else {
				// Update existing CloudWatch logging options.
				mOldCloudWatchLoggingOption := o.([]interface{})[0].(map[string]interface{})
				mNewCloudWatchLoggingOption := n.([]interface{})[0].(map[string]interface{})

				input.ApplicationUpdate.CloudWatchLoggingOptionUpdates = []*kinesisanalytics.CloudWatchLoggingOptionUpdate{
					{
						CloudWatchLoggingOptionId: aws.String(mOldCloudWatchLoggingOption["id"].(string)),
						LogStreamARNUpdate:        aws.String(mNewCloudWatchLoggingOption["log_stream_arn"].(string)),
						RoleARNUpdate:             aws.String(mNewCloudWatchLoggingOption["role_arn"].(string)),
					},
				}

				updateApplication = true
			}
		}

		if d.HasChange("code") {
			input.ApplicationUpdate.ApplicationCodeUpdate = aws.String(d.Get("code").(string))

			updateApplication = true
		}

		if d.HasChange("inputs") {
		}

		if d.HasChange("outputs") {
		}

		if d.HasChange("reference_data_sources") {
			o, n := d.GetChange("reference_data_sources")

			if len(o.([]interface{})) == 0 {
				// Add new reference data source.
				input := &kinesisanalytics.AddApplicationReferenceDataSourceInput{
					ApplicationName:             aws.String(applicationName),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionId),
					ReferenceDataSource:         expandKinesisAnalyticsV1ReferenceDataSource(n.([]interface{})),
				}

				log.Printf("[DEBUG] Adding Kinesis Analytics Application (%s) reference data source: %s", d.Id(), input)

				_, err := waiter.IAMPropagation(func() (interface{}, error) {
					return conn.AddApplicationReferenceDataSource(input)
				})

				if err != nil {
					return fmt.Errorf("error adding Kinesis Analytics Application (%s) reference data source: %w", d.Id(), err)
				}

				currentApplicationVersionId += 1
			} else if len(n.([]interface{})) == 0 {
				// Delete existing reference data source.
				mOldReferenceDataSource := o.([]interface{})[0].(map[string]interface{})

				input := &kinesisanalytics.DeleteApplicationReferenceDataSourceInput{
					ApplicationName:             aws.String(applicationName),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionId),
					ReferenceId:                 aws.String(mOldReferenceDataSource["id"].(string)),
				}

				log.Printf("[DEBUG] Deleting Kinesis Analytics Application (%s) reference data source: %s", d.Id(), input)

				_, err := waiter.IAMPropagation(func() (interface{}, error) {
					return conn.DeleteApplicationReferenceDataSource(input)
				})

				if err != nil {
					return fmt.Errorf("error deleting Kinesis Analytics Application (%s) reference data source: %w", d.Id(), err)
				}

				currentApplicationVersionId += 1
			} else {
				// Update existing reference data source.
				referenceDataSourceUpdate := expandKinesisAnalyticsV1ReferenceDataSourceUpdate(n.([]interface{}))

				input.ApplicationUpdate.ReferenceDataSourceUpdates = []*kinesisanalytics.ReferenceDataSourceUpdate{referenceDataSourceUpdate}

				updateApplication = true
			}
		}

		if updateApplication {
			input.CurrentApplicationVersionId = aws.Int64(currentApplicationVersionId)

			log.Printf("[DEBUG] Updating Kinesis Analytics Application (%s): %s", d.Id(), input)

			_, err := waiter.IAMPropagation(func() (interface{}, error) {
				return conn.UpdateApplication(input)
			})

			if err != nil {
				return fmt.Errorf("error updating Kinesis Analytics Application (%s): %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags")
		if err := keyvaluetags.KinesisanalyticsUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Kinesis Analytics Application (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsKinesisAnalyticsApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsApplicationUpdateOld(d *schema.ResourceData, meta interface{}) error {
	var version int
	conn := meta.(*AWSClient).kinesisanalyticsconn
	name := d.Get("name").(string)

	if v, ok := d.GetOk("version"); ok {
		version = v.(int)
	} else {
		version = 1
	}

	if !d.IsNewResource() {
		updateApplicationOpts := &kinesisanalytics.UpdateApplicationInput{
			ApplicationName:             aws.String(name),
			CurrentApplicationVersionId: aws.Int64(int64(version)),
		}

		applicationUpdate := createApplicationUpdateOpts(d)

		if !reflect.DeepEqual(applicationUpdate, &kinesisanalytics.ApplicationUpdate{}) {
			updateApplicationOpts.SetApplicationUpdate(applicationUpdate)
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
				addOpts := &kinesisanalytics.AddApplicationCloudWatchLoggingOptionInput{
					ApplicationName:             aws.String(name),
					CurrentApplicationVersionId: aws.Int64(int64(version)),
					CloudWatchLoggingOption:     cloudwatchLoggingOption,
				}
				// Retry for IAM eventual consistency
				err := resource.Retry(1*time.Minute, func() *resource.RetryError {
					_, err := conn.AddApplicationCloudWatchLoggingOption(addOpts)
					if err != nil {
						if isAWSErr(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
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

		oldInputs, newInputs := d.GetChange("inputs")
		if len(oldInputs.([]interface{})) == 0 && len(newInputs.([]interface{})) > 0 {
			if v, ok := d.GetOk("inputs"); ok {
				i := v.([]interface{})[0].(map[string]interface{})
				input := expandKinesisAnalyticsInputs(i)
				addOpts := &kinesisanalytics.AddApplicationInputInput{
					ApplicationName:             aws.String(name),
					CurrentApplicationVersionId: aws.Int64(int64(version)),
					Input:                       input,
				}
				// Retry for IAM eventual consistency
				err := resource.Retry(1*time.Minute, func() *resource.RetryError {
					_, err := conn.AddApplicationInput(addOpts)
					if err != nil {
						if isAWSErr(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
							return resource.RetryableError(err)
						}
						// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
						if isAWSErr(err, kinesisanalytics.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
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
				addOpts := &kinesisanalytics.AddApplicationOutputInput{
					ApplicationName:             aws.String(name),
					CurrentApplicationVersionId: aws.Int64(int64(version)),
					Output:                      output,
				}
				// Retry for IAM eventual consistency
				err := resource.Retry(1*time.Minute, func() *resource.RetryError {
					_, err := conn.AddApplicationOutput(addOpts)
					if err != nil {
						if isAWSErr(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
							return resource.RetryableError(err)
						}
						// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
						if isAWSErr(err, kinesisanalytics.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
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

		arn := d.Get("arn").(string)
		if d.HasChange("tags") {
			o, n := d.GetChange("tags")

			if err := keyvaluetags.KinesisanalyticsUpdateTags(conn, arn, o, n); err != nil {
				return fmt.Errorf("error updating Kinesis Analytics Application (%s) tags: %s", arn, err)
			}
		}
	}

	oldReferenceData, newReferenceData := d.GetChange("reference_data_sources")
	if len(oldReferenceData.([]interface{})) == 0 && len(newReferenceData.([]interface{})) > 0 {
		if v := d.Get("reference_data_sources").([]interface{}); len(v) > 0 {
			for _, r := range v {
				rd := r.(map[string]interface{})
				referenceData := expandKinesisAnalyticsReferenceData(rd)
				addOpts := &kinesisanalytics.AddApplicationReferenceDataSourceInput{
					ApplicationName:             aws.String(name),
					CurrentApplicationVersionId: aws.Int64(int64(version)),
					ReferenceDataSource:         referenceData,
				}
				// Retry for IAM eventual consistency
				err := resource.Retry(1*time.Minute, func() *resource.RetryError {
					_, err := conn.AddApplicationReferenceDataSource(addOpts)
					if err != nil {
						if isAWSErr(err, kinesisanalytics.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
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

	return resourceAwsKinesisAnalyticsApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn

	createTimestamp, err := time.Parse(time.RFC3339, d.Get("create_timestamp").(string))
	if err != nil {
		return fmt.Errorf("error parsing create_timestamp: %w", err)
	}

	applicationName := d.Get("name").(string)

	log.Printf("[DEBUG] Deleting Kinesis Analytics Application (%s)", d.Id())
	_, err = conn.DeleteApplication(&kinesisanalytics.DeleteApplicationInput{
		ApplicationName: aws.String(applicationName),
		CreateTimestamp: aws.Time(createTimestamp),
	})

	if isAWSErr(err, kinesisanalytics.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Kinesis Analytics Application (%s): %w", d.Id(), err)
	}

	_, err = waiter.ApplicationDeleted(conn, applicationName, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for Kinesis Analytics Application (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsKinesisAnalyticsApplicationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func expandKinesisAnalyticsCloudwatchLoggingOption(clo map[string]interface{}) *kinesisanalytics.CloudWatchLoggingOption {
	cloudwatchLoggingOption := &kinesisanalytics.CloudWatchLoggingOption{
		LogStreamARN: aws.String(clo["log_stream_arn"].(string)),
		RoleARN:      aws.String(clo["role_arn"].(string)),
	}

	return cloudwatchLoggingOption
}

func expandKinesisAnalyticsInputs(i map[string]interface{}) *kinesisanalytics.Input {
	input := &kinesisanalytics.Input{
		NamePrefix: aws.String(i["name_prefix"].(string)),
	}

	if v := i["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfi := &kinesisanalytics.KinesisFirehoseInput{
			ResourceARN: aws.String(kf["resource_arn"].(string)),
			RoleARN:     aws.String(kf["role_arn"].(string)),
		}
		input.KinesisFirehoseInput = kfi
	}

	if v := i["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		ksi := &kinesisanalytics.KinesisStreamsInput{
			ResourceARN: aws.String(ks["resource_arn"].(string)),
			RoleARN:     aws.String(ks["role_arn"].(string)),
		}
		input.KinesisStreamsInput = ksi
	}

	if v := i["parallelism"].([]interface{}); len(v) > 0 {
		p := v[0].(map[string]interface{})

		if c, ok := p["count"]; ok {
			ip := &kinesisanalytics.InputParallelism{
				Count: aws.Int64(int64(c.(int))),
			}
			input.InputParallelism = ip
		}
	}

	if v := i["processing_configuration"].([]interface{}); len(v) > 0 {
		pc := v[0].(map[string]interface{})

		if l := pc["lambda"].([]interface{}); len(l) > 0 {
			lp := l[0].(map[string]interface{})
			ipc := &kinesisanalytics.InputProcessingConfiguration{
				InputLambdaProcessor: &kinesisanalytics.InputLambdaProcessor{
					ResourceARN: aws.String(lp["resource_arn"].(string)),
					RoleARN:     aws.String(lp["role_arn"].(string)),
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

func expandKinesisAnalyticsSourceSchema(vL map[string]interface{}) *kinesisanalytics.SourceSchema {
	ss := &kinesisanalytics.SourceSchema{}
	if v := vL["record_columns"].([]interface{}); len(v) > 0 {
		var rcs []*kinesisanalytics.RecordColumn

		for _, rc := range v {
			rcD := rc.(map[string]interface{})
			rc := &kinesisanalytics.RecordColumn{
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
		rf := &kinesisanalytics.RecordFormat{}

		if v := vL["mapping_parameters"].([]interface{}); len(v) > 0 {
			vL := v[0].(map[string]interface{})
			mp := &kinesisanalytics.MappingParameters{}

			if v := vL["csv"].([]interface{}); len(v) > 0 {
				cL := v[0].(map[string]interface{})
				cmp := &kinesisanalytics.CSVMappingParameters{
					RecordColumnDelimiter: aws.String(cL["record_column_delimiter"].(string)),
					RecordRowDelimiter:    aws.String(cL["record_row_delimiter"].(string)),
				}
				mp.CSVMappingParameters = cmp
				rf.RecordFormatType = aws.String("CSV")
			}

			if v := vL["json"].([]interface{}); len(v) > 0 {
				jL := v[0].(map[string]interface{})
				jmp := &kinesisanalytics.JSONMappingParameters{
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

func expandKinesisAnalyticsOutputs(o map[string]interface{}) *kinesisanalytics.Output {
	output := &kinesisanalytics.Output{
		Name: aws.String(o["name"].(string)),
	}

	if v := o["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfo := &kinesisanalytics.KinesisFirehoseOutput{
			ResourceARN: aws.String(kf["resource_arn"].(string)),
			RoleARN:     aws.String(kf["role_arn"].(string)),
		}
		output.KinesisFirehoseOutput = kfo
	}

	if v := o["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		kso := &kinesisanalytics.KinesisStreamsOutput{
			ResourceARN: aws.String(ks["resource_arn"].(string)),
			RoleARN:     aws.String(ks["role_arn"].(string)),
		}
		output.KinesisStreamsOutput = kso
	}

	if v := o["lambda"].([]interface{}); len(v) > 0 {
		l := v[0].(map[string]interface{})
		lo := &kinesisanalytics.LambdaOutput{
			ResourceARN: aws.String(l["resource_arn"].(string)),
			RoleARN:     aws.String(l["role_arn"].(string)),
		}
		output.LambdaOutput = lo
	}

	if v := o["schema"].([]interface{}); len(v) > 0 {
		ds := v[0].(map[string]interface{})
		dso := &kinesisanalytics.DestinationSchema{
			RecordFormatType: aws.String(ds["record_format_type"].(string)),
		}
		output.DestinationSchema = dso
	}

	return output
}

func expandKinesisAnalyticsReferenceData(rd map[string]interface{}) *kinesisanalytics.ReferenceDataSource {
	referenceData := &kinesisanalytics.ReferenceDataSource{
		TableName: aws.String(rd["table_name"].(string)),
	}

	if v := rd["s3"].([]interface{}); len(v) > 0 {
		s3 := v[0].(map[string]interface{})
		s3rds := &kinesisanalytics.S3ReferenceDataSource{
			BucketARN:        aws.String(s3["bucket_arn"].(string)),
			FileKey:          aws.String(s3["file_key"].(string)),
			ReferenceRoleARN: aws.String(s3["role_arn"].(string)),
		}
		referenceData.S3ReferenceDataSource = s3rds
	}

	if v := rd["schema"].([]interface{}); len(v) > 0 {
		ss := expandKinesisAnalyticsSourceSchema(v[0].(map[string]interface{}))
		referenceData.ReferenceSchema = ss
	}

	return referenceData
}

func createApplicationUpdateOpts(d *schema.ResourceData) *kinesisanalytics.ApplicationUpdate {
	applicationUpdate := &kinesisanalytics.ApplicationUpdate{}

	if d.HasChange("code") {
		if v, ok := d.GetOk("code"); ok {
			applicationUpdate.ApplicationCodeUpdate = aws.String(v.(string))
		}
	}

	oldLoggingOptions, newLoggingOptions := d.GetChange("cloudwatch_logging_options")
	if len(oldLoggingOptions.([]interface{})) > 0 && len(newLoggingOptions.([]interface{})) > 0 {
		if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
			clo := v.([]interface{})[0].(map[string]interface{})
			cloudwatchLoggingOption := expandKinesisAnalyticsCloudwatchLoggingOptionUpdate(clo)
			applicationUpdate.CloudWatchLoggingOptionUpdates = []*kinesisanalytics.CloudWatchLoggingOptionUpdate{cloudwatchLoggingOption}
		}
	}

	oldInputs, newInputs := d.GetChange("inputs")
	if len(oldInputs.([]interface{})) > 0 && len(newInputs.([]interface{})) > 0 {
		if v, ok := d.GetOk("inputs"); ok {
			vL := v.([]interface{})[0].(map[string]interface{})
			inputUpdate := expandKinesisAnalyticsInputUpdate(vL)
			applicationUpdate.InputUpdates = []*kinesisanalytics.InputUpdate{inputUpdate}
		}
	}

	oldOutputs, newOutputs := d.GetChange("outputs")
	if len(oldOutputs.([]interface{})) > 0 && len(newOutputs.([]interface{})) > 0 {
		if v, ok := d.GetOk("outputs"); ok {
			vL := v.([]interface{})[0].(map[string]interface{})
			outputUpdate := expandKinesisAnalyticsOutputUpdate(vL)
			applicationUpdate.OutputUpdates = []*kinesisanalytics.OutputUpdate{outputUpdate}
		}
	}

	oldReferenceData, newReferenceData := d.GetChange("reference_data_sources")
	if len(oldReferenceData.([]interface{})) > 0 && len(newReferenceData.([]interface{})) > 0 {
		if v := d.Get("reference_data_sources").([]interface{}); len(v) > 0 {
			var rdsus []*kinesisanalytics.ReferenceDataSourceUpdate
			for _, rd := range v {
				rdL := rd.(map[string]interface{})
				rdsu := &kinesisanalytics.ReferenceDataSourceUpdate{
					ReferenceId:     aws.String(rdL["id"].(string)),
					TableNameUpdate: aws.String(rdL["table_name"].(string)),
				}

				if v := rdL["s3"].([]interface{}); len(v) > 0 {
					vL := v[0].(map[string]interface{})
					s3rdsu := &kinesisanalytics.S3ReferenceDataSourceUpdate{
						BucketARNUpdate:        aws.String(vL["bucket_arn"].(string)),
						FileKeyUpdate:          aws.String(vL["file_key"].(string)),
						ReferenceRoleARNUpdate: aws.String(vL["role_arn"].(string)),
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
			applicationUpdate.ReferenceDataSourceUpdates = rdsus
		}
	}

	return applicationUpdate
}

func expandKinesisAnalyticsInputUpdate(vL map[string]interface{}) *kinesisanalytics.InputUpdate {
	inputUpdate := &kinesisanalytics.InputUpdate{
		InputId:          aws.String(vL["id"].(string)),
		NamePrefixUpdate: aws.String(vL["name_prefix"].(string)),
	}

	if v := vL["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfiu := &kinesisanalytics.KinesisFirehoseInputUpdate{
			ResourceARNUpdate: aws.String(kf["resource_arn"].(string)),
			RoleARNUpdate:     aws.String(kf["role_arn"].(string)),
		}
		inputUpdate.KinesisFirehoseInputUpdate = kfiu
	}

	if v := vL["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		ksiu := &kinesisanalytics.KinesisStreamsInputUpdate{
			ResourceARNUpdate: aws.String(ks["resource_arn"].(string)),
			RoleARNUpdate:     aws.String(ks["role_arn"].(string)),
		}
		inputUpdate.KinesisStreamsInputUpdate = ksiu
	}

	if v := vL["parallelism"].([]interface{}); len(v) > 0 {
		p := v[0].(map[string]interface{})

		if c, ok := p["count"]; ok {
			ipu := &kinesisanalytics.InputParallelismUpdate{
				CountUpdate: aws.Int64(int64(c.(int))),
			}
			inputUpdate.InputParallelismUpdate = ipu
		}
	}

	if v := vL["processing_configuration"].([]interface{}); len(v) > 0 {
		pc := v[0].(map[string]interface{})

		if l := pc["lambda"].([]interface{}); len(l) > 0 {
			lp := l[0].(map[string]interface{})
			ipc := &kinesisanalytics.InputProcessingConfigurationUpdate{
				InputLambdaProcessorUpdate: &kinesisanalytics.InputLambdaProcessorUpdate{
					ResourceARNUpdate: aws.String(lp["resource_arn"].(string)),
					RoleARNUpdate:     aws.String(lp["role_arn"].(string)),
				},
			}
			inputUpdate.InputProcessingConfigurationUpdate = ipc
		}
	}

	if v := vL["schema"].([]interface{}); len(v) > 0 {
		ss := &kinesisanalytics.InputSchemaUpdate{}
		vL := v[0].(map[string]interface{})

		if v := vL["record_columns"].([]interface{}); len(v) > 0 {
			var rcs []*kinesisanalytics.RecordColumn

			for _, rc := range v {
				rcD := rc.(map[string]interface{})
				rc := &kinesisanalytics.RecordColumn{
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
			rf := &kinesisanalytics.RecordFormat{}

			if v := vL["mapping_parameters"].([]interface{}); len(v) > 0 {
				vL := v[0].(map[string]interface{})
				mp := &kinesisanalytics.MappingParameters{}

				if v := vL["csv"].([]interface{}); len(v) > 0 {
					cL := v[0].(map[string]interface{})
					cmp := &kinesisanalytics.CSVMappingParameters{
						RecordColumnDelimiter: aws.String(cL["record_column_delimiter"].(string)),
						RecordRowDelimiter:    aws.String(cL["record_row_delimiter"].(string)),
					}
					mp.CSVMappingParameters = cmp
					rf.RecordFormatType = aws.String("CSV")
				}

				if v := vL["json"].([]interface{}); len(v) > 0 {
					jL := v[0].(map[string]interface{})
					jmp := &kinesisanalytics.JSONMappingParameters{
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

func expandKinesisAnalyticsOutputUpdate(vL map[string]interface{}) *kinesisanalytics.OutputUpdate {
	outputUpdate := &kinesisanalytics.OutputUpdate{
		OutputId:   aws.String(vL["id"].(string)),
		NameUpdate: aws.String(vL["name"].(string)),
	}

	if v := vL["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfou := &kinesisanalytics.KinesisFirehoseOutputUpdate{
			ResourceARNUpdate: aws.String(kf["resource_arn"].(string)),
			RoleARNUpdate:     aws.String(kf["role_arn"].(string)),
		}
		outputUpdate.KinesisFirehoseOutputUpdate = kfou
	}

	if v := vL["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		ksou := &kinesisanalytics.KinesisStreamsOutputUpdate{
			ResourceARNUpdate: aws.String(ks["resource_arn"].(string)),
			RoleARNUpdate:     aws.String(ks["role_arn"].(string)),
		}
		outputUpdate.KinesisStreamsOutputUpdate = ksou
	}

	if v := vL["lambda"].([]interface{}); len(v) > 0 {
		l := v[0].(map[string]interface{})
		lou := &kinesisanalytics.LambdaOutputUpdate{
			ResourceARNUpdate: aws.String(l["resource_arn"].(string)),
			RoleARNUpdate:     aws.String(l["role_arn"].(string)),
		}
		outputUpdate.LambdaOutputUpdate = lou
	}

	if v := vL["schema"].([]interface{}); len(v) > 0 {
		ds := v[0].(map[string]interface{})
		dsu := &kinesisanalytics.DestinationSchema{
			RecordFormatType: aws.String(ds["record_format_type"].(string)),
		}
		outputUpdate.DestinationSchemaUpdate = dsu
	}

	return outputUpdate
}

func expandKinesisAnalyticsCloudwatchLoggingOptionUpdate(clo map[string]interface{}) *kinesisanalytics.CloudWatchLoggingOptionUpdate {
	cloudwatchLoggingOption := &kinesisanalytics.CloudWatchLoggingOptionUpdate{
		CloudWatchLoggingOptionId: aws.String(clo["id"].(string)),
		LogStreamARNUpdate:        aws.String(clo["log_stream_arn"].(string)),
		RoleARNUpdate:             aws.String(clo["role_arn"].(string)),
	}
	return cloudwatchLoggingOption
}

func flattenKinesisAnalyticsCloudwatchLoggingOptions(options []*kinesisanalytics.CloudWatchLoggingOptionDescription) []interface{} {
	s := []interface{}{}
	for _, v := range options {
		option := map[string]interface{}{
			"id":             aws.StringValue(v.CloudWatchLoggingOptionId),
			"log_stream_arn": aws.StringValue(v.LogStreamARN),
			"role_arn":       aws.StringValue(v.RoleARN),
		}
		s = append(s, option)
	}
	return s
}

func flattenKinesisAnalyticsInputs(inputs []*kinesisanalytics.InputDescription) []interface{} {
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

func flattenKinesisAnalyticsOutputs(outputs []*kinesisanalytics.OutputDescription) []interface{} {
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

func flattenKinesisAnalyticsReferenceDataSources(dataSources []*kinesisanalytics.ReferenceDataSourceDescription) []interface{} {
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

//
// New expand/flatten functions.
// TODO Remove 'V1'.
//

func expandKinesisAnalyticsV1CloudWatchLoggingOptions(vCloudWatchLoggingOptions []interface{}) []*kinesisanalytics.CloudWatchLoggingOption {
	if len(vCloudWatchLoggingOptions) == 0 || vCloudWatchLoggingOptions[0] == nil {
		return nil
	}

	cloudWatchLoggingOption := &kinesisanalytics.CloudWatchLoggingOption{}

	mCloudWatchLoggingOption := vCloudWatchLoggingOptions[0].(map[string]interface{})

	if vLogStreamArn, ok := mCloudWatchLoggingOption["log_stream_arn"].(string); ok && vLogStreamArn != "" {
		cloudWatchLoggingOption.LogStreamARN = aws.String(vLogStreamArn)
	}
	if vRoleArn, ok := mCloudWatchLoggingOption["role_arn"].(string); ok && vRoleArn != "" {
		cloudWatchLoggingOption.RoleARN = aws.String(vRoleArn)
	}

	return []*kinesisanalytics.CloudWatchLoggingOption{cloudWatchLoggingOption}
}

func expandKinesisAnalyticsV1Inputs(vInputs []interface{}) []*kinesisanalytics.Input {
	if len(vInputs) == 0 || vInputs[0] == nil {
		return nil
	}

	return []*kinesisanalytics.Input{expandKinesisAnalyticsV1Input(vInputs)}
}

func expandKinesisAnalyticsV1Input(vInput []interface{}) *kinesisanalytics.Input {
	if len(vInput) == 0 || vInput[0] == nil {
		return nil
	}

	input := &kinesisanalytics.Input{}

	mInput := vInput[0].(map[string]interface{})

	if vInputParallelism, ok := mInput["parallelism"].([]interface{}); ok && len(vInputParallelism) > 0 && vInputParallelism[0] != nil {
		inputParallelism := &kinesisanalytics.InputParallelism{}

		mInputParallelism := vInputParallelism[0].(map[string]interface{})

		if vCount, ok := mInputParallelism["count"].(int); ok {
			inputParallelism.Count = aws.Int64(int64(vCount))
		}

		input.InputParallelism = inputParallelism
	}

	if vInputProcessingConfiguration, ok := mInput["processing_configuration"].([]interface{}); ok {
		input.InputProcessingConfiguration = expandKinesisAnalyticsV1InputProcessingConfiguration(vInputProcessingConfiguration)
	}

	if vInputSchema, ok := mInput["schema"].([]interface{}); ok {
		input.InputSchema = expandKinesisAnalyticsV1SourceSchema(vInputSchema)
	}

	if vKinesisFirehoseInput, ok := mInput["kinesis_firehose"].([]interface{}); ok && len(vKinesisFirehoseInput) > 0 && vKinesisFirehoseInput[0] != nil {
		kinesisFirehoseInput := &kinesisanalytics.KinesisFirehoseInput{}

		mKinesisFirehoseInput := vKinesisFirehoseInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseInput["resource_arn"].(string); ok && vResourceArn != "" {
			kinesisFirehoseInput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisFirehoseInput["role_arn"].(string); ok && vRoleArn != "" {
			kinesisFirehoseInput.RoleARN = aws.String(vRoleArn)
		}

		input.KinesisFirehoseInput = kinesisFirehoseInput
	}

	if vKinesisStreamsInput, ok := mInput["kinesis_stream"].([]interface{}); ok && len(vKinesisStreamsInput) > 0 && vKinesisStreamsInput[0] != nil {
		kinesisStreamsInput := &kinesisanalytics.KinesisStreamsInput{}

		mKinesisStreamsInput := vKinesisStreamsInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsInput["resource_arn"].(string); ok && vResourceArn != "" {
			kinesisStreamsInput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisStreamsInput["role_arn"].(string); ok && vRoleArn != "" {
			kinesisStreamsInput.RoleARN = aws.String(vRoleArn)
		}

		input.KinesisStreamsInput = kinesisStreamsInput
	}

	if vNamePrefix, ok := mInput["name_prefix"].(string); ok && vNamePrefix != "" {
		input.NamePrefix = aws.String(vNamePrefix)
	}

	return input
}

func expandKinesisAnalyticsV1InputProcessingConfiguration(vInputProcessingConfiguration []interface{}) *kinesisanalytics.InputProcessingConfiguration {
	if len(vInputProcessingConfiguration) == 0 || vInputProcessingConfiguration[0] == nil {
		return nil
	}

	inputProcessingConfiguration := &kinesisanalytics.InputProcessingConfiguration{}

	mInputProcessingConfiguration := vInputProcessingConfiguration[0].(map[string]interface{})

	if vInputLambdaProcessor, ok := mInputProcessingConfiguration["lambda"].([]interface{}); ok && len(vInputLambdaProcessor) > 0 && vInputLambdaProcessor[0] != nil {
		inputLambdaProcessor := &kinesisanalytics.InputLambdaProcessor{}

		mInputLambdaProcessor := vInputLambdaProcessor[0].(map[string]interface{})

		if vResourceArn, ok := mInputLambdaProcessor["resource_arn"].(string); ok && vResourceArn != "" {
			inputLambdaProcessor.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mInputLambdaProcessor["role_arn"].(string); ok && vRoleArn != "" {
			inputLambdaProcessor.RoleARN = aws.String(vRoleArn)
		}

		inputProcessingConfiguration.InputLambdaProcessor = inputLambdaProcessor
	}

	return inputProcessingConfiguration
}

func expandKinesisAnalyticsV1Output(vOutput interface{}) *kinesisanalytics.Output {
	if vOutput == nil {
		return nil
	}

	output := &kinesisanalytics.Output{}

	mOutput := vOutput.(map[string]interface{})

	if vDestinationSchema, ok := mOutput["destination_schema"].([]interface{}); ok {
		destinationSchema := &kinesisanalytics.DestinationSchema{}

		mDestinationSchema := vDestinationSchema[0].(map[string]interface{})

		if vRecordFormatType, ok := mDestinationSchema["record_format_type"].(string); ok && vRecordFormatType != "" {
			destinationSchema.RecordFormatType = aws.String(vRecordFormatType)
		}

		output.DestinationSchema = destinationSchema
	}

	if vKinesisFirehoseOutput, ok := mOutput["kinesis_firehose"].([]interface{}); ok && len(vKinesisFirehoseOutput) > 0 && vKinesisFirehoseOutput[0] != nil {
		kinesisFirehoseOutput := &kinesisanalytics.KinesisFirehoseOutput{}

		mKinesisFirehoseOutput := vKinesisFirehoseOutput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseOutput["resource_arn"].(string); ok && vResourceArn != "" {
			kinesisFirehoseOutput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisFirehoseOutput["role_arn"].(string); ok && vRoleArn != "" {
			kinesisFirehoseOutput.RoleARN = aws.String(vRoleArn)
		}

		output.KinesisFirehoseOutput = kinesisFirehoseOutput
	}

	if vKinesisStreamsOutput, ok := mOutput["kinesis_stream"].([]interface{}); ok && len(vKinesisStreamsOutput) > 0 && vKinesisStreamsOutput[0] != nil {
		kinesisStreamsOutput := &kinesisanalytics.KinesisStreamsOutput{}

		mKinesisStreamsOutput := vKinesisStreamsOutput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsOutput["resource_arn"].(string); ok && vResourceArn != "" {
			kinesisStreamsOutput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisStreamsOutput["role_arn"].(string); ok && vRoleArn != "" {
			kinesisStreamsOutput.RoleARN = aws.String(vRoleArn)
		}

		output.KinesisStreamsOutput = kinesisStreamsOutput
	}

	if vLambdaOutput, ok := mOutput["lambda"].([]interface{}); ok && len(vLambdaOutput) > 0 && vLambdaOutput[0] != nil {
		lambdaOutput := &kinesisanalytics.LambdaOutput{}

		mLambdaOutput := vLambdaOutput[0].(map[string]interface{})

		if vResourceArn, ok := mLambdaOutput["resource_arn"].(string); ok && vResourceArn != "" {
			lambdaOutput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mLambdaOutput["role_arn"].(string); ok && vRoleArn != "" {
			lambdaOutput.RoleARN = aws.String(vRoleArn)
		}

		output.LambdaOutput = lambdaOutput
	}

	if vName, ok := mOutput["name"].(string); ok && vName != "" {
		output.Name = aws.String(vName)
	}

	return output
}

func expandKinesisAnalyticsV1Outputs(vOutputs []interface{}) []*kinesisanalytics.Output {
	if len(vOutputs) == 0 {
		return nil
	}

	outputs := []*kinesisanalytics.Output{}

	for _, vOutput := range vOutputs {
		output := expandKinesisAnalyticsV1Output(vOutput)

		if output != nil {
			outputs = append(outputs, output)
		}
	}

	return outputs
}

func expandKinesisAnalyticsV1RecordColumns(vRecordColumns []interface{}) []*kinesisanalytics.RecordColumn {
	recordColumns := []*kinesisanalytics.RecordColumn{}

	for _, vRecordColumn := range vRecordColumns {
		recordColumn := &kinesisanalytics.RecordColumn{}

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

func expandKinesisAnalyticsV1RecordFormat(vRecordFormat []interface{}) *kinesisanalytics.RecordFormat {
	if len(vRecordFormat) == 0 || vRecordFormat[0] == nil {
		return nil
	}

	recordFormat := &kinesisanalytics.RecordFormat{}

	mRecordFormat := vRecordFormat[0].(map[string]interface{})

	if vMappingParameters, ok := mRecordFormat["mapping_parameters"].([]interface{}); ok && len(vMappingParameters) > 0 && vMappingParameters[0] != nil {
		mappingParameters := &kinesisanalytics.MappingParameters{}

		mMappingParameters := vMappingParameters[0].(map[string]interface{})

		if vCsvMappingParameters, ok := mMappingParameters["csv"].([]interface{}); ok && len(vCsvMappingParameters) > 0 && vCsvMappingParameters[0] != nil {
			csvMappingParameters := &kinesisanalytics.CSVMappingParameters{}

			mCsvMappingParameters := vCsvMappingParameters[0].(map[string]interface{})

			if vRecordColumnDelimiter, ok := mCsvMappingParameters["record_column_delimiter"].(string); ok && vRecordColumnDelimiter != "" {
				csvMappingParameters.RecordColumnDelimiter = aws.String(vRecordColumnDelimiter)
			}
			if vRecordRowDelimiter, ok := mCsvMappingParameters["record_row_delimiter"].(string); ok && vRecordRowDelimiter != "" {
				csvMappingParameters.RecordRowDelimiter = aws.String(vRecordRowDelimiter)
			}

			mappingParameters.CSVMappingParameters = csvMappingParameters

			recordFormat.RecordFormatType = aws.String(kinesisanalytics.RecordFormatTypeCsv)
		}

		if vJsonMappingParameters, ok := mMappingParameters["json"].([]interface{}); ok && len(vJsonMappingParameters) > 0 && vJsonMappingParameters[0] != nil {
			jsonMappingParameters := &kinesisanalytics.JSONMappingParameters{}

			mJsonMappingParameters := vJsonMappingParameters[0].(map[string]interface{})

			if vRecordRowPath, ok := mJsonMappingParameters["record_row_path"].(string); ok && vRecordRowPath != "" {
				jsonMappingParameters.RecordRowPath = aws.String(vRecordRowPath)
			}

			mappingParameters.JSONMappingParameters = jsonMappingParameters

			recordFormat.RecordFormatType = aws.String(kinesisanalytics.RecordFormatTypeJson)
		}

		recordFormat.MappingParameters = mappingParameters
	}

	return recordFormat
}

func expandKinesisAnalyticsV1ReferenceDataSource(vReferenceDataSource []interface{}) *kinesisanalytics.ReferenceDataSource {
	if len(vReferenceDataSource) == 0 || vReferenceDataSource[0] == nil {
		return nil
	}

	referenceDataSource := &kinesisanalytics.ReferenceDataSource{}

	mReferenceDataSource := vReferenceDataSource[0].(map[string]interface{})

	if vReferenceSchema, ok := mReferenceDataSource["schema"].([]interface{}); ok {
		referenceDataSource.ReferenceSchema = expandKinesisAnalyticsV1SourceSchema(vReferenceSchema)
	}

	if vS3ReferenceDataSource, ok := mReferenceDataSource["s3"].([]interface{}); ok && len(vS3ReferenceDataSource) > 0 && vS3ReferenceDataSource[0] != nil {
		s3ReferenceDataSource := &kinesisanalytics.S3ReferenceDataSource{}

		mS3ReferenceDataSource := vS3ReferenceDataSource[0].(map[string]interface{})

		if vBucketArn, ok := mS3ReferenceDataSource["bucket_arn"].(string); ok && vBucketArn != "" {
			s3ReferenceDataSource.BucketARN = aws.String(vBucketArn)
		}
		if vFileKey, ok := mS3ReferenceDataSource["file_key"].(string); ok && vFileKey != "" {
			s3ReferenceDataSource.FileKey = aws.String(vFileKey)
		}
		if vReferenceRoleArn, ok := mS3ReferenceDataSource["role_arn"].(string); ok && vReferenceRoleArn != "" {
			s3ReferenceDataSource.ReferenceRoleARN = aws.String(vReferenceRoleArn)
		}

		referenceDataSource.S3ReferenceDataSource = s3ReferenceDataSource
	}

	if vTableName, ok := mReferenceDataSource["table_name"].(string); ok && vTableName != "" {
		referenceDataSource.TableName = aws.String(vTableName)
	}

	return referenceDataSource
}

func expandKinesisAnalyticsV1ReferenceDataSourceUpdate(vReferenceDataSource []interface{}) *kinesisanalytics.ReferenceDataSourceUpdate {
	if len(vReferenceDataSource) == 0 || vReferenceDataSource[0] == nil {
		return nil
	}

	referenceDataSourceUpdate := &kinesisanalytics.ReferenceDataSourceUpdate{}

	mReferenceDataSource := vReferenceDataSource[0].(map[string]interface{})

	if vReferenceId, ok := mReferenceDataSource["id"].(string); ok && vReferenceId != "" {
		referenceDataSourceUpdate.ReferenceId = aws.String(vReferenceId)
	}

	if vReferenceSchema, ok := mReferenceDataSource["schema"].([]interface{}); ok {
		referenceDataSourceUpdate.ReferenceSchemaUpdate = expandKinesisAnalyticsV1SourceSchema(vReferenceSchema)
	}

	if vS3ReferenceDataSource, ok := mReferenceDataSource["s3"].([]interface{}); ok && len(vS3ReferenceDataSource) > 0 && vS3ReferenceDataSource[0] != nil {
		s3ReferenceDataSourceUpdate := &kinesisanalytics.S3ReferenceDataSourceUpdate{}

		mS3ReferenceDataSource := vS3ReferenceDataSource[0].(map[string]interface{})

		if vBucketArn, ok := mS3ReferenceDataSource["bucket_arn"].(string); ok && vBucketArn != "" {
			s3ReferenceDataSourceUpdate.BucketARNUpdate = aws.String(vBucketArn)
		}
		if vFileKey, ok := mS3ReferenceDataSource["file_key"].(string); ok && vFileKey != "" {
			s3ReferenceDataSourceUpdate.FileKeyUpdate = aws.String(vFileKey)
		}
		if vReferenceRoleArn, ok := mS3ReferenceDataSource["role_arn"].(string); ok && vReferenceRoleArn != "" {
			s3ReferenceDataSourceUpdate.ReferenceRoleARNUpdate = aws.String(vReferenceRoleArn)
		}

		referenceDataSourceUpdate.S3ReferenceDataSourceUpdate = s3ReferenceDataSourceUpdate
	}

	if vTableName, ok := mReferenceDataSource["table_name"].(string); ok && vTableName != "" {
		referenceDataSourceUpdate.TableNameUpdate = aws.String(vTableName)
	}

	return referenceDataSourceUpdate
}

func expandKinesisAnalyticsV1SourceSchema(vSourceSchema []interface{}) *kinesisanalytics.SourceSchema {
	if len(vSourceSchema) == 0 || vSourceSchema[0] == nil {
		return nil
	}

	sourceSchema := &kinesisanalytics.SourceSchema{}

	mSourceSchema := vSourceSchema[0].(map[string]interface{})

	if vRecordColumns, ok := mSourceSchema["record_columns"].(*schema.Set); ok && vRecordColumns.Len() > 0 {
		sourceSchema.RecordColumns = expandKinesisAnalyticsV1RecordColumns(vRecordColumns.List())
	}

	if vRecordEncoding, ok := mSourceSchema["record_encoding"].(string); ok && vRecordEncoding != "" {
		sourceSchema.RecordEncoding = aws.String(vRecordEncoding)
	}

	if vRecordFormat, ok := mSourceSchema["record_format"].([]interface{}); ok && len(vRecordFormat) > 0 && vRecordFormat[0] != nil {
		sourceSchema.RecordFormat = expandKinesisAnalyticsV1RecordFormat(vRecordFormat)
	}

	return sourceSchema
}
