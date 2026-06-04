// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_connect_approved_origin")
func newApprovedOriginResourceAsListResource() list.ListResourceWithConfigure {
	return &approvedOriginListResource{}
}

var _ list.ListResource = &approvedOriginListResource{}

type approvedOriginListResource struct {
	approvedOriginResource
	framework.WithList
}

func (l *approvedOriginListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrInstanceID: listschema.StringAttribute{
				Required:    true,
				Description: "The identifier of the Amazon Connect instance.",
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *approvedOriginListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ConnectClient(ctx)

	var query listApprovedOriginsModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	instanceID := query.InstanceID.ValueString()

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &connect.ListApprovedOriginsInput{
			InstanceId: aws.String(instanceID),
		}

		pages := connect.NewListApprovedOriginsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, origin := range page.Origins {
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("origin"), origin)

				result := request.NewListResult(ctx)

				var data approvedOriginResourceModel
				l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
					data.InstanceID = types.StringValue(instanceID)
					data.Origin = types.StringValue(origin)
					result.DisplayName = origin
				})

				if !yield(result) {
					return
				}
			}
		}
	}
}

type listApprovedOriginsModel struct {
	InstanceID types.String `tfsdk:"instance_id"`
}
