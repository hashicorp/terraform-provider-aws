// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_monitoring_schedule")
// @Tags(identifierAttribute="arn")
func ResourceMonitoringSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMonitoringScheduleCreate,
		ReadWithoutTimeout:   resourceMonitoringScheduleRead,
		UpdateWithoutTimeout: resourceMonitoringScheduleUpdate,
		DeleteWithoutTimeout: resourceMonitoringScheduleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			"monitoring_schedule_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"monitoring_job_definition_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validName,
						},
						"monitoring_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.MonitoringType_Values(), false),
						},
						"schedule_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrScheduleExpression: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexache.MustCompile(`^cron`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
								},
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMonitoringScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	var name string
	if v, ok := d.GetOk(names.AttrName); ok {
		name = v.(string)
	} else {
		name = id.UniqueId()
	}

	createOpts := &sagemaker.CreateMonitoringScheduleInput{
		MonitoringScheduleConfig: expandMonitoringScheduleConfig(d.Get("monitoring_schedule_config").([]interface{})),
		MonitoringScheduleName:   aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	_, err := conn.CreateMonitoringScheduleWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Monitoring Schedule (%s): %s", name, err)
	}

	d.SetId(name)
	if _, err := WaitMonitoringScheduleScheduled(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Monitoring Schedule (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceMonitoringScheduleRead(ctx, d, meta)...)
}

func resourceMonitoringScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	monitoringSchedule, err := FindMonitoringScheduleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		log.Printf("[WARN] SageMaker Monitoring Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Monitoring Schedule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, monitoringSchedule.MonitoringScheduleArn)
	d.Set(names.AttrName, monitoringSchedule.MonitoringScheduleName)

	if err := d.Set("monitoring_schedule_config", flattenMonitoringScheduleConfig(monitoringSchedule.MonitoringScheduleConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting monitoring_schedule_config for SageMaker Monitoring Schedule (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceMonitoringScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	if d.HasChanges("monitoring_schedule_config") {
		modifyOpts := &sagemaker.UpdateMonitoringScheduleInput{
			MonitoringScheduleName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("monitoring_schedule_config"); ok && (len(v.([]interface{})) > 0) {
			modifyOpts.MonitoringScheduleConfig = expandMonitoringScheduleConfig(v.([]interface{}))
		}

		log.Printf("[INFO] Modifying monitoring_schedule_config attribute for %s: %#v", d.Id(), modifyOpts)
		if _, err := conn.UpdateMonitoringScheduleWithContext(ctx, modifyOpts); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Monitoring Schedule (%s): %s", d.Id(), err)
		}
		if _, err := WaitMonitoringScheduleScheduled(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating SageMaker Monitoring Schedule (%s): waiting for completion: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMonitoringScheduleRead(ctx, d, meta)...)
}

func resourceMonitoringScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	deleteOpts := &sagemaker.DeleteMonitoringScheduleInput{
		MonitoringScheduleName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Monitoring Schedule : %s", d.Id())

	_, err := conn.DeleteMonitoringScheduleWithContext(ctx, deleteOpts)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Monitoring Schedule (%s): %s", d.Id(), err)
	}

	if _, err := WaitMonitoringScheduleNotFound(ctx, conn, d.Id()); err != nil {
		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Monitoring Schedule (%s) to stop: %s", d.Id(), err)
		}
	}
	return diags
}

func FindMonitoringScheduleByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeMonitoringScheduleOutput, error) {
	input := &sagemaker.DescribeMonitoringScheduleInput{
		MonitoringScheduleName: aws.String(name),
	}

	output, err := conn.DescribeMonitoringScheduleWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
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

func expandMonitoringScheduleConfig(configured []interface{}) *sagemaker.MonitoringScheduleConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringScheduleConfig{}

	if v, ok := m["monitoring_job_definition_name"].(string); ok && v != "" {
		c.MonitoringJobDefinitionName = aws.String(v)
	}

	if v, ok := m["monitoring_type"].(string); ok && v != "" {
		c.MonitoringType = aws.String(v)
	}

	if v, ok := m["schedule_config"].([]interface{}); ok && len(v) > 0 {
		c.ScheduleConfig = expandScheduleConfig(v)
	}

	return c
}

func expandScheduleConfig(configured []interface{}) *sagemaker.ScheduleConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.ScheduleConfig{}

	if v, ok := m[names.AttrScheduleExpression].(string); ok && v != "" {
		c.ScheduleExpression = aws.String(v)
	}

	return c
}

func flattenMonitoringScheduleConfig(config *sagemaker.MonitoringScheduleConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.MonitoringJobDefinitionName != nil {
		m["monitoring_job_definition_name"] = aws.StringValue(config.MonitoringJobDefinitionName)
	}

	if config.MonitoringType != nil {
		m["monitoring_type"] = aws.StringValue(config.MonitoringType)
	}

	if config.ScheduleConfig != nil {
		m["schedule_config"] = flattenScheduleConfig(config.ScheduleConfig)
	}

	return []map[string]interface{}{m}
}

func flattenScheduleConfig(config *sagemaker.ScheduleConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.ScheduleExpression != nil {
		m[names.AttrScheduleExpression] = aws.StringValue(config.ScheduleExpression)
	}

	return []map[string]interface{}{m}
}
