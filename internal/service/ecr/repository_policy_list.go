// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_ecr_repository_policy")
func newRepositoryPolicyResourceAsListResource() inttypes.ListResourceForSDK {
	l := repositoryPolicyListResource{}
	l.SetResourceSchema(resourceRepositoryPolicy())
	return &l
}

var _ list.ListResource = &repositoryPolicyListResource{}

type repositoryPolicyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type repositoryPolicyListResourceModel struct {
	framework.WithRegionModel
}

func (l *repositoryPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query repositoryPolicyListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	conn := l.Meta().ECRClient(ctx)

	tflog.Info(ctx, "Listing ECR repository policies")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &ecr.DescribeRepositoriesInput{}

		for repository, err := range listRepositories(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			repositoryName := aws.ToString(repository.RepositoryName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("repository"), repositoryName)

			output, err := findRepositoryPolicyByRepositoryName(ctx, conn, repositoryName)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading ECR Repository Policy (%s): %w", repositoryName, err))
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(repositoryName)
			rd.Set("repository", repositoryName)

			if request.IncludeResource {
				policy, err := structure.NormalizeJsonString(aws.ToString(output.PolicyText))
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("normalizing ECR Repository Policy (%s): %w", repositoryName, err))
					yield(result)
					return
				}

				resourceRepositoryPolicyFlatten(rd, output, policy)
			}

			result.DisplayName = repositoryName

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
