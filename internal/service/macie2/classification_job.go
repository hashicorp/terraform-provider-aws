// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_classification_job", name="Classification Job")
// @Tags
func ResourceClassificationJob() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClassificationJobCreate,
		ReadWithoutTimeout:   resourceClassificationJobRead,
		UpdateWithoutTimeout: resourceClassificationJobUpdate,
		DeleteWithoutTimeout: resourceClassificationJobDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"custom_data_identifier_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"schedule_frequency": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"daily_schedule": {
							Type:          schema.TypeBool,
							Optional:      true,
							ConflictsWith: []string{"schedule_frequency.0.weekly_schedule", "schedule_frequency.0.monthly_schedule"},
						},
						"weekly_schedule": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"schedule_frequency.0.daily_schedule", "schedule_frequency.0.monthly_schedule"},
						},
						"monthly_schedule": {
							Type:          schema.TypeInt,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"schedule_frequency.0.daily_schedule", "schedule_frequency.0.weekly_schedule"},
						},
					},
				},
			},
			"sampling_percentage": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validation.StringLenBetween(0, 500),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validation.StringLenBetween(0, 500-id.UniqueIDSuffixLength),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"initial_run": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"job_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(macie2.JobType_Values(), false),
			},
			"s3_job_definition": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_definitions": {
							ConflictsWith: []string{"s3_job_definition.0.bucket_criteria"},
							Type:          schema.TypeList,
							Optional:      true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAccountID: {
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
						"bucket_criteria": {
							ConflictsWith: []string{"s3_job_definition.0.bucket_definitions"},
							Type:          schema.TypeList,
							Optional:      true,
							Computed:      true,
							MaxItems:      1,
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
															"simple_criterion": {
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
																		names.AttrValues: {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		names.AttrKey: {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																	},
																},
															},
															"tag_criterion": {
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
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					names.AttrValue: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																					names.AttrKey: {
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
															"simple_criterion": {
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
																		names.AttrValues: {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		names.AttrKey: {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																		},
																	},
																},
															},
															"tag_criterion": {
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
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					names.AttrValue: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																					names.AttrKey: {
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
																		names.AttrValues: {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		names.AttrKey: {
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
																					names.AttrValue: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																					names.AttrKey: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																		},
																		names.AttrKey: {
																			Type:         schema.TypeString,
																			Optional:     true,
																			Computed:     true,
																			ValidateFunc: validation.StringInSlice(tagScopeTermKey_Values(), false),
																		},
																		names.AttrTarget: {
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
																		names.AttrValues: {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		names.AttrKey: {
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
																					names.AttrValue: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																					names.AttrKey: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																					},
																				},
																			},
																		},
																		names.AttrKey: {
																			Type:         schema.TypeString,
																			Optional:     true,
																			Computed:     true,
																			ValidateFunc: validation.StringInSlice(tagScopeTermKey_Values(), false),
																		},
																		names.AttrTarget: {
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
								},
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
				ValidateFunc: validation.StringInSlice([]string{macie2.JobStatusCancelled, macie2.JobStatusRunning, macie2.JobStatusUserPaused}, false),
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_paused_details": {
				Type:     schema.TypeList,
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
		},
		CustomizeDiff: resourceClassificationJobCustomizeDiff,
	}
}
func resourceClassificationJobCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	//TagScopeTerm() enforces the `target` key even though documentation marks it as optional.
	//ClassificationJobs criteria and scoping cannot be updated.
	//The API as of Aug 7, 2022 returns an empty string (even if a target was sent), causing a diff on new plans.
	//The following will clear the diff for these keys if the object exists already in the state.
	if diff.Id() != "" {
		for _, key := range diff.GetChangedKeysPrefix("s3_job_definition.0.scoping.0.excludes") {
			if strings.Contains(key, "tag_scope_term") && strings.Contains(key, names.AttrTarget) {
				err := diff.Clear(key)
				if err != nil {
					return err
				}
			}
		}
		for _, key := range diff.GetChangedKeysPrefix("s3_job_definition.0.scoping.0.includes") {
			if strings.Contains(key, "tag_scope_term") && strings.Contains(key, names.AttrTarget) {
				err := diff.Clear(key)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func resourceClassificationJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.CreateClassificationJobInput{
		ClientToken:     aws.String(id.UniqueId()),
		Name:            aws.String(create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))),
		JobType:         aws.String(d.Get("job_type").(string)),
		S3JobDefinition: expandS3JobDefinition(d.Get("s3_job_definition").([]interface{})),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("custom_data_identifier_ids"); ok {
		input.CustomDataIdentifierIds = flex.ExpandStringList(v.([]interface{}))
	}
	if v, ok := d.GetOk("schedule_frequency"); ok {
		input.ScheduleFrequency = expandScheduleFrequency(v.([]interface{}))
	}
	if v, ok := d.GetOk("sampling_percentage"); ok {
		input.SamplingPercentage = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("initial_run"); ok {
		input.InitialRun = aws.Bool(v.(bool))
	}

	var err error
	var output *macie2.CreateClassificationJobOutput
	err = retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		output, err = conn.CreateClassificationJobWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateClassificationJobWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie ClassificationJob: %s", err)
	}

	d.SetId(aws.StringValue(output.JobId))

	return append(diags, resourceClassificationJobRead(ctx, d, meta)...)
}

func resourceClassificationJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.DescribeClassificationJobInput{
		JobId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeClassificationJobWithContext(ctx, input)

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") ||
		tfawserr.ErrMessageContains(err, macie2.ErrCodeValidationException, "cannot update cancelled job for job")) {
		log.Printf("[WARN] Macie ClassificationJob (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie ClassificationJob (%s): %s", d.Id(), err)
	}

	if err = d.Set("custom_data_identifier_ids", flex.FlattenStringList(resp.CustomDataIdentifierIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for Macie ClassificationJob (%s): %s", "custom_data_identifier_ids", d.Id(), err)
	}
	if err = d.Set("schedule_frequency", flattenScheduleFrequency(resp.ScheduleFrequency)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for Macie ClassificationJob (%s): %s", "schedule_frequency", d.Id(), err)
	}
	d.Set("sampling_percentage", resp.SamplingPercentage)
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.StringValue(resp.Name)))
	d.Set(names.AttrDescription, resp.Description)
	d.Set("initial_run", resp.InitialRun)
	d.Set("job_type", resp.JobType)
	if err = d.Set("s3_job_definition", flattenS3JobDefinition(resp.S3JobDefinition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for Macie ClassificationJob (%s): %s", "s3_job_definition", d.Id(), err)
	}

	setTagsOut(ctx, resp.Tags)

	d.Set("job_id", resp.JobId)
	d.Set("job_arn", resp.JobArn)
	status := aws.StringValue(resp.JobStatus)
	if status == macie2.JobStatusComplete || status == macie2.JobStatusIdle || status == macie2.JobStatusPaused {
		status = macie2.JobStatusRunning
	}
	d.Set("job_status", status)
	d.Set(names.AttrCreatedAt, aws.TimeValue(resp.CreatedAt).Format(time.RFC3339))
	if err = d.Set("user_paused_details", flattenUserPausedDetails(resp.UserPausedDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for Macie ClassificationJob (%s): %s", "user_paused_details", d.Id(), err)
	}

	return diags
}

func resourceClassificationJobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.UpdateClassificationJobInput{
		JobId: aws.String(d.Id()),
	}

	if d.HasChange("job_status") {
		status := d.Get("job_status").(string)

		if status == macie2.JobStatusCancelled {
			return sdkdiag.AppendErrorf(diags, "updating Macie ClassificationJob (%s): %s", d.Id(), fmt.Sprintf("%s cannot be set", macie2.JobStatusCancelled))
		}

		input.JobStatus = aws.String(status)
	}

	_, err := conn.UpdateClassificationJobWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Macie ClassificationJob (%s): %s", d.Id(), err)
	}

	return append(diags, resourceClassificationJobRead(ctx, d, meta)...)
}

func resourceClassificationJobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.UpdateClassificationJobInput{
		JobId:     aws.String(d.Id()),
		JobStatus: aws.String(macie2.JobStatusCancelled),
	}

	_, err := conn.UpdateClassificationJobWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeValidationException, "cannot update cancelled job for job") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Macie ClassificationJob (%s): %s", d.Id(), err)
	}

	return diags
}

func expandS3JobDefinition(s3JobDefinitionObj []interface{}) *macie2.S3JobDefinition {
	if len(s3JobDefinitionObj) == 0 {
		return nil
	}

	var s3JobDefinition macie2.S3JobDefinition

	s3JobMap := s3JobDefinitionObj[0].(map[string]interface{})

	if v1, ok1 := s3JobMap["bucket_criteria"]; ok1 && len(v1.([]interface{})) > 0 {
		s3JobDefinition.BucketCriteria = expandS3BucketCriteriaForJob(v1.([]interface{}))
	}
	if v1, ok1 := s3JobMap["bucket_definitions"]; ok1 && len(v1.([]interface{})) > 0 {
		s3JobDefinition.BucketDefinitions = expandBucketDefinitions(v1.([]interface{}))
	}
	if v1, ok1 := s3JobMap["scoping"]; ok1 && len(v1.([]interface{})) > 0 {
		s3JobDefinition.Scoping = expandScoping(v1.([]interface{}))
	}

	return &s3JobDefinition
}

func expandS3BucketCriteriaForJob(criteria []interface{}) *macie2.S3BucketCriteriaForJob {
	if len(criteria) == 0 {
		return nil
	}

	var criteriaObj macie2.S3BucketCriteriaForJob

	criteriaMap := criteria[0].(map[string]interface{})

	if v, ok := criteriaMap["excludes"]; ok && len(v.([]interface{})) > 0 {
		v1 := v.([]interface{})
		andMap := v1[0].(map[string]interface{})
		if v2, ok1 := andMap["and"]; ok1 && len(v2.([]interface{})) > 0 {
			criteriaObj.Excludes = &macie2.CriteriaBlockForJob{
				And: expandCriteriaBlockForJob(v2.([]interface{})),
			}
		}
	}
	if v, ok := criteriaMap["includes"]; ok && len(v.([]interface{})) > 0 {
		v1 := v.([]interface{})
		andMap := v1[0].(map[string]interface{})
		if v2, ok1 := andMap["and"]; ok1 && len(v2.([]interface{})) > 0 {
			criteriaObj.Includes = &macie2.CriteriaBlockForJob{
				And: expandCriteriaBlockForJob(v2.([]interface{})),
			}
		}
	}

	return &criteriaObj
}

func expandCriteriaBlockForJob(criteriaBlocks []interface{}) []*macie2.CriteriaForJob {
	if len(criteriaBlocks) == 0 {
		return nil
	}

	var criteriaBlocksList []*macie2.CriteriaForJob

	for _, v := range criteriaBlocks {
		v1 := v.(map[string]interface{})
		var criteriaBlock macie2.CriteriaForJob

		if v2, ok1 := v1["simple_criterion"]; ok1 && len(v2.([]interface{})) > 0 {
			criteriaBlock.SimpleCriterion = expandSimpleCriterionForJob(v2.([]interface{}))
		}
		if v2, ok1 := v1["tag_criterion"]; ok1 && len(v2.([]interface{})) > 0 {
			criteriaBlock.TagCriterion = expandTagCriterionForJob(v2.([]interface{}))
		}

		criteriaBlocksList = append(criteriaBlocksList, &criteriaBlock)
	}

	return criteriaBlocksList
}

func expandSimpleCriterionForJob(criterion []interface{}) *macie2.SimpleCriterionForJob {
	if len(criterion) == 0 {
		return nil
	}

	var simpleCriterion macie2.SimpleCriterionForJob

	simpleCriterionMap := criterion[0].(map[string]interface{})

	if v, ok := simpleCriterionMap["comparator"]; ok && v.(string) != "" {
		simpleCriterion.Comparator = aws.String(v.(string))
	}
	if v, ok := simpleCriterionMap[names.AttrKey]; ok && v.(string) != "" {
		simpleCriterion.Key = aws.String(v.(string))
	}
	if v, ok := simpleCriterionMap[names.AttrValues]; ok && len(v.([]interface{})) > 0 {
		simpleCriterion.Values = flex.ExpandStringList(v.([]interface{}))
	}

	return &simpleCriterion
}

func expandTagCriterionForJob(criterion []interface{}) *macie2.TagCriterionForJob {
	if len(criterion) == 0 {
		return nil
	}

	var tagCriterion macie2.TagCriterionForJob

	tagCriterionMap := criterion[0].(map[string]interface{})

	if v, ok := tagCriterionMap["comparator"]; ok && v.(string) != "" {
		tagCriterion.Comparator = aws.String(v.(string))
	}
	if v, ok := tagCriterionMap["tag_values"]; ok && len(v.([]interface{})) > 0 {
		tagCriterion.TagValues = expandTagCriterionPairForJob(v.([]interface{}))
	}

	return &tagCriterion
}

func expandTagCriterionPairForJob(tagValues []interface{}) []*macie2.TagCriterionPairForJob {
	if len(tagValues) == 0 {
		return nil
	}

	var tagValuesList []*macie2.TagCriterionPairForJob

	for _, v := range tagValues {
		v1 := v.(map[string]interface{})
		var tagValue macie2.TagCriterionPairForJob

		if v2, ok := v1[names.AttrValue]; ok && v2.(string) != "" {
			tagValue.Value = aws.String(v2.(string))
		}
		if v2, ok := v1[names.AttrKey]; ok && v2.(string) != "" {
			tagValue.Key = aws.String(v2.(string))
		}
		tagValuesList = append(tagValuesList, &tagValue)
	}

	return tagValuesList
}

func expandBucketDefinitions(definitions []interface{}) []*macie2.S3BucketDefinitionForJob {
	if len(definitions) == 0 {
		return nil
	}

	var bucketDefinitions []*macie2.S3BucketDefinitionForJob

	for _, v := range definitions {
		v1 := v.(map[string]interface{})

		bucketDefinition := &macie2.S3BucketDefinitionForJob{
			Buckets:   flex.ExpandStringList(v1["buckets"].([]interface{})),
			AccountId: aws.String(v1[names.AttrAccountID].(string)),
		}

		bucketDefinitions = append(bucketDefinitions, bucketDefinition)
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
			scopingObj.Excludes = &macie2.JobScopingBlock{
				And: expandJobScopeTerm(v2.([]interface{})),
			}
		}
	}
	if v, ok := scopingMap["includes"]; ok && len(v.([]interface{})) > 0 {
		v1 := v.([]interface{})
		andMap := v1[0].(map[string]interface{})
		if v2, ok1 := andMap["and"]; ok1 && len(v2.([]interface{})) > 0 {
			scopingObj.Includes = &macie2.JobScopingBlock{
				And: expandJobScopeTerm(v2.([]interface{})),
			}
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
			scopeTerm.SimpleScopeTerm = expandSimpleScopeTerm(v2.([]interface{}))
		}
		if v2, ok1 := v1["tag_scope_term"]; ok1 && len(v2.([]interface{})) > 0 {
			scopeTerm.TagScopeTerm = expandTagScopeTerm(v2.([]interface{}))
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

	if v, ok := simpleScopeTermMap[names.AttrKey]; ok && v.(string) != "" {
		simpleTerm.Key = aws.String(v.(string))
	}
	if v, ok := simpleScopeTermMap[names.AttrValues]; ok && len(v.([]interface{})) > 0 {
		simpleTerm.Values = flex.ExpandStringList(v.([]interface{}))
	}
	if v, ok := simpleScopeTermMap["comparator"]; ok && v.(string) != "" {
		simpleTerm.Comparator = aws.String(v.(string))
	}

	return &simpleTerm
}

func expandTagScopeTerm(tagScopeTerm []interface{}) *macie2.TagScopeTerm {
	if len(tagScopeTerm) == 0 {
		return nil
	}

	var tagTerm macie2.TagScopeTerm

	tagScopeTermMap := tagScopeTerm[0].(map[string]interface{})

	if v, ok := tagScopeTermMap[names.AttrKey]; ok && v.(string) != "" {
		tagTerm.Key = aws.String(v.(string))
	}
	if v, ok := tagScopeTermMap["tag_values"]; ok && len(v.([]interface{})) > 0 {
		tagTerm.TagValues = expandTagValues(v.([]interface{}))
	}
	if v, ok := tagScopeTermMap["comparator"]; ok && v.(string) != "" {
		tagTerm.Comparator = aws.String(v.(string))
	}
	if v, ok := tagScopeTermMap[names.AttrTarget]; ok && v.(string) != "" {
		tagTerm.Target = aws.String(v.(string))
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

		if v2, ok := v1[names.AttrValue]; ok && v2.(string) != "" {
			tagValue.Value = aws.String(v2.(string))
		}
		if v2, ok := v1[names.AttrKey]; ok && v2.(string) != "" {
			tagValue.Key = aws.String(v2.(string))
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
		jobScheduleFrequency.DailySchedule = &macie2.DailySchedule{}
	}
	if v1, ok1 := scheduleMap["weekly_schedule"]; ok1 && v1.(string) != "" {
		jobScheduleFrequency.WeeklySchedule = &macie2.WeeklySchedule{
			DayOfWeek: aws.String(v1.(string)),
		}
	}
	if v1, ok1 := scheduleMap["monthly_schedule"]; ok1 && v1.(int) > 0 {
		jobScheduleFrequency.MonthlySchedule = &macie2.MonthlySchedule{
			DayOfMonth: aws.Int64(int64(v1.(int))),
		}
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
		"bucket_criteria":    flattenS3BucketCriteriaForJob(s3JobDefinition.BucketCriteria),
		"bucket_definitions": flattenBucketDefinition(s3JobDefinition.BucketDefinitions),
		"scoping":            flattenScoping(s3JobDefinition.Scoping),
	})

	return jobDefinitions
}

func flattenS3BucketCriteriaForJob(criteria *macie2.S3BucketCriteriaForJob) []map[string]interface{} {
	if criteria == nil {
		return nil
	}

	var criteriaList []map[string]interface{}

	criteriaList = append(criteriaList, map[string]interface{}{
		"excludes": flattenCriteriaBlockForJob(criteria.Excludes),
		"includes": flattenCriteriaBlockForJob(criteria.Includes),
	})

	return criteriaList
}

func flattenCriteriaBlockForJob(criteriaBlock *macie2.CriteriaBlockForJob) []map[string]interface{} {
	if criteriaBlock == nil {
		return nil
	}

	var criteriaBlockList []map[string]interface{}

	criteriaBlockList = append(criteriaBlockList, map[string]interface{}{
		"and": flattenCriteriaForJob(criteriaBlock.And),
	})

	return criteriaBlockList
}

func flattenCriteriaForJob(criteria []*macie2.CriteriaForJob) []map[string]interface{} {
	if criteria == nil {
		return nil
	}

	var criteriaList []map[string]interface{}

	for _, criterion := range criteria {
		criteriaList = append(criteriaList, map[string]interface{}{
			"simple_criterion": flattenSimpleCriterionForJob(criterion.SimpleCriterion),
			"tag_criterion":    flattenTagCriterionForJob(criterion.TagCriterion),
		})
	}

	return criteriaList
}

func flattenSimpleCriterionForJob(criterion *macie2.SimpleCriterionForJob) []map[string]interface{} {
	if criterion == nil {
		return nil
	}

	var simpleCriterionList []map[string]interface{}

	simpleCriterionList = append(simpleCriterionList, map[string]interface{}{
		"comparator":     aws.StringValue(criterion.Comparator),
		names.AttrKey:    aws.StringValue(criterion.Key),
		names.AttrValues: flex.FlattenStringList(criterion.Values),
	})

	return simpleCriterionList
}

func flattenTagCriterionForJob(criterion *macie2.TagCriterionForJob) []map[string]interface{} {
	if criterion == nil {
		return nil
	}

	var tagCriterionList []map[string]interface{}

	tagCriterionList = append(tagCriterionList, map[string]interface{}{
		"comparator": aws.StringValue(criterion.Comparator),
		"tag_values": flattenTagCriterionPairForJob(criterion.TagValues),
	})

	return tagCriterionList
}

func flattenTagCriterionPairForJob(tagValues []*macie2.TagCriterionPairForJob) []map[string]interface{} {
	if len(tagValues) == 0 {
		return nil
	}

	var tagValuesList []map[string]interface{}

	for _, tagValue := range tagValues {
		tagValuesList = append(tagValuesList, map[string]interface{}{
			names.AttrValue: aws.StringValue(tagValue.Value),
			names.AttrKey:   aws.StringValue(tagValue.Key),
		})
	}

	return tagValuesList
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
			names.AttrAccountID: aws.StringValue(bucket.AccountId),
			"buckets":           flex.FlattenStringList(bucket.Buckets),
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
		names.AttrKey:    aws.StringValue(simpleScopeTerm.Key),
		"comparator":     aws.StringValue(simpleScopeTerm.Comparator),
		names.AttrValues: flex.FlattenStringList(simpleScopeTerm.Values),
	})

	return simpleScopeTermList
}

func flattenTagScopeTerm(tagScopeTerm *macie2.TagScopeTerm) []map[string]interface{} {
	if tagScopeTerm == nil {
		return nil
	}

	var tagScopeTermList []map[string]interface{}

	tagScopeTermList = append(tagScopeTermList, map[string]interface{}{
		names.AttrKey:    aws.StringValue(tagScopeTerm.Key),
		"comparator":     aws.StringValue(tagScopeTerm.Comparator),
		names.AttrTarget: aws.StringValue(tagScopeTerm.Target),
		"tag_values":     flattenTagValues(tagScopeTerm.TagValues),
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
			names.AttrValue: aws.StringValue(tagValue.Value),
			names.AttrKey:   aws.StringValue(tagValue.Key),
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
