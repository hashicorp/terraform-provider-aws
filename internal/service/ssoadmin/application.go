// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Application")
// @Tags
func newResourceApplication(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceApplication{}, nil
}

const (
	ResNameApplication = "Application"
)

type resourceApplication struct {
	framework.ResourceWithConfigure
}

func (r *resourceApplication) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ssoadmin_application"
}

func (r *resourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_account": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_provider_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"client_token": schema.StringAttribute{
				Optional: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.ApplicationStatus](),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"portal_options": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"visibility": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.ApplicationVisibility](),
								// If explicitly set, require that sign_in_options also be configured
								// to ensure the flattener correctly reads both values into state.
								stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("sign_in_options")),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"sign_in_options": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"origin": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.SignInOrigin](),
										},
									},
									"application_url": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan resourceApplicationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.CreateApplicationInput{
		ApplicationProviderArn: aws.String(plan.ApplicationProviderARN.ValueString()),
		InstanceArn:            aws.String(plan.InstanceARN.ValueString()),
		Name:                   aws.String(plan.Name.ValueString()),
		Tags:                   getTagsIn(ctx),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}
	if !plan.PortalOptions.IsNull() {
		var tfList []portalOptionsData
		resp.Diagnostics.Append(plan.PortalOptions.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		portalOptions, d := expandPortalOptions(ctx, tfList)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.PortalOptions = portalOptions
	}
	if !plan.Status.IsNull() {
		in.Status = awstypes.ApplicationStatus(plan.Status.ValueString())
	}

	out, err := conn.CreateApplication(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplication, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplication, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ApplicationARN = flex.StringToFrameworkARN(ctx, out.ApplicationArn)
	plan.ID = flex.StringToFramework(ctx, out.ApplicationArn)

	// Read after create to get computed attributes omitted from the create response
	readOut, err := findApplicationByID(ctx, conn, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplication, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	plan.ApplicationAccount = flex.StringToFramework(ctx, readOut.ApplicationAccount)
	plan.Status = flex.StringValueToFramework(ctx, readOut.Status)
	portalOptions, d := flattenPortalOptions(ctx, readOut.PortalOptions)
	resp.Diagnostics.Append(d...)
	plan.PortalOptions = portalOptions

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ApplicationAccount = flex.StringToFramework(ctx, out.ApplicationAccount)
	state.ApplicationARN = flex.StringToFrameworkARN(ctx, out.ApplicationArn)
	state.ApplicationProviderARN = flex.StringToFrameworkARN(ctx, out.ApplicationProviderArn)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.ID = flex.StringToFramework(ctx, out.ApplicationArn)
	state.InstanceARN = flex.StringToFrameworkARN(ctx, out.InstanceArn)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.Status = flex.StringValueToFramework(ctx, out.Status)

	portalOptions, d := flattenPortalOptions(ctx, out.PortalOptions)
	resp.Diagnostics.Append(d...)
	state.PortalOptions = portalOptions

	// listTags requires both application and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	tags, err := listTags(ctx, conn, state.ApplicationARN.ValueString(), state.InstanceARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	setTagsOut(ctx, Tags(tags))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan, state resourceApplicationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.PortalOptions.Equal(state.PortalOptions) ||
		!plan.Status.Equal(state.Status) {
		in := &ssoadmin.UpdateApplicationInput{
			ApplicationArn: aws.String(plan.ApplicationARN.ValueString()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}
		if !plan.Name.IsNull() {
			in.Name = aws.String(plan.Name.ValueString())
		}
		if !plan.PortalOptions.IsNull() {
			var tfList []portalOptionsData
			resp.Diagnostics.Append(plan.PortalOptions.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			portalOptions, d := expandPortalOptionsUpdate(ctx, tfList)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.PortalOptions = portalOptions
		}
		if !plan.Status.IsNull() {
			in.Status = awstypes.ApplicationStatus(plan.Status.ValueString())
		}

		out, err := conn.UpdateApplication(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionUpdating, ResNameApplication, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionUpdating, ResNameApplication, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	// updateTags requires both application and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	if oldTagsAll, newTagsAll := state.TagsAll, plan.TagsAll; !newTagsAll.Equal(oldTagsAll) {
		if err := updateTags(ctx, conn, plan.ApplicationARN.ValueString(), plan.InstanceARN.ValueString(), oldTagsAll, newTagsAll); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionUpdating, ResNameApplication, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.DeleteApplicationInput{
		ApplicationArn: aws.String(state.ApplicationARN.ValueString()),
	}

	_, err := conn.DeleteApplication(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionDeleting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceApplication) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func findApplicationByID(ctx context.Context, conn *ssoadmin.Client, id string) (*ssoadmin.DescribeApplicationOutput, error) {
	in := &ssoadmin.DescribeApplicationInput{
		ApplicationArn: aws.String(id),
	}

	out, err := conn.DescribeApplication(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func flattenPortalOptions(ctx context.Context, apiObject *awstypes.PortalOptions) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: portalOptionsAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}
	// Skip writing to state if only the visibilty attribute is returned
	// to avoid a nested computed attribute causing a diff.
	if apiObject.SignInOptions == nil {
		return types.ListNull(elemType), diags
	}

	signInOptions, d := flattenSignInOptions(ctx, apiObject.SignInOptions)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"visibility":      flex.StringValueToFramework(ctx, apiObject.Visibility),
		"sign_in_options": signInOptions,
	}
	objVal, d := types.ObjectValue(portalOptionsAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenSignInOptions(ctx context.Context, apiObject *awstypes.SignInOptions) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: signInOptionsAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"application_url": flex.StringToFramework(ctx, apiObject.ApplicationUrl),
		"origin":          flex.StringValueToFramework(ctx, apiObject.Origin),
	}
	objVal, d := types.ObjectValue(signInOptionsAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func expandPortalOptions(ctx context.Context, tfList []portalOptionsData) (*awstypes.PortalOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	tfObj := tfList[0]

	var signInOptions []signInOptionsData
	diags.Append(tfObj.SignInOptions.ElementsAs(ctx, &signInOptions, false)...)

	apiObject := &awstypes.PortalOptions{
		Visibility:    awstypes.ApplicationVisibility(tfObj.Visibility.ValueString()),
		SignInOptions: expandSignInOptions(signInOptions),
	}

	return apiObject, diags
}

// expandPortalOptionsUpdate is a variant of the expander for update opertations which
// does not include the visibility argument.
func expandPortalOptionsUpdate(ctx context.Context, tfList []portalOptionsData) (*awstypes.UpdateApplicationPortalOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	tfObj := tfList[0]

	var signInOptions []signInOptionsData
	diags.Append(tfObj.SignInOptions.ElementsAs(ctx, &signInOptions, false)...)

	apiObject := &awstypes.UpdateApplicationPortalOptions{
		SignInOptions: expandSignInOptions(signInOptions),
	}

	return apiObject, diags
}

func expandSignInOptions(tfList []signInOptionsData) *awstypes.SignInOptions {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.SignInOptions{
		Origin: awstypes.SignInOrigin(tfObj.Origin.ValueString()),
	}

	if !tfObj.ApplicationURL.IsNull() {
		apiObject.ApplicationUrl = aws.String(tfObj.ApplicationURL.ValueString())
	}

	return apiObject
}

type resourceApplicationData struct {
	ApplicationAccount     types.String `tfsdk:"application_account"`
	ApplicationARN         fwtypes.ARN  `tfsdk:"application_arn"`
	ApplicationProviderARN fwtypes.ARN  `tfsdk:"application_provider_arn"`
	ClientToken            types.String `tfsdk:"client_token"`
	Description            types.String `tfsdk:"description"`
	ID                     types.String `tfsdk:"id"`
	InstanceARN            fwtypes.ARN  `tfsdk:"instance_arn"`
	Name                   types.String `tfsdk:"name"`
	PortalOptions          types.List   `tfsdk:"portal_options"`
	Status                 types.String `tfsdk:"status"`
	Tags                   types.Map    `tfsdk:"tags"`
	TagsAll                types.Map    `tfsdk:"tags_all"`
}

type portalOptionsData struct {
	SignInOptions types.List   `tfsdk:"sign_in_options"`
	Visibility    types.String `tfsdk:"visibility"`
}

type signInOptionsData struct {
	ApplicationURL types.String `tfsdk:"application_url"`
	Origin         types.String `tfsdk:"origin"`
}

var portalOptionsAttrTypes = map[string]attr.Type{
	"sign_in_options": types.ListType{ElemType: types.ObjectType{AttrTypes: signInOptionsAttrTypes}},
	"visibility":      types.StringType,
}

var signInOptionsAttrTypes = map[string]attr.Type{
	"application_url": types.StringType,
	"origin":          types.StringType,
}
