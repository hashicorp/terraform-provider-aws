package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceQueue() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceQueueCreate,
		ReadContext:   resourceQueueRead,
		UpdateContext: resourceQueueUpdate,
		// Queues do not support deletion today. NoOp the Delete method.
		// Users can rename their queues manually if they want.
		DeleteContext: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			"hours_of_operation_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"max_contacts": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 127),
			},
			"outbound_caller_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"outbound_caller_id_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"outbound_caller_id_number_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"outbound_flow_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 500),
						},
					},
				},
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"queue_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"quick_connect_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(connect.QueueStatus_Values(), false), // Valid Values: ENABLED | DISABLED
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID, queueID, err := QueueParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeQueueWithContext(ctx, &connect.DescribeQueueInput{
		InstanceId: aws.String(instanceID),
		QueueId:    aws.String(queueID),
	})

	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, connect.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Connect Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Queue (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.Queue == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Queue (%s): empty response", d.Id()))
	}

	if err := d.Set("outbound_caller_config", flattenOutboundCallerConfig(resp.Queue.OutboundCallerConfig)); err != nil {
		return diag.FromErr(err)
	}

	d.Set("arn", resp.Queue.QueueArn)
	d.Set("description", resp.Queue.Description)
	d.Set("hours_of_operation_id", resp.Queue.HoursOfOperationId)
	d.Set("instance_id", instanceID)
	d.Set("max_contacts", resp.Queue.MaxContacts)
	d.Set("name", resp.Queue.Name)
	d.Set("queue_id", resp.Queue.QueueId)
	d.Set("status", resp.Queue.Status)

	tags := KeyValueTags(resp.Queue.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}
func expandOutboundCallerConfig(outboundCallerConfig []interface{}) *connect.OutboundCallerConfig {
	if len(outboundCallerConfig) == 0 || outboundCallerConfig[0] == nil {
		return nil
	}

	tfMap, ok := outboundCallerConfig[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &connect.OutboundCallerConfig{}

	if v, ok := tfMap["outbound_caller_id_name"].(string); ok && v != "" {
		result.OutboundCallerIdName = aws.String(v)
	}

	// passing an empty string leads to an InvalidParameterException
	if v, ok := tfMap["outbound_caller_id_number_id"].(string); ok && v != "" {
		result.OutboundCallerIdNumberId = aws.String(v)
	}

	// passing an empty string leads to an InvalidParameterException
	if v, ok := tfMap["outbound_flow_id"].(string); ok && v != "" {
		result.OutboundFlowId = aws.String(v)
	}

	return result
}

func flattenOutboundCallerConfig(outboundCallerConfig *connect.OutboundCallerConfig) []interface{} {
	if outboundCallerConfig == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if v := outboundCallerConfig.OutboundCallerIdName; v != nil {
		values["outbound_caller_id_name"] = aws.StringValue(v)
	}

	if v := outboundCallerConfig.OutboundCallerIdNumberId; v != nil {
		values["outbound_caller_id_number_id"] = aws.StringValue(v)
	}

	if v := outboundCallerConfig.OutboundFlowId; v != nil {
		values["outbound_flow_id"] = aws.StringValue(v)
	}

	return []interface{}{values}
}

func QueueParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:queueID", id)
	}

	return parts[0], parts[1], nil
}
