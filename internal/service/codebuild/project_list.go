// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
		pages := codebuild.NewListProjectsPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing CodeBuild Projects: %w", err))
				yield(result)
				return
			}

			if len(page.Projects) == 0 {
				continue
			}

			// Batch-get full project details when IncludeResource is set.
			var projects []types.Project
			if request.IncludeResource {
				projects, err = findProjects(ctx, conn, &codebuild.BatchGetProjectsInput{
					Names: page.Projects,
				})
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading CodeBuild Projects: %w", err))
					yield(result)
					return
				}
			}

			// Build a map for lookup by name.
			projectsByName := make(map[string]*types.Project, len(projects))
			for i := range projects {
				projectsByName[aws.ToString(projects[i].Name)] = &projects[i]
			}

			for _, projectName := range page.Projects {
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), projectName)

				result := request.NewListResult(ctx)
				rd := l.ResourceData()
				rd.SetId(projectName)
				rd.Set(names.AttrARN, awsClient.RegionalARN(ctx, "codebuild", "project/"+projectName))

				if project, ok := projectsByName[projectName]; ok {
					diags := resourceProjectFlatten(ctx, rd, project)
					if diags.HasError() {
						tflog.Error(ctx, "Error reading CodeBuild project", map[string]any{
							"error": sdkdiag.DiagnosticsString(diags),
						})
						continue
					}
				}
				result.DisplayName = projectName

				l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
				if result.Diagnostics.HasError() {
					tflog.Error(ctx, "Error setting result for CodeBuild project", map[string]any{
						"error": result.Diagnostics,
					})
					continue
				}

				if !yield(result) {
					return
				}
			}
		}
	}
}
