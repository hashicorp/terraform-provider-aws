// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_bedrockagentcore_registry_record")
func newRegistryRecordResourceAsListResource() list.ListResourceWithConfigure {
	return &registryRecordListResource{}
}

var _ list.ListResource = &registryRecordListResource{}

type registryRecordListResource struct {
	registryRecordResource
	framework.WithList
}

func (l *registryRecordListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"registry_id": listschema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (l *registryRecordListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().BedrockAgentCoreClient(ctx)

	var query listRegistryRecordModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	registryID := fwflex.StringValueFromFramework(ctx, query.RegistryID)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("registry_id"): registryID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := bedrockagentcorecontrol.ListRegistryRecordsInput{
			RegistryId: aws.String(registryID),
		}
		for item, err := range listRegistryRecords(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), aws.ToString(item.RecordArn))

			output, err := findRegistryRecordByTwoPartKey(ctx, conn, registryID, aws.ToString(item.RecordId))
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data registryRecordResourceModel
			// bedrockagentcorecontrol.GetRegistryRecordOutput holds registry ARN, but not registry ID.
			data.RegistryID = fwflex.StringValueToFramework(ctx, registryID)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listRegistryRecordModel struct {
	framework.WithRegionModel
	RegistryID types.String `tfsdk:"registry_id"`
}

func listRegistryRecords(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.ListRegistryRecordsInput) iter.Seq2[awstypes.RegistryRecordSummary, error] {
	return func(yield func(awstypes.RegistryRecordSummary, error) bool) {
		pages := bedrockagentcorecontrol.NewListRegistryRecordsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.RegistryRecordSummary](), fmt.Errorf("listing Bedrock AgentCore Registry Records: %w", err))
				return
			}

			for _, item := range page.RegistryRecords {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
