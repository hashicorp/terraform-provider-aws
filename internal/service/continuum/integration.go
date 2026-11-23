// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tflistplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/listplanmodifier"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	iamPropagationTimeout = 2 * time.Minute
)

// @FrameworkResource("aws_continuum_integration", name="Integration")
// @IdentityAttribute("id")
// @Testing(preCheck="testAccIntegrationPreCheck")
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importIgnore="provider_configuration.0.s3.0.external_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/continuum;continuum.GetIntegrationOutput")
func newIntegrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &integrationResource{}, nil
}

type integrationResource struct {
	framework.ResourceWithModel[integrationResourceModel]
	framework.WithImportByIdentity
}

func (r *integrationResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	exactlyOneProvider := listvalidator.ExactlyOneOf(
		path.MatchRelative().AtParent().AtName("github"),
		path.MatchRelative().AtParent().AtName("gitlab"),
		path.MatchRelative().AtParent().AtName("s3"),
	)

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"integration_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IntegrationType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"provider_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[providerConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"github": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[githubConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								exactlyOneProvider,
							},
							PlanModifiers: []planmodifier.List{
								tflistplanmodifier.RequiresReplaceIfEmptied,
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"organization_name": schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
											stringplanmodifier.RequiresReplace(),
										},
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 39),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z]([0-9A-Za-z-]*[0-9A-Za-z])?$`), ""),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"authorization": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[githubAuthorizationConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"authorization_code": schema.StringAttribute{
													Required:  true,
													Sensitive: true,
													PlanModifiers: []planmodifier.String{
														requiresReplaceUnlessImported(),
													},
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 512),
													},
												},
												"redirect_url": schema.StringAttribute{
													Optional: true,
													PlanModifiers: []planmodifier.String{
														requiresReplaceUnlessImported(),
													},
													Validators: []validator.String{
														stringvalidator.LengthBetween(9, 2083),
														stringvalidator.RegexMatches(regexache.MustCompile(`^https://`), "must start with https://"),
													},
												},
												names.AttrState: schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														requiresReplaceUnlessImported(),
													},
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 1024),
													},
												},
											},
										},
									},
								},
							},
						},
						"gitlab": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[gitlabConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								exactlyOneProvider,
							},
							PlanModifiers: []planmodifier.List{
								tflistplanmodifier.RequiresReplaceIfEmptied,
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"access_token": schema.StringAttribute{
										Required:  true,
										Sensitive: true,
										PlanModifiers: []planmodifier.String{
											requiresReplaceUnlessImported(),
										},
									},
									"expires_at": schema.StringAttribute{
										CustomType: timetypes.RFC3339Type{},
										Computed:   true,
									},
									"instance_url": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"target_path": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"target_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.GitLabTargetType](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"token_rotation": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.GitLabTokenRotation](),
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
						"s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								exactlyOneProvider,
							},
							PlanModifiers: []planmodifier.List{
								tflistplanmodifier.RequiresReplaceIfEmptied,
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucket: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
										Validators: []validator.String{
											stringvalidator.LengthBetween(3, 63),
										},
									},
									names.AttrExternalID: schema.StringAttribute{
										Required:  true,
										Sensitive: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(2, 1224),
											stringvalidator.RegexMatches(regexache.MustCompile(`^[\w+=,.@:/-]+$`), ""),
										},
									},
									names.AttrRoleARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
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

func (r *integrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan integrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	name := plan.Name.ValueString()
	var input continuum.CreateIntegrationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func(ctx context.Context) (*continuum.CreateIntegrationOutput, error) {
			// A client token replays the recorded result, including terminal failures, so use a fresh one per attempt.
			input.ClientToken = aws.String(create.UniqueId(ctx))
			return conn.CreateIntegration(ctx, &input)
		},
		isRoleNotAssumable,
	)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// CreateIntegration returns only the ID, ARN and creation time.
	id := aws.ToString(output.IntegrationId)
	integration, err := findIntegrationByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, plan.flatten(ctx, integration, integration.ProviderDetails))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *integrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	id := state.ID.ValueString()
	output, err := findIntegrationByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, state.flatten(ctx, output, output.ProviderDetails))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *integrationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state integrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	id := plan.ID.ValueString()

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	providerConfiguration, d := plan.ProviderConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	var (
		integration     any
		providerDetails awstypes.ProviderDetails
	)
	// Only the S3 role and external ID can change in place. Every other change forces replacement,
	// except filling in GitHub or GitLab secrets after import, which needs no API call.
	if diff.HasChanges() && !providerConfiguration.S3.IsNull() {
		var input continuum.UpdateIntegrationInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Integration")))
		if response.Diagnostics.HasError() {
			return
		}

		output, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
			func(ctx context.Context) (*continuum.UpdateIntegrationOutput, error) {
				return conn.UpdateIntegration(ctx, &input)
			},
			isRoleNotAssumable,
		)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		integration, providerDetails = output, output.ProviderDetails
	} else {
		// Nothing changed in AWS, but computed values such as updated_at are unknown in the plan.
		output, err := findIntegrationByID(ctx, conn, id)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		integration, providerDetails = output, output.ProviderDetails
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, plan.flatten(ctx, integration, providerDetails))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *integrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	id := state.ID.ValueString()
	input := continuum.DeleteIntegrationInput{
		IntegrationId: aws.String(id),
	}
	_, err := conn.DeleteIntegration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

func findIntegrationByID(ctx context.Context, conn *continuum.Client, id string) (*continuum.GetIntegrationOutput, error) {
	input := continuum.GetIntegrationInput{
		IntegrationId: aws.String(id),
	}

	return findIntegration(ctx, conn, &input)
}

func findIntegration(ctx context.Context, conn *continuum.Client, input *continuum.GetIntegrationInput) (*continuum.GetIntegrationOutput, error) {
	output, err := conn.GetIntegration(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

type integrationResourceModel struct {
	framework.WithRegionModel
	ARN                   types.String                                                `tfsdk:"arn"`
	CreatedAt             timetypes.RFC3339                                           `tfsdk:"created_at"`
	ID                    types.String                                                `tfsdk:"id"`
	IntegrationType       fwtypes.StringEnum[awstypes.IntegrationType]                `tfsdk:"integration_type"`
	Name                  types.String                                                `tfsdk:"name"`
	ProviderConfiguration fwtypes.ListNestedObjectValueOf[providerConfigurationModel] `tfsdk:"provider_configuration"`
	UpdatedAt             timetypes.RFC3339                                           `tfsdk:"updated_at"`
}

// flatten updates the model from a GetIntegration output or a ListIntegrations item. The service returns
// provider details in place of the provider configuration and never returns secrets, so the prior secret
// values are kept.
func (m *integrationResourceModel) flatten(ctx context.Context, integration any, providerDetails awstypes.ProviderDetails) diag.Diagnostics {
	var diags diag.Diagnostics

	prior, d := m.ProviderConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() {
		return diags
	}

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, integration, m, fwflex.WithFieldNamePrefix("Integration")))
	if diags.HasError() {
		return diags
	}

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, providerDetails, &m.ProviderConfiguration))
	if diags.HasError() {
		return diags
	}

	if prior == nil {
		return diags
	}

	current, d := m.ProviderConfiguration.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || current == nil {
		return diags
	}

	smerr.AddEnrich(ctx, &diags, keepPrior(ctx, &current.GitHub, prior.GitHub, func(current, prior *githubConfigurationModel) {
		current.Authorization = prior.Authorization
	}))
	smerr.AddEnrich(ctx, &diags, keepPrior(ctx, &current.GitLab, prior.GitLab, func(current, prior *gitlabConfigurationModel) {
		current.AccessToken = prior.AccessToken
	}))
	smerr.AddEnrich(ctx, &diags, keepPrior(ctx, &current.S3, prior.S3, func(current, prior *s3ConfigurationModel) {
		current.ExternalID = prior.ExternalID
	}))
	if diags.HasError() {
		return diags
	}

	m.ProviderConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, current)

	return diags
}

// keepPrior copies values that the service does not return from the prior nested object into the current one.
func keepPrior[T any](ctx context.Context, current *fwtypes.ListNestedObjectValueOf[T], prior fwtypes.ListNestedObjectValueOf[T], keep func(current, prior *T)) diag.Diagnostics {
	var diags diag.Diagnostics

	c, d := current.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || c == nil {
		return diags
	}

	p, d := prior.ToPtr(ctx)
	smerr.AddEnrich(ctx, &diags, d)
	if diags.HasError() || p == nil {
		return diags
	}

	keep(c, p)
	*current = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, c)

	return diags
}

type providerConfigurationModel struct {
	GitHub fwtypes.ListNestedObjectValueOf[githubConfigurationModel] `tfsdk:"github"`
	GitLab fwtypes.ListNestedObjectValueOf[gitlabConfigurationModel] `tfsdk:"gitlab"`
	S3     fwtypes.ListNestedObjectValueOf[s3ConfigurationModel]     `tfsdk:"s3"`
}

var (
	_ fwflex.TypedExpander = providerConfigurationModel{}
	_ fwflex.Flattener     = &providerConfigurationModel{}
)

func (m providerConfigurationModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[awstypes.ProviderConfiguration]():
		return m.expandToProviderConfiguration(ctx)

	case reflect.TypeFor[awstypes.UpdateProviderConfiguration]():
		return m.expandToUpdateProviderConfiguration(ctx)
	}

	return nil, nil
}

func (m providerConfigurationModel) expandToProviderConfiguration(ctx context.Context) (awstypes.ProviderConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.GitHub.IsNull():
		var result awstypes.ProviderConfigurationMemberGithub
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.GitHub, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags

	case !m.GitLab.IsNull():
		var result awstypes.ProviderConfigurationMemberGitlab
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.GitLab, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags

	case !m.S3.IsNull():
		var result awstypes.ProviderConfigurationMemberS3
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.S3, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	}

	return nil, diags
}

// expandToUpdateProviderConfiguration expands the S3 configuration, the only provider that can be updated.
func (m providerConfigurationModel) expandToUpdateProviderConfiguration(ctx context.Context) (awstypes.UpdateProviderConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !m.S3.IsNull() {
		var result awstypes.UpdateProviderConfigurationMemberS3
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.S3, &result.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	}

	return nil, diags
}

// Flatten flattens the provider details. Values that the service does not return, such as secrets, are null.
func (m *providerConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ProviderDetailsMemberGithub:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, &t.Value, &m.GitHub))

	case awstypes.ProviderDetailsMemberGitlab:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, &t.Value, &m.GitLab))

	case awstypes.ProviderDetailsMemberS3:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, &t.Value, &m.S3))

	default:
		smerr.AddError(ctx, &diags, fmt.Errorf("unexpected integration provider details type: %T", v))
	}

	return diags
}

type githubConfigurationModel struct {
	Authorization    fwtypes.ListNestedObjectValueOf[githubAuthorizationConfigurationModel] `tfsdk:"authorization"`
	OrganizationName types.String                                                           `tfsdk:"organization_name"`
}

type githubAuthorizationConfigurationModel struct {
	AuthorizationCode types.String `tfsdk:"authorization_code"`
	RedirectURL       types.String `tfsdk:"redirect_url"`
	State             types.String `tfsdk:"state"`
}

type gitlabConfigurationModel struct {
	AccessToken   types.String                                     `tfsdk:"access_token"`
	ExpiresAt     timetypes.RFC3339                                `tfsdk:"expires_at"`
	InstanceURL   types.String                                     `tfsdk:"instance_url"`
	TargetPath    types.String                                     `tfsdk:"target_path"`
	TargetType    fwtypes.StringEnum[awstypes.GitLabTargetType]    `tfsdk:"target_type"`
	TokenRotation fwtypes.StringEnum[awstypes.GitLabTokenRotation] `tfsdk:"token_rotation"`
}

type s3ConfigurationModel struct {
	Bucket     types.String `tfsdk:"bucket"`
	ExternalID types.String `tfsdk:"external_id"`
	RoleARN    fwtypes.ARN  `tfsdk:"role_arn"`
}

// isRoleNotAssumable reports whether Continuum could not yet assume, or use, a freshly created IAM role.
// IAM changes are eventually consistent, so the call is retried.
func isRoleNotAssumable(err error) (bool, error) {
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "could not be assumed") ||
		errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "cannot access the S3 bucket") ||
		errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "role is not authorized to access this Collector") {
		return true, err
	}

	return false, err
}

// requiresReplaceUnlessImported replaces the integration when a secret changes. The service never returns
// secrets, so they are null in state after import; filling them in from configuration is not a change.
func requiresReplaceUnlessImported() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			response.RequiresReplace = !request.StateValue.IsNull()
		},
		"Requires replacement when the value changes, unless there is no prior value.",
		"Requires replacement when the value changes, unless there is no prior value.",
	)
}
