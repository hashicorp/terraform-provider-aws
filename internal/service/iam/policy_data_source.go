// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_iam_policy")
func DataSourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePolicyRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{"name", "path_prefix"},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"arn"},
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"arn"},
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)
	name := d.Get("name").(string)
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

		arn = aws.StringValue((outputRaw.(*iam.Policy)).Arn)
	}

	// We need to make a call to `iam.GetPolicy` because `iam.ListPolicies` doesn't return all values
	policy, err := FindPolicyByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s): %s", arn, err)
	}

	arn = aws.StringValue(policy.Arn)

	d.SetId(arn)
	d.Set("arn", arn)
	d.Set("description", policy.Description)
	d.Set("name", policy.PolicyName)
	d.Set("path", policy.Path)
	d.Set("policy_id", policy.PolicyId)

	if err := d.Set("tags", KeyValueTags(ctx, policy.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	outputRaw, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout,
		func() (interface{}, error) {
			return findPolicyVersion(ctx, conn, arn, aws.StringValue(policy.DefaultVersionId))
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy (%s) default version: %s", arn, err)
	}

	policyDocument, err := url.QueryUnescape(aws.StringValue(outputRaw.(*iam.PolicyVersion).Document))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing IAM Policy (%s) document: %s", arn, err)
	}

	d.Set("policy", policyDocument)

	return diags
}
