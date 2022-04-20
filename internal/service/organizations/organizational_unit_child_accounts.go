package organizations

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceOrganizationalUnitChildAccounts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOrganizationalUnitChildAccountsRead,

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"joined_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"joined_timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceOrganizationalUnitChildAccountsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	parent_id := d.Get("parent_id").(string)

	params := &organizations.ListAccountsForParentInput{
		ParentId: aws.String(parent_id),
	}

	var accounts []*organizations.Account

	err := conn.ListAccountsForParentPages(params,
		func(page *organizations.ListAccountsForParentOutput, lastPage bool) bool {
			accounts = append(accounts, page.Accounts...)

			return !lastPage
		})

	if err != nil {
		return fmt.Errorf("error listing Organizations Accounts for parent (%s): %w", parent_id, err)
	}

	d.SetId(parent_id)

	if err := d.Set("accounts", flattenAccounts(accounts)); err != nil {
		return fmt.Errorf("error setting accounts: %w", err)
	}

	return nil
}

func flattenAccounts(accounts []*organizations.Account) []map[string]interface{} {
	if len(accounts) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, account := range accounts {
		result = append(result, map[string]interface{}{
			"arn":              aws.StringValue(account.Arn),
			"email":            aws.StringValue(account.Email),
			"id":               aws.StringValue(account.Id),
			"joined_method":    aws.StringValue(account.JoinedMethod),
			"joined_timestamp": aws.TimeValue(account.JoinedTimestamp),
			"name":             aws.StringValue(account.Name),
			"status":           aws.StringValue(account.Status),
		})
	}
	return result
}
