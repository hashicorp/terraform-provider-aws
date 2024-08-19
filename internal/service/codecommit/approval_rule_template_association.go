// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codecommit"
	"github.com/aws/aws-sdk-go-v2/service/codecommit/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codecommit_approval_rule_template_association", name="Approval Rule Template Association")
func resourceApprovalRuleTemplateAssociation() *schema.Resource {
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
			names.AttrRepositoryName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`[\w\.-]+`), ""),
				),
			},
		},
	}
}

func resourceApprovalRuleTemplateAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	approvalRuleTemplateName := d.Get("approval_rule_template_name").(string)
	repositoryName := d.Get(names.AttrRepositoryName).(string)
	id := approvalRuleTemplateAssociationCreateResourceID(approvalRuleTemplateName, repositoryName)
	input := &codecommit.AssociateApprovalRuleTemplateWithRepositoryInput{
		ApprovalRuleTemplateName: aws.String(approvalRuleTemplateName),
		RepositoryName:           aws.String(repositoryName),
	}

	_, err := conn.AssociateApprovalRuleTemplateWithRepository(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeCommit Approval Rule Template Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceApprovalRuleTemplateAssociationRead(ctx, d, meta)...)
}

func resourceApprovalRuleTemplateAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	approvalRuleTemplateName, repositoryName, err := approvalRuleTemplateAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findApprovalRuleTemplateAssociationByTwoPartKey(ctx, conn, approvalRuleTemplateName, repositoryName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeCommit Approval Rule Template Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeCommit Approval Rule Template Association (%s): %s", d.Id(), err)
	}

	d.Set("approval_rule_template_name", approvalRuleTemplateName)
	d.Set(names.AttrRepositoryName, repositoryName)

	return diags
}

func resourceApprovalRuleTemplateAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCommitClient(ctx)

	approvalRuleTemplateName, repositoryName, err := approvalRuleTemplateAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting CodeCommit Approval Rule Template Association: %s", d.Id())
	_, err = conn.DisassociateApprovalRuleTemplateFromRepository(ctx, &codecommit.DisassociateApprovalRuleTemplateFromRepositoryInput{
		ApprovalRuleTemplateName: aws.String(approvalRuleTemplateName),
		RepositoryName:           aws.String(repositoryName),
	})

	if errs.IsA[*types.ApprovalRuleTemplateDoesNotExistException](err) || errs.IsA[*types.RepositoryDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeCommit Approval Rule Template Association (%s): %s", d.Id(), err)
	}

	return diags
}

const approvalRuleTemplateAssociationResourceIDSeparator = ","

func approvalRuleTemplateAssociationCreateResourceID(approvalRuleTemplateName, repositoryName string) string {
	parts := []string{approvalRuleTemplateName, repositoryName}
	id := strings.Join(parts, approvalRuleTemplateAssociationResourceIDSeparator)

	return id
}

func approvalRuleTemplateAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, approvalRuleTemplateAssociationResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected APPROVAL_RULE_TEMPLATE_NAME%[2]sREPOSITORY_NAME", id, approvalRuleTemplateAssociationResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findApprovalRuleTemplateAssociationByTwoPartKey(ctx context.Context, conn *codecommit.Client, approvalRuleTemplateName, repositoryName string) (*string, error) {
	input := &codecommit.ListRepositoriesForApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(approvalRuleTemplateName),
	}

	output, err := findApprovalRuleTemplateRepositories(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	output = tfslices.Filter(output, func(v string) bool {
		return v == repositoryName
	})

	return tfresource.AssertSingleValueResult(output)
}

func findApprovalRuleTemplateRepositories(ctx context.Context, conn *codecommit.Client, input *codecommit.ListRepositoriesForApprovalRuleTemplateInput) ([]string, error) {
	var output []string

	pages := codecommit.NewListRepositoriesForApprovalRuleTemplatePaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.ApprovalRuleTemplateDoesNotExistException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.RepositoryNames...)
	}

	return output, nil
}
