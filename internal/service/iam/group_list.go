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

// @SDKListResource("aws_iam_group")
func newGroupResourceAsListResource() inttypes.ListResourceForSDK {
	l := groupListResource{}
	l.SetResourceSchema(resourceGroup())

	return &l
}

var _ list.ListResource = &groupListResource{}

type groupListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listGroupModel struct {
	PathPrefix types.String `tfsdk:"path_prefix"`
}

func (l *groupListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"path_prefix": listschema.StringAttribute{
				Optional:    true,
				Description: "Path prefix on which to filter group names.",
				Validators: []validator.String{
					validPolicyPathFramework,
				},
			},
		},
	}
}

func (l *groupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().IAMClient(ctx)

	var query listGroupModel
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
		input := iam.ListGroupsInput{
			PathPrefix: query.PathPrefix.ValueStringPointer(),
		}
		for item, err := range listGroups(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}
			name := aws.ToString(item.GroupName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				resourceGroupFlatten(&item, rd)
			}

			result.DisplayName = name

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

func listGroups(ctx context.Context, conn *iam.Client, input *iam.ListGroupsInput) iter.Seq2[awstypes.Group, error] {
	return func(yield func(awstypes.Group, error) bool) {
		pages := iam.NewListGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Group{}, fmt.Errorf("listing IAM Group resources: %w", err))
				return
			}

			for _, item := range page.Groups {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
