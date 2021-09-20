package ses

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceActiveReceiptRuleSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceActiveReceiptRuleSetUpdate,
		Update: resourceActiveReceiptRuleSetUpdate,
		Read:   resourceActiveReceiptRuleSetRead,
		Delete: resourceActiveReceiptRuleSetDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_set_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceActiveReceiptRuleSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	ruleSetName := d.Get("rule_set_name").(string)

	createOpts := &ses.SetActiveReceiptRuleSetInput{
		RuleSetName: aws.String(ruleSetName),
	}

	_, err := conn.SetActiveReceiptRuleSet(createOpts)
	if err != nil {
		return fmt.Errorf("Error setting active SES rule set: %s", err)
	}

	d.SetId(ruleSetName)

	return resourceActiveReceiptRuleSetRead(d, meta)
}

func resourceActiveReceiptRuleSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	describeOpts := &ses.DescribeActiveReceiptRuleSetInput{}

	response, err := conn.DescribeActiveReceiptRuleSet(describeOpts)
	if err != nil {
		if tfawserr.ErrMessageContains(err, ses.ErrCodeRuleSetDoesNotExistException, "") {
			log.Printf("[WARN] SES Receipt Rule Set (%s) belonging to SES Active Receipt Rule Set not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if response.Metadata == nil {
		log.Print("[WARN] No active Receipt Rule Set found")
		d.SetId("")
		return nil
	}

	d.Set("rule_set_name", response.Metadata.Name)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("receipt-rule-set/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceActiveReceiptRuleSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	deleteOpts := &ses.SetActiveReceiptRuleSetInput{
		RuleSetName: nil,
	}

	_, err := conn.SetActiveReceiptRuleSet(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting active SES rule set: %s", err)
	}

	return nil
}
