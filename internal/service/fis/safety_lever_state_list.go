// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_fis_safety_lever_state")
func newSafetyLeverStateResourceAsListResource() list.ListResourceWithConfigure {
	return &safetyLeverStateListResource{}
}

var _ list.ListResource = &safetyLeverStateListResource{}

type safetyLeverStateListResource struct {
	safetyLeverStateResource
	framework.WithList
}

func (l *safetyLeverStateListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().FISClient(ctx)

	tflog.Info(ctx, "Listing FIS Safety Lever States")

	stream.Results = func(yield func(list.ListResult) bool) {
		out, err := findSafetyLever(ctx, conn, safetyLeverDefaultID)
		if retry.NotFound(err) {
			return
		}
		if err != nil {
			yield(fwdiag.NewListResultErrorDiagnostic(err))
			return
		}

		result := request.NewListResult(ctx)
		var data safetyLeverStateResourceModel

		l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
			if request.IncludeResource {
				result.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
				if result.Diagnostics.HasError() {
					return
				}
			}

			result.DisplayName = safetyLeverDefaultID
		})

		yield(result)
	}
}
