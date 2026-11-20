// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"

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

// @SDKListResource("aws_ram_principal_association")
func newPrincipalAssociationResourceAsListResource() inttypes.ListResourceForSDK {
	l := principalAssociationListResource{}
	l.SetResourceSchema(resourcePrincipalAssociation())
	return &l
}

var _ list.ListResource = &principalAssociationListResource{}

type principalAssociationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *principalAssociationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().RAMClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ram.GetResourceShareAssociationsInput{
			AssociationType: awstypes.ResourceShareAssociationTypePrincipal,
		}

		for item, err := range listResourceAssociations(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			resourceShareARN := aws.ToString(item.ResourceShareArn)
			principal := aws.ToString(item.AssociatedEntity)

			if item.Status == awstypes.ResourceShareAssociationStatusDisassociated || resourceShareARN == "" || principal == "" {
				continue
			}

			id := createPrincipalAssociationResourceID(resourceShareARN, principal)

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set(names.AttrPrincipal, principal)
			rd.Set("resource_share_arn", resourceShareARN)

			result.DisplayName = principal

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
