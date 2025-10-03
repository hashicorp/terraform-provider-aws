// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_transfer_web_app_customization", name="Web App Customization")
func newResourceWebAppCustomization(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceWebAppCustomization{}

	return r, nil
}

const (
	ResNameWebAppCustomization = "Web App Customization"
)

type resourceWebAppCustomization struct {
	framework.ResourceWithModel[resourceWebAppCustomizationModel]
}

func (r *resourceWebAppCustomization) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"favicon_file": schema.StringAttribute{
				// If faviconFile is not specified when calling the UpdateWebAppCustomization API,
				// the existing favicon remains unchanged.
				// Therefore, this field is marked as Optional and Computed.
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 20960),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(), // TODO TEMP Remove
			"logo_file": schema.StringAttribute{
				// Same as favicon_file
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 51200),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"web_app_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceWebAppCustomization) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().TransferClient(ctx)

	var plan resourceWebAppCustomizationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input transfer.UpdateWebAppCustomizationInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.UpdateWebAppCustomization(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionCreating, ResNameWebAppCustomization, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionCreating, ResNameWebAppCustomization, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.WebAppId)

	rout, _ := findWebAppCustomizationByID(ctx, conn, plan.ID.ValueString())
	resp.Diagnostics.Append(flex.Flatten(ctx, rout, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceWebAppCustomization) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().TransferClient(ctx)

	var state resourceWebAppCustomizationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findWebAppCustomizationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionReading, ResNameWebAppCustomization, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = flex.StringToFramework(ctx, out.WebAppId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceWebAppCustomization) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().TransferClient(ctx)

	var plan, state resourceWebAppCustomizationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input transfer.UpdateWebAppCustomizationInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateWebAppCustomization(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Transfer, create.ErrActionUpdating, ResNameWebAppCustomization, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Transfer, create.ErrActionUpdating, ResNameWebAppCustomization, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	rout, _ := findWebAppCustomizationByID(ctx, conn, plan.ID.ValueString())
	resp.Diagnostics.Append(flex.Flatten(ctx, rout, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceWebAppCustomization) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().TransferClient(ctx)

	var state resourceWebAppCustomizationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := transfer.DeleteWebAppCustomizationInput{
		WebAppId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteWebAppCustomization(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionDeleting, ResNameWebAppCustomization, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceWebAppCustomization) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("web_app_id"), request.ID)...)
}

func findWebAppCustomizationByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedWebAppCustomization, error) {
	input := transfer.DescribeWebAppCustomizationInput{
		WebAppId: aws.String(id),
	}

	out, err := conn.DescribeWebAppCustomization(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.WebAppCustomization == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.WebAppCustomization, nil
}

type resourceWebAppCustomizationModel struct {
	framework.WithRegionModel
	ARN         types.String `tfsdk:"arn"`
	FaviconFile types.String `tfsdk:"favicon_file"`
	LogoFile    types.String `tfsdk:"logo_file"`
	Title       types.String `tfsdk:"title"`
	WebAppID    types.String `tfsdk:"web_app_id"`

	// TEMP TODO Remove
	ID types.String `tfsdk:"id"`
}

var (
	_ flex.Expander  = resourceWebAppCustomizationModel{}
	_ flex.Flattener = &resourceWebAppCustomizationModel{}
)

func (m resourceWebAppCustomizationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var input transfer.UpdateWebAppCustomizationInput
	var diags diag.Diagnostics
	input.WebAppId = m.WebAppID.ValueStringPointer()

	if !m.FaviconFile.IsNull() && m.FaviconFile.ValueString() != "" {
		if v, err := itypes.Base64Decode(m.FaviconFile.ValueString()); err != nil {
			diags.AddError(
				"Favicon File Decode Error",
				"An unexpected error occurred while decoding the Favicon File. ",
			)
		} else {
			input.FaviconFile = v
		}
	} else {
		input.FaviconFile = nil
	}
	if !m.LogoFile.IsNull() && m.LogoFile.ValueString() != "" {
		if v, err := itypes.Base64Decode(m.LogoFile.ValueString()); err != nil {
			diags.AddError(
				"Logo File Decode Error",
				"An unexpected error occurred while decoding the Logo File. ",
			)
		} else {
			input.LogoFile = v
		}
	} else {
		input.LogoFile = nil
	}
	if !m.Title.IsNull() && m.Title.ValueString() != "" {
		input.Title = m.Title.ValueStringPointer()
	} else {
		input.Title = aws.String("")
	}
	return &input, nil
}

func (m *resourceWebAppCustomizationModel) Flatten(ctx context.Context, in any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := in.(type) {
	case awstypes.DescribedWebAppCustomization:
		m.ARN = flex.StringToFramework(ctx, t.Arn)
		m.FaviconFile = flex.StringToFramework(ctx, aws.String(itypes.Base64Encode(t.FaviconFile)))
		m.ID = flex.StringToFramework(ctx, t.WebAppId)
		m.LogoFile = flex.StringToFramework(ctx, aws.String(itypes.Base64Encode(t.LogoFile)))
		m.Title = flex.StringToFramework(ctx, t.Title)
		m.WebAppID = flex.StringToFramework(ctx, t.WebAppId)
	case transfer.UpdateWebAppCustomizationOutput:
		m.WebAppID = flex.StringToFramework(ctx, t.WebAppId)
	default:
		diags.AddError("Interface Conversion Error", fmt.Sprintf("cannot flatten %T into %T", in, m))
	}
	return diags
}
