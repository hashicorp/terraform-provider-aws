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

// @SDKDataSource("aws_organizations_policy")
func DataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePolicyRead,

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy_summary": {
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
				},
			},
		},
	}
}

const (
	DSNamePolicy = "Organization Policy Data Source"
)

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).OrganizationsConn()
	policyID := d.Get("policy_id").(string)

	policy, err := findPolicyByPolicyID(ctx, conn, policyID)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Policy (%s): %s", policyID, err)
	}

	d.SetId(policyID)
	if err := d.Set("policy", FlattenOrganizationPolicy(policy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting policy: %s", err)
	}

	return diags
}
func findPolicyByPolicyID(ctx context.Context, conn *organizations.Organizations, id string) ([]*organizations.Policy, error) {
	input := &organizations.DescribePolicyInput{
		PolicyId: aws.String(id),
	}
	var output []*organizations.Policy

	output, err := conn.DescribePolicyWithContext(ctx, input, func(page *organizations.DescribePolicyOutput, error) {
		output = append(output, page.Policies...)

		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func FlattenOrganizationPolicy(policies []*organizations.Policy) []map[string]interface{} {
	if len(policies) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, policy := range policies {
		result = append(result, map[string]interface{}{
			"context":     aws.StringValue(policy.Context),
			"arn":         aws.StringValue(policy.PolicySummary.Arn),
			"aws_managed": aws.BoolValue(policy.PolicySummary.AwsManaged),
			"description": aws.StringValue(policy.PolicySummary.Description),
			"id":          aws.StringValue(policy.PolicySummary.Id),
			"name":        aws.StringValue(policy.PolicySummary.Name),
			"type":        aws.StringValue(policy.PolicySummary.Type),
		})
	}
	return result
}
