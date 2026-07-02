// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_wafv2_web_acl_rule")
func newResourceWebACLRuleAsListResource() list.ListResourceWithConfigure {
	return &webACLRuleListResource{}
}

var _ list.ListResource = &webACLRuleListResource{}

type webACLRuleListResource struct {
	resourceWebACLRule
	framework.WithList
}

func (r *webACLRuleListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"web_acl_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the Web ACL whose rules to list.",
			},
		},
	}
}

func (r *webACLRuleListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().WAFV2Client(ctx)

	var query listWebACLRuleModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	webACLARN := query.WebACLARN.ValueString()
	webACLID, webACLName, webACLScope, err := parseWebACLARN(webACLARN)
	if err != nil {
		stream.Results = func(yield func(list.ListResult) bool) {
			yield(fwdiag.NewListResultErrorDiagnostic(err))
		}
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)

		webACL, err := findWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
		if retry.NotFound(err) {
			return
		}
		if err != nil {
			result = fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		for i := range webACL.WebACL.Rules {
			rule := &webACL.WebACL.Rules[i]
			ruleName := aws.ToString(rule.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), ruleName)

			var data webACLRuleModel
			r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(r.flattenWebACLRule(ctx, rule, &data)...)
				if result.Diagnostics.HasError() {
					return
				}
				data.WebACLARN = fwflex.StringValueToFramework(ctx, webACLARN)
				result.DisplayName = ruleName
			})

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

type listWebACLRuleModel struct {
	framework.WithRegionModel
	WebACLARN types.String `tfsdk:"web_acl_arn"`
}
