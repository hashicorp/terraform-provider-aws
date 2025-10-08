// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"reflect"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_oauth2_credential_provider", name="OAuth2 Credential Provider")
func newOAuth2CredentialProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &oauth2CredentialProviderResource{}
	return r, nil
}

type oauth2CredentialProviderResource struct {
	framework.ResourceWithModel[oauth2CredentialProviderResourceModel]
}

func clientCredentialAttributes() map[string]schema.Attribute {
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
				stringvalidator.ConflictsWith(path.Expressions{
					path.MatchRelative().AtParent().AtName("client_id_wo"),
				}...),
				stringvalidator.AlsoRequires(path.Expressions{
					path.MatchRelative().AtParent().AtName(names.AttrClientSecret),
				}...),
			},
		},
		"client_id_wo": schema.StringAttribute{
			Optional:  true,
			WriteOnly: true,
			Sensitive: true,
			Validators: []validator.String{
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
				stringvalidator.ConflictsWith(path.Expressions{
					path.MatchRelative().AtParent().AtName("client_secret_wo"),
				}...),
				stringvalidator.AlsoRequires(path.Expressions{
					path.MatchRelative().AtParent().AtName(names.AttrClientID),
				}...),
			},
		},
		"client_secret_wo": schema.StringAttribute{
			Optional:  true,
			WriteOnly: true,
			Sensitive: true,
			Validators: []validator.String{
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

func basicOAuth2ProviderBlock(ctx context.Context) schema.ListNestedBlock {
	attrs := clientCredentialAttributes()
	attrs["oauth_discovery"] = schema.ListAttribute{
		CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2DiscoveryModel](ctx),
		Computed:   true,
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2ProviderConfigModel](ctx),
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
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.CredentialProviderVendorType](),
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9\-_]+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2ProviderConfigInputModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"custom": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[oauth2ProviderConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: clientCredentialAttributes(),
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
						"github":     basicOAuth2ProviderBlock(ctx),
						"google":     basicOAuth2ProviderBlock(ctx),
						"microsoft":  basicOAuth2ProviderBlock(ctx),
						"salesforce": basicOAuth2ProviderBlock(ctx),
						"slack":      basicOAuth2ProviderBlock(ctx),
					},
				},
			},
		},
	}
}

func (r *oauth2CredentialProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan, config oauth2CredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	creds, d := config.CredsValue(ctx)
	smerr.EnrichAppend(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}
	ctxWithCreds := withOAuth2Creds(ctx, creds)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input bedrockagentcorecontrol.CreateOauth2CredentialProviderInput
	smerr.EnrichAppend(ctx, &response.Diagnostics,
		fwflex.Expand(ctxWithCreds, plan, &input,
			fwflex.WithFieldNameSuffix("Input"),
		))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateOauth2CredentialProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// Refresh from GET as oauth_discovery is not returned in CreateOauth2CredentialProviderOutput.
	provider, err := findOAuth2CredentialProviderByName(ctx, conn, name)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics,
		fwflex.Flatten(ctxWithCreds, provider, &plan,
			fwflex.WithFieldNameSuffix("Output"),
		))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *oauth2CredentialProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data oauth2CredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	out, err := findOAuth2CredentialProviderByName(ctx, conn, name)
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// TODO Remove before AgentCore GA
	// Store clientId/clientSecret in context before flattening zeroes them
	// This won't be necessary in AgentCore GA - AWS API already returns those values but awstypes don't
	creds, d := data.CredsValue(ctx)
	smerr.EnrichAppend(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}
	ctxWithCreds := withOAuth2Creds(ctx, creds)

	smerr.EnrichAppend(ctx, &response.Diagnostics,
		fwflex.Flatten(ctxWithCreds, out, &data,
			fwflex.WithFieldNameSuffix("Output")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *oauth2CredentialProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state, config oauth2CredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	credsConfig, d := config.CredsValue(ctx)
	smerr.EnrichAppend(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}
	ctxWithCreds := withOAuth2Creds(ctx, credsConfig)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		name := fwflex.StringValueFromFramework(ctx, plan.Name)
		var input bedrockagentcorecontrol.UpdateOauth2CredentialProviderInput
		smerr.EnrichAppend(ctx, &response.Diagnostics,
			fwflex.Expand(ctxWithCreds, plan, &input,
				fwflex.WithFieldNameSuffix("Input")))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateOauth2CredentialProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}

		// Refresh from GET as oauth_discovery is not returned in CreateOauth2CredentialProviderOutput.
		got, err := findOAuth2CredentialProviderByName(ctx, conn, name)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}

		smerr.EnrichAppend(ctx, &response.Diagnostics,
			fwflex.Flatten(ctxWithCreds, got, &plan,
				fwflex.WithFieldNameSuffix("Output"),
			))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *oauth2CredentialProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data oauth2CredentialProviderResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
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
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *oauth2CredentialProviderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

func (r oauth2CredentialProviderResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if !request.Plan.Raw.IsNull() {
		var plan oauth2CredentialProviderResourceModel
		smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
		if response.Diagnostics.HasError() {
			return
		}

		vendor, diags := plan.VendorValue(ctx)
		smerr.EnrichAppend(ctx, &response.Diagnostics, diags)
		if response.Diagnostics.HasError() {
			return
		}
		var previousVendor attr.Value
		response.Plan.GetAttribute(ctx, path.Root("credential_provider_vendor"), &previousVendor)
		newVendorValue := fwtypes.StringEnumValue(awstypes.CredentialProviderVendorType(vendor))
		if !previousVendor.IsNull() && !previousVendor.IsUnknown() && !previousVendor.Equal(newVendorValue) {
			response.RequiresReplace = []path.Path{path.Root("config").AtListIndex(0)}
		}
		response.Plan.SetAttribute(ctx, path.Root("credential_provider_vendor"), newVendorValue)
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
		return nil, smarterr.NewError(&sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type oauth2CredsKey struct{}

type oauth2Creds struct {
	ClientID             types.String
	ClientIDWO           types.String
	ClientSecret         types.String
	ClientSecretWO       types.String
	ClientCredsWOVersion types.Int64
}

func withOAuth2Creds(ctx context.Context, c oauth2Creds) context.Context {
	return context.WithValue(ctx, oauth2CredsKey{}, c)
}

func oauth2CredsFrom(ctx context.Context) (oauth2Creds, bool) {
	v := ctx.Value(oauth2CredsKey{})
	c, ok := v.(oauth2Creds)
	return c, ok
}

type oauth2CredentialProviderResourceModel struct {
	framework.WithRegionModel
	ClientSecretARN          fwtypes.ListNestedObjectValueOf[secretModel]                    `tfsdk:"client_secret_arn"`
	CredentialProviderARN    types.String                                                    `tfsdk:"credential_provider_arn"`
	CredentialProviderVendor fwtypes.StringEnum[awstypes.CredentialProviderVendorType]       `tfsdk:"credential_provider_vendor"`
	Name                     types.String                                                    `tfsdk:"name"`
	OAuth2ProviderConfig     fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigInputModel] `tfsdk:"oauth2_provider_config"`
}

type oauth2ProviderConfigInputModel struct {
	Custom     fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"custom"`
	Github     fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"github"`
	Google     fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"google"`
	Microsoft  fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"microsoft"`
	Salesforce fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"salesforce"`
	Slack      fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"slack"`
}

func (m *oauth2CredentialProviderResourceModel) VendorValue(ctx context.Context) (string, diag.Diagnostics) {
	var vendor string

	if m.OAuth2ProviderConfig.IsNull() || m.OAuth2ProviderConfig.IsUnknown() {
		return vendor, nil
	}
	c, diags := m.OAuth2ProviderConfig.ToPtr(ctx)
	if diags.HasError() {
		return vendor, diags
	}

	switch {
	case !c.Custom.IsNull():
		vendor = string(awstypes.CredentialProviderVendorTypeCustomOauth2)
	case !c.Github.IsNull():
		vendor = string(awstypes.CredentialProviderVendorTypeGithubOauth2)
	case !c.Google.IsNull():
		vendor = string(awstypes.CredentialProviderVendorTypeGoogleOauth2)
	case !c.Microsoft.IsNull():
		vendor = string(awstypes.CredentialProviderVendorTypeMicrosoftOauth2)
	case !c.Salesforce.IsNull():
		vendor = string(awstypes.CredentialProviderVendorTypeSalesforceOauth2)
	case !c.Slack.IsNull():
		vendor = string(awstypes.CredentialProviderVendorTypeSlackOauth2)
	default:
		diags.AddError(
			"Invalid OAuth2 Provider Configuration",
			"At least one OAuth2 provider must be configured: custom, github, google, microsoft, salesforce, or slack",
		)
	}

	return vendor, nil
}

func (m *oauth2CredentialProviderResourceModel) CredsValue(ctx context.Context) (oauth2Creds, diag.Diagnostics) {
	if m.OAuth2ProviderConfig.IsNull() || m.OAuth2ProviderConfig.IsUnknown() {
		return oauth2Creds{}, nil
	}
	c, diags := m.OAuth2ProviderConfig.ToPtr(ctx)
	if diags.HasError() {
		return oauth2Creds{}, diags
	}

	var list fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel]
	switch {
	case !c.Custom.IsNull():
		list = c.Custom
	case !c.Github.IsNull():
		list = c.Github
	case !c.Google.IsNull():
		list = c.Google
	case !c.Microsoft.IsNull():
		list = c.Microsoft
	case !c.Salesforce.IsNull():
		list = c.Salesforce
	case !c.Slack.IsNull():
		list = c.Slack
	default:
		return oauth2Creds{}, nil
	}

	model, diags := list.ToPtr(ctx)
	if diags.HasError() || model == nil {
		return oauth2Creds{}, diags
	}

	return oauth2Creds{
		ClientID:             model.ClientID,
		ClientSecret:         model.ClientSecret,
		ClientIDWO:           model.ClientIDWo,
		ClientSecretWO:       model.ClientSecretWo,
		ClientCredsWOVersion: model.ClientCredsWoVersion,
	}, diags
}

var (
	_ fwflex.Flattener = &oauth2ProviderConfigInputModel{}
	_ fwflex.Expander  = &oauth2ProviderConfigInputModel{}
)

func (m *oauth2ProviderConfigInputModel) Flatten(ctxWithCreds context.Context, v any) (diags diag.Diagnostics) {
	var model oauth2ProviderConfigModel

	// We need to flatten "read-write" credentials and "write-only" credentials version as only those are stored in the state.
	// AutoFlex zeroes them because those are not returned by the AWS API.
	if creds, ok := oauth2CredsFrom(ctxWithCreds); ok {
		model.ClientID = creds.ClientID
		model.ClientSecret = creds.ClientSecret
		model.ClientCredsWoVersion = creds.ClientCredsWOVersion
	}

	switch t := v.(type) {
	case awstypes.Oauth2ProviderConfigOutputMemberCustomOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, fwflex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Custom = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberGithubOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, fwflex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Github = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberGoogleOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, fwflex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Google = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberMicrosoftOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, fwflex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Microsoft = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberSalesforceOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, fwflex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Salesforce = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberSlackOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, fwflex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Slack = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)
	}

	return diags
}

func (m oauth2ProviderConfigInputModel) Expand(ctxWithCreds context.Context) (result any, diags diag.Diagnostics) {
	var from *oauth2ProviderConfigModel
	var to any

	switch {
	case !m.Custom.IsNull():
		var r awstypes.Oauth2ProviderConfigInputMemberCustomOauth2ProviderConfig
		from, diags = m.Custom.ToPtr(ctxWithCreds)
		to = &r.Value
		result = &r
	case !m.Github.IsNull():
		var r awstypes.Oauth2ProviderConfigInputMemberGithubOauth2ProviderConfig
		from, diags = m.Github.ToPtr(ctxWithCreds)
		to = &r.Value
		result = &r
	case !m.Google.IsNull():
		var r awstypes.Oauth2ProviderConfigInputMemberGoogleOauth2ProviderConfig
		from, diags = m.Google.ToPtr(ctxWithCreds)
		to = &r.Value
		result = &r

	case !m.Microsoft.IsNull():
		var r awstypes.Oauth2ProviderConfigInputMemberMicrosoftOauth2ProviderConfig
		from, diags = m.Microsoft.ToPtr(ctxWithCreds)
		to = &r.Value
		result = &r

	case !m.Salesforce.IsNull():
		var r awstypes.Oauth2ProviderConfigInputMemberSalesforceOauth2ProviderConfig
		from, diags = m.Salesforce.ToPtr(ctxWithCreds)
		to = &r.Value
		result = &r

	case !m.Slack.IsNull():
		var r awstypes.Oauth2ProviderConfigInputMemberSlackOauth2ProviderConfig
		from, diags = m.Slack.ToPtr(ctxWithCreds)
		to = &r.Value
		result = &r
	default:
		diags.AddError(
			"Invalid OAuth2 Provider Configuration",
			"At least one OAuth2 provider must be configured: custom, github, google, microsoft, salesforce, or slack",
		)
		return nil, diags
	}

	smerr.EnrichAppend(ctxWithCreds, &diags, diags)
	if diags.HasError() {
		return nil, diags
	}

	if creds, ok := oauth2CredsFrom(ctxWithCreds); ok {
		if !creds.ClientIDWO.IsNull() && !creds.ClientSecretWO.IsNull() {
			from.ClientID = creds.ClientIDWO
			from.ClientSecret = creds.ClientSecretWO
		}
	}

	smerr.EnrichAppend(ctxWithCreds, &diags, fwflex.Expand(ctxWithCreds, from, to))
	if diags.HasError() {
		return nil, diags
	}
	return result, diags
}

type oauth2ProviderConfigModel struct {
	ClientCredsWoVersion types.Int64                                           `tfsdk:"client_credentials_wo_version"`
	ClientID             types.String                                          `tfsdk:"client_id"`
	ClientIDWo           types.String                                          `tfsdk:"client_id_wo"`
	ClientSecret         types.String                                          `tfsdk:"client_secret"`
	ClientSecretWo       types.String                                          `tfsdk:"client_secret_wo"`
	OauthDiscovery       fwtypes.ListNestedObjectValueOf[oauth2DiscoveryModel] `tfsdk:"oauth_discovery"`
}

type oauth2DiscoveryModel struct {
	AuthorizationServerMetadata fwtypes.ListNestedObjectValueOf[oauth2AuthorizationServerMetadataModel] `tfsdk:"authorization_server_metadata"`
	DiscoveryUrl                types.String                                                            `tfsdk:"discovery_url"`
}

var (
	_ fwflex.Flattener = &oauth2DiscoveryModel{}
	_ fwflex.Expander  = &oauth2DiscoveryModel{}
)

func (m *oauth2DiscoveryModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.Oauth2DiscoveryMemberDiscoveryUrl:
		m.DiscoveryUrl = types.StringValue(t.Value)

	case awstypes.Oauth2DiscoveryMemberAuthorizationServerMetadata:
		var model oauth2AuthorizationServerMetadataModel
		smerr.EnrichAppend(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.AuthorizationServerMetadata = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("oauth2 discovery configuration flatten: %s", reflect.TypeOf(v).String()),
		)
	}
	return diags
}

func (m oauth2DiscoveryModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.DiscoveryUrl.IsNull():
		var r awstypes.Oauth2DiscoveryMemberDiscoveryUrl
		r.Value = m.DiscoveryUrl.ValueString()
		return &r, diags

	case !m.AuthorizationServerMetadata.IsNull():
		model, d := m.AuthorizationServerMetadata.ToPtr(ctx)
		smerr.EnrichAppend(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.Oauth2DiscoveryMemberAuthorizationServerMetadata
		smerr.EnrichAppend(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	default:
		diags.AddError(
			"Invalid OAuth2 Discovery Configuration",
			"Either discovery_url or authorization_server_metadata must be configured",
		)
		return nil, diags
	}
}

type oauth2AuthorizationServerMetadataModel struct {
	AuthorizationEndpoint types.String        `tfsdk:"authorization_endpoint"`
	Issuer                types.String        `tfsdk:"issuer"`
	ResponseTypes         fwtypes.SetOfString `tfsdk:"response_types"`
	TokenEndpoint         types.String        `tfsdk:"token_endpoint"`
}
