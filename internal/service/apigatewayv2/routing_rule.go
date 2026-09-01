// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_apigatewayv2_routing_rule", name="Routing Rule")
// @Testing(importStateIdAttribute="arn")
func newRoutingRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &routingRuleResource{}
	return r, nil
}

type routingRuleResource struct {
	framework.ResourceWithModel[routingRuleResourceModel]
}

func (r *routingRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPriority: schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.Between(1, 1000000),
				},
			},
			"routing_rule_arn": framework.ARNAttributeComputedOnly(),
			"routing_rule_id":  framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			names.AttrAction: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[routingRuleActionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"invoke_api": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[routingRuleActionInvokeAPIModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"api_id": schema.StringAttribute{
										Required: true,
									},
									names.AttrStage: schema.StringAttribute{
										Required: true,
									},
									"strip_base_path": schema.BoolAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrCondition: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[routingRuleConditionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("match_base_paths"),
							path.MatchRelative().AtName("match_headers"),
						),
					},
					Blocks: map[string]schema.Block{
						"match_base_paths": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[routingRuleMatchBasePathsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"any_of": schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
										ElementType: types.StringType,
										Required:    true,
									},
								},
							},
						},
						"match_headers": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[routingRuleMatchHeadersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"any_of": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[routingRuleMatchHeaderValueModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrHeader: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthAtMost(40),
													},
												},
												"value_glob": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthAtMost(128),
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
		},
	}
}

func (r *routingRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routingRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	var input apigatewayv2.CreateRoutingRuleInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateRoutingRule(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("creating API Gateway v2 Routing Rule", err.Error())
		return
	}

	// Set values for unknowns.
	plan.RoutingRuleARN = fwflex.StringToFramework(ctx, out.RoutingRuleArn)
	plan.RoutingRuleID = fwflex.StringToFramework(ctx, out.RoutingRuleId)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *routingRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routingRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	domainName, ruleID := fwflex.StringValueFromFramework(ctx, state.DomainName), fwflex.StringValueFromFramework(ctx, state.RoutingRuleID)
	out, err := findRoutingRuleByTwoPartKey(ctx, conn, domainName, ruleID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading API Gateway v2 Routing Rule (%s)", ruleID), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routingRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state routingRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		ruleID := fwflex.StringValueFromFramework(ctx, plan.RoutingRuleID)
		var input apigatewayv2.PutRoutingRuleInput
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.PutRoutingRule(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating API Gateway v2 Routing Rule (%s)", ruleID), err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routingRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routingRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	domainName, ruleID := fwflex.StringValueFromFramework(ctx, state.DomainName), fwflex.StringValueFromFramework(ctx, state.RoutingRuleID)
	input := apigatewayv2.DeleteRoutingRuleInput{
		DomainName:    aws.String(domainName),
		RoutingRuleId: aws.String(ruleID),
	}
	_, err := conn.DeleteRoutingRule(ctx, &input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting API Gateway v2 Routing Rule (%s)", ruleID), err.Error())
		return
	}
}

func (r *routingRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	domainName, ruleID, err := parseRoutingRuleARN(req.ID)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrDomainName), domainName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("routing_rule_id"), ruleID)...)
}

func findRoutingRuleByTwoPartKey(ctx context.Context, conn *apigatewayv2.Client, domainName, ruleID string) (*apigatewayv2.GetRoutingRuleOutput, error) {
	input := apigatewayv2.GetRoutingRuleInput{
		DomainName:    aws.String(domainName),
		RoutingRuleId: aws.String(ruleID),
	}

	return findRoutingRule(ctx, conn, &input)
}

func findRoutingRule(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetRoutingRuleInput) (*apigatewayv2.GetRoutingRuleOutput, error) {
	output, err := conn.GetRoutingRule(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func parseRoutingRuleARN(v string) (string, string, error) {
	// arn:${Partition}:apigateway:${Region}:${Account}:/domainnames/${DomainName}/routingrules/${RoutingRuleId}
	arn, err := arn.Parse(v)
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(strings.TrimPrefix(arn.Resource, "/domainnames/"), "/routingrules/")
	if len(parts) != 2 {
		return "", "", errors.New("invalid routing rule ARN")
	}

	return parts[0], parts[1], nil
}

type routingRuleResourceModel struct {
	framework.WithRegionModel
	Actions        fwtypes.ListNestedObjectValueOf[routingRuleActionModel]    `tfsdk:"action"`
	Conditions     fwtypes.ListNestedObjectValueOf[routingRuleConditionModel] `tfsdk:"condition"`
	DomainName     types.String                                               `tfsdk:"domain_name"`
	Priority       types.Int32                                                `tfsdk:"priority"`
	RoutingRuleARN types.String                                               `tfsdk:"routing_rule_arn"`
	RoutingRuleID  types.String                                               `tfsdk:"routing_rule_id"`
}

type routingRuleActionModel struct {
	InvokeAPI fwtypes.ListNestedObjectValueOf[routingRuleActionInvokeAPIModel] `tfsdk:"invoke_api"`
}

type routingRuleActionInvokeAPIModel struct {
	ApiID         types.String `tfsdk:"api_id"`
	Stage         types.String `tfsdk:"stage"`
	StripBasePath types.Bool   `tfsdk:"strip_base_path"`
}

type routingRuleConditionModel struct {
	MatchBasePaths fwtypes.ListNestedObjectValueOf[routingRuleMatchBasePathsModel] `tfsdk:"match_base_paths"`
	MatchHeaders   fwtypes.ListNestedObjectValueOf[routingRuleMatchHeadersModel]   `tfsdk:"match_headers"`
}

type routingRuleMatchBasePathsModel struct {
	AnyOf fwtypes.SetOfString `tfsdk:"any_of"`
}

type routingRuleMatchHeadersModel struct {
	AnyOf fwtypes.ListNestedObjectValueOf[routingRuleMatchHeaderValueModel] `tfsdk:"any_of"`
}

type routingRuleMatchHeaderValueModel struct {
	Header    types.String `tfsdk:"header"`
	ValueGLOB types.String `tfsdk:"value_glob"`
}
