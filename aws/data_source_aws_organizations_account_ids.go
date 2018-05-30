package aws

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationsAccountIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsAccountIDsRead,
		Schema: map[string]*schema.Schema{

			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsOrganizationsAccountIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	req := &organizations.ListAccountsInput{}

	accounts := make([]string, 0)

	log.Printf("[DEBUG] ListAccounts %s\n", req)
	if err := conn.ListAccountsPages(req, func(resp *organizations.ListAccountsOutput, lastPage bool) bool {
		for _, account := range resp.Accounts {
			accounts = append(accounts, *account.Id)
		}
		return true
	}); err != nil {
		return err
	}

	log.Printf("[DEBUG] ListAccountsResponse %s\n", accounts)

	d.SetId(time.Now().UTC().String())
	d.Set("ids", accounts)

	return nil
}
