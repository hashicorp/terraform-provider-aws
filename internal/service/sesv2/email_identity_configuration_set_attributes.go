// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sesv2_email_identity_configuration_set_attributes", name="Email Identity Configuration Set Attributes")
func newEmailIdentityConfigurationSetAttributesResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &emailIdentityConfigurationSetAttributesResource{}
	return r, nil
}

const (
	ResNameEmailIdentityConfigurationSetAttributes = "Email Identity Configuration Set Attributes"
)

type emailIdentityConfigurationSetAttributesResource struct {
	framework.ResourceWithModel[emailIdentityConfigurationSetAttributesResourceModel]
	framework.WithImportByID
}

func (r *emailIdentityConfigurationSetAttributesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"configuration_set_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"email_identity": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (r *emailIdentityConfigurationSetAttributesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emailIdentityConfigurationSetAttributesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	emailIdentity := plan.EmailIdentity.ValueString()
	input := &sesv2.PutEmailIdentityConfigurationSetAttributesInput{
		EmailIdentity:        aws.String(emailIdentity),
		ConfigurationSetName: plan.ConfigurationSetName.ValueStringPointer(),
	}

	_, err := conn.PutEmailIdentityConfigurationSetAttributes(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("creating SESv2 Email Identity Configuration Set Attributes (%s)", emailIdentity),
			err.Error(),
		)
		return
	}

	plan.ID = fwflex.StringValueToFramework(ctx, emailIdentity)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *emailIdentityConfigurationSetAttributesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state emailIdentityConfigurationSetAttributesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	out, err := findEmailIdentityByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("reading SESv2 Email Identity Configuration Set Attributes (%s)", state.ID.ValueString()),
			err.Error(),
		)
		return
	}

	state.EmailIdentity = state.ID
	state.ConfigurationSetName = fwflex.StringToFramework(ctx, out.ConfigurationSetName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *emailIdentityConfigurationSetAttributesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan emailIdentityConfigurationSetAttributesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	input := &sesv2.PutEmailIdentityConfigurationSetAttributesInput{
		EmailIdentity:        plan.ID.ValueStringPointer(),
		ConfigurationSetName: plan.ConfigurationSetName.ValueStringPointer(),
	}

	_, err := conn.PutEmailIdentityConfigurationSetAttributes(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("updating SESv2 Email Identity Configuration Set Attributes (%s)", plan.ID.ValueString()),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *emailIdentityConfigurationSetAttributesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state emailIdentityConfigurationSetAttributesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SESV2Client(ctx)

	_, err := conn.PutEmailIdentityConfigurationSetAttributes(ctx, &sesv2.PutEmailIdentityConfigurationSetAttributesInput{
		EmailIdentity:        state.ID.ValueStringPointer(),
		ConfigurationSetName: aws.String(""),
	})
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("deleting SESv2 Email Identity Configuration Set Attributes (%s)", state.ID.ValueString()),
			err.Error(),
		)
		return
	}
}

type emailIdentityConfigurationSetAttributesResourceModel struct {
	framework.WithRegionModel
	ConfigurationSetName types.String `tfsdk:"configuration_set_name"`
	EmailIdentity        types.String `tfsdk:"email_identity"`
	ID                   types.String `tfsdk:"id"`
}
