// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_classification_job", name="Classification Job")
// @Tags(identifierAttribute="job_arn")
func resourceClassificationJob() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClassificationJobCreate,
		ReadWithoutTimeout:   resourceClassificationJobRead,
		UpdateWithoutTimeout: resourceClassificationJobUpdate,
		DeleteWithoutTimeout: resourceClassificationJobDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_data_identifier_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"job_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(enum.Slice(awstypes.JobStatusCancelled, awstypes.JobStatusRunning, awstypes.JobStatusUserPaused), false),
			},
			"job_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.JobType](),
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
							ForceNew:      true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAccountID: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"buckets": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
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
							ForceNew:      true,
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"excludes": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"and": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"simple_criterion": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																ForceNew: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ForceNew:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.JobComparator](),
																		},
																		names.AttrValues: {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		names.AttrKey: {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																		},
																	},
																},
															},
															"tag_criterion": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																ForceNew: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ForceNew:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.JobComparator](),
																		},
																		"tag_values": {
																			Type:     schema.TypeList,
																			Optional: true,
																			ForceNew: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					names.AttrValue: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																						ForceNew: true,
																					},
																					names.AttrKey: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																						ForceNew: true,
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
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"and": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"simple_criterion": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																ForceNew: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ForceNew:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.JobComparator](),
																		},
																		names.AttrValues: {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		names.AttrKey: {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																		},
																	},
																},
															},
															"tag_criterion": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																ForceNew: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ForceNew:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.JobComparator](),
																		},
																		"tag_values": {
																			Type:     schema.TypeList,
																			Optional: true,
																			ForceNew: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					names.AttrValue: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																						ForceNew: true,
																					},
																					names.AttrKey: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																						ForceNew: true,
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
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"excludes": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"and": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"simple_scope_term": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																ForceNew: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ForceNew:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.JobComparator](),
																		},
																		names.AttrValues: {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		names.AttrKey: {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ForceNew:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.ScopeFilterKey](),
																		},
																	},
																},
															},
															"tag_scope_term": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																ForceNew: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ForceNew:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.JobComparator](),
																		},
																		"tag_values": {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					names.AttrValue: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																						ForceNew: true,
																					},
																					names.AttrKey: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																						ForceNew: true,
																					},
																				},
																			},
																		},
																		names.AttrKey: {
																			Type:         schema.TypeString,
																			Optional:     true,
																			Computed:     true,
																			ForceNew:     true,
																			ValidateFunc: validation.StringInSlice(tagScopeTermKey_Values(), false),
																		},
																		names.AttrTarget: {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ForceNew:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.TagTarget](),
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
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"and": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"simple_scope_term": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																ForceNew: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																		},
																		names.AttrValues: {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		names.AttrKey: {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																		},
																	},
																},
															},
															"tag_scope_term": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																ForceNew: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"comparator": {
																			Type:     schema.TypeString,
																			Optional: true,
																			Computed: true,
																			ForceNew: true,
																		},
																		"tag_values": {
																			Type:     schema.TypeList,
																			Optional: true,
																			ForceNew: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					names.AttrValue: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																						ForceNew: true,
																					},
																					names.AttrKey: {
																						Type:     schema.TypeString,
																						Optional: true,
																						Computed: true,
																						ForceNew: true,
																					},
																				},
																			},
																		},
																		names.AttrKey: {
																			Type:         schema.TypeString,
																			Optional:     true,
																			Computed:     true,
																			ForceNew:     true,
																			ValidateFunc: validation.StringInSlice(tagScopeTermKey_Values(), false),
																		},
																		names.AttrTarget: {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ForceNew:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.TagTarget](),
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
			"sampling_percentage": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"schedule_frequency": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"daily_schedule": {
							Type:          schema.TypeBool,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"schedule_frequency.0.weekly_schedule", "schedule_frequency.0.monthly_schedule"},
						},
						"weekly_schedule": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"schedule_frequency.0.daily_schedule", "schedule_frequency.0.monthly_schedule"},
						},
						"monthly_schedule": {
							Type:          schema.TypeInt,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"schedule_frequency.0.daily_schedule", "schedule_frequency.0.weekly_schedule"},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_paused_details": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"job_expires_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"job_imminent_expiration_health_event_arn": {
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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Minute),
			Update: schema.DefaultTimeout(4 * time.Minute),
		},
	}
}

func resourceClassificationJobCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta any) error {
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

func resourceClassificationJobCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := macie2.CreateClassificationJobInput{
		ClientToken:     aws.String(id.UniqueId()),
		JobType:         awstypes.JobType(d.Get("job_type").(string)),
		Name:            aws.String(name),
		S3JobDefinition: expandS3JobDefinition(d.Get("s3_job_definition").([]any)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("custom_data_identifier_ids"); ok {
		input.CustomDataIdentifierIds = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("initial_run"); ok {
		input.InitialRun = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("sampling_percentage"); ok {
		input.SamplingPercentage = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("schedule_frequency"); ok {
		input.ScheduleFrequency = expandScheduleFrequency(v.([]any))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.CreateClassificationJob(ctx, &input)
	}, errCodeClientError)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie Classification Job (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*macie2.CreateClassificationJobOutput).JobId))

	return append(diags, resourceClassificationJobRead(ctx, d, meta)...)
}

func resourceClassificationJobRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	output, err := findClassificationJobByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Macie Classification Job (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie Classification Job (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCreatedAt, aws.ToTime(output.CreatedAt).Format(time.RFC3339))
	if err = d.Set("custom_data_identifier_ids", output.CustomDataIdentifierIds); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting custom_data_identifier_ids: %s", err)
	}
	d.Set(names.AttrDescription, output.Description)
	d.Set("initial_run", output.InitialRun)
	d.Set("job_arn", output.JobArn)
	d.Set("job_id", output.JobId)
	jobStatus := output.JobStatus
	if jobStatus == awstypes.JobStatusComplete || jobStatus == awstypes.JobStatusIdle || jobStatus == awstypes.JobStatusPaused {
		jobStatus = awstypes.JobStatusRunning
	}
	d.Set("job_status", jobStatus)
	d.Set("job_type", output.JobType)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.Name)))
	if err = d.Set("s3_job_definition", flattenS3JobDefinition(output.S3JobDefinition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3_job_definition: %s", err)
	}
	d.Set("sampling_percentage", output.SamplingPercentage)
	if err = d.Set("schedule_frequency", flattenScheduleFrequency(output.ScheduleFrequency)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting schedule_frequency: %s", err)
	}
	if err = d.Set("user_paused_details", flattenUserPausedDetails(output.UserPausedDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting user_paused_details: %s", err)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceClassificationJobUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	if d.HasChange("job_status") {
		jobStatus := awstypes.JobStatus(d.Get("job_status").(string))

		if jobStatus == awstypes.JobStatusCancelled {
			return sdkdiag.AppendErrorf(diags, "updating Macie Classification Job (%s): %s", d.Id(), fmt.Errorf("%s status cannot be set", awstypes.JobStatusCancelled))
		}

		input := macie2.UpdateClassificationJobInput{
			JobId:     aws.String(d.Id()),
			JobStatus: jobStatus,
		}

		_, err := conn.UpdateClassificationJob(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Macie Classification Job (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceClassificationJobRead(ctx, d, meta)...)
}

func resourceClassificationJobDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := macie2.UpdateClassificationJobInput{
		JobId:     aws.String(d.Id()),
		JobStatus: awstypes.JobStatusCancelled,
	}

	_, err := conn.UpdateClassificationJob(ctx, &input)

	if isClassificationJobNotFoundError(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Macie Classification Job (%s): %s", d.Id(), err)
	}

	return diags
}

func findClassificationJobByID(ctx context.Context, conn *macie2.Client, id string) (*macie2.DescribeClassificationJobOutput, error) {
	input := macie2.DescribeClassificationJobInput{
		JobId: aws.String(id),
	}

	return findClassificationJob(ctx, conn, &input)
}

func findClassificationJob(ctx context.Context, conn *macie2.Client, input *macie2.DescribeClassificationJobInput) (*macie2.DescribeClassificationJobOutput, error) {
	output, err := conn.DescribeClassificationJob(ctx, input)

	if isClassificationJobNotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func isClassificationJobNotFoundError(err error) bool {
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return true
	}
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
		return true
	}
	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "cannot update cancelled job for job") {
		return true
	}

	return false
}

func expandS3JobDefinition(s3JobDefinitionObj []any) *awstypes.S3JobDefinition {
	if len(s3JobDefinitionObj) == 0 {
		return nil
	}

	var s3JobDefinition awstypes.S3JobDefinition

	s3JobMap := s3JobDefinitionObj[0].(map[string]any)

	if v1, ok1 := s3JobMap["bucket_criteria"]; ok1 && len(v1.([]any)) > 0 {
		s3JobDefinition.BucketCriteria = expandS3BucketCriteriaForJob(v1.([]any))
	}
	if v1, ok1 := s3JobMap["bucket_definitions"]; ok1 && len(v1.([]any)) > 0 {
		s3JobDefinition.BucketDefinitions = expandBucketDefinitions(v1.([]any))
	}
	if v1, ok1 := s3JobMap["scoping"]; ok1 && len(v1.([]any)) > 0 {
		s3JobDefinition.Scoping = expandScoping(v1.([]any))
	}

	return &s3JobDefinition
}

func expandS3BucketCriteriaForJob(criteria []any) *awstypes.S3BucketCriteriaForJob {
	if len(criteria) == 0 {
		return nil
	}

	var criteriaObj awstypes.S3BucketCriteriaForJob

	criteriaMap := criteria[0].(map[string]any)

	if v, ok := criteriaMap["excludes"]; ok && len(v.([]any)) > 0 {
		v1 := v.([]any)
		andMap := v1[0].(map[string]any)
		if v2, ok1 := andMap["and"]; ok1 && len(v2.([]any)) > 0 {
			criteriaObj.Excludes = &awstypes.CriteriaBlockForJob{
				And: expandCriteriaBlockForJob(v2.([]any)),
			}
		}
	}
	if v, ok := criteriaMap["includes"]; ok && len(v.([]any)) > 0 {
		v1 := v.([]any)
		andMap := v1[0].(map[string]any)
		if v2, ok1 := andMap["and"]; ok1 && len(v2.([]any)) > 0 {
			criteriaObj.Includes = &awstypes.CriteriaBlockForJob{
				And: expandCriteriaBlockForJob(v2.([]any)),
			}
		}
	}

	return &criteriaObj
}

func expandCriteriaBlockForJob(criteriaBlocks []any) []awstypes.CriteriaForJob {
	if len(criteriaBlocks) == 0 {
		return nil
	}

	var criteriaBlocksList []awstypes.CriteriaForJob

	for _, v := range criteriaBlocks {
		v1 := v.(map[string]any)
		var criteriaBlock awstypes.CriteriaForJob

		if v2, ok1 := v1["simple_criterion"]; ok1 && len(v2.([]any)) > 0 {
			criteriaBlock.SimpleCriterion = expandSimpleCriterionForJob(v2.([]any))
		}
		if v2, ok1 := v1["tag_criterion"]; ok1 && len(v2.([]any)) > 0 {
			criteriaBlock.TagCriterion = expandTagCriterionForJob(v2.([]any))
		}

		criteriaBlocksList = append(criteriaBlocksList, criteriaBlock)
	}

	return criteriaBlocksList
}

func expandSimpleCriterionForJob(criterion []any) *awstypes.SimpleCriterionForJob {
	if len(criterion) == 0 {
		return nil
	}

	var simpleCriterion awstypes.SimpleCriterionForJob

	simpleCriterionMap := criterion[0].(map[string]any)

	if v, ok := simpleCriterionMap["comparator"]; ok && v.(string) != "" {
		simpleCriterion.Comparator = awstypes.JobComparator(v.(string))
	}
	if v, ok := simpleCriterionMap[names.AttrKey]; ok && v.(string) != "" {
		simpleCriterion.Key = awstypes.SimpleCriterionKeyForJob(v.(string))
	}
	if v, ok := simpleCriterionMap[names.AttrValues]; ok && len(v.([]any)) > 0 {
		simpleCriterion.Values = flex.ExpandStringValueList(v.([]any))
	}

	return &simpleCriterion
}

func expandTagCriterionForJob(criterion []any) *awstypes.TagCriterionForJob {
	if len(criterion) == 0 {
		return nil
	}

	var tagCriterion awstypes.TagCriterionForJob

	tagCriterionMap := criterion[0].(map[string]any)

	if v, ok := tagCriterionMap["comparator"]; ok && v.(string) != "" {
		tagCriterion.Comparator = awstypes.JobComparator(v.(string))
	}
	if v, ok := tagCriterionMap["tag_values"]; ok && len(v.([]any)) > 0 {
		tagCriterion.TagValues = expandTagCriterionPairForJob(v.([]any))
	}

	return &tagCriterion
}

func expandTagCriterionPairForJob(tagValues []any) []awstypes.TagCriterionPairForJob {
	if len(tagValues) == 0 {
		return nil
	}

	var tagValuesList []awstypes.TagCriterionPairForJob

	for _, v := range tagValues {
		v1 := v.(map[string]any)
		var tagValue awstypes.TagCriterionPairForJob

		if v2, ok := v1[names.AttrValue]; ok && v2.(string) != "" {
			tagValue.Value = aws.String(v2.(string))
		}
		if v2, ok := v1[names.AttrKey]; ok && v2.(string) != "" {
			tagValue.Key = aws.String(v2.(string))
		}
		tagValuesList = append(tagValuesList, tagValue)
	}

	return tagValuesList
}

func expandBucketDefinitions(definitions []any) []awstypes.S3BucketDefinitionForJob {
	if len(definitions) == 0 {
		return nil
	}

	var bucketDefinitions []awstypes.S3BucketDefinitionForJob

	for _, v := range definitions {
		v1 := v.(map[string]any)

		bucketDefinition := &awstypes.S3BucketDefinitionForJob{
			Buckets:   flex.ExpandStringValueList(v1["buckets"].([]any)),
			AccountId: aws.String(v1[names.AttrAccountID].(string)),
		}

		bucketDefinitions = append(bucketDefinitions, *bucketDefinition)
	}

	return bucketDefinitions
}

func expandScoping(scoping []any) *awstypes.Scoping {
	if len(scoping) == 0 {
		return nil
	}

	var scopingObj awstypes.Scoping

	scopingMap := scoping[0].(map[string]any)

	if v, ok := scopingMap["excludes"]; ok && len(v.([]any)) > 0 {
		v1 := v.([]any)
		andMap := v1[0].(map[string]any)
		if v2, ok1 := andMap["and"]; ok1 && len(v2.([]any)) > 0 {
			scopingObj.Excludes = &awstypes.JobScopingBlock{
				And: expandJobScopeTerm(v2.([]any)),
			}
		}
	}
	if v, ok := scopingMap["includes"]; ok && len(v.([]any)) > 0 {
		v1 := v.([]any)
		andMap := v1[0].(map[string]any)
		if v2, ok1 := andMap["and"]; ok1 && len(v2.([]any)) > 0 {
			scopingObj.Includes = &awstypes.JobScopingBlock{
				And: expandJobScopeTerm(v2.([]any)),
			}
		}
	}

	return &scopingObj
}

func expandJobScopeTerm(scopeTerms []any) []awstypes.JobScopeTerm {
	if len(scopeTerms) == 0 {
		return nil
	}

	var scopeTermsList []awstypes.JobScopeTerm

	for _, v := range scopeTerms {
		v1 := v.(map[string]any)
		var scopeTerm awstypes.JobScopeTerm

		if v2, ok1 := v1["simple_scope_term"]; ok1 && len(v2.([]any)) > 0 {
			scopeTerm.SimpleScopeTerm = expandSimpleScopeTerm(v2.([]any))
		}
		if v2, ok1 := v1["tag_scope_term"]; ok1 && len(v2.([]any)) > 0 {
			scopeTerm.TagScopeTerm = expandTagScopeTerm(v2.([]any))
		}
		scopeTermsList = append(scopeTermsList, scopeTerm)
	}

	return scopeTermsList
}

func expandSimpleScopeTerm(simpleScopeTerm []any) *awstypes.SimpleScopeTerm {
	if len(simpleScopeTerm) == 0 {
		return nil
	}

	var simpleTerm awstypes.SimpleScopeTerm

	simpleScopeTermMap := simpleScopeTerm[0].(map[string]any)

	if v, ok := simpleScopeTermMap[names.AttrKey]; ok && v.(string) != "" {
		simpleTerm.Key = awstypes.ScopeFilterKey(v.(string))
	}
	if v, ok := simpleScopeTermMap[names.AttrValues]; ok && len(v.([]any)) > 0 {
		simpleTerm.Values = flex.ExpandStringValueList(v.([]any))
	}
	if v, ok := simpleScopeTermMap["comparator"]; ok && v.(string) != "" {
		simpleTerm.Comparator = awstypes.JobComparator(v.(string))
	}

	return &simpleTerm
}

func expandTagScopeTerm(tagScopeTerm []any) *awstypes.TagScopeTerm {
	if len(tagScopeTerm) == 0 {
		return nil
	}

	var tagTerm awstypes.TagScopeTerm

	tagScopeTermMap := tagScopeTerm[0].(map[string]any)

	if v, ok := tagScopeTermMap[names.AttrKey]; ok && v.(string) != "" {
		tagTerm.Key = aws.String(v.(string))
	}
	if v, ok := tagScopeTermMap["tag_values"]; ok && len(v.([]any)) > 0 {
		tagTerm.TagValues = expandTagValues(v.([]any))
	}
	if v, ok := tagScopeTermMap["comparator"]; ok && v.(string) != "" {
		tagTerm.Comparator = awstypes.JobComparator(v.(string))
	}
	if v, ok := tagScopeTermMap[names.AttrTarget]; ok && v.(string) != "" {
		tagTerm.Target = awstypes.TagTarget(v.(string))
	}

	return &tagTerm
}

func expandTagValues(tagValues []any) []awstypes.TagValuePair {
	if len(tagValues) == 0 {
		return nil
	}

	var tagValuesList []awstypes.TagValuePair

	for _, v := range tagValues {
		v1 := v.(map[string]any)
		var tagValue awstypes.TagValuePair

		if v2, ok := v1[names.AttrValue]; ok && v2.(string) != "" {
			tagValue.Value = aws.String(v2.(string))
		}
		if v2, ok := v1[names.AttrKey]; ok && v2.(string) != "" {
			tagValue.Key = aws.String(v2.(string))
		}
		tagValuesList = append(tagValuesList, tagValue)
	}

	return tagValuesList
}

func expandScheduleFrequency(schedules []any) *awstypes.JobScheduleFrequency {
	if len(schedules) == 0 {
		return nil
	}

	var jobScheduleFrequency awstypes.JobScheduleFrequency

	scheduleMap := schedules[0].(map[string]any)

	if v1, ok1 := scheduleMap["daily_schedule"]; ok1 && v1.(bool) {
		jobScheduleFrequency.DailySchedule = &awstypes.DailySchedule{}
	}
	if v1, ok1 := scheduleMap["weekly_schedule"]; ok1 && v1.(string) != "" {
		jobScheduleFrequency.WeeklySchedule = &awstypes.WeeklySchedule{
			DayOfWeek: awstypes.DayOfWeek(v1.(string)),
		}
	}
	if v1, ok1 := scheduleMap["monthly_schedule"]; ok1 && v1.(int) > 0 {
		jobScheduleFrequency.MonthlySchedule = &awstypes.MonthlySchedule{
			DayOfMonth: aws.Int32(int32(v1.(int))),
		}
	}

	return &jobScheduleFrequency
}

func flattenScheduleFrequency(schedule *awstypes.JobScheduleFrequency) []map[string]any {
	if schedule == nil {
		return nil
	}

	var schedulesList []map[string]any
	schedMap := map[string]any{}
	if schedule.DailySchedule != nil {
		schedMap["daily_schedule"] = true
	}
	if schedule.WeeklySchedule != nil {
		schedMap["weekly_schedule"] = schedule.WeeklySchedule.DayOfWeek
	}
	if schedule.MonthlySchedule != nil && schedule.MonthlySchedule.DayOfMonth != nil {
		schedMap["monthly_schedule"] = schedule.MonthlySchedule.DayOfMonth
	}
	schedulesList = append(schedulesList, schedMap)

	return schedulesList
}

func flattenS3JobDefinition(s3JobDefinition *awstypes.S3JobDefinition) []map[string]any {
	if s3JobDefinition == nil {
		return nil
	}

	var jobDefinitions []map[string]any

	jobDefinitions = append(jobDefinitions, map[string]any{
		"bucket_criteria":    flattenS3BucketCriteriaForJob(s3JobDefinition.BucketCriteria),
		"bucket_definitions": flattenBucketDefinition(s3JobDefinition.BucketDefinitions),
		"scoping":            flattenScoping(s3JobDefinition.Scoping),
	})

	return jobDefinitions
}

func flattenS3BucketCriteriaForJob(criteria *awstypes.S3BucketCriteriaForJob) []map[string]any {
	if criteria == nil {
		return nil
	}

	var criteriaList []map[string]any

	criteriaList = append(criteriaList, map[string]any{
		"excludes": flattenCriteriaBlockForJob(criteria.Excludes),
		"includes": flattenCriteriaBlockForJob(criteria.Includes),
	})

	return criteriaList
}

func flattenCriteriaBlockForJob(criteriaBlock *awstypes.CriteriaBlockForJob) []map[string]any {
	if criteriaBlock == nil {
		return nil
	}

	var criteriaBlockList []map[string]any

	criteriaBlockList = append(criteriaBlockList, map[string]any{
		"and": flattenCriteriaForJob(criteriaBlock.And),
	})

	return criteriaBlockList
}

func flattenCriteriaForJob(criteria []awstypes.CriteriaForJob) []map[string]any {
	if criteria == nil {
		return nil
	}

	var criteriaList []map[string]any

	for _, criterion := range criteria {
		criteriaList = append(criteriaList, map[string]any{
			"simple_criterion": flattenSimpleCriterionForJob(criterion.SimpleCriterion),
			"tag_criterion":    flattenTagCriterionForJob(criterion.TagCriterion),
		})
	}

	return criteriaList
}

func flattenSimpleCriterionForJob(criterion *awstypes.SimpleCriterionForJob) []map[string]any {
	if criterion == nil {
		return nil
	}

	var simpleCriterionList []map[string]any

	simpleCriterionList = append(simpleCriterionList, map[string]any{
		"comparator":     string(criterion.Comparator),
		names.AttrKey:    string(criterion.Key),
		names.AttrValues: flex.FlattenStringValueList(criterion.Values),
	})

	return simpleCriterionList
}

func flattenTagCriterionForJob(criterion *awstypes.TagCriterionForJob) []map[string]any {
	if criterion == nil {
		return nil
	}

	var tagCriterionList []map[string]any

	tagCriterionList = append(tagCriterionList, map[string]any{
		"comparator": string(criterion.Comparator),
		"tag_values": flattenTagCriterionPairForJob(criterion.TagValues),
	})

	return tagCriterionList
}

func flattenTagCriterionPairForJob(tagValues []awstypes.TagCriterionPairForJob) []map[string]any {
	if len(tagValues) == 0 {
		return nil
	}

	var tagValuesList []map[string]any

	for _, tagValue := range tagValues {
		tagValuesList = append(tagValuesList, map[string]any{
			names.AttrValue: aws.ToString(tagValue.Value),
			names.AttrKey:   aws.ToString(tagValue.Key),
		})
	}

	return tagValuesList
}

func flattenBucketDefinition(bucketDefinitions []awstypes.S3BucketDefinitionForJob) []map[string]any {
	if len(bucketDefinitions) == 0 {
		return nil
	}

	var bucketDefinitionList []map[string]any

	for _, bucket := range bucketDefinitions {
		bucketDefinitionList = append(bucketDefinitionList, map[string]any{
			names.AttrAccountID: aws.ToString(bucket.AccountId),
			"buckets":           flex.FlattenStringValueList(bucket.Buckets),
		})
	}

	return bucketDefinitionList
}

func flattenScoping(scoping *awstypes.Scoping) []map[string]any {
	if scoping == nil {
		return nil
	}

	var scopingList []map[string]any

	scopingList = append(scopingList, map[string]any{
		"excludes": flattenJobScopingBlock(scoping.Excludes),
		"includes": flattenJobScopingBlock(scoping.Includes),
	})

	return scopingList
}

func flattenJobScopingBlock(scopeTerm *awstypes.JobScopingBlock) []map[string]any {
	if scopeTerm == nil {
		return nil
	}

	var scopeTermList []map[string]any

	scopeTermList = append(scopeTermList, map[string]any{
		"and": flattenJobScopeTerm(scopeTerm.And),
	})

	return scopeTermList
}

func flattenJobScopeTerm(scopeTerms []awstypes.JobScopeTerm) []map[string]any {
	if scopeTerms == nil {
		return nil
	}

	var scopeTermList []map[string]any

	for _, scopeTerm := range scopeTerms {
		scopeTermList = append(scopeTermList, map[string]any{
			"simple_scope_term": flattenSimpleScopeTerm(scopeTerm.SimpleScopeTerm),
			"tag_scope_term":    flattenTagScopeTerm(scopeTerm.TagScopeTerm),
		})
	}

	return scopeTermList
}

func flattenSimpleScopeTerm(simpleScopeTerm *awstypes.SimpleScopeTerm) []map[string]any {
	if simpleScopeTerm == nil {
		return nil
	}

	var simpleScopeTermList []map[string]any

	simpleScopeTermList = append(simpleScopeTermList, map[string]any{
		names.AttrKey:    string(simpleScopeTerm.Key),
		"comparator":     string(simpleScopeTerm.Comparator),
		names.AttrValues: flex.FlattenStringValueList(simpleScopeTerm.Values),
	})

	return simpleScopeTermList
}

func flattenTagScopeTerm(tagScopeTerm *awstypes.TagScopeTerm) []map[string]any {
	if tagScopeTerm == nil {
		return nil
	}

	var tagScopeTermList []map[string]any

	tagScopeTermList = append(tagScopeTermList, map[string]any{
		names.AttrKey:    aws.ToString(tagScopeTerm.Key),
		"comparator":     string(tagScopeTerm.Comparator),
		names.AttrTarget: string(tagScopeTerm.Target),
		"tag_values":     flattenTagValues(tagScopeTerm.TagValues),
	})

	return tagScopeTermList
}

func flattenTagValues(tagValues []awstypes.TagValuePair) []map[string]any {
	if len(tagValues) == 0 {
		return nil
	}

	var tagValuesList []map[string]any

	for _, tagValue := range tagValues {
		tagValuesList = append(tagValuesList, map[string]any{
			names.AttrValue: aws.ToString(tagValue.Value),
			names.AttrKey:   aws.ToString(tagValue.Key),
		})
	}

	return tagValuesList
}

func flattenUserPausedDetails(userPausedDetail *awstypes.UserPausedDetails) []map[string]any {
	if userPausedDetail == nil {
		return nil
	}

	var userDetails []map[string]any

	userDetails = append(userDetails, map[string]any{
		"job_imminent_expiration_health_event_arn": aws.ToString(userPausedDetail.JobImminentExpirationHealthEventArn),
		"job_expires_at": userPausedDetail.JobExpiresAt.String(),
		"job_paused_at":  userPausedDetail.JobPausedAt.String(),
	})

	return userDetails
}
