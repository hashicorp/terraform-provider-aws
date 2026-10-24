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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_ec2_application_status_check_association")
func newApplicationStatusCheckAssociationResourceAsListResource() list.ListResourceWithConfigure {
	return &applicationStatusCheckAssociationListResource{}
}

var _ list.ListResource = &applicationStatusCheckAssociationListResource{}

type applicationStatusCheckAssociationListResource struct {
	applicationStatusCheckAssociationResource
	framework.WithList
}

func (l *applicationStatusCheckAssociationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EC2Client(ctx)

	tflog.Info(ctx, "Listing EC2 Application Status Check Associations")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input ec2.DescribeApplicationStatusCheckAssociationsInput
		for item, err := range listApplicationStatusCheckAssociations(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			checkID := aws.ToString(item.ApplicationStatusCheckId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("application_status_check_id"), checkID)
			switch item.AssociationType {
			case awstypes.AssociationTypeEnumInstanceId:
				ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrInstanceID), aws.ToString(item.Value))
			case awstypes.AssociationTypeEnumTag:
				ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("target_tag_key"), aws.ToString(item.Key))
				ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("target_tag_value"), aws.ToString(item.Value))
			}
			result := request.NewListResult(ctx)
			var data applicationStatusCheckAssociationResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				flattenApplicationStatusCheckAssociation(&item, &data)
				result.DisplayName = applicationStatusCheckAssociationImportIDString(checkID, data.InstanceID.ValueString(), data.TargetTagKey.ValueString(), data.TargetTagValue.ValueString())
			})
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

func listApplicationStatusCheckAssociations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeApplicationStatusCheckAssociationsInput) iter.Seq2[awstypes.ApplicationStatusCheckAssociationObject, error] {
	return func(yield func(awstypes.ApplicationStatusCheckAssociationObject, error) bool) {
		err := describeApplicationStatusCheckAssociationsPages(ctx, conn, input, func(page *ec2.DescribeApplicationStatusCheckAssociationsOutput, _ bool) bool {
			for _, item := range page.Associations {
				if !yield(item, nil) {
					return false
				}
			}
			return true
		})
		if err != nil {
			yield(awstypes.ApplicationStatusCheckAssociationObject{}, fmt.Errorf("listing EC2 Application Status Check Associations: %w", err))
		}
	}
}
