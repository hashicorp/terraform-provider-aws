package docdb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"customer_aws_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEventSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if v, ok := d.GetOk("name"); ok {
		d.Set("name", v.(string))
	} else if v, ok := d.GetOk("name_prefix"); ok {
		d.Set("name", resource.PrefixedUniqueId(v.(string)))
	} else {
		d.Set("name", resource.PrefixedUniqueId("tf-"))
	}

	request := &docdb.CreateEventSubscriptionInput{
		SubscriptionName: aws.String(d.Get("name").(string)),
		SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		Tags:             Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("source_ids"); ok {
		request.SourceIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("event_categories"); ok {
		request.EventCategories = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("source_type"); ok {
		request.SourceType = aws.String(v.(string))
	}

	log.Println("[DEBUG] Create DocDB Event Subscription:", request)

	output, err := conn.CreateEventSubscription(request)
	if err != nil || output.EventSubscription == nil {
		return diag.FromErr(fmt.Errorf("Error creating DocDB Event Subscription %s: %w", d.Get("name").(string), err))
	}

	d.SetId(aws.StringValue(output.EventSubscription.CustSubscriptionId))

	if _, err := waitEventSubscriptionActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(fmt.Errorf("Error waiting for DocDB Event Subscription (%s) state to be Active: %w", d.Id(), err))
	}

	return resourceEventSubscriptionRead(ctx, d, meta)
}

func resourceEventSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	sub, err := FindEventSubscriptionById(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN]DocDB Event Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading DocDB Event Subscription %s: %w", d.Id(), err))
	}

	d.Set("arn", sub.EventSubscriptionArn)
	d.Set("name", sub.CustSubscriptionId)
	d.Set("sns_topic_arn", sub.SnsTopicArn)
	d.Set("enabled", sub.Enabled)
	d.Set("customer_aws_id", sub.CustomerAwsId)

	if sub.SourceType != nil {
		d.Set("source_type", sub.SourceType)
	}

	if sub.SourceIdsList != nil {
		if err := d.Set("source_ids", flex.FlattenStringList(sub.SourceIdsList)); err != nil {
			return diag.FromErr(fmt.Errorf("Error saving Source IDs to state for DocDB Event Subscription (%s): %s", d.Id(), err))
		}
	}

	if sub.EventCategoriesList != nil {
		if err := d.Set("event_categories", flex.FlattenStringList(sub.EventCategoriesList)); err != nil {
			return diag.FromErr(fmt.Errorf("Error saving Event Categories to state for DocDB Event Subscription (%s): %s", d.Id(), err))
		}
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for DocDB Event Subscription (%s): %s", d.Get("arn").(string), err))
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceEventSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn

	if d.HasChangesExcept("tags", "tags_all", "source_ids") {
		req := &docdb.ModifyEventSubscriptionInput{
			SubscriptionName: aws.String(d.Id()),
		}

		if d.HasChange("event_categories") {
			req.EventCategories = flex.ExpandStringSet(d.Get("event_categories").(*schema.Set))
			req.SourceType = aws.String(d.Get("source_type").(string))
		}

		if d.HasChange("enabled") {
			req.Enabled = aws.Bool(d.Get("enabled").(bool))
		}

		if d.HasChange("sns_topic_arn") {
			req.SnsTopicArn = aws.String(d.Get("sns_topic_arn").(string))
		}

		if d.HasChange("source_type") {
			req.SourceType = aws.String(d.Get("source_type").(string))
		}

		log.Printf("[DEBUG] DocDB Event Subscription modification request: %#v", req)
		_, err := conn.ModifyEventSubscription(req)
		if err != nil {
			return diag.FromErr(fmt.Errorf("Modifying DocDB Event Subscription %s failed: %w", d.Id(), err))
		}

		if _, err := waitEventSubscriptionActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(fmt.Errorf("Error waiting for DocDB Event Subscription (%s) state to be Modified: %w", d.Id(), err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating DocDB Cluster Event Subscription (%s) tags: %s", d.Get("arn").(string), err))
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
				_, err := conn.RemoveSourceIdentifierFromSubscription(&docdb.RemoveSourceIdentifierFromSubscriptionInput{
					SourceIdentifier: removing,
					SubscriptionName: aws.String(d.Id()),
				})
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		if len(add) > 0 {
			for _, adding := range add {
				log.Printf("[INFO] Adding %s as a Source Identifier to %q", *adding, d.Id())
				_, err := conn.AddSourceIdentifierToSubscription(&docdb.AddSourceIdentifierToSubscriptionInput{
					SourceIdentifier: adding,
					SubscriptionName: aws.String(d.Id()),
				})
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return resourceEventSubscriptionRead(ctx, d, meta)
}

func resourceEventSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn

	log.Printf("[DEBUG] Deleting DocDB Event Subscription: %s", d.Id())
	_, err := conn.DeleteEventSubscription(&docdb.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeSubscriptionNotFoundFault) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting DocDB Event Subscription (%s): %w", d.Id(), err))
	}

	if _, err := waitEventSubscriptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(fmt.Errorf("Error waiting for DocDB Event Subscription (%s) state to be Deleted: %w", d.Id(), err))
	}

	return nil
}
