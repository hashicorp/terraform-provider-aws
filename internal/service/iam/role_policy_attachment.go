// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRolePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	role := d.Get("role").(string)
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
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	role := d.Get("role").(string)
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
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	if err := detachPolicyFromRole(ctx, conn, d.Get("role").(string), d.Get("policy_arn").(string)); err != nil {
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

	d.Set("role", roleName)
	d.Set("policy_arn", policyARN)
	d.SetId(fmt.Sprintf("%s-%s", roleName, policyARN))

	return []*schema.ResourceData{d}, nil
}

func attachPolicyToRole(ctx context.Context, conn *iam.IAM, role, policyARN string) error {
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.AttachRolePolicyWithContext(ctx, &iam.AttachRolePolicyInput{
			PolicyArn: aws.String(policyARN),
			RoleName:  aws.String(role),
		})
	}, iam.ErrCodeConcurrentModificationException)

	if err != nil {
		return fmt.Errorf("attaching IAM Policy (%s) to IAM Role (%s): %w", policyARN, role, err)
	}

	return nil
}

func detachPolicyFromRole(ctx context.Context, conn *iam.IAM, role, policyARN string) error {
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.DetachRolePolicyWithContext(ctx, &iam.DetachRolePolicyInput{
			PolicyArn: aws.String(policyARN),
			RoleName:  aws.String(role),
		})
	}, iam.ErrCodeConcurrentModificationException)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching IAM Policy (%s) from IAM Role (%s): %w", policyARN, role, err)
	}

	return nil
}

func findAttachedRolePolicyByTwoPartKey(ctx context.Context, conn *iam.IAM, roleName, policyARN string) (*iam.AttachedPolicy, error) {
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	return findAttachedRolePolicy(ctx, conn, input, func(v *iam.AttachedPolicy) bool {
		return aws.StringValue(v.PolicyArn) == policyARN
	})
}

func findAttachedRolePolicy(ctx context.Context, conn *iam.IAM, input *iam.ListAttachedRolePoliciesInput, filter tfslices.Predicate[*iam.AttachedPolicy]) (*iam.AttachedPolicy, error) {
	output, err := findAttachedRolePolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findAttachedRolePolicies(ctx context.Context, conn *iam.IAM, input *iam.ListAttachedRolePoliciesInput, filter tfslices.Predicate[*iam.AttachedPolicy]) ([]*iam.AttachedPolicy, error) {
	var output []*iam.AttachedPolicy

	err := conn.ListAttachedRolePoliciesPagesWithContext(ctx, input, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AttachedPolicies {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
