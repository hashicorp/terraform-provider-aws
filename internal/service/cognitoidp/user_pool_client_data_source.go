// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cognito_user_pool_client", name="User Pool Client")
func dataSourceUserPoolClient() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserPoolClientRead,

		Schema: map[string]*schema.Schema{
			"access_token_validity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"allowed_oauth_flows": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"allowed_oauth_flows_user_pool_client": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"allowed_oauth_scopes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"analytics_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrApplicationID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"application_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrExternalID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrRoleARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_data_shared": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"callback_urls": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrClientID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrClientSecret: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"default_redirect_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_token_revocation": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_propagate_additional_user_context_data": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"explicit_auth_flows": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"generate_secret": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"id_token_validity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"logout_urls": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prevent_user_existence_errors": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"read_attributes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"refresh_token_validity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"supported_identity_providers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"token_validity_units": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_token": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id_token": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"refresh_token": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrUserPoolID: {
				Type:     schema.TypeString,
				Required: true,
			},
			"write_attributes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceUserPoolClientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	clientID := d.Get(names.AttrClientID).(string)
	userPoolClient, err := findUserPoolClientByTwoPartKey(ctx, conn, d.Get(names.AttrUserPoolID).(string), clientID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito User Pool Client (%s): %s", clientID, err)
	}

	d.SetId(clientID)
	d.Set("access_token_validity", userPoolClient.AccessTokenValidity)
	d.Set("allowed_oauth_flows", userPoolClient.AllowedOAuthFlows)
	d.Set("allowed_oauth_flows_user_pool_client", userPoolClient.AllowedOAuthFlowsUserPoolClient)
	d.Set("allowed_oauth_scopes", userPoolClient.AllowedOAuthScopes)
	if err := d.Set("analytics_configuration", flattenUserPoolClientAnalyticsConfig(userPoolClient.AnalyticsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting analytics_configuration: %s", err)
	}
	d.Set("callback_urls", userPoolClient.CallbackURLs)
	d.Set(names.AttrClientSecret, userPoolClient.ClientSecret)
	d.Set("default_redirect_uri", userPoolClient.DefaultRedirectURI)
	d.Set("enable_propagate_additional_user_context_data", userPoolClient.EnablePropagateAdditionalUserContextData)
	d.Set("enable_token_revocation", userPoolClient.EnableTokenRevocation)
	d.Set("explicit_auth_flows", userPoolClient.ExplicitAuthFlows)
	d.Set("id_token_validity", userPoolClient.IdTokenValidity)
	d.Set("logout_urls", userPoolClient.LogoutURLs)
	d.Set(names.AttrName, userPoolClient.ClientName)
	d.Set("prevent_user_existence_errors", userPoolClient.PreventUserExistenceErrors)
	d.Set("read_attributes", userPoolClient.ReadAttributes)
	d.Set("refresh_token_validity", userPoolClient.RefreshTokenValidity)
	d.Set("supported_identity_providers", userPoolClient.SupportedIdentityProviders)
	if err := d.Set("token_validity_units", flattenUserPoolClientTokenValidityUnitsType(userPoolClient.TokenValidityUnits)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting token_validity_units: %s", err)
	}

	d.Set(names.AttrUserPoolID, userPoolClient.UserPoolId)
	d.Set("write_attributes", userPoolClient.WriteAttributes)

	return diags
}

func flattenUserPoolClientAnalyticsConfig(apiObject *awstypes.AnalyticsConfigurationType) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"user_data_shared": apiObject.UserDataShared,
	}

	if apiObject.ApplicationArn != nil {
		tfMap["application_arn"] = aws.ToString(apiObject.ApplicationArn)
	}

	if apiObject.ApplicationId != nil {
		tfMap[names.AttrApplicationID] = aws.ToString(apiObject.ApplicationId)
	}

	if apiObject.ExternalId != nil {
		tfMap[names.AttrExternalID] = aws.ToString(apiObject.ExternalId)
	}

	if apiObject.RoleArn != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(apiObject.RoleArn)
	}

	return []interface{}{tfMap}
}

func flattenUserPoolClientTokenValidityUnitsType(apiObject *awstypes.TokenValidityUnitsType) []interface{} {
	if apiObject == nil {
		return nil
	}

	//tokenValidityConfig is never nil and if everything is empty it causes diffs
	if apiObject.IdToken == "" && apiObject.AccessToken == "" && apiObject.RefreshToken == "" {
		return nil
	}

	tfMap := map[string]interface{}{
		"access_token":  apiObject.AccessToken,
		"id_token":      apiObject.IdToken,
		"refresh_token": apiObject.RefreshToken,
	}

	return []interface{}{tfMap}
}
