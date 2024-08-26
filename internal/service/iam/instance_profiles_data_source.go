// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_instance_profiles", name="Instance Profiles")
func dataSourceInstanceProfiles() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceProfilesRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrNames: {
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

	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	roleName := d.Get("role_name").(string)
	instanceProfiles, err := findInstanceProfilesForRole(ctx, conn, roleName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Instance Profiles for Role (%s): %s", roleName, err)
	}

	var arns, nms, paths []string

	for _, v := range instanceProfiles {
		arns = append(arns, aws.ToString(v.Arn))
		nms = append(nms, aws.ToString(v.InstanceProfileName))
		paths = append(paths, aws.ToString(v.Path))
	}

	d.SetId(roleName)
	d.Set(names.AttrARNs, arns)
	d.Set(names.AttrNames, nms)
	d.Set("paths", paths)

	return diags
}

func findInstanceProfilesForRole(ctx context.Context, conn *iam.Client, roleName string) ([]awstypes.InstanceProfile, error) {
	input := &iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	}
	var output []awstypes.InstanceProfile

	pages := iam.NewListInstanceProfilesForRolePaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.InstanceProfiles {
			if !reflect.ValueOf(v).IsZero() {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
