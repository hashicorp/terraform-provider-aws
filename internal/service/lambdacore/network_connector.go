// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdacore

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdacore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdacore/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_lambdacore_network_connector", name="Network Connector")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambdacore;lambdacore.GetNetworkConnectorOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newNetworkConnectorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &networkConnectorResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameNetworkConnector = "Network Connector"
)

type networkConnectorResource struct {
	framework.ResourceWithModel[networkConnectorResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *networkConnectorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operator_role": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.NetworkConnectorState](),
				Computed:   true,
			},
			"state_reason": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"vpc_egress_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcEgressConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"associated_compute_resource_types": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringEnumType[awstypes.ComputeResourceType](),
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
						},
						"network_protocol": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.NetworkProtocol](),
							Optional:   true,
							Computed:   true,
						},
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

func (r *networkConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaCoreClient(ctx)

	var plan networkConnectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambdacore.CreateNetworkConnectorInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	input.ClientToken = aws.String(create.UniqueId(ctx))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateNetworkConnector(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)

	outWait, err := waitNetworkConnectorCreated(ctx, conn, plan.Arn.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, outWait, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *networkConnectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaCoreClient(ctx)

	var state networkConnectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNetworkConnectorByARN(ctx, conn, state.Arn.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Arn.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *networkConnectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaCoreClient(ctx)

	var plan, state networkConnectorResourceModel
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
		var input lambdacore.UpdateNetworkConnectorInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		input.Identifier = plan.Arn.ValueStringPointer()
		input.ClientToken = aws.String(create.UniqueId(ctx))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateNetworkConnector(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Arn.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Arn.String())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		outWait, err := waitNetworkConnectorUpdated(ctx, conn, plan.Arn.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Arn.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, outWait, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *networkConnectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaCoreClient(ctx)

	var state networkConnectorResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := lambdacore.DeleteNetworkConnectorInput{
		Identifier: state.Arn.ValueStringPointer(),
	}

	_, err := conn.DeleteNetworkConnector(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Arn.ValueString())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitNetworkConnectorDeleted(ctx, conn, state.Arn.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Arn.ValueString())
		return
	}
}

func waitNetworkConnectorCreated(ctx context.Context, conn *lambdacore.Client, arn string, timeout time.Duration) (*lambdacore.GetNetworkConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.NetworkConnectorStatePending),
		Target:                    enum.Slice(awstypes.NetworkConnectorStateActive),
		Refresh:                   statusNetworkConnector(conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdacore.GetNetworkConnectorOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitNetworkConnectorUpdated(ctx context.Context, conn *lambdacore.Client, arn string, timeout time.Duration) (*lambdacore.GetNetworkConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.NetworkConnectorStatePending),
		Target:                    enum.Slice(awstypes.NetworkConnectorStateActive),
		Refresh:                   statusNetworkConnector(conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdacore.GetNetworkConnectorOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitNetworkConnectorDeleted(ctx context.Context, conn *lambdacore.Client, arn string, timeout time.Duration) (*lambdacore.GetNetworkConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NetworkConnectorStateActive, awstypes.NetworkConnectorStateDeleting),
		Target:  []string{},
		Refresh: statusNetworkConnector(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdacore.GetNetworkConnectorOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusNetworkConnector(conn *lambdacore.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findNetworkConnectorByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findNetworkConnectorByARN(ctx context.Context, conn *lambdacore.Client, arn string) (*lambdacore.GetNetworkConnectorOutput, error) {
	input := lambdacore.GetNetworkConnectorInput{
		Identifier: aws.String(arn),
	}

	out, err := conn.GetNetworkConnector(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type networkConnectorResourceModel struct {
	framework.WithRegionModel
	Arn           types.String                                                 `tfsdk:"arn"`
	Configuration fwtypes.ListNestedObjectValueOf[vpcEgressConfigurationModel] `tfsdk:"vpc_egress_configuration"`
	Name          types.String                                                 `tfsdk:"name"`
	OperatorRole  fwtypes.ARN                                                  `tfsdk:"operator_role"`
	State         fwtypes.StringEnum[awstypes.NetworkConnectorState]           `tfsdk:"state"`
	StateReason   types.String                                                 `tfsdk:"state_reason"`
	Timeouts      timeouts.Value                                               `tfsdk:"timeouts"`
}

type vpcEgressConfigurationModel struct {
	AssociatedComputeResourceTypes fwtypes.ListOfStringEnum[awstypes.ComputeResourceType] `tfsdk:"associated_compute_resource_types"`
	NetworkProtocol                fwtypes.StringEnum[awstypes.NetworkProtocol]           `tfsdk:"network_protocol"`
	SecurityGroupIDs               fwtypes.SetOfString                                    `tfsdk:"security_group_ids"`
	SubnetIDs                      fwtypes.SetOfString                                    `tfsdk:"subnet_ids"`
}

var _ flex.Expander = vpcEgressConfigurationModel{}

// Configuration is a tagged union in the service API; VpcEgressConfiguration is
// its only member today, so the resource models it as a flat block and this
// expander wraps it back into the union member.
func (m vpcEgressConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var value awstypes.NetworkConnectorVpcEgressConfiguration
	diags.Append(flex.Expand(ctx, m, &value)...)
	if diags.HasError() {
		return nil, diags
	}

	return &awstypes.NetworkConnectorConfigurationMemberVpcEgressConfiguration{
		Value: value,
	}, diags
}

var _ flex.Flattener = &vpcEgressConfigurationModel{}

func (m *vpcEgressConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.NetworkConnectorConfigurationMemberVpcEgressConfiguration:
		diags.Append(flex.Flatten(ctx, t.Value, m)...)
	}

	return diags
}

func sweepNetworkConnectors(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := lambdacore.ListNetworkConnectorsInput{}
	conn := client.LambdaCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := lambdacore.NewListNetworkConnectorsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.NetworkConnectors {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newNetworkConnectorResource, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}
