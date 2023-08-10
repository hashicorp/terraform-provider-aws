// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_organizations_organizational_units")
func DataSourceOrganizationalUnits() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationalUnitsRead,

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"children": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceOrganizationalUnitsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	parentID := d.Get("parent_id").(string)
	children, err := findOUsForParent(ctx, conn, parentID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Organization Units for parent (%s): %s", parentID, err)
	}

	d.SetId(parentID)

	if err := d.Set("children", FlattenOrganizationalUnits(children)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting children: %s", err)
	}

	return diags
}

func findOUsForParent(ctx context.Context, conn *organizations.Organizations, id string) ([]*organizations.OrganizationalUnit, error) {
	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(id),
	}
	var output []*organizations.OrganizationalUnit

	err := conn.ListOrganizationalUnitsForParentPagesWithContext(ctx, input, func(page *organizations.ListOrganizationalUnitsForParentOutput, lastPage bool) bool {
		output = append(output, page.OrganizationalUnits...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FlattenOrganizationalUnits(ous []*organizations.OrganizationalUnit) []map[string]interface{} {
	if len(ous) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, ou := range ous {
		result = append(result, map[string]interface{}{
			"arn":  aws.StringValue(ou.Arn),
			"id":   aws.StringValue(ou.Id),
			"name": aws.StringValue(ou.Name),
		})
	}
	return result
}
