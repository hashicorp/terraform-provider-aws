package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/service/organizations"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationAttributes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsAttributesRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"feature_set": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"available_policy_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsOrganizationsAttributesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	resp, err := conn.DescribeOrganization(&organizations.DescribeOrganizationInput{})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Organization attributes %d", resp.Organization)

	d.SetId(*resp.Organization.Arn)
	d.Set("arn", resp.Organization.Arn)
	d.Set("arn", resp.Organization.Id)
	d.Set("master_account_arn", resp.Organization.MasterAccountArn)
	d.Set("master_account_email", resp.Organization.MasterAccountEmail)
	d.Set("master_account_id", resp.Organization.MasterAccountId)
	d.Set("feature_set", resp.Organization.FeatureSet)

	var policies []map[string]string

	for _, policyItem := range resp.Organization.AvailablePolicyTypes {
		policy := map[string]string{
			"status": *policyItem.Status,
			"type":   *policyItem.Type,
		}
		policies = append(policies, policy)

	}
	d.Set("available_policy_types", policies)

	return nil
}
