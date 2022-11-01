package ses

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
		return fmt.Errorf("reading SES Active Receipt Rule Set: %s", err)
	}
	if data == nil || data.Metadata == nil {
		return tfresource.NewEmptyResultError(nil)
	}

	name := aws.StringValue(data.Metadata.Name)
	d.SetId(name)
	d.Set("rule_set_name", name)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("receipt-rule-set/%s", name),
	}.String()
	d.Set("arn", arn)

	return nil
}
