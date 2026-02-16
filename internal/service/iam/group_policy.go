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

// @SDKResource("aws_iam_group_policy", name="Group Policy")
func resourceGroupPolicy() *schema.Resource {
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
		},
	}
}

func resourceGroupPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policyDoc, err := verify.LegacyPolicyNormalize(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	groupName, policyName := d.Get("group").(string), create.Name(ctx, d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := iam.PutGroupPolicyInput{
		GroupName:      aws.String(groupName),
		PolicyDocument: aws.String(policyDoc),
		PolicyName:     aws.String(policyName),
	}

	_, err = conn.PutGroupPolicy(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting IAM Group (%s) Policy (%s): %s", groupName, policyName, err)
	}

	if d.IsNewResource() {
		d.SetId(groupPolicyCreateResourceID(groupName, policyName))

		_, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return findGroupPolicyByTwoPartKey(ctx, conn, groupName, policyName)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IAM Group Policy (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupPolicyRead(ctx, d, meta)...)
}

func resourceGroupPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	groupName, policyName, err := groupPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyDocument, err := findGroupPolicyByTwoPartKey(ctx, conn, groupName, policyName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IAM Group Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group Policy (%s): %s", d.Id(), err)
	}

	policy, err := url.QueryUnescape(policyDocument)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err := verify.LegacyPolicyToSet(d.Get(names.AttrPolicy).(string), policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("group", groupName)
	d.Set(names.AttrName, policyName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(policyName))
	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceGroupPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	groupName, policyName, err := groupPolicyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IAM Group Policy: %s", d.Id())
	input := iam.DeleteGroupPolicyInput{
		GroupName:  aws.String(groupName),
		PolicyName: aws.String(policyName),
	}
	_, err = conn.DeleteGroupPolicy(ctx, &input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Group Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findGroupPolicyByTwoPartKey(ctx context.Context, conn *iam.Client, groupName, policyName string) (string, error) {
	input := iam.GetGroupPolicyInput{
		GroupName:  aws.String(groupName),
		PolicyName: aws.String(policyName),
	}

	return findGroupPolicy(ctx, conn, &input)
}

func findGroupPolicy(ctx context.Context, conn *iam.Client, input *iam.GetGroupPolicyInput) (string, error) {
	output, err := conn.GetGroupPolicy(ctx, input)

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

const groupPolicyResourceIDSeparator = ":"

func groupPolicyCreateResourceID(groupName, policyName string) string {
	parts := []string{groupName, policyName}
	id := strings.Join(parts, groupPolicyResourceIDSeparator)

	return id
}

func groupPolicyParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, groupPolicyResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected GROUP-NAME%[2]sPOLICY-NAME", id, groupPolicyResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}
