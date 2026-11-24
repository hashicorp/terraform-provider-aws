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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkListResource("aws_ec2_public_ipv4_pool")
func newPublicIPv4PoolResourceAsListResource() list.ListResourceWithConfigure {
	return &publicIPv4PoolListResource{}
}

var _ list.ListResource = &publicIPv4PoolListResource{}

type publicIPv4PoolListResource struct {
	publicIPv4PoolResource
	framework.WithList
}

func (l *publicIPv4PoolListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	tflog.Info(ctx, "Listing EC2 Public IPv4 Pools")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input ec2.DescribePublicIpv4PoolsInput
		for item, err := range listPublicIPv4Pools(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(item.PoolId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("pool_id"), id)

			var output *awstypes.PublicIpv4Pool
			if request.IncludeResource {
				var err error
				output, err = findPublicIPv4PoolByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)
			tags := keyValueTags(ctx, item.Tags)

			var data publicIPv4PoolResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.PoolID = fwflex.StringValueToFramework(ctx, id)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data))
					if result.Diagnostics.HasError() {
						return
					}
				}

				if v, ok := tags["Name"]; ok {
					result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), id)
				} else {
					result.DisplayName = id
				}
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

func listPublicIPv4Pools(ctx context.Context, conn *ec2.Client, input *ec2.DescribePublicIpv4PoolsInput) iter.Seq2[awstypes.PublicIpv4Pool, error] {
	return func(yield func(awstypes.PublicIpv4Pool, error) bool) {
		pages := ec2.NewDescribePublicIpv4PoolsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.PublicIpv4Pool{}, fmt.Errorf("listing EC2 Public IPv4 Pools: %w", err))
				return
			}

			for _, item := range page.PublicIpv4Pools {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
