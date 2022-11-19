package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceAttachedGroupPolicies() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAttachedGroupPoliciesRead,

		Schema: map[string]*schema.Schema{
			"group": {
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

func dataSourceAttachedGroupPoliciesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	group := d.Get("group").(string)
	pathPrefix := d.Get("path_prefix").(string)
	input := &iam.ListAttachedGroupPoliciesInput{
		GroupName:  aws.String(group),
		PathPrefix: aws.String(pathPrefix),
	}

	var arns, names []string

	log.Printf("[DEBUG] Reading IAM Attached Group Policies: %s", input)
	err := conn.ListAttachedGroupPoliciesPages(input,
		func(page *iam.ListAttachedGroupPoliciesOutput, lastPage bool) bool {
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
		return fmt.Errorf("error getting attached group policies: %w", err)
	}

	d.SetId(group)
	d.Set("arns", arns)
	d.Set("names", names)

	return nil
}
