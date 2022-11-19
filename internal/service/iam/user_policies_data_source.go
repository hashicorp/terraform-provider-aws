package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceUserPolicies() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserPoliciesRead,

		Schema: map[string]*schema.Schema{
			"user": {
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

func dataSourceUserPoliciesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	user := d.Get("user").(string)
	input := &iam.ListUserPoliciesInput{
		UserName: aws.String(user),
	}

	var names []string

	log.Printf("[DEBUG] Reading IAM User Policies: %s", input)
	err := conn.ListUserPoliciesPages(input,
		func(page *iam.ListUserPoliciesOutput, lastPage bool) bool {
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
		return fmt.Errorf("error getting user policies: %w", err)
	}

	d.SetId(user)
	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("error setting names: %w", err)
	}

	return nil
}
