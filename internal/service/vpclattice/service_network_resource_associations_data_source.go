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
// @FrameworkDataSource("aws_vpclattice_service_network_resource_associations", name="Service Network Resource Associations")
func newDataSourceServiceNetworkResourceAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServiceNetworkResourceAssociations{}, nil
}

const (
	DSNameServiceNetworkResourceAssociations = "Service Network Resource Associations Data Source"
)

type dataSourceServiceNetworkResourceAssociations struct {
	framework.DataSourceWithModel[dataSourceServiceNetworkResourceAssociationsModel]
}

func (d *dataSourceServiceNetworkResourceAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_network_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "ID or ARN of the Service Network.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("resource_configuration_identifier"),
					),
				},
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
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"associations": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN:                 framework.ARNAttributeComputedOnly(),
						"created_by":                  schema.StringAttribute{Computed: true},
						"failure_code":                schema.StringAttribute{Computed: true},
						names.AttrID:                  schema.StringAttribute{Computed: true},
						"is_managed_association":      schema.BoolAttribute{Computed: true},
						"resource_configuration_arn":  schema.StringAttribute{Computed: true},
						"resource_configuration_id":   schema.StringAttribute{Computed: true},
						"resource_configuration_name": schema.StringAttribute{Computed: true},
						"service_network_arn":         schema.StringAttribute{Computed: true},
						"service_network_id":          schema.StringAttribute{Computed: true},
						"service_network_name":        schema.StringAttribute{Computed: true},
						names.AttrStatus:              schema.StringAttribute{Computed: true},
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
						"private_dns_entry": schema.ListNestedBlock{
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

func (d *dataSourceServiceNetworkResourceAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().VPCLatticeClient(ctx)

	var associations []*awstypes.ServiceNetworkResourceAssociationSummary

	var data dataSourceServiceNetworkResourceAssociationsModel

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

		items, err := listServiceNetworkResourceAssociationsByServiceNetworkIdentifier(ctx, conn, snID)
		if err != nil {
			resp.Diagnostics.AddError("Error listing associations", err.Error())
			return
		}
		associations = items
	} else if !data.ResourceConfigurationIdentifier.IsNull() {
		sID := data.ResourceConfigurationIdentifier.ValueString()
		data.ID = types.StringValue(sID)

		if _, err := findResourceConfigurationByID(ctx, conn, sID); err != nil {
			resp.Diagnostics.AddError("Error reading Resource Configuration", err.Error())
			return
		}

		items, err := listServiceNetworkResourceAssociationsByResourceConfigurationIdentifier(ctx, conn, sID)
		if err != nil {
			resp.Diagnostics.AddError("Error listing associations", err.Error())
			return
		}
		associations = items
	}

	data.Associations = flattenServiceNetworkResourceAssociations(associations)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenServiceNetworkResourceAssociations(objs []*awstypes.ServiceNetworkResourceAssociationSummary) []associationModel {
	result := make([]associationModel, 0, len(objs))
	for _, obj := range objs {
		result = append(result, flattenServiceNetworkResourceAssociation(obj))
	}
	return result
}

func flattenServiceNetworkResourceAssociation(obj *awstypes.ServiceNetworkResourceAssociationSummary) associationModel {
	var dnsEntries []dnsEntryModel
	var privateDnsEntries []dnsEntryModel
	if obj.DnsEntry != nil {
		dnsEntries = []dnsEntryModel{flattenDNSEntryFramework(obj.DnsEntry)}
	}
	if obj.PrivateDnsEntry != nil {
		privateDnsEntries = []dnsEntryModel{flattenDNSEntryFramework(obj.PrivateDnsEntry)}
	}

	return associationModel{
		Arn:                       types.StringPointerValue(obj.Arn),
		CreatedBy:                 types.StringPointerValue(obj.CreatedBy),
		DnsEntry:                  dnsEntries,
		FailureCode:               types.StringPointerValue(obj.FailureCode),
		ID:                        types.StringPointerValue(obj.Id),
		IsManagedAssociation:      types.BoolPointerValue(obj.IsManagedAssociation),
		PrivateDnsEntry:           privateDnsEntries,
		ResourceConfigurationArn:  types.StringPointerValue(obj.ResourceConfigurationArn),
		ResourceConfigurationId:   types.StringPointerValue(obj.ResourceConfigurationId),
		ResourceConfigurationName: types.StringPointerValue(obj.ResourceConfigurationName),
		ServiceNetworkArn:         types.StringPointerValue(obj.ServiceNetworkArn),
		ServiceNetworkId:          types.StringPointerValue(obj.ServiceNetworkId),
		ServiceNetworkName:        types.StringPointerValue(obj.ServiceNetworkName),
		Status:                    types.StringValue(string(obj.Status)),
	}
}

func flattenDNSEntryFramework(obj *awstypes.DnsEntry) dnsEntryModel {
	return dnsEntryModel{
		DomainName:   types.StringPointerValue(obj.DomainName),
		HostedZoneID: types.StringPointerValue(obj.HostedZoneId),
	}
}

func listServiceNetworkResourceAssociationsByResourceConfigurationIdentifier(ctx context.Context, conn *vpclattice.Client, resourceConfigurationIdentifier string) ([]*awstypes.ServiceNetworkResourceAssociationSummary, error) {
	input := vpclattice.ListServiceNetworkResourceAssociationsInput{
		ResourceConfigurationIdentifier: aws.String(resourceConfigurationIdentifier),
	}
	return listServiceNetworkResourceAssociations(ctx, conn, &input)
}

func listServiceNetworkResourceAssociationsByServiceNetworkIdentifier(ctx context.Context, conn *vpclattice.Client, serviceNetworkIdentifier string) ([]*awstypes.ServiceNetworkResourceAssociationSummary, error) {
	input := vpclattice.ListServiceNetworkResourceAssociationsInput{
		ServiceNetworkIdentifier: aws.String(serviceNetworkIdentifier),
	}
	return listServiceNetworkResourceAssociations(ctx, conn, &input)
}

func listServiceNetworkResourceAssociations(ctx context.Context, conn *vpclattice.Client, input *vpclattice.ListServiceNetworkResourceAssociationsInput) ([]*awstypes.ServiceNetworkResourceAssociationSummary, error) {
	var output []*awstypes.ServiceNetworkResourceAssociationSummary
	paginator := vpclattice.NewListServiceNetworkResourceAssociationsPaginator(conn, input, func(opts *vpclattice.ListServiceNetworkResourceAssociationsPaginatorOptions) {
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

type dataSourceServiceNetworkResourceAssociationsModel struct {
	framework.WithRegionModel
	ServiceNetworkIdentifier        types.String       `tfsdk:"service_network_identifier"`
	ResourceConfigurationIdentifier types.String       `tfsdk:"resource_configuration_identifier"`
	Associations                    []associationModel `tfsdk:"associations"`
	ID                              types.String       `tfsdk:"id"`
}

type associationModel struct {
	Arn                       types.String    `tfsdk:"arn"`
	CreatedBy                 types.String    `tfsdk:"created_by"`
	DnsEntry                  []dnsEntryModel `tfsdk:"dns_entry"`
	FailureCode               types.String    `tfsdk:"failure_code"`
	ID                        types.String    `tfsdk:"id"`
	IsManagedAssociation      types.Bool      `tfsdk:"is_managed_association"`
	PrivateDnsEntry           []dnsEntryModel `tfsdk:"private_dns_entry"`
	ResourceConfigurationArn  types.String    `tfsdk:"resource_configuration_arn"`
	ResourceConfigurationId   types.String    `tfsdk:"resource_configuration_id"`
	ResourceConfigurationName types.String    `tfsdk:"resource_configuration_name"`
	ServiceNetworkArn         types.String    `tfsdk:"service_network_arn"`
	ServiceNetworkId          types.String    `tfsdk:"service_network_id"`
	ServiceNetworkName        types.String    `tfsdk:"service_network_name"`
	Status                    types.String    `tfsdk:"status"`
}
