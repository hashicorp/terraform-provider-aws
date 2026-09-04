// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_odb_gi_minor_versions", name="GI Minor Versions")
func newDataSourceGIMinorVersions(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &giMinorVersionsDataSource{}, nil
}

type giMinorVersionsDataSource struct {
	framework.DataSourceWithModel[giMinorVersionsDataSourceModel]
}

func (d *giMinorVersionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone": schema.StringAttribute{
				Optional:    true,
				Description: "Availability Zone to filter GI minor versions.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"availability_zone_id": schema.StringAttribute{
				Optional:    true,
				Description: "Availability Zone ID to filter GI minor versions.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"gi_minor_versions": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewListNestedObjectTypeOf[giMinorVersionSummaryModel](ctx),
				Description: "Available GI minor versions and their Grid Infrastructure software image IDs.",
			},
			"gi_version": schema.StringAttribute{
				Required:    true,
				Description: "GI major version.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"shape_family": schema.StringAttribute{
				Required:    true,
				Description: "Shape family for the GI minor versions.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
		},
	}
}

func (d *giMinorVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)

	var data giMinorVersionsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.ListGiMinorVersionsInput{
		GiVersion:   data.GIVersion.ValueStringPointer(),
		ShapeFamily: data.ShapeFamily.ValueStringPointer(),
	}
	if !data.AvailabilityZone.IsNull() {
		input.AvailabilityZone = data.AvailabilityZone.ValueStringPointer()
	}
	if !data.AvailabilityZoneID.IsNull() {
		input.AvailabilityZoneId = data.AvailabilityZoneID.ValueStringPointer()
	}

	var output odb.ListGiMinorVersionsOutput
	paginator := odb.NewListGiMinorVersionsPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		if page == nil {
			break
		}

		output.GiMinorVersions = append(output.GiMinorVersions, page.GiMinorVersions...)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type giMinorVersionsDataSourceModel struct {
	framework.WithRegionModel
	AvailabilityZone   types.String                                                `tfsdk:"availability_zone"`
	AvailabilityZoneID types.String                                                `tfsdk:"availability_zone_id"`
	GiMinorVersions    fwtypes.ListNestedObjectValueOf[giMinorVersionSummaryModel] `tfsdk:"gi_minor_versions"`
	GIVersion          types.String                                                `tfsdk:"gi_version"`
	ShapeFamily        types.String                                                `tfsdk:"shape_family"`
}

type giMinorVersionSummaryModel struct {
	GridImageID types.String `tfsdk:"grid_image_id"`
	Version     types.String `tfsdk:"version"`
}
