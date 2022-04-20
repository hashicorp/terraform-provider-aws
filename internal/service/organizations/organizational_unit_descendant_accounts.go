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
func getDescendantOrganizationalUnitsIDs(parent_id string, meta interface{}) ([]string, error) {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	params := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parent_id),
	}

	// Descendants will hold IDs of all generations of OUs under parent_id.
	var descendants []string
	// Children will hold immediate child OUs of parent_id.
	var children []*organizations.OrganizationalUnit
	// ChildOUIds will hold the OU IDs of child ous.
	var childOUIds []string
	// result will hold any descendants of immediate child OUs.
	var result []string

	// Get all immediate children OUs.
	err := conn.ListOrganizationalUnitsForParentPages(params,
		func(page *organizations.ListOrganizationalUnitsForParentOutput, lastPage bool) bool {
			children = append(children, page.OrganizationalUnits...)

			return !lastPage
		})

	if err != nil {
		return nil, fmt.Errorf("error listing Organizations Organization Units for parent (%s): %w", parent_id, err)
	}

	if childOUIds = getIDsFromOUs(children); childOUIds != nil {
		for _, id := range childOUIds {
			// Append this child ou Id and all it's descendants Ids.
			descendants = append(descendants, id)

			result, err = getDescendantOrganizationalUnitsIDs(id, meta)
			if err != nil {
				return descendants, err
			} else if descendants != nil {
				descendants = append(descendants, result...)
			} else {
				return nil, nil
			}
		}

		return descendants, nil
	}

	//Base Case. ParentId has no children.
	return nil, nil
}

func dataSourceOrganizationalUnitDescendantAccountsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OrganizationsConn

	parent_id := d.Get("parent_id").(string)

	// Collect all the OUs of which we need to list children.
	var organizationalUnitIDs []string
	organizationalUnitIDs = append(organizationalUnitIDs, parent_id)

	// Get all descendant organizational units.
	// 	var descendantOUs []*organizations.OrganizationalUnit

	descendantOUs, err := getDescendantOrganizationalUnitsIDs(organizationalUnitIDs[0], meta)
	if err != nil {
		return err
	} else if descendantOUs != nil {
		// If descendant OUs are found, add them to organizationalUnits.
		organizationalUnitIDs = append(organizationalUnitIDs, descendantOUs...)
	}

	var accounts []*organizations.Account

	for _, id := range organizationalUnitIDs {
		// Get immediate child accounts of ou.
		params := &organizations.ListAccountsForParentInput{
			ParentId: aws.String(id),
		}

		err := conn.ListAccountsForParentPages(params,
			func(page *organizations.ListAccountsForParentOutput, lastPage bool) bool {
				accounts = append(accounts, page.Accounts...)

				return !lastPage
			})

		if err != nil {
			return fmt.Errorf("error listing Organizations Accounts for parent (%s): %w", id, err)
		}
	}

	d.SetId(parent_id)

	if err := d.Set("accounts", flattenAccounts(accounts)); err != nil {
		return fmt.Errorf("error setting accounts: %w", err)
	}

	return nil
}

func getIDsFromOUs(ous []*organizations.OrganizationalUnit) []string {
	if len(ous) == 0 {
		return nil
	}
	var result []string
	for _, ou := range ous {
		result = append(result, aws.StringValue(ou.Id))
	}
	return result
}

// func flattenAccounts(accounts []*organizations.Account) []map[string]interface{} {
// 	if len(accounts) == 0 {
// 		return nil
// 	}
// 	var result []map[string]interface{}
// 	for _, account := range accounts {
// 		result = append(result, map[string]interface{}{
// 			"arn":              aws.StringValue(account.Arn),
// 			"email":            aws.StringValue(account.Email),
// 			"id":               aws.StringValue(account.Id),
// 			"joined_method":    aws.StringValue(account.JoinedMethod),
// 			"joined_timestamp": aws.StringValue(account.JoinedTimestamp),
// 			"name":             aws.StringValue(account.Name),
// 			"status":           aws.StringValue(account.Status),
// 		})
// 	}
// 	return result
// }
