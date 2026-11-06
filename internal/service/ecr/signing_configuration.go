// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecr_signing_configuration", name="Signing Configuration")
// @SingletonIdentity
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// Identity tests are hand-written in signing_configuration_identity_test.go because the
// generated regional-singleton tests assume an `id` attribute.
// @Testing(identityTest=false)
func newSigningConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &signingConfigurationResource{}, nil
}

type signingConfigurationResource struct {
	framework.ResourceWithModel[signingConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *signingConfigurationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"registry_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[signingRuleModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeBetween(1, 10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"signing_profile_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"repository_filter": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[signingRepositoryFilterModel](ctx),
							Validators: []validator.Set{
								setvalidator.SizeAtMost(100),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrFilter: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 256),
											stringvalidator.RegexMatches(
												regexache.MustCompile(`^(?:[a-z0-9*]+(?:[._-][a-z0-9*]+)*/)*[a-z0-9*]+(?:[._-][a-z0-9*]+)*$`),
												"must contain only lowercase alphanumeric characters, wildcards (*), and the separators '.', '_', '-', and '/'",
											),
										},
									},
									"filter_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SigningRepositoryFilterType](),
										Required:   true,
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

func (r *signingConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan signingConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECRClient(ctx)

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.put(ctx, conn, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *signingConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state signingConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECRClient(ctx)

	output, err := findSigningConfiguration(ctx, conn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RegistryID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *signingConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state signingConfigurationResourceModel
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
		conn := r.Meta().ECRClient(ctx)

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.put(ctx, conn, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *signingConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state signingConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECRClient(ctx)

	var input ecr.DeleteSigningConfigurationInput
	_, err := conn.DeleteSigningConfiguration(ctx, &input)
	if errs.IsA[*awstypes.SigningConfigurationNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RegistryID.String())
	}
}

// put replaces the registry's signing configuration with the rules in data, then reads it back
// to populate computed attributes. PutSigningConfiguration does not return the registry ID.
func (r *signingConfigurationResource) put(ctx context.Context, conn *ecr.Client, data *signingConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	var config awstypes.SigningConfiguration
	smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &config))
	if diags.HasError() {
		return diags
	}

	input := ecr.PutSigningConfigurationInput{
		SigningConfiguration: &config,
	}
	if _, err := conn.PutSigningConfiguration(ctx, &input); err != nil {
		smerr.AddError(ctx, &diags, err)
		return diags
	}

	output, err := findSigningConfiguration(ctx, conn)
	if err != nil {
		smerr.AddError(ctx, &diags, err)
		return diags
	}

	smerr.AddEnrich(ctx, &diags, r.flatten(ctx, output, data))

	return diags
}

// flatten is shared by Read and the list resource. The rules are nested one level down in the
// API response, so both levels are flattened into the model.
func (r *signingConfigurationResource) flatten(ctx context.Context, output *ecr.GetSigningConfigurationOutput, data *signingConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, output, data))
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, output.SigningConfiguration, data))

	return diags
}

func findSigningConfiguration(ctx context.Context, conn *ecr.Client) (*ecr.GetSigningConfigurationOutput, error) {
	var input ecr.GetSigningConfigurationInput
	output, err := conn.GetSigningConfiguration(ctx, &input)
	if errs.IsA[*awstypes.SigningConfigurationNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.SigningConfiguration == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

type signingConfigurationResourceModel struct {
	framework.WithRegionModel
	RegistryID types.String                                     `tfsdk:"registry_id"`
	Rules      fwtypes.SetNestedObjectValueOf[signingRuleModel] `tfsdk:"rule"`
}

type signingRuleModel struct {
	RepositoryFilters fwtypes.SetNestedObjectValueOf[signingRepositoryFilterModel] `tfsdk:"repository_filter" autoflex:",omitempty"`
	SigningProfileARN fwtypes.ARN                                                  `tfsdk:"signing_profile_arn"`
}

type signingRepositoryFilterModel struct {
	Filter     types.String                                             `tfsdk:"filter"`
	FilterType fwtypes.StringEnum[awstypes.SigningRepositoryFilterType] `tfsdk:"filter_type"`
}
