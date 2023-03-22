package sagemaker

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_sagemaker_monitoring_schedule")
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"schedule_expression": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexp.MustCompile(`^cron`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
								},
							},
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMonitoringScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	createOpts := &sagemaker.CreateMonitoringScheduleInput{
		MonitoringScheduleName:   aws.String(name),
		MonitoringScheduleConfig: expandMonitoringScheduleConfig(d.Get("monitoring_schedule_config").([]interface{})),
	}

	if len(tags) > 0 {
		createOpts.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] SageMaker Monitoring Schedule create config: %#v", *createOpts)
	_, err := conn.CreateMonitoringScheduleWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Monitoring Schedule: %s", err)
	}
	d.SetId(name)
	if _, err := WaitMonitoringScheduleScheduled(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Monitoring Schedule (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceMonitoringScheduleRead(ctx, d, meta)...)
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

	if v, ok := m["schedule_expression"].(string); ok && v != "" {
		c.ScheduleExpression = aws.String(v)
	}

	return c
}

func resourceMonitoringScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	monitoringSchedule, err := FindMonitoringScheduleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		log.Printf("[WARN] SageMaker Monitoring Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Monitoring Schedule (%s): %s", d.Id(), err)
	}

	d.Set("arn", monitoringSchedule.MonitoringScheduleArn)
	d.Set("name", monitoringSchedule.MonitoringScheduleName)

	if err := d.Set("monitoring_schedule_config", flattenMonitoringScheduleConfig(monitoringSchedule.MonitoringScheduleConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting monitoring_schedule_config for SageMaker Monitoring Schedule (%s): %s", d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, aws.StringValue(monitoringSchedule.MonitoringScheduleArn))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SageMaker Monitoring Schedule (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func flattenMonitoringScheduleConfig(monitoringScheduleConfig *sagemaker.MonitoringScheduleConfig) []map[string]interface{} {
	if monitoringScheduleConfig == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if monitoringScheduleConfig.MonitoringJobDefinitionName != nil {
		spec["monitoring_job_definition_name"] = aws.StringValue(monitoringScheduleConfig.MonitoringJobDefinitionName)
	}

	if monitoringScheduleConfig.MonitoringType != nil {
		spec["monitoring_type"] = aws.StringValue(monitoringScheduleConfig.MonitoringType)
	}

	if monitoringScheduleConfig.ScheduleConfig != nil {
		spec["schedule_config"] = flattenScheduleConfig(monitoringScheduleConfig.ScheduleConfig)
	}

	return []map[string]interface{}{spec}
}

func flattenScheduleConfig(scheduleConfig *sagemaker.ScheduleConfig) []map[string]interface{} {
	if scheduleConfig == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if scheduleConfig.ScheduleExpression != nil {
		spec["schedule_expression"] = aws.StringValue(scheduleConfig.ScheduleExpression)
	}

	return []map[string]interface{}{spec}
}

func resourceMonitoringScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Monitoring Schedule (%s) tags: %s", d.Id(), err)
		}
	}

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
	conn := meta.(*conns.AWSClient).SageMakerConn()

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
