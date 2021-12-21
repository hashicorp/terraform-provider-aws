package ses

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceActiveReceiptRuleSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceActiveReceiptRuleSetRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_set_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceActiveReceiptRuleSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn
	data, err := conn.DescribeActiveReceiptRuleSet(&ses.DescribeActiveReceiptRuleSetInput{})
	if err != nil {
		if tfawserr.ErrMessageContains(err, ses.ErrCodeRuleSetDoesNotExistException, "") {
			return fmt.Errorf("SES Active Receipt Rule Set not found: %s", err)
		}
		return fmt.Errorf("Error getting SES Active Receipt Rule Set: %s", err)
	}
	d.SetId(*data.Metadata.Name)
	d.Set("rule_set_name", data.Metadata.Name)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("receipt-rule-set/%s", *data.Metadata.Name),
	}.String()
	d.Set("arn", arn)
	return nil
}
