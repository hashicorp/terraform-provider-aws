// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3files_mount_target", name="Mount Target")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetMountTargetOutput")
// @Testing(existsTakesT=true, destroyTakesT=true)
// @Testing(hasNoPreExistingResource="true")
func newMountTargetResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &mountTargetResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type mountTargetResource struct {
	framework.ResourceWithModel[mountTargetResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *mountTargetResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Availability Zone ID",
			},
			names.AttrFileSystemID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "File system ID",
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrIPAddressType: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.IpAddressType](),
				Optional:    true,
				Description: "IP address type",
			},
			"ipv4_address": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "IPv4 address",
			},
			"ipv6_address": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "IPv6 address",
			},
			names.AttrNetworkInterfaceID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Network interface ID",
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "AWS account ID of the owner",
			},
			names.AttrSecurityGroups: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Security group IDs",
			},
			names.AttrStatus: schema.StringAttribute{
				Computed:    true,
				Description: "Mount target status",
			},
			names.AttrStatusMessage: schema.StringAttribute{
				Computed:    true,
				Description: "Status message",
			},
			names.AttrSubnetID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Subnet ID",
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "VPC ID",
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *mountTargetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data mountTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.CreateMountTargetInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix("MountTarget")))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateMountTarget(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.MountTargetId)

	mountTarget, err := waitMountTargetCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, mountTarget, &data, fwflex.WithFieldNamePrefix("MountTarget")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *mountTargetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data mountTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	output, err := findMountTargetByID(ctx, conn, data.ID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	flattenMountTargetResource(ctx, output, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func flattenMountTargetResource(ctx context.Context, output *s3files.GetMountTargetOutput, data *mountTargetResourceModel, diags *diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	smerr.AddEnrich(ctx, diags, fwflex.Flatten(ctx, output, data, fwflex.WithFieldNamePrefix("MountTarget")))
}

func (r *mountTargetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new mountTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	if !new.SecurityGroups.Equal(old.SecurityGroups) {
		input := s3files.UpdateMountTargetInput{
			MountTargetId: new.ID.ValueStringPointer(),
		}
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input, fwflex.WithFieldNamePrefix("MountTarget")))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateMountTarget(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ID.ValueString())
			return
		}

		_, err = waitMountTargetUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ID.ValueString())
			return
		}

		// Read the updated resource to get all computed values
		output, err := findMountTargetByID(ctx, conn, new.ID.ValueString())
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ID.ValueString())
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &new, fwflex.WithFieldNamePrefix("MountTarget")))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, new))
}

func (r *mountTargetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data mountTargetResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.DeleteMountTargetInput{
		MountTargetId: data.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteMountTarget(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return // Resource already deleted
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	_, err = waitMountTargetDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
	}
}

type mountTargetResourceModel struct {
	framework.WithRegionModel
	AvailabilityZoneID types.String                               `tfsdk:"availability_zone_id"`
	FileSystemID       types.String                               `tfsdk:"file_system_id"`
	ID                 types.String                               `tfsdk:"id"`
	IpAddressType      fwtypes.StringEnum[awstypes.IpAddressType] `tfsdk:"ip_address_type"`
	Ipv4Address        types.String                               `tfsdk:"ipv4_address"`
	Ipv6Address        types.String                               `tfsdk:"ipv6_address"`
	NetworkInterfaceID types.String                               `tfsdk:"network_interface_id"`
	OwnerID            types.String                               `tfsdk:"owner_id"`
	SecurityGroups     fwtypes.SetValueOf[types.String]           `tfsdk:"security_groups"`
	Status             types.String                               `tfsdk:"status"`
	StatusMessage      types.String                               `tfsdk:"status_message"`
	SubnetID           types.String                               `tfsdk:"subnet_id"`
	Timeouts           timeouts.Value                             `tfsdk:"timeouts"`
	VPCID              types.String                               `tfsdk:"vpc_id"`
}

func findMountTargetByID(ctx context.Context, conn *s3files.Client, id string) (*s3files.GetMountTargetOutput, error) {
	input := s3files.GetMountTargetInput{
		MountTargetId: &id,
	}

	output, err := conn.GetMountTarget(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func waitMountTargetCreated(ctx context.Context, conn *s3files.Client, id string, timeout time.Duration) (*s3files.GetMountTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateCreating),
		Target:  enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateError),
		Refresh: statusMountTarget(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3files.GetMountTargetOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMountTargetUpdated(ctx context.Context, conn *s3files.Client, id string, timeout time.Duration) (*s3files.GetMountTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateUpdating),
		Target:  enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateError),
		Refresh: statusMountTarget(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3files.GetMountTargetOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMountTargetDeleted(ctx context.Context, conn *s3files.Client, id string, timeout time.Duration) (*s3files.GetMountTargetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateDeleting),
		Target:  []string{},
		Refresh: statusMountTarget(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3files.GetMountTargetOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusMountTarget(_ context.Context, conn *s3files.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findMountTargetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		if output != nil && output.Status == awstypes.LifeCycleStateError {
			return output, string(output.Status), smarterr.Errorf("in \"%s\" state with status message: %s", string(output.Status), aws.ToString(output.StatusMessage))
		}

		return output, string(output.Status), nil
	}
}
