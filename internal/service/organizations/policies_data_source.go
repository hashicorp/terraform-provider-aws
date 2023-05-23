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

// @SDKDataSource("aws_organizations_policies", name="Policies")
func DataSourcePolicies() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePoliciesRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"aws_managed": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"description": {
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
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	filter := d.Get("filter").(string)

	policies, err := listPolicies(ctx, conn, filter)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Policies with filter(%s): %s", filter, err)
	}

	d.SetId(filter)

	if err := d.Set("policies", flattenPolicySummaries(policies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting policies: %s", err)
	}

	return diags
}

func listPolicies(ctx context.Context, conn *organizations.Organizations, filter string) ([]*organizations.PolicySummary, error) {
	input := &organizations.ListPoliciesInput{
		Filter: aws.String(filter),
	}
	var output []*organizations.PolicySummary
	err := conn.ListPoliciesPagesWithContext(ctx, input, func(page *organizations.ListPoliciesOutput, lastPage bool) bool {
		output = append(output, page.Policies...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func flattenPolicySummaries(summaries []*organizations.PolicySummary) []map[string]interface{} {
	if len(summaries) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, s := range summaries {
		result = append(result, map[string]interface{}{
			"arn":         aws.StringValue(s.Arn),
			"aws_managed": aws.BoolValue(s.AwsManaged),
			"description": aws.StringValue(s.Description),
			"id":          aws.StringValue(s.Id),
			"name":        aws.StringValue(s.Name),
			"type":        aws.StringValue(s.Type),
		})
	}
	return result
}
