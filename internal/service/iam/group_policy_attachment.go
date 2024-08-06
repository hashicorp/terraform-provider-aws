// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"reflect"
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
)

// @SDKResource("aws_iam_group_policy_attachment", name="Group Policy Attachment")
func resourceGroupPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupPolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceGroupPolicyAttachmentRead,
		DeleteWithoutTimeout: resourceGroupPolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceGroupPolicyAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceGroupPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	group := d.Get("group").(string)
	policyARN := d.Get("policy_arn").(string)

	if err := attachPolicyToGroup(ctx, conn, group, policyARN); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.PrefixedUniqueId(fmt.Sprintf("%s-", group)))

	return append(diags, resourceGroupPolicyAttachmentRead(ctx, d, meta)...)
}

func resourceGroupPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	group := d.Get("group").(string)
	policyARN := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive.
	id := fmt.Sprintf("%s:%s", group, policyARN)

	_, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findAttachedGroupPolicyByTwoPartKey(ctx, conn, group, policyARN)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Group Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group Policy Attachment (%s): %s", id, err)
	}

	return diags
}

func resourceGroupPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if err := detachPolicyFromGroup(ctx, conn, d.Get("group").(string), d.Get("policy_arn").(string)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceGroupPolicyAttachmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <group-name>/<policy_arn>", d.Id())
	}

	groupName := idParts[0]
	policyARN := idParts[1]

	d.Set("group", groupName)
	d.Set("policy_arn", policyARN)
	d.SetId(fmt.Sprintf("%s-%s", groupName, policyARN))

	return []*schema.ResourceData{d}, nil
}

func attachPolicyToGroup(ctx context.Context, conn *iam.Client, group, policyARN string) error {
	var errConcurrentModificationException *awstypes.ConcurrentModificationException
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.AttachGroupPolicy(ctx, &iam.AttachGroupPolicyInput{
			GroupName: aws.String(group),
			PolicyArn: aws.String(policyARN),
		})
	}, errConcurrentModificationException.ErrorCode())

	if err != nil {
		return fmt.Errorf("attaching IAM Policy (%s) to IAM Group (%s): %w", policyARN, group, err)
	}

	return nil
}

func detachPolicyFromGroup(ctx context.Context, conn *iam.Client, group, policyARN string) error {
	var errConcurrentModificationException *awstypes.ConcurrentModificationException
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.DetachGroupPolicy(ctx, &iam.DetachGroupPolicyInput{
			GroupName: aws.String(group),
			PolicyArn: aws.String(policyARN),
		})
	}, errConcurrentModificationException.ErrorCode())

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching IAM Policy (%s) from IAM Group (%s): %w", policyARN, group, err)
	}

	return nil
}

func findAttachedGroupPolicyByTwoPartKey(ctx context.Context, conn *iam.Client, groupName, policyARN string) (*awstypes.AttachedPolicy, error) {
	input := &iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	return findAttachedGroupPolicy(ctx, conn, input, func(v awstypes.AttachedPolicy) bool {
		return aws.ToString(v.PolicyArn) == policyARN
	})
}

func findAttachedGroupPolicy(ctx context.Context, conn *iam.Client, input *iam.ListAttachedGroupPoliciesInput, filter tfslices.Predicate[awstypes.AttachedPolicy]) (*awstypes.AttachedPolicy, error) {
	output, err := findAttachedGroupPolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAttachedGroupPolicies(ctx context.Context, conn *iam.Client, input *iam.ListAttachedGroupPoliciesInput, filter tfslices.Predicate[awstypes.AttachedPolicy]) ([]awstypes.AttachedPolicy, error) {
	var output []awstypes.AttachedPolicy

	pages := iam.NewListAttachedGroupPoliciesPaginator(conn, input)
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
			if !reflect.ValueOf(v).IsZero() && filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
