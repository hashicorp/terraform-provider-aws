package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
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
			"queue_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"quick_connect_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"quick_connect_ids_associated": {
				Type:     schema.TypeSet,
				Computed: true,
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

func resourceQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)

	input := &connect.CreateQueueInput{
		InstanceId: aws.String(instanceID),
		Name:       aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("hours_of_operation_id"); ok {
		input.HoursOfOperationId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_contacts"); ok {
		input.MaxContacts = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("outbound_caller_config"); ok {
		input.OutboundCallerConfig = expandOutboundCallerConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("quick_connect_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.QuickConnectIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Connect Queue %s", input)
	output, err := conn.CreateQueueWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Queue (%s): %w", name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Queue (%s): empty output", name))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.QueueId)))

	return resourceQueueRead(ctx, d, meta)
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

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
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

	// reading quick_connect_ids requires a separate API call
	quickConnectIds, err := getQueueQuickConnectIDs(ctx, conn, instanceID, queueID)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Queue Quick Connect ID for Queue (%s): %w", queueID, err))
	}

	d.Set("quick_connect_ids", flex.FlattenStringSet(quickConnectIds))
	d.Set("quick_connect_ids_associated", flex.FlattenStringSet(quickConnectIds))

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

func resourceQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	instanceID, queueID, err := QueueParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	// Queue has 6 update APIs
	// UpdateQueueHoursOfOperationWithContext: Updates the hours_of_operation_id of a queue.
	// UpdateQueueMaxContactsWithContext: Updates the max_contacts of a queue.
	// UpdateQueueNameWithContext: Updates the name and description of a queue.
	// UpdateQueueOutboundCallerConfigWithContext: Updates the outbound_caller_config of a queue.
	// UpdateQueueStatusWithContext: Updates the status of a queue. Valid Values: ENABLED | DISABLED
	// AssociateQueueQuickConnectsWithContext: Associates a set of quick connects with a queue. There is also DisassociateQueueQuickConnectsWithContext

	// updates to hours_of_operation_id
	if d.HasChange("hours_of_operation_id") {
		input := &connect.UpdateQueueHoursOfOperationInput{
			InstanceId:         aws.String(instanceID),
			QueueId:            aws.String(queueID),
			HoursOfOperationId: aws.String(d.Get("hours_of_operation_id").(string)),
		}
		_, err = conn.UpdateQueueHoursOfOperationWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error updating Queue Hours of Operation (%s): %w", d.Id(), err))
		}
	}

	// updates to max_contacts
	if d.HasChange("max_contacts") {
		input := &connect.UpdateQueueMaxContactsInput{
			InstanceId:  aws.String(instanceID),
			QueueId:     aws.String(queueID),
			MaxContacts: aws.Int64(int64(d.Get("max_contacts").(int))),
		}
		_, err = conn.UpdateQueueMaxContactsWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error updating Queue Max Contacts (%s): %w", d.Id(), err))
		}
	}

	// updates to name and/or description
	if d.HasChanges("name", "description") {
		input := &connect.UpdateQueueNameInput{
			InstanceId:  aws.String(instanceID),
			QueueId:     aws.String(queueID),
			Name:        aws.String(d.Get("name").(string)),
			Description: aws.String(d.Get("description").(string)),
		}
		_, err = conn.UpdateQueueNameWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error updating Queue Name and/or Description (%s): %w", d.Id(), err))
		}
	}

	// updates to outbound_caller_config
	if d.HasChange("outbound_caller_config") {
		input := &connect.UpdateQueueOutboundCallerConfigInput{
			InstanceId:           aws.String(instanceID),
			QueueId:              aws.String(queueID),
			OutboundCallerConfig: expandOutboundCallerConfig(d.Get("outbound_caller_config").([]interface{})),
		}
		_, err = conn.UpdateQueueOutboundCallerConfigWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error updating Queue Outbound Caller Config (%s): %w", d.Id(), err))
		}
	}

	// updates to status
	if d.HasChange("status") {
		input := &connect.UpdateQueueStatusInput{
			InstanceId: aws.String(instanceID),
			QueueId:    aws.String(queueID),
			Status:     aws.String(d.Get("status").(string)),
		}
		_, err = conn.UpdateQueueStatusWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error updating Queue Status (%s): %w", d.Id(), err))
		}
	}

	// updates to quick_connect_ids
	if d.HasChange("quick_connect_ids") {
		// first disassociate all existing quick connects
		if v, ok := d.GetOk("quick_connect_ids_associated"); ok && v.(*schema.Set).Len() > 0 {
			input := &connect.DisassociateQueueQuickConnectsInput{
				InstanceId: aws.String(instanceID),
				QueueId:    aws.String(queueID),
			}
			input.QuickConnectIds = flex.ExpandStringSet(v.(*schema.Set))
			_, err = conn.DisassociateQueueQuickConnectsWithContext(ctx, input)
			if err != nil {
				return diag.FromErr(fmt.Errorf("[ERROR] Error updating Queues Quick Connect IDs, specifically disassociating quick connects from queue (%s): %w", d.Id(), err))
			}
		}

		// re-associate the quick connects
		if v, ok := d.GetOk("quick_connect_ids"); ok && v.(*schema.Set).Len() > 0 {
			input := &connect.AssociateQueueQuickConnectsInput{
				InstanceId: aws.String(instanceID),
				QueueId:    aws.String(queueID),
			}
			input.QuickConnectIds = flex.ExpandStringSet(v.(*schema.Set))
			_, err = conn.AssociateQueueQuickConnectsWithContext(ctx, input)
			if err != nil {
				return diag.FromErr(fmt.Errorf("[ERROR] Error updating Queues Quick Connect IDs, specifically associating quick connects to queue (%s): %w", d.Id(), err))
			}
		}
	}

	// updates to tags
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceQueueRead(ctx, d, meta)
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

func getQueueQuickConnectIDs(ctx context.Context, conn *connect.Connect, instanceID, queueID string) ([]*string, error) {
	var result []*string

	input := &connect.ListQueueQuickConnectsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListQueueQuickConnectsMaxResults),
		QueueId:    aws.String(queueID),
	}

	err := conn.ListQueueQuickConnectsPagesWithContext(ctx, input, func(page *connect.ListQueueQuickConnectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qc := range page.QuickConnectSummaryList {
			if qc == nil {
				continue
			}

			result = append(result, qc.Id)
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func QueueParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:queueID", id)
	}

	return parts[0], parts[1], nil
}
