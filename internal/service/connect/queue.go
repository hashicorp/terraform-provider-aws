// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_queue", name="Queue")
// @Tags(identifierAttribute="arn")
func resourceQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueueCreate,
		ReadWithoutTimeout:   resourceQueueRead,
		UpdateWithoutTimeout: resourceQueueUpdate,
		DeleteWithoutTimeout: resourceQueueDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.QueueStatus](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceQueueCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

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
		input.MaxContacts = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("outbound_caller_config"); ok {
		input.OutboundCallerConfig = expandOutboundCallerConfig(v.([]any))
	}

	if v, ok := d.GetOk("quick_connect_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.QuickConnectIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateQueue(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Queue (%s): %s", name, err)
	}

	id := queueCreateResourceID(instanceID, aws.ToString(output.QueueId))
	d.SetId(id)

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, queueID, err := queueParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	queue, err := findQueueByTwoPartKey(ctx, conn, instanceID, queueID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Queue (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, queue.QueueArn)
	d.Set(names.AttrDescription, queue.Description)
	d.Set("hours_of_operation_id", queue.HoursOfOperationId)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("max_contacts", queue.MaxContacts)
	d.Set(names.AttrName, queue.Name)
	if err := d.Set("outbound_caller_config", flattenOutboundCallerConfig(queue.OutboundCallerConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting outbound_caller_config: %s", err)
	}
	d.Set("queue_id", queue.QueueId)
	d.Set(names.AttrStatus, queue.Status)

	quickConnects, err := findQueueQuickConnectSummariesByTwoPartKey(ctx, conn, instanceID, queueID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Queue (%s) Quick Connect summaries: %s", d.Id(), err)
	}

	d.Set("quick_connect_ids", tfslices.ApplyToAll(quickConnects, func(v awstypes.QuickConnectSummary) string {
		return aws.ToString(v.Id)
	}))

	setTagsOut(ctx, queue.Tags)

	return diags
}

func resourceQueueUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, queueID, err := queueParseResourceID(d.Id())
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
			HoursOfOperationId: aws.String(d.Get("hours_of_operation_id").(string)),
			InstanceId:         aws.String(instanceID),
			QueueId:            aws.String(queueID),
		}

		_, err = conn.UpdateQueueHoursOfOperation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Queue (%s) HoursOfOperation: %s", d.Id(), err)
		}
	}

	// updates to max_contacts
	if d.HasChange("max_contacts") {
		input := &connect.UpdateQueueMaxContactsInput{
			InstanceId:  aws.String(instanceID),
			MaxContacts: aws.Int32(int32(d.Get("max_contacts").(int))),
			QueueId:     aws.String(queueID),
		}

		_, err = conn.UpdateQueueMaxContacts(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Queue (%s) MaxContacts: %s", d.Id(), err)
		}
	}

	// updates to name and/or description
	if d.HasChanges(names.AttrName, names.AttrDescription) {
		input := &connect.UpdateQueueNameInput{
			Description: aws.String(d.Get(names.AttrDescription).(string)),
			InstanceId:  aws.String(instanceID),
			Name:        aws.String(d.Get(names.AttrName).(string)),
			QueueId:     aws.String(queueID),
		}

		_, err = conn.UpdateQueueName(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Queue (%s) Name: %s", d.Id(), err)
		}
	}

	// updates to outbound_caller_config
	if d.HasChange("outbound_caller_config") {
		input := &connect.UpdateQueueOutboundCallerConfigInput{
			InstanceId:           aws.String(instanceID),
			OutboundCallerConfig: expandOutboundCallerConfig(d.Get("outbound_caller_config").([]any)),
			QueueId:              aws.String(queueID),
		}

		_, err = conn.UpdateQueueOutboundCallerConfig(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Queue (%s) OutboundCallerConfig: %s", d.Id(), err)
		}
	}

	// updates to status
	if d.HasChange(names.AttrStatus) {
		input := &connect.UpdateQueueStatusInput{
			InstanceId: aws.String(instanceID),
			QueueId:    aws.String(queueID),
			Status:     awstypes.QueueStatus(d.Get(names.AttrStatus).(string)),
		}

		_, err = conn.UpdateQueueStatus(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Queue (%s) Status: %s", d.Id(), err)
		}
	}

	// updates to quick_connect_ids
	if d.HasChange("quick_connect_ids") {
		o, n := d.GetChange("quick_connect_ids")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

		if len(add) > 0 {
			input := &connect.AssociateQueueQuickConnectsInput{
				InstanceId:      aws.String(instanceID),
				QueueId:         aws.String(queueID),
				QuickConnectIds: add,
			}

			_, err = conn.AssociateQueueQuickConnects(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating Connect Queue (%s) Quick Connects: %s", d.Id(), err)
			}
		}

		if len(del) > 0 {
			input := &connect.DisassociateQueueQuickConnectsInput{
				InstanceId:      aws.String(instanceID),
				QueueId:         aws.String(queueID),
				QuickConnectIds: del,
			}

			_, err = conn.DisassociateQueueQuickConnects(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating Connect Queue (%s) Quick Connects: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, queueID, err := queueParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Queue: %s", d.Id())
	const (
		timeout = 1 * time.Minute
	)
	_, err = tfresource.RetryWhenIsA[*awstypes.ResourceInUseException](ctx, timeout, func() (any, error) {
		return conn.DeleteQueue(ctx, &connect.DeleteQueueInput{
			InstanceId: aws.String(instanceID),
			QueueId:    aws.String(queueID),
		})
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Queue (%s): %s", d.Id(), err)
	}

	return diags
}

const queueResourceIDSeparator = ":"

func queueCreateResourceID(instanceID, queueID string) string {
	parts := []string{instanceID, queueID}
	id := strings.Join(parts, queueResourceIDSeparator)

	return id
}

func queueParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, queueResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]squeueID", id, queueResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findQueueByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, queueID string) (*awstypes.Queue, error) {
	input := &connect.DescribeQueueInput{
		InstanceId: aws.String(instanceID),
		QueueId:    aws.String(queueID),
	}

	return findQueue(ctx, conn, input)
}

func findQueue(ctx context.Context, conn *connect.Client, input *connect.DescribeQueueInput) (*awstypes.Queue, error) {
	output, err := conn.DescribeQueue(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Queue == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Queue, nil
}

func findQueueQuickConnectSummariesByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, queueID string) ([]awstypes.QuickConnectSummary, error) {
	const maxResults = 60
	input := &connect.ListQueueQuickConnectsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
		QueueId:    aws.String(queueID),
	}

	return findQueueQuickConnectSummaries(ctx, conn, input)
}

func findQueueQuickConnectSummaries(ctx context.Context, conn *connect.Client, input *connect.ListQueueQuickConnectsInput) ([]awstypes.QuickConnectSummary, error) {
	var output []awstypes.QuickConnectSummary

	pages := connect.NewListQueueQuickConnectsPaginator(conn, input)
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

		output = append(output, page.QuickConnectSummaryList...)
	}

	return output, nil
}

func expandOutboundCallerConfig(tfList []any) *awstypes.OutboundCallerConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.OutboundCallerConfig{}

	if v, ok := tfMap["outbound_caller_id_name"].(string); ok && v != "" {
		apiObject.OutboundCallerIdName = aws.String(v)
	}

	// passing an empty string leads to an InvalidParameterException
	if v, ok := tfMap["outbound_caller_id_number_id"].(string); ok && v != "" {
		apiObject.OutboundCallerIdNumberId = aws.String(v)
	}

	// passing an empty string leads to an InvalidParameterException
	if v, ok := tfMap["outbound_flow_id"].(string); ok && v != "" {
		apiObject.OutboundFlowId = aws.String(v)
	}

	return apiObject
}

func flattenOutboundCallerConfig(apiObject *awstypes.OutboundCallerConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if v := apiObject.OutboundCallerIdName; v != nil {
		tfMap["outbound_caller_id_name"] = aws.ToString(v)
	}

	if v := apiObject.OutboundCallerIdNumberId; v != nil {
		tfMap["outbound_caller_id_number_id"] = aws.ToString(v)
	}

	if v := apiObject.OutboundFlowId; v != nil {
		tfMap["outbound_flow_id"] = aws.ToString(v)
	}

	return []any{tfMap}
}
