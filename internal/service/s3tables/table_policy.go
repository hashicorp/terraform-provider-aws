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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table_policy", name="Table Policy")
func newTablePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &tablePolicyResource{}, nil
}

type tablePolicyResource struct {
	framework.ResourceWithModel[tablePolicyResourceModel]
}

func (r *tablePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrNamespace: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_policy": schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			"table_bucket_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *tablePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tablePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input s3tables.PutTablePolicyInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTablePolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Tables Table Policy (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *tablePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tablePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name, namespace, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Name), fwflex.StringValueFromFramework(ctx, data.Namespace), fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	output, err := findTablePolicyByThreePartKey(ctx, conn, tableBucketARN, namespace, name)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Table Policy (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *tablePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new tablePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, new.Name)
	var input s3tables.PutTablePolicyInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTablePolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Tables Table Policy (%s)", name), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *tablePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tablePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	name, namespace, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Name), fwflex.StringValueFromFramework(ctx, data.Namespace), fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	input := s3tables.DeleteTablePolicyInput{
		Name:           aws.String(name),
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(tableBucketARN),
	}
	_, err := conn.DeleteTablePolicy(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Tables Table Policy (%s)", name), err.Error())

		return
	}
}

func (r *tablePolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	identifier, err := parseTableIdentifier(request.ID)
	if err != nil {
		response.Diagnostics.AddError(
			"Invalid Import ID",
			"Import IDs for S3 Tables Table Policies must use the format <table bucket ARN>"+tableIDSeparator+"<namespace>"+tableIDSeparator+"<table name>.\n"+
				fmt.Sprintf("Had %q", request.ID),
		)
		return
	}

	identifier.PopulateState(ctx, &response.State, &response.Diagnostics)
}

func findTablePolicyByThreePartKey(ctx context.Context, conn *s3tables.Client, tableBucketARN, namespace, name string) (*s3tables.GetTablePolicyOutput, error) {
	input := s3tables.GetTablePolicyInput{
		Name:           aws.String(name),
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(tableBucketARN),
	}

	return findTablePolicy(ctx, conn, &input)
}

func findTablePolicy(ctx context.Context, conn *s3tables.Client, input *s3tables.GetTablePolicyInput) (*s3tables.GetTablePolicyOutput, error) {
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
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type tablePolicyResourceModel struct {
	framework.WithRegionModel
	Name           types.String      `tfsdk:"name"`
	Namespace      types.String      `tfsdk:"namespace"`
	ResourcePolicy fwtypes.IAMPolicy `tfsdk:"resource_policy"`
	TableBucketARN fwtypes.ARN       `tfsdk:"table_bucket_arn"`
}
