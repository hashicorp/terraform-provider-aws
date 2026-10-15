// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_connect_phone_number_contact_flow_association", name="Phone Number Contact Flow Association")
func newPhoneNumberContactFlowAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &phoneNumberContactFlowAssociationResource{}

	return r, nil
}

type phoneNumberContactFlowAssociationResource struct {
	framework.ResourceWithModel[phoneNumberContactFlowAssociationResourceModel]
	framework.WithNoUpdate
}

func (r *phoneNumberContactFlowAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"contact_flow_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrInstanceID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"phone_number_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *phoneNumberContactFlowAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data phoneNumberContactFlowAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectClient(ctx)

	phoneNumberID, instanceID, contactFlowID := fwflex.StringValueFromFramework(ctx, data.PhoneNumberID), fwflex.StringValueFromFramework(ctx, data.InstanceID), fwflex.StringValueFromFramework(ctx, data.ContactFlowID)
	input := connect.AssociatePhoneNumberContactFlowInput{
		ContactFlowId: aws.String(contactFlowID),
		InstanceId:    aws.String(instanceID),
		PhoneNumberId: aws.String(phoneNumberID),
	}

	_, err := conn.AssociatePhoneNumberContactFlow(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Connect Phone Number (%s) Contact Flow (%s) Association", phoneNumberID, contactFlowID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *phoneNumberContactFlowAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data phoneNumberContactFlowAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectClient(ctx)

	phoneNumberID, instanceID, contactFlowID := fwflex.StringValueFromFramework(ctx, data.PhoneNumberID), fwflex.StringValueFromFramework(ctx, data.InstanceID), fwflex.StringValueFromFramework(ctx, data.ContactFlowID)
	_, err := findPhoneNumberContactFlowAssociationByThreePartKey(ctx, conn, phoneNumberID, instanceID, contactFlowID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Connect Phone Number (%s) Contact Flow (%s) Association", phoneNumberID, contactFlowID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *phoneNumberContactFlowAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data phoneNumberContactFlowAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectClient(ctx)

	phoneNumberID, instanceID, contactFlowID := fwflex.StringValueFromFramework(ctx, data.PhoneNumberID), fwflex.StringValueFromFramework(ctx, data.InstanceID), fwflex.StringValueFromFramework(ctx, data.ContactFlowID)
	input := connect.DisassociatePhoneNumberContactFlowInput{
		InstanceId:    aws.String(instanceID),
		PhoneNumberId: aws.String(phoneNumberID),
	}
	_, err := conn.DisassociatePhoneNumberContactFlow(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Connect Phone Number (%s) Contact Flow (%s) Association", phoneNumberID, contactFlowID), err.Error())

		return
	}
}

func (r *phoneNumberContactFlowAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		phoneNumberContactFlowAssociationIDParts = 3
	)
	parts, err := intflex.ExpandResourceId(request.ID, phoneNumberContactFlowAssociationIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("phone_number_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrInstanceID), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("contact_flow_id"), parts[2])...)
}

func findPhoneNumberContactFlowAssociationByThreePartKey(ctx context.Context, conn *connect.Client, phoneNumberID, instanceID, contactFlowID string) (*awstypes.FlowAssociationSummary, error) {
	phoneNumber, err := findPhoneNumberByID(ctx, conn, phoneNumberID)

	if err != nil {
		return nil, err
	}

	contactFlow, err := findContactFlowByTwoPartKey(ctx, conn, instanceID, contactFlowID)

	if err != nil {
		return nil, err
	}

	input := connect.ListFlowAssociationsInput{
		InstanceId:   aws.String(instanceID),
		ResourceType: awstypes.ListFlowAssociationResourceTypeVoicePhoneNumber,
	}

	return findFlowAssociation(ctx, conn, &input, func(v *awstypes.FlowAssociationSummary) bool {
		return aws.ToString(v.ResourceId) == aws.ToString(phoneNumber.PhoneNumberArn) && aws.ToString(v.FlowId) == aws.ToString(contactFlow.Arn)
	})
}

func findFlowAssociation(ctx context.Context, conn *connect.Client, input *connect.ListFlowAssociationsInput, filter tfslices.Predicate[*awstypes.FlowAssociationSummary]) (*awstypes.FlowAssociationSummary, error) {
	output, err := findFlowAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFlowAssociations(ctx context.Context, conn *connect.Client, input *connect.ListFlowAssociationsInput, filter tfslices.Predicate[*awstypes.FlowAssociationSummary]) ([]awstypes.FlowAssociationSummary, error) {
	var output []awstypes.FlowAssociationSummary

	pages := connect.NewListFlowAssociationsPaginator(conn, input)
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

		for _, v := range page.FlowAssociationSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

type phoneNumberContactFlowAssociationResourceModel struct {
	framework.WithRegionModel
	ContactFlowID types.String `tfsdk:"contact_flow_id"`
	InstanceID    types.String `tfsdk:"instance_id"`
	PhoneNumberID types.String `tfsdk:"phone_number_id"`
}
