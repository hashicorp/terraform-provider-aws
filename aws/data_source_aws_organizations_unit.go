package aws

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsOrganizationUnit() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationUnitRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsOrganizationUnitRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.ListRootsInput{}
	result, err := conn.ListRoots(input)
	if err != nil {
		return err
	}

	// there should be exactly one, per https://docs.aws.amazon.com/organizations/latest/userguide/orgs_getting-started_concepts.html#root
	root := result.Roots[0]
	if root == nil {
		return errors.New("Root organizational unit not found")
	}

	d.SetId(*root.Id)
	d.Set("arn", root.Arn)

	return nil
}
