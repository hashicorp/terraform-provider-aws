package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationsAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsAccountRead,
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_user_access_to_billing": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsOrganizationsAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	accountID := d.Get("account_id").(string)
	req := &organizations.DescribeAccountInput{
		AccountId: aws.String(accountID),
	}

	log.Printf("[DEBUG] DescribeAccount %s\n", req)
	resp, err := conn.DescribeAccount(req)
	if err != nil {
		return err
	}

	account := resp.Account

	log.Printf("[DEBUG] DescribeAccountResponse %s\n", account)

	d.SetId(*account.Id)
	d.Set("account_id", account.Id)
	d.Set("arn", account.Arn)
	d.Set("email", account.Email)
	d.Set("joined_method", account.JoinedMethod)
	d.Set("joined_timestamp", account.JoinedTimestamp)
	d.Set("name", account.Name)
	d.Set("status", account.Status)

	return nil
}
