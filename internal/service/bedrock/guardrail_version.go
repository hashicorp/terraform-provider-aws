// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_guardrail_version", name="Guardrail Version")
func newGuardrailVersionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &guardrailVersionResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type guardrailVersionResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[guardrailVersionResourceModel]
	framework.WithTimeouts
}

func (r *guardrailVersionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 200),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"guardrail_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
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
				Delete: true,
			}),
		},
	}
}

func (r *guardrailVersionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data guardrailVersionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	guardrailARN := data.GuardrailARN.ValueString()
	input := &bedrock.CreateGuardrailVersionInput{
		Description:         fwflex.StringFromFramework(ctx, data.Description),
		GuardrailIdentifier: aws.String(guardrailARN),
	}

	output, err := conn.CreateGuardrailVersion(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Bedrock Guardrail Version (%s)", guardrailARN), err.Error())

		return
	}

	data.Version = fwflex.StringToFramework(ctx, output.Version)

	if _, err := waitGuardrailCreated(ctx, conn, data.GuardrailARN.ValueString(), data.Version.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Guardrail (%s) Version (%s) create", data.GuardrailARN.ValueString(), data.Version.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *guardrailVersionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data guardrailVersionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	output, err := findGuardrailByTwoPartKey(ctx, conn, data.GuardrailARN.ValueString(), data.Version.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Guardrail (%s) Version (%s)", data.GuardrailARN.ValueString(), data.Version.ValueString()), err.Error())

		return
	}

	data.Description = fwflex.StringToFramework(ctx, output.Description)
	data.GuardrailARN = fwtypes.ARNValue(aws.ToString(output.GuardrailArn))
	data.Version = fwflex.StringToFramework(ctx, output.Version)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *guardrailVersionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data guardrailVersionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	if data.SkipDestroy.ValueBool() {
		return
	}

	input := bedrock.DeleteGuardrailInput{
		GuardrailIdentifier: data.GuardrailARN.ValueStringPointer(),
		GuardrailVersion:    data.Version.ValueStringPointer(),
	}
	_, err := conn.DeleteGuardrail(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Guardrail (%s) Version (%s)", data.GuardrailARN.ValueString(), data.Version.ValueString()), err.Error())

		return
	}

	if _, err := waitGuardrailDeleted(ctx, conn, data.GuardrailARN.ValueString(), data.Version.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Guardrail (%s) Version (%s) delete", data.GuardrailARN.ValueString(), data.Version.ValueString()), err.Error())

		return
	}
}

func (r *guardrailVersionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts, err := flex.ExpandResourceId(request.ID, guardrailIDParts, false)
	if err != nil {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: guardrail_identifier,version. Got: %q", request.ID),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("guardrail_arn"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrVersion), parts[1])...)
}

type guardrailVersionResourceModel struct {
	Description  types.String   `tfsdk:"description"`
	GuardrailARN fwtypes.ARN    `tfsdk:"guardrail_arn"`
	SkipDestroy  types.Bool     `tfsdk:"skip_destroy"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
	Version      types.String   `tfsdk:"version"`
}
