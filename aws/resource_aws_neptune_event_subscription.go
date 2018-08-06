package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsNeptuneEventSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNeptuneEventSubscriptionCreate,
		Read:   resourceAwsNeptuneEventSubscriptionRead,
		Update: resourceAwsNeptuneEventSubscriptionUpdate,
		Delete: resourceAwsNeptuneEventSubscriptionDelete,

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
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateNeptuneEventSubscriptionName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateNeptuneEventSubscriptionNamePrefix,
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
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
			"customer_aws_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsNeptuneEventSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

	if v, ok := d.GetOk("name"); ok {
		d.Set("name", v.(string))
	} else if v, ok := d.GetOk("name_prefix"); ok {
		d.Set("name", resource.PrefixedUniqueId(v.(string)))
	} else {
		d.Set("name", resource.PrefixedUniqueId("tf-"))
	}

	tags := tagsFromMapNeptune(d.Get("tags").(map[string]interface{}))

	request := &neptune.CreateEventSubscriptionInput{
		SubscriptionName: aws.String(d.Get("name").(string)),
		SnsTopicArn:      aws.String(d.Get("sns_topic_arn").(string)),
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		Tags:             tags,
	}

	if v, ok := d.GetOk("source_ids"); ok {
		sourceIdsSet := v.(*schema.Set)
		sourceIds := make([]*string, sourceIdsSet.Len())
		for i, sourceId := range sourceIdsSet.List() {
			sourceIds[i] = aws.String(sourceId.(string))
		}
		request.SourceIds = sourceIds
	}

	if v, ok := d.GetOk("event_categories"); ok {
		eventCategoriesSet := v.(*schema.Set)
		eventCategories := make([]*string, eventCategoriesSet.Len())
		for i, eventCategory := range eventCategoriesSet.List() {
			eventCategories[i] = aws.String(eventCategory.(string))
		}
		request.EventCategories = eventCategories
	}

	if v, ok := d.GetOk("source_type"); ok {
		request.SourceType = aws.String(v.(string))
	}

	log.Println("[DEBUG] Create Neptune Event Subscription:", request)

	output, err := conn.CreateEventSubscription(request)
	if err != nil || output.EventSubscription == nil {
		return fmt.Errorf("Error creating Neptune Event Subscription %s: %s", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(output.EventSubscription.CustSubscriptionId))

	if err := setTagsNeptune(conn, d, aws.StringValue(output.EventSubscription.EventSubscriptionArn)); err != nil {
		return fmt.Errorf("Error creating Neptune Event Subscription (%s) tags: %s", d.Id(), err)
	}

	log.Println("[INFO] Waiting for Neptune Event Subscription to be ready")

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active"},
		Refresh:    resourceAwsNeptuneEventSubscriptionRefreshFunc(d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Creating Neptune Event Subscription %s failed: %s", d.Id(), err)
	}

	return resourceAwsNeptuneEventSubscriptionRead(d, meta)
}
