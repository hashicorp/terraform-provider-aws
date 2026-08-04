// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package transfer

import (
	"context"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @FrameworkResource("aws_transfer_web_app_customization", name="Web App Customization")
func newWebAppCustomizationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &webAppCustomizationResource{}

	return r, nil
}

type webAppCustomizationResource struct {
	framework.ResourceWithModel[webAppCustomizationResourceModel]
}

func (r *webAppCustomizationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			"logo_file": schema.StringAttribute{
				// Same as favicon_file.
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

func (r *webAppCustomizationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data webAppCustomizationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	var input transfer.UpdateWebAppCustomizationInput
	response.Diagnostics.Append(expandUpdateWebAppCustomizationInput(ctx, &data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	webAppID := fwflex.StringValueFromFramework(ctx, data.WebAppID)
	_, err := conn.UpdateWebAppCustomization(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Transfer Web App (%s) Customization", webAppID), err.Error())

		return
	}

	if data.FaviconFile.IsUnknown() {
		data.FaviconFile = types.StringNull()
	}
	if data.LogoFile.IsUnknown() {
		data.LogoFile = types.StringNull()
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *webAppCustomizationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data webAppCustomizationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	webAppID := fwflex.StringValueFromFramework(ctx, data.WebAppID)
	out, err := findWebAppCustomizationByID(ctx, conn, webAppID)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Transfer Web App (%s) Customization", webAppID), err.Error())

		return
	}

	flattenDescribedWebAppCustomization(ctx, out, &data)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *webAppCustomizationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old webAppCustomizationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input transfer.UpdateWebAppCustomizationInput
		response.Diagnostics.Append(expandUpdateWebAppCustomizationInput(ctx, &new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		webAppID := fwflex.StringValueFromFramework(ctx, new.WebAppID)
		_, err := conn.UpdateWebAppCustomization(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Transfer Web App (%s) Customization", webAppID), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *webAppCustomizationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data webAppCustomizationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TransferClient(ctx)

	webAppID := fwflex.StringValueFromFramework(ctx, data.WebAppID)
	input := transfer.DeleteWebAppCustomizationInput{
		WebAppId: aws.String(webAppID),
	}
	_, err := conn.DeleteWebAppCustomization(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Transfer Web App (%s) Customization", webAppID), err.Error())

		return
	}
}

func (r *webAppCustomizationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("web_app_id"), request, response)
}

func findWebAppCustomizationByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedWebAppCustomization, error) {
	input := transfer.DescribeWebAppCustomizationInput{
		WebAppId: aws.String(id),
	}

	return findWebAppCustomization(ctx, conn, &input)
}

func findWebAppCustomization(ctx context.Context, conn *transfer.Client, input *transfer.DescribeWebAppCustomizationInput) (*awstypes.DescribedWebAppCustomization, error) {
	out, err := conn.DescribeWebAppCustomization(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.WebAppCustomization == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.WebAppCustomization, nil
}

type webAppCustomizationResourceModel struct {
	framework.WithRegionModel
	FaviconFile types.String `tfsdk:"favicon_file"`
	LogoFile    types.String `tfsdk:"logo_file"`
	Title       types.String `tfsdk:"title"`
	WebAppID    types.String `tfsdk:"web_app_id"`
}

func expandUpdateWebAppCustomizationInput(ctx context.Context, data *webAppCustomizationResourceModel, apiObject *transfer.UpdateWebAppCustomizationInput) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	if !data.FaviconFile.IsNull() && !data.FaviconFile.IsUnknown() {
		if v, err := inttypes.Base64Decode(fwflex.StringValueFromFramework(ctx, data.FaviconFile)); err != nil {
			diags.AddError("Favicon File Decode Error", err.Error())
		} else {
			apiObject.FaviconFile = v
		}
	}

	if !data.LogoFile.IsNull() && !data.LogoFile.IsUnknown() {
		if v, err := inttypes.Base64Decode(fwflex.StringValueFromFramework(ctx, data.LogoFile)); err != nil {
			diags.AddError("Logo File Decode Error", err.Error())
		} else {
			apiObject.LogoFile = v
		}
	}

	if !data.Title.IsNull() {
		apiObject.Title = fwflex.StringFromFramework(ctx, data.Title)
	} else {
		apiObject.Title = aws.String("")
	}

	apiObject.WebAppId = fwflex.StringFromFramework(ctx, data.WebAppID)

	return diags
}

func flattenDescribedWebAppCustomization(ctx context.Context, apiObject *awstypes.DescribedWebAppCustomization, data *webAppCustomizationResourceModel) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	if v := apiObject.FaviconFile; v != nil {
		data.FaviconFile = fwflex.StringToFramework(ctx, aws.String(inttypes.Base64Encode(v)))
	} else {
		data.FaviconFile = types.StringNull()
	}
	if v := apiObject.LogoFile; v != nil {
		data.LogoFile = fwflex.StringToFramework(ctx, aws.String(inttypes.Base64Encode(v)))
	} else {
		data.LogoFile = types.StringNull()
	}
	data.Title = fwflex.StringToFramework(ctx, apiObject.Title)
	data.WebAppID = fwflex.StringToFramework(ctx, apiObject.WebAppId)
}
