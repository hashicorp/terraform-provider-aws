// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_batch_job_definition")
func jobDefinitionResourceAsListResource() inttypes.ListResourceForSDK {
	l := jobDefinitionListResource{}
	l.SetResourceSchema(resourceJobDefinition())
	return &l
}

type jobDefinitionListResource struct {
	framework.ResourceWithConfigure
	framework.ListResourceWithSDKv2Resource
	framework.ListResourceWithSDKv2Tags
}

type jobDefinitionListResourceModel struct {
	framework.WithRegionModel
}

func (l *jobDefinitionListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{},
		Blocks:     map[string]listschema.Block{},
	}
}

func (l *jobDefinitionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.BatchClient(ctx)

	var query jobDefinitionListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input batch.DescribeJobDefinitionsInput

	tflog.Info(ctx, "Listing Batch job definitions")

	stream.Results = func(yield func(list.ListResult) bool) {
		pages := batch.NewDescribeJobDefinitionsPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, jobDef := range page.JobDefinitions {
				arn := aws.ToString(jobDef.JobDefinitionArn)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), arn)

				result := request.NewListResult(ctx)
				rd := l.ResourceData()
				rd.SetId(arn)

				tflog.Info(ctx, "Reading Batch job definition")
				diags := resourceJobDefinitionRead(ctx, rd, awsClient)
				if diags.HasError() {
					result = fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading Batch job definition %s", arn))
					yield(result)
					return
				}
				if rd.Id() == "" {
					// Resource is logically deleted
					continue
				}

				err = l.SetTags(ctx, awsClient, rd)
				if err != nil {
					result = fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}

				result.DisplayName = aws.ToString(jobDef.JobDefinitionName)

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
