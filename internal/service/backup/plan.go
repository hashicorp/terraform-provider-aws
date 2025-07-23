// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	defaultPlanRuleScheduleExpressionTimezone = "Etc/UTC"
)

// @SDKResource("aws_backup_plan", name="Plan")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/backup;backup.GetBackupPlanOutput")
func resourcePlan() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlanCreate,
		ReadWithoutTimeout:   resourcePlanRead,
		UpdateWithoutTimeout: resourcePlanUpdate,
		DeleteWithoutTimeout: resourcePlanDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    planResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: planStateUpgradeV0,
				Version: 0,
			},
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
						"schedule_expression_timezone": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  defaultPlanRuleScheduleExpressionTimezone,
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
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePlanCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &backup.CreateBackupPlanInput{
		BackupPlan: &awstypes.BackupPlanInput{
			AdvancedBackupSettings: expandAdvancedBackupSetting(d.Get("advanced_backup_setting").(*schema.Set).List()),
			BackupPlanName:         aws.String(name),
			Rules:                  expandBackupRuleInputs(ctx, d.Get(names.AttrRule).(*schema.Set).List()),
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

func resourcePlanRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	output, err := findPlanByID(ctx, conn, d.Id())

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
	if err := d.Set("advanced_backup_setting", flattenAdvancedBackupSettings(output.AdvancedBackupSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_backup_setting: %s", err)
	}
	d.Set(names.AttrARN, output.BackupPlanArn)
	d.Set(names.AttrName, output.BackupPlan.BackupPlanName)
	if err := d.Set(names.AttrRule, flattenBackupRules(ctx, output.BackupPlan.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}
	d.Set(names.AttrVersion, output.VersionId)

	return diags
}

func resourcePlanUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	if d.HasChanges("advanced_backup_setting", names.AttrRule) {
		input := &backup.UpdateBackupPlanInput{
			BackupPlan: &awstypes.BackupPlanInput{
				AdvancedBackupSettings: expandAdvancedBackupSetting(d.Get("advanced_backup_setting").(*schema.Set).List()),
				BackupPlanName:         aws.String(d.Get(names.AttrName).(string)),
				Rules:                  expandBackupRuleInputs(ctx, d.Get(names.AttrRule).(*schema.Set).List()),
			},
			BackupPlanId: aws.String(d.Id()),
		}

		_, err := conn.UpdateBackupPlan(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Backup Plan (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePlanRead(ctx, d, meta)...)
}

func resourcePlanDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	log.Printf("[DEBUG] Deleting Backup Plan: %s", d.Id())
	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidRequestException](ctx, timeout, func() (any, error) {
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

func findPlanByID(ctx context.Context, conn *backup.Client, id string) (*backup.GetBackupPlanOutput, error) {
	input := &backup.GetBackupPlanInput{
		BackupPlanId: aws.String(id),
	}

	return findPlan(ctx, conn, input)
}

func findPlan(ctx context.Context, conn *backup.Client, input *backup.GetBackupPlanInput) (*backup.GetBackupPlanOutput, error) {
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

func expandBackupRuleInputs(ctx context.Context, tfList []any) []awstypes.BackupRuleInput { // nosemgrep:ci.backup-in-func-name
	apiObjects := []awstypes.BackupRuleInput{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.BackupRuleInput{}

		if v, ok := tfMap["completion_window"].(int); ok {
			apiObject.CompletionWindowMinutes = aws.Int64(int64(v))
		}
		if v := expandCopyActions(tfMap["copy_action"].(*schema.Set).List()); len(v) > 0 {
			apiObject.CopyActions = v
		}
		if v, ok := tfMap["enable_continuous_backup"].(bool); ok {
			apiObject.EnableContinuousBackup = aws.Bool(v)
		}
		if v, ok := tfMap["lifecycle"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.Lifecycle = expandLifecycle(v[0].(map[string]any))
		}
		if v, ok := tfMap["recovery_point_tags"].(map[string]any); ok && len(v) > 0 {
			apiObject.RecoveryPointTags = svcTags(tftags.New(ctx, v).IgnoreAWS())
		}
		if v, ok := tfMap["rule_name"].(string); ok && v != "" {
			apiObject.RuleName = aws.String(v)
		} else {
			continue
		}
		if v, ok := tfMap[names.AttrSchedule].(string); ok && v != "" {
			apiObject.ScheduleExpression = aws.String(v)
		}
		if v, ok := tfMap["schedule_expression_timezone"].(string); ok && v != "" {
			apiObject.ScheduleExpressionTimezone = aws.String(v)
		}
		if v, ok := tfMap["start_window"].(int); ok {
			apiObject.StartWindowMinutes = aws.Int64(int64(v))
		}
		if v, ok := tfMap["target_vault_name"].(string); ok && v != "" {
			apiObject.TargetBackupVaultName = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandAdvancedBackupSetting(tfList []any) []awstypes.AdvancedBackupSetting { // nosemgrep:ci.backup-in-func-name
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []awstypes.AdvancedBackupSetting{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.AdvancedBackupSetting{}

		if v, ok := tfMap["backup_options"].(map[string]any); ok && v != nil {
			apiObject.BackupOptions = flex.ExpandStringValueMap(v)
		}
		if v, ok := tfMap[names.AttrResourceType].(string); ok && v != "" {
			apiObject.ResourceType = aws.String(v)
		}

		// https://github.com/hashicorp/terraform-plugin-sdk/issues/588
		// Map in Set may add empty element. Ignore it.
		if apiObject.ResourceType == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCopyActions(tfList []any) []awstypes.CopyAction {
	apiObjects := []awstypes.CopyAction{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.CopyAction{}

		apiObject.DestinationBackupVaultArn = aws.String(tfMap["destination_vault_arn"].(string))

		if v, ok := tfMap["lifecycle"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.Lifecycle = expandLifecycle(v[0].(map[string]any))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLifecycle(tfMap map[string]any) *awstypes.Lifecycle {
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

func flattenBackupRules(ctx context.Context, apiObjects []awstypes.BackupRule) []any { // nosemgrep:ci.backup-in-func-name
	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"completion_window":            aws.ToInt64(apiObject.CompletionWindowMinutes),
			"enable_continuous_backup":     aws.ToBool(apiObject.EnableContinuousBackup),
			"rule_name":                    aws.ToString(apiObject.RuleName),
			names.AttrSchedule:             aws.ToString(apiObject.ScheduleExpression),
			"schedule_expression_timezone": aws.ToString(apiObject.ScheduleExpressionTimezone),
			"start_window":                 aws.ToInt64(apiObject.StartWindowMinutes),
			"target_vault_name":            aws.ToString(apiObject.TargetBackupVaultName),
		}

		if v := apiObject.CopyActions; len(v) > 0 {
			tfMap["copy_action"] = flattenCopyActions(v)
		}

		if v := apiObject.Lifecycle; v != nil {
			tfMap["lifecycle"] = flattenLifecycle(v)
		}

		if v := keyValueTags(ctx, apiObject.RecoveryPointTags).IgnoreAWS().Map(); len(v) > 0 {
			tfMap["recovery_point_tags"] = v
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAdvancedBackupSettings(apiObjects []awstypes.AdvancedBackupSetting) []any { // nosemgrep:ci.backup-in-func-name
	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"backup_options":       apiObject.BackupOptions,
			names.AttrResourceType: aws.ToString(apiObject.ResourceType),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCopyActions(apiObjects []awstypes.CopyAction) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, copyAction := range apiObjects {
		tfMap := map[string]any{
			"destination_vault_arn": aws.ToString(copyAction.DestinationBackupVaultArn),
		}

		if copyAction.Lifecycle != nil {
			tfMap["lifecycle"] = flattenLifecycle(copyAction.Lifecycle)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenLifecycle(apiObject *awstypes.Lifecycle) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"delete_after":       aws.ToInt64(apiObject.DeleteAfterDays),
		"cold_storage_after": aws.ToInt64(apiObject.MoveToColdStorageAfterDays),
		"opt_in_to_archive_for_supported_resources": aws.ToBool(apiObject.OptInToArchiveForSupportedResources),
	}

	return []any{tfMap}
}
