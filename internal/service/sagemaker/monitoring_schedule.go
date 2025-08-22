// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_monitoring_schedule", name="Monitoring Schedule")
// @Tags(identifierAttribute="arn")
func resourceMonitoringSchedule() *schema.Resource {
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.MonitoringType](),
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
	}
}

func resourceMonitoringScheduleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	var name string
	if v, ok := d.GetOk(names.AttrName); ok {
		name = v.(string)
	} else {
		name = id.UniqueId()
	}

	createOpts := &sagemaker.CreateMonitoringScheduleInput{
		MonitoringScheduleConfig: expandMonitoringScheduleConfig(d.Get("monitoring_schedule_config").([]any)),
		MonitoringScheduleName:   aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	_, err := conn.CreateMonitoringSchedule(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Monitoring Schedule (%s): %s", name, err)
	}

	d.SetId(name)
	if err := waitMonitoringScheduleScheduled(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Monitoring Schedule (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceMonitoringScheduleRead(ctx, d, meta)...)
}

func resourceMonitoringScheduleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	monitoringSchedule, err := findMonitoringScheduleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Monitoring Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Monitoring Schedule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, monitoringSchedule.MonitoringScheduleArn)
	d.Set(names.AttrName, monitoringSchedule.MonitoringScheduleName)

	if err := d.Set("monitoring_schedule_config", flattenMonitoringScheduleConfig(monitoringSchedule.MonitoringScheduleConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting monitoring_schedule_config for SageMaker AI Monitoring Schedule (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceMonitoringScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChanges("monitoring_schedule_config") {
		modifyOpts := &sagemaker.UpdateMonitoringScheduleInput{
			MonitoringScheduleName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("monitoring_schedule_config"); ok && (len(v.([]any)) > 0) {
			modifyOpts.MonitoringScheduleConfig = expandMonitoringScheduleConfig(v.([]any))
		}

		log.Printf("[INFO] Modifying monitoring_schedule_config attribute for %s: %#v", d.Id(), modifyOpts)
		if _, err := conn.UpdateMonitoringSchedule(ctx, modifyOpts); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Monitoring Schedule (%s): %s", d.Id(), err)
		}
		if err := waitMonitoringScheduleScheduled(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Monitoring Schedule (%s): waiting for completion: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMonitoringScheduleRead(ctx, d, meta)...)
}

func resourceMonitoringScheduleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	deleteOpts := &sagemaker.DeleteMonitoringScheduleInput{
		MonitoringScheduleName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker AI Monitoring Schedule : %s", d.Id())

	_, err := conn.DeleteMonitoringSchedule(ctx, deleteOpts)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Monitoring Schedule (%s): %s", d.Id(), err)
	}

	if _, err := waitMonitoringScheduleNotFound(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Monitoring Schedule (%s) to delete: %s", d.Id(), err)
	}
	return diags
}

func findMonitoringScheduleByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeMonitoringScheduleOutput, error) {
	input := &sagemaker.DescribeMonitoringScheduleInput{
		MonitoringScheduleName: aws.String(name),
	}

	output, err := conn.DescribeMonitoringSchedule(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
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

func expandMonitoringScheduleConfig(configured []any) *awstypes.MonitoringScheduleConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.MonitoringScheduleConfig{}

	if v, ok := m["monitoring_job_definition_name"].(string); ok && v != "" {
		c.MonitoringJobDefinitionName = aws.String(v)
	}

	if v, ok := m["monitoring_type"].(string); ok && v != "" {
		c.MonitoringType = awstypes.MonitoringType(v)
	}

	if v, ok := m["schedule_config"].([]any); ok && len(v) > 0 {
		c.ScheduleConfig = expandScheduleConfig(v)
	}

	return c
}

func expandScheduleConfig(configured []any) *awstypes.ScheduleConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.ScheduleConfig{}

	if v, ok := m[names.AttrScheduleExpression].(string); ok && v != "" {
		c.ScheduleExpression = aws.String(v)
	}

	return c
}

func flattenMonitoringScheduleConfig(config *awstypes.MonitoringScheduleConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.MonitoringJobDefinitionName != nil {
		m["monitoring_job_definition_name"] = aws.ToString(config.MonitoringJobDefinitionName)
	}

	m["monitoring_type"] = config.MonitoringType

	if config.ScheduleConfig != nil {
		m["schedule_config"] = flattenScheduleConfig(config.ScheduleConfig)
	}

	return []map[string]any{m}
}

func flattenScheduleConfig(config *awstypes.ScheduleConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if config.ScheduleExpression != nil {
		m[names.AttrScheduleExpression] = aws.ToString(config.ScheduleExpression)
	}

	return []map[string]any{m}
}
