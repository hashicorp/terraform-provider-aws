// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_organizational_unit", name="Organizational Unit")
func DataSourceOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationalUnitRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"parent_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[0-9a-z]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_CreateOrganizationalUnit.html#organizations-CreateOrganizationalUnit-request-ParentId"),
			},
		},
	}
}

func dataSourceOrganizationalUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	name := d.Get(names.AttrName).(string)
	parentID := d.Get("parent_id").(string)
	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parentID),
	}

	ou, err := findOrganizationalUnitForParent(ctx, conn, input, func(v *organizations.OrganizationalUnit) bool {
		return aws.StringValue(v.Name) == name
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Organizational Unit (%s/%s): %s", parentID, name, err)
	}

	d.SetId(aws.StringValue(ou.Id))
	d.Set(names.AttrARN, ou.Arn)

	return diags
}
