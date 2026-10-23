// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_observabilityadmin_dataset_integration", name="Dataset Integration")
// @ArnIdentity
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/observabilityadmin;observabilityadmin;observabilityadmin.GetDatasetIntegrationOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccDatasetIntegrationPreCheck")
// @Testing(serialize=true)
func newDatasetIntegrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &datasetIntegrationResource{}, nil
}

type datasetIntegrationResource struct {
	framework.ResourceWithModel[datasetIntegrationResourceModel]
	framework.WithImportByIdentity
}

func (r *datasetIntegrationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (r *datasetIntegrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data datasetIntegrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	var input observabilityadmin.CreateDatasetIntegrationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateDatasetIntegration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *datasetIntegrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data datasetIntegrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	output, err := findDatasetIntegrationByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *datasetIntegrationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old datasetIntegrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, new.ARN)
	if !new.RoleARN.Equal(old.RoleARN) {
		input := observabilityadmin.UpdateDatasetIntegrationInput{
			Arn:     aws.String(arn),
			RoleArn: fwflex.StringFromFramework(ctx, new.RoleARN),
		}
		if _, err := conn.UpdateDatasetIntegration(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
			return
		}
	}

	output, err := findDatasetIntegrationByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &new))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *datasetIntegrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data datasetIntegrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	input := observabilityadmin.DeleteDatasetIntegrationInput{
		Arn: aws.String(arn),
	}
	_, err := conn.DeleteDatasetIntegration(ctx, &input)
	switch {
	case errs.IsA[*awstypes.ResourceNotFoundException](err):
		return
	case err != nil:
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}
}

type datasetIntegrationResourceModel struct {
	framework.WithRegionModel
	ARN       types.String      `tfsdk:"arn"`
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	RoleARN   fwtypes.ARN       `tfsdk:"role_arn"`
	Tags      tftags.Map        `tfsdk:"tags"`
	TagsAll   tftags.Map        `tfsdk:"tags_all"`
	UpdatedAt timetypes.RFC3339 `tfsdk:"updated_at"`
}

func findDatasetIntegrationByARN(ctx context.Context, conn *observabilityadmin.Client, arn string) (*observabilityadmin.GetDatasetIntegrationOutput, error) {
	input := observabilityadmin.GetDatasetIntegrationInput{
		Arn: aws.String(arn),
	}

	return findDatasetIntegration(ctx, conn, &input)
}

func findDatasetIntegration(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetDatasetIntegrationInput) (*observabilityadmin.GetDatasetIntegrationOutput, error) {
	output, err := conn.GetDatasetIntegration(ctx, input)
	switch {
	case errs.IsA[*awstypes.ResourceNotFoundException](err):
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	case err != nil:
		return nil, smarterr.NewError(err)
	case output == nil:
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func findDatasetIntegrations(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.ListDatasetIntegrationsInput) ([]awstypes.DatasetIntegrationSummary, error) {
	var output []awstypes.DatasetIntegrationSummary

	pages := observabilityadmin.NewListDatasetIntegrationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.DatasetIntegrationSummaries...)
	}

	return output, nil
}
