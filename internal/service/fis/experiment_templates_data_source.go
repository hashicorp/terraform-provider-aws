// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_fis_experiment_templates", name="Experiment Templates")
func newExperimentTemplatesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &experimentTemplatesDataSource{}, nil
}

type experimentTemplatesDataSource struct {
	framework.DataSourceWithModel[experimentTemplatesDataSourceDataSourceModel]
}

func (d *experimentTemplatesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrIDs: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrTags: tftags.TagsAttribute(),
		},
	}
}

func (d *experimentTemplatesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data experimentTemplatesDataSourceDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().FISClient(ctx)

	input := fis.ListExperimentTemplatesInput{}
	filter := tfslices.PredicateTrue[*awstypes.ExperimentTemplateSummary]()
	if tagsToMatch := tftags.New(ctx, data.Tags); len(tagsToMatch) > 0 {
		filter = func(v *awstypes.ExperimentTemplateSummary) bool {
			return keyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch)
		}
	}

	output, err := findExperimentTemplates(ctx, conn, &input, filter)

	if err != nil {
		response.Diagnostics.AddError("reading FIS Experiment Templates", err.Error())

		return
	}

	data.IDs = fwflex.FlattenFrameworkStringValueListOfString(ctx, tfslices.ApplyToAll(output, func(v awstypes.ExperimentTemplateSummary) string {
		return aws.ToString(v.Id)
	}))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type experimentTemplatesDataSourceDataSourceModel struct {
	framework.WithRegionModel
	IDs  fwtypes.ListOfString `tfsdk:"ids"`
	Tags tftags.Map           `tfsdk:"tags"`
}

func findExperimentTemplates(ctx context.Context, conn *fis.Client, input *fis.ListExperimentTemplatesInput, filter tfslices.Predicate[*awstypes.ExperimentTemplateSummary]) ([]awstypes.ExperimentTemplateSummary, error) {
	var output []awstypes.ExperimentTemplateSummary

	pages := fis.NewListExperimentTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ExperimentTemplates {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
