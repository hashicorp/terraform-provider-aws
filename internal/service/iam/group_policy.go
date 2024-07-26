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
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policyDoc, err := verify.LegacyPolicyNormalize(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	groupName := d.Get("group").(string)
	policyName := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	request := &iam.PutGroupPolicyInput{
		GroupName:      aws.String(groupName),
		PolicyDocument: aws.String(policyDoc),
		PolicyName:     aws.String(policyName),
	}

	_, err = conn.PutGroupPolicy(ctx, request)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting IAM Group (%s) Policy (%s): %s", groupName, policyName, err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("%s:%s", groupName, policyName))

		_, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
			return FindGroupPolicyByTwoPartKey(ctx, conn, groupName, policyName)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IAM Group Policy (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupPolicyRead(ctx, d, meta)...)
}

func resourceGroupPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	groupName, policyName, err := GroupPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyDocument, err := FindGroupPolicyByTwoPartKey(ctx, conn, groupName, policyName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
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

func resourceGroupPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	groupName, policyName, err := GroupPolicyParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IAM Group Policy: %s", d.Id())
	_, err = conn.DeleteGroupPolicy(ctx, &iam.DeleteGroupPolicyInput{
		GroupName:  aws.String(groupName),
		PolicyName: aws.String(policyName),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Group Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindGroupPolicyByTwoPartKey(ctx context.Context, conn *iam.Client, groupName, policyName string) (string, error) {
	input := &iam.GetGroupPolicyInput{
		GroupName:  aws.String(groupName),
		PolicyName: aws.String(policyName),
	}

	output, err := conn.GetGroupPolicy(ctx, input)

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
