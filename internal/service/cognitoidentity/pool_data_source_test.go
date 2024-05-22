// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIdentityPoolDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ip cognitoidentity.DescribeIdentityPoolOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_identity_pool.test"
	resourceName := "aws_cognito_identity_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CognitoIdentityEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &ip),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "developer_provider_name", resourceName, "developer_provider_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allow_unauthenticated_identities", resourceName, "allow_unauthenticated_identities"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allow_classic_flow", resourceName, "allow_classic_flow"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPoolDataSource_openidConnectProviderARNs(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ip cognitoidentity.DescribeIdentityPoolOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_identity_pool.test"
	resourceName := "aws_cognito_identity_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CognitoIdentityEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_openidConnectProviderARNs(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &ip),
					resource.TestCheckResourceAttrPair(dataSourceName, "openid_connect_provider_arns", resourceName, "openid_connect_provider_arns"),
					resource.TestCheckResourceAttr(dataSourceName, "openid_connect_provider_arns.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPoolDataSource_cognitoIdentityProviders(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	name := sdkacctest.RandString(10)
	var ip cognitoidentity.DescribeIdentityPoolOutput
	dataSourceName := "data.aws_cognito_identity_pool.test"
	resourceName := "aws_cognito_identity_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CognitoIdentityEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_identityProviders(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &ip),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "cognito_identity_providers.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPoolDataSource_samlProviderARNs(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	name := sdkacctest.RandString(10)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	var ip cognitoidentity.DescribeIdentityPoolOutput
	dataSourceName := "data.aws_cognito_identity_pool.test"
	resourceName := "aws_cognito_identity_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CognitoIdentityEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_samlProviderARNs(name, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &ip),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "saml_provider_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "saml_provider_arns.0", "aws_iam_saml_provider.default", names.AttrARN)),
			},
		},
	})
}

func TestAccCognitoIdentityPoolDataSource_supportedLoginProviders(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	name := sdkacctest.RandString(10)
	var ip cognitoidentity.DescribeIdentityPoolOutput
	dataSourceName := "data.aws_cognito_identity_pool.test"
	resourceName := "aws_cognito_identity_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CognitoIdentityEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_supportedLoginProviders(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &ip),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "supported_login_providers.%", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "supported_login_providers.graph.facebook.com", "7346241598935555")),
			},
		},
	})
}

func TestAccCognitoIdentityPoolDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ip cognitoidentity.DescribeIdentityPoolOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_identity_pool.test"
	resourceName := "aws_cognito_identity_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CognitoIdentityEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIdentityServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_tags(name, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &ip),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccPoolDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "test" {
  identity_pool_name               = "%s"
  allow_unauthenticated_identities = false
}

data "aws_cognito_identity_pool" "test" {
  identity_pool_name = aws_cognito_identity_pool.test.identity_pool_name
}
	`, rName)
}

func testAccPoolDataSourceConfig_openidConnectProviderARNs(rName string) string {
	return acctest.ConfigCompose(testAccPoolConfig_openidConnectProviderARNs(rName), `
data "aws_cognito_identity_pool" "test" {
  identity_pool_name = aws_cognito_identity_pool.test.identity_pool_name
}
`)
}

func testAccPoolDataSourceConfig_identityProviders(rName string) string {
	return acctest.ConfigCompose(testAccPoolConfig_identityProviders(rName), `
data "aws_cognito_identity_pool" "test" {
  identity_pool_name = aws_cognito_identity_pool.test.identity_pool_name
}
`)
}

func testAccPoolDataSourceConfig_samlProviderARNs(name, idpEntityId string) string {
	return acctest.ConfigCompose(testAccPoolConfig_samlProviderARNs(name, idpEntityId), `
data "aws_cognito_identity_pool" "test" {
  identity_pool_name = aws_cognito_identity_pool.test.identity_pool_name
}
`)
}

func testAccPoolDataSourceConfig_supportedLoginProviders(name string) string {
	return acctest.ConfigCompose(testAccPoolConfig_supportedLoginProviders(name), `
data "aws_cognito_identity_pool" "test" {
  identity_pool_name = aws_cognito_identity_pool.test.identity_pool_name
}
`)
}

func testAccPoolDataSourceConfig_tags(name, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccPoolConfig_tags1(name, tagKey1, tagValue1), `
data "aws_cognito_identity_pool" "test" {
  identity_pool_name = aws_cognito_identity_pool.test.identity_pool_name
}
`)
}
