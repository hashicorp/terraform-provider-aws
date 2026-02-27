// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workmail/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_workmail_organization")
func newOrganizationResourceAsListResource() list.ListResourceWithConfigure {
	return &organizationListResource{}
}

var _ list.ListResource = &organizationListResource{}

type organizationListResource struct {
	organizationResource
	framework.WithList
}

func (r *organizationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().WorkMailClient(ctx)

	var query listOrganizationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input workmail.ListOrganizationsInput
		for orgSummary, err := range listOrganizations(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			// Skip deleted organizations
			if aws.ToString(orgSummary.State) == statusDeleted {
				continue
			}

			orgID := aws.ToString(orgSummary.OrganizationId)
			out, err := findOrganizationByID(ctx, conn, orgID)
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data organizationResourceModel

			r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
				if diags := fwflex.Flatten(ctx, out, &data); diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				data.OrganizationAlias = fwflex.StringToFramework(ctx, out.Alias)
				result.DisplayName = orgID
			})

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listOrganizationModel struct {
	framework.WithRegionModel
}

func listOrganizations(ctx context.Context, conn *workmail.Client, input *workmail.ListOrganizationsInput) iter.Seq2[awstypes.OrganizationSummary, error] {
	return func(yield func(awstypes.OrganizationSummary, error) bool) {
		pages := workmail.NewListOrganizationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.OrganizationSummary{}, fmt.Errorf("listing WorkMail Organization resources: %w", err))
				return
			}

			for _, item := range page.OrganizationSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
