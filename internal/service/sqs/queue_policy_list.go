// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_sqs_queue_policy")
func newQueuePolicyResourceAsListResource() inttypes.ListResourceForSDK {
	l := queuePolicyListResource{}
	l.SetResourceSchema(resourceQueuePolicy())
	return &l
}

var _ list.ListResource = &queuePolicyListResource{}

type queuePolicyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listQueuePolicyModel struct {
	framework.WithRegionModel
}

func (l *queuePolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SQSClient(ctx)

	var query listQueuePolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	h := &queueAttributeHandler{
		AttributeName: sqstypes.QueueAttributeNamePolicy,
		SchemaKey:     names.AttrPolicy,
		ToSet:         verify.PolicyToSet,
	}

	tflog.Info(ctx, "Listing SQS queue policies")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input sqs.ListQueuesInput
		for queueURL, err := range listQueues(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("queue_url"), queueURL)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(queueURL)
			rd.Set("queue_url", queueURL)

			policy, err := findQueueAttributeByTwoPartKey(ctx, conn, queueURL, sqstypes.QueueAttributeNamePolicy)
			if err != nil {
				if errors.Is(err, tfresource.ErrEmptyResult) || retry.NotFound(err) {
					continue
				}

				tflog.Error(ctx, "Reading SQS queue policy", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			name, err := queueNameFromURL(queueURL)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			if request.IncludeResource {
				if err := h.flattenResourceData(rd, aws.ToString(policy)); err != nil {
					tflog.Error(ctx, "Reading SQS queue policy", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = name

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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
