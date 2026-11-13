// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ecs

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecs_gateway", name="Gateway")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckGateways")
func newGatewayResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)

	return r, nil
}

type gatewayResource struct {
	framework.ResourceWithModel[gatewayResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *gatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccessType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDeletionProtection: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"infrastructure_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			names.AttrPropagateTags: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PropagateGatewayTags](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("host_shared_alb"),
						),
					},
					Blocks: map[string]schema.Block{
						"host_shared_alb": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayHostSharedALBModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"http_desync": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.GatewayAlbDesyncMitigationMode](),
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"max_alb_count": schema.Int32Attribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Int32{
											int32planmodifier.UseStateForUnknown(),
										},
									},
									"scheme": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.GatewayElbScheme](),
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
											stringplanmodifier.RequiresReplace(),
										},
									},
									"web_acl_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"access_logs": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayAccessLogsModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												names.AttrDestination: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayAccessLogDestinationModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Validators: []validator.Object{
															tfobjectvalidator.ExactlyOneOfChildren(
																path.MatchRelative().AtName("s3"),
															),
														},
														Blocks: map[string]schema.Block{
															"s3": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayS3AccessLogDestinationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		names.AttrBucket: schema.StringAttribute{
																			Required: true,
																		},
																		names.AttrPrefix: schema.StringAttribute{
																			Optional: true,
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
										CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayNetworkConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inbound_security_group": schema.StringAttribute{
													Optional: true,
												},
												names.AttrIPAddressType: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.GatewayElbIpAddressType](),
													Optional:   true,
													Computed:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(),
														stringplanmodifier.RequiresReplace(),
													},
												},
												names.AttrSubnets: schema.SetAttribute{
													CustomType: fwtypes.SetOfStringType,
													Required:   true,
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.RequiresReplace(),
													},
												},
											},
										},
									},
									"test_traffic_rules": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayTestTrafficRulesModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												names.AttrHeader: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayTestTrafficHeaderModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrName: schema.StringAttribute{
																Required: true,
															},
															names.AttrValues: schema.SetAttribute{
																CustomType: fwtypes.SetOfStringType,
																Required:   true,
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
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *gatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan gatewayResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input ecs.CreateGatewayInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := tfresource.RetryWhen(ctx, propagationTimeout, func(ctx context.Context) (*ecs.CreateGatewayOutput, error) {
		return conn.CreateGateway(ctx, &input)
	}, gatewayRoleRetryable)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrARN), plan.GatewayARN))
	if response.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, plan.GatewayARN)
	gateway, err := waitGatewayStable(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, gateway, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *gatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data gatewayResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.GatewayARN)
	gateway, err := findGatewayByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, gateway, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayResource) flatten(ctx context.Context, gateway *awstypes.Gateway, data *gatewayResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, gateway, data)...)
	if diags.HasError() {
		return diags
	}

	setTagsOut(ctx, gateway.Tags)

	return diags
}

func (r *gatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state gatewayResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, state.GatewayARN)
		var input ecs.UpdateGatewayInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}

		input.Gateway = aws.String(arn)

		_, err := tfresource.RetryWhen(ctx, propagationTimeout, func(ctx context.Context) (*ecs.UpdateGatewayOutput, error) {
			return conn.UpdateGateway(ctx, &input)
		}, gatewayRoleRetryable)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
			return
		}

		gateway, err := waitGatewayStable(ctx, conn, arn, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, gateway, &plan))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *gatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data gatewayResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.GatewayARN)
	input := ecs.DeleteGatewayInput{
		Gateway: aws.String(arn),
	}
	_, err := conn.DeleteGateway(ctx, &input)
	if isGatewayNotFoundError(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
	}
}

func gatewayRoleRetryable(err error) (bool, error) {
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Could not assume the infrastructure role") {
		return true, err
	}

	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "AWS was not able to validate the provided access credentials") {
		return true, err
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "could not be assumed") {
		return true, err
	}

	if errs.IsAErrorMessageContains[*awstypes.ClientException](err, "/ECSGateway is not authorized to perform") {
		return true, err
	}

	return false, err
}

func isGatewayNotFoundError(err error) bool {
	return errs.IsA[*awstypes.GatewayNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ClientException](err, "No gateway found")
}

func findGatewayByARN(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.Gateway, error) {
	input := ecs.DescribeGatewayInput{
		Gateway: aws.String(arn),
	}

	output, err := conn.DescribeGateway(ctx, &input)
	if isGatewayNotFoundError(err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if output == nil || output.Gateway == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	if output.Gateway.Status == awstypes.GatewayStatusDeleted {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: string(output.Gateway.Status),
		})
	}

	return output.Gateway, nil
}

func statusGateway(conn *ecs.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGatewayByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Status), nil
	}
}

func waitGatewayStable(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) (*awstypes.Gateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GatewayStatusUpdateInProgress),
		Target:                    enum.Slice(awstypes.GatewayStatusActive, awstypes.GatewayStatusUpdateComplete),
		Refresh:                   statusGateway(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Gateway); ok {
		retry.SetLastError(err, gatewayStatusError(output))
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func gatewayStatusError(gateway *awstypes.Gateway) error {
	var causes []error

	if v := aws.ToString(gateway.StatusReason); v != "" {
		causes = append(causes, errors.New(v))
	}

	for _, failure := range gateway.Failures {
		for _, e := range failure.Errors {
			causes = append(causes, fmt.Errorf("%s: %s %s: %s", aws.ToString(failure.Arn), aws.ToString(e.Operation), aws.ToString(e.Code), aws.ToString(e.Message)))
		}
	}

	return errors.Join(causes...)
}

type gatewayResourceModel struct {
	framework.WithRegionModel
	AccessType            fwtypes.StringEnum[awstypes.AccessType]                    `tfsdk:"access_type"`
	Configuration         fwtypes.ListNestedObjectValueOf[gatewayConfigurationModel] `tfsdk:"configuration"`
	DeletionProtection    types.Bool                                                 `tfsdk:"deletion_protection"`
	GatewayARN            types.String                                               `tfsdk:"arn"`
	InfrastructureRoleARN fwtypes.ARN                                                `tfsdk:"infrastructure_role_arn"`
	Name                  types.String                                               `tfsdk:"name"`
	PropagateTags         fwtypes.StringEnum[awstypes.PropagateGatewayTags]          `tfsdk:"propagate_tags"`
	Tags                  tftags.Map                                                 `tfsdk:"tags"`
	TagsAll               tftags.Map                                                 `tfsdk:"tags_all"`
	Timeouts              timeouts.Value                                             `tfsdk:"timeouts"`
}

type gatewayConfigurationModel struct {
	HostSharedALB fwtypes.ListNestedObjectValueOf[gatewayHostSharedALBModel] `tfsdk:"host_shared_alb"`
}

type gatewayHostSharedALBModel struct {
	AccessLogs           fwtypes.ListNestedObjectValueOf[gatewayAccessLogsModel]           `tfsdk:"access_logs"`
	HTTPDesync           fwtypes.StringEnum[awstypes.GatewayAlbDesyncMitigationMode]       `tfsdk:"http_desync"`
	MaxALBCount          types.Int32                                                       `tfsdk:"max_alb_count"`
	NetworkConfiguration fwtypes.ListNestedObjectValueOf[gatewayNetworkConfigurationModel] `tfsdk:"network_configuration"`
	Scheme               fwtypes.StringEnum[awstypes.GatewayElbScheme]                     `tfsdk:"scheme"`
	TestTrafficRules     fwtypes.ListNestedObjectValueOf[gatewayTestTrafficRulesModel]     `tfsdk:"test_traffic_rules"`
	WebACLARN            fwtypes.ARN                                                       `tfsdk:"web_acl_arn"`
}

type gatewayNetworkConfigurationModel struct {
	InboundSecurityGroup types.String                                         `tfsdk:"inbound_security_group"`
	IPAddressType        fwtypes.StringEnum[awstypes.GatewayElbIpAddressType] `tfsdk:"ip_address_type"`
	Subnets              fwtypes.SetOfString                                  `tfsdk:"subnets"`
}

type gatewayAccessLogsModel struct {
	Destination fwtypes.ListNestedObjectValueOf[gatewayAccessLogDestinationModel] `tfsdk:"destination"`
}

type gatewayAccessLogDestinationModel struct {
	S3 fwtypes.ListNestedObjectValueOf[gatewayS3AccessLogDestinationModel] `tfsdk:"s3"`
}

type gatewayS3AccessLogDestinationModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Prefix types.String `tfsdk:"prefix"`
}

type gatewayTestTrafficRulesModel struct {
	Header fwtypes.ListNestedObjectValueOf[gatewayTestTrafficHeaderModel] `tfsdk:"header"`
}

type gatewayTestTrafficHeaderModel struct {
	Name   types.String        `tfsdk:"name"`
	Values fwtypes.SetOfString `tfsdk:"values"`
}

var (
	_ fwflex.TypedExpander = gatewayConfigurationModel{}
	_ fwflex.Flattener     = &gatewayConfigurationModel{}
	_ fwflex.Expander      = gatewayAccessLogDestinationModel{}
	_ fwflex.Flattener     = &gatewayAccessLogDestinationModel{}
)

func (m gatewayConfigurationModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	alb, d := m.HostSharedALB.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	if alb == nil {
		diags.AddError("Invalid Union", "gateway configuration expand: no member is set")
		return nil, diags
	}

	switch targetType {
	case reflect.TypeFor[awstypes.GatewayConfiguration]():
		var r awstypes.GatewayConfigurationMemberHostSharedAlb
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alb, &r.Value))
		return &r, diags

	case reflect.TypeFor[awstypes.UpdateGatewayConfiguration]():
		var r awstypes.UpdateGatewayConfigurationMemberHostSharedAlb
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, alb, &r.Value))
		return &r, diags
	}

	diags.AddError("Unsupported Type", fmt.Sprintf("gateway configuration expand: %s", targetType))

	return nil, diags
}

func (m *gatewayConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.DescribeGatewayConfigurationMemberHostSharedAlb:
		var model gatewayHostSharedALBModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.HostSharedALB = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.UnknownUnionMember:
		fwflex.HandleFlattenUnknownUnionMember(ctx, t.Tag, &diags)

	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("gateway configuration flatten: %T", v))
	}

	return diags
}

func (m gatewayAccessLogDestinationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	s3, d := m.S3.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return nil, diags
	}
	if s3 == nil {
		diags.AddError("Invalid Union", "gateway access log destination expand: no member is set")
		return nil, diags
	}

	var r awstypes.GatewayElbAccessLogDestinationMemberS3
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, s3, &r.Value))

	return &r, diags
}

func (m *gatewayAccessLogDestinationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.GatewayElbAccessLogDestinationMemberS3:
		var model gatewayS3AccessLogDestinationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.UnknownUnionMember:
		fwflex.HandleFlattenUnknownUnionMember(ctx, t.Tag, &diags)

	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("gateway access log destination flatten: %T", v))
	}

	return diags
}
