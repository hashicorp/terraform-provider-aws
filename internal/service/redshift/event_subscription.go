package redshift

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEventSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceEventSubscriptionCreate,
		Read:   resourceEventSubscriptionRead,
		Update: resourceEventSubscriptionUpdate,
		Delete: resourceEventSubscriptionDelete,
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

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
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
			"severity": {
				Type:     schema.TypeString,
				Optional: true,
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

func resourceEventSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	request := &redshift.CreateEventSubscriptionInput{
		SubscriptionName: aws.String(d.Get("name").(string)),
		SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		SourceIds:        flex.ExpandStringSet(d.Get("source_ids").(*schema.Set)),
		SourceType:       aws.String(d.Get("source_type").(string)),
		Severity:         aws.String(d.Get("severity").(string)),
		EventCategories:  flex.ExpandStringSet(d.Get("event_categories").(*schema.Set)),
		Tags:             Tags(tags.IgnoreAws()),
	}

	log.Println("[DEBUG] Create Redshift Event Subscription:", request)

	output, err := conn.CreateEventSubscription(request)
	if err != nil || output.EventSubscription == nil {
		return fmt.Errorf("Error creating Redshift Event Subscription %s: %s", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(output.EventSubscription.CustSubscriptionId))

	return resourceEventSubscriptionRead(d, meta)
}

func resourceEventSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("eventsubscription:%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	sub, err := resourceAwsRedshiftEventSubscriptionRetrieve(d.Id(), conn)
	if err != nil {
		return fmt.Errorf("Error retrieving Redshift Event Subscription %s: %s", d.Id(), err)
	}
	if sub == nil {
		log.Printf("[WARN] Redshift Event Subscription (%s) not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("name", sub.CustSubscriptionId); err != nil {
		return err
	}
	if err := d.Set("sns_topic_arn", sub.SnsTopicArn); err != nil {
		return err
	}
	if err := d.Set("status", sub.Status); err != nil {
		return err
	}
	if err := d.Set("source_type", sub.SourceType); err != nil {
		return err
	}
	if err := d.Set("severity", sub.Severity); err != nil {
		return err
	}
	if err := d.Set("enabled", sub.Enabled); err != nil {
		return err
	}
	if err := d.Set("source_ids", flex.FlattenStringList(sub.SourceIdsList)); err != nil {
		return err
	}
	if err := d.Set("event_categories", flex.FlattenStringList(sub.EventCategoriesList)); err != nil {
		return err
	}
	if err := d.Set("customer_aws_id", sub.CustomerAwsId); err != nil {
		return err
	}
	tags := KeyValueTags(sub.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsRedshiftEventSubscriptionRetrieve(name string, conn *redshift.Redshift) (*redshift.EventSubscription, error) {

	request := &redshift.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(name),
	}

	describeResp, err := conn.DescribeEventSubscriptions(request)
	if err != nil {
		if tfawserr.ErrMessageContains(err, redshift.ErrCodeSubscriptionNotFoundFault, "") {
			log.Printf("[WARN] No Redshift Event Subscription by name (%s) found", name)
			return nil, nil
		}
		return nil, fmt.Errorf("Error reading Redshift Event Subscription %s: %s", name, err)
	}

	if len(describeResp.EventSubscriptionsList) != 1 {
		return nil, fmt.Errorf("Unable to find Redshift Event Subscription: %#v", describeResp.EventSubscriptionsList)
	}

	return describeResp.EventSubscriptionsList[0], nil
}

func resourceEventSubscriptionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	req := &redshift.ModifyEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
		SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		SourceIds:        flex.ExpandStringSet(d.Get("source_ids").(*schema.Set)),
		SourceType:       aws.String(d.Get("source_type").(string)),
		Severity:         aws.String(d.Get("severity").(string)),
		EventCategories:  flex.ExpandStringSet(d.Get("event_categories").(*schema.Set)),
	}

	log.Printf("[DEBUG] Redshift Event Subscription modification request: %#v", req)
	_, err := conn.ModifyEventSubscription(req)
	if err != nil {
		return fmt.Errorf("Modifying Redshift Event Subscription %s failed: %s", d.Id(), err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Event Subscription (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return nil
}

func resourceEventSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	deleteOpts := redshift.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteEventSubscription(&deleteOpts); err != nil {
		if tfawserr.ErrMessageContains(err, redshift.ErrCodeSubscriptionNotFoundFault, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Redshift Event Subscription %s: %s", d.Id(), err)
	}

	return nil
}
