// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	intOrg "github.com/hashicorp/terraform-provider-aws/internal/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_organizational_unit", name="Organizational Unit")
func dataSourceOrganizationalUnit() *schema.Resource {
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
			"principal_org_path": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The full path of the organizational unit within the AWS Organization. " +
					"See https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_last-accessed-view-data-orgs.html#access_policies_last-accessed-viewing-orgs-entity-path",
			},
		},
	}
}

func dataSourceOrganizationalUnitRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	name := d.Get(names.AttrName).(string)
	parentID := d.Get("parent_id").(string)
	input := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parentID),
	}

	ou, err := findOrganizationalUnitForParent(ctx, conn, input, func(v *awstypes.OrganizationalUnit) bool {
		return aws.ToString(v.Name) == name
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Organizational Unit (%s/%s): %s", parentID, name, err)
	}

	d.SetId(aws.ToString(ou.Id))
	d.Set(names.AttrARN, ou.Arn)

	principalOrgPath, err := intOrg.BuildPrincipalOrgPath(ctx, conn, aws.ToString(ou.Id))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "computing principal org path for OU %s: %s", aws.ToString(ou.Id), err)
	}

	d.Set("principal_org_path", principalOrgPath)

	return diags
}
