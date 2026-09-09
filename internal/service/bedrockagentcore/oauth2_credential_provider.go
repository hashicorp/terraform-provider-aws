// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	oauth2ClientCredentialsCtxKey = inttypes.NewContextKey[oauth2ClientCredentialsModel]()
)

// @FrameworkResource("aws_bedrockagentcore_oauth2_credential_provider", name="OAuth2 Credential Provider")
// @IdentityAttribute("name")
// @Tags(identifierAttribute="credential_provider_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol;bedrockagentcorecontrol;bedrockagentcorecontrol.GetOauth2CredentialProviderOutput")
// @Testing(importIgnore="oauth2_provider_config.0.github_oauth2_provider_config.0.client_id;oauth2_provider_config.0.github_oauth2_provider_config.0.client_secret")
// @Testing(importStateIdAttribute="name")
// @Testing(preCheck="testAccPreCheckOAuth2CredentialProviders")
// @Testing(preIdentityVersion="v6.63.0")
func newOAuth2CredentialProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &oauth2CredentialProviderResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type oauth2CredentialProviderResource struct {
	framework.ResourceWithModel[oauth2CredentialProviderResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func oauth2ClientCredentialsAttributes(context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"client_credentials_wo_version": schema.Int64Attribute{
			Optional: true,
			Validators: []validator.Int64{
				int64validator.AlsoRequires(path.Expressions{
					path.MatchRelative().AtParent().AtName("client_id_wo"),
				}...),
			},
		},
		names.AttrClientID: schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 256),
				stringvalidator.ConflictsWith(path.Expressions{
					path.MatchRelative().AtParent().AtName("client_id_wo"),
				}...),
				stringvalidator.AlsoRequires(path.Expressions{
					path.MatchRelative().AtParent().AtName(names.AttrClientSecret),
				}...),
				//stringvalidator.PreferWriteOnlyAttribute(path.MatchRelative().AtParent().AtName("client_id_wo")),
			},
		},
		"client_id_wo": schema.StringAttribute{
			Optional:  true,
			WriteOnly: true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 256),
				stringvalidator.ConflictsWith(path.Expressions{
					path.MatchRelative().AtParent().AtName(names.AttrClientID),
				}...),
				stringvalidator.AlsoRequires(path.Expressions{
					path.MatchRelative().AtParent().AtName("client_credentials_wo_version"),
					path.MatchRelative().AtParent().AtName("client_secret_wo"),
				}...),
			},
		},
		names.AttrClientSecret: schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 2048),
				stringvalidator.ConflictsWith(path.Expressions{
					path.MatchRelative().AtParent().AtName("client_secret_wo"),
				}...),
				stringvalidator.AlsoRequires(path.Expressions{
					path.MatchRelative().AtParent().AtName(names.AttrClientID),
				}...),
				//stringvalidator.PreferWriteOnlyAttribute(path.MatchRelative().AtParent().AtName("client_secret_wo")),
			},
		},
		"client_secret_wo": schema.StringAttribute{
			Optional:  true,
			WriteOnly: true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 2048),
				stringvalidator.ConflictsWith(path.Expressions{
					path.MatchRelative().AtParent().AtName(names.AttrClientSecret),
				}...),
				stringvalidator.AlsoRequires(path.Expressions{
					path.MatchRelative().AtParent().AtName("client_credentials_wo_version"),
					path.MatchRelative().AtParent().AtName("client_id_wo"),
				}...),
			},
		},
	}
}

func basicOAuth2ProviderConfigBlock[T any](ctx context.Context) schema.Block {
	attrs := oauth2ClientCredentialsAttributes(ctx)
	attrs["oauth_discovery"] = framework.ResourceComputedListOfObjectsAttribute[oauth2DiscoveryModel](ctx)

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: attrs,
		},
	}
}

func (r *oauth2CredentialProviderResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_secret_arn":       framework.ResourceComputedListOfObjectsAttribute[secretModel](ctx, listplanmodifier.UseStateForUnknown()),
			"credential_provider_arn": framework.ARNAttributeComputedOnly(),
			"credential_provider_vendor": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CredentialProviderVendorType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-_]{1,128}$`), "Valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 50 characters."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"oauth2_provider_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2ProviderConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"custom_oauth2_provider_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customOAuth2ProviderConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: oauth2ClientCredentialsAttributes(ctx),
								Blocks: map[string]schema.Block{
									"oauth_discovery": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2DiscoveryModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"discovery_url": schema.StringAttribute{
													Optional: true,
												},
											},
											Blocks: map[string]schema.Block{
												"authorization_server_metadata": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2AuthorizationServerMetadataModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"authorization_endpoint": schema.StringAttribute{
																Required: true,
															},
															names.AttrIssuer: schema.StringAttribute{
																Required: true,
															},
															"response_types": schema.SetAttribute{
																CustomType:  fwtypes.SetOfStringType,
																ElementType: types.StringType,
																Optional:    true,
															},
															"token_endpoint": schema.StringAttribute{
																Required: true,
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
						"github_oauth2_provider_config":     basicOAuth2ProviderConfigBlock[githubOAuth2ProviderConfigModel](ctx),
						"google_oauth2_provider_config":     basicOAuth2ProviderConfigBlock[googleOAuth2ProviderConfigModel](ctx),
						"microsoft_oauth2_provider_config":  basicOAuth2ProviderConfigBlock[microsoftOAuth2ProviderConfigModel](ctx),
						"salesforce_oauth2_provider_config": basicOAuth2ProviderConfigBlock[salesforceOAuth2ProviderConfigModel](ctx),
						"slack_oauth2_provider_config":      basicOAuth2ProviderConfigBlock[slackOAuth2ProviderConfigModel](ctx),
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

func (r *oauth2CredentialProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan, config oauth2CredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	// Get the effective client credentials.
	clientCredentials, d := plan.clientCredentials(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	// Write-only attribute are only in Config.
	fromConfig, d := config.clientCredentials(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}
	clientCredentials.ClientIDWO = fromConfig.ClientIDWO
	clientCredentials.ClientSecretWO = fromConfig.ClientSecretWO

	// Stuff the client credentials into Context for AutoFlEx.
	ctx = oauth2ClientCredentialsCtxKey.NewContext(ctx, clientCredentials)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input bedrockagentcorecontrol.CreateOauth2CredentialProviderInput
	smerr.AddEnrich(ctx, &response.Diagnostics,
		fwflex.Expand(ctx, plan, &input,
			fwflex.WithFieldNameSuffix("Input"),
		))
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateOauth2CredentialProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	provider, err := waitOAuth2CredentialProviderCreated(ctx, conn, name, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root(names.AttrName), name)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics,
		fwflex.Flatten(ctx, provider, &plan,
			fwflex.WithFieldNameSuffix("Output"),
		))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *oauth2CredentialProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data oauth2CredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// Get the client credentials from State.
	clientCredentials, d := data.clientCredentials(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	out, err := findOAuth2CredentialProviderByName(ctx, conn, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	// Stuff the client credentials into Context for AutoFlEx.
	ctx = oauth2ClientCredentialsCtxKey.NewContext(ctx, clientCredentials)

	smerr.AddEnrich(ctx, &response.Diagnostics,
		fwflex.Flatten(ctx, out, &data,
			fwflex.WithFieldNameSuffix("Output")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *oauth2CredentialProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state, config oauth2CredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		// Get the effective client credentials.
		clientCredentials, d := plan.clientCredentials(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}

		// Write-only attribute are only in Config.
		fromConfig, d := config.clientCredentials(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}
		clientCredentials.ClientIDWO = fromConfig.ClientIDWO
		clientCredentials.ClientSecretWO = fromConfig.ClientSecretWO

		// Stuff the client credentials into Context for AutoFlEx.
		ctx = oauth2ClientCredentialsCtxKey.NewContext(ctx, clientCredentials)

		name := fwflex.StringValueFromFramework(ctx, plan.Name)
		var input bedrockagentcorecontrol.UpdateOauth2CredentialProviderInput
		smerr.AddEnrich(ctx, &response.Diagnostics,
			fwflex.Expand(ctx, plan, &input,
				fwflex.WithFieldNameSuffix("Input")))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateOauth2CredentialProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
			return
		}

		got, err := waitOAuth2CredentialProviderUpdated(ctx, conn, name, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics,
			fwflex.Flatten(ctx, got, &plan,
				fwflex.WithFieldNameSuffix("Output"),
			))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *oauth2CredentialProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data oauth2CredentialProviderResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	input := bedrockagentcorecontrol.DeleteOauth2CredentialProviderInput{
		Name: aws.String(name),
	}
	_, err := conn.DeleteOauth2CredentialProvider(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	if _, err := waitOAuth2CredentialProviderDeleted(ctx, conn, name, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}
}

func findOAuth2CredentialProviderByName(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) {
	input := bedrockagentcorecontrol.GetOauth2CredentialProviderInput{
		Name: aws.String(name),
	}
	return findOAuth2CredentialProvider(ctx, conn, &input)
}

func findOAuth2CredentialProvider(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetOauth2CredentialProviderInput) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) {
	out, err := conn.GetOauth2CredentialProvider(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

func statusOAuth2CredentialProvider(conn *bedrockagentcorecontrol.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findOAuth2CredentialProviderByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func waitOAuth2CredentialProviderCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string, timeout time.Duration) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusCreating),
		Target:                    enum.Slice(awstypes.StatusReady),
		Refresh:                   statusOAuth2CredentialProvider(conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitOAuth2CredentialProviderUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string, timeout time.Duration) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusUpdating),
		Target:                    enum.Slice(awstypes.StatusReady),
		Refresh:                   statusOAuth2CredentialProvider(conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitOAuth2CredentialProviderDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string, timeout time.Duration) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusDeleting, awstypes.StatusReady),
		Target:  []string{},
		Refresh: statusOAuth2CredentialProvider(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type oauth2CredentialProviderResourceModel struct {
	framework.WithRegionModel
	ClientSecretARN          fwtypes.ListNestedObjectValueOf[secretModel]               `tfsdk:"client_secret_arn"`
	CredentialProviderARN    types.String                                               `tfsdk:"credential_provider_arn"`
	CredentialProviderVendor fwtypes.StringEnum[awstypes.CredentialProviderVendorType]  `tfsdk:"credential_provider_vendor"`
	Name                     types.String                                               `tfsdk:"name"`
	OAuth2ProviderConfig     fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"oauth2_provider_config"`
	Tags                     tftags.Map                                                 `tfsdk:"tags"`
	TagsAll                  tftags.Map                                                 `tfsdk:"tags_all"`
	Timeouts                 timeouts.Value                                             `tfsdk:"timeouts"`
}

type oauth2ProviderConfigModel struct {
	CustomOAuth2ProviderConfig     fwtypes.ListNestedObjectValueOf[customOAuth2ProviderConfigModel]     `tfsdk:"custom_oauth2_provider_config"`
	GithubOAuth2ProviderConfig     fwtypes.ListNestedObjectValueOf[githubOAuth2ProviderConfigModel]     `tfsdk:"github_oauth2_provider_config"`
	GoogleOAuth2ProviderConfig     fwtypes.ListNestedObjectValueOf[googleOAuth2ProviderConfigModel]     `tfsdk:"google_oauth2_provider_config"`
	MicrosoftOAuth2ProviderConfig  fwtypes.ListNestedObjectValueOf[microsoftOAuth2ProviderConfigModel]  `tfsdk:"microsoft_oauth2_provider_config"`
	SalesforceOAuth2ProviderConfig fwtypes.ListNestedObjectValueOf[salesforceOAuth2ProviderConfigModel] `tfsdk:"salesforce_oauth2_provider_config"`
	SlackOAuth2ProviderConfig      fwtypes.ListNestedObjectValueOf[slackOAuth2ProviderConfigModel]      `tfsdk:"slack_oauth2_provider_config"`
}

func (m *oauth2CredentialProviderResourceModel) clientCredentials(ctx context.Context) (oauth2ClientCredentialsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	data, d := m.OAuth2ProviderConfig.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || data == nil {
		return inttypes.Zero[oauth2ClientCredentialsModel](), diags
	}

	v, d := data.clientCredentials(ctx)
	diags.Append(d...)

	return v, diags
}

var (
	_ fwflex.Flattener = &oauth2ProviderConfigModel{}
	_ fwflex.Expander  = &oauth2ProviderConfigModel{}
)

func (m *oauth2ProviderConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Propagate client credentials from State.
	clientCredentials := oauth2ClientCredentialsCtxKey.FromContext(ctx)

	switch t := v.(type) {
	case awstypes.Oauth2ProviderConfigOutputMemberCustomOauth2ProviderConfig:
		var model customOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ClientCredentialsModel = clientCredentials
		m.CustomOAuth2ProviderConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberGithubOauth2ProviderConfig:
		var model githubOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ClientCredentialsModel = clientCredentials
		m.GithubOAuth2ProviderConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberGoogleOauth2ProviderConfig:
		var model googleOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ClientCredentialsModel = clientCredentials
		m.GoogleOAuth2ProviderConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberMicrosoftOauth2ProviderConfig:
		var model microsoftOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ClientCredentialsModel = clientCredentials
		m.MicrosoftOAuth2ProviderConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberSalesforceOauth2ProviderConfig:
		var model salesforceOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ClientCredentialsModel = clientCredentials
		m.SalesforceOAuth2ProviderConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberSlackOauth2ProviderConfig:
		var model slackOAuth2ProviderConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		model.oauth2ClientCredentialsModel = clientCredentials
		m.SlackOAuth2ProviderConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("oauth2_provider_config flatten: %T", v))
	}

	return diags
}

func (m oauth2ProviderConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Set API fields for client credentials.
	clientCredentials := oauth2ClientCredentialsCtxKey.FromContext(ctx)
	if !clientCredentials.ClientIDWO.IsNull() && !clientCredentials.ClientSecretWO.IsNull() {
		clientCredentials.ClientID = clientCredentials.ClientIDWO
		clientCredentials.ClientSecret = clientCredentials.ClientSecretWO
	}

	switch {
	case !m.CustomOAuth2ProviderConfig.IsNull():
		data, d := m.CustomOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		data.oauth2ClientCredentialsModel = clientCredentials
		var r awstypes.Oauth2ProviderConfigInputMemberCustomOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.GithubOAuth2ProviderConfig.IsNull():
		data, d := m.GithubOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		data.oauth2ClientCredentialsModel = clientCredentials
		var r awstypes.Oauth2ProviderConfigInputMemberGithubOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.GoogleOAuth2ProviderConfig.IsNull():
		data, d := m.GoogleOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		data.oauth2ClientCredentialsModel = clientCredentials
		var r awstypes.Oauth2ProviderConfigInputMemberGoogleOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.MicrosoftOAuth2ProviderConfig.IsNull():
		data, d := m.MicrosoftOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		data.oauth2ClientCredentialsModel = clientCredentials
		var r awstypes.Oauth2ProviderConfigInputMemberMicrosoftOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.SalesforceOAuth2ProviderConfig.IsNull():
		data, d := m.SalesforceOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		data.oauth2ClientCredentialsModel = clientCredentials
		var r awstypes.Oauth2ProviderConfigInputMemberSalesforceOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.SlackOAuth2ProviderConfig.IsNull():
		data, d := m.SlackOAuth2ProviderConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		data.oauth2ClientCredentialsModel = clientCredentials
		var r awstypes.Oauth2ProviderConfigInputMemberSlackOauth2ProviderConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

func (m *oauth2ProviderConfigModel) clientCredentials(ctx context.Context) (oauth2ClientCredentialsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.CustomOAuth2ProviderConfig.IsNull():
		v, d := m.CustomOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2ClientCredentialsModel](), diags
		}
		return v.oauth2ClientCredentialsModel, diags

	case !m.GithubOAuth2ProviderConfig.IsNull():
		v, d := m.GithubOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2ClientCredentialsModel](), diags
		}
		return v.oauth2ClientCredentialsModel, diags

	case !m.GoogleOAuth2ProviderConfig.IsNull():
		v, d := m.GoogleOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2ClientCredentialsModel](), diags
		}
		return v.oauth2ClientCredentialsModel, diags

	case !m.MicrosoftOAuth2ProviderConfig.IsNull():
		v, d := m.MicrosoftOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2ClientCredentialsModel](), diags
		}
		return v.oauth2ClientCredentialsModel, diags

	case !m.SalesforceOAuth2ProviderConfig.IsNull():
		v, d := m.SalesforceOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2ClientCredentialsModel](), diags
		}
		return v.oauth2ClientCredentialsModel, diags

	case !m.SlackOAuth2ProviderConfig.IsNull():
		v, d := m.SlackOAuth2ProviderConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return inttypes.Zero[oauth2ClientCredentialsModel](), diags
		}
		return v.oauth2ClientCredentialsModel, diags
	}

	return inttypes.Zero[oauth2ClientCredentialsModel](), diags
}

type oauth2ClientCredentialsModel struct {
	ClientCredentialsWOVersion types.Int64  `tfsdk:"client_credentials_wo_version"`
	ClientID                   types.String `tfsdk:"client_id"`
	ClientIDWO                 types.String `tfsdk:"client_id_wo"`
	ClientSecret               types.String `tfsdk:"client_secret"`
	ClientSecretWO             types.String `tfsdk:"client_secret_wo"`
}

type oauth2DiscoveryModel struct {
	AuthorizationServerMetadata fwtypes.ListNestedObjectValueOf[oauth2AuthorizationServerMetadataModel] `tfsdk:"authorization_server_metadata"`
	DiscoveryURL                types.String                                                            `tfsdk:"discovery_url"`
}

type basicOAuth2ProviderConfigModel struct {
	oauth2ClientCredentialsModel
	OAuthDiscovery fwtypes.ListNestedObjectValueOf[oauth2DiscoveryModel] `tfsdk:"oauth_discovery"`
}

type customOAuth2ProviderConfigModel struct {
	oauth2ClientCredentialsModel
	OAuthDiscovery fwtypes.ListNestedObjectValueOf[oauth2DiscoveryModel] `tfsdk:"oauth_discovery"`
}

type githubOAuth2ProviderConfigModel struct {
	basicOAuth2ProviderConfigModel
}

type googleOAuth2ProviderConfigModel struct {
	basicOAuth2ProviderConfigModel
}

type microsoftOAuth2ProviderConfigModel struct {
	basicOAuth2ProviderConfigModel
}

type salesforceOAuth2ProviderConfigModel struct {
	basicOAuth2ProviderConfigModel
}

type slackOAuth2ProviderConfigModel struct {
	basicOAuth2ProviderConfigModel
}

var (
	_ fwflex.Flattener = &oauth2DiscoveryModel{}
	_ fwflex.Expander  = &oauth2DiscoveryModel{}
)

func (m *oauth2DiscoveryModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.Oauth2DiscoveryMemberDiscoveryUrl:
		m.DiscoveryURL = fwflex.StringValueToFramework(ctx, t.Value)

	case awstypes.Oauth2DiscoveryMemberAuthorizationServerMetadata:
		var model oauth2AuthorizationServerMetadataModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.AuthorizationServerMetadata = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("oauth2DiscoveryModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m oauth2DiscoveryModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.DiscoveryURL.IsNull():
		var r awstypes.Oauth2DiscoveryMemberDiscoveryUrl
		r.Value = fwflex.StringValueFromFramework(ctx, m.DiscoveryURL)
		return &r, diags

	case !m.AuthorizationServerMetadata.IsNull():
		model, d := m.AuthorizationServerMetadata.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.Oauth2DiscoveryMemberAuthorizationServerMetadata
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type oauth2AuthorizationServerMetadataModel struct {
	AuthorizationEndpoint types.String        `tfsdk:"authorization_endpoint"`
	Issuer                types.String        `tfsdk:"issuer"`
	ResponseTypes         fwtypes.SetOfString `tfsdk:"response_types"`
	TokenEndpoint         types.String        `tfsdk:"token_endpoint"`
}
