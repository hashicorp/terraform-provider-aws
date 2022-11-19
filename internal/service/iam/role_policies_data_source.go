package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceRolePolicies() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRolePoliciesRead,

		Schema: map[string]*schema.Schema{
			"role": {
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

func dataSourceRolePoliciesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	role := d.Get("role").(string)
	input := &iam.ListRolePoliciesInput{
		RoleName: aws.String(role),
	}

	var names []string

	log.Printf("[DEBUG] Reading IAM Role Policies: %s", input)
	err := conn.ListRolePoliciesPages(input,
		func(page *iam.ListRolePoliciesOutput, lastPage bool) bool {
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
		return fmt.Errorf("error getting role policies: %w", err)
	}

	d.SetId(role)
	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("error setting names: %w", err)
	}

	return nil
}
