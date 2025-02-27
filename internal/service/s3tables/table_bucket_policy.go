// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table_bucket_policy", name="Table Bucket Policy")
func newResourceTableBucketPolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTableBucketPolicy{}, nil
}

const (
	ResNameTableBucketPolicy = "Table Bucket Policy"
)

type resourceTableBucketPolicy struct {
	framework.ResourceWithConfigure
}

func (r *resourceTableBucketPolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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

func (r *resourceTableBucketPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan resourceTableBucketPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3tables.PutTableBucketPolicyInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTableBucketPolicy(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTableBucketPolicy, plan.TableBucketARN.String(), err),
			err.Error(),
		)
		return
	}

	out, err := findTableBucketPolicy(ctx, conn, plan.TableBucketARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTableBucketPolicy, plan.TableBucketARN.String(), err),
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

func (r *resourceTableBucketPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableBucketPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTableBucketPolicy(ctx, conn, state.TableBucketARN.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionReading, ResNameTableBucketPolicy, state.TableBucketARN.String(), err),
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

func (r *resourceTableBucketPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan resourceTableBucketPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3tables.PutTableBucketPolicyInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutTableBucketPolicy(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTableBucketPolicy, plan.TableBucketARN.String(), err),
			err.Error(),
		)
		return
	}

	out, err := findTableBucketPolicy(ctx, conn, plan.TableBucketARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTableBucketPolicy, plan.TableBucketARN.String(), err),
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

func (r *resourceTableBucketPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableBucketPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := s3tables.DeleteTableBucketPolicyInput{
		TableBucketARN: state.TableBucketARN.ValueStringPointer(),
	}

	_, err := conn.DeleteTableBucketPolicy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionDeleting, ResNameTableBucketPolicy, state.TableBucketARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTableBucketPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("table_bucket_arn"), req, resp)
}

func findTableBucketPolicy(ctx context.Context, conn *s3tables.Client, tableBucketARN string) (*s3tables.GetTableBucketPolicyOutput, error) {
	in := s3tables.GetTableBucketPolicyInput{
		TableBucketARN: aws.String(tableBucketARN),
	}

	out, err := conn.GetTableBucketPolicy(ctx, &in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	return out, nil
}

type resourceTableBucketPolicyModel struct {
	ResourcePolicy fwtypes.IAMPolicy `tfsdk:"resource_policy"`
	TableBucketARN fwtypes.ARN       `tfsdk:"table_bucket_arn"`
}
