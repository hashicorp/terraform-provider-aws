package cognitoidentity_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccCognitoIdentityPool_basic(t *testing.T) {
	var v1, v2 cognitoidentity.IdentityPool
	name := sdkacctest.RandString(10)
	updatedName := sdkacctest.RandString(10)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cognito-identity", regexp.MustCompile(`identitypool/.+`)),
					resource.TestCheckResourceAttr(resourceName, "allow_unauthenticated_identities", "false"),
					resource.TestCheckResourceAttr(resourceName, "developer_provider_name", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolConfig_basic(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v2),
					testAccCheckPoolRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", updatedName)),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPool_DeveloperProviderName(t *testing.T) {
	var v1, v2 cognitoidentity.IdentityPool
	name := sdkacctest.RandString(10)
	developerProviderName := sdkacctest.RandString(10)
	developerProviderNameUpdated := sdkacctest.RandString(10)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_DeveloperProviderName(name, developerProviderName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "developer_provider_name", developerProviderName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolConfig_DeveloperProviderName(name, developerProviderNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v2),
					testAccCheckPoolRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "developer_provider_name", developerProviderNameUpdated),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPool_supportedLoginProviders(t *testing.T) {
	var v1, v2, v3 cognitoidentity.IdentityPool
	name := sdkacctest.RandString(10)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_supportedLoginProviders(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "supported_login_providers.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "supported_login_providers.graph.facebook.com", "7346241598935555"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolConfig_supportedLoginProvidersModified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v2),
					testAccCheckPoolNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "supported_login_providers.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "supported_login_providers.graph.facebook.com", "7346241598935552"),
					resource.TestCheckResourceAttr(resourceName, "supported_login_providers.accounts.google.com", "123456789012.apps.googleusercontent.com"),
				),
			},
			{
				Config: testAccPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v3),
					testAccCheckPoolNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "supported_login_providers.%", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPool_openidConnectProviderARNs(t *testing.T) {
	var v1, v2, v3 cognitoidentity.IdentityPool
	name := sdkacctest.RandString(10)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_openidConnectProviderARNs(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_provider_arns.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolConfig_openidConnectProviderARNsModified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v2),
					testAccCheckPoolNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_provider_arns.#", "2"),
				),
			},
			{
				Config: testAccPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v3),
					testAccCheckPoolNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_provider_arns.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPool_samlProviderARNs(t *testing.T) {
	var v1, v2, v3 cognitoidentity.IdentityPool
	name := sdkacctest.RandString(10)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	secondaryIdpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_samlProviderARNs(name, idpEntityId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "saml_provider_arns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "saml_provider_arns.0", "aws_iam_saml_provider.default", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolConfig_samlProviderARNsModified(name, idpEntityId, secondaryIdpEntityId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v2),
					testAccCheckPoolNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "saml_provider_arns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "saml_provider_arns.0", "aws_iam_saml_provider.secondary", "arn"),
				),
			},
			{
				Config: testAccPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v3),
					testAccCheckPoolNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "saml_provider_arns.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPool_cognitoIdentityProviders(t *testing.T) {
	var v1, v2, v3 cognitoidentity.IdentityPool
	name := sdkacctest.RandString(10)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_cognitoIdentityProviders(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cognito_identity_providers.*", map[string]string{
						"client_id":               "7lhlkkfbfb4q5kpp90urffao",
						"provider_name":           fmt.Sprintf("cognito-idp.%[1]s.%[2]s/%[1]s_Zr231apJu", acctest.Region(), acctest.PartitionDNSSuffix()),
						"server_side_token_check": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cognito_identity_providers.*", map[string]string{
						"client_id":               "7lhlkkfbfb4q5kpp90urffao",
						"provider_name":           fmt.Sprintf("cognito-idp.%[1]s.%[2]s/%[1]s_Ab129faBb", acctest.Region(), acctest.PartitionDNSSuffix()),
						"server_side_token_check": "false",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolConfig_cognitoIdentityProvidersModified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v2),
					testAccCheckPoolNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cognito_identity_providers.*", map[string]string{
						"client_id":               "6lhlkkfbfb4q5kpp90urffae",
						"provider_name":           fmt.Sprintf("cognito-idp.%[1]s.%[2]s/%[1]s_Zr231apJu", acctest.Region(), acctest.PartitionDNSSuffix()),
						"server_side_token_check": "false",
					}),
				),
			},
			{
				Config: testAccPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v3),
					testAccCheckPoolNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPool_addingNewProviderKeepsOldProvider(t *testing.T) {
	var v1, v2, v3 cognitoidentity.IdentityPool
	name := sdkacctest.RandString(10)
	resourceName := "aws_cognito_identity_pool.main"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_cognitoIdentityProviders(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolConfig_cognitoIdentityProvidersAndOpenidConnectProviderARNs(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v2),
					testAccCheckPoolNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_provider_arns.#", "1"),
				),
			},
			{
				Config: testAccPoolConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v3),
					testAccCheckPoolNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "identity_pool_name", fmt.Sprintf("identity pool %s", name)),
					resource.TestCheckResourceAttr(resourceName, "cognito_identity_providers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "openid_connect_provider_arns.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIdentityPool_tags(t *testing.T) {
	var v1, v2, v3 cognitoidentity.IdentityPool
	name := sdkacctest.RandString(10)
	resourceName := "aws_cognito_identity_pool.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_Tags1(name, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPoolConfig_Tags2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v2),
					testAccCheckPoolNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccPoolConfig_Tags1(name, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(resourceName, &v3),
					testAccCheckPoolNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckPoolExists(n string, identityPool *cognitoidentity.IdentityPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Cognito Identity Pool ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIdentityConn

		result, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
			IdentityPoolId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("Cognito Identity Pool (%s) not found", rs.Primary.ID)
		}

		*identityPool = *result

		return err
	}
}

func testAccCheckPoolDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIdentityConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_identity_pool" {
			continue
		}

		_, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
			IdentityPoolId: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, cognitoidentity.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIdentityConn

	input := &cognitoidentity.ListIdentityPoolsInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.ListIdentityPools(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckPoolRecreated(i, j *cognitoidentity.IdentityPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if poolIdentityEqual(i, j) {
			return fmt.Errorf("Cognito Identity Pool not recreated")
		}
		return nil
	}
}

func testAccCheckPoolNotRecreated(i, j *cognitoidentity.IdentityPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !poolIdentityEqual(i, j) {
			return fmt.Errorf("Cognito Identity Pool recreated")
		}
		return nil
	}
}

func poolIdentity(v *cognitoidentity.IdentityPool) string {
	return aws.StringValue(v.IdentityPoolId)
}

func poolIdentityEqual(i, j *cognitoidentity.IdentityPool) bool {
	return poolIdentity(i) == poolIdentity(j)
}

func testAccPoolConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false
}
`, name)
}

func testAccPoolConfig_DeveloperProviderName(name, developerProviderName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %[1]s"
  allow_unauthenticated_identities = false
  developer_provider_name          = %[2]q
}
`, name, developerProviderName)
}

func testAccPoolConfig_supportedLoginProviders(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  supported_login_providers = {
    "graph.facebook.com" = "7346241598935555"
  }
}
`, name)
}

func testAccPoolConfig_supportedLoginProvidersModified(name string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  supported_login_providers = {
    "graph.facebook.com"  = "7346241598935552"
    "accounts.google.com" = "123456789012.apps.googleusercontent.com"
  }
}
`, name)
}

func testAccPoolConfig_openidConnectProviderARNs(name string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  openid_connect_provider_arns = ["arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/server.example.com"]
}
`, name)
}

func testAccPoolConfig_openidConnectProviderARNsModified(name string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  openid_connect_provider_arns = ["arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/modified-1.example.com", "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/modified-2.example.com"]
}
`, name)
}

func testAccPoolConfig_samlProviderARNs(name, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "myprovider-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %[1]s"
  allow_unauthenticated_identities = false

  saml_provider_arns = [aws_iam_saml_provider.default.arn]
}
`, name, idpEntityId)
}

func testAccPoolConfig_samlProviderARNsModified(name, idpEntityId, secondaryIdpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "default" {
  name                   = "default-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}

resource "aws_iam_saml_provider" "secondary" {
  name                   = "secondary-%[1]s"
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[3]q })
}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %[1]s"
  allow_unauthenticated_identities = false

  saml_provider_arns = [aws_iam_saml_provider.secondary.arn]
}
`, name, idpEntityId, secondaryIdpEntityId)
}

func testAccPoolConfig_cognitoIdentityProviders(name string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  cognito_identity_providers {
    client_id               = "7lhlkkfbfb4q5kpp90urffao"
    provider_name           = "cognito-idp.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${data.aws_region.current.name}_Ab129faBb"
    server_side_token_check = false
  }

  cognito_identity_providers {
    client_id               = "7lhlkkfbfb4q5kpp90urffao"
    provider_name           = "cognito-idp.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${data.aws_region.current.name}_Zr231apJu"
    server_side_token_check = false
  }
}
`, name)
}

func testAccPoolConfig_cognitoIdentityProvidersModified(name string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  cognito_identity_providers {
    client_id               = "6lhlkkfbfb4q5kpp90urffae"
    provider_name           = "cognito-idp.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${data.aws_region.current.name}_Zr231apJu"
    server_side_token_check = false
  }
}
`, name)
}

func testAccPoolConfig_cognitoIdentityProvidersAndOpenidConnectProviderARNs(name string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool %s"
  allow_unauthenticated_identities = false

  cognito_identity_providers {
    client_id               = "7lhlkkfbfb4q5kpp90urffao"
    provider_name           = "cognito-idp.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${data.aws_region.current.name}_Ab129faBb"
    server_side_token_check = false
  }

  cognito_identity_providers {
    client_id               = "7lhlkkfbfb4q5kpp90urffao"
    provider_name           = "cognito-idp.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${data.aws_region.current.name}_Zr231apJu"
    server_side_token_check = false
  }

  openid_connect_provider_arns = ["arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/server.example.com"]
}
`, name)
}

func testAccPoolConfig_Tags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = %q
  allow_unauthenticated_identities = false

  tags = {
    %q = %q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccPoolConfig_Tags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = %q
  allow_unauthenticated_identities = false

  tags = {
    %q = %q
    %q = %q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}
