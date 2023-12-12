// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Collection")
// @Tags(identifierAttribute="arn")
func newResourceCollection(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCollection{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultReadTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type resourceCollection struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

const (
	ResNameCollection = "Collection"
)

func (r *resourceCollection) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rekognition_collection"
}

func (r *resourceCollection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	collectionRegex, _ := regexp.Compile(`^[a-zA-Z0-9_.\-]+$`)

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"collection_id": schema.StringAttribute{
				Description: "The name of the Rekognition collection",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(collectionRegex, "must conform to: ^[a-zA-Z0-9_.\\-]+$"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id":                 framework.IDAttribute(),
			"face_model_version": schema.StringAttribute{Computed: true},
			names.AttrTags:       tftags.TagsAttribute(),
			names.AttrTagsAll:    tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
		Version: 1,
	}
}

func (r *resourceCollection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceCollection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var plan resourceCollectionData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rekognition.CreateCollectionInput{
		CollectionId: plan.CollectionId.ValueStringPointer(),
		Tags:         getTagsIn(ctx),
	}

	out, err := conn.CreateCollection(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameCollection, plan.CollectionId.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.CollectionArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameCollection, plan.CollectionId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	state.Id = plan.CollectionId
	state.Arn = flex.StringToFramework(ctx, out.CollectionArn)
	state.FaceModelVersion = flex.StringToFramework(ctx, out.FaceModelVersion)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceCollection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state resourceCollectionData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindCollectionByID(ctx, conn, state.Id.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionReading, ResNameCollection, state.Id.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.Arn = flex.StringToFramework(ctx, out.CollectionARN)
	state.FaceModelVersion = flex.StringToFramework(ctx, out.FaceModelVersion)
	state.CollectionId = flex.StringToFramework(ctx, state.Id.ValueStringPointer())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCollection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// resource update not supported, but tag updates are supported
	var plan resourceCollectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state resourceCollectionData
	resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)

	state.Tags = plan.Tags
	state.TagsAll = plan.TagsAll
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCollection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state resourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rekognition.DeleteCollectionInput{
		CollectionId: state.Id.ValueStringPointer(),
	}

	_, err := conn.DeleteCollection(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionDeleting, ResNameCollection, state.Id.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCollection) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func FindCollectionByID(ctx context.Context, conn *rekognition.Client, id string) (*rekognition.DescribeCollectionOutput, error) {
	in := &rekognition.DescribeCollectionInput{
		CollectionId: aws.String(id),
	}

	out, err := conn.DescribeCollection(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.CollectionARN == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceCollectionData struct {
	Arn              types.String   `tfsdk:"arn"`
	CollectionId     types.String   `tfsdk:"collection_id"`
	FaceModelVersion types.String   `tfsdk:"face_model_version"`
	Id               types.String   `tfsdk:"id"`
	Tags             types.Map      `tfsdk:"tags"`
	TagsAll          types.Map      `tfsdk:"tags_all"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}
