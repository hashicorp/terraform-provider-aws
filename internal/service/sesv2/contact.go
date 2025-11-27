// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	intretry "github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_sesv2_contact", name="Contact")
func newResourceContact(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceContact{}

	return r, nil
}

const (
	ResNameContact = "Contact"
)

type resourceContact struct {
	framework.ResourceWithModel[resourceContactModel]
}

func (r *resourceContact) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"contact_list_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email_address": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"unsubscribe_all": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"topic_preferences": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[topicPreference](ctx),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"subscription_status": schema.StringAttribute{
							Required: true,
						},
						"topic_name": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceContact) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var plan resourceContactModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input sesv2.CreateContactInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Contact")))
	if resp.Diagnostics.HasError() {
		return
	}

	var preferences []topicPreference
	smerr.AddEnrich(ctx, &resp.Diagnostics, plan.TopicPreferences.ElementsAs(ctx, &preferences, false))
	if resp.Diagnostics.HasError() {
		return
	}

	if len(preferences) > 0 {
		input.TopicPreferences = expandTopicPreferences(ctx, preferences)
	}

	out, err := conn.CreateContact(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.name())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.name())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithIgnoredFieldNames([]string{"TopicPreferences"})))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceContact) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var state resourceContactModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindContact(ctx, conn, state.ContactListName.ValueString(), state.EmailAddress.ValueString())
	if intretry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.name())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithIgnoredFieldNames([]string{"TopicPreferences"})))
	if resp.Diagnostics.HasError() {
		return
	}
	if len(out.TopicPreferences) > 0 {
		state.TopicPreferences = flattenTopicPreferences(ctx, out.TopicPreferences)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceContact) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var plan, state resourceContactModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input sesv2.UpdateContactInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Contact")))
		if resp.Diagnostics.HasError() {
			return
		}

		var preferences []topicPreference
		smerr.AddEnrich(ctx, &resp.Diagnostics, plan.TopicPreferences.ElementsAs(ctx, &preferences, false))
		if resp.Diagnostics.HasError() {
			return
		}
		if len(preferences) > 0 {
			input.TopicPreferences = expandTopicPreferences(ctx, preferences)
		}

		out, err := conn.UpdateContact(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.name())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.name())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan), flex.WithIgnoredFieldNames([]string{"TopicPreferences"}))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceContact) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SESV2Client(ctx)

	var state resourceContactModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	var input sesv2.DeleteContactInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, state, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteContact(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.name())
		return
	}
}

func FindContact(ctx context.Context, conn *sesv2.Client, contactList, email string) (*sesv2.GetContactOutput, error) {
	input := &sesv2.GetContactInput{
		ContactListName: aws.String(contactList),
		EmailAddress:    aws.String(email),
	}

	output, err := conn.GetContact(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type resourceContactModel struct {
	ContactListName  types.String                                    `tfsdk:"contact_list_name"`
	EmailAddress     types.String                                    `tfsdk:"email_address"`
	TopicPreferences fwtypes.SetNestedObjectValueOf[topicPreference] `tfsdk:"topic_preferences"`
	UnsubscribeAll   types.Bool                                      `tfsdk:"unsubscribe_all"`
}

func (r *resourceContactModel) name() string {
	return fmt.Sprintf("%s %s", r.ContactListName.ValueString(), r.EmailAddress.ValueString())
}

type topicPreference struct {
	SubscriptionStatus types.String `tfsdk:"subscription_status"`
	TopicName          types.String `tfsdk:"topic_name"`
}

func expandTopicPreferences(_ context.Context, tfList []topicPreference) []awstypes.TopicPreference {
	var preferences []awstypes.TopicPreference
	for _, item := range tfList {
		preference := awstypes.TopicPreference{
			SubscriptionStatus: awstypes.SubscriptionStatus(item.SubscriptionStatus.ValueString()),
			TopicName:          item.TopicName.ValueStringPointer(),
		}
		preferences = append(preferences, preference)
	}
	return preferences
}

func flattenTopicPreferences(ctx context.Context, preferences []awstypes.TopicPreference) fwtypes.SetNestedObjectValueOf[topicPreference] {
	var result []*topicPreference
	for _, preference := range preferences {
		result = append(result, &topicPreference{
			SubscriptionStatus: flex.StringValueToFramework(ctx, preference.SubscriptionStatus),
			TopicName:          flex.StringToFramework(ctx, preference.TopicName),
		})
	}
	return fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, result)
}

func sweepContacts(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SESV2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	contactListPages := sesv2.NewListContactListsPaginator(conn, &sesv2.ListContactListsInput{})
	for contactListPages.HasMorePages() {
		contactListPage, err := contactListPages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, contactList := range contactListPage.ContactLists {
			input := &sesv2.ListContactsInput{
				ContactListName: contactList.ContactListName,
			}

			pages := sesv2.NewListContactsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.Contacts {
					sweepResources = append(sweepResources, sweepfw.NewSweepResource(
						newResourceContact,
						client,
						sweepfw.NewAttribute("contact_list_name", aws.ToString(contactList.ContactListName)),
						sweepfw.NewAttribute("email_address", aws.ToString(v.EmailAddress)),
					),
					)
				}
			}
		}
	}

	return sweepResources, nil
}
