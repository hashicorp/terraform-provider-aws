// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dms_event_subscription", name="Event Subscription")
// @Tags(identifierAttribute="arn")
func ResourceEventSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventSubscriptionCreate,
		ReadWithoutTimeout:   resourceEventSubscriptionRead,
		UpdateWithoutTimeout: resourceEventSubscriptionUpdate,
		DeleteWithoutTimeout: resourceEventSubscriptionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"source_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: true,
			},
			"source_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"replication-instance",
					"replication-task",
				}, false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEventSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &dms.CreateEventSubscriptionInput{
		Enabled:          aws.Bool(d.Get(names.AttrEnabled).(bool)),
		EventCategories:  flex.ExpandStringSet(d.Get("event_categories").(*schema.Set)),
		SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
		SourceType:       aws.String(d.Get("source_type").(string)),
		SubscriptionName: aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("source_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SourceIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	_, err := conn.CreateEventSubscriptionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Event Subscription (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitEventSubscriptionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Event Subscription (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	subscription, err := FindEventSubscriptionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Event Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Event Subscription (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "dms",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("es:%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrEnabled, subscription.Enabled)
	d.Set("event_categories", aws.StringValueSlice(subscription.EventCategoriesList))
	d.Set(names.AttrName, d.Id())
	d.Set("sns_topic_arn", subscription.SnsTopicArn)
	d.Set("source_ids", aws.StringValueSlice(subscription.SourceIdsList))
	d.Set("source_type", subscription.SourceType)

	return diags
}

func resourceEventSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &dms.ModifyEventSubscriptionInput{
			Enabled:          aws.Bool(d.Get(names.AttrEnabled).(bool)),
			EventCategories:  flex.ExpandStringSet(d.Get("event_categories").(*schema.Set)),
			SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
			SourceType:       aws.String(d.Get("source_type").(string)),
			SubscriptionName: aws.String(d.Id()),
		}

		_, err := conn.ModifyEventSubscriptionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DMS Event Subscription (%s): %s", d.Id(), err)
		}

		if _, err := waitEventSubscriptionUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DMS Event Subscription (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	log.Printf("[DEBUG] Deleting DMS Event Subscription: %s", d.Id())
	_, err := conn.DeleteEventSubscriptionWithContext(ctx, &dms.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Event Subscription (%s): %s", d.Id(), err)
	}

	if _, err := waitEventSubscriptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Event Subscription (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindEventSubscriptionByName(ctx context.Context, conn *dms.DatabaseMigrationService, name string) (*dms.EventSubscription, error) {
	input := &dms.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(name),
	}

	return findEventSubscription(ctx, conn, input)
}

func findEventSubscription(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeEventSubscriptionsInput) (*dms.EventSubscription, error) {
	output, err := findEventSubscriptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findEventSubscriptions(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeEventSubscriptionsInput) ([]*dms.EventSubscription, error) {
	var output []*dms.EventSubscription

	err := conn.DescribeEventSubscriptionsPagesWithContext(ctx, input, func(page *dms.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EventSubscriptionsList {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusEventSubscription(ctx context.Context, conn *dms.DatabaseMigrationService, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEventSubscriptionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitEventSubscriptionCreated(ctx context.Context, conn *dms.DatabaseMigrationService, name string, timeout time.Duration) (*dms.EventSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{eventSubscriptionStatusCreating, eventSubscriptionStatusModifying},
		Target:     []string{eventSubscriptionStatusActive},
		Refresh:    statusEventSubscription(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitEventSubscriptionUpdated(ctx context.Context, conn *dms.DatabaseMigrationService, name string, timeout time.Duration) (*dms.EventSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{eventSubscriptionStatusModifying},
		Target:     []string{eventSubscriptionStatusActive},
		Refresh:    statusEventSubscription(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitEventSubscriptionDeleted(ctx context.Context, conn *dms.DatabaseMigrationService, name string, timeout time.Duration) (*dms.EventSubscription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{eventSubscriptionStatusDeleting},
		Target:     []string{},
		Refresh:    statusEventSubscription(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.EventSubscription); ok {
		return output, err
	}

	return nil, err
}
