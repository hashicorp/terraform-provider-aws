// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_event_subscription", name="Event Subscription")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/redshift/types;awstypes;awstypes.EventSubscription")
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
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						names.AttrConfiguration,
						"management",
						"monitoring",
						"security",
						"pending",
					}, false),
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"severity": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "INFO",
				ValidateFunc: validation.StringInSlice([]string{
					"ERROR",
					"INFO",
				}, false),
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.SourceType](),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceEventSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := redshift.CreateEventSubscriptionInput{
		Enabled:          aws.Bool(d.Get(names.AttrEnabled).(bool)),
		SnsTopicArn:      aws.String(d.Get(names.AttrSNSTopicARN).(string)),
		SubscriptionName: aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("event_categories"); ok && v.(*schema.Set).Len() > 0 {
		input.EventCategories = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("severity"); ok {
		input.Severity = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SourceIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSourceType); ok {
		input.SourceType = aws.String(v.(string))
	}

	output, err := conn.CreateEventSubscription(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Event Subscription (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.EventSubscription.CustSubscriptionId))

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.RedshiftClient(ctx)

	sub, err := findEventSubscriptionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift Event Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Event Subscription (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, eventSubscriptionARN(ctx, c, d.Id()))
	d.Set("customer_aws_id", sub.CustomerAwsId)
	d.Set(names.AttrEnabled, sub.Enabled)
	d.Set("event_categories", sub.EventCategoriesList)
	d.Set(names.AttrName, sub.CustSubscriptionId)
	d.Set("severity", sub.Severity)
	d.Set(names.AttrSNSTopicARN, sub.SnsTopicArn)
	d.Set("source_ids", sub.SourceIdsList)
	d.Set(names.AttrSourceType, sub.SourceType)
	d.Set(names.AttrStatus, sub.Status)

	setTagsOut(ctx, sub.Tags)

	return diags
}

func resourceEventSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := redshift.ModifyEventSubscriptionInput{
			Enabled:          aws.Bool(d.Get(names.AttrEnabled).(bool)),
			EventCategories:  flex.ExpandStringValueSet(d.Get("event_categories").(*schema.Set)),
			Severity:         aws.String(d.Get("severity").(string)),
			SnsTopicArn:      aws.String(d.Get(names.AttrSNSTopicARN).(string)),
			SourceIds:        flex.ExpandStringValueSet(d.Get("source_ids").(*schema.Set)),
			SourceType:       aws.String(d.Get(names.AttrSourceType).(string)),
			SubscriptionName: aws.String(d.Id()),
		}
		_, err := conn.ModifyEventSubscription(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Event Subscription (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceEventSubscriptionRead(ctx, d, meta)...)
}

func resourceEventSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Event Subscription: %s", d.Id())
	input := redshift.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	}
	_, err := conn.DeleteEventSubscription(ctx, &input)

	if errs.IsA[*awstypes.SubscriptionNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Event Subscription (%s): %s", d.Id(), err)
	}

	return diags
}

func eventSubscriptionARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, names.Redshift, "eventsubscription:"+id)
}
