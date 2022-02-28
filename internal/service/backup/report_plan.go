package backup

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReportPlan() *schema.Resource {
	return &schema.Resource{
		Create: resourceReportPlanCreate,
		Read:   resourceReportPlanRead,
		Update: resourceReportPlanUpdate,
		Delete: resourceReportPlanDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validReportPlanName,
			},
			"report_delivery_channel": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"formats": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"CSV",
									"JSON",
								}, false),
							},
						},
						"s3_bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"s3_key_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"report_setting": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"framework_arns": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"number_of_frameworks": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						// A report plan template cannot be updated
						"report_template": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"RESOURCE_COMPLIANCE_REPORT",
								"CONTROL_COMPLIANCE_REPORT",
								"BACKUP_JOB_REPORT",
								"COPY_JOB_REPORT",
								"RESTORE_JOB_REPORT",
							}, false),
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

func resourceReportPlanCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &backup.CreateReportPlanInput{
		IdempotencyToken:      aws.String(resource.UniqueId()),
		ReportDeliveryChannel: expandReportDeliveryChannel(d.Get("report_delivery_channel").([]interface{})),
		ReportPlanName:        aws.String(name),
		ReportSetting:         expandReportSetting(d.Get("report_setting").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		input.ReportPlanDescription = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.ReportPlanTags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Backup Report Plan: %#v", input)
	resp, err := conn.CreateReportPlan(input)
	if err != nil {
		return fmt.Errorf("error creating Backup Report Plan: %w", err)
	}

	// Set ID with the name since the name is unique for the report plan
	d.SetId(aws.StringValue(resp.ReportPlanName))

	return resourceReportPlanRead(d, meta)
}

func resourceReportPlanRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeReportPlan(&backup.DescribeReportPlanInput{
		ReportPlanName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Backup Report Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading Backup Report Plan (%s): %w", d.Id(), err)
	}

	d.Set("arn", resp.ReportPlan.ReportPlanArn)
	d.Set("deployment_status", resp.ReportPlan.DeploymentStatus)
	d.Set("description", resp.ReportPlan.ReportPlanDescription)
	d.Set("name", resp.ReportPlan.ReportPlanName)

	if err := d.Set("creation_time", resp.ReportPlan.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %s", err)
	}

	if err := d.Set("report_delivery_channel", flattenReportDeliveryChannel(resp.ReportPlan.ReportDeliveryChannel)); err != nil {
		return fmt.Errorf("error setting report_delivery_channel: %w", err)
	}

	if err := d.Set("report_setting", flattenReportSetting(resp.ReportPlan.ReportSetting)); err != nil {
		return fmt.Errorf("error setting report_delivery_channel: %w", err)
	}

	tags, err := ListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for Backup Report Plan (%s): %w", d.Id(), err)
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceReportPlanUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	if d.HasChanges("description", "report_delivery_channel", "report_plan_description", "report_setting") {
		input := &backup.UpdateReportPlanInput{
			IdempotencyToken:      aws.String(resource.UniqueId()),
			ReportDeliveryChannel: expandReportDeliveryChannel(d.Get("report_delivery_channel").([]interface{})),
			ReportPlanDescription: aws.String(d.Get("description").(string)),
			ReportPlanName:        aws.String(d.Id()),
			ReportSetting:         expandReportSetting(d.Get("report_setting").([]interface{})),
		}

		log.Printf("[DEBUG] Updating Backup Report Plan: %#v", input)
		_, err := conn.UpdateReportPlan(input)
		if err != nil {
			return fmt.Errorf("error updating Backup Report Plan (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for Backup Report Plan (%s): %w", d.Id(), err)
		}
	}

	return resourceReportPlanRead(d, meta)
}

func resourceReportPlanDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	input := &backup.DeleteReportPlanInput{
		ReportPlanName: aws.String(d.Id()),
	}

	_, err := conn.DeleteReportPlan(input)
	if err != nil {
		return fmt.Errorf("error deleting Backup Report Plan: %s", err)
	}

	return nil
}

func expandReportDeliveryChannel(reportDeliveryChannel []interface{}) *backup.ReportDeliveryChannel {
	if len(reportDeliveryChannel) == 0 || reportDeliveryChannel[0] == nil {
		return nil
	}

	tfMap, ok := reportDeliveryChannel[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &backup.ReportDeliveryChannel{
		S3BucketName: aws.String(tfMap["s3_bucket_name"].(string)),
	}

	if v, ok := tfMap["formats"]; ok && v.(*schema.Set).Len() > 0 {
		result.Formats = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := tfMap["s3_key_prefix"].(string); ok && v != "" {
		result.S3KeyPrefix = aws.String(v)
	}

	return result
}

func expandReportSetting(reportSetting []interface{}) *backup.ReportSetting {
	if len(reportSetting) == 0 || reportSetting[0] == nil {
		return nil
	}

	tfMap, ok := reportSetting[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &backup.ReportSetting{
		ReportTemplate: aws.String(tfMap["report_template"].(string)),
	}

	if v, ok := tfMap["framework_arns"]; ok && v.(*schema.Set).Len() > 0 {
		result.FrameworkArns = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := tfMap["number_of_frameworks"].(int); ok && v > 0 {
		result.NumberOfFrameworks = aws.Int64(int64(v))
	}

	return result
}

func flattenReportDeliveryChannel(reportDeliveryChannel *backup.ReportDeliveryChannel) []interface{} {
	if reportDeliveryChannel == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"s3_bucket_name": aws.StringValue(reportDeliveryChannel.S3BucketName),
	}

	if reportDeliveryChannel.Formats != nil && len(reportDeliveryChannel.Formats) > 0 {
		values["formats"] = flex.FlattenStringSet(reportDeliveryChannel.Formats)
	}

	if v := reportDeliveryChannel.S3KeyPrefix; v != nil {
		values["s3_key_prefix"] = aws.StringValue(v)
	}

	return []interface{}{values}
}

func flattenReportSetting(reportSetting *backup.ReportSetting) []interface{} {
	if reportSetting == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"report_template": aws.StringValue(reportSetting.ReportTemplate),
	}

	if reportSetting.FrameworkArns != nil && len(reportSetting.FrameworkArns) > 0 {
		values["framework_arns"] = flex.FlattenStringSet(reportSetting.FrameworkArns)
	}

	if reportSetting.NumberOfFrameworks != nil {
		values["number_of_frameworks"] = aws.Int64Value(reportSetting.NumberOfFrameworks)
	}

	return []interface{}{values}
}
