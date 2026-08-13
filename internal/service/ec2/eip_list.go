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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_eip")
func newEIPResourceAsListResource() inttypes.ListResourceForSDK {
	l := eipListResource{}
	l.SetResourceSchema(resourceEIP())

	return &l
}

var _ list.ListResource = &eipListResource{}

type eipListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *eipListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listEIPs(ctx, conn, &ec2.DescribeAddressesInput{}) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.AllocationId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			setTagsOut(ctx, item.Tags)

			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				resourceEIPFlatten(ctx, l.Meta(), &item, rd)
			}

			result.DisplayName = id

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

func listEIPs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAddressesInput) iter.Seq2[awstypes.Address, error] {
	return func(yield func(awstypes.Address, error) bool) {
		items, err := findEIPs(ctx, conn, input)
		if err != nil {
			yield(awstypes.Address{}, fmt.Errorf("listing EC2 EIPs: %w", err))
			return
		}

		for _, item := range items {
			if !yield(item, nil) {
				return
			}
		}
	}
}
