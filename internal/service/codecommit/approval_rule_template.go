// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_codecommit_approval_rule_template")
func ResourceApprovalRuleTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApprovalRuleTemplateCreate,
		ReadWithoutTimeout:   resourceApprovalRuleTemplateRead,
		UpdateWithoutTimeout: resourceApprovalRuleTemplateUpdate,
		DeleteWithoutTimeout: resourceApprovalRuleTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"content": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				ValidateFunc: validation.All(
					validation.StringIsJSON,
					validation.StringLenBetween(1, 3000),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"approval_rule_template_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
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

func resourceApprovalRuleTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	name := d.Get("name").(string)

	input := &codecommit.CreateApprovalRuleTemplateInput{
		ApprovalRuleTemplateName:    aws.String(name),
		ApprovalRuleTemplateContent: aws.String(d.Get("content").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.ApprovalRuleTemplateDescription = aws.String(v.(string))
	}

	_, err := conn.CreateApprovalRuleTemplateWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeCommit Approval Rule Template (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceApprovalRuleTemplateRead(ctx, d, meta)...)
}

func resourceApprovalRuleTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	input := &codecommit.GetApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(d.Id()),
	}

	resp, err := conn.GetApprovalRuleTemplateWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codecommit.ErrCodeApprovalRuleTemplateDoesNotExistException) {
		log.Printf("[WARN] CodeCommit Approval Rule Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Approval Rule Template (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.ApprovalRuleTemplate == nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Approval Rule Template (%s): empty output", d.Id())
	}

	result := resp.ApprovalRuleTemplate

	d.Set("name", result.ApprovalRuleTemplateName)
	d.Set("approval_rule_template_id", result.ApprovalRuleTemplateId)
	d.Set("description", result.ApprovalRuleTemplateDescription)
	d.Set("content", result.ApprovalRuleTemplateContent)
	d.Set("creation_date", result.CreationDate.Format(time.RFC3339))
	d.Set("last_modified_date", result.LastModifiedDate.Format(time.RFC3339))
	d.Set("last_modified_user", result.LastModifiedUser)
	d.Set("rule_content_sha256", result.RuleContentSha256)

	return diags
}

func resourceApprovalRuleTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	if d.HasChange("description") {
		input := &codecommit.UpdateApprovalRuleTemplateDescriptionInput{
			ApprovalRuleTemplateDescription: aws.String(d.Get("description").(string)),
			ApprovalRuleTemplateName:        aws.String(d.Id()),
		}

		_, err := conn.UpdateApprovalRuleTemplateDescriptionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeCommit Approval Rule Template (%s) description: %s", d.Id(), err)
		}
	}

	if d.HasChange("content") {
		input := &codecommit.UpdateApprovalRuleTemplateContentInput{
			ApprovalRuleTemplateName:  aws.String(d.Id()),
			ExistingRuleContentSha256: aws.String(d.Get("rule_content_sha256").(string)),
			NewRuleContent:            aws.String(d.Get("content").(string)),
		}

		_, err := conn.UpdateApprovalRuleTemplateContentWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeCommit Approval Rule Template (%s) content: %s", d.Id(), err)
		}
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)

		input := &codecommit.UpdateApprovalRuleTemplateNameInput{
			NewApprovalRuleTemplateName: aws.String(newName),
			OldApprovalRuleTemplateName: aws.String(d.Id()),
		}

		_, err := conn.UpdateApprovalRuleTemplateNameWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeCommit Approval Rule Template (%s) name: %s", d.Id(), err)
		}

		d.SetId(newName)
	}

	return append(diags, resourceApprovalRuleTemplateRead(ctx, d, meta)...)
}

func resourceApprovalRuleTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	input := &codecommit.DeleteApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(d.Id()),
	}

	_, err := conn.DeleteApprovalRuleTemplateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codecommit.ErrCodeApprovalRuleTemplateDoesNotExistException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeCommit Approval Rule Template (%s): %s", d.Id(), err)
	}

	return diags
}
