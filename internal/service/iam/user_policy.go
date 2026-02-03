// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
				StateFunc: func(v any) string {
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

func resourceUserPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policyDoc, err := verify.LegacyPolicyNormalize(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	userName, policyName := d.Get("user").(string), create.Name(ctx, d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := iam.PutUserPolicyInput{
		PolicyDocument: aws.String(policyDoc),
		PolicyName:     aws.String(policyName),
		UserName:       aws.String(userName),
	}

	_, err = conn.PutUserPolicy(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting IAM User (%s) Policy (%s): %s", userName, policyName, err)
	}

	if d.IsNewResource() {
		d.SetId(userPolicyCreateResourceID(userName, policyName))

		_, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return findUserPolicyByTwoPartKey(ctx, conn, userName, policyName)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IAM User Policy (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserPolicyRead(ctx, d, meta)...)
}

func resourceUserPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	userName, policyName, err := userPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyDocument, err := findUserPolicyByTwoPartKey(ctx, conn, userName, policyName)

	if !d.IsNewResource() && retry.NotFound(err) {
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

func resourceUserPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	userName, policyName, err := userPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IAM User Policy: %s", d.Id())
	input := iam.DeleteUserPolicyInput{
		PolicyName: aws.String(policyName),
		UserName:   aws.String(userName),
	}
	_, err = conn.DeleteUserPolicy(ctx, &input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findUserPolicyByTwoPartKey(ctx context.Context, conn *iam.Client, userName, policyName string) (string, error) {
	input := iam.GetUserPolicyInput{
		PolicyName: aws.String(policyName),
		UserName:   aws.String(userName),
	}

	return findUserPolicy(ctx, conn, &input)
}

func findUserPolicy(ctx context.Context, conn *iam.Client, input *iam.GetUserPolicyInput) (string, error) {
	output, err := conn.GetUserPolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return "", &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.PolicyDocument == nil {
		return "", tfresource.NewEmptyResultError()
	}

	return aws.ToString(output.PolicyDocument), nil
}

const userPolicyResourceIDSeparator = ":"

func userPolicyCreateResourceID(userName, policyName string) string {
	parts := []string{userName, policyName}
	id := strings.Join(parts, userPolicyResourceIDSeparator)

	return id
}

func userPolicyParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, userPolicyResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected USER-NAME%[2]sPOLICY-NAME", id, userPolicyResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}
