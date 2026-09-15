// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_ssoadmin_region")
func newRegionResourceAsListResource() list.ListResourceWithConfigure {
	return &regionListResource{}
}

var _ list.ListResource = &regionListResource{}

type regionListResource struct {
	regionResource
	framework.WithList
}

func (l *regionListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"instance_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the IAM Identity Center instance to list Regions from.",
			},
		},
	}
}

func (l *regionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SSOAdminClient(ctx)

	var query listRegionModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	instanceARN := query.InstanceARN.ValueString()

	tflog.Info(ctx, "Listing SSO Admin Regions", map[string]any{
		"instance_arn": instanceARN,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := ssoadmin.ListRegionsInput{
			InstanceArn: aws.String(instanceARN),
		}

		for item, err := range listRegions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			regionName := aws.ToString(item.RegionName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("region_name"), regionName)

			result := request.NewListResult(ctx)

			var data regionResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.InstanceARN = fwtypes.ARNValue(instanceARN)
				data.RegionName = fwflex.StringValueToFramework(ctx, regionName)

				if request.IncludeResource {
					result.Diagnostics.Append(fwflex.Flatten(ctx, &item, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = regionName
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listRegionModel struct {
	framework.WithRegionModel
	InstanceARN fwtypes.ARN `tfsdk:"instance_arn"`
}

func listRegions(ctx context.Context, conn *ssoadmin.Client, input *ssoadmin.ListRegionsInput) iter.Seq2[awstypes.RegionMetadata, error] {
	return func(yield func(awstypes.RegionMetadata, error) bool) {
		pages := ssoadmin.NewListRegionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.RegionMetadata{}, fmt.Errorf("listing SSO Admin Region resources: %w", err))
				return
			}

			for _, item := range page.Regions {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
