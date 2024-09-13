// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_delegated_administrators", name="Delegated Administrators")
func dataSourceDelegatedAdministrators() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDelegatedAdministratorsRead,

		Schema: map[string]*schema.Schema{
			"delegated_administrators": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"delegation_enabled_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEmail: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"joined_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"joined_timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"service_principal": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func dataSourceDelegatedAdministratorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	input := &organizations.ListDelegatedAdministratorsInput{}

	if v, ok := d.GetOk("service_principal"); ok {
		input.ServicePrincipal = aws.String(v.(string))
	}

	output, err := findDelegatedAdministrators(ctx, conn, input, tfslices.PredicateTrue[*awstypes.DelegatedAdministrator]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Delegated Administrators: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)
	if err = d.Set("delegated_administrators", flattenDelegatedAdministrators(output)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting delegated_administrators: %s", err)
	}

	return nil
}

func flattenDelegatedAdministrators(apiObjects []awstypes.DelegatedAdministrator) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			names.AttrARN:             aws.ToString(apiObject.Arn),
			"delegation_enabled_date": aws.ToTime(apiObject.DelegationEnabledDate).Format(time.RFC3339),
			names.AttrEmail:           aws.ToString(apiObject.Email),
			names.AttrID:              aws.ToString(apiObject.Id),
			"joined_method":           apiObject.JoinedMethod,
			"joined_timestamp":        aws.ToTime(apiObject.JoinedTimestamp).Format(time.RFC3339),
			names.AttrName:            aws.ToString(apiObject.Name),
			names.AttrStatus:          apiObject.Status,
		})
	}

	return tfList
}
