// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_apigatewayv2_routing_rule", name="Routing Rule")
// @Testing(importStateIdAttribute="arn")
func newResourceRoutingRule(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceRoutingRule{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameRoutingRule = "Routing Rule"
)

var (
	flexOpt = flex.WithFieldNamePrefix("RoutingRule")
)

type resourceRoutingRule struct {
	framework.ResourceWithModel[resourceRoutingRuleModel]
	framework.WithTimeouts
}

func (r *resourceRoutingRule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 512),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrPriority: schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1000000),
				},
			},
			"domain_name_id": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"conditions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[conditionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(3),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("match_headers"),
							path.MatchRelative().AtName("match_base_paths"),
						),
					},
					Blocks: map[string]schema.Block{
						"match_headers": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[matchHeadersModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"any_of": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[anyOfModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrHeader: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(
															regexache.MustCompile("^[a-zA-Z0-9*?!#$%&'.^_`|~-]{1,39}$"),
															"must be less than 40 characters and the only allowed characters are a-z, A-Z, 0-9, and the following special characters: *?!#$%&'.^_`|~-",
														),
													},
												},
												"value_glob": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(
															regexache.MustCompile("^[a-zA-Z0-9*?!#$%&'.^_`|~-]{1,127}$"),
															"must be less than 128 characters and the only allowed characters are a-z, A-Z, 0-9, and the following special characters: *?!#$%&'.^_`|~-",
														),
													},
												},
											},
										},
									},
								},
							},
						},
						"match_base_paths": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[matchBasePathsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"any_of": schema.ListAttribute{
										ElementType: types.StringType,
										Required:    true,
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrActions: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[actionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"invoke_api": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[invokeAPIModel](ctx),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceRoutingRule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().APIGatewayV2Client(ctx)

	var plan resourceRoutingRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input apigatewayv2.CreateRoutingRuleInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flexOpt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateRoutingRule(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGatewayV2, create.ErrActionCreating, ResNameRoutingRule, plan.DomainName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGatewayV2, create.ErrActionCreating, ResNameRoutingRule, plan.DomainName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan, flexOpt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceRoutingRule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().APIGatewayV2Client(ctx)

	var state resourceRoutingRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRoutingRuleByARN(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGatewayV2, create.ErrActionReading, ResNameRoutingRule, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flexOpt)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceRoutingRule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().APIGatewayV2Client(ctx)

	var plan, state resourceRoutingRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input apigatewayv2.PutRoutingRuleInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flexOpt)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.PutRoutingRule(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.APIGatewayV2, create.ErrActionUpdating, ResNameRoutingRule, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.APIGatewayV2, create.ErrActionUpdating, ResNameRoutingRule, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan, flexOpt)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceRoutingRule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().APIGatewayV2Client(ctx)

	var state resourceRoutingRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := apigatewayv2.DeleteRoutingRuleInput{
		RoutingRuleId: state.ID.ValueStringPointer(),
		DomainName:    state.DomainName.ValueStringPointer(),
	}

	_, err := conn.DeleteRoutingRule(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGatewayV2, create.ErrActionDeleting, ResNameRoutingRule, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceRoutingRule) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
	domainName, _, err := parseRoutingRuleARN(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGatewayV2, create.ErrActionImporting, ResNameRoutingRule, req.ID, err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrDomainName), domainName)...)
}

func findRoutingRuleByARN(ctx context.Context, conn *apigatewayv2.Client, id string) (*apigatewayv2.GetRoutingRuleOutput, error) {
	domainName, ruleID, err := parseRoutingRuleARN(id)
	if err != nil {
		return nil, err
	}

	input := apigatewayv2.GetRoutingRuleInput{
		DomainName:    aws.String(domainName),
		RoutingRuleId: aws.String(ruleID),
	}
	out, err := conn.GetRoutingRule(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, tfresource.NewEmptyResultError(&input)
		}
		return nil, err
	}

	return out, nil
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

type resourceRoutingRuleModel struct {
	framework.WithRegionModel
	ARN          types.String                                     `tfsdk:"arn"`
	Actions      fwtypes.ListNestedObjectValueOf[actionsModel]    `tfsdk:"actions"`
	Conditions   fwtypes.ListNestedObjectValueOf[conditionsModel] `tfsdk:"conditions"`
	DomainName   types.String                                     `tfsdk:"domain_name"`
	DomainNameID types.String                                     `tfsdk:"domain_name_id"`
	ID           types.String                                     `tfsdk:"id"`
	Priority     types.Int64                                      `tfsdk:"priority"`
	Timeouts     timeouts.Value                                   `tfsdk:"timeouts"`
}

type conditionsModel struct {
	MatchBasePaths fwtypes.ListNestedObjectValueOf[matchBasePathsModel] `tfsdk:"match_base_paths"`
	MatchHeaders   fwtypes.ListNestedObjectValueOf[matchHeadersModel]   `tfsdk:"match_headers"`
}

type matchHeadersModel struct {
	AnyOf fwtypes.ListNestedObjectValueOf[anyOfModel] `tfsdk:"any_of"`
}

type anyOfModel struct {
	Header    types.String `tfsdk:"header"`
	ValueGlob types.String `tfsdk:"value_glob"`
}

type matchBasePathsModel struct {
	AnyOf fwtypes.ListValueOf[types.String] `tfsdk:"any_of"`
}

type actionsModel struct {
	InvokeAPI fwtypes.ListNestedObjectValueOf[invokeAPIModel] `tfsdk:"invoke_api"`
}

type invokeAPIModel struct {
	ApiId         types.String `tfsdk:"api_id"`
	Stage         types.String `tfsdk:"stage"`
	StripBasePath types.Bool   `tfsdk:"strip_base_path"`
}
