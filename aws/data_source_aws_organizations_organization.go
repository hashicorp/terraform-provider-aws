package aws

import (
	"log"

	organizations "github.com/aws/aws-sdk-go/service/organizations"
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

func dataSourceAwsOrganizationsOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	log.Printf("[INFO] Reading Organization")
	org, err := conn.DescribeOrganization(&organizations.DescribeOrganizationInput{})
	if err != nil {
		if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
			log.Printf("[WARN] Organization does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	d.SetId(*org.Organization.Id)
	d.Set("arn", org.Organization.Arn)
	d.Set("feature_set", org.Organization.FeatureSet)
	d.Set("master_account_arn", org.Organization.MasterAccountArn)
	d.Set("master_account_email", org.Organization.MasterAccountEmail)
	d.Set("master_account_id", org.Organization.MasterAccountId)
	availablePolicyTypes := availablePolicyTypeListToMap(org.Organization.AvailablePolicyTypes)
	d.Set("available_policy_types", availablePolicyTypes)
	return nil
}

func availablePolicyTypeListToMap(list []*organizations.PolicyTypeSummary) []map[string]interface{} {
	var output []map[string]interface{}
	for _, o := range list {
		policyTypeSummary := map[string]interface{}{
			"status": *o.Status,
			"type":   *o.Type,
		}
		output = append(output, policyTypeSummary)
	}
	return output
}
