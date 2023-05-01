package opensearchserverless

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
// @Tags(identifierAttribute="arn")
func newResourceCollection(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceCollection{}, nil
}

type resourceCollectionData struct {
	Arn         types.String `tfsdk:"arn"`
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Tags        types.Map    `tfsdk:"tags"`
	TagsAll     types.Map    `tfsdk:"tags_all"`
	Type        types.String `tfsdk:"type"`
}

const (
	ResNameCollection = "Collection"
)

type resourceCollection struct {
	framework.ResourceWithConfigure
}

func (r *resourceCollection) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_opensearchserverless_collection"
}

func (r *resourceCollection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
				},
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z][a-z0-9-]+$`),
						`must start with any lower case letter and can can include any lower case letter, number, or "-"`),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"type": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.CollectionType](),
				},
			},
		},
	}
}

func (r *resourceCollection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceCollectionData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient()

	in := &opensearchserverless.CreateCollectionInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(plan.Name.ValueString()),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.Type.IsNull() {
		in.Type = awstypes.CollectionType(plan.Type.ValueString())
	}

	out, err := conn.CreateCollection(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionCreating, ResNameCollection, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan
	state.Arn = flex.StringToFramework(ctx, out.CreateCollectionDetail.Arn)
	state.Description = flex.StringToFramework(ctx, out.CreateCollectionDetail.Description)
	state.ID = flex.StringToFramework(ctx, out.CreateCollectionDetail.Id)
	state.Name = flex.StringToFramework(ctx, out.CreateCollectionDetail.Name)
	state.Type = flex.StringValueToFramework(ctx, out.CreateCollectionDetail.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCollection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient()

	var state resourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCollectionByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	state.Arn = flex.StringToFramework(ctx, out.Arn)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.ID = flex.StringToFramework(ctx, out.Id)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.Type = flex.StringValueToFramework(ctx, out.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCollection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient()

	var plan, state resourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) {
		input := &opensearchserverless.UpdateCollectionInput{
			ClientToken: aws.String(id.UniqueId()),
			Id:          flex.StringFromFramework(ctx, plan.ID),
			Description: flex.StringFromFramework(ctx, plan.Description),
		}

		out, err := conn.UpdateCollection(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating Collection (%s)", plan.Name.ValueString()), err.Error())
			return
		}

		state.Arn = flex.StringToFramework(ctx, out.UpdateCollectionDetail.Arn)
		state.Description = flex.StringToFramework(ctx, out.UpdateCollectionDetail.Description)
		state.ID = flex.StringToFramework(ctx, out.UpdateCollectionDetail.Id)
		state.Name = flex.StringToFramework(ctx, out.UpdateCollectionDetail.Name)
		state.Type = flex.StringValueToFramework(ctx, out.UpdateCollectionDetail.Type)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCollection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient()

	var state resourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteCollection(ctx, &opensearchserverless.DeleteCollectionInput{
		ClientToken: aws.String(id.UniqueId()),
		Id:          aws.String(state.ID.ValueString()),
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionDeleting, ResNameCollection, state.Name.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceCollection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
