// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ecs_service")
func newServiceResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceService{}
	l.SetResourceSchema(resourceService())
	return &l
}

var _ list.ListResource = &listResourceService{}

type listResourceService struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceService) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"cluster": listschema.StringAttribute{
				Required:    true,
				Description: `The name of the ECS cluster`,
			},
			"launch_type": listschema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.LaunchType](),
				Optional:    true,
				Description: `The launch type to use when filtering the ListServices results.`,
			},
		},
	}
}

func (l *listResourceService) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ECSClient(ctx)

	var query listServiceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ecs.ListServicesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing ECS (Elastic Container) Service")
	stream.Results = func(yield func(list.ListResult) bool) {
		for serviceArn, err := range listServices(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), serviceArn)

			cluster := query.Cluster.ValueString()
			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(serviceArn)
			rd.Set("cluster", cluster)

			tflog.Info(ctx, "Reading ECS (Elastic Container) Service")
			service, err := findServiceByTwoPartKey(ctx, conn, serviceArn, cluster)
			if err != nil {
				tflog.Error(ctx, "Reading ECS (Elastic Container) Service", map[string]any{
					"err": err.Error(),
				})
				continue
			}

			if status := aws.ToString(service.Status); status != serviceStatusActive {
				continue
			}

			diags := resourceServiceFlatten(ctx, rd, service, cluster)
			if diags.HasError() {
				tflog.Error(ctx, "Flatten ECS (Elastic Container) Service", map[string]any{
					"diags": sdkdiag.DiagnosticsString(diags),
				})
				continue
			}

			result.DisplayName = aws.ToString(service.ServiceName)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
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

type listServiceModel struct {
	framework.WithRegionModel
	Cluster    types.String                            `tfsdk:"cluster"`
	LaunchType fwtypes.StringEnum[awstypes.LaunchType] `tfsdk:"launch_type"`
}

func listServices(ctx context.Context, conn *ecs.Client, input *ecs.ListServicesInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := ecs.NewListServicesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield("", fmt.Errorf("listing ECS (Elastic Container) Service resources: %w", err))
				return
			}

			for _, arnStr := range page.ServiceArns {
				if !yield(arnStr, nil) {
					return
				}
			}
		}
	}
}
