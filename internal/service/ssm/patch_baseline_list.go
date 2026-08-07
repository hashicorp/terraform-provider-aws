// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ssm_patch_baseline")
func newPatchBaselineResourceAsListResource() inttypes.ListResourceForSDK {
	l := patchBaselineListResource{}
	l.SetResourceSchema(resourcePatchBaseline())
	return &l
}

type patchBaselineListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type patchBaselineListResourceModel struct {
	framework.WithRegionModel
}

func (l *patchBaselineListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SSMClient(ctx)

	var query patchBaselineListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ssm.DescribePatchBaselinesInput

	tflog.Info(ctx, "Listing SSM Patch Baselines")

	stream.Results = func(yield func(list.ListResult) bool) {
		for baselineIdentity, err := range listPatchBaselines(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			baselineID, err := extractPatchBaselineIDFromARN(aws.ToString(baselineIdentity.BaselineId))
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), baselineID)
			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(baselineID)

			if request.IncludeResource {
				baseline, err := findPatchBaselineByID(ctx, conn, baselineID)
				if err != nil {
					tflog.Error(ctx, "Reading SSM Patch Baseline", map[string]any{
						"error": err,
					})
					continue
				}

				if err := resourcePatchBaselineFlatten(ctx, awsClient, rd, baseline); err != nil {
					tflog.Error(ctx, "Flattening SSM Patch Baseline", map[string]any{
						"error": err,
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(baselineIdentity.BaselineName)

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

func listPatchBaselines(ctx context.Context, conn *ssm.Client, input *ssm.DescribePatchBaselinesInput) iter.Seq2[awstypes.PatchBaselineIdentity, error] {
	return func(yield func(awstypes.PatchBaselineIdentity, error) bool) {
		pages := ssm.NewDescribePatchBaselinesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.PatchBaselineIdentity{}, fmt.Errorf("listing SSM Patch Baselines: %w", err))
				return
			}

			for _, baseline := range page.BaselineIdentities {
				if !yield(baseline, nil) {
					return
				}
			}
		}
	}
}

// extractPatchBaselineIDFromARN extracts the baseline ID (pb-xxx) from an ARN or returns the input if it's already a baseline ID
func extractPatchBaselineIDFromARN(arnOrID string) (string, error) {
	if arnOrID == "" {
		return "", errors.New("empty string")
	}
	if !arn.IsARN(arnOrID) {
		return arnOrID, nil
	}

	arnParts, err := arn.Parse(arnOrID)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(arnParts.Resource, "patchbaseline/") {
		return "", fmt.Errorf("arn resource %s is not a patchbaseline", arnOrID)
	}

	return strings.TrimPrefix(arnParts.Resource, "patchbaseline/"), nil
}
