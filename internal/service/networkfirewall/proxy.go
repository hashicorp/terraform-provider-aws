// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_proxy", name="Proxy")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preIdentityVersion="v5.100.0")
func newResourceProxy(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProxy{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultUpdateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(60 * time.Minute)

	return r, nil
}

type resourceProxy struct {
	framework.ResourceWithModel[resourceProxyModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *resourceProxy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreateTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			"nat_gateway_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_dns_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"proxy_configuration_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("proxy_configuration_arn"),
						path.MatchRoot("proxy_configuration_name"),
					),
				},
			},
			"proxy_configuration_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("proxy_configuration_arn"),
						path.MatchRoot("proxy_configuration_name"),
					),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"update_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"update_token": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vpc_endpoint_service_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"listener_properties": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[listenerPropertiesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(2),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrPort: schema.Int32Attribute{
							Required: true,
						},
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ListenerPropertyType](),
							Required:   true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"tls_intercept_properties": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tlsInterceptPropertiesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"pca_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
						"tls_intercept_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.TlsInterceptMode](),
							Computed:   true,
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceProxy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan resourceProxyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input networkfirewall.CreateProxyInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Proxy"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateProxy(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating NetworkFirewall Proxy (%s)", plan.ProxyName.ValueString()), err.Error())
		return
	}
	if out == nil || out.Proxy == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating NetworkFirewall Proxy (%s)", plan.ProxyName.ValueString()), errors.New("empty output").Error())
		return
	}

	arn := aws.ToString(out.Proxy.ProxyArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	describeOut, err := waitProxyCreated(ctx, conn, arn, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall Proxy (%s) create", arn), err.Error())
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, describeOut.Proxy, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.ProxyArn
	plan.UpdateToken = flex.StringToFramework(ctx, describeOut.UpdateToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceProxy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceProxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findProxyByARN(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading NetworkFirewall Proxy (%s)", state.ID.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.Proxy, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)

	setTagsOut(ctx, out.Proxy.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceProxy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan, state resourceProxyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state, flex.WithIgnoredField("Tags"), flex.WithIgnoredField("TagsAll"))
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := networkfirewall.UpdateProxyInput{
			ProxyArn:     state.ProxyArn.ValueStringPointer(),
			NatGatewayId: state.NatGatewayId.ValueStringPointer(),
			UpdateToken:  state.UpdateToken.ValueStringPointer(),
		}

		// Handle TlsInterceptProperties
		resp.Diagnostics.Append(flex.Expand(ctx, plan.TlsInterceptProperties, &input.TlsInterceptProperties)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Handle ListenerProperties changes - determine what to add/remove
		var planListeners, stateListeners []listenerPropertiesModel
		resp.Diagnostics.Append(plan.ListenerProperties.ElementsAs(ctx, &planListeners, false)...)
		resp.Diagnostics.Append(state.ListenerProperties.ElementsAs(ctx, &stateListeners, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Build maps for comparison
		stateListenerMap := make(map[string]listenerPropertiesModel)
		for _, l := range stateListeners {
			key := fmt.Sprintf("%d-%s", l.Port.ValueInt32(), l.Type.ValueString())
			stateListenerMap[key] = l
		}

		planListenerMap := make(map[string]listenerPropertiesModel)
		for _, l := range planListeners {
			key := fmt.Sprintf("%d-%s", l.Port.ValueInt32(), l.Type.ValueString())
			planListenerMap[key] = l
		}

		// Find listeners to add (in plan but not in state)
		for key, l := range planListenerMap {
			if _, exists := stateListenerMap[key]; !exists {
				input.ListenerPropertiesToAdd = append(input.ListenerPropertiesToAdd, awstypes.ListenerPropertyRequest{
					Port: l.Port.ValueInt32Pointer(),
					Type: awstypes.ListenerPropertyType(l.Type.ValueString()),
				})
			}
		}

		// Find listeners to remove (in state but not in plan)
		for key, l := range stateListenerMap {
			if _, exists := planListenerMap[key]; !exists {
				input.ListenerPropertiesToRemove = append(input.ListenerPropertiesToRemove, awstypes.ListenerPropertyRequest{
					Port: l.Port.ValueInt32Pointer(),
					Type: awstypes.ListenerPropertyType(l.Type.ValueString()),
				})
			}
		}

		out, err := conn.UpdateProxy(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating NetworkFirewall Proxy (%s)", state.ID.ValueString()), err.Error())
			return
		}
		if out == nil || out.Proxy == nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating NetworkFirewall Proxy (%s)", state.ID.ValueString()), errors.New("empty output").Error())
			return
		}

		// Wait for modification to complete
		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		describeOut, err := waitProxyUpdated(ctx, conn, state.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall Proxy (%s) update", state.ID.ValueString()), err.Error())
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, describeOut.Proxy, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		plan.UpdateToken = flex.StringToFramework(ctx, describeOut.UpdateToken)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceProxy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceProxyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := networkfirewall.DeleteProxyInput{
		ProxyArn:     state.ProxyArn.ValueStringPointer(),
		NatGatewayId: state.NatGatewayId.ValueStringPointer(),
	}

	_, err := conn.DeleteProxy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("deleting NetworkFirewall Proxy (%s)", state.ID.ValueString()), err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitProxyDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall Proxy (%s) delete", state.ID.ValueString()), err.Error())
		return
	}
}

func findProxyByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeProxyOutput, error) {
	input := networkfirewall.DescribeProxyInput{
		ProxyArn: aws.String(arn),
	}

	out, err := conn.DescribeProxy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.Proxy == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if out.Proxy.DeleteTime != nil {
		return nil, &sdkretry.NotFoundError{
			Message: "resource is deleted",
		}
	}

	return out, nil
}

func statusProxy(ctx context.Context, conn *networkfirewall.Client, arn string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findProxyByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Proxy.ProxyState), nil
	}
}

func statusProxyModify(ctx context.Context, conn *networkfirewall.Client, arn string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findProxyByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Proxy.ProxyModifyState), nil
	}
}

func waitProxyCreated(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeProxyOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProxyStateAttaching),
		Target:                    enum.Slice(awstypes.ProxyStateAttached),
		Refresh:                   statusProxy(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeProxyOutput); ok {
		return out, err
	}

	return nil, err
}

func waitProxyUpdated(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeProxyOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProxyModifyStateModifying),
		Target:                    enum.Slice(awstypes.ProxyModifyStateCompleted),
		Refresh:                   statusProxyModify(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeProxyOutput); ok {
		return out, err
	}

	return nil, err
}

func waitProxyDeleted(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeProxyOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProxyStateAttached, awstypes.ProxyStateDetaching),
		Target:  []string{},
		Refresh: statusProxy(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeProxyOutput); ok {
		return out, err
	}

	return nil, err
}

type resourceProxyModel struct {
	framework.WithRegionModel
	CreateTime             timetypes.RFC3339                                            `tfsdk:"create_time"`
	ID                     types.String                                                 `tfsdk:"id"`
	ListenerProperties     fwtypes.ListNestedObjectValueOf[listenerPropertiesModel]     `tfsdk:"listener_properties"`
	NatGatewayId           types.String                                                 `tfsdk:"nat_gateway_id"`
	PrivateDNSName         types.String                                                 `tfsdk:"private_dns_name"`
	ProxyArn               types.String                                                 `tfsdk:"arn"`
	ProxyConfigurationArn  fwtypes.ARN                                                  `tfsdk:"proxy_configuration_arn"`
	ProxyConfigurationName types.String                                                 `tfsdk:"proxy_configuration_name"`
	ProxyName              types.String                                                 `tfsdk:"name"`
	Tags                   tftags.Map                                                   `tfsdk:"tags"`
	TagsAll                tftags.Map                                                   `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                               `tfsdk:"timeouts"`
	TlsInterceptProperties fwtypes.ListNestedObjectValueOf[tlsInterceptPropertiesModel] `tfsdk:"tls_intercept_properties"`
	UpdateTime             timetypes.RFC3339                                            `tfsdk:"update_time"`
	UpdateToken            types.String                                                 `tfsdk:"update_token"`
	VpcEndpointServiceName types.String                                                 `tfsdk:"vpc_endpoint_service_name"`
}

type listenerPropertiesModel struct {
	Port types.Int32                                       `tfsdk:"port"`
	Type fwtypes.StringEnum[awstypes.ListenerPropertyType] `tfsdk:"type"`
}

type tlsInterceptPropertiesModel struct {
	PcaArn           fwtypes.ARN                                   `tfsdk:"pca_arn"`
	TlsInterceptMode fwtypes.StringEnum[awstypes.TlsInterceptMode] `tfsdk:"tls_intercept_mode"`
}
