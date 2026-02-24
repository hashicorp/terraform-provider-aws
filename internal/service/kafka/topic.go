// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package kafka

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const topicResourceIDPartCount = 2

// @SDKResource("aws_msk_topic", name="Topic")
func resourceTopic() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTopicCreate,
		ReadWithoutTimeout:   resourceTopicRead,
		UpdateWithoutTimeout: resourceTopicUpdate,
		DeleteWithoutTimeout: resourceTopicDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"configs": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Base64-encoded Kafka topic configurations",
			},
			"partition_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"replication_factor": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"topic_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTopicCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	clusterArn := d.Get("cluster_arn").(string)
	topicName := d.Get("topic_name").(string)

	partitionCount := int32(d.Get("partition_count").(int))
	replicationFactor := int32(d.Get("replication_factor").(int))
	input := &kafka.CreateTopicInput{
		ClusterArn:        aws.String(clusterArn),
		PartitionCount:    &partitionCount,
		ReplicationFactor: &replicationFactor,
		TopicName:         aws.String(topicName),
	}

	if v, ok := d.GetOk("configs"); ok {
		input.Configs = aws.String(v.(string))
	}

	_, err := conn.CreateTopic(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Topic (%s/%s): %s", clusterArn, topicName, err)
	}

	id, err := flex.FlattenResourceId([]string{clusterArn, topicName}, topicResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.SetId(id)

	if _, err := waitTopicCreated(ctx, conn, clusterArn, topicName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Topic (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTopicRead(ctx, d, meta)...)
}

func resourceTopicRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), topicResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	clusterArn, topicName := parts[0], parts[1]
	output, err := findTopicByTwoPartKey(ctx, conn, clusterArn, topicName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] MSK Topic (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Topic (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.TopicArn)
	d.Set("cluster_arn", clusterArn)
	d.Set("configs", output.Configs)
	d.Set("partition_count", output.PartitionCount)
	d.Set("replication_factor", output.ReplicationFactor)
	d.Set("topic_name", output.TopicName)

	return diags
}

func resourceTopicUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	if d.HasChanges("partition_count", "configs") {
		parts, err := flex.ExpandResourceId(d.Id(), topicResourceIDPartCount, false)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		clusterArn, topicName := parts[0], parts[1]
		input := &kafka.UpdateTopicInput{
			ClusterArn: aws.String(clusterArn),
			TopicName:  aws.String(topicName),
		}

		if d.HasChange("partition_count") {
			v := int32(d.Get("partition_count").(int))
			input.PartitionCount = &v
		}

		if d.HasChange("configs") {
			input.Configs = aws.String(d.Get("configs").(string))
		}

		_, err = conn.UpdateTopic(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Topic (%s): %s", d.Id(), err)
		}

		if _, err := waitTopicUpdated(ctx, conn, clusterArn, topicName, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Topic (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTopicRead(ctx, d, meta)...)
}

func resourceTopicDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), topicResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	clusterArn, topicName := parts[0], parts[1]

	log.Printf("[INFO] Deleting MSK Topic: %s", d.Id())
	_, err = conn.DeleteTopic(ctx, &kafka.DeleteTopicInput{
		ClusterArn: aws.String(clusterArn),
		TopicName:  aws.String(topicName),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK Topic (%s): %s", d.Id(), err)
	}

	if _, err := waitTopicDeleted(ctx, conn, clusterArn, topicName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Topic (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func waitTopicCreated(ctx context.Context, conn *kafka.Client, clusterArn, topicName string, timeout time.Duration) (*kafka.DescribeTopicOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.TopicStateCreating),
		Target:                    enum.Slice(types.TopicStateActive),
		Refresh:                   statusTopic(conn, clusterArn, topicName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeTopicOutput); ok {
		return output, err
	}

	return nil, err
}

func waitTopicUpdated(ctx context.Context, conn *kafka.Client, clusterArn, topicName string, timeout time.Duration) (*kafka.DescribeTopicOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TopicStateUpdating),
		Target:  enum.Slice(types.TopicStateActive),
		Refresh: statusTopic(conn, clusterArn, topicName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeTopicOutput); ok {
		return output, err
	}

	return nil, err
}

func waitTopicDeleted(ctx context.Context, conn *kafka.Client, clusterArn, topicName string, timeout time.Duration) (*kafka.DescribeTopicOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TopicStateDeleting, types.TopicStateActive),
		Target:  []string{},
		Refresh: statusTopic(conn, clusterArn, topicName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeTopicOutput); ok {
		return output, err
	}

	return nil, err
}

func statusTopic(conn *kafka.Client, clusterArn, topicName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTopicByTwoPartKey(ctx, conn, clusterArn, topicName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findTopicByTwoPartKey(ctx context.Context, conn *kafka.Client, clusterArn, topicName string) (*kafka.DescribeTopicOutput, error) {
	input := &kafka.DescribeTopicInput{
		ClusterArn: aws.String(clusterArn),
		TopicName:  aws.String(topicName),
	}

	output, err := conn.DescribeTopic(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
