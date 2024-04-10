// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_backup_restore_testing_plan", name="RestoreTestingPlan")
// @Tags(identifierAttribute="arn")
func ResourceRestoreTestingPlan() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRestoreTestingPlanCreate,
		ReadWithoutTimeout:   resourceRestoreTestingPlanRead,
		UpdateWithoutTimeout: resourceRestoreTestingPlanUpdate,
		DeleteWithoutTimeout: resourceRestoreTestingPlanDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"schedule_timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "UTC",
			},
			"start_window": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      24,
				ValidateFunc: validation.IntBetween(1, 168),
			},
			"recovery_point_selection": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"algorithm": {
							Type:     schema.TypeString,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(restoreTestingRecoveryPointSelectionAlgorithm_Values(), false),
							},
						},
						"exclude_vaults": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"include_vaults": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"recovery_point_types": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(restoreTestingRecoveryPointType_Values(), false),
							},
						},
						"selection_window": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      30,
							ValidateFunc: validation.IntBetween(1, 365),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRestoreTestingPlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	name := d.Get("name").(string)

	input := &backup.CreateRestoreTestingPlanInput{
		RestoreTestingPlan: &backup.RestoreTestingPlanForCreate{
			RestoreTestingPlanName:     aws.String(name),
			ScheduleExpression:         aws.String(d.Get("schedule").(string)),
			ScheduleExpressionTimezone: aws.String(d.Get("schedule_timezone").(string)),
			StartWindowHours:           aws.Int64(int64(d.Get("start_window").(int))),
		},
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("recovery_point_selection"); ok && len(v.([]interface{})) > 0 {
		input.RestoreTestingPlan.RecoveryPointSelection = expandRecoveryPointSelection(d.Get("recovery_point_selection").([]interface{}))
	}

	_, err := conn.CreateRestoreTestingPlanWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Restore Testing Plan (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceRestoreTestingPlanRead(ctx, d, meta)...)
}

func resourceRestoreTestingPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	output, err := FindRestoreTestingPlanByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Restore Testing Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Restore Testing Plan (%s): %s", d.Id(), err)
	}

	if err := d.Set("recovery_point_selection", flattenRestoreTestingPlanRecoveryPointSelection(output.RestoreTestingPlan.RecoveryPointSelection)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting recovery_point_selection: %s", err)
	}

	d.Set("arn", output.RestoreTestingPlan.RestoreTestingPlanArn)
	d.Set("name", output.RestoreTestingPlan.RestoreTestingPlanName)
	d.Set("schedule", output.RestoreTestingPlan.ScheduleExpression)
	d.Set("schedule_timezone", output.RestoreTestingPlan.ScheduleExpressionTimezone)
	d.Set("start_window", output.RestoreTestingPlan.StartWindowHours)

	return diags
}

func FindRestoreTestingPlanByID(ctx context.Context, conn *backup.Backup, id string) (*backup.GetRestoreTestingPlanOutput, error) {
	input := &backup.GetRestoreTestingPlanInput{
		RestoreTestingPlanName: aws.String(id),
	}

	output, err := conn.GetRestoreTestingPlanWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RestoreTestingPlan == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func resourceRestoreTestingPlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	if d.HasChanges("schedule", "schedule_timezone", "start_window", "recovery_point_selection") {
		input := &backup.UpdateRestoreTestingPlanInput{
			RestoreTestingPlanName: aws.String(d.Id()),
			RestoreTestingPlan: &backup.RestoreTestingPlanForUpdate{
				RecoveryPointSelection:     expandRecoveryPointSelection(d.Get("recovery_point_selection").([]interface{})),
				ScheduleExpression:         aws.String(d.Get("schedule").(string)),
				ScheduleExpressionTimezone: aws.String(d.Get("schedule_timezone").(string)),
				StartWindowHours:           aws.Int64(int64(d.Get("start_window").(int))),
			},
		}

		_, err := conn.UpdateRestoreTestingPlanWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Restore Testing Plan (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRestoreTestingPlanRead(ctx, d, meta)...)
}

func resourceRestoreTestingPlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	log.Printf("[DEBUG] Deleting Restore Testing Plan: %s", d.Id())
	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, timeout, func() (interface{}, error) {
		return conn.DeleteRestoreTestingPlanWithContext(ctx, &backup.DeleteRestoreTestingPlanInput{
			RestoreTestingPlanName: aws.String(d.Id()),
		})
	}, backup.ErrCodeInvalidRequestException, "Related recovery point selections must be deleted prior to restore testing plan")

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Restore Testing Plan (%s): %s", d.Id(), err)
	}

	return diags
}

func expandRecoveryPointSelection(recoveryPointSelectionMaps []interface{}) *backup.RestoreTestingRecoveryPointSelection {
	if len(recoveryPointSelectionMaps) == 0 {
		return nil
	}

	recoveryPointSelectionMap := recoveryPointSelectionMaps[0].(map[string]interface{})

	return &backup.RestoreTestingRecoveryPointSelection{
		Algorithm:           aws.String(recoveryPointSelectionMap["algorithm"].(string)),
		ExcludeVaults:       flex.ExpandStringSet(recoveryPointSelectionMap["exclude_vaults"].(*schema.Set)),
		IncludeVaults:       flex.ExpandStringSet(recoveryPointSelectionMap["include_vaults"].(*schema.Set)),
		RecoveryPointTypes:  flex.ExpandStringSet(recoveryPointSelectionMap["recovery_point_types"].(*schema.Set)),
		SelectionWindowDays: aws.Int64(int64(recoveryPointSelectionMap["selection_window"].(int))),
	}
}

func flattenRestoreTestingPlanRecoveryPointSelection(recoveryPointSelection *backup.RestoreTestingRecoveryPointSelection) []map[string]interface{} {
	vRecoveryPointSelection := make(map[string]interface{})

	vRecoveryPointSelection["algorithm"] = aws.StringValue(recoveryPointSelection.Algorithm)
	vRecoveryPointSelection["exclude_vaults"] = aws.StringValueSlice(recoveryPointSelection.ExcludeVaults)
	vRecoveryPointSelection["include_vaults"] = aws.StringValueSlice(recoveryPointSelection.IncludeVaults)
	vRecoveryPointSelection["recovery_point_types"] = aws.StringValueSlice(recoveryPointSelection.RecoveryPointTypes)
	vRecoveryPointSelection["selection_window"] = aws.Int64Value(recoveryPointSelection.SelectionWindowDays)

	return []map[string]interface{}{vRecoveryPointSelection}
}
