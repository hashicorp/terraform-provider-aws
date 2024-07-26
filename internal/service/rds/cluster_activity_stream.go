// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_cluster_activity_stream")
func ResourceClusterActivityStream() *schema.Resource {
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(rds.ActivityStreamMode_Values(), false),
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

func resourceClusterActivityStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	arn := d.Get(names.AttrResourceARN).(string)
	input := &rds.StartActivityStreamInput{
		ApplyImmediately:                aws.Bool(true),
		EngineNativeAuditFieldsIncluded: aws.Bool(d.Get("engine_native_audit_fields_included").(bool)),
		KmsKeyId:                        aws.String(d.Get(names.AttrKMSKeyID).(string)),
		Mode:                            aws.String(d.Get(names.AttrMode).(string)),
		ResourceArn:                     aws.String(arn),
	}

	_, err := conn.StartActivityStreamWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Cluster Activity Stream (%s): %s", arn, err)
	}

	d.SetId(arn)

	if err := waitActivityStreamStarted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Activity Stream (%s) start: %s", d.Id(), err)
	}

	return append(diags, resourceClusterActivityStreamRead(ctx, d, meta)...)
}

func resourceClusterActivityStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	output, err := FindDBClusterWithActivityStream(ctx, conn, d.Id())

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

func resourceClusterActivityStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	log.Printf("[DEBUG] Deleting RDS Cluster Activity Stream: %s", d.Id())
	_, err := conn.StopActivityStreamWithContext(ctx, &rds.StopActivityStreamInput{
		ApplyImmediately: aws.Bool(true),
		ResourceArn:      aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "stopping RDS Cluster Activity Stream (%s): %s", d.Id(), err)
	}

	if err := waitActivityStreamStopped(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Activity Stream (%s) stop: %s", d.Id(), err)
	}

	return diags
}

func FindDBClusterWithActivityStream(ctx context.Context, conn *rds.RDS, arn string) (*rds.DBCluster, error) {
	output, err := FindDBClusterByID(ctx, conn, arn)
	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.ActivityStreamStatus); status == rds.ActivityStreamStatusStopped {
		return nil, &retry.NotFoundError{
			Message: status,
		}
	}

	return output, nil
}

func statusDBClusterActivityStream(ctx context.Context, conn *rds.RDS, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBClusterWithActivityStream(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ActivityStreamStatus), nil
	}
}

const (
	dbClusterActivityStreamStartedTimeout = 30 * time.Minute
	dbClusterActivityStreamStoppedTimeout = 30 * time.Minute
)

func waitActivityStreamStarted(ctx context.Context, conn *rds.RDS, arn string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{rds.ActivityStreamStatusStarting},
		Target:     []string{rds.ActivityStreamStatusStarted},
		Refresh:    statusDBClusterActivityStream(ctx, conn, arn),
		Timeout:    dbClusterActivityStreamStartedTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitActivityStreamStopped(ctx context.Context, conn *rds.RDS, arn string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{rds.ActivityStreamStatusStopping},
		Target:     []string{},
		Refresh:    statusDBClusterActivityStream(ctx, conn, arn),
		Timeout:    dbClusterActivityStreamStoppedTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
