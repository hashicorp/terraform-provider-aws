package codecommit

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceApprovalRuleTemplate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceApprovalRuleTemplateRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"approval_rule_template_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_user": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_content_sha256": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceApprovalRuleTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn()

	templateName := d.Get("name").(string)
	input := &codecommit.GetApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(templateName),
	}

	output, err := conn.GetApprovalRuleTemplateWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Approval Rule Template (%s): %s", templateName, err)
	}

	if output == nil || output.ApprovalRuleTemplate == nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Approval Rule Template (%s): empty output", templateName)
	}

	result := output.ApprovalRuleTemplate

	d.SetId(aws.StringValue(result.ApprovalRuleTemplateName))
	d.Set("name", result.ApprovalRuleTemplateName)
	d.Set("approval_rule_template_id", result.ApprovalRuleTemplateId)
	d.Set("content", result.ApprovalRuleTemplateContent)
	d.Set("creation_date", result.CreationDate.Format(time.RFC3339))
	d.Set("description", result.ApprovalRuleTemplateDescription)
	d.Set("last_modified_date", result.LastModifiedDate.Format(time.RFC3339))
	d.Set("last_modified_user", result.LastModifiedUser)
	d.Set("rule_content_sha256", result.RuleContentSha256)

	return diags
}
