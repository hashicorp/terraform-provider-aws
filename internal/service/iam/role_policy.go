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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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

const (
	rolePolicyNameMaxLen       = 128
	rolePolicyNamePrefixMaxLen = rolePolicyNameMaxLen - id.UniqueIDSuffixLength
)

// @SDKResource("aws_iam_role_policy", name="Role Policy")
func resourceRolePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRolePolicyPut,
		ReadWithoutTimeout:   resourceRolePolicyRead,
		UpdateWithoutTimeout: resourceRolePolicyPut,
		DeleteWithoutTimeout: resourceRolePolicyDelete,

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
				ValidateFunc:  validRolePolicyName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validResourceName(rolePolicyNamePrefixMaxLen),
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
			names.AttrRole: {
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
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policy, err := verify.LegacyPolicyNormalize(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyName := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	roleName := d.Get(names.AttrRole).(string)
	input := &iam.PutRolePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(policyName),
		RoleName:       aws.String(roleName),
	}

	_, err = conn.PutRolePolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting IAM Role (%s) Policy (%s): %s", roleName, policyName, err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("%s:%s", roleName, policyName))

		_, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
			return FindRolePolicyByTwoPartKey(ctx, conn, roleName, policyName)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IAM Role Policy (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRolePolicyRead(ctx, d, meta)...)
}

func resourceRolePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	roleName, policyName, err := RolePolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyDocument, err := FindRolePolicyByTwoPartKey(ctx, conn, roleName, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Role Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy (%s): %s", d.Id(), err)
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
	d.Set(names.AttrRole, roleName)

	return diags
}

func resourceRolePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	roleName, policyName, err := RolePolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IAM Role Policy: %s", d.Id())
	_, err = conn.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
		PolicyName: aws.String(policyName),
		RoleName:   aws.String(roleName),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Role Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindRolePolicyByTwoPartKey(ctx context.Context, conn *iam.Client, roleName, policyName string) (string, error) {
	input := &iam.GetRolePolicyInput{
		PolicyName: aws.String(policyName),
		RoleName:   aws.String(roleName),
	}

	output, err := conn.GetRolePolicy(ctx, input)

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
