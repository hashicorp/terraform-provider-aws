// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_backup_report_plan", name="Report Plan")
// @Tags(identifierAttribute="arn")
func ResourceReportPlan() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReportPlanCreate,
		ReadWithoutTimeout:   resourceReportPlanRead,
		UpdateWithoutTimeout: resourceReportPlanUpdate,
		DeleteWithoutTimeout: resourceReportPlanDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(reportDeliveryChannelFormat_Values(), false),
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
						"accounts": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
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
						"organization_units": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"regions": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						// A report plan template cannot be updated
						"report_template": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(reportSettingTemplate_Values(), false),
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

func resourceReportPlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	name := d.Get("name").(string)
	input := &backup.CreateReportPlanInput{
		IdempotencyToken:      aws.String(id.UniqueId()),
		ReportDeliveryChannel: expandReportDeliveryChannel(d.Get("report_delivery_channel").([]interface{})),
		ReportPlanName:        aws.String(name),
		ReportPlanTags:        getTagsIn(ctx),
		ReportSetting:         expandReportSetting(d.Get("report_setting").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		input.ReportPlanDescription = aws.String(v.(string))
	}

	output, err := conn.CreateReportPlanWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Report Plan (%s): %s", name, err)
	}

	// Set ID with the name since the name is unique for the report plan.
	d.SetId(aws.StringValue(output.ReportPlanName))

	if _, err := waitReportPlanCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Backup Report Plan (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceReportPlanRead(ctx, d, meta)...)
}

func resourceReportPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	reportPlan, err := FindReportPlanByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Report Plan %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Report Plan (%s): %s", d.Id(), err)
	}

	d.Set("arn", reportPlan.ReportPlanArn)
	d.Set("creation_time", reportPlan.CreationTime.Format(time.RFC3339))
	d.Set("deployment_status", reportPlan.DeploymentStatus)
	d.Set("description", reportPlan.ReportPlanDescription)
	d.Set("name", reportPlan.ReportPlanName)

	if err := d.Set("report_delivery_channel", flattenReportDeliveryChannel(reportPlan.ReportDeliveryChannel)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting report_delivery_channel: %s", err)
	}

	if err := d.Set("report_setting", flattenReportSetting(reportPlan.ReportSetting)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting report_setting: %s", err)
	}

	return diags
}

func resourceReportPlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	if d.HasChangesExcept("tags_all", "tags") {
		input := &backup.UpdateReportPlanInput{
			IdempotencyToken:      aws.String(id.UniqueId()),
			ReportDeliveryChannel: expandReportDeliveryChannel(d.Get("report_delivery_channel").([]interface{})),
			ReportPlanDescription: aws.String(d.Get("description").(string)),
			ReportPlanName:        aws.String(d.Id()),
			ReportSetting:         expandReportSetting(d.Get("report_setting").([]interface{})),
		}

		log.Printf("[DEBUG] Updating Backup Report Plan: %s", input)
		_, err := conn.UpdateReportPlanWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Backup Report Plan (%s): %s", d.Id(), err)
		}

		if _, err := waitReportPlanUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Backup Report Plan (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReportPlanRead(ctx, d, meta)...)
}

func resourceReportPlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)

	log.Printf("[DEBUG] Deleting Backup Report Plan: %s", d.Id())
	_, err := conn.DeleteReportPlanWithContext(ctx, &backup.DeleteReportPlanInput{
		ReportPlanName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Report Plan (%s): %s", d.Id(), err)
	}

	if _, err := waitReportPlanDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Backup Report Plan (%s) delete: %s", d.Id(), err)
	}

	return diags
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

	if v, ok := tfMap["accounts"]; ok && v.(*schema.Set).Len() > 0 {
		result.Accounts = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := tfMap["framework_arns"]; ok && v.(*schema.Set).Len() > 0 {
		result.FrameworkArns = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := tfMap["number_of_frameworks"].(int); ok && v > 0 {
		result.NumberOfFrameworks = aws.Int64(int64(v))
	}

	if v, ok := tfMap["organization_units"]; ok && v.(*schema.Set).Len() > 0 {
		result.OrganizationUnits = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := tfMap["regions"]; ok && v.(*schema.Set).Len() > 0 {
		result.Regions = flex.ExpandStringSet(v.(*schema.Set))
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

	if reportSetting.Accounts != nil && len(reportSetting.Accounts) > 0 {
		values["accounts"] = flex.FlattenStringSet(reportSetting.Accounts)
	}

	if reportSetting.FrameworkArns != nil && len(reportSetting.FrameworkArns) > 0 {
		values["framework_arns"] = flex.FlattenStringSet(reportSetting.FrameworkArns)
	}

	if reportSetting.NumberOfFrameworks != nil {
		values["number_of_frameworks"] = aws.Int64Value(reportSetting.NumberOfFrameworks)
	}

	if reportSetting.OrganizationUnits != nil && len(reportSetting.OrganizationUnits) > 0 {
		values["organization_units"] = flex.FlattenStringSet(reportSetting.OrganizationUnits)
	}

	if reportSetting.Regions != nil && len(reportSetting.Regions) > 0 {
		values["regions"] = flex.FlattenStringSet(reportSetting.Regions)
	}

	return []interface{}{values}
}

func FindReportPlanByName(ctx context.Context, conn *backup.Backup, name string) (*backup.ReportPlan, error) {
	input := &backup.DescribeReportPlanInput{
		ReportPlanName: aws.String(name),
	}

	output, err := conn.DescribeReportPlanWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ReportPlan == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ReportPlan, nil
}

func statusReportPlanDeployment(ctx context.Context, conn *backup.Backup, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindReportPlanByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DeploymentStatus), nil
	}
}

func waitReportPlanCreated(ctx context.Context, conn *backup.Backup, name string, timeout time.Duration) (*backup.ReportPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{reportPlanDeploymentStatusCreateInProgress},
		Target:  []string{reportPlanDeploymentStatusCompleted},
		Timeout: timeout,
		Refresh: statusReportPlanDeployment(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.ReportPlan); ok {
		return output, err
	}

	return nil, err
}

func waitReportPlanDeleted(ctx context.Context, conn *backup.Backup, name string, timeout time.Duration) (*backup.ReportPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{reportPlanDeploymentStatusDeleteInProgress},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusReportPlanDeployment(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.ReportPlan); ok {
		return output, err
	}

	return nil, err
}

func waitReportPlanUpdated(ctx context.Context, conn *backup.Backup, name string, timeout time.Duration) (*backup.ReportPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{reportPlanDeploymentStatusUpdateInProgress},
		Target:  []string{reportPlanDeploymentStatusCompleted},
		Timeout: timeout,
		Refresh: statusReportPlanDeployment(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.ReportPlan); ok {
		return output, err
	}

	return nil, err
}
