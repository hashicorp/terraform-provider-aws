// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_user_hierarchy_structure", name="User Hierarchy Structure")
func dataSourceUserHierarchyStructure() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserHierarchyStructureRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"hierarchy_structure": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"level_one":   sdkv2.DataSourcePropertyFromResourceProperty(hierarchyStructureLevelSchema()),
							"level_two":   sdkv2.DataSourcePropertyFromResourceProperty(hierarchyStructureLevelSchema()),
							"level_three": sdkv2.DataSourcePropertyFromResourceProperty(hierarchyStructureLevelSchema()),
							"level_four":  sdkv2.DataSourcePropertyFromResourceProperty(hierarchyStructureLevelSchema()),
							"level_five":  sdkv2.DataSourcePropertyFromResourceProperty(hierarchyStructureLevelSchema()),
						},
					},
				},
				names.AttrInstanceID: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 100),
				},
			}
		},
	}
}

func dataSourceUserHierarchyStructureRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	hierarchyStructure, err := findUserHierarchyStructureByID(ctx, conn, instanceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect User Hierarchy Structure (%s): %s", instanceID, err)
	}

	d.SetId(instanceID)
	if err := d.Set("hierarchy_structure", flattenHierarchyStructure(hierarchyStructure)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting hierarchy_structure: %s", err)
	}

	return diags
}
