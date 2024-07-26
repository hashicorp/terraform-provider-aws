// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cur

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	cur "github.com/aws/aws-sdk-go-v2/service/costandusagereportservice"
	"github.com/aws/aws-sdk-go-v2/service/costandusagereportservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cur_report_definition", name="Report Definition")
// @Tags(identifierAttribute="report_name")
func resourceReportDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReportDefinitionCreate,
		ReadWithoutTimeout:   resourceReportDefinitionRead,
		UpdateWithoutTimeout: resourceReportDefinitionUpdate,
		DeleteWithoutTimeout: resourceReportDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"additional_artifacts": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{Type: schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.AdditionalArtifact](),
				},
			},
			"additional_schema_elements": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.SchemaElement](),
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compression": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.CompressionFormat](),
			},
			names.AttrFormat: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.ReportFormat](),
			},
			"refresh_closed_reports": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"report_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z!\-_.*\'()]+`), "The name must be unique, is case sensitive, and can't include spaces."),
				),
			},
			"report_versioning": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				Default:          types.ReportVersioningCreateNewReport,
				ValidateDiagFunc: enum.Validate[types.ReportVersioning](),
			},
			names.AttrS3Bucket: {
				Type:     schema.TypeString,
				Required: true,
			},
			"s3_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"s3_region": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.AWSRegion](),
			},
			"time_unit": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.TimeUnit](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReportDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CURClient(ctx)

	reportName := d.Get("report_name").(string)
	additionalArtifacts := flex.ExpandStringyValueSet[types.AdditionalArtifact](d.Get("additional_artifacts").(*schema.Set))
	compression := types.CompressionFormat(d.Get("compression").(string))
	format := types.ReportFormat(d.Get(names.AttrFormat).(string))
	prefix := d.Get("s3_prefix").(string)
	reportVersioning := types.ReportVersioning(d.Get("report_versioning").(string))

	if err := checkReportDefinitionPropertyCombination(
		additionalArtifacts,
		compression,
		format,
		prefix,
		reportVersioning,
	); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &cur.PutReportDefinitionInput{
		ReportDefinition: &types.ReportDefinition{
			AdditionalArtifacts:      additionalArtifacts,
			AdditionalSchemaElements: flex.ExpandStringyValueSet[types.SchemaElement](d.Get("additional_schema_elements").(*schema.Set)),
			Compression:              compression,
			Format:                   format,
			RefreshClosedReports:     aws.Bool(d.Get("refresh_closed_reports").(bool)),
			ReportName:               aws.String(reportName),
			ReportVersioning:         reportVersioning,
			S3Bucket:                 aws.String(d.Get(names.AttrS3Bucket).(string)),
			S3Prefix:                 aws.String(prefix),
			S3Region:                 types.AWSRegion(d.Get("s3_region").(string)),
			TimeUnit:                 types.TimeUnit(d.Get("time_unit").(string)),
		},
		Tags: getTagsIn(ctx),
	}

	_, err := conn.PutReportDefinition(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cost And Usage Report Definition (%s): %s", reportName, err)
	}

	d.SetId(reportName)

	return append(diags, resourceReportDefinitionRead(ctx, d, meta)...)
}

func resourceReportDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CURClient(ctx)

	reportDefinition, err := findReportDefinitionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cost And Usage Report Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cost And Usage Report Definition (%s): %s", d.Id(), err)
	}

	reportName := aws.ToString(reportDefinition.ReportName)
	d.SetId(reportName)
	d.Set("additional_artifacts", reportDefinition.AdditionalArtifacts)
	d.Set("additional_schema_elements", reportDefinition.AdditionalSchemaElements)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.CUR,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "definition/" + reportName,
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("compression", reportDefinition.Compression)
	d.Set(names.AttrFormat, reportDefinition.Format)
	d.Set("refresh_closed_reports", reportDefinition.RefreshClosedReports)
	d.Set("report_name", reportName)
	d.Set("report_versioning", reportDefinition.ReportVersioning)
	d.Set(names.AttrS3Bucket, reportDefinition.S3Bucket)
	d.Set("s3_prefix", reportDefinition.S3Prefix)
	d.Set("s3_region", reportDefinition.S3Region)
	d.Set("time_unit", reportDefinition.TimeUnit)

	return diags
}

func resourceReportDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CURClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		additionalArtifacts := flex.ExpandStringyValueSet[types.AdditionalArtifact](d.Get("additional_artifacts").(*schema.Set))
		compression := types.CompressionFormat(d.Get("compression").(string))
		format := types.ReportFormat(d.Get(names.AttrFormat).(string))
		prefix := d.Get("s3_prefix").(string)
		reportVersioning := types.ReportVersioning(d.Get("report_versioning").(string))

		if err := checkReportDefinitionPropertyCombination(
			additionalArtifacts,
			compression,
			format,
			prefix,
			reportVersioning,
		); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &cur.ModifyReportDefinitionInput{
			ReportDefinition: &types.ReportDefinition{
				AdditionalArtifacts:      additionalArtifacts,
				AdditionalSchemaElements: flex.ExpandStringyValueSet[types.SchemaElement](d.Get("additional_schema_elements").(*schema.Set)),
				Compression:              compression,
				Format:                   format,
				RefreshClosedReports:     aws.Bool(d.Get("refresh_closed_reports").(bool)),
				ReportName:               aws.String(d.Id()),
				ReportVersioning:         reportVersioning,
				S3Bucket:                 aws.String(d.Get(names.AttrS3Bucket).(string)),
				S3Prefix:                 aws.String(prefix),
				S3Region:                 types.AWSRegion(d.Get("s3_region").(string)),
				TimeUnit:                 types.TimeUnit(d.Get("time_unit").(string)),
			},
			ReportName: aws.String(d.Id()),
		}

		_, err := conn.ModifyReportDefinition(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Cost And Usage Report Definition (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceReportDefinitionRead(ctx, d, meta)...)
}

func resourceReportDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CURClient(ctx)

	log.Printf("[DEBUG] Deleting Cost And Usage Report Definition: %s", d.Id())
	_, err := conn.DeleteReportDefinition(ctx, &cur.DeleteReportDefinitionInput{
		ReportName: aws.String(d.Id()),
	})

	if errs.IsA[*types.ValidationException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cost And Usage Report Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func checkReportDefinitionPropertyCombination(additionalArtifacts []types.AdditionalArtifact, compression types.CompressionFormat, format types.ReportFormat, prefix string, reportVersioning types.ReportVersioning) error {
	// perform various combination checks, AWS API unhelpfully just returns an empty ValidationException
	// these combinations have been determined from the Create Report AWS Console Web Form

	if slices.Contains(additionalArtifacts, types.AdditionalArtifactAthena) {
		if len(additionalArtifacts) > 1 {
			return fmt.Errorf(
				"When %s exists within additional_artifacts, no other artifact type can be declared",
				types.AdditionalArtifactAthena,
			)
		}

		if len(prefix) == 0 {
			return fmt.Errorf(
				"When %s exists within additional_artifacts, prefix cannot be empty",
				types.AdditionalArtifactAthena,
			)
		}

		if reportVersioning != types.ReportVersioningOverwriteReport {
			return fmt.Errorf(
				"When %s exists within additional_artifacts, report_versioning must be %s",
				types.AdditionalArtifactAthena,
				types.ReportVersioningOverwriteReport,
			)
		}

		if format != types.ReportFormatParquet {
			return fmt.Errorf(
				"When %s exists within additional_artifacts, both format and compression must be %s",
				types.AdditionalArtifactAthena,
				types.ReportFormatParquet,
			)
		}
	} else if len(additionalArtifacts) > 0 && (format == types.ReportFormatParquet) {
		return fmt.Errorf(
			"When additional_artifacts includes %s and/or %s, format must not be %s",
			types.AdditionalArtifactQuicksight,
			types.AdditionalArtifactRedshift,
			types.ReportFormatParquet,
		)
	}

	if format == types.ReportFormatParquet {
		if compression != types.CompressionFormatParquet {
			return fmt.Errorf(
				"When format is %s, compression must also be %s",
				types.ReportFormatParquet,
				types.CompressionFormatParquet,
			)
		}
	} else {
		if compression == types.CompressionFormatParquet {
			return fmt.Errorf(
				"When format is %s, compression must not be %s",
				format,
				types.CompressionFormatParquet,
			)
		}
	}
	// end checks

	return nil
}

func findReportDefinitionByName(ctx context.Context, conn *cur.Client, name string) (*types.ReportDefinition, error) {
	input := &cur.DescribeReportDefinitionsInput{}

	return findReportDefinition(ctx, conn, input, func(v *types.ReportDefinition) bool {
		return aws.ToString(v.ReportName) == name
	})
}

func findReportDefinition(ctx context.Context, conn *cur.Client, input *cur.DescribeReportDefinitionsInput, filter tfslices.Predicate[*types.ReportDefinition]) (*types.ReportDefinition, error) {
	output, err := findReportDefinitions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReportDefinitions(ctx context.Context, conn *cur.Client, input *cur.DescribeReportDefinitionsInput, filter tfslices.Predicate[*types.ReportDefinition]) ([]types.ReportDefinition, error) {
	var output []types.ReportDefinition

	pages := cur.NewDescribeReportDefinitionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ReportDefinitions {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
