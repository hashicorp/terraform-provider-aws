// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_queue", name="Queue")
// @Tags(identifierAttribute="arn")
func ResourceQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueueCreate,
		ReadWithoutTimeout:   resourceQueueRead,
		UpdateWithoutTimeout: resourceQueueUpdate,
		DeleteWithoutTimeout: resourceQueueDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			"hours_of_operation_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"max_contacts": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			names.AttrName: {
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
			names.AttrStatus: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(connect.QueueStatus_Values(), false), // Valid Values: ENABLED | DISABLED
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	input := &connect.CreateQueueInput{
		InstanceId: aws.String(instanceID),
		Name:       aws.String(name),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
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

	log.Printf("[DEBUG] Creating Connect Queue %s", input)
	output, err := conn.CreateQueueWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Queue (%s): %s", name, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Queue (%s): empty output", name)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.QueueId)))

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, queueID, err := QueueParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := conn.DescribeQueueWithContext(ctx, &connect.DescribeQueueInput{
		InstanceId: aws.String(instanceID),
		QueueId:    aws.String(queueID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Queue (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.Queue == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Queue (%s): empty response", d.Id())
	}

	if err := d.Set("outbound_caller_config", flattenOutboundCallerConfig(resp.Queue.OutboundCallerConfig)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrARN, resp.Queue.QueueArn)
	d.Set(names.AttrDescription, resp.Queue.Description)
	d.Set("hours_of_operation_id", resp.Queue.HoursOfOperationId)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("max_contacts", resp.Queue.MaxContacts)
	d.Set(names.AttrName, resp.Queue.Name)
	d.Set("queue_id", resp.Queue.QueueId)
	d.Set(names.AttrStatus, resp.Queue.Status)

	// reading quick_connect_ids requires a separate API call
	quickConnectIds, err := getQueueQuickConnectIDs(ctx, conn, instanceID, queueID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Queue Quick Connect ID for Queue (%s): %s", queueID, err)
	}

	d.Set("quick_connect_ids", aws.StringValueSlice(quickConnectIds))

	setTagsOut(ctx, resp.Queue.Tags)

	return diags
}

func resourceQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, queueID, err := QueueParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
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
			return sdkdiag.AppendErrorf(diags, "updating Queue Hours of Operation (%s): %s", d.Id(), err)
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
			return sdkdiag.AppendErrorf(diags, "updating Queue Max Contacts (%s): %s", d.Id(), err)
		}
	}

	// updates to name and/or description
	if d.HasChanges(names.AttrName, names.AttrDescription) {
		input := &connect.UpdateQueueNameInput{
			InstanceId:  aws.String(instanceID),
			QueueId:     aws.String(queueID),
			Name:        aws.String(d.Get(names.AttrName).(string)),
			Description: aws.String(d.Get(names.AttrDescription).(string)),
		}
		_, err = conn.UpdateQueueNameWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Queue Name and/or Description (%s): %s", d.Id(), err)
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
			return sdkdiag.AppendErrorf(diags, "updating Queue Outbound Caller Config (%s): %s", d.Id(), err)
		}
	}

	// updates to status
	if d.HasChange(names.AttrStatus) {
		input := &connect.UpdateQueueStatusInput{
			InstanceId: aws.String(instanceID),
			QueueId:    aws.String(queueID),
			Status:     aws.String(d.Get(names.AttrStatus).(string)),
		}
		_, err = conn.UpdateQueueStatusWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Queue Status (%s): %s", d.Id(), err)
		}
	}

	// updates to quick_connect_ids
	if d.HasChange("quick_connect_ids") {
		o, n := d.GetChange("quick_connect_ids")

		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		quickConnectIdsUpdateAdd := ns.Difference(os)
		quickConnectIdsUpdateRemove := os.Difference(ns)

		if len(quickConnectIdsUpdateAdd.List()) > 0 { // nosemgrep:ci.semgrep.migrate.aws-api-context
			_, err = conn.AssociateQueueQuickConnectsWithContext(ctx, &connect.AssociateQueueQuickConnectsInput{
				InstanceId:      aws.String(instanceID),
				QueueId:         aws.String(queueID),
				QuickConnectIds: flex.ExpandStringSet(quickConnectIdsUpdateAdd),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Queues Quick Connect IDs, specifically associating quick connects to queue (%s): %s", d.Id(), err)
			}
		}

		if len(quickConnectIdsUpdateRemove.List()) > 0 { // nosemgrep:ci.semgrep.migrate.aws-api-context
			_, err = conn.DisassociateQueueQuickConnectsWithContext(ctx, &connect.DisassociateQueueQuickConnectsInput{
				InstanceId:      aws.String(instanceID),
				QueueId:         aws.String(queueID),
				QuickConnectIds: flex.ExpandStringSet(quickConnectIdsUpdateRemove),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Queues Quick Connect IDs, specifically disassociating quick connects from queue (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, queueID, err := QueueParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteQueueWithContext(ctx, &connect.DeleteQueueInput{
		InstanceId: aws.String(instanceID),
		QueueId:    aws.String(queueID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Queue (%s): %s", d.Id(), err)
	}

	return diags
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
