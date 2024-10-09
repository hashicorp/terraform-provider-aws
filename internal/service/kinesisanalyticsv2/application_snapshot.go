// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_kinesisanalyticsv2_application_snapshot", name="Application Snapshot")
func resourceApplicationSnapshot() *schema.Resource {
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
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
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
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},
		},
	}
}

func resourceApplicationSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	applicationName := d.Get("application_name").(string)
	snapshotName := d.Get("snapshot_name").(string)
	id := applicationSnapshotCreateResourceID(applicationName, snapshotName)
	input := &kinesisanalyticsv2.CreateApplicationSnapshotInput{
		ApplicationName: aws.String(applicationName),
		SnapshotName:    aws.String(snapshotName),
	}

	_, err := conn.CreateApplicationSnapshot(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Analytics v2 Application Snapshot (%s): %s", id, err)
	}

	d.SetId(id)

	snapshot, err := waitSnapshotCreated(ctx, conn, applicationName, snapshotName, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		if snapshot != nil && snapshot.SnapshotCreationTimestamp != nil {
			// SnapshotCreationTimestamp is required for deletion, so persist to state now in case of subsequent errors and destroy being called without refresh.
			d.Set("snapshot_creation_timestamp", aws.ToTime(snapshot.SnapshotCreationTimestamp).Format(time.RFC3339))
		}

		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application Snapshot (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceApplicationSnapshotRead(ctx, d, meta)...)
}

func resourceApplicationSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	applicationName, snapshotName, err := applicationSnapshotParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	snapshot, err := findSnapshotDetailsByTwoPartKey(ctx, conn, applicationName, snapshotName)

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
	d.Set("snapshot_creation_timestamp", aws.ToTime(snapshot.SnapshotCreationTimestamp).Format(time.RFC3339))
	d.Set("snapshot_name", snapshotName)

	return diags
}

func resourceApplicationSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisAnalyticsV2Client(ctx)

	applicationName, snapshotName, err := applicationSnapshotParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	snapshotCreationTimestamp, err := time.Parse(time.RFC3339, d.Get("snapshot_creation_timestamp").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing snapshot_creation_timestamp: %s", err)
	}

	log.Printf("[DEBUG] Deleting Kinesis Analytics v2 Application Snapshot (%s)", d.Id())
	_, err = conn.DeleteApplicationSnapshot(ctx, &kinesisanalyticsv2.DeleteApplicationSnapshotInput{
		ApplicationName:           aws.String(applicationName),
		SnapshotCreationTimestamp: aws.Time(snapshotCreationTimestamp),
		SnapshotName:              aws.String(snapshotName),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Analytics v2 Application Snapshot (%s): %s", d.Id(), err)
	}

	if _, err := waitSnapshotDeleted(ctx, conn, applicationName, snapshotName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Analytics v2 Application Snapshot (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const applicationSnapshotIDSeparator = "/"

func applicationSnapshotCreateResourceID(applicationName, snapshotName string) string {
	parts := []string{applicationName, snapshotName}
	id := strings.Join(parts, applicationSnapshotIDSeparator)

	return id
}

func applicationSnapshotParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, applicationSnapshotIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected application-name%[2]ssnapshot-name", id, applicationSnapshotIDSeparator)
}

func findSnapshotDetailsByTwoPartKey(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, snapshotName string) (*awstypes.SnapshotDetails, error) {
	input := &kinesisanalyticsv2.DescribeApplicationSnapshotInput{
		ApplicationName: aws.String(applicationName),
		SnapshotName:    aws.String(snapshotName),
	}

	return findSnapshotDetails(ctx, conn, input)
}

func findSnapshotDetails(ctx context.Context, conn *kinesisanalyticsv2.Client, input *kinesisanalyticsv2.DescribeApplicationSnapshotInput) (*awstypes.SnapshotDetails, error) {
	output, err := conn.DescribeApplicationSnapshot(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SnapshotDetails == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SnapshotDetails, nil
}

func statusSnapshotDetails(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, snapshotName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSnapshotDetailsByTwoPartKey(ctx, conn, applicationName, snapshotName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.SnapshotStatus), nil
	}
}

func waitSnapshotCreated(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, snapshotName string, timeout time.Duration) (*awstypes.SnapshotDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotStatusCreating),
		Target:  enum.Slice(awstypes.SnapshotStatusReady),
		Refresh: statusSnapshotDetails(ctx, conn, applicationName, snapshotName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SnapshotDetails); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotDeleted(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, snapshotName string, timeout time.Duration) (*awstypes.SnapshotDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotStatusDeleting),
		Target:  []string{},
		Refresh: statusSnapshotDetails(ctx, conn, applicationName, snapshotName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SnapshotDetails); ok {
		return output, err
	}

	return nil, err
}
