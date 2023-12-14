// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_organizations_delegated_services")
func DataSourceDelegatedServices() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDelegatedServicesRead,
		Schema: map[string]*schema.Schema{
			"account_id": {
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
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	accountID := d.Get("account_id").(string)
	output, err := findDelegatedServicesByAccountID(ctx, conn, accountID)

	if err != nil {
		return diag.Errorf("reading Organizations Delegated Services (%s): %s", accountID, err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)
	if err = d.Set("delegated_services", flattenDelegatedServices(output)); err != nil {
		return diag.Errorf("setting delegated_services: %s", err)
	}

	return nil
}

func findDelegatedServicesByAccountID(ctx context.Context, conn *organizations.Organizations, accountID string) ([]*organizations.DelegatedService, error) {
	input := &organizations.ListDelegatedServicesForAccountInput{
		AccountId: aws.String(accountID),
	}
	var output []*organizations.DelegatedService

	err := conn.ListDelegatedServicesForAccountPagesWithContext(ctx, input, func(page *organizations.ListDelegatedServicesForAccountOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.DelegatedServices...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func flattenDelegatedServices(apiObjects []*organizations.DelegatedService) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			"delegation_enabled_date": aws.TimeValue(apiObject.DelegationEnabledDate).Format(time.RFC3339),
			"service_principal":       aws.StringValue(apiObject.ServicePrincipal),
		})
	}

	return tfList
}
