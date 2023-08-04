// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_codecommit_approval_rule_template_association")
func ResourceApprovalRuleTemplateAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApprovalRuleTemplateAssociationCreate,
		ReadWithoutTimeout:   resourceApprovalRuleTemplateAssociationRead,
		DeleteWithoutTimeout: resourceApprovalRuleTemplateAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"approval_rule_template_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`[\w\.-]+`), ""),
				),
			},
		},
	}
}

func resourceApprovalRuleTemplateAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	approvalRuleTemplateName := d.Get("approval_rule_template_name").(string)
	repositoryName := d.Get("repository_name").(string)

	input := &codecommit.AssociateApprovalRuleTemplateWithRepositoryInput{
		ApprovalRuleTemplateName: aws.String(approvalRuleTemplateName),
		RepositoryName:           aws.String(repositoryName),
	}

	_, err := conn.AssociateApprovalRuleTemplateWithRepositoryWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating CodeCommit Approval Rule Template (%s) with repository (%s): %s", approvalRuleTemplateName, repositoryName, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", approvalRuleTemplateName, repositoryName))

	return append(diags, resourceApprovalRuleTemplateAssociationRead(ctx, d, meta)...)
}

func resourceApprovalRuleTemplateAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	approvalRuleTemplateName, repositoryName, err := ApprovalRuleTemplateAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Approval Rule Template Association (%s): %s", d.Id(), err)
	}

	err = FindApprovalRuleTemplateAssociation(ctx, conn, approvalRuleTemplateName, repositoryName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeCommit Approval Rule Template Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Approval Rule Template Association (%s): %s", d.Id(), err)
	}

	d.Set("approval_rule_template_name", approvalRuleTemplateName)
	d.Set("repository_name", repositoryName)

	return diags
}

func resourceApprovalRuleTemplateAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitConn(ctx)

	approvalRuleTemplateName, repositoryName, err := ApprovalRuleTemplateAssociationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeCommit Approval Rule Template (%s) from repository (%s): %s", approvalRuleTemplateName, repositoryName, err)
	}

	input := &codecommit.DisassociateApprovalRuleTemplateFromRepositoryInput{
		ApprovalRuleTemplateName: aws.String(approvalRuleTemplateName),
		RepositoryName:           aws.String(repositoryName),
	}

	_, err = conn.DisassociateApprovalRuleTemplateFromRepositoryWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codecommit.ErrCodeApprovalRuleTemplateDoesNotExistException, codecommit.ErrCodeRepositoryDoesNotExistException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeCommit Approval Rule Template (%s) from repository (%s): %s", approvalRuleTemplateName, repositoryName, err)
	}

	return diags
}

func ApprovalRuleTemplateAssociationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected APPROVAL_RULE_TEMPLATE_NAME,REPOSITORY_NAME", id)
	}

	return parts[0], parts[1], nil
}
