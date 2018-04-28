package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationsAccountIds() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationAccountIdsRead,

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsOrganizationAccountIdsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	accountIds := make([]string, 0)

	if parentId, ok := d.GetOk("parent_id"); ok {
		input := &organizations.ListAccountsForParentInput{
			ParentId: aws.String(parentId.(string)),
		}
		result, err := conn.ListAccountsForParent(input)

		if err != nil {
			return err
		}

		for _, account := range result.Accounts {
			accountIds = append(accountIds, *account.Id)
		}
	} else {
		input := &organizations.ListAccountsInput{}
		result, err := conn.ListAccounts(input)

		if err != nil {
			return err
		}

		for _, account := range result.Accounts {
			accountIds = append(accountIds, *account.Id)
		}
	}

	d.SetId(time.Now().UTC().String())
	d.Set("ids", accountIds)

	return nil
}
