// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_firewall_policy_v2", name="Firewall Policy V2")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @ArnFormat("firewall-policy-v2/{name}")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networkfirewall;networkfirewall.DescribeFirewallPolicyV2Output")
// @Testing(importIgnore="update_token", plannableImportAction="NoOp")
// @Testing(hasNoPreExistingResource=true)
func newFirewallPolicyV2Resource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &firewallPolicyV2Resource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type firewallPolicyV2Resource struct {
	framework.ResourceWithModel[firewallPolicyV2ResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *firewallPolicyV2Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	ruleVariablesBlock := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ruleVariablesV2Model](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"ip_set":   keyedDefinitionSetBlock[ipSetV2Model](ctx),
				"port_set": keyedDefinitionSetBlock[portSetV2Model](ctx),
			},
		},
	}
	referenceSetsBlock := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[referenceSetsV2Model](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"ip_set_reference": schema.SetNestedBlock{
					CustomType: fwtypes.NewSetNestedObjectTypeOf[ipSetReferenceV2Model](ctx),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							names.AttrKey: schema.StringAttribute{
								Required: true,
							},
							"reference_arn": schema.StringAttribute{
								CustomType: fwtypes.ARNType,
								Required:   true,
							},
						},
					},
				},
			},
		},
	}
	addressBlock := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[addressV2Model](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"address_definition": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 512),
				},
			},
			names.AttrEncryptionConfiguration: schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[encryptionConfigurationModel](ctx),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric characters and hyphens"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"update_token": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"firewall_policy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[firewallPolicyV2Model](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"firewall_default_action": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.PolicyDefaultAction](),
							Optional:   true,
						},
						"policy_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.PolicyMode](),
							Optional:   true,
						},
						"tls_inspection_configuration_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"firewall_rule_group_entry": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[firewallRuleGroupEntryModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrPriority: schema.Int32Attribute{
										Optional: true,
									},
									names.AttrResourceARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"inline_rule_group": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[firewallRuleGroupV2Model](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"reference_sets": referenceSetsBlock,
												"rule_variables": ruleVariablesBlock,
												"rules_source": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[firewallRulesSourceModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"rules_string": schema.StringAttribute{
																Optional: true,
															},
														},
														Blocks: map[string]schema.Block{
															"firewall_rule": firewallRuleBlock(ctx),
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"ips_rule_group_entry": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ipsRuleGroupEntryModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrResourceARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"inline_rule_group": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[ipsRuleGroupV2Model](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"reference_sets": referenceSetsBlock,
												"rule_variables": ruleVariablesBlock,
												"rules_source": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[ipsRulesSourceModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"rules_string": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"override": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[ipsRuleGroupOverrideModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrAction: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.OverrideAction](),
													Optional:   true,
												},
											},
										},
									},
								},
							},
						},
						"policy_variables": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[policyVariablesV2Model](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"rule_variable": keyedDefinitionSetBlock[ipSetV2Model](ctx),
								},
							},
						},
						"route53_resolved_ip_validation": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[route53ResolvedIPValidationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrAction: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.Route53ResolvedIPValidationAction](),
										Required:   true,
									},
									"dns_expiry_exception_policy": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.DNSExpiryExceptionPolicy](),
										Required:   true,
									},
									"dns_firewall_rule_group_id": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"bypass_destination_validation_ip_range": addressBlock,
									"source_ip_range":                        addressBlock,
								},
							},
						},
						"stateful_engine_options": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[statefulEngineOptionsV2Model](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"stream_exception_policy": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.StreamExceptionPolicyV2](),
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"flow_timeouts": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[flowTimeoutsV2Model](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"tcp_idle_timeout_seconds": schema.Int32Attribute{
													Optional: true,
													Validators: []validator.Int32{
														int32validator.Between(60, 6000),
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
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func firewallRuleBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[firewallRuleModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"accept_prior_hooks": schema.BoolAttribute{
					Optional: true,
				},
				names.AttrAction: schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.FirewallRuleAction](),
					Required:   true,
				},
				"action_scope": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.ActionScope](),
					Required:   true,
				},
				"alert_enabled": schema.BoolAttribute{
					Optional: true,
				},
				"bypass_ips_enabled": schema.BoolAttribute{
					Optional: true,
				},
				"dest_ips": schema.StringAttribute{
					Optional: true,
				},
				"dest_ports": schema.StringAttribute{
					Optional: true,
				},
				"hook": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.HookKey](),
					Required:   true,
				},
				names.AttrProtocol: schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.FirewallRuleProtocol](),
					Required:   true,
				},
				"source_ips": schema.StringAttribute{
					Optional: true,
				},
				"source_ports": schema.StringAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"rule_metadata": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ruleMetadataModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"rule_description": schema.StringAttribute{
								Optional: true,
							},
							"rule_enabled": schema.BoolAttribute{
								Optional: true,
							},
						},
					},
				},
				"rule_option": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ruleOptionV2Model](ctx),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"keyword": schema.StringAttribute{
								Required: true,
							},
							"settings": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								ElementType: types.StringType,
								Optional:    true,
							},
						},
					},
				},
			},
		},
	}
}

// keyedDefinitionSetBlock returns the schema for a map of name to list-of-definitions,
// e.g. IP set or port set variables.
func keyedDefinitionSetBlock[T any](ctx context.Context) schema.SetNestedBlock {
	return schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[T](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"definition": schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringType,
					ElementType: types.StringType,
					Required:    true,
				},
				names.AttrKey: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}
}

func (r *firewallPolicyV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan firewallPolicyV2ResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input networkfirewall.CreateFirewallPolicyV2Input
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("FirewallPolicyV2")))
	if resp.Diagnostics.HasError() {
		return
	}

	// An empty policy is valid but FirewallPolicyV2 is a required member.
	if input.FirewallPolicyV2 == nil {
		input.FirewallPolicyV2 = &awstypes.FirewallPolicyV2{}
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateFirewallPolicyV2(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	arn := aws.ToString(out.FirewallPolicyV2Response.FirewallPolicyV2Arn)

	outD, err := waitFirewallPolicyV2Created(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root(names.AttrARN), arn)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, outD, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *firewallPolicyV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state firewallPolicyV2ResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findFirewallPolicyV2ByARN(ctx, conn, arn)
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

	setTagsOut(ctx, out.FirewallPolicyV2Response.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *firewallPolicyV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan, state firewallPolicyV2ResourceModel
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
		arn := fwflex.StringValueFromFramework(ctx, state.ARN)
		var input networkfirewall.UpdateFirewallPolicyV2Input
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("FirewallPolicyV2")))
		if resp.Diagnostics.HasError() {
			return
		}

		if input.FirewallPolicyV2 == nil {
			input.FirewallPolicyV2 = &awstypes.FirewallPolicyV2{}
		}
		input.FirewallPolicyV2Arn = aws.String(arn)
		// An empty description clears the existing value; omitting it leaves the value unchanged.
		input.Description = plan.Description.ValueStringPointer()
		if input.Description == nil {
			input.Description = aws.String("")
		}
		input.UpdateToken = state.UpdateToken.ValueStringPointer()

		_, err := conn.UpdateFirewallPolicyV2(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		outD, err := waitFirewallPolicyV2Updated(ctx, conn, arn, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, outD, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.UpdateToken = state.UpdateToken
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *firewallPolicyV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state firewallPolicyV2ResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := networkfirewall.DeleteFirewallPolicyV2Input{
		FirewallPolicyV2Arn: aws.String(arn),
	}
	_, err := conn.DeleteFirewallPolicyV2(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitFirewallPolicyV2Deleted(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func (r *firewallPolicyV2Resource) flatten(ctx context.Context, out *networkfirewall.DescribeFirewallPolicyV2Output, data *firewallPolicyV2ResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, out.FirewallPolicyV2Response, data, fwflex.WithFieldNamePrefix("FirewallPolicyV2"))...)
	if diags.HasError() {
		return diags
	}

	policy := out.FirewallPolicyV2
	if policy == nil {
		policy = &awstypes.FirewallPolicyV2{}
	}
	diags.Append(fwflex.Flatten(ctx, policy, &data.FirewallPolicyV2)...)
	if diags.HasError() {
		return diags
	}

	data.UpdateToken = fwflex.StringToFramework(ctx, out.UpdateToken)

	return diags
}

func waitFirewallPolicyV2Created(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeFirewallPolicyV2Output, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceStatusUpdating),
		Target:                    enum.Slice(awstypes.ResourceStatusActive),
		Refresh:                   statusFirewallPolicyV2(conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeFirewallPolicyV2Output); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FirewallPolicyV2Response.StatusMessage)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFirewallPolicyV2Updated(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeFirewallPolicyV2Output, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceStatusUpdating),
		Target:                    enum.Slice(awstypes.ResourceStatusActive),
		Refresh:                   statusFirewallPolicyV2(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeFirewallPolicyV2Output); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FirewallPolicyV2Response.StatusMessage)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFirewallPolicyV2Deleted(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeFirewallPolicyV2Output, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusActive, awstypes.ResourceStatusDeleting),
		Target:  []string{},
		Refresh: statusFirewallPolicyV2(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeFirewallPolicyV2Output); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FirewallPolicyV2Response.StatusMessage)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusFirewallPolicyV2(conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findFirewallPolicyV2ByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.FirewallPolicyV2Response.FirewallPolicyV2Status), nil
	}
}

func findFirewallPolicyV2ByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeFirewallPolicyV2Output, error) {
	input := networkfirewall.DescribeFirewallPolicyV2Input{
		FirewallPolicyV2Arn: aws.String(arn),
	}

	return findFirewallPolicyV2(ctx, conn, &input)
}

func findFirewallPolicyV2(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.DescribeFirewallPolicyV2Input) (*networkfirewall.DescribeFirewallPolicyV2Output, error) {
	out, err := conn.DescribeFirewallPolicyV2(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.FirewallPolicyV2Response == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type firewallPolicyV2ResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                                  `tfsdk:"arn"`
	Description             types.String                                                  `tfsdk:"description"`
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	FirewallPolicyV2        fwtypes.ListNestedObjectValueOf[firewallPolicyV2Model]        `tfsdk:"firewall_policy"`
	Name                    types.String                                                  `tfsdk:"name"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
	UpdateToken             types.String                                                  `tfsdk:"update_token"`
}

type firewallPolicyV2Model struct {
	FirewallDefaultAction         fwtypes.StringEnum[awstypes.PolicyDefaultAction]                  `tfsdk:"firewall_default_action"`
	FirewallRuleGroupEntries      fwtypes.ListNestedObjectValueOf[firewallRuleGroupEntryModel]      `tfsdk:"firewall_rule_group_entry"`
	IPSRuleGroupEntries           fwtypes.ListNestedObjectValueOf[ipsRuleGroupEntryModel]           `tfsdk:"ips_rule_group_entry"`
	PolicyMode                    fwtypes.StringEnum[awstypes.PolicyMode]                           `tfsdk:"policy_mode"`
	PolicyVariables               fwtypes.ListNestedObjectValueOf[policyVariablesV2Model]           `tfsdk:"policy_variables"`
	Route53ResolvedIPValidation   fwtypes.ListNestedObjectValueOf[route53ResolvedIPValidationModel] `tfsdk:"route53_resolved_ip_validation"`
	StatefulEngineOptions         fwtypes.ListNestedObjectValueOf[statefulEngineOptionsV2Model]     `tfsdk:"stateful_engine_options"`
	TLSInspectionConfigurationARN fwtypes.ARN                                                       `tfsdk:"tls_inspection_configuration_arn"`
}

type firewallRuleGroupEntryModel struct {
	InlineRuleGroup fwtypes.ListNestedObjectValueOf[firewallRuleGroupV2Model] `tfsdk:"inline_rule_group"`
	Priority        types.Int32                                               `tfsdk:"priority"`
	ResourceARN     fwtypes.ARN                                               `tfsdk:"resource_arn"`
}

type firewallRuleGroupV2Model struct {
	ReferenceSets fwtypes.ListNestedObjectValueOf[referenceSetsV2Model]     `tfsdk:"reference_sets"`
	RuleVariables fwtypes.ListNestedObjectValueOf[ruleVariablesV2Model]     `tfsdk:"rule_variables"`
	RulesSource   fwtypes.ListNestedObjectValueOf[firewallRulesSourceModel] `tfsdk:"rules_source"`
}

type firewallRulesSourceModel struct {
	FirewallRules fwtypes.ListNestedObjectValueOf[firewallRuleModel] `tfsdk:"firewall_rule"`
	RulesString   types.String                                       `tfsdk:"rules_string"`
}

type firewallRuleModel struct {
	AcceptPriorHooks types.Bool                                         `tfsdk:"accept_prior_hooks"`
	Action           fwtypes.StringEnum[awstypes.FirewallRuleAction]    `tfsdk:"action"`
	ActionScope      fwtypes.StringEnum[awstypes.ActionScope]           `tfsdk:"action_scope"`
	AlertEnabled     types.Bool                                         `tfsdk:"alert_enabled"`
	BypassIPSEnabled types.Bool                                         `tfsdk:"bypass_ips_enabled"`
	DestIPs          types.String                                       `tfsdk:"dest_ips"`
	DestPorts        types.String                                       `tfsdk:"dest_ports"`
	Hook             fwtypes.StringEnum[awstypes.HookKey]               `tfsdk:"hook"`
	Protocol         fwtypes.StringEnum[awstypes.FirewallRuleProtocol]  `tfsdk:"protocol"`
	RuleMetadata     fwtypes.ListNestedObjectValueOf[ruleMetadataModel] `tfsdk:"rule_metadata"`
	RuleOptions      fwtypes.ListNestedObjectValueOf[ruleOptionV2Model] `tfsdk:"rule_option"`
	SourceIPs        types.String                                       `tfsdk:"source_ips"`
	SourcePorts      types.String                                       `tfsdk:"source_ports"`
}

type ruleMetadataModel struct {
	RuleDescription types.String `tfsdk:"rule_description"`
	RuleEnabled     types.Bool   `tfsdk:"rule_enabled"`
}

type ruleOptionV2Model struct {
	Keyword  types.String         `tfsdk:"keyword"`
	Settings fwtypes.ListOfString `tfsdk:"settings"`
}

type ipsRuleGroupEntryModel struct {
	InlineRuleGroup fwtypes.ListNestedObjectValueOf[ipsRuleGroupV2Model]       `tfsdk:"inline_rule_group"`
	Override        fwtypes.ListNestedObjectValueOf[ipsRuleGroupOverrideModel] `tfsdk:"override"`
	ResourceARN     fwtypes.ARN                                                `tfsdk:"resource_arn"`
}

type ipsRuleGroupV2Model struct {
	ReferenceSets fwtypes.ListNestedObjectValueOf[referenceSetsV2Model] `tfsdk:"reference_sets"`
	RuleVariables fwtypes.ListNestedObjectValueOf[ruleVariablesV2Model] `tfsdk:"rule_variables"`
	RulesSource   fwtypes.ListNestedObjectValueOf[ipsRulesSourceModel]  `tfsdk:"rules_source"`
}

type ipsRulesSourceModel struct {
	RulesString types.String `tfsdk:"rules_string"`
}

type ipsRuleGroupOverrideModel struct {
	Action fwtypes.StringEnum[awstypes.OverrideAction] `tfsdk:"action"`
}

type referenceSetsV2Model struct {
	IPSetReferences fwtypes.SetNestedObjectValueOf[ipSetReferenceV2Model] `tfsdk:"ip_set_reference"`
}

type ipSetReferenceV2Model struct {
	MapBlockKey  types.String `tfsdk:"key"`
	ReferenceARN fwtypes.ARN  `tfsdk:"reference_arn"`
}

type ruleVariablesV2Model struct {
	IPSets   fwtypes.SetNestedObjectValueOf[ipSetV2Model]   `tfsdk:"ip_set"`
	PortSets fwtypes.SetNestedObjectValueOf[portSetV2Model] `tfsdk:"port_set"`
}

type ipSetV2Model struct {
	Definition  fwtypes.ListOfString `tfsdk:"definition"`
	MapBlockKey types.String         `tfsdk:"key"`
}

type portSetV2Model struct {
	Definition  fwtypes.ListOfString `tfsdk:"definition"`
	MapBlockKey types.String         `tfsdk:"key"`
}

type policyVariablesV2Model struct {
	RuleVariables fwtypes.SetNestedObjectValueOf[ipSetV2Model] `tfsdk:"rule_variable"`
}

type route53ResolvedIPValidationModel struct {
	Action                              fwtypes.StringEnum[awstypes.Route53ResolvedIPValidationAction] `tfsdk:"action"`
	BypassDestinationValidationIPRanges fwtypes.ListNestedObjectValueOf[addressV2Model]                `tfsdk:"bypass_destination_validation_ip_range"`
	DNSExpiryExceptionPolicy            fwtypes.StringEnum[awstypes.DNSExpiryExceptionPolicy]          `tfsdk:"dns_expiry_exception_policy"`
	DNSFirewallRuleGroupID              types.String                                                   `tfsdk:"dns_firewall_rule_group_id"`
	SourceIPRanges                      fwtypes.ListNestedObjectValueOf[addressV2Model]                `tfsdk:"source_ip_range"`
}

type addressV2Model struct {
	AddressDefinition types.String `tfsdk:"address_definition"`
}

type statefulEngineOptionsV2Model struct {
	FlowTimeouts          fwtypes.ListNestedObjectValueOf[flowTimeoutsV2Model] `tfsdk:"flow_timeouts"`
	StreamExceptionPolicy fwtypes.StringEnum[awstypes.StreamExceptionPolicyV2] `tfsdk:"stream_exception_policy"`
}

type flowTimeoutsV2Model struct {
	TCPIdleTimeoutSeconds types.Int32 `tfsdk:"tcp_idle_timeout_seconds"`
}
