// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_directory_service_ip_route")
func newIPRouteResourceAsListResource() list.ListResourceWithConfigure {
	return &ipRouteListResource{}
}

var _ list.ListResource = &ipRouteListResource{}

type ipRouteListResource struct {
	ipRouteResource
	framework.WithList
}

func (l *ipRouteListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"directory_id": listschema.StringAttribute{
				Required:    true,
				Description: "Identifier of the directory whose IP routes to list.",
				Validators: []validator.String{
					directoryIDValidator,
				},
			},
		},
	}
}

func (l *ipRouteListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query ipRouteListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	conn := l.Meta().DSClient(ctx)
	directoryID := fwflex.StringValueFromFramework(ctx, query.DirectoryID)
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("directory_id"), directoryID)

	stream.Results = func(yield func(list.ListResult) bool) {
		routes, err := findIPRoutesByDirectoryID(ctx, conn, directoryID)
		if err != nil {
			yield(smerr.NewListResultError(ctx, err, smerr.ID, directoryID))
			return
		}

		for _, route := range routes {
			cidr := ipRouteInfoKey(route)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("cidr"), cidr)

			result := request.NewListResult(ctx)

			var data ipRouteResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.DirectoryID = types.StringValue(directoryID)
				l.flatten(ctx, &route, &data)

				result.DisplayName = cidr
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type ipRouteListModel struct {
	framework.WithRegionModel
	DirectoryID types.String `tfsdk:"directory_id"`
}
