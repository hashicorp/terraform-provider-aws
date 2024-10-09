// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_policy", name="Policy")
// @Tags
func dataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePolicyRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{names.AttrName, "path_prefix"},
			},
			"attachment_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrARN},
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrARN},
			},
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	arn := d.Get(names.AttrARN).(string)
	name := d.Get(names.AttrName).(string)
	pathPrefix := d.Get("path_prefix").(string)

	if arn == "" {
		outputRaw, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout,
			func() (interface{}, error) {
				return findPolicyByTwoPartKey(ctx, conn, name, pathPrefix)
			},
		)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("IAM Policy", err))
		}

		arn = aws.ToString((outputRaw.(*awstypes.Policy)).Arn)
	}

	// We need to make a call to `iam.GetPolicy` because `iam.ListPolicies` doesn't return all values
	policy, err := findPolicyByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s): %s", arn, err)
	}

	arn = aws.ToString(policy.Arn)

	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set("attachment_count", policy.AttachmentCount)
	d.Set(names.AttrDescription, policy.Description)
	d.Set(names.AttrName, policy.PolicyName)
	d.Set(names.AttrPath, policy.Path)
	d.Set("policy_id", policy.PolicyId)

	setTagsOut(ctx, policy.Tags)

	outputRaw, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout,
		func() (interface{}, error) {
			return findPolicyVersion(ctx, conn, arn, aws.ToString(policy.DefaultVersionId))
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s) default version: %s", arn, err)
	}

	policyDocument, err := url.QueryUnescape(aws.ToString(outputRaw.(*awstypes.PolicyVersion).Document))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing IAM Policy (%s) document: %s", arn, err)
	}

	d.Set(names.AttrPolicy, policyDocument)

	return diags
}
