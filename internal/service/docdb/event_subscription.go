// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_docdb_event_subscription", name="Event Subscription")
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
				ValidateFunc:  validEventSubscriptionNamePrefix,
			},
			"sns_topic_arn": {
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
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEventSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &docdb.CreateEventSubscriptionInput{
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
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

	output, err := conn.CreateEventSubscriptionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating DocumentDB Event Subscription (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.EventSubscription.CustSubscriptionId))

	if _, err := waitEventSubscriptionActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for DocumentDB Event Subscription (%s) to become active: %s", d.Id(), err)
	}

	return resourceEventSubscriptionRead(ctx, d, meta)
}

func resourceEventSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	output, err := FindEventSubscriptionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DocumentDB Event Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading DocumentDB Event Subscription (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.EventSubscriptionArn)
	d.Set("customer_aws_id", output.CustomerAwsId)
	d.Set("enabled", output.Enabled)
	d.Set("event_categories", aws.StringValueSlice(output.EventCategoriesList))
	d.Set("name", output.CustSubscriptionId)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(output.CustSubscriptionId)))
	d.Set("sns_topic_arn", output.SnsTopicArn)
	d.Set("source_ids", aws.StringValueSlice(output.SourceIdsList))
	d.Set("source_type", output.SourceType)

	return nil
}

func resourceEventSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	if d.HasChangesExcept("tags", "tags_all", "source_ids") {
		input := &docdb.ModifyEventSubscriptionInput{
			SubscriptionName: aws.String(d.Id()),
		}

		if d.HasChange("enabled") {
			input.Enabled = aws.Bool(d.Get("enabled").(bool))
		}

		if d.HasChange("event_categories") {
			input.EventCategories = flex.ExpandStringSet(d.Get("event_categories").(*schema.Set))
			input.SourceType = aws.String(d.Get("source_type").(string))
		}

		if d.HasChange("sns_topic_arn") {
			input.SnsTopicArn = aws.String(d.Get("sns_topic_arn").(string))
		}

		if d.HasChange("source_type") {
			input.SourceType = aws.String(d.Get("source_type").(string))
		}

		_, err := conn.ModifyEventSubscriptionWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating DocumentDB Event Subscription (%s): %s", d.Id(), err)
		}

		if _, err := waitEventSubscriptionActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for DocumentDB Event Subscription (%s) to become active: %s", d.Id(), err)
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
				_, err := conn.RemoveSourceIdentifierFromSubscriptionWithContext(ctx, &docdb.RemoveSourceIdentifierFromSubscriptionInput{
					SourceIdentifier: v,
					SubscriptionName: aws.String(d.Id()),
				})

				if err != nil {
					return diag.Errorf("removing DocumentDB Cluster Event Subscription (%s) source identifier: %s", d.Id(), err)
				}
			}
		}

		if len(add) > 0 {
			for _, v := range add {
				_, err := conn.AddSourceIdentifierToSubscriptionWithContext(ctx, &docdb.AddSourceIdentifierToSubscriptionInput{
					SourceIdentifier: v,
					SubscriptionName: aws.String(d.Id()),
				})

				if err != nil {
					return diag.Errorf("adding DocumentDB Cluster Event Subscription (%s) source identifier: %s", d.Id(), err)
				}
			}
		}
	}

	return resourceEventSubscriptionRead(ctx, d, meta)
}

func resourceEventSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	log.Printf("[DEBUG] Deleting DocumentDB Event Subscription: %s", d.Id())
	_, err := conn.DeleteEventSubscriptionWithContext(ctx, &docdb.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeSubscriptionNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting DocumentDB Event Subscription (%s): %s", d.Id(), err)
	}

	if _, err := waitEventSubscriptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for DocumentDB Event Subscription (%s) delete: %s", d.Id(), err)
	}

	return nil
}
