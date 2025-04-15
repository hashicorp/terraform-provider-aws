// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpclattice_service_network_resource_association", name="Service Network Resource Association")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/vpclattice;vpclattice.GetServiceNetworkResourceAssociationOutput")
func newServiceNetworkResourceAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &serviceNetworkResourceAssociationResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type serviceNetworkResourceAssociationResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[serviceNetworkResourceAssociationResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *serviceNetworkResourceAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"dns_entry":   framework.ResourceComputedListOfObjectsAttribute[dnsEntryModel](ctx, listplanmodifier.UseStateForUnknown()),
			names.AttrID:  framework.IDAttribute(),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *serviceNetworkResourceAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data serviceNetworkResourceAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	var input vpclattice.CreateServiceNetworkResourceAssociationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.ResourceConfigurationIdentifier = fwflex.StringFromFramework(ctx, data.ResourceConfigurationID)
	input.ServiceNetworkIdentifier = fwflex.StringFromFramework(ctx, data.ServiceNetworkID)
	input.Tags = getTagsIn(ctx)

	outputCSNRA, err := conn.CreateServiceNetworkResourceAssociation(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating VPCLattice Service Network Resource Association", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, outputCSNRA.Id)

	outputGSNRA, err := waitServiceNetworkResourceAssociationCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPCLattice Service Network Resource Association (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGSNRA, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *serviceNetworkResourceAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data serviceNetworkResourceAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	output, err := findServiceNetworkResourceAssociationByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPCLattice Resource Gateway (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *serviceNetworkResourceAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data serviceNetworkResourceAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	input := vpclattice.DeleteServiceNetworkResourceAssociationInput{
		ServiceNetworkResourceAssociationIdentifier: data.ID.ValueStringPointer(),
	}
	_, err := conn.DeleteServiceNetworkResourceAssociation(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPCLattice Service Network Resource Association (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitServiceNetworkResourceAssociationDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for VPCLattice Service Network Resource Association (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findServiceNetworkResourceAssociationByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetServiceNetworkResourceAssociationOutput, error) {
	input := vpclattice.GetServiceNetworkResourceAssociationInput{
		ServiceNetworkResourceAssociationIdentifier: aws.String(id),
	}

	output, err := conn.GetServiceNetworkResourceAssociation(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusServiceNetworkResourceAssociation(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findServiceNetworkResourceAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitServiceNetworkResourceAssociationCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkResourceAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ServiceNetworkServiceAssociationStatusCreateInProgress),
		Target:                    enum.Slice(awstypes.ServiceNetworkServiceAssociationStatusActive),
		Refresh:                   statusServiceNetworkResourceAssociation(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetServiceNetworkResourceAssociationOutput); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureReason)))

		return output, err
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

	if output, ok := outputRaw.(*vpclattice.GetServiceNetworkResourceAssociationOutput); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureReason)))

		return output, err
	}

	return nil, err
}

type serviceNetworkResourceAssociationResourceModel struct {
	ARN                     types.String                                   `tfsdk:"arn"`
	ID                      types.String                                   `tfsdk:"id"`
	DNSEntry                fwtypes.ListNestedObjectValueOf[dnsEntryModel] `tfsdk:"dns_entry"`
	ResourceConfigurationID types.String                                   `tfsdk:"resource_configuration_identifier"`
	ServiceNetworkID        types.String                                   `tfsdk:"service_network_identifier"`
	Tags                    tftags.Map                                     `tfsdk:"tags"`
	TagsAll                 tftags.Map                                     `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                 `tfsdk:"timeouts"`
}

type dnsEntryModel struct {
	DomainName   types.String `tfsdk:"domain_name"`
	HostedZoneID types.String `tfsdk:"hosted_zone_id"`
}
