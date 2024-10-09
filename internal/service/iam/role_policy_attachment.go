// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_role_policy_attachment", name="Role Policy Attachment")
func resourceRolePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRolePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceRolePolicyAttachmentRead,
		DeleteWithoutTimeout: resourceRolePolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRolePolicyAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrRole: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRolePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	role := d.Get(names.AttrRole).(string)
	policyARN := d.Get("policy_arn").(string)

	if err := attachPolicyToRole(ctx, conn, role, policyARN); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.PrefixedUniqueId(fmt.Sprintf("%s-", role)))

	return append(diags, resourceRolePolicyAttachmentRead(ctx, d, meta)...)
}

func resourceRolePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	role := d.Get(names.AttrRole).(string)
	policyARN := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive.
	id := fmt.Sprintf("%s:%s", role, policyARN)

	_, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findAttachedRolePolicyByTwoPartKey(ctx, conn, role, policyARN)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Role Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy Attachment (%s): %s", id, err)
	}

	return diags
}

func resourceRolePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if err := detachPolicyFromRole(ctx, conn, d.Get(names.AttrRole).(string), d.Get("policy_arn").(string)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceRolePolicyAttachmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <role-name>/<policy_arn>", d.Id())
	}

	roleName := idParts[0]
	policyARN := idParts[1]

	d.Set(names.AttrRole, roleName)
	d.Set("policy_arn", policyARN)
	d.SetId(fmt.Sprintf("%s-%s", roleName, policyARN))

	return []*schema.ResourceData{d}, nil
}

func attachPolicyToRole(ctx context.Context, conn *iam.Client, role, policyARN string) error {
	var errConcurrentModificationException *awstypes.ConcurrentModificationException
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			PolicyArn: aws.String(policyARN),
			RoleName:  aws.String(role),
		})
	}, errConcurrentModificationException.ErrorCode())

	if err != nil {
		return fmt.Errorf("attaching IAM Policy (%s) to IAM Role (%s): %w", policyARN, role, err)
	}

	return nil
}

func detachPolicyFromRole(ctx context.Context, conn *iam.Client, role, policyARN string) error {
	var errConcurrentModificationException *awstypes.ConcurrentModificationException
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
			PolicyArn: aws.String(policyARN),
			RoleName:  aws.String(role),
		})
	}, errConcurrentModificationException.ErrorCode())

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching IAM Policy (%s) from IAM Role (%s): %w", policyARN, role, err)
	}

	return nil
}

func findAttachedRolePolicyByTwoPartKey(ctx context.Context, conn *iam.Client, roleName, policyARN string) (*awstypes.AttachedPolicy, error) {
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	return findAttachedRolePolicy(ctx, conn, input, func(v awstypes.AttachedPolicy) bool {
		return aws.ToString(v.PolicyArn) == policyARN
	})
}

func findAttachedRolePolicy(ctx context.Context, conn *iam.Client, input *iam.ListAttachedRolePoliciesInput, filter tfslices.Predicate[awstypes.AttachedPolicy]) (*awstypes.AttachedPolicy, error) {
	output, err := findAttachedRolePolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAttachedRolePolicies(ctx context.Context, conn *iam.Client, input *iam.ListAttachedRolePoliciesInput, filter tfslices.Predicate[awstypes.AttachedPolicy]) ([]awstypes.AttachedPolicy, error) {
	var output []awstypes.AttachedPolicy

	pages := iam.NewListAttachedRolePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AttachedPolicies {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
