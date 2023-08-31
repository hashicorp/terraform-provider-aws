// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_kinesis_stream_consumer")
func ResourceStreamConsumer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamConsumerCreate,
		ReadWithoutTimeout:   resourceStreamConsumerRead,
		DeleteWithoutTimeout: resourceStreamConsumerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stream_arn": {
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
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)

	name := d.Get("name").(string)
	input := &kinesis.RegisterStreamConsumerInput{
		ConsumerName: aws.String(name),
		StreamARN:    aws.String(d.Get("stream_arn").(string)),
	}

	log.Printf("[DEBUG] Registering Kinesis Stream Consumer: %s", input)
	output, err := conn.RegisterStreamConsumerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kinesis Stream Consumer (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Consumer.ConsumerARN))

	if _, err := waitStreamConsumerCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kinesis Stream Consumer (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStreamConsumerRead(ctx, d, meta)...)
}

func resourceStreamConsumerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)

	consumer, err := FindStreamConsumerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Stream Consumer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Stream Consumer (%s): %s", d.Id(), err)
	}

	d.Set("arn", consumer.ConsumerARN)
	d.Set("creation_timestamp", aws.TimeValue(consumer.ConsumerCreationTimestamp).Format(time.RFC3339))
	d.Set("name", consumer.ConsumerName)
	d.Set("stream_arn", consumer.StreamARN)

	return diags
}

func resourceStreamConsumerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)

	log.Printf("[DEBUG] Deregistering Kinesis Stream Consumer: (%s)", d.Id())
	_, err := conn.DeregisterStreamConsumerWithContext(ctx, &kinesis.DeregisterStreamConsumerInput{
		ConsumerARN: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
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

func FindStreamConsumerByARN(ctx context.Context, conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	input := &kinesis.DescribeStreamConsumerInput{
		ConsumerARN: aws.String(arn),
	}

	output, err := conn.DescribeStreamConsumerWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
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

func statusStreamConsumer(ctx context.Context, conn *kinesis.Kinesis, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStreamConsumerByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ConsumerStatus), nil
	}
}

const (
	streamConsumerCreatedTimeout = 5 * time.Minute
	streamConsumerDeletedTimeout = 5 * time.Minute
)

func waitStreamConsumerCreated(ctx context.Context, conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesis.ConsumerStatusCreating},
		Target:  []string{kinesis.ConsumerStatusActive},
		Refresh: statusStreamConsumer(ctx, conn, arn),
		Timeout: streamConsumerCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kinesis.ConsumerDescription); ok {
		return output, err
	}

	return nil, err
}

func waitStreamConsumerDeleted(ctx context.Context, conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kinesis.ConsumerStatusDeleting},
		Target:  []string{},
		Refresh: statusStreamConsumer(ctx, conn, arn),
		Timeout: streamConsumerDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kinesis.ConsumerDescription); ok {
		return output, err
	}

	return nil, err
}
