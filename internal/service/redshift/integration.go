// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_redshift_integration", name="Integration")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/redshift/types;awstypes;awstypes.Integration")
// @Testing(preIdentityVersion="6.19.0")
func newIntegrationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &integrationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type integrationResource struct {
	framework.ResourceWithModel[integrationResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *integrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					stringplanmodifier.RequiresReplaceIfConfigured(),
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

func (r *integrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data integrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	var input redshift.CreateIntegrationInput
	response.Diagnostics.Append(flex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.TagList = getTagsIn(ctx)

	output, err := conn.CreateIntegration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Redshift Integration (%s)", data.IntegrationName.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.IntegrationARN = flex.StringToFramework(ctx, output.IntegrationArn)

	integration, err := waitIntegrationCreated(ctx, conn, data.IntegrationARN.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Redshift Integration (%s) create", data.IntegrationARN.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.KMSKeyID = flex.StringToFramework(ctx, integration.KMSKeyId)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *integrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data integrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	output, err := findIntegrationByARN(ctx, conn, data.IntegrationARN.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Redshift Integration (%s)", data.IntegrationARN.ValueString()), err.Error())

		return
	}

	prevAdditionalEncryptionContext := data.AdditionalEncryptionContext

	response.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Null vs. empty map handling.
	if prevAdditionalEncryptionContext.IsNull() && !data.AdditionalEncryptionContext.IsNull() && len(data.AdditionalEncryptionContext.Elements()) == 0 {
		data.AdditionalEncryptionContext = prevAdditionalEncryptionContext
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *integrationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old integrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.IntegrationName.Equal(old.IntegrationName) {
		var input redshift.ModifyIntegrationInput
		response.Diagnostics.Append(flex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.ModifyIntegration(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Redshift Integration (%s)", new.IntegrationARN.ValueString()), err.Error())

			return
		}

		if _, err := waitIntegrationUpdated(ctx, conn, new.IntegrationARN.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Redshift Integration (%s) update", new.IntegrationARN.ValueString()), err.Error())

			return
		}
	}

	new.KMSKeyID = old.KMSKeyID

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *integrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data integrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	input := redshift.DeleteIntegrationInput{
		IntegrationArn: flex.StringFromFramework(ctx, data.IntegrationARN),
	}
	_, err := conn.DeleteIntegration(ctx, &input)

	if errs.IsA[*awstypes.IntegrationNotFoundFault](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Redshift Integration (%s)", data.IntegrationARN.ValueString()), err.Error())

		return
	}

	if _, err := waitIntegrationDeleted(ctx, conn, data.IntegrationARN.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Redshift Integration (%s) delete", data.IntegrationARN.ValueString()), err.Error())

		return
	}
}

func integrationError(v awstypes.IntegrationError) error {
	return fmt.Errorf("%s: %s", aws.ToString(v.ErrorCode), aws.ToString(v.ErrorMessage))
}

type integrationResourceModel struct {
	framework.WithRegionModel
	AdditionalEncryptionContext fwtypes.MapOfString `tfsdk:"additional_encryption_context"`
	Description                 types.String        `tfsdk:"description"`
	IntegrationARN              types.String        `tfsdk:"arn"`
	IntegrationName             types.String        `tfsdk:"integration_name"`
	KMSKeyID                    types.String        `tfsdk:"kms_key_id"`
	SourceARN                   fwtypes.ARN         `tfsdk:"source_arn"`
	Tags                        tftags.Map          `tfsdk:"tags"`
	TagsAll                     tftags.Map          `tfsdk:"tags_all"`
	TargetARN                   fwtypes.ARN         `tfsdk:"target_arn"`
	Timeouts                    timeouts.Value      `tfsdk:"timeouts"`
}
