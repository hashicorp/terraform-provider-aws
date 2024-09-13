// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_user_policy", name="User Policy")
func resourceUserPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserPolicyPut,
		ReadWithoutTimeout:   resourceUserPolicyRead,
		UpdateWithoutTimeout: resourceUserPolicyPut,
		DeleteWithoutTimeout: resourceUserPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
			},
			names.AttrPolicy: {
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
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policyDoc, err := verify.LegacyPolicyNormalize(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	userName := d.Get("user").(string)
	policyName := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &iam.PutUserPolicyInput{
		PolicyDocument: aws.String(policyDoc),
		PolicyName:     aws.String(policyName),
		UserName:       aws.String(userName),
	}

	_, err = conn.PutUserPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting IAM User (%s) Policy (%s): %s", userName, policyName, err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("%s:%s", userName, policyName))

		_, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
			return FindUserPolicyByTwoPartKey(ctx, conn, userName, policyName)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IAM User Policy (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserPolicyRead(ctx, d, meta)...)
}

func resourceUserPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	userName, policyName, err := UserPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyDocument, err := FindUserPolicyByTwoPartKey(ctx, conn, userName, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM User Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Policy (%s): %s", d.Id(), err)
	}

	policy, err := url.QueryUnescape(policyDocument)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err := verify.LegacyPolicyToSet(d.Get(names.AttrPolicy).(string), policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrName, policyName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(policyName))
	d.Set(names.AttrPolicy, policyToSet)
	d.Set("user", userName)

	return diags
}

func resourceUserPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	userName, policyName, err := UserPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IAM User Policy: %s", d.Id())
	_, err = conn.DeleteUserPolicy(ctx, &iam.DeleteUserPolicyInput{
		PolicyName: aws.String(policyName),
		UserName:   aws.String(userName),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindUserPolicyByTwoPartKey(ctx context.Context, conn *iam.Client, userName, policyName string) (string, error) {
	input := &iam.GetUserPolicyInput{
		PolicyName: aws.String(policyName),
		UserName:   aws.String(userName),
	}

	output, err := conn.GetUserPolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return "", &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.PolicyDocument == nil {
		return "", tfresource.NewEmptyResultError(input)
	}

	return aws.ToString(output.PolicyDocument), nil
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
