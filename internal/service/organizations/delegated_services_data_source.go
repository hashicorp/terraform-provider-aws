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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_delegated_services", name="Delegated Services")
func dataSourceDelegatedServices() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDelegatedServicesRead,
		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"delegated_services": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delegation_enabled_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_principal": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDelegatedServicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	output, err := findDelegatedServicesByAccountID(ctx, conn, accountID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Delegated Services (%s): %s", accountID, err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)
	if err = d.Set("delegated_services", flattenDelegatedServices(output)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting delegated_services: %s", err)
	}

	return nil
}

func findDelegatedServicesByAccountID(ctx context.Context, conn *organizations.Client, accountID string) ([]awstypes.DelegatedService, error) {
	input := &organizations.ListDelegatedServicesForAccountInput{
		AccountId: aws.String(accountID),
	}

	return findDelegatedServices(ctx, conn, input)
}

func findDelegatedServices(ctx context.Context, conn *organizations.Client, input *organizations.ListDelegatedServicesForAccountInput) ([]awstypes.DelegatedService, error) {
	var output []awstypes.DelegatedService

	pages := organizations.NewListDelegatedServicesForAccountPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.DelegatedServices...)
	}

	return output, nil
}

func flattenDelegatedServices(apiObjects []awstypes.DelegatedService) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			"delegation_enabled_date": aws.ToTime(apiObject.DelegationEnabledDate).Format(time.RFC3339),
			"service_principal":       aws.ToString(apiObject.ServicePrincipal),
		})
	}

	return tfList
}
