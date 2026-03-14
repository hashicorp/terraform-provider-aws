// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package lambda

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkResource("aws_lambda_capacity_provider", name="Capacity Provider")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambda/types;types.CapacityProvider")
// @Testing(importStateIdAttribute="name")
// @Testing(importStateIdFunc=testAccCheckCapacityProviderImportStateID)
// @Testing(preIdentityVersion="v6.25.0")
// @Testing(preCheck="testAccCapacityProviderPreCheck")
func newResourceCapacityProvider(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCapacityProvider{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	capacityProviderNamePrefix = "CapacityProvider"
	ResNameCapacityProvider    = "Capacity Provider"
)

type resourceCapacityProvider struct {
	framework.ResourceWithModel[resourceCapacityProviderModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *resourceCapacityProvider) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:                      framework.ARNAttributeComputedOnly(),
			"capacity_provider_scaling_config": framework.ResourceOptionalComputedListOfObjectsAttribute[capacityProviderScalingConfigModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
			"instance_requirements":            framework.ResourceOptionalComputedListOfObjectsAttribute[instanceRequirementsModel](ctx, 1, nil, listplanmodifier.RequiresReplaceIfConfigured(), listplanmodifier.UseStateForUnknown()),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"permissions_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[permissionConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"capacity_provider_operator_role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			names.AttrVPCConfig: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSubnetIDs: schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
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

func (r *resourceCapacityProvider) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan resourceCapacityProviderModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	var input lambda.CreateCapacityProviderInput
	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix(capacityProviderNamePrefix)))
	if response.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := tfresource.RetryWhenIsA[*lambda.CreateCapacityProviderOutput, *awstypes.InvalidParameterValueException](ctx, iamPropagationTimeout, func(ctx context.Context) (*lambda.CreateCapacityProviderOutput, error) {
		return conn.CreateCapacityProvider(ctx, &input)
	})

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	if out == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := waitCapacityProviderActive(ctx, conn, plan.Name.ValueString(), createTimeout)
	if err != nil {
		smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrName), plan.Name.ValueString()))
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Flatten(ctx, output, &plan, flex.WithFieldNamePrefix(capacityProviderNamePrefix)))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *resourceCapacityProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state resourceCapacityProviderModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findCapacityProviderByName(ctx, conn, state.Name.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output, &state, flex.WithFieldNamePrefix(capacityProviderNamePrefix)))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceCapacityProvider) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan, state resourceCapacityProviderModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input lambda.UpdateCapacityProviderInput
		smerr.AddEnrich(ctx, &response.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix(capacityProviderNamePrefix)))
		if response.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdateCapacityProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if output == nil {
			smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, flex.Flatten(ctx, output, &plan, flex.WithFieldNamePrefix(capacityProviderNamePrefix)))
		if response.Diagnostics.HasError() {
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitCapacityProviderActive(ctx, conn, plan.Name.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *resourceCapacityProvider) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state resourceCapacityProviderModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	input := lambda.DeleteCapacityProviderInput{
		CapacityProviderName: state.Name.ValueStringPointer(),
	}

	deleteCapacityProvider := func(startTime time.Time) (time.Duration, error) {
		functionDeletePropagationTimeout := time.Minute * 5

		_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ResourceConflictException](ctx, functionDeletePropagationTimeout, func(ctx context.Context) (any, error) {
			return conn.DeleteCapacityProvider(ctx, &input)
		}, "To delete this capacity provider, first remove its association")

		return time.Since(startTime), err
	}

	timeElapsed, err := deleteCapacityProvider(time.Now())
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitCapacityProviderDeleted(ctx, conn, state.Name.ValueString(), deleteTimeout-timeElapsed)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *resourceCapacityProvider) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data resourceCapacityProviderModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if !data.InstanceRequirements.IsNull() {
		ir, d := data.InstanceRequirements.ToPtr(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}

		if ir != nil {
			if ir.Architectures.IsNull() {
				smerr.AddError(ctx, &response.Diagnostics, errors.New("architectures must be specified when instance_requirements is configured"), smerr.ID)
				return
			}
		}
	}

	if !data.CapacityProviderScalingConfig.IsNull() {
		sc, d := data.CapacityProviderScalingConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}

		if sc != nil {
			if sc.ScalingMode.IsNull() {
				smerr.AddError(ctx, &response.Diagnostics, errors.New("scaling_mode must be specified when capacity_provider_scaling_config is configured"), smerr.ID)
				return
			}
		}
	}
}

func waitCapacityProviderActive(ctx context.Context, conn *lambda.Client, name string, timeout time.Duration) (*awstypes.CapacityProvider, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.CapacityProviderStatePending),
		Target:                    enum.Slice(awstypes.CapacityProviderStateActive),
		Refresh:                   statusCapacityProvider(conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.CapacityProvider); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitCapacityProviderDeleted(ctx context.Context, conn *lambda.Client, name string, timeout time.Duration) (*awstypes.CapacityProvider, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.CapacityProviderStatePending, awstypes.CapacityProviderStateDeleting),
		Target:                    []string{},
		Refresh:                   statusCapacityProvider(conn, name),
		Timeout:                   timeout,
		Delay:                     time.Second * 5,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.CapacityProvider); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusCapacityProvider(conn *lambda.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findCapacityProviderByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findCapacityProviderByName(ctx context.Context, conn *lambda.Client, name string) (*awstypes.CapacityProvider, error) {
	input := lambda.GetCapacityProviderInput{
		CapacityProviderName: aws.String(name),
	}

	out, err := conn.GetCapacityProvider(ctx, &input)
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

	return out.CapacityProvider, nil
}

type resourceCapacityProviderModel struct {
	framework.WithRegionModel
	ARN                           types.String                                                        `tfsdk:"arn"`
	Name                          types.String                                                        `tfsdk:"name"`
	KMSKeyARN                     fwtypes.ARN                                                         `tfsdk:"kms_key_arn"`
	PermissionsConfig             fwtypes.ListNestedObjectValueOf[permissionConfigModel]              `tfsdk:"permissions_config"`
	VpcConfig                     fwtypes.ListNestedObjectValueOf[vpcConfigModel]                     `tfsdk:"vpc_config"`
	CapacityProviderScalingConfig fwtypes.ListNestedObjectValueOf[capacityProviderScalingConfigModel] `tfsdk:"capacity_provider_scaling_config"`
	InstanceRequirements          fwtypes.ListNestedObjectValueOf[instanceRequirementsModel]          `tfsdk:"instance_requirements"`
	Tags                          tftags.Map                                                          `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                          `tfsdk:"tags_all"`
	Timeouts                      timeouts.Value                                                      `tfsdk:"timeouts"`
}

type vpcConfigModel struct {
	SecurityGroupIDs fwtypes.SetOfString `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetOfString `tfsdk:"subnet_ids"`
}

type permissionConfigModel struct {
	CapacityProviderOperatorRoleARN fwtypes.ARN `tfsdk:"capacity_provider_operator_role_arn"`
}

type instanceRequirementsModel struct {
	Architectures         fwtypes.ListOfStringEnum[awstypes.Architecture] `tfsdk:"architectures"`
	AllowedInstanceTypes  fwtypes.ListOfString                            `tfsdk:"allowed_instance_types"`
	ExcludedInstanceTypes fwtypes.ListOfString                            `tfsdk:"excluded_instance_types"`
}

type capacityProviderScalingConfigModel struct {
	MaxVCpuCount    types.Int32                                              `tfsdk:"max_vcpu_count"`
	ScalingMode     fwtypes.StringEnum[awstypes.CapacityProviderScalingMode] `tfsdk:"scaling_mode"`
	ScalingPolicies fwtypes.ListNestedObjectValueOf[scalingPoliciesModel]    `tfsdk:"scaling_policies"`
}

type scalingPoliciesModel struct {
	PredefinedMetricType fwtypes.StringEnum[awstypes.CapacityProviderPredefinedMetricType] `tfsdk:"predefined_metric_type"`
	TargetValue          types.Float64                                                     `tfsdk:"target_value"`
}
