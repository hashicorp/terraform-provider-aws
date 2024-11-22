// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codebuild_report_group", name="Report Group")
// @Tags
func resourceReportGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReportGroupCreate,
		ReadWithoutTimeout:   resourceReportGroupRead,
		UpdateWithoutTimeout: resourceReportGroupUpdate,
		DeleteWithoutTimeout: resourceReportGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delete_reports": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"export_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrBucket: {
										Type:     schema.TypeString,
										Required: true,
									},
									"encryption_disabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"encryption_key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"packaging": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          types.ReportPackagingTypeNone,
										ValidateDiagFunc: enum.Validate[types.ReportPackagingType](),
									},
									names.AttrPath: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ReportExportConfigType](),
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 128),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ReportType](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReportGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &codebuild.CreateReportGroupInput{
		ExportConfig: expandReportGroupExportConfig(d.Get("export_config").([]interface{})),
		Name:         aws.String(name),
		Tags:         getTagsIn(ctx),
		Type:         types.ReportType(d.Get(names.AttrType).(string)),
	}

	output, err := conn.CreateReportGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeBuild Report Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ReportGroup.Arn))

	return append(diags, resourceReportGroupRead(ctx, d, meta)...)
}

func resourceReportGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	reportGroup, err := findReportGroupByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeBuild Report Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeBuild Report Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, reportGroup.Arn)
	d.Set("created", reportGroup.Created.Format(time.RFC3339))
	if err := d.Set("export_config", flattenReportGroupExportConfig(reportGroup.ExportConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting export config: %s", err)
	}
	d.Set(names.AttrName, reportGroup.Name)
	d.Set(names.AttrType, reportGroup.Type)

	setTagsOut(ctx, reportGroup.Tags)

	return diags
}

func resourceReportGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	input := &codebuild.UpdateReportGroupInput{
		Arn: aws.String(d.Id()),
	}

	if d.HasChange("export_config") {
		input.ExportConfig = expandReportGroupExportConfig(d.Get("export_config").([]interface{}))
	}

	if d.HasChange(names.AttrTagsAll) {
		input.Tags = getTagsIn(ctx)
	}

	_, err := conn.UpdateReportGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CodeBuild Report Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceReportGroupRead(ctx, d, meta)...)
}

func resourceReportGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	log.Printf("[INFO] Deleting CodeBuild Report Group: %s", d.Id())
	_, err := conn.DeleteReportGroup(ctx, &codebuild.DeleteReportGroupInput{
		Arn:           aws.String(d.Id()),
		DeleteReports: d.Get("delete_reports").(bool),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Report Group (%s): %s", d.Id(), err)
	}

	if _, err := waitReportGroupDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "while waiting for CodeBuild Report Group (%s) to become deleted: %s", d.Id(), err)
	}

	return diags
}

func findReportGroupByARN(ctx context.Context, conn *codebuild.Client, arn string) (*types.ReportGroup, error) {
	input := &codebuild.BatchGetReportGroupsInput{
		ReportGroupArns: []string{arn},
	}

	return findReportGroup(ctx, conn, input)
}

func findReportGroup(ctx context.Context, conn *codebuild.Client, input *codebuild.BatchGetReportGroupsInput) (*types.ReportGroup, error) {
	output, err := findReportGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReportGroups(ctx context.Context, conn *codebuild.Client, input *codebuild.BatchGetReportGroupsInput) ([]types.ReportGroup, error) {
	output, err := conn.BatchGetReportGroups(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ReportGroups, nil
}

func statusReportGroup(ctx context.Context, conn *codebuild.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findReportGroupByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitReportGroupDeleted(ctx context.Context, conn *codebuild.Client, arn string) (*types.ReportGroup, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ReportGroupStatusTypeDeleting),
		Target:  []string{},
		Refresh: statusReportGroup(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ReportGroup); ok {
		return output, err
	}

	return nil, err
}

func expandReportGroupExportConfig(tfList []interface{}) *types.ReportExportConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &types.ReportExportConfig{}

	if v, ok := tfMap["s3_destination"]; ok {
		apiObject.S3Destination = expandReportGroupS3ReportExportConfig(v.([]interface{}))
	}

	if v, ok := tfMap[names.AttrType]; ok {
		apiObject.ExportConfigType = types.ReportExportConfigType(v.(string))
	}

	return apiObject
}

func flattenReportGroupExportConfig(apiObject *types.ReportExportConfig) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"s3_destination": flattenReportGroupS3ReportExportConfig(apiObject.S3Destination),
		names.AttrType:   apiObject.ExportConfigType,
	}

	return []map[string]interface{}{tfMap}
}

func expandReportGroupS3ReportExportConfig(tfList []interface{}) *types.S3ReportExportConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &types.S3ReportExportConfig{}

	if v, ok := tfMap[names.AttrBucket]; ok {
		apiObject.Bucket = aws.String(v.(string))
	}

	if v, ok := tfMap["encryption_disabled"]; ok {
		apiObject.EncryptionDisabled = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["encryption_key"]; ok {
		apiObject.EncryptionKey = aws.String(v.(string))
	}

	if v, ok := tfMap["packaging"]; ok {
		apiObject.Packaging = types.ReportPackagingType(v.(string))
	}

	if v, ok := tfMap[names.AttrPath]; ok {
		apiObject.Path = aws.String(v.(string))
	}

	return apiObject
}

func flattenReportGroupS3ReportExportConfig(apiObject *types.S3ReportExportConfig) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"packaging": apiObject.Packaging,
	}

	if v := apiObject.Bucket; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.EncryptionDisabled; v != nil {
		tfMap["encryption_disabled"] = aws.ToBool(v)
	}

	if v := apiObject.EncryptionKey; v != nil {
		tfMap["encryption_key"] = aws.ToString(v)
	}

	if v := apiObject.Path; v != nil {
		tfMap[names.AttrPath] = aws.ToString(v)
	}

	return []map[string]interface{}{tfMap}
}
