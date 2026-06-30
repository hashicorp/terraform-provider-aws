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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_vpclattice_service_network_service_associations", name="Service Network Service Associations")
func newDataSourceServiceNetworkServiceAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServiceNetworkServiceAssociations{}, nil
}

const (
	DSNameServiceNetworkServiceAssociations = "Service Network Service Associations Data Source"
)

type dataSourceServiceNetworkServiceAssociations struct {
	framework.DataSourceWithModel[dataSourceServiceNetworkServiceAssociationsModel]
}

func (d *dataSourceServiceNetworkServiceAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_network_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "ID or ARN of the Service Network.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("service_identifier"),
					),
				},
			},
			"service_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "ID or ARN of the Service.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("service_network_identifier"),
					),
				},
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"associations": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN:          framework.ARNAttributeComputedOnly(),
						"created_by":           schema.StringAttribute{Computed: true},
						"custom_domain_name":   schema.StringAttribute{Computed: true},
						names.AttrID:           schema.StringAttribute{Computed: true},
						"service_arn":          schema.StringAttribute{Computed: true},
						"service_id":           schema.StringAttribute{Computed: true},
						names.AttrServiceName:  schema.StringAttribute{Computed: true},
						"service_network_arn":  schema.StringAttribute{Computed: true},
						"service_network_id":   schema.StringAttribute{Computed: true},
						"service_network_name": schema.StringAttribute{Computed: true},
						names.AttrStatus:       schema.StringAttribute{Computed: true},
					},
					Blocks: map[string]schema.Block{
						"dns_entry": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDomainName:   schema.StringAttribute{Computed: true},
									names.AttrHostedZoneID: schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceServiceNetworkServiceAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().VPCLatticeClient(ctx)

	var associations []*awstypes.ServiceNetworkServiceAssociationSummary

	var data dataSourceServiceNetworkServiceAssociationsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ServiceNetworkIdentifier.IsNull() {
		snID := data.ServiceNetworkIdentifier.ValueString()
		data.ID = types.StringValue(snID)

		if _, err := findServiceNetworkByID(ctx, conn, snID); err != nil {
			resp.Diagnostics.AddError("Error reading Service Network", err.Error())
			return
		}

		items, err := listServiceNetworkServiceAssociationsByServiceNetworkIdentifier(ctx, conn, snID)
		if err != nil {
			resp.Diagnostics.AddError("Error listing associations", err.Error())
			return
		}
		associations = items
	} else if !data.ServiceIdentifier.IsNull() {
		sID := data.ServiceIdentifier.ValueString()
		data.ID = types.StringValue(sID)

		if _, err := findServiceByID(ctx, conn, sID); err != nil {
			resp.Diagnostics.AddError("Error reading Service", err.Error())
			return
		}

		items, err := listServiceNetworkServiceAssociationsByServiceIdentifier(ctx, conn, sID)
		if err != nil {
			resp.Diagnostics.AddError("Error listing associations", err.Error())
			return
		}
		associations = items
	}

	data.Associations = flattenServiceNetworkServiceAssociations(associations)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenServiceNetworkServiceAssociations(objs []*awstypes.ServiceNetworkServiceAssociationSummary) []associationModel {
	result := make([]associationModel, 0, len(objs))
	for _, obj := range objs {
		result = append(result, flattenServiceNetworkServiceAssociation(obj))
	}
	return result
}

func flattenServiceNetworkServiceAssociation(obj *awstypes.ServiceNetworkServiceAssociationSummary) associationModel {
	var dnsEntries []dnsEntryModel
	if obj.DnsEntry != nil {
		dnsEntries = []dnsEntryModel{flattenDNSEntryFramework(obj.DnsEntry)}
	}

	return associationModel{
		Arn:                types.StringPointerValue(obj.Arn),
		CreatedBy:          types.StringPointerValue(obj.CreatedBy),
		CustomDomainName:   types.StringPointerValue(obj.CustomDomainName),
		DnsEntry:           dnsEntries,
		ID:                 types.StringPointerValue(obj.Id),
		ServiceArn:         types.StringPointerValue(obj.ServiceArn),
		ServiceId:          types.StringPointerValue(obj.ServiceId),
		ServiceName:        types.StringPointerValue(obj.ServiceName),
		ServiceNetworkArn:  types.StringPointerValue(obj.ServiceNetworkArn),
		ServiceNetworkId:   types.StringPointerValue(obj.ServiceNetworkId),
		ServiceNetworkName: types.StringPointerValue(obj.ServiceNetworkName),
		Status:             types.StringValue(string(obj.Status)),
	}
}

func flattenDNSEntryFramework(obj *awstypes.DnsEntry) dnsEntryModel {
	return dnsEntryModel{
		DomainName:   types.StringPointerValue(obj.DomainName),
		HostedZoneID: types.StringPointerValue(obj.HostedZoneId),
	}
}

func listServiceNetworkServiceAssociationsByServiceIdentifier(ctx context.Context, conn *vpclattice.Client, serviceIdentifier string) ([]*awstypes.ServiceNetworkServiceAssociationSummary, error) {
	input := vpclattice.ListServiceNetworkServiceAssociationsInput{
		ServiceIdentifier: aws.String(serviceIdentifier),
	}
	return listServiceNetworkServiceAssociations(ctx, conn, &input)
}

func listServiceNetworkServiceAssociationsByServiceNetworkIdentifier(ctx context.Context, conn *vpclattice.Client, serviceNetworkIdentifier string) ([]*awstypes.ServiceNetworkServiceAssociationSummary, error) {
	input := vpclattice.ListServiceNetworkServiceAssociationsInput{
		ServiceNetworkIdentifier: aws.String(serviceNetworkIdentifier),
	}
	return listServiceNetworkServiceAssociations(ctx, conn, &input)
}

func listServiceNetworkServiceAssociations(ctx context.Context, conn *vpclattice.Client, input *vpclattice.ListServiceNetworkServiceAssociationsInput) ([]*awstypes.ServiceNetworkServiceAssociationSummary, error) {
	var output []*awstypes.ServiceNetworkServiceAssociationSummary
	paginator := vpclattice.NewListServiceNetworkServiceAssociationsPaginator(conn, input, func(opts *vpclattice.ListServiceNetworkServiceAssociationsPaginatorOptions) {
		opts.Limit = 100
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for idx := range page.Items {
			v := page.Items[idx]
			output = append(output, &v)
		}
	}
	return output, nil
}

type dataSourceServiceNetworkServiceAssociationsModel struct {
	framework.WithRegionModel
	ServiceNetworkIdentifier types.String       `tfsdk:"service_network_identifier"`
	ServiceIdentifier        types.String       `tfsdk:"service_identifier"`
	Associations             []associationModel `tfsdk:"associations"`
	ID                       types.String       `tfsdk:"id"`
}

type associationModel struct {
	Arn                types.String    `tfsdk:"arn"`
	CreatedBy          types.String    `tfsdk:"created_by"`
	CustomDomainName   types.String    `tfsdk:"custom_domain_name"`
	DnsEntry           []dnsEntryModel `tfsdk:"dns_entry"`
	ID                 types.String    `tfsdk:"id"`
	ServiceArn         types.String    `tfsdk:"service_arn"`
	ServiceId          types.String    `tfsdk:"service_id"`
	ServiceName        types.String    `tfsdk:"service_name"`
	ServiceNetworkArn  types.String    `tfsdk:"service_network_arn"`
	ServiceNetworkId   types.String    `tfsdk:"service_network_id"`
	ServiceNetworkName types.String    `tfsdk:"service_network_name"`
	Status             types.String    `tfsdk:"status"`
}
