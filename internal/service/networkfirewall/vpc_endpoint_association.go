// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package networkfirewall

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_vpc_endpoint_association", name="VPC Endpoint Association")
// @Tags(identifierAttribute="vpc_endpoint_association_arn")
func newVPCEndpointAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcEndpointAssociationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type vpcEndpointAssociationResource struct {
	framework.ResourceWithModel[vpcEndpointAssociationResourceModel]
	framework.WithTimeouts
}

func (r *vpcEndpointAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"firewall_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"vpc_endpoint_association_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vpc_endpoint_association_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vpc_endpoint_association_status": framework.ResourceComputedListOfObjectsAttribute[vpcEndpointAssociationStatusModel](ctx, listplanmodifier.UseStateForUnknown()),
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"subnet_mapping": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[subnetMappingModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrIPAddressType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.IPAddressType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrSubnetID: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *vpcEndpointAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcEndpointAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFirewallClient(ctx)

	var input networkfirewall.CreateVpcEndpointAssociationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	outputCVEA, err := conn.CreateVpcEndpointAssociation(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating NetworkFirewall VPC Endpoint Association", err.Error())

		return
	}

	arn := aws.ToString(outputCVEA.VpcEndpointAssociation.VpcEndpointAssociationArn)

	outputDVEA, err := waitVPCEndpointAssociationCreated(ctx, conn, arn, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall VPC Endpoint Association (%s) create", arn), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputDVEA.VpcEndpointAssociation.SubnetMapping, &data.SubnetMapping)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.VPCEndpointAssociationARN = fwflex.StringValueToFramework(ctx, arn)
	data.VPCEndpointAssociationID = fwflex.StringToFramework(ctx, outputDVEA.VpcEndpointAssociation.VpcEndpointAssociationId)
	vpcEndpointAssociationStatus, diags := flattenVPCEndpointAssociationStatus(ctx, outputDVEA.VpcEndpointAssociationStatus)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	data.VpcEndpointAssociationStatus = vpcEndpointAssociationStatus

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *vpcEndpointAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcEndpointAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFirewallClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.VPCEndpointAssociationARN)
	output, err := findVPCEndpointAssociationByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading NetworkFirewall VPC Endpoint Association (%s)", arn), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.VpcEndpointAssociation, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	vpcEndpointAssociationStatus, diags := flattenVPCEndpointAssociationStatus(ctx, output.VpcEndpointAssociationStatus)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	data.VpcEndpointAssociationStatus = vpcEndpointAssociationStatus

	setTagsOut(ctx, output.VpcEndpointAssociation.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcEndpointAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpcEndpointAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFirewallClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.VPCEndpointAssociationARN)
	input := networkfirewall.DeleteVpcEndpointAssociationInput{
		VpcEndpointAssociationArn: aws.String(arn),
	}
	_, err := conn.DeleteVpcEndpointAssociation(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting NetworkFirewall VPC Endpoint Association (%s)", arn), err.Error())

		return
	}

	if _, err := waitVPCEndpointAssociationDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall VPC Endpoint Association (%s) delete", arn), err.Error())

		return
	}
}

func (r *vpcEndpointAssociationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("vpc_endpoint_association_arn"), request, response)
}

func findVPCEndpointAssociation(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.DescribeVpcEndpointAssociationInput) (*networkfirewall.DescribeVpcEndpointAssociationOutput, error) {
	output, err := conn.DescribeVpcEndpointAssociation(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VpcEndpointAssociation == nil || output.VpcEndpointAssociationStatus == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findVPCEndpointAssociationByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeVpcEndpointAssociationOutput, error) {
	input := networkfirewall.DescribeVpcEndpointAssociationInput{
		VpcEndpointAssociationArn: aws.String(arn),
	}

	return findVPCEndpointAssociation(ctx, conn, &input)
}

func statusVPCEndpointAssociation(conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCEndpointAssociationByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.VpcEndpointAssociationStatus.Status), nil
	}
}

func waitVPCEndpointAssociationCreated(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeVpcEndpointAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.FirewallStatusValueProvisioning),
		Target:                    enum.Slice(awstypes.FirewallStatusValueReady),
		Refresh:                   statusVPCEndpointAssociation(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeVpcEndpointAssociationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointAssociationDeleted(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeVpcEndpointAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FirewallStatusValueReady, awstypes.FirewallStatusValueDeleting),
		Target:  []string{},
		Refresh: statusVPCEndpointAssociation(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeVpcEndpointAssociationOutput); ok {
		return output, err
	}

	return nil, err
}

func flattenVPCEndpointAssociationStatus(ctx context.Context, veas *awstypes.VpcEndpointAssociationStatus) (fwtypes.ListNestedObjectValueOf[vpcEndpointAssociationStatusModel], diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	if veas == nil {
		return fwtypes.NewListNestedObjectValueOfNull[vpcEndpointAssociationStatusModel](ctx), diags
	}

	var models []*associationSyncStateModel
	for az, syncState := range veas.AssociationSyncState {
		a := syncState.Attachment
		if a == nil {
			continue
		}

		attachment, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &attachmentModel{
			EndpointID:    fwflex.StringToFramework(ctx, a.EndpointId),
			SubnetID:      fwflex.StringToFramework(ctx, a.SubnetId),
			Status:        fwtypes.StringEnumValue(a.Status),
			StatusMessage: fwflex.StringToFramework(ctx, a.StatusMessage),
		})
		diags.Append(d...)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfNull[vpcEndpointAssociationStatusModel](ctx), diags
		}

		models = append(models, &associationSyncStateModel{
			Attachment:       attachment,
			AvailabilityZone: fwflex.StringValueToFramework(ctx, az),
		})
	}

	associationSyncState, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, models, nil)
	diags.Append(d...)
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfNull[vpcEndpointAssociationStatusModel](ctx), diags
	}

	vpcEndpointAssociationStatus, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &vpcEndpointAssociationStatusModel{
		AssociationSyncState: associationSyncState,
	})
	diags.Append(d...)
	if diags.HasError() {
		return fwtypes.NewListNestedObjectValueOfNull[vpcEndpointAssociationStatusModel](ctx), diags
	}

	return vpcEndpointAssociationStatus, diags
}

type vpcEndpointAssociationResourceModel struct {
	framework.WithRegionModel
	Description                  types.String                                                       `tfsdk:"description"`
	FirewallARN                  fwtypes.ARN                                                        `tfsdk:"firewall_arn"`
	SubnetMapping                fwtypes.ListNestedObjectValueOf[subnetMappingModel]                `tfsdk:"subnet_mapping"`
	Tags                         tftags.Map                                                         `tfsdk:"tags"`
	TagsAll                      tftags.Map                                                         `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                                     `tfsdk:"timeouts"`
	VPCEndpointAssociationARN    types.String                                                       `tfsdk:"vpc_endpoint_association_arn"`
	VPCEndpointAssociationID     types.String                                                       `tfsdk:"vpc_endpoint_association_id"`
	VpcEndpointAssociationStatus fwtypes.ListNestedObjectValueOf[vpcEndpointAssociationStatusModel] `tfsdk:"vpc_endpoint_association_status"`
	VPCID                        types.String                                                       `tfsdk:"vpc_id"`
}

type subnetMappingModel struct {
	SubnetId      types.String                               `tfsdk:"subnet_id"`
	IPAddressType fwtypes.StringEnum[awstypes.IPAddressType] `tfsdk:"ip_address_type"`
}

type vpcEndpointAssociationStatusModel struct {
	AssociationSyncState fwtypes.SetNestedObjectValueOf[associationSyncStateModel] `tfsdk:"association_sync_state"`
}

type associationSyncStateModel struct {
	Attachment       fwtypes.ListNestedObjectValueOf[attachmentModel] `tfsdk:"attachment"`
	AvailabilityZone types.String                                     `tfsdk:"availability_zone"`
}

type attachmentModel struct {
	EndpointID    types.String                                  `tfsdk:"endpoint_id"`
	SubnetID      types.String                                  `tfsdk:"subnet_id"`
	Status        fwtypes.StringEnum[awstypes.AttachmentStatus] `tfsdk:"status"`
	StatusMessage types.String                                  `tfsdk:"status_message"`
}
