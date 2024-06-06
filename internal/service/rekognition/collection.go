// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
	r.SetDefaultCreateTimeout(2 * time.Minute)

	return r, nil
}

type resourceCollection struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

const (
	ResNameCollection = "Collection"
)

func (r *resourceCollection) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rekognition_collection"
}

func (r *resourceCollection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	collectionRegex := regexache.MustCompile(`^[a-zA-Z0-9_.\-]+$`)

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
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
			names.AttrID: framework.IDAttribute(),
			"face_model_version": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}

	if s.Blocks == nil {
		s.Blocks = make(map[string]schema.Block)
	}
	s.Blocks[names.AttrTimeouts] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
	})

	resp.Schema = s
}

func (r *resourceCollection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var plan resourceCollectionData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rekognition.CreateCollectionInput{
		CollectionId: plan.CollectionID.ValueStringPointer(),
		Tags:         getTagsIn(ctx),
	}

	_, err := conn.CreateCollection(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameCollection, plan.CollectionID.ValueString(), err),
			err.Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)

	out, err := tfresource.RetryWhenNotFound(ctx, createTimeout, func() (interface{}, error) {
		return findCollectionByID(ctx, conn, plan.CollectionID.ValueString())
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameCollection, plan.CollectionID.ValueString(), err),
			err.Error(),
		)
		return
	}

	output := out.(*rekognition.DescribeCollectionOutput)

	state := plan
	state.ID = plan.CollectionID
	state.ARN = flex.StringToFramework(ctx, output.CollectionARN)
	state.FaceModelVersion = flex.StringToFramework(ctx, output.FaceModelVersion)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCollection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state resourceCollectionData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCollectionByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionReading, ResNameCollection, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.CollectionARN)
	state.FaceModelVersion = flex.StringToFramework(ctx, out.FaceModelVersion)
	state.CollectionID = flex.StringToFramework(ctx, state.ID.ValueStringPointer())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCollection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceCollectionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCollection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state resourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rekognition.DeleteCollectionInput{
		CollectionId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteCollection(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionDeleting, ResNameCollection, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCollection) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func findCollectionByID(ctx context.Context, conn *rekognition.Client, id string) (*rekognition.DescribeCollectionOutput, error) {
	in := &rekognition.DescribeCollectionInput{
		CollectionId: aws.String(id),
	}

	out, err := conn.DescribeCollection(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.CollectionARN == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceCollectionData struct {
	ARN              types.String   `tfsdk:"arn"`
	CollectionID     types.String   `tfsdk:"collection_id"`
	FaceModelVersion types.String   `tfsdk:"face_model_version"`
	ID               types.String   `tfsdk:"id"`
	Tags             types.Map      `tfsdk:"tags"`
	TagsAll          types.Map      `tfsdk:"tags_all"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}
