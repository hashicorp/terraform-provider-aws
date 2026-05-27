// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"

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
// @Testing(tagsIdentifierAttribute="arn", tagsResourceType="Policy")
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
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{names.AttrARN},
				ValidateDiagFunc: validPolicyPath,
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

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	arn := d.Get(names.AttrARN).(string)
	name := d.Get(names.AttrName).(string)
	pathPrefix := d.Get("path_prefix").(string)

	if arn == "" {
		output, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout,
			func(ctx context.Context) (*awstypes.Policy, error) {
				return findPolicyByTwoPartKey(ctx, conn, name, pathPrefix)
			},
		)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("IAM Policy", err))
		}

		arn = aws.ToString(output.Arn)
	}

	// We need to make a call to `iam.GetPolicy` because `iam.ListPolicies` doesn't return all values
	policy, err := findPolicyByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s): %s", arn, err)
	}

	arn = aws.ToString(policy.Arn)

	d.SetId(arn)
	resourcePolicyFlatten(ctx, policy, d)

	policyVersion, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout,
		func(ctx context.Context) (*awstypes.PolicyVersion, error) {
			return findPolicyVersionByTwoPartKey(ctx, conn, arn, aws.ToString(policy.DefaultVersionId))
		},
	)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s) default version: %s", arn, err)
	}

	if err := resourcePolicyFlattenPolicyDocument(aws.ToString(policyVersion.Document), d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}
