// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecs_gateway", name="Gateway")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
func newGatewayResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

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
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must contain only alphanumeric characters and hyphens"),
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
					Blocks: map[string]schema.Block{
						"host_shared_alb": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayHostSharedALBConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"current_alb_count": schema.Int32Attribute{
										Computed: true,
									},
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
										Validators: []validator.Int32{
											int32validator.AtLeast(1),
										},
									},
									names.AttrPort: schema.Int32Attribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Int32{
											int32planmodifier.RequiresReplaceIfConfigured(),
											int32planmodifier.UseStateForUnknown(),
										},
										Validators: []validator.Int32{
											int32validator.Between(1, 65535),
										},
									},
									names.AttrProtocol: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.GatewayAlbProtocol](),
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplaceIfConfigured(),
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"scheme": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.GatewayElbScheme](),
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplaceIfConfigured(),
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"web_acl_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"access_logs": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayAccessLogConfigurationModel](ctx),
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
														Blocks: map[string]schema.Block{
															"s3": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayAccessLogS3DestinationModel](ctx),
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeAtLeast(1),
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
													Computed: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(),
													},
												},
												names.AttrIPAddressType: schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.GatewayElbIpAddressType](),
													Optional:   true,
													Computed:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplaceIfConfigured(),
														stringplanmodifier.UseStateForUnknown(),
													},
												},
												names.AttrSubnets: schema.SetAttribute{
													CustomType: fwtypes.SetOfStringType,
													Required:   true,
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.RequiresReplace(),
													},
													Validators: []validator.Set{
														setvalidator.SizeAtLeast(2),
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
															names.AttrValues: schema.ListAttribute{
																CustomType: fwtypes.ListOfStringType,
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
				Delete: true,
			}),
		},
	}
}

func (r *gatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan gatewayResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	name := plan.Name.ValueString()
	var input ecs.CreateGatewayInput
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := createGatewayWithRetry(ctx, conn, &input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(r.Meta().Partition(ctx), err) {
		input.Tags = nil
		output, err = createGatewayWithRetry(ctx, conn, &input)
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating ECS Gateway (%s)", name), err.Error())
		return
	}

	arn := aws.ToString(output.GatewayArn)

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		if err := createTags(ctx, conn, arn, tags); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("setting ECS Gateway (%s) tags", arn), err.Error())
			return
		}
	}

	// Save the ARN so that Terraform tracks the resource if the waiter times out.
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrARN), arn)...)
	if response.Diagnostics.HasError() {
		return
	}

	gateway, err := waitGatewayActive(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Gateway (%s) create", arn), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, gateway, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *gatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state gatewayResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	arn := state.GatewayARN.ValueString()
	gateway, err := findGatewayByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Gateway (%s)", arn), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, gateway, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *gatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state gatewayResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	arn := plan.GatewayARN.ValueString()
	if diff.HasChanges() {
		var input ecs.UpdateGatewayInput
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.Gateway = aws.String(arn)

		if _, err := updateGatewayWithRetry(ctx, conn, &input); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating ECS Gateway (%s)", arn), err.Error())
			return
		}

		gateway, err := waitGatewayActive(ctx, conn, arn, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Gateway (%s) update", arn), err.Error())
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, gateway, &plan)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *gatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state gatewayResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECSClient(ctx)

	arn := state.GatewayARN.ValueString()
	input := ecs.DeleteGatewayInput{
		Gateway: aws.String(arn),
	}
	_, err := conn.DeleteGateway(ctx, &input)

	if errs.IsA[*awstypes.GatewayNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting ECS Gateway (%s)", arn), err.Error())
		return
	}

	if _, err := waitGatewayDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for ECS Gateway (%s) delete", arn), err.Error())
		return
	}
}

func createGatewayWithRetry(ctx context.Context, conn *ecs.Client, input *ecs.CreateGatewayInput) (*ecs.CreateGatewayOutput, error) {
	// The infrastructure role may not be assumable straight after it is created.
	return tfresource.RetryWhenIsAErrorMessageContains[*ecs.CreateGatewayOutput, *awstypes.AccessDeniedException](ctx, propagationTimeout,
		func(ctx context.Context) (*ecs.CreateGatewayOutput, error) {
			return conn.CreateGateway(ctx, input)
		},
		"Could not assume the infrastructure role",
	)
}

func updateGatewayWithRetry(ctx context.Context, conn *ecs.Client, input *ecs.UpdateGatewayInput) (*ecs.UpdateGatewayOutput, error) {
	return tfresource.RetryWhenIsAErrorMessageContains[*ecs.UpdateGatewayOutput, *awstypes.AccessDeniedException](ctx, propagationTimeout,
		func(ctx context.Context) (*ecs.UpdateGatewayOutput, error) {
			return conn.UpdateGateway(ctx, input)
		},
		"Could not assume the infrastructure role",
	)
}

func findGatewayByARN(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.Gateway, error) {
	input := ecs.DescribeGatewayInput{
		Gateway: aws.String(arn),
	}
	output, err := conn.DescribeGateway(ctx, &input)

	if errs.IsA[*awstypes.GatewayNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Gateway == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if status := output.Gateway.Status; status == awstypes.GatewayStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
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
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitGatewayActive(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) (*awstypes.Gateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GatewayStatusUpdateInProgress),
		Target:  enum.Slice(awstypes.GatewayStatusActive, awstypes.GatewayStatusUpdateComplete),
		Refresh: statusGateway(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Gateway); ok {
		retry.SetLastError(err, gatewayError(output))
		return output, err
	}

	return nil, err
}

func waitGatewayDeleted(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) (*awstypes.Gateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GatewayStatusActive, awstypes.GatewayStatusUpdateInProgress, awstypes.GatewayStatusUpdateComplete, awstypes.GatewayStatusUpdateFailed),
		Target:  []string{},
		Refresh: statusGateway(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Gateway); ok {
		retry.SetLastError(err, gatewayError(output))
		return output, err
	}

	return nil, err
}

func gatewayError(apiObject *awstypes.Gateway) error {
	if reason := aws.ToString(apiObject.StatusReason); reason != "" {
		return fmt.Errorf("%s", reason)
	}

	return nil
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
	HostSharedALB fwtypes.ListNestedObjectValueOf[gatewayHostSharedALBConfigurationModel] `tfsdk:"host_shared_alb"`
}

var (
	_ fwflex.TypedExpander = gatewayConfigurationModel{}
	_ fwflex.Flattener     = &gatewayConfigurationModel{}
)

func (m gatewayConfigurationModel) ExpandTo(ctx context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[awstypes.GatewayConfiguration]():
		if !m.HostSharedALB.IsNull() {
			var r awstypes.GatewayConfigurationMemberHostSharedAlb
			diags.Append(fwflex.Expand(ctx, m.HostSharedALB, &r.Value)...)
			if diags.HasError() {
				return nil, diags
			}

			return &r, diags
		}

	case reflect.TypeFor[awstypes.UpdateGatewayConfiguration]():
		if !m.HostSharedALB.IsNull() {
			var r awstypes.UpdateGatewayConfigurationMemberHostSharedAlb
			diags.Append(fwflex.Expand(ctx, m.HostSharedALB, &r.Value)...)
			if diags.HasError() {
				return nil, diags
			}

			return &r, diags
		}
	}

	return nil, diags
}

func (m *gatewayConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.DescribeGatewayConfigurationMemberHostSharedAlb:
		var model gatewayHostSharedALBConfigurationModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &model)...)
		if diags.HasError() {
			return diags
		}

		m.HostSharedALB = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	}

	return diags
}

type gatewayHostSharedALBConfigurationModel struct {
	AccessLogs           fwtypes.ListNestedObjectValueOf[gatewayAccessLogConfigurationModel] `tfsdk:"access_logs"`
	CurrentALBCount      types.Int32                                                         `tfsdk:"current_alb_count"`
	HTTPDesync           fwtypes.StringEnum[awstypes.GatewayAlbDesyncMitigationMode]         `tfsdk:"http_desync"`
	MaxALBCount          types.Int32                                                         `tfsdk:"max_alb_count"`
	NetworkConfiguration fwtypes.ListNestedObjectValueOf[gatewayNetworkConfigurationModel]   `tfsdk:"network_configuration"`
	Port                 types.Int32                                                         `tfsdk:"port"`
	Protocol             fwtypes.StringEnum[awstypes.GatewayAlbProtocol]                     `tfsdk:"protocol"`
	Scheme               fwtypes.StringEnum[awstypes.GatewayElbScheme]                       `tfsdk:"scheme"`
	TestTrafficRules     fwtypes.ListNestedObjectValueOf[gatewayTestTrafficRulesModel]       `tfsdk:"test_traffic_rules"`
	WebACLARN            fwtypes.ARN                                                         `tfsdk:"web_acl_arn"`
}

type gatewayAccessLogConfigurationModel struct {
	Destination fwtypes.ListNestedObjectValueOf[gatewayAccessLogDestinationModel] `tfsdk:"destination"`
}

type gatewayAccessLogDestinationModel struct {
	S3 fwtypes.ListNestedObjectValueOf[gatewayAccessLogS3DestinationModel] `tfsdk:"s3"`
}

var (
	_ fwflex.Expander  = gatewayAccessLogDestinationModel{}
	_ fwflex.Flattener = &gatewayAccessLogDestinationModel{}
)

func (m gatewayAccessLogDestinationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	if !m.S3.IsNull() {
		var r awstypes.GatewayElbAccessLogDestinationMemberS3
		diags.Append(fwflex.Expand(ctx, m.S3, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

func (m *gatewayAccessLogDestinationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.GatewayElbAccessLogDestinationMemberS3:
		var model gatewayAccessLogS3DestinationModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &model)...)
		if diags.HasError() {
			return diags
		}

		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	}

	return diags
}

type gatewayAccessLogS3DestinationModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Prefix types.String `tfsdk:"prefix"`
}

type gatewayNetworkConfigurationModel struct {
	InboundSecurityGroup types.String                                         `tfsdk:"inbound_security_group"`
	IPAddressType        fwtypes.StringEnum[awstypes.GatewayElbIpAddressType] `tfsdk:"ip_address_type"`
	Subnets              fwtypes.SetOfString                                  `tfsdk:"subnets"`
}

type gatewayTestTrafficRulesModel struct {
	Header fwtypes.ListNestedObjectValueOf[gatewayTestTrafficHeaderModel] `tfsdk:"header"`
}

type gatewayTestTrafficHeaderModel struct {
	Name   types.String         `tfsdk:"name"`
	Values fwtypes.ListOfString `tfsdk:"values"`
}
