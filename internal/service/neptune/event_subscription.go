// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_neptune_event_subscription", name="Event Subscription")
// @Tags(identifierAttribute="arn")
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
				ValidateFunc:  validEventSubscriptionNamePrefix,
			},
			names.AttrSNSTopicARN: {
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
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceEventSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrName).(string)),
		create.WithConfiguredPrefix(d.Get(names.AttrNamePrefix).(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()
	input := &neptune.CreateEventSubscriptionInput{
		Enabled:          aws.Bool(d.Get(names.AttrEnabled).(bool)),
		SnsTopicArn:      aws.String(d.Get(names.AttrSNSTopicARN).(string)),
		SubscriptionName: aws.String(name),
		Tags:             getTagsIn(ctx),
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
		return sdkdiag.AppendErrorf(diags, "creating Neptune Event Subscription (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.EventSubscription.CustSubscriptionId))

	if _, err := waitEventSubscriptionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Event Subscription (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	output, err := findEventSubscriptionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Neptune Event Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Event Subscription (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.EventSubscriptionArn)
	d.Set("customer_aws_id", output.CustomerAwsId)
	d.Set(names.AttrEnabled, output.Enabled)
	d.Set("event_categories", output.EventCategoriesList)
	d.Set(names.AttrName, output.CustSubscriptionId)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.CustSubscriptionId)))
	d.Set(names.AttrSNSTopicARN, output.SnsTopicArn)
	d.Set("source_ids", output.SourceIdsList)
	d.Set(names.AttrSourceType, output.SourceType)

	return diags
}

func resourceEventSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "source_ids") {
		input := &neptune.ModifyEventSubscriptionInput{
			SubscriptionName: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrEnabled) {
			input.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))
		}

		if d.HasChange("event_categories") {
			input.EventCategories = flex.ExpandStringValueSet(d.Get("event_categories").(*schema.Set))
			input.SourceType = aws.String(d.Get(names.AttrSourceType).(string))
		}

		if d.HasChange(names.AttrSNSTopicARN) {
			input.SnsTopicArn = aws.String(d.Get(names.AttrSNSTopicARN).(string))
		}

		if d.HasChange(names.AttrSourceType) {
			input.SourceType = aws.String(d.Get(names.AttrSourceType).(string))
		}

		_, err := conn.ModifyEventSubscription(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Event Subscription (%s): %s", d.Id(), err)
		}

		if _, err := waitEventSubscriptionUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Neptune Event Subscription (%s) update: %s", d.Id(), err)
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
			for _, v := range remove {
				_, err := conn.RemoveSourceIdentifierFromSubscription(ctx, &neptune.RemoveSourceIdentifierFromSubscriptionInput{
					SourceIdentifier: v,
					SubscriptionName: aws.String(d.Id()),
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "removing Neptune Event Subscription (%s) source identifier: %s", d.Id(), err)
				}
			}
		}

		if len(add) > 0 {
			for _, v := range add {
				_, err := conn.AddSourceIdentifierToSubscription(ctx, &neptune.AddSourceIdentifierToSubscriptionInput{
					SourceIdentifier: v,
					SubscriptionName: aws.String(d.Id()),
				})

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "adding Neptune Event Subscription (%s) source identifier: %s", d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	log.Printf("[DEBUG] Deleting Neptune Event Subscription: %s", d.Id())
	_, err := conn.DeleteEventSubscription(ctx, &neptune.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.SubscriptionNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Event Subscription (%s): %s", d.Id(), err)
	}

	if _, err := waitEventSubscriptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Event Subscription (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findEventSubscriptionByName(ctx context.Context, conn *neptune.Client, name string) (*awstypes.EventSubscription, error) {
	input := &neptune.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(name),
	}
	output, err := findEventSubscription(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.CustSubscriptionId) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEventSubscription(ctx context.Context, conn *neptune.Client, input *neptune.DescribeEventSubscriptionsInput) (*awstypes.EventSubscription, error) {
	output, err := findEventSubscriptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEventSubscriptions(ctx context.Context, conn *neptune.Client, input *neptune.DescribeEventSubscriptionsInput) ([]awstypes.EventSubscription, error) {
	var output []awstypes.EventSubscription

	pages := neptune.NewDescribeEventSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.SubscriptionNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.EventSubscriptionsList...)
	}

	return output, nil
}

func statusEventSubscription(ctx context.Context, conn *neptune.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findEventSubscriptionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitEventSubscriptionCreated(ctx context.Context, conn *neptune.Client, name string, timeout time.Duration) (*awstypes.EventSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{eventSubscriptionStatusCreating},
		Target:     []string{eventSubscriptionStatusActive},
		Refresh:    statusEventSubscription(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitEventSubscriptionUpdated(ctx context.Context, conn *neptune.Client, name string, timeout time.Duration) (*awstypes.EventSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{eventSubscriptionStatusModifying},
		Target:     []string{eventSubscriptionStatusActive},
		Refresh:    statusEventSubscription(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitEventSubscriptionDeleted(ctx context.Context, conn *neptune.Client, name string, timeout time.Duration) (*awstypes.EventSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{eventSubscriptionStatusDeleting},
		Target:     []string{},
		Refresh:    statusEventSubscription(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.EventSubscription); ok {
		return output, err
	}

	return nil, err
}
