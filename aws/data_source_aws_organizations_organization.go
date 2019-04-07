package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
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
			"feature_set": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_email": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"master_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsOrganizationsOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	svc := meta.(*AWSClient).organizationsconn

	log.Println("[DEBUG] Describing Organization")
	out, err := svc.DescribeOrganization(&organizations.DescribeOrganizationInput{})
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(out.Organization.Id))
	d.Set("arn", out.Organization.Arn)

	availablePolicyTypes := make([]map[string]interface{}, 0)
	for _, v := range out.Organization.AvailablePolicyTypes {
		apt := map[string]interface{}{}
		apt["status"] = aws.StringValue(v.Status)
		apt["type"] = aws.StringValue(v.Type)
		availablePolicyTypes = append(availablePolicyTypes, apt)
	}
	if err := d.Set("available_policy_types", availablePolicyTypes); err != nil {
		return fmt.Errorf("error setting available_policy_types: %s", err)
	}

	d.Set("feature_set", out.Organization.FeatureSet)
	d.Set("master_account_arn", out.Organization.MasterAccountArn)
	d.Set("master_account_email", out.Organization.MasterAccountEmail)
	d.Set("master_account_id", out.Organization.MasterAccountId)

	return nil
}
