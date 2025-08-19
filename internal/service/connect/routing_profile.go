// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	routingProfileQueueAssociationChunkSize = 10
)

// @SDKResource("aws_connect_routing_profile", name="Routing Profile")
// @Tags(identifierAttribute="arn")
func resourceRoutingProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRoutingProfileCreate,
		ReadWithoutTimeout:   resourceRoutingProfileRead,
		UpdateWithoutTimeout: resourceRoutingProfileUpdate,
		DeleteWithoutTimeout: resourceRoutingProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_outbound_queue_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"media_concurrencies": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Channel](),
						},
						"concurrency": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 10),
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 127),
			},
			"queue_configs": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Channel](),
						},
						"delay": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 9999),
						},
						names.AttrPriority: {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 99),
						},
						"queue_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"queue_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"queue_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"routing_profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRoutingProfileCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	input := &connect.CreateRoutingProfileInput{
		DefaultOutboundQueueId: aws.String(d.Get("default_outbound_queue_id").(string)),
		Description:            aws.String(d.Get(names.AttrDescription).(string)),
		InstanceId:             aws.String(instanceID),
		MediaConcurrencies:     expandMediaConcurrencies(d.Get("media_concurrencies").(*schema.Set).List()),
		Name:                   aws.String(name),
		Tags:                   getTagsIn(ctx),
	}

	var queueConfigs []awstypes.RoutingProfileQueueConfig
	if v, ok := d.GetOk("queue_configs"); ok && v.(*schema.Set).Len() > 0 {
		queueConfigs = expandRoutingProfileQueueConfigs(v.(*schema.Set).List())
	}

	if len(queueConfigs) <= routingProfileQueueAssociationChunkSize {
		input.QueueConfigs = queueConfigs
	}

	output, err := conn.CreateRoutingProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Routing Profile (%s): %s", name, err)
	}

	routingProfileID := aws.ToString(output.RoutingProfileId)
	id := routingProfileCreateResourceID(instanceID, routingProfileID)
	d.SetId(id)

	// call the batched association API if the number of queues to associate with the routing profile is > CreateRoutingProfileQueuesMaxItems
	if len(queueConfigs) > routingProfileQueueAssociationChunkSize {
		if err := updateRoutingProfileQueueAssociations(ctx, conn, instanceID, routingProfileID, queueConfigs, nil); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRoutingProfileRead(ctx, d, meta)...)
}

func resourceRoutingProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, routingProfileID, err := routingProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	routingProfile, err := findRoutingProfileByTwoPartKey(ctx, conn, instanceID, routingProfileID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Routing Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Routing Profile (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, routingProfile.RoutingProfileArn)
	d.Set("default_outbound_queue_id", routingProfile.DefaultOutboundQueueId)
	d.Set(names.AttrDescription, routingProfile.Description)
	d.Set(names.AttrInstanceID, instanceID)
	if err := d.Set("media_concurrencies", flattenMediaConcurrencies(routingProfile.MediaConcurrencies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting media_concurrencies: %s", err)
	}
	d.Set(names.AttrName, routingProfile.Name)
	d.Set("routing_profile_id", routingProfile.RoutingProfileId)

	queueConfigs, err := findRoutingConfigQueueConfigSummariesByTwoPartKey(ctx, conn, instanceID, routingProfileID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Routing Profile (%s) Queue Config summaries: %s", d.Id(), err)
	}

	if err := d.Set("queue_configs", flattenRoutingConfigQueueConfigSummaries(queueConfigs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting queue_configs: %s", err)
	}

	setTagsOut(ctx, routingProfile.Tags)

	return diags
}

func resourceRoutingProfileUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, routingProfileID, err := routingProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// RoutingProfile has 4 update APIs
	// UpdateRoutingProfileConcurrency: Updates the channels that agents can handle in the Contact Control Panel (CCP) for a routing profile.
	// UpdateRoutingProfileDefaultOutboundQueue: Updates the default outbound queue of a routing profile.
	// UpdateRoutingProfileName: Updates the name and description of a routing profile.
	// UpdateRoutingProfileQueues: Updates the properties associated with a set of queues for a routing profile.

	if d.HasChange("media_concurrencies") {
		// updates to concurrency
		input := &connect.UpdateRoutingProfileConcurrencyInput{
			InstanceId:         aws.String(instanceID),
			MediaConcurrencies: expandMediaConcurrencies(d.Get("media_concurrencies").(*schema.Set).List()),
			RoutingProfileId:   aws.String(routingProfileID),
		}

		_, err = conn.UpdateRoutingProfileConcurrency(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Routing Profile (%s) Concurrency: %s", d.Id(), err)
		}
	}

	if d.HasChange("default_outbound_queue_id") {
		// updates to default outbound queue id
		input := &connect.UpdateRoutingProfileDefaultOutboundQueueInput{
			DefaultOutboundQueueId: aws.String(d.Get("default_outbound_queue_id").(string)),
			InstanceId:             aws.String(instanceID),
			RoutingProfileId:       aws.String(routingProfileID),
		}

		_, err = conn.UpdateRoutingProfileDefaultOutboundQueue(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Routing Profile (%s) DefaultOutboundQueue: %s", d.Id(), err)
		}
	}

	if d.HasChanges(names.AttrName, names.AttrDescription) {
		// updates to name and/or description
		input := &connect.UpdateRoutingProfileNameInput{
			Description:      aws.String(d.Get(names.AttrDescription).(string)),
			InstanceId:       aws.String(instanceID),
			Name:             aws.String(d.Get(names.AttrName).(string)),
			RoutingProfileId: aws.String(routingProfileID),
		}

		_, err = conn.UpdateRoutingProfileName(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Routing Profile (%s) Name: %s", d.Id(), err)
		}
	}

	// updates to queue configs
	// There are 3 APIs for this
	// AssociateRoutingProfileQueues - Associates a set of queues with a routing profile.
	// DisassociateRoutingProfileQueues - Disassociates a set of queues from a routing profile.
	// UpdateRoutingProfileQueues - Updates the properties associated with a set of queues for a routing profile.
	// since the update only updates the existing queues that are associated, we will instead disassociate and associate
	// the respective queues based on the diff detected
	if d.HasChange("queue_configs") {
		o, n := d.GetChange("queue_configs")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := expandRoutingProfileQueueConfigs(ns.Difference(os).List()), expandRoutingProfileQueueConfigs(os.Difference(ns).List())

		if err := updateRoutingProfileQueueAssociations(ctx, conn, instanceID, routingProfileID, add, del); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRoutingProfileRead(ctx, d, meta)...)
}

func resourceRoutingProfileDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, routingProfileID, err := routingProfileParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Routing Profile: %s", d.Id())
	input := connect.DeleteRoutingProfileInput{
		InstanceId:       aws.String(instanceID),
		RoutingProfileId: aws.String(routingProfileID),
	}
	_, err = conn.DeleteRoutingProfile(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Routing Profile (%s): %s", d.Id(), err)
	}

	return diags
}

const routingProfileResourceIDSeparator = ":"

func routingProfileCreateResourceID(instanceID, routingProfileID string) string {
	parts := []string{instanceID, routingProfileID}
	id := strings.Join(parts, routingProfileResourceIDSeparator)

	return id
}

func routingProfileParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, routingProfileResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]sroutingProfileID", id, routingProfileResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func updateRoutingProfileQueueAssociations(ctx context.Context, conn *connect.Client, instanceID, routingProfileID string, add, del []awstypes.RoutingProfileQueueConfig) error {
	// updates to queue configs
	// There are 3 APIs for this
	// AssociateRoutingProfileQueues - Associates a set of queues with a routing profile.
	// DisassociateRoutingProfileQueues - Disassociates a set of queues from a routing profile.
	// UpdateRoutingProfileQueues - Updates the properties associated with a set of queues for a routing profile.
	// since the update only updates the existing queues that are associated, we will instead disassociate and associate
	// the respective queues based on the diff detected

	// disassociate first since Queue and channel type combination cannot be duplicated
	for chunk := range slices.Chunk(del, routingProfileQueueAssociationChunkSize) {
		var queueReferences []awstypes.RoutingProfileQueueReference
		for _, v := range chunk {
			if v := v.QueueReference; v != nil {
				queueReferences = append(queueReferences, *v)
			}
		}

		if len(queueReferences) > 0 {
			input := &connect.DisassociateRoutingProfileQueuesInput{
				InstanceId:       aws.String(instanceID),
				QueueReferences:  queueReferences,
				RoutingProfileId: aws.String(routingProfileID),
			}

			_, err := conn.DisassociateRoutingProfileQueues(ctx, input)

			if err != nil {
				return fmt.Errorf("disassociating Connect Routing Profile (%s) queues: %s", routingProfileID, err)
			}
		}
	}

	for chunk := range slices.Chunk(add, routingProfileQueueAssociationChunkSize) {
		input := &connect.AssociateRoutingProfileQueuesInput{
			InstanceId:       aws.String(instanceID),
			QueueConfigs:     chunk,
			RoutingProfileId: aws.String(routingProfileID),
		}

		_, err := conn.AssociateRoutingProfileQueues(ctx, input)

		if err != nil {
			return fmt.Errorf("associating Connect Routing Profile (%s) queues: %s", routingProfileID, err)
		}
	}

	return nil
}

func findRoutingProfileByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, routingProfileID string) (*awstypes.RoutingProfile, error) {
	input := &connect.DescribeRoutingProfileInput{
		InstanceId:       aws.String(instanceID),
		RoutingProfileId: aws.String(routingProfileID),
	}

	return findRoutingProfile(ctx, conn, input)
}

func findRoutingProfile(ctx context.Context, conn *connect.Client, input *connect.DescribeRoutingProfileInput) (*awstypes.RoutingProfile, error) {
	output, err := conn.DescribeRoutingProfile(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RoutingProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RoutingProfile, nil
}

func findRoutingConfigQueueConfigSummariesByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, routingProfileID string) ([]awstypes.RoutingProfileQueueConfigSummary, error) {
	const maxResults = 60
	input := &connect.ListRoutingProfileQueuesInput{
		InstanceId:       aws.String(instanceID),
		MaxResults:       aws.Int32(maxResults),
		RoutingProfileId: aws.String(routingProfileID),
	}

	return findRoutingConfigQueueConfigSummaries(ctx, conn, input)
}

func findRoutingConfigQueueConfigSummaries(ctx context.Context, conn *connect.Client, input *connect.ListRoutingProfileQueuesInput) ([]awstypes.RoutingProfileQueueConfigSummary, error) {
	var output []awstypes.RoutingProfileQueueConfigSummary

	pages := connect.NewListRoutingProfileQueuesPaginator(conn, input)
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

		output = append(output, page.RoutingProfileQueueConfigSummaryList...)
	}

	return output, nil
}

func expandMediaConcurrencies(tfList []any) []awstypes.MediaConcurrency {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []awstypes.MediaConcurrency{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.MediaConcurrency{
			Channel:     awstypes.Channel(tfMap["channel"].(string)),
			Concurrency: aws.Int32(int32(tfMap["concurrency"].(int))),
		}
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenMediaConcurrencies(apiObjects []awstypes.MediaConcurrency) []any {
	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"channel":     apiObject.Channel,
			"concurrency": aws.ToInt32(apiObject.Concurrency),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandRoutingProfileQueueConfigs(tfList []any) []awstypes.RoutingProfileQueueConfig {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []awstypes.RoutingProfileQueueConfig{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.RoutingProfileQueueConfig{
			Delay:    aws.Int32(int32(tfMap["delay"].(int))),
			Priority: aws.Int32(int32(tfMap[names.AttrPriority].(int))),
			QueueReference: &awstypes.RoutingProfileQueueReference{
				Channel: awstypes.Channel(tfMap["channel"].(string)),
				QueueId: aws.String(tfMap["queue_id"].(string)),
			},
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenRoutingConfigQueueConfigSummaries(apiObjects []awstypes.RoutingProfileQueueConfigSummary) []any {
	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"channel":          apiObject.Channel,
			"delay":            apiObject.Delay,
			names.AttrPriority: aws.ToInt32(apiObject.Priority),
			"queue_arn":        aws.ToString(apiObject.QueueArn),
			"queue_id":         aws.ToString(apiObject.QueueId),
			"queue_name":       aws.ToString(apiObject.QueueName),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
