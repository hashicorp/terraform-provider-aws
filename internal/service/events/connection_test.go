// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsConnection_apiKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeConnectionOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	authorizationType := "API_KEY"
	description := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	nameModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	descriptionModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	keyModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	valueModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_cloudwatch_event_connection.api_key"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v2),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf("connection/%s/%s", nameModified, uuidRegex))),
					testAccCheckConnectionRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, nameModified),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionModified),
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v3),
					testAccCheckConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, nameModified),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.0.key", keyModified),
				),
			},
		},
	})
}

func TestAccEventsConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 eventbridge.DescribeConnectionOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	authorizationType := "BASIC"
	description := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	username := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	password := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	nameModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	descriptionModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	usernameModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	passwordModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_connection.basic"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_basic(
					name,
					description,
					authorizationType,
					username,
					password,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.password", password),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.username", username),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "invocation_connectivity_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, "kms_key_identifier", ""),
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v2),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf("connection/%s/%s", nameModified, uuidRegex))),
					testAccCheckConnectionRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.password", passwordModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.basic.0.username", usernameModified),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.invocation_http_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", authorizationType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionModified),
					resource.TestCheckResourceAttr(resourceName, "invocation_connectivity_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, nameModified),
					resource.TestCheckResourceAttr(resourceName, "kms_key_identifier", ""),
				),
			},
		},
	})
}

func TestAccEventsConnection_oAuth(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeConnectionOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	authorizationType := "OAUTH_CLIENT_CREDENTIALS"
	description := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	// oauth
	authorizationEndpoint := "https://example.com/auth"
	httpMethod := "POST"

	// client_parameters
	clientID := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clientSecret := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// oauth_http_parameters
	bodyKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyValue := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyIsSecretValue := true

	headerKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerValue := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerIsSecretValue := true

	queryStringKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringValue := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringIsSecretValue := true

	// modified
	nameModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	descriptionModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	// oauth
	authorizationEndpointModified := "https://example.com/auth-modified"
	httpMethodModified := "GET"

	// client_parameters
	clientIDModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clientSecretModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// oauth_http_parameters modified
	bodyKeyModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyValueModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyIsSecretValueModified := false

	headerKeyModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerValueModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerIsSecretValueModified := false

	queryStringKeyModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringValueModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringIsSecretValueModified := false

	resourceName := "aws_cloudwatch_event_connection.oauth"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_oauthHTTPParametersEmpty(
					nameModified,
					descriptionModified,
					authorizationType,
					authorizationEndpointModified,
					httpMethod,
				),
				ExpectError: regexache.MustCompile("Missing required argument"),
			},
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v2),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "events", regexache.MustCompile(fmt.Sprintf("connection/%s/%s", nameModified, uuidRegex))),
					testAccCheckConnectionRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, nameModified),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionModified),
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v3),
					testAccCheckConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, nameModified),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionModified),
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

func TestAccEventsConnection_connectivityParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 eventbridge.DescribeConnectionOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	oAuthorizationType := "OAUTH_CLIENT_CREDENTIALS"

	description := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// oauth
	authorizationEndpoint := "https://example.com/auth"
	httpMethod := "POST"

	// client_parameters
	clientID := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	clientSecret := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// oauth_http_parameters
	bodyKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyValue := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyIsSecretValue := true

	headerKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerValue := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerIsSecretValue := true

	queryStringKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringValue := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringIsSecretValue := true

	resourceName := "aws_cloudwatch_event_connection.oauth"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_oauthConnectivityParameters(
					name,
					description,
					oAuthorizationType,
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", oAuthorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.authorization_endpoint", authorizationEndpoint),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.http_method", httpMethod),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.client_parameters.0.client_id", clientID),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.key", bodyKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.body.0.is_value_secret", strconv.FormatBool(bodyIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.key", headerKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.header.0.is_value_secret", strconv.FormatBool(headerIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.key", queryStringKey),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.oauth.0.oauth_http_parameters.0.query_string.0.is_value_secret", strconv.FormatBool(queryStringIsSecretValue)),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.connectivity_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.connectivity_parameters.0.resource_parameters.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "auth_parameters.0.connectivity_parameters.0.resource_parameters.0.resource_configuration_arn", "aws_vpclattice_resource_configuration.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "invocation_connectivity_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "invocation_connectivity_parameters.0.resource_parameters.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "invocation_connectivity_parameters.0.resource_parameters.0.resource_configuration_arn", "aws_vpclattice_resource_configuration.test", names.AttrARN),
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
				Config: testAccConnectionConfig_oauthConnectivityParameters(
					name,
					description,
					oAuthorizationType,
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
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v2),
					testAccCheckConnectionNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", oAuthorizationType),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.connectivity_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auth_parameters.0.connectivity_parameters.0.resource_parameters.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "auth_parameters.0.connectivity_parameters.0.resource_parameters.0.resource_configuration_arn", "aws_vpclattice_resource_configuration.test2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "invocation_connectivity_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "invocation_connectivity_parameters.0.resource_parameters.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "invocation_connectivity_parameters.0.resource_parameters.0.resource_configuration_arn", "aws_vpclattice_resource_configuration.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccEventsConnection_invocationHTTPParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 eventbridge.DescribeConnectionOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	authorizationType := "API_KEY"
	description := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	// invocation_http_parameters
	bodyKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyValue := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyIsSecretValue := true

	headerKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerValue := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerIsSecretValue := true

	queryStringKey := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringValue := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringIsSecretValue := true

	// invocation_http_parameters modified
	bodyKeyModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyValueModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bodyIsSecretValueModified := false

	headerKeyModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerValueModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	headerIsSecretValueModified := false

	queryStringKeyModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringValueModified := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	queryStringIsSecretValueModified := false

	resourceName := "aws_cloudwatch_event_connection.invocation_http_parameters"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccConnectionConfig_invocationHTTPParametersEmpty(name, description),
				ExpectError: regexache.MustCompile("Missing required argument"),
			},
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v2),
					testAccCheckConnectionNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v3),
					testAccCheckConnectionNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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
	ctx := acctest.Context(t)
	var v eventbridge.DescribeConnectionOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	authorizationType := "API_KEY"
	description := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	value := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_connection.api_key"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
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
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfevents.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEventsConnection_invocationConnectivityParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeConnectionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_connection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_invocationConnectivityParameters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "invocation_connectivity_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "invocation_connectivity_parameters.0.resource_parameters.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_parameters.0.basic.0.password"},
			},
		},
	})
}

func TestAccEventsConnection_kmsKeyIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridge.DescribeConnectionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_connection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig_kmsKeyIdentifier(rName, "${aws_kms_key.test_1.id}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_key.test_1", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_parameters.0.basic.0.password"},
			},
			{
				Config: testAccConnectionConfig_kmsKeyIdentifier(rName, "${aws_kms_key.test_2.arn}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_key.test_2", names.AttrARN),
				),
			},
			{
				Config: testAccConnectionConfig_kmsKeyIdentifier(rName, "${aws_kms_alias.test_1.name}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_alias.test_1", names.AttrName),
				),
			},
			{
				Config: testAccConnectionConfig_kmsKeyIdentifier(rName, "${aws_kms_alias.test_1.arn}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_identifier", "aws_kms_alias.test_1", names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckConnectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_event_connection" {
				continue
			}

			_, err := tfevents.FindConnectionByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Connection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConnectionExists(ctx context.Context, t *testing.T, n string, v *eventbridge.DescribeConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		output, err := tfevents.FindConnectionByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectionRecreated(i, j *eventbridge.DescribeConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.ConnectionArn) == aws.ToString(j.ConnectionArn) {
			return fmt.Errorf("EventBridge Connection not recreated")
		}
		return nil
	}
}

func testAccCheckConnectionNotRecreated(i, j *eventbridge.DescribeConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.ConnectionArn) != aws.ToString(j.ConnectionArn) {
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

func testAccConnectionConfig_invocationHTTPParametersEmpty(name, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "invocation_http_parameters" {
  name        = %[1]q
  description = %[2]q
  auth_parameters {
    invocation_http_parameters {
    }
  }
}
`, name, description)
}

func testAccConnectionConfig_oauthHTTPParametersEmpty(
	name,
	description,
	authorizationType,
	authorizationEndpoint,
	httpMethod string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_connection" "oauth" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  auth_parameters {
    oauth {
      authorization_endpoint = %[4]q
      http_method            = %[5]q
      oauth_http_parameters {
      }
    }
  }
}
`, name, description, authorizationType, authorizationEndpoint, httpMethod)
}

func testAccConnectionConfig_oauthConnectivityParameters(
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
	queryStringIsSecretValue bool,
	updated ...bool) string {
	useUpdated := len(updated) > 0 && updated[0]

	resourceConfigRef := "aws_vpclattice_resource_configuration.test"
	additionalResourceConfig := ""
	if useUpdated {
		resourceConfigRef = "aws_vpclattice_resource_configuration.test2"
		additionalResourceConfig = fmt.Sprintf(`
resource "aws_vpclattice_resource_configuration" "test2" {
  name = "%s-updated"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["443"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}
`, name)
	}

	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(name, 1), additionalResourceConfig, fmt.Sprintf(`
resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_cloudwatch_event_connection" "oauth" {
  name               = %[1]q
  description        = %[2]q
  authorization_type = %[3]q
  invocation_connectivity_parameters {
    resource_parameters {
      resource_configuration_arn = %[17]s.arn
    }
  }
  auth_parameters {
    connectivity_parameters {
      resource_parameters {
        resource_configuration_arn = %[17]s.arn
      }
    }
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
		queryStringIsSecretValue,
		resourceConfigRef))
}

func testAccConnectionConfig_invocationConnectivityParameters(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_cloudwatch_event_connection" "test" {
  name               = %[1]q
  authorization_type = "BASIC"

  auth_parameters {
    basic {
      username = "tfacctest"
      password = "avoid-plaintext-passwords"
    }
  }

  invocation_connectivity_parameters {
    resource_parameters {
      resource_configuration_arn = aws_vpclattice_resource_configuration.test.arn
    }
  }
}
`, rName))
}

func testAccConnectionConfig_kmsKeyIdentifier(name, kmsKeyIdentifier string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_kms_key" "test_1" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "key-policy-example"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow use of the key"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        Action = [
          "kms:DescribeKey",
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ],
        Resource = "*"
        Condition = {
          StringLike = {
            "kms:ViaService" = "secretsmanager.*.amazonaws.com"
            "kms:EncryptionContext:SecretARN" = [
              "arn:${data.aws_partition.current.partition}:secretsmanager:*:*:secret:events!connection/*"
            ]
          }
        }
      }
    ]
  })
  tags = {
    EventBridgeApiDestinations = "true"
  }
}

resource "aws_kms_alias" "test_1" {
  name          = "alias/test-1"
  target_key_id = aws_kms_key.test_1.key_id
}

resource "aws_kms_key" "test_2" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "key-policy-example"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow use of the key"
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        Action = [
          "kms:DescribeKey",
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ],
        Resource = "*"
        Condition = {
          StringLike = {
            "kms:ViaService" = "secretsmanager.*.amazonaws.com"
            "kms:EncryptionContext:SecretARN" = [
              "arn:${data.aws_partition.current.partition}:secretsmanager:*:*:secret:events!connection/*"
            ]
          }
        }
      }
    ]
  })
  tags = {
    EventBridgeApiDestinations = "true"
  }
}

resource "aws_cloudwatch_event_connection" "test" {
  name               = %[1]q
  authorization_type = "BASIC"
  auth_parameters {
    basic {
      username = "tfacctest"
      password = "avoid-plaintext-passwords"
    }
  }
  kms_key_identifier = %[2]q
}
`, name, kmsKeyIdentifier)
}
