// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_redshiftserverless_scheduled_action", name="Scheduled Action")
// @Tags(identifierAttribute="arn")
func resourceScheduledAction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScheduledActionCreate,
		ReadWithoutTimeout:   resourceScheduledActionRead,
		UpdateWithoutTimeout: resourceScheduledActionUpdate,
		DeleteWithoutTimeout: resourceScheduledActionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"end_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-z0-9-]+{3,60}$`), "must contain only lowercase alphanumeric characters, or hyphen, and between 3 and 60 characters"),
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"start_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"target_action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create_snapshot": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"snapshot_name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9a-z-]{3,60}$`), "must contain only lowercase alphanumeric characters, or hyphen, and between 3 and 60 characters"),
									},
									"namespace_name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9a-z-]{1,235}$`), "must contain only lowercase alphanumeric characters, or hyphen, and at most 235 characters"),
									},
									"retention_period": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-z0-9-]+{3,64}$`), "must contain only lowercase alphanumeric characters, or hyphen, and between 3 and 64 characters"),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceScheduledActionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	name := d.Get("name").(string)
	input := &redshiftserverless.CreateScheduledActionInput{
		Enabled:             aws.Bool(d.Get("enabled").(bool)),
		NamespaceName:       aws.String(d.Get("namespace_name").(string)),
		ScheduledActionName: aws.String(name),
		RoleArn:             aws.String(d.Get("role_arn").(string)),
		Schedule:            aws.String(d.Get("schedule").(string)),
		TargetAction:        expandTargetAction(d.Get("target_action").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		input.ScheduledActionDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("end_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))

		input.EndTime = aws.Time(t)
	}

	if v, ok := d.GetOk("start_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))

		input.StartTime = aws.Time(t)
	}

	log.Printf("[DEBUG] Creating Redshift Serverless Scheduled Action: %s", input)
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateScheduledActionWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "The IAM role must delegate access to Amazon Redshift scheduler") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Scheduled Action (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*redshift.CreateScheduledActionOutput).ScheduledActionName))

	return append(diags, resourceScheduledActionRead(ctx, d, meta)...)
}

func resourceScheduledActionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	scheduledAction, err := FindScheduledActionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Scheduled Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Scheduled Action (%s): %s", d.Id(), err)
	}

	d.Set("namespace_name", output.NamespaceName)
	d.Set("description", scheduledAction.ScheduledActionDescription)
	if aws.StringValue(scheduledAction.State) == redshiftserverless.ScheduledActionStateActive {
		d.Set("enabled", true)
	} else {
		d.Set("enabled", false)
	}
	if scheduledAction.EndTime != nil {
		d.Set("end_time", aws.TimeValue(scheduledAction.EndTime).Format(time.RFC3339))
	} else {
		d.Set("end_time", nil)
	}
	d.Set("role_arn", scheduledAction.RoleArn)
	d.Set("name", scheduledAction.ScheduledActionName)
	d.Set("schedule", scheduledAction.Schedule)
	if scheduledAction.StartTime != nil {
		d.Set("start_time", aws.TimeValue(scheduledAction.StartTime).Format(time.RFC3339))
	} else {
		d.Set("start_time", nil)
	}

	if scheduledAction.TargetAction != nil {
		if err := d.Set("target_action", []interface{}{flattenTargetAction(scheduledAction.TargetAction)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_action: %s", err)
		}
	} else {
		d.Set("target_action", nil)
	}

	return diags

}

func resourceScheduledActionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	input := &redshiftserverless.UpdateScheduledActionInput{
		ScheduledActionName: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("description") {
		input.ScheduledActionDescription = aws.String(d.Get("description").(string))
	}

	if d.HasChange("enabled") {
		input.Enabled = aws.Bool(d.Get("enabled").(bool))
	}

	if hasChange, v := d.HasChange("end_time"), d.Get("end_time").(string); hasChange && v != "" {
		t, _ := time.Parse(time.RFC3339, v)

		input.EndTime = aws.Time(t)
	}

	if d.HasChange("role_arn") {
		input.IamRole = aws.String(d.Get("role_arn").(string))
	}

	if d.HasChange("schedule") {
		input.Schedule = aws.String(d.Get("schedule").(string))
	}

	if hasChange, v := d.HasChange("start_time"), d.Get("start_time").(string); hasChange && v != "" {
		t, _ := time.Parse(time.RFC3339, v)

		input.StartTime = aws.Time(t)
	}

	if d.HasChange("target_action") {
		input.TargetAction = expandTargetAction(d.Get("target_action").([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Updating Redshift Serverless Scheduled Action: %s", input)
	_, err := conn.UpdateScheduledActionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Redshift Serverless Scheduled Action (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceScheduledActionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Scheduled Action: %s", d.Id())
	_, err := conn.DeleteScheduledActionWithContext(ctx, &redshiftserverless.DeleteScheduledActionInput{
		ScheduledActionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeScheduledActionNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Scheduled Action (%s): %s", d.Id(), err)
	}

	return diags
}

func expandTargetAction(tfMap map[string]interface{}) *types.TargetActionMemberCreateSnapshot {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.TargetActionMemberCreateSnapshot{}

	if v, ok := tfMap["create_snapshot"].([]interface{}); ok && len(v) > 0 {
		apiObject.CreateSnapshot = expandCreateSnapshotScheduleActionParameters(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCreateSnapshotScheduleActionParameters(tfMap map[string]interface{}) *redshiftserverless.CreateSnapshotScheduleActionParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &redshiftserverless.CreateSnapshotScheduleActionParameters{}

	if v, ok := tfMap["namespace_name"].(string); ok && v != "" {
		apiObject.namespaceName = aws.String(v)
	}

	if v, ok := tfMap["snapshot_name_prefix"].(string); ok && v != "" {
		apiObject.snapshotNamePrefix = aws.String(v)
	}

	if v, ok := tfMap["retention_period"].(int); ok && v != 0 {
		apiObject.retentionPeriod = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenTargetAction(apiObject *redshiftserverless.TargetActionMemberCreateSnapshot) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ResizeCluster; v != nil {
		tfMap["create_snapshot"] = []interface{}{flattenCreateSnapshotScheduleActionParameters(v)}
	}

	return tfMap
}

func flattenCreateSnapshotScheduleActionParameters(apiObject *redshiftserverless.CreateSnapshotScheduleActionParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.namespaceName; v != nil {
		tfMap["namespace_name"] = aws.StringValue(v)
	}

	if v := apiObject.snapshotNamePrefix; v != nil {
		tfMap["snapshot_name_prefix"] = aws.StringValue(v)
	}

	if v := apiObject.retentionPeriod; v != nil {
		tfMap["retention_period"] = aws.Int64Value(v)
	}

	return tfMap
}