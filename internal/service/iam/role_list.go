// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/go-cty/cty"
	frameworkdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_iam_role")
func newRoleResourceAsListResource() inttypes.ListResourceForSDK {
	l := roleListResource{}
	l.SetResourceSchema(resourceRole())

	return &l
}

type roleListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type roleListResourceModel struct{}

func (l *roleListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	var query roleListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input iam.ListRolesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for role, err := range listNonServiceLinkedRoles(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := resourceRoleListItemLoggingContext(ctx, role)
			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(aws.ToString(role.RoleName))

			tflog.Info(ctx, "Reading resource")
			result.Diagnostics.Append(translateDiags(resourceRoleFlatten(ctx, &role, rd))...)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			result.DisplayName = resourceRoleDisplayName(role)

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

func resourceRoleDisplayName(role awstypes.Role) string {
	var buf strings.Builder

	path := aws.ToString(role.Path)
	buf.WriteString(strings.TrimPrefix(path, "/"))

	buf.WriteString(aws.ToString(role.RoleName))

	return buf.String()
}

func translateDiags(in diag.Diagnostics) frameworkdiag.Diagnostics {
	out := make(frameworkdiag.Diagnostics, len(in))
	for i, diagIn := range in {
		var diagOut frameworkdiag.Diagnostic
		if diagIn.Severity == diag.Error {
			if len(diagIn.AttributePath) == 0 {
				diagOut = frameworkdiag.NewErrorDiagnostic(diagIn.Summary, diagIn.Detail)
			} else {
				diagOut = frameworkdiag.NewAttributeErrorDiagnostic(translatePath(diagIn.AttributePath), diagIn.Summary, diagIn.Detail)
			}
		} else {
			if len(diagIn.AttributePath) == 0 {
				diagOut = frameworkdiag.NewWarningDiagnostic(diagIn.Summary, diagIn.Detail)
			} else {
				diagOut = frameworkdiag.NewAttributeWarningDiagnostic(translatePath(diagIn.AttributePath), diagIn.Summary, diagIn.Detail)
			}
		}
		out[i] = diagOut
	}
	return out
}

func translatePath(in cty.Path) path.Path {
	var out path.Path

	if len(in) == 0 {
		return out
	}

	step := in[0]
	switch v := step.(type) {
	case cty.GetAttrStep:
		out = path.Root(v.Name)
	}

	for i := 1; i < len(in); i++ {
		step := in[i]
		switch v := step.(type) {
		case cty.GetAttrStep:
			out = out.AtName(v.Name)

		case cty.IndexStep:
			switch v.Key.Type() {
			case cty.Number:
				v, _ := v.Key.AsBigFloat().Int64()
				out = out.AtListIndex(int(v))
			case cty.String:
				out = out.AtMapKey(v.Key.AsString())
			}
		}
	}

	return out
}

func resourceRoleListItemLoggingContext(ctx context.Context, role awstypes.Role) context.Context {
	return tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), aws.ToString(role.RoleName))
}
