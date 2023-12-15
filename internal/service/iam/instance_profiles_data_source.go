// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_iam_instance_profiles")
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
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	roleName := d.Get("role_name").(string)
	instanceProfiles, err := findInstanceProfilesForRole(ctx, conn, roleName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Instance Profiles for Role (%s): %s", roleName, err)
	}

	var arns, names, paths []string

	for _, v := range instanceProfiles {
		arns = append(arns, aws.StringValue(v.Arn))
		names = append(names, aws.StringValue(v.InstanceProfileName))
		paths = append(paths, aws.StringValue(v.Path))
	}

	d.SetId(roleName)
	d.Set("arns", arns)
	d.Set("names", names)
	d.Set("paths", paths)

	return diags
}

func findInstanceProfilesForRole(ctx context.Context, conn *iam.IAM, roleName string) ([]*iam.InstanceProfile, error) {
	input := &iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	}
	var output []*iam.InstanceProfile

	err := conn.ListInstanceProfilesForRolePagesWithContext(ctx, input, func(page *iam.ListInstanceProfilesForRoleOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceProfiles {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
