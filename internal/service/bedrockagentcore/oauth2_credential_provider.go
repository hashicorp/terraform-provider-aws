// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_oauth2_credential_provider", name="OAuth2 Credential Provider")
func newResourceOAuth2CredentialProvider(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceOauth2CredentialProvider{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameOAuth2CredentialProvider = "OAuth2 Credential Provider"
)

type resourceOauth2CredentialProvider struct {
	framework.ResourceWithModel[resourceOauth2CredentialProviderModel]
	framework.WithTimeouts
}

func clientCredentialAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
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
		"client_credentials_wo_version": schema.Int64Attribute{
			Optional: true,
			Validators: []validator.Int64{
				int64validator.AlsoRequires(path.Expressions{
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

func (r *resourceOauth2CredentialProvider) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	vendorType := fwtypes.StringEnumType[awstypes.CredentialProviderVendorType]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"client_secret_arn": schema.StringAttribute{
				Computed: true,
			},
			"vendor": schema.StringAttribute{
				Computed:    true,
				CustomType:  vendorType,
				Description: "The credential provider vendor type, automatically determined from config block type",
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
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
															names.AttrIssuer: schema.StringAttribute{
																Required: true,
															},
															"authorization_endpoint": schema.StringAttribute{
																Required: true,
															},
															"token_endpoint": schema.StringAttribute{
																Required: true,
															},
															"response_types": schema.SetAttribute{
																CustomType:  fwtypes.SetOfStringType,
																ElementType: types.StringType,
																Optional:    true,
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r resourceOauth2CredentialProvider) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.Plan.Raw.IsNull() {
		var plan resourceOauth2CredentialProviderModel
		smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
		if resp.Diagnostics.HasError() {
			return
		}

		vendor, diags := plan.VendorValue(ctx)
		smerr.EnrichAppend(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}
		var previousVendor attr.Value
		resp.Plan.GetAttribute(ctx, path.Root("vendor"), &previousVendor)
		newVendorValue := fwtypes.StringEnumValue(awstypes.CredentialProviderVendorType(vendor))
		if !previousVendor.IsNull() && !previousVendor.IsUnknown() && !previousVendor.Equal(newVendorValue) {
			resp.RequiresReplace = []path.Path{path.Root("config").AtListIndex(0)}
		}
		resp.Plan.SetAttribute(ctx, path.Root("vendor"), newVendorValue)
	}
}

func (r *resourceOauth2CredentialProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourceOauth2CredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract credentials from the raw configuration because write‑only
	// attributes are not present in plan or state.
	var configModel resourceOauth2CredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &configModel))
	if resp.Diagnostics.HasError() {
		return
	}
	creds, d := configModel.CredsValue(ctx)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	ctxWithCreds := withOAuth2Creds(ctx, creds)

	var input bedrockagentcorecontrol.CreateOauth2CredentialProviderInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics,
		flex.Expand(ctxWithCreds, plan, &input,
			flex.WithFieldNamePrefix("CredentialProvider"),
			flex.WithFieldNameSuffix("Input"),
		))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateOauth2CredentialProvider(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	// Refresh from GET as oauth_discovery is not returned in CreateOauth2CredentialProviderOutput.
	got, err := findOAuth2CredentialProviderByName(ctx, conn, plan.Name.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics,
		flex.Flatten(ctxWithCreds, got, &plan,
			flex.WithFieldNamePrefix("CredentialProvider"),
			flex.WithFieldNameSuffix("Output"),
		))
	if resp.Diagnostics.HasError() {
		return
	}

	if got.ClientSecretArn != nil && got.ClientSecretArn.SecretArn != nil {
		plan.ClientSecretArn = types.StringValue(*got.ClientSecretArn.SecretArn)
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceOauth2CredentialProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceOauth2CredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findOAuth2CredentialProviderByName(ctx, conn, state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	// TODO Remove before AgentCore GA
	// Store clientId/clientSecret in context before flattening zeroes them
	// This won't be necessary in AgentCore GA - AWS API already returns those values but awstypes don't
	creds, d := state.CredsValue(ctx)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	ctxWithCreds := withOAuth2Creds(ctx, creds)

	smerr.EnrichAppend(ctx, &resp.Diagnostics,
		flex.Flatten(ctxWithCreds, out, &state,
			flex.WithFieldNamePrefix("CredentialProvider"),
			flex.WithFieldNameSuffix("Output")))
	if resp.Diagnostics.HasError() {
		return
	}

	if out.ClientSecretArn != nil && out.ClientSecretArn.SecretArn != nil {
		state.ClientSecretArn = types.StringValue(*out.ClientSecretArn.SecretArn)
	}
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceOauth2CredentialProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state resourceOauth2CredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract credentials from the raw configuration because write‑only
	// attributes are not present in plan or state.
	var configModel resourceOauth2CredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &configModel))
	if resp.Diagnostics.HasError() {
		return
	}
	credsConfig, d := configModel.CredsValue(ctx)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	ctxWithCreds := withOAuth2Creds(ctx, credsConfig)

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateOauth2CredentialProviderInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics,
			flex.Expand(ctxWithCreds, plan, &input,
				flex.WithFieldNamePrefix("CredentialProvider"),
				flex.WithFieldNameSuffix("Input")))
		if resp.Diagnostics.HasError() {
			return
		}
		out, err := conn.UpdateOauth2CredentialProvider(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		// Refresh from GET as oauth_discovery is not returned in CreateOauth2CredentialProviderOutput.
		got, err := findOAuth2CredentialProviderByName(ctx, conn, plan.Name.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics,
			flex.Flatten(ctxWithCreds, got, &plan,
				flex.WithFieldNamePrefix("CredentialProvider"),
				flex.WithFieldNameSuffix("Output"),
			))
		if resp.Diagnostics.HasError() {
			return
		}

		if got.ClientSecretArn != nil && got.ClientSecretArn.SecretArn != nil {
			plan.ClientSecretArn = types.StringValue(*got.ClientSecretArn.SecretArn)
		}
	}
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceOauth2CredentialProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceOauth2CredentialProviderModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteOauth2CredentialProviderInput{
		Name: state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteOauth2CredentialProvider(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *resourceOauth2CredentialProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), req, resp)
}

func findOAuth2CredentialProviderByName(ctx context.Context, conn *bedrockagentcorecontrol.Client, name string) (*bedrockagentcorecontrol.GetOauth2CredentialProviderOutput, error) {
	input := bedrockagentcorecontrol.GetOauth2CredentialProviderInput{
		Name: aws.String(name),
	}

	out, err := conn.GetOauth2CredentialProvider(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type oauth2CredsKey struct{}

type oauth2Creds struct {
	ClientId             types.String
	ClientSecret         types.String
	ClientIdWo           types.String
	ClientSecretWo       types.String
	ClientCredsWoVersion types.Int64
}

func withOAuth2Creds(ctx context.Context, c oauth2Creds) context.Context {
	return context.WithValue(ctx, oauth2CredsKey{}, c)
}

func oauth2CredsFrom(ctx context.Context) (oauth2Creds, bool) {
	v := ctx.Value(oauth2CredsKey{})
	c, ok := v.(oauth2Creds)
	return c, ok
}

type resourceOauth2CredentialProviderModel struct {
	framework.WithRegionModel
	ARN                  types.String                                                    `tfsdk:"arn"`
	ClientSecretArn      types.String                                                    `tfsdk:"client_secret_arn" autoflex:"-"`
	Vendor               fwtypes.StringEnum[awstypes.CredentialProviderVendorType]       `tfsdk:"vendor"`
	Name                 types.String                                                    `tfsdk:"name"`
	Oauth2ProviderConfig fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigInputModel] `tfsdk:"config"`
	Timeouts             timeouts.Value                                                  `tfsdk:"timeouts"`
}

type oauth2ProviderConfigInputModel struct {
	Custom     fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"custom"`
	Github     fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"github"`
	Google     fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"google"`
	Microsoft  fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"microsoft"`
	Salesforce fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"salesforce"`
	Slack      fwtypes.ListNestedObjectValueOf[oauth2ProviderConfigModel] `tfsdk:"slack"`
}

func (m *resourceOauth2CredentialProviderModel) VendorValue(ctx context.Context) (string, diag.Diagnostics) {
	var vendor string

	if m.Oauth2ProviderConfig.IsNull() || m.Oauth2ProviderConfig.IsUnknown() {
		return vendor, nil
	}
	c, diags := m.Oauth2ProviderConfig.ToPtr(ctx)
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

func (m *resourceOauth2CredentialProviderModel) CredsValue(ctx context.Context) (oauth2Creds, diag.Diagnostics) {
	if m.Oauth2ProviderConfig.IsNull() || m.Oauth2ProviderConfig.IsUnknown() {
		return oauth2Creds{}, nil
	}
	c, diags := m.Oauth2ProviderConfig.ToPtr(ctx)
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
		ClientId:             model.ClientId,
		ClientSecret:         model.ClientSecret,
		ClientIdWo:           model.ClientIdWo,
		ClientSecretWo:       model.ClientSecretWo,
		ClientCredsWoVersion: model.ClientCredsWoVersion,
	}, diags
}

type oauth2ProviderConfigModel struct {
	ClientId             types.String                                          `tfsdk:"client_id"`
	ClientSecret         types.String                                          `tfsdk:"client_secret"`
	ClientIdWo           types.String                                          `tfsdk:"client_id_wo"`
	ClientSecretWo       types.String                                          `tfsdk:"client_secret_wo"`
	ClientCredsWoVersion types.Int64                                           `tfsdk:"client_credentials_wo_version"`
	OauthDiscovery       fwtypes.ListNestedObjectValueOf[oauth2DiscoveryModel] `tfsdk:"oauth_discovery"`
}

type oauth2DiscoveryModel struct {
	DiscoveryUrl                types.String                                                            `tfsdk:"discovery_url"`
	AuthorizationServerMetadata fwtypes.ListNestedObjectValueOf[oauth2AuthorizationServerMetadataModel] `tfsdk:"authorization_server_metadata"`
}

type oauth2AuthorizationServerMetadataModel struct {
	Issuer                types.String        `tfsdk:"issuer"`
	AuthorizationEndpoint types.String        `tfsdk:"authorization_endpoint"`
	TokenEndpoint         types.String        `tfsdk:"token_endpoint"`
	ResponseTypes         fwtypes.SetOfString `tfsdk:"response_types"`
}

func (m *oauth2ProviderConfigInputModel) Flatten(ctxWithCreds context.Context, v any) (diags diag.Diagnostics) {
	var model oauth2ProviderConfigModel

	// We need to flatten "read-write" credentials and "write-only" credentials version as only those are stored in the state.
	// AutoFlex zeroes them because those are not returned by the AWS API.
	if creds, ok := oauth2CredsFrom(ctxWithCreds); ok {
		model.ClientId = creds.ClientId
		model.ClientSecret = creds.ClientSecret
		model.ClientCredsWoVersion = creds.ClientCredsWoVersion
	}

	switch t := v.(type) {
	case awstypes.Oauth2ProviderConfigOutputMemberCustomOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, flex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Custom = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberGithubOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, flex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Github = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberGoogleOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, flex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Google = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberMicrosoftOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, flex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Microsoft = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberSalesforceOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, flex.Flatten(ctxWithCreds, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.Salesforce = fwtypes.NewListNestedObjectValueOfPtrMust(ctxWithCreds, &model)

	case awstypes.Oauth2ProviderConfigOutputMemberSlackOauth2ProviderConfig:
		smerr.EnrichAppend(ctxWithCreds, &diags, flex.Flatten(ctxWithCreds, t.Value, &model))
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
		if !creds.ClientIdWo.IsNull() && !creds.ClientSecretWo.IsNull() {
			from.ClientId = creds.ClientIdWo
			from.ClientSecret = creds.ClientSecretWo
		}
	}

	smerr.EnrichAppend(ctxWithCreds, &diags, flex.Expand(ctxWithCreds, from, to))
	if diags.HasError() {
		return nil, diags
	}
	return result, diags
}

func (m *oauth2DiscoveryModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.Oauth2DiscoveryMemberDiscoveryUrl:
		m.DiscoveryUrl = types.StringValue(t.Value)
		return diags

	case awstypes.Oauth2DiscoveryMemberAuthorizationServerMetadata:
		var model oauth2AuthorizationServerMetadataModel
		smerr.EnrichAppend(ctx, &diags, flex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.AuthorizationServerMetadata = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
		return diags

	default:
		return diags
	}
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
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, model, &r.Value))
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
