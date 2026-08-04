// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
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

// @SDKListResource("aws_route53_vpc_association_authorization")
func newVPCAssociationAuthorizationResourceAsListResource() inttypes.ListResourceForSDK {
	l := vpcAssociationAuthorizationListResource{}
	l.SetResourceSchema(resourceVPCAssociationAuthorization())
	return &l
}

var _ list.ListResource = &vpcAssociationAuthorizationListResource{}

type vpcAssociationAuthorizationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type vpcAssociationAuthorizationListResourceModel struct {
	ZoneID types.String `tfsdk:"zone_id"`
}

func (l *vpcAssociationAuthorizationListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"zone_id": listschema.StringAttribute{
				Required:    true,
				Description: "ID of the hosted zone to list VPC association authorizations for.",
			},
		},
	}
}

func (l *vpcAssociationAuthorizationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query vpcAssociationAuthorizationListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	awsClient := l.Meta()
	conn := awsClient.Route53Client(ctx)

	zoneID := query.ZoneID.ValueString()

	tflog.Info(ctx, "Listing Route53 VPC Association Authorizations", map[string]any{
		"zone_id": zoneID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &route53.ListVPCAssociationAuthorizationsInput{
			HostedZoneId: aws.String(zoneID),
		}

		err := listVPCAssociationAuthorizationsPages(ctx, conn, input, func(page *route53.ListVPCAssociationAuthorizationsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, vpc := range page.VPCs {
				vpcID := aws.ToString(vpc.VPCId)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrVPCID), vpcID)
				id := vpcAssociationAuthorizationCreateResourceID(zoneID, vpcID)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.SetId(id)
				rd.Set("zone_id", zoneID)
				rd.Set(names.AttrVPCID, vpcID)

				if request.IncludeResource {
					resourceVPCAssociationAuthorizationFlatten(rd, zoneID, &vpc)
				}

				result.DisplayName = id

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
