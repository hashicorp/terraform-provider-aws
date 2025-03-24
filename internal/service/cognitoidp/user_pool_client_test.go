// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	oauthScopeEmail                     = "email"
	oauthScopeOpenID                    = "openid"
	oauthScopePhone                     = "phone"
	oauthScopeProfile                   = "profile"
	oauthScopeAWSCognitoSignInUserAdmin = "aws.cognito.signin.user.admin" // nosemgrep:ci.aws-in-const-name,ci.aws-in-var-name
)

func TestAccCognitoIDPUserPoolClient_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "access_token_validity", "0"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows_user_pool_client", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auth_session_validity", "3"),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientSecret, ""),
					resource.TestCheckResourceAttr(resourceName, "default_redirect_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_propagate_additional_user_context_data", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_token_revocation", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "generate_secret"),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "0"),
					resource.TestCheckResourceAttr(resourceName, "prevent_user_existence_errors", ""),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "30"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserPoolID, "aws_cognito_user_pool.test", names.AttrID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_flows"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_scopes"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("analytics_configuration"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("callback_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("explicit_auth_flows"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("ADMIN_NO_SRP_AUTH"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logout_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("read_attributes"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("supported_identity_providers"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_validity_units"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("write_attributes"), knownvalue.SetExact([]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_enableRevocation(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_revocation(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "enable_token_revocation", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_revocation(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "enable_token_revocation", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_accessTokenValidity(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_accessTokenValidity(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "access_token_validity", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_accessTokenValidity(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "access_token_validity", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_accessTokenValidity_error(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserPoolClientConfig_accessTokenValidity(rName, 25),
				ExpectError: regexache.MustCompile(`Attribute access_token_validity must have a duration between 5m0s and\s+24h0m0s, got: 25h0m0s`),
			},
			{
				Config:      testAccUserPoolClientConfig_accessTokenValidityUnit(rName, 2, string(awstypes.TimeUnitsTypeDays)),
				ExpectError: regexache.MustCompile(`Attribute access_token_validity must have a duration between 5m0s and\s+24h0m0s, got: 48h0m0s`),
			},
			{
				Config:      testAccUserPoolClientConfig_accessTokenValidityUnit(rName, 4, string(awstypes.TimeUnitsTypeMinutes)),
				ExpectError: regexache.MustCompile(`Attribute access_token_validity must have a duration between 5m0s and\s+24h0m0s, got: 4m0s`),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_idTokenValidity(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_idTokenValidity(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_idTokenValidity(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_idTokenValidity_error(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserPoolClientConfig_idTokenValidity(rName, 25),
				ExpectError: regexache.MustCompile(`Attribute id_token_validity must have a duration between 5m0s and\s+24h0m0s,\s+got: 25h0m0s`),
			},
			{
				Config:      testAccUserPoolClientConfig_idTokenValidityUnit(rName, 2, string(awstypes.TimeUnitsTypeDays)),
				ExpectError: regexache.MustCompile(`Attribute id_token_validity must have a duration between 5m0s and\s+24h0m0s,\s+got: 48h0m0s`),
			},
			{
				Config:      testAccUserPoolClientConfig_idTokenValidityUnit(rName, 4, string(awstypes.TimeUnitsTypeMinutes)),
				ExpectError: regexache.MustCompile(`Attribute id_token_validity must have a duration between 5m0s and\s+24h0m0s,\s+got: 4m0s`),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_refreshTokenValidity(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_refreshTokenValidity(rName, 60),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "60"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_refreshTokenValidity(rName, 120),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_refreshTokenValidity_error(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserPoolClientConfig_refreshTokenValidity(rName, 10*365+1),
				ExpectError: regexache.MustCompile(`Attribute refresh_token_validity must have a duration between 1h0m0s and\s+87600h0m0s,\s+got: 87624h0m0s`),
			},
			{
				Config:      testAccUserPoolClientConfig_refreshTokenValidityUnit(rName, 59, string(awstypes.TimeUnitsTypeMinutes)),
				ExpectError: regexache.MustCompile(`Attribute refresh_token_validity must have a duration between 1h0m0s and\s+87600h0m0s,\s+got: 59m0s`),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_tokenValidityUnits(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnits(rName, "days"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_validity_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"access_token":  knownvalue.StringExact("days"),
							"id_token":      knownvalue.StringExact("days"),
							"refresh_token": knownvalue.StringExact("days"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnits(rName, "hours"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_validity_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"access_token":  knownvalue.StringExact("hours"),
							"id_token":      knownvalue.StringExact("hours"),
							"refresh_token": knownvalue.StringExact("hours"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_validity_units"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_tokenValidityUnits_explicitDefaults(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnits_explicitDefaults(rName, "days"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_validity_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"access_token":  knownvalue.StringExact("hours"),
							"id_token":      knownvalue.StringExact("hours"),
							"refresh_token": knownvalue.StringExact("days"),
						}),
					})),
				},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_tokenValidityUnits_AccessToken(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnits_Unit(rName, "access_token", "days"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_validity_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"access_token":  knownvalue.StringExact("days"),
							"id_token":      knownvalue.StringExact("hours"),
							"refresh_token": knownvalue.StringExact("days"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnits(rName, "hours"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_validity_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"access_token":  knownvalue.StringExact("hours"),
							"id_token":      knownvalue.StringExact("hours"),
							"refresh_token": knownvalue.StringExact("hours"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_tokenValidityUnitsWTokenValidity(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnitsTokenValidity(rName, "days"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_validity_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"access_token":  knownvalue.StringExact("days"),
							"id_token":      knownvalue.StringExact("days"),
							"refresh_token": knownvalue.StringExact("days"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_tokenValidityUnitsTokenValidity(rName, "hours"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "id_token_validity", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_validity_units"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"access_token":  knownvalue.StringExact("hours"),
							"id_token":      knownvalue.StringExact("hours"),
							"refresh_token": knownvalue.StringExact("hours"),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_name(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_name(rName, "name1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_name(rName, "name2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "name2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_allFields(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_allFields(rName, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "generate_secret", acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, names.AttrClientSecret, regexache.MustCompile(`\w+`)),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "300"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows_user_pool_client", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_redirect_uri", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "prevent_user_existence_errors", "LEGACY"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_flows"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("code"),
						knownvalue.StringExact("implicit"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_scopes"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(oauthScopeEmail),
						knownvalue.StringExact(oauthScopeOpenID),
						knownvalue.StringExact(oauthScopePhone),
						knownvalue.StringExact(oauthScopeProfile),
						knownvalue.StringExact(oauthScopeAWSCognitoSignInUserAdmin),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("analytics_configuration"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("callback_urls"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("https://www.example.com/callback"),
						knownvalue.StringExact("https://www.example.com/redirect"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("explicit_auth_flows"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("CUSTOM_AUTH_FLOW_ONLY"),
						knownvalue.StringExact("USER_PASSWORD_AUTH"),
						knownvalue.StringExact("ADMIN_NO_SRP_AUTH"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logout_urls"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("https://www.example.com/login"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("read_attributes"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(names.AttrEmail),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("write_attributes"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(names.AttrEmail),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"generate_secret",
				},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_allFieldsUpdatingOneField(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_allFields(rName, 300),
			},
			{
				Config: testAccUserPoolClientConfig_allFields(rName, 299),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "generate_secret", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "refresh_token_validity", "299"),
					resource.TestCheckResourceAttr(resourceName, "allowed_oauth_flows_user_pool_client", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_redirect_uri", "https://www.example.com/redirect"),
					resource.TestCheckResourceAttr(resourceName, "prevent_user_existence_errors", "LEGACY"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_flows"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("code"),
						knownvalue.StringExact("implicit"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_scopes"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(oauthScopeEmail),
						knownvalue.StringExact(oauthScopeOpenID),
						knownvalue.StringExact(oauthScopePhone),
						knownvalue.StringExact(oauthScopeProfile),
						knownvalue.StringExact(oauthScopeAWSCognitoSignInUserAdmin),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("callback_urls"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("https://www.example.com/callback"),
						knownvalue.StringExact("https://www.example.com/redirect"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("explicit_auth_flows"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("CUSTOM_AUTH_FLOW_ONLY"),
						knownvalue.StringExact("USER_PASSWORD_AUTH"),
						knownvalue.StringExact("ADMIN_NO_SRP_AUTH"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logout_urls"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("https://www.example.com/login"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("read_attributes"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(names.AttrEmail),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("write_attributes"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(names.AttrEmail),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"generate_secret",
				},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_analyticsApplicationID(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"
	pinpointResourceName := "aws_pinpoint_app.analytics"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIdentityProvider(ctx, t)
			acctest.PreCheckPinpointApp(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_analyticsApplicationID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("analytics_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrExternalID: knownvalue.StringExact(rName),
							"user_data_shared":   knownvalue.Bool(false),
							"application_arn":    knownvalue.Null(),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("analytics_configuration").AtSliceIndex(0).AtMapKey(names.AttrApplicationID), pinpointResourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("analytics_configuration").AtSliceIndex(0).AtMapKey(names.AttrRoleARN), "aws_iam_role.analytics", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_analyticsShareData(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("analytics_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrExternalID: knownvalue.StringExact(rName),
							"user_data_shared":   knownvalue.Bool(true),
							"application_arn":    knownvalue.Null(),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("analytics_configuration").AtSliceIndex(0).AtMapKey(names.AttrApplicationID), pinpointResourceName, tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("analytics_configuration").AtSliceIndex(0).AtMapKey(names.AttrRoleARN), "aws_iam_role.analytics", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("analytics_configuration"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_analyticsWithARN(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"
	pinpointResourceName := "aws_pinpoint_app.analytics"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIdentityProvider(ctx, t)
			acctest.PreCheckPinpointApp(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_analyticsARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("analytics_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrExternalID:    knownvalue.Null(),
							"user_data_shared":      knownvalue.Bool(false),
							names.AttrApplicationID: knownvalue.Null(),
							names.AttrRoleARN:       acctest.GlobalARN("iam", "role/aws-service-role/cognito-idp.amazonaws.com/AWSServiceRoleForAmazonCognitoIdp"),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("analytics_configuration").AtSliceIndex(0).AtMapKey("application_arn"), pinpointResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_analyticsARNShareData(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("analytics_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrExternalID:    knownvalue.Null(),
							"user_data_shared":      knownvalue.Bool(true),
							names.AttrApplicationID: knownvalue.Null(),
							names.AttrRoleARN:       acctest.GlobalARN("iam", "role/aws-service-role/cognito-idp.amazonaws.com/AWSServiceRoleForAmazonCognitoIdp"),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("analytics_configuration").AtSliceIndex(0).AtMapKey("application_arn"), pinpointResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_authSessionValidity(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_authSessionValidity(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "auth_session_validity", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPoolClientConfig_authSessionValidity(rName, 15),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "auth_session_validity", "15"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUserPoolClient, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_Disappears_userPool(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceUserPool(), "aws_cognito_user_pool.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_emptySets(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_emptySets(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_flows"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_scopes"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("callback_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("explicit_auth_flows"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logout_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("read_attributes"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("supported_identity_providers"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("write_attributes"), knownvalue.SetExact([]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccUserPoolClientConfig_nulls(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_nulls(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_nulls(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_flows"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_scopes"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("callback_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("explicit_auth_flows"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logout_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("read_attributes"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("supported_identity_providers"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("write_attributes"), knownvalue.SetExact([]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccUserPoolClientConfig_emptySets(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_frameworkMigration_nulls(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		CheckDestroy: testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.59.0",
					},
				},
				Config: testAccUserPoolClientConfig_nulls(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_flows"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_scopes"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("callback_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("explicit_auth_flows"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logout_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("read_attributes"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("supported_identity_providers"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("write_attributes"), knownvalue.Null()),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccUserPoolClientConfig_nulls(rName),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_frameworkMigration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		CheckDestroy: testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.59.0",
					},
				},
				Config: testAccUserPoolClientConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccUserPoolClientConfig_basic(rName),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_frameworkMigration_emptySet(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		CheckDestroy: testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.59.0",
					},
				},
				Config: testAccUserPoolClientConfig_emptySets(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_flows"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_oauth_scopes"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("callback_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("explicit_auth_flows"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("logout_urls"), knownvalue.SetExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("read_attributes"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("supported_identity_providers"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("write_attributes"), knownvalue.Null()),
				},
			},
			/*
				TODO Investigate...
				{
					ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
					Config:                   testAccUserPoolClientConfig_emptySets(rName),
					PlanOnly:                 true,
				},
			*/
		},
	})
}

func TestAccCognitoIDPUserPoolClient_supportedIdentityProviders_utf8(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"
	identityProvider := strings.Repeat("あ", 32)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_supportedIdentityProviders_utf8(rName, identityProvider),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "supported_identity_providers.0", identityProvider),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCognitoIDPUserPoolClient_urls_utf8(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.UserPoolClientType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_user_pool_client.test"
	defaultRedirectURI := "https://www.example.com/redirect/" + strings.Repeat("あ", 991)
	logoutURL := "https://www.example.com/logout/" + strings.Repeat("あ", 993)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolClientDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolClientConfig_urls_utf8(rName, defaultRedirectURI, logoutURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserPoolClientExists(ctx, resourceName, &client),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.0", defaultRedirectURI),
					resource.TestCheckResourceAttr(resourceName, "default_redirect_uri", defaultRedirectURI),
					resource.TestCheckResourceAttr(resourceName, "logout_urls.0", logoutURL),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccUserPoolClientImportStateIDFunc(ctx, resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserPoolClientImportStateIDFunc(ctx context.Context, n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		userPoolID := rs.Primary.Attributes[names.AttrUserPoolID]
		clientID := rs.Primary.ID
		_, err := tfcognitoidp.FindUserPoolClientByTwoPartKey(ctx, conn, userPoolID, clientID)

		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s/%s", userPoolID, clientID), nil
	}
}

func testAccCheckUserPoolClientDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_user_pool_client" {
				continue
			}

			_, err := tfcognitoidp.FindUserPoolClientByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito User Pool Client %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserPoolClientExists(ctx context.Context, n string, v *awstypes.UserPoolClientType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindUserPoolClientByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccUserPoolClientConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccUserPoolClientConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name                = %[1]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}
`, rName))
}

func testAccUserPoolClientConfig_revocation(rName string, revoke bool) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name                    = %[1]q
  user_pool_id            = aws_cognito_user_pool.test.id
  explicit_auth_flows     = ["ADMIN_NO_SRP_AUTH"]
  enable_token_revocation = %[2]t
}
`, rName, revoke))
}

func testAccUserPoolClientConfig_accessTokenValidity(rName string, validity int) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  access_token_validity = %[2]d
}
`, rName, validity))
}

func testAccUserPoolClientConfig_accessTokenValidityUnit(rName string, validity int, unit string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  access_token_validity = %[2]d

  token_validity_units {
    access_token = %[3]q
  }
}
`, rName, validity, unit))
}

func testAccUserPoolClientConfig_idTokenValidity(rName string, validity int) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  id_token_validity = %[2]d
}
`, rName, validity))
}

func testAccUserPoolClientConfig_idTokenValidityUnit(rName string, validity int, unit string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  id_token_validity = %[2]d

  token_validity_units {
    id_token = %[3]q
  }
}
`, rName, validity, unit))
}

func testAccUserPoolClientConfig_refreshTokenValidity(rName string, refreshTokenValidity int) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  refresh_token_validity = %[2]d
}
`, rName, refreshTokenValidity))
}

func testAccUserPoolClientConfig_refreshTokenValidityUnit(rName string, refreshTokenValidity int, unit string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  refresh_token_validity = %[2]d

  token_validity_units {
    refresh_token = %[3]q
  }
}
`, rName, refreshTokenValidity, unit))
}

func testAccUserPoolClientConfig_tokenValidityUnits(rName, value string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  token_validity_units {
    access_token  = %[2]q
    id_token      = %[2]q
    refresh_token = %[2]q
  }
}
`, rName, value))
}

func testAccUserPoolClientConfig_tokenValidityUnits_Unit(rName, unit, value string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  token_validity_units {
    %[2]s = %[3]q
  }
}
`, rName, unit, value))
}

func testAccUserPoolClientConfig_tokenValidityUnits_explicitDefaults(rName, value string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  token_validity_units {
    access_token  = "hours"
    id_token      = "hours"
    refresh_token = "days"
  }
}
`, rName, value))
}

func testAccUserPoolClientConfig_tokenValidityUnitsTokenValidity(rName, units string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
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
`, rName, units))
}

func testAccUserPoolClientConfig_name(rName, name string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}
`, name))
}

func testAccUserPoolClientConfig_allFields(rName string, refreshTokenValidity int) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
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
`, rName, refreshTokenValidity))
}

func testAccUserPoolClientConfig_baseAnalytics(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_pinpoint_app" "analytics" {
  name = %[1]q
}

resource "aws_iam_role" "analytics" {
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

resource "aws_iam_role_policy" "analytics" {
  name = %[1]q
  role = aws_iam_role.analytics.id

  policy = <<-EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "mobiletargeting:UpdateEndpoint",
        "mobiletargeting:PutEvents"
      ],
      "Effect": "Allow",
      "Resource": "arn:${data.aws_partition.current.partition}:mobiletargeting:*:${data.aws_caller_identity.current.account_id}:apps/${aws_pinpoint_app.analytics.application_id}*"
    }
  ]
}
EOF
}
`, rName))
}

func testAccUserPoolClientConfig_analyticsApplicationID(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_baseAnalytics(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_id = aws_pinpoint_app.analytics.application_id
    external_id    = %[1]q
    role_arn       = aws_iam_role.analytics.arn
  }
}
`, rName))
}

func testAccUserPoolClientConfig_analyticsShareData(rName string, share bool) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_baseAnalytics(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_id   = aws_pinpoint_app.analytics.application_id
    external_id      = %[1]q
    role_arn         = aws_iam_role.analytics.arn
    user_data_shared = %[2]t
  }
}
`, rName, share))
}

func testAccUserPoolClientConfig_analyticsARN(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_baseAnalytics(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_arn = aws_pinpoint_app.analytics.arn
  }
}
`, rName))
}

func testAccUserPoolClientConfig_analyticsARNShareData(rName string, share bool) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_baseAnalytics(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_arn  = aws_pinpoint_app.analytics.arn
    user_data_shared = %[2]t
  }
}
`, rName, share))
}

func testAccUserPoolClientConfig_authSessionValidity(rName string, validity int) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name                  = %[1]q
  auth_session_validity = %[2]d
  user_pool_id          = aws_cognito_user_pool.test.id
  explicit_auth_flows   = ["ADMIN_NO_SRP_AUTH"]
}
`, rName, validity))
}

func testAccUserPoolClientConfig_emptySets(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  allowed_oauth_flows          = []
  allowed_oauth_scopes         = []
  callback_urls                = []
  explicit_auth_flows          = []
  logout_urls                  = []
  read_attributes              = []
  supported_identity_providers = []
  write_attributes             = []
}
`, rName))
}

func testAccUserPoolClientConfig_nulls(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName))
}

func testAccUserPoolClientConfig_supportedIdentityProviders_utf8(rName, identityProvider string) string {
	return acctest.ConfigCompose(
		testAccIdentityProviderConfig_saml(identityProvider, rName, acctest.CtTrue),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  supported_identity_providers = [%[2]q]

  depends_on = [aws_cognito_identity_provider.test]
}
`, rName, identityProvider))
}

func testAccUserPoolClientConfig_urls_utf8(rName, defaultRedirectUri, logoutUrl string) string {
	return acctest.ConfigCompose(
		testAccUserPoolClientConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool_client" "test" {
  name         = %[1]q
  user_pool_id = aws_cognito_user_pool.test.id

  callback_urls        = [%[2]q]
  default_redirect_uri = %[2]q
  logout_urls          = [%[3]q]
}
`, rName, defaultRedirectUri, logoutUrl))
}
