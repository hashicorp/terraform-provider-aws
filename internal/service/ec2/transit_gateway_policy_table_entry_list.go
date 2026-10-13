// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
)

// @FrameworkListResource("aws_ec2_transit_gateway_policy_table_entry")
func newTransitGatewayPolicyTableEntryResourceAsListResource() list.ListResourceWithConfigure {
	return &transitGatewayPolicyTableEntryListResource{}
}

var _ list.ListResource = &transitGatewayPolicyTableEntryListResource{}

type transitGatewayPolicyTableEntryListResource struct {
	transitGatewayPolicyTableEntryResource
	framework.WithList
}

func (l *transitGatewayPolicyTableEntryListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"transit_gateway_policy_table_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the transit gateway policy table.",
			},
		},
	}
}

func (l *transitGatewayPolicyTableEntryListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	var query listTransitGatewayPolicyTableEntryModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	policyTableID := fwflex.StringValueFromFramework(ctx, query.TransitGatewayPolicyTableID)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("transit_gateway_policy_table_id"): policyTableID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := ec2.GetTransitGatewayPolicyTableEntriesInput{
			TransitGatewayPolicyTableId: aws.String(policyTableID),
		}
		for item, err := range listTransitGatewayPolicyTableEntries(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data transitGatewayPolicyTableEntryResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.TransitGatewayPolicyTableID = fwflex.StringValueToFramework(ctx, policyTableID)
				result.Diagnostics.Append(l.flatten(ctx, &item, nil, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = aws.ToString(item.PolicyRuleNumber)
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listTransitGatewayPolicyTableEntryModel struct {
	framework.WithRegionModel
	TransitGatewayPolicyTableID types.String `tfsdk:"transit_gateway_policy_table_id"`
}
