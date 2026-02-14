// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_codebuild_project")
func newProjectResourceAsListResource() inttypes.ListResourceForSDK {
	l := projectListResource{}
	l.SetResourceSchema(resourceProject())
	return &l
}

type projectListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type projectListResourceModel struct {
	framework.WithRegionModel
}

func (l *projectListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.CodeBuildClient(ctx)

	var query projectListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input codebuild.ListProjectsInput

	tflog.Info(ctx, "Listing CodeBuild projects")

	stream.Results = func(yield func(list.ListResult) bool) {
		for projectName, err := range listProjects(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), projectName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(projectName)

			tflog.Info(ctx, "Listing CodeBuild project")
			diags := resourceProjectRead(ctx, rd, awsClient)
			if diags.HasError() {
				tflog.Error(ctx, "Error reading CodeBuild project", map[string]any{
					"project_name": projectName,
					"error":        diags,
				})
				continue
			}

			if rd.Id() == "" {
				continue
			}

			result.DisplayName = projectName

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				tflog.Error(ctx, "Error setting result for CodeBuild project", map[string]any{
					"project_name": projectName,
					"error":        result.Diagnostics,
				})
				continue
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listProjects(ctx context.Context, conn *codebuild.Client, input *codebuild.ListProjectsInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := codebuild.NewListProjectsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield("", fmt.Errorf("listing CodeBuild Projects: %w", err))
				return
			}

			for _, projectName := range page.Projects {
				if !yield(projectName, nil) {
					return
				}
			}
		}
	}
}
