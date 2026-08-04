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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ssm_association", name="Association")
func newAssociationResourceAsListResource() inttypes.ListResourceForSDK {
	l := associationListResource{}
	l.SetResourceSchema(resourceAssociation())
	return &l
}

var _ list.ListResource = &associationListResource{}

type associationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *associationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SSMClient(ctx)

	var query listAssociationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Resources", map[string]any{})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := ssm.ListAssociationsInput{}
		for item, err := range listAssociations(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}
			id := aws.ToString(item.AssociationId)
			name := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set(names.AttrAssociationID, id)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				out, err := findAssociationByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					result = fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading SSM Association (%s): %w", name, err))
					yield(result)
					return
				}

				if err := resourceAssociationFlatten(ctx, l.Meta(), out, rd); err != nil {
					tflog.Error(ctx, "Flatten SSM Association", map[string]any{
						"error": err.Error(),
					})
					continue
				}
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

type listAssociationModel struct {
	framework.WithRegionModel
}

func listAssociations(ctx context.Context, conn *ssm.Client, input *ssm.ListAssociationsInput) iter.Seq2[awstypes.Association, error] {
	return func(yield func(awstypes.Association, error) bool) {
		pages := ssm.NewListAssociationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Association{}, fmt.Errorf("listing SSM (Systems Manager) Association resources: %w", err))
				return
			}

			for _, item := range page.Associations {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
