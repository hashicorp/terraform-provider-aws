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

// @SDKResource("aws_iam_user_policy_attachment", name="User Policy Attachment")
func resourceUserPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserPolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceUserPolicyAttachmentRead,
		DeleteWithoutTimeout: resourceUserPolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceUserPolicyAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"user": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceUserPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	user := d.Get("user").(string)
	policyARN := d.Get("policy_arn").(string)

	if err := attachPolicyToUser(ctx, conn, user, policyARN); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.PrefixedUniqueId(fmt.Sprintf("%s-", user)))

	return append(diags, resourceUserPolicyAttachmentRead(ctx, d, meta)...)
}

func resourceUserPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	user := d.Get("user").(string)
	policyARN := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive.
	id := fmt.Sprintf("%s:%s", user, policyARN)

	_, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findAttachedUserPolicyByTwoPartKey(ctx, conn, user, policyARN)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM User Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Policy Attachment (%s): %s", id, err)
	}

	return diags
}

func resourceUserPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	if err := detachPolicyFromUser(ctx, conn, d.Get("user").(string), d.Get("policy_arn").(string)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceUserPolicyAttachmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <user-name>/<policy_arn>", d.Id())
	}

	userName := idParts[0]
	policyARN := idParts[1]

	d.Set("user", userName)
	d.Set("policy_arn", policyARN)
	d.SetId(fmt.Sprintf("%s-%s", userName, policyARN))

	return []*schema.ResourceData{d}, nil
}

func attachPolicyToUser(ctx context.Context, conn *iam.IAM, user, policyARN string) error {
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.AttachUserPolicyWithContext(ctx, &iam.AttachUserPolicyInput{
			PolicyArn: aws.String(policyARN),
			UserName:  aws.String(user),
		})
	}, iam.ErrCodeConcurrentModificationException)

	if err != nil {
		return fmt.Errorf("attaching IAM Policy (%s) to IAM User (%s): %w", policyARN, user, err)
	}

	return nil
}

func detachPolicyFromUser(ctx context.Context, conn *iam.IAM, user, policyARN string) error {
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.DetachUserPolicyWithContext(ctx, &iam.DetachUserPolicyInput{
			PolicyArn: aws.String(policyARN),
			UserName:  aws.String(user),
		})
	}, iam.ErrCodeConcurrentModificationException)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching IAM Policy (%s) from IAM User (%s): %w", policyARN, user, err)
	}

	return nil
}

func findAttachedUserPolicyByTwoPartKey(ctx context.Context, conn *iam.IAM, userName, policyARN string) (*iam.AttachedPolicy, error) {
	input := &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(userName),
	}

	return findAttachedUserPolicy(ctx, conn, input, func(v *iam.AttachedPolicy) bool {
		return aws.StringValue(v.PolicyArn) == policyARN
	})
}

func findAttachedUserPolicy(ctx context.Context, conn *iam.IAM, input *iam.ListAttachedUserPoliciesInput, filter tfslices.Predicate[*iam.AttachedPolicy]) (*iam.AttachedPolicy, error) {
	output, err := findAttachedUserPolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findAttachedUserPolicies(ctx context.Context, conn *iam.IAM, input *iam.ListAttachedUserPoliciesInput, filter tfslices.Predicate[*iam.AttachedPolicy]) ([]*iam.AttachedPolicy, error) {
	var output []*iam.AttachedPolicy

	err := conn.ListAttachedUserPoliciesPagesWithContext(ctx, input, func(page *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool {
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
