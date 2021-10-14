package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsOrganizationsAccounts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsAccountsRead,

		Schema: map[string]*schema.Schema{
			"account_ids": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
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
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
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
						"tags": tagsSchemaComputed(),
					},
				},
			},
		},
	}
}

func dataSourceAwsOrganizationsAccountsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	var account_ids = expandStringList(d.Get("account_ids").([]interface{}))

	var accounts []map[string]interface{}

	for _, account_id := range account_ids {
		params := &organizations.DescribeAccountInput{
			AccountId: aws.String(*account_id),
		}

		acc, err := conn.DescribeAccount(params)
		if err != nil {
			return fmt.Errorf("Error describing account: %w", err)
		}

		params_tags := &organizations.ListTagsForResourceInput{
			ResourceId: aws.String(*account_id),
		}

		rt, err := conn.ListTagsForResource(params_tags)
		valueTags := keyvaluetags.OrganizationsKeyValueTags(rt.Tags)
		tags := valueTags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()
		var account = map[string]interface{}{
			"arn":              aws.StringValue(acc.Account.Arn),
			"email":            aws.StringValue(acc.Account.Email),
			"name":             aws.StringValue(acc.Account.Name),
			"status":           aws.StringValue(acc.Account.Status),
			"joined_method":    aws.StringValue(acc.Account.JoinedMethod),
			"joined_timestamp": aws.TimeValue(acc.Account.JoinedTimestamp).Format(time.RFC3339),
			"tags":             tags,
		}

		accounts = append(accounts, account)
	}

	if err := d.Set("accounts", accounts); err != nil {
		return fmt.Errorf("error setting accounts: %w", err)
	}

	d.SetId(meta.(*AWSClient).accountid)

	return nil
}
