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

// @SDKDataSource("aws_organizations_organizational_policies")
func DataSourceOrganizationPolicies() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationPoliciesRead,

		Schema: map[string]*schema.Schema{
			"target_id": {
				Type:     schema.TypeString,
				Required: true,
			},
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

const (
	DSNameOrganizationalPolicies = "Organizational Policies Data Source"
)

func dataSourceOrganizationPoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).OrganizationsConn()

	targetID := d.Get("target_id").(string)
	filter := d.Get("filter").(string)

	policies, err := findPoliciesForTarget(ctx, conn, targetID, filter)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Policies for target (%s): %s", targetID, err)
	}

	d.SetId(targetID)
	d.Set("filter", filter)

	if err := d.Set("policies", FlattenOrganizationalPolicies(policies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting policies: %s", err)
	}

	return diags
}

func findPoliciesForTarget(ctx context.Context, conn *organizations.Organizations, id string, filter string) ([]*organizations.PolicySummary, error) {
	input := &organizations.ListPoliciesForTargetInput{
		TargetId: aws.String(id),
		Filter:   aws.String(filter),
	}
	var output []*organizations.PolicySummary

	err := conn.ListPoliciesForTargetPagesWithContext(ctx, input, func(page *organizations.ListPoliciesForTargetOutput, lastPage bool) bool {
		output = append(output, page.Policies...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FlattenOrganizationalPolicies(policies []*organizations.PolicySummary) []map[string]interface{} {
	if len(policies) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, policy := range policies {
		result = append(result, map[string]interface{}{
			"arn":         aws.StringValue(policy.Arn),
			"aws_managed": aws.BoolValue(policy.AwsManaged),
			"description": aws.StringValue(policy.Description),
			"id":          aws.StringValue(policy.Id),
			"name":        aws.StringValue(policy.Name),
			"type":        aws.StringValue(policy.Type),
		})
	}
	return result
}
