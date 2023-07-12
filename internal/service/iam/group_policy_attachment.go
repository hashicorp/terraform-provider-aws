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

// @SDKResource("aws_iam_group_policy_attachment")
func ResourceGroupPolicyAttachment() *schema.Resource {
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGroupPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	group := d.Get("group").(string)
	arn := d.Get("policy_arn").(string)

	err := attachPolicyToGroup(ctx, conn, group, arn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching policy %s to IAM group %s: %v", arn, group, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.PrefixedUniqueId(fmt.Sprintf("%s-", group)))

	return append(diags, resourceGroupPolicyAttachmentRead(ctx, d, meta)...)
}

func resourceGroupPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)
	group := d.Get("group").(string)
	arn := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s:%s", group, arn)

	var attachedPolicy *iam.AttachedPolicy

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		attachedPolicy, err = FindGroupAttachedPolicy(ctx, conn, group, arn)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if d.IsNewResource() && attachedPolicy == nil {
			return retry.RetryableError(&retry.NotFoundError{
				LastError: fmt.Errorf("IAM Group Managed Policy Attachment (%s) not found", id),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		attachedPolicy, err = FindGroupAttachedPolicy(ctx, conn, group, arn)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group Managed Policy Attachment (%s): %s", id, err)
	}

	if attachedPolicy == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading IAM User Managed Policy Attachment (%s): not found after creation", id)
		}

		log.Printf("[WARN] IAM Group Managed Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	return diags
}

func resourceGroupPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)
	group := d.Get("group").(string)
	arn := d.Get("policy_arn").(string)

	err := detachPolicyFromGroup(ctx, conn, group, arn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "removing policy %s from IAM Group %s: %v", arn, group, err)
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

func attachPolicyToGroup(ctx context.Context, conn *iam.IAM, group string, arn string) error {
	_, err := conn.AttachGroupPolicyWithContext(ctx, &iam.AttachGroupPolicyInput{
		GroupName: aws.String(group),
		PolicyArn: aws.String(arn),
	})
	return err
}

func detachPolicyFromGroup(ctx context.Context, conn *iam.IAM, group string, arn string) error {
	_, err := conn.DetachGroupPolicyWithContext(ctx, &iam.DetachGroupPolicyInput{
		GroupName: aws.String(group),
		PolicyArn: aws.String(arn),
	})
	return err
}
