// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_guardrail_version", name="Guardrail Version")
func newResourceGuardrailVersion(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGuardrailVersion{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameGuardrailVersion = "Guardrail Version"
)

type resourceGuardrailVersion struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceGuardrailVersion) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrock_guardrail_version"
}

func (r *resourceGuardrailVersion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 200),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"guardrail_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(2048),
					stringvalidator.RegexMatches(regexp.MustCompile("^((arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:[0-9]{12}:guardrail/[a-z0-9]+))$"), ""),
				},
				PlanModifiers: []planmodifier.String{ /*START PLAN MODIFIERS*/
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSkipDestroy: schema.BoolAttribute{
				Optional: true,
			},
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceGuardrailVersion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	conn := r.Meta().BedrockClient(ctx)

	var plan resourceGuardrailVersionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &bedrock.CreateGuardrailVersionInput{
		GuardrailIdentifier: aws.String(plan.GuardrailIdentifier.ValueString()),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	out, err := conn.CreateGuardrailVersion(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameGuardrailVersion, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Version == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameGuardrailVersion, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.Version = flex.StringToFramework(ctx, out.Version)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitGuardrailCreated(ctx, conn, plan.GuardrailIdentifier.ValueString(), plan.Version.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForCreation, ResNameGuardrailVersion, plan.Version.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceGuardrailVersion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state resourceGuardrailVersionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGuardrailByID(ctx, conn, state.GuardrailIdentifier.ValueString(), state.Version.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionSetting, ResNameGuardrailVersion, "", err),
			err.Error(),
		)
		return
	}

	state.Version = flex.StringToFramework(ctx, out.Version)
	state.Description = flex.StringToFramework(ctx, out.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op
func (r *resourceGuardrailVersion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *resourceGuardrailVersion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state resourceGuardrailVersionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SkipDestroy.ValueBool() {
		return
	}

	in := &bedrock.DeleteGuardrailInput{
		GuardrailIdentifier: aws.String(state.GuardrailIdentifier.ValueString()),
		GuardrailVersion:    aws.String(state.Version.ValueString()),
	}

	if _, err := conn.DeleteGuardrail(ctx, in); err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionDeleting, ResNameGuardrail, state.GuardrailIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if _, err := waitGuardrailDeleted(ctx, conn, state.GuardrailIdentifier.ValueString(), state.Version.ValueString(), deleteTimeout); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionWaitingForDeletion, ResNameGuardrail, state.GuardrailIdentifier.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceGuardrailVersion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, guardrailIDParts, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: guardrail_identifier,version. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("guardrail_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrVersion), parts[1])...)
}

type resourceGuardrailVersionData struct {
	Description         types.String   `tfsdk:"description"`
	GuardrailIdentifier types.String   `tfsdk:"guardrail_identifier"`
	Version             types.String   `tfsdk:"version"`
	SkipDestroy         types.Bool     `tfsdk:"skip_destroy"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}
