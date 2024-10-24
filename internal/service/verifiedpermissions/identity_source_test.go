// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfverifiedpermissions "github.com/hashicorp/terraform-provider-aws/internal/service/verifiedpermissions"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedPermissionsIdentitySource_Cognito_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var identitySource verifiedpermissions.GetIdentitySourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_identity_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentitySourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentitySourceConfig_Cognito_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "policy_store_id", "aws_verifiedpermissions_policy_store.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "principal_entity_type", "AWS::Cognito"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.client_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.group_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.cognito_user_pool_configuration.0.user_pool_arn", "aws_cognito_user_pool.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIdentitySourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVerifiedPermissionsIdentitySource_OpenID_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var identitySource verifiedpermissions.GetIdentitySourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_identity_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentitySourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentitySourceConfig_OpenID_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "policy_store_id", "aws_verifiedpermissions_policy_store.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "principal_entity_type", "MyCorp::User"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.entity_id_prefix"),
					testAccCheckPairAsHTTPSURL(resourceName, "configuration.0.open_id_connect_configuration.0.issuer", "aws_cognito_user_pool.test", names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.0", "https://myapp.example.com"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.principal_id_claim", "sub"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.identity_token_only.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIdentitySourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVerifiedPermissionsIdentitySource_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var identitySource verifiedpermissions.GetIdentitySourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_identity_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentitySourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentitySourceConfig_Cognito_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfverifiedpermissions.ResourceIdentitySource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVerifiedPermissionsIdentitySource_Cognito_update(t *testing.T) {
	ctx := acctest.Context(t)

	var identitySource verifiedpermissions.GetIdentitySourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_identity_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentitySourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentitySourceConfig_Cognito_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "policy_store_id", "aws_verifiedpermissions_policy_store.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "principal_entity_type", "AWS::Cognito"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.client_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.group_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.cognito_user_pool_configuration.0.user_pool_arn", "aws_cognito_user_pool.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct0),
				),
			},
			{
				Config: testAccIdentitySourceConfig_Cognito_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource),
					resource.TestCheckResourceAttrPair(resourceName, "policy_store_id", "aws_verifiedpermissions_policy_store.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "principal_entity_type", "AWS::Cognito"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.client_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.cognito_user_pool_configuration.0.client_ids.0", "aws_cognito_user_pool_client.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.group_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.cognito_user_pool_configuration.0.user_pool_arn", "aws_cognito_user_pool.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct0),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIdentitySourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVerifiedPermissionsIdentitySource_OpenID_update(t *testing.T) {
	ctx := acctest.Context(t)

	var identitySource verifiedpermissions.GetIdentitySourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_identity_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentitySourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentitySourceConfig_OpenID_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.entity_id_prefix"),
					testAccCheckPairAsHTTPSURL(resourceName, "configuration.0.open_id_connect_configuration.0.issuer", "aws_cognito_user_pool.test", names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.0", "https://myapp.example.com"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.principal_id_claim", "sub"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.identity_token_only.#", acctest.Ct0),
				),
			},
			{
				Config: testAccIdentitySourceConfig_updateOpenIDConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.entity_id_prefix", "MyOIDCProvider"),
					testAccCheckPairAsHTTPSURL(resourceName, "configuration.0.open_id_connect_configuration.0.issuer", "aws_cognito_user_pool.test", names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.0.group_claim", "groups"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.0.group_entity_type", "MyCorp::UserGroup"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.0", "https://myapp.example.com"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.principal_id_claim", "sub"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.identity_token_only.#", acctest.Ct0),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIdentitySourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccIdentitySourceConfig_updateOpenIDConfigurationTokenSelection(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.entity_id_prefix", "MyOIDCProvider"),
					testAccCheckPairAsHTTPSURL(resourceName, "configuration.0.open_id_connect_configuration.0.issuer", "aws_cognito_user_pool.test", names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.0.group_claim", "groups"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.0.group_entity_type", "MyCorp::UserGroup"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.identity_token_only.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.identity_token_only.0.client_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.identity_token_only.0.client_ids.0", "1example23456789"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.identity_token_only.0.principal_id_claim", "sub"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIdentitySourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// The AWS Console claims that changing the identity provider type requires recreating the identity source,
// but this does not seem to be the case using the API
func TestAccVerifiedPermissionsIdentitySource_Cognito_convertToOpenID(t *testing.T) {
	ctx := acctest.Context(t)

	var identitySource1, identitySource2 verifiedpermissions.GetIdentitySourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_identity_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentitySourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentitySourceConfig_Cognito_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "policy_store_id", "aws_verifiedpermissions_policy_store.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "principal_entity_type", "AWS::Cognito"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.client_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.group_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.cognito_user_pool_configuration.0.user_pool_arn", "aws_cognito_user_pool.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct0),
				),
			},
			{
				Config: testAccIdentitySourceConfig_OpenID_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource2),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "policy_store_id", "aws_verifiedpermissions_policy_store.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "principal_entity_type", "MyCorp::User"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.entity_id_prefix"),
					testAccCheckPairAsHTTPSURL(resourceName, "configuration.0.open_id_connect_configuration.0.issuer", "aws_cognito_user_pool.test", names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.0", "https://myapp.example.com"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.principal_id_claim", "sub"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.identity_token_only.#", acctest.Ct0),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIdentitySourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// The AWS Console claims that changing the identity provider type requires recreating the identity source,
// but this does not seem to be the case using the API
func TestAccVerifiedPermissionsIdentitySource_OpenID_convertToCognito(t *testing.T) {
	ctx := acctest.Context(t)

	var identitySource1, identitySource2 verifiedpermissions.GetIdentitySourceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_identity_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentitySourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentitySourceConfig_OpenID_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "policy_store_id", "aws_verifiedpermissions_policy_store.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "principal_entity_type", "MyCorp::User"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.entity_id_prefix"),
					testAccCheckPairAsHTTPSURL(resourceName, "configuration.0.open_id_connect_configuration.0.issuer", "aws_cognito_user_pool.test", names.AttrEndpoint),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.group_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.audiences.0", "https://myapp.example.com"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.access_token_only.0.principal_id_claim", "sub"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.0.token_selection.0.identity_token_only.#", acctest.Ct0),
				),
			},
			{
				Config: testAccIdentitySourceConfig_Cognito_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentitySourceExists(ctx, resourceName, &identitySource2),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "policy_store_id", "aws_verifiedpermissions_policy_store.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "principal_entity_type", "AWS::Cognito"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.client_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cognito_user_pool_configuration.0.group_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.cognito_user_pool_configuration.0.user_pool_arn", "aws_cognito_user_pool.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.open_id_connect_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIdentitySourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIdentitySourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedpermissions_identity_source" {
				continue
			}

			_, err := tfverifiedpermissions.FindIdentitySourceByIDAndPolicyStoreID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["policy_store_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingDestroyed, tfverifiedpermissions.ResNameIdentitySource, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckIdentitySourceExists(ctx context.Context, name string, identitySource *verifiedpermissions.GetIdentitySourceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNameIdentitySource, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNameIdentitySource, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)
		resp, err := tfverifiedpermissions.FindIdentitySourceByIDAndPolicyStoreID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["policy_store_id"])

		if err != nil {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNameIdentitySource, rs.Primary.ID, err)
		}

		*identitySource = *resp

		return nil
	}
}

func testAccIdentitySourceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["policy_store_id"], rs.Primary.ID), nil
	}
}

func testAccCheckPairAsHTTPSURL(nameFirst, keyFirst, nameSecond, keySecond string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		first, err := acctest.PrimaryInstanceState(s, nameFirst)
		if err != nil {
			return err
		}

		vFirst, ok := first.Attributes[keyFirst]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", nameFirst, keyFirst)
		}

		second, err := acctest.PrimaryInstanceState(s, nameSecond)
		if err != nil {
			return err
		}

		vSecond, ok := second.Attributes[keySecond]
		if !ok {
			return fmt.Errorf("%s: No attribute %q found", nameSecond, keySecond)
		}

		vSecond = fmt.Sprintf("https://%s", vSecond)

		if vFirst != vSecond {
			return fmt.Errorf(
				"%s: Attribute '%s' expected %#v, got %#v",
				nameFirst,
				keyFirst,
				vSecond,
				vFirst)
		}

		return nil
	}
}

func testAccIdentitySourceConfig_base() string {
	return `
resource "aws_verifiedpermissions_policy_store" "test" {
  description = "Terraform acceptance test"
  validation_settings {
    mode = "OFF"
  }
}
`
}

func testAccIdentitySourceConfig_Cognito_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccIdentitySourceConfig_base(),
		fmt.Sprintf(`
resource "aws_verifiedpermissions_identity_source" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id
  configuration {
    cognito_user_pool_configuration {
      user_pool_arn = aws_cognito_user_pool.test.arn
    }
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName))
}

func testAccIdentitySourceConfig_Cognito_update(rName string) string {
	return acctest.ConfigCompose(
		testAccIdentitySourceConfig_base(),
		fmt.Sprintf(`
resource "aws_verifiedpermissions_identity_source" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id
  configuration {
    cognito_user_pool_configuration {
      user_pool_arn = aws_cognito_user_pool.test.arn
      client_ids    = [aws_cognito_user_pool_client.test.id]
    }
  }
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name                = %[1]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}
`, rName))
}

func testAccIdentitySourceConfig_OpenID_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccIdentitySourceConfig_base(),
		fmt.Sprintf(`
resource "aws_verifiedpermissions_identity_source" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id
  configuration {
    open_id_connect_configuration {
      issuer = "https://${aws_cognito_user_pool.test.endpoint}"
      token_selection {
        access_token_only {
          audiences          = ["https://myapp.example.com"]
          principal_id_claim = "sub"
        }
      }
    }
  }
  principal_entity_type = "MyCorp::User"
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName))
}

func testAccIdentitySourceConfig_updateOpenIDConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccIdentitySourceConfig_base(),
		fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_verifiedpermissions_identity_source" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id
  configuration {
    open_id_connect_configuration {
      issuer = "https://${aws_cognito_user_pool.test.endpoint}"
      token_selection {
        access_token_only {
          audiences          = ["https://myapp.example.com"]
          principal_id_claim = "sub"
        }
      }
      entity_id_prefix = "MyOIDCProvider"
      group_configuration {
        group_claim       = "groups"
        group_entity_type = "MyCorp::UserGroup"
      }
    }
  }
  principal_entity_type = "MyCorp::User"
}
`, rName))
}

func testAccIdentitySourceConfig_updateOpenIDConfigurationTokenSelection(rName string) string {
	return acctest.ConfigCompose(
		testAccIdentitySourceConfig_base(),
		fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_verifiedpermissions_identity_source" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id
  configuration {
    open_id_connect_configuration {
      issuer = "https://${aws_cognito_user_pool.test.endpoint}"
      token_selection {
        identity_token_only {
          client_ids         = ["1example23456789"]
          principal_id_claim = "sub"
        }
      }
      entity_id_prefix = "MyOIDCProvider"
      group_configuration {
        group_claim       = "groups"
        group_entity_type = "MyCorp::UserGroup"
      }
    }
  }
  principal_entity_type = "MyCorp::User"
}
`, rName))
}
