// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	AssociateRoutingProfileQueuesMaxItems    = 10
	DisassociateRoutingProfileQueuesMaxItems = 10
	CreateRoutingProfileQueuesMaxItems       = 10
)

// @SDKResource("aws_connect_routing_profile", name="Routing Profile")
// @Tags(identifierAttribute="arn")
func ResourceRoutingProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRoutingProfileCreate,
		ReadWithoutTimeout:   resourceRoutingProfileRead,
		UpdateWithoutTimeout: resourceRoutingProfileUpdate,
		DeleteWithoutTimeout: resourceRoutingProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_outbound_queue_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			"instance_id": {
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
			"name": {
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
						"priority": {
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

func resourceRoutingProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)
	input := &connect.CreateRoutingProfileInput{
		DefaultOutboundQueueId: aws.String(d.Get("default_outbound_queue_id").(string)),
		Description:            aws.String(d.Get("description").(string)),
		InstanceId:             aws.String(instanceID),
		MediaConcurrencies:     expandRoutingProfileMediaConcurrencies(d.Get("media_concurrencies").(*schema.Set).List()),
		Name:                   aws.String(name),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("queue_configs"); ok && v.(*schema.Set).Len() > 0 && v.(*schema.Set).Len() <= CreateRoutingProfileQueuesMaxItems {
		input.QueueConfigs = expandRoutingProfileQueueConfigs(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Creating Connect Routing Profile %+v", input)
	output, err := conn.CreateRoutingProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Routing Profile (%s): %s", name, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Routing Profile (%s): empty output", name)
	}

	// call the batched association API if the number of queues to associate with the routing profile is > CreateRoutingProfileQueuesMaxItems
	if v, ok := d.GetOk("queue_configs"); ok && v.(*schema.Set).Len() > CreateRoutingProfileQueuesMaxItems {
		queueConfigsUpdateRemove := make([]interface{}, 0)
		err = updateQueueConfigs(ctx, conn, instanceID, aws.ToString(output.RoutingProfileId), v.(*schema.Set).List(), queueConfigsUpdateRemove)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.ToString(output.RoutingProfileId)))

	return append(diags, resourceRoutingProfileRead(ctx, d, meta)...)
}

func resourceRoutingProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, routingProfileID, err := RoutingProfileParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := conn.DescribeRoutingProfile(ctx, &connect.DescribeRoutingProfileInput{
		InstanceId:       aws.String(instanceID),
		RoutingProfileId: aws.String(routingProfileID),
	})

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Connect Routing Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Routing Profile (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.RoutingProfile == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Routing Profile (%s): empty response", d.Id())
	}

	routingProfile := resp.RoutingProfile

	if err := d.Set("media_concurrencies", flattenRoutingProfileMediaConcurrencies(routingProfile.MediaConcurrencies)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("arn", routingProfile.RoutingProfileArn)
	d.Set("default_outbound_queue_id", routingProfile.DefaultOutboundQueueId)
	d.Set("description", routingProfile.Description)
	d.Set("instance_id", instanceID)
	d.Set("name", routingProfile.Name)

	d.Set("routing_profile_id", routingProfile.RoutingProfileId)

	// getting the routing profile queues uses a separate API: ListRoutingProfileQueues
	queueConfigs, err := getRoutingProfileQueueConfigs(ctx, conn, instanceID, routingProfileID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "finding Connect Routing Profile Queue Configs Summary by Routing Profile ID (%s): %s", routingProfileID, err)
	}

	d.Set("queue_configs", queueConfigs)

	setTagsOut(ctx, resp.RoutingProfile.Tags)

	return diags
}

func resourceRoutingProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, routingProfileID, err := RoutingProfileParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// RoutingProfile has 4 update APIs
	// UpdateRoutingProfileConcurrency: Updates the channels that agents can handle in the Contact Control Panel (CCP) for a routing profile.
	// UpdateRoutingProfileDefaultOutboundQueue: Updates the default outbound queue of a routing profile.
	// UpdateRoutingProfileName: Updates the name and description of a routing profile.
	// UpdateRoutingProfileQueues: Updates the properties associated with a set of queues for a routing profile.

	// updates to concurrency
	inputConcurrency := &connect.UpdateRoutingProfileConcurrencyInput{
		InstanceId:       aws.String(instanceID),
		RoutingProfileId: aws.String(routingProfileID),
	}

	if d.HasChange("media_concurrencies") {
		mediaConcurrencies := expandRoutingProfileMediaConcurrencies(d.Get("media_concurrencies").(*schema.Set).List())
		inputConcurrency.MediaConcurrencies = mediaConcurrencies
		_, err = conn.UpdateRoutingProfileConcurrency(ctx, inputConcurrency)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RoutingProfile Media Concurrency (%s): %s", d.Id(), err)
		}
	}

	// updates to default outbound queue id
	inputDefaultOutboundQueue := &connect.UpdateRoutingProfileDefaultOutboundQueueInput{
		InstanceId:       aws.String(instanceID),
		RoutingProfileId: aws.String(routingProfileID),
	}

	if d.HasChange("default_outbound_queue_id") {
		inputDefaultOutboundQueue.DefaultOutboundQueueId = aws.String(d.Get("default_outbound_queue_id").(string))
		_, err = conn.UpdateRoutingProfileDefaultOutboundQueue(ctx, inputDefaultOutboundQueue)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RoutingProfile Default Outbound Queue ID (%s): %s", d.Id(), err)
		}
	}

	// updates to name and/or description
	inputNameDesc := &connect.UpdateRoutingProfileNameInput{
		InstanceId:       aws.String(instanceID),
		RoutingProfileId: aws.String(routingProfileID),
	}

	if d.HasChanges("name", "description") {
		inputNameDesc.Name = aws.String(d.Get("name").(string))
		inputNameDesc.Description = aws.String(d.Get("description").(string))
		_, err = conn.UpdateRoutingProfileName(ctx, inputNameDesc)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RoutingProfile Name (%s): %s", d.Id(), err)
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

		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		queueConfigsUpdateAdd := ns.Difference(os).List()
		queueConfigsUpdateRemove := os.Difference(ns).List()

		err = updateQueueConfigs(ctx, conn, instanceID, routingProfileID, queueConfigsUpdateAdd, queueConfigsUpdateRemove)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRoutingProfileRead(ctx, d, meta)...)
}

func updateQueueConfigs(ctx context.Context, conn *connect.Client, instanceID, routingProfileID string, queueConfigsUpdateAdd, queueConfigsUpdateRemove []interface{}) error {
	// updates to queue configs
	// There are 3 APIs for this
	// AssociateRoutingProfileQueues - Associates a set of queues with a routing profile.
	// DisassociateRoutingProfileQueues - Disassociates a set of queues from a routing profile.
	// UpdateRoutingProfileQueues - Updates the properties associated with a set of queues for a routing profile.
	// since the update only updates the existing queues that are associated, we will instead disassociate and associate
	// the respective queues based on the diff detected

	// disassociate first since Queue and channel type combination cannot be duplicated
	if len(queueConfigsUpdateRemove) > 0 {
		for i := 0; i < len(queueConfigsUpdateRemove); i += DisassociateRoutingProfileQueuesMaxItems {
			j := i + DisassociateRoutingProfileQueuesMaxItems
			if j > len(queueConfigsUpdateRemove) {
				j = len(queueConfigsUpdateRemove)
			}
			_, err := conn.DisassociateRoutingProfileQueues(ctx, &connect.DisassociateRoutingProfileQueuesInput{
				InstanceId:       aws.String(instanceID),
				QueueReferences:  expandRoutingProfileQueueReferences(queueConfigsUpdateRemove[i:j]),
				RoutingProfileId: aws.String(routingProfileID),
			})
			if err != nil {
				return fmt.Errorf("updating RoutingProfile Queue Configs, specifically disassociating queues from routing profile (%s): %s", routingProfileID, err)
			}
		}
	}

	if len(queueConfigsUpdateAdd) > 0 {
		for i := 0; i < len(queueConfigsUpdateAdd); i += AssociateRoutingProfileQueuesMaxItems {
			j := i + AssociateRoutingProfileQueuesMaxItems
			if j > len(queueConfigsUpdateAdd) {
				j = len(queueConfigsUpdateAdd)
			}
			_, err := conn.AssociateRoutingProfileQueues(ctx, &connect.AssociateRoutingProfileQueuesInput{
				InstanceId:       aws.String(instanceID),
				QueueConfigs:     expandRoutingProfileQueueConfigs(queueConfigsUpdateAdd[i:j]),
				RoutingProfileId: aws.String(routingProfileID),
			})
			if err != nil {
				return fmt.Errorf("updating RoutingProfile Queue Configs, specifically associating queues to routing profile (%s): %s", routingProfileID, err)
			}
		}
	}

	return nil
}

func resourceRoutingProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, routingProfileID, err := RoutingProfileParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteRoutingProfile(ctx, &connect.DeleteRoutingProfileInput{
		InstanceId:       aws.String(instanceID),
		RoutingProfileId: aws.String(routingProfileID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RoutingProfile (%s): %s", d.Id(), err)
	}

	return diags
}

func expandRoutingProfileMediaConcurrencies(mediaConcurrencies []interface{}) []awstypes.MediaConcurrency {
	if len(mediaConcurrencies) == 0 {
		return nil
	}

	mediaConcurrenciesExpanded := []awstypes.MediaConcurrency{}

	for _, mediaConcurrency := range mediaConcurrencies {
		data := mediaConcurrency.(map[string]interface{})
		mediaConcurrencyExpanded := awstypes.MediaConcurrency{
			Channel:     awstypes.Channel(data["channel"].(string)),
			Concurrency: aws.Int32(int32(data["concurrency"].(int))),
		}
		mediaConcurrenciesExpanded = append(mediaConcurrenciesExpanded, mediaConcurrencyExpanded)
	}

	return mediaConcurrenciesExpanded
}

func flattenRoutingProfileMediaConcurrencies(mediaConcurrencies []awstypes.MediaConcurrency) []interface{} {
	mediaConcurrenciesList := []interface{}{}

	for _, mediaConcurrency := range mediaConcurrencies {
		values := map[string]interface{}{
			"channel":     string(mediaConcurrency.Channel),
			"concurrency": aws.ToInt32(mediaConcurrency.Concurrency),
		}

		mediaConcurrenciesList = append(mediaConcurrenciesList, values)
	}
	return mediaConcurrenciesList
}

func expandRoutingProfileQueueConfigs(queueConfigs []interface{}) []awstypes.RoutingProfileQueueConfig {
	if len(queueConfigs) == 0 {
		return nil
	}

	queueConfigsExpanded := []awstypes.RoutingProfileQueueConfig{}

	for _, queueConfig := range queueConfigs {
		data := queueConfig.(map[string]interface{})
		queueConfigExpanded := awstypes.RoutingProfileQueueConfig{
			Delay:    aws.Int32(int32(data["delay"].(int))),
			Priority: aws.Int32(int32(data["priority"].(int))),
		}

		qr := awstypes.RoutingProfileQueueReference{
			Channel: awstypes.Channel(data["channel"].(string)),
			QueueId: aws.String(data["queue_id"].(string)),
		}
		queueConfigExpanded.QueueReference = &qr

		queueConfigsExpanded = append(queueConfigsExpanded, queueConfigExpanded)
	}

	return queueConfigsExpanded
}

func expandRoutingProfileQueueReferences(queueConfigs []interface{}) []awstypes.RoutingProfileQueueReference {
	if len(queueConfigs) == 0 {
		return nil
	}

	queueReferencesExpanded := []awstypes.RoutingProfileQueueReference{}

	for _, queueConfig := range queueConfigs {
		data := queueConfig.(map[string]interface{})
		queueReferenceExpanded := awstypes.RoutingProfileQueueReference{
			Channel: awstypes.Channel(data["channel"].(string)),
			QueueId: aws.String(data["queue_id"].(string)),
		}

		queueReferencesExpanded = append(queueReferencesExpanded, queueReferenceExpanded)
	}

	return queueReferencesExpanded
}

func getRoutingProfileQueueConfigs(ctx context.Context, conn *connect.Client, instanceID, routingProfileID string) ([]interface{}, error) {
	queueConfigsList := []interface{}{}

	input := &connect.ListRoutingProfileQueuesInput{
		InstanceId:       aws.String(instanceID),
		MaxResults:       aws.Int32(ListRoutingProfileQueuesMaxResults),
		RoutingProfileId: aws.String(routingProfileID),
	}

	pages := connect.NewListRoutingProfileQueuesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, qc := range page.RoutingProfileQueueConfigSummaryList {
			values := map[string]interface{}{
				"channel":    string(qc.Channel),
				"delay":      qc.Delay,
				"priority":   aws.ToInt32(qc.Priority),
				"queue_arn":  aws.ToString(qc.QueueArn),
				"queue_id":   aws.ToString(qc.QueueId),
				"queue_name": aws.ToString(qc.QueueName),
			}

			queueConfigsList = append(queueConfigsList, values)
		}
	}

	return queueConfigsList, nil
}

func RoutingProfileParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:routingProfileID", id)
	}

	return parts[0], parts[1], nil
}
