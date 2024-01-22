// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/m2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Application")
// @Tags(identifierAttribute="arn")
func newResourceApplication(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceApplication{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameApplication = "Application"
)

type resourceApplication struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceApplication) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_m2_application"
}

func (r *resourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"current_version": schema.Int64Attribute{
				Computed: true,
			},
			"client_token": schema.StringAttribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"engine_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(enum.Values[awstypes.EngineType]()...),
				},
			},
			"kms_key_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"definition": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"content": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.ExactlyOneOf(
								path.MatchRelative().AtParent().AtName("content"),
								path.MatchRelative().AtParent().AtName("s3_location"),
							),
						},
					},
					"s3_location": schema.StringAttribute{
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().M2Client(ctx)

	var plan resourceApplicationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &m2.CreateApplicationInput{
		Tags: getTagsIn(ctx),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var definition applicationDefinition
	resp.Diagnostics.Append(plan.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{})...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiDefinition := expandApplicationDefinition(ctx, definition)
	in.Definition = apiDefinition

	out, err := conn.CreateApplication(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionCreating, ResNameApplication, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ApplicationId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionCreating, ResNameApplication, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.ApplicationArn)
	plan.ID = flex.StringToFramework(ctx, out.ApplicationId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	app, err := waitApplicationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForCreation, ResNameApplication, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(plan.refreshFromOutput(ctx, app)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().M2Client(ctx)

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
			create.ProblemStandardMessage(names.M2, create.ErrActionSetting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	version, err := findApplicationVersion(ctx, conn, state.ID.ValueString(), *out.LatestVersion.ApplicationVersion)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionSetting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, out)...)
	resp.Diagnostics.Append(state.refreshFromVersion(ctx, version)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().M2Client(ctx)

	var plan, state resourceApplicationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	update := false

	in := &m2.UpdateApplicationInput{
		ApplicationId:             flex.StringFromFramework(ctx, state.ID),
		CurrentApplicationVersion: flex.Int32FromFramework(ctx, state.CurrentVersion),
	}

	if !plan.Definition.Equal(state.Definition) {
		var definition applicationDefinition
		resp.Diagnostics.Append(plan.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		apiDefinition := expandApplicationDefinition(ctx, definition)
		in.Definition = apiDefinition
		update = true
	}

	if !plan.Description.Equal(state.Description) {
		in.Description = flex.StringFromFramework(ctx, plan.Description)
		update = true
	}

	if update {
		out, err := conn.UpdateApplication(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameApplication, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.ApplicationVersion == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameApplication, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		version, err := waitApplicationUpdated(ctx, conn, plan.ID.ValueString(), *out.ApplicationVersion, updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForUpdate, ResNameApplication, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		resp.Diagnostics.Append(plan.refreshFromVersion(ctx, version)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *resourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().M2Client(ctx)

	var state resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &m2.DeleteApplicationInput{
		ApplicationId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteApplication(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionDeleting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitApplicationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForDeletion, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *resourceApplication) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *resourceApplicationData) refreshFromOutput(ctx context.Context, app *m2.GetApplicationOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(flex.Flatten(ctx, app, r)...)
	r.ARN = flex.StringToFramework(ctx, app.ApplicationArn)
	r.ID = flex.StringToFramework(ctx, app.ApplicationId)
	r.CurrentVersion = flex.Int32ToFramework(ctx, app.LatestVersion.ApplicationVersion)

	return diags
}
func (r *resourceApplicationData) refreshFromVersion(ctx context.Context, version *m2.GetApplicationVersionOutput) diag.Diagnostics {
	var diags diag.Diagnostics
	definition, d := flattenApplicationDefinitionFromVersion(ctx, version)
	r.Definition = definition
	diags.Append(d...)
	r.CurrentVersion = flex.Int32ToFramework(ctx, version.ApplicationVersion)
	return diags
}

func waitApplicationCreated(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ApplicationLifecycleCreating),
		Target:                    enum.Slice(awstypes.ApplicationLifecycleCreated, awstypes.ApplicationLifecycleAvailable),
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationUpdated(ctx context.Context, conn *m2.Client, id string, version int32, timeout time.Duration) (*m2.GetApplicationVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ApplicationVersionLifecycleCreating),
		Target:                    enum.Slice(awstypes.ApplicationVersionLifecycleAvailable),
		Refresh:                   statusApplicationVersion(ctx, conn, id, version),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetApplicationVersionOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationDeleted(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationLifecycleDeleting, awstypes.ApplicationLifecycleDeletingFromEnvironment),
		Target:  []string{},
		Refresh: statusApplication(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationDeletedFromEnvironment(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationLifecycleDeletingFromEnvironment),
		Target:  enum.Slice(awstypes.ApplicationLifecycleAvailable),
		Refresh: statusApplication(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationStopped(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ApplicationLifecycleStopping),
		Target:                    enum.Slice(awstypes.ApplicationLifecycleStopped),
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationRunning(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ApplicationLifecycleStarting),
		Target:                    enum.Slice(awstypes.ApplicationLifecycleRunning),
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusApplication(ctx context.Context, conn *m2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findApplicationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func statusApplicationVersion(ctx context.Context, conn *m2.Client, id string, version int32) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findApplicationVersion(ctx, conn, id, version)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findApplicationByID(ctx context.Context, conn *m2.Client, id string) (*m2.GetApplicationOutput, error) {
	in := &m2.GetApplicationInput{
		ApplicationId: aws.String(id),
	}

	out, err := conn.GetApplication(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ApplicationId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func findApplicationVersion(ctx context.Context, conn *m2.Client, id string, version int32) (*m2.GetApplicationVersionOutput, error) {
	in := &m2.GetApplicationVersionInput{
		ApplicationId:      aws.String(id),
		ApplicationVersion: aws.Int32(version),
	}
	out, err := conn.GetApplicationVersion(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.ApplicationVersion == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func startApplication(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stopInput := &m2.StartApplicationInput{
		ApplicationId: &id,
	}

	_, err := conn.StartApplication(ctx, stopInput)
	if err != nil {
		return nil, err
	}

	app, err := waitApplicationRunning(ctx, conn, id, timeout)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func stopApplicationIfRunning(ctx context.Context, conn *m2.Client, id string, forceStop bool, timeout time.Duration) error {

	app, err := findApplicationByID(ctx, conn, id)
	if err != nil {
		if tfresource.NotFound(err) {
			return nil
		}
		return err
	}

	if app.Status != awstypes.ApplicationLifecycleRunning {
		return nil
	}

	stopInput := &m2.StopApplicationInput{
		ApplicationId: &id,
		ForceStop:     forceStop,
	}

	_, err = conn.StopApplication(ctx, stopInput)
	if err != nil {
		return err
	}

	_, err = waitApplicationStopped(ctx, conn, id, timeout)
	return err
}

func expandApplicationDefinition(ctx context.Context, definition applicationDefinition) awstypes.Definition {
	if !definition.S3Location.IsNull() {
		return &awstypes.DefinitionMemberS3Location{
			Value: *flex.StringFromFramework(ctx, definition.S3Location),
		}
	}

	if !definition.Content.IsNull() {
		return &awstypes.DefinitionMemberContent{
			Value: *flex.StringFromFramework(ctx, definition.Content),
		}
	}

	return nil
}

func flattenApplicationDefinitionFromVersion(ctx context.Context, version *m2.GetApplicationVersionOutput) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	obj := map[string]attr.Value{
		"content":     flex.StringToFramework(ctx, version.DefinitionContent),
		"s3_location": types.StringNull(), // This value is never returned...
	}

	definitionValue, d := types.ObjectValue(applicationDefinitionAttrs, obj)
	diags.Append(d...)

	return definitionValue, diags
}

type resourceApplicationData struct {
	ARN            types.String   `tfsdk:"arn"`
	ClientToken    types.String   `tfsdk:"client_token"`
	CurrentVersion types.Int64    `tfsdk:"current_version"`
	Definition     types.Object   `tfsdk:"definition"`
	Description    types.String   `tfsdk:"description"`
	ID             types.String   `tfsdk:"id"`
	EngineType     types.String   `tfsdk:"engine_type"`
	KmsKeyId       types.String   `tfsdk:"kms_key_id"`
	Name           types.String   `tfsdk:"name"`
	RoleArn        fwtypes.ARN    `tfsdk:"role_arn"`
	Tags           types.Map      `tfsdk:"tags"`
	TagsAll        types.Map      `tfsdk:"tags_all"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

type applicationDefinition struct {
	Content    types.String `tfsdk:"content"`
	S3Location types.String `tfsdk:"s3_location"`
}

var (
	applicationDefinitionAttrs = map[string]attr.Type{
		"content":     types.StringType,
		"s3_location": types.StringType,
	}
)
