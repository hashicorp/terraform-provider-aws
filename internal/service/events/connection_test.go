package events_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEventsConnection_apiKey(t *testing.T) {
	var v1, v2, v3 eventbridge.DescribeConnectionOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authorizationType := "API_KEY"
	description := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	nameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	descriptionModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	valueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudwatch_event_connection.api_key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_apiKey(
					name,
					description,
					authorizationType,
					key,
					value,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", key),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_parameters.0.api_key.0.value"},
			},
			{
				Config: testAccConnectionConfig_apiKey(
					nameModified,
					descriptionModified,
					authorizationType,
					keyModified,
					valueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v2),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("connection/%s/%s", nameModified, uuidRegex))),
					testAccCheckConnectionRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", keyModified),
				),
			},
			{
				Config: testAccConnectionConfig_apiKey(
					nameModified,
					descriptionModified,
					authorizationType,
					keyModified,
					valueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v3),
					testAccCheckConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", keyModified),
				),
			},
		},
	})
}

func TestAccEventsConnection_basic(t *testing.T) {
	var v1, v2, v3 eventbridge.DescribeConnectionOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authorizationType := "BASIC"
	description := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	username := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	password := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	nameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	descriptionModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	usernameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	passwordModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudwatch_event_connection.basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(
					name,
					description,
					authorizationType,
					username,
					password,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.username", username),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_parameters.0.basic.0.password"},
			},
			{
				Config: testAccConnectionConfig_basic(
					nameModified,
					descriptionModified,
					authorizationType,
					usernameModified,
					passwordModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v2),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("connection/%s/%s", nameModified, uuidRegex))),
					testAccCheckConnectionRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.username", usernameModified),
				),
			},
			{
				Config: testAccConnectionConfig_basic(
					nameModified,
					descriptionModified,
					authorizationType,
					usernameModified,
					passwordModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v3),
					testAccCheckConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.username", usernameModified),
				),
			},
		},
	})
}

func TestAccEventsConnection_oAuth(t *testing.T) {
	var v1, v2, v3 eventbridge.DescribeConnectionOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authorizationType := "OAUTH_CLIENT_CREDENTIALS"
	description := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// oauth
	authorizationEndpoint := "https://example.com/auth"
	httpMethod := "POST"

	// client_parameters
	clientID := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clientSecret := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// oauth_http_parameters
	bodyKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bodyValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bodyIsSecretValue := true

	headerKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerIsSecretValue := true

	queryStringKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringIsSecretValue := true

	// modified
	nameModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	descriptionModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// oauth
	authorizationEndpointModified := "https://example.com/auth-modified"
	httpMethodModified := "GET"

	// client_parameters
	clientIDModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clientSecretModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// oauth_http_parameters modified
	bodyKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bodyValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bodyIsSecretValueModified := false

	headerKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerIsSecretValueModified := false

	queryStringKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringIsSecretValueModified := false

	resourceName := "aws_cloudwatch_event_connection.oauth"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_oauth(
					name,
					description,
					authorizationType,
					authorizationEndpoint,
					httpMethod,
					clientID,
					clientSecret,
					bodyKey,
					bodyValue,
					bodyIsSecretValue,
					headerKey,
					headerValue,
					headerIsSecretValue,
					queryStringKey,
					queryStringValue,
					queryStringIsSecretValue,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.authorization_endpoint", authorizationEndpoint),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.http_method", httpMethod),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.client_parameters.0.client_id", clientID),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.key", bodyKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.key", headerKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.key", queryStringKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValue)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auth_parameters.0.oauth.0.client_parameters.0.client_secret",
					"auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.value",
					"auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.value",
					"auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.value",
				},
			},
			{
				Config: testAccConnectionConfig_oauth(
					nameModified,
					descriptionModified,
					authorizationType,
					authorizationEndpointModified,
					httpMethodModified,
					clientIDModified,
					clientSecretModified,
					bodyKeyModified,
					bodyValueModified,
					bodyIsSecretValueModified,
					headerKeyModified,
					headerValueModified,
					headerIsSecretValueModified,
					queryStringKeyModified,
					queryStringValueModified,
					queryStringIsSecretValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v2),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf("connection/%s/%s", nameModified, uuidRegex))),
					testAccCheckConnectionRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.authorization_endpoint", authorizationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.client_parameters.0.client_id", clientIDModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.key", bodyKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
				),
			},
			{
				Config: testAccConnectionConfig_oauth(
					nameModified,
					descriptionModified,
					authorizationType,
					authorizationEndpointModified,
					httpMethodModified,
					clientIDModified,
					clientSecretModified,
					bodyKeyModified,
					bodyValueModified,
					bodyIsSecretValueModified,
					headerKeyModified,
					headerValueModified,
					headerIsSecretValueModified,
					queryStringKeyModified,
					queryStringValueModified,
					queryStringIsSecretValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v3),
					testAccCheckConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", nameModified),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.authorization_endpoint", authorizationEndpointModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.http_method", httpMethodModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.client_parameters.0.client_id", clientIDModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.key", bodyKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
				),
			},
		},
	})
}

func TestAccEventsConnection_invocationHTTPParameters(t *testing.T) {
	var v1, v2, v3 eventbridge.DescribeConnectionOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authorizationType := "API_KEY"
	description := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// invocation_http_parameters
	bodyKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bodyValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bodyIsSecretValue := true

	headerKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerIsSecretValue := true

	queryStringKey := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringValue := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringIsSecretValue := true

	// invocation_http_parameters modified
	bodyKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bodyValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bodyIsSecretValueModified := false

	headerKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	headerIsSecretValueModified := false

	queryStringKeyModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringValueModified := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	queryStringIsSecretValueModified := false

	resourceName := "aws_cloudwatch_event_connection.invocation_http_parameters"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_invocationHTTPParameters(
					name,
					description,
					authorizationType,
					key,
					value,
					bodyKey,
					bodyValue,
					bodyIsSecretValue,
					headerKey,
					headerValue,
					headerIsSecretValue,
					queryStringKey,
					queryStringValue,
					queryStringIsSecretValue,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.key", bodyKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.key", fmt.Sprintf("second-%s", bodyKey)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.is_value_secret", strconv.FormatBool(bodyIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.key", headerKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.key", fmt.Sprintf("second-%s", headerKey)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.is_value_secret", strconv.FormatBool(headerIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.key", queryStringKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.key", fmt.Sprintf("second-%s", queryStringKey)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.is_value_secret", strconv.FormatBool(queryStringIsSecretValue)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auth_parameters.0.api_key.0.value",
					"auth_parameters.0.invocation_http_parameters.0.body.0.value",
					"auth_parameters.0.invocation_http_parameters.0.body.1.value",
					"auth_parameters.0.invocation_http_parameters.0.header.0.value",
					"auth_parameters.0.invocation_http_parameters.0.header.1.value",
					"auth_parameters.0.invocation_http_parameters.0.query_string.0.value",
					"auth_parameters.0.invocation_http_parameters.0.query_string.1.value",
				},
			},
			{
				Config: testAccConnectionConfig_invocationHTTPParameters(
					name,
					description,
					authorizationType,
					key,
					value,
					bodyKeyModified,
					bodyValueModified,
					bodyIsSecretValueModified,
					headerKeyModified,
					headerValueModified,
					headerIsSecretValueModified,
					queryStringKeyModified,
					queryStringValueModified,
					queryStringIsSecretValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v2),
					testAccCheckConnectionNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.key", bodyKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.key", fmt.Sprintf("second-%s", bodyKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.key", fmt.Sprintf("second-%s", headerKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.key", fmt.Sprintf("second-%s", queryStringKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
				),
			},
			{
				Config: testAccConnectionConfig_invocationHTTPParameters(
					name,
					description,
					authorizationType,
					key,
					value,
					bodyKeyModified,
					bodyValueModified,
					bodyIsSecretValueModified,
					headerKeyModified,
					headerValueModified,
					headerIsSecretValueModified,
					queryStringKeyModified,
					queryStringValueModified,
					queryStringIsSecretValueModified,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v3),
					testAccCheckConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.key", bodyKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.key", fmt.Sprintf("second-%s", bodyKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.body.1.is_value_secret", strconv.FormatBool(bodyIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.key", headerKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.key", fmt.Sprintf("second-%s", headerKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.header.1.is_value_secret", strconv.FormatBool(headerIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.key", queryStringKeyModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.key", fmt.Sprintf("second-%s", queryStringKeyModified)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.0.query_string.1.is_value_secret", strconv.FormatBool(queryStringIsSecretValueModified)),
				),
			},
		},
	})
}

func TestAccEventsConnection_disappears(t *testing.T) {
	var v eventbridge.DescribeConnectionOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authorizationType := "API_KEY"
	description := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	value := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_connection.api_key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, eventbridge.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_apiKey(
					name,
					description,
					authorizationType,
					key,
					value,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfevents.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConnectionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_connection" {
			continue
		}

		_, err := tfevents.FindConnectionByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EventBridge connection %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckConnectionExists(n string, v *eventbridge.DescribeConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EventsConn

		output, err := tfevents.FindConnectionByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectionRecreated(i, j *eventbridge.DescribeConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ConnectionArn) == aws.StringValue(j.ConnectionArn) {
			return fmt.Errorf("EventBridge Connection not recreated")
		}
		return nil
	}
}

func testAccCheckConnectionNotRecreated(i, j *eventbridge.DescribeConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ConnectionArn) != aws.StringValue(j.ConnectionArn) {
			return fmt.Errorf("EventBridge Connection was recreated")
		}
		return nil
	}
}

func testAccConnectionConfig_apiKey(name, description, authorizationType, key, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "api_key" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  auth_parameters {
    api_key {
      key   = %[4]q
      value = %[5]q
    }
  }
}
`, name,
		description,
		authorizationType,
		key,
		value)
}

func testAccConnectionConfig_basic(name, description, authorizationType, username, password string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "basic" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  auth_parameters {
    basic {
      username = %[4]q
      password = %[5]q
    }
  }
}
`, name,
		description,
		authorizationType,
		username,
		password)
}

func testAccConnectionConfig_oauth(
	name,
	description,
	authorizationType,
	authorizationEndpoint,
	httpMethod,
	clientID,
	clientSecret string,
	bodyKey string,
	bodyValue string,
	bodyIsSecretValue bool,
	headerKey string,
	headerValue string,
	headerIsSecretValue bool,
	queryStringKey string,
	queryStringValue string,
	queryStringIsSecretValue bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "oauth" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  auth_parameters {
    oauth {
      authorization_endpoint = %[4]q
      http_method            = %[5]q
      client_parameters {
        client_id     = %[6]q
        client_secret = %[7]q
      }

      oauth_http_parameters {
        body {
          key             = %[8]q
          value           = %[9]q
          is_value_secret = %[10]t
        }

        header {
          key             = %[11]q
          value           = %[12]q
          is_value_secret = %[13]t
        }

        query_string {
          key             = %[14]q
          value           = %[15]q
          is_value_secret = %[16]t
        }
      }
    }
  }
}
`, name,
		description,
		authorizationType,
		authorizationEndpoint,
		httpMethod,
		clientID,
		clientSecret,
		bodyKey,
		bodyValue,
		bodyIsSecretValue,
		headerKey,
		headerValue,
		headerIsSecretValue,
		queryStringKey,
		queryStringValue,
		queryStringIsSecretValue)
}

func testAccConnectionConfig_invocationHTTPParameters(
	name,
	description,
	authorizationType,
	key,
	value string,
	bodyKey string,
	bodyValue string,
	bodyIsSecretValue bool,
	headerKey string,
	headerValue string,
	headerIsSecretValue bool,
	queryStringKey string,
	queryStringValue string,
	queryStringIsSecretValue bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "invocation_http_parameters" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  auth_parameters {
    api_key {
      key   = %[4]q
      value = %[5]q
    }

    invocation_http_parameters {
      body {
        key             = %[6]q
        value           = %[7]q
        is_value_secret = %[8]t
      }

      body {
        key             = "second-%[6]s"
        value           = "second-%[7]s"
        is_value_secret = %[8]t
      }

      header {
        key             = %[9]q
        value           = %[10]q
        is_value_secret = %[11]t
      }

      header {
        key             = "second-%[9]s"
        value           = "second-%[10]s"
        is_value_secret = %[11]t
      }

      query_string {
        key             = %[12]q
        value           = %[13]q
        is_value_secret = %[14]t
      }

      query_string {
        key             = "second-%[12]s"
        value           = "second-%[13]s"
        is_value_secret = %[14]t
      }
    }
  }
}
`, name,
		description,
		authorizationType,
		key,
		value,
		bodyKey,
		bodyValue,
		bodyIsSecretValue,
		headerKey,
		headerValue,
		headerIsSecretValue,
		queryStringKey,
		queryStringValue,
		queryStringIsSecretValue)
}
