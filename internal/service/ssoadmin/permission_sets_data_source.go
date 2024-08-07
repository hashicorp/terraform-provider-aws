// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssoadmin_permission_sets")
func DataSourcePermissionSets() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePermissionSetsRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func dataSourcePermissionSetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	instanceArn := d.Get("instance_arn").(string)

	input := &ssoadmin.ListPermissionSetsInput{
		InstanceArn: aws.String(instanceArn),
	}

	var permissionSetArns []string
	paginator := ssoadmin.NewListPermissionSetsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing SSO Permission Sets: %s", err)
		}

		permissionSetArns = append(permissionSetArns, page.PermissionSets...)
	}

	d.SetId(instanceArn)
	d.Set(names.AttrARNs, permissionSetArns)

	return diags
}
