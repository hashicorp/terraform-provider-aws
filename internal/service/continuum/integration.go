// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"
	"errors"

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_continuum_integration", name="Integration")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(identityTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/continuum;continuum.GetIntegrationOutput")
func newIntegrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &integrationResource{}, nil
}

type integrationResource struct {
	framework.ResourceWithModel[integrationResourceModel]
	framework.WithImportByIdentity
}

func (r *integrationResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	oneOf := func() validator.List {
		return listvalidator.ExactlyOneOf(
			path.MatchRoot("github"),
			path.MatchRoot("gitlab"),
			path.MatchRoot("s3"),
		)
	}

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
			"github": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[githubIntegrationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					oneOf(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"organization_name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "GitHub organization that the integration installs into. Omit for a personal account.",
						},
					},
					Blocks: map[string]schema.Block{
						"authorization": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[githubAuthorizationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"authorization_code": schema.StringAttribute{
										Required:    true,
										Sensitive:   true,
										Description: "Single-use OAuth authorization code returned by GitHub. Obtain it with `StartIntegrationAuthorization`.",
									},
									"redirect_url": schema.StringAttribute{
										Required: true,
									},
									names.AttrState: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"gitlab": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[gitlabIntegrationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					oneOf(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"access_token": schema.StringAttribute{
							Required:  true,
							Sensitive: true,
						},
						"expires_at": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"instance_url": schema.StringAttribute{
							Required: true,
						},
						"target_path": schema.StringAttribute{
							Required: true,
						},
						"target_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.GitLabTargetType](),
							Required:   true,
						},
						"token_rotation": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.GitLabTokenRotation](),
							Computed:   true,
						},
					},
				},
			},
			"s3": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3IntegrationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					oneOf(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrBucket: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrExternalID: schema.StringAttribute{
							Optional:    true,
							Sensitive:   true,
							Description: "External ID that Continuum presents when assuming `role_arn`. The service does not return this value.",
						},
						names.AttrRoleARN: schema.StringAttribute{
							CustomType:  fwtypes.ARNType,
							Required:    true,
							Description: "ARN of the IAM role that Continuum assumes to read the bucket.",
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

	providerConfig, diags := plan.expandProviderConfiguration(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	input := continuum.CreateIntegrationInput{
		ClientToken:           aws.String(create.UniqueId(ctx)),
		Name:                  aws.String(name),
		ProviderConfiguration: providerConfig,
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
	if output == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, name)
		return
	}

	id := aws.ToString(output.IntegrationId)
	integration, err := findIntegrationByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, plan.flatten(ctx, integration))
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

	output, err := findIntegrationByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, state.flatten(ctx, output))
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

	// Only the S3 role and external ID can change in place; every other change forces replacement.
	if !plan.S3.Equal(state.S3) {
		s3, d := plan.S3.ToPtr(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}

		input := continuum.UpdateIntegrationInput{
			IntegrationId: plan.ID.ValueStringPointer(),
			ProviderConfiguration: &awstypes.UpdateProviderConfigurationMemberS3{
				Value: awstypes.UpdateS3Configuration{
					RoleArn:    s3.RoleARN.ValueStringPointer(),
					ExternalId: s3.ExternalID.ValueStringPointer(),
				},
			},
		}

		_, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
			func(ctx context.Context) (*continuum.UpdateIntegrationOutput, error) {
				return conn.UpdateIntegration(ctx, &input)
			},
			isRoleNotAssumable,
		)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	output, err := findIntegrationByID(ctx, conn, plan.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, plan.flatten(ctx, output))
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

	input := continuum.DeleteIntegrationInput{
		IntegrationId: state.ID.ValueStringPointer(),
	}
	_, err := conn.DeleteIntegration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
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

type integrationResourceModel struct {
	framework.WithRegionModel
	ARN             types.String                                            `tfsdk:"arn"`
	CreatedAt       timetypes.RFC3339                                       `tfsdk:"created_at"`
	GitHub          fwtypes.ListNestedObjectValueOf[githubIntegrationModel] `tfsdk:"github"`
	GitLab          fwtypes.ListNestedObjectValueOf[gitlabIntegrationModel] `tfsdk:"gitlab"`
	ID              types.String                                            `tfsdk:"id"`
	IntegrationType fwtypes.StringEnum[awstypes.IntegrationType]            `tfsdk:"integration_type"`
	Name            types.String                                            `tfsdk:"name"`
	S3              fwtypes.ListNestedObjectValueOf[s3IntegrationModel]     `tfsdk:"s3"`
	UpdatedAt       timetypes.RFC3339                                       `tfsdk:"updated_at"`
}

type githubIntegrationModel struct {
	Authorization    fwtypes.ListNestedObjectValueOf[githubAuthorizationModel] `tfsdk:"authorization"`
	OrganizationName types.String                                              `tfsdk:"organization_name"`
}

type githubAuthorizationModel struct {
	AuthorizationCode types.String `tfsdk:"authorization_code"`
	RedirectURL       types.String `tfsdk:"redirect_url"`
	State             types.String `tfsdk:"state"`
}

type gitlabIntegrationModel struct {
	AccessToken   types.String                                     `tfsdk:"access_token"`
	ExpiresAt     timetypes.RFC3339                                `tfsdk:"expires_at"`
	InstanceURL   types.String                                     `tfsdk:"instance_url"`
	TargetPath    types.String                                     `tfsdk:"target_path"`
	TargetType    fwtypes.StringEnum[awstypes.GitLabTargetType]    `tfsdk:"target_type"`
	TokenRotation fwtypes.StringEnum[awstypes.GitLabTokenRotation] `tfsdk:"token_rotation"`
}

type s3IntegrationModel struct {
	Bucket     types.String `tfsdk:"bucket"`
	ExternalID types.String `tfsdk:"external_id"`
	RoleARN    fwtypes.ARN  `tfsdk:"role_arn"`
}

func (m *integrationResourceModel) expandProviderConfiguration(ctx context.Context) (awstypes.ProviderConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.S3.IsNull():
		s3, d := m.S3.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.ProviderConfigurationMemberS3{
			Value: awstypes.S3Configuration{
				Bucket:     s3.Bucket.ValueStringPointer(),
				ExternalId: s3.ExternalID.ValueStringPointer(),
				RoleArn:    s3.RoleARN.ValueStringPointer(),
			},
		}, diags

	case !m.GitLab.IsNull():
		gl, d := m.GitLab.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.ProviderConfigurationMemberGitlab{
			Value: awstypes.GitLabConfiguration{
				AccessToken: gl.AccessToken.ValueStringPointer(),
				InstanceUrl: gl.InstanceURL.ValueStringPointer(),
				TargetPath:  gl.TargetPath.ValueStringPointer(),
				TargetType:  gl.TargetType.ValueEnum(),
			},
		}, diags

	case !m.GitHub.IsNull():
		gh, d := m.GitHub.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		auth, d := gh.Authorization.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.ProviderConfigurationMemberGithub{
			Value: awstypes.GitHubConfiguration{
				OrganizationName: gh.OrganizationName.ValueStringPointer(),
				Authorization: &awstypes.GitHubAuthorizationConfiguration{
					AuthorizationCode: auth.AuthorizationCode.ValueStringPointer(),
					RedirectUrl:       auth.RedirectURL.ValueStringPointer(),
					State:             auth.State.ValueStringPointer(),
				},
			},
		}, diags
	}

	diags.AddError("Missing provider configuration", "exactly one of github, gitlab or s3 must be configured")
	return nil, diags
}

// flatten updates the model from the API response. Secret values (external IDs, tokens,
// authorization codes) are never returned by the service, so the prior values are kept.
func (m *integrationResourceModel) flatten(ctx context.Context, output *continuum.GetIntegrationOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ARN = types.StringPointerValue(output.IntegrationArn)
	m.CreatedAt = timetypes.NewRFC3339TimePointerValue(output.CreatedAt)
	m.ID = types.StringPointerValue(output.IntegrationId)
	m.IntegrationType = fwtypes.StringEnumValue(output.IntegrationType)
	m.Name = types.StringPointerValue(output.Name)
	m.UpdatedAt = timetypes.NewRFC3339TimePointerValue(output.UpdatedAt)

	switch v := output.ProviderDetails.(type) {
	case *awstypes.ProviderDetailsMemberS3:
		var s3 s3IntegrationModel
		if !m.S3.IsNull() && !m.S3.IsUnknown() {
			prior, d := m.S3.ToPtr(ctx)
			diags.Append(d...)
			if prior != nil {
				s3 = *prior
			}
		}
		s3.Bucket = types.StringPointerValue(v.Value.Bucket)
		s3.RoleARN = fwtypes.ARNValue(aws.ToString(v.Value.RoleArn))
		if s3.ExternalID.IsUnknown() {
			s3.ExternalID = types.StringNull()
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &s3)
		m.GitHub = fwtypes.NewListNestedObjectValueOfNull[githubIntegrationModel](ctx)
		m.GitLab = fwtypes.NewListNestedObjectValueOfNull[gitlabIntegrationModel](ctx)

	case *awstypes.ProviderDetailsMemberGitlab:
		var gl gitlabIntegrationModel
		if !m.GitLab.IsNull() && !m.GitLab.IsUnknown() {
			prior, d := m.GitLab.ToPtr(ctx)
			diags.Append(d...)
			if prior != nil {
				gl = *prior
			}
		}
		gl.ExpiresAt = timetypes.NewRFC3339TimePointerValue(v.Value.ExpiresAt)
		gl.InstanceURL = types.StringPointerValue(v.Value.InstanceUrl)
		gl.TargetPath = types.StringPointerValue(v.Value.TargetPath)
		gl.TargetType = fwtypes.StringEnumValue(v.Value.TargetType)
		gl.TokenRotation = fwtypes.StringEnumValue(v.Value.TokenRotation)
		if gl.AccessToken.IsUnknown() {
			gl.AccessToken = types.StringNull()
		}
		m.GitLab = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gl)
		m.GitHub = fwtypes.NewListNestedObjectValueOfNull[githubIntegrationModel](ctx)
		m.S3 = fwtypes.NewListNestedObjectValueOfNull[s3IntegrationModel](ctx)

	case *awstypes.ProviderDetailsMemberGithub:
		var gh githubIntegrationModel
		if !m.GitHub.IsNull() && !m.GitHub.IsUnknown() {
			prior, d := m.GitHub.ToPtr(ctx)
			diags.Append(d...)
			if prior != nil {
				gh = *prior
			}
		}
		gh.OrganizationName = types.StringPointerValue(v.Value.OrganizationName)
		if gh.Authorization.IsUnknown() {
			gh.Authorization = fwtypes.NewListNestedObjectValueOfNull[githubAuthorizationModel](ctx)
		}
		m.GitHub = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gh)
		m.GitLab = fwtypes.NewListNestedObjectValueOfNull[gitlabIntegrationModel](ctx)
		m.S3 = fwtypes.NewListNestedObjectValueOfNull[s3IntegrationModel](ctx)

	default:
		diags.AddError("Unexpected integration provider", enum.Slice(output.IntegrationType)[0])
	}

	return diags
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
