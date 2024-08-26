// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_backup_plan", name="Plan")
// @Tags(identifierAttribute="arn")
func ResourcePlan() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlanCreate,
		ReadWithoutTimeout:   resourcePlanRead,
		UpdateWithoutTimeout: resourcePlanUpdate,
		DeleteWithoutTimeout: resourcePlanDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"advanced_backup_setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_options": {
							Type:     schema.TypeMap,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrResourceType: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"EC2",
							}, false),
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrRule: {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"completion_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  180,
						},
						"copy_action": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_vault_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"lifecycle": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cold_storage_after": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"delete_after": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"opt_in_to_archive_for_supported_resources": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"enable_continuous_backup": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"lifecycle": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cold_storage_after": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"delete_after": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"opt_in_to_archive_for_supported_resources": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"recovery_point_tags": tftags.TagsSchema(),
						"rule_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 50),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must contain only alphanumeric characters, hyphens, underscores, and periods"),
							),
						},
						names.AttrSchedule: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60,
						},
						"target_vault_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(2, 50),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric characters, hyphens, and underscores"),
							),
						},
					},
				},
				Set: planHash,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &backup.CreateBackupPlanInput{
		BackupPlan: &awstypes.BackupPlanInput{
			AdvancedBackupSettings: expandPlanAdvancedSettings(d.Get("advanced_backup_setting").(*schema.Set)),
			BackupPlanName:         aws.String(name),
			Rules:                  expandPlanRules(ctx, d.Get(names.AttrRule).(*schema.Set)),
		},
		BackupPlanTags: getTagsIn(ctx),
	}

	output, err := conn.CreateBackupPlan(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Plan (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.BackupPlanId))

	return append(diags, resourcePlanRead(ctx, d, meta)...)
}

func resourcePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	output, err := FindPlanByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Plan (%s): %s", d.Id(), err)
	}

	// AdvancedBackupSettings being read direct from output and not from under
	// output.BackupPlan is deliberate - the latter always contains nil.
	if err := d.Set("advanced_backup_setting", flattenPlanAdvancedSettings(output.AdvancedBackupSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_backup_setting: %s", err)
	}
	d.Set(names.AttrARN, output.BackupPlanArn)
	d.Set(names.AttrName, output.BackupPlan.BackupPlanName)
	if err := d.Set(names.AttrRule, flattenPlanRules(ctx, output.BackupPlan.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}
	d.Set(names.AttrVersion, output.VersionId)

	return diags
}

func resourcePlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	if d.HasChanges(names.AttrRule, "advanced_backup_setting") {
		input := &backup.UpdateBackupPlanInput{
			BackupPlanId: aws.String(d.Id()),
			BackupPlan: &awstypes.BackupPlanInput{
				AdvancedBackupSettings: expandPlanAdvancedSettings(d.Get("advanced_backup_setting").(*schema.Set)),
				BackupPlanName:         aws.String(d.Get(names.AttrName).(string)),
				Rules:                  expandPlanRules(ctx, d.Get(names.AttrRule).(*schema.Set)),
			},
		}

		_, err := conn.UpdateBackupPlan(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Backup Plan (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePlanRead(ctx, d, meta)...)
}

func resourcePlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	log.Printf("[DEBUG] Deleting Backup Plan: %s", d.Id())
	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidRequestException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteBackupPlan(ctx, &backup.DeleteBackupPlanInput{
			BackupPlanId: aws.String(d.Id()),
		})
	}, "Related backup plan selections must be deleted prior to backup")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Plan (%s): %s", d.Id(), err)
	}

	return diags
}

func FindPlanByID(ctx context.Context, conn *backup.Client, id string) (*backup.GetBackupPlanOutput, error) {
	input := &backup.GetBackupPlanInput{
		BackupPlanId: aws.String(id),
	}

	output, err := conn.GetBackupPlan(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.BackupPlan == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandPlanRules(ctx context.Context, vRules *schema.Set) []awstypes.BackupRuleInput {
	rules := []awstypes.BackupRuleInput{}

	for _, vRule := range vRules.List() {
		rule := awstypes.BackupRuleInput{}

		mRule := vRule.(map[string]interface{})

		if vRuleName, ok := mRule["rule_name"].(string); ok && vRuleName != "" {
			rule.RuleName = aws.String(vRuleName)
		} else {
			continue
		}
		if vTargetVaultName, ok := mRule["target_vault_name"].(string); ok && vTargetVaultName != "" {
			rule.TargetBackupVaultName = aws.String(vTargetVaultName)
		}
		if vSchedule, ok := mRule[names.AttrSchedule].(string); ok && vSchedule != "" {
			rule.ScheduleExpression = aws.String(vSchedule)
		}
		if vEnableContinuousBackup, ok := mRule["enable_continuous_backup"].(bool); ok {
			rule.EnableContinuousBackup = aws.Bool(vEnableContinuousBackup)
		}
		if vStartWindow, ok := mRule["start_window"].(int); ok {
			rule.StartWindowMinutes = aws.Int64(int64(vStartWindow))
		}
		if vCompletionWindow, ok := mRule["completion_window"].(int); ok {
			rule.CompletionWindowMinutes = aws.Int64(int64(vCompletionWindow))
		}

		if vRecoveryPointTags, ok := mRule["recovery_point_tags"].(map[string]interface{}); ok && len(vRecoveryPointTags) > 0 {
			rule.RecoveryPointTags = Tags(tftags.New(ctx, vRecoveryPointTags).IgnoreAWS())
		}

		if v, ok := mRule["lifecycle"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Lifecycle = expandPlanLifecycle(v[0].(map[string]interface{}))
		}

		if vCopyActions := expandPlanCopyActions(mRule["copy_action"].(*schema.Set).List()); len(vCopyActions) > 0 {
			rule.CopyActions = vCopyActions
		}

		rules = append(rules, rule)
	}

	return rules
}

func expandPlanAdvancedSettings(vAdvancedBackupSettings *schema.Set) []awstypes.AdvancedBackupSetting {
	advancedBackupSettings := []awstypes.AdvancedBackupSetting{}

	for _, vAdvancedBackupSetting := range vAdvancedBackupSettings.List() {
		advancedBackupSetting := awstypes.AdvancedBackupSetting{}

		mAdvancedBackupSetting := vAdvancedBackupSetting.(map[string]interface{})

		if v, ok := mAdvancedBackupSetting["backup_options"].(map[string]interface{}); ok && v != nil {
			advancedBackupSetting.BackupOptions = flex.ExpandStringValueMap(v)
		}
		if v, ok := mAdvancedBackupSetting[names.AttrResourceType].(string); ok && v != "" {
			advancedBackupSetting.ResourceType = aws.String(v)
		}

		// https://github.com/hashicorp/terraform-plugin-sdk/issues/588
		// Map in Set may add empty element. Ignore it.
		if advancedBackupSetting.ResourceType == nil {
			continue
		}

		advancedBackupSettings = append(advancedBackupSettings, advancedBackupSetting)
	}

	return advancedBackupSettings
}

func expandPlanCopyActions(actionList []interface{}) []awstypes.CopyAction {
	actions := []awstypes.CopyAction{}

	for _, i := range actionList {
		item := i.(map[string]interface{})
		action := awstypes.CopyAction{}

		action.DestinationBackupVaultArn = aws.String(item["destination_vault_arn"].(string))

		if v, ok := item["lifecycle"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			action.Lifecycle = expandPlanLifecycle(v[0].(map[string]interface{}))
		}

		actions = append(actions, action)
	}

	return actions
}

func expandPlanLifecycle(tfMap map[string]interface{}) *awstypes.Lifecycle {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Lifecycle{}

	if v, ok := tfMap["delete_after"].(int); ok && v != 0 {
		apiObject.DeleteAfterDays = aws.Int64(int64(v))
	}

	if v, ok := tfMap["cold_storage_after"].(int); ok && v != 0 {
		apiObject.MoveToColdStorageAfterDays = aws.Int64(int64(v))
	}

	if v, ok := tfMap["opt_in_to_archive_for_supported_resources"].(bool); ok && v {
		apiObject.OptInToArchiveForSupportedResources = aws.Bool(v)
	}

	return apiObject
}

func flattenPlanRules(ctx context.Context, rules []awstypes.BackupRule) *schema.Set {
	vRules := []interface{}{}

	for _, rule := range rules {
		mRule := map[string]interface{}{
			"rule_name":                aws.ToString(rule.RuleName),
			"target_vault_name":        aws.ToString(rule.TargetBackupVaultName),
			names.AttrSchedule:         aws.ToString(rule.ScheduleExpression),
			"enable_continuous_backup": aws.ToBool(rule.EnableContinuousBackup),
			"start_window":             int(aws.ToInt64(rule.StartWindowMinutes)),
			"completion_window":        int(aws.ToInt64(rule.CompletionWindowMinutes)),
			"recovery_point_tags":      KeyValueTags(ctx, rule.RecoveryPointTags).IgnoreAWS().Map(),
		}

		if lifecycle := rule.Lifecycle; lifecycle != nil {
			mRule["lifecycle"] = flattenPlanCopyActionLifecycle(lifecycle)
		}

		mRule["copy_action"] = flattenPlanCopyActions(rule.CopyActions)

		vRules = append(vRules, mRule)
	}

	return schema.NewSet(planHash, vRules)
}

func flattenPlanAdvancedSettings(advancedBackupSettings []awstypes.AdvancedBackupSetting) *schema.Set {
	vAdvancedBackupSettings := []interface{}{}

	for _, advancedBackupSetting := range advancedBackupSettings {
		mAdvancedBackupSetting := map[string]interface{}{
			"backup_options":       advancedBackupSetting.BackupOptions,
			names.AttrResourceType: aws.ToString(advancedBackupSetting.ResourceType),
		}

		vAdvancedBackupSettings = append(vAdvancedBackupSettings, mAdvancedBackupSetting)
	}

	return schema.NewSet(planHash, vAdvancedBackupSettings)
}

func flattenPlanCopyActions(copyActions []awstypes.CopyAction) []interface{} {
	if len(copyActions) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, copyAction := range copyActions {
		tfMap := map[string]interface{}{
			"destination_vault_arn": aws.ToString(copyAction.DestinationBackupVaultArn),
		}

		if copyAction.Lifecycle != nil {
			tfMap["lifecycle"] = flattenPlanCopyActionLifecycle(copyAction.Lifecycle)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPlanCopyActionLifecycle(copyActionLifecycle *awstypes.Lifecycle) []interface{} {
	if copyActionLifecycle == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"delete_after":       aws.ToInt64(copyActionLifecycle.DeleteAfterDays),
		"cold_storage_after": aws.ToInt64(copyActionLifecycle.MoveToColdStorageAfterDays),
		"opt_in_to_archive_for_supported_resources": aws.ToBool(copyActionLifecycle.OptInToArchiveForSupportedResources),
	}

	return []interface{}{m}
}

func planHash(vRule interface{}) int {
	var buf bytes.Buffer

	mRule := vRule.(map[string]interface{})

	if v, ok := mRule["rule_name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := mRule["target_vault_name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := mRule[names.AttrSchedule].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := mRule["enable_continuous_backup"].(bool); ok {
		buf.WriteString(fmt.Sprintf("%t-", v))
	}
	if v, ok := mRule["start_window"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}
	if v, ok := mRule["completion_window"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}

	if vRecoveryPointTags, ok := mRule["recovery_point_tags"].(map[string]interface{}); ok && len(vRecoveryPointTags) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", tftags.New(context.Background(), vRecoveryPointTags).Hash()))
	}

	if vLifecycle, ok := mRule["lifecycle"].([]interface{}); ok && len(vLifecycle) > 0 && vLifecycle[0] != nil {
		mLifecycle := vLifecycle[0].(map[string]interface{})

		if v, ok := mLifecycle["delete_after"].(int); ok {
			buf.WriteString(fmt.Sprintf("%d-", v))
		}
		if v, ok := mLifecycle["cold_storage_after"].(int); ok {
			buf.WriteString(fmt.Sprintf("%d-", v))
		}
	}

	if vCopyActions, ok := mRule["copy_action"].(*schema.Set); ok && vCopyActions.Len() > 0 {
		for _, a := range vCopyActions.List() {
			action := a.(map[string]interface{})
			if mLifecycle, ok := action["lifecycle"].([]interface{}); ok {
				for _, l := range mLifecycle {
					lifecycle := l.(map[string]interface{})
					if v, ok := lifecycle["delete_after"].(int); ok {
						buf.WriteString(fmt.Sprintf("%d-", v))
					}
					if v, ok := lifecycle["cold_storage_after"].(int); ok {
						buf.WriteString(fmt.Sprintf("%d-", v))
					}
				}
			}

			if v, ok := action["destination_vault_arn"].(string); ok {
				buf.WriteString(fmt.Sprintf("%s-", v))
			}
		}
	}

	return create.StringHashcode(buf.String())
}
