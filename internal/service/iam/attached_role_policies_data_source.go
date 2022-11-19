package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceAttachedRolePolicies() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAttachedRolePoliciesRead,

		Schema: map[string]*schema.Schema{
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAttachedRolePoliciesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	role := d.Get("role").(string)
	pathPrefix := d.Get("path_prefix").(string)
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName:   aws.String(role),
		PathPrefix: aws.String(pathPrefix),
	}

	var arns, names []string

	log.Printf("[DEBUG] Reading IAM Attached Role Policies: %s", input)
	err := conn.ListAttachedRolePoliciesPages(input,
		func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, p := range page.AttachedPolicies {
				if p == nil {
					continue
				}

				arns = append(arns, aws.StringValue(p.PolicyArn))
				names = append(names, aws.StringValue(p.PolicyName))
			}

			return !lastPage
		},
	)
	if err != nil {
		return fmt.Errorf("error getting attached role policies: %w", err)
	}

	d.SetId(role)
	d.Set("arns", arns)
	d.Set("names", names)

	return nil
}
