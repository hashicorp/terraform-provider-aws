// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_organizational_unit", name="Organizational Unit")
func DataSourceOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationalUnitRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parent_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[0-9a-z]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_CreateOrganizationalUnit.html#organizations-CreateOrganizationalUnit-request-ParentId"),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameOrganizationalUnit = "Organizational Unit Data Source"
)

func dataSourceOrganizationalUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	name := d.Get("name").(string)
	parentID := d.Get("parent_id").(string)

	output, err := findOrganizationalUnitsForParent(ctx, conn, parentID)
	if err != nil {
		return append(diags, create.DiagError(names.Organizations, create.ErrActionReading, DSNameOrganizationalUnit, name, err)...)
	}
	if len(output) == 0 {
		return sdkdiag.AppendErrorf(diags, "Organizational parent not found (%s)", parentID)
	}

	for _, v := range output {
		if v.Name != nil && aws.StringValue(v.Name) == name && v.Id != nil {
			d.SetId(aws.StringValue(v.Id))
			d.Set("arn", v.Arn)
			return diags
		}
	}

	return sdkdiag.AppendErrorf(diags, "No matching organization for (%s): %s", parentID, name)
}
