package organizations

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_organizations_policy_attachments")
func DataSourcePolicyAttachments() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePolicyAttachmentsRead,

		Schema: map[string]*schema.Schema{
			"target_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(organizations.PolicyType_Values(), false),
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

func dataSourcePolicyAttachmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	target_id := d.Get("target_id").(string)
	filter := d.Get("filter").(string)

	params := &organizations.ListPoliciesForTargetInput{
		TargetId: aws.String(target_id),
		Filter:   aws.String(filter),
	}

	var policies []*organizations.PolicySummary

	err := conn.ListPoliciesForTargetPagesWithContext(ctx, params,
		func(page *organizations.ListPoliciesForTargetOutput, lastPage bool) bool {
			policies = append(policies, page.Policies...)

			return !lastPage
		})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing Policy Attachments for target (%s): %w", target_id, err))
	}

	if err := d.Set("policies", flattenPolicySummaries(policies)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting policies: %w", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", target_id, filter))

	return nil
}

func flattenPolicySummaries(summaries []*organizations.PolicySummary) []map[string]interface{} {
	if len(summaries) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, summary := range summaries {
		result = append(result, map[string]interface{}{
			"arn":         aws.StringValue(summary.Arn),
			"aws_managed": aws.BoolValue(summary.AwsManaged),
			"description": aws.StringValue(summary.Description),
			"id":          aws.StringValue(summary.Id),
			"name":        aws.StringValue(summary.Name),
			"type":        aws.StringValue(summary.Type),
		})
	}
	return result
}
