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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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
func resourceReportPlan() *schema.Resource {
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
		IdempotencyToken:      aws.String(sdkid.UniqueId()),
		ReportDeliveryChannel: expandReportDeliveryChannel(d.Get("report_delivery_channel").([]interface{})),
		ReportPlanName:        aws.String(name),
		ReportPlanTags:        getTagsIn(ctx),
		ReportSetting:         expandReportSetting(d.Get("report_setting").([]interface{})),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.ReportPlanDescription = aws.String(v.(string))
	}

	_, err := conn.CreateReportPlan(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Report Plan (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitReportPlanCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Backup Report Plan (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceReportPlanRead(ctx, d, meta)...)
}

func resourceReportPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	reportPlan, err := findReportPlanByName(ctx, conn, d.Id())

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
			IdempotencyToken:      aws.String(sdkid.UniqueId()),
			ReportDeliveryChannel: expandReportDeliveryChannel(d.Get("report_delivery_channel").([]interface{})),
			ReportPlanDescription: aws.String(d.Get(names.AttrDescription).(string)),
			ReportPlanName:        aws.String(d.Id()),
			ReportSetting:         expandReportSetting(d.Get("report_setting").([]interface{})),
		}

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

func expandReportDeliveryChannel(tfList []interface{}) *awstypes.ReportDeliveryChannel {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ReportDeliveryChannel{
		S3BucketName: aws.String(tfMap[names.AttrS3BucketName].(string)),
	}

	if v, ok := tfMap["formats"]; ok && v.(*schema.Set).Len() > 0 {
		apiObject.Formats = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap[names.AttrS3KeyPrefix].(string); ok && v != "" {
		apiObject.S3KeyPrefix = aws.String(v)
	}

	return apiObject
}

func expandReportSetting(tfList []interface{}) *awstypes.ReportSetting {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ReportSetting{
		ReportTemplate: aws.String(tfMap["report_template"].(string)),
	}

	if v, ok := tfMap["accounts"]; ok && v.(*schema.Set).Len() > 0 {
		apiObject.Accounts = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["framework_arns"]; ok && v.(*schema.Set).Len() > 0 {
		apiObject.FrameworkArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["number_of_frameworks"].(int); ok && v > 0 {
		apiObject.NumberOfFrameworks = int32(v)
	}

	if v, ok := tfMap["organization_units"]; ok && v.(*schema.Set).Len() > 0 {
		apiObject.OrganizationUnits = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["regions"]; ok && v.(*schema.Set).Len() > 0 {
		apiObject.Regions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	return apiObject
}

func flattenReportDeliveryChannel(apiObject *awstypes.ReportDeliveryChannel) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrS3BucketName: aws.ToString(apiObject.S3BucketName),
	}

	if len(apiObject.Formats) > 0 {
		tfMap["formats"] = apiObject.Formats
	}

	if v := apiObject.S3KeyPrefix; v != nil {
		tfMap[names.AttrS3KeyPrefix] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenReportSetting(apiObject *awstypes.ReportSetting) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"report_template": aws.ToString(apiObject.ReportTemplate),
	}

	if len(apiObject.Accounts) > 0 {
		tfMap["accounts"] = apiObject.Accounts
	}

	if len(apiObject.FrameworkArns) > 0 {
		tfMap["framework_arns"] = apiObject.FrameworkArns
	}

	tfMap["number_of_frameworks"] = apiObject.NumberOfFrameworks

	if len(apiObject.OrganizationUnits) > 0 {
		tfMap["organization_units"] = apiObject.OrganizationUnits
	}

	if len(apiObject.Regions) > 0 {
		tfMap["regions"] = apiObject.Regions
	}

	return []interface{}{tfMap}
}

func findReportPlanByName(ctx context.Context, conn *backup.Client, name string) (*awstypes.ReportPlan, error) {
	input := &backup.DescribeReportPlanInput{
		ReportPlanName: aws.String(name),
	}

	return findReportPlan(ctx, conn, input)
}

func findReportPlan(ctx context.Context, conn *backup.Client, input *backup.DescribeReportPlanInput) (*awstypes.ReportPlan, error) {
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

func statusReportPlan(ctx context.Context, conn *backup.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findReportPlanByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.DeploymentStatus), nil
	}
}

const (
	reportPlanDeploymentStatusCompleted        = "COMPLETED"
	reportPlanDeploymentStatusCreateInProgress = "CREATE_IN_PROGRESS"
	reportPlanDeploymentStatusDeleteInProgress = "DELETE_IN_PROGRESS"
	reportPlanDeploymentStatusUpdateInProgress = "UPDATE_IN_PROGRESS"
)

func waitReportPlanCreated(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*awstypes.ReportPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{reportPlanDeploymentStatusCreateInProgress},
		Target:  []string{reportPlanDeploymentStatusCompleted},
		Timeout: timeout,
		Refresh: statusReportPlan(ctx, conn, name),
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
		Refresh: statusReportPlan(ctx, conn, name),
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
		Refresh: statusReportPlan(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReportPlan); ok {
		return output, err
	}

	return nil, err
}
