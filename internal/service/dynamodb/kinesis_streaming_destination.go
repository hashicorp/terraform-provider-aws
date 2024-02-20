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
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"stream_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceKinesisStreamingDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	streamARN := d.Get("stream_arn").(string)
	tableName := d.Get("table_name").(string)
	id := errs.Must(flex.FlattenResourceId([]string{tableName, streamARN}, kinesisStreamingDestinationResourceIDPartCount, false))
	input := &dynamodb.EnableKinesisStreamingDestinationInput{
		StreamArn: aws.String(streamARN),
		TableName: aws.String(tableName),
	}

	_, err := conn.EnableKinesisStreamingDestinationWithContext(ctx, input)

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
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

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

	d.Set("stream_arn", output.StreamArn)
	d.Set("table_name", tableName)

	return diags
}

func resourceKinesisStreamingDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), kinesisStreamingDestinationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	tableName, streamARN := parts[0], parts[1]
	_, err = findKinesisDataStreamDestinationByTwoPartKey(ctx, conn, streamARN, tableName)

	if tfresource.NotFound(err) {
		return diags
	}

	log.Printf("[DEBUG] Deleting DynamoDB Kinesis Streaming Destination: %s", d.Id())
	_, err = conn.DisableKinesisStreamingDestinationWithContext(ctx, &dynamodb.DisableKinesisStreamingDestinationInput{
		TableName: aws.String(tableName),
		StreamArn: aws.String(streamARN),
	})

	if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
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

func kinesisDataStreamDestinationForStream(arn string) tfslices.Predicate[*dynamodb.KinesisDataStreamDestination] {
	return func(v *dynamodb.KinesisDataStreamDestination) bool {
		return aws.StringValue(v.StreamArn) == arn
	}
}

func findKinesisDataStreamDestinationByTwoPartKey(ctx context.Context, conn *dynamodb.DynamoDB, streamARN, tableName string) (*dynamodb.KinesisDataStreamDestination, error) {
	input := &dynamodb.DescribeKinesisStreamingDestinationInput{
		TableName: aws.String(tableName),
	}
	output, err := findKinesisDataStreamDestination(ctx, conn, input, kinesisDataStreamDestinationForStream(streamARN))

	if err != nil {
		return nil, err
	}

	if aws.StringValue(output.DestinationStatus) == dynamodb.DestinationStatusDisabled {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findKinesisDataStreamDestination(ctx context.Context, conn *dynamodb.DynamoDB, input *dynamodb.DescribeKinesisStreamingDestinationInput, filter tfslices.Predicate[*dynamodb.KinesisDataStreamDestination]) (*dynamodb.KinesisDataStreamDestination, error) {
	output, err := findKinesisDataStreamDestinations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findKinesisDataStreamDestinations(ctx context.Context, conn *dynamodb.DynamoDB, input *dynamodb.DescribeKinesisStreamingDestinationInput, filter tfslices.Predicate[*dynamodb.KinesisDataStreamDestination]) ([]*dynamodb.KinesisDataStreamDestination, error) {
	output, err := conn.DescribeKinesisStreamingDestinationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
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

func statusKinesisStreamingDestination(ctx context.Context, conn *dynamodb.DynamoDB, streamARN, tableName string) retry.StateRefreshFunc {
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

		return output, aws.StringValue(output.DestinationStatus), nil
	}
}

func waitKinesisStreamingDestinationActive(ctx context.Context, conn *dynamodb.DynamoDB, streamARN, tableName string) (*dynamodb.KinesisDataStreamDestination, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusDisabled, dynamodb.DestinationStatusEnabling},
		Target:  []string{dynamodb.DestinationStatusActive},
		Timeout: timeout,
		Refresh: statusKinesisStreamingDestination(ctx, conn, streamARN, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.KinesisDataStreamDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.DestinationStatusDescription)))
		return output, err
	}

	return nil, err
}

func waitKinesisStreamingDestinationDisabled(ctx context.Context, conn *dynamodb.DynamoDB, streamARN, tableName string) (*dynamodb.KinesisDataStreamDestination, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusActive, dynamodb.DestinationStatusDisabling},
		Target:  []string{dynamodb.DestinationStatusDisabled},
		Timeout: timeout,
		Refresh: statusKinesisStreamingDestination(ctx, conn, streamARN, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.KinesisDataStreamDestination); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.DestinationStatusDescription)))
		return output, err
	}

	return nil, err
}
