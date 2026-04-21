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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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

const (
	rolePolicyNameMaxLen       = 128
	rolePolicyNamePrefixMaxLen = rolePolicyNameMaxLen - sdkid.UniqueIDSuffixLength
)

// @SDKResource("aws_iam_role_policy", name="Role Policy")
// @IdentityAttribute("role")
// @IdentityAttribute("name")
// @IdAttrFormat("{role}:{name}")
// @ImportIDHandler("rolePolicyImportID")
// @Testing(existsType="string")
// @Testing(preIdentityVersion="6.0.0")
func resourceRolePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRolePolicyPut,
		ReadWithoutTimeout:   resourceRolePolicyRead,
		UpdateWithoutTimeout: resourceRolePolicyPut,
		DeleteWithoutTimeout: resourceRolePolicyDelete,

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
				StateFunc: func(v any) string {
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

func resourceRolePolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policy, err := verify.LegacyPolicyNormalize(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyName := create.Name(ctx, d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
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
		d.SetId(createRolePolicyImportID(roleName, policyName))

		_, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return findRolePolicyByTwoPartKey(ctx, conn, roleName, policyName)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IAM Role Policy (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRolePolicyRead(ctx, d, meta)...)
}

func resourceRolePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	roleName, policyName, err := rolePolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyDocument, err := findRolePolicyByTwoPartKey(ctx, conn, roleName, policyName)

	if !d.IsNewResource() && retry.NotFound(err) {
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

func resourceRolePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	roleName, policyName, err := rolePolicyParseID(d.Id())
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

func findRolePolicyByTwoPartKey(ctx context.Context, conn *iam.Client, roleName, policyName string) (string, error) {
	input := &iam.GetRolePolicyInput{
		PolicyName: aws.String(policyName),
		RoleName:   aws.String(roleName),
	}

	output, err := conn.GetRolePolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return "", &retry.NotFoundError{
			LastError: err,
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

func createRolePolicyImportID(roleName, policyName string) string {
	return fmt.Sprintf("%s:%s", roleName, policyName)
}

func rolePolicyParseID(id string) (roleName, policyName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("role_policy id must be of the form <role name>:<policy name>")
		return
	}

	roleName = parts[0]
	policyName = parts[1]
	return
}

type rolePolicyImportID struct{}

func (rolePolicyImportID) Create(d *schema.ResourceData) string {
	return createRolePolicyImportID(d.Get(names.AttrRole).(string), d.Get(names.AttrName).(string))
}

func (rolePolicyImportID) Parse(id string) (string, map[string]any, error) {
	roleName, policyName, err := rolePolicyParseID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		names.AttrRole: roleName,
		names.AttrName: policyName,
	}
	return id, result, nil
}
