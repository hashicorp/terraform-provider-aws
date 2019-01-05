package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationsAccounts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsAccountsRead,

		Schema: map[string]*schema.Schema{
			"account_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsOrganizationsAccountsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	accounts := make([]string, 0)
	input := &organizations.ListAccountsInput{}
	for {
		output, err := conn.ListAccounts(input)
		if err != nil {
			return fmt.Errorf("error reading AWS Organization: %s", err)
		}
		for _, account := range output.Accounts {
			accounts = append(accounts, aws.StringValue(account.Id))
		}
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	d.Set("account_ids", accounts)

	return nil
}
