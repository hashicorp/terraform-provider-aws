// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	kinesisStreamingDestinationResourceIDPartCount = 2
)

// @SDKResource("aws_dynamodb_kinesis_streaming_destination", name="Kinesis Streaming Destination")
func resourceKinesisStreamingDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKinesisStreamingDestinationCreate,
		ReadWithoutTimeout:   resourceKinesisStreamingDestinationRead,
		DeleteWithoutTimeout: resourceKinesisStreamingDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrStreamARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTableName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceKinesisStreamingDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	streamARN := d.Get(names.AttrStreamARN).(string)
	tableName := d.Get(names.AttrTableName).(string)
	id := errs.Must(flex.FlattenResourceId([]string{tableName, streamARN}, kinesisStreamingDestinationResourceIDPartCount, false))
	input := &dynamodb.EnableKinesisStreamingDestinationInput{
		StreamArn: aws.String(streamARN),
		TableName: aws.String(tableName),
	}

	_, err := conn.EnableKinesisStreamingDestination(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling DynamoDB Kinesis Streaming Destination (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitKinesisStreamingDestinationActive(ctx, conn, streamARN, tableName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Kinesis Streaming Destination (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceKinesisStreamingDestinationRead(ctx, d, meta)...)
}

func resourceKinesisStreamingDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), kinesisStreamingDestinationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	tableName, streamARN := parts[0], parts[1]
	output, err := findKinesisDataStreamDestinationByTwoPartKey(ctx, conn, streamARN, tableName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DynamoDB Kinesis Streaming Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Kinesis Streaming Destination (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrStreamARN, output.StreamArn)
	d.Set(names.AttrTableName, tableName)

	return diags
}

func resourceKinesisStreamingDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), kinesisStreamingDestinationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	tableName, streamARN := parts[0], parts[1]
	_, err = findKinesisDataStreamDestinationByTwoPartKey(ctx, conn, streamARN, tableName)

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling DynamoDB Kinesis Streaming Destination (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting DynamoDB Kinesis Streaming Destination: %s", d.Id())
	_, err = conn.DisableKinesisStreamingDestination(ctx, &dynamodb.DisableKinesisStreamingDestinationInput{
		TableName: aws.String(tableName),
		StreamArn: aws.String(streamARN),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling DynamoDB Kinesis Streaming Destination (%s): %s", d.Id(), err)
	}

	if _, err := waitKinesisStreamingDestinationDisabled(ctx, conn, streamARN, tableName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Kinesis Streaming Destination (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func kinesisDataStreamDestinationForStream(arn string) tfslices.Predicate[awstypes.KinesisDataStreamDestination] {
	return func(v awstypes.KinesisDataStreamDestination) bool {
		return aws.ToString(v.StreamArn) == arn
	}
}

func findKinesisDataStreamDestinationByTwoPartKey(ctx context.Context, conn *dynamodb.Client, streamARN, tableName string) (*awstypes.KinesisDataStreamDestination, error) {
	input := &dynamodb.DescribeKinesisStreamingDestinationInput{
		TableName: aws.String(tableName),
	}
	output, err := findKinesisDataStreamDestination(ctx, conn, input, kinesisDataStreamDestinationForStream(streamARN))

	if err != nil {
		return nil, err
	}

	if output.DestinationStatus == awstypes.DestinationStatusDisabled {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findKinesisDataStreamDestination(ctx context.Context, conn *dynamodb.Client, input *dynamodb.DescribeKinesisStreamingDestinationInput, filter tfslices.Predicate[awstypes.KinesisDataStreamDestination]) (*awstypes.KinesisDataStreamDestination, error) {
	output, err := findKinesisDataStreamDestinations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findKinesisDataStreamDestinations(ctx context.Context, conn *dynamodb.Client, input *dynamodb.DescribeKinesisStreamingDestinationInput, filter tfslices.Predicate[awstypes.KinesisDataStreamDestination]) ([]awstypes.KinesisDataStreamDestination, error) {
	output, err := conn.DescribeKinesisStreamingDestination(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.KinesisDataStreamDestinations, filter), nil
}

func statusKinesisStreamingDestination(ctx context.Context, conn *dynamodb.Client, streamARN, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &dynamodb.DescribeKinesisStreamingDestinationInput{
			TableName: aws.String(tableName),
		}
		output, err := findKinesisDataStreamDestination(ctx, conn, input, kinesisDataStreamDestinationForStream(streamARN))

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.DestinationStatus), nil
	}
}

func waitKinesisStreamingDestinationActive(ctx context.Context, conn *dynamodb.Client, streamARN, tableName string) (*awstypes.KinesisDataStreamDestination, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DestinationStatusDisabled, awstypes.DestinationStatusEnabling),
		Target:  enum.Slice(awstypes.DestinationStatusActive),
		Timeout: timeout,
		Refresh: statusKinesisStreamingDestination(ctx, conn, streamARN, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KinesisDataStreamDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.DestinationStatusDescription)))
		return output, err
	}

	return nil, err
}

func waitKinesisStreamingDestinationDisabled(ctx context.Context, conn *dynamodb.Client, streamARN, tableName string) (*awstypes.KinesisDataStreamDestination, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DestinationStatusActive, awstypes.DestinationStatusDisabling),
		Target:  enum.Slice(awstypes.DestinationStatusDisabled),
		Timeout: timeout,
		Refresh: statusKinesisStreamingDestination(ctx, conn, streamARN, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.KinesisDataStreamDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.DestinationStatusDescription)))
		return output, err
	}

	return nil, err
}
