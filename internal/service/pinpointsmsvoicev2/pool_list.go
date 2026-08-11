// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_pinpointsmsvoicev2_pool")
func newPoolResourceAsListResource() list.ListResourceWithConfigure {
	return &poolListResource{}
}

var _ list.ListResource = &poolListResource{}

type poolListResource struct {
	poolResource
	framework.WithList
}

func (l *poolListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().PinpointSMSVoiceV2Client(ctx)

	var query listPoolModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing End User Messaging SMS Pools")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := pinpointsmsvoicev2.DescribePoolsInput{}
		pools, err := findPools(ctx, conn, &input)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		for i := range pools {
			pool := &pools[i]
			poolID := aws.ToString(pool.PoolId)

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), poolID)

			result := request.NewListResult(ctx)

			var data poolResourceModel
			skipPool := false
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, fwflex.Flatten(ctx, pool, &data))
				if result.Diagnostics.HasError() {
					return
				}

				associated, err := findPoolOriginationIdentities(ctx, conn, poolID)
				if retry.NotFound(err) {
					tflog.Debug(ctx, "Pool deleted concurrently; skipping in list enumeration")
					skipPool = true
					return
				}
				if err != nil {
					smerr.AddError(ctx, &result.Diagnostics, err, smerr.ID, poolID)
					return
				}
				data.OriginationIdentities = fwflex.FlattenFrameworkStringValueSetOfString(ctx, associated)

				result.DisplayName = poolID
			})

			if skipPool {
				continue
			}
			if !yield(result) {
				return
			}
		}
	}
}

type listPoolModel struct {
	framework.WithRegionModel
}
