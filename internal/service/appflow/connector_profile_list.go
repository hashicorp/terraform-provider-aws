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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_appflow_connector_profile")
func newConnectorProfileResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceConnectorProfile(), framework.SDKListSpec[awstypes.ConnectorProfile, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.ConnectorProfile, error], diag.Diagnostics) {
			var input appflow.DescribeConnectorProfilesInput

			return listConnectorProfiles(ctx, c.AppFlowClient(ctx), &input), nil
		},
		ID: func(cp awstypes.ConnectorProfile, _ framework.RegionQuery) string {
			return aws.ToString(cp.ConnectorProfileName)
		},
		Flatten: resourceConnectorProfileFlatten,
		DisplayName: func(cp awstypes.ConnectorProfile) string {
			return aws.ToString(cp.ConnectorProfileName)
		},
	})
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
