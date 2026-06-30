// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_vpclattice_service_network_resource_associations", name="Service Network Resource Associations")
func newDataSourceServiceNetworkResourceAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServiceNetworkResourceAssociations{}, nil
}

type dataSourceServiceNetworkResourceAssociations struct {
	framework.DataSourceWithModel[dataSourceServiceNetworkResourceAssociationsModel]
}

func (d *dataSourceServiceNetworkResourceAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"associations": framework.DataSourceComputedListOfObjectAttribute[associationModel](ctx),
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			"resource_configuration_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "ID or ARN of the Resource Configuration.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("service_network_identifier"),
					),
				},
			},
			"service_network_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "ID or ARN of the Service Network.",
			},
		},
	}
}

func (d *dataSourceServiceNetworkResourceAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().VPCLatticeClient(ctx)

	var data dataSourceServiceNetworkResourceAssociationsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var associations []awstypes.ServiceNetworkResourceAssociationSummary

	switch {
	case !data.ServiceNetworkIdentifier.IsNull():
		snID := data.ServiceNetworkIdentifier.ValueString()
		data.ID = types.StringValue(snID)

		if _, err := findServiceNetworkByID(ctx, conn, snID); err != nil {
			resp.Diagnostics.AddError("reading VPC Lattice Service Network ("+snID+")", err.Error())
			return
		}

		input := vpclattice.ListServiceNetworkResourceAssociationsInput{
			ServiceNetworkIdentifier: aws.String(snID),
		}
		output, err := listServiceNetworkResourceAssociations(ctx, conn, &input)
		if err != nil {
			resp.Diagnostics.AddError("listing VPC Lattice Service Network Resource Associations", err.Error())
			return
		}
		associations = output
	case !data.ResourceConfigurationIdentifier.IsNull():
		rcID := data.ResourceConfigurationIdentifier.ValueString()
		data.ID = types.StringValue(rcID)

		if _, err := findResourceConfigurationByID(ctx, conn, rcID); err != nil {
			resp.Diagnostics.AddError("reading VPC Lattice Resource Configuration ("+rcID+")", err.Error())
			return
		}

		input := vpclattice.ListServiceNetworkResourceAssociationsInput{
			ResourceConfigurationIdentifier: aws.String(rcID),
		}
		output, err := listServiceNetworkResourceAssociations(ctx, conn, &input)
		if err != nil {
			resp.Diagnostics.AddError("listing VPC Lattice Service Network Resource Associations", err.Error())
			return
		}
		associations = output
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, associations, &data.Associations)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func listServiceNetworkResourceAssociations(ctx context.Context, conn *vpclattice.Client, input *vpclattice.ListServiceNetworkResourceAssociationsInput) ([]awstypes.ServiceNetworkResourceAssociationSummary, error) {
	var output []awstypes.ServiceNetworkResourceAssociationSummary

	paginator := vpclattice.NewListServiceNetworkResourceAssociationsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		output = append(output, page.Items...)
	}

	return output, nil
}

type dataSourceServiceNetworkResourceAssociationsModel struct {
	framework.WithRegionModel
	Associations                    fwtypes.ListNestedObjectValueOf[associationModel] `tfsdk:"associations"`
	ID                              types.String                                      `tfsdk:"id"`
	ResourceConfigurationIdentifier types.String                                      `tfsdk:"resource_configuration_identifier"`
	ServiceNetworkIdentifier        types.String                                      `tfsdk:"service_network_identifier"`
}

type associationModel struct {
	ARN                       types.String                                                         `tfsdk:"arn"`
	CreatedBy                 types.String                                                         `tfsdk:"created_by"`
	DNSEntry                  fwtypes.ListNestedObjectValueOf[dnsEntryModel]                       `tfsdk:"dns_entry"`
	FailureCode               types.String                                                         `tfsdk:"failure_code"`
	ID                        types.String                                                         `tfsdk:"id"`
	IsManagedAssociation      types.Bool                                                           `tfsdk:"is_managed_association"`
	PrivateDNSEntry           fwtypes.ListNestedObjectValueOf[dnsEntryModel]                       `tfsdk:"private_dns_entry"`
	ResourceConfigurationARN  types.String                                                         `tfsdk:"resource_configuration_arn"`
	ResourceConfigurationID   types.String                                                         `tfsdk:"resource_configuration_id"`
	ResourceConfigurationName types.String                                                         `tfsdk:"resource_configuration_name"`
	ServiceNetworkARN         types.String                                                         `tfsdk:"service_network_arn"`
	ServiceNetworkID          types.String                                                         `tfsdk:"service_network_id"`
	ServiceNetworkName        types.String                                                         `tfsdk:"service_network_name"`
	Status                    fwtypes.StringEnum[awstypes.ServiceNetworkResourceAssociationStatus] `tfsdk:"status"`
}
