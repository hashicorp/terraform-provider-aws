// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdacore

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdacore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdacore/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 140),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operator_role": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[networkConnectorConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("vpc_egress_configuration"),
						),
					},
					Blocks: map[string]schema.Block{
						"vpc_egress_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[networkConnectorVPCEgressConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"associated_compute_resource_types": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringEnumType[awstypes.ComputeResourceType](),
										Required:    true,
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

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input lambdacore.CreateNetworkConnectorInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := tfresource.RetryWhenIsAErrorMessageContains[*lambdacore.CreateNetworkConnectorOutput, *awstypes.InvalidParameterValueException](ctx, propagationTimeout, func(ctx context.Context) (*lambdacore.CreateNetworkConnectorOutput, error) {
		return conn.CreateNetworkConnector(ctx, &input)
	}, "The service is unable to assume")
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	// Set values for unknowns.
	arn := aws.ToString(out.Arn)
	plan.ARN = fwflex.StringValueToFramework(ctx, arn)

	if _, err := waitNetworkConnectorCreated(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts)); err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root(names.AttrARN), arn)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
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

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findNetworkConnectorByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
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

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, plan.ARN)
		var input lambdacore.UpdateNetworkConnectorInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(create.UniqueId(ctx))
		input.Identifier = aws.String(arn)

		_, err := tfresource.RetryWhenIsAErrorMessageContains[*lambdacore.UpdateNetworkConnectorOutput, *awstypes.InvalidParameterValueException](ctx, propagationTimeout, func(ctx context.Context) (*lambdacore.UpdateNetworkConnectorOutput, error) {
			return conn.UpdateNetworkConnector(ctx, &input)
		}, "The service is unable to assume")
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		if _, err := waitNetworkConnectorUpdated(ctx, conn, arn, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
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

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := lambdacore.DeleteNetworkConnectorInput{
		Identifier: aws.String(arn),
	}

	_, err := conn.DeleteNetworkConnector(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitNetworkConnectorDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func (r *networkConnectorResource) flatten(ctx context.Context, connector *lambdacore.GetNetworkConnectorOutput, data *networkConnectorResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, connector, data)...)
	if diags.HasError() {
		return diags
	}

	return diags
}

func waitNetworkConnectorCreated(ctx context.Context, conn *lambdacore.Client, arn string, timeout time.Duration) (*lambdacore.GetNetworkConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.NetworkConnectorStatePending),
		Target:                    enum.Slice(awstypes.NetworkConnectorStateActive),
		Refresh:                   statusNetworkConnector(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdacore.GetNetworkConnectorOutput); ok {
		if stateReason := aws.ToString(out.StateReason); stateReason != "" {
			retry.SetLastError(err, fmt.Errorf("%s: %s", out.StateReasonCode, stateReason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitNetworkConnectorUpdated(ctx context.Context, conn *lambdacore.Client, arn string, timeout time.Duration) (*lambdacore.GetNetworkConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.NetworkConnectorLastUpdateStatusInProgress),
		Target:                    enum.Slice(awstypes.NetworkConnectorLastUpdateStatusSuccessful),
		Refresh:                   lastUpdateStatusNetworkConnector(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdacore.GetNetworkConnectorOutput); ok {
		if stateReason := aws.ToString(out.LastUpdateStatusReason); stateReason != "" {
			retry.SetLastError(err, fmt.Errorf("%s: %s", out.LastUpdateStatusReasonCode, stateReason))
		}
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
		if stateReason := aws.ToString(out.StateReason); stateReason != "" {
			retry.SetLastError(err, fmt.Errorf("%s: %s", out.StateReasonCode, stateReason))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func lastUpdateStatusNetworkConnector(conn *lambdacore.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findNetworkConnectorByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.LastUpdateStatus), nil
	}
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

	return findNetworkConnector(ctx, conn, &input)
}

func findNetworkConnector(ctx context.Context, conn *lambdacore.Client, input *lambdacore.GetNetworkConnectorInput) (*lambdacore.GetNetworkConnectorOutput, error) {
	out, err := conn.GetNetworkConnector(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type networkConnectorResourceModel struct {
	framework.WithRegionModel
	ARN           types.String                                                        `tfsdk:"arn"`
	Configuration fwtypes.ListNestedObjectValueOf[networkConnectorConfigurationModel] `tfsdk:"configuration"`
	Name          types.String                                                        `tfsdk:"name"`
	OperatorRole  fwtypes.ARN                                                         `tfsdk:"operator_role"`
	Timeouts      timeouts.Value                                                      `tfsdk:"timeouts"`
}

type networkConnectorConfigurationModel struct {
	VPCEgressConfiguration fwtypes.ListNestedObjectValueOf[networkConnectorVPCEgressConfigurationModel] `tfsdk:"vpc_egress_configuration"`
}

var (
	_ fwflex.Expander  = networkConnectorConfigurationModel{}
	_ fwflex.Flattener = &networkConnectorConfigurationModel{}
)

func (m *networkConnectorConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.NetworkConnectorConfigurationMemberVpcEgressConfiguration:
		var data networkConnectorVPCEgressConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.VPCEgressConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("networkConnectorConfigurationModel.Flatten: %T", v))
	}
	return diags
}

func (m networkConnectorConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.VPCEgressConfiguration.IsNull():
		data, d := m.VPCEgressConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.NetworkConnectorConfigurationMemberVpcEgressConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

type networkConnectorVPCEgressConfigurationModel struct {
	AssociatedComputeResourceTypes fwtypes.ListOfStringEnum[awstypes.ComputeResourceType] `tfsdk:"associated_compute_resource_types"`
	NetworkProtocol                fwtypes.StringEnum[awstypes.NetworkProtocol]           `tfsdk:"network_protocol"`
	SecurityGroupIDs               fwtypes.SetOfString                                    `tfsdk:"security_group_ids"`
	SubnetIDs                      fwtypes.SetOfString                                    `tfsdk:"subnet_ids"`
}
