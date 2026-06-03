// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_s3_table_source_association", name="S3 Table Source Association")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types;awstypes;*awstypes.S3TableIntegrationSource")
// @Testing(tagsTest=false)
func newS3TableSourceAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &s3TableSourceAssociationResource{}

	return r, nil
}

type s3TableSourceAssociationResource struct {
	framework.ResourceWithModel[s3TableSourceAssociationResourceModel]
	framework.WithNoUpdate
}

func (r *s3TableSourceAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"datasource_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("*"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"datasource_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("*"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"integration_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *s3TableSourceAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data s3TableSourceAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	input := cloudwatchlogs.AssociateSourceToS3TableIntegrationInput{
		IntegrationArn: fwflex.StringFromFramework(ctx, data.IntegrationARN),
		DataSource: &awstypes.DataSource{
			Name: fwflex.StringFromFramework(ctx, data.DatasourceName),
			Type: fwflex.StringFromFramework(ctx, data.DatasourceType),
		},
	}

	output, err := conn.AssociateSourceToS3TableIntegration(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("creating CloudWatch Logs S3 Table Source Association (%s/%s)", data.IntegrationARN.ValueString(), data.DatasourceName.ValueString()),
			err.Error(),
		)
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.Identifier)

	// AssociateSourceToS3TableIntegration only returns identifier; read back to populate computed fields.
	out, err := findS3TableSourceAssociationByTwoPartKey(ctx, conn, data.IntegrationARN.ValueString(), aws.ToString(output.Identifier))
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("reading CloudWatch Logs S3 Table Source Association (%s) after create", aws.ToString(output.Identifier)),
			err.Error(),
		)
		return
	}

	if out.DataSource != nil {
		data.DatasourceName = fwflex.StringToFramework(ctx, out.DataSource.Name)
		data.DatasourceType = fwflex.StringToFramework(ctx, out.DataSource.Type)
	}
	data.Status = fwflex.StringValueToFramework(ctx, string(out.Status))
	data.StatusReason = fwflex.StringToFramework(ctx, out.StatusReason)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *s3TableSourceAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data s3TableSourceAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	out, err := findS3TableSourceAssociationByTwoPartKey(ctx, conn, data.IntegrationARN.ValueString(), data.ID.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("reading CloudWatch Logs S3 Table Source Association (%s)", data.ID.ValueString()),
			err.Error(),
		)
		return
	}

	if out.DataSource != nil {
		data.DatasourceName = fwflex.StringToFramework(ctx, out.DataSource.Name)
		data.DatasourceType = fwflex.StringToFramework(ctx, out.DataSource.Type)
	}
	data.Status = fwflex.StringValueToFramework(ctx, string(out.Status))
	data.StatusReason = fwflex.StringToFramework(ctx, out.StatusReason)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *s3TableSourceAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data s3TableSourceAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	_, err := conn.DisassociateSourceFromS3TableIntegration(ctx, &cloudwatchlogs.DisassociateSourceFromS3TableIntegrationInput{
		Identifier: fwflex.StringFromFramework(ctx, data.ID),
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("deleting CloudWatch Logs S3 Table Source Association (%s)", data.ID.ValueString()),
			err.Error(),
		)
		return
	}
}

func (r *s3TableSourceAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
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

func findS3TableSourceAssociationByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.Client, integrationARN, identifier string) (*awstypes.S3TableIntegrationSource, error) {
	input := cloudwatchlogs.ListSourcesForS3TableIntegrationInput{
		IntegrationArn: aws.String(integrationARN),
	}
	return findS3TableSourceAssociation(ctx, conn, &input, func(v awstypes.S3TableIntegrationSource) bool {
		return aws.ToString(v.Identifier) == identifier
	})
}

func findS3TableSourceAssociation(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.ListSourcesForS3TableIntegrationInput, filter func(awstypes.S3TableIntegrationSource) bool) (*awstypes.S3TableIntegrationSource, error) {
	paginator := cloudwatchlogs.NewListSourcesForS3TableIntegrationPaginator(conn, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || tfawserr.ErrMessageContains(err, "ValidationException", "Integration not found") {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.Sources {
			if filter(v) {
				return &v, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError()
}

type s3TableSourceAssociationResourceModel struct {
	framework.WithRegionModel
	DatasourceName types.String `tfsdk:"datasource_name"`
	DatasourceType types.String `tfsdk:"datasource_type"`
	ID             types.String `tfsdk:"id"`
	IntegrationARN fwtypes.ARN  `tfsdk:"integration_arn"`
	Status         types.String `tfsdk:"status"`
	StatusReason   types.String `tfsdk:"status_reason"`
}
