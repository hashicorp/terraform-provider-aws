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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table_policy", name="Table Policy")
func newResourceTablePolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTablePolicy{}, nil
}

const (
	ResNameTablePolicy = "Table Policy"
)

type resourceTablePolicy struct {
	framework.ResourceWithConfigure
}

func (r *resourceTablePolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Required:   true,
				Validators: tableNameValidator,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrNamespace: schema.StringAttribute{
				Required:   true,
				Validators: namespaceNameValidator,
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

func (r *resourceTablePolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan resourceTablePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3tables.PutTablePolicyInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTablePolicy(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTablePolicy, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	out, err := findTablePolicy(ctx, conn, plan.TableBucketARN.ValueString(), plan.Namespace.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTableBucketPolicy, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTablePolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTablePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTablePolicy(ctx, conn, state.TableBucketARN.ValueString(), state.Namespace.ValueString(), state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionReading, ResNameTablePolicy, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTablePolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan resourceTablePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3tables.PutTablePolicyInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTablePolicy(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionUpdating, ResNameTablePolicy, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	out, err := findTablePolicy(ctx, conn, plan.TableBucketARN.ValueString(), plan.Namespace.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTableBucketPolicy, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTablePolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTablePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := s3tables.DeleteTablePolicyInput{
		Name:           state.Name.ValueStringPointer(),
		Namespace:      state.Namespace.ValueStringPointer(),
		TableBucketARN: state.TableBucketARN.ValueStringPointer(),
	}

	_, err := conn.DeleteTablePolicy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionDeleting, ResNameTablePolicy, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTablePolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	identifier, err := parseTableIdentifier(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import IDs for S3 Tables Table Policies must use the format <table bucket ARN>"+tableIDSeparator+"<namespace>"+tableIDSeparator+"<table name>.\n"+
				fmt.Sprintf("Had %q", req.ID),
		)
		return
	}

	identifier.PopulateState(ctx, &resp.State, &resp.Diagnostics)
}

func findTablePolicy(ctx context.Context, conn *s3tables.Client, bucketARN, namespace, name string) (*s3tables.GetTablePolicyOutput, error) {
	in := s3tables.GetTablePolicyInput{
		Name:           aws.String(name),
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(bucketARN),
	}

	out, err := conn.GetTablePolicy(ctx, &in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceTablePolicyModel struct {
	Name           types.String      `tfsdk:"name"`
	Namespace      types.String      `tfsdk:"namespace"`
	ResourcePolicy fwtypes.IAMPolicy `tfsdk:"resource_policy"`
	TableBucketARN fwtypes.ARN       `tfsdk:"table_bucket_arn"`
}
