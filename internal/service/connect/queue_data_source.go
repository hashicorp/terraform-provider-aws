package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceQueue() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceQueueRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hours_of_operation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"max_contacts": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "queue_id"},
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
				ExactlyOneOf: []string{"queue_id", "name"},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeQueueInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("queue_id"); ok {
		input.QueueId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		queueSummary, err := dataSourceGetQueueSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Queue Summary by name (%s): %w", name, err))
		}

		if queueSummary == nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Queue Summary by name (%s): not found", name))
		}

		input.QueueId = queueSummary.Id
	}

	resp, err := conn.DescribeQueueWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Queue: %w", err))
	}

	if resp == nil || resp.Queue == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Queue: empty response"))
	}

	queue := resp.Queue

	d.Set("arn", queue.QueueArn)
	d.Set("description", queue.Description)
	d.Set("hours_of_operation_id", queue.HoursOfOperationId)
	d.Set("max_contacts", queue.MaxContacts)
	d.Set("name", queue.Name)
	d.Set("queue_id", queue.QueueId)
	d.Set("status", queue.Status)

	if err := d.Set("outbound_caller_config", flattenOutboundCallerConfig(queue.OutboundCallerConfig)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting outbound_caller_config: %s", err))
	}

	if err := d.Set("tags", KeyValueTags(queue.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(queue.QueueId)))

	return nil
}

func dataSourceGetQueueSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.QueueSummary, error) {
	var result *connect.QueueSummary

	input := &connect.ListQueuesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListQueuesMaxResults),
	}

	err := conn.ListQueuesPagesWithContext(ctx, input, func(page *connect.ListQueuesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qs := range page.QueueSummaryList {
			if qs == nil {
				continue
			}

			if aws.StringValue(qs.Name) == name {
				result = qs
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
