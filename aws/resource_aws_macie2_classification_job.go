package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
	"time"
)

const (
	errorMacie2ClassificationJobCreate   = "error creating Macie2 ClassificationJob: %s"
	errorMacie2ClassificationJobRead     = "error reading Macie2 ClassificationJob (%s): %w"
	errorMacie2ClassificationJobUpdating = "error updating Macie2 ClassificationJob (%s): %w"
	errorMacie2ClassificationJobSetting  = "error setting `%s` for Macie2 ClassificationJob (%s): %s"
)

func resourceAwsMacie2ClassificationJob() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMacie2ClassificationJobCreate,
		ReadContext:   resourceMacie2ClassificationJobRead,
		UpdateContext: resourceMacie2ClassificationJobUpdate,
		DeleteContext: resourceMacie2ClassificationJobDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"custom_data_identifier_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"schedule_frequency": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"daily_schedule": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"weekly_schedule": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"monthly_schedule": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"sampling_percentage": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"client_token": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"initial_run": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"job_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ONE_TIME", "SCHEDULED"}, false),
			},
			"s3_job_definition": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_definitions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"buckets": {
										Type:     schema.TypeList,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"scoping": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"excludes": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"and": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"simple_scope_term": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			Computed:     true,
																			ValidateFunc: validation.StringInSlice([]string{"EQ", "GT", "GTE", "LT", "LTE", "NE", "CONTAINS", "STARTS_WITH"}, false),
																		},
																		"values": {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		"key": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			Computed:     true,
																			ValidateFunc: validation.StringInSlice([]string{"BUCKET_CREATION_DATE", "OBJECT_EXTENSION", "OBJECT_LAST_MODIFIED_DATE", "OBJECT_SIZE", "TAG", "OBJECT_KEY"}, false),
																		},
																	},
																},
															},
															"tag_scope_term": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			Computed:     true,
																			ValidateFunc: validation.StringInSlice([]string{"EQ", "GT", "GTE", "LT", "LTE", "NE", "CONTAINS", "STARTS_WITH"}, false),
																		},
																		"tag_values": {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"value": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																					"key": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																		},
																		"key": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																		"target": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			Computed:     true,
																			ValidateFunc: validation.StringInSlice([]string{"S3_OBJECT"}, false),
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
									"includes": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"and": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"simple_scope_term": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																		"values": {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		"key": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																	},
																},
															},
															"tag_scope_term": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																		"tag_values": {
																			Type:     schema.TypeList,
																			Optional: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"value": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																					"key": {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																		},
																		"key": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																		"target": {
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
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchemaComputed(),
			"job_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"RUNNING", "CANCELLED", "USER_PAUSED"}, false),
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_run_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_paused_details": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"job_imminent_expiration_health_event_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"job_expires_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"job_paused_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"last_run_error_status": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"statistics": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number_of_runs": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"approximate_number_of_objects_to_process": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceMacie2ClassificationJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.CreateClassificationJobInput{
		ClientToken:     aws.String(d.Get("client_token").(string)),
		Name:            aws.String(d.Get("name").(string)),
		JobType:         aws.String(d.Get("job_type").(string)),
		S3JobDefinition: expandS3JobDefinition(d),
	}

	if v, ok := d.GetOk("custom_data_identifier_ids"); ok {
		input.SetCustomDataIdentifierIds(expandStringList(v.([]interface{})))
	}
	if v, ok := d.GetOk("schedule_frequency"); ok {
		input.SetScheduleFrequency(expandScheduleFrequency(v.([]interface{})))
	}
	if v, ok := d.GetOk("sampling_percentage"); ok {
		input.SetSamplingPercentage(int64(v.(int)))
	}
	if v, ok := d.GetOk("description"); ok {
		input.SetDescription(v.(string))
	}
	if v, ok := d.GetOk("initial_run"); ok {
		input.SetInitialRun(v.(bool))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.SetTags(keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().AppsyncTags())
	}

	log.Printf("[DEBUG] Creating Macie2 ClassificationJob: %v", input)

	var err error
	var output macie2.CreateClassificationJobOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		resp, err := conn.CreateClassificationJobWithContext(ctx, input)
		if err != nil {
			if isAWSErr(err, macie2.ErrorCodeClientError, "") {
				log.Printf(errorMacie2ClassificationJobCreate, err)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		output = *resp

		return nil
	})

	if isResourceTimeoutError(err) {
		_, _ = conn.CreateClassificationJobWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobCreate, err))
	}

	d.SetId(aws.StringValue(output.JobId))

	return resourceMacie2ClassificationJobRead(ctx, d, meta)
}

func resourceMacie2ClassificationJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	input := &macie2.DescribeClassificationJobInput{
		JobId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Macie2 ClassificationJob: %s", input)
	resp, err := conn.DescribeClassificationJobWithContext(ctx, input)
	if err != nil {
		if isAWSErr(err, macie2.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Macie2 ClassificationJob does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobRead, d.Id(), err))
	}

	if err = d.Set("custom_data_identifier_ids", flattenStringList(resp.CustomDataIdentifierIds)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "custom_data_identifier_ids", d.Id(), err))
	}
	if err = d.Set("schedule_frequency", flattenScheduleFrequency(resp.ScheduleFrequency)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "schedule_frequency", d.Id(), err))
	}
	if err = d.Set("sampling_percentage", aws.Int64Value(resp.SamplingPercentage)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "sampling_percentage", d.Id(), err))
	}
	if err = d.Set("client_token", aws.StringValue(resp.ClientToken)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "client_token", d.Id(), err))
	}
	if err = d.Set("name", aws.StringValue(resp.Name)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "name", d.Id(), err))
	}
	if err = d.Set("description", aws.StringValue(resp.Description)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "description", d.Id(), err))
	}
	if err = d.Set("initial_run", aws.BoolValue(resp.InitialRun)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "initial_run", d.Id(), err))
	}
	if err = d.Set("job_type", aws.StringValue(resp.JobType)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "job_type", d.Id(), err))
	}
	if err = d.Set("s3_job_definition", flattenS3JobDefinition(resp.S3JobDefinition)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "s3_job_definition", d.Id(), err))
	}
	if err = d.Set("tags", keyvaluetags.AppsyncKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "tags", d.Id(), err))
	}
	if err = d.Set("job_id", aws.StringValue(resp.JobId)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "job_id", d.Id(), err))
	}
	if err = d.Set("job_arn", aws.StringValue(resp.JobArn)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "job_arn", d.Id(), err))
	}
	if err = d.Set("job_status", aws.StringValue(resp.JobStatus)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "job_status", d.Id(), err))
	}
	if err = d.Set("created_at", resp.CreatedAt.String()); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "created_at", d.Id(), err))
	}
	if err = d.Set("last_run_time", resp.LastRunTime.String()); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "last_run_time", d.Id(), err))
	}
	if err = d.Set("user_paused_details", flattenUserPausedDetails(resp.UserPausedDetails)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "user_paused_details", d.Id(), err))
	}
	if err = d.Set("last_run_error_status", flattenLastRunErrorStatus(resp.LastRunErrorStatus)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "last_run_error_status", d.Id(), err))
	}
	if err = d.Set("statistics", flattenStatistics(resp.Statistics)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobSetting, "statistics", d.Id(), err))
	}

	return nil
}

func resourceMacie2ClassificationJobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.UpdateClassificationJobInput{
		JobId: aws.String(d.Id()),
	}

	if d.HasChange("job_status") {
		input.SetJobStatus(d.Get("job_status").(string))
	}

	log.Printf("[DEBUG] Updating Macie2 ClassificationJob: %s", input)
	_, err := conn.UpdateClassificationJobWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2ClassificationJobUpdating, d.Id(), err))
	}

	return resourceMacie2ClassificationJobRead(ctx, d, meta)
}

func resourceMacie2ClassificationJobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func expandS3JobDefinition(d *schema.ResourceData) *macie2.S3JobDefinition {
	var s3JobDefinition macie2.S3JobDefinition

	if v, ok := d.GetOk("s3_job_definition"); ok {
		s3Job := v.([]interface{})
		if len(s3Job) > 0 {
			if s3Job[0] != nil {
				s3JobMap := s3Job[0].(map[string]interface{})

				if v1, ok1 := s3JobMap["bucket_definitions"]; ok1 && len(v1.([]interface{})) > 0 {
					s3JobDefinition.SetBucketDefinitions(expandBucketDefinitions(v1.([]interface{})))
				}
				if v1, ok1 := s3JobMap["scoping"]; ok1 && len(v1.([]interface{})) > 0 {
					s3JobDefinition.SetScoping(expandScoping(v1.([]interface{})))
				}
			}
		}
	}

	return &s3JobDefinition
}

func expandBucketDefinitions(definitions []interface{}) []*macie2.S3BucketDefinitionForJob {
	bucketDefinitions := make([]*macie2.S3BucketDefinitionForJob, len(definitions))

	for i, v := range definitions {
		v1 := v.(map[string]interface{})

		var bucketDefinition macie2.S3BucketDefinitionForJob

		bucketDefinition.Buckets = expandStringList(v1["buckets"].([]interface{}))
		bucketDefinition.AccountId = aws.String(v1["account_id"].(string))

		bucketDefinitions[i] = &bucketDefinition
	}

	return bucketDefinitions
}

func expandScoping(scoping []interface{}) *macie2.Scoping {
	var scopingObj macie2.Scoping

	if len(scoping) > 0 {
		scopingMap := scoping[0].(map[string]interface{})

		if v, ok := scopingMap["excludes"]; ok && len(v.([]interface{})) > 0 {
			v1 := v.([]interface{})
			andMap := v1[0].(map[string]interface{})
			if v2, ok1 := andMap["and"]; ok1 && len(v2.([]interface{})) > 0 {
				scopingObj.SetExcludes(&macie2.JobScopingBlock{
					And: expandJobScopeTerm(v2.([]interface{})),
				})
			}
		}
		if v, ok := scopingMap["includes"]; ok && len(v.([]interface{})) > 0 {
			v1 := v.([]interface{})
			andMap := v1[0].(map[string]interface{})
			if v2, ok1 := andMap["and"]; ok1 && len(v2.([]interface{})) > 0 {
				scopingObj.SetIncludes(&macie2.JobScopingBlock{
					And: expandJobScopeTerm(v2.([]interface{})),
				})
			}
		}
	}

	return &scopingObj
}

func expandJobScopeTerm(scopeTerms []interface{}) []*macie2.JobScopeTerm {
	scopeTermsList := make([]*macie2.JobScopeTerm, len(scopeTerms))

	for i, v := range scopeTerms {
		v1 := v.(map[string]interface{})
		var scopeTerm macie2.JobScopeTerm

		if v2, ok1 := v1["simple_scope_term"]; ok1 && len(v2.([]interface{})) > 0 {
			scopeTerm.SetSimpleScopeTerm(expandSimpleScopeTerm(v2.([]interface{})))
		}
		if v2, ok1 := v1["tag_scope_term"]; ok1 && len(v2.([]interface{})) > 0 {
			scopeTerm.SetTagScopeTerm(expandTagScopeTerm(v2.([]interface{})))
		}
		scopeTermsList[i] = &scopeTerm
	}

	return scopeTermsList
}

func expandSimpleScopeTerm(simpleScopeTerm []interface{}) *macie2.SimpleScopeTerm {
	var simpleTerm macie2.SimpleScopeTerm

	if len(simpleScopeTerm) > 0 {
		simpleScopeTermMap := simpleScopeTerm[0].(map[string]interface{})

		if v, ok := simpleScopeTermMap["key"]; ok && v.(string) != "" {
			simpleTerm.SetKey(v.(string))
		}
		if v, ok := simpleScopeTermMap["values"]; ok && len(v.([]interface{})) > 0 {
			simpleTerm.SetValues(expandStringList(v.([]interface{})))
		}
		if v, ok := simpleScopeTermMap["comparator"]; ok && v.(string) != "" {
			simpleTerm.SetComparator(v.(string))
		}
	}

	return &simpleTerm
}

func expandTagScopeTerm(tagScopeTerm []interface{}) *macie2.TagScopeTerm {
	var tagTerm macie2.TagScopeTerm

	if len(tagScopeTerm) > 0 {
		tagScopeTermMap := tagScopeTerm[0].(map[string]interface{})

		if v, ok := tagScopeTermMap["key"]; ok && v.(string) != "" {
			tagTerm.SetKey(v.(string))
		}
		if v, ok := tagScopeTermMap["tag_values"]; ok && len(v.([]interface{})) > 0 {
			tagTerm.SetTagValues(expandTagValues(v.([]interface{})))
		}
		if v, ok := tagScopeTermMap["comparator"]; ok && v.(string) != "" {
			tagTerm.SetComparator(v.(string))
		}
		if v, ok := tagScopeTermMap["target"]; ok && v.(string) != "" {
			tagTerm.SetTarget(v.(string))
		}
	}

	return &tagTerm
}

func expandTagValues(tagValues []interface{}) []*macie2.TagValuePair {
	tagValuesList := make([]*macie2.TagValuePair, len(tagValues))

	for k, v := range tagValues {
		v1 := v.(map[string]interface{})
		var tagValue macie2.TagValuePair

		if v2, ok := v1["value"]; ok && v2.(string) != "" {
			tagValue.SetValue(v2.(string))
		}
		if v2, ok := v1["key"]; ok && v2.(string) != "" {
			tagValue.SetKey(v2.(string))
		}
		tagValuesList[k] = &tagValue
	}

	return tagValuesList
}

func expandScheduleFrequency(schedules []interface{}) *macie2.JobScheduleFrequency {
	var jobScheduleFrequency macie2.JobScheduleFrequency

	if len(schedules) > 0 {
		scheduleMap := schedules[0].(map[string]interface{})

		if v1, ok1 := scheduleMap["daily_schedule"]; ok1 && v1.(bool) {
			jobScheduleFrequency.SetDailySchedule(&macie2.DailySchedule{})
		}
		if v1, ok1 := scheduleMap["weekly_schedule"]; ok1 && v1.(string) != "" {
			jobScheduleFrequency.SetWeeklySchedule(&macie2.WeeklySchedule{
				DayOfWeek: aws.String(v1.(string)),
			})
		}
		if v1, ok1 := scheduleMap["monthly_schedule"]; ok1 && v1.(int) > 0 {
			jobScheduleFrequency.SetMonthlySchedule(&macie2.MonthlySchedule{
				DayOfMonth: aws.Int64(int64(v1.(int))),
			})
		}
	}

	return &jobScheduleFrequency
}

func flattenScheduleFrequency(schedule *macie2.JobScheduleFrequency) []map[string]interface{} {
	schedulesList := make([]map[string]interface{}, 0)
	if schedule != nil {
		schedMap := map[string]interface{}{}
		if schedule.DailySchedule != nil {
			schedMap["daily_schedule"] = true
		}
		if schedule.WeeklySchedule != nil && schedule.WeeklySchedule.DayOfWeek != nil {
			schedMap["weekly_schedule"] = schedule.WeeklySchedule.DayOfWeek
		}
		if schedule.MonthlySchedule != nil && schedule.MonthlySchedule.DayOfMonth != nil {
			schedMap["monthly_schedule"] = schedule.MonthlySchedule.DayOfMonth
		}
		schedulesList = append(schedulesList, schedMap)
	}

	return schedulesList
}

func flattenS3JobDefinition(s3JobDefinition *macie2.S3JobDefinition) []map[string]interface{} {
	jobDefinitions := make([]map[string]interface{}, 0)

	if s3JobDefinition != nil {
		jobDefinitions = append(jobDefinitions, map[string]interface{}{
			"bucket_definitions": flattenBucketDefinition(s3JobDefinition.BucketDefinitions),
			"scoping":            flattenScoping(s3JobDefinition.Scoping),
		})
	}

	return jobDefinitions
}

func flattenBucketDefinition(bucketDefinitions []*macie2.S3BucketDefinitionForJob) []map[string]interface{} {
	bucketDefinitionList := make([]map[string]interface{}, 0)

	if bucketDefinitions != nil {

		for _, bucket := range bucketDefinitions {
			bucketDefinitionList = append(bucketDefinitionList, map[string]interface{}{
				"account_id": aws.StringValue(bucket.AccountId),
				"buckets":    flattenStringList(bucket.Buckets),
			})
		}
	}

	return bucketDefinitionList
}

func flattenScoping(scoping *macie2.Scoping) []map[string]interface{} {
	scopingList := make([]map[string]interface{}, 0)

	if scoping != nil {
		scopingList = append(scopingList, map[string]interface{}{
			"excludes": flattenJobScopingBlock(scoping.Excludes),
			"includes": flattenJobScopingBlock(scoping.Includes),
		})
	}

	return scopingList
}

func flattenJobScopingBlock(scopeTerm *macie2.JobScopingBlock) []map[string]interface{} {
	scopeTermList := make([]map[string]interface{}, 0)

	if scopeTerm != nil {
		scopeTermList = append(scopeTermList, map[string]interface{}{
			"and": flattenJobScopeTerm(scopeTerm.And),
		})
	}

	return scopeTermList
}

func flattenJobScopeTerm(scopeTerms []*macie2.JobScopeTerm) []map[string]interface{} {
	scopeTermList := make([]map[string]interface{}, 0)

	if scopeTerms != nil {
		for _, scopeTerm := range scopeTerms {
			scopeTermList = append(scopeTermList, map[string]interface{}{
				"simple_scope_term": flattenSimpleScopeTerm(scopeTerm.SimpleScopeTerm),
				"tag_scope_term":    flattenTagScopeTerm(scopeTerm.TagScopeTerm),
			})
		}
	}

	return scopeTermList
}

func flattenSimpleScopeTerm(simpleScopeTerm *macie2.SimpleScopeTerm) []map[string]interface{} {
	simpleScopeTermList := make([]map[string]interface{}, 0)

	if simpleScopeTerm != nil {
		simpleScopeTermList = append(simpleScopeTermList, map[string]interface{}{
			"key":        aws.StringValue(simpleScopeTerm.Key),
			"comparator": aws.StringValue(simpleScopeTerm.Comparator),
			"values":     flattenStringList(simpleScopeTerm.Values),
		})
	}

	return simpleScopeTermList
}

func flattenTagScopeTerm(tagScopeTerm *macie2.TagScopeTerm) []map[string]interface{} {
	tagScopeTermList := make([]map[string]interface{}, 0)

	if tagScopeTerm != nil {
		tagScopeTermList = append(tagScopeTermList, map[string]interface{}{
			"key":        aws.StringValue(tagScopeTerm.Key),
			"comparator": aws.StringValue(tagScopeTerm.Comparator),
			"target":     aws.StringValue(tagScopeTerm.Target),
			"tag_values": flattenTagValues(tagScopeTerm.TagValues),
		})
	}

	return tagScopeTermList
}

func flattenTagValues(tagValues []*macie2.TagValuePair) []map[string]interface{} {
	tagValuesList := make([]map[string]interface{}, 0)

	if tagValues != nil {
		for _, tagValue := range tagValues {
			tagValuesList = append(tagValuesList, map[string]interface{}{
				"value": aws.StringValue(tagValue.Value),
				"key":   aws.StringValue(tagValue.Key),
			})
		}
	}

	return tagValuesList
}

func flattenUserPausedDetails(userPausedDetail *macie2.UserPausedDetails) []map[string]interface{} {
	userDetails := make([]map[string]interface{}, 0)

	if userPausedDetail != nil {
		userDetails = append(userDetails, map[string]interface{}{
			"job_imminent_expiration_health_event_arn": aws.StringValue(userPausedDetail.JobImminentExpirationHealthEventArn),
			"job_expires_at": userPausedDetail.JobExpiresAt.String(),
			"job_paused_at":  userPausedDetail.JobPausedAt.String(),
		})
	}

	return userDetails
}

func flattenLastRunErrorStatus(lastErrorStatus *macie2.LastRunErrorStatus) []map[string]interface{} {
	lastError := make([]map[string]interface{}, 0)

	if lastErrorStatus != nil {
		lastError = append(lastError, map[string]interface{}{
			"code": aws.StringValue(lastErrorStatus.Code),
		})
	}

	return lastError
}

func flattenStatistics(statistics *macie2.Statistics) []map[string]interface{} {
	statisticsList := make([]map[string]interface{}, 0)

	if statistics != nil {
		statisticsList = append(statisticsList, map[string]interface{}{
			"approximate_number_of_objects_to_process": aws.Float64Value(statistics.ApproximateNumberOfObjectsToProcess),
			"number_of_runs": aws.Float64Value(statistics.NumberOfRuns),
		})
	}

	return statisticsList
}
