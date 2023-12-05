package shield

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Proactive Engagement Association")
func newResourceProactiveEngagementAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProactiveEngagementAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameProactiveEngagementAssociation = "Proactive Engagement Association"
	emergencyContactsEmailKey             = "email_address"
	emergencyContactsNotesKey             = "contact_notes"
	emergencyContactsPhoneKey             = "phone_number"
)

type resourceProactiveEngagementAssociation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceProactiveEngagementAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_shield_proactive_engagement_association"
}

func (r *resourceProactiveEngagementAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"enabled": schema.BoolAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"emergency_contacts": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"contact_notes": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"email_address": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"phone_number": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceProactiveEngagementAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var plan resourceProactiveEngagementAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var emergencyContactsList []*shield.EmergencyContact
	if plan.Enabled.ValueBool() {
		if !plan.EmergencyContacts.IsNull() {
			var emergencyContacts []emergencyContactData
			resp.Diagnostics.Append(plan.EmergencyContacts.ElementsAs(ctx, &emergencyContacts, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			emergencyContactsList = expandEmergencyContacts(ctx, emergencyContacts)
		} else {
			err := errors.New("At least one emergency_contacts block is required.")
			resp.Diagnostics.AddError(create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		describeIn := &shield.DescribeEmergencyContactSettingsInput{}
		eContactSettings, err := conn.DescribeEmergencyContactSettingsWithContext(ctx, describeIn)

		if err == nil && len(eContactSettings.EmergencyContactList) == 0 {
			r.executeUpdateExistingAssociation(ctx, req, resp, plan, conn, emergencyContactsList)
		} else if err != nil {
			var ioe *shield.InvalidOperationException
			var nfe *shield.ResourceNotFoundException
			if errors.As(err, &ioe) && strings.Contains(ioe.Message(), "Enable/DisableProactiveEngagement") {
				r.executeUpdateExistingAssociation(ctx, req, resp, plan, conn, emergencyContactsList)
			} else if errors.As(err, &nfe) {
				r.executeCreateNewAssociation(ctx, req, resp, plan, conn, emergencyContactsList)
			} else {
				resp.Diagnostics.AddError(create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
					err.Error(),
				)
				return
			}
		} else {
			r.executeUpdateExistingAssociation(ctx, req, resp, plan, conn, emergencyContactsList)
		}

		createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
		_, err = waitProactiveEngagementAssociationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForCreation, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	plan.ID = types.StringValue(r.Meta().AccountID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceProactiveEngagementAssociation) executeUpdateExistingAssociation(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, plan resourceProactiveEngagementAssociationData, conn *shield.Shield, emergencyContactsList []*shield.EmergencyContact) {
	in := &shield.UpdateEmergencyContactSettingsInput{
		EmergencyContactList: emergencyContactsList,
	}
	_, err := conn.UpdateEmergencyContactSettingsWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	enableIn := &shield.EnableProactiveEngagementInput{}
	_, err = conn.EnableProactiveEngagementWithContext(ctx, enableIn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceProactiveEngagementAssociation) executeCreateNewAssociation(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, plan resourceProactiveEngagementAssociationData, conn *shield.Shield, emergencyContactsList []*shield.EmergencyContact) {
	in := &shield.AssociateProactiveEngagementDetailsInput{
		EmergencyContactList: emergencyContactsList,
	}
	in.EmergencyContactList = emergencyContactsList
	_, err := conn.AssociateProactiveEngagementDetailsWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceProactiveEngagementAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ShieldConn(ctx)
	var state resourceProactiveEngagementAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &shield.DescribeEmergencyContactSettingsInput{}

	out, err := conn.DescribeEmergencyContactSettingsWithContext(ctx, in)

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionSetting, ResNameProactiveEngagementAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if out != nil && out.EmergencyContactList != nil {
		emergencyContacts, d := flattenEmergencyContacts(ctx, out.EmergencyContactList)
		resp.Diagnostics.Append(d...)
		state.EmergencyContacts = emergencyContacts
	}
	state.Enabled = types.BoolValue(true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceProactiveEngagementAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var plan, state resourceProactiveEngagementAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &shield.UpdateEmergencyContactSettingsInput{}
	var emergencyContactsList []*shield.EmergencyContact
	if !plan.EmergencyContacts.IsNull() {
		var emergencyContacts []emergencyContactData
		resp.Diagnostics.Append(plan.EmergencyContacts.ElementsAs(ctx, &emergencyContacts, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		emergencyContactsList = expandEmergencyContacts(ctx, emergencyContacts)
	} else {
		err := errors.New("At least one emergency_contacts block is required.")
		resp.Diagnostics.AddError(create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	in.EmergencyContactList = emergencyContactsList

	_, err := conn.UpdateEmergencyContactSettingsWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	if plan.Enabled.ValueBool() && len(emergencyContactsList) > 0 {
		_, err = conn.EnableProactiveEngagementWithContext(ctx, &shield.EnableProactiveEngagementInput{})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitProactiveEngagementAssociationUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForUpdate, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	} else {
		_, err = conn.DisableProactiveEngagementWithContext(ctx, &shield.DisableProactiveEngagementInput{})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitProactiveEngagementAssociationDeleted(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForUpdate, ResNameProactiveEngagementAssociation, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceProactiveEngagementAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var state resourceProactiveEngagementAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &shield.DisableProactiveEngagementInput{}

	_, err := conn.DisableProactiveEngagementWithContext(ctx, in)
	if err != nil {
		var nfe *shield.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionDeleting, ResNameProactiveEngagementAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	updateIn := &shield.UpdateEmergencyContactSettingsInput{}
	updateIn.EmergencyContactList = []*shield.EmergencyContact{}
	_, err = conn.UpdateEmergencyContactSettingsWithContext(ctx, updateIn)
	if err != nil {
		var nfe *shield.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionDeleting, ResNameProactiveEngagementAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitProactiveEngagementAssociationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForDeletion, ResNameProactiveEngagementAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func waitProactiveEngagementAssociationCreated(ctx context.Context, conn *shield.Shield, id string, timeout time.Duration) (*shield.DescribeEmergencyContactSettingsOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusProactiveEngagementAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            2,
		ContinuousTargetOccurence: 2,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeEmergencyContactSettingsOutput); ok {
		return out, err
	}
	return nil, err
}

func waitProactiveEngagementAssociationUpdated(ctx context.Context, conn *shield.Shield, id string, timeout time.Duration) (*shield.DescribeEmergencyContactSettingsOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated, statusNormal},
		Refresh:                   statusProactiveEngagementAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeEmergencyContactSettingsOutput); ok {
		return out, err
	}

	return nil, err
}

func waitProactiveEngagementAssociationDeleted(ctx context.Context, conn *shield.Shield, id string, timeout time.Duration) (*shield.DescribeEmergencyContactSettingsOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusProactiveEngagementAssociationDeleted(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeEmergencyContactSettingsOutput); ok {
		return out, err
	}
	return nil, err
}

func statusProactiveEngagementAssociation(ctx context.Context, conn *shield.Shield, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := describeEmergencyContactSettings(ctx, conn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		return out, statusNormal, nil
	}
}

func statusProactiveEngagementAssociationDeleted(ctx context.Context, conn *shield.Shield, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := describeEmergencyContactSettings(ctx, conn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if out == nil || out.EmergencyContactList == nil || len(out.EmergencyContactList) == 0 {
			return nil, "", nil
		}
		return out, statusNormal, nil
	}
}

func describeEmergencyContactSettings(ctx context.Context, conn *shield.Shield) (*shield.DescribeEmergencyContactSettingsOutput, error) {
	in := &shield.DescribeEmergencyContactSettingsInput{}

	out, err := conn.DescribeEmergencyContactSettingsWithContext(ctx, in)
	if err != nil {
		var nfe *shield.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
	}

	if out == nil || out.EmergencyContactList == nil || len(out.EmergencyContactList) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}
	return out, nil
}

// func flattenEmergencyContactList(ctx context.Context, apiObjects []*shield.EmergencyContact) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	attributeTypes := fwtypes.AttributeTypesMust[emergencyContactData](ctx)
// 	elemType := types.ObjectType{AttrTypes: attributeTypes}

// 	if len(apiObjects) == 0 {
// 		return types.ListNull(elemType), diags
// 	}

// 	elems := []attr.Value{}
// 	for _, apiObject := range apiObjects {
// 		if apiObject == nil {
// 			continue
// 		}

// 		obj := map[string]attr.Value{
// 			"email_address": flex.StringValueToFramework(ctx, *apiObject.EmailAddress),
// 			"phone_number":  flex.StringValueToFramework(ctx, *apiObject.PhoneNumber),
// 			"contact_notes": flex.StringValueToFramework(ctx, *apiObject.ContactNotes),
// 		}
// 		objVal, d := types.ObjectValue(attributeTypes, obj)
// 		diags.Append(d...)

// 		elems = append(elems, objVal)
// 	}

// 	listVal, d := types.ListValue(elemType, elems)
// 	diags.Append(d...)

// 	return listVal, diags
// }

func expandEmergencyContacts(ctx context.Context, tfList []emergencyContactData) []*shield.EmergencyContact {
	if len(tfList) == 0 {
		return nil
	}

	apiList := []*shield.EmergencyContact{}
	for _, tfObj := range tfList {
		apiObject := &shield.EmergencyContact{}

		if !tfObj.EmailAddress.IsNull() {
			apiObject.EmailAddress = flex.StringFromFramework(ctx, tfObj.EmailAddress)
		}
		if !tfObj.ContactNotes.IsNull() {
			apiObject.ContactNotes = flex.StringFromFramework(ctx, tfObj.ContactNotes)
		}
		if !tfObj.PhoneNumber.IsNull() {
			apiObject.PhoneNumber = flex.StringFromFramework(ctx, tfObj.PhoneNumber)
		}
		apiList = append(apiList, apiObject)
	}

	return apiList
}

func flattenEmergencyContacts(ctx context.Context, apiObject []*shield.EmergencyContact) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: emergencyContactsAttrTypes}

	elems := []attr.Value{}
	for _, t := range apiObject {
		obj := map[string]attr.Value{
			"email_address": flex.StringToFramework(ctx, t.EmailAddress),
			"contact_notes": flex.StringToFramework(ctx, t.ContactNotes),
			"phone_number":  flex.StringToFramework(ctx, t.PhoneNumber),
		}
		objVal, d := types.ObjectValue(emergencyContactsAttrTypes, obj)
		diags.Append(d...)
		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

var emergencyContactsAttrTypes = map[string]attr.Type{
	"email_address": types.StringType,
	"contact_notes": types.StringType,
	"phone_number":  types.StringType,
}

type emergencyContactData struct {
	EmailAddress types.String `tfsdk:"email_address"`
	ContactNotes types.String `tfsdk:"contact_notes"`
	PhoneNumber  types.String `tfsdk:"phone_number"`
}

type resourceProactiveEngagementAssociationData struct {
	ID                types.String   `tfsdk:"id"`
	Enabled           types.Bool     `tfsdk:"enabled"`
	EmergencyContacts types.List     `tfsdk:"emergency_contacts"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}
