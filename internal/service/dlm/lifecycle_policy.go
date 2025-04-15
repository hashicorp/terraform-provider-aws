// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
						"resource_types": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.ResourceTypeValues](),
							},
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
						"policy_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.PolicyTypeValuesEbsSnapshotManagement,
							ValidateDiagFunc: enum.Validate[awstypes.PolicyTypeValues](),
						},
						names.AttrSchedule: {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
													Required:     true,
													ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[\w:\-\/\*]+$`), ""),
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
	const createRetryTimeout = 2 * time.Minute
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DLMClient(ctx)

	input := dlm.CreateLifecyclePolicyInput{
		Description:      aws.String(d.Get(names.AttrDescription).(string)),
		ExecutionRoleArn: aws.String(d.Get(names.AttrExecutionRoleARN).(string)),
		PolicyDetails:    expandPolicyDetails(d.Get("policy_details").([]any)),
		State:            awstypes.SettablePolicyStateValues(d.Get(names.AttrState).(string)),
		Tags:             getTagsIn(ctx),
	}

	out, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, createRetryTimeout, func() (any, error) {
		return conn.CreateLifecyclePolicy(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DLM Lifecycle Policy: %s", err)
	}

	d.SetId(aws.ToString(out.(*dlm.CreateLifecyclePolicyOutput).PolicyId))

	return append(diags, resourceLifecyclePolicyRead(ctx, d, meta)...)
}

func resourceLifecyclePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DLMClient(ctx)

	log.Printf("[INFO] Reading DLM lifecycle policy: %s", d.Id())
	out, err := findLifecyclePolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DLM Lifecycle Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DLM Lifecycle Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.Policy.PolicyArn)
	d.Set(names.AttrDescription, out.Policy.Description)
	d.Set(names.AttrExecutionRoleARN, out.Policy.ExecutionRoleArn)
	d.Set(names.AttrState, out.Policy.State)
	if err := d.Set("policy_details", flattenPolicyDetails(out.Policy.PolicyDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting policy details %s", err)
	}

	setTagsOut(ctx, out.Policy.Tags)

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
		if d.HasChange(names.AttrState) {
			input.State = awstypes.SettablePolicyStateValues(d.Get(names.AttrState).(string))
		}
		if d.HasChange("policy_details") {
			input.PolicyDetails = expandPolicyDetails(d.Get("policy_details").([]any))
		}

		log.Printf("[INFO] Updating lifecycle policy %s", d.Id())
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
	input := &dlm.GetLifecyclePolicyInput{
		PolicyId: aws.String(id),
	}

	output, err := conn.GetLifecyclePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastRequest: input,
			LastError:   err,
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

func expandPolicyDetails(cfg []any) *awstypes.PolicyDetails {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]any)
	policyType := m["policy_type"].(string)

	policyDetails := &awstypes.PolicyDetails{
		PolicyType: awstypes.PolicyTypeValues(policyType),
	}
	if v, ok := m["resource_types"].([]any); ok && len(v) > 0 {
		policyDetails.ResourceTypes = flex.ExpandStringyValueList[awstypes.ResourceTypeValues](v)
	}
	if v, ok := m["resource_locations"].([]any); ok && len(v) > 0 {
		policyDetails.ResourceLocations = flex.ExpandStringyValueList[awstypes.ResourceLocationValues](v)
	}
	if v, ok := m[names.AttrSchedule].([]any); ok && len(v) > 0 {
		policyDetails.Schedules = expandSchedules(v)
	}
	if v, ok := m[names.AttrAction].([]any); ok && len(v) > 0 {
		policyDetails.Actions = expandActions(v)
	}
	if v, ok := m["event_source"].([]any); ok && len(v) > 0 {
		policyDetails.EventSource = expandEventSource(v)
	}
	if v, ok := m["target_tags"].(map[string]any); ok && len(v) > 0 {
		policyDetails.TargetTags = expandTags(v)
	}
	if v, ok := m[names.AttrParameters].([]any); ok && len(v) > 0 {
		policyDetails.Parameters = expandParameters(v, policyType)
	}

	return policyDetails
}

func flattenPolicyDetails(policyDetails *awstypes.PolicyDetails) []map[string]any {
	result := make(map[string]any)
	result["resource_types"] = flex.FlattenStringyValueList(policyDetails.ResourceTypes)
	result["resource_locations"] = flex.FlattenStringyValueList(policyDetails.ResourceLocations)
	result[names.AttrAction] = flattenActions(policyDetails.Actions)
	result["event_source"] = flattenEventSource(policyDetails.EventSource)
	result[names.AttrSchedule] = flattenSchedules(policyDetails.Schedules)
	result["target_tags"] = flattenTags(policyDetails.TargetTags)
	result["policy_type"] = string(policyDetails.PolicyType)

	if policyDetails.Parameters != nil {
		result[names.AttrParameters] = flattenParameters(policyDetails.Parameters)
	}

	return []map[string]any{result}
}

func expandSchedules(cfg []any) []awstypes.Schedule {
	schedules := make([]awstypes.Schedule, len(cfg))
	for i, c := range cfg {
		schedule := awstypes.Schedule{}
		m := c.(map[string]any)
		if v, ok := m["copy_tags"]; ok {
			schedule.CopyTags = aws.Bool(v.(bool))
		}
		if v, ok := m["create_rule"]; ok {
			schedule.CreateRule = expandCreateRule(v.([]any))
		}
		if v, ok := m["cross_region_copy_rule"].(*schema.Set); ok && v.Len() > 0 {
			schedule.CrossRegionCopyRules = expandCrossRegionCopyRules(v.List())
		}
		if v, ok := m[names.AttrName]; ok {
			schedule.Name = aws.String(v.(string))
		}
		if v, ok := m["deprecate_rule"]; ok {
			schedule.DeprecateRule = expandDeprecateRule(v.([]any))
		}
		if v, ok := m["fast_restore_rule"]; ok {
			schedule.FastRestoreRule = expandFastRestoreRule(v.([]any))
		}
		if v, ok := m["share_rule"]; ok {
			schedule.ShareRules = expandShareRule(v.([]any))
		}
		if v, ok := m["retain_rule"]; ok {
			schedule.RetainRule = expandRetainRule(v.([]any))
		}
		if v, ok := m["tags_to_add"]; ok {
			schedule.TagsToAdd = expandTags(v.(map[string]any))
		}
		if v, ok := m["variable_tags"]; ok {
			schedule.VariableTags = expandTags(v.(map[string]any))
		}

		schedules[i] = schedule
	}

	return schedules
}

func flattenSchedules(schedules []awstypes.Schedule) []map[string]any {
	result := make([]map[string]any, len(schedules))
	for i, s := range schedules {
		m := make(map[string]any)
		m["copy_tags"] = aws.ToBool(s.CopyTags)
		m["create_rule"] = flattenCreateRule(s.CreateRule)
		m["cross_region_copy_rule"] = flattenCrossRegionCopyRules(s.CrossRegionCopyRules)
		m[names.AttrName] = aws.ToString(s.Name)
		m["retain_rule"] = flattenRetainRule(s.RetainRule)
		m["tags_to_add"] = flattenTags(s.TagsToAdd)
		m["variable_tags"] = flattenTags(s.VariableTags)

		if s.DeprecateRule != nil {
			m["deprecate_rule"] = flattenDeprecateRule(s.DeprecateRule)
		}

		if s.FastRestoreRule != nil {
			m["fast_restore_rule"] = flattenFastRestoreRule(s.FastRestoreRule)
		}

		if s.ShareRules != nil {
			m["share_rule"] = flattenShareRule(s.ShareRules)
		}

		result[i] = m
	}

	return result
}

func expandActions(cfg []any) []awstypes.Action {
	actions := make([]awstypes.Action, len(cfg))
	for i, c := range cfg {
		action := awstypes.Action{}
		m := c.(map[string]any)
		if v, ok := m["cross_region_copy"].(*schema.Set); ok {
			action.CrossRegionCopy = expandActionCrossRegionCopyRules(v.List())
		}
		if v, ok := m[names.AttrName]; ok {
			action.Name = aws.String(v.(string))
		}

		actions[i] = action
	}

	return actions
}

func flattenActions(actions []awstypes.Action) []map[string]any {
	result := make([]map[string]any, len(actions))
	for i, s := range actions {
		m := make(map[string]any)

		m[names.AttrName] = aws.ToString(s.Name)

		if s.CrossRegionCopy != nil {
			m["cross_region_copy"] = flattenActionCrossRegionCopyRules(s.CrossRegionCopy)
		}

		result[i] = m
	}

	return result
}

func expandActionCrossRegionCopyRules(l []any) []awstypes.CrossRegionCopyAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var rules []awstypes.CrossRegionCopyAction

	for _, tfMapRaw := range l {
		m, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		rule := awstypes.CrossRegionCopyAction{}
		if v, ok := m[names.AttrEncryptionConfiguration].([]any); ok {
			rule.EncryptionConfiguration = expandActionCrossRegionCopyRuleEncryptionConfiguration(v)
		}
		if v, ok := m["retain_rule"].([]any); ok && len(v) > 0 && v[0] != nil {
			rule.RetainRule = expandCrossRegionCopyRuleRetainRule(v)
		}
		if v, ok := m[names.AttrTarget].(string); ok && v != "" {
			rule.Target = aws.String(v)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenActionCrossRegionCopyRules(rules []awstypes.CrossRegionCopyAction) []any {
	if len(rules) == 0 {
		return []any{}
	}

	var result []any

	for _, rule := range rules {
		m := map[string]any{
			names.AttrEncryptionConfiguration: flattenActionCrossRegionCopyRuleEncryptionConfiguration(rule.EncryptionConfiguration),
			"retain_rule":                     flattenCrossRegionCopyRuleRetainRule(rule.RetainRule),
			names.AttrTarget:                  aws.ToString(rule.Target),
		}

		result = append(result, m)
	}

	return result
}

func expandActionCrossRegionCopyRuleEncryptionConfiguration(l []any) *awstypes.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	config := &awstypes.EncryptionConfiguration{
		Encrypted: aws.Bool(m[names.AttrEncrypted].(bool)),
	}

	if v, ok := m["cmk_arn"].(string); ok && v != "" {
		config.CmkArn = aws.String(v)
	}
	return config
}

func flattenActionCrossRegionCopyRuleEncryptionConfiguration(rule *awstypes.EncryptionConfiguration) []any {
	if rule == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEncrypted: aws.ToBool(rule.Encrypted),
		"cmk_arn":           aws.ToString(rule.CmkArn),
	}

	return []any{m}
}

func expandEventSource(l []any) *awstypes.EventSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	config := &awstypes.EventSource{
		Type: awstypes.EventSourceValues(m[names.AttrType].(string)),
	}

	if v, ok := m[names.AttrParameters].([]any); ok && len(v) > 0 {
		config.Parameters = expandEventSourceParameters(v)
	}

	return config
}

func flattenEventSource(rule *awstypes.EventSource) []any {
	if rule == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrParameters: flattenEventSourceParameters(rule.Parameters),
		names.AttrType:       string(rule.Type),
	}

	return []any{m}
}

func expandEventSourceParameters(l []any) *awstypes.EventParameters {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	config := &awstypes.EventParameters{
		DescriptionRegex: aws.String(m["description_regex"].(string)),
		EventType:        awstypes.EventTypeValues(m["event_type"].(string)),
		SnapshotOwner:    flex.ExpandStringValueSet(m["snapshot_owner"].(*schema.Set)),
	}

	return config
}

func flattenEventSourceParameters(rule *awstypes.EventParameters) []any {
	if rule == nil {
		return []any{}
	}

	m := map[string]any{
		"description_regex": aws.ToString(rule.DescriptionRegex),
		"event_type":        string(rule.EventType),
		"snapshot_owner":    flex.FlattenStringValueSet(rule.SnapshotOwner),
	}

	return []any{m}
}

func expandCrossRegionCopyRules(l []any) []awstypes.CrossRegionCopyRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var rules []awstypes.CrossRegionCopyRule

	for _, tfMapRaw := range l {
		m, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		rule := awstypes.CrossRegionCopyRule{}

		if v, ok := m["cmk_arn"].(string); ok && v != "" {
			rule.CmkArn = aws.String(v)
		}
		if v, ok := m["copy_tags"].(bool); ok {
			rule.CopyTags = aws.Bool(v)
		}
		if v, ok := m["deprecate_rule"].([]any); ok && len(v) > 0 && v[0] != nil {
			rule.DeprecateRule = expandCrossRegionCopyRuleDeprecateRule(v)
		}
		if v, ok := m[names.AttrEncrypted].(bool); ok {
			rule.Encrypted = aws.Bool(v)
		}
		if v, ok := m["retain_rule"].([]any); ok && len(v) > 0 && v[0] != nil {
			rule.RetainRule = expandCrossRegionCopyRuleRetainRule(v)
		}
		if v, ok := m[names.AttrTarget].(string); ok && v != "" {
			rule.Target = aws.String(v)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenCrossRegionCopyRules(rules []awstypes.CrossRegionCopyRule) []any {
	if len(rules) == 0 {
		return []any{}
	}

	var result []any

	for _, rule := range rules {
		m := map[string]any{
			"cmk_arn":           aws.ToString(rule.CmkArn),
			"copy_tags":         aws.ToBool(rule.CopyTags),
			"deprecate_rule":    flattenCrossRegionCopyRuleDeprecateRule(rule.DeprecateRule),
			names.AttrEncrypted: aws.ToBool(rule.Encrypted),
			"retain_rule":       flattenCrossRegionCopyRuleRetainRule(rule.RetainRule),
			names.AttrTarget:    aws.ToString(rule.Target),
		}

		result = append(result, m)
	}

	return result
}

func expandCrossRegionCopyRuleDeprecateRule(l []any) *awstypes.CrossRegionCopyDeprecateRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.CrossRegionCopyDeprecateRule{
		Interval:     aws.Int32(int32(m[names.AttrInterval].(int))),
		IntervalUnit: awstypes.RetentionIntervalUnitValues(m["interval_unit"].(string)),
	}
}

func expandCrossRegionCopyRuleRetainRule(l []any) *awstypes.CrossRegionCopyRetainRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.CrossRegionCopyRetainRule{
		Interval:     aws.Int32(int32(m[names.AttrInterval].(int))),
		IntervalUnit: awstypes.RetentionIntervalUnitValues(m["interval_unit"].(string)),
	}
}

func flattenCrossRegionCopyRuleDeprecateRule(rule *awstypes.CrossRegionCopyDeprecateRule) []any {
	if rule == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrInterval: int(aws.ToInt32(rule.Interval)),
		"interval_unit":    string(rule.IntervalUnit),
	}

	return []any{m}
}

func flattenCrossRegionCopyRuleRetainRule(rule *awstypes.CrossRegionCopyRetainRule) []any {
	if rule == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrInterval: int(aws.ToInt32(rule.Interval)),
		"interval_unit":    string(rule.IntervalUnit),
	}

	return []any{m}
}

func expandCreateRule(cfg []any) *awstypes.CreateRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	c := cfg[0].(map[string]any)
	createRule := &awstypes.CreateRule{}

	if v, ok := c["times"].([]any); ok && len(v) > 0 {
		createRule.Times = flex.ExpandStringValueList(v)
	}

	if v, ok := c[names.AttrInterval].(int); ok && v > 0 {
		createRule.Interval = aws.Int32(int32(v))
	}

	if v, ok := c[names.AttrLocation].(string); ok && v != "" {
		createRule.Location = awstypes.LocationValues(v)
	}

	if v, ok := c["interval_unit"].(string); ok && v != "" {
		createRule.IntervalUnit = awstypes.IntervalUnitValues(v)
	} else {
		createRule.IntervalUnit = awstypes.IntervalUnitValuesHours
	}

	if v, ok := c["cron_expression"].(string); ok && v != "" {
		createRule.CronExpression = aws.String(v)
		createRule.IntervalUnit = "" // sets interval unit to empty string so that all fields related to interval are ignored
	}

	return createRule
}

func flattenCreateRule(createRule *awstypes.CreateRule) []map[string]any {
	if createRule == nil {
		return []map[string]any{}
	}

	result := make(map[string]any)
	result["times"] = flex.FlattenStringValueList(createRule.Times)

	if createRule.Interval != nil {
		result[names.AttrInterval] = aws.ToInt32(createRule.Interval)
	}

	result["interval_unit"] = string(createRule.IntervalUnit)

	result[names.AttrLocation] = string(createRule.Location)

	if createRule.CronExpression != nil {
		result["cron_expression"] = aws.ToString(createRule.CronExpression)
	}

	return []map[string]any{result}
}

func expandRetainRule(cfg []any) *awstypes.RetainRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]any)
	rule := &awstypes.RetainRule{}

	if v, ok := m["count"].(int); ok && v > 0 {
		rule.Count = aws.Int32(int32(v))
	}

	if v, ok := m[names.AttrInterval].(int); ok && v > 0 {
		rule.Interval = aws.Int32(int32(v))
	}

	if v, ok := m["interval_unit"].(string); ok && v != "" {
		rule.IntervalUnit = awstypes.RetentionIntervalUnitValues(v)
	}

	return rule
}

func flattenRetainRule(retainRule *awstypes.RetainRule) []map[string]any {
	result := make(map[string]any)
	result["count"] = aws.ToInt32(retainRule.Count)
	result["interval_unit"] = string(retainRule.IntervalUnit)
	result[names.AttrInterval] = aws.ToInt32(retainRule.Interval)

	return []map[string]any{result}
}

func expandDeprecateRule(cfg []any) *awstypes.DeprecateRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]any)
	rule := &awstypes.DeprecateRule{}

	if v, ok := m["count"].(int); ok && v > 0 {
		rule.Count = aws.Int32(int32(v))
	}

	if v, ok := m[names.AttrInterval].(int); ok && v > 0 {
		rule.Interval = aws.Int32(int32(v))
	}

	if v, ok := m["interval_unit"].(string); ok && v != "" {
		rule.IntervalUnit = awstypes.RetentionIntervalUnitValues(v)
	}

	return rule
}

func flattenDeprecateRule(rule *awstypes.DeprecateRule) []map[string]any {
	result := make(map[string]any)
	result["count"] = aws.ToInt32(rule.Count)
	result["interval_unit"] = string(rule.IntervalUnit)
	result[names.AttrInterval] = aws.ToInt32(rule.Interval)

	return []map[string]any{result}
}

func expandFastRestoreRule(cfg []any) *awstypes.FastRestoreRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]any)
	rule := &awstypes.FastRestoreRule{
		AvailabilityZones: flex.ExpandStringValueSet(m[names.AttrAvailabilityZones].(*schema.Set)),
	}

	if v, ok := m["count"].(int); ok && v > 0 {
		rule.Count = aws.Int32(int32(v))
	}

	if v, ok := m[names.AttrInterval].(int); ok && v > 0 {
		rule.Interval = aws.Int32(int32(v))
	}

	if v, ok := m["interval_unit"].(string); ok && v != "" {
		rule.IntervalUnit = awstypes.RetentionIntervalUnitValues(v)
	}

	return rule
}

func flattenFastRestoreRule(rule *awstypes.FastRestoreRule) []map[string]any {
	result := make(map[string]any)
	result["count"] = aws.ToInt32(rule.Count)
	result["interval_unit"] = string(rule.IntervalUnit)
	result[names.AttrInterval] = aws.ToInt32(rule.Interval)
	result[names.AttrAvailabilityZones] = flex.FlattenStringValueSet(rule.AvailabilityZones)

	return []map[string]any{result}
}

func expandShareRule(cfg []any) []awstypes.ShareRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}

	rules := make([]awstypes.ShareRule, 0)

	for _, shareRule := range cfg {
		m := shareRule.(map[string]any)

		rule := awstypes.ShareRule{
			TargetAccounts: flex.ExpandStringValueSet(m["target_accounts"].(*schema.Set)),
		}

		if v, ok := m["unshare_interval"].(int); ok && v > 0 {
			rule.UnshareInterval = aws.Int32(int32(v))
		}

		if v, ok := m["unshare_interval_unit"].(string); ok && v != "" {
			rule.UnshareIntervalUnit = awstypes.RetentionIntervalUnitValues(v)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenShareRule(rules []awstypes.ShareRule) []map[string]any {
	values := make([]map[string]any, 0)

	for _, v := range rules {
		rule := make(map[string]any)

		if v.TargetAccounts != nil {
			rule["target_accounts"] = flex.FlattenStringValueSet(v.TargetAccounts)
		}

		rule["unshare_interval_unit"] = string(v.UnshareIntervalUnit)

		if v.UnshareInterval != nil {
			rule["unshare_interval"] = aws.ToInt32(v.UnshareInterval)
		}

		values = append(values, rule)
	}

	return values
}

func expandTags(m map[string]any) []awstypes.Tag {
	var result []awstypes.Tag
	for k, v := range m {
		result = append(result, awstypes.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return result
}

func flattenTags(tags []awstypes.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range tags {
		result[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}

	return result
}

func expandParameters(cfg []any, policyType string) *awstypes.Parameters {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]any)
	parameters := &awstypes.Parameters{}

	if v, ok := m["exclude_boot_volume"].(bool); ok && policyType == string(awstypes.PolicyTypeValuesEbsSnapshotManagement) {
		parameters.ExcludeBootVolume = aws.Bool(v)
	}

	if v, ok := m["no_reboot"].(bool); ok && policyType == string(awstypes.PolicyTypeValuesImageManagement) {
		parameters.NoReboot = aws.Bool(v)
	}

	return parameters
}

func flattenParameters(parameters *awstypes.Parameters) []map[string]any {
	result := make(map[string]any)
	if parameters.ExcludeBootVolume != nil {
		result["exclude_boot_volume"] = aws.ToBool(parameters.ExcludeBootVolume)
	}

	if parameters.NoReboot != nil {
		result["no_reboot"] = aws.ToBool(parameters.NoReboot)
	}

	return []map[string]any{result}
}
