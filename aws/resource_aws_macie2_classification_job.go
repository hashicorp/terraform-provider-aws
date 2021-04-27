package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func resourceAwsMacie2ClassificationJob() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2ClassificationJobCreate,
		ReadWithoutTimeout:   resourceMacie2ClassificationJobRead,
		UpdateWithoutTimeout: resourceMacie2ClassificationJobUpdate,
		DeleteWithoutTimeout: resourceMacie2ClassificationJobDelete,
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
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
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
				ValidateFunc: validation.StringInSlice(macie2.JobType_Values(), false),
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
																			ValidateFunc: validation.StringInSlice(macie2.JobComparator_Values(), false),
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
																			ValidateFunc: validation.StringInSlice(macie2.ScopeFilterKey_Values(), false),
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
																			ValidateFunc: validation.StringInSlice(macie2.JobComparator_Values(), false),
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
																			ValidateFunc: validation.StringInSlice(macie2.TagTarget_Values(), false),
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
				ValidateFunc: validation.StringInSlice(macie2.JobStatus_Values(), false),
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
		ClientToken:     aws.String(resource.UniqueId()),
		Name:            aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))),
		JobType:         aws.String(d.Get("job_type").(string)),
		S3JobDefinition: expandS3JobDefinition(d.Get("s3_job_definition").([]interface{})),
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

	log.Printf("[DEBUG] Creating Macie ClassificationJob: %v", input)

	var err error
	var output macie2.CreateClassificationJobOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		resp, err := conn.CreateClassificationJobWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		output = *resp

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateClassificationJobWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie ClassificationJob: %w", err))
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

	resp, err := conn.DescribeClassificationJobWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Macie ClassificationJob does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Macie ClassificationJob (%s): %w", d.Id(), err))
	}

	if err = d.Set("custom_data_identifier_ids", flattenStringList(resp.CustomDataIdentifierIds)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie ClassificationJob (%s): %w", "custom_data_identifier_ids", d.Id(), err))
	}
	if err = d.Set("schedule_frequency", flattenScheduleFrequency(resp.ScheduleFrequency)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie ClassificationJob (%s): %w", "schedule_frequency", d.Id(), err))
	}
	d.Set("sampling_percentage", aws.Int64Value(resp.SamplingPercentage))
	d.Set("name", aws.StringValue(resp.Name))
	d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(resp.Name)))
	d.Set("description", aws.StringValue(resp.Description))
	d.Set("initial_run", aws.BoolValue(resp.InitialRun))
	d.Set("job_type", aws.StringValue(resp.JobType))
	if err = d.Set("s3_job_definition", flattenS3JobDefinition(resp.S3JobDefinition)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie ClassificationJob (%s): %w", "s3_job_definition", d.Id(), err))
	}
	if err = d.Set("tags", keyvaluetags.AppsyncKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie ClassificationJob (%s): %w", "tags", d.Id(), err))
	}
	d.Set("job_id", aws.StringValue(resp.JobId))
	d.Set("job_arn", aws.StringValue(resp.JobArn))
	d.Set("job_status", aws.StringValue(resp.JobStatus))
	d.Set("created_at", aws.TimeValue(resp.CreatedAt).Format(time.RFC3339))
	d.Set("last_run_time", aws.TimeValue(resp.LastRunTime).Format(time.RFC3339))
	if err = d.Set("user_paused_details", flattenUserPausedDetails(resp.UserPausedDetails)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie ClassificationJob (%s): %w", "user_paused_details", d.Id(), err))
	}
	if err = d.Set("last_run_error_status", flattenLastRunErrorStatus(resp.LastRunErrorStatus)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie ClassificationJob (%s): %w", "last_run_error_status", d.Id(), err))
	}
	if err = d.Set("statistics", flattenStatistics(resp.Statistics)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie ClassificationJob (%s): %w", "statistics", d.Id(), err))
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

	_, err := conn.UpdateClassificationJobWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Macie ClassificationJob (%s): %w", d.Id(), err))
	}

	return resourceMacie2ClassificationJobRead(ctx, d, meta)
}

func resourceMacie2ClassificationJobDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func expandS3JobDefinition(s3JobDefinitionObj []interface{}) *macie2.S3JobDefinition {
	if len(s3JobDefinitionObj) == 0 {
		return nil
	}

	var s3JobDefinition macie2.S3JobDefinition

	s3JobMap := s3JobDefinitionObj[0].(map[string]interface{})

	if v1, ok1 := s3JobMap["bucket_definitions"]; ok1 && len(v1.([]interface{})) > 0 {
		s3JobDefinition.SetBucketDefinitions(expandBucketDefinitions(v1.([]interface{})))
	}
	if v1, ok1 := s3JobMap["scoping"]; ok1 && len(v1.([]interface{})) > 0 {
		s3JobDefinition.SetScoping(expandScoping(v1.([]interface{})))
	}

	return &s3JobDefinition
}

func expandBucketDefinitions(definitions []interface{}) []*macie2.S3BucketDefinitionForJob {
	if len(definitions) == 0 {
		return nil
	}

	var bucketDefinitions []*macie2.S3BucketDefinitionForJob

	for _, v := range definitions {
		v1 := v.(map[string]interface{})

		var bucketDefinition macie2.S3BucketDefinitionForJob

		bucketDefinition.Buckets = expandStringList(v1["buckets"].([]interface{}))
		bucketDefinition.AccountId = aws.String(v1["account_id"].(string))

		bucketDefinitions = append(bucketDefinitions, &bucketDefinition)
	}

	return bucketDefinitions
}

func expandScoping(scoping []interface{}) *macie2.Scoping {
	if len(scoping) == 0 {
		return nil
	}

	var scopingObj macie2.Scoping

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

	return &scopingObj
}

func expandJobScopeTerm(scopeTerms []interface{}) []*macie2.JobScopeTerm {
	if len(scopeTerms) == 0 {
		return nil
	}

	var scopeTermsList []*macie2.JobScopeTerm

	for _, v := range scopeTerms {
		v1 := v.(map[string]interface{})
		var scopeTerm macie2.JobScopeTerm

		if v2, ok1 := v1["simple_scope_term"]; ok1 && len(v2.([]interface{})) > 0 {
			scopeTerm.SetSimpleScopeTerm(expandSimpleScopeTerm(v2.([]interface{})))
		}
		if v2, ok1 := v1["tag_scope_term"]; ok1 && len(v2.([]interface{})) > 0 {
			scopeTerm.SetTagScopeTerm(expandTagScopeTerm(v2.([]interface{})))
		}
		scopeTermsList = append(scopeTermsList, &scopeTerm)
	}

	return scopeTermsList
}

func expandSimpleScopeTerm(simpleScopeTerm []interface{}) *macie2.SimpleScopeTerm {
	if len(simpleScopeTerm) == 0 {
		return nil
	}

	var simpleTerm macie2.SimpleScopeTerm

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

	return &simpleTerm
}

func expandTagScopeTerm(tagScopeTerm []interface{}) *macie2.TagScopeTerm {
	if len(tagScopeTerm) == 0 {
		return nil
	}

	var tagTerm macie2.TagScopeTerm

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

	return &tagTerm
}

func expandTagValues(tagValues []interface{}) []*macie2.TagValuePair {
	if len(tagValues) == 0 {
		return nil
	}

	var tagValuesList []*macie2.TagValuePair

	for _, v := range tagValues {
		v1 := v.(map[string]interface{})
		var tagValue macie2.TagValuePair

		if v2, ok := v1["value"]; ok && v2.(string) != "" {
			tagValue.SetValue(v2.(string))
		}
		if v2, ok := v1["key"]; ok && v2.(string) != "" {
			tagValue.SetKey(v2.(string))
		}
		tagValuesList = append(tagValuesList, &tagValue)
	}

	return tagValuesList
}

func expandScheduleFrequency(schedules []interface{}) *macie2.JobScheduleFrequency {
	if len(schedules) == 0 {
		return nil
	}

	var jobScheduleFrequency macie2.JobScheduleFrequency

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

	return &jobScheduleFrequency
}

func flattenScheduleFrequency(schedule *macie2.JobScheduleFrequency) []map[string]interface{} {
	if schedule == nil {
		return nil
	}

	var schedulesList []map[string]interface{}
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

	return schedulesList
}

func flattenS3JobDefinition(s3JobDefinition *macie2.S3JobDefinition) []map[string]interface{} {
	if s3JobDefinition == nil {
		return nil
	}

	var jobDefinitions []map[string]interface{}

	jobDefinitions = append(jobDefinitions, map[string]interface{}{
		"bucket_definitions": flattenBucketDefinition(s3JobDefinition.BucketDefinitions),
		"scoping":            flattenScoping(s3JobDefinition.Scoping),
	})

	return jobDefinitions
}

func flattenBucketDefinition(bucketDefinitions []*macie2.S3BucketDefinitionForJob) []map[string]interface{} {
	if len(bucketDefinitions) == 0 {
		return nil
	}

	var bucketDefinitionList []map[string]interface{}

	for _, bucket := range bucketDefinitions {
		if bucket == nil {
			continue
		}
		bucketDefinitionList = append(bucketDefinitionList, map[string]interface{}{
			"account_id": aws.StringValue(bucket.AccountId),
			"buckets":    flattenStringList(bucket.Buckets),
		})
	}

	return bucketDefinitionList
}

func flattenScoping(scoping *macie2.Scoping) []map[string]interface{} {
	if scoping == nil {
		return nil
	}

	var scopingList []map[string]interface{}

	scopingList = append(scopingList, map[string]interface{}{
		"excludes": flattenJobScopingBlock(scoping.Excludes),
		"includes": flattenJobScopingBlock(scoping.Includes),
	})

	return scopingList
}

func flattenJobScopingBlock(scopeTerm *macie2.JobScopingBlock) []map[string]interface{} {
	if scopeTerm == nil {
		return nil
	}

	var scopeTermList []map[string]interface{}

	scopeTermList = append(scopeTermList, map[string]interface{}{
		"and": flattenJobScopeTerm(scopeTerm.And),
	})

	return scopeTermList
}

func flattenJobScopeTerm(scopeTerms []*macie2.JobScopeTerm) []map[string]interface{} {
	if scopeTerms == nil {
		return nil
	}

	var scopeTermList []map[string]interface{}

	for _, scopeTerm := range scopeTerms {
		scopeTermList = append(scopeTermList, map[string]interface{}{
			"simple_scope_term": flattenSimpleScopeTerm(scopeTerm.SimpleScopeTerm),
			"tag_scope_term":    flattenTagScopeTerm(scopeTerm.TagScopeTerm),
		})
	}

	return scopeTermList
}

func flattenSimpleScopeTerm(simpleScopeTerm *macie2.SimpleScopeTerm) []map[string]interface{} {
	if simpleScopeTerm == nil {
		return nil
	}

	var simpleScopeTermList []map[string]interface{}

	simpleScopeTermList = append(simpleScopeTermList, map[string]interface{}{
		"key":        aws.StringValue(simpleScopeTerm.Key),
		"comparator": aws.StringValue(simpleScopeTerm.Comparator),
		"values":     flattenStringList(simpleScopeTerm.Values),
	})

	return simpleScopeTermList
}

func flattenTagScopeTerm(tagScopeTerm *macie2.TagScopeTerm) []map[string]interface{} {
	if tagScopeTerm == nil {
		return nil
	}

	var tagScopeTermList []map[string]interface{}

	tagScopeTermList = append(tagScopeTermList, map[string]interface{}{
		"key":        aws.StringValue(tagScopeTerm.Key),
		"comparator": aws.StringValue(tagScopeTerm.Comparator),
		"target":     aws.StringValue(tagScopeTerm.Target),
		"tag_values": flattenTagValues(tagScopeTerm.TagValues),
	})

	return tagScopeTermList
}

func flattenTagValues(tagValues []*macie2.TagValuePair) []map[string]interface{} {
	if len(tagValues) == 0 {
		return nil
	}

	var tagValuesList []map[string]interface{}

	for _, tagValue := range tagValues {
		tagValuesList = append(tagValuesList, map[string]interface{}{
			"value": aws.StringValue(tagValue.Value),
			"key":   aws.StringValue(tagValue.Key),
		})
	}

	return tagValuesList
}

func flattenUserPausedDetails(userPausedDetail *macie2.UserPausedDetails) []map[string]interface{} {
	if userPausedDetail == nil {
		return nil
	}

	var userDetails []map[string]interface{}

	userDetails = append(userDetails, map[string]interface{}{
		"job_imminent_expiration_health_event_arn": aws.StringValue(userPausedDetail.JobImminentExpirationHealthEventArn),
		"job_expires_at": userPausedDetail.JobExpiresAt.String(),
		"job_paused_at":  userPausedDetail.JobPausedAt.String(),
	})

	return userDetails
}

func flattenLastRunErrorStatus(lastErrorStatus *macie2.LastRunErrorStatus) []map[string]interface{} {
	if lastErrorStatus == nil {
		return nil
	}

	var lastError []map[string]interface{}

	lastError = append(lastError, map[string]interface{}{
		"code": aws.StringValue(lastErrorStatus.Code),
	})

	return lastError
}

func flattenStatistics(statistics *macie2.Statistics) []map[string]interface{} {
	if statistics == nil {
		return nil
	}

	var statisticsList []map[string]interface{}

	statisticsList = append(statisticsList, map[string]interface{}{
		"approximate_number_of_objects_to_process": aws.Float64Value(statistics.ApproximateNumberOfObjectsToProcess),
		"number_of_runs": aws.Float64Value(statistics.NumberOfRuns),
	})

	return statisticsList
}
