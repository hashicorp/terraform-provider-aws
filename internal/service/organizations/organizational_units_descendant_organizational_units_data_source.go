// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_organizational_unit_descendant_organizational_units", name="Organizational Unit Descendant Organization Units")
func dataSourceOrganizationalUnitDescendantOrganizationalUnits() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationalUnitDescendantOrganizationalUnitsRead,

		Schema: map[string]*schema.Schema{
			"children": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceOrganizationalUnitDescendantOrganizationalUnitsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	parentID := d.Get("parent_id").(string)
	organizationUnits, err := findAllOrganizationalUnitsForParentAndBelow(ctx, conn, parentID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Organizational Units for parent (%s) and descendants: %s", parentID, err)
	}

	d.SetId(parentID)

	if err := d.Set("children", flattenOrganizationalUnits(organizationUnits)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting children: %s", err)
	}

	return diags
}

// findAllOrganizationalUnitsForParentAndBelow recurses down an OU tree, returning all organizational units at the specified parent and below.
func findAllOrganizationalUnitsForParentAndBelow(ctx context.Context, conn *organizations.Client, id string) ([]awstypes.OrganizationalUnit, error) {
	var output []awstypes.OrganizationalUnit

	ous, err := findOrganizationalUnitsForParentByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	output = append(output, ous...)

	for _, ou := range ous {
		organizationUnits, err := findAllOrganizationalUnitsForParentAndBelow(ctx, conn, aws.ToString(ou.Id))

		if err != nil {
			return nil, err
		}

		output = append(output, organizationUnits...)
	}

	return output, nil
}
