// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
// @FrameworkResource("aws_cloudfront_invalidation", name="Invalidation")
func newResourceInvalidation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceInvalidation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameInvalidation = "Invalidation"
)

type resourceInvalidation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceInvalidation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_cloudfront_invalidation"
}

func (r *resourceInvalidation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"distribution_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"paths": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTriggers: schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceInvalidation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var plan resourceInvalidationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var paths []string
	for _, p := range flex.ExpandFrameworkStringList(ctx, plan.Paths) {
		paths = append(paths, *p)
	}

	in := &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(plan.DistributionID.ValueString()),
		InvalidationBatch: &awstypes.InvalidationBatch{
			CallerReference: aws.String(uuid.NewString()),
			Paths: &awstypes.Paths{
				Quantity: aws.Int32(int32(len(paths))),
				Items:    paths,
			},
		},
	}

	out, err := conn.CreateInvalidation(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionCreating, ResNameInvalidation, plan.DistributionID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Invalidation == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionCreating, ResNameInvalidation, plan.DistributionID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	plan.ID = flex.StringToFramework(ctx, out.Invalidation.Id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	waiter := cloudfront.NewInvalidationCompletedWaiter(conn)
	waitResp, err := waiter.WaitForOutput(ctx, &cloudfront.GetInvalidationInput{
		DistributionId: aws.String(plan.DistributionID.ValueString()),
		Id:             out.Invalidation.Id,
	}, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionWaitingForCreation, ResNameInvalidation, plan.DistributionID.String(), err),
			err.Error(),
		)
		return
	}
	plan.Status = flex.StringToFramework(ctx, waitResp.Invalidation.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceInvalidation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var state resourceInvalidationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findInvalidationByID(ctx, conn, state.DistributionID.ValueString(), state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionSetting, ResNameInvalidation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.Id)
	state.Status = flex.StringToFramework(ctx, out.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// There is no update API, so this method is a no-op
func (r *resourceInvalidation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// There is no update API, so this method is a no-op
func (r *resourceInvalidation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func findInvalidationByID(ctx context.Context, conn *cloudfront.Client, distributionId, id string) (*awstypes.Invalidation, error) {
	in := &cloudfront.GetInvalidationInput{
		DistributionId: aws.String(distributionId),
		Id:             aws.String(id),
	}

	out, err := conn.GetInvalidation(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.NoSuchInvalidation](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Invalidation == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Invalidation, nil
}

type resourceInvalidationData struct {
	Paths          types.List     `tfsdk:"paths"`
	ID             types.String   `tfsdk:"id"`
	Status         types.String   `tfsdk:"status"`
	DistributionID types.String   `tfsdk:"distribution_id"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}
