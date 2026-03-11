// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_iam_user")
func newUserResourceAsListResource() inttypes.ListResourceForSDK {
	l := userListResource{}
	l.SetResourceSchema(resourceUser())
	return &l
}

var _ list.ListResource = &userListResource{}

type userListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listUserModel struct {
	PathPrefix types.String `tfsdk:"path_prefix"`
}

func (l *userListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"path_prefix": listschema.StringAttribute{
				Optional:    true,
				Description: "Path prefix on which to filter user names.",
				Validators: []validator.String{
					validPolicyPathFramework,
				},
			},
		},
	}
}

func (l *userListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().IAMClient(ctx)

	var query listUserModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	pathPrefix := query.PathPrefix.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("path_prefix"): pathPrefix,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := iam.ListUsersInput{
			PathPrefix: query.PathPrefix.ValueStringPointer(),
		}
		for item, err := range listUsers(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}
			name := aws.ToString(item.UserName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				resourceUserFlatten(&item, rd)
			}

			result.DisplayName = aws.ToString(item.UserName)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
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

func listUsers(ctx context.Context, conn *iam.Client, input *iam.ListUsersInput) iter.Seq2[awstypes.User, error] {
	return func(yield func(awstypes.User, error) bool) {
		pages := iam.NewListUsersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.User{}, fmt.Errorf("listing IAM User resources: %w", err))
				return
			}

			for _, item := range page.Users {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
