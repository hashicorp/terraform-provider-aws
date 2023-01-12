package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSpotDataFeedSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceSpotDataFeedSubscriptionCreate,
		Read:   resourceSpotDataFeedSubscriptionRead,
		Delete: resourceSpotDataFeedSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSpotDataFeedSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.CreateSpotDatafeedSubscriptionInput{
		Bucket: aws.String(d.Get("bucket").(string)),
	}

	if v, ok := d.GetOk("prefix"); ok {
		input.Prefix = aws.String(v.(string))
	}

	_, err := conn.CreateSpotDatafeedSubscription(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Spot Datafeed Subscription: %w", err)
	}

	d.SetId("spot-datafeed-subscription")

	return resourceSpotDataFeedSubscriptionRead(d, meta)
}

func resourceSpotDataFeedSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn()

	subscription, err := FindSpotDatafeedSubscription(conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Spot Datafeed Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Spot Datafeed Subscription (%s): %w", d.Id(), err)
	}

	d.Set("bucket", subscription.Bucket)
	d.Set("prefix", subscription.Prefix)

	return nil
}

func resourceSpotDataFeedSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 Spot Datafeed Subscription: %s", d.Id())
	_, err := conn.DeleteSpotDatafeedSubscription(&ec2.DeleteSpotDatafeedSubscriptionInput{})

	if err != nil {
		return fmt.Errorf("deleting EC2 Spot Datafeed Subscription (%s): %w", d.Id(), err)
	}

	return nil
}
