// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_organizational_units")
func DataSourceOrganizationalUnits() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

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

func findOrganizationalUnitsForParentByID(ctx context.Context, conn *organizations.Organizations, id string) ([]*organizations.OrganizationalUnit, error) {
	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(id),
	}

	return findOrganizationalUnitsForParent(ctx, conn, input, tfslices.PredicateTrue[*organizations.OrganizationalUnit]())
}

func findOrganizationalUnitForParent(ctx context.Context, conn *organizations.Organizations, input *organizations.ListOrganizationalUnitsForParentInput, filter tfslices.Predicate[*organizations.OrganizationalUnit]) (*organizations.OrganizationalUnit, error) {
	output, err := findOrganizationalUnitsForParent(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findOrganizationalUnitsForParent(ctx context.Context, conn *organizations.Organizations, input *organizations.ListOrganizationalUnitsForParentInput, filter tfslices.Predicate[*organizations.OrganizationalUnit]) ([]*organizations.OrganizationalUnit, error) {
	var output []*organizations.OrganizationalUnit

	err := conn.ListOrganizationalUnitsForParentPagesWithContext(ctx, input, func(page *organizations.ListOrganizationalUnitsForParentOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.OrganizationalUnits {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeParentNotFoundException) {
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

func flattenOrganizationalUnits(ous []*organizations.OrganizationalUnit) []map[string]interface{} {
	if len(ous) == 0 {
		return nil
	}

	var result []map[string]interface{}
	for _, ou := range ous {
		result = append(result, map[string]interface{}{
			names.AttrARN:  aws.StringValue(ou.Arn),
			names.AttrID:   aws.StringValue(ou.Id),
			names.AttrName: aws.StringValue(ou.Name),
		})
	}
	return result
}
