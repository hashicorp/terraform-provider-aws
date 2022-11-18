package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceAttachedUserPolicies() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAttachedUserPoliciesRead,

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAttachedUserPoliciesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	user := d.Get("user").(string)
	pathPrefix := d.Get("path_prefix").(string)
	req := &iam.ListAttachedUserPoliciesInput{
		UserName:   aws.String(user),
		PathPrefix: aws.String(pathPrefix),
	}

	arns := []string{}
	names := []string{}

	log.Printf("[DEBUG] Reading IAM Attached User Policies: %s", req)
	err := conn.ListAttachedUserPoliciesPages(req,
		func(resp *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool {
			for _, p := range (*resp).AttachedPolicies {
				arns = append(arns, *p.PolicyArn)
				names = append(names, *p.PolicyName)
			}
			return !lastPage
		},
	)
	if err != nil {
		return fmt.Errorf("error getting attached user policies: %w", err)
	}

	d.Set("arns", arns)
	d.Set("names", names)

	return nil
}
