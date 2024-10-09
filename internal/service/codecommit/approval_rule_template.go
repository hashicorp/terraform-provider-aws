// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codecommit"
	"github.com/aws/aws-sdk-go-v2/service/codecommit/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codecommit_approval_rule_template", name="Approval Rule Template")
func resourceApprovalRuleTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApprovalRuleTemplateCreate,
		ReadWithoutTimeout:   resourceApprovalRuleTemplateRead,
		UpdateWithoutTimeout: resourceApprovalRuleTemplateUpdate,
		DeleteWithoutTimeout: resourceApprovalRuleTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"approval_rule_template_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContent: {
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
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
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

func resourceApprovalRuleTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &codecommit.CreateApprovalRuleTemplateInput{
		ApprovalRuleTemplateName:    aws.String(name),
		ApprovalRuleTemplateContent: aws.String(d.Get(names.AttrContent).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.ApprovalRuleTemplateDescription = aws.String(v.(string))
	}

	_, err := conn.CreateApprovalRuleTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeCommit Approval Rule Template (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceApprovalRuleTemplateRead(ctx, d, meta)...)
}

func resourceApprovalRuleTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	result, err := findApprovalRuleTemplateByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeCommit Approval Rule Template %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Approval Rule Template (%s): %s", d.Id(), err)
	}

	d.Set("approval_rule_template_id", result.ApprovalRuleTemplateId)
	d.Set(names.AttrDescription, result.ApprovalRuleTemplateDescription)
	d.Set(names.AttrContent, result.ApprovalRuleTemplateContent)
	d.Set(names.AttrCreationDate, result.CreationDate.Format(time.RFC3339))
	d.Set("last_modified_date", result.LastModifiedDate.Format(time.RFC3339))
	d.Set("last_modified_user", result.LastModifiedUser)
	d.Set(names.AttrName, result.ApprovalRuleTemplateName)
	d.Set("rule_content_sha256", result.RuleContentSha256)

	return diags
}

func resourceApprovalRuleTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	if d.HasChange(names.AttrDescription) {
		input := &codecommit.UpdateApprovalRuleTemplateDescriptionInput{
			ApprovalRuleTemplateDescription: aws.String(d.Get(names.AttrDescription).(string)),
			ApprovalRuleTemplateName:        aws.String(d.Id()),
		}

		_, err := conn.UpdateApprovalRuleTemplateDescription(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeCommit Approval Rule Template (%s) description: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrContent) {
		input := &codecommit.UpdateApprovalRuleTemplateContentInput{
			ApprovalRuleTemplateName:  aws.String(d.Id()),
			ExistingRuleContentSha256: aws.String(d.Get("rule_content_sha256").(string)),
			NewRuleContent:            aws.String(d.Get(names.AttrContent).(string)),
		}

		_, err := conn.UpdateApprovalRuleTemplateContent(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeCommit Approval Rule Template (%s) content: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrName) {
		newName := d.Get(names.AttrName).(string)

		input := &codecommit.UpdateApprovalRuleTemplateNameInput{
			NewApprovalRuleTemplateName: aws.String(newName),
			OldApprovalRuleTemplateName: aws.String(d.Id()),
		}

		_, err := conn.UpdateApprovalRuleTemplateName(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeCommit Approval Rule Template (%s) name: %s", d.Id(), err)
		}

		d.SetId(newName)
	}

	return append(diags, resourceApprovalRuleTemplateRead(ctx, d, meta)...)
}

func resourceApprovalRuleTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	log.Printf("[INFO] Deleting CodeCommit Approval Rule Template: %s", d.Id())
	_, err := conn.DeleteApprovalRuleTemplate(ctx, &codecommit.DeleteApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(d.Id()),
	})

	if errs.IsA[*types.ApprovalRuleTemplateDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeCommit Approval Rule Template (%s): %s", d.Id(), err)
	}

	return diags
}

func findApprovalRuleTemplateByName(ctx context.Context, conn *codecommit.Client, name string) (*types.ApprovalRuleTemplate, error) {
	input := &codecommit.GetApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(name),
	}

	output, err := conn.GetApprovalRuleTemplate(ctx, input)

	if errs.IsA[*types.ApprovalRuleTemplateDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApprovalRuleTemplate == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApprovalRuleTemplate, nil
}
