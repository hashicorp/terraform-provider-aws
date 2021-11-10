package meta

import (
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// See http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-getting-started.html#step-2
var billingAccountId = "386209384616"

func DataSourceBillingServiceAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceBillingServiceAccountRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBillingServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(billingAccountId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "iam",
		AccountID: billingAccountId,
		Resource:  "root",
	}.String()
	d.Set("arn", arn)

	return nil
}
