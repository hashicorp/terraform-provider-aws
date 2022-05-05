package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceInstanceProfiles() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceInstanceProfilesRead,

		Schema: map[string]*schema.Schema{
			"role_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validIamResourceName(roleNameMaxLen),
			},
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"paths": {
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

func dataSourceInstanceProfilesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	roleName := d.Get("role_name").(string)

	req := &iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	}

	log.Printf("[DEBUG] Reading IAM Instance Profiles for given role %s: %s", roleName, req)
	resp, err := conn.ListInstanceProfilesForRole(req)
	if err != nil {
		return fmt.Errorf("Error getting instance profiles: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("no IAM instance profiles found for role %s", roleName)
	}

	instanceProfiles := resp.InstanceProfiles

	var arns, paths, names []string
	for _, profile := range instanceProfiles {
		arns = append(arns, aws.StringValue(profile.Arn))
		paths = append(paths, aws.StringValue(profile.Path))
		names = append(names, aws.StringValue(profile.InstanceProfileName))
	}

	d.SetId(roleName)
	d.Set("arns", arns)
	d.Set("paths", paths)
	d.Set("names", names)

	return nil
}
