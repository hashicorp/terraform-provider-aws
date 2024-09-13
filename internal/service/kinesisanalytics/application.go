// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalytics

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalytics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalytics/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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

// @SDKResource("aws_kinesis_analytics_application", name="Application")
// @Tags(identifierAttribute="arn")
func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationCreate,
		ReadWithoutTimeout:   resourceApplicationRead,
		UpdateWithoutTimeout: resourceApplicationUpdate,
		DeleteWithoutTimeout: resourceApplicationDelete,

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			customdiff.ForceNewIfChange("inputs", func(_ context.Context, old, new, meta interface{}) bool {
				// An existing input configuration cannot be deleted.
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
		),

		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationImport,
		},

		Schema: map[string]*schema.Schema{
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
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"log_stream_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
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
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"inputs": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kinesis_firehose": {
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
									names.AttrRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
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
									names.AttrResourceARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
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
						"parallelism": {
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
												names.AttrResourceARN: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},

												names.AttrRoleARN: {
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
						names.AttrSchema: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"record_columns": {
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
															names.AttrJSON: {
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
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"starting_position": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.InputStartingPosition](),
									},
								},
							},
						},
						"stream_names": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
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
			"outputs": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kinesis_firehose": {
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
									names.AttrRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
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
									names.AttrResourceARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
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
									names.AttrResourceARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrRoleARN: {
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
						names.AttrSchema: {
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
					},
				},
			},
			"reference_data_sources": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
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
										ValidateFunc: verify.ValidARN,
									},
									"file_key": {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						names.AttrSchema: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"record_columns": {
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
															names.AttrJSON: {
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
						names.AttrTableName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 32),
						},
					},
				},
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
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsClient(ctx)

	applicationName := d.Get(names.AttrName).(string)
	input := &kinesisanalytics.CreateApplicationInput{
		ApplicationCode:          aws.String(d.Get("code").(string)),
		ApplicationDescription:   aws.String(d.Get(names.AttrDescription).(string)),
		ApplicationName:          aws.String(applicationName),
		CloudWatchLoggingOptions: expandCloudWatchLoggingOptions(d.Get("cloudwatch_logging_options").([]interface{})),
		Inputs:                   expandInputs(d.Get("inputs").([]interface{})),
		Outputs:                  expandOutputs(d.Get("outputs").(*schema.Set).List()),
		Tags:                     getTagsIn(ctx),
	}

	outputRaw, err := waitIAMPropagation(ctx, func() (interface{}, error) {
		return conn.CreateApplication(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Analytics Application (%s): %s", applicationName, err)
	}

	applicationSummary := outputRaw.(*kinesisanalytics.CreateApplicationOutput).ApplicationSummary

	d.SetId(aws.ToString(applicationSummary.ApplicationARN))

	if v := d.Get("reference_data_sources").([]interface{}); len(v) > 0 && v[0] != nil {
		// Add new reference data source.
		input := &kinesisanalytics.AddApplicationReferenceDataSourceInput{
			ApplicationName:             aws.String(applicationName),
			CurrentApplicationVersionId: aws.Int64(1), // Newly created application version.
			ReferenceDataSource:         expandReferenceDataSource(v),
		}

		_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
			return conn.AddApplicationReferenceDataSource(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics Application (%s) reference data source: %+v", d.Id(), err)
		}
	}

	if _, ok := d.GetOk("start_application"); ok {
		if v, ok := d.GetOk("inputs"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			var inputStartingPosition string

			if v, ok := tfMap["starting_position_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				tfMap := v[0].(map[string]interface{})

				if v, ok := tfMap["starting_position"].(string); ok && v != "" {
					inputStartingPosition = v
				}
			}

			application, err := findApplicationDetailByName(ctx, conn, applicationName)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics Application (%s): %s", d.Id(), err)
			}

			if err := startApplication(ctx, conn, application, inputStartingPosition); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsClient(ctx)

	application, err := findApplicationDetailByName(ctx, conn, d.Get(names.AttrName).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Analytics Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics Application (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(application.ApplicationARN)
	d.Set(names.AttrARN, arn)
	if err := d.Set("cloudwatch_logging_options", flattenCloudWatchLoggingOptionDescriptions(application.CloudWatchLoggingOptionDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cloudwatch_logging_options: %s", err)
	}
	d.Set("code", application.ApplicationCode)
	d.Set("create_timestamp", aws.ToTime(application.CreateTimestamp).Format(time.RFC3339))
	d.Set(names.AttrDescription, application.ApplicationDescription)
	if err := d.Set("inputs", flattenInputDescriptions(application.InputDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting inputs: %s", err)
	}
	d.Set("last_update_timestamp", aws.ToTime(application.LastUpdateTimestamp).Format(time.RFC3339))
	d.Set(names.AttrName, application.ApplicationName)
	if err := d.Set("outputs", flattenOutputDescriptions(application.OutputDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting outputs: %s", err)
	}
	if err := d.Set("reference_data_sources", flattenReferenceDataSourceDescriptions(application.ReferenceDataSourceDescriptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting reference_data_sources: %s", err)
	}
	d.Set(names.AttrStatus, application.ApplicationStatus)
	d.Set(names.AttrVersion, application.ApplicationVersionId)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsClient(ctx)

	if d.HasChanges("cloudwatch_logging_options", "code", "inputs", "outputs", "reference_data_sources") {
		applicationName := d.Get(names.AttrName).(string)
		currentApplicationVersionID := int64(d.Get(names.AttrVersion).(int))
		updateApplication := false

		input := &kinesisanalytics.UpdateApplicationInput{
			ApplicationName:   aws.String(applicationName),
			ApplicationUpdate: &awstypes.ApplicationUpdate{},
		}

		if d.HasChange("cloudwatch_logging_options") {
			o, n := d.GetChange("cloudwatch_logging_options")

			if len(o.([]interface{})) == 0 {
				// Add new CloudWatch logging options.
				mNewCloudWatchLoggingOption := n.([]interface{})[0].(map[string]interface{})

				input := &kinesisanalytics.AddApplicationCloudWatchLoggingOptionInput{
					ApplicationName: aws.String(applicationName),
					CloudWatchLoggingOption: &awstypes.CloudWatchLoggingOption{
						LogStreamARN: aws.String(mNewCloudWatchLoggingOption["log_stream_arn"].(string)),
						RoleARN:      aws.String(mNewCloudWatchLoggingOption[names.AttrRoleARN].(string)),
					},
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
				}

				_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
					return conn.AddApplicationCloudWatchLoggingOption(ctx, input)
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics Application (%s) CloudWatch logging option: %s", d.Id(), err)
				}

				if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
				}

				currentApplicationVersionID += 1
			} else if len(n.([]interface{})) == 0 {
				// Delete existing CloudWatch logging options.
				mOldCloudWatchLoggingOption := o.([]interface{})[0].(map[string]interface{})

				input := &kinesisanalytics.DeleteApplicationCloudWatchLoggingOptionInput{
					ApplicationName:             aws.String(applicationName),
					CloudWatchLoggingOptionId:   aws.String(mOldCloudWatchLoggingOption[names.AttrID].(string)),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
				}

				_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
					return conn.DeleteApplicationCloudWatchLoggingOption(ctx, input)
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics Application (%s) CloudWatch logging option: %s", d.Id(), err)
				}

				if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
				}

				currentApplicationVersionID += 1
			} else {
				// Update existing CloudWatch logging options.
				mOldCloudWatchLoggingOption := o.([]interface{})[0].(map[string]interface{})
				mNewCloudWatchLoggingOption := n.([]interface{})[0].(map[string]interface{})

				input.ApplicationUpdate.CloudWatchLoggingOptionUpdates = []awstypes.CloudWatchLoggingOptionUpdate{
					{
						CloudWatchLoggingOptionId: aws.String(mOldCloudWatchLoggingOption[names.AttrID].(string)),
						LogStreamARNUpdate:        aws.String(mNewCloudWatchLoggingOption["log_stream_arn"].(string)),
						RoleARNUpdate:             aws.String(mNewCloudWatchLoggingOption[names.AttrRoleARN].(string)),
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
			o, n := d.GetChange("inputs")

			if len(o.([]interface{})) == 0 {
				// Add new input.
				input := &kinesisanalytics.AddApplicationInputInput{
					ApplicationName:             aws.String(applicationName),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
					Input:                       expandInput(n.([]interface{})),
				}

				_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
					return conn.AddApplicationInput(ctx, input)
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics Application (%s) input: %s", d.Id(), err)
				}

				if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
				}

				currentApplicationVersionID += 1
			} else if len(n.([]interface{})) == 0 {
				// The existing input cannot be deleted.
				// This should be handled by the CustomizeDiff function above.
				return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics Application (%s) input", d.Id())
			} else {
				// Update existing input.
				inputUpdate := expandInputUpdate(n.([]interface{}))

				if d.HasChange("inputs.0.processing_configuration") {
					o, n := d.GetChange("inputs.0.processing_configuration")

					// Update of existing input processing configuration is handled via the updating of the existing input.

					if len(o.([]interface{})) == 0 {
						// Add new input processing configuration.
						input := &kinesisanalytics.AddApplicationInputProcessingConfigurationInput{
							ApplicationName:              aws.String(applicationName),
							CurrentApplicationVersionId:  aws.Int64(currentApplicationVersionID),
							InputId:                      inputUpdate.InputId,
							InputProcessingConfiguration: expandInputProcessingConfiguration(n.([]interface{})),
						}

						_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
							return conn.AddApplicationInputProcessingConfiguration(ctx, input)
						})

						if err != nil {
							return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics Application (%s) input processing configuration: %s", d.Id(), err)
						}

						if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
							return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
						}

						currentApplicationVersionID += 1
					} else if len(n.([]interface{})) == 0 {
						// Delete existing input processing configuration.
						input := &kinesisanalytics.DeleteApplicationInputProcessingConfigurationInput{
							ApplicationName:             aws.String(applicationName),
							CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
							InputId:                     inputUpdate.InputId,
						}

						_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
							return conn.DeleteApplicationInputProcessingConfiguration(ctx, input)
						})

						if err != nil {
							return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics Application (%s) input processing configuration: %s", d.Id(), err)
						}

						if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
							return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
						}

						currentApplicationVersionID += 1
					}
				}

				input.ApplicationUpdate.InputUpdates = []awstypes.InputUpdate{inputUpdate}

				updateApplication = true
			}
		}

		if d.HasChange("outputs") {
			o, n := d.GetChange("outputs")
			os := o.(*schema.Set)
			ns := n.(*schema.Set)

			additions := []interface{}{}
			deletions := []string{}

			// Additions.
			for _, vOutput := range ns.Difference(os).List() {
				if outputId, ok := vOutput.(map[string]interface{})[names.AttrID].(string); ok && outputId != "" {
					// Shouldn't be attempting to add an output with an ID.
					log.Printf("[WARN] Attempting to add invalid Kinesis Analytics Application (%s) output: %#v", d.Id(), vOutput)
				} else {
					additions = append(additions, vOutput)
				}
			}

			// Deletions.
			for _, vOutput := range os.Difference(ns).List() {
				if outputId, ok := vOutput.(map[string]interface{})[names.AttrID].(string); ok && outputId != "" {
					deletions = append(deletions, outputId)
				} else {
					// Shouldn't be attempting to delete an output without an ID.
					log.Printf("[WARN] Attempting to delete invalid Kinesis Analytics Application (%s) output: %#v", d.Id(), vOutput)
				}
			}

			// Delete existing outputs.
			for _, outputID := range deletions {
				input := &kinesisanalytics.DeleteApplicationOutputInput{
					ApplicationName:             aws.String(applicationName),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
					OutputId:                    aws.String(outputID),
				}

				_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
					return conn.DeleteApplicationOutput(ctx, input)
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics Application (%s) output: %s", d.Id(), err)
				}

				if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
				}

				currentApplicationVersionID += 1
			}

			// Add new outputs.
			for _, vOutput := range additions {
				input := &kinesisanalytics.AddApplicationOutputInput{
					ApplicationName:             aws.String(applicationName),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
					Output:                      expandOutput(vOutput),
				}

				_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
					return conn.AddApplicationOutput(ctx, input)
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics Application (%s) output: %s", d.Id(), err)
				}

				if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
				}

				currentApplicationVersionID += 1
			}
		}

		if d.HasChange("reference_data_sources") {
			o, n := d.GetChange("reference_data_sources")

			if len(o.([]interface{})) == 0 {
				// Add new reference data source.
				input := &kinesisanalytics.AddApplicationReferenceDataSourceInput{
					ApplicationName:             aws.String(applicationName),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
					ReferenceDataSource:         expandReferenceDataSource(n.([]interface{})),
				}

				_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
					return conn.AddApplicationReferenceDataSource(ctx, input)
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "adding Kinesis Analytics Application (%s) reference data source: %s", d.Id(), err)
				}

				if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
				}

				currentApplicationVersionID += 1
			} else if len(n.([]interface{})) == 0 {
				// Delete existing reference data source.
				mOldReferenceDataSource := o.([]interface{})[0].(map[string]interface{})

				input := &kinesisanalytics.DeleteApplicationReferenceDataSourceInput{
					ApplicationName:             aws.String(applicationName),
					CurrentApplicationVersionId: aws.Int64(currentApplicationVersionID),
					ReferenceId:                 aws.String(mOldReferenceDataSource[names.AttrID].(string)),
				}

				_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
					return conn.DeleteApplicationReferenceDataSource(ctx, input)
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics Application (%s) reference data source: %s", d.Id(), err)
				}

				if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
				}

				currentApplicationVersionID += 1
			} else {
				// Update existing reference data source.
				referenceDataSourceUpdate := expandReferenceDataSourceUpdate(n.([]interface{}))

				input.ApplicationUpdate.ReferenceDataSourceUpdates = []awstypes.ReferenceDataSourceUpdate{referenceDataSourceUpdate}

				updateApplication = true
			}
		}

		if updateApplication {
			input.CurrentApplicationVersionId = aws.Int64(currentApplicationVersionID)

			_, err := waitIAMPropagation(ctx, func() (interface{}, error) {
				return conn.UpdateApplication(ctx, input)
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Kinesis Analytics Application (%s): %s", d.Id(), err)
			}

			if _, err := waitApplicationUpdated(ctx, conn, applicationName); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) update: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("start_application") {
		application, err := findApplicationDetailByName(ctx, conn, d.Get(names.AttrName).(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics Application (%s): %s", d.Id(), err)
		}

		if _, ok := d.GetOk("start_application"); ok {
			if v, ok := d.GetOk("inputs"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				tfMap := v.([]interface{})[0].(map[string]interface{})

				var inputStartingPosition string

				if v, ok := tfMap["starting_position_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
					tfMap := v[0].(map[string]interface{})

					if v, ok := tfMap["starting_position"].(string); ok && v != "" {
						inputStartingPosition = v
					}
				}

				if err := startApplication(ctx, conn, application, inputStartingPosition); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		} else {
			if err := stopApplication(ctx, conn, application); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsClient(ctx)

	createTimestamp, err := time.Parse(time.RFC3339, d.Get("create_timestamp").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	applicationName := d.Get(names.AttrName).(string)

	log.Printf("[DEBUG] Deleting Kinesis Analytics Application (%s)", d.Id())
	_, err = conn.DeleteApplication(ctx, &kinesisanalytics.DeleteApplicationInput{
		ApplicationName: aws.String(applicationName),
		CreateTimestamp: aws.Time(createTimestamp),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics Application (%s): %s", d.Id(), err)
	}

	if _, err := waitApplicationDeleted(ctx, conn, applicationName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics Application (%s) delete: %s", d.Id(), err)
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
		return []*schema.ResourceData{}, fmt.Errorf("Unexpected ARN format: %q", d.Id())
	}

	d.Set(names.AttrName, parts[1])

	return []*schema.ResourceData{d}, nil
}

func startApplication(ctx context.Context, conn *kinesisanalytics.Client, application *awstypes.ApplicationDetail, inputStartingPosition string) error {
	applicationARN := aws.ToString(application.ApplicationARN)
	applicationName := aws.ToString(application.ApplicationName)

	if actual, expected := string(application.ApplicationStatus), string(awstypes.ApplicationStatusReady); actual != expected {
		log.Printf("[DEBUG] Kinesis Analytics Application (%s) has status %s. An application can only be started if it's in the %s state", applicationARN, actual, expected)
		return nil
	}

	if len(application.InputDescriptions) == 0 {
		log.Printf("[DEBUG] Kinesis Analytics Application (%s) has no input description", applicationARN)
		return nil
	}

	input := &kinesisanalytics.StartApplicationInput{
		ApplicationName: aws.String(applicationName),
		InputConfigurations: []awstypes.InputConfiguration{{
			Id:                                 application.InputDescriptions[0].InputId,
			InputStartingPositionConfiguration: &awstypes.InputStartingPositionConfiguration{},
		}},
	}

	if inputStartingPosition != "" {
		input.InputConfigurations[0].InputStartingPositionConfiguration.InputStartingPosition = awstypes.InputStartingPosition(inputStartingPosition)
	}

	if _, err := conn.StartApplication(ctx, input); err != nil {
		return fmt.Errorf("starting Kinesis Analytics Application (%s): %w", applicationARN, err)
	}

	if _, err := waitApplicationStarted(ctx, conn, applicationName); err != nil {
		return fmt.Errorf("waiting for Kinesis Analytics Application (%s) start: %w", applicationARN, err)
	}

	return nil
}

func stopApplication(ctx context.Context, conn *kinesisanalytics.Client, application *awstypes.ApplicationDetail) error {
	applicationARN := aws.ToString(application.ApplicationARN)
	applicationName := aws.ToString(application.ApplicationName)

	if actual, expected := string(application.ApplicationStatus), string(awstypes.ApplicationStatusRunning); actual != expected {
		log.Printf("[DEBUG] Kinesis Analytics Application (%s) has status %s. An application can only be stopped if it's in the %s state", applicationARN, actual, expected)
		return nil
	}

	input := &kinesisanalytics.StopApplicationInput{
		ApplicationName: aws.String(applicationName),
	}

	if _, err := conn.StopApplication(ctx, input); err != nil {
		return fmt.Errorf("stopping Kinesis Analytics Application (%s): %w", applicationARN, err)
	}

	if _, err := waitApplicationStopped(ctx, conn, applicationName); err != nil {
		return fmt.Errorf("waiting for Kinesis Analytics Application (%s) stop: %w", applicationARN, err)
	}

	return nil
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
	if vRoleArn, ok := mCloudWatchLoggingOption[names.AttrRoleARN].(string); ok && vRoleArn != "" {
		cloudWatchLoggingOption.RoleARN = aws.String(vRoleArn)
	}

	return []awstypes.CloudWatchLoggingOption{cloudWatchLoggingOption}
}

func expandInputs(vInputs []interface{}) []awstypes.Input {
	if len(vInputs) == 0 || vInputs[0] == nil {
		return []awstypes.Input{}
	}

	return []awstypes.Input{*expandInput(vInputs)}
}

func expandInput(vInput []interface{}) *awstypes.Input {
	if len(vInput) == 0 || vInput[0] == nil {
		return nil
	}

	input := &awstypes.Input{}

	mInput := vInput[0].(map[string]interface{})

	if vInputParallelism, ok := mInput["parallelism"].([]interface{}); ok && len(vInputParallelism) > 0 && vInputParallelism[0] != nil {
		inputParallelism := &awstypes.InputParallelism{}

		mInputParallelism := vInputParallelism[0].(map[string]interface{})

		if vCount, ok := mInputParallelism["count"].(int); ok {
			inputParallelism.Count = aws.Int32(int32(vCount))
		}

		input.InputParallelism = inputParallelism
	}

	if vInputProcessingConfiguration, ok := mInput["processing_configuration"].([]interface{}); ok {
		input.InputProcessingConfiguration = expandInputProcessingConfiguration(vInputProcessingConfiguration)
	}

	if vInputSchema, ok := mInput[names.AttrSchema].([]interface{}); ok {
		input.InputSchema = expandSourceSchema(vInputSchema)
	}

	if vKinesisFirehoseInput, ok := mInput["kinesis_firehose"].([]interface{}); ok && len(vKinesisFirehoseInput) > 0 && vKinesisFirehoseInput[0] != nil {
		kinesisFirehoseInput := &awstypes.KinesisFirehoseInput{}

		mKinesisFirehoseInput := vKinesisFirehoseInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseInput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisFirehoseInput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisFirehoseInput[names.AttrRoleARN].(string); ok && vRoleArn != "" {
			kinesisFirehoseInput.RoleARN = aws.String(vRoleArn)
		}

		input.KinesisFirehoseInput = kinesisFirehoseInput
	}

	if vKinesisStreamsInput, ok := mInput["kinesis_stream"].([]interface{}); ok && len(vKinesisStreamsInput) > 0 && vKinesisStreamsInput[0] != nil {
		kinesisStreamsInput := &awstypes.KinesisStreamsInput{}

		mKinesisStreamsInput := vKinesisStreamsInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsInput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisStreamsInput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisStreamsInput[names.AttrRoleARN].(string); ok && vRoleArn != "" {
			kinesisStreamsInput.RoleARN = aws.String(vRoleArn)
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

	if vInputLambdaProcessor, ok := mInputProcessingConfiguration["lambda"].([]interface{}); ok && len(vInputLambdaProcessor) > 0 && vInputLambdaProcessor[0] != nil {
		inputLambdaProcessor := &awstypes.InputLambdaProcessor{}

		mInputLambdaProcessor := vInputLambdaProcessor[0].(map[string]interface{})

		if vResourceArn, ok := mInputLambdaProcessor[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			inputLambdaProcessor.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mInputLambdaProcessor[names.AttrRoleARN].(string); ok && vRoleArn != "" {
			inputLambdaProcessor.RoleARN = aws.String(vRoleArn)
		}

		inputProcessingConfiguration.InputLambdaProcessor = inputLambdaProcessor
	}

	return inputProcessingConfiguration
}

func expandInputUpdate(vInput []interface{}) awstypes.InputUpdate {
	inputUpdate := awstypes.InputUpdate{}

	if len(vInput) == 0 || vInput[0] == nil {
		return inputUpdate
	}

	mInput := vInput[0].(map[string]interface{})

	if vInputId, ok := mInput[names.AttrID].(string); ok && vInputId != "" {
		inputUpdate.InputId = aws.String(vInputId)
	}

	if vInputParallelism, ok := mInput["parallelism"].([]interface{}); ok && len(vInputParallelism) > 0 && vInputParallelism[0] != nil {
		inputParallelismUpdate := &awstypes.InputParallelismUpdate{}

		mInputParallelism := vInputParallelism[0].(map[string]interface{})

		if vCount, ok := mInputParallelism["count"].(int); ok {
			inputParallelismUpdate.CountUpdate = aws.Int32(int32(vCount))
		}

		inputUpdate.InputParallelismUpdate = inputParallelismUpdate
	}

	if vInputProcessingConfiguration, ok := mInput["processing_configuration"].([]interface{}); ok && len(vInputProcessingConfiguration) > 0 && vInputProcessingConfiguration[0] != nil {
		inputProcessingConfigurationUpdate := &awstypes.InputProcessingConfigurationUpdate{}

		mInputProcessingConfiguration := vInputProcessingConfiguration[0].(map[string]interface{})

		if vInputLambdaProcessor, ok := mInputProcessingConfiguration["lambda"].([]interface{}); ok && len(vInputLambdaProcessor) > 0 && vInputLambdaProcessor[0] != nil {
			inputLambdaProcessorUpdate := &awstypes.InputLambdaProcessorUpdate{}

			mInputLambdaProcessor := vInputLambdaProcessor[0].(map[string]interface{})

			if vResourceArn, ok := mInputLambdaProcessor[names.AttrResourceARN].(string); ok && vResourceArn != "" {
				inputLambdaProcessorUpdate.ResourceARNUpdate = aws.String(vResourceArn)
			}
			if vRoleArn, ok := mInputLambdaProcessor[names.AttrRoleARN].(string); ok && vRoleArn != "" {
				inputLambdaProcessorUpdate.RoleARNUpdate = aws.String(vRoleArn)
			}

			inputProcessingConfigurationUpdate.InputLambdaProcessorUpdate = inputLambdaProcessorUpdate
		}

		inputUpdate.InputProcessingConfigurationUpdate = inputProcessingConfigurationUpdate
	}

	if vInputSchema, ok := mInput[names.AttrSchema].([]interface{}); ok && len(vInputSchema) > 0 && vInputSchema[0] != nil {
		inputSchemaUpdate := &awstypes.InputSchemaUpdate{}

		mInputSchema := vInputSchema[0].(map[string]interface{})

		if vRecordColumns, ok := mInputSchema["record_columns"].([]interface{}); ok {
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

	if vKinesisFirehoseInput, ok := mInput["kinesis_firehose"].([]interface{}); ok && len(vKinesisFirehoseInput) > 0 && vKinesisFirehoseInput[0] != nil {
		kinesisFirehoseInputUpdate := &awstypes.KinesisFirehoseInputUpdate{}

		mKinesisFirehoseInput := vKinesisFirehoseInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseInput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisFirehoseInputUpdate.ResourceARNUpdate = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisFirehoseInput[names.AttrRoleARN].(string); ok && vRoleArn != "" {
			kinesisFirehoseInputUpdate.RoleARNUpdate = aws.String(vRoleArn)
		}

		inputUpdate.KinesisFirehoseInputUpdate = kinesisFirehoseInputUpdate
	}

	if vKinesisStreamsInput, ok := mInput["kinesis_stream"].([]interface{}); ok && len(vKinesisStreamsInput) > 0 && vKinesisStreamsInput[0] != nil {
		kinesisStreamsInputUpdate := &awstypes.KinesisStreamsInputUpdate{}

		mKinesisStreamsInput := vKinesisStreamsInput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsInput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisStreamsInputUpdate.ResourceARNUpdate = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisStreamsInput[names.AttrRoleARN].(string); ok && vRoleArn != "" {
			kinesisStreamsInputUpdate.RoleARNUpdate = aws.String(vRoleArn)
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

	if vDestinationSchema, ok := mOutput[names.AttrSchema].([]interface{}); ok && len(vDestinationSchema) > 0 && vDestinationSchema[0] != nil {
		destinationSchema := &awstypes.DestinationSchema{}

		mDestinationSchema := vDestinationSchema[0].(map[string]interface{})

		if vRecordFormatType, ok := mDestinationSchema["record_format_type"].(string); ok && vRecordFormatType != "" {
			destinationSchema.RecordFormatType = awstypes.RecordFormatType(vRecordFormatType)
		}

		output.DestinationSchema = destinationSchema
	}

	if vKinesisFirehoseOutput, ok := mOutput["kinesis_firehose"].([]interface{}); ok && len(vKinesisFirehoseOutput) > 0 && vKinesisFirehoseOutput[0] != nil {
		kinesisFirehoseOutput := &awstypes.KinesisFirehoseOutput{}

		mKinesisFirehoseOutput := vKinesisFirehoseOutput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisFirehoseOutput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisFirehoseOutput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisFirehoseOutput[names.AttrRoleARN].(string); ok && vRoleArn != "" {
			kinesisFirehoseOutput.RoleARN = aws.String(vRoleArn)
		}

		output.KinesisFirehoseOutput = kinesisFirehoseOutput
	}

	if vKinesisStreamsOutput, ok := mOutput["kinesis_stream"].([]interface{}); ok && len(vKinesisStreamsOutput) > 0 && vKinesisStreamsOutput[0] != nil {
		kinesisStreamsOutput := &awstypes.KinesisStreamsOutput{}

		mKinesisStreamsOutput := vKinesisStreamsOutput[0].(map[string]interface{})

		if vResourceArn, ok := mKinesisStreamsOutput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			kinesisStreamsOutput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mKinesisStreamsOutput[names.AttrRoleARN].(string); ok && vRoleArn != "" {
			kinesisStreamsOutput.RoleARN = aws.String(vRoleArn)
		}

		output.KinesisStreamsOutput = kinesisStreamsOutput
	}

	if vLambdaOutput, ok := mOutput["lambda"].([]interface{}); ok && len(vLambdaOutput) > 0 && vLambdaOutput[0] != nil {
		lambdaOutput := &awstypes.LambdaOutput{}

		mLambdaOutput := vLambdaOutput[0].(map[string]interface{})

		if vResourceArn, ok := mLambdaOutput[names.AttrResourceARN].(string); ok && vResourceArn != "" {
			lambdaOutput.ResourceARN = aws.String(vResourceArn)
		}
		if vRoleArn, ok := mLambdaOutput[names.AttrRoleARN].(string); ok && vRoleArn != "" {
			lambdaOutput.RoleARN = aws.String(vRoleArn)
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
			outputs = append(outputs, *output)
		}
	}

	return outputs
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

		if vCsvMappingParameters, ok := mMappingParameters["csv"].([]interface{}); ok && len(vCsvMappingParameters) > 0 && vCsvMappingParameters[0] != nil {
			csvMappingParameters := &awstypes.CSVMappingParameters{}

			mCsvMappingParameters := vCsvMappingParameters[0].(map[string]interface{})

			if vRecordColumnDelimiter, ok := mCsvMappingParameters["record_column_delimiter"].(string); ok && vRecordColumnDelimiter != "" {
				csvMappingParameters.RecordColumnDelimiter = aws.String(vRecordColumnDelimiter)
			}
			if vRecordRowDelimiter, ok := mCsvMappingParameters["record_row_delimiter"].(string); ok && vRecordRowDelimiter != "" {
				csvMappingParameters.RecordRowDelimiter = aws.String(vRecordRowDelimiter)
			}

			mappingParameters.CSVMappingParameters = csvMappingParameters

			recordFormat.RecordFormatType = awstypes.RecordFormatTypeCsv
		}

		if vJsonMappingParameters, ok := mMappingParameters[names.AttrJSON].([]interface{}); ok && len(vJsonMappingParameters) > 0 && vJsonMappingParameters[0] != nil {
			jsonMappingParameters := &awstypes.JSONMappingParameters{}

			mJsonMappingParameters := vJsonMappingParameters[0].(map[string]interface{})

			if vRecordRowPath, ok := mJsonMappingParameters["record_row_path"].(string); ok && vRecordRowPath != "" {
				jsonMappingParameters.RecordRowPath = aws.String(vRecordRowPath)
			}

			mappingParameters.JSONMappingParameters = jsonMappingParameters

			recordFormat.RecordFormatType = awstypes.RecordFormatTypeJson
		}

		recordFormat.MappingParameters = mappingParameters
	}

	return recordFormat
}

func expandReferenceDataSource(vReferenceDataSource []interface{}) *awstypes.ReferenceDataSource {
	if len(vReferenceDataSource) == 0 || vReferenceDataSource[0] == nil {
		return nil
	}

	referenceDataSource := &awstypes.ReferenceDataSource{}

	mReferenceDataSource := vReferenceDataSource[0].(map[string]interface{})

	if vReferenceSchema, ok := mReferenceDataSource[names.AttrSchema].([]interface{}); ok {
		referenceDataSource.ReferenceSchema = expandSourceSchema(vReferenceSchema)
	}

	if vS3ReferenceDataSource, ok := mReferenceDataSource["s3"].([]interface{}); ok && len(vS3ReferenceDataSource) > 0 && vS3ReferenceDataSource[0] != nil {
		s3ReferenceDataSource := &awstypes.S3ReferenceDataSource{}

		mS3ReferenceDataSource := vS3ReferenceDataSource[0].(map[string]interface{})

		if vBucketArn, ok := mS3ReferenceDataSource["bucket_arn"].(string); ok && vBucketArn != "" {
			s3ReferenceDataSource.BucketARN = aws.String(vBucketArn)
		}
		if vFileKey, ok := mS3ReferenceDataSource["file_key"].(string); ok && vFileKey != "" {
			s3ReferenceDataSource.FileKey = aws.String(vFileKey)
		}
		if vReferenceRoleArn, ok := mS3ReferenceDataSource[names.AttrRoleARN].(string); ok && vReferenceRoleArn != "" {
			s3ReferenceDataSource.ReferenceRoleARN = aws.String(vReferenceRoleArn)
		}

		referenceDataSource.S3ReferenceDataSource = s3ReferenceDataSource
	}

	if vTableName, ok := mReferenceDataSource[names.AttrTableName].(string); ok && vTableName != "" {
		referenceDataSource.TableName = aws.String(vTableName)
	}

	return referenceDataSource
}

func expandReferenceDataSourceUpdate(vReferenceDataSource []interface{}) awstypes.ReferenceDataSourceUpdate {
	referenceDataSourceUpdate := awstypes.ReferenceDataSourceUpdate{}

	if len(vReferenceDataSource) == 0 || vReferenceDataSource[0] == nil {
		return referenceDataSourceUpdate
	}

	mReferenceDataSource := vReferenceDataSource[0].(map[string]interface{})

	if vReferenceId, ok := mReferenceDataSource[names.AttrID].(string); ok && vReferenceId != "" {
		referenceDataSourceUpdate.ReferenceId = aws.String(vReferenceId)
	}

	if vReferenceSchema, ok := mReferenceDataSource[names.AttrSchema].([]interface{}); ok {
		referenceDataSourceUpdate.ReferenceSchemaUpdate = expandSourceSchema(vReferenceSchema)
	}

	if vS3ReferenceDataSource, ok := mReferenceDataSource["s3"].([]interface{}); ok && len(vS3ReferenceDataSource) > 0 && vS3ReferenceDataSource[0] != nil {
		s3ReferenceDataSourceUpdate := &awstypes.S3ReferenceDataSourceUpdate{}

		mS3ReferenceDataSource := vS3ReferenceDataSource[0].(map[string]interface{})

		if vBucketArn, ok := mS3ReferenceDataSource["bucket_arn"].(string); ok && vBucketArn != "" {
			s3ReferenceDataSourceUpdate.BucketARNUpdate = aws.String(vBucketArn)
		}
		if vFileKey, ok := mS3ReferenceDataSource["file_key"].(string); ok && vFileKey != "" {
			s3ReferenceDataSourceUpdate.FileKeyUpdate = aws.String(vFileKey)
		}
		if vReferenceRoleArn, ok := mS3ReferenceDataSource[names.AttrRoleARN].(string); ok && vReferenceRoleArn != "" {
			s3ReferenceDataSourceUpdate.ReferenceRoleARNUpdate = aws.String(vReferenceRoleArn)
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

	if vRecordColumns, ok := mSourceSchema["record_columns"].([]interface{}); ok {
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

func flattenCloudWatchLoggingOptionDescriptions(cloudWatchLoggingOptionDescriptions []awstypes.CloudWatchLoggingOptionDescription) []interface{} {
	if len(cloudWatchLoggingOptionDescriptions) == 0 {
		return []interface{}{}
	}

	cloudWatchLoggingOptionDescription := cloudWatchLoggingOptionDescriptions[0]

	mCloudWatchLoggingOption := map[string]interface{}{
		names.AttrID:      aws.ToString(cloudWatchLoggingOptionDescription.CloudWatchLoggingOptionId),
		"log_stream_arn":  aws.ToString(cloudWatchLoggingOptionDescription.LogStreamARN),
		names.AttrRoleARN: aws.ToString(cloudWatchLoggingOptionDescription.RoleARN),
	}

	return []interface{}{mCloudWatchLoggingOption}
}

func flattenInputDescriptions(inputDescriptions []awstypes.InputDescription) []interface{} {
	if len(inputDescriptions) == 0 {
		return []interface{}{}
	}

	inputDescription := inputDescriptions[0]

	mInput := map[string]interface{}{
		names.AttrID:         aws.ToString(inputDescription.InputId),
		names.AttrNamePrefix: aws.ToString(inputDescription.NamePrefix),
		"stream_names":       flex.FlattenStringValueList(inputDescription.InAppStreamNames),
	}

	if inputParallelism := inputDescription.InputParallelism; inputParallelism != nil {
		mInputParallelism := map[string]interface{}{
			"count": int(aws.ToInt32(inputParallelism.Count)),
		}

		mInput["parallelism"] = []interface{}{mInputParallelism}
	}

	if inputSchema := inputDescription.InputSchema; inputSchema != nil {
		mInput[names.AttrSchema] = flattenSourceSchema(inputSchema)
	}

	if inputProcessingConfigurationDescription := inputDescription.InputProcessingConfigurationDescription; inputProcessingConfigurationDescription != nil {
		mInputProcessingConfiguration := map[string]interface{}{}

		if inputLambdaProcessorDescription := inputProcessingConfigurationDescription.InputLambdaProcessorDescription; inputLambdaProcessorDescription != nil {
			mInputLambdaProcessor := map[string]interface{}{
				names.AttrResourceARN: aws.ToString(inputLambdaProcessorDescription.ResourceARN),
				names.AttrRoleARN:     aws.ToString(inputLambdaProcessorDescription.RoleARN),
			}

			mInputProcessingConfiguration["lambda"] = []interface{}{mInputLambdaProcessor}
		}

		mInput["processing_configuration"] = []interface{}{mInputProcessingConfiguration}
	}

	if inputStartingPositionConfiguration := inputDescription.InputStartingPositionConfiguration; inputStartingPositionConfiguration != nil {
		mInputStartingPositionConfiguration := map[string]interface{}{
			"starting_position": string(inputStartingPositionConfiguration.InputStartingPosition),
		}

		mInput["starting_position_configuration"] = []interface{}{mInputStartingPositionConfiguration}
	}

	if kinesisFirehoseInputDescription := inputDescription.KinesisFirehoseInputDescription; kinesisFirehoseInputDescription != nil {
		mKinesisFirehoseInput := map[string]interface{}{
			names.AttrResourceARN: aws.ToString(kinesisFirehoseInputDescription.ResourceARN),
			names.AttrRoleARN:     aws.ToString(kinesisFirehoseInputDescription.RoleARN),
		}

		mInput["kinesis_firehose"] = []interface{}{mKinesisFirehoseInput}
	}

	if kinesisStreamsInputDescription := inputDescription.KinesisStreamsInputDescription; kinesisStreamsInputDescription != nil {
		mKinesisStreamsInput := map[string]interface{}{
			names.AttrResourceARN: aws.ToString(kinesisStreamsInputDescription.ResourceARN),
			names.AttrRoleARN:     aws.ToString(kinesisStreamsInputDescription.RoleARN),
		}

		mInput["kinesis_stream"] = []interface{}{mKinesisStreamsInput}
	}

	return []interface{}{mInput}
}

func flattenOutputDescriptions(outputDescriptions []awstypes.OutputDescription) []interface{} {
	if len(outputDescriptions) == 0 {
		return []interface{}{}
	}

	vOutputs := []interface{}{}

	for _, outputDescription := range outputDescriptions {
		mOutput := map[string]interface{}{
			names.AttrID:   aws.ToString(outputDescription.OutputId),
			names.AttrName: aws.ToString(outputDescription.Name),
		}

		if destinationSchema := outputDescription.DestinationSchema; destinationSchema != nil {
			mDestinationSchema := map[string]interface{}{
				"record_format_type": string(destinationSchema.RecordFormatType),
			}

			mOutput[names.AttrSchema] = []interface{}{mDestinationSchema}
		}

		if kinesisFirehoseOutputDescription := outputDescription.KinesisFirehoseOutputDescription; kinesisFirehoseOutputDescription != nil {
			mKinesisFirehoseOutput := map[string]interface{}{
				names.AttrResourceARN: aws.ToString(kinesisFirehoseOutputDescription.ResourceARN),
				names.AttrRoleARN:     aws.ToString(kinesisFirehoseOutputDescription.RoleARN),
			}

			mOutput["kinesis_firehose"] = []interface{}{mKinesisFirehoseOutput}
		}

		if kinesisStreamsOutputDescription := outputDescription.KinesisStreamsOutputDescription; kinesisStreamsOutputDescription != nil {
			mKinesisStreamsOutput := map[string]interface{}{
				names.AttrResourceARN: aws.ToString(kinesisStreamsOutputDescription.ResourceARN),
				names.AttrRoleARN:     aws.ToString(kinesisStreamsOutputDescription.RoleARN),
			}

			mOutput["kinesis_stream"] = []interface{}{mKinesisStreamsOutput}
		}

		if lambdaOutputDescription := outputDescription.LambdaOutputDescription; lambdaOutputDescription != nil {
			mLambdaOutput := map[string]interface{}{
				names.AttrResourceARN: aws.ToString(lambdaOutputDescription.ResourceARN),
				names.AttrRoleARN:     aws.ToString(lambdaOutputDescription.RoleARN),
			}

			mOutput["lambda"] = []interface{}{mLambdaOutput}
		}

		vOutputs = append(vOutputs, mOutput)
	}

	return vOutputs
}

func flattenReferenceDataSourceDescriptions(referenceDataSourceDescriptions []awstypes.ReferenceDataSourceDescription) []interface{} {
	if len(referenceDataSourceDescriptions) == 0 {
		return []interface{}{}
	}

	referenceDataSourceDescription := referenceDataSourceDescriptions[0]

	mReferenceDataSource := map[string]interface{}{
		names.AttrID:        aws.ToString(referenceDataSourceDescription.ReferenceId),
		names.AttrTableName: aws.ToString(referenceDataSourceDescription.TableName),
	}

	if referenceSchema := referenceDataSourceDescription.ReferenceSchema; referenceSchema != nil {
		mReferenceDataSource[names.AttrSchema] = flattenSourceSchema(referenceSchema)
	}

	if s3ReferenceDataSource := referenceDataSourceDescription.S3ReferenceDataSourceDescription; s3ReferenceDataSource != nil {
		mS3ReferenceDataSource := map[string]interface{}{
			"bucket_arn":      aws.ToString(s3ReferenceDataSource.BucketARN),
			"file_key":        aws.ToString(s3ReferenceDataSource.FileKey),
			names.AttrRoleARN: aws.ToString(s3ReferenceDataSource.ReferenceRoleARN),
		}

		mReferenceDataSource["s3"] = []interface{}{mS3ReferenceDataSource}
	}

	return []interface{}{mReferenceDataSource}
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

		mSourceSchema["record_columns"] = vRecordColumns
	}

	if recordFormat := sourceSchema.RecordFormat; recordFormat != nil {
		mRecordFormat := map[string]interface{}{
			"record_format_type": string(recordFormat.RecordFormatType),
		}

		if mappingParameters := recordFormat.MappingParameters; mappingParameters != nil {
			mMappingParameters := map[string]interface{}{}

			if csvMappingParameters := mappingParameters.CSVMappingParameters; csvMappingParameters != nil {
				mCsvMappingParameters := map[string]interface{}{
					"record_column_delimiter": aws.ToString(csvMappingParameters.RecordColumnDelimiter),
					"record_row_delimiter":    aws.ToString(csvMappingParameters.RecordRowDelimiter),
				}

				mMappingParameters["csv"] = []interface{}{mCsvMappingParameters}
			}

			if jsonMappingParameters := mappingParameters.JSONMappingParameters; jsonMappingParameters != nil {
				mJsonMappingParameters := map[string]interface{}{
					"record_row_path": aws.ToString(jsonMappingParameters.RecordRowPath),
				}

				mMappingParameters[names.AttrJSON] = []interface{}{mJsonMappingParameters}
			}

			mRecordFormat["mapping_parameters"] = []interface{}{mMappingParameters}
		}

		mSourceSchema["record_format"] = []interface{}{mRecordFormat}
	}

	return []interface{}{mSourceSchema}
}
