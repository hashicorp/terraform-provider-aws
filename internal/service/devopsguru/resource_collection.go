// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/devopsguru"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// @FrameworkResource(name="Resource Collection")
func newResourceResourceCollection(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceResourceCollection{}, nil
}

const (
	ResNameResourceCollection = "Resource Collection"
)

type resourceResourceCollection struct {
	framework.ResourceWithConfigure
}

func (r *resourceResourceCollection) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_devopsguru_resource_collection"
}

func (r *resourceResourceCollection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"type": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ResourceCollectionType](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"cloudformation": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[cloudformationData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"stack_names": schema.ListAttribute{
							Required:    true,
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"tags": schema.ListNestedBlock{
				// Attempting to specify multiple app boundary keys will result in a ValidationException
				//
				//   ValidationException: Multiple app boundary keys are not supported
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[tagsData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"app_boundary_key": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"tag_values": schema.ListAttribute{
							Required:    true,
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceResourceCollection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var plan resourceResourceCollectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(plan.Type.ValueString())

	rc := &awstypes.UpdateResourceCollectionFilter{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, rc)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Tags.IsNull() {
		// Fields named "Tags" are currently hardcoded to be ignored by AutoFlex. Expanding plan.Tags
		// into the request structs Tags field is a temporary workaround until the AutoFlex
		// options implementation can be merged.
		//
		// Ref: https://github.com/hashicorp/terraform-provider-aws/pull/36437
		resp.Diagnostics.Append(flex.Expand(ctx, plan.Tags, &rc.Tags)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	in := &devopsguru.UpdateResourceCollectionInput{
		Action:             awstypes.UpdateResourceCollectionActionAdd,
		ResourceCollection: rc,
	}

	out, err := conn.UpdateResourceCollection(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionCreating, ResNameResourceCollection, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionCreating, ResNameResourceCollection, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceResourceCollection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var state resourceResourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourceCollectionByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionSetting, ResNameResourceCollection, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fields named "Tags" are currently hardcoded to be ignored by AutoFlex. Flattening the Tags
	// struct from the response into state.Tags is a temporary workaround until the AutoFlex
	// options implementation can be merged.
	//
	// Ref: https://github.com/hashicorp/terraform-provider-aws/pull/36437
	resp.Diagnostics.Append(flex.Flatten(ctx, out.Tags, &state.Tags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Copy from ID on read to support import
	state.Type = fwtypes.StringEnumValue(awstypes.ResourceCollectionType(state.ID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceResourceCollection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is a no-op
}

func (r *resourceResourceCollection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var state resourceResourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rc := &awstypes.UpdateResourceCollectionFilter{}
	resp.Diagnostics.Append(flex.Expand(ctx, state, rc)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.Tags.IsNull() {
		// Fields named "Tags" are currently hardcoded to be ignored by AutoFlex. Expanding state.Tags
		// into the request structs Tags field is a temporary workaround until the AutoFlex
		// options implementation can be merged.
		//
		// Ref: https://github.com/hashicorp/terraform-provider-aws/pull/36437
		resp.Diagnostics.Append(flex.Expand(ctx, state.Tags, &rc.Tags)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	in := &devopsguru.UpdateResourceCollectionInput{
		Action:             awstypes.UpdateResourceCollectionActionRemove,
		ResourceCollection: rc,
	}

	_, err := conn.UpdateResourceCollection(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionDeleting, ResNameResourceCollection, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResourceCollection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findResourceCollectionByID(ctx context.Context, conn *devopsguru.Client, id string) (*awstypes.ResourceCollectionFilter, error) {
	collectionType := awstypes.ResourceCollectionType(id)
	in := &devopsguru.GetResourceCollectionInput{
		ResourceCollectionType: collectionType,
	}

	out, err := conn.GetResourceCollection(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ResourceCollection == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	switch collectionType {
	case awstypes.ResourceCollectionTypeAwsCloudFormation, awstypes.ResourceCollectionTypeAwsService:
		// AWS_CLOUD_FORMATION and AWS_SERVICE collection types should have
		// a non-empty array of stack names
		if out.ResourceCollection.CloudFormation == nil ||
			len(out.ResourceCollection.CloudFormation.StackNames) == 0 {
			return nil, &retry.NotFoundError{
				LastRequest: in,
			}
		}
	case awstypes.ResourceCollectionTypeAwsTags:
		// AWS_TAGS collection types should have a Tags array with 1 item,
		// and that object should have a TagValues array with at least 1 item
		if len(out.ResourceCollection.Tags) == 0 ||
			len(out.ResourceCollection.Tags) == 1 && len(out.ResourceCollection.Tags[0].TagValues) == 0 {
			return nil, &retry.NotFoundError{
				LastRequest: in,
			}
		}
	}

	return out.ResourceCollection, nil
}

type resourceResourceCollectionData struct {
	CloudFormation fwtypes.ListNestedObjectValueOf[cloudformationData] `tfsdk:"cloudformation"`
	ID             types.String                                        `tfsdk:"id"`
	Tags           fwtypes.ListNestedObjectValueOf[tagsData]           `tfsdk:"tags"`
	Type           fwtypes.StringEnum[awstypes.ResourceCollectionType] `tfsdk:"type"`
}

type cloudformationData struct {
	StackNames fwtypes.ListValueOf[types.String] `tfsdk:"stack_names"`
}

type tagsData struct {
	AppBoundaryKey types.String                      `tfsdk:"app_boundary_key"`
	TagValues      fwtypes.ListValueOf[types.String] `tfsdk:"tag_values"`
}
