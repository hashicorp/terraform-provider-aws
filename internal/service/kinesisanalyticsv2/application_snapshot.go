// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_kinesisanalyticsv2_application_snapshot")
func ResourceApplicationSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationSnapshotCreate,
		ReadWithoutTimeout:   resourceApplicationSnapshotRead,
		DeleteWithoutTimeout: resourceApplicationSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"application_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},

			"application_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"snapshot_creation_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"snapshot_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},
		},
	}
}

func resourceApplicationSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Conn(ctx)
	applicationName := d.Get("application_name").(string)
	snapshotName := d.Get("snapshot_name").(string)

	input := &kinesisanalyticsv2.CreateApplicationSnapshotInput{
		ApplicationName: aws.String(applicationName),
		SnapshotName:    aws.String(snapshotName),
	}

	log.Printf("[DEBUG] Creating Kinesis Analytics v2 Application Snapshot: %s", input)

	_, err := conn.CreateApplicationSnapshotWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Analytics v2 Application Snapshot (%s/%s): %s", applicationName, snapshotName, err)
	}

	d.SetId(applicationSnapshotCreateID(applicationName, snapshotName))

	_, err = waitSnapshotCreated(ctx, conn, applicationName, snapshotName, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application Snapshot (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceApplicationSnapshotRead(ctx, d, meta)...)
}

func resourceApplicationSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Conn(ctx)

	applicationName, snapshotName, err := applicationSnapshotParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics v2 Application Snapshot (%s): %s", d.Id(), err)
	}

	snapshot, err := FindSnapshotDetailsByApplicationAndSnapshotNames(ctx, conn, applicationName, snapshotName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Analytics v2 Application Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Analytics v2 Application Snapshot (%s): %s", d.Id(), err)
	}

	d.Set("application_name", applicationName)
	d.Set("application_version_id", snapshot.ApplicationVersionId)
	d.Set("snapshot_creation_timestamp", aws.TimeValue(snapshot.SnapshotCreationTimestamp).Format(time.RFC3339))
	d.Set("snapshot_name", snapshotName)

	return diags
}

func resourceApplicationSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Conn(ctx)

	applicationName, snapshotName, err := applicationSnapshotParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application Snapshot (%s): %s", d.Id(), err)
	}

	snapshotCreationTimestamp, err := time.Parse(time.RFC3339, d.Get("snapshot_creation_timestamp").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing snapshot_creation_timestamp: %s", err)
	}

	log.Printf("[DEBUG] Deleting Kinesis Analytics v2 Application Snapshot (%s)", d.Id())
	_, err = conn.DeleteApplicationSnapshotWithContext(ctx, &kinesisanalyticsv2.DeleteApplicationSnapshotInput{
		ApplicationName:           aws.String(applicationName),
		SnapshotCreationTimestamp: aws.Time(snapshotCreationTimestamp),
		SnapshotName:              aws.String(snapshotName),
	})

	if tfawserr.ErrCodeEquals(err, kinesisanalyticsv2.ErrCodeResourceNotFoundException) {
		return diags
	}

	if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application Snapshot (%s): %s", d.Id(), err)
	}

	_, err = waitSnapshotDeleted(ctx, conn, applicationName, snapshotName, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application Snapshot (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}
