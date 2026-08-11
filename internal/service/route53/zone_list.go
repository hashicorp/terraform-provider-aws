// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_route53_zone")
func newZoneResourceAsListResource() inttypes.ListResourceForSDK {
	l := zoneListResource{}
	l.SetResourceSchema(resourceZone())
	return &l
}

var _ list.ListResource = &zoneListResource{}

type zoneListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type zoneListResourceModel struct {
	PrivateZone types.Bool `tfsdk:"private_zone"`
}

func (l *zoneListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"private_zone": listschema.BoolAttribute{
				Optional:    true,
				Description: "When true, only private hosted zones are returned.",
			},
		},
	}
}

func (l *zoneListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.Route53Client(ctx)

	var query zoneListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Route 53 Hosted Zones")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &route53.ListHostedZonesInput{}
		if !query.PrivateZone.IsNull() && query.PrivateZone.ValueBool() {
			input.HostedZoneType = awstypes.HostedZoneTypePrivateHostedZone
		}

		for item, err := range listZones(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			zoneID := cleanZoneID(aws.ToString(item.Id))
			itemCtx := tflog.SetField(ctx, logging.ResourceAttributeKey("zone_id"), zoneID)

			result := request.NewListResult(itemCtx)
			rd := l.ResourceData()
			rd.SetId(zoneID)
			rd.Set("zone_id", zoneID)

			if request.IncludeResource {
				output, err := findHostedZoneByID(itemCtx, conn, zoneID)
				if err != nil {
					tflog.Error(itemCtx, "Reading Route 53 Hosted Zone", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				zoneTags, err := listTags(itemCtx, conn, zoneID, string(awstypes.TagResourceTypeHostedzone))
				if err != nil {
					tflog.Error(itemCtx, "Reading Route 53 Hosted Zone tags", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				setTagsOut(itemCtx, svcTags(zoneTags))

				diags := resourceZoneFlatten(itemCtx, conn, rd, awsClient, output)
				if diags.HasError() {
					tflog.Error(itemCtx, "Flattening Route 53 Hosted Zone", map[string]any{
						"diags": sdkdiag.DiagnosticsString(diags),
					})
					continue
				}

				if rd.Id() == "" {
					continue
				}
			}

			result.DisplayName = normalizeDomainName(aws.ToString(item.Name))

			l.SetResult(itemCtx, awsClient, request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				tflog.Error(itemCtx, "Setting Route 53 Hosted Zone result", map[string]any{
					"diags": result.Diagnostics,
				})
				yield(result)
				continue
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listZones(ctx context.Context, conn *route53.Client, input *route53.ListHostedZonesInput) iter.Seq2[awstypes.HostedZone, error] {
	return func(yield func(awstypes.HostedZone, error) bool) {
		pages := route53.NewListHostedZonesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.HostedZone{}, fmt.Errorf("listing Route 53 Hosted Zone resources: %w", err))
				return
			}

			for _, item := range page.HostedZones {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
