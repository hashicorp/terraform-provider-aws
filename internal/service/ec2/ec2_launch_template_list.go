// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_launch_template")
func newLaunchTemplateResourceAsListResource() inttypes.ListResourceForSDK {
	l := launchTemplateListResource{}
	l.SetResourceSchema(resourceLaunchTemplate())
	return &l
}

var _ list.ListResource = &launchTemplateListResource{}

type launchTemplateListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listLaunchTemplateModel struct {
	framework.WithRegionModel
	LaunchTemplateIDs   fwtypes.ListValueOf[types.String] `tfsdk:"launch_template_ids"`
	LaunchTemplateNames fwtypes.ListValueOf[types.String] `tfsdk:"launch_template_names"`
	Filters             customListFilters                 `tfsdk:"filter"`
}

func (l *launchTemplateListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"launch_template_ids": listschema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			"launch_template_names": listschema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
		},
		Blocks: map[string]listschema.Block{
			names.AttrFilter: customListFiltersBlock(ctx),
		},
	}
}

func (l *launchTemplateListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	var query listLaunchTemplateModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeLaunchTemplatesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listLaunchTemplates(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.LaunchTemplateId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			lt, err := findLaunchTemplateByID(ctx, conn, id)
			if err != nil {
				tflog.Error(ctx, "Reading EC2 Launch Template", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			version := flex.Int64ToStringValue(lt.LatestVersionNumber)
			ltv, err := findLaunchTemplateVersionByTwoPartKey(ctx, conn, id, version)
			if err != nil {
				tflog.Error(ctx, "Reading EC2 Launch Template Version", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := resourceLaunchTemplateFlatten(ctx, conn, l.Meta(), lt, ltv, rd); err != nil {
					tflog.Error(ctx, "Flattening EC2 Launch Template", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.LaunchTemplateName)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listLaunchTemplates(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLaunchTemplatesInput) iter.Seq2[awstypes.LaunchTemplate, error] {
	return func(yield func(awstypes.LaunchTemplate, error) bool) {
		pages := ec2.NewDescribeLaunchTemplatesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.LaunchTemplate](), fmt.Errorf("listing EC2 Launch Template resources: %w", err))
				return
			}

			for _, item := range page.LaunchTemplates {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
