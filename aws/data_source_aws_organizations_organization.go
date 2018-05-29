package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationsOrganization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsOrganizationRead,

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func dataSourceAwsOrganizationsOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	req := &organizations.DescribeOrganizationInput{}

	log.Printf("[DEBUG] Reading Organization: %s", req)
	resp, err := conn.DescribeOrganization(req)
	if err != nil {
		return err
	}

	org := resp.Organization

	d.SetId(*org.Id)
	d.Set("arn", org.Arn)
	d.Set("feature_set", org.FeatureSet)
	d.Set("master_account_arn", org.MasterAccountArn)
	d.Set("master_account_email", org.MasterAccountEmail)
	d.Set("master_account_id", org.MasterAccountId)

	log.Printf("[DEBUG] Reading Organization: %+v\n", org)

	return nil
}
