// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_event_subscription", name="Event Subscription")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceEventSubscription() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_aws_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"event_categories": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validEventSubscriptionName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validEventSubscriptionName,
			},
			"sns_topic": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"source_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSourceType: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.SourceType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEventSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &rds.CreateEventSubscriptionInput{
		Enabled:          aws.Bool(d.Get(names.AttrEnabled).(bool)),
		SnsTopicArn:      aws.String(d.Get("sns_topic").(string)),
		SubscriptionName: aws.String(name),
		Tags:             getTagsInV2(ctx),
	}

	if v, ok := d.GetOk("event_categories"); ok && v.(*schema.Set).Len() > 0 {
		input.EventCategories = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("source_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SourceIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSourceType); ok {
		input.SourceType = aws.String(v.(string))
	}

	output, err := conn.CreateEventSubscription(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Event Subscription (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.EventSubscription.CustSubscriptionId))

	if _, err := waitEventSubscriptionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Event Subscription (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	sub, err := findEventSubscriptionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Event Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Event Subscription (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, sub.EventSubscriptionArn)
	d.Set("customer_aws_id", sub.CustomerAwsId)
	d.Set(names.AttrEnabled, sub.Enabled)
	d.Set("event_categories", sub.EventCategoriesList)
	d.Set(names.AttrName, sub.CustSubscriptionId)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(sub.CustSubscriptionId)))
	d.Set("sns_topic", sub.SnsTopicArn)
	d.Set("source_ids", sub.SourceIdsList)
	d.Set(names.AttrSourceType, sub.SourceType)

	return diags
}

func resourceEventSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "source_ids") {
		input := &rds.ModifyEventSubscriptionInput{
			SubscriptionName: aws.String(d.Id()),
		}

		input.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))

		if d.HasChange("event_categories") {
			input.EventCategories = flex.ExpandStringValueSet(d.Get("event_categories").(*schema.Set))
			input.SourceType = aws.String(d.Get(names.AttrSourceType).(string))
		}

		if d.HasChange(names.AttrSourceType) {
			input.SourceType = aws.String(d.Get(names.AttrSourceType).(string))
		}

		if d.HasChange("sns_topic") {
			input.SnsTopicArn = aws.String(d.Get("sns_topic").(string))
		}

		_, err := conn.ModifyEventSubscription(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Event Subscription (%s): %s", d.Id(), err)
		}

		if _, err := waitEventSubscriptionUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Event Subscription (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("source_ids") {
		o, n := d.GetChange("source_ids")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

		for _, del := range del {
			input := &rds.RemoveSourceIdentifierFromSubscriptionInput{
				SourceIdentifier: aws.String(del),
				SubscriptionName: aws.String(d.Id()),
			}

			_, err := conn.RemoveSourceIdentifierFromSubscription(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "removing RDS Event Subscription (%s) source ID (%s): %s", d.Id(), del, err)
			}
		}

		for _, add := range add {
			input := &rds.AddSourceIdentifierToSubscriptionInput{
				SourceIdentifier: aws.String(add),
				SubscriptionName: aws.String(d.Id()),
			}

			_, err := conn.AddSourceIdentifierToSubscription(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "adding RDS Event Subscription (%s) source ID (%s): %s", d.Id(), add, err)
			}
		}
	}

	return diags
}

func resourceEventSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting RDS Event Subscription: (%s)", d.Id())
	_, err := conn.DeleteEventSubscription(ctx, &rds.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	})

	if errs.IsA[*types.SubscriptionNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Event Subscription (%s): %s", d.Id(), err)
	}

	if _, err := waitEventSubscriptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Event Subscription (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findEventSubscriptionByID(ctx context.Context, conn *rds.Client, id string) (*types.EventSubscription, error) {
	input := &rds.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(id),
	}
	output, err := findEventSubscription(ctx, conn, input, tfslices.PredicateTrue[*types.EventSubscription]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.CustSubscriptionId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEventSubscription(ctx context.Context, conn *rds.Client, input *rds.DescribeEventSubscriptionsInput, filter tfslices.Predicate[*types.EventSubscription]) (*types.EventSubscription, error) {
	output, err := findEventSubscriptions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEventSubscriptions(ctx context.Context, conn *rds.Client, input *rds.DescribeEventSubscriptionsInput, filter tfslices.Predicate[*types.EventSubscription]) ([]types.EventSubscription, error) {
	var output []types.EventSubscription

	pages := rds.NewDescribeEventSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.SubscriptionNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.EventSubscriptionsList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusEventSubscription(ctx context.Context, conn *rds.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findEventSubscriptionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitEventSubscriptionCreated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.EventSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{eventSubscriptionStatusCreating},
		Target:     []string{eventSubscriptionStatusActive},
		Refresh:    statusEventSubscription(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitEventSubscriptionDeleted(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.EventSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{eventSubscriptionStatusDeleting},
		Target:     []string{},
		Refresh:    statusEventSubscription(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitEventSubscriptionUpdated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.EventSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{eventSubscriptionStatusModifying},
		Target:                    []string{eventSubscriptionStatusActive},
		Refresh:                   statusEventSubscription(ctx, conn, id),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		Delay:                     30 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.EventSubscription); ok {
		return output, err
	}

	return nil, err
}
