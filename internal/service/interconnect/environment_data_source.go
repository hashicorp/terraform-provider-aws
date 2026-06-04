// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/interconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/interconnect/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_interconnect_environment", name="Environment")
func newEnvironmentDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &environmentDataSource{}, nil
}

type environmentDataSource struct {
	framework.DataSourceWithModel[environmentDataSourceModel]
}

func (d *environmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"activation_page_url": schema.StringAttribute{
				Computed: true,
			},
			"environment_id": schema.StringAttribute{
				Required: true,
			},
			"interconnect_provider": schema.StringAttribute{
				Computed: true,
			},
			names.AttrLocation: schema.StringAttribute{
				Computed: true,
			},
			"remote_identifier_type": schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.RemoteAccountIdentifierType](),
			},
			names.AttrState: schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.EnvironmentState](),
			},
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
			"bandwidths": framework.DataSourceComputedListOfObjectAttribute[bandwidthsModel](ctx),
		},
	}
}

func (d *environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().InterconnectClient(ctx)

	var data environmentDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEnvironmentByID(ctx, conn, data.EnvironmentID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.EnvironmentID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}
	data.InterconnectProvider = flattenProvider(out.Provider)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findEnvironmentByID(ctx context.Context, conn *interconnect.Client, id string) (*awstypes.Environment, error) {
	input := interconnect.GetEnvironmentInput{
		Id: aws.String(id),
	}

	out, err := conn.GetEnvironment(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.Environment == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.Environment, nil
}

type environmentDataSourceModel struct {
	framework.WithRegionModel
	ActivationPageURL    types.String                                             `tfsdk:"activation_page_url"`
	Bandwidths           fwtypes.ListNestedObjectValueOf[bandwidthsModel]         `tfsdk:"bandwidths"`
	EnvironmentID        types.String                                             `tfsdk:"environment_id"`
	InterconnectProvider types.String                                             `tfsdk:"interconnect_provider" autoflex:"-"`
	Location             types.String                                             `tfsdk:"location"`
	RemoteIdentifierType fwtypes.StringEnum[awstypes.RemoteAccountIdentifierType] `tfsdk:"remote_identifier_type"`
	State                fwtypes.StringEnum[awstypes.EnvironmentState]            `tfsdk:"state"`
	Type                 types.String                                             `tfsdk:"type"`
}

type bandwidthsModel struct {
	Available fwtypes.ListOfString `tfsdk:"available"`
	Supported fwtypes.ListOfString `tfsdk:"supported"`
}
