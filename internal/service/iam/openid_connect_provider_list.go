// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_iam_openid_connect_provider")
func newOpenIDConnectProviderResourceAsListResource() inttypes.ListResourceForSDK {
	l := openIDConnectProviderListResource{}
	l.SetResourceSchema(resourceOpenIDConnectProvider())

	return &l
}

var _ list.ListResource = &openIDConnectProviderListResource{}

type openIDConnectProviderListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *openIDConnectProviderListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	tflog.Info(ctx, "Listing IAM OIDC Providers")
	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listOpenIDConnectProviders(ctx, conn, &iam.ListOpenIDConnectProvidersInput{}) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.Arn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				tflog.Info(ctx, "Reading IAM OIDC Provider")
				provider, err := findOpenIDConnectProviderByARN(ctx, conn, arn)
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}

				resourceOpenIDConnectProviderFlatten(ctx, provider, rd)
			}

			displayName, err := urlFromOpenIDConnectProviderARN(arn)
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}
			result.DisplayName = displayName

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

func listOpenIDConnectProviders(ctx context.Context, conn *iam.Client, input *iam.ListOpenIDConnectProvidersInput) iter.Seq2[awstypes.OpenIDConnectProviderListEntry, error] {
	return func(yield func(awstypes.OpenIDConnectProviderListEntry, error) bool) {
		output, err := conn.ListOpenIDConnectProviders(ctx, input)
		if err != nil {
			yield(awstypes.OpenIDConnectProviderListEntry{}, fmt.Errorf("listing IAM OIDC providers: %w", err))
			return
		}

		for _, item := range output.OpenIDConnectProviderList {
			if !yield(item, nil) {
				return
			}
		}
	}
}
