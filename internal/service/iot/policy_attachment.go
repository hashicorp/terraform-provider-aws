// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_policy_attachment", nmw="Policy Attachment")
func resourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourcePolicyAttachmentRead,
		DeleteWithoutTimeout: resourcePolicyAttachmentDelete,

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTarget: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	policyName := d.Get(names.AttrPolicy).(string)
	target := d.Get(names.AttrTarget).(string)
	id := policyAttachmentCreateResourceID(policyName, target)
	input := &iot.AttachPolicyInput{
		PolicyName: aws.String(policyName),
		Target:     aws.String(target),
	}

	_, err := conn.AttachPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Policy Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	policyName, target, err := policyAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findAttachedPolicyByTwoPartKey(ctx, conn, policyName, target)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Policy Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Policy Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourcePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	policyName, target, err := policyAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting IoT Policy Attachment: %s", d.Id())
	_, err = conn.DetachPolicy(ctx, &iot.DetachPolicyInput{
		PolicyName: aws.String(policyName),
		Target:     aws.String(target),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Policy Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func findAttachedPolicyByTwoPartKey(ctx context.Context, conn *iot.Client, policyName, target string) (*awstypes.Policy, error) {
	input := &iot.ListAttachedPoliciesInput{
		PageSize:  aws.Int32(250),
		Recursive: false,
		Target:    aws.String(target),
	}

	return findAttachedPolicy(ctx, conn, input)
}

func findAttachedPolicy(ctx context.Context, conn *iot.Client, input *iot.ListAttachedPoliciesInput) (*awstypes.Policy, error) {
	output, err := findAttachedPolicies(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertFirstValueResult(output)
}

func findAttachedPolicies(ctx context.Context, conn *iot.Client, input *iot.ListAttachedPoliciesInput) ([]awstypes.Policy, error) {
	var output []awstypes.Policy

	pages := iot.NewListAttachedPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Policies...)
	}

	return output, nil
}

const policyAttachmentResourceIDSeparator = "|"

func policyAttachmentCreateResourceID(policyName, target string) string {
	parts := []string{policyName, target}
	id := strings.Join(parts, policyAttachmentResourceIDSeparator)

	return id
}

func policyAttachmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, policyAttachmentResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected policy-name%[2]starget", id, policyAttachmentResourceIDSeparator)
}
