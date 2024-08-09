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

func (*resourceEmailTemplate) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_pinpoint_email_template"
}

func (r *resourceEmailTemplate) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"template_name": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"request": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[requestData](ctx),
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
						"template_description": schema.StringAttribute{
							Optional: true,
						},
						"text_part": schema.StringAttribute{
							Optional: true,
						},
						names.AttrTags:    tftags.TagsAttribute(),
						names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
					},
					Blocks: map[string]schema.Block{
						"header": schema.ListNestedBlock{
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

	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.EmailTemplateRequest.Tags = getTagsIn(ctx)

	out, err := conn.CreateEmailTemplate(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameEmailTemplate, plan.TemplateName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameEmailTemplate, plan.TemplateName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
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
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameEmailTemplate, state.TemplateName.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEmailTemplate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().PinpointClient(ctx)

	var plan, state emailTemplateData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.TemplateName.Equal(state.TemplateName) {
		in := &pinpoint.UpdateEmailTemplateInput{
			TemplateName: aws.String(plan.TemplateName.ValueString()),
		}
		resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.TemplateName = plan.TemplateName.ValueStringPointer()
		out, err := conn.UpdateEmailTemplate(ctx, in)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameEmailTemplate, plan.TemplateName.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameEmailTemplate, plan.TemplateName.String(), nil),
				errors.New("empty output from email template update").Error(),
			)
			return
		}
		resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEmailTemplate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PinpointClient(ctx)
	var state emailTemplateData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &pinpoint.DeleteEmailTemplateInput{
		TemplateName: aws.String(state.TemplateName.ValueString()),
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
	TemplateName types.String                                 `tfsdk:"template_name"`
	Request      fwtypes.ListNestedObjectValueOf[requestData] `tfsdk:"request"`
}

type requestData struct {
	DefaultSubstitutions types.String                                `tfsdk:"default_substitutions"`
	Header               fwtypes.ListNestedObjectValueOf[headerData] `tfsdk:"header"`
	HtmlPart             types.String                                `tfsdk:"html_part"`
	RecommenderId        types.String                                `tfsdk:"recommenderId"`
	Subject              types.String                                `tfsdk:"subject"`
	TemplateDescription  types.String                                `tfsdk:"template_description"`
	TextPart             types.String                                `tfsdk:"text_part"`
	Tags                 types.Map                                   `tfsdk:"tags"`
	TagsAll              types.Map                                   `tfsdk:"tags_all"`
}

type headerData struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
