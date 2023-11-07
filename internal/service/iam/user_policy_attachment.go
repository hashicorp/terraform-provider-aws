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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_iam_user_policy_attachment")
func ResourceUserPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserPolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceUserPolicyAttachmentRead,
		DeleteWithoutTimeout: resourceUserPolicyAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserPolicyAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"policy_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	user := d.Get("user").(string)
	arn := d.Get("policy_arn").(string)

	err := attachPolicyToUser(ctx, conn, user, arn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching policy %s to IAM User %s: %v", arn, user, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.PrefixedUniqueId(fmt.Sprintf("%s-", user)))

	return append(diags, resourceUserPolicyAttachmentRead(ctx, d, meta)...)
}

func resourceUserPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)
	user := d.Get("user").(string)
	arn := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s:%s", user, arn)

	var attachedPolicy *iam.AttachedPolicy

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		attachedPolicy, err = FindUserAttachedPolicy(ctx, conn, user, arn)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if d.IsNewResource() && attachedPolicy == nil {
			return retry.RetryableError(&retry.NotFoundError{
				LastError: fmt.Errorf("IAM User Managed Policy Attachment (%s) not found", id),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		attachedPolicy, err = FindUserAttachedPolicy(ctx, conn, user, arn)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Managed Policy Attachment (%s): %s", id, err)
	}

	if attachedPolicy == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading IAM User Managed Policy Attachment (%s): not found after creation", id)
		}

		log.Printf("[WARN] IAM User Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	return diags
}

func resourceUserPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)
	user := d.Get("user").(string)
	arn := d.Get("policy_arn").(string)

	err := DetachPolicyFromUser(ctx, conn, user, arn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "removing policy %s from IAM User %s: %v", arn, user, err)
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

func attachPolicyToUser(ctx context.Context, conn *iam.IAM, user string, arn string) error {
	_, err := conn.AttachUserPolicyWithContext(ctx, &iam.AttachUserPolicyInput{
		UserName:  aws.String(user),
		PolicyArn: aws.String(arn),
	})
	return err
}

func DetachPolicyFromUser(ctx context.Context, conn *iam.IAM, user string, arn string) error {
	_, err := conn.DetachUserPolicyWithContext(ctx, &iam.DetachUserPolicyInput{
		UserName:  aws.String(user),
		PolicyArn: aws.String(arn),
	})
	return err
}
