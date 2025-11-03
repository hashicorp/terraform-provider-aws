// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/service/lambda/mocks"
	awsmocktypes "github.com/hashicorp/terraform-provider-aws/internal/service/lambda/mocks/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambda_capacity_provider", name="Capacity Provider")
// @Tags(identifierAttribute="arn")
func newResourceCapacityProvider(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCapacityProvider{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameCapacityProvider = "Capacity Provider"
)

type resourceCapacityProvider struct {
	framework.ResourceWithModel[resourceCapacityProviderModel]
	framework.WithTimeouts
}

func (r *resourceCapacityProvider) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"type": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"vpc_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"subnet_ids": schema.SetAttribute{
							Required:   true,
							CustomType: fwtypes.SetOfStringType,
						},
						"security_group_ids": schema.SetAttribute{
							Optional:   true,
							CustomType: fwtypes.SetOfStringType,
						},
					},
				},
			},
			"permissions_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[permissionConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"instance_profile_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						"capacity_provider_operator_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
			"capacity_provider_overrides": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderOverridesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"kms_key_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"instance_requirements": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[instanceRequirementsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"architecture": schema.StringAttribute{
										Optional: true,
									},
									"allowed_instance_types": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Optional:   true,
									},
									"excluded_instance_types": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Optional:   true,
									},
								},
							},
						},
						"capacity_provider_autoscaling": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[capacityProviderAutoscalingModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"max_instance_count": schema.Int64Attribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"instance_scaling_targets_percent": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[instanceScalingTargetsPercentageModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"target_cpu": schema.Int64Attribute{
													Optional: true,
												},
												"target_memory": schema.Int64Attribute{
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceCapacityProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := mocks.LambdaClient(ctx)

	var plan resourceCapacityProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input awsmocktypes.CreateCapacityProviderInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("CapacityProvider")))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateCapacityProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitCapacityProviderActive(ctx, conn, plan.ARN.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceCapacityProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := mocks.LambdaClient(ctx)

	var state resourceCapacityProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCapacityProviderByArn(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("CapacityProvider")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceCapacityProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := mocks.LambdaClient(ctx)

	var plan, state resourceCapacityProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input awsmocktypes.UpdateCapacityProviderInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("CapacityProvider")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateCapacityProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ARN.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitCapacityProviderActive(ctx, conn, plan.ARN.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.String())
			return
		}
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceCapacityProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := mocks.LambdaClient(ctx)

	var state resourceCapacityProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := awsmocktypes.DeleteCapacityProviderInput{}

	_, err := conn.DeleteCapacityProvider(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitCapacityProviderDeleted(ctx, conn, state.ARN.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
		return
	}
}

func (r *resourceCapacityProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitCapacityProviderActive(ctx context.Context, conn *mocks.Client, id string, timeout time.Duration) (*awsmocktypes.GetCapacityProviderOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:                   enum.Slice(awsmocktypes.StatePending),
		Target:                    enum.Slice(awsmocktypes.StateActive),
		Refresh:                   statusCapacityProvider(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awsmocktypes.GetCapacityProviderOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitCapacityProviderDeleted(ctx context.Context, conn *mocks.Client, id string, timeout time.Duration) (*awsmocktypes.DeleteCapacityProviderOutput, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(awsmocktypes.StatePending, awsmocktypes.StateDeleting),
		Target:  []string{},
		Refresh: statusCapacityProvider(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awsmocktypes.DeleteCapacityProviderOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusCapacityProvider(ctx context.Context, conn *mocks.Client, id string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findCapacityProviderByArn(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findCapacityProviderByArn(ctx context.Context, conn *mocks.Client, id string) (*awsmocktypes.GetCapacityProviderOutput, error) {
	input := awsmocktypes.GetCapacityProviderInput{
		CapacityProviderName: aws.String(id),
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
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type resourceCapacityProviderModel struct {
	framework.WithRegionModel
	ARN                       types.String                                                    `tfsdk:"arn"`
	Name                      types.String                                                    `tfsdk:"name"`
	PermissionConfig          fwtypes.ListNestedObjectValueOf[permissionConfigModel]          `tfsdk:"permission_config"`
	VpcConfig                 fwtypes.ListNestedObjectValueOf[vpcConfigModel]                 `tfsdk:"vpc_config"`
	CapacityProviderOverrides fwtypes.ListNestedObjectValueOf[capacityProviderOverridesModel] `tfsdk:"capacity_provider_overrides"`
	Tags                      tftags.Map                                                      `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                      `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                  `tfsdk:"timeouts"`
}

type vpcConfigModel struct {
	SecurityGroupIDs fwtypes.SetOfString `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetOfString `tfsdk:"subnet_ids"`
}

type permissionConfigModel struct {
	InstanceProfileARN          fwtypes.ARN `tfsdk:"instance_profile_arn"`
	CapacityProviderOperatorARN fwtypes.ARN `tfsdk:"capacity_provider_operator_arn"`
}

type capacityProviderOverridesModel struct {
	InstanceRequirements        fwtypes.ListNestedObjectValueOf[instanceRequirementsModel]        `tfsdk:"instance_requirements"`
	CapacityProviderAutoscaling fwtypes.ListNestedObjectValueOf[capacityProviderAutoscalingModel] `tfsdk:"capacity_provider_autoscaling"`
	KmsKeyARN                   fwtypes.ARN                                                       `tfsdk:"kms_key_arn"`
}

type instanceRequirementsModel struct {
	Architecture          types.String         `tfsdk:"architecture"`
	AllowedInstanceTypes  fwtypes.ListOfString `tfsdk:"allowed_instance_types"`
	ExcludedInstanceTypes fwtypes.ListOfString `tfsdk:"excluded_instance_types"`
}

type capacityProviderAutoscalingModel struct {
	MaxInstanceCount              types.Int64                                                            `tfsdk:"max_instance_count"`
	InstanceScalingTargetsPercent fwtypes.ListNestedObjectValueOf[instanceScalingTargetsPercentageModel] `tfsdk:"instance_scaling_targets_percent"`
}

type instanceScalingTargetsPercentageModel struct {
	TargetCPU    types.Int64 `tfsdk:"target_cpu"`
	TargetMemory types.Int64 `tfsdk:"target_memory"`
}

//func sweepCapacityProviders(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
//	input := lambda.ListCapacityProvidersInput{}
//	conn := client.LambdaClient(ctx)
//	var sweepResources []sweep.Sweepable
//
//	pages := lambda.NewListCapacityProvidersPaginator(conn, &input)
//	for pages.HasMorePages() {
//		page, err := pages.NextPage(ctx)
//		if err != nil {
//			return nil, smarterr.NewError(err)
//		}
//
//		for _, v := range page.CapacityProviders {
//			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceCapacityProvider, client,
//				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.CapacityProviderId))),
//			)
//		}
//	}
//
//	return sweepResources, nil
//}
