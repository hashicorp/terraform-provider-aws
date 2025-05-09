// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notificationscontacts

import (
	"context"
	"errors"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notificationscontacts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notificationscontacts/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_notificationscontacts_email_contact", name="Email Contact")
// @Tags(identifierAttribute="arn")
func newResourceEmailContact(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEmailContact{}

	return r, nil
}

const (
	ResNameEmailContact = "Email Contact"
)

type resourceEmailContact struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[resourceEmailContactModel]
}

func (r *resourceEmailContact) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"email_address": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 254),
					stringvalidator.RegexMatches(regexache.MustCompile(`(.+)@(.+)`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`[\w.~-]+`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EmailContactStatus](),
				Computed:   true,
				Default:    stringdefault.StaticString(string(awstypes.EmailContactStatusInactive)),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *resourceEmailContact) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NotificationsContactsClient(ctx)

	var plan resourceEmailContactModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input notificationscontacts.CreateEmailContactInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateEmailContact(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NotificationsContacts, create.ErrActionCreating, ResNameEmailContact, plan.EmailAddress.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Arn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NotificationsContacts, create.ErrActionCreating, ResNameEmailContact, plan.EmailAddress.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	var activateIn notificationscontacts.SendActivationCodeInput
	activateIn.Arn = out.Arn
	_, err = conn.SendActivationCode(ctx, &activateIn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NotificationsContacts, "activating", ResNameEmailContact, plan.EmailAddress.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEmailContact) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NotificationsContactsClient(ctx)

	var state resourceEmailContactModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEmailContactByARN(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NotificationsContacts, create.ErrActionReading, ResNameEmailContact, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	state.EmailAddress = flex.StringToFramework(ctx, out.Address)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEmailContact) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NotificationsContactsClient(ctx)

	var state resourceEmailContactModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := notificationscontacts.DeleteEmailContactInput{
		Arn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteEmailContact(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NotificationsContacts, create.ErrActionDeleting, ResNameEmailContact, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEmailContact) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

func findEmailContactByARN(ctx context.Context, conn *notificationscontacts.Client, arn string) (*awstypes.EmailContact, error) {
	input := notificationscontacts.GetEmailContactInput{
		Arn: aws.String(arn),
	}

	out, err := conn.GetEmailContact(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.EmailContact == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.EmailContact, nil
}

type resourceEmailContactModel struct {
	ARN          types.String                                    `tfsdk:"arn"`
	EmailAddress types.String                                    `tfsdk:"email_address"`
	Name         types.String                                    `tfsdk:"name"`
	Status       fwtypes.StringEnum[awstypes.EmailContactStatus] `tfsdk:"status"`
	Tags         tftags.Map                                      `tfsdk:"tags"`
	TagsAll      tftags.Map                                      `tfsdk:"tags_all"`
}

func sweepEmailContacts(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := notificationscontacts.ListEmailContactsInput{}
	conn := client.NotificationsContactsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := notificationscontacts.NewListEmailContactsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.EmailContacts {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceEmailContact, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}
