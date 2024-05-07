// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dynamodb_table_export")
func ResourceTableExport() *schema.Resource {
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(dynamodb.ExportFormat_Values(), false),
				ForceNew:     true,
				Default:      dynamodb.ExportFormatDynamodbJson,
			},
			"export_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"export_time": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidUTCTimestamp,
				ForceNew:     true,
			},
			"item_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"manifest_files_s3_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"s3_bucket_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
				ForceNew:     true,
			},
			"s3_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
				ForceNew:     true,
			},
			"s3_sse_algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(dynamodb.S3SseAlgorithm_Values(), false),
				ForceNew:     true,
			},
			"s3_sse_kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
				ForceNew:     true,
			},
			"start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"table_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	ResNameTableExport = "Table Export"
)

func resourceTableExportCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	s3Bucket := d.Get("s3_bucket").(string)
	tableArn := d.Get("table_arn").(string)

	in := &dynamodb.ExportTableToPointInTimeInput{
		S3Bucket: aws.String(s3Bucket),
		TableArn: aws.String(tableArn),
	}

	if v, ok := d.GetOk("export_format"); ok {
		in.ExportFormat = aws.String(v.(string))
	}
	if v, ok := d.GetOk("export_time"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		in.ExportTime = aws.Time(v)
	}

	if v, ok := d.GetOk("s3_bucket_owner"); ok {
		in.S3BucketOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("s3_sse_algorithm"); ok {
		in.S3SseAlgorithm = aws.String(v.(string))
	}

	if v, ok := d.GetOk("s3_prefix"); ok {
		in.S3Prefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("s3_sse_kms_key_id"); ok {
		in.S3SseKmsKeyId = aws.String(v.(string))
	}

	log.Printf("Creating export table: %s", in)

	out, err := conn.ExportTableToPointInTimeWithContext(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, ResNameTableExport, d.Get("table_arn").(string), err)
	}

	if out == nil || out.ExportDescription == nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, ResNameTableExport, d.Get("table_arn").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.ExportDescription.ExportArn))

	if _, err := waitTableExportCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, ResNameTableExport, d.Id(), err)
	}

	return append(diags, resourceTableExportRead(ctx, d, meta)...)
}

func resourceTableExportRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	out, err := FindTableExportByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DynamoDB TableExport (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionReading, ResNameTableExport, d.Id(), err)
	}
	desc := out.ExportDescription

	d.Set(names.AttrARN, desc.ExportArn)
	d.Set("billed_size_in_bytes", desc.BilledSizeBytes)
	d.Set("item_count", desc.ItemCount)
	d.Set("manifest_files_s3_key", desc.ExportManifest)
	d.Set("table_arn", desc.TableArn)
	d.Set("s3_bucket", desc.S3Bucket)
	d.Set("s3_bucket_owner", desc.S3BucketOwner)
	d.Set("export_format", desc.ExportFormat)
	d.Set("s3_prefix", desc.S3Prefix)
	d.Set("s3_sse_algorithm", desc.S3SseAlgorithm)
	d.Set("s3_sse_kms_key_id", desc.S3SseKmsKeyId)
	d.Set("export_status", desc.ExportStatus)
	if desc.EndTime != nil {
		d.Set("end_time", aws.TimeValue(desc.EndTime).Format(time.RFC3339))
	}
	if desc.StartTime != nil {
		d.Set("start_time", aws.TimeValue(desc.StartTime).Format(time.RFC3339))
	}
	if desc.ExportTime != nil {
		d.Set("export_time", aws.TimeValue(desc.ExportTime).Format(time.RFC3339))
	}

	return diags
}
