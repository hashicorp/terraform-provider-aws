// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package mailmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mailmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_mailmanager_ingress_point", name="Ingress Point")
// @IdentityAttribute("id")
// @Tags(identifierAttribute="arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckIngressPoint")
// @Testing(serialize=true)
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newIngressPointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &ingressPointResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const ResNameIngressPoint = "Ingress Point"

type ingressPointResource struct {
	framework.ResourceWithModel[ingressPointResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *ingressPointResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"a_record": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"last_updated_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"rule_set_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IngressPointStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"tls_policy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TlsPolicy](),
				Optional:   true,
				Computed:   true,
			},
			"traffic_policy_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IngressPointType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"ingress_point_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ingressPointConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"secret_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
						"smtp_password_wo": schema.StringAttribute{
							Optional:  true,
							WriteOnly: true,
							Sensitive: true,
							Validators: []validator.String{
								stringvalidator.AlsoRequires(path.Expressions{
									path.MatchRelative().AtParent().AtName("smtp_password_wo_version"),
								}...),
							},
						},
						"smtp_password_wo_version": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.AlsoRequires(path.Expressions{
									path.MatchRelative().AtParent().AtName("smtp_password_wo"),
								}...),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"tls_auth_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[tlsAuthConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"trust_store": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[trustStoreModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"ca_content": schema.StringAttribute{
													Required: true,
												},
												"crl_content": schema.StringAttribute{
													Optional: true,
												},
												names.AttrKMSKeyARN: schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Optional:   true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrNetworkConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[networkConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"private_network_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[privateNetworkConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrVPCEndpointID: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"public_network_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[publicNetworkConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"ip_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.IpType](),
										Required:   true,
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

func (r *ingressPointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data, config ingressPointResourceModel

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MailManagerClient(ctx)

	var input mailmanager.CreateIngressPointInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input, flex.WithFieldNamePrefix("IngressPoint")))
	if resp.Diagnostics.HasError() {
		return
	}

	// smtp_password_wo is write-only. it is only present in Config, not Plan.
	configIPCfg, d := config.IngressPointConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	if configIPCfg != nil && !configIPCfg.SMTPPasswordWO.IsNull() {
		input.IngressPointConfiguration = &awstypes.IngressPointConfigurationMemberSmtpPassword{
			Value: configIPCfg.SMTPPasswordWO.ValueString(),
		}
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateIngressPoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	ingressPointID := aws.ToString(out.IngressPointId)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrID), ingressPointID))

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	ingressPointOut, err := waitIngressPointActive(ctx, conn, ingressPointID, createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ingressPointID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, ingressPointOut, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *ingressPointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ingressPointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MailManagerClient(ctx)

	ingressPointID := state.ID.ValueString()
	out, err := findIngressPointByID(ctx, conn, ingressPointID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ingressPointID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *ingressPointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config ingressPointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MailManagerClient(ctx)

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input mailmanager.UpdateIngressPointInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("IngressPoint")))
		if resp.Diagnostics.HasError() {
			return
		}

		// smtp_password_wo is write-only: it is only present in Config, not Plan.
		updateConfigIPCfg, d := config.IngressPointConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}
		if updateConfigIPCfg != nil && !updateConfigIPCfg.SMTPPasswordWO.IsNull() {
			input.IngressPointConfiguration = &awstypes.IngressPointConfigurationMemberSmtpPassword{
				Value: updateConfigIPCfg.SMTPPasswordWO.ValueString(),
			}
		}

		_, err := conn.UpdateIngressPoint(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		ingressPointOut, err := waitIngressPointActive(ctx, conn, state.ID.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, ingressPointOut, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.LastUpdatedTimestamp = state.LastUpdatedTimestamp
		plan.TLSPolicy = state.TLSPolicy
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *ingressPointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ingressPointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MailManagerClient(ctx)

	ingressPointID := state.ID.ValueString()
	input := mailmanager.DeleteIngressPointInput{
		IngressPointId: aws.String(ingressPointID),
	}

	_, err := conn.DeleteIngressPoint(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ingressPointID)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitIngressPointDeleted(ctx, conn, ingressPointID, deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ingressPointID)
	}
}

func (r *ingressPointResource) flatten(ctx context.Context, apiObject *mailmanager.GetIngressPointOutput, data *ingressPointResourceModel) diag.Diagnostics {
	// AWS always returns a NetworkConfiguration even when the user did not configure one.
	// Preserve the existing value to avoid drift.
	priorNetworkConfiguration := data.NetworkConfiguration

	diags := flex.Flatten(ctx, apiObject, data, flex.WithFieldNamePrefix("IngressPoint"))
	if diags.HasError() {
		return diags
	}

	if priorNetworkConfiguration.IsNull() {
		data.NetworkConfiguration = priorNetworkConfiguration
	}

	return diags
}

func waitIngressPointActive(ctx context.Context, conn *mailmanager.Client, id string, timeout time.Duration) (*mailmanager.GetIngressPointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.IngressPointStatusProvisioning,
			awstypes.IngressPointStatusUpdating,
		),
		Target:                    enum.Slice(awstypes.IngressPointStatusActive),
		Refresh:                   statusIngressPoint(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mailmanager.GetIngressPointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitIngressPointDeleted(ctx context.Context, conn *mailmanager.Client, id string, timeout time.Duration) (*mailmanager.GetIngressPointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.IngressPointStatusActive,
			awstypes.IngressPointStatusDeprovisioning,
		),
		Target:  []string{},
		Refresh: statusIngressPoint(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mailmanager.GetIngressPointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusIngressPoint(conn *mailmanager.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findIngressPointByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findIngressPointByID(ctx context.Context, conn *mailmanager.Client, id string) (*mailmanager.GetIngressPointOutput, error) {
	input := mailmanager.GetIngressPointInput{
		IngressPointId: aws.String(id),
	}

	out, err := conn.GetIngressPoint(ctx, &input)
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

var (
	_ flex.Expander  = ingressPointConfigurationModel{}
	_ flex.Flattener = &ingressPointConfigurationModel{}
	_ flex.Expander  = networkConfigurationModel{}
	_ flex.Flattener = &networkConfigurationModel{}
)

func (m ingressPointConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.SMTPPasswordWO.IsNull() && !m.SMTPPasswordWO.IsUnknown():
		return &awstypes.IngressPointConfigurationMemberSmtpPassword{
			Value: m.SMTPPasswordWO.ValueString(),
		}, diags
	case !m.SMTPPasswordWOVersion.IsNull():
		return &awstypes.IngressPointConfigurationMemberSmtpPassword{}, diags
	case !m.SecretARN.IsNull():
		return &awstypes.IngressPointConfigurationMemberSecretArn{
			Value: m.SecretARN.ValueString(),
		}, diags
	case !m.TLSAuthConfiguration.IsNull():
		var x awstypes.IngressPointConfigurationMemberTlsAuthConfiguration
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.TLSAuthConfiguration, &x.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &x, diags
	}

	return nil, diags
}

func (m *ingressPointConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.IngressPointAuthConfiguration:
		if t.SecretArn != nil {
			m.SecretARN = fwtypes.ARNValue(aws.ToString(t.SecretArn))
		}
		if t.TlsAuthConfiguration != nil {
			var tlsConf tlsAuthConfigurationModel
			smerr.AddEnrich(ctx, &diags, flex.Flatten(ctx, t.TlsAuthConfiguration, &tlsConf))
			if diags.HasError() {
				return diags
			}
			m.TLSAuthConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tlsConf)
		}
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("ingress point configuration flatten: %T", v))
	}

	return diags
}

func (m networkConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.PublicNetworkConfiguration.IsNull():
		var x awstypes.NetworkConfigurationMemberPublicNetworkConfiguration
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.PublicNetworkConfiguration, &x.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &x, diags
	case !m.PrivateNetworkConfiguration.IsNull():
		var x awstypes.NetworkConfigurationMemberPrivateNetworkConfiguration
		smerr.AddEnrich(ctx, &diags, flex.Expand(ctx, m.PrivateNetworkConfiguration, &x.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &x, diags
	}

	return nil, diags
}

func (m *networkConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	m.PrivateNetworkConfiguration = fwtypes.NewListNestedObjectValueOfNull[privateNetworkConfigurationModel](ctx)
	m.PublicNetworkConfiguration = fwtypes.NewListNestedObjectValueOfNull[publicNetworkConfigurationModel](ctx)

	switch t := v.(type) {
	case awstypes.NetworkConfigurationMemberPublicNetworkConfiguration:
		pub := publicNetworkConfigurationModel{
			IPType: fwtypes.StringEnumValue(t.Value.IpType),
		}
		m.PublicNetworkConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &pub)

	case awstypes.NetworkConfigurationMemberPrivateNetworkConfiguration:
		priv := privateNetworkConfigurationModel{
			VPCEndpointID: types.StringPointerValue(t.Value.VpcEndpointId),
		}
		m.PrivateNetworkConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &priv)

	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("network configuration flatten: %T", v))
	}

	return diags
}

type ingressPointResourceModel struct {
	framework.WithRegionModel
	ARecord                   types.String                                                    `tfsdk:"a_record"`
	ARN                       types.String                                                    `tfsdk:"arn"`
	CreatedTimestamp          timetypes.RFC3339                                               `tfsdk:"created_timestamp"`
	ID                        types.String                                                    `tfsdk:"id"`
	IngressPointConfiguration fwtypes.ListNestedObjectValueOf[ingressPointConfigurationModel] `tfsdk:"ingress_point_configuration"`
	LastUpdatedTimestamp      timetypes.RFC3339                                               `tfsdk:"last_updated_timestamp"`
	Name                      types.String                                                    `tfsdk:"name"`
	NetworkConfiguration      fwtypes.ListNestedObjectValueOf[networkConfigurationModel]      `tfsdk:"network_configuration"`
	RuleSetID                 types.String                                                    `tfsdk:"rule_set_id"`
	Status                    fwtypes.StringEnum[awstypes.IngressPointStatus]                 `tfsdk:"status"`
	Tags                      tftags.Map                                                      `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                      `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                  `tfsdk:"timeouts"`
	TLSPolicy                 fwtypes.StringEnum[awstypes.TlsPolicy]                          `tfsdk:"tls_policy"`
	TrafficPolicyID           types.String                                                    `tfsdk:"traffic_policy_id"`
	Type                      fwtypes.StringEnum[awstypes.IngressPointType]                   `tfsdk:"type"`
}

type ingressPointConfigurationModel struct {
	SecretARN             fwtypes.ARN                                                `tfsdk:"secret_arn"`
	SMTPPasswordWO        types.String                                               `tfsdk:"smtp_password_wo"`
	SMTPPasswordWOVersion types.Int64                                                `tfsdk:"smtp_password_wo_version"`
	TLSAuthConfiguration  fwtypes.ListNestedObjectValueOf[tlsAuthConfigurationModel] `tfsdk:"tls_auth_configuration"`
}

type tlsAuthConfigurationModel struct {
	TrustStore fwtypes.ListNestedObjectValueOf[trustStoreModel] `tfsdk:"trust_store"`
}

type trustStoreModel struct {
	CAContent  types.String `tfsdk:"ca_content"`
	CRLContent types.String `tfsdk:"crl_content"`
	KMSKeyARN  fwtypes.ARN  `tfsdk:"kms_key_arn"`
}

type networkConfigurationModel struct {
	PrivateNetworkConfiguration fwtypes.ListNestedObjectValueOf[privateNetworkConfigurationModel] `tfsdk:"private_network_configuration"`
	PublicNetworkConfiguration  fwtypes.ListNestedObjectValueOf[publicNetworkConfigurationModel]  `tfsdk:"public_network_configuration"`
}

type privateNetworkConfigurationModel struct {
	VPCEndpointID types.String `tfsdk:"vpc_endpoint_id"`
}

type publicNetworkConfigurationModel struct {
	IPType fwtypes.StringEnum[awstypes.IpType] `tfsdk:"ip_type"`
}
