package iam

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceInstanceProfiles() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceProfilesRead,

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
				ValidateFunc: validResourceName(roleNameMaxLen),
			},
		},
	}
}

func dataSourceInstanceProfilesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IAMConn

	roleName := d.Get("role_name").(string)
	input := &iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	}
	var arns, names, paths []string

	err := conn.ListInstanceProfilesForRolePagesWithContext(ctx, input, func(page *iam.ListInstanceProfilesForRoleOutput, lastPage bool) bool {
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
		return diag.Errorf("listing IAM Instance Profiles for Role (%s): %s", roleName, err)
	}

	d.SetId(roleName)
	d.Set("arns", arns)
	d.Set("names", names)
	d.Set("paths", paths)

	return nil
}
