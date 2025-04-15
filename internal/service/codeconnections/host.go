// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeconnections

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections"
	awstypes "github.com/aws/aws-sdk-go-v2/service/codeconnections/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_codeconnections_host", name="Host")
// @Tags(identifierAttribute="arn")
func newHostResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &hostResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameHost = "Host"
)

type hostResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *hostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"provider_endpoint": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 512),
				},
			},
			"provider_type": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ProviderType](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrVPCConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelVPCConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							ElementType: types.StringType,
						},
						names.AttrSubnetIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							ElementType: types.StringType,
						},
						"tls_certificate": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 16384),
							},
						},
						names.AttrVPCID: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(12, 21),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CodeConnectionsClient(ctx)

	var data hostResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input codeconnections.CreateHostInput

	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix("Host"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateHost(ctx, &input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionCreating, ResNameHost, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns
	data.HostArn = fwflex.StringToFramework(ctx, output.HostArn)
	data.ID = fwflex.StringToFramework(ctx, output.HostArn)

	host, err := waitHostPendingOrAvailable(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionWaitingForCreation, ResNameHost, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	data.Name = fwflex.StringToFramework(ctx, host.Name)
	data.ProviderEndpoint = fwflex.StringToFramework(ctx, host.ProviderEndpoint)
	data.ProviderType = fwtypes.StringEnumValue(host.ProviderType)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CodeConnectionsClient(ctx)

	var data hostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpcConfiguration, d := data.VPCConfiguration.ToPtr(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findHostByARN(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionSetting, ResNameHost, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	if vpcConfiguration != nil && out.VpcConfiguration != nil {
		out.VpcConfiguration.TlsCertificate = fwflex.StringFromFramework(ctx, vpcConfiguration.TlsCertificate)
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var new, old hostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CodeConnectionsClient(ctx)

	if !new.ProviderEndpoint.Equal(old.ProviderEndpoint) ||
		!new.VPCConfiguration.Equal(old.VPCConfiguration) {
		input := codeconnections.UpdateHostInput{
			HostArn: new.HostArn.ValueStringPointer(),
		}

		out, err := conn.UpdateHost(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CodeConnections, create.ErrActionUpdating, ResNameHost, new.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CodeConnections, create.ErrActionUpdating, ResNameHost, new.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, new.Timeouts)
		_, err = waitHostPendingOrAvailable(ctx, conn, new.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CodeConnections, create.ErrActionWaitingForUpdate, ResNameHost, new.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CodeConnectionsClient(ctx)

	var state hostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := codeconnections.DeleteHostInput{
		HostArn: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteHost(ctx, &input)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionDeleting, ResNameHost, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitHostDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionWaitingForDeletion, ResNameHost, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *hostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

const (
	hostStatusAvailable         = "AVAILABLE"
	hostStatusPending           = "PENDING"
	hostStatusVPCConfigDeleting = "VPC_CONFIG_DELETING"
	// hostStatusVPCConfigFailedInitialization = "VPC_CONFIG_FAILED_INITIALIZATION"
	hostStatusVPCConfigInitializing = "VPC_CONFIG_INITIALIZING"
)

func waitHostPendingOrAvailable(ctx context.Context, conn *codeconnections.Client, id string, timeout time.Duration) (*awstypes.Host, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{hostStatusVPCConfigInitializing},
		Target:                    []string{hostStatusPending, hostStatusAvailable},
		Refresh:                   statusHost(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Host); ok {
		return out, err
	}

	return nil, err
}

func waitHostDeleted(ctx context.Context, conn *codeconnections.Client, id string, timeout time.Duration) (*awstypes.Host, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{hostStatusVPCConfigDeleting},
		Target:  []string{},
		Refresh: statusHost(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Host); ok {
		return out, err
	}

	return nil, err
}

func statusHost(ctx context.Context, conn *codeconnections.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findHostByARN(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

func findHostByARN(ctx context.Context, conn *codeconnections.Client, arn string) (*awstypes.Host, error) {
	input := &codeconnections.GetHostInput{
		HostArn: aws.String(arn),
	}

	output, err := conn.GetHost(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Name == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	host := &awstypes.Host{
		HostArn:          aws.String(arn),
		Name:             output.Name,
		ProviderEndpoint: output.ProviderEndpoint,
		ProviderType:     output.ProviderType,
		Status:           output.Status,
		VpcConfiguration: output.VpcConfiguration,
	}

	return host, nil
}

type hostResourceModel struct {
	HostArn          types.String                                                      `tfsdk:"arn"`
	ID               types.String                                                      `tfsdk:"id"`
	Name             types.String                                                      `tfsdk:"name"`
	ProviderEndpoint types.String                                                      `tfsdk:"provider_endpoint"`
	ProviderType     fwtypes.StringEnum[awstypes.ProviderType]                         `tfsdk:"provider_type"`
	Tags             tftags.Map                                                        `tfsdk:"tags"`
	TagsAll          tftags.Map                                                        `tfsdk:"tags_all"`
	Timeouts         timeouts.Value                                                    `tfsdk:"timeouts"`
	VPCConfiguration fwtypes.ListNestedObjectValueOf[customModelVPCConfigurationModel] `tfsdk:"vpc_configuration"`
}

type customModelVPCConfigurationModel struct {
	SecurityGroupIDs fwtypes.SetValueOf[types.String] `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
	TlsCertificate   types.String                     `tfsdk:"tls_certificate"`
	VpcId            types.String                     `tfsdk:"vpc_id"`
}
