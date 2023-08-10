// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ssoadmin_permission_set_inline_policy")
func ResourcePermissionSetInlinePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionSetInlinePolicyPut,
		ReadWithoutTimeout:   resourcePermissionSetInlinePolicyRead,
		UpdateWithoutTimeout: resourcePermissionSetInlinePolicyPut,
		DeleteWithoutTimeout: resourcePermissionSetInlinePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"inline_policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     verify.ValidIAMPolicyJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},

			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"permission_set_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourcePermissionSetInlinePolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)

	policy, err := structure.NormalizeJsonString(d.Get("inline_policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get("inline_policy").(string), err)
	}

	input := &ssoadmin.PutInlinePolicyToPermissionSetInput{
		InlinePolicy:     aws.String(policy),
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err = conn.PutInlinePolicyToPermissionSetWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Inline Policy for SSO Permission Set (%s): %s", permissionSetArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", permissionSetArn, instanceArn))

	// (Re)provision ALL accounts after making the above changes
	if err := provisionPermissionSet(ctx, conn, permissionSetArn, instanceArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "provisioning SSO Permission Set (%s): %s", permissionSetArn, err)
	}

	return append(diags, resourcePermissionSetInlinePolicyRead(ctx, d, meta)...)
}

func resourcePermissionSetInlinePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	permissionSetArn, instanceArn, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing SSO Permission Set Inline Policy ID: %s", err)
	}

	input := &ssoadmin.GetInlinePolicyForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	output, err := conn.GetInlinePolicyForPermissionSetWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Inline Policy for SSO Permission Set (%s) not found, removing from state", permissionSetArn)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inline Policy for SSO Permission Set (%s): %s", permissionSetArn, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading Inline Policy for SSO Permission Set (%s): empty output", permissionSetArn)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("inline_policy").(string), aws.StringValue(output.InlinePolicy))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inline Policy for SSO Permission Set (%s): %s", permissionSetArn, err)
	}

	d.Set("inline_policy", policyToSet)

	d.Set("instance_arn", instanceArn)
	d.Set("permission_set_arn", permissionSetArn)

	return diags
}

func resourcePermissionSetInlinePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	permissionSetArn, instanceArn, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing SSO Permission Set Inline Policy ID: %s", err)
	}

	input := &ssoadmin.DeleteInlinePolicyFromPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err = conn.DeleteInlinePolicyFromPermissionSetWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "detaching Inline Policy from SSO Permission Set (%s): %s", permissionSetArn, err)
	}

	return diags
}
