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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_iam_group_policy")
func ResourceGroupPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupPolicyPut,
		ReadWithoutTimeout:   resourceGroupPolicyRead,
		UpdateWithoutTimeout: resourceGroupPolicyPut,
		DeleteWithoutTimeout: resourceGroupPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
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
		},
	}
}

func resourceGroupPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	policyDoc, err := verify.LegacyPolicyNormalize(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	groupName := d.Get("group").(string)
	policyName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	request := &iam.PutGroupPolicyInput{
		GroupName:      aws.String(groupName),
		PolicyDocument: aws.String(policyDoc),
		PolicyName:     aws.String(policyName),
	}

	_, err = conn.PutGroupPolicyWithContext(ctx, request)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting IAM Group (%s) Policy (%s): %s", groupName, policyName, err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("%s:%s", groupName, policyName))
	}

	return append(diags, resourceGroupPolicyRead(ctx, d, meta)...)
}

func resourceGroupPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	group, name, err := GroupPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group Policy (%s): %s", d.Id(), err)
	}

	request := &iam.GetGroupPolicyInput{
		PolicyName: aws.String(name),
		GroupName:  aws.String(group),
	}

	var getResp *iam.GetGroupPolicyOutput

	err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		getResp, err = conn.GetGroupPolicyWithContext(ctx, request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getResp, err = conn.GetGroupPolicyWithContext(ctx, request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Group Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group Policy (%s): %s", d.Id(), err)
	}

	if getResp == nil || getResp.PolicyDocument == nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group Policy (%s): empty response", d.Id())
	}

	policy, err := url.QueryUnescape(*getResp.PolicyDocument)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.LegacyPolicyToSet(d.Get("policy").(string), policy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group Policy (%s): setting policy: %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	if err := d.Set("name", name); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting name: %s", err)
	}

	if err := d.Set("group", group); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting group: %s", err)
	}

	return diags
}

func resourceGroupPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	group, name, err := GroupPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Group Policy (%s): %s", d.Id(), err)
	}

	request := &iam.DeleteGroupPolicyInput{
		PolicyName: aws.String(name),
		GroupName:  aws.String(group),
	}

	if _, err := conn.DeleteGroupPolicyWithContext(ctx, request); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting IAM Group Policy (%s): %s", d.Id(), err)
	}
	return diags
}

func GroupPolicyParseID(id string) (groupName, policyName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = fmt.Errorf("group_policy id must be of the form <group name>:<policy name>")
		return
	}

	groupName = parts[0]
	policyName = parts[1]
	return
}
