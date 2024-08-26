// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_codecommit_approval_rule_template", name="Approval Rule Template")
func dataSourceApprovalRuleTemplate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceApprovalRuleTemplateRead,

		Schema: map[string]*schema.Schema{
			"approval_rule_template_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContent: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
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
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	templateName := d.Get(names.AttrName).(string)
	result, err := findApprovalRuleTemplateByName(ctx, conn, templateName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Approval Rule Template (%s): %s", templateName, err)
	}

	d.SetId(aws.ToString(result.ApprovalRuleTemplateName))
	d.Set("approval_rule_template_id", result.ApprovalRuleTemplateId)
	d.Set(names.AttrContent, result.ApprovalRuleTemplateContent)
	d.Set(names.AttrCreationDate, result.CreationDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, result.ApprovalRuleTemplateDescription)
	d.Set("last_modified_date", result.LastModifiedDate.Format(time.RFC3339))
	d.Set("last_modified_user", result.LastModifiedUser)
	d.Set(names.AttrName, result.ApprovalRuleTemplateName)
	d.Set("rule_content_sha256", result.RuleContentSha256)

	return diags
}
