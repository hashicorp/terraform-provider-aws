package aws

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsKinesisAnalyticsApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisAnalyticsApplicationCreate,
		Read:   resourceAwsKinesisAnalyticsApplicationRead,
		Update: resourceAwsKinesisAnalyticsApplicationUpdate,
		Delete: resourceAwsKinesisAnalyticsApplicationDelete,

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

			"code": {
				Type:     schema.TypeString,
				Optional: true,
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

						"log_stream": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},

						"role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
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
									"resource": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},

									"role": {
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
									"resource": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},

									"role": {
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
												"resource": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validateArn,
												},

												"role": {
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
													Required: true,
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
		},
	}
}

func resourceAwsKinesisAnalyticsApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn
	name := d.Get("name").(string)
	createOpts := &kinesisanalytics.CreateApplicationInput{
		ApplicationName: aws.String(name),
	}

	if v, ok := d.GetOk("code"); ok && v.(string) != "" {
		createOpts.ApplicationCode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
		clo := v.([]interface{})[0].(map[string]interface{})
		cloudwatchLoggingOption := createCloudwatchLoggingOption(clo)
		createOpts.CloudWatchLoggingOptions = []*kinesisanalytics.CloudWatchLoggingOption{cloudwatchLoggingOption}
	}

	if v, ok := d.GetOk("inputs"); ok {
		i := v.([]interface{})[0].(map[string]interface{})
		inputs := createInputs(i)
		createOpts.Inputs = []*kinesisanalytics.Input{inputs}
	}

	_, err := conn.CreateApplication(createOpts)
	if err != nil {
		return fmt.Errorf("Unable to create Kinesis Analytics Application: %s", err)
	}

	return resourceAwsKinesisAnalyticsApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn
	name := d.Get("name").(string)

	describeOpts := &kinesisanalytics.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}
	resp, err := conn.DescribeApplication(describeOpts)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("[WARN] Error reading Kinesis Analytics Application: \"%s\", code: \"%s\"", awsErr.Message(), awsErr.Code())
		}
		return err
	}

	d.SetId(aws.StringValue(resp.ApplicationDetail.ApplicationARN))
	d.Set("name", aws.StringValue(resp.ApplicationDetail.ApplicationName))
	d.Set("arn", aws.StringValue(resp.ApplicationDetail.ApplicationARN))
	d.Set("code", aws.StringValue(resp.ApplicationDetail.ApplicationCode))
	d.Set("create_timestamp", aws.TimeValue(resp.ApplicationDetail.CreateTimestamp).Format(time.RFC3339))
	d.Set("description", aws.StringValue(resp.ApplicationDetail.ApplicationDescription))
	d.Set("last_update_timestamp", aws.TimeValue(resp.ApplicationDetail.LastUpdateTimestamp).Format(time.RFC3339))
	d.Set("status", aws.StringValue(resp.ApplicationDetail.ApplicationStatus))
	d.Set("version", int(aws.Int64Value(resp.ApplicationDetail.ApplicationVersionId)))

	if err := d.Set("cloudwatch_logging_options", getCloudwatchLoggingOptions(resp.ApplicationDetail.CloudWatchLoggingOptionDescriptions)); err != nil {
		return err
	}

	if err := d.Set("inputs", getInputs(resp.ApplicationDetail.InputDescriptions)); err != nil {
		return err
	}

	return nil
}

func resourceAwsKinesisAnalyticsApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn

	if !d.IsNewResource() {
		applicationUpdate := &kinesisanalytics.ApplicationUpdate{}
		name := d.Get("name").(string)
		version := d.Get("version").(int)

		updateApplicationOpts := &kinesisanalytics.UpdateApplicationInput{
			ApplicationName:             aws.String(name),
			CurrentApplicationVersionId: aws.Int64(int64(version)),
		}

		applicationUpdate, err := createApplicationUpdateOpts(d)
		if err != nil {
			return err
		}

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
				cloudwatchLoggingOption := createCloudwatchLoggingOption(clo)
				addOpts := &kinesisanalytics.AddApplicationCloudWatchLoggingOptionInput{
					ApplicationName:             aws.String(name),
					CurrentApplicationVersionId: aws.Int64(int64(version)),
					CloudWatchLoggingOption:     cloudwatchLoggingOption,
				}
				conn.AddApplicationCloudWatchLoggingOption(addOpts)
			}
		}
	}

	return resourceAwsKinesisAnalyticsApplicationRead(d, meta)
}

func resourceAwsKinesisAnalyticsApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsconn
	name := d.Get("name").(string)
	createTimestamp, parseErr := time.Parse(time.RFC3339, d.Get("create_timestamp").(string))
	if parseErr != nil {
		return parseErr
	}

	log.Printf("[DEBUG] Kinesis Analytics Application destroy: %v", d.Id())
	deleteOpts := &kinesisanalytics.DeleteApplicationInput{
		ApplicationName: aws.String(name),
		CreateTimestamp: aws.Time(createTimestamp),
	}
	_, deleteErr := conn.DeleteApplication(deleteOpts)
	if deleteErr != nil {
		return deleteErr
	}

	log.Printf("[DEBUG] Kinesis Analytics Application deleted: %v", d.Id())
	return nil
}

func createCloudwatchLoggingOption(clo map[string]interface{}) *kinesisanalytics.CloudWatchLoggingOption {
	cloudwatchLoggingOption := &kinesisanalytics.CloudWatchLoggingOption{
		LogStreamARN: aws.String(clo["log_stream"].(string)),
		RoleARN:      aws.String(clo["role"].(string)),
	}
	return cloudwatchLoggingOption
}

func createInputs(i map[string]interface{}) *kinesisanalytics.Input {
	input := &kinesisanalytics.Input{}

	if v, ok := i["name_prefix"]; ok {
		input.NamePrefix = aws.String(v.(string))
	}

	if v := i["kinesis_firehose"].([]interface{}); len(v) > 0 {
		kf := v[0].(map[string]interface{})
		kfi := &kinesisanalytics.KinesisFirehoseInput{
			ResourceARN: aws.String(kf["resource"].(string)),
			RoleARN:     aws.String(kf["role"].(string)),
		}
		input.KinesisFirehoseInput = kfi
	}

	if v := i["kinesis_stream"].([]interface{}); len(v) > 0 {
		ks := v[0].(map[string]interface{})
		ksi := &kinesisanalytics.KinesisStreamsInput{
			ResourceARN: aws.String(ks["resource"].(string)),
			RoleARN:     aws.String(ks["role"].(string)),
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
					ResourceARN: aws.String(lp["resource"].(string)),
					RoleARN:     aws.String(lp["role"].(string)),
				},
			}
			input.InputProcessingConfiguration = ipc
		}
	}

	if v := i["schema"].([]interface{}); len(v) > 0 {
		ss := &kinesisanalytics.SourceSchema{}
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

		input.InputSchema = ss
	}

	return input
}

func createApplicationUpdateOpts(d *schema.ResourceData) (*kinesisanalytics.ApplicationUpdate, error) {
	applicationUpdate := &kinesisanalytics.ApplicationUpdate{}

	if d.HasChange("code") {
		if v, ok := d.GetOk("code"); ok && v.(string) != "" {
			applicationUpdate.ApplicationCodeUpdate = aws.String(v.(string))
		}
	}

	oldLoggingOptions, _ := d.GetChange("cloudwatch_logging_options")
	if len(oldLoggingOptions.([]interface{})) > 0 {
		if v, ok := d.GetOk("cloudwatch_logging_options"); ok {
			var cloudwatchLoggingOptions []*kinesisanalytics.CloudWatchLoggingOptionUpdate
			clo := v.([]interface{})[0].(map[string]interface{})
			cloudwatchLoggingOption := &kinesisanalytics.CloudWatchLoggingOptionUpdate{
				CloudWatchLoggingOptionId: aws.String(clo["id"].(string)),
				LogStreamARNUpdate:        aws.String(clo["log_stream"].(string)),
				RoleARNUpdate:             aws.String(clo["role"].(string)),
			}
			cloudwatchLoggingOptions = append(cloudwatchLoggingOptions, cloudwatchLoggingOption)
			applicationUpdate.CloudWatchLoggingOptionUpdates = cloudwatchLoggingOptions
		}
	}

	return applicationUpdate, nil
}

func getCloudwatchLoggingOptions(options []*kinesisanalytics.CloudWatchLoggingOptionDescription) []interface{} {
	s := []interface{}{}
	for _, v := range options {
		option := map[string]interface{}{
			"id":         aws.StringValue(v.CloudWatchLoggingOptionId),
			"log_stream": aws.StringValue(v.LogStreamARN),
			"role":       aws.StringValue(v.RoleARN),
		}
		s = append(s, option)
	}
	return s
}

func getInputs(inputs []*kinesisanalytics.InputDescription) []interface{} {
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
				input["processing_configurations"] = []interface{}{
					map[string]interface{}{
						"lambda": []interface{}{
							map[string]interface{}{
								"resource": aws.StringValue(ipcd.InputLambdaProcessorDescription.ResourceARN),
								"role":     aws.StringValue(ipcd.InputLambdaProcessorDescription.RoleARN),
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
					"resource": aws.StringValue(id.KinesisFirehoseInputDescription.ResourceARN),
					"role":     aws.StringValue(id.KinesisFirehoseInputDescription.RoleARN),
				},
			}
		}

		if id.KinesisStreamsInputDescription != nil {
			input["kinesis_stream"] = []interface{}{
				map[string]interface{}{
					"resource": aws.StringValue(id.KinesisStreamsInputDescription.ResourceARN),
					"role":     aws.StringValue(id.KinesisStreamsInputDescription.RoleARN),
				},
			}
		}

		s = append(s, input)
	}
	return s
}
