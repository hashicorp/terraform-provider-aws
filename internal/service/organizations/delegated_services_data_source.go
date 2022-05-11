package organizations

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
	conn := meta.(*conns.AWSClient).OrganizationsConn

	input := &organizations.ListDelegatedServicesForAccountInput{
		AccountId: aws.String(d.Get("account_id").(string)),
	}

	var delegators []*organizations.DelegatedService
	err := conn.ListDelegatedServicesForAccountPagesWithContext(ctx, input, func(page *organizations.ListDelegatedServicesForAccountOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		delegators = append(delegators, page.DelegatedServices...)

		return !lastPage
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error describing organizations delegated services: %w", err))
	}

	if err = d.Set("delegated_services", flattenDelegatedServices(delegators)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting delegated_services: %w", err))
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return nil
}

func flattenDelegatedServices(delegatedServices []*organizations.DelegatedService) []map[string]interface{} {
	if len(delegatedServices) == 0 {
		return nil
	}

	var result []map[string]interface{}
	for _, delegated := range delegatedServices {
		result = append(result, map[string]interface{}{
			"delegation_enabled_date": aws.TimeValue(delegated.DelegationEnabledDate).Format(time.RFC3339),
			"service_principal":       aws.StringValue(delegated.ServicePrincipal),
		})
	}
	return result
}
