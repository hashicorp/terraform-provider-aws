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

// @SDKResource("aws_db_activity_stream", name="Activity Stream")
func resourceActivityStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceActivityStreamCreate,
		ReadWithoutTimeout:   resourceActivityStreamRead,
		DeleteWithoutTimeout: resourceActivityStreamDelete,

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
				ValidateFunc: verify.ValidARNCheck(validDBInstanceARN),
			},
		},
	}
}

func resourceActivityStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return sdkdiag.AppendErrorf(diags, "creating DB Activity Stream (%s): %s", arn, err)
	}

	d.SetId(arn)

	if _, err := waitActivityStreamStarted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DB Activity Stream (%s) start: %s", d.Id(), err)
	}

	return append(diags, resourceActivityStreamRead(ctx, d, meta)...)
}

func resourceActivityStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	output, err := findDBInstanceWithActivityStream(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DB Activity Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DB Activity Stream (%s): %s", d.Id(), err)
	}

	d.Set("engine_native_audit_fields_included", output.ActivityStreamEngineNativeAuditFieldsIncluded)
	d.Set("kinesis_stream_name", output.ActivityStreamKinesisStreamName)
	d.Set(names.AttrKMSKeyID, output.ActivityStreamKmsKeyId)
	d.Set(names.AttrMode, output.ActivityStreamMode)
	d.Set(names.AttrResourceARN, output.DBInstanceArn)

	return diags
}

func resourceActivityStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting DB Activity Stream: %s", d.Id())
	_, err := conn.StopActivityStream(ctx, &rds.StopActivityStreamInput{
		ApplyImmediately: aws.Bool(true),
		ResourceArn:      aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterCombination, "Activity Streams feature expected to be STARTED, but is STOPPED") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "stopping DB Activity Stream (%s): %s", d.Id(), err)
	}

	if _, err := waitActivityStreamStopped(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DB Activity Stream (%s) stop: %s", d.Id(), err)
	}

	return diags
}

func findDBInstanceWithActivityStream(ctx context.Context, conn *rds.Client, arn string) (*types.DBInstance, error) {
	output, err := findDBInstanceByID(ctx, conn, arn)

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

func statusActivityStream(ctx context.Context, conn *rds.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBInstanceByID(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ActivityStreamStatus), nil
	}
}

func waitActivityStreamStarted(ctx context.Context, conn *rds.Client, arn string) (*types.DBInstance, error) {
	const (
		timeout = 30 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ActivityStreamStatusStarting),
		Target:     enum.Slice(types.ActivityStreamStatusStarted),
		Refresh:    statusActivityStream(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitActivityStreamStopped(ctx context.Context, conn *rds.Client, arn string) (*types.DBInstance, error) {
	const (
		timeout = 30 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ActivityStreamStatusStopping),
		Target:     enum.Slice(types.ActivityStreamStatusStopped),
		Refresh:    statusActivityStream(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBInstance); ok {
		return output, err
	}

	return nil, err
}
