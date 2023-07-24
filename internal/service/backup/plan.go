// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rule": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 50),
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]+$`), "must contain only alphanumeric characters, hyphens, underscores, and periods"),
							),
						},
						"target_vault_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(2, 50),
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_]+$`), "must contain only alphanumeric characters, hyphens, and underscores"),
							),
						},
						"schedule": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enable_continuous_backup": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"start_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60,
						},
						"completion_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  180,
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
								},
							},
						},
						"copy_action": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
											},
										},
									},
									"destination_vault_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"recovery_point_tags": tftags.TagsSchema(),
					},
				},
				Set: planHash,
			},
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
						"resource_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"EC2",
							}, false),
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	input := &backup.CreateBackupPlanInput{
		BackupPlan: &backup.PlanInput{
			BackupPlanName:         aws.String(d.Get("name").(string)),
			Rules:                  expandPlanRules(ctx, d.Get("rule").(*schema.Set)),
			AdvancedBackupSettings: expandPlanAdvancedSettings(d.Get("advanced_backup_setting").(*schema.Set)),
		},
		BackupPlanTags: getTagsIn(ctx),
	}

	resp, err := conn.CreateBackupPlanWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Plan: %s", err)
	}

	d.SetId(aws.StringValue(resp.BackupPlanId))

	return append(diags, resourcePlanRead(ctx, d, meta)...)
}

func resourcePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	resp, err := conn.GetBackupPlanWithContext(ctx, &backup.GetBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Backup Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Plan (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.BackupPlanArn)
	d.Set("name", resp.BackupPlan.BackupPlanName)
	d.Set("version", resp.VersionId)

	if err := d.Set("rule", flattenPlanRules(ctx, resp.BackupPlan.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	// AdvancedBackupSettings being read direct from resp and not from under
	// resp.BackupPlan is deliberate - the latter always contains null
	if err := d.Set("advanced_backup_setting", flattenPlanAdvancedSettings(resp.AdvancedBackupSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_backup_setting: %s", err)
	}

	return diags
}

func resourcePlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	if d.HasChanges("rule", "advanced_backup_setting") {
		input := &backup.UpdateBackupPlanInput{
			BackupPlanId: aws.String(d.Id()),
			BackupPlan: &backup.PlanInput{
				BackupPlanName:         aws.String(d.Get("name").(string)),
				Rules:                  expandPlanRules(ctx, d.Get("rule").(*schema.Set)),
				AdvancedBackupSettings: expandPlanAdvancedSettings(d.Get("advanced_backup_setting").(*schema.Set)),
			},
		}

		log.Printf("[DEBUG] Updating Backup Plan: %#v", input)
		_, err := conn.UpdateBackupPlanWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Backup Plan (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePlanRead(ctx, d, meta)...)
}

func resourcePlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	input := &backup.DeleteBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Backup Plan: %s", d.Id())
	err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteBackupPlanWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidRequestException, "Related backup plan selections must be deleted prior to backup") {
			return retry.RetryableError(err)
		}

		if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
			return nil
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteBackupPlanWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Plan (%s): %s", d.Id(), err)
	}

	return diags
}

func expandPlanRules(ctx context.Context, vRules *schema.Set) []*backup.RuleInput {
	rules := []*backup.RuleInput{}

	for _, vRule := range vRules.List() {
		rule := &backup.RuleInput{}

		mRule := vRule.(map[string]interface{})

		if vRuleName, ok := mRule["rule_name"].(string); ok && vRuleName != "" {
			rule.RuleName = aws.String(vRuleName)
		} else {
			continue
		}
		if vTargetVaultName, ok := mRule["target_vault_name"].(string); ok && vTargetVaultName != "" {
			rule.TargetBackupVaultName = aws.String(vTargetVaultName)
		}
		if vSchedule, ok := mRule["schedule"].(string); ok && vSchedule != "" {
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

		if vLifecycle, ok := mRule["lifecycle"].([]interface{}); ok && len(vLifecycle) > 0 {
			rule.Lifecycle = expandPlanLifecycle(vLifecycle)
		}

		if vCopyActions := expandPlanCopyActions(mRule["copy_action"].(*schema.Set).List()); len(vCopyActions) > 0 {
			rule.CopyActions = vCopyActions
		}

		rules = append(rules, rule)
	}

	return rules
}

func expandPlanAdvancedSettings(vAdvancedBackupSettings *schema.Set) []*backup.AdvancedBackupSetting {
	advancedBackupSettings := []*backup.AdvancedBackupSetting{}

	for _, vAdvancedBackupSetting := range vAdvancedBackupSettings.List() {
		advancedBackupSetting := &backup.AdvancedBackupSetting{}

		mAdvancedBackupSetting := vAdvancedBackupSetting.(map[string]interface{})

		if v, ok := mAdvancedBackupSetting["backup_options"].(map[string]interface{}); ok && v != nil {
			advancedBackupSetting.BackupOptions = flex.ExpandStringMap(v)
		}
		if v, ok := mAdvancedBackupSetting["resource_type"].(string); ok && v != "" {
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

func expandPlanCopyActions(actionList []interface{}) []*backup.CopyAction {
	actions := []*backup.CopyAction{}

	for _, i := range actionList {
		item := i.(map[string]interface{})
		action := &backup.CopyAction{}

		action.DestinationBackupVaultArn = aws.String(item["destination_vault_arn"].(string))

		if v, ok := item["lifecycle"].([]interface{}); ok && len(v) > 0 {
			action.Lifecycle = expandPlanLifecycle(v)
		}

		actions = append(actions, action)
	}

	return actions
}

func expandPlanLifecycle(l []interface{}) *backup.Lifecycle {
	lifecycle := new(backup.Lifecycle)

	for _, i := range l {
		lc := i.(map[string]interface{})
		if vDeleteAfter, ok := lc["delete_after"]; ok && vDeleteAfter.(int) > 0 {
			lifecycle.DeleteAfterDays = aws.Int64(int64(vDeleteAfter.(int)))
		}
		if vMoveToColdStorageAfterDays, ok := lc["cold_storage_after"]; ok && vMoveToColdStorageAfterDays.(int) > 0 {
			lifecycle.MoveToColdStorageAfterDays = aws.Int64(int64(vMoveToColdStorageAfterDays.(int)))
		}
	}

	return lifecycle
}

func flattenPlanRules(ctx context.Context, rules []*backup.Rule) *schema.Set {
	vRules := []interface{}{}

	for _, rule := range rules {
		mRule := map[string]interface{}{
			"rule_name":                aws.StringValue(rule.RuleName),
			"target_vault_name":        aws.StringValue(rule.TargetBackupVaultName),
			"schedule":                 aws.StringValue(rule.ScheduleExpression),
			"enable_continuous_backup": aws.BoolValue(rule.EnableContinuousBackup),
			"start_window":             int(aws.Int64Value(rule.StartWindowMinutes)),
			"completion_window":        int(aws.Int64Value(rule.CompletionWindowMinutes)),
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

func flattenPlanAdvancedSettings(advancedBackupSettings []*backup.AdvancedBackupSetting) *schema.Set {
	vAdvancedBackupSettings := []interface{}{}

	for _, advancedBackupSetting := range advancedBackupSettings {
		mAdvancedBackupSetting := map[string]interface{}{
			"backup_options": aws.StringValueMap(advancedBackupSetting.BackupOptions),
			"resource_type":  aws.StringValue(advancedBackupSetting.ResourceType),
		}

		vAdvancedBackupSettings = append(vAdvancedBackupSettings, mAdvancedBackupSetting)
	}

	return schema.NewSet(planHash, vAdvancedBackupSettings)
}

func flattenPlanCopyActions(copyActions []*backup.CopyAction) []interface{} {
	if len(copyActions) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, copyAction := range copyActions {
		if copyAction == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"destination_vault_arn": aws.StringValue(copyAction.DestinationBackupVaultArn),
		}

		if copyAction.Lifecycle != nil {
			tfMap["lifecycle"] = flattenPlanCopyActionLifecycle(copyAction.Lifecycle)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenPlanCopyActionLifecycle(copyActionLifecycle *backup.Lifecycle) []interface{} {
	if copyActionLifecycle == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"delete_after":       aws.Int64Value(copyActionLifecycle.DeleteAfterDays),
		"cold_storage_after": aws.Int64Value(copyActionLifecycle.MoveToColdStorageAfterDays),
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
	if v, ok := mRule["schedule"].(string); ok {
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
