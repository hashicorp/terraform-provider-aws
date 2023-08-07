package shield

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/shield"
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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"

	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="DRT Access Role Arn Association")
func newResourceDRTAccessRoleArnAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDRTAccessRoleArnAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDRTAccessRoleArnAssociation = "DRT Access Role Arn Association"
)

type resourceDRTAccessRoleArnAssociation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDRTAccessRoleArnAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_shield_drt_access_role_arn_association"
}

func (r *resourceDRTAccessRoleArnAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{ // required by hashicorps terraform plugin testing framework
				DeprecationMessage:  "id is only for framework compatability and not used by the provider",
				MarkdownDescription: "The ID of the directory.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_arn": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^arn:aws:iam::\d{12}:role/?[a-zA-Z_0-9+=,.@\-_/]+`),
						"must match arn pattern",
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Read:   true,
			}),
		},
	}
}

func (r *resourceDRTAccessRoleArnAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var plan resourceDRTAccessRoleArnAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {

		return
	}

	in := &shield.AssociateDRTRoleInput{
		RoleArn: aws.String(plan.RoleArn.ValueString()),
	}

	out, err := conn.AssociateDRTRoleWithContext(ctx, in)
	if err != nil {

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameDRTAccessRoleArnAssociation, plan.RoleArn.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameDRTAccessRoleArnAssociation, plan.RoleArn.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDRTAccessRoleArnAssociationCreated(ctx, conn, plan.RoleArn.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForCreation, ResNameDRTAccessRoleArnAssociation, plan.RoleArn.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(plan.RoleArn.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDRTAccessRoleArnAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	conn := r.Meta().ShieldConn(ctx)

	var state resourceDRTAccessRoleArnAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &shield.DescribeDRTAccessInput{}

	out, err := conn.DescribeDRTAccessWithContext(ctx, in)

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionSetting, ResNameDRTAccessRoleArnAssociation, state.RoleArn.String(), err),
			err.Error(),
		)
		return
	}
	if state.ID.IsNull() || state.ID.IsUnknown() {
		// Setting ID of state - required by hashicorps terraform plugin testing framework for Import. See issue https://github.com/hashicorp/terraform-plugin-testing/issues/84
		state.ID = types.StringValue(fmt.Sprintf("%s", state.RoleArn.ValueString()))
	}

	state.RoleArn = flex.StringToFramework(ctx, out.RoleArn)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDRTAccessRoleArnAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ShieldConn(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceDRTAccessRoleArnAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.RoleArn.Equal(state.RoleArn) {

		in := &shield.AssociateDRTRoleInput{
			RoleArn: aws.String(plan.RoleArn.ValueString()),
		}

		out, err := conn.AssociateDRTRoleWithContext(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameDRTAccessRoleArnAssociation, plan.RoleArn.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameDRTAccessRoleArnAssociation, plan.RoleArn.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitDRTAccessRoleArnAssociationUpdated(ctx, conn, plan.RoleArn.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForUpdate, ResNameDRTAccessRoleArnAssociation, plan.RoleArn.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDRTAccessRoleArnAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var state resourceDRTAccessRoleArnAssociationData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := &shield.DisassociateDRTRoleInput{}

	_, err := conn.DisassociateDRTRoleWithContext(ctx, in)
	if err != nil {
		var nfe *shield.ResourceNotFoundException
		if errors.As(err, &nfe) {

			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionDeleting, ResNameDRTAccessRoleArnAssociation, state.RoleArn.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDRTAccessRoleArnAssociationDeleted(ctx, conn, state.RoleArn.ValueString(), deleteTimeout)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForDeletion, ResNameDRTAccessRoleArnAssociation, state.RoleArn.String(), err),
			err.Error(),
		)
		return
	}

}

func waitDRTAccessRoleArnAssociationCreated(ctx context.Context, conn *shield.Shield, roleArn string, timeout time.Duration) (*shield.DescribeDRTAccessOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusDRTAccessRoleArnAssociation(ctx, conn, roleArn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeDRTAccessOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDRTAccessRoleArnAssociationUpdated(ctx context.Context, conn *shield.Shield, roleArn string, timeout time.Duration) (*shield.DescribeDRTAccessOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusDRTAccessRoleArnAssociation(ctx, conn, roleArn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeDRTAccessOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDRTAccessRoleArnAssociationDeleted(ctx context.Context, conn *shield.Shield, roleArn string, timeout time.Duration) (*shield.DescribeDRTAccessOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusDRTAccessRoleArnAssociationDeleted(ctx, conn, roleArn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*shield.DescribeDRTAccessOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDRTAccessRoleArnAssociation(ctx context.Context, conn *shield.Shield, roleArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := describeDRTAccessRoleArnAssociation(ctx, conn, roleArn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func statusDRTAccessRoleArnAssociationDeleted(ctx context.Context, conn *shield.Shield, roleArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := describeDRTAccessRoleArnAssociation(ctx, conn, roleArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if out.RoleArn != nil && aws.ToString(out.RoleArn) == roleArn {
			return out, statusDeleting, nil
		}

		return out, statusDeleting, nil
	}
}

func describeDRTAccessRoleArnAssociation(ctx context.Context, conn *shield.Shield, roleArn string) (*shield.DescribeDRTAccessOutput, error) {
	in := &shield.DescribeDRTAccessInput{}

	out, err := conn.DescribeDRTAccessWithContext(ctx, in)
	if err != nil {
		var nfe *shield.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
	}

	if out == nil || out.RoleArn == nil || aws.ToString(out.RoleArn) != roleArn {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceDRTAccessRoleArnAssociationData struct {
	ID       types.String   `tfsdk:"id"`
	RoleArn  types.String   `tfsdk:"role_arn"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
