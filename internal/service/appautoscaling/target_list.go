// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appautoscaling

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_appautoscaling_target")
func newTargetResourceAsListResource() inttypes.ListResourceForSDK {
	l := targetListResource{}
	l.SetResourceSchema(resourceTarget())
	return &l
}

var _ list.ListResource = &targetListResource{}

type targetListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listTargetModel struct {
	framework.WithRegionModel
	ServiceNamespace types.String `tfsdk:"service_namespace"`
}

func (l *targetListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"service_namespace": listschema.StringAttribute{
				Required:    true,
				Description: "Namespace of the AWS service that owns the scalable target.",
			},
		},
	}
}

func (l *targetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().AppAutoScalingClient(ctx)

	var query listTargetModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	serviceNamespace := query.ServiceNamespace.ValueString()

	tflog.Info(ctx, "Listing Application Auto Scaling Targets", map[string]any{
		logging.ResourceAttributeKey("service_namespace"): serviceNamespace,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := applicationautoscaling.DescribeScalableTargetsInput{
			ServiceNamespace: awstypes.ServiceNamespace(serviceNamespace),
		}
		for item, err := range listScalableTargetsForList(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			resourceID := aws.ToString(item.ResourceId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrResourceID), resourceID)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(resourceID)
			rd.Set("service_namespace", item.ServiceNamespace)
			rd.Set(names.AttrResourceID, item.ResourceId)
			rd.Set("scalable_dimension", item.ScalableDimension)

			if request.IncludeResource {
				if err := resourceTargetFlatten(&item, rd); err != nil {
					tflog.Error(ctx, "Reading Application Auto Scaling Target", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = resourceID

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

func listScalableTargetsForList(ctx context.Context, conn *applicationautoscaling.Client, input *applicationautoscaling.DescribeScalableTargetsInput) iter.Seq2[awstypes.ScalableTarget, error] {
	return func(yield func(awstypes.ScalableTarget, error) bool) {
		pages := applicationautoscaling.NewDescribeScalableTargetsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ScalableTarget{}, fmt.Errorf("listing Application Auto Scaling Scalable Targets: %w", err))
				return
			}

			for _, item := range page.ScalableTargets {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
