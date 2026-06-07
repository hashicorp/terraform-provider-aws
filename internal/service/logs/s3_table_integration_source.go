// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_s3_table_integration_source", name="S3 Table Integration Data Source Association")
func newS3TableIntegrationSourceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &s3TableIntegrationSourceResource{}

	return r, nil
}

type s3TableIntegrationSourceResource struct {
	framework.ResourceWithModel[s3TableIntegrationSourceResourceModel]
	framework.WithNoUpdate
}

func (r *s3TableIntegrationSourceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"integration_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"data_source": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrType: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *s3TableIntegrationSourceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data s3TableIntegrationSourceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	var input cloudwatchlogs.AssociateSourceToS3TableIntegrationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.AssociateSourceToS3TableIntegration(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating CloudWatch Logs S3 Table Integration Data Source Association", err.Error())
		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.Identifier)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *s3TableIntegrationSourceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data s3TableIntegrationSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	integrationARN, id := fwflex.StringValueFromFramework(ctx, data.IntegrationARN), fwflex.StringValueFromFramework(ctx, data.ID)
	output, err := findS3TableIntegrationSourceByTwoPartKey(ctx, conn, integrationARN, id)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs S3 Table Integration (%s) Data Source Association (%s)", integrationARN, id), err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *s3TableIntegrationSourceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data s3TableIntegrationSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	input := cloudwatchlogs.DisassociateSourceFromS3TableIntegrationInput{
		Identifier: aws.String(id),
	}
	_, err := conn.DisassociateSourceFromS3TableIntegration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudWatch Logs S3 Table Integration Data Source Association (%s)", id), err.Error())
		return
	}
}

func (r *s3TableIntegrationSourceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Import ID format: <integration_arn>,<identifier>
	// The comma separator is used because integration ARNs contain slashes.
	parts := strings.SplitN(request.ID, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: integration_arn,id. Got: %q", request.ID),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("integration_arn"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
}

func findS3TableIntegrationSourceByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.Client, integrationARN, identifier string) (*awstypes.S3TableIntegrationSource, error) {
	input := cloudwatchlogs.ListSourcesForS3TableIntegrationInput{
		IntegrationArn: aws.String(integrationARN),
	}
	return findS3TableIntegrationSource(ctx, conn, &input, func(v awstypes.S3TableIntegrationSource) bool {
		return aws.ToString(v.Identifier) == identifier
	})
}

func listS3TableIntegrationSources(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.ListSourcesForS3TableIntegrationInput, filter tfslices.Predicate[awstypes.S3TableIntegrationSource]) iter.Seq2[awstypes.S3TableIntegrationSource, error] {
	return func(yield func(awstypes.S3TableIntegrationSource, error) bool) {
		pages := cloudwatchlogs.NewListSourcesForS3TableIntegrationPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.S3TableIntegrationSource](), fmt.Errorf("listing CloudWatch Logs S3TableIntegrationSources: %w", err))
				return
			}

			for _, v := range page.Sources {
				if filter(v) {
					if !yield(v, nil) {
						return
					}
				}
			}
		}
	}
}

func findS3TableIntegrationSource(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.ListSourcesForS3TableIntegrationInput, filter tfslices.Predicate[awstypes.S3TableIntegrationSource]) (*awstypes.S3TableIntegrationSource, error) {
	var output []awstypes.S3TableIntegrationSource
	for v, err := range listS3TableIntegrationSources(ctx, conn, input, filter) {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Integration not found") {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, v)
	}

	return tfresource.AssertSingleValueResult(output)
}

type s3TableIntegrationSourceResourceModel struct {
	framework.WithRegionModel
	DataSource     fwtypes.ListNestedObjectValueOf[dataSourceModel] `tfsdk:"data_source"`
	ID             types.String                                     `tfsdk:"id"`
	IntegrationARN fwtypes.ARN                                      `tfsdk:"integration_arn"`
}

type dataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}
