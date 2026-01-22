// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package notifications

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_notifications_managed_notification_account_contact_association", name="Managed Notification Account Contact Association")
func newManagedNotificationAccountContactAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &managedNotificationAccountContactAssociationResource{}

	return r, nil
}

type managedNotificationAccountContactAssociationResource struct {
	framework.ResourceWithModel[managedNotificationAccountContactAssociationResourceModel]
	framework.WithNoUpdate
}

func (r *managedNotificationAccountContactAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"contact_identifier": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccountContactType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"managed_notification_configuration_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *managedNotificationAccountContactAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data managedNotificationAccountContactAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	managedNotificationConfigurationARN, contactIdentifier := fwflex.StringValueFromFramework(ctx, data.ManagedNotificationConfigurationARN), data.ContactIdentifier.ValueEnum()
	input := notifications.AssociateManagedNotificationAccountContactInput{
		ContactIdentifier:                   contactIdentifier,
		ManagedNotificationConfigurationArn: aws.String(managedNotificationConfigurationARN),
	}
	_, err := conn.AssociateManagedNotificationAccountContact(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating User Notifications Managed Notification Account Contact Association (%s,%s)", managedNotificationConfigurationARN, contactIdentifier), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *managedNotificationAccountContactAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data managedNotificationAccountContactAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	managedNotificationConfigurationARN, contactIdentifier := fwflex.StringValueFromFramework(ctx, data.ManagedNotificationConfigurationARN), fwflex.StringValueFromFramework(ctx, data.ContactIdentifier)
	_, err := findManagedNotificationAccountContactAssociationByTwoPartKey(ctx, conn, managedNotificationConfigurationARN, contactIdentifier)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading User Notifications Managed Notification Account Contact Association (%s,%s)", managedNotificationConfigurationARN, contactIdentifier), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *managedNotificationAccountContactAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data managedNotificationAccountContactAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NotificationsClient(ctx)

	managedNotificationConfigurationARN, contactIdentifier := fwflex.StringValueFromFramework(ctx, data.ManagedNotificationConfigurationARN), data.ContactIdentifier.ValueEnum()
	input := notifications.DisassociateManagedNotificationAccountContactInput{
		ContactIdentifier:                   contactIdentifier,
		ManagedNotificationConfigurationArn: aws.String(managedNotificationConfigurationARN),
	}
	_, err := conn.DisassociateManagedNotificationAccountContact(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting User Notifications Managed Notification Account Contact Association (%s,%s)", managedNotificationConfigurationARN, contactIdentifier), err.Error())

		return
	}
}

func (r *managedNotificationAccountContactAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		managedNotificationAccountContactAssociationIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, managedNotificationAccountContactAssociationIDParts, false)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("contact_identifier"), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("managed_notification_configuration_arn"), parts[0])...)
}

func findManagedNotificationAccountContactAssociationByTwoPartKey(ctx context.Context, conn *notifications.Client, managedNotificationConfigurationARN, contactIdentifier string) (*awstypes.ManagedNotificationChannelAssociationSummary, error) {
	input := notifications.ListManagedNotificationChannelAssociationsInput{
		ManagedNotificationConfigurationArn: aws.String(managedNotificationConfigurationARN),
	}

	return findManagedNotificationChannelAssociation(ctx, conn, &input, func(v *awstypes.ManagedNotificationChannelAssociationSummary) bool {
		return aws.ToString(v.ChannelIdentifier) == contactIdentifier && v.ChannelType == awstypes.ChannelTypeAccountContact
	})
}

func findManagedNotificationChannelAssociation(ctx context.Context, conn *notifications.Client, input *notifications.ListManagedNotificationChannelAssociationsInput, filter tfslices.Predicate[*awstypes.ManagedNotificationChannelAssociationSummary]) (*awstypes.ManagedNotificationChannelAssociationSummary, error) {
	output, err := findManagedNotificationChannelAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findManagedNotificationChannelAssociations(ctx context.Context, conn *notifications.Client, input *notifications.ListManagedNotificationChannelAssociationsInput, filter tfslices.Predicate[*awstypes.ManagedNotificationChannelAssociationSummary]) ([]awstypes.ManagedNotificationChannelAssociationSummary, error) {
	var output []awstypes.ManagedNotificationChannelAssociationSummary

	pages := notifications.NewListManagedNotificationChannelAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ChannelAssociations {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

type managedNotificationAccountContactAssociationResourceModel struct {
	ContactIdentifier                   fwtypes.StringEnum[awstypes.AccountContactType] `tfsdk:"contact_identifier"`
	ManagedNotificationConfigurationARN fwtypes.ARN                                     `tfsdk:"managed_notification_configuration_arn"`
}
