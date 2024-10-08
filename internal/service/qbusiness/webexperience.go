// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	awstypes "github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_qbusiness_webexperience", name="Webexperience")
// @Tags(identifierAttribute="arn")
func newResourceWebexperience(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceWebexperience{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameWebexperience = "Webexperience"
)

type resourceWebexperience struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceWebexperience) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrApplicationID: schema.StringAttribute{
				Description: "Identifier of the Amazon Q application associated with the webexperience",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				},
			},
			"iam_service_role_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "The Amazon Resource Name (ARN) of the service role attached to your web experience",
				Optional:    true,
			},
			"sample_prompts_control_mode": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.WebExperienceSamplePromptsControlMode](),
				Required:    true,
				Description: "Status information about whether file upload functionality is activated or deactivated for your end user.",
			},
			"subtitle": schema.StringAttribute{
				Description: "The subtitle for your Amazon Q Business web experience.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			"title": schema.StringAttribute{
				Description: "The title for your Amazon Q Business web experience.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			"webexperience_id": schema.StringAttribute{
				Computed: true,
			},
			"welcome_message": schema.StringAttribute{
				Description: "The customized welcome message for end users of an Amazon Q Business web experience.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

type resourceWebexperienceData struct {
	ApplicationId            types.String                                                       `tfsdk:"application_id"`
	ID                       types.String                                                       `tfsdk:"id"`
	RoleArn                  fwtypes.ARN                                                        `tfsdk:"iam_service_role_arn"`
	SamplePromptsControlMode fwtypes.StringEnum[awstypes.WebExperienceSamplePromptsControlMode] `tfsdk:"sample_prompts_control_mode"`
	Subtitle                 types.String                                                       `tfsdk:"subtitle"`
	Tags                     types.Map                                                          `tfsdk:"tags"`
	TagsAll                  types.Map                                                          `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                                     `tfsdk:"timeouts"`
	Title                    types.String                                                       `tfsdk:"title"`
	WebexperienceArn         types.String                                                       `tfsdk:"arn"`
	WebexperienceId          types.String                                                       `tfsdk:"webexperience_id"`
	WelcomeMessage           types.String                                                       `tfsdk:"welcome_message"`
}

func (r *resourceWebexperience) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceWebexperienceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	input := &qbusiness.CreateWebExperienceInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(id.UniqueId())

	out, err := conn.CreateWebExperience(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Q Business webexperience", err.Error())
		return
	}

	data.WebexperienceArn = fwflex.StringToFramework(ctx, out.WebExperienceArn)
	data.WebexperienceId = fwflex.StringToFramework(ctx, out.WebExperienceId)

	data.setID()

	if _, err := waitWebexperienceCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Amazon Q Webexperience creation", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceWebexperience) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceWebexperienceData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	out, err := FindWebexperienceByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to retrieve Q Business webexperience (%s)", data.ID.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceWebexperience) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceWebexperienceData

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := new.initFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	if !old.SamplePromptsControlMode.Equal(new.SamplePromptsControlMode) ||
		!old.RoleArn.Equal(new.RoleArn) ||
		!old.Subtitle.Equal(new.Subtitle) ||
		!old.Title.Equal(new.Title) ||
		!old.WelcomeMessage.Equal(new.WelcomeMessage) {
		conn := r.Meta().QBusinessClient(ctx)

		input := &qbusiness.UpdateWebExperienceInput{}

		resp.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateWebExperience(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError("failed to update Amazon Q webexperience", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceWebexperience) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceWebexperienceData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	if err := data.initFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	webexperienceId := data.WebexperienceId.ValueString()
	appId := data.ApplicationId.ValueString()

	input := &qbusiness.DeleteWebExperienceInput{
		WebExperienceId: aws.String(webexperienceId),
		ApplicationId:   aws.String(appId),
	}

	_, err := conn.DeleteWebExperience(ctx, input)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("failed to delete Q Business webexperience (%s)", data.WebexperienceId.ValueString()), err.Error())
		return
	}

	if _, err := waitWebexperienceDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Q Business webexperience deletion", err.Error())
		return
	}
}

func (r *resourceWebexperience) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *resourceWebexperience) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_qbusiness_webexperience"
}

const (
	indexWebexperienceIDPartCount = 2
)

func (r *resourceWebexperienceData) setID() {
	r.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{r.ApplicationId.ValueString(), r.WebexperienceId.ValueString()},
		indexWebexperienceIDPartCount, false)))
}

func (r *resourceWebexperienceData) initFromID() error {
	parts, err := flex.ExpandResourceId(r.ID.ValueString(), indexWebexperienceIDPartCount, false)
	if err != nil {
		return err
	}

	r.ApplicationId = types.StringValue(parts[0])
	r.WebexperienceId = types.StringValue(parts[1])
	return nil
}

func FindWebexperienceByID(ctx context.Context, conn *qbusiness.Client, webexperience_id string) (*qbusiness.GetWebExperienceOutput, error) {
	parts, err := flex.ExpandResourceId(webexperience_id, indexWebexperienceIDPartCount, false)

	if err != nil {
		return nil, err
	}

	input := &qbusiness.GetWebExperienceInput{
		ApplicationId:   aws.String(parts[0]),
		WebExperienceId: aws.String(parts[1]),
	}

	output, err := conn.GetWebExperience(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
