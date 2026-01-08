// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dlm

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dlm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dlm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dlm_lifecycle_policy", name="Lifecycle Policy")
// @Tags(identifierAttribute="arn")
func resourceLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLifecyclePolicyCreate,
		ReadWithoutTimeout:   resourceLifecyclePolicyRead,
		UpdateWithoutTimeout: resourceLifecyclePolicyUpdate,
		DeleteWithoutTimeout: resourceLifecyclePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile("^[0-9A-Za-z _-]+$"), "see https://docs.aws.amazon.com/cli/latest/reference/dlm/create-lifecycle-policy.html"),
					validation.StringLenBetween(1, 500),
				),
			},
			"default_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DefaultPolicyTypeValues](),
			},
			names.AttrExecutionRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"policy_details": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAction: {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cross_region_copy": {
										Type:     schema.TypeSet,
										Required: true,
										MaxItems: 3,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEncryptionConfiguration: {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"cmk_arn": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: verify.ValidARN,
															},
															names.AttrEncrypted: {
																Type:     schema.TypeBool,
																Optional: true,
																Default:  false,
															},
														},
													},
												},
												"retain_rule": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrInterval: {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"interval_unit": {
																Type:             schema.TypeString,
																Required:         true,
																ValidateDiagFunc: enum.Validate[awstypes.RetentionIntervalUnitValues](),
															},
														},
													},
												},
												names.AttrTarget: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[\w:\-\/\*]+$`), ""),
												},
											},
										},
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 120),
											validation.StringMatch(regexache.MustCompile("^[0-9A-Za-z _-]+$"), "see https://docs.aws.amazon.com/dlm/latest/APIReference/API_Action.html"),
										),
									},
								},
							},
						},
						"copy_tags": {
							Type:          schema.TypeBool,
							Optional:      true,
							Default:       false,
							ConflictsWith: []string{"policy_details.0.schedule"},
							RequiredWith:  []string{"default_policy"},
						},
						"create_interval": {
							Type:          schema.TypeInt,
							Optional:      true,
							Default:       1,
							ValidateFunc:  validation.IntBetween(1, 7),
							ConflictsWith: []string{"policy_details.0.schedule"},
							RequiredWith:  []string{"default_policy"},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if d.Get("default_policy").(string) == "" {
									if old == "0" && new == "1" {
										return true
									}
								}
								return false
							},
						},
						"exclusions": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							RequiredWith:  []string{"default_policy"},
							ConflictsWith: []string{"policy_details.0.resource_types", "policy_details.0.schedule"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"exclude_boot_volumes": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"exclude_tags": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"exclude_volume_types": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 6,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"extend_deletion": {
							Type:          schema.TypeBool,
							Optional:      true,
							Default:       false,
							ConflictsWith: []string{"policy_details.0.schedule"},
							RequiredWith:  []string{"default_policy"},
						},
						"event_source": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrParameters: {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"description_regex": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 1000),
												},
												"event_type": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.EventTypeValues](),
												},
												"snapshot_owner": {
													Type:     schema.TypeSet,
													Required: true,
													MaxItems: 50,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidAccountID,
													},
												},
											},
										},
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EventSourceValues](),
									},
								},
							},
						},
						names.AttrParameters: {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"exclude_boot_volume": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"no_reboot": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"policy_language": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PolicyLanguageValues](),
						},
						"policy_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.PolicyTypeValuesEbsSnapshotManagement,
							ValidateDiagFunc: enum.Validate[awstypes.PolicyTypeValues](),
						},
						"resource_locations": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.ResourceLocationValues](),
							},
						},
						names.AttrResourceType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ResourceTypeValues](),
							ConflictsWith:    []string{"policy_details.0.resource_types", "policy_details.0.schedule"},
							RequiredWith:     []string{"default_policy"},
						},
						"resource_types": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.ResourceTypeValues](),
							},
							ConflictsWith: []string{"policy_details.0.resource_type", "default_policy"},
						},
						"retain_interval": {
							Type:          schema.TypeInt,
							Optional:      true,
							Default:       7,
							ValidateFunc:  validation.IntBetween(2, 14),
							ConflictsWith: []string{"policy_details.0.schedule"},
							RequiredWith:  []string{"default_policy"},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if d.Get("default_policy").(string) == "" {
									if old == "0" && new == "7" {
										return true
									}
								}
								return false
							},
						},
						names.AttrSchedule: {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"archive_rule": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"archive_retain_rule": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"retention_archive_tier": {
																Type:     schema.TypeList,
																Required: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"count": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntBetween(1, 1000),
																		},
																		names.AttrInterval: {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		"interval_unit": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.RetentionIntervalUnitValues](),
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
									"copy_tags": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"create_rule": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cron_expression": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringMatch(regexache.MustCompile("^cron\\([^\n]{11,100}\\)$"), "see https://docs.aws.amazon.com/dlm/latest/APIReference/API_CreateRule.html"),
												},
												names.AttrInterval: {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntInSlice([]int{1, 2, 3, 4, 6, 8, 12, 24}),
												},
												"interval_unit": {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[awstypes.IntervalUnitValues](),
												},
												names.AttrLocation: {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[awstypes.LocationValues](),
												},
												"scripts": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"execute_operation_on_script_failure": {
																Type:     schema.TypeBool,
																Computed: true,
																Optional: true,
															},
															"execution_handler": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(0, 200),
																	validation.StringMatch(regexache.MustCompile("^([a-zA-Z0-9_\\-.]{3,128}|[a-zA-Z0-9_\\-.:/]{3,200}|[A-Z0-9_]+)$"), "see https://docs.aws.amazon.com/dlm/latest/APIReference/API_Action.html"),
																),
															},
															"execution_handler_service": {
																Type:             schema.TypeString,
																Computed:         true,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[awstypes.ExecutionHandlerServiceValues](),
															},
															"execution_timeout": {
																Type:         schema.TypeInt,
																Computed:     true,
																Optional:     true,
																ValidateFunc: validation.IntBetween(1, 120),
															},
															"maximum_retry_count": {
																Type:         schema.TypeInt,
																Computed:     true,
																Optional:     true,
																ValidateFunc: validation.IntBetween(1, 3),
															},
															"stages": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 1,
																MaxItems: 2,
																Elem: &schema.Schema{
																	Type:             schema.TypeString,
																	ValidateDiagFunc: enum.Validate[awstypes.StageValues](),
																},
															},
														},
													},
												},
												"times": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													MaxItems: 1,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringMatch(regexache.MustCompile("^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$"), "see https://docs.aws.amazon.com/dlm/latest/APIReference/API_CreateRule.html#dlm-Type-CreateRule-Times"),
													},
												},
											},
										},
									},
									"cross_region_copy_rule": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 3,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cmk_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"copy_tags": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
												"deprecate_rule": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrInterval: {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"interval_unit": {
																Type:             schema.TypeString,
																Required:         true,
																ValidateDiagFunc: enum.Validate[awstypes.RetentionIntervalUnitValues](),
															},
														},
													},
												},
												names.AttrEncrypted: {
													Type:     schema.TypeBool,
													Required: true,
												},
												"retain_rule": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrInterval: {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"interval_unit": {
																Type:             schema.TypeString,
																Required:         true,
																ValidateDiagFunc: enum.Validate[awstypes.RetentionIntervalUnitValues](),
															},
														},
													},
												},
												names.AttrTarget: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[\w:\-\/\*]+$`), ""),
												},
												"target_region": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidRegionName,
												},
											},
										},
									},
									"deprecate_rule": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"count": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 1000),
												},
												names.AttrInterval: {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"interval_unit": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.RetentionIntervalUnitValues](),
												},
											},
										},
									},
									"fast_restore_rule": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrAvailabilityZones: {
													Type:     schema.TypeSet,
													Required: true,
													MinItems: 1,
													MaxItems: 10,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"count": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 1000),
												},
												names.AttrInterval: {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"interval_unit": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.RetentionIntervalUnitValues](),
												},
											},
										},
									},
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 120),
									},
									"retain_rule": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"count": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 1000),
												},
												names.AttrInterval: {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"interval_unit": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.RetentionIntervalUnitValues](),
												},
											},
										},
									},
									"share_rule": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target_accounts": {
													Type:     schema.TypeSet,
													Required: true,
													MinItems: 1,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidAccountID,
													},
												},
												"unshare_interval": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"unshare_interval_unit": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.RetentionIntervalUnitValues](),
												},
											},
										},
									},
									"tags_to_add": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"variable_tags": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"target_tags": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrState: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.SettablePolicyStateValuesEnabled,
				ValidateDiagFunc: enum.Validate[awstypes.SettablePolicyStateValues](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	ResNameLifecyclePolicy = "Lifecycle Policy"
)

func resourceLifecyclePolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DLMClient(ctx)

	input := dlm.CreateLifecyclePolicyInput{
		Description:      aws.String(d.Get(names.AttrDescription).(string)),
		ExecutionRoleArn: aws.String(d.Get(names.AttrExecutionRoleARN).(string)),
		PolicyDetails:    expandPolicyDetails(d.Get("policy_details").([]any), d.Get("default_policy").(string)),
		State:            awstypes.SettablePolicyStateValues(d.Get(names.AttrState).(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("default_policy"); ok {
		input.DefaultPolicy = awstypes.DefaultPolicyTypeValues(v.(string))
	}

	const (
		timeout = 2 * time.Minute
	)
	output, err := tfresource.RetryWhenIsA[*dlm.CreateLifecyclePolicyOutput, *awstypes.InvalidRequestException](ctx, timeout, func(ctx context.Context) (*dlm.CreateLifecyclePolicyOutput, error) {
		return conn.CreateLifecyclePolicy(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DLM Lifecycle Policy: %s", err)
	}

	d.SetId(aws.ToString(output.PolicyId))

	return append(diags, resourceLifecyclePolicyRead(ctx, d, meta)...)
}

func resourceLifecyclePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DLMClient(ctx)

	output, err := findLifecyclePolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] DLM Lifecycle Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DLM Lifecycle Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Policy.PolicyArn)
	if aws.ToBool(output.Policy.DefaultPolicy) {
		d.Set("default_policy", d.Get("default_policy"))
	}
	d.Set(names.AttrDescription, output.Policy.Description)
	d.Set(names.AttrExecutionRoleARN, output.Policy.ExecutionRoleArn)
	if err := d.Set("policy_details", flattenPolicyDetails(output.Policy.PolicyDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting policy details %s", err)
	}
	d.Set(names.AttrState, output.Policy.State)

	setTagsOut(ctx, output.Policy.Tags)

	return diags
}

func resourceLifecyclePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DLMClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := dlm.UpdateLifecyclePolicyInput{
			PolicyId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}
		if d.HasChange(names.AttrExecutionRoleARN) {
			input.ExecutionRoleArn = aws.String(d.Get(names.AttrExecutionRoleARN).(string))
		}
		if d.HasChange("policy_details") {
			input.PolicyDetails = expandPolicyDetails(d.Get("policy_details").([]any), d.Get("default_policy").(string))
		}
		if d.HasChange(names.AttrState) {
			input.State = awstypes.SettablePolicyStateValues(d.Get(names.AttrState).(string))
		}

		_, err := conn.UpdateLifecyclePolicy(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DLM Lifecycle Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLifecyclePolicyRead(ctx, d, meta)...)
}

func resourceLifecyclePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DLMClient(ctx)

	log.Printf("[INFO] Deleting DLM lifecycle policy: %s", d.Id())
	input := dlm.DeleteLifecyclePolicyInput{
		PolicyId: aws.String(d.Id()),
	}
	_, err := conn.DeleteLifecyclePolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DLM Lifecycle Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findLifecyclePolicyByID(ctx context.Context, conn *dlm.Client, id string) (*dlm.GetLifecyclePolicyOutput, error) {
	input := dlm.GetLifecyclePolicyInput{
		PolicyId: aws.String(id),
	}

	return findLifecyclePolicy(ctx, conn, &input)
}

func findLifecyclePolicy(ctx context.Context, conn *dlm.Client, input *dlm.GetLifecyclePolicyInput) (*dlm.GetLifecyclePolicyOutput, error) {
	output, err := conn.GetLifecyclePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandPolicyDetails(tfList []any, defaultPolicyValue string) *awstypes.PolicyDetails {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	policyType := awstypes.PolicyTypeValues(tfMap["policy_type"].(string))
	apiObject := &awstypes.PolicyDetails{
		PolicyType: policyType,
	}

	if defaultPolicyValue != "" {
		if v, ok := tfMap["copy_tags"].(bool); ok {
			apiObject.CopyTags = aws.Bool(v)
		}
		if v, ok := tfMap["create_interval"].(int); ok {
			apiObject.CreateInterval = aws.Int32(int32(v))
		}
		if v, ok := tfMap["exclusions"].([]any); ok && len(v) > 0 {
			apiObject.Exclusions = expandExclusions(v)
		}
		if v, ok := tfMap["extend_deletion"].(bool); ok {
			apiObject.ExtendDeletion = aws.Bool(v)
		}
		if v, ok := tfMap[names.AttrResourceType].(string); ok {
			apiObject.ResourceType = awstypes.ResourceTypeValues(v)
		}
		if v, ok := tfMap["retain_interval"].(int); ok {
			apiObject.RetainInterval = aws.Int32(int32(v))
		}
	}

	if v, ok := tfMap[names.AttrAction].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandActions(v)
	}
	if v, ok := tfMap["event_source"].([]any); ok && len(v) > 0 {
		apiObject.EventSource = expandEventSource(v)
	}
	if v, ok := tfMap[names.AttrParameters].([]any); ok && len(v) > 0 {
		apiObject.Parameters = expandParameters(v, policyType)
	}
	if v, ok := tfMap["policy_language"].(string); ok {
		apiObject.PolicyLanguage = awstypes.PolicyLanguageValues(v)
	}
	if v, ok := tfMap["resource_types"].([]any); ok && len(v) > 0 {
		apiObject.ResourceTypes = flex.ExpandStringyValueList[awstypes.ResourceTypeValues](v)
	}
	if v, ok := tfMap["resource_locations"].([]any); ok && len(v) > 0 {
		apiObject.ResourceLocations = flex.ExpandStringyValueList[awstypes.ResourceLocationValues](v)
	}
	if v, ok := tfMap[names.AttrSchedule].([]any); ok && len(v) > 0 {
		apiObject.Schedules = expandSchedules(v)
	}
	if v, ok := tfMap["target_tags"].(map[string]any); ok && len(v) > 0 {
		apiObject.TargetTags = expandTags(v)
	}

	return apiObject
}

func flattenPolicyDetails(apiObject *awstypes.PolicyDetails) []any {
	tfMap := make(map[string]any)
	tfMap[names.AttrAction] = flattenActions(apiObject.Actions)
	tfMap["copy_tags"] = aws.ToBool(apiObject.CopyTags)
	tfMap["create_interval"] = aws.ToInt32(apiObject.CreateInterval)
	tfMap["event_source"] = flattenEventSource(apiObject.EventSource)
	tfMap["exclusions"] = flattenExclusions(apiObject.Exclusions)
	tfMap["extend_deletion"] = aws.ToBool(apiObject.ExtendDeletion)
	if apiObject.Parameters != nil {
		tfMap[names.AttrParameters] = flattenParameters(apiObject.Parameters)
	}
	tfMap["policy_language"] = apiObject.PolicyLanguage
	tfMap["policy_type"] = apiObject.PolicyType
	tfMap[names.AttrResourceType] = apiObject.ResourceType
	tfMap["resource_types"] = apiObject.ResourceTypes
	tfMap["resource_locations"] = apiObject.ResourceLocations
	tfMap["retain_interval"] = aws.ToInt32(apiObject.RetainInterval)
	tfMap[names.AttrSchedule] = flattenSchedules(apiObject.Schedules)
	tfMap["target_tags"] = flattenTags(apiObject.TargetTags)

	return []any{tfMap}
}

func expandSchedules(tfList []any) []awstypes.Schedule {
	apiObjects := make([]awstypes.Schedule, len(tfList))

	for i, tfMapRaw := range tfList {
		apiObject := awstypes.Schedule{}
		tfMap := tfMapRaw.(map[string]any)

		if v, ok := tfMap["archive_rule"].([]any); ok && len(v) > 0 {
			apiObject.ArchiveRule = expandArchiveRule(v)
		}
		if v, ok := tfMap["copy_tags"]; ok {
			apiObject.CopyTags = aws.Bool(v.(bool))
		}
		if v, ok := tfMap["create_rule"]; ok {
			apiObject.CreateRule = expandCreateRule(v.([]any))
		}
		if v, ok := tfMap["cross_region_copy_rule"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.CrossRegionCopyRules = expandCrossRegionCopyRules(v.List())
		}
		if v, ok := tfMap["deprecate_rule"]; ok {
			apiObject.DeprecateRule = expandDeprecateRule(v.([]any))
		}
		if v, ok := tfMap["fast_restore_rule"]; ok {
			apiObject.FastRestoreRule = expandFastRestoreRule(v.([]any))
		}
		if v, ok := tfMap[names.AttrName]; ok {
			apiObject.Name = aws.String(v.(string))
		}
		if v, ok := tfMap["retain_rule"]; ok {
			apiObject.RetainRule = expandRetainRule(v.([]any))
		}
		if v, ok := tfMap["share_rule"]; ok {
			apiObject.ShareRules = expandShareRule(v.([]any))
		}
		if v, ok := tfMap["tags_to_add"]; ok {
			apiObject.TagsToAdd = expandTags(v.(map[string]any))
		}
		if v, ok := tfMap["variable_tags"]; ok {
			apiObject.VariableTags = expandTags(v.(map[string]any))
		}

		apiObjects[i] = apiObject
	}

	return apiObjects
}

func flattenSchedules(apiObjects []awstypes.Schedule) []any {
	tfList := make([]any, len(apiObjects))

	for i, apiObject := range apiObjects {
		tfMap := make(map[string]any)
		tfMap["archive_rule"] = flattenArchiveRule(apiObject.ArchiveRule)
		tfMap["copy_tags"] = aws.ToBool(apiObject.CopyTags)
		tfMap["create_rule"] = flattenCreateRule(apiObject.CreateRule)
		tfMap["cross_region_copy_rule"] = flattenCrossRegionCopyRules(apiObject.CrossRegionCopyRules)
		if apiObject.DeprecateRule != nil {
			tfMap["deprecate_rule"] = flattenDeprecateRule(apiObject.DeprecateRule)
		}
		if apiObject.FastRestoreRule != nil {
			tfMap["fast_restore_rule"] = flattenFastRestoreRule(apiObject.FastRestoreRule)
		}
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		tfMap["retain_rule"] = flattenRetainRule(apiObject.RetainRule)
		if apiObject.ShareRules != nil {
			tfMap["share_rule"] = flattenShareRule(apiObject.ShareRules)
		}
		tfMap["tags_to_add"] = flattenTags(apiObject.TagsToAdd)
		tfMap["variable_tags"] = flattenTags(apiObject.VariableTags)

		tfList[i] = tfMap
	}

	return tfList
}

func expandActions(tfList []any) []awstypes.Action {
	apiObjects := make([]awstypes.Action, len(tfList))

	for i, tfMapRaw := range tfList {
		apiObject := awstypes.Action{}
		tfMap := tfMapRaw.(map[string]any)

		if v, ok := tfMap["cross_region_copy"].(*schema.Set); ok {
			apiObject.CrossRegionCopy = expandActionCrossRegionCopyRules(v.List())
		}
		if v, ok := tfMap[names.AttrName]; ok {
			apiObject.Name = aws.String(v.(string))
		}

		apiObjects[i] = apiObject
	}

	return apiObjects
}

func flattenActions(apiObjects []awstypes.Action) []any {
	tfList := make([]any, len(apiObjects))

	for i, apiObject := range apiObjects {
		tfMap := make(map[string]any)
		if apiObject.CrossRegionCopy != nil {
			tfMap["cross_region_copy"] = flattenActionCrossRegionCopyRules(apiObject.CrossRegionCopy)
		}
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)

		tfList[i] = tfMap
	}

	return tfList
}

func expandActionCrossRegionCopyRules(tfList []any) []awstypes.CrossRegionCopyAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []awstypes.CrossRegionCopyAction

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.CrossRegionCopyAction{}
		if v, ok := tfMap[names.AttrEncryptionConfiguration].([]any); ok {
			apiObject.EncryptionConfiguration = expandActionCrossRegionCopyRuleEncryptionConfiguration(v)
		}
		if v, ok := tfMap["retain_rule"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.RetainRule = expandCrossRegionCopyRuleRetainRule(v)
		}
		if v, ok := tfMap[names.AttrTarget].(string); ok && v != "" {
			apiObject.Target = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenActionCrossRegionCopyRules(apiObjects []awstypes.CrossRegionCopyAction) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrEncryptionConfiguration: flattenActionCrossRegionCopyRuleEncryptionConfiguration(apiObject.EncryptionConfiguration),
			"retain_rule":                     flattenCrossRegionCopyRuleRetainRule(apiObject.RetainRule),
			names.AttrTarget:                  aws.ToString(apiObject.Target),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandActionCrossRegionCopyRuleEncryptionConfiguration(tfList []any) *awstypes.EncryptionConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.EncryptionConfiguration{
		Encrypted: aws.Bool(tfMap[names.AttrEncrypted].(bool)),
	}

	if v, ok := tfMap["cmk_arn"].(string); ok && v != "" {
		apiObject.CmkArn = aws.String(v)
	}

	return apiObject
}

func flattenActionCrossRegionCopyRuleEncryptionConfiguration(apiObject *awstypes.EncryptionConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"cmk_arn":           aws.ToString(apiObject.CmkArn),
		names.AttrEncrypted: aws.ToBool(apiObject.Encrypted),
	}

	return []any{tfMap}
}

func expandEventSource(tfList []any) *awstypes.EventSource {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.EventSource{
		Type: awstypes.EventSourceValues(tfMap[names.AttrType].(string)),
	}

	if v, ok := tfMap[names.AttrParameters].([]any); ok && len(v) > 0 {
		apiObject.Parameters = expandEventSourceParameters(v)
	}

	return apiObject
}

func flattenEventSource(apiObject *awstypes.EventSource) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrParameters: flattenEventSourceParameters(apiObject.Parameters),
		names.AttrType:       apiObject.Type,
	}

	return []any{tfMap}
}

func expandEventSourceParameters(tfList []any) *awstypes.EventParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.EventParameters{
		DescriptionRegex: aws.String(tfMap["description_regex"].(string)),
		EventType:        awstypes.EventTypeValues(tfMap["event_type"].(string)),
		SnapshotOwner:    flex.ExpandStringValueSet(tfMap["snapshot_owner"].(*schema.Set)),
	}

	return apiObject
}

func flattenEventSourceParameters(apiObject *awstypes.EventParameters) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"description_regex": aws.ToString(apiObject.DescriptionRegex),
		"event_type":        apiObject.EventType,
		"snapshot_owner":    apiObject.SnapshotOwner,
	}

	return []any{tfMap}
}

func expandCrossRegionCopyRules(tfList []any) []awstypes.CrossRegionCopyRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []awstypes.CrossRegionCopyRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.CrossRegionCopyRule{}

		if v, ok := tfMap["cmk_arn"].(string); ok && v != "" {
			apiObject.CmkArn = aws.String(v)
		}
		if v, ok := tfMap["copy_tags"].(bool); ok {
			apiObject.CopyTags = aws.Bool(v)
		}
		if v, ok := tfMap["deprecate_rule"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.DeprecateRule = expandCrossRegionCopyRuleDeprecateRule(v)
		}
		if v, ok := tfMap[names.AttrEncrypted].(bool); ok {
			apiObject.Encrypted = aws.Bool(v)
		}
		if v, ok := tfMap["retain_rule"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.RetainRule = expandCrossRegionCopyRuleRetainRule(v)
		}
		if v, ok := tfMap[names.AttrTarget].(string); ok && v != "" {
			apiObject.Target = aws.String(v)
		}
		if v, ok := tfMap["target_region"].(string); ok && v != "" {
			apiObject.TargetRegion = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCrossRegionCopyRules(apiObjects []awstypes.CrossRegionCopyRule) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"cmk_arn":           aws.ToString(apiObject.CmkArn),
			"copy_tags":         aws.ToBool(apiObject.CopyTags),
			"deprecate_rule":    flattenCrossRegionCopyRuleDeprecateRule(apiObject.DeprecateRule),
			names.AttrEncrypted: aws.ToBool(apiObject.Encrypted),
			"retain_rule":       flattenCrossRegionCopyRuleRetainRule(apiObject.RetainRule),
			names.AttrTarget:    aws.ToString(apiObject.Target),
			"target_region":     aws.ToString(apiObject.TargetRegion),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandCrossRegionCopyRuleDeprecateRule(tfList []any) *awstypes.CrossRegionCopyDeprecateRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	return &awstypes.CrossRegionCopyDeprecateRule{
		Interval:     aws.Int32(int32(tfMap[names.AttrInterval].(int))),
		IntervalUnit: awstypes.RetentionIntervalUnitValues(tfMap["interval_unit"].(string)),
	}
}

func expandCrossRegionCopyRuleRetainRule(tfList []any) *awstypes.CrossRegionCopyRetainRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	return &awstypes.CrossRegionCopyRetainRule{
		Interval:     aws.Int32(int32(tfMap[names.AttrInterval].(int))),
		IntervalUnit: awstypes.RetentionIntervalUnitValues(tfMap["interval_unit"].(string)),
	}
}

func flattenCrossRegionCopyRuleDeprecateRule(apiObject *awstypes.CrossRegionCopyDeprecateRule) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrInterval: aws.ToInt32(apiObject.Interval),
		"interval_unit":    apiObject.IntervalUnit,
	}

	return []any{tfMap}
}

func flattenCrossRegionCopyRuleRetainRule(apiObject *awstypes.CrossRegionCopyRetainRule) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrInterval: aws.ToInt32(apiObject.Interval),
		"interval_unit":    apiObject.IntervalUnit,
	}

	return []any{tfMap}
}

func expandCreateRule(tfList []any) *awstypes.CreateRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.CreateRule{}

	if v, ok := tfMap[names.AttrInterval].(int); ok && v > 0 {
		apiObject.Interval = aws.Int32(int32(v))
	}
	if v, ok := tfMap["interval_unit"].(string); ok && v != "" {
		apiObject.IntervalUnit = awstypes.IntervalUnitValues(v)
	} else {
		apiObject.IntervalUnit = awstypes.IntervalUnitValuesHours
	}
	if v, ok := tfMap[names.AttrLocation].(string); ok && v != "" {
		apiObject.Location = awstypes.LocationValues(v)
	}
	if v, ok := tfMap["scripts"]; ok {
		apiObject.Scripts = expandScripts(v.([]any))
	}
	if v, ok := tfMap["times"].([]any); ok && len(v) > 0 {
		apiObject.Times = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["cron_expression"].(string); ok && v != "" {
		apiObject.CronExpression = aws.String(v)
		apiObject.IntervalUnit = "" // sets interval unit to empty string so that all fields related to interval are ignored
	}

	return apiObject
}

func flattenCreateRule(apiObject *awstypes.CreateRule) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)
	if apiObject.CronExpression != nil {
		tfMap["cron_expression"] = aws.ToString(apiObject.CronExpression)
	}
	if apiObject.Interval != nil {
		tfMap[names.AttrInterval] = aws.ToInt32(apiObject.Interval)
	}
	tfMap["interval_unit"] = apiObject.IntervalUnit
	tfMap[names.AttrLocation] = apiObject.Location
	if apiObject.Scripts != nil {
		tfMap["scripts"] = flattenScripts(apiObject.Scripts)
	}
	tfMap["times"] = apiObject.Times

	return []any{tfMap}
}

func expandRetainRule(tfList []any) *awstypes.RetainRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.RetainRule{}

	if v, ok := tfMap["count"].(int); ok && v > 0 {
		apiObject.Count = aws.Int32(int32(v))
	}
	if v, ok := tfMap[names.AttrInterval].(int); ok && v > 0 {
		apiObject.Interval = aws.Int32(int32(v))
	}
	if v, ok := tfMap["interval_unit"].(string); ok && v != "" {
		apiObject.IntervalUnit = awstypes.RetentionIntervalUnitValues(v)
	}

	return apiObject
}

func flattenRetainRule(apiObject *awstypes.RetainRule) []any {
	tfMap := make(map[string]any)
	tfMap["count"] = aws.ToInt32(apiObject.Count)
	tfMap[names.AttrInterval] = aws.ToInt32(apiObject.Interval)
	tfMap["interval_unit"] = apiObject.IntervalUnit

	return []any{tfMap}
}

func expandDeprecateRule(tfList []any) *awstypes.DeprecateRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.DeprecateRule{}

	if v, ok := tfMap["count"].(int); ok && v > 0 {
		apiObject.Count = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrInterval].(int); ok && v > 0 {
		apiObject.Interval = aws.Int32(int32(v))
	}

	if v, ok := tfMap["interval_unit"].(string); ok && v != "" {
		apiObject.IntervalUnit = awstypes.RetentionIntervalUnitValues(v)
	}

	return apiObject
}

func flattenDeprecateRule(apiObject *awstypes.DeprecateRule) []any {
	tfMap := make(map[string]any)
	tfMap["count"] = aws.ToInt32(apiObject.Count)
	tfMap[names.AttrInterval] = aws.ToInt32(apiObject.Interval)
	tfMap["interval_unit"] = apiObject.IntervalUnit

	return []any{tfMap}
}

func expandFastRestoreRule(tfList []any) *awstypes.FastRestoreRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.FastRestoreRule{
		AvailabilityZones: flex.ExpandStringValueSet(tfMap[names.AttrAvailabilityZones].(*schema.Set)),
	}

	if v, ok := tfMap["count"].(int); ok && v > 0 {
		apiObject.Count = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrInterval].(int); ok && v > 0 {
		apiObject.Interval = aws.Int32(int32(v))
	}

	if v, ok := tfMap["interval_unit"].(string); ok && v != "" {
		apiObject.IntervalUnit = awstypes.RetentionIntervalUnitValues(v)
	}

	return apiObject
}

func flattenFastRestoreRule(apiObject *awstypes.FastRestoreRule) []any {
	tfMap := make(map[string]any)
	tfMap[names.AttrAvailabilityZones] = apiObject.AvailabilityZones
	tfMap["count"] = aws.ToInt32(apiObject.Count)
	tfMap[names.AttrInterval] = aws.ToInt32(apiObject.Interval)
	tfMap["interval_unit"] = apiObject.IntervalUnit

	return []any{tfMap}
}

func expandShareRule(tfList []any) []awstypes.ShareRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.ShareRule, 0)

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.ShareRule{
			TargetAccounts: flex.ExpandStringValueSet(tfMap["target_accounts"].(*schema.Set)),
		}

		if v, ok := tfMap["unshare_interval"].(int); ok && v > 0 {
			apiObject.UnshareInterval = aws.Int32(int32(v))
		}

		if v, ok := tfMap["unshare_interval_unit"].(string); ok && v != "" {
			apiObject.UnshareIntervalUnit = awstypes.RetentionIntervalUnitValues(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenShareRule(apiObjects []awstypes.ShareRule) []any {
	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)
		if apiObject.TargetAccounts != nil {
			tfMap["target_accounts"] = apiObject.TargetAccounts
		}
		if apiObject.UnshareInterval != nil {
			tfMap["unshare_interval"] = aws.ToInt32(apiObject.UnshareInterval)
		}
		tfMap["unshare_interval_unit"] = apiObject.UnshareIntervalUnit

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandTags(tfMap map[string]any) []awstypes.Tag {
	var apiObjects []awstypes.Tag

	for k, v := range tfMap {
		apiObjects = append(apiObjects, awstypes.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return apiObjects
}

func flattenTags(apiObjects []awstypes.Tag) map[string]string {
	tfMap := make(map[string]string)

	for _, apiObject := range apiObjects {
		tfMap[aws.ToString(apiObject.Key)] = aws.ToString(apiObject.Value)
	}

	return tfMap
}

func expandParameters(tfList []any, policyType awstypes.PolicyTypeValues) *awstypes.Parameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.Parameters{}

	if v, ok := tfMap["exclude_boot_volume"].(bool); ok && policyType == awstypes.PolicyTypeValuesEbsSnapshotManagement {
		apiObject.ExcludeBootVolume = aws.Bool(v)
	}

	if v, ok := tfMap["no_reboot"].(bool); ok && policyType == awstypes.PolicyTypeValuesImageManagement {
		apiObject.NoReboot = aws.Bool(v)
	}

	return apiObject
}

func flattenParameters(apiObject *awstypes.Parameters) []any {
	tfMap := make(map[string]any)
	if apiObject.ExcludeBootVolume != nil {
		tfMap["exclude_boot_volume"] = aws.ToBool(apiObject.ExcludeBootVolume)
	}
	if apiObject.NoReboot != nil {
		tfMap["no_reboot"] = aws.ToBool(apiObject.NoReboot)
	}

	return []any{tfMap}
}

func expandScripts(tfList []any) []awstypes.Script {
	apiObjects := make([]awstypes.Script, len(tfList))

	for i, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.Script{}

		if v, ok := tfMap["execute_operation_on_script_failure"].(bool); ok {
			apiObject.ExecuteOperationOnScriptFailure = aws.Bool(v)
		}
		if v, ok := tfMap["execution_handler"].(string); ok {
			apiObject.ExecutionHandler = aws.String(v)
		}
		if v, ok := tfMap["execution_handler_service"].(string); ok && v != "" {
			apiObject.ExecutionHandlerService = awstypes.ExecutionHandlerServiceValues(v)
		}
		if v, ok := tfMap["execution_timeout"].(int); ok && v > 0 {
			apiObject.ExecutionTimeout = aws.Int32(int32(v))
		}
		if v, ok := tfMap["maximum_retry_count"].(int); ok && v > 0 {
			apiObject.MaximumRetryCount = aws.Int32(int32(v))
		}
		if v, ok := tfMap["stages"].([]any); ok && len(v) > 0 {
			apiObject.Stages = flex.ExpandStringyValueList[awstypes.StageValues](v)
		}

		apiObjects[i] = apiObject
	}

	return apiObjects
}

func flattenScripts(apiObjects []awstypes.Script) []any {
	tfList := make([]any, len(apiObjects))

	for i, apiObject := range apiObjects {
		tfMap := make(map[string]any)
		tfMap["execute_operation_on_script_failure"] = aws.ToBool(apiObject.ExecuteOperationOnScriptFailure)
		tfMap["execution_handler"] = aws.ToString(apiObject.ExecutionHandler)
		tfMap["execution_handler_service"] = apiObject.ExecutionHandlerService
		tfMap["execution_timeout"] = aws.ToInt32(apiObject.ExecutionTimeout)
		tfMap["maximum_retry_count"] = aws.ToInt32(apiObject.MaximumRetryCount)
		tfMap["stages"] = apiObject.Stages

		tfList[i] = tfMap
	}

	return tfList
}

func expandArchiveRule(tfList []any) *awstypes.ArchiveRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	return &awstypes.ArchiveRule{
		RetainRule: expandArchiveRetainRule(tfMap["archive_retain_rule"].([]any)),
	}
}

func flattenArchiveRule(apiObject *awstypes.ArchiveRule) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)
	tfMap["archive_retain_rule"] = flattenArchiveRetainRule(apiObject.RetainRule)

	return []any{tfMap}
}

func expandArchiveRetainRule(tfList []any) *awstypes.ArchiveRetainRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	return &awstypes.ArchiveRetainRule{
		RetentionArchiveTier: expandRetentionArchiveTier(tfMap["retention_archive_tier"].([]any)),
	}
}

func flattenArchiveRetainRule(apiObject *awstypes.ArchiveRetainRule) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)
	tfMap["retention_archive_tier"] = flattenRetentionArchiveTier(apiObject.RetentionArchiveTier)

	return []any{tfMap}
}

func expandRetentionArchiveTier(tfList []any) *awstypes.RetentionArchiveTier {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.RetentionArchiveTier{}

	if v, ok := tfMap["count"].(int); ok && v > 0 {
		apiObject.Count = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrInterval].(int); ok && v > 0 {
		apiObject.Interval = aws.Int32(int32(v))
	}

	if v, ok := tfMap["interval_unit"].(string); ok && v != "" {
		apiObject.IntervalUnit = awstypes.RetentionIntervalUnitValues(v)
	}

	return apiObject
}

func flattenRetentionArchiveTier(apiObject *awstypes.RetentionArchiveTier) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)
	tfMap["count"] = aws.ToInt32(apiObject.Count)
	tfMap[names.AttrInterval] = aws.ToInt32(apiObject.Interval)
	tfMap["interval_unit"] = apiObject.IntervalUnit

	return []any{tfMap}
}

func expandExclusions(tfList []any) *awstypes.Exclusions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.Exclusions{}

	if v, ok := tfMap["exclude_boot_volumes"].(bool); ok {
		apiObject.ExcludeBootVolumes = aws.Bool(v)
	}
	if v, ok := tfMap["exclude_tags"].(map[string]any); ok {
		apiObject.ExcludeTags = expandTags(v)
	}
	if v, ok := tfMap["exclude_volume_types"].([]any); ok && len(v) > 0 {
		apiObject.ExcludeVolumeTypes = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func flattenExclusions(apiObject *awstypes.Exclusions) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)
	tfMap["exclude_boot_volumes"] = aws.ToBool(apiObject.ExcludeBootVolumes)
	tfMap["exclude_tags"] = flattenTags(apiObject.ExcludeTags)
	tfMap["exclude_volume_types"] = apiObject.ExcludeVolumeTypes

	return []any{tfMap}
}
