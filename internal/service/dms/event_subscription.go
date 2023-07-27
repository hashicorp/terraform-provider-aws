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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"event_categories": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
			},
			"name": {
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
				Set:      schema.HashString,
				ForceNew: true,
				Optional: true,
			},
			"source_type": {
				Type:     schema.TypeString,
				Optional: true,
				// The API suppors modification but doing so loses all source_ids
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

	request := &dms.CreateEventSubscriptionInput{
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
		SubscriptionName: aws.String(d.Get("name").(string)),
		SourceType:       aws.String(d.Get("source_type").(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("event_categories"); ok {
		request.EventCategories = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("source_ids"); ok {
		request.SourceIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	_, err := conn.CreateEventSubscriptionWithContext(ctx, request)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Event Subscription (%s): %s", d.Get("name").(string), err)
	}

	d.SetId(d.Get("name").(string))

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "modifying"},
		Target:     []string{"active"},
		Refresh:    resourceEventSubscriptionStateRefreshFunc(ctx, conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Event Subscription (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	request := &dms.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(d.Id()),
	}

	response, err := conn.DescribeEventSubscriptionsWithContext(ctx, request)

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		log.Printf("[WARN] DMS event subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS event subscription: %s", err)
	}

	if response == nil || len(response.EventSubscriptionsList) == 0 || response.EventSubscriptionsList[0] == nil {
		log.Printf("[WARN] DMS event subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	subscription := response.EventSubscriptionsList[0]

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "dms",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("es:%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("enabled", subscription.Enabled)
	d.Set("sns_topic_arn", subscription.SnsTopicArn)
	d.Set("source_type", subscription.SourceType)
	d.Set("name", d.Id())
	d.Set("event_categories", flex.FlattenStringList(subscription.EventCategoriesList))
	d.Set("source_ids", flex.FlattenStringList(subscription.SourceIdsList))

	return diags
}

func resourceEventSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if d.HasChanges("enabled", "event_categories", "sns_topic_arn", "source_type") {
		request := &dms.ModifyEventSubscriptionInput{
			Enabled:          aws.Bool(d.Get("enabled").(bool)),
			SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
			SubscriptionName: aws.String(d.Get("name").(string)),
			SourceType:       aws.String(d.Get("source_type").(string)),
		}

		if v, ok := d.GetOk("event_categories"); ok {
			request.EventCategories = flex.ExpandStringSet(v.(*schema.Set))
		}

		_, err := conn.ModifyEventSubscriptionWithContext(ctx, request)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DMS Event Subscription (%s): %s", d.Id(), err)
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"modifying"},
			Target:     []string{"active"},
			Refresh:    resourceEventSubscriptionStateRefreshFunc(ctx, conn, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
			Delay:      10 * time.Second,
		}

		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DMS Event Subscription (%s) modification: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	request := &dms.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	}

	_, err := conn.DeleteEventSubscriptionWithContext(ctx, request)

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Event Subscription (%s): %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{"deleting"},
		Target:     []string{},
		Refresh:    resourceEventSubscriptionStateRefreshFunc(ctx, conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Event Subscription (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func resourceEventSubscriptionStateRefreshFunc(ctx context.Context, conn *dms.DatabaseMigrationService, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := conn.DescribeEventSubscriptionsWithContext(ctx, &dms.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(name),
		})

		if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if v == nil || len(v.EventSubscriptionsList) == 0 || v.EventSubscriptionsList[0] == nil {
			return nil, "", nil
		}

		return v, aws.StringValue(v.EventSubscriptionsList[0].Status), nil
	}
}
