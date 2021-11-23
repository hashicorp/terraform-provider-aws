package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	conn := meta.(*conns.AWSClient).EC2Conn

	params := &ec2.CreateSpotDatafeedSubscriptionInput{
		Bucket: aws.String(d.Get("bucket").(string)),
	}

	if v, ok := d.GetOk("prefix"); ok {
		params.Prefix = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating Spot Datafeed Subscription")
	_, err := conn.CreateSpotDatafeedSubscription(params)
	if err != nil {
		return fmt.Errorf("Error Creating Spot Datafeed Subscription: %s", err)
	}

	d.SetId("spot-datafeed-subscription")

	return resourceSpotDataFeedSubscriptionRead(d, meta)
}
func resourceSpotDataFeedSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	resp, err := conn.DescribeSpotDatafeedSubscription(&ec2.DescribeSpotDatafeedSubscriptionInput{})
	if err != nil {
		cgw, ok := err.(awserr.Error)
		if ok && cgw.Code() == "InvalidSpotDatafeed.NotFound" {
			log.Printf("[WARNING] Spot Datafeed Subscription Not Found so refreshing from state")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Describing Spot Datafeed Subscription: %s", err)
	}

	if resp == nil {
		log.Printf("[WARNING] Spot Datafeed Subscription Not Found so refreshing from state")
		d.SetId("")
		return nil
	}

	subscription := *resp.SpotDatafeedSubscription
	d.Set("bucket", subscription.Bucket)
	d.Set("prefix", subscription.Prefix)

	return nil
}
func resourceSpotDataFeedSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting Spot Datafeed Subscription")
	_, err := conn.DeleteSpotDatafeedSubscription(&ec2.DeleteSpotDatafeedSubscriptionInput{})
	if err != nil {
		return fmt.Errorf("Error deleting Spot Datafeed Subscription: %s", err)
	}
	return nil
}
