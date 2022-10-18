package cognitoidp

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceUserPoolClient() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserPoolClientRead,

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
						"application_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"application_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_arn": {
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
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_secret": {
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
			"name": {
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
			"user_pool_id": {
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

func dataSourceUserPoolClientRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	clientId := d.Get("client_id").(string)
	d.SetId(clientId)

	userPoolClient, err := FindCognitoUserPoolClient(conn, d.Get("user_pool_id").(string), d.Id())

	if err != nil {
		return fmt.Errorf("error reading Cognito User Pool Client (%s): %w", clientId, err)
	}

	d.Set("user_pool_id", userPoolClient.UserPoolId)
	d.Set("name", userPoolClient.ClientName)
	d.Set("explicit_auth_flows", flex.FlattenStringSet(userPoolClient.ExplicitAuthFlows))
	d.Set("read_attributes", flex.FlattenStringSet(userPoolClient.ReadAttributes))
	d.Set("write_attributes", flex.FlattenStringSet(userPoolClient.WriteAttributes))
	d.Set("refresh_token_validity", userPoolClient.RefreshTokenValidity)
	d.Set("access_token_validity", userPoolClient.AccessTokenValidity)
	d.Set("id_token_validity", userPoolClient.IdTokenValidity)
	d.Set("client_secret", userPoolClient.ClientSecret)
	d.Set("allowed_oauth_flows", flex.FlattenStringSet(userPoolClient.AllowedOAuthFlows))
	d.Set("allowed_oauth_flows_user_pool_client", userPoolClient.AllowedOAuthFlowsUserPoolClient)
	d.Set("allowed_oauth_scopes", flex.FlattenStringSet(userPoolClient.AllowedOAuthScopes))
	d.Set("callback_urls", flex.FlattenStringSet(userPoolClient.CallbackURLs))
	d.Set("default_redirect_uri", userPoolClient.DefaultRedirectURI)
	d.Set("logout_urls", flex.FlattenStringSet(userPoolClient.LogoutURLs))
	d.Set("prevent_user_existence_errors", userPoolClient.PreventUserExistenceErrors)
	d.Set("supported_identity_providers", flex.FlattenStringSet(userPoolClient.SupportedIdentityProviders))
	d.Set("enable_token_revocation", userPoolClient.EnableTokenRevocation)
	d.Set("enable_propagate_additional_user_context_data", userPoolClient.EnablePropagateAdditionalUserContextData)

	if err := d.Set("analytics_configuration", flattenUserPoolClientAnalyticsConfig(userPoolClient.AnalyticsConfiguration)); err != nil {
		return fmt.Errorf("error setting analytics_configuration: %w", err)
	}

	if err := d.Set("token_validity_units", flattenUserPoolClientTokenValidityUnitsType(userPoolClient.TokenValidityUnits)); err != nil {
		return fmt.Errorf("error setting token_validity_units: %w", err)
	}

	return nil
}
