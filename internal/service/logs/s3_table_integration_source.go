// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_s3_table_integration_source", name="S3 Table Integration Source")
// @IdentityAttribute("integration_arn")
// @IdentityAttribute("id")
// @ImportIDHandler("s3TableIntegrationSourceImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdFunc=testAccS3TableIntegrationSourceImportStateIDFunc)
func newS3TableIntegrationSourceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &s3TableIntegrationSourceResource{}

	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type s3TableIntegrationSourceResource struct {
	framework.ResourceWithModel[s3TableIntegrationSourceResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
	framework.WithTimeouts
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Delete: true,
			}),
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
	response.Diagnostics.Append(r.flatten(ctx, output, &data)...)
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

	integrationARN, id := fwflex.StringValueFromFramework(ctx, data.IntegrationARN), fwflex.StringValueFromFramework(ctx, data.ID)
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

	if _, err := waitS3TableIntegrationDeleted(ctx, conn, integrationARN, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudWatch Logs S3 Table Integration (%s) Data Source Association (%s) delete", integrationARN, id), err.Error())
		return
	}
}

func (r *s3TableIntegrationSourceResource) flatten(ctx context.Context, s3TableIntegrationSource *awstypes.S3TableIntegrationSource, data *s3TableIntegrationSourceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, s3TableIntegrationSource, data)...)
	return diags
}

const s3TableIntegrationSourceImportIDSeparator = intflex.ResourceIdSeparator

func s3TableIntegrationSourceParseImportID(id string) (string, string, error) {
	parts := strings.Split(id, s3TableIntegrationSourceImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected integration-arn%[2]sidentifier", id, s3TableIntegrationSourceImportIDSeparator)
}

var (
	_ inttypes.ImportIDParser = s3TableIntegrationSourceImportID{}
)

type s3TableIntegrationSourceImportID struct{}

func (s3TableIntegrationSourceImportID) Parse(identifier string) (string, map[string]any, error) {
	integrationARN, identifier, err := s3TableIntegrationSourceParseImportID(identifier)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		names.AttrID:      identifier,
		"integration_arn": integrationARN,
	}

	return identifier, result, nil
}

func findS3TableIntegrationSourceByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.Client, integrationARN, identifier string) (*awstypes.S3TableIntegrationSource, error) {
	input := cloudwatchlogs.ListSourcesForS3TableIntegrationInput{
		IntegrationArn: aws.String(integrationARN),
	}

	return findS3TableIntegrationSource(ctx, conn, &input, tfslices.WithFilter(func(v awstypes.S3TableIntegrationSource) bool {
		return aws.ToString(v.Identifier) == identifier
	}))
}

func findS3TableIntegrationSource(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.ListSourcesForS3TableIntegrationInput, optFns ...tfslices.FinderOptionsFunc[awstypes.S3TableIntegrationSource]) (*awstypes.S3TableIntegrationSource, error) {
	output, err := findS3TableIntegrationSources(ctx, conn, input, optFns...)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findS3TableIntegrationSources(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.ListSourcesForS3TableIntegrationInput, optFns ...tfslices.FinderOptionsFunc[awstypes.S3TableIntegrationSource]) ([]awstypes.S3TableIntegrationSource, error) {
	output, err := tfslices.CollectAndConcatWithError(listS3TableIntegrationSourcePages(ctx, conn, input), optFns...)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || tfawserr.ErrMessageContains(err, errCodeValidationException, "Integration not found") || tfawserr.ErrMessageContains(err, errCodeValidationException, "Invalid integration ARN") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func listS3TableIntegrationSourcePages(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.ListSourcesForS3TableIntegrationInput, optFns ...func(*cloudwatchlogs.Options)) iter.Seq2[[]awstypes.S3TableIntegrationSource, error] {
	return func(yield func([]awstypes.S3TableIntegrationSource, error) bool) {
		pages := cloudwatchlogs.NewListSourcesForS3TableIntegrationPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Logs S3TableIntegrationSources: %w", err))
				return
			}

			if !yield(page.Sources, nil) {
				return
			}
		}
	}
}

func listS3TableIntegrationSources(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.ListSourcesForS3TableIntegrationInput, optFns ...func(*cloudwatchlogs.Options)) iter.Seq2[awstypes.S3TableIntegrationSource, error] {
	return tfiter.ConcatValuesWithError(listS3TableIntegrationSourcePages(ctx, conn, input, optFns...))
}

func statusS3TableIntegrationSource(conn *cloudwatchlogs.Client, integrationARN, identifier string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findS3TableIntegrationSourceByTwoPartKey(ctx, conn, integrationARN, identifier)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitS3TableIntegrationDeleted(ctx context.Context, conn *cloudwatchlogs.Client, integrationARN, identifier string, timeout time.Duration) (*awstypes.S3TableIntegrationSource, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{"DELETING"}, // Undocumented status value observed in API responses.
		Target:  []string{},
		Refresh: statusS3TableIntegrationSource(conn, integrationARN, identifier),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.S3TableIntegrationSource); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		return output, err
	}

	return nil, err
}

type s3TableIntegrationSourceResourceModel struct {
	framework.WithRegionModel
	DataSource     fwtypes.ListNestedObjectValueOf[dataSourceModel] `tfsdk:"data_source"`
	ID             types.String                                     `tfsdk:"id"`
	IntegrationARN fwtypes.ARN                                      `tfsdk:"integration_arn"`
	Timeouts       timeouts.Value                                   `tfsdk:"timeouts"`
}

type dataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}
