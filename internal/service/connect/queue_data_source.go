// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_queue", name="Queue")
// @Tags
func dataSourceQueue() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQueueRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hours_of_operation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"max_contacts": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "queue_id"},
			},
			"outbound_caller_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"outbound_caller_id_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"outbound_caller_id_number_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"outbound_flow_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"queue_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"queue_id", names.AttrName},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQueueRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.DescribeQueueInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("queue_id"); ok {
		input.QueueId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		queueSummary, err := findQueueSummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Queue (%s) summary: %s", name, err)
		}

		input.QueueId = queueSummary.Id
	}

	queue, err := findQueue(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Queue: %s", err)
	}

	queueID := aws.ToString(queue.QueueId)
	id := queueCreateResourceID(instanceID, queueID)
	d.SetId(id)
	d.Set(names.AttrARN, queue.QueueArn)
	d.Set(names.AttrDescription, queue.Description)
	d.Set("hours_of_operation_id", queue.HoursOfOperationId)
	d.Set("max_contacts", queue.MaxContacts)
	d.Set(names.AttrName, queue.Name)
	if err := d.Set("outbound_caller_config", flattenOutboundCallerConfig(queue.OutboundCallerConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting outbound_caller_config: %s", err)
	}
	d.Set("queue_id", queueID)
	d.Set(names.AttrStatus, queue.Status)

	setTagsOut(ctx, queue.Tags)

	return diags
}

func findQueueSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.QueueSummary, error) {
	const maxResults = 60
	input := &connect.ListQueuesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findQueueSummary(ctx, conn, input, func(v *awstypes.QueueSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findQueueSummary(ctx context.Context, conn *connect.Client, input *connect.ListQueuesInput, filter tfslices.Predicate[*awstypes.QueueSummary]) (*awstypes.QueueSummary, error) {
	output, err := findQueueSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findQueueSummaries(ctx context.Context, conn *connect.Client, input *connect.ListQueuesInput, filter tfslices.Predicate[*awstypes.QueueSummary]) ([]awstypes.QueueSummary, error) {
	var output []awstypes.QueueSummary

	pages := connect.NewListQueuesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.QueueSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
