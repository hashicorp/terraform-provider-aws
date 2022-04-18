package organizations

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceOrganizationalUnitDescendantAccounts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOrganizationalUnitDescendantAccountsRead,

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

// Todo: Get all child OUs
func getDescendantOrganizationalUnits(parent_id string) []*organizations.OrganizationalUnit {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	params := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parent_id),
	}

    // Descendants will hold all generations of OUs under parent_id.
	var descendants []*organizations.OrganizationalUnit
	// Children will hold immediate child OUs of parent_id.
	var children []*organizations.OrganizationalUnit
	// result will hold any descendants of immediate child OUs.
	var result []*organizations.OrganizationalUnit

    // Get all immediate children OUs.
	err := conn.ListOrganizationalUnitsForParentPages(params,
		func(page *organizations.ListOrganizationalUnitsForParentOutput, lastPage bool) bool {
			children = append(children, page.OrganizationalUnits...)

			return !lastPage
		})

	if err != nil {
		return fmt.Errorf("error listing Organizations Organization Units for parent (%s): %w", parent_id, err)
	}

    // If child OUs exist, get all their descendants.
	if children = flattenOrganizationalUnits(children); children != nil {
        for _, ou := range children {
            // Append this child ou and all it's descendants.
            descendants = append(descendants, ou)
            if result = getDescendantOrganizationalUnits(ou); result != nil{
                descendants = append(descendants, result)
            }
        }

        return descendants
	}

    //Base Case. ParentId has no children.
	return nil
}

func dataSourceOrganizationalUnitDescendantAccountsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

    // Collect all the OUs of which we need to list children.
	var organizationalUnits []*organizations.OrganizationalUnit
	organizationalUnits = append(organizationalUnits, d.Get("parent_id").(string))

    // Get all descendant organizational units.
    var descendantOUs []*organizations.OrganizationalUnit

    if descendantOUs = getDescendantOrganizationalUnits(parent_id); descendantOUs != nil{
	    organizationalUnits = append(organizationalUnits, descendantOUs)
	}

    var accounts []*organizations.Account

    for _, ou := range organizationalUnits {
         // Get immediate child accounts of ou.
         parent_id = ou.get("Id")
        params := &organizations.ListAccountsForParentInput{
            ParentId: aws.String(parent_id),
        }

        err := conn.ListAccountsForParentPages(params,
            func(page *organizations.ListAccountsForParentOutput, lastPage bool) bool {
                accounts = append(accounts, page.Accounts...)

                return !lastPage
            })

        if err != nil {
            return fmt.Errorf("error listing Organizations Accounts for parent (%s): %w", parent_id, err)
        }
    }

	d.SetId(parent_id)

	if err := d.Set("accounts", flattenAccounts(accounts)); err != nil {
		return fmt.Errorf("error setting accounts: %w", err)
	}

	return nil
}

func flattenOrganizationalUnits(ous []*organizations.OrganizationalUnit) []map[string]interface{} {
	if len(ous) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, ou := range ous {
		result = append(result, map[string]interface{}{
			"arn":  aws.StringValue(ou.Arn),
			"id":   aws.StringValue(ou.Id),
			"name": aws.StringValue(ou.Name),
		})
	}
	return result
}

func flattenAccounts(accounts []*organizations.Account) []map[string]interface{} {
	if len(ous) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, account := range accounts {
		result = append(result, map[string]interface{}{
			"arn":              aws.StringValue(account.Arn),
			"email":            aws.StringValue(account.Email),
			"id":               aws.StringValue(account.Id),
			"joined_method":    aws.StringValue(account.JoinedMethod),
			"joined_timestamp": aws.StringValue(account.JoinedTimestamp),
			"name":             aws.StringValue(account.Name),
			"status":           aws.StringValue(account.Status),
		})
	}
	return result
}
