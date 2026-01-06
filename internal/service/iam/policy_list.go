// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_iam_policy")
func newPolicyResourceAsListResource() inttypes.ListResourceForSDK {
	l := policyListResource{}
	l.SetResourceSchema(resourcePolicy())

	return &l
}

var _ list.ListResourceWithRawV5Schemas = &policyListResource{}

type policyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type policyListResourceModel struct {
	PathPrefix types.String `tfsdk:"path_prefix"`
}

func (l *policyListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"path_prefix": listschema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validPolicyPathFramework,
				},
			},
		},
	}
}

func (l *policyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	var query policyListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input iam.ListPoliciesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}
	input.Scope = awstypes.PolicyScopeTypeLocal

	tflog.Info(ctx, "Listing IAM Policy resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for policy, err := range listPolicies(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := resourcePolicyListItemLoggingContext(ctx, policy)
			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(aws.ToString(policy.Arn))

			tflog.Info(ctx, "Reading resource")
			resourcePolicyFlatten(ctx, &policy, rd)

			result.DisplayName = resourcePolicyDisplayName(rd)

			if request.IncludeResource {
				tflog.Info(ctx, "Reading additional resource data")

				policyVersion, err := findPolicyVersionByTwoPartKey(ctx, conn, aws.ToString(policy.Arn), aws.ToString(policy.DefaultVersionId))
				if retry.NotFound(err) {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping")
					continue
				}
				if err != nil {
					result = fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}

				if err := resourcePolicyFlattenPolicyDocument(aws.ToString(policyVersion.Document), rd); err != nil {
					result = fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}
			}

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listPolicies(ctx context.Context, conn *iam.Client, input *iam.ListPoliciesInput) iter.Seq2[awstypes.Policy, error] {
	return func(yield func(awstypes.Policy, error) bool) {
		pages := iam.NewListPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Policy{}, fmt.Errorf("listing IAM Policies: %w", err))
				return
			}
			for _, policy := range page.Policies {
				if !yield(policy, nil) {
					return
				}
			}
		}
	}
}

func resourcePolicyListItemLoggingContext(ctx context.Context, policy awstypes.Policy) context.Context {
	return tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), aws.ToString(policy.Arn))
}

func resourcePolicyDisplayName(d *schema.ResourceData) string {
	var buf strings.Builder

	path := d.Get(names.AttrPath).(string)
	buf.WriteString(strings.TrimPrefix(path, "/"))

	buf.WriteString(d.Get(names.AttrName).(string))

	return buf.String()
}
