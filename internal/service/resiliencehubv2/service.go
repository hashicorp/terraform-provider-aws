// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehubv2_service", name="Service")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.Service")
// @Testing(hasNoPreExistingResource=true)
func newResourceService(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceService{}, nil
}

type resourceService struct {
	framework.ResourceWithModel[resourceServiceModel]
	framework.WithImportByIdentity
}

func (r *resourceService) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
			},
			"policy_arn": fwschema.StringAttribute{
				Optional: true,
			},
			"regions": fwschema.ListAttribute{
				Required:   true,
				CustomType: fwtypes.ListOfStringType,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"permission_model": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[permissionModelModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"invoker_role_name": fwschema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceService) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceServiceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.CreateServiceInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateService(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.Service, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ARN = types.StringValue(aws.ToString(output.Service.ServiceArn))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceService) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceServiceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	svc, err := findServiceByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, svc, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.ARN = types.StringValue(aws.ToString(svc.ServiceArn))

	tags, err := listTags(ctx, conn, state.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
	setTagsOut(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *resourceService) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceServiceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.UpdateServiceInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ServiceArn = state.ARN.ValueStringPointer()

	output, err := conn.UpdateService(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.Service, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ARN = types.StringValue(aws.ToString(output.Service.ServiceArn))

	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := updateTags(ctx, conn, state.ARN.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceService) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceServiceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.DeleteServiceInput{
		ServiceArn: state.ARN.ValueStringPointer(),
	}
	_, err := conn.DeleteService(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
	}
}

func findServiceByARN(ctx context.Context, conn *resiliencehubv2.Client, arn string) (*awstypes.Service, error) {
	input := resiliencehubv2.GetServiceInput{
		ServiceArn: aws.String(arn),
	}
	output, err := conn.GetService(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		return nil, smarterr.NewError(err)
	}
	if output == nil || output.Service == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return output.Service, nil
}

type resourceServiceModel struct {
	framework.WithRegionModel
	ARN             types.String                                          `tfsdk:"arn"`
	Description     types.String                                          `tfsdk:"description"`
	Name            types.String                                          `tfsdk:"name"`
	PermissionModel fwtypes.ListNestedObjectValueOf[permissionModelModel] `tfsdk:"permission_model"`
	PolicyArn       types.String                                          `tfsdk:"policy_arn"`
	Regions         fwtypes.ListOfString                                  `tfsdk:"regions"`
	Tags            tftags.Map                                            `tfsdk:"tags"`
	TagsAll         tftags.Map                                            `tfsdk:"tags_all"`
}

type permissionModelModel struct {
	InvokerRoleName types.String `tfsdk:"invoker_role_name"`
}
