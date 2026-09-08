// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ram_resource_association")
func newResourceAssociationResourceAsListResource() inttypes.ListResourceForSDK {
	l := resourceAssociationListResource{}
	l.SetResourceSchema(resourceResourceAssociation())
	return &l
}

var _ list.ListResource = &resourceAssociationListResource{}

type resourceAssociationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *resourceAssociationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().RAMClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ram.GetResourceShareAssociationsInput{
			AssociationType: awstypes.ResourceShareAssociationTypeResource,
		}

		for item, err := range listResourceAssociations(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			resourceShareARN := aws.ToString(item.ResourceShareArn)
			resourceARN := aws.ToString(item.AssociatedEntity)

			if resourceShareARN == "" || resourceARN == "" {
				continue
			}

			id := createResourceAssociationResourceID(resourceShareARN, resourceARN)

			itemCtx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)
			itemCtx = tflog.SetField(itemCtx, names.AttrResourceARN, resourceARN)
			itemCtx = tflog.SetField(itemCtx, "resource_share_arn", resourceShareARN)

			result := request.NewListResult(itemCtx)

			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set(names.AttrResourceARN, resourceARN)
			rd.Set("resource_share_arn", resourceShareARN)

			result.DisplayName = resourceARN

			l.SetResult(itemCtx, l.Meta(), request.IncludeResource, rd, &result)
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

func listResourceAssociations(ctx context.Context, conn *ram.Client, input *ram.GetResourceShareAssociationsInput) iter.Seq2[awstypes.ResourceShareAssociation, error] {
	return func(yield func(awstypes.ResourceShareAssociation, error) bool) {
		pages := ram.NewGetResourceShareAssociationsPaginator(conn, input)

		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ResourceShareAssociation{}, fmt.Errorf("listing RAM Resource Associations: %w", err))
				return
			}

			for _, item := range page.ResourceShareAssociations {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
