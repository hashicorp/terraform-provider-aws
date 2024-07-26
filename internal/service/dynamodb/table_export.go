// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dynamodb_table_export", name="Table Export")
func resourceTableExport() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableExportCreate,
		ReadWithoutTimeout:   resourceTableExportRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"billed_size_in_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"end_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"export_format": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.ExportFormatDynamodbJson,
				ValidateDiagFunc: enum.Validate[awstypes.ExportFormat](),
			},
			"export_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"export_time": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidUTCTimestamp,
			},
			"item_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"manifest_files_s3_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrS3Bucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"s3_bucket_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"s3_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"s3_sse_algorithm": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.S3SseAlgorithm](),
			},
			"s3_sse_kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			names.AttrStartTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"table_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceTableExportCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	s3Bucket := d.Get(names.AttrS3Bucket).(string)
	tableARN := d.Get("table_arn").(string)
	input := &dynamodb.ExportTableToPointInTimeInput{
		S3Bucket: aws.String(s3Bucket),
		TableArn: aws.String(tableARN),
	}

	if v, ok := d.GetOk("export_format"); ok {
		input.ExportFormat = awstypes.ExportFormat(v.(string))
	}

	if v, ok := d.GetOk("export_time"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		input.ExportTime = aws.Time(v)
	}

	if v, ok := d.GetOk("s3_bucket_owner"); ok {
		input.S3BucketOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("s3_prefix"); ok {
		input.S3Prefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("s3_sse_algorithm"); ok {
		input.S3SseAlgorithm = awstypes.S3SseAlgorithm(v.(string))
	}

	if v, ok := d.GetOk("s3_sse_kms_key_id"); ok {
		input.S3SseKmsKeyId = aws.String(v.(string))
	}

	output, err := conn.ExportTableToPointInTime(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "exporting DynamoDB Table (%s): %s", tableARN, err)
	}

	d.SetId(aws.ToString(output.ExportDescription.ExportArn))

	if _, err := waitTableExportCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Table Export (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTableExportRead(ctx, d, meta)...)
}

func resourceTableExportRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	desc, err := findTableExportByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DynamoDB Table Export (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Table Export (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, desc.ExportArn)
	d.Set("billed_size_in_bytes", desc.BilledSizeBytes)
	if desc.EndTime != nil {
		d.Set("end_time", aws.ToTime(desc.EndTime).Format(time.RFC3339))
	}
	d.Set("export_format", desc.ExportFormat)
	d.Set("export_status", desc.ExportStatus)
	if desc.ExportTime != nil {
		d.Set("export_time", aws.ToTime(desc.ExportTime).Format(time.RFC3339))
	}
	d.Set("item_count", desc.ItemCount)
	d.Set("manifest_files_s3_key", desc.ExportManifest)
	d.Set(names.AttrS3Bucket, desc.S3Bucket)
	d.Set("s3_bucket_owner", desc.S3BucketOwner)
	d.Set("s3_prefix", desc.S3Prefix)
	d.Set("s3_sse_algorithm", desc.S3SseAlgorithm)
	d.Set("s3_sse_kms_key_id", desc.S3SseKmsKeyId)
	if desc.StartTime != nil {
		d.Set(names.AttrStartTime, aws.ToTime(desc.StartTime).Format(time.RFC3339))
	}
	d.Set("table_arn", desc.TableArn)

	return diags
}

func findTableExportByARN(ctx context.Context, conn *dynamodb.Client, arn string) (*awstypes.ExportDescription, error) {
	input := &dynamodb.DescribeExportInput{
		ExportArn: aws.String(arn),
	}

	output, err := conn.DescribeExport(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || output.ExportDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ExportDescription, nil
}

func statusTableExport(ctx context.Context, conn *dynamodb.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTableExportByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ExportStatus), nil
	}
}

func waitTableExportCreated(ctx context.Context, conn *dynamodb.Client, id string, timeout time.Duration) (*awstypes.ExportDescription, error) {
	const (
		maxTimeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ExportStatusInProgress),
		Target:  enum.Slice(awstypes.ExportStatusCompleted, awstypes.ExportStatusFailed),
		Refresh: statusTableExport(ctx, conn, id),
		Timeout: max(maxTimeout, timeout),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ExportDescription); ok {
		return output, err
	}

	return nil, err
}
