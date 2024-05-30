// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_scheduled_action", name="Scheduled Action")
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
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"end_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"iam_role": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9a-z-]{1,63}$`), ""),
			},
			names.AttrSchedule: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrStartTime: {
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
						"pause_cluster": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrClusterIdentifier: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							ExactlyOneOf: []string{
								"target_action.0.pause_cluster",
								"target_action.0.resize_cluster",
								"target_action.0.resume_cluster",
							},
						},
						"resize_cluster": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"classic": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									names.AttrClusterIdentifier: {
										Type:     schema.TypeString,
										Required: true,
									},
									"cluster_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"node_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"number_of_nodes": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
							ExactlyOneOf: []string{
								"target_action.0.pause_cluster",
								"target_action.0.resize_cluster",
								"target_action.0.resume_cluster",
							},
						},
						"resume_cluster": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrClusterIdentifier: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							ExactlyOneOf: []string{
								"target_action.0.pause_cluster",
								"target_action.0.resize_cluster",
								"target_action.0.resume_cluster",
							},
						},
					},
				},
			},
		},
	}
}

func resourceScheduledActionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &redshift.CreateScheduledActionInput{
		Enable:              aws.Bool(d.Get("enable").(bool)),
		IamRole:             aws.String(d.Get("iam_role").(string)),
		Schedule:            aws.String(d.Get(names.AttrSchedule).(string)),
		ScheduledActionName: aws.String(name),
		TargetAction:        expandScheduledActionType(d.Get("target_action").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.ScheduledActionDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("end_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))

		input.EndTime = aws.Time(t)
	}

	if v, ok := d.GetOk(names.AttrStartTime); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))

		input.StartTime = aws.Time(t)
	}

	log.Printf("[DEBUG] Creating Redshift Scheduled Action: %s", input)
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
		return sdkdiag.AppendErrorf(diags, "creating Redshift Scheduled Action (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*redshift.CreateScheduledActionOutput).ScheduledActionName))

	return append(diags, resourceScheduledActionRead(ctx, d, meta)...)
}

func resourceScheduledActionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	scheduledAction, err := findScheduledActionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Scheduled Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Scheduled Action (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDescription, scheduledAction.ScheduledActionDescription)
	if aws.StringValue(scheduledAction.State) == redshift.ScheduledActionStateActive {
		d.Set("enable", true)
	} else {
		d.Set("enable", false)
	}
	if scheduledAction.EndTime != nil {
		d.Set("end_time", aws.TimeValue(scheduledAction.EndTime).Format(time.RFC3339))
	} else {
		d.Set("end_time", nil)
	}
	d.Set("iam_role", scheduledAction.IamRole)
	d.Set(names.AttrName, scheduledAction.ScheduledActionName)
	d.Set(names.AttrSchedule, scheduledAction.Schedule)
	if scheduledAction.StartTime != nil {
		d.Set(names.AttrStartTime, aws.TimeValue(scheduledAction.StartTime).Format(time.RFC3339))
	} else {
		d.Set(names.AttrStartTime, nil)
	}

	if scheduledAction.TargetAction != nil {
		if err := d.Set("target_action", []interface{}{flattenScheduledActionType(scheduledAction.TargetAction)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_action: %s", err)
		}
	} else {
		d.Set("target_action", nil)
	}

	return diags
}

func resourceScheduledActionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	input := &redshift.ModifyScheduledActionInput{
		ScheduledActionName: aws.String(d.Get(names.AttrName).(string)),
	}

	if d.HasChange(names.AttrDescription) {
		input.ScheduledActionDescription = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange("enable") {
		input.Enable = aws.Bool(d.Get("enable").(bool))
	}

	if hasChange, v := d.HasChange("end_time"), d.Get("end_time").(string); hasChange && v != "" {
		t, _ := time.Parse(time.RFC3339, v)

		input.EndTime = aws.Time(t)
	}

	if d.HasChange("iam_role") {
		input.IamRole = aws.String(d.Get("iam_role").(string))
	}

	if d.HasChange(names.AttrSchedule) {
		input.Schedule = aws.String(d.Get(names.AttrSchedule).(string))
	}

	if hasChange, v := d.HasChange(names.AttrStartTime), d.Get(names.AttrStartTime).(string); hasChange && v != "" {
		t, _ := time.Parse(time.RFC3339, v)

		input.StartTime = aws.Time(t)
	}

	if d.HasChange("target_action") {
		input.TargetAction = expandScheduledActionType(d.Get("target_action").([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Updating Redshift Scheduled Action: %s", input)
	_, err := conn.ModifyScheduledActionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Redshift Scheduled Action (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceScheduledActionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Scheduled Action: %s", d.Id())
	_, err := conn.DeleteScheduledActionWithContext(ctx, &redshift.DeleteScheduledActionInput{
		ScheduledActionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeScheduledActionNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Scheduled Action (%s): %s", d.Id(), err)
	}

	return diags
}

func expandScheduledActionType(tfMap map[string]interface{}) *redshift.ScheduledActionType {
	if tfMap == nil {
		return nil
	}

	apiObject := &redshift.ScheduledActionType{}

	if v, ok := tfMap["pause_cluster"].([]interface{}); ok && len(v) > 0 {
		apiObject.PauseCluster = expandPauseClusterMessage(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["resize_cluster"].([]interface{}); ok && len(v) > 0 {
		apiObject.ResizeCluster = expandResizeClusterMessage(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["resume_cluster"].([]interface{}); ok && len(v) > 0 {
		apiObject.ResumeCluster = expandResumeClusterMessage(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPauseClusterMessage(tfMap map[string]interface{}) *redshift.PauseClusterMessage {
	if tfMap == nil {
		return nil
	}

	apiObject := &redshift.PauseClusterMessage{}

	if v, ok := tfMap[names.AttrClusterIdentifier].(string); ok && v != "" {
		apiObject.ClusterIdentifier = aws.String(v)
	}

	return apiObject
}

func expandResizeClusterMessage(tfMap map[string]interface{}) *redshift.ResizeClusterMessage {
	if tfMap == nil {
		return nil
	}

	apiObject := &redshift.ResizeClusterMessage{}

	if v, ok := tfMap["classic"].(bool); ok {
		apiObject.Classic = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrClusterIdentifier].(string); ok && v != "" {
		apiObject.ClusterIdentifier = aws.String(v)
	}

	if v, ok := tfMap["cluster_type"].(string); ok && v != "" {
		apiObject.ClusterType = aws.String(v)
	}

	if v, ok := tfMap["node_type"].(string); ok && v != "" {
		apiObject.NodeType = aws.String(v)
	}

	if v, ok := tfMap["number_of_nodes"].(int); ok && v != 0 {
		apiObject.NumberOfNodes = aws.Int64(int64(v))
	}

	return apiObject
}

func expandResumeClusterMessage(tfMap map[string]interface{}) *redshift.ResumeClusterMessage {
	if tfMap == nil {
		return nil
	}

	apiObject := &redshift.ResumeClusterMessage{}

	if v, ok := tfMap[names.AttrClusterIdentifier].(string); ok && v != "" {
		apiObject.ClusterIdentifier = aws.String(v)
	}

	return apiObject
}

func flattenScheduledActionType(apiObject *redshift.ScheduledActionType) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PauseCluster; v != nil {
		tfMap["pause_cluster"] = []interface{}{flattenPauseClusterMessage(v)}
	}

	if v := apiObject.ResizeCluster; v != nil {
		tfMap["resize_cluster"] = []interface{}{flattenResizeClusterMessage(v)}
	}

	if v := apiObject.ResumeCluster; v != nil {
		tfMap["resume_cluster"] = []interface{}{flattenResumeClusterMessage(v)}
	}

	return tfMap
}

func flattenPauseClusterMessage(apiObject *redshift.PauseClusterMessage) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ClusterIdentifier; v != nil {
		tfMap[names.AttrClusterIdentifier] = aws.StringValue(v)
	}

	return tfMap
}

func flattenResizeClusterMessage(apiObject *redshift.ResizeClusterMessage) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Classic; v != nil {
		tfMap["classic"] = aws.BoolValue(v)
	}

	if v := apiObject.ClusterIdentifier; v != nil {
		tfMap[names.AttrClusterIdentifier] = aws.StringValue(v)
	}

	if v := apiObject.ClusterType; v != nil {
		tfMap["cluster_type"] = aws.StringValue(v)
	}

	if v := apiObject.NodeType; v != nil {
		tfMap["node_type"] = aws.StringValue(v)
	}

	if v := apiObject.NumberOfNodes; v != nil {
		tfMap["number_of_nodes"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenResumeClusterMessage(apiObject *redshift.ResumeClusterMessage) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ClusterIdentifier; v != nil {
		tfMap[names.AttrClusterIdentifier] = aws.StringValue(v)
	}

	return tfMap
}
