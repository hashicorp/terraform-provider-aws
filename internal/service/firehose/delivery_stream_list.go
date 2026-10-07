// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package firehose

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_kinesis_firehose_delivery_stream")
func newDeliveryStreamResourceAsListResource() inttypes.ListResourceForSDK {
	l := deliveryStreamListResource{}
	l.SetResourceSchema(resourceDeliveryStream())
	return &l
}

type deliveryStreamListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *deliveryStreamListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.FirehoseClient(ctx)

	var input firehose.ListDeliveryStreamsInput
	stream.Results = func(yield func(list.ListResult) bool) {
		for name, err := range listDeliveryStreamNames(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			// Construct ARN from stream name to avoid per-result finder calls
			// when IncludeResource is not set
			arn := awsClient.RegionalARN(ctx, "firehose", fmt.Sprintf("deliverystream/%s", name))
			ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				s, err := findDeliveryStreamByName(ctx, conn, name)
				if err != nil {
					tflog.Error(ctx, "Reading Kinesis Firehose Delivery Stream", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourceDeliveryStreamFlatten(ctx, rd, s); err != nil {
					tflog.Error(ctx, "Reading Kinesis Firehose Delivery Stream", map[string]any{
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

func listDeliveryStreamNames(ctx context.Context, conn *firehose.Client, input *firehose.ListDeliveryStreamsInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		var stopped bool
		err := listDeliveryStreamsPages(ctx, conn, input, func(page *firehose.ListDeliveryStreamsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}
			for _, name := range page.DeliveryStreamNames {
				if !yield(name, nil) {
					stopped = true
					return false
				}
			}
			return !lastPage
		})
		if !stopped && err != nil {
			yield("", fmt.Errorf("listing Kinesis Firehose Delivery Streams: %w", err))
		}
	}
}
