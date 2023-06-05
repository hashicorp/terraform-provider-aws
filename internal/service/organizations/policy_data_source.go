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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	policyID := d.Get("policy_id").(string)
	input := &organizations.DescribePolicyInput{
		PolicyId: aws.String(policyID),
	}

	output, err := conn.DescribePolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Policy (%s): %s", policyID, err)
	}

	d.SetId(aws.StringValue(output.Policy.PolicySummary.Id))
	d.Set("arn", output.Policy.PolicySummary.Arn)
	d.Set("aws_managed", output.Policy.PolicySummary.AwsManaged)
	d.Set("content", output.Policy.Content)
	d.Set("description", output.Policy.PolicySummary.Description)
	d.Set("name", output.Policy.PolicySummary.Name)
	d.Set("type", output.Policy.PolicySummary.Type)

	return diags
}
