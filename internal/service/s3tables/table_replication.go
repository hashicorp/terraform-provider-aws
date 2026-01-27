// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3tables

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table_replication", name="Table Replication")
// @ArnIdentity("table_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3tables;s3tables.GetTableReplicationOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importIgnore="version_token")
// @Testing(plannableImportAction="NoOp")
func newTableReplicationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &tableReplicationResource{}, nil
}

type tableReplicationResource struct {
	framework.ResourceWithModel[tableReplicationResourceModel]
	framework.WithImportByIdentity
}

func (r *tableReplicationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrRole: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"table_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version_token": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tableReplicationRuleModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrDestination: schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[replicationDestinationModel](ctx),
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.SizeAtMost(5),
								setvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"destination_table_bucket_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *tableReplicationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tableReplicationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	var configuration awstypes.TableReplicationConfiguration
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &configuration)...)
	if response.Diagnostics.HasError() {
		return
	}

	tableARN := fwflex.StringValueFromFramework(ctx, data.TableARN)
	input := s3tables.PutTableReplicationInput{
		Configuration: &configuration,
		TableArn:      aws.String(tableARN),
	}

	output, err := conn.PutTableReplication(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Tables Table Replication (%s)", tableARN), err.Error())

		return
	}

	// Set values for unknowns.
	data.VersionToken = fwflex.StringToFramework(ctx, output.VersionToken)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *tableReplicationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tableReplicationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableARN := fwflex.StringValueFromFramework(ctx, data.TableARN)
	output, err := findTableReplicationByARN(ctx, conn, tableARN)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Replication (%s)", tableARN), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.Configuration, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *tableReplicationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new tableReplicationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	var configuration awstypes.TableReplicationConfiguration
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &configuration)...)
	if response.Diagnostics.HasError() {
		return
	}

	tableARN := fwflex.StringValueFromFramework(ctx, new.TableARN)
	input := s3tables.PutTableReplicationInput{
		Configuration: &configuration,
		TableArn:      aws.String(tableARN),
		VersionToken:  fwflex.StringFromFramework(ctx, old.VersionToken),
	}

	output, err := conn.PutTableReplication(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Tables Table Replication (%s)", tableARN), err.Error())

		return
	}

	// Set values for unknowns.
	new.VersionToken = fwflex.StringToFramework(ctx, output.VersionToken)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *tableReplicationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tableReplicationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableARN := fwflex.StringValueFromFramework(ctx, data.TableARN)
	input := s3tables.DeleteTableReplicationInput{
		TableArn:     aws.String(tableARN),
		VersionToken: fwflex.StringFromFramework(ctx, data.VersionToken),
	}
	_, err := conn.DeleteTableReplication(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Tables Replication (%s)", tableARN), err.Error())

		return
	}
}

func findTableReplicationByARN(ctx context.Context, conn *s3tables.Client, tableARN string) (*s3tables.GetTableReplicationOutput, error) {
	input := s3tables.GetTableReplicationInput{
		TableArn: aws.String(tableARN),
	}

	return findTableReplication(ctx, conn, &input)
}

func findTableReplication(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTableReplicationInput) (*s3tables.GetTableReplicationOutput, error) {
	output, err := conn.GetTableReplication(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Configuration == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type tableReplicationResourceModel struct {
	framework.WithRegionModel
	Role         fwtypes.ARN                                                `tfsdk:"role"`
	Rules        fwtypes.ListNestedObjectValueOf[tableReplicationRuleModel] `tfsdk:"rule"`
	TableARN     fwtypes.ARN                                                `tfsdk:"table_arn"`
	VersionToken types.String                                               `tfsdk:"version_token"`
}

type tableReplicationRuleModel struct {
	Destinations fwtypes.SetNestedObjectValueOf[replicationDestinationModel] `tfsdk:"destination"`
}
