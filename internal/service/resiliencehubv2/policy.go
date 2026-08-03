// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// @FrameworkResource("aws_resiliencehubv2_policy", name="Policy")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.Policy")
// @Testing(hasNoPreExistingResource=true)
func newPolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &policyResource{}, nil
}

type policyResource struct {
	framework.ResourceWithModel[resourcePolicyModel]
	framework.WithImportByIdentity
}

func (r *policyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 615),
				},
			},
			names.AttrKMSKeyID: fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_\-]{1,59}$`),
						`Must start with a letter or number. 2-60 characters. Use letters, numbers, hyphens, and underscores only.`,
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"availability_slo": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[availabilitySLOModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.AtLeastOneOf(
						path.MatchRoot("availability_slo"),
						path.MatchRoot("data_recovery"),
						path.MatchRoot("multi_az"),
						path.MatchRoot("multi_region"),
					),
				},
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						names.AttrTarget: fwschema.Float64Attribute{
							Required: true,
							Validators: []validator.Float64{
								float64validator.Between(0, 100),
							},
						},
					},
				},
			},
			"data_recovery": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataRecoveryTargetsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"time_between_backups_in_minutes": fwschema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
							},
						},
					},
				},
			},
			"multi_az": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiAZTargetsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"disaster_recovery_approach": fwschema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MultiAzDisasterRecoveryApproach](),
							Required:   true,
						},
						"rpo_in_minutes": fwschema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
							},
						},
						"rto_in_minutes": fwschema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
							},
						},
					},
				},
			},
			"multi_region": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiRegionTargetsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"disaster_recovery_approach": fwschema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MultiRegionDisasterRecoveryApproach](),
							Required:   true,
						},
						"rpo_in_minutes": fwschema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
							},
						},
						"rto_in_minutes": fwschema.Int32Attribute{
							Optional: true,
							Validators: []validator.Int32{
								int32validator.AtLeast(0),
							},
						},
					},
				},
			},
		},
	}
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourcePolicyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.CreatePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	// Set values for unknowns.
	plan.PolicyARN = fwflex.StringToFramework(ctx, output.Policy.PolicyArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourcePolicyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.PolicyARN)
	policy, err := findPolicyByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, policy, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourcePolicyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, plan.PolicyARN)
		var input resiliencehubv2.UpdatePolicyInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdatePolicy(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output.Policy, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourcePolicyModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.PolicyARN)
	input := resiliencehubv2.DeletePolicyInput{
		PolicyArn: aws.String(arn),
	}
	_, err := conn.DeletePolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
	}
}

func (r *policyResource) flatten(ctx context.Context, policy *awstypes.Policy, data *resourcePolicyModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, policy, data)...)

	return diags
}

func findPolicyByARN(ctx context.Context, conn *resiliencehubv2.Client, arn string) (*awstypes.Policy, error) {
	input := resiliencehubv2.GetPolicyInput{
		PolicyArn: aws.String(arn),
	}

	return findPolicy(ctx, conn, &input)
}

func findPolicy(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.GetPolicyInput) (*awstypes.Policy, error) {
	output, err := conn.GetPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.Policy == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.Policy, nil
}

type resourcePolicyModel struct {
	framework.WithRegionModel
	AvailabilitySlo fwtypes.ListNestedObjectValueOf[availabilitySLOModel]     `tfsdk:"availability_slo"`
	DataRecovery    fwtypes.ListNestedObjectValueOf[dataRecoveryTargetsModel] `tfsdk:"data_recovery"`
	Description     types.String                                              `tfsdk:"description"`
	KMSKeyID        fwtypes.ARN                                               `tfsdk:"kms_key_id"`
	MultiAz         fwtypes.ListNestedObjectValueOf[multiAZTargetsModel]      `tfsdk:"multi_az"`
	MultiRegion     fwtypes.ListNestedObjectValueOf[multiRegionTargetsModel]  `tfsdk:"multi_region"`
	Name            types.String                                              `tfsdk:"name"`
	PolicyARN       types.String                                              `tfsdk:"arn"`
	Tags            tftags.Map                                                `tfsdk:"tags"`
	TagsAll         tftags.Map                                                `tfsdk:"tags_all"`
}

type availabilitySLOModel struct {
	Target types.Float64 `tfsdk:"target"`
}

type dataRecoveryTargetsModel struct {
	TimeBetweenBackupsInMinutes types.Int32 `tfsdk:"time_between_backups_in_minutes"`
}

type multiAZTargetsModel struct {
	DisasterRecoveryApproach fwtypes.StringEnum[awstypes.MultiAzDisasterRecoveryApproach] `tfsdk:"disaster_recovery_approach"`
	RPOInMinutes             types.Int32                                                  `tfsdk:"rpo_in_minutes"`
	RTOInMinutes             types.Int32                                                  `tfsdk:"rto_in_minutes"`
}

type multiRegionTargetsModel struct {
	DisasterRecoveryApproach fwtypes.StringEnum[awstypes.MultiRegionDisasterRecoveryApproach] `tfsdk:"disaster_recovery_approach"`
	RPOIntervalInMinutes     types.Int32                                                      `tfsdk:"rpo_in_minutes"`
	RTOInMinutes             types.Int32                                                      `tfsdk:"rto_in_minutes"`
}
