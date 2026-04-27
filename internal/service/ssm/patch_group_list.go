// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_ssm_patch_group")
func newPatchGroupResourceAsListResource() inttypes.ListResourceForSDK {
	l := patchGroupListResource{}
	l.SetResourceSchema(resourcePatchGroup())
	return &l
}

var _ list.ListResource = &patchGroupListResource{}

type patchGroupListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *patchGroupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SSMClient(ctx)

	var query listPatchGroupModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		input := ssm.DescribePatchGroupsInput{}
		for item, err := range listPatchGroups(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var groupBaselineID string
			if item.BaselineIdentity != nil {
				groupBaselineID = aws.ToString(item.BaselineIdentity.BaselineId)
			}
			patchGroupName := aws.ToString(item.PatchGroup)

			id, err := flex.FlattenResourceId([]string{patchGroupName, groupBaselineID}, patchGroupResourceIDPartCount, false)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("generating id: %w", err))
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("patch_group"), patchGroupName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set("patch_group", patchGroupName)
			rd.Set("baseline_id", groupBaselineID)

			if request.IncludeResource { //nolint:revive,staticcheck // Be explicit about IncludeResource handling
				// No-op, all attributes already set
			}

			result.DisplayName = patchGroupName

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

type listPatchGroupModel struct {
	framework.WithRegionModel
}

func listPatchGroups(ctx context.Context, conn *ssm.Client, input *ssm.DescribePatchGroupsInput) iter.Seq2[awstypes.PatchGroupPatchBaselineMapping, error] {
	return func(yield func(awstypes.PatchGroupPatchBaselineMapping, error) bool) {
		pages := ssm.NewDescribePatchGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.PatchGroupPatchBaselineMapping{}, fmt.Errorf("listing SSM (Systems Manager) Patch Group resources: %w", err))
				return
			}

			for _, item := range page.Mappings {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
