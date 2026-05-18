// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_route53_zone_association")
func newZoneAssociationResourceAsListResource() inttypes.ListResourceForSDK {
	l := zoneAssociationListResource{}
	l.SetResourceSchema(resourceZoneAssociation())
	return &l
}

var _ list.ListResource = &zoneAssociationListResource{}

type zoneAssociationListResource struct {
	framework.ListResourceWithSDKv2Resource
}
type zoneAssociationListResourceModel struct {
	VPCID     types.String `tfsdk:"vpc_id"`
	VPCRegion types.String `tfsdk:"vpc_region"`
}

func (l *zoneAssociationListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"vpc_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the VPC to list hosted zone associations for.",
			},
			"vpc_region": listschema.StringAttribute{
				Optional:    true,
				Description: "Region of the VPC. Defaults to the provider region.",
			},
		},
	}
}

func (l *zoneAssociationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query zoneAssociationListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	awsClient := l.Meta()
	conn := awsClient.Route53Client(ctx)

	vpcID := query.VPCID.ValueString()
	vpcRegion := query.VPCRegion.ValueString()
	if vpcRegion == "" {
		vpcRegion = awsClient.Region(ctx)
	}

	tflog.Info(ctx, "Listing Route 53 Zone Associations", map[string]any{
		logging.ResourceAttributeKey(names.AttrVPCID): vpcID,
		logging.ResourceAttributeKey("vpc_region"):    vpcRegion,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &route53.ListHostedZonesByVPCInput{
			VPCId:     aws.String(vpcID),
			VPCRegion: awstypes.VPCRegion(vpcRegion),
		}

		err := listHostedZonesByVPCPages(ctx, conn, input, func(page *route53.ListHostedZonesByVPCOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, summary := range page.HostedZoneSummaries {
				hostedZoneID := aws.ToString(summary.HostedZoneId)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("zone_id"), hostedZoneID)
				id := zoneAssociationCreateResourceID(hostedZoneID, vpcID, vpcRegion)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.SetId(id)
				rd.Set("zone_id", hostedZoneID)
				rd.Set(names.AttrVPCID, vpcID)
				rd.Set("vpc_region", vpcRegion)

				if request.IncludeResource {
					resourceZoneAssociationFlatten(rd, &summary, vpcID, vpcRegion)
				}

				result.DisplayName = normalizeDomainName(aws.ToString(summary.Name))

				l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
				if result.Diagnostics.HasError() {
					yield(result)
					return false
				}

				if !yield(result) {
					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}
	}
}
