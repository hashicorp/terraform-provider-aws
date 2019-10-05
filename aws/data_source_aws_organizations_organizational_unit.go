package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationsOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsOrganizationalUnitRead,

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"children": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
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
					},
				},
			},
		},
	}
}

func dataSourceAwsOrganizationsOrganizationalUnitRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	parent_id := d.Get("parent_id").(string)
	d.SetId(resource.UniqueId())

	params := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parent_id),
	}

	var children []*organizations.OrganizationalUnit
	for {
		ous, err := conn.ListOrganizationalUnitsForParent(params)

		if err != nil {
			return fmt.Errorf("Error listing organizational units for parent: %s", err)
		}

		for _, ou := range ous.OrganizationalUnits {
			children = append(children, ou)
		}

		if ous.NextToken == nil {
			break
		}
		params.NextToken = ous.NextToken
	}

	if err := d.Set("children", flattenOrganizationsOrganizationalUnits(children)); err != nil {
		return fmt.Errorf("Error setting children.")
	}

	return nil
}
