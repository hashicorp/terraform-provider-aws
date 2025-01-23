// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_vpclattice_service_network_resource_association", name="Service Network Resource Association")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/vpclattice;vpclattice.GetServiceNetworkResourceAssociationOutput")
func newResourceServiceNetworkResourceAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceServiceNetworkResourceAssociation{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameServiceNetworkResourceAssociation = "Service Network Resource Association"
)

type resourceServiceNetworkResourceAssociation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceServiceNetworkResourceAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_vpclattice_service_network_resource_association"
}

func (r *resourceServiceNetworkResourceAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"dns_entry": schema.ListAttribute{
				Computed:   true,
				CustomType: fwtypes.NewListNestedObjectTypeOf[dnsEntry](ctx),
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[dnsEntry](ctx),
				},
			},
			"id": framework.IDAttribute(),
			"resource_configuration_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_network_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceServiceNetworkResourceAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().VPCLatticeClient(ctx)

	var plan resourceServiceNetworkResourceAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input vpclattice.CreateServiceNetworkResourceAssociationInput

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("ServiceNetworkResourceAssociation"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateServiceNetworkResourceAssociation(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionCreating, ResNameServiceNetworkResourceAssociation, plan.ResourceConfigurationIdentifier.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionCreating, ResNameServiceNetworkResourceAssociation, plan.ResourceConfigurationIdentifier.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitServiceNetworkResourceAssociationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForCreation, ResNameServiceNetworkResourceAssociation, plan.ResourceConfigurationIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, created, &plan)...)
	plan.ResourceConfigurationIdentifier = flex.StringToFramework(ctx, created.ResourceConfigurationId)
	plan.ServiceNetworkIdentifier = flex.StringToFramework(ctx, created.ServiceNetworkId)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceServiceNetworkResourceAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().VPCLatticeClient(ctx)

	var state resourceServiceNetworkResourceAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findServiceNetworkResourceAssociationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionSetting, ResNameServiceNetworkResourceAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	state.ResourceConfigurationIdentifier = flex.StringToFramework(ctx, out.ResourceConfigurationId)
	state.ServiceNetworkIdentifier = flex.StringToFramework(ctx, out.ServiceNetworkId)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceServiceNetworkResourceAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceServiceNetworkResourceAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().VPCLatticeClient(ctx)

	var state resourceServiceNetworkResourceAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := vpclattice.DeleteServiceNetworkResourceAssociationInput{
		ServiceNetworkResourceAssociationIdentifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteServiceNetworkResourceAssociation(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionDeleting, ResNameServiceNetworkResourceAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitServiceNetworkResourceAssociationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameServiceNetworkResourceAssociation, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceServiceNetworkResourceAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceServiceNetworkResourceAssociation) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func waitServiceNetworkResourceAssociationCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkResourceAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ServiceNetworkServiceAssociationStatusCreateInProgress),
		Target:                    enum.Slice(awstypes.ServiceNetworkServiceAssociationStatusActive),
		Refresh:                   statusServiceNetworkResourceAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceNetworkResourceAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitServiceNetworkResourceAssociationDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkResourceAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ServiceNetworkServiceAssociationStatusActive, awstypes.ServiceNetworkServiceAssociationStatusDeleteInProgress),
		Target:  []string{},
		Refresh: statusServiceNetworkResourceAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceNetworkResourceAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusServiceNetworkResourceAssociation(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findServiceNetworkResourceAssociationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findServiceNetworkResourceAssociationByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetServiceNetworkResourceAssociationOutput, error) {
	in := &vpclattice.GetServiceNetworkResourceAssociationInput{
		ServiceNetworkResourceAssociationIdentifier: aws.String(id),
	}

	out, err := conn.GetServiceNetworkResourceAssociation(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

type resourceServiceNetworkResourceAssociationModel struct {
	ARN                             types.String                              `tfsdk:"arn"`
	ID                              types.String                              `tfsdk:"id"`
	DnsEntry                        fwtypes.ListNestedObjectValueOf[dnsEntry] `tfsdk:"dns_entry"`
	ResourceConfigurationIdentifier types.String                              `tfsdk:"resource_configuration_identifier"`
	ServiceNetworkIdentifier        types.String                              `tfsdk:"service_network_identifier"`
	Tags                            tftags.Map                                `tfsdk:"tags"`
	TagsAll                         tftags.Map                                `tfsdk:"tags_all"`
	Timeouts                        timeouts.Value                            `tfsdk:"timeouts"`
}

type dnsEntry struct {
	DomainName   types.String `tfsdk:"domain_name"`
	HostedZoneId types.String `tfsdk:"hosted_zone_id"`
}
