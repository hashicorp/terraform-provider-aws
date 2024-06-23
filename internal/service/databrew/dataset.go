// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package databrew

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databrew"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databrew/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_databrew_dataset", name="Dataset")
func newResourceDataset(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDataset{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDataset = "Dataset"
)

type resourceDataset struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDataset) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_databrew_dataset"
}

func (r *resourceDataset) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"input": schema.MapNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "s3_input_definition": schema.MapNestedAttribute{
                            NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"bucket": schema.StringAttribute{
										Required: true,
									},
								},
							},
                        },
                    },
                },
                Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDataset) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var plan resourceDatasetData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &databrew.CreateDatasetInput{
		Name:        aws.String(plan.Name.ValueString()),
	}

	out, err := conn.CreateDataset(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionCreating, ResNameDataset, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Name == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionCreating, ResNameDataset, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.Name = flex.StringToFramework(ctx, out.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDataset) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var state resourceDatasetData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDatasetByName(ctx, conn, state.Name.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionSetting, ResNameDataset, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	state.Name = flex.StringToFramework(ctx, out.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDataset) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var plan, state resourceDatasetData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		in := &databrew.UpdateDatasetInput{
			Name: aws.String(plan.Name.ValueString()),
		}

		out, err := conn.UpdateDataset(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataBrew, create.ErrActionUpdating, ResNameDataset, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Name == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataBrew, create.ErrActionUpdating, ResNameDataset, plan.Name.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.Name = flex.StringToFramework(ctx, out.Name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDataset) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var state resourceDatasetData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &databrew.DeleteDatasetInput{
		Name: aws.String(state.Name.ValueString()),
	}

	_, err := conn.DeleteDataset(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionDeleting, ResNameDataset, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	// deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	// _, err = waitDatasetDeleted(ctx, conn, state.Name.ValueString(), deleteTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.DataBrew, create.ErrActionWaitingForDeletion, ResNameDataset, state.Name.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }
}

func (r *resourceDataset) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func findDatasetByName(ctx context.Context, conn *databrew.Client, name string) (*databrew.DescribeDatasetOutput, error) {
	in := &databrew.DescribeDatasetInput{
		Name: aws.String(name),
	}

	out, err := conn.DescribeDataset(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Name == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceDatasetData struct {
	Name        types.String   `tfsdk:"name"`
	Input types.String   `tfsdk:"input"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}
