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
// @SDKListResource("aws_appautoscaling_policy")
func newPolicyResourceAsListResource() inttypes.ListResourceForSDK {
	l := policyListResource{}
	l.SetResourceSchema(resourcePolicy())
	return &l
}

var _ list.ListResource = &policyListResource{}

type policyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listPolicyModel struct {
	framework.WithRegionModel
	ServiceNamespace types.String `tfsdk:"service_namespace"`
}

func (l *policyListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"service_namespace": listschema.StringAttribute{
				Required:    true,
				Description: "Namespace of the AWS service that owns the scalable target.",
			},
		},
	}
}

func (l *policyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().AppAutoScalingClient(ctx)

	var query listPolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	serviceNamespace := query.ServiceNamespace.ValueString()

	tflog.Info(ctx, "Listing Application Auto Scaling Policies", map[string]any{
		logging.ResourceAttributeKey("service_namespace"): serviceNamespace,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := applicationautoscaling.DescribeScalingPoliciesInput{
			ServiceNamespace: awstypes.ServiceNamespace(serviceNamespace),
		}
		for item, err := range listScalingPoliciesForList(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			policyName := aws.ToString(item.PolicyName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), policyName)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(policyName)
			rd.Set(names.AttrName, policyName)
			rd.Set("service_namespace", item.ServiceNamespace)
			rd.Set(names.AttrResourceID, item.ResourceId)
			rd.Set("scalable_dimension", item.ScalableDimension)

			if request.IncludeResource {
				if err := resourcePolicyFlatten(&item, rd); err != nil {
					tflog.Error(ctx, "Reading Application Auto Scaling Policy", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = policyName

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

func listScalingPoliciesForList(ctx context.Context, conn *applicationautoscaling.Client, input *applicationautoscaling.DescribeScalingPoliciesInput) iter.Seq2[awstypes.ScalingPolicy, error] {
	return func(yield func(awstypes.ScalingPolicy, error) bool) {
		pages := applicationautoscaling.NewDescribeScalingPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ScalingPolicy{}, fmt.Errorf("listing Application Auto Scaling Policies: %w", err))
				return
			}

			for _, item := range page.ScalingPolicies {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
