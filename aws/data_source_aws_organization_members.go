package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/organizations"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationMembers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsMembersRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"accounts": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsOrganizationsMembersRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	var accounts []string

	err := conn.ListAccountsPages(&organizations.ListAccountsInput{}, func(resp *organizations.ListAccountsOutput, isLast bool) bool {
		for _, res := range resp.Accounts {
			accounts = append(accounts, *res.Id)
		}
		return !isLast
	})
	if err != nil {
		if orgerr, ok := err.(awserr.Error); ok && orgerr.Code() == "ParentNotFoundException" {
			log.Printf("[WARN] Organization %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
	}

	if len(accounts) < 1 {
		return fmt.Errorf("Your query returned no results. Please check your account ID and try again")
	}

	log.Printf("[DEBUG] Found %d accounts", len(accounts))

	d.SetId(resource.UniqueId())
	err = d.Set("accounts", accounts)
	if err != nil {
		return err
	}

	return nil
}
