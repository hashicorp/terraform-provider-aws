// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appflow

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appflow"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appflow/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_appflow_connector_profile")
func newConnectorProfileResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceConnectorProfile{}
	l.SetResourceSchema(resourceConnectorProfile())
	return &l
}

var _ list.ListResource = &listResourceConnectorProfile{}

type listResourceConnectorProfile struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceConnectorProfile) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.AppFlowClient(ctx)

	var query listConnectorProfileModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing AppFlow Connector Profile")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input appflow.DescribeConnectorProfilesInput
		for item, err := range listConnectorProfiles(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.ConnectorProfileName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), name)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(name)
			_ = rd.Set(names.AttrName, name)

			tflog.Info(ctx, "Reading AppFlow Connector Profile")
			resourceConnectorProfileFlatten(ctx, &item, rd)

			result.DisplayName = aws.ToString(item.ConnectorProfileName)

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
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

type listConnectorProfileModel struct {
	framework.WithRegionModel
}

func listConnectorProfiles(ctx context.Context, conn *appflow.Client, input *appflow.DescribeConnectorProfilesInput) iter.Seq2[awstypes.ConnectorProfile, error] {
	return func(yield func(awstypes.ConnectorProfile, error) bool) {
		pages := appflow.NewDescribeConnectorProfilesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ConnectorProfile{}, fmt.Errorf("listing AppFlow Connector Profile resources: %w", err))
				return
			}

			for _, item := range page.ConnectorProfileDetails {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
