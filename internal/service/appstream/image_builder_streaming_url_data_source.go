// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/appstream"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_appstream_image_builder_streaming_url", name="Image Builder Streaming URL")
func newImageBuilderStreamingURLDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &imageBuilderStreamingURLDataSource{}, nil
}

type imageBuilderStreamingURLDataSource struct {
	framework.DataSourceWithModel[imageBuilderStreamingURLDataSourceModel]
}

func (d *imageBuilderStreamingURLDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for creating an AppStream Image Builder Streaming URL.",
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of the image builder.",
				Required:    true,
			},
			"streaming_url": schema.StringAttribute{
				Description: "Streaming URL for the image builder.",
				Computed:    true,
			},
			"expires": schema.StringAttribute{
				Description: "Time when the streaming URL expires.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"validity": schema.Int64Attribute{
				Description: "Duration (in seconds) for which the streaming URL will be valid. Must be a value from 1 to 604800 (1 week). Defaults to 3600 (1 hour).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 604800),
				},
			},
		},
	}
}
func (d *imageBuilderStreamingURLDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().AppStreamClient(ctx)

	var data imageBuilderStreamingURLDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := &appstream.CreateImageBuilderStreamingURLInput{
		Name: data.Name.ValueStringPointer(),
	}

	// Set validity if provided, otherwise use default of 3600 (1 hour)
	validity := int64(3600)
	if !data.Validity.IsNull() {
		validity = data.Validity.ValueInt64()
	}
	input.Validity = &validity

	output, err := conn.CreateImageBuilderStreamingURL(ctx, input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	data.ID = data.Name
	data.StreamingURL = types.StringPointerValue(output.StreamingURL)
	if output.Expires != nil {
		data.Expires = timetypes.NewRFC3339TimePointerValue(output.Expires)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.Name.String())
}

type imageBuilderStreamingURLDataSourceModel struct {
	Expires      timetypes.RFC3339 `tfsdk:"expires"`
	ID           types.String      `tfsdk:"id"`
	Name         types.String      `tfsdk:"name"`
	StreamingURL types.String      `tfsdk:"streaming_url"`
	Validity     types.Int64       `tfsdk:"validity"`
}
