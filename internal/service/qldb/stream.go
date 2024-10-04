// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qldb

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qldb"
	"github.com/aws/aws-sdk-go-v2/service/qldb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_qldb_stream", name="Stream")
// @Tags(identifierAttribute="arn")
func resourceStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamCreate,
		ReadWithoutTimeout:   resourceStreamRead,
		UpdateWithoutTimeout: resourceStreamUpdate,
		DeleteWithoutTimeout: resourceStreamDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(8 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exclusive_end_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"inclusive_start_time": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"kinesis_configuration": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aggregation_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						names.AttrStreamARN: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"ledger_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
				),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"stream_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	ledgerName := d.Get("ledger_name").(string)
	name := d.Get("stream_name").(string)
	input := &qldb.StreamJournalToKinesisInput{
		LedgerName: aws.String(ledgerName),
		RoleArn:    aws.String(d.Get(names.AttrRoleARN).(string)),
		StreamName: aws.String(name),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk("exclusive_end_time"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		input.ExclusiveEndTime = aws.Time(v)
	}

	if v, ok := d.GetOk("inclusive_start_time"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))
		input.InclusiveStartTime = aws.Time(v)
	}

	if v, ok := d.GetOk("kinesis_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.KinesisConfiguration = expandKinesisConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.StreamJournalToKinesis(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QLDB Stream (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.StreamId))

	if _, err := waitStreamCreated(ctx, conn, ledgerName, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QLDB Stream (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStreamRead(ctx, d, meta)...)
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	ledgerName := d.Get("ledger_name").(string)
	stream, err := findStreamByTwoPartKey(ctx, conn, ledgerName, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QLDB Stream %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QLDB Stream (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, stream.Arn)
	if stream.ExclusiveEndTime != nil {
		d.Set("exclusive_end_time", aws.ToTime(stream.ExclusiveEndTime).Format(time.RFC3339))
	} else {
		d.Set("exclusive_end_time", nil)
	}
	if stream.InclusiveStartTime != nil {
		d.Set("inclusive_start_time", aws.ToTime(stream.InclusiveStartTime).Format(time.RFC3339))
	} else {
		d.Set("inclusive_start_time", nil)
	}
	if stream.KinesisConfiguration != nil {
		if err := d.Set("kinesis_configuration", []interface{}{flattenKinesisConfiguration(stream.KinesisConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting kinesis_configuration: %s", err)
		}
	} else {
		d.Set("kinesis_configuration", nil)
	}
	d.Set("ledger_name", stream.LedgerName)
	d.Set(names.AttrRoleARN, stream.RoleArn)
	d.Set("stream_name", stream.StreamName)

	return diags
}

func resourceStreamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceStreamRead(ctx, d, meta)
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)

	ledgerName := d.Get("ledger_name").(string)
	input := &qldb.CancelJournalKinesisStreamInput{
		LedgerName: aws.String(ledgerName),
		StreamId:   aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting QLDB Stream: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.CancelJournalKinesisStream(ctx, input)
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QLDB Stream (%s): %s", d.Id(), err)
	}

	if _, err := waitStreamDeleted(ctx, conn, ledgerName, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for QLDB Stream (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findStreamByTwoPartKey(ctx context.Context, conn *qldb.Client, ledgerName, streamID string) (*types.JournalKinesisStreamDescription, error) {
	input := &qldb.DescribeJournalKinesisStreamInput{
		LedgerName: aws.String(ledgerName),
		StreamId:   aws.String(streamID),
	}

	output, err := findJournalKinesisStream(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/qldb/latest/developerguide/streams.create.html#streams.create.states.
	switch status := output.Status; status {
	case types.StreamStatusCompleted, types.StreamStatusCanceled, types.StreamStatusFailed:
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findJournalKinesisStream(ctx context.Context, conn *qldb.Client, input *qldb.DescribeJournalKinesisStreamInput) (*types.JournalKinesisStreamDescription, error) {
	output, err := conn.DescribeJournalKinesisStream(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Stream == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Stream, nil
}

func statusStreamCreated(ctx context.Context, conn *qldb.Client, ledgerName, streamID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindStream as it maps useful statuses to NotFoundError.
		output, err := findJournalKinesisStream(ctx, conn, &qldb.DescribeJournalKinesisStreamInput{
			LedgerName: aws.String(ledgerName),
			StreamId:   aws.String(streamID),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitStreamCreated(ctx context.Context, conn *qldb.Client, ledgerName, streamID string, timeout time.Duration) (*types.JournalKinesisStreamDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.StreamStatusImpaired),
		Target:     enum.Slice(types.StreamStatusActive),
		Refresh:    statusStreamCreated(ctx, conn, ledgerName, streamID),
		Timeout:    timeout,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.JournalKinesisStreamDescription); ok {
		tfresource.SetLastError(err, errors.New(string(output.ErrorCause)))

		return output, err
	}

	return nil, err
}

func statusStreamDeleted(ctx context.Context, conn *qldb.Client, ledgerName, streamID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findStreamByTwoPartKey(ctx, conn, ledgerName, streamID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitStreamDeleted(ctx context.Context, conn *qldb.Client, ledgerName, streamID string, timeout time.Duration) (*types.JournalKinesisStreamDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.StreamStatusActive, types.StreamStatusImpaired),
		Target:     []string{},
		Refresh:    statusStreamDeleted(ctx, conn, ledgerName, streamID),
		Timeout:    timeout,
		MinTimeout: 1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.JournalKinesisStreamDescription); ok {
		tfresource.SetLastError(err, errors.New(string(output.ErrorCause)))

		return output, err
	}

	return nil, err
}

func expandKinesisConfiguration(tfMap map[string]interface{}) *types.KinesisConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.KinesisConfiguration{}

	if v, ok := tfMap["aggregation_enabled"].(bool); ok {
		apiObject.AggregationEnabled = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrStreamARN].(string); ok && v != "" {
		apiObject.StreamArn = aws.String(v)
	}

	return apiObject
}

func flattenKinesisConfiguration(apiObject *types.KinesisConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AggregationEnabled; v != nil {
		tfMap["aggregation_enabled"] = aws.ToBool(v)
	}

	if v := apiObject.StreamArn; v != nil {
		tfMap[names.AttrStreamARN] = aws.ToString(v)
	}

	return tfMap
}
