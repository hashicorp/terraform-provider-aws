// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

	tflog.Info(ctx, "Listing IAM Roles")
	stream.Results = func(yield func(list.ListResult) bool) {
		for role, err := range listNonServiceLinkedRoles(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			roleName := aws.ToString(role.RoleName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), roleName)
			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(roleName)

			tflog.Info(ctx, "Reading IAM Role")
			diags := resourceRoleFlatten(ctx, &role, rd)
			if diags.HasError() {
				result = fwdiag.NewListResultSDKDiagnostics(diags)
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
