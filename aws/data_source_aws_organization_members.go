package aws

import (
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationMembers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationMembersRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"emails": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"accounts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
			},
			"total_accounts": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsOrganizationMembersRead(d *schema.ResourceData, meta interface{}) error {
	var token *string
	var accountList []map[string]string
	var idList []string
	var nameList []string
	var emailList []string
	var arnList []string

	svc := meta.(*AWSClient).organizationsconn

	for {
		input := &organizations.ListAccountsInput{
			NextToken: token,
		}

		result, err := svc.ListAccounts(input)
		if err != nil {
			return err
		}

		for _, account := range result.Accounts {
			accountInfo := make(map[string]string)
			accountInfo["Arn"] = *account.Arn
			accountInfo["Email"] = *account.Email
			accountInfo["Id"] = *account.Id
			accountInfo["JoinedMethod"] = *account.JoinedMethod
			accountInfo["JoinedTimestamp"] = (*account.JoinedTimestamp).String()
			accountInfo["Name"] = *account.Name
			accountInfo["Status"] = *account.Status

			accountList = append(accountList, accountInfo)
			arnList = append(arnList, accountInfo["Arn"])
			emailList = append(emailList, accountInfo["Email"])
			idList = append(idList, accountInfo["Id"])
			nameList = append(nameList, accountInfo["Name"])
		}

		if result.NextToken == nil {
			break
		} else {
			token = result.NextToken
		}
	}

	d.SetId(resource.UniqueId())
	d.Set("accounts", accountList)
	d.Set("arns", arnList)
	d.Set("emails", emailList)
	d.Set("ids", idList)
	d.Set("names", nameList)
	d.Set("total_accounts", len(idList))

	return nil
}
