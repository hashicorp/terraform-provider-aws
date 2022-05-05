package iam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceInstanceProfiles() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceInstanceProfilesRead,

		Schema: map[string]*schema.Schema{
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
			"paths": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"role_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validIamResourceName(roleNameMaxLen),
			},
		},
	}
}

func dataSourceInstanceProfilesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	roleName := d.Get("role_name").(string)
	input := &iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	}
	var arns, names, paths []string

	err := conn.ListInstanceProfilesForRolePages(input, func(page *iam.ListInstanceProfilesForRoleOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceProfiles {
			arns = append(arns, aws.StringValue(v.Arn))
			names = append(names, aws.StringValue(v.InstanceProfileName))
			paths = append(paths, aws.StringValue(v.Path))
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("listing IAM Instance Profiles for Role (%s): %w", roleName, err)
	}

	d.SetId(roleName)
	d.Set("arns", arns)
	d.Set("names", names)
	d.Set("paths", paths)

	return nil
}
