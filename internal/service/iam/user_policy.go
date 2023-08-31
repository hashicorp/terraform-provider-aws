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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_iam_user_policy")
func ResourceUserPolicy() *schema.Resource {
	return &schema.Resource{
		// PutUserPolicy API is idempotent, so these can be the same.
		CreateWithoutTimeout: resourceUserPolicyPut,
		ReadWithoutTimeout:   resourceUserPolicyRead,
		UpdateWithoutTimeout: resourceUserPolicyPut,
		DeleteWithoutTimeout: resourceUserPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	p, err := verify.LegacyPolicyNormalize(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", p, err)
	}

	request := &iam.PutUserPolicyInput{
		UserName:       aws.String(d.Get("user").(string)),
		PolicyDocument: aws.String(p),
	}

	var policyName string
	if !d.IsNewResource() {
		_, policyName, err = UserPolicyParseID(d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting IAM User Policy %s: %s", d.Id(), err)
		}
	} else if v, ok := d.GetOk("name"); ok {
		policyName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		policyName = id.PrefixedUniqueId(v.(string))
	} else {
		policyName = id.UniqueId()
	}
	request.PolicyName = aws.String(policyName)

	if _, err := conn.PutUserPolicyWithContext(ctx, request); err != nil {
		return sdkdiag.AppendErrorf(diags, "putting IAM User Policy %s: %s", aws.StringValue(request.PolicyName), err)
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(request.UserName), aws.StringValue(request.PolicyName)))
	return diags
}

func resourceUserPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	user, name, err := UserPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Policy (%s): %s", d.Id(), err)
	}

	request := &iam.GetUserPolicyInput{
		PolicyName: aws.String(name),
		UserName:   aws.String(user),
	}

	var getResp *iam.GetUserPolicyOutput

	err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		getResp, err = conn.GetUserPolicyWithContext(ctx, request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getResp, err = conn.GetUserPolicyWithContext(ctx, request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Policy (%s): %s", d.Id(), err)
	}

	if getResp == nil || getResp.PolicyDocument == nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Policy (%s): empty response", d.Id())
	}

	policy, err := url.QueryUnescape(*getResp.PolicyDocument)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.LegacyPolicyToSet(d.Get("policy").(string), policy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Policy (%s): setting policy: %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	d.Set("name", name)
	d.Set("user", user)

	return diags
}

func resourceUserPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	user, name, err := UserPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User Policy %s: %s", d.Id(), err)
	}

	request := &iam.DeleteUserPolicyInput{
		PolicyName: aws.String(name),
		UserName:   aws.String(user),
	}

	if _, err := conn.DeleteUserPolicyWithContext(ctx, request); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting IAM User Policy %s: %s", d.Id(), err)
	}
	return diags
}

func UserPolicyParseID(id string) (userName, policyName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("user_policy id must be of the form <user name>:<policy name>")
		return
	}

	userName = parts[0]
	policyName = parts[1]
	return
}
