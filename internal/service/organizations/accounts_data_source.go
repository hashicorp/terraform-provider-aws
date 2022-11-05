package organizations

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceAccounts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAccountsRead,
		Schema: map[string]*schema.Schema{
			"accounts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	DSNameAccounts = "Organizations Accounts Data Source"
)

func dataSourceAccountsRead(d *schema.ResourceData, meta interface{}) error {
	var accountIds []string
	var err error
	var listAccountInput organizations.ListAccountsForParentInput
	var listAccountOutput *organizations.ListAccountsForParentOutput

	conn := meta.(*conns.AWSClient).OrganizationsConn
	parentId := d.Get("parent_id").(string)
	for {
		listAccountInput = organizations.ListAccountsForParentInput{ParentId: &parentId}
		listAccountOutput, err = conn.ListAccountsForParent(&listAccountInput)
		if err != nil {
			return err
		}
		for i := 0; i < len(listAccountOutput.Accounts); i++ {
			accountIds = append(accountIds, *(listAccountOutput.Accounts[i]).Id)
		}
		if listAccountOutput.NextToken == nil {
			break
		} else {
			parentId = aws.StringValue(listAccountOutput.NextToken)
		}
	}
	d.SetId(d.Get("parent_id").(string))
	d.Set("accounts", accountIds)

	return nil
}
