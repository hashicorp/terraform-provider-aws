// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_pinpointsmsvoicev2_event_destination")
func newEventDestinationResourceAsListResource() list.ListResourceWithConfigure {
	return &eventDestinationListResource{}
}

var _ list.ListResource = &eventDestinationListResource{}

type eventDestinationListResource struct {
	eventDestinationResource
	framework.WithList
}

func (l *eventDestinationListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"configuration_set_names": listschema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Description: "Names of configuration sets to list event destinations for.",
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *eventDestinationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	var query eventDestinationListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var configurationSetNames []string
	if diags := fwflex.Expand(ctx, query.ConfigurationSetNames, &configurationSetNames); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	input := pinpointsmsvoicev2.DescribeConfigurationSetsInput{
		ConfigurationSetNames: configurationSetNames,
	}

	tflog.Info(ctx, "Listing End User Messaging SMS Event Destinations")

	stream.Results = func(yield func(list.ListResult) bool) {
		configSets, err := findConfigurationSets(ctx, conn, &input)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		for i := range configSets {
			configSet := &configSets[i]
			configSetName := aws.ToString(configSet.ConfigurationSetName)
			configSetARN := aws.ToString(configSet.ConfigurationSetArn)

			for j := range configSet.EventDestinations {
				item := &configSet.EventDestinations[j]
				eventDestinationName := aws.ToString(item.EventDestinationName)

				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("configuration_set_name"), configSetName)
				ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("event_destination_name"), eventDestinationName)

				result := request.NewListResult(ctx)

				var data eventDestinationResourceModel
				l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
					smerr.AddEnrich(ctx, &result.Diagnostics, fwflex.Flatten(ctx, item, &data))
					if result.Diagnostics.HasError() {
						return
					}

					data.ConfigurationSetName = fwflex.StringToFramework(ctx, configSet.ConfigurationSetName)
					data.ConfigurationSetARN = fwtypes.ARNValue(configSetARN)

					result.DisplayName = eventDestinationName
				})

				if !yield(result) {
					return
				}
			}
		}
	}
}

type eventDestinationListModel struct {
	framework.WithRegionModel
	ConfigurationSetNames fwtypes.ListOfString `tfsdk:"configuration_set_names"`
}
