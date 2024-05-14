// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Proactive Engagement")
func newProactiveEngagementResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &proactiveEngagementResource{}, nil
}

type proactiveEngagementResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *proactiveEngagementResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_shield_proactive_engagement"
}

func (r *proactiveEngagementResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrEnabled: schema.BoolAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"emergency_contact": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[emergencyContactModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"contact_notes": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 1024),
							},
						},
						"email_address": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 150),
							},
						},
						"phone_number": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(`^\+[1-9]\d{1,14}$`), ""),
							},
						},
					},
				},
			},
		},
	}
}

func (r *proactiveEngagementResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data proactiveEngagementResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	input := &shield.AssociateProactiveEngagementDetailsInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.AssociateProactiveEngagementDetails(ctx, input)

	// "InvalidOperationException: Proactive engagement details are already associated with the subscription. Please use Enable/DisableProactiveEngagement APIs to update it's status".
	if err != nil && !errs.IsA[*awstypes.InvalidOperationException](err) {
		response.Diagnostics.AddError("creating Shield Proactive Engagement", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = types.StringValue(r.Meta().AccountID)

	response.Diagnostics.Append(updateEmergencyContactSettings(ctx, conn, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(putProactiveEngagementStatus(ctx, conn, data.Enabled.ValueBool())...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *proactiveEngagementResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data proactiveEngagementResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	subscription, err := findSubscription(ctx, conn)

	if err == nil && subscription.ProactiveEngagementStatus == "" {
		err = tfresource.NewEmptyResultError(nil)
	}

	var emergencyContacts []awstypes.EmergencyContact

	if err == nil {
		emergencyContacts, err = findEmergencyContactSettings(ctx, conn)
	}

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading Shield Proactive Engagement", err.Error())

		return
	}

	data.EmergencyContactList = fwtypes.NewListNestedObjectValueOfValueSliceMust[emergencyContactModel](ctx, tfslices.ApplyToAll(emergencyContacts, func(apiObject awstypes.EmergencyContact) emergencyContactModel {
		return emergencyContactModel{
			ContactNotes: fwflex.StringToFramework(ctx, apiObject.ContactNotes),
			EmailAddress: fwflex.StringToFramework(ctx, apiObject.EmailAddress),
			PhoneNumber:  fwflex.StringToFramework(ctx, apiObject.PhoneNumber),
		}
	}))
	data.Enabled = types.BoolValue(subscription.ProactiveEngagementStatus == awstypes.ProactiveEngagementStatusEnabled)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *proactiveEngagementResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new proactiveEngagementResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	if !new.EmergencyContactList.Equal(old.EmergencyContactList) {
		response.Diagnostics.Append(updateEmergencyContactSettings(ctx, conn, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if !new.Enabled.Equal(old.Enabled) {
		response.Diagnostics.Append(putProactiveEngagementStatus(ctx, conn, new.Enabled.ValueBool())...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *proactiveEngagementResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().ShieldClient(ctx)

	inputD := &shield.DisableProactiveEngagementInput{}

	_, err := conn.DisableProactiveEngagement(ctx, inputD)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError("disabling Shield proactive engagement", err.Error())

		return
	}

	inputU := &shield.UpdateEmergencyContactSettingsInput{
		EmergencyContactList: []awstypes.EmergencyContact{},
	}

	_, err = conn.UpdateEmergencyContactSettings(ctx, inputU)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError("updating Shield emergency contact settings", err.Error())

		return
	}
}

func disableProactiveEngagement(ctx context.Context, conn *shield.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	input := &shield.DisableProactiveEngagementInput{}

	_, err := conn.DisableProactiveEngagement(ctx, input)

	if err != nil {
		diags.AddError("disabling Shield proactive engagement", err.Error())

		return diags
	}

	return diags
}

func enableProactiveEngagement(ctx context.Context, conn *shield.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	input := &shield.EnableProactiveEngagementInput{}

	_, err := conn.EnableProactiveEngagement(ctx, input)

	if err != nil {
		diags.AddError("enabling Shield proactive engagement", err.Error())

		return diags
	}

	return diags
}

func putProactiveEngagementStatus(ctx context.Context, conn *shield.Client, enabled bool) diag.Diagnostics {
	var diags diag.Diagnostics

	if enabled {
		diags.Append(enableProactiveEngagement(ctx, conn)...)
	} else {
		diags.Append(disableProactiveEngagement(ctx, conn)...)
	}

	return diags
}

func updateEmergencyContactSettings(ctx context.Context, conn *shield.Client, data *proactiveEngagementResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	input := &shield.UpdateEmergencyContactSettingsInput{}

	diags.Append(fwflex.Expand(ctx, data, input)...)
	if diags.HasError() {
		return diags
	}

	_, err := conn.UpdateEmergencyContactSettings(ctx, input)

	if err != nil {
		diags.AddError("updating Shield emergency contact settings", err.Error())

		return diags
	}

	return diags
}

func findEmergencyContactSettings(ctx context.Context, conn *shield.Client) ([]awstypes.EmergencyContact, error) {
	input := &shield.DescribeEmergencyContactSettingsInput{}

	output, err := conn.DescribeEmergencyContactSettings(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.EmergencyContactList) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.EmergencyContactList, nil
}

func findSubscription(ctx context.Context, conn *shield.Client) (*awstypes.Subscription, error) {
	input := &shield.DescribeSubscriptionInput{}

	output, err := conn.DescribeSubscription(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Subscription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Subscription, nil
}

type proactiveEngagementResourceModel struct {
	EmergencyContactList fwtypes.ListNestedObjectValueOf[emergencyContactModel] `tfsdk:"emergency_contact"`
	Enabled              types.Bool                                             `tfsdk:"enabled"`
	ID                   types.String                                           `tfsdk:"id"`
}

type emergencyContactModel struct {
	ContactNotes types.String `tfsdk:"contact_notes"`
	EmailAddress types.String `tfsdk:"email_address"`
	PhoneNumber  types.String `tfsdk:"phone_number"`
}
