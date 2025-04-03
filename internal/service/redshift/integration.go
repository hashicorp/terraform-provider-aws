// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_redshift_integration", name="Integration")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newResourceIntegration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceIntegration{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	integrationStatusActive         = "active"
	integrationStatusCreating       = "creating"
	integrationStatusDeleting       = "deleting"
	integrationStatusFailed         = "failed"
	integrationStatusModifying      = "modifying"
	integrationStatusNeedsAttention = "needs_attention"
	integrationStatusSyncing        = "syncing"
)

const (
	ResNameIntegration = "Integration"
)

type resourceIntegration struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceIntegration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"additional_encryption_context": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"integration_name": schema.StringAttribute{
				Required: true,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					&ignoreKmsKeyIdForS3Modifier{},
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrTargetARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

func (r *resourceIntegration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var plan resourceIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input redshift.CreateIntegrationInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.TagList = getTagsIn(ctx)

	out, err := conn.CreateIntegration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionCreating, ResNameIntegration, plan.IntegrationName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.IntegrationArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionCreating, ResNameIntegration, plan.IntegrationName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// Set values for unknowns.
	plan.IntegrationARN = flex.StringToFramework(ctx, out.IntegrationArn)
	plan.setID()

	prevAdditionalEncryptionContext := plan.AdditionalEncryptionContext

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Null vs. empty map handling.
	if prevAdditionalEncryptionContext.IsNull() && !plan.AdditionalEncryptionContext.IsNull() && len(plan.AdditionalEncryptionContext.Elements()) == 0 {
		plan.AdditionalEncryptionContext = prevAdditionalEncryptionContext
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	integration, err := waitIntegrationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.State.SetAttribute(ctx, path.Root(names.AttrID), plan.ID) // Set 'id' so as to taint the resource.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionWaitingForCreation, ResNameIntegration, plan.IntegrationName.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	plan.KMSKeyID = flex.StringToFramework(ctx, integration.KMSKeyId)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceIntegration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var state resourceIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := state.InitFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	out, err := findIntegrationByARN(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionReading, ResNameIntegration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	prevAdditionalEncryptionContext := state.AdditionalEncryptionContext

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Null vs. empty map handling.
	if prevAdditionalEncryptionContext.IsNull() && !state.AdditionalEncryptionContext.IsNull() && len(state.AdditionalEncryptionContext.Elements()) == 0 {
		state.AdditionalEncryptionContext = prevAdditionalEncryptionContext
	}

	setTagsOut(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceIntegration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var plan, state resourceIntegrationModel
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
		var input redshift.ModifyIntegrationInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.ModifyIntegration(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Redshift, create.ErrActionUpdating, ResNameIntegration, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.IntegrationArn == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Redshift, create.ErrActionUpdating, ResNameIntegration, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		prevAdditionalEncryptionContext := plan.AdditionalEncryptionContext

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Null vs. empty map handling.
		if prevAdditionalEncryptionContext.IsNull() && !plan.AdditionalEncryptionContext.IsNull() && len(plan.AdditionalEncryptionContext.Elements()) == 0 {
			plan.AdditionalEncryptionContext = prevAdditionalEncryptionContext
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitIntegrationUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionWaitingForUpdate, ResNameIntegration, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceIntegration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var state resourceIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := redshift.DeleteIntegrationInput{
		IntegrationArn: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteIntegration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.IntegrationNotFoundFault](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionDeleting, ResNameIntegration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitIntegrationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Redshift, create.ErrActionWaitingForDeletion, ResNameIntegration, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func integrationError(v awstypes.IntegrationError) error {
	return fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage))
}

type resourceIntegrationModel struct {
	AdditionalEncryptionContext fwtypes.MapValueOf[types.String] `tfsdk:"additional_encryption_context"`
	Description                 types.String                     `tfsdk:"description"`
	ID                          types.String                     `tfsdk:"id"`
	IntegrationARN              types.String                     `tfsdk:"arn"`
	IntegrationName             types.String                     `tfsdk:"integration_name"`
	KMSKeyID                    types.String                     `tfsdk:"kms_key_id"`
	SourceARN                   fwtypes.ARN                      `tfsdk:"source_arn"`
	Tags                        tftags.Map                       `tfsdk:"tags"`
	TagsAll                     tftags.Map                       `tfsdk:"tags_all"`
	TargetARN                   fwtypes.ARN                      `tfsdk:"target_arn"`
	Timeouts                    timeouts.Value                   `tfsdk:"timeouts"`
}

func (model *resourceIntegrationModel) InitFromID() error {
	model.IntegrationARN = model.ID

	return nil
}

func (model *resourceIntegrationModel) setID() {
	model.ID = model.IntegrationARN
}

type ignoreKmsKeyIdForS3Modifier struct{}

func (m *ignoreKmsKeyIdForS3Modifier) Description(_ context.Context) string {
	return "If source_arn is an S3 ARN and ConfigValue is null, do not show any differences."
}

func (m *ignoreKmsKeyIdForS3Modifier) MarkdownDescription(_ context.Context) string {
	return "If source_arn is an S3 ARN and ConfigValue is null, do not show any differences."
}

func (m *ignoreKmsKeyIdForS3Modifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var plan resourceIntegrationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// when source_arn is an S3 ARN and ConfigValue is null, do not show any differences.
	matched, _ := regexp.MatchString(`^arn:aws[a-z\-]*:s3:.*$`, plan.SourceARN.ValueString())
	if matched && req.ConfigValue.IsNull() {
		resp.PlanValue = req.StateValue
	}
}
