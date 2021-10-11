package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsDbEventSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbEventSubscriptionCreate,
		Read:   resourceAwsDbEventSubscriptionRead,
		Update: resourceAwsDbEventSubscriptionUpdate,
		Delete: resourceAwsDbEventSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				ValidateFunc:  validateDbEventSubscriptionName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateDbEventSubscriptionName,
			},
			"sns_topic": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
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
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsDbEventSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	name := naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &rds.CreateEventSubscriptionInput{
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		SnsTopicArn:      aws.String(d.Get("sns_topic").(string)),
		SubscriptionName: aws.String(name),
	}

	if v, ok := d.GetOk("event_categories"); ok && v.(*schema.Set).Len() > 0 {
		input.EventCategories = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("source_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SourceIds = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("source_type"); ok {
		input.SourceType = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().RdsTags()
	}

	log.Printf("[DEBUG] Creating RDS Event Subscription: %s", input)
	output, err := conn.CreateEventSubscription(input)

	if err != nil {
		return fmt.Errorf("error creating RDS Event Subscription (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.EventSubscription.CustSubscriptionId))

	if _, err = waiter.EventSubscriptionCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for RDS Event Subscription (%s) create: %w", d.Id(), err)
	}

	return resourceAwsDbEventSubscriptionRead(d, meta)
}

func resourceAwsDbEventSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	sub, err := finder.EventSubscriptionByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Event Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS Event Subscription (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(sub.EventSubscriptionArn)
	d.Set("arn", arn)
	d.Set("customer_aws_id", sub.CustomerAwsId)
	d.Set("enabled", sub.Enabled)
	d.Set("event_categories", aws.StringValueSlice(sub.EventCategoriesList))
	d.Set("name", sub.CustSubscriptionId)
	d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(sub.CustSubscriptionId)))
	d.Set("sns_topic", sub.SnsTopicArn)
	d.Set("source_ids", aws.StringValueSlice(sub.SourceIdsList))
	d.Set("source_type", sub.SourceType)

	tags, err := keyvaluetags.RdsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for RDS Event Subscription (%s): %w", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsDbEventSubscriptionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	if d.HasChangesExcept("tags", "tags_all", "source_ids") {
		input := &rds.ModifyEventSubscriptionInput{
			SubscriptionName: aws.String(d.Id()),
		}

		if d.HasChange("enabled") {
			input.Enabled = aws.Bool(d.Get("enabled").(bool))
		}

		if d.HasChange("event_categories") {
			input.EventCategories = expandStringSet(d.Get("event_categories").(*schema.Set))
		}

		if d.HasChange("source_type") {
			input.SourceType = aws.String(d.Get("source_type").(string))
		}

		if d.HasChange("sns_topic") {
			input.SnsTopicArn = aws.String(d.Get("sns_topic").(string))
		}

		log.Printf("[DEBUG] Updating RDS Event Subscription: %s", input)
		_, err := conn.ModifyEventSubscription(input)

		if err != nil {
			return fmt.Errorf("error updating RDS Event Subscription (%s): %w", d.Id(), err)
		}

		if _, err = waiter.EventSubscriptionUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for RDS Event Subscription (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.RdsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating RDS Event Subscription (%s) tags: %w", d.Get("arn").(string), err)
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
			_, err := conn.RemoveSourceIdentifierFromSubscription(&rds.RemoveSourceIdentifierFromSubscriptionInput{
				SourceIdentifier: aws.String(del),
				SubscriptionName: aws.String(d.Id()),
			})

			if err != nil {
				return fmt.Errorf("error removing RDS Event Subscription (%s) source ID (%s): %w", d.Id(), del, err)
			}
		}

		for _, add := range add {
			add := add.(string)
			_, err := conn.AddSourceIdentifierToSubscription(&rds.AddSourceIdentifierToSubscriptionInput{
				SourceIdentifier: aws.String(add),
				SubscriptionName: aws.String(d.Id()),
			})

			if err != nil {
				return fmt.Errorf("error adding RDS Event Subscription (%s) source ID (%s): %w", d.Id(), add, err)
			}
		}
	}

	return nil
}

func resourceAwsDbEventSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	log.Printf("[DEBUG] Deleting RDS Event Subscription: (%s)", d.Id())
	_, err := conn.DeleteEventSubscription(&rds.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeSubscriptionNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting RDS Event Subscription (%s): %w", d.Id(), err)
	}

	if _, err = waiter.EventSubscriptionDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for RDS Event Subscription (%s) delete: %w", d.Id(), err)
	}

	return nil
}
