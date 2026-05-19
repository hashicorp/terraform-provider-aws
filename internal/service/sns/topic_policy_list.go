// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_sns_topic_policy")
func newTopicPolicyResourceAsListResource() inttypes.ListResourceForSDK {
	l := topicPolicyListResource{}
	l.SetResourceSchema(resourceTopicPolicy())
	return &l
}

var _ list.ListResource = &topicPolicyListResource{}

type topicPolicyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listTopicPolicyModel struct {
	framework.WithRegionModel
}

func (l *topicPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SNSClient(ctx)

	var query listTopicPolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing SNS Topic Policies")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input sns.ListTopicsInput
		for item, err := range listTopics(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			topicARN := aws.ToString(item.TopicArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), topicARN)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(topicARN)
			rd.Set(names.AttrARN, topicARN)

			if request.IncludeResource {
				attributes, err := findTopicAttributesWithValidAWSPrincipalsByARN(ctx, conn, topicARN)
				if err != nil {
					tflog.Error(ctx, "Reading SNS Topic Policy", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if diags := resourceTopicPolicyFlatten(ctx, rd, attributes); diags.HasError() {
					tflog.Error(ctx, "Reading SNS Topic Policy", map[string]any{
						"diags": sdkdiag.DiagnosticsString(diags),
					})
					continue
				}
			}

			name, err := parseTopicNameFromARN(topicARN)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result.DisplayName = name

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
