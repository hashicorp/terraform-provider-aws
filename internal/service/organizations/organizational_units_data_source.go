// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_organizational_units", name="Organizational Unit")
func dataSourceOrganizationalUnits() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationalUnitsRead,

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

func dataSourceOrganizationalUnitsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	parentID := d.Get("parent_id").(string)
	children, err := findOrganizationalUnitsForParentByID(ctx, conn, parentID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Organization Units for parent (%s): %s", parentID, err)
	}

	d.SetId(parentID)
	if err := d.Set("children", flattenOrganizationalUnits(children)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting children: %s", err)
	}

	return diags
}

func findOrganizationalUnitsForParentByID(ctx context.Context, conn *organizations.Client, id string) ([]awstypes.OrganizationalUnit, error) {
	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(id),
	}

	return findOrganizationalUnitsForParent(ctx, conn, input, tfslices.PredicateTrue[*awstypes.OrganizationalUnit]())
}

func findOrganizationalUnitForParent(ctx context.Context, conn *organizations.Client, input *organizations.ListOrganizationalUnitsForParentInput, filter tfslices.Predicate[*awstypes.OrganizationalUnit]) (*awstypes.OrganizationalUnit, error) {
	output, err := findOrganizationalUnitsForParent(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOrganizationalUnitsForParent(ctx context.Context, conn *organizations.Client, input *organizations.ListOrganizationalUnitsForParentInput, filter tfslices.Predicate[*awstypes.OrganizationalUnit]) ([]awstypes.OrganizationalUnit, error) {
	var output []awstypes.OrganizationalUnit

	pages := organizations.NewListOrganizationalUnitsForParentPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ParentNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.OrganizationalUnits {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func flattenOrganizationalUnits(apiObjects []awstypes.OrganizationalUnit) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, ou := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			names.AttrARN:  aws.ToString(ou.Arn),
			names.AttrID:   aws.ToString(ou.Id),
			names.AttrName: aws.ToString(ou.Name),
		})
	}

	return tfList
}
