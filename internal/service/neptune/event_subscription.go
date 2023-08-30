// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_neptune_event_subscription", name="Event Subscription")
// @Tags(identifierAttribute="arn")
func ResourceEventSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventSubscriptionCreate,
		ReadWithoutTimeout:   resourceEventSubscriptionRead,
		UpdateWithoutTimeout: resourceEventSubscriptionUpdate,
		DeleteWithoutTimeout: resourceEventSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validEventSubscriptionName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validEventSubscriptionNamePrefix,
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"event_categories": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"source_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"source_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"customer_aws_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEventSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	if v, ok := d.GetOk("name"); ok {
		d.Set("name", v.(string))
	} else if v, ok := d.GetOk("name_prefix"); ok {
		d.Set("name", id.PrefixedUniqueId(v.(string)))
	} else {
		d.Set("name", id.PrefixedUniqueId("tf-"))
	}

	input := &neptune.CreateEventSubscriptionInput{
		SubscriptionName: aws.String(d.Get("name").(string)),
		SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("source_ids"); ok {
		sourceIdsSet := v.(*schema.Set)
		sourceIds := make([]*string, sourceIdsSet.Len())
		for i, sourceId := range sourceIdsSet.List() {
			sourceIds[i] = aws.String(sourceId.(string))
		}
		input.SourceIds = sourceIds
	}

	if v, ok := d.GetOk("event_categories"); ok {
		eventCategoriesSet := v.(*schema.Set)
		eventCategories := make([]*string, eventCategoriesSet.Len())
		for i, eventCategory := range eventCategoriesSet.List() {
			eventCategories[i] = aws.String(eventCategory.(string))
		}
		input.EventCategories = eventCategories
	}

	if v, ok := d.GetOk("source_type"); ok {
		input.SourceType = aws.String(v.(string))
	}

	log.Println("[DEBUG] Create Neptune Event Subscription:", input)

	output, err := conn.CreateEventSubscriptionWithContext(ctx, input)
	if err != nil || output.EventSubscription == nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Event Subscription %s: %s", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(output.EventSubscription.CustSubscriptionId))

	log.Println("[INFO] Waiting for Neptune Event Subscription to be ready")

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active"},
		Refresh:    resourceEventSubscriptionRefreshFunc(ctx, d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Event Subscription state to be \"active\": %s", err)
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	sub, err := resourceEventSubscriptionRetrieve(ctx, d.Id(), conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Event Subscription %s: %s", d.Id(), err)
	}

	if sub == nil {
		log.Printf("[DEBUG] Neptune Event Subscription (%s) not found - removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", sub.EventSubscriptionArn)
	d.Set("name", sub.CustSubscriptionId)
	d.Set("sns_topic_arn", sub.SnsTopicArn)
	d.Set("enabled", sub.Enabled)
	d.Set("customer_aws_id", sub.CustomerAwsId)
	d.Set("source_type", sub.SourceType)

	if sub.SourceIdsList != nil {
		if err := d.Set("source_ids", flex.FlattenStringList(sub.SourceIdsList)); err != nil {
			return sdkdiag.AppendErrorf(diags, "saving Source IDs to state for Neptune Event Subscription (%s): %s", d.Id(), err)
		}
	}

	if sub.EventCategoriesList != nil {
		if err := d.Set("event_categories", flex.FlattenStringList(sub.EventCategoriesList)); err != nil {
			return sdkdiag.AppendErrorf(diags, "saving Event Categories to state for Neptune Event Subscription (%s): %s", d.Id(), err)
		}
	}

	return diags
}

func resourceEventSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	requestUpdate := false

	req := &neptune.ModifyEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	}

	if d.HasChange("event_categories") {
		eventCategoriesSet := d.Get("event_categories").(*schema.Set)
		req.EventCategories = make([]*string, eventCategoriesSet.Len())
		for i, eventCategory := range eventCategoriesSet.List() {
			req.EventCategories[i] = aws.String(eventCategory.(string))
		}
		req.SourceType = aws.String(d.Get("source_type").(string))
		requestUpdate = true
	}

	if d.HasChange("enabled") {
		req.Enabled = aws.Bool(d.Get("enabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("sns_topic_arn") {
		req.SnsTopicArn = aws.String(d.Get("sns_topic_arn").(string))
		requestUpdate = true
	}

	if d.HasChange("source_type") {
		req.SourceType = aws.String(d.Get("source_type").(string))
		requestUpdate = true
	}

	log.Printf("[DEBUG] Send Neptune Event Subscription modification request: %#v", requestUpdate)
	if requestUpdate {
		log.Printf("[DEBUG] Neptune Event Subscription modification request: %#v", req)
		_, err := conn.ModifyEventSubscriptionWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Event Subscription (%s): %s", d.Id(), err)
		}

		log.Println("[INFO] Waiting for Neptune Event Subscription modification to finish")

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"modifying"},
			Target:     []string{"active"},
			Refresh:    resourceEventSubscriptionRefreshFunc(ctx, d.Id(), conn),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
			Delay:      30 * time.Second,
		}

		// Wait, catching any errors
		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Event Subscription (%s): waiting for completion: %s", d.Id(), err)
		}
	}

	if d.HasChange("source_ids") {
		o, n := d.GetChange("source_ids")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := flex.ExpandStringSet(os.Difference(ns))
		add := flex.ExpandStringSet(ns.Difference(os))

		if len(remove) > 0 {
			for _, removing := range remove {
				log.Printf("[INFO] Removing %s as a Source Identifier from %q", *removing, d.Id())
				_, err := conn.RemoveSourceIdentifierFromSubscriptionWithContext(ctx, &neptune.RemoveSourceIdentifierFromSubscriptionInput{
					SourceIdentifier: removing,
					SubscriptionName: aws.String(d.Id()),
				})
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Neptune Event Subscription (%s): removing Source Identifier (%s): %s", d.Id(), aws.StringValue(removing), err)
				}
			}
		}

		if len(add) > 0 {
			for _, adding := range add {
				log.Printf("[INFO] Adding %s as a Source Identifier to %q", *adding, d.Id())
				_, err := conn.AddSourceIdentifierToSubscriptionWithContext(ctx, &neptune.AddSourceIdentifierToSubscriptionInput{
					SourceIdentifier: adding,
					SubscriptionName: aws.String(d.Id()),
				})
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Neptune Event Subscription (%s): adding Source Identifier (%s): %s", d.Id(), aws.StringValue(adding), err)
				}
			}
		}
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	log.Printf("[DEBUG] Deleting Neptune Event Subscription: %s", d.Id())
	_, err := conn.DeleteEventSubscriptionWithContext(ctx, &neptune.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeSubscriptionNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Event Subscription (%s): %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting"},
		Target:     []string{},
		Refresh:    resourceEventSubscriptionRefreshFunc(ctx, d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Event Subscription (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceEventSubscriptionRefreshFunc(ctx context.Context, name string, conn *neptune.Neptune) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		sub, err := resourceEventSubscriptionRetrieve(ctx, name, conn)

		if err != nil {
			log.Printf("Error on retrieving Neptune Event Subscription when waiting: %s", err)
			return nil, "", err
		}

		if sub == nil {
			return nil, "", nil
		}

		if sub.Status != nil {
			log.Printf("[DEBUG] Neptune Event Subscription status for %s: %s", name, aws.StringValue(sub.Status))
		}

		return sub, aws.StringValue(sub.Status), nil
	}
}

func resourceEventSubscriptionRetrieve(ctx context.Context, name string, conn *neptune.Neptune) (*neptune.EventSubscription, error) {
	request := &neptune.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(name),
	}

	describeResp, err := conn.DescribeEventSubscriptionsWithContext(ctx, request)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeSubscriptionNotFoundFault) {
			log.Printf("[DEBUG] Neptune Event Subscription (%s) not found", name)
			return nil, nil
		}
		return nil, err
	}

	if len(describeResp.EventSubscriptionsList) != 1 ||
		aws.StringValue(describeResp.EventSubscriptionsList[0].CustSubscriptionId) != name {
		return nil, nil
	}

	return describeResp.EventSubscriptionsList[0], nil
}
