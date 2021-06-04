package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func resourceAwsDirectoryServiceLogSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDirectoryServiceLogSubscriptionCreate,
		Read:   resourceAwsDirectoryServiceLogSubscriptionRead,
		Delete: resourceAwsDirectoryServiceLogSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"log_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDirectoryServiceLogSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	DirectoryConn := meta.(*awsprovider.AWSClient).DirectoryConn

	directoryId := d.Get("directory_id")
	logGroupName := d.Get("log_group_name")

	input := directoryservice.CreateLogSubscriptionInput{
		DirectoryId:  aws.String(directoryId.(string)),
		LogGroupName: aws.String(logGroupName.(string)),
	}

	_, err := DirectoryConn.CreateLogSubscription(&input)
	if err != nil {
		return fmt.Errorf("error creating Directory Service Log Subscription: %s", err)
	}

	d.SetId(directoryId.(string))

	return resourceAwsDirectoryServiceLogSubscriptionRead(d, meta)
}

func resourceAwsDirectoryServiceLogSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	DirectoryConn := meta.(*awsprovider.AWSClient).DirectoryConn

	directoryId := d.Id()

	input := directoryservice.ListLogSubscriptionsInput{
		DirectoryId: aws.String(directoryId),
	}

	out, err := DirectoryConn.ListLogSubscriptions(&input)
	if err != nil {
		return fmt.Errorf("error listing Directory Service Log Subscription: %s", err)
	}

	if len(out.LogSubscriptions) == 0 {
		log.Printf("[WARN] No log subscriptions for directory %s found", directoryId)
		d.SetId("")
		return nil
	}

	logSubscription := out.LogSubscriptions[0]
	d.Set("directory_id", logSubscription.DirectoryId)
	d.Set("log_group_name", logSubscription.LogGroupName)

	return nil
}

func resourceAwsDirectoryServiceLogSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	DirectoryConn := meta.(*awsprovider.AWSClient).DirectoryConn

	directoryId := d.Id()

	input := directoryservice.DeleteLogSubscriptionInput{
		DirectoryId: aws.String(directoryId),
	}

	_, err := DirectoryConn.DeleteLogSubscription(&input)
	if err != nil {
		return fmt.Errorf("error deleting Directory Service Log Subscription: %s", err)
	}

	return nil
}
