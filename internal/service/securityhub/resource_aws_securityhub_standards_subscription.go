package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/securityhub/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/securityhub/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStandardsSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceStandardsSubscriptionCreate,
		Read:   resourceStandardsSubscriptionRead,
		Delete: resourceStandardsSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"standards_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceStandardsSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	standardsARN := d.Get("standards_arn").(string)
	input := &securityhub.BatchEnableStandardsInput{
		StandardsSubscriptionRequests: []*securityhub.StandardsSubscriptionRequest{
			{
				StandardsArn: aws.String(standardsARN),
			},
		},
	}

	log.Printf("[DEBUG] Creating Security Hub Standards Subscription: %s", input)
	output, err := conn.BatchEnableStandards(input)

	if err != nil {
		return fmt.Errorf("error enabling Security Hub Standard (%s): %w", standardsARN, err)
	}

	d.SetId(aws.StringValue(output.StandardsSubscriptions[0].StandardsSubscriptionArn))

	_, err = waiter.waitStandardsSubscriptionCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Security Hub Standards Subscription (%s) to create: %w", d.Id(), err)
	}

	return resourceStandardsSubscriptionRead(d, meta)
}

func resourceStandardsSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	output, err := finder.FindStandardsSubscriptionByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Standards Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("standards_arn", output.StandardsArn)

	return nil
}

func resourceStandardsSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	log.Printf("[DEBUG] Deleting Security Hub Standards Subscription: %s", d.Id())
	_, err := conn.BatchDisableStandards(&securityhub.BatchDisableStandardsInput{
		StandardsSubscriptionArns: aws.StringSlice([]string{d.Id()}),
	})

	if err != nil {
		return fmt.Errorf("error disabling Security Hub Standard (%s): %w", d.Id(), err)
	}

	_, err = waiter.waitStandardsSubscriptionDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Security Hub Standards Subscription (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
