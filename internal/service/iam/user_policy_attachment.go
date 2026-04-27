// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_iam_user_policy_attachment", name="User Policy Attachment")
// @IdentityAttribute("user")
// @IdentityAttribute("policy_arn")
// @IdAttrFormat("{user}/{policy_arn}")
// @ImportIDHandler("userPolicyAttachmentImportID")
// @Testing(preIdentityVersion="v6.41.0")
func resourceUserPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserPolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceUserPolicyAttachmentRead,
		DeleteWithoutTimeout: resourceUserPolicyAttachmentDelete,

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

func resourceUserPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	user := d.Get("user").(string)
	policyARN := d.Get("policy_arn").(string)

	if err := attachPolicyToUser(ctx, conn, user, policyARN); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(createUserPolicyAttachmentImportID(d))

	return append(diags, resourceUserPolicyAttachmentRead(ctx, d, meta)...)
}

func resourceUserPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	user := d.Get("user").(string)
	policyARN := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive.
	id := fmt.Sprintf("%s:%s", user, policyARN)

	attachedPolicy, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func(ctx context.Context) (*awstypes.AttachedPolicy, error) {
		return findAttachedUserPolicyByTwoPartKey(ctx, conn, user, policyARN)
	}, d.IsNewResource())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IAM User Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Policy Attachment (%s): %s", id, err)
	}

	resourceUserPolicyAttachmentFlatten(d, user, attachedPolicy)

	return diags
}

func resourceUserPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if err := detachPolicyFromUser(ctx, conn, d.Get("user").(string), d.Get("policy_arn").(string)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func attachPolicyToUser(ctx context.Context, conn *iam.Client, user, policyARN string) error {
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConcurrentModificationException](ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.AttachUserPolicy(ctx, &iam.AttachUserPolicyInput{
			PolicyArn: aws.String(policyARN),
			UserName:  aws.String(user),
		})
	})

	if err != nil {
		return fmt.Errorf("attaching IAM Policy (%s) to IAM User (%s): %w", policyARN, user, err)
	}

	return nil
}

func detachPolicyFromUser(ctx context.Context, conn *iam.Client, user, policyARN string) error {
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConcurrentModificationException](ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.DetachUserPolicy(ctx, &iam.DetachUserPolicyInput{
			PolicyArn: aws.String(policyARN),
			UserName:  aws.String(user),
		})
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching IAM Policy (%s) from IAM User (%s): %w", policyARN, user, err)
	}

	return nil
}

func findAttachedUserPolicyByTwoPartKey(ctx context.Context, conn *iam.Client, userName, policyARN string) (*awstypes.AttachedPolicy, error) {
	input := &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(userName),
	}

	return findAttachedUserPolicy(ctx, conn, input, func(v awstypes.AttachedPolicy) bool {
		return aws.ToString(v.PolicyArn) == policyARN
	})
}

func findAttachedUserPolicy(ctx context.Context, conn *iam.Client, input *iam.ListAttachedUserPoliciesInput, filter tfslices.Predicate[awstypes.AttachedPolicy]) (*awstypes.AttachedPolicy, error) {
	output, err := findAttachedUserPolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAttachedUserPolicies(ctx context.Context, conn *iam.Client, input *iam.ListAttachedUserPoliciesInput, filter tfslices.Predicate[awstypes.AttachedPolicy]) ([]awstypes.AttachedPolicy, error) {
	var output []awstypes.AttachedPolicy

	pages := iam.NewListAttachedUserPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
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

func resourceUserPolicyAttachmentFlatten(d *schema.ResourceData, user string, attachedPolicy *awstypes.AttachedPolicy) {
	d.Set("user", user)
	d.Set("policy_arn", attachedPolicy.PolicyArn)
}

func createUserPolicyAttachmentImportID(d *schema.ResourceData) string {
	return (userPolicyAttachmentImportID{}).Create(d)
}

type userPolicyAttachmentImportID struct{}

func (v userPolicyAttachmentImportID) Create(d *schema.ResourceData) string {
	return v.create(d.Get("user").(string), d.Get("policy_arn").(string))
}

func (userPolicyAttachmentImportID) create(userName, policyARN string) string {
	return fmt.Sprintf("%s/%s", userName, policyARN)
}

func (userPolicyAttachmentImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", nil, fmt.Errorf("unexpected format for Import ID (%q), expected <user-name>/<policy_arn>", id)
	}

	result := map[string]any{
		"user":       parts[0],
		"policy_arn": parts[1],
	}

	return id, result, nil
}
