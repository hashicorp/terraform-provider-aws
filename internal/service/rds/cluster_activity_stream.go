// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_cluster_activity_stream", name="Cluster Activity Stream")
func resourceClusterActivityStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterActivityStreamCreate,
		ReadWithoutTimeout:   resourceClusterActivityStreamRead,
		DeleteWithoutTimeout: resourceClusterActivityStreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"engine_native_audit_fields_included": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"kinesis_stream_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrMode: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ActivityStreamMode](),
			},
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceClusterActivityStreamCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	arn := d.Get(names.AttrResourceARN).(string)
	input := &rds.StartActivityStreamInput{
		ApplyImmediately:                aws.Bool(true),
		EngineNativeAuditFieldsIncluded: aws.Bool(d.Get("engine_native_audit_fields_included").(bool)),
		KmsKeyId:                        aws.String(d.Get(names.AttrKMSKeyID).(string)),
		Mode:                            types.ActivityStreamMode(d.Get(names.AttrMode).(string)),
		ResourceArn:                     aws.String(arn),
	}

	_, err := conn.StartActivityStream(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Cluster Activity Stream (%s): %s", arn, err)
	}

	d.SetId(arn)

	if _, err := waitActivityStreamStarted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Activity Stream (%s) start: %s", d.Id(), err)
	}

	return append(diags, resourceClusterActivityStreamRead(ctx, d, meta)...)
}

func resourceClusterActivityStreamRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	output, err := findDBClusterWithActivityStream(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Cluster Activity Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Activity Stream (%s): %s", d.Id(), err)
	}

	d.Set("kinesis_stream_name", output.ActivityStreamKinesisStreamName)
	d.Set(names.AttrKMSKeyID, output.ActivityStreamKmsKeyId)
	d.Set(names.AttrMode, output.ActivityStreamMode)
	d.Set(names.AttrResourceARN, output.DBClusterArn)

	return diags
}

func resourceClusterActivityStreamDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting RDS Cluster Activity Stream: %s", d.Id())
	_, err := conn.StopActivityStream(ctx, &rds.StopActivityStreamInput{
		ApplyImmediately: aws.Bool(true),
		ResourceArn:      aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterCombination, "Activity Streams feature expected to be started, but is stopped") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "stopping RDS Cluster Activity Stream (%s): %s", d.Id(), err)
	}

	if _, err := waitActivityStreamStopped(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Activity Stream (%s) stop: %s", d.Id(), err)
	}

	return diags
}

func findDBClusterWithActivityStream(ctx context.Context, conn *rds.Client, arn string) (*types.DBCluster, error) {
	output, err := findDBClusterByID(ctx, conn, arn)

	if err != nil {
		return nil, err
	}

	if status := output.ActivityStreamStatus; status == types.ActivityStreamStatusStopped {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func statusDBClusterActivityStream(ctx context.Context, conn *rds.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDBClusterWithActivityStream(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ActivityStreamStatus), nil
	}
}

func waitActivityStreamStarted(ctx context.Context, conn *rds.Client, arn string) (*types.DBCluster, error) {
	const (
		timeout = 30 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ActivityStreamStatusStarting),
		Target:     enum.Slice(types.ActivityStreamStatusStarted),
		Refresh:    statusDBClusterActivityStream(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitActivityStreamStopped(ctx context.Context, conn *rds.Client, arn string) (*types.DBCluster, error) {
	const (
		timeout = 30 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ActivityStreamStatusStopping),
		Target:     []string{},
		Refresh:    statusDBClusterActivityStream(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBCluster); ok {
		return output, err
	}

	return nil, err
}
