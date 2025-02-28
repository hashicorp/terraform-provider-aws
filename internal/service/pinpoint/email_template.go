// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @FrameworkResource("aws_pinpoint_email_template", name="Email Template")
// @Tags(identifierAttribute="arn")
func newResourceEmailTemplate(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEmailTemplate{}

	r.SetDefaultCreateTimeout(40 * time.Minute)
	r.SetDefaultUpdateTimeout(80 * time.Minute)
	r.SetDefaultDeleteTimeout(40 * time.Minute)

	return r, nil
}

const (
	ResNameEmailTemplate = "Email Template"
)

type resourceEmailTemplate struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceEmailTemplate) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"template_name": schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"email_template": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[emailTemplate](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"default_substitutions": schema.StringAttribute{
							Optional: true,
						},
						"html_part": schema.StringAttribute{
							Optional: true,
						},
						"recommender_id": schema.StringAttribute{
							Optional: true,
						},
						"subject": schema.StringAttribute{
							Optional: true,
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						"text_part": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrHeader: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[headerData](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrName: schema.StringAttribute{
										Optional: true,
									},
									names.AttrValue: schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceEmailTemplate) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().PinpointClient(ctx)

	var plan emailTemplateData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &pinpoint.CreateEmailTemplateInput{}

	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in, flex.WithFieldNameSuffix("Request"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.EmailTemplateRequest.Tags = getTagsIn(ctx)

	out, err := conn.CreateEmailTemplate(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Pinpoint, create.ErrActionCreating, ResNameEmailTemplate, plan.TemplateName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Pinpoint, create.ErrActionCreating, ResNameEmailTemplate, plan.TemplateName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.Arn = flex.StringToFramework(ctx, out.CreateTemplateMessageBody.Arn)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEmailTemplate) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().PinpointClient(ctx)

	var state emailTemplateData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEmailTemplateByName(ctx, conn, state.TemplateName.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Pinpoint, create.ErrActionSetting, ResNameEmailTemplate, state.TemplateName.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithFieldNameSuffix("Response"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Arn = flex.StringToFramework(ctx, out.EmailTemplateResponse.Arn)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEmailTemplate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().PinpointClient(ctx)

	var old, new emailTemplateData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &old)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !old.TemplateName.Equal(new.TemplateName) ||
		!old.Arn.Equal(new.Arn) ||
		!old.EmailTemplate.Equal(new.EmailTemplate) {
		in := &pinpoint.UpdateEmailTemplateInput{}

		resp.Diagnostics.Append(flex.Expand(ctx, &old, in, flex.WithFieldNameSuffix("Request"))...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.TemplateName = old.TemplateName.ValueStringPointer()

		_, err := conn.UpdateEmailTemplate(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Pinpoint, create.ErrActionUpdating, ResNameEmailTemplate, new.TemplateName.String(), err),
				err.Error(),
			)
			return
		}
	}

	output, err := findEmailTemplateByName(ctx, conn, old.TemplateName.ValueString())
	if err != nil {
		create.AddError(&resp.Diagnostics, names.Pinpoint, create.ErrActionWaitingForUpdate, ResNameEmailTemplate, old.TemplateName.ValueString(), err)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &new, flex.WithFieldNameSuffix("Response"))...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceEmailTemplate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PinpointClient(ctx)
	var state emailTemplateData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &pinpoint.DeleteEmailTemplateInput{
		TemplateName: state.TemplateName.ValueStringPointer(),
	}

	_, err := conn.DeleteEmailTemplate(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Pinpoint, create.ErrActionDeleting, ResNameEmailTemplate, state.TemplateName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEmailTemplate) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("template_name"), request, response)
}

func findEmailTemplateByName(ctx context.Context, conn *pinpoint.Client, name string) (*pinpoint.GetEmailTemplateOutput, error) {
	in := &pinpoint.GetEmailTemplateInput{
		TemplateName: aws.String(name),
	}

	out, err := conn.GetEmailTemplate(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type emailTemplateData struct {
	TemplateName  types.String                                   `tfsdk:"template_name"`
	EmailTemplate fwtypes.ListNestedObjectValueOf[emailTemplate] `tfsdk:"email_template"`
	Arn           types.String                                   `tfsdk:"arn"`
	Tags          tftags.Map                                     `tfsdk:"tags"`
	TagsAll       tftags.Map                                     `tfsdk:"tags_all"`
}

type emailTemplate struct {
	DefaultSubstitutions types.String                                `tfsdk:"default_substitutions"`
	Header               fwtypes.ListNestedObjectValueOf[headerData] `tfsdk:"header"`
	HtmlPart             types.String                                `tfsdk:"html_part"`
	RecommenderId        types.String                                `tfsdk:"recommender_id"`
	Subject              types.String                                `tfsdk:"subject"`
	Description          types.String                                `tfsdk:"description"`
	TextPart             types.String                                `tfsdk:"text_part"`
}

type headerData struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
