// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_vpclattice_service_network_service_associations", name="Service Network Service Associations")
func newDataSourceServiceNetworkServiceAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServiceNetworkServiceAssociations{}, nil
}

type dataSourceServiceNetworkServiceAssociations struct {
	framework.DataSourceWithModel[dataSourceServiceNetworkServiceAssociationsModel]
}

func (d *dataSourceServiceNetworkServiceAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"items": framework.DataSourceComputedListOfObjectAttribute[itemModel](ctx),
			"service_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "ID or ARN of the Service.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("service_network_identifier"),
					),
				},
			},
			"service_network_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "ID or ARN of the Service Network.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("service_identifier"),
					),
				},
			},
		},
	}
}

func (d *dataSourceServiceNetworkServiceAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().VPCLatticeClient(ctx)

	var data dataSourceServiceNetworkServiceAssociationsModel

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input vpclattice.ListServiceNetworkServiceAssociationsInput
	if !data.ServiceNetworkIdentifier.IsNull() {
		snID := data.ServiceNetworkIdentifier.ValueString()
		if _, err := findServiceNetworkByID(ctx, conn, snID); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
			return
		}
		input.ServiceNetworkIdentifier = data.ServiceNetworkIdentifier.ValueStringPointer()
	} else if !data.ServiceIdentifier.IsNull() {
		sID := data.ServiceIdentifier.ValueString()
		if _, err := findServiceByID(ctx, conn, sID); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
			return
		}
		input.ServiceIdentifier = data.ServiceIdentifier.ValueStringPointer()
	}

	out, err := findServiceNetworkServiceAssociations(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data.Items))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findServiceNetworkServiceAssociations(ctx context.Context, conn *vpclattice.Client, input *vpclattice.ListServiceNetworkServiceAssociationsInput) ([]awstypes.ServiceNetworkServiceAssociationSummary, error) {
	var output []awstypes.ServiceNetworkServiceAssociationSummary
	paginator := vpclattice.NewListServiceNetworkServiceAssociationsPaginator(conn, input, func(opts *vpclattice.ListServiceNetworkServiceAssociationsPaginatorOptions) {
		opts.Limit = 100
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		output = append(output, page.Items...)
	}
	return output, nil
}

type dataSourceServiceNetworkServiceAssociationsModel struct {
	framework.WithRegionModel
	Items                    fwtypes.ListNestedObjectValueOf[itemModel] `tfsdk:"items"`
	ServiceIdentifier        types.String                               `tfsdk:"service_identifier"`
	ServiceNetworkIdentifier types.String                               `tfsdk:"service_network_identifier"`
}

type itemModel struct {
	ARN                types.String                                                        `tfsdk:"arn"`
	CreatedAt          timetypes.RFC3339                                                   `tfsdk:"created_at"`
	CreatedBy          types.String                                                        `tfsdk:"created_by"`
	CustomDomainName   types.String                                                        `tfsdk:"custom_domain_name"`
	DNSEntry           fwtypes.ListNestedObjectValueOf[dnsEntryModel]                      `tfsdk:"dns_entry"`
	ID                 types.String                                                        `tfsdk:"id"`
	ServiceARN         types.String                                                        `tfsdk:"service_arn"`
	ServiceID          types.String                                                        `tfsdk:"service_id"`
	ServiceName        types.String                                                        `tfsdk:"service_name"`
	ServiceNetworkARN  types.String                                                        `tfsdk:"service_network_arn"`
	ServiceNetworkID   types.String                                                        `tfsdk:"service_network_id"`
	ServiceNetworkName types.String                                                        `tfsdk:"service_network_name"`
	Status             fwtypes.StringEnum[awstypes.ServiceNetworkServiceAssociationStatus] `tfsdk:"status"`
}
