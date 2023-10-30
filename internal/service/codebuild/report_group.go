// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codebuild_report_group", name="Report Group")
// @Tags
func ResourceReportGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReportGroupCreate,
		ReadWithoutTimeout:   resourceReportGroupRead,
		UpdateWithoutTimeout: resourceReportGroupUpdate,
		DeleteWithoutTimeout: resourceReportGroupDelete,

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
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 128),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(codebuild.ReportType_Values(), false),
			},
			"export_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codebuild.ReportExportConfigType_Values(), false),
						},
						"s3_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
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
										Type:         schema.TypeString,
										Optional:     true,
										Default:      codebuild.ReportPackagingTypeNone,
										ValidateFunc: validation.StringInSlice(codebuild.ReportPackagingType_Values(), false),
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReportGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	input := &codebuild.CreateReportGroupInput{
		Name:         aws.String(d.Get("name").(string)),
		Type:         aws.String(d.Get("type").(string)),
		ExportConfig: expandReportGroupExportConfig(d.Get("export_config").([]interface{})),
		Tags:         getTagsIn(ctx),
	}

	resp, err := conn.CreateReportGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeBuild Report Group: %s", err)
	}

	d.SetId(aws.StringValue(resp.ReportGroup.Arn))

	return append(diags, resourceReportGroupRead(ctx, d, meta)...)
}

func resourceReportGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	reportGroup, err := FindReportGroupByARN(ctx, conn, d.Id())
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codebuild.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CodeBuild, create.ErrActionReading, ResNameReportGroup, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CodeBuild, create.ErrActionReading, ResNameReportGroup, d.Id(), err)
	}

	if !d.IsNewResource() && reportGroup == nil {
		create.LogNotFoundRemoveState(names.CodeBuild, create.ErrActionReading, ResNameReportGroup, d.Id())
		d.SetId("")
		return diags
	}

	if reportGroup == nil {
		return create.DiagError(names.CodeBuild, create.ErrActionReading, ResNameReportGroup, d.Id(), errors.New("not found after creation"))
	}

	d.Set("arn", reportGroup.Arn)
	d.Set("type", reportGroup.Type)
	d.Set("name", reportGroup.Name)

	if err := d.Set("created", reportGroup.Created.Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting created: %s", err)
	}

	if err := d.Set("export_config", flattenReportGroupExportConfig(reportGroup.ExportConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting export config: %s", err)
	}

	setTagsOut(ctx, reportGroup.Tags)

	return diags
}

func resourceReportGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	input := &codebuild.UpdateReportGroupInput{
		Arn: aws.String(d.Id()),
	}

	if d.HasChange("export_config") {
		input.ExportConfig = expandReportGroupExportConfig(d.Get("export_config").([]interface{}))
	}

	if d.HasChange("tags_all") {
		input.Tags = getTagsIn(ctx)
	}

	_, err := conn.UpdateReportGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CodeBuild Report Group: %s", err)
	}

	return append(diags, resourceReportGroupRead(ctx, d, meta)...)
}

func resourceReportGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	deleteOpts := &codebuild.DeleteReportGroupInput{
		Arn:           aws.String(d.Id()),
		DeleteReports: aws.Bool(d.Get("delete_reports").(bool)),
	}

	if _, err := conn.DeleteReportGroupWithContext(ctx, deleteOpts); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Report Group (%s): %s", d.Id(), err)
	}

	if _, err := waitReportGroupDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "while waiting for CodeBuild Report Group (%s) to become deleted: %s", d.Id(), err)
	}

	return diags
}

func expandReportGroupExportConfig(config []interface{}) *codebuild.ReportExportConfig {
	if len(config) == 0 {
		return nil
	}

	s := config[0].(map[string]interface{})
	exportConfig := &codebuild.ReportExportConfig{}

	if v, ok := s["type"]; ok {
		exportConfig.ExportConfigType = aws.String(v.(string))
	}

	if v, ok := s["s3_destination"]; ok {
		exportConfig.S3Destination = expandReportGroupS3ReportExportConfig(v.([]interface{}))
	}

	return exportConfig
}

func flattenReportGroupExportConfig(config *codebuild.ReportExportConfig) []map[string]interface{} {
	settings := make(map[string]interface{})

	if config == nil {
		return nil
	}

	settings["s3_destination"] = flattenReportGroupS3ReportExportConfig(config.S3Destination)
	settings["type"] = aws.StringValue(config.ExportConfigType)

	return []map[string]interface{}{settings}
}

func expandReportGroupS3ReportExportConfig(config []interface{}) *codebuild.S3ReportExportConfig {
	if len(config) == 0 {
		return nil
	}

	s := config[0].(map[string]interface{})
	s3ReportExportConfig := &codebuild.S3ReportExportConfig{}

	if v, ok := s["bucket"]; ok {
		s3ReportExportConfig.Bucket = aws.String(v.(string))
	}
	if v, ok := s["encryption_disabled"]; ok {
		s3ReportExportConfig.EncryptionDisabled = aws.Bool(v.(bool))
	}

	if v, ok := s["encryption_key"]; ok {
		s3ReportExportConfig.EncryptionKey = aws.String(v.(string))
	}

	if v, ok := s["packaging"]; ok {
		s3ReportExportConfig.Packaging = aws.String(v.(string))
	}

	if v, ok := s["path"]; ok {
		s3ReportExportConfig.Path = aws.String(v.(string))
	}

	return s3ReportExportConfig
}

func flattenReportGroupS3ReportExportConfig(config *codebuild.S3ReportExportConfig) []map[string]interface{} {
	settings := make(map[string]interface{})

	if config == nil {
		return nil
	}

	settings["path"] = aws.StringValue(config.Path)
	settings["bucket"] = aws.StringValue(config.Bucket)
	settings["packaging"] = aws.StringValue(config.Packaging)
	settings["encryption_disabled"] = aws.BoolValue(config.EncryptionDisabled)

	if config.EncryptionKey != nil {
		settings["encryption_key"] = aws.StringValue(config.EncryptionKey)
	}

	return []map[string]interface{}{settings}
}
