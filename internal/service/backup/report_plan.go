// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			names.AttrName: {
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
						names.AttrS3BucketName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrS3KeyPrefix: {
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
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &backup.CreateReportPlanInput{
		IdempotencyToken:      aws.String(id.UniqueId()),
		ReportDeliveryChannel: expandReportDeliveryChannel(d.Get("report_delivery_channel").([]interface{})),
		ReportPlanName:        aws.String(name),
		ReportPlanTags:        getTagsIn(ctx),
		ReportSetting:         expandReportSetting(d.Get("report_setting").([]interface{})),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.ReportPlanDescription = aws.String(v.(string))
	}

	output, err := conn.CreateReportPlan(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Report Plan (%s): %s", name, err)
	}

	// Set ID with the name since the name is unique for the report plan.
	d.SetId(aws.ToString(output.ReportPlanName))

	if _, err := waitReportPlanCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Backup Report Plan (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceReportPlanRead(ctx, d, meta)...)
}

func resourceReportPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	reportPlan, err := FindReportPlanByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Report Plan %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Report Plan (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, reportPlan.ReportPlanArn)
	d.Set(names.AttrCreationTime, reportPlan.CreationTime.Format(time.RFC3339))
	d.Set("deployment_status", reportPlan.DeploymentStatus)
	d.Set(names.AttrDescription, reportPlan.ReportPlanDescription)
	d.Set(names.AttrName, reportPlan.ReportPlanName)

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
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &backup.UpdateReportPlanInput{
			IdempotencyToken:      aws.String(id.UniqueId()),
			ReportDeliveryChannel: expandReportDeliveryChannel(d.Get("report_delivery_channel").([]interface{})),
			ReportPlanDescription: aws.String(d.Get(names.AttrDescription).(string)),
			ReportPlanName:        aws.String(d.Id()),
			ReportSetting:         expandReportSetting(d.Get("report_setting").([]interface{})),
		}

		log.Printf("[DEBUG] Updating Backup Report Plan: %+v", input)
		_, err := conn.UpdateReportPlan(ctx, input)

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
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	log.Printf("[DEBUG] Deleting Backup Report Plan: %s", d.Id())
	_, err := conn.DeleteReportPlan(ctx, &backup.DeleteReportPlanInput{
		ReportPlanName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Report Plan (%s): %s", d.Id(), err)
	}

	if _, err := waitReportPlanDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Backup Report Plan (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandReportDeliveryChannel(reportDeliveryChannel []interface{}) *awstypes.ReportDeliveryChannel {
	if len(reportDeliveryChannel) == 0 || reportDeliveryChannel[0] == nil {
		return nil
	}

	tfMap, ok := reportDeliveryChannel[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.ReportDeliveryChannel{
		S3BucketName: aws.String(tfMap[names.AttrS3BucketName].(string)),
	}

	if v, ok := tfMap["formats"]; ok && v.(*schema.Set).Len() > 0 {
		result.Formats = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap[names.AttrS3KeyPrefix].(string); ok && v != "" {
		result.S3KeyPrefix = aws.String(v)
	}

	return result
}

func expandReportSetting(reportSetting []interface{}) *awstypes.ReportSetting {
	if len(reportSetting) == 0 || reportSetting[0] == nil {
		return nil
	}

	tfMap, ok := reportSetting[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.ReportSetting{
		ReportTemplate: aws.String(tfMap["report_template"].(string)),
	}

	if v, ok := tfMap["accounts"]; ok && v.(*schema.Set).Len() > 0 {
		result.Accounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["framework_arns"]; ok && v.(*schema.Set).Len() > 0 {
		result.FrameworkArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["number_of_frameworks"].(int); ok && v > 0 {
		result.NumberOfFrameworks = int32(v)
	}

	if v, ok := tfMap["organization_units"]; ok && v.(*schema.Set).Len() > 0 {
		result.OrganizationUnits = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["regions"]; ok && v.(*schema.Set).Len() > 0 {
		result.Regions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	return result
}

func flattenReportDeliveryChannel(reportDeliveryChannel *awstypes.ReportDeliveryChannel) []interface{} {
	if reportDeliveryChannel == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		names.AttrS3BucketName: aws.ToString(reportDeliveryChannel.S3BucketName),
	}

	if reportDeliveryChannel.Formats != nil && len(reportDeliveryChannel.Formats) > 0 {
		values["formats"] = flex.FlattenStringValueSet(reportDeliveryChannel.Formats)
	}

	if v := reportDeliveryChannel.S3KeyPrefix; v != nil {
		values[names.AttrS3KeyPrefix] = aws.ToString(v)
	}

	return []interface{}{values}
}

func flattenReportSetting(reportSetting *awstypes.ReportSetting) []interface{} {
	if reportSetting == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"report_template": aws.ToString(reportSetting.ReportTemplate),
	}

	if reportSetting.Accounts != nil && len(reportSetting.Accounts) > 0 {
		values["accounts"] = flex.FlattenStringValueSet(reportSetting.Accounts)
	}

	if reportSetting.FrameworkArns != nil && len(reportSetting.FrameworkArns) > 0 {
		values["framework_arns"] = flex.FlattenStringValueSet(reportSetting.FrameworkArns)
	}

	values["number_of_frameworks"] = reportSetting.NumberOfFrameworks

	if reportSetting.OrganizationUnits != nil && len(reportSetting.OrganizationUnits) > 0 {
		values["organization_units"] = flex.FlattenStringValueSet(reportSetting.OrganizationUnits)
	}

	if reportSetting.Regions != nil && len(reportSetting.Regions) > 0 {
		values["regions"] = flex.FlattenStringValueSet(reportSetting.Regions)
	}

	return []interface{}{values}
}

func FindReportPlanByName(ctx context.Context, conn *backup.Client, name string) (*awstypes.ReportPlan, error) {
	input := &backup.DescribeReportPlanInput{
		ReportPlanName: aws.String(name),
	}

	output, err := conn.DescribeReportPlan(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func statusReportPlanDeployment(ctx context.Context, conn *backup.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindReportPlanByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.DeploymentStatus), nil
	}
}

func waitReportPlanCreated(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*awstypes.ReportPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{reportPlanDeploymentStatusCreateInProgress},
		Target:  []string{reportPlanDeploymentStatusCompleted},
		Timeout: timeout,
		Refresh: statusReportPlanDeployment(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReportPlan); ok {
		return output, err
	}

	return nil, err
}

func waitReportPlanDeleted(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*awstypes.ReportPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{reportPlanDeploymentStatusDeleteInProgress},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusReportPlanDeployment(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReportPlan); ok {
		return output, err
	}

	return nil, err
}

func waitReportPlanUpdated(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*awstypes.ReportPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{reportPlanDeploymentStatusUpdateInProgress},
		Target:  []string{reportPlanDeploymentStatusCompleted},
		Timeout: timeout,
		Refresh: statusReportPlanDeployment(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReportPlan); ok {
		return output, err
	}

	return nil, err
}
