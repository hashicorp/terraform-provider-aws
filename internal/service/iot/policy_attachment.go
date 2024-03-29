// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_iot_policy_attachment")
func ResourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourcePolicyAttachmentRead,
		DeleteWithoutTimeout: resourcePolicyAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	policyName := d.Get("policy").(string)
	target := d.Get("target").(string)
	id := policyAttachmentCreateResourceID(policyName, target)
	input := &iot.AttachPolicyInput{
		PolicyName: aws.String(policyName),
		Target:     aws.String(target),
	}

	_, err := conn.AttachPolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Policy Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	policyName, target, err := policyAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = FindAttachedPolicyByTwoPartKey(ctx, conn, policyName, target)

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
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	policyName, target, err := policyAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting IoT Policy Attachment: %s", d.Id())
	_, err = conn.DetachPolicyWithContext(ctx, &iot.DetachPolicyInput{
		PolicyName: aws.String(policyName),
		Target:     aws.String(target),
	})

	// DetachPolicy doesn't return an error if the policy doesn't exist,
	// but it returns an error if the Target is not found.
	if tfawserr.ErrMessageContains(err, iot.ErrCodeInvalidRequestException, "Invalid Target") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Policy Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func FindAttachedPolicyByTwoPartKey(ctx context.Context, conn *iot.IoT, policyName, target string) (*iot.Policy, error) {
	input := &iot.ListAttachedPoliciesInput{
		PageSize:  aws.Int64(250),
		Recursive: aws.Bool(false),
		Target:    aws.String(target),
	}

	return findAttachedPolicy(ctx, conn, input, func(v *iot.Policy) bool {
		return aws.StringValue(v.PolicyName) == policyName
	})
}

func findAttachedPolicy(ctx context.Context, conn *iot.IoT, input *iot.ListAttachedPoliciesInput, filter tfslices.Predicate[*iot.Policy]) (*iot.Policy, error) {
	output, err := findAttachedPolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findAttachedPolicies(ctx context.Context, conn *iot.IoT, input *iot.ListAttachedPoliciesInput, filter tfslices.Predicate[*iot.Policy]) ([]*iot.Policy, error) {
	var output []*iot.Policy

	err := conn.ListAttachedPoliciesPagesWithContext(ctx, input, func(page *iot.ListAttachedPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Policies {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
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
