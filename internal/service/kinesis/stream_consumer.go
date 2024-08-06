// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kinesis_stream_consumer", name="Stream Consumer")
func resourceStreamConsumer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamConsumerCreate,
		ReadWithoutTimeout:   resourceStreamConsumerRead,
		DeleteWithoutTimeout: resourceStreamConsumerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStreamARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceStreamConsumerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &kinesis.RegisterStreamConsumerInput{
		ConsumerName: aws.String(name),
		StreamARN:    aws.String(d.Get(names.AttrStreamARN).(string)),
	}

	output, err := conn.RegisterStreamConsumer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Stream Consumer (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Consumer.ConsumerARN))

	if _, err := waitStreamConsumerCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream Consumer (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStreamConsumerRead(ctx, d, meta)...)
}

func resourceStreamConsumerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	consumer, err := findStreamConsumerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Stream Consumer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Stream Consumer (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, consumer.ConsumerARN)
	d.Set("creation_timestamp", aws.ToTime(consumer.ConsumerCreationTimestamp).Format(time.RFC3339))
	d.Set(names.AttrName, consumer.ConsumerName)
	d.Set(names.AttrStreamARN, consumer.StreamARN)

	return diags
}

func resourceStreamConsumerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	log.Printf("[DEBUG] Deregistering Kinesis Stream Consumer: (%s)", d.Id())
	_, err := conn.DeregisterStreamConsumer(ctx, &kinesis.DeregisterStreamConsumerInput{
		ConsumerARN: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kinesis Stream Consumer (%s): %s", d.Id(), err)
	}

	if _, err := waitStreamConsumerDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream Consumer (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findStreamConsumerByARN(ctx context.Context, conn *kinesis.Client, arn string) (*types.ConsumerDescription, error) {
	input := &kinesis.DescribeStreamConsumerInput{
		ConsumerARN: aws.String(arn),
	}

	output, err := conn.DescribeStreamConsumer(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ConsumerDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ConsumerDescription, nil
}

func statusStreamConsumer(ctx context.Context, conn *kinesis.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findStreamConsumerByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ConsumerStatus), nil
	}
}

func waitStreamConsumerCreated(ctx context.Context, conn *kinesis.Client, arn string) (*types.ConsumerDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConsumerStatusCreating),
		Target:  enum.Slice(types.ConsumerStatusActive),
		Refresh: statusStreamConsumer(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ConsumerDescription); ok {
		return output, err
	}

	return nil, err
}

func waitStreamConsumerDeleted(ctx context.Context, conn *kinesis.Client, arn string) (*types.ConsumerDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConsumerStatusDeleting),
		Target:  []string{},
		Refresh: statusStreamConsumer(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ConsumerDescription); ok {
		return output, err
	}

	return nil, err
}
