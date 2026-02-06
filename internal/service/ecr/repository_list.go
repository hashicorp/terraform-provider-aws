// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ecr_repository")
func repositoryResourceAsListResource() inttypes.ListResourceForSDK {
	l := repositoryListResource{}
	l.SetResourceSchema(resourceRepository())
	return &l
}

type repositoryListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type repositoryListResourceModel struct {
	framework.WithRegionModel
}

func (l *repositoryListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query repositoryListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	awsClient := l.Meta()
	conn := awsClient.ECRClient(ctx)

	tflog.Info(ctx, "Listing ECR repositories")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ecr.DescribeRepositoriesInput{}

		paginator := ecr.NewDescribeRepositoriesPaginator(conn, input)

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, repo := range page.Repositories {
				name := aws.ToString(repo.RepositoryName)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), name)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.SetId(name)

				diags := resourceRepositoryRead(ctx, rd, awsClient)
				if diags.HasError() || rd.Id() == "" {
					tflog.Error(ctx, "Reading ECR repository", map[string]any{
						names.AttrID: name,
						"diags":      sdkdiag.DiagnosticsString(diags),
					})
					continue
				}

				result.DisplayName = name

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
}
