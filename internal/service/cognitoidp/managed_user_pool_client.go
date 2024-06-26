// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Managed User Pool Client")
func newManagedUserPoolClientResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &managedUserPoolClientResource{}, nil
}

type managedUserPoolClientResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
	userPoolClientResourceWithImport
	userPoolClientResourceWithConfigValidators
}

func (*managedUserPoolClientResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cognito_managed_user_pool_client"
}

// Schema returns the schema for this resource.
func (r *managedUserPoolClientResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := userPoolClientResourceSchema(ctx)

	// Overwrite the "name" attribute.
	s.Attributes[names.AttrName] = schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	// Additional attributes.
	s.Attributes["name_pattern"] = schema.StringAttribute{
		CustomType: fwtypes.RegexpType,
		Optional:   true,
		Validators: append(
			userPoolClientNameValidator,
			stringvalidator.ExactlyOneOf(
				path.MatchRelative().AtParent().AtName(names.AttrNamePrefix),
				path.MatchRelative().AtParent().AtName("name_pattern"),
			),
		),
	}
	s.Attributes[names.AttrNamePrefix] = schema.StringAttribute{
		Optional:   true,
		Validators: userPoolClientNameValidator,
	}

	response.Schema = s
}

func (r *managedUserPoolClientResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data managedUserPoolClientResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	filter := tfslices.PredicateTrue[*awstypes.UserPoolClientDescription]()
	if namePattern := data.NamePattern; !namePattern.IsUnknown() && !namePattern.IsNull() {
		filter = func(v *awstypes.UserPoolClientDescription) bool {
			return namePattern.ValueRegexp().MatchString(aws.ToString(v.ClientName))
		}
	}
	if namePrefix := data.NamePrefix; !namePrefix.IsUnknown() && !namePrefix.IsNull() {
		filter = func(v *awstypes.UserPoolClientDescription) bool {
			return strings.HasPrefix(aws.ToString(v.ClientName), namePrefix.ValueString())
		}
	}
	userPoolID := data.UserPoolID.ValueString()

	output, err := findUserPoolClientByName(ctx, conn, userPoolID, filter)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito User Pool Client (%s)", userPoolID), err.Error())

		return
	}

	var current managedUserPoolClientResourceModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &current)...)
	if response.Diagnostics.HasError() {
		return
	}

	needsUpdate := false

	if !data.AccessTokenValidity.IsUnknown() && !data.AccessTokenValidity.Equal(current.AccessTokenValidity) {
		needsUpdate = true
		current.AccessTokenValidity = data.AccessTokenValidity
	}
	if !data.AllowedOAuthFlows.IsUnknown() && !data.AllowedOAuthFlows.Equal(current.AllowedOAuthFlows) {
		needsUpdate = true
		current.AllowedOAuthFlows = data.AllowedOAuthFlows
	}
	if !data.AllowedOAuthFlowsUserPoolClient.IsUnknown() && !data.AllowedOAuthFlowsUserPoolClient.Equal(current.AllowedOAuthFlowsUserPoolClient) {
		needsUpdate = true
		current.AllowedOAuthFlowsUserPoolClient = data.AllowedOAuthFlowsUserPoolClient
	}
	if !data.AllowedOAuthScopes.IsUnknown() && !data.AllowedOAuthScopes.Equal(current.AllowedOAuthScopes) {
		needsUpdate = true
		current.AllowedOAuthScopes = data.AllowedOAuthScopes
	}
	if !data.AnalyticsConfiguration.IsUnknown() && !data.AnalyticsConfiguration.Equal(current.AnalyticsConfiguration) {
		needsUpdate = true
		current.AnalyticsConfiguration = data.AnalyticsConfiguration
	}
	if !data.AuthSessionValidity.IsUnknown() && !data.AuthSessionValidity.Equal(current.AuthSessionValidity) {
		needsUpdate = true
		current.AuthSessionValidity = data.AuthSessionValidity
	}
	if !data.CallbackURLs.IsUnknown() && !data.CallbackURLs.Equal(current.CallbackURLs) {
		needsUpdate = true
		current.CallbackURLs = data.CallbackURLs
	}
	if !data.DefaultRedirectURI.IsUnknown() && !data.DefaultRedirectURI.Equal(current.DefaultRedirectURI) {
		needsUpdate = true
		current.DefaultRedirectURI = data.DefaultRedirectURI
	}
	if !data.EnablePropagateAdditionalUserContextData.IsUnknown() && !data.EnablePropagateAdditionalUserContextData.Equal(current.EnablePropagateAdditionalUserContextData) {
		needsUpdate = true
		current.EnablePropagateAdditionalUserContextData = data.EnablePropagateAdditionalUserContextData
	}
	if !data.EnableTokenRevocation.IsUnknown() && !data.EnableTokenRevocation.Equal(current.EnableTokenRevocation) {
		needsUpdate = true
		current.EnableTokenRevocation = data.EnableTokenRevocation
	}
	if !data.ExplicitAuthFlows.IsUnknown() && !data.ExplicitAuthFlows.Equal(current.ExplicitAuthFlows) {
		needsUpdate = true
		current.ExplicitAuthFlows = data.ExplicitAuthFlows
	}
	if !data.IDTokenValidity.IsUnknown() && !data.IDTokenValidity.Equal(current.IDTokenValidity) {
		needsUpdate = true
		current.IDTokenValidity = data.IDTokenValidity
	}
	if !data.LogoutURLs.IsUnknown() && !data.LogoutURLs.Equal(current.LogoutURLs) {
		needsUpdate = true
		current.LogoutURLs = data.LogoutURLs
	}
	if !data.PreventUserExistenceErrors.IsUnknown() && !data.PreventUserExistenceErrors.Equal(current.PreventUserExistenceErrors) {
		needsUpdate = true
		current.PreventUserExistenceErrors = data.PreventUserExistenceErrors
	}
	if !data.ReadAttributes.IsUnknown() && !data.ReadAttributes.Equal(current.ReadAttributes) {
		needsUpdate = true
		current.ReadAttributes = data.ReadAttributes
	}
	if !data.RefreshTokenValidity.IsUnknown() && !data.RefreshTokenValidity.Equal(current.RefreshTokenValidity) {
		needsUpdate = true
		current.RefreshTokenValidity = data.RefreshTokenValidity
	}
	if !data.SupportedIdentityProviders.IsUnknown() && !data.SupportedIdentityProviders.Equal(current.SupportedIdentityProviders) {
		needsUpdate = true
		current.SupportedIdentityProviders = data.SupportedIdentityProviders
	}
	if !data.TokenValidityUnits.IsUnknown() && !data.TokenValidityUnits.Equal(current.TokenValidityUnits) {
		needsUpdate = true
		current.TokenValidityUnits = data.TokenValidityUnits
	}
	if !data.WriteAttributes.IsUnknown() && !data.WriteAttributes.Equal(current.WriteAttributes) {
		needsUpdate = true
		current.WriteAttributes = data.WriteAttributes
	}

	if needsUpdate {
		input := &cognitoidentityprovider.UpdateUserPoolClientInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, current, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		const (
			timeout = 2 * time.Minute
		)
		outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ConcurrentModificationException](ctx, timeout, func() (interface{}, error) {
			return conn.UpdateUserPoolClient(ctx, input)
		})

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Cognito Managed User Pool Client (%s)", current.ClientID.ValueString()), err.Error())

			return
		}

		// Set values for unknowns.
		response.Diagnostics.Append(fwflex.Flatten(ctx, outputRaw.(*cognitoidentityprovider.UpdateUserPoolClientOutput).UserPoolClient, &current)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &current)...)
}

func (r *managedUserPoolClientResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data managedUserPoolClientResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	output, err := findUserPoolClientByTwoPartKey(ctx, conn, data.UserPoolID.ValueString(), data.ClientID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito Managed User Pool Client (%s)", data.ClientID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	// if isDefaultTokenValidityUnits(output.TokenValidityUnits) {
	// 	output.TokenValidityUnits = nil
	// }
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *managedUserPoolClientResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new managedUserPoolClientResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	input := &cognitoidentityprovider.UpdateUserPoolClientInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// If removing `token_validity_units`, reset to defaults.
	// if !old.TokenValidityUnits.IsNull() && new.TokenValidityUnits.IsNull() {
	// 	input.TokenValidityUnits.AccessToken = awstypes.TimeUnitsTypeHours
	// 	input.TokenValidityUnits.IdToken = awstypes.TimeUnitsTypeHours
	// 	input.TokenValidityUnits.RefreshToken = awstypes.TimeUnitsTypeDays
	// }

	const (
		timeout = 2 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ConcurrentModificationException](ctx, timeout, func() (interface{}, error) {
		return conn.UpdateUserPoolClient(ctx, input)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Managed Cognito User Pool Client (%s)", new.ClientID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputRaw.(*cognitoidentityprovider.UpdateUserPoolClientOutput).UserPoolClient, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func findUserPoolClientByName(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID string, filter tfslices.Predicate[*awstypes.UserPoolClientDescription]) (*awstypes.UserPoolClientType, error) {
	input := &cognitoidentityprovider.ListUserPoolClientsInput{
		UserPoolId: aws.String(userPoolID),
	}
	var userPoolClients []awstypes.UserPoolClientDescription

	pages := cognitoidentityprovider.NewListUserPoolClientsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.UserPoolClients {
			if filter(&v) {
				userPoolClients = append(userPoolClients, v)
			}
		}
	}

	userPoolClient, err := tfresource.AssertSingleValueResult(userPoolClients)

	if err != nil {
		return nil, err
	}

	return findUserPoolClientByTwoPartKey(ctx, conn, userPoolID, aws.ToString(userPoolClient.ClientId))
}

type managedUserPoolClientResourceModel struct {
	userPoolClientResourceModel

	NamePattern fwtypes.Regexp `tfsdk:"name_pattern"`
	NamePrefix  types.String   `tfsdk:"name_prefix"`
}
