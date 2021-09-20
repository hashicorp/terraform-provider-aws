package rds

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
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
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validEventSubscriptionName,
			},
			"sns_topic": {
				Type:     schema.TypeString,
				Required: true,
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
				// ValidateFunc: validateDbEventSubscriptionSourceIds,
				// requires source_type to be set, does not seem to be a way to validate this
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

func resourceEventSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	sourceIdsSet := d.Get("source_ids").(*schema.Set)
	sourceIds := make([]*string, sourceIdsSet.Len())
	for i, sourceId := range sourceIdsSet.List() {
		sourceIds[i] = aws.String(sourceId.(string))
	}

	eventCategoriesSet := d.Get("event_categories").(*schema.Set)
	eventCategories := make([]*string, eventCategoriesSet.Len())
	for i, eventCategory := range eventCategoriesSet.List() {
		eventCategories[i] = aws.String(eventCategory.(string))
	}

	request := &rds.CreateEventSubscriptionInput{
		SubscriptionName: aws.String(name),
		SnsTopicArn:      aws.String(d.Get("sns_topic").(string)),
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		SourceIds:        sourceIds,
		SourceType:       aws.String(d.Get("source_type").(string)),
		EventCategories:  eventCategories,
		Tags:             tags.IgnoreAws().RdsTags(),
	}

	log.Println("[DEBUG] Create RDS Event Subscription:", request)

	output, err := conn.CreateEventSubscription(request)
	if err != nil || output.EventSubscription == nil {
		return fmt.Errorf("Error creating RDS Event Subscription %s: %s", name, err)
	}

	d.SetId(aws.StringValue(output.EventSubscription.CustSubscriptionId))

	log.Println(
		"[INFO] Waiting for RDS Event Subscription to be ready")

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active"},
		Refresh:    resourceAwsDbEventSubscriptionRefreshFunc(d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Creating RDS Event Subscription %s failed: %s", d.Id(), err)
	}

	return resourceEventSubscriptionRead(d, meta)
}

func resourceEventSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	sub, err := EventSubscriptionRetrieve(d.Id(), conn)

	if tfawserr.ErrMessageContains(err, rds.ErrCodeSubscriptionNotFoundFault, "") {
		log.Printf("[WARN] RDS Event Subscription (%s) not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving RDS Event Subscription (%s): %s", d.Id(), err)
	}

	if sub == nil {
		log.Printf("[WARN] RDS Event Subscription (%s) not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", sub.EventSubscriptionArn)
	if err := d.Set("name", sub.CustSubscriptionId); err != nil {
		return err
	}
	if err := d.Set("sns_topic", sub.SnsTopicArn); err != nil {
		return err
	}
	if err := d.Set("source_type", sub.SourceType); err != nil {
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

	tags, err := tftags.RdsListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for RDS Event Subscription (%s): %s", d.Get("arn").(string), err)
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

func EventSubscriptionRetrieve(name string, conn *rds.RDS) (*rds.EventSubscription, error) {
	input := &rds.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(name),
	}

	var eventSubscription *rds.EventSubscription

	err := conn.DescribeEventSubscriptionsPages(input, func(page *rds.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, es := range page.EventSubscriptionsList {
			if es == nil {
				continue
			}

			if aws.StringValue(es.CustSubscriptionId) == name {
				eventSubscription = es
				return false
			}
		}

		return !lastPage
	})

	return eventSubscription, err
}

func resourceEventSubscriptionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	requestUpdate := false

	req := &rds.ModifyEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	}

	if d.HasChange("event_categories") {
		eventCategoriesSet := d.Get("event_categories").(*schema.Set)
		req.EventCategories = make([]*string, eventCategoriesSet.Len())
		for i, eventCategory := range eventCategoriesSet.List() {
			req.EventCategories[i] = aws.String(eventCategory.(string))
		}
		req.SourceType = aws.String(d.Get("source_type").(string))
		requestUpdate = true
	}

	if d.HasChange("enabled") {
		req.Enabled = aws.Bool(d.Get("enabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("sns_topic") {
		req.SnsTopicArn = aws.String(d.Get("sns_topic").(string))
		requestUpdate = true
	}

	if d.HasChange("source_type") {
		req.SourceType = aws.String(d.Get("source_type").(string))
		requestUpdate = true
	}

	log.Printf("[DEBUG] Send RDS Event Subscription modification request: %#v", requestUpdate)
	if requestUpdate {
		log.Printf("[DEBUG] RDS Event Subscription modification request: %#v", req)
		_, err := conn.ModifyEventSubscription(req)
		if err != nil {
			return fmt.Errorf("Modifying RDS Event Subscription %s failed: %s", d.Id(), err)
		}

		log.Println(
			"[INFO] Waiting for RDS Event Subscription modification to finish")

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"modifying"},
			Target:     []string{"active"},
			Refresh:    resourceAwsDbEventSubscriptionRefreshFunc(d.Id(), conn),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
			Delay:      30 * time.Second, // Wait 30 secs before starting
		}

		// Wait, catching any errors
		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("Modifying RDS Event Subscription %s failed: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.RdsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating RDS Event Subscription (%s) tags: %s", d.Get("arn").(string), err)
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
				_, err := conn.RemoveSourceIdentifierFromSubscription(&rds.RemoveSourceIdentifierFromSubscriptionInput{
					SourceIdentifier: removing,
					SubscriptionName: aws.String(d.Id()),
				})
				if err != nil {
					return err
				}
			}
		}

		if len(add) > 0 {
			for _, adding := range add {
				log.Printf("[INFO] Adding %s as a Source Identifier to %q", *adding, d.Id())
				_, err := conn.AddSourceIdentifierToSubscription(&rds.AddSourceIdentifierToSubscriptionInput{
					SourceIdentifier: adding,
					SubscriptionName: aws.String(d.Id()),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func resourceEventSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	deleteOpts := rds.DeleteEventSubscriptionInput{
		SubscriptionName: aws.String(d.Id()),
	}

	_, err := conn.DeleteEventSubscription(&deleteOpts)

	if tfawserr.ErrMessageContains(err, rds.ErrCodeSubscriptionNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting RDS Event Subscription (%s): %s", d.Id(), err)
	}

	err = WaitForEventSubscriptionDeletion(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for RDS Event Subscription (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsDbEventSubscriptionRefreshFunc(name string, conn *rds.RDS) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		sub, err := EventSubscriptionRetrieve(name, conn)

		if tfawserr.ErrMessageContains(err, rds.ErrCodeSubscriptionNotFoundFault, "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if sub == nil {
			return nil, "", nil
		}

		return sub, aws.StringValue(sub.Status), nil
	}
}

func WaitForEventSubscriptionDeletion(conn *rds.RDS, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleting"},
		Target:     []string{},
		Refresh:    resourceAwsDbEventSubscriptionRefreshFunc(name, conn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	_, err := stateConf.WaitForState()

	return err
}
