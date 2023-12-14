// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_event_subscription", name="Event Subscription")
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
			"customer_aws_id": {
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
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"source_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(rds.SourceType_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEventSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &rds.CreateEventSubscriptionInput{
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		SnsTopicArn:      aws.String(d.Get("sns_topic").(string)),
		SubscriptionName: aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("event_categories"); ok && v.(*schema.Set).Len() > 0 {
		input.EventCategories = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("source_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SourceIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("source_type"); ok {
		input.SourceType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating RDS Event Subscription: %s", input)
	output, err := conn.CreateEventSubscriptionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Event Subscription (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.EventSubscription.CustSubscriptionId))

	if _, err = waitEventSubscriptionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Event Subscription (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	sub, err := FindEventSubscriptionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Event Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Event Subscription (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(sub.EventSubscriptionArn)
	d.Set("arn", arn)
	d.Set("customer_aws_id", sub.CustomerAwsId)
	d.Set("enabled", sub.Enabled)
	d.Set("event_categories", aws.StringValueSlice(sub.EventCategoriesList))
	d.Set("name", sub.CustSubscriptionId)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(sub.CustSubscriptionId)))
	d.Set("sns_topic", sub.SnsTopicArn)
	d.Set("source_ids", aws.StringValueSlice(sub.SourceIdsList))
	d.Set("source_type", sub.SourceType)

	return diags
}

func resourceEventSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all", "source_ids") {
		input := &rds.ModifyEventSubscriptionInput{
			SubscriptionName: aws.String(d.Id()),
		}

		input.Enabled = aws.Bool(d.Get("enabled").(bool))

		if d.HasChange("event_categories") {
			input.EventCategories = flex.ExpandStringSet(d.Get("event_categories").(*schema.Set))
			input.SourceType = aws.String(d.Get("source_type").(string))
		}

		if d.HasChange("source_type") {
			input.SourceType = aws.String(d.Get("source_type").(string))
		}

		if d.HasChange("sns_topic") {
			input.SnsTopicArn = aws.String(d.Get("sns_topic").(string))
		}

		log.Printf("[DEBUG] Updating RDS Event Subscription: %s", input)
		_, err := conn.ModifyEventSubscriptionWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Event Subscription (%s): %s", d.Id(), err)
		}

		if _, err = waitEventSubscriptionUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Event Subscription (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("source_ids") {
		o, n := d.GetChange("source_ids")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		add := ns.Difference(os).List()
		del := os.Difference(ns).List()

		for _, del := range del {
			del := del.(string)
			_, err := conn.RemoveSourceIdentifierFromSubscriptionWithContext(ctx, &rds.RemoveSourceIdentifierFromSubscriptionInput{
				SourceIdentifier: aws.String(del),
				SubscriptionName: aws.String(d.Id()),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "removing RDS Event Subscription (%s) source ID (%s): %s", d.Id(), del, err)
			}
		}

		for _, add := range add {
			add := add.(string)
			_, err := conn.AddSourceIdentifierToSubscriptionWithContext(ctx, &rds.AddSourceIdentifierToSubscriptionInput{
				SourceIdentifier: aws.String(add),
				SubscriptionName: aws.String(d.Id()),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "adding RDS Event Subscription (%s) source ID (%s): %s", d.Id(), add, err)
			}
		}
	}

	return diags
}

func resourceEventSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	log.Printf("[DEBUG] Deleting RDS Event Subscription: (%s)", d.Id())
	_, err := conn.DeleteEventSubscriptionWithContext(ctx, &rds.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeSubscriptionNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Event Subscription (%s): %s", d.Id(), err)
	}

	if _, err = waitEventSubscriptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Event Subscription (%s) delete: %s", d.Id(), err)
	}

	return diags
}
