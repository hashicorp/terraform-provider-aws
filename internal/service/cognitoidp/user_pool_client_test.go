package cognitoidp_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCognitoIDPUserPoolClient_basic(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "explicit_auth_flows.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_enableRevocation(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_revocation(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enable_token_revocation", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_revocation(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enable_token_revocation", "false"),
				),
			},
			{
				Config: testAccUserPoolClientConfig_revocation(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enable_token_revocation", "true"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_refreshTokenValidity(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_refreshTokenValidity(rName, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "60"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_refreshTokenValidity(rName, 120),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "120"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_accessTokenValidity(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_accessTokenValidity(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "access_token_validity", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_accessTokenValidity(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "access_token_validity", "1"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_idTokenValidity(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_idTokenValidity(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_idTokenValidity(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "1"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_tokenValidityUnits(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnits(rName, "days"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.access_token", "days"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.id_token", "days"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.refresh_token", "days"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnits(rName, "hours"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.access_token", "hours"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.id_token", "hours"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.refresh_token", "hours"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_tokenValidityUnitsWTokenValidity(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnitsTokenValidity(rName, "days"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.access_token", "days"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.id_token", "days"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.refresh_token", "days"),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnitsTokenValidity(rName, "hours"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.access_token", "hours"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.id_token", "hours"),
					resource.TestCheckResourceAttr(resourceName, "token_validity_units.0.refresh_token", "hours"),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "1"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_name(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_name(rName, "name1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_name(rName, "name2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", "name2"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_allFields(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_allFields(rName, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "explicit_auth_flows.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "CUSTOM_AUTH_FLOW_ONLY"),
					resource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "USER_PASSWORD_AUTH"),
					resource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr(resourceName, "generate_secret", "true"),
					resource.TestCheckResourceAttr(resourceName, "read_attributes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "read_attributes.*", "email"),
					resource.TestCheckResourceAttr(resourceName, "write_attributes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "write_attributes.*", "email"),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "300"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_flows.*", "code"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_flows.*", "implicit"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows_user_pool_client", "true"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_scopes.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "email"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "phone"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "aws.cognito.signin.user.admin"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "profile"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "callback_urls.*", "https://www.example.com/callback"),
					resource.TestCheckTypeSetElemAttr(resourceName, "callback_urls.*", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "default_redirect_uri", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "logout_urls.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "logout_urls.*", "https://www.example.com/login"),
					resource.TestCheckResourceAttr(resourceName, "prevent_user_existence_errors", "LEGACY"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"generate_secret"},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_allFieldsUpdatingOneField(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_allFields(rName, 300),
			},
			{
				Config: testAccUserPoolClientConfig_allFields(rName, 299),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "explicit_auth_flows.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "CUSTOM_AUTH_FLOW_ONLY"),
					resource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "USER_PASSWORD_AUTH"),
					resource.TestCheckTypeSetElemAttr(resourceName, "explicit_auth_flows.*", "ADMIN_NO_SRP_AUTH"),
					resource.TestCheckResourceAttr(resourceName, "generate_secret", "true"),
					resource.TestCheckResourceAttr(resourceName, "read_attributes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "read_attributes.*", "email"),
					resource.TestCheckResourceAttr(resourceName, "write_attributes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "write_attributes.*", "email"),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "299"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_flows.*", "code"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_flows.*", "implicit"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows_user_pool_client", "true"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_scopes.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "email"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "phone"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "aws.cognito.signin.user.admin"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_oauth_scopes.*", "profile"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "callback_urls.*", "https://www.example.com/callback"),
					resource.TestCheckTypeSetElemAttr(resourceName, "callback_urls.*", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "default_redirect_uri", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "logout_urls.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "logout_urls.*", "https://www.example.com/login"),
					resource.TestCheckResourceAttr(resourceName, "prevent_user_existence_errors", "LEGACY"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"generate_secret"},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_analytics(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"
	pinpointResourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckIdentityProvider(t)
			testAccPreCheckPinpointApp(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_analytics(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "analytics_configuration.0.application_id", pinpointResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.0.external_id", rName),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.0.user_data_shared", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "0"),
				),
			},
			{
				Config: testAccUserPoolClientConfig_analyticsShareData(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "analytics_configuration.0.application_id", pinpointResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.0.external_id", rName),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.0.user_data_shared", "true"),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_analyticsWithARN(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckIdentityProvider(t)
			testAccPreCheckPinpointApp(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_analyticsARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "analytics_configuration.0.application_arn", "aws_pinpoint_app.test", "arn"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "analytics_configuration.0.role_arn", "iam", "role/aws-service-role/cognito-idp.amazonaws.com/AWSServiceRoleForAmazonCognitoIdp"),
					resource.TestCheckResourceAttr(resourceName, "analytics_configuration.0.user_data_shared", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_disappears(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					acctest.CheckResourceDisappears(acctest.Provider, tfcognitoidp.ResourceUserPoolClient(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_Disappears_userPool(t *testing.T) {
	var client cognitoidentityprovider.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckIdentityProvider(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentityprovider.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserPoolClientDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(resourceName, &client),
					acctest.CheckResourceDisappears(acctest.Provider, tfcognitoidp.ResourceUserPool(), "aws_cognito_user_pool.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccUserPoolClientImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return "", errors.New("No Cognito User Pool Client ID set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn
		userPoolId := rs.Primary.Attributes["user_pool_id"]
		clientId := rs.Primary.ID

		_, err := tfcognitoidp.FindCognitoUserPoolClient(conn, userPoolId, clientId)

		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s/%s", userPoolId, clientId), nil
	}
}

func testAccCheckUserPoolClientDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_client" {
			continue
		}

		_, err := tfcognitoidp.FindCognitoUserPoolClient(conn, rs.Primary.Attributes["user_pool_id"], rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckUserPoolClientExists(name string, client *cognitoidentityprovider.UserPoolClientType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito User Pool Client ID set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPConn

		resp, err := tfcognitoidp.FindCognitoUserPoolClient(conn, rs.Primary.Attributes["user_pool_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		*client = *resp

		return nil
	}
}

func testAccUserPoolClientBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccUserPoolClientConfig_basic(rName string) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name                = %[1]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}
`, rName)
}

func testAccUserPoolClientConfig_revocation(rName string, revoke bool) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name                    = %[1]q
  user_pool_id            = aws_cognito_user_pool.test.id
  explicit_auth_flows     = ["ADMIN_NO_SRP_AUTH"]
  enable_token_revocation = %[2]t
}
`, rName, revoke)
}

func testAccUserPoolClientConfig_refreshTokenValidity(rName string, refreshTokenValidity int) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name                   = %[1]q
  refresh_token_validity = %[2]d
  user_pool_id           = aws_cognito_user_pool.test.id
}
`, rName, refreshTokenValidity)
}

func testAccUserPoolClientConfig_accessTokenValidity(rName string, validity int) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name                  = %[1]q
  access_token_validity = %[2]d
  user_pool_id          = aws_cognito_user_pool.test.id
}
`, rName, validity)
}

func testAccUserPoolClientConfig_idTokenValidity(rName string, validity int) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name              = %[1]q
  id_token_validity = %[2]d
  user_pool_id      = aws_cognito_user_pool.test.id
}
`, rName, validity)
}

func testAccUserPoolClientConfig_tokenValidityUnits(rName, units string) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  token_validity_units {
    access_token  = %[2]q
    id_token      = %[2]q
    refresh_token = %[2]q
  }
}
`, rName, units)
}

func testAccUserPoolClientConfig_tokenValidityUnitsTokenValidity(rName, units string) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name              = %[1]q
  user_pool_id      = aws_cognito_user_pool.test.id
  id_token_validity = 1

  token_validity_units {
    access_token  = %[2]q
    id_token      = %[2]q
    refresh_token = %[2]q
  }
}
`, rName, units)
}

func testAccUserPoolClientConfig_name(rName, name string) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}
`, name)
}

func testAccUserPoolClientConfig_allFields(rName string, refreshTokenValidity int) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name = %[1]q

  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH", "CUSTOM_AUTH_FLOW_ONLY", "USER_PASSWORD_AUTH"]

  generate_secret = "true"

  read_attributes  = ["email"]
  write_attributes = ["email"]

  refresh_token_validity        = %[2]d
  prevent_user_existence_errors = "LEGACY"

  allowed_oauth_flows                  = ["code", "implicit"]
  allowed_oauth_flows_user_pool_client = "true"
  allowed_oauth_scopes                 = ["phone", "email", "openid", "profile", "aws.cognito.signin.user.admin"]

  callback_urls        = ["https://www.example.com/redirect", "https://www.example.com/callback"]
  default_redirect_uri = "https://www.example.com/redirect"
  logout_urls          = ["https://www.example.com/login"]
}
`, rName, refreshTokenValidity)
}

func testAccUserPoolClientAnalyticsBaseConfig(rName string) string {
	return testAccUserPoolClientBaseConfig(rName) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_pinpoint_app" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "cognito-idp.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<-EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "mobiletargeting:UpdateEndpoint",
        "mobiletargeting:PutItems"
      ],
      "Effect": "Allow",
      "Resource": "arn:${data.aws_partition.current.partition}:mobiletargeting:*:${data.aws_caller_identity.current.account_id}:apps/${aws_pinpoint_app.test.application_id}*"
    }
  ]
}
EOF
}
`, rName)
}

func testAccUserPoolClientConfig_analytics(rName string) string {
	return testAccUserPoolClientAnalyticsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_id = aws_pinpoint_app.test.application_id
    external_id    = %[1]q
    role_arn       = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccUserPoolClientConfig_analyticsShareData(rName string) string {
	return testAccUserPoolClientAnalyticsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_id   = aws_pinpoint_app.test.application_id
    external_id      = %[1]q
    role_arn         = aws_iam_role.test.arn
    user_data_shared = true
  }
}
`, rName)
}

func testAccUserPoolClientConfig_analyticsARN(rName string) string {
	return testAccUserPoolClientAnalyticsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_arn = aws_pinpoint_app.test.arn
  }
}
`, rName)
}

func testAccPreCheckPinpointApp(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

	input := &pinpoint.GetAppsInput{}

	_, err := conn.GetApps(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
