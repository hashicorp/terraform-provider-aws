// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_s3tables_table_replication", name="Table Replication")
// @ArnIdentity("table_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3tables;s3tables.GetTableReplicationOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newTableReplicationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &tableReplicationResource{}, nil
}

type tableReplicationResource struct {
	framework.ResourceWithModel[tableReplicationResourceModel]
}

func (r *tableReplicationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"table_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

	tableARN := fwflex.StringValueFromFramework(ctx, data.TableARN)

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

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Replication (%s)", tableARN), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *tableReplicationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new tableReplicationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableARN := fwflex.StringValueFromFramework(ctx, new.TableARN)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *tableReplicationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tableReplicationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	tableARN := fwflex.StringValueFromFramework(ctx, new.TableARN)
}

func findTableReplicationByARN(ctx context.Context, conn *s3tables.Client, tableARN string) (*s3tables.GetTablePolicyOutput, error) {
	input := s3tables.GetTablePolicyInput{}

	return findTableReplication(ctx, conn, &input)
}

func findTableReplication(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTablePolicyInput) (*s3tables.GetTablePolicyOutput, error) {
	output, err := conn.GetTablePolicy(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || aws.ToString(output.ResourcePolicy) == "" {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type tableReplicationResourceModel struct {
	framework.WithRegionModel
	TableARN fwtypes.ARN `tfsdk:"table_arn"`
}
