// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	rolePolicyNameMaxLen       = 128
	rolePolicyNamePrefixMaxLen = rolePolicyNameMaxLen - id.UniqueIDSuffixLength
)

// @SDKResource("aws_iam_role_policy")
func ResourceRolePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRolePolicyPut,
		ReadWithoutTimeout:   resourceRolePolicyRead,
		UpdateWithoutTimeout: resourceRolePolicyPut,
		DeleteWithoutTimeout: resourceRolePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validRolePolicyName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validResourceName(rolePolicyNamePrefixMaxLen),
			},
			"policy": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          verify.ValidIAMPolicyJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := verify.LegacyPolicyNormalize(v)
					return json
				},
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validRolePolicyRole,
			},
		},
	}
}

func resourceRolePolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	policy, err := verify.LegacyPolicyNormalize(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	roleName := d.Get("role").(string)
	input := &iam.PutRolePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(policyName),
		RoleName:       aws.String(roleName),
	}

	_, err = conn.PutRolePolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting IAM Role (%s) Policy (%s): %s", roleName, policyName, err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("%s:%s", roleName, policyName))
	}

	return append(diags, resourceRolePolicyRead(ctx, d, meta)...)
}

func resourceRolePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	role, name, err := RolePolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): %s", d.Id(), err)
	}

	request := &iam.GetRolePolicyInput{
		PolicyName: aws.String(name),
		RoleName:   aws.String(role),
	}

	var getResp *iam.GetRolePolicyOutput

	err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		getResp, err = conn.GetRolePolicyWithContext(ctx, request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getResp, err = conn.GetRolePolicyWithContext(ctx, request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Role Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): %s", d.Id(), err)
	}

	if getResp == nil || getResp.PolicyDocument == nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): empty response", d.Id())
	}

	policy, err := url.QueryUnescape(*getResp.PolicyDocument)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.LegacyPolicyToSet(d.Get("policy").(string), policy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): setting policy: %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	d.Set("name", name)
	d.Set("role", role)

	return diags
}

func resourceRolePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	role, name, err := RolePolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM role policy (%s): %s", d.Id(), err)
	}

	request := &iam.DeleteRolePolicyInput{
		PolicyName: aws.String(name),
		RoleName:   aws.String(role),
	}

	if _, err := conn.DeleteRolePolicyWithContext(ctx, request); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting IAM role policy (%s): %s", d.Id(), err)
	}
	return diags
}

func RolePolicyParseID(id string) (roleName, policyName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("role_policy id must be of the form <role name>:<policy name>")
		return
	}

	roleName = parts[0]
	policyName = parts[1]
	return
}
