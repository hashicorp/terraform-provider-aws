package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceGroupPolicies() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGroupPoliciesRead,

		Schema: map[string]*schema.Schema{
			"group": {
				Type:     schema.TypeString,
				Required: true,
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceGroupPoliciesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	group := d.Get("group").(string)
	input := &iam.ListGroupPoliciesInput{
		GroupName: aws.String(group),
	}

	var names []string

	log.Printf("[DEBUG] Reading IAM Group Policies: %s", input)
	err := conn.ListGroupPoliciesPages(input,
		func(page *iam.ListGroupPoliciesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, name := range page.PolicyNames {
				if name == nil {
					continue
				}

				names = append(names, aws.StringValue(name))
			}

			return !lastPage
		},
	)
	if err != nil {
		return fmt.Errorf("error getting group policies: %w", err)
	}

	d.SetId(group)
	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("error setting names: %w", err)
	}

	return nil
}
