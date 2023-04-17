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
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
}

const (
	DSNamePolicy = "Organization Policy Data Source"
)

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).OrganizationsConn()

	input := &organizations.DescribeOrganizationInput{
		PolicyId: aws.String(d.Get("policy_id").(string)),
	}
	output, err := conn.DescribeOrganizationWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Policy (%s): %s", policyID, err)
	}

	d.SetId(aws.StringValue(output.Policy.PolicySummary.Id))
	d.Set("name", aws.StringValue(output.Policy.PolicySummary.Name))
	d.Set("description", aws.StringValue(output.Policy.PolicySummary.Description))
	d.Set("type", aws.StringValue(output.Policy.PolicySummary.Type))
	d.Set("context", aws.StringValue(output.Policy.Context))
	d.Set("aws_managed", aws.StringValue(output.Policy.PolicySummary.AwsManaged))

	return diags
}

